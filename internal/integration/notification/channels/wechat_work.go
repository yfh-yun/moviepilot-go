package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"moviepilot-go/internal/integration/notification"
)

// WeChatWorkConfig 企业微信通知配置
type WeChatWorkConfig struct {
	CorpID  string `json:"corp_id"`  // 企业ID
	AgentID string `json:"agent_id"` // 应用AgentID
	Secret  string `json:"secret"`   // 应用Secret
	ToUser  string `json:"to_user"`  // 指定接收用户，@all表示所有人
	MsgType string `json:"msg_type"` // 消息类型，支持text、markdown等
}

// WeChatWorkSender 企业微信通知发送器
type WeChatWorkSender struct {
	config       *WeChatWorkConfig
	httpClient   *http.Client
	accessToken  string
	tokenExpires time.Time
}

// NewWeChatWorkSender 创建新的企业微信通知发送器
func NewWeChatWorkSender(config *WeChatWorkConfig) *WeChatWorkSender {
	return &WeChatWorkSender{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name 返回发送器名称
func (w *WeChatWorkSender) Name() string {
	return "wechat-work"
}

// SupportedLevels 返回支持的通知级别
func (w *WeChatWorkSender) SupportedLevels() []notification.NotificationLevel {
	return []notification.NotificationLevel{
		notification.LevelInfo,
		notification.LevelWarning,
		notification.LevelError,
		notification.LevelSuccess,
	}
}

// getAccessToken 获取访问令牌
func (w *WeChatWorkSender) getAccessToken(ctx context.Context) (string, error) {
	if w.accessToken != "" && time.Now().Before(w.tokenExpires) {
		return w.accessToken, nil
	}

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

	var tokenResp struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("wechat work API error: %s (code: %d)", tokenResp.ErrMsg, tokenResp.ErrCode)
	}

	w.accessToken = tokenResp.AccessToken
	w.tokenExpires = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return w.accessToken, nil
}

// Validate 验证消息
func (w *WeChatWorkSender) Validate(message *notification.Message) error {
	if message.Title == "" && message.Content == "" {
		return fmt.Errorf("message title and content cannot both be empty")
	}
	return nil
}

// Send 发送企业微信通知
func (w *WeChatWorkSender) Send(ctx context.Context, message *notification.Message) error {
	accessToken, err := w.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	msgBody := w.buildMessageBody(message)
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

	var msgResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if msgResp.ErrCode != 0 {
		return fmt.Errorf("wechat work API error: %s (code: %d)", msgResp.ErrMsg, msgResp.ErrCode)
	}

	return nil
}

// buildMessageBody 构建消息体
func (w *WeChatWorkSender) buildMessageBody(message *notification.Message) string {
	msgType := w.config.MsgType
	if msgType == "" {
		msgType = "text"
	}

	content := message.Content
	if message.Title != "" {
		content = fmt.Sprintf("%s\n\n%s", message.Title, message.Content)
	}

	// 添加级别标识
	switch message.Level {
	case notification.LevelError:
		content = "🚨 " + content
	case notification.LevelWarning:
		content = "⚠️ " + content
	case notification.LevelSuccess:
		content = "✅ " + content
	}

	msgBody := map[string]interface{}{
		"touser":  w.config.ToUser,
		"msgtype": msgType,
		"agentid": w.config.AgentID,
	}

	switch msgType {
	case "text":
		msgBody["text"] = map[string]string{
			"content": content,
		}
	case "markdown":
		msgBody["markdown"] = map[string]string{
			"content": content,
		}
	default:
		msgBody["text"] = map[string]string{
			"content": content,
		}
	}

	jsonData, _ := json.Marshal(msgBody)
	return string(jsonData)
}

// HealthCheck 健康检查
func (w *WeChatWorkSender) HealthCheck(ctx context.Context) error {
	_, err := w.getAccessToken(ctx)
	return err
}

// Close 关闭发送器
func (w *WeChatWorkSender) Close() error {
	w.accessToken = ""
	w.tokenExpires = time.Time{}
	return nil
}

// CreateWeChatWorkChannel 创建企业微信通知渠道
func CreateWeChatWorkChannel(config *WeChatWorkConfig) *notification.Channel {
	return &notification.Channel{
		Name:        "wechat-work",
		Description: "企业微信通知渠道",
		Enabled:     config.CorpID != "" && config.AgentID != "" && config.Secret != "",
		Sender:      NewWeChatWorkSender(config),
		Config: map[string]string{
			"corp_id":  config.CorpID,
			"agent_id": config.AgentID,
			"msg_type": config.MsgType,
		},
	}
}
