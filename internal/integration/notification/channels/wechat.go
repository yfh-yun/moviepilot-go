package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/notification"
)

// WeChatConfig 微信通知配置
type WeChatConfig struct {
	CorpID  string `json:"corp_id"`  // 企业ID
	AgentID string `json:"agent_id"` // 应用AgentID
	Secret  string `json:"secret"`   // 应用Secret
	ToUser  string `json:"to_user"`  // 指定接收用户，多个用 | 分隔
	ToParty string `json:"to_party"` // 指定接收部门，多个用 | 分隔
	ToTag   string `json:"to_tag"`   // 指定接收标签，多个用 | 分隔
	Safe    int    `json:"safe"`     // 是否保密消息，0表示否，1表示是
	MsgType string `json:"msg_type"` // 消息类型，支持text、markdown、image等
}

// WeChatSender 微信通知发送器
type WeChatSender struct {
	config       *WeChatConfig
	httpClient   *http.Client
	accessToken  string
	tokenExpires time.Time
}

// AccessTokenResponse 微信API访问令牌响应
type AccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// MessageResponse 发送消息响应
type MessageResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgID   string `json:"msgid"`
}

// NewWeChatSender 创建新的微信通知发送器
func NewWeChatSender(config *WeChatConfig) *WeChatSender {
	return &WeChatSender{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name 返回发送器名称
func (w *WeChatSender) Name() string {
	return "wechat"
}

// SupportedLevels 返回支持的通知级别
func (w *WeChatSender) SupportedLevels() []notification.NotificationLevel {
	return []notification.NotificationLevel{
		notification.LevelInfo,
		notification.LevelWarning,
		notification.LevelError,
		notification.LevelSuccess,
	}
}

// getAccessToken 获取访问令牌
func (w *WeChatSender) getAccessToken(ctx context.Context) (string, error) {
	// 检查令牌是否有效
	if w.accessToken != "" && time.Now().Before(w.tokenExpires) {
		return w.accessToken, nil
	}

	// 构建URL
	apiURL := fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		url.QueryEscape(w.config.CorpID),
		url.QueryEscape(w.config.Secret),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("wechat API error: %s (code: %d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	// 更新令牌和过期时间
	w.accessToken = tokenResp.AccessToken
	w.tokenExpires = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second) // 提前5分钟过期

	return w.accessToken, nil
}

// Validate 验证消息是否适合此发送器
func (w *WeChatSender) Validate(message *notification.Message) error {
	if message.Title == "" && message.Content == "" {
		return fmt.Errorf("message title and content cannot both be empty")
	}

	// 检查消息长度限制
	if len(message.Content) > 2048 {
		return fmt.Errorf("message content too long (max 2048 characters)")
	}

	return nil
}

// Send 发送微信通知
func (w *WeChatSender) Send(ctx context.Context, message *notification.Message) error {
	accessToken, err := w.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// 构建消息体
	msgBody := w.buildMessageBody(message)

	// 发送消息
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", accessToken)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(msgBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var msgResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if msgResp.ErrCode != 0 {
		return fmt.Errorf("wechat API error: %s (code: %d)", msgResp.ErrMsg, msgResp.ErrCode)
	}

	return nil
}

// buildMessageBody 构建消息体
func (w *WeChatSender) buildMessageBody(message *notification.Message) string {
	msgType := w.config.MsgType
	if msgType == "" {
		msgType = "text"
	}

	content := message.Content
	if message.Title != "" {
		content = fmt.Sprintf("%s\n\n%s", message.Title, message.Content)
	}

	// 根据消息级别添加表情
	switch message.Level {
	case notification.LevelError:
		content = "🚨 " + content
	case notification.LevelWarning:
		content = "⚠️ " + content
	case notification.LevelSuccess:
		content = "✅ " + content
	case notification.LevelInfo:
		content = "ℹ️ " + content
	}

	msgBody := map[string]interface{}{
		"touser":  w.config.ToUser,
		"toparty": w.config.ToParty,
		"totag":   w.config.ToTag,
		"msgtype": msgType,
		"agentid": w.config.AgentID,
		"safe":    w.config.Safe,
	}

	// 根据消息类型设置具体内容
	switch msgType {
	case "text":
		msgBody["text"] = map[string]string{
			"content": content,
		}
	case "markdown":
		msgBody["markdown"] = map[string]string{
			"content": content,
		}
	case "textcard":
		msgBody["textcard"] = map[string]string{
			"title":       message.Title,
			"description": message.Content,
			"url":         message.LinkURL,
			"btntxt":      "查看详情",
		}
	case "news":
		articles := []map[string]string{
			{
				"title":       message.Title,
				"description": message.Content,
				"url":         message.LinkURL,
				"picurl":      message.ImageURL,
			},
		}
		msgBody["news"] = map[string]interface{}{
			"articles": articles,
		}
	default:
		// 默认为文本消息
		msgBody["text"] = map[string]string{
			"content": content,
		}
	}

	jsonData, _ := json.Marshal(msgBody)
	return string(jsonData)
}

// HealthCheck 检查发送器健康状态
func (w *WeChatSender) HealthCheck(ctx context.Context) error {
	_, err := w.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("wechat channel unhealthy: %w", err)
	}
	return nil
}

// Close 关闭发送器
func (w *WeChatSender) Close() error {
	w.accessToken = ""
	w.tokenExpires = time.Time{}
	return nil
}

// CreateWeChatChannel 创建微信通知渠道
func CreateWeChatChannel(config *WeChatConfig) *notification.Channel {
	return &notification.Channel{
		Name:        "wechat",
		Description: "微信企业号通知渠道",
		Enabled:     config.CorpID != "" && config.AgentID != "" && config.Secret != "",
		Sender:      NewWeChatSender(config),
		Config: map[string]string{
			"corp_id":  config.CorpID,
			"agent_id": config.AgentID,
			"msg_type": config.MsgType,
		},
	}
}
