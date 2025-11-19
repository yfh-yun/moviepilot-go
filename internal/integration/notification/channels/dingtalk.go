package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/notification"
)

// DingTalkConfig 钉钉通知配置
type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"` // 钉钉群机器人Webhook URL
	Secret     string `json:"secret"`      // 安全密钥（可选）
	MsgType    string `json:"msg_type"`    // 消息类型，支持text、markdown等
}

// DingTalkSender 钉钉通知发送器
type DingTalkSender struct {
	config     *DingTalkConfig
	httpClient *http.Client
}

// NewDingTalkSender 创建新的钉钉通知发送器
func NewDingTalkSender(config *DingTalkConfig) *DingTalkSender {
	return &DingTalkSender{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name 返回发送器名称
func (d *DingTalkSender) Name() string {
	return "dingtalk"
}

// SupportedLevels 返回支持的通知级别
func (d *DingTalkSender) SupportedLevels() []notification.NotificationLevel {
	return []notification.NotificationLevel{
		notification.LevelInfo,
		notification.LevelWarning,
		notification.LevelError,
		notification.LevelSuccess,
	}
}

// Validate 验证消息
func (d *DingTalkSender) Validate(message *notification.Message) error {
	if message.Title == "" && message.Content == "" {
		return fmt.Errorf("message title and content cannot both be empty")
	}

	if len(message.Content) > 5000 {
		return fmt.Errorf("message content too long (max 5000 characters)")
	}

	return nil
}

// Send 发送钉钉通知
func (d *DingTalkSender) Send(ctx context.Context, message *notification.Message) error {
	if d.config.WebhookURL == "" {
		return fmt.Errorf("dingtalk webhook URL is not configured")
	}

	// 构建消息体
	msgBody := d.buildMessageBody(message)

	req, err := http.NewRequestWithContext(ctx, "POST", d.config.WebhookURL, strings.NewReader(msgBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.ErrCode != 0 {
		return fmt.Errorf("dingtalk API error: %s (code: %d)", apiResp.ErrMsg, apiResp.ErrCode)
	}

	return nil
}

// buildMessageBody 构建消息体
func (d *DingTalkSender) buildMessageBody(message *notification.Message) string {
	msgType := d.config.MsgType
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
		"msgtype": msgType,
	}

	switch msgType {
	case "text":
		msgBody["text"] = map[string]string{
			"content": content,
		}
	case "markdown":
		markdownContent := fmt.Sprintf("### %s\n\n%s", message.Title, message.Content)
		if message.LinkURL != "" {
			markdownContent += fmt.Sprintf("\n\n[查看详情](%s)", message.LinkURL)
		}
		msgBody["markdown"] = map[string]string{
			"title": message.Title,
			"text":  markdownContent,
		}
	case "actionCard":
		actionCard := map[string]string{
			"title":          message.Title,
			"text":           content,
			"hideAvatar":     "0",
			"btnOrientation": "0",
		}

		if message.LinkURL != "" {
			actionCard["singleTitle"] = "查看详情"
			actionCard["singleURL"] = message.LinkURL
		}

		msgBody["actionCard"] = actionCard
	case "feedCard":
		if message.LinkURL != "" && message.ImageURL != "" {
			links := []map[string]string{
				{
					"title":      message.Title,
					"messageURL": message.LinkURL,
					"picURL":     message.ImageURL,
				},
			}
			msgBody["feedCard"] = map[string]interface{}{
				"links": links,
			}
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
func (d *DingTalkSender) HealthCheck(ctx context.Context) error {
	// 测试发送空消息
	testMsg := &notification.Message{
		Title:   "健康检查",
		Content: "钉钉通知渠道健康检查测试",
		Level:   notification.LevelInfo,
	}

	return d.Validate(testMsg)
}

// Close 关闭发送器
func (d *DingTalkSender) Close() error {
	return nil
}

// CreateDingTalkChannel 创建钉钉通知渠道
func CreateDingTalkChannel(config *DingTalkConfig) *notification.Channel {
	return &notification.Channel{
		Name:        "dingtalk",
		Description: "钉钉通知渠道",
		Enabled:     config.WebhookURL != "",
		Sender:      NewDingTalkSender(config),
		Config: map[string]string{
			"webhook_url": config.WebhookURL,
			"msg_type":    config.MsgType,
		},
	}
}
