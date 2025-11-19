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

	"github.com/yfh-yun/moviepilot-go/internal/modules/notification"
)

// VoceChatProvider VoceChat通知提供商
type VoceChatProvider struct {
	config  *VoceChatConfig
	client  *http.Client
	logger  *zap.Logger
}

// VoceChatConfig VoceChat配置
type VoceChatConfig struct {
	ServerURL    string `json:"server_url"`
	AccessToken string `json:"access_token"`
	ChannelID    string `json:"channel_id"`
	Timeout      int    `json:"timeout"`
	ProxyURL     string `json:"proxy_url"`
}

// VoceChatMessage VoceChat消息格式
type VoceChatMessage struct {
	ChannelID string      `json:"channel_id"`
	Content   string      `json:"content"`
	Type      string      `json:"type"` // text, markdown, image
	AttachURL string      `json:"attach_url,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// VoceChatResponse VoceChat响应
type VoceChatResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewVoceChatProvider 创建VoceChat提供商
func NewVoceChatProvider(logger *zap.Logger) notification.NotificationProvider {
	return &VoceChatProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetName 获取提供商名称
func (p *VoceChatProvider) GetName() string {
	return "VoceChat"
}

// GetType 获取提供商类型
func (p *VoceChatProvider) GetType() string {
	return "vocechat"
}

// Initialize 初始化提供商
func (p *VoceChatProvider) Initialize(config map[string]interface{}) error {
	// 解析配置
	configBytes, _ := json.Marshal(config)
	p.config = &VoceChatConfig{}
	if err := json.Unmarshal(configBytes, p.config); err != nil {
		return fmt.Errorf("解析VoceChat配置失败: %w", err)
	}
	
	// 验证配置
	if p.config.ServerURL == "" {
		return fmt.Errorf("ServerURL不能为空")
	}
	
	if p.config.AccessToken == "" {
		return fmt.Errorf("AccessToken不能为空")
	}
	
	if p.config.ChannelID == "" {
		return fmt.Errorf("ChannelID不能为空")
	}
	
	// 确保ServerURL以/结尾
	if !strings.HasSuffix(p.config.ServerURL, "/") {
		p.config.ServerURL += "/"
	}
	
	// 设置超时
	if p.config.Timeout > 0 {
		p.client.Timeout = time.Duration(p.config.Timeout) * time.Second
	}
	
	p.logger.Info("VoceChat提供商初始化成功")
	return nil
}

// Send 发送通知消息
func (p *VoceChatProvider) Send(ctx context.Context, message *notification.NotificationMessage) error {
	voceMessage := p.buildMessage(message)
	
	jsonData, err := json.Marshal(voceMessage)
	if err != nil {
		return fmt.Errorf("编码VoceChat消息失败: %w", err)
	}
	
	apiURL := p.config.ServerURL + "api/v1/messages"
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.AccessToken)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送VoceChat消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取VoceChat响应失败: %w", err)
	}
	
	var response VoceChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析VoceChat响应失败: %w", err)
	}
	
	if !response.Success {
		errorMsg := response.Error
		if errorMsg == "" {
			errorMsg = response.Message
		}
		return fmt.Errorf("VoceChat发送失败: %s", errorMsg)
	}
	
	return nil
}

// SendBatch 批量发送通知消息
func (p *VoceChatProvider) SendBatch(ctx context.Context, messages []*notification.NotificationMessage) error {
	// VoceChat暂不支持真正的批量发送，这里进行逐个发送
	for _, message := range messages {
		if err := p.Send(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

// ValidateConfig 验证配置
func (p *VoceChatProvider) ValidateConfig(config map[string]interface{}) error {
	configBytes, _ := json.Marshal(config)
	var cfg VoceChatConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}
	
	if cfg.ServerURL == "" {
		return fmt.Errorf("ServerURL不能为空")
	}
	
	if cfg.AccessToken == "" {
		return fmt.Errorf("AccessToken不能为空")
	}
	
	if cfg.ChannelID == "" {
		return fmt.Errorf("ChannelID不能为空")
	}
	
	return nil
}

// IsHealthy 健康检查
func (p *VoceChatProvider) IsHealthy(ctx context.Context) bool {
	// 检查连接到VoceChat服务器
	testURL := p.config.ServerURL + "api/v1/health"
	
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		p.logger.Debug("创建VoceChat健康检查请求失败", zap.Error(err))
		return false
	}
	
	req.Header.Set("Authorization", "Bearer "+p.config.AccessToken)
	
	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Debug("VoceChat健康检查失败", zap.Error(err))
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// GetConfig 获取当前配置
func (p *VoceChatProvider) GetConfig() map[string]interface{} {
	if p.config == nil {
		return nil
	}
	
	configBytes, _ := json.Marshal(p.config)
	var result map[string]interface{}
	json.Unmarshal(configBytes, &result)
	
	// 隐藏敏感信息
	if accessToken, exists := result["access_token"]; exists {
		if str, ok := accessToken.(string); ok && len(str) > 20 {
			result["access_token"] = str[:20] + "***"
		}
	}
	
	return result
}

// Close 关闭提供商
func (p *VoceChatProvider) Close() error {
	return nil
}

// buildMessage 构建消息
func (p *VoceChatProvider) buildMessage(message *notification.NotificationMessage) VoceChatMessage {
	var contentBuilder strings.Builder
	
	// 添加标题
	if message.Title != "" {
		contentBuilder.WriteString(fmt.Sprintf("# %s\n\n", message.Title))
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
		finalContent += "\n\n## 详细信息"
		for key, value := range message.Metadata {
			finalContent += fmt.Sprintf("\n• **%s:** %v", key, value)
		}
	}
	
	// 添加图片
	if message.ImageURL != "" {
		finalContent += fmt.Sprintf("\n\n📷 ![图片](%s)", message.ImageURL)
	}
	
	voceMessage := VoceChatMessage{
		ChannelID: p.config.ChannelID,
		Content:   finalContent,
		Type:      "markdown", // 使用Markdown格式
	}
	
	// 如果有图片URL，也可以设置为附件
	if message.ImageURL != "" {
		voceMessage.AttachURL = message.ImageURL
	}
	
	return voceMessage
}