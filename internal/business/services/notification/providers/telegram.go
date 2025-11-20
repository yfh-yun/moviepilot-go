package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/integration/notification"
)

// TelegramProvider Telegram通知提供商
type TelegramProvider struct {
	config  *TelegramConfig
	client  *http.Client
	logger  *zap.Logger
}

// TelegramConfig Telegram配置
type TelegramConfig struct {
	BotToken      string `json:"bot_token"`
	ChatID        string `json:"chat_id"`
	WebhookURL    string `json:"webhook_url"`
	ProxyURL      string `json:"proxy_url"`
	Timeout       int    `json:"timeout"`
	ParseMode     string `json:"parse_mode"` // HTML, Markdown, MarkdownV2
	DisablePreview bool   `json:"disable_preview"`
}

// TelegramMessage Telegram消息格式
type TelegramMessage struct {
	ChatID                   string                    `json:"chat_id"`
	Text                     string                    `json:"text"`
	ParseMode                string                    `json:"parse_mode,omitempty"`
	DisableWebPagePreview    bool                      `json:"disable_web_page_preview,omitempty"`
	DisableNotification      bool                      `json:"disable_notification,omitempty"`
	ReplyToMessageID         int                       `json:"reply_to_message_id,omitempty"`
	ReplyMarkup              *ReplyMarkup               `json:"reply_markup,omitempty"`
}

// ReplyMarkup 回复标记
type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
	Keyboard       [][]KeyboardButton        `json:"keyboard,omitempty"`
}

// InlineKeyboardButton 内联键盘按钮
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

// KeyboardButton 键盘按钮
type KeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact,omitempty"`
	RequestLocation bool   `json:"request_location,omitempty"`
}

// TelegramResponse Telegram API响应
type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
	ErrorCode   int    `json:"error_code,omitempty"`
}

// NewTelegramProvider 创建Telegram提供商
func NewTelegramProvider(logger *zap.Logger) notification.NotificationProvider {
	return &TelegramProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetName 获取提供商名称
func (p *TelegramProvider) GetName() string {
	return "Telegram"
}

// GetType 获取提供商类型
func (p *TelegramProvider) GetType() string {
	return "telegram"
}

// Initialize 初始化提供商
func (p *TelegramProvider) Initialize(config map[string]interface{}) error {
	// 解析配置
	configBytes, _ := json.Marshal(config)
	p.config = &TelegramConfig{}
	if err := json.Unmarshal(configBytes, p.config); err != nil {
		return fmt.Errorf("解析Telegram配置失败: %w", err)
	}
	
	// 验证配置
	if p.config.BotToken == "" {
		return fmt.Errorf("BotToken不能为空")
	}
	
	if p.config.ChatID == "" {
		return fmt.Errorf("ChatID不能为空")
	}
	
	// 设置默认解析模式
	if p.config.ParseMode == "" {
		p.config.ParseMode = "HTML"
	}
	
	// 设置超时
	if p.config.Timeout > 0 {
		p.client.Timeout = time.Duration(p.config.Timeout) * time.Second
	}
	
	p.logger.Info("Telegram提供商初始化成功")
	return nil
}

// Send 发送通知消息
func (p *TelegramProvider) Send(ctx context.Context, message *notification.NotificationMessage) error {
	telegramMessage := p.buildMessage(message)
	
	// 使用Webhook（如果配置了）
	if p.config.WebhookURL != "" {
		return p.sendViaWebhook(ctx, telegramMessage)
	}
	
	// 使用Bot Token API
	return p.sendViaAPI(ctx, telegramMessage)
}

// SendBatch 批量发送通知消息
func (p *TelegramProvider) SendBatch(ctx context.Context, messages []*notification.NotificationMessage) error {
	// Telegram暂不支持真正的批量发送，这里进行逐个发送
	for _, message := range messages {
		if err := p.Send(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

// ValidateConfig 验证配置
func (p *TelegramProvider) ValidateConfig(config map[string]interface{}) error {
	configBytes, _ := json.Marshal(config)
	var cfg TelegramConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}
	
	if cfg.BotToken == "" {
		return fmt.Errorf("BotToken不能为空")
	}
	
	if cfg.ChatID == "" {
		return fmt.Errorf("ChatID不能为空")
	}
	
	return nil
}

// IsHealthy 健康检查
func (p *TelegramProvider) IsHealthy(ctx context.Context) bool {
	// 发送测试消息
	testMessage := TelegramMessage{
		ChatID: p.config.ChatID,
		Text:   "Health Check",
	}
	
	err := p.sendViaAPI(ctx, testMessage)
	if err != nil {
		p.logger.Debug("Telegram健康检查失败", zap.Error(err))
		return false
	}
	
	return true
}

// GetConfig 获取当前配置
func (p *TelegramProvider) GetConfig() map[string]interface{} {
	if p.config == nil {
		return nil
	}
	
	configBytes, _ := json.Marshal(p.config)
	var result map[string]interface{}
	json.Unmarshal(configBytes, &result)
	
	// 隐藏敏感信息
	if botToken, exists := result["bot_token"]; exists {
		if str, ok := botToken.(string); ok && len(str) > 20 {
			result["bot_token"] = str[:20] + "***"
		}
	}
	
	return result
}

// Close 关闭提供商
func (p *TelegramProvider) Close() error {
	return nil
}

// buildMessage 构建Telegram消息
func (p *TelegramProvider) buildMessage(message *notification.NotificationMessage) TelegramMessage {
	var textBuilder strings.Builder
	
	// 添加标题
	if message.Title != "" {
		textBuilder.WriteString(fmt.Sprintf("<b>%s</b>\n\n", message.Title))
	}
	
	// 添加内容
	textBuilder.WriteString(message.Content)
	
	// 添加图片
	if message.ImageURL != "" {
		textBuilder.WriteString(fmt.Sprintf("\n\n📷 %s", message.ImageURL))
	}
	
	// 添加元数据
	if len(message.Metadata) > 0 {
		textBuilder.WriteString("\n\n📋 <b>详细信息</b>")
		for key, value := range message.Metadata {
			textBuilder.WriteString(fmt.Sprintf("\n• <b>%s:</b> %v", key, value))
		}
	}
	
	// 根据级别添加图标
	var levelIcon string
	switch message.Level {
	case notification.LevelError, notification.LevelCritical:
		levelIcon = "🚨"
	case notification.LevelWarning:
		levelIcon = "⚠️"
	case notification.LevelSuccess:
		levelIcon = "✅"
	default:
		levelIcon = "ℹ️"
	}
	
	text := fmt.Sprintf("%s %s", levelIcon, textBuilder.String())
	
	telegramMessage := TelegramMessage{
		ChatID:                p.config.ChatID,
		Text:                  text,
		ParseMode:             p.config.ParseMode,
		DisableWebPagePreview: p.config.DisablePreview,
	}
	
	// 如果是严重错误，启用静默通知
	if message.Level == notification.LevelCritical {
		telegramMessage.DisableNotification = false
	}
	
	return telegramMessage
}

// sendViaAPI 通过API发送
func (p *TelegramProvider) sendViaAPI(ctx context.Context, message TelegramMessage) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", p.config.BotToken)
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("编码Telegram消息失败: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Telegram消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取Telegram响应失败: %w", err)
	}
	
	var response TelegramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析Telegram响应失败: %w", err)
	}
	
	if !response.OK {
		return fmt.Errorf("Telegram API错误: %s (错误代码: %d)", response.Description, response.ErrorCode)
	}
	
	return nil
}

// sendViaWebhook 通过Webhook发送
func (p *TelegramProvider) sendViaWebhook(ctx context.Context, message TelegramMessage) error {
	// Webhook格式与API稍有不同
	webhookMessage := map[string]interface{}{
		"method": "sendMessage",
		"chat_id": message.ChatID,
		"text": message.Text,
	}
	
	if message.ParseMode != "" {
		webhookMessage["parse_mode"] = message.ParseMode
	}
	
	if message.DisableWebPagePreview {
		webhookMessage["disable_web_page_preview"] = true
	}
	
	if message.DisableNotification {
		webhookMessage["disable_notification"] = true
	}
	
	jsonData, err := json.Marshal(webhookMessage)
	if err != nil {
		return fmt.Errorf("编码Telegram Webhook消息失败: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Telegram Webhook消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取Telegram Webhook响应失败: %w", err)
	}
	
	var response TelegramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析Telegram Webhook响应失败: %w", err)
	}
	
	if !response.OK {
		return fmt.Errorf("Telegram Webhook错误: %s (错误代码: %d)", response.Description, response.ErrorCode)
	}
	
	return nil
}

// SendPhoto 发送图片消息
func (p *TelegramProvider) SendPhoto(ctx context.Context, chatID, photoURL, caption string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", p.config.BotToken)
	
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("photo", photoURL)
	if caption != "" {
		data.Set("caption", caption)
	}
	if p.config.ParseMode != "" {
		data.Set("parse_mode", p.config.ParseMode)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Telegram图片失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取Telegram图片响应失败: %w", err)
	}
	
	var response TelegramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析Telegram图片响应失败: %w", err)
	}
	
	if !response.OK {
		return fmt.Errorf("Telegram图片发送错误: %s (错误代码: %d)", response.Description, response.ErrorCode)
	}
	
	return nil
}

// SendDocument 发送文档消息
func (p *TelegramProvider) SendDocument(ctx context.Context, chatID, documentURL, caption string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", p.config.BotToken)
	
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("document", documentURL)
	if caption != "" {
		data.Set("caption", caption)
	}
	if p.config.ParseMode != "" {
		data.Set("parse_mode", p.config.ParseMode)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Telegram文档失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取Telegram文档响应失败: %w", err)
	}
	
	var response TelegramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析Telegram文档响应失败: %w", err)
	}
	
	if !response.OK {
		return fmt.Errorf("Telegram文档发送错误: %s (错误代码: %d)", response.Description, response.ErrorCode)
	}
	
	return nil
}

// ValidateChatID 验证ChatID格式
func (p *TelegramProvider) ValidateChatID(chatID string) error {
	// ChatID可以是纯数字、@username或者chat id
	if strings.HasPrefix(chatID, "@") {
		// @username格式
		if len(chatID) < 2 {
			return fmt.Errorf("用户名格式无效")
		}
	} else {
		// 纯数字格式
		if _, err := strconv.ParseInt(chatID, 10, 64); err != nil {
			return fmt.Errorf("ChatID必须是数字或@username格式")
		}
	}
	return nil
}