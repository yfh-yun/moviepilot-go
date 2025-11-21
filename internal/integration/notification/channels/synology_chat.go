package channels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/utils"
)

// SynologyChatConfig Synology Chat配置
type SynologyChatConfig struct {
	WebhookURL string `json:"webhook_url"`
	Token      string `json:"token"`
}

// SynologyChatClient Synology Chat客户端
type SynologyChatClient struct {
	config *SynologyChatConfig
	client *http.Client
}

// NewSynologyChatClient 创建Synology Chat客户端
func NewSynologyChatClient(config map[string]interface{}) *SynologyChatClient {
	webhookURL, _ := config["webhook_url"].(string)
	token, _ := config["token"].(string)

	return &SynologyChatClient{
		config: &SynologyChatConfig{
			WebhookURL: webhookURL,
			Token:      token,
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendMessage 发送消息
func (s *SynologyChatClient) SendMessage(ctx context.Context, title, content string, params map[string]interface{}) error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("Synology Chat webhook URL is required")
	}

	// 构建Synology Chat消息
	message := s.buildMessage(title, content, params)

	// 发送消息
	return s.sendToSynologyChat(ctx, message)
}

// buildMessage 构建Synology Chat消息
func (s *SynologyChatClient) buildMessage(title, content string, params map[string]interface{}) map[string]interface{} {
	message := map[string]interface{}{
		"text": fmt.Sprintf("**%s**\n%s", title, content),
	}

	// 添加token
	if s.config.Token != "" {
		message["token"] = s.config.Token
	}

	// 构建文件URL
	if fileURL, ok := params["file_url"].(string); ok && fileURL != "" {
		message["file_url"] = fileURL
	}

	// 添加附件
	if attachments := s.buildAttachments(params); len(attachments) > 0 {
		message["attachments"] = attachments
	}

	return message
}

// buildAttachments 构建附件
func (s *SynologyChatClient) buildAttachments(params map[string]interface{}) []map[string]interface{} {
	var attachments []map[string]interface{}

	// 主附件
	attachment := map[string]interface{}{
		"color": s.getColor(params),
		"fields": []map[string]interface{}{
			{
				"title": "时间",
				"value": time.Now().Format("2006-01-02 15:04:05"),
				"short": true,
			},
		},
	}

	// 添加类型字段
	if msgType, ok := params["type"].(string); ok && msgType != "" {
		attachment["fields"] = append(attachment["fields"].([]map[string]interface{}), map[string]interface{}{
			"title": "类型",
			"value": msgType,
			"short": true,
		})
	}

	// 添加大小字段
	if size, ok := params["size"].(int64); ok && size > 0 {
		attachment["fields"] = append(attachment["fields"].([]map[string]interface{}), map[string]interface{}{
			"title": "大小",
			"value": utils.FormatFileSize(size),
			"short": true,
		})
	}

	// 添加自定义字段
	if customFields, ok := params["fields"].(map[string]interface{}); ok {
		for key, value := range customFields {
			attachment["fields"] = append(attachment["fields"].([]map[string]interface{}), map[string]interface{}{
				"title": key,
				"value": fmt.Sprintf("%v", value),
				"short": true,
			})
		}
	}

	// 添加图片
	if imageURL, ok := params["image_url"].(string); ok && imageURL != "" {
		attachment["image_url"] = imageURL
	}

	attachments = append(attachments, attachment)
	return attachments
}

// getColor 获取消息颜色
func (s *SynologyChatClient) getColor(params map[string]interface{}) string {
	if msgType, ok := params["type"].(string); ok {
		switch msgType {
		case "download", "transfer":
			return "#36a64f" // 绿色
		case "warning":
			return "#ff9500" // 橙色
		case "error":
			return "#ff3b30" // 红色
		default:
			return "#007aff" // 蓝色
		}
	}
	return "#007aff"
}

// sendToSynologyChat 发送消息到Synology Chat
func (s *SynologyChatClient) sendToSynologyChat(ctx context.Context, message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Synology Chat message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Synology Chat message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Synology Chat API returned status %d", resp.StatusCode)
	}

	logger.Info("Synology Chat message sent successfully")
	return nil
}

// Test 测试连接
func (s *SynologyChatClient) Test(ctx context.Context) error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("Synology Chat webhook URL is required")
	}

	testMessage := map[string]interface{}{
		"text": "MoviePilot Synology Chat integration test",
	}

	if s.config.Token != "" {
		testMessage["token"] = s.config.Token
	}

	jsonData, err := json.Marshal(testMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal test message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send test message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Synology Chat test failed with status %d", resp.StatusCode)
	}

	logger.Info("Synology Chat integration test successful")
	return nil
}

// GetConfig 获取配置模板
func GetSynologyChatConfigTemplate() map[string]interface{} {
	return map[string]interface{}{
		"webhook_url": map[string]interface{}{
			"type":        "string",
			"required":    true,
			"label":       "Webhook URL",
			"placeholder": "https://your-synology:5001/webapi/entry.cgi?...",
			"description": "Synology Chat Webhook URL",
		},
		"token": map[string]interface{}{
			"type":        "string",
			"required":    false,
			"label":       "Token",
			"placeholder": "Optional authentication token",
			"description": "Authentication token (optional)",
		},
	}
}

// ValidateConfig 验证配置
func ValidateSynologyChatConfig(config map[string]interface{}) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	// 验证URL格式
	if !utils.IsValidURL(webhookURL) {
		return fmt.Errorf("invalid webhook URL format")
	}

	return nil
}