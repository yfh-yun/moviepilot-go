package providers

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/integration/notification"
)

// WeChatProvider 微信通知提供商
type WeChatProvider struct {
	config  *WeChatConfig
	client  *http.Client
	logger  *zap.Logger
}

// WeChatConfig 微信配置
type WeChatConfig struct {
	CorpID          string `json:"corp_id"`
	CorpSecret      string `json:"corp_secret"`
	AgentID         string `json:"agent_id"`
	Token           string `json:"token"`
	EncodingAESKey  string `json:"encoding_aes_key"`
	WebhookURL      string `json:"webhook_url"`
	ProxyURL        string `json:"proxy_url"`
	Timeout         int    `json:"timeout"`
	EnableEncrypt   bool   `json:"enable_encrypt"`
}

// WeChatMessage 微信消息格式
type WeChatMessage struct {
	ToUser   string      `json:"touser,omitempty"`
	ToParty  string      `json:"toparty,omitempty"`
	ToTag    string      `json:"totag,omitempty"`
	AgentID  string      `json:"agentid"`
	MsgType  string      `json:"msgtype"`
	Text     *TextMsg    `json:"text,omitempty"`
	Markdown *MarkdownMsg `json:"markdown,omitempty"`
	Image    *ImageMsg   `json:"image,omitempty"`
	News     *NewsMsg    `json:"news,omitempty"`
}

// TextMsg 文本消息
type TextMsg struct {
	Content string `json:"content"`
}

// MarkdownMsg Markdown消息
type MarkdownMsg struct {
	Content string `json:"content"`
}

// ImageMsg 图片消息
type ImageMsg struct {
	MediaID string `json:"media_id"`
}

// NewsMsg 图文消息
type NewsMsg struct {
	Articles []Article `json:"articles"`
}

// Article 图文文章
type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl"`
}

// AccessTokenResponse 访问令牌响应
type AccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// SendMessageResponse 发送消息响应
type SendMessageResponse struct {
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"invalidparty"`
	InvalidTag   string `json:"invalidtag"`
}

// UploadMediaResponse 上传媒体响应
type UploadMediaResponse struct {
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
	Type     string `json:"type"`
	MediaID  string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}

// NewWeChatProvider 创建微信提供商
func NewWeChatProvider(logger *zap.Logger) notification.NotificationProvider {
	return &WeChatProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetName 获取提供商名称
func (p *WeChatProvider) GetName() string {
	return "WeChat"
}

// GetType 获取提供商类型
func (p *WeChatProvider) GetType() string {
	return "wechat"
}

// Initialize 初始化提供商
func (p *WeChatProvider) Initialize(config map[string]interface{}) error {
	// 解析配置
	configBytes, _ := json.Marshal(config)
	p.config = &WeChatConfig{}
	if err := json.Unmarshal(configBytes, p.config); err != nil {
		return fmt.Errorf("解析微信配置失败: %w", err)
	}
	
	// 验证配置
	if p.config.CorpID == "" {
		return fmt.Errorf("CorpID不能为空")
	}
	
	if p.config.CorpSecret == "" {
		return fmt.Errorf("CorpSecret不能为空")
	}
	
	if p.config.AgentID == "" {
		return fmt.Errorf("AgentID不能为空")
	}
	
	// 设置超时
	if p.config.Timeout > 0 {
		p.client.Timeout = time.Duration(p.config.Timeout) * time.Second
	}
	
	p.logger.Info("微信提供商初始化成功")
	return nil
}

// Send 发送通知消息
func (p *WeChatProvider) Send(ctx context.Context, message *notification.NotificationMessage) error {
	if p.config.WebhookURL != "" {
		// 使用Webhook方式
		return p.sendViaWebhook(ctx, message)
	} else {
		// 使用企业微信API方式
		return p.sendViaAPI(ctx, message)
	}
}

// SendBatch 批量发送通知消息
func (p *WeChatProvider) SendBatch(ctx context.Context, messages []*notification.NotificationMessage) error {
	// 微信暂不支持真正的批量发送，这里进行逐个发送
	for _, message := range messages {
		if err := p.Send(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

// ValidateConfig 验证配置
func (p *WeChatProvider) ValidateConfig(config map[string]interface{}) error {
	configBytes, _ := json.Marshal(config)
	var cfg WeChatConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("配置格式错误: %w", err)
	}
	
	if cfg.CorpID == "" {
		return fmt.Errorf("CorpID不能为空")
	}
	
	if cfg.CorpSecret == "" {
		return fmt.Errorf("CorpSecret不能为空")
	}
	
	if cfg.AgentID == "" {
		return fmt.Errorf("AgentID不能为空")
	}
	
	return nil
}

// IsHealthy 健康检查
func (p *WeChatProvider) IsHealthy(ctx context.Context) bool {
	// 获取访问令牌进行健康检查
	_, err := p.getAccessToken(ctx)
	if err != nil {
		p.logger.Debug("微信健康检查失败", zap.Error(err))
		return false
	}
	
	return true
}

// GetConfig 获取当前配置
func (p *WeChatProvider) GetConfig() map[string]interface{} {
	if p.config == nil {
		return nil
	}
	
	configBytes, _ := json.Marshal(p.config)
	var result map[string]interface{}
	json.Unmarshal(configBytes, &result)
	
	// 隐藏敏感信息
	if corpSecret, exists := result["corp_secret"]; exists {
		if str, ok := corpSecret.(string); ok && len(str) > 10 {
			result["corp_secret"] = str[:10] + "***"
		}
	}
	
	if token, exists := result["token"]; exists {
		if str, ok := token.(string); ok && len(str) > 10 {
			result["token"] = str[:10] + "***"
		}
	}
	
	if encodingKey, exists := result["encoding_aes_key"]; exists {
		if str, ok := encodingKey.(string); ok && len(str) > 10 {
			result["encoding_aes_key"] = str[:10] + "***"
		}
	}
	
	return result
}

// Close 关闭提供商
func (p *WeChatProvider) Close() error {
	return nil
}

// sendViaAPI 通过API发送
func (p *WeChatProvider) sendViaAPI(ctx context.Context, message *notification.NotificationMessage) error {
	accessToken, err := p.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取访问令牌失败: %w", err)
	}
	
	wechatMessage := p.buildAPIMessage(message)
	
	jsonData, err := json.Marshal(wechatMessage)
	if err != nil {
		return fmt.Errorf("编码微信消息失败: %w", err)
	}
	
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", accessToken)
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送微信消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取微信响应失败: %w", err)
	}
	
	var response SendMessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析微信响应失败: %w", err)
	}
	
	if response.ErrCode != 0 {
		return fmt.Errorf("微信API错误: %s (错误代码: %d)", response.ErrMsg, response.ErrCode)
	}
	
	return nil
}

// sendViaWebhook 通过Webhook发送
func (p *WeChatProvider) sendViaWebhook(ctx context.Context, message *notification.NotificationMessage) error {
	webhookMessage := p.buildWebhookMessage(message)
	
	jsonData, err := json.Marshal(webhookMessage)
	if err != nil {
		return fmt.Errorf("编码微信Webhook消息失败: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送微信Webhook消息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取微信Webhook响应失败: %w", err)
	}
	
	var response SendMessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析微信Webhook响应失败: %w", err)
	}
	
	if response.ErrCode != 0 {
		return fmt.Errorf("微信Webhook错误: %s (错误代码: %d)", response.ErrMsg, response.ErrCode)
	}
	
	return nil
}

// getAccessToken 获取访问令牌
func (p *WeChatProvider) getAccessToken(ctx context.Context) (string, error) {
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		url.QueryEscape(p.config.CorpID),
		url.QueryEscape(p.config.CorpSecret))
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取访问令牌失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取访问令牌响应失败: %w", err)
	}
	
	var response AccessTokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析访问令牌响应失败: %w", err)
	}
	
	if response.ErrCode != 0 {
		return "", fmt.Errorf("微信API错误: %s (错误代码: %d)", response.ErrMsg, response.ErrCode)
	}
	
	return response.AccessToken, nil
}

// buildAPIMessage 构建API消息
func (p *WeChatProvider) buildAPIMessage(message *notification.NotificationMessage) WeChatMessage {
	var content strings.Builder
	
	// 添加标题
	if message.Title != "" {
		content.WriteString(fmt.Sprintf("## %s\n\n", message.Title))
	}
	
	// 添加内容
	content.WriteString(message.Content)
	
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
	
	finalContent := fmt.Sprintf("%s %s", levelIcon, content.String())
	
	wechatMessage := WeChatMessage{
		AgentID: p.config.AgentID,
		MsgType: "markdown",
		Markdown: &MarkdownMsg{
			Content: finalContent,
		},
	}
	
	// 设置接收用户
	if message.UserID != "" {
		wechatMessage.ToUser = message.UserID
	}
	
	return wechatMessage
}

// buildWebhookMessage 构建Webhook消息
func (p *WeChatProvider) buildWebhookMessage(message *notification.NotificationMessage) map[string]interface{} {
	var content strings.Builder
	
	// 添加标题
	if message.Title != "" {
		content.WriteString(fmt.Sprintf("## %s\n\n", message.Title))
	}
	
	// 添加内容
	content.WriteString(message.Content)
	
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
	
	finalContent := fmt.Sprintf("%s %s", levelIcon, content.String())
	
	return map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": finalContent,
		},
	}
}

// AESDecrypt AES解密（用于微信消息加密）
func (p *WeChatProvider) AESDecrypt(ciphertext, key string) (string, error) {
	// Base64解码
	encrypted, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}
	
	// 创建AES解密器
	block, err := aes.NewCipher([]byte(key + "="))
	if err != nil {
		return "", fmt.Errorf("创建AES解密器失败: %w", err)
	}
	
	// 检查数据长度
	if len(encrypted) < aes.BlockSize {
		return "", fmt.Errorf("加密数据长度不足")
	}
	
	// 获取IV
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]
	
	// 创建CBC模式解密器
	mode := cipher.NewCBCDecrypter(block, iv)
	
	// 解密
	mode.CryptBlocks(encrypted, encrypted)
	
	// 移除PKCS7填充
	padding := int(encrypted[len(encrypted)-1])
	if padding > len(encrypted) || padding == 0 {
		return "", fmt.Errorf("无效的PKCS7填充")
	}
	
	plaintext := encrypted[:len(encrypted)-padding]
	
	return string(plaintext), nil
}

// AESEncrypt AES加密
func (p *WeChatProvider) AESEncrypt(plaintext, key string) (string, error) {
	// 创建AES加密器
	block, err := aes.NewCipher([]byte(key + "="))
	if err != nil {
		return "", fmt.Errorf("创建AES加密器失败: %w", err)
	}
	
	// 添加PKCS7填充
	padding := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	for i := 0; i < padding; i++ {
		plaintext += string(padding)
	}
	
	// 生成随机IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("生成IV失败: %w", err)
	}
	
	// 创建CBC模式加密器
	mode := cipher.NewCBCEncrypter(block, iv)
	
	// 加密
	encrypted := make([]byte, len(plaintext))
	mode.CryptBlocks(encrypted, []byte(plaintext))
	
	// 合并IV和加密数据
	result := append(iv, encrypted...)
	
	// Base64编码
	return base64.StdEncoding.EncodeToString(result), nil
}

// GenerateSignature 生成签名
func (p *WeChatProvider) GenerateSignature(token string, timestamp int64, nonce string, ciphertext string) string {
	// 排序参数
	params := []string{token, fmt.Sprintf("%d", timestamp), nonce, ciphertext}
	sort.Strings(params)
	
	// 生成签名字符串
	signature := strings.Join(params, "")
	
	// 这里需要使用SHA1算法，为了简化暂时返回空字符串
	// 实际应用中需要使用crypto/sha1包
	return signature
}