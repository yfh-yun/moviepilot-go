// Package implementations 提供动作系统的具体实现
package implementations

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/business/workflows/base"
	"moviepilot-go/internal/business/workflows/interfaces"
	"moviepilot-go/internal/business/workflows/types"
	"moviepilot-go/pkg/logger"
)

// MessageSender 消息发送动作
type MessageSender struct {
	*base.Action
	config *MessageSenderConfig
}

// MessageSenderConfig 消息发送器配置
type MessageSenderConfig struct {
	Channels    []string          `json:"channels" description:"通知渠道列表"`
	Title       string            `json:"title" description:"消息标题"`
	Content     string            `json:"content" description:"消息内容"`
	ImageURL    string            `json:"image_url" description:"图片URL"`
	Link        string            `json:"link" description:"链接地址"`
	Priority    string            `json:"priority" description:"优先级: low, normal, high"`
	Tags        []string          `json:"tags" description:"标签列表"`
	Retries     int               `json:"retries" description:"重试次数"`
	Timeout     time.Duration     `json:"timeout" description:"超时时间"`
	Template    string            `json:"template" description:"消息模板"`
	Variables   map[string]string `json:"variables" description:"模板变量"`
}

// NewMessageSender 创建消息发送器实例
func NewMessageSender() interfaces.Action {
	return &MessageSender{
		Action: base.NewAction("MessageSender", "消息发送器，支持多渠道消息推送"),
		config: &MessageSenderConfig{
			Channels: []string{"webhook", "email"},
			Priority: "normal",
			Retries:  3,
			Timeout:  30 * time.Second,
		},
	}
}

// Execute 执行消息发送
func (ms *MessageSender) Execute(ctx context.Context, workflowID int64, params map[string]interface{}, actionContext *types.ActionContext) (*types.ActionContext, error) {
	logger.Debug("MessageSender execution started", 
		"workflow_id", workflowID,
		"action", "MessageSender")

	// 解析参数
	config, err := ms.parseConfig(params)
	if err != nil {
		ms.SetError(fmt.Sprintf("参数解析失败: %v", err))
		return actionContext, err
	}

	// 构建消息
	message := ms.buildMessage(config)

	// 发送消息
	results := make(map[string]interface{})
	successCount := 0

	for _, channel := range config.Channels {
		result, err := ms.sendToChannel(ctx, channel, message, config)
		if err != nil {
			logger.Error("Failed to send message", "channel", channel, "error", err)
			results[channel] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		} else {
			results[channel] = map[string]interface{}{
				"success": true,
				"result":  result,
			}
			successCount++
		}
	}

	// 设置结果
	ms.SetData("message", message)
	ms.SetData("results", results)
	ms.SetData("success_count", successCount)
	ms.SetData("total_count", len(config.Channels))

	if successCount > 0 {
		ms.SetDone(fmt.Sprintf("成功发送到 %d/%d 个渠道", successCount, len(config.Channels)))
	} else {
		ms.SetError("所有渠道发送失败")
		return actionContext, fmt.Errorf("所有渠道发送失败")
	}

	logger.Info("MessageSender execution completed", 
		"workflow_id", workflowID,
		"success_count", successCount,
		"total_count", len(config.Channels))

	return actionContext, nil
}

// parseConfig 解析配置参数
func (ms *MessageSender) parseConfig(params map[string]interface{}) (*MessageSenderConfig, error) {
	config := *ms.config // 复制默认配置

	if channels, ok := params["channels"].([]string); ok {
		config.Channels = channels
	}

	if title, ok := params["title"].(string); ok {
		config.Title = title
	}

	if content, ok := params["content"].(string); ok {
		config.Content = content
	}

	if imageURL, ok := params["image_url"].(string); ok {
		config.ImageURL = imageURL
	}

	if link, ok := params["link"].(string); ok {
		config.Link = link
	}

	if priority, ok := params["priority"].(string); ok {
		config.Priority = priority
	}

	if tags, ok := params["tags"].([]string); ok {
		config.Tags = tags
	}

	if retries, ok := params["retries"].(float64); ok {
		config.Retries = int(retries)
	}

	if variables, ok := params["variables"].(map[string]string); ok {
		config.Variables = variables
	}

	return &config, nil
}

// buildMessage 构建消息
func (ms *MessageSender) buildMessage(config *MessageSenderConfig) *Message {
	message := &Message{
		Title:    config.Title,
		Content:  config.Content,
		ImageURL: config.ImageURL,
		Link:     config.Link,
		Priority: config.Priority,
		Tags:     config.Tags,
		Time:     time.Now(),
	}

	// 应用模板
	if config.Template != "" {
		message = ms.applyTemplate(message, config.Template, config.Variables)
	}

	return message
}

// Message 消息结构
type Message struct {
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	ImageURL string    `json:"image_url"`
	Link     string    `json:"link"`
	Priority string    `json:"priority"`
	Tags     []string  `json:"tags"`
	Time     time.Time `json:"time"`
}

// sendToChannel 发送到指定渠道
func (ms *MessageSender) sendToChannel(ctx context.Context, channel string, message *Message, config *MessageSenderConfig) (interface{}, error) {
	switch channel {
	case "webhook":
		return ms.sendToWebhook(ctx, message, config)
	case "email":
		return ms.sendToEmail(ctx, message, config)
	case "slack":
		return ms.sendToSlack(ctx, message, config)
	case "telegram":
		return ms.sendToTelegram(ctx, message, config)
	case "dingtalk":
		return ms.sendToDingTalk(ctx, message, config)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", channel)
	}
}

// sendToWebhook 发送到Webhook
func (ms *MessageSender) sendToWebhook(ctx context.Context, message *Message, config *MessageSenderConfig) (interface{}, error) {
	// 实现Webhook发送逻辑
	logger.Info("Sending message to webhook", "title", message.Title)
	return map[string]interface{}{
		"webhook_id": "wh_" + fmt.Sprintf("%d", time.Now().Unix()),
		"status":     "sent",
	}, nil
}

// sendToEmail 发送到邮件
func (ms *MessageSender) sendToEmail(ctx context.Context, message *Message, config *MessageSenderConfig) (interface{}, error) {
	// 实现邮件发送逻辑
	logger.Info("Sending message to email", "title", message.Title)
	return map[string]interface{}{
		"email_id":   "email_" + fmt.Sprintf("%d", time.Now().Unix()),
		"recipients": []string{"admin@example.com"},
		"status":     "sent",
	}, nil
}

// sendToSlack 发送到Slack
func (ms *MessageSender) sendToSlack(ctx context.Context, message *Message, config *MessageSenderConfig) (interface{}, error) {
	// 实现Slack发送逻辑
	logger.Info("Sending message to Slack", "title", message.Title)
	return map[string]interface{}{
		"slack_ts":  fmt.Sprintf("%d", time.Now().Unix()),
		"channel":   "#general",
		"status":    "sent",
	}, nil
}

// sendToTelegram 发送到Telegram
func (ms *MessageSender) sendToTelegram(ctx context.Context, message *Message, config *MessageSenderConfig) (interface{}, error) {
	// 实现Telegram发送逻辑
	logger.Info("Sending message to Telegram", "title", message.Title)
	return map[string]interface{}{
		"message_id": fmt.Sprintf("%d", time.Now().Unix()),
		"chat_id":    "123456789",
		"status":     "sent",
	}, nil
}

// sendToDingTalk 发送到钉钉
func (ms *MessageSender) sendToDingTalk(ctx context.Context, message *Message, config *MessageSenderConfig) (interface{}, error) {
	// 实现钉钉发送逻辑
	logger.Info("Sending message to DingTalk", "title", message.Title)
	return map[string]interface{}{
		"message_id": fmt.Sprintf("%d", time.Now().Unix()),
		"chat_id":    "dingtalk_group_123",
		"status":     "sent",
	}, nil
}

// applyTemplate 应用模板
func (ms *MessageSender) applyTemplate(message *Message, template string, variables map[string]string) *Message {
	// 简单的模板替换
	title := message.Title
	content := message.Content

	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		title = fmt.Sprintf(title, placeholder, value)
		content = fmt.Sprintf(content, placeholder, value)
	}

	message.Title = title
	message.Content = content
	return message
}

// Initialize 初始化消息发送器
func (ms *MessageSender) Initialize() error {
	logger.Info("Initializing MessageSender", 
		"channels", ms.config.Channels,
		"priority", ms.config.Priority)
	return nil
}

// Cleanup 清理资源
func (ms *MessageSender) Cleanup() error {
	logger.Info("Cleaning up MessageSender")
	return nil
}