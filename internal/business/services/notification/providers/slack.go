package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/integration/notification"
)

// SlackProvider Slack通知提供商
type SlackProvider struct {
	config  *SlackConfig
	client  *http.Client
	logger  *zap.Logger
}

// SlackConfig Slack配置
type SlackConfig struct {
	BotToken      string `json:"bot_token"`
	WebhookURL    string `json:"webhook_url"`
	Channel       string `json:"channel"`
	DefaultUser   string `json:"default_user"`
	DefaultIcon   string `json:"default_icon"`
	ProxyURL      string `json:"proxy_url"`
	Timeout       int    `json:"timeout"`
	EnableProxy   bool   `json:"enable_proxy"`
}

// SlackMessage Slack消息格式
type SlackMessage struct {
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Blocks      []Block      `json:"blocks,omitempty"`
}

// Attachment Slack附件
type Attachment struct {
	Color     string   `json:"color,omitempty"`
	Title     string   `json:"title,omitempty"`
	TitleLink string   `json:"title_link,omitempty"`
	Text      string   `json:"text,omitempty"`
	ImageURL  string   `json:"image_url,omitempty"`
	Timestamp int64    `json:"ts,omitempty"`
	Fields    []Field  `json:"fields,omitempty"`
}

// Field Slack字段
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short,omitempty"`
}

// Block Slack块
type Block struct {
	Type      string      `json:"type"`
	Text      *TextBlock  `json:"text,omitempty"`
	Image     *ImageBlock `json:"image,omitempty"`
	Fields    []Field     `json:"fields,omitempty"`
}

// TextBlock 文本块
type TextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ImageBlock 图片块
type ImageBlock struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
	AltText  string `json:"alt_text"`
}

// NewSlackProvider 创建Slack提供商
func NewSlackProvider(logger *zap.Logger) notification.NotificationProvider {
	return &SlackProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetName 获取提供商名称
func (p *SlackProvider) GetName() string {
	return "Slack"
}

// GetType 获取提供商类型
func (p *SlackProvider) GetType() string {
	return "slack"
}

// Initialize 初始化提供商
func (p *SlackProvider) Initialize(config map[string]interface{}) error {
	// 解析配置
	configBytes, _ := json.Marshal(config)
	p.config = &SlackConfig{}
	if err := json.Unmarshal(configBytes, p.config); err != nil {
		return fmt.Errorf("解析Slack配置失败: %w", err)
	}
	
	// 验证配置
	if p.config.WebhookURL == "" && p.config.BotToken == "" {
		return fmt.Errorf("必须提供WebhookURL或BotToken")
	}
	
	// 设置超时
	if p.config.Timeout > 0 {
		p.client.Timeout = time.Duration(p.config.Timeout) * time.Second
	}
	
	p.logger.Info("Slack提供商初始化成功")
	return nil
}

// Send 发送通知消息
func (p *SlackProvider) Send(ctx context.Context, message *notification.NotificationMessage) error {
	var slackMessage SlackMessage
	
	// 根据配置选择发送方式
	if p.config.WebhookURL != "" {
		// 使用Webhook
		slackMessage = p.buildWebhookMessage(message)
		return p.sendViaWebhook(ctx, slackMessage)
	} else {
		// 使用Bot Token
		slackMessage = p.buildBotMessage(message)
		return p.sendViaBot(ctx, slackMessage)
	}
}

// SendBatch 批量发送通知消息
func (p *SlackProvider) SendBatch(ctx context.Context, messages []*notification.NotificationMessage) error {
	// Slack暂不支持真正的批量发送，这里进行逐个发送
	for _, message := range messages {
		if err := p.Send(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

// ValidateConfig 验证配置
func (p *SlackProvider) ValidateConfig(config map[string]interface{}) error {
	configBytes, _ := json.Marshal(config)
	var cfg SlackConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}
	
	if cfg.WebhookURL == "" && cfg.BotToken == "" {
		return fmt.Errorf("必须提供WebhookURL或BotToken")
	}
	
	return nil
}

// IsHealthy 健康检查
func (p *SlackProvider) IsHealthy(ctx context.Context) bool {
	if p.config.WebhookURL != "" {
		// 测试Webhook连接
		testMessage := SlackMessage{
			Text: "Health Check",
		}
		
		err := p.sendViaWebhook(ctx, testMessage)
		if err != nil {
			p.logger.Debug("Slack健康检查失败", zap.Error(err))
			return false
		}
		return true
	}
	
	// Bot Token健康检查需要调用API
	return true
}

// GetConfig 获取当前配置
func (p *SlackProvider) GetConfig() map[string]interface{} {
	if p.config == nil {
		return nil
	}
	
	configBytes, _ := json.Marshal(p.config)
	var result map[string]interface{}
	json.Unmarshal(configBytes, &result)
	
	// 隐藏敏感信息
	if webhookURL, exists := result["webhook_url"]; exists {
		if str, ok := webhookURL.(string); ok && len(str) > 20 {
			result["webhook_url"] = str[:20] + "***"
		}
	}
	
	if botToken, exists := result["bot_token"]; exists {
		if str, ok := botToken.(string); ok && len(str) > 20 {
			result["bot_token"] = str[:20] + "***"
		}
	}
	
	return result
}

// Close 关闭提供商
func (p *SlackProvider) Close() error {
	return nil
}

// buildWebhookMessage 构建Webhook消息
func (p *SlackProvider) buildWebhookMessage(message *notification.NotificationMessage) SlackMessage {
	slackMessage := SlackMessage{
		Text:     message.Content,
		Username: p.config.DefaultUser,
		IconEmoji: p.config.DefaultIcon,
	}
	
	// 设置频道
	if p.config.Channel != "" {
		slackMessage.Channel = p.config.Channel
	}
	
	// 添加附件
	attachment := Attachment{
		Title: message.Title,
		Text:  message.Content,
	}
	
	// 根据级别设置颜色
	switch message.Level {
	case notification.LevelError, notification.LevelCritical:
		attachment.Color = "danger"
	case notification.LevelWarning:
		attachment.Color = "warning"
	case notification.LevelSuccess:
		attachment.Color = "good"
	default:
		attachment.Color = "#439FE0"
	}
	
	// 添加图片
	if message.ImageURL != "" {
		attachment.ImageURL = message.ImageURL
	}
	
	// 添加字段
	if len(message.Metadata) > 0 {
		for key, value := range message.Metadata {
			attachment.Fields = append(attachment.Fields, Field{
				Title: key,
				Value: fmt.Sprintf("%v", value),
				Short: true,
			})
		}
	}
	
	slackMessage.Attachments = []Attachment{attachment}
	
	return slackMessage
}

// buildBotMessage 构建Bot消息
func (p *SlackProvider) buildBotMessage(message *notification.NotificationMessage) SlackMessage {
	slackMessage := SlackMessage{
		Channel: p.config.Channel,
		Text:    fmt.Sprintf("*%s*\n%s", message.Title, message.Content),
	}
	
	if p.config.DefaultUser != "" {
		slackMessage.Username = p.config.DefaultUser
	}
	
	if p.config.DefaultIcon != "" {
		if strings.HasPrefix(p.config.DefaultIcon, "http") {
			slackMessage.IconURL = p.config.DefaultIcon
		} else {
			slackMessage.IconEmoji = p.config.DefaultIcon
		}
	}
	
	return slackMessage
}

// sendViaWebhook 通过Webhook发送
func (p *SlackProvider) sendViaWebhook(ctx context.Context, message SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("编码Slack消息失败: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Slack消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Slack返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// sendViaBot 通过Bot Token发送
func (p *SlackProvider) sendViaBot(ctx context.Context, message SlackMessage) error {
	apiURL := "https://slack.com/api/chat.postMessage"
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("编码Slack消息失败: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.BotToken)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Slack消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取Slack响应失败: %w", err)
	}
	
	var response struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析Slack响应失败: %w", err)
	}
	
	if !response.OK {
		return fmt.Errorf("Slack API错误: %s", response.Error)
	}
	
	return nil
}