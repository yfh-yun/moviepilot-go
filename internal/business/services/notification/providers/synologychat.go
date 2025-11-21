package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/integration/notification"
)

// SynologyChatProvider SynologyChat通知提供商
type SynologyChatProvider struct {
	config  *SynologyChatConfig
	client  *http.Client
	logger  *zap.Logger
}

// SynologyChatConfig SynologyChat配置
type SynologyChatConfig struct {
	WebhookURL string `json:"webhook_url"`
	Token      string `json:"token"`
	Timeout    int    `json:"timeout"`
	ProxyURL   string `json:"proxy_url"`
}

// SynologyChatMessage SynologyChat消息格式
type SynologyChatMessage struct {
	Text string `json:"text"`
}

// SynologyChatResponse SynologyChat响应
type SynologyChatResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// NewSynologyChatProvider 创建SynologyChat提供商
func NewSynologyChatProvider(logger *zap.Logger) notification.NotificationProvider {
	return &SynologyChatProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetName 获取提供商名称
func (p *SynologyChatProvider) GetName() string {
	return "SynologyChat"
}

// GetType 获取提供商类型
func (p *SynologyChatProvider) GetType() string {
	return "synologychat"
}

// Initialize 初始化提供商
func (p *SynologyChatProvider) Initialize(config map[string]interface{}) error {
	// 解析配置
	configBytes, _ := json.Marshal(config)
	p.config = &SynologyChatConfig{}
	if err := json.Unmarshal(configBytes, p.config); err != nil {
		return fmt.Errorf("解析SynologyChat配置失败: %w", err)
	}
	
	// 验证配置
	if p.config.WebhookURL == "" {
		return fmt.Errorf("WebhookURL不能为空")
	}
	
	// 设置超时
	if p.config.Timeout > 0 {
		p.client.Timeout = time.Duration(p.config.Timeout) * time.Second
	}
	
	p.logger.Info("SynologyChat提供商初始化成功")
	return nil
}

// Send 发送通知消息
func (p *SynologyChatProvider) Send(ctx context.Context, message *notification.NotificationMessage) error {
	synologyMessage := p.buildMessage(message)
	
	jsonData, err := json.Marshal(synologyMessage)
	if err != nil {
		return fmt.Errorf("编码SynologyChat消息失败: %w", err)
	}
	
	// 添加Token到URL
	webhookURL := p.config.WebhookURL
	if p.config.Token != "" {
		if strings.Contains(webhookURL, "?") {
			webhookURL += "&token=" + p.config.Token
		} else {
			webhookURL += "?token=" + p.config.Token
		}
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送SynologyChat消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取SynologyChat响应失败: %w", err)
	}
	
	var response SynologyChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析SynologyChat响应失败: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("SynologyChat发送失败: %s", response.Error)
	}
	
	return nil
}

// SendBatch 批量发送通知消息
func (p *SynologyChatProvider) SendBatch(ctx context.Context, messages []*notification.NotificationMessage) error {
	// SynologyChat暂不支持真正的批量发送，这里进行逐个发送
	for _, message := range messages {
		if err := p.Send(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

// ValidateConfig 验证配置
func (p *SynologyChatProvider) ValidateConfig(config map[string]interface{}) error {
	configBytes, _ := json.Marshal(config)
	var cfg SynologyChatConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}
	
	if cfg.WebhookURL == "" {
		return fmt.Errorf("WebhookURL不能为空")
	}
	
	return nil
}

// IsHealthy 健康检查
func (p *SynologyChatProvider) IsHealthy(ctx context.Context) bool {
	testMessage := SynologyChatMessage{
		Text: "Health Check",
	}
	
	err := p.sendMessage(ctx, testMessage)
	if err != nil {
		p.logger.Debug("SynologyChat健康检查失败", zap.Error(err))
		return false
	}
	
	return true
}

// GetConfig 获取当前配置
func (p *SynologyChatProvider) GetConfig() map[string]interface{} {
	if p.config == nil {
		return nil
	}
	
	configBytes, _ := json.Marshal(p.config)
	var result map[string]interface{}
	json.Unmarshal(configBytes, &result)
	
	// 隐藏敏感信息
	if token, exists := result["token"]; exists {
		if str, ok := token.(string); ok && len(str) > 10 {
			result["token"] = str[:10] + "***"
		}
	}
	
	return result
}

// Close 关闭提供商
func (p *SynologyChatProvider) Close() error {
	return nil
}

// buildMessage 构建消息
func (p *SynologyChatProvider) buildMessage(message *notification.NotificationMessage) SynologyChatMessage {
	var contentBuilder strings.Builder
	
	// 添加标题
	if message.Title != "" {
		contentBuilder.WriteString(fmt.Sprintf("📢 %s\n\n", message.Title))
	}
	
	// 添加内容
	contentBuilder.WriteString(message.Content)
	
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
	
	finalContent := fmt.Sprintf("%s %s", levelIcon, contentBuilder.String())
	
	// 添加元数据
	if len(message.Metadata) > 0 {
		finalContent += "\n\n📋 详细信息"
		for key, value := range message.Metadata {
			finalContent += fmt.Sprintf("\n• %s: %v", key, value)
		}
	}
	
	// 添加图片
	if message.ImageURL != "" {
		finalContent += fmt.Sprintf("\n\n📷 %s", message.ImageURL)
	}
	
	return SynologyChatMessage{
		Text: finalContent,
	}
}

// sendMessage 发送消息（内部方法）
func (p *SynologyChatProvider) sendMessage(ctx context.Context, message SynologyChatMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("编码SynologyChat消息失败: %w", err)
	}
	
	webhookURL := p.config.WebhookURL
	if p.config.Token != "" {
		if strings.Contains(webhookURL, "?") {
			webhookURL += "&token=" + p.config.Token
		} else {
			webhookURL += "?token=" + p.config.Token
		}
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送SynologyChat消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取SynologyChat响应失败: %w", err)
	}
	
	var response SynologyChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析SynologyChat响应失败: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("SynologyChat发送失败: %s", response.Error)
	}
	
	return nil
}