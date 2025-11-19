package channels

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/notification"
)

// FeishuConfig 飞书通知配置
type FeishuConfig struct {
	WebhookURL string `json:"webhook_url"` // 飞书群机器人Webhook URL
	Secret     string `json:"secret"`      // 安全密钥（可选）
	MsgType    string `json:"msg_type"`    // 消息类型，支持text、post、interactive等
}

// FeishuSender 飞书通知发送器
type FeishuSender struct {
	config     *FeishuConfig
	httpClient *http.Client
}

// NewFeishuSender 创建新的飞书通知发送器
func NewFeishuSender(config *FeishuConfig) *FeishuSender {
	return &FeishuSender{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name 返回发送器名称
func (f *FeishuSender) Name() string {
	return "feishu"
}

// SupportedLevels 返回支持的通知级别
func (f *FeishuSender) SupportedLevels() []notification.NotificationLevel {
	return []notification.NotificationLevel{
		notification.LevelInfo,
		notification.LevelWarning,
		notification.LevelError,
		notification.LevelSuccess,
	}
}

// Validate 验证消息
func (f *FeishuSender) Validate(message *notification.Message) error {
	if message.Title == "" && message.Content == "" {
		return fmt.Errorf("message title and content cannot both be empty")
	}

	if len(message.Content) > 4096 {
		return fmt.Errorf("message content too long (max 4096 characters)")
	}

	return nil
}

// generateSign 生成签名（如果配置了安全密钥）
func (f *FeishuSender) generateSign(timestamp int64) string {
	if f.config.Secret == "" {
		return ""
	}

	signString := fmt.Sprintf("%d\n%s", timestamp, f.config.Secret)
	hmac := hmac.New(sha256.New, []byte(f.config.Secret))
	hmac.Write([]byte(signString))
	return base64.StdEncoding.EncodeToString(hmac.Sum(nil))
}

// Send 发送飞书通知
func (f *FeishuSender) Send(ctx context.Context, message *notification.Message) error {
	if f.config.WebhookURL == "" {
		return fmt.Errorf("feishu webhook URL is not configured")
	}

	// 构建消息体
	msgBody := f.buildMessageBody(message)

	req, err := http.NewRequestWithContext(ctx, "POST", f.config.WebhookURL, strings.NewReader(msgBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Code != 0 {
		return fmt.Errorf("feishu API error: %s (code: %d)", apiResp.Msg, apiResp.Code)
	}

	return nil
}

// buildMessageBody 构建消息体
func (f *FeishuSender) buildMessageBody(message *notification.Message) string {
	msgType := f.config.MsgType
	if msgType == "" {
		msgType = "text"
	}

	// 生成时间戳和签名
	timestamp := time.Now().Unix()
	sign := f.generateSign(timestamp)

	msgBody := map[string]interface{}{
		"msg_type": msgType,
	}

	// 如果有签名，添加到请求中
	if sign != "" {
		msgBody["timestamp"] = timestamp
		msgBody["sign"] = sign
	}

	switch msgType {
	case "text":
		content := message.Content
		if message.Title != "" {
			content = fmt.Sprintf("%s\n%s", message.Title, message.Content)
		}

		msgBody["content"] = map[string]interface{}{
			"text": content,
		}

	case "post":
		content := fmt.Sprintf("%s\n%s", message.Title, message.Content)

		msgBody["content"] = map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title": message.Title,
					"content": [][]map[string]string{
						{
							{
								"tag":  "text",
								"text": message.Content,
							},
						},
					},
				},
			},
		}

	case "interactive":
		// 构建交互式消息卡片
		card := map[string]interface{}{
			"config": map[string]string{
				"wide_screen_mode": "true",
			},
			"header": map[string]interface{}{
				"title": map[string]string{
					"tag":     "plain_text",
					"content": message.Title,
				},
				"template": f.getHeaderTemplate(message.Level),
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]string{
						"tag":     "lark_md",
						"content": message.Content,
					},
				},
			},
		}

		// 添加操作按钮（如果有链接）
		if message.LinkURL != "" {
			actions := map[string]interface{}{
				"tag": "action",
				"actions": []map[string]string{
					{
						"tag": "button",
						"text": map[string]string{
							"tag":     "plain_text",
							"content": "查看详情",
						},
						"type": "primary",
						"url":  message.LinkURL,
					},
				},
			}
			card["elements"] = append(card["elements"].([]map[string]interface{}), actions)
		}

		msgBody["card"] = card

	default:
		// 默认使用文本消息
		content := message.Content
		if message.Title != "" {
			content = fmt.Sprintf("%s\n%s", message.Title, message.Content)
		}

		msgBody["content"] = map[string]interface{}{
			"text": content,
		}
	}

	jsonData, _ := json.Marshal(msgBody)
	return string(jsonData)
}

// getHeaderTemplate 根据消息级别获取标题模板颜色
func (f *FeishuSender) getHeaderTemplate(level notification.NotificationLevel) string {
	switch level {
	case notification.LevelError:
		return "red"
	case notification.LevelWarning:
		return "orange"
	case notification.LevelSuccess:
		return "green"
	case notification.LevelInfo:
		return "blue"
	default:
		return "blue"
	}
}

// HealthCheck 健康检查
func (f *FeishuSender) HealthCheck(ctx context.Context) error {
	// 测试发送空消息
	testMsg := &notification.Message{
		Title:   "健康检查",
		Content: "飞书通知渠道健康检查测试",
		Level:   notification.LevelInfo,
	}

	return f.Validate(testMsg)
}

// Close 关闭发送器
func (f *FeishuSender) Close() error {
	return nil
}

// CreateFeishuChannel 创建飞书通知渠道
func CreateFeishuChannel(config *FeishuConfig) *notification.Channel {
	return &notification.Channel{
		Name:        "feishu",
		Description: "飞书通知渠道",
		Enabled:     config.WebhookURL != "",
		Sender:      NewFeishuSender(config),
		Config: map[string]string{
			"webhook_url": config.WebhookURL,
			"msg_type":    config.MsgType,
		},
	}
}
