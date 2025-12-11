package wechat

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
	"moviepilot-go/pkg/logger"
)

// Config WeChat 企业微信配置
type Config struct {
	// CorpID 企业ID
	CorpID string
	// AgentID 应用ID
	AgentID string
	// Secret 应用密钥
	Secret string
	// Timeout 请求超时时间
	Timeout time.Duration
}

// Client WeChat 企业微信通知客户端
type Client struct {
	corpID  string
	agentID string
	secret  string
	client  *http.Client
	logger  *zap.Logger

	// 缓存的 access_token
	accessToken string
	tokenExpiry time.Time
}

// NewClient 创建 WeChat 企业微信客户端
func NewClient(cfg Config) (*Client, error) {
	if cfg.CorpID == "" {
		return nil, fmt.Errorf("corp ID is required")
	}
	if cfg.AgentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}
	if cfg.Secret == "" {
		return nil, fmt.Errorf("secret is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	return &Client{
		corpID:  cfg.CorpID,
		agentID: cfg.AgentID,
		secret:  cfg.Secret,
		client:  &http.Client{Timeout: cfg.Timeout},
		logger:  logger.GetLogger(),
	}, nil
}

// Name 返回通知渠道名称
func (c *Client) Name() string {
	return "wechat"
}

// SendText 发送文本消息
func (c *Client) SendText(ctx context.Context, message string) error {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取 access_token 失败: %w", err)
	}

	payload := map[string]any{
		"touser":  "@all",
		"msgtype": "text",
		"agentid": c.agentID,
		"text": map[string]string{
			"content": message,
		},
	}

	return c.sendMessage(ctx, token, payload)
}

// SendImage 发送图片消息
func (c *Client) SendImage(ctx context.Context, imageURL string, caption string) error {
	// 企业微信发送图片需要先上传获取 media_id
	// 这里简化处理，发送文本+图片URL
	message := caption
	if message == "" {
		message = "图片"
	}
	message += "\n" + imageURL

	return c.SendText(ctx, message)
}

// SendFile 发送文件消息
func (c *Client) SendFile(ctx context.Context, fileURL string, filename string) error {
	// 企业微信发送文件需要先上传获取 media_id
	// 这里简化处理，发送文本+文件URL
	message := filename
	if message == "" {
		message = "文件"
	}
	message += "\n" + fileURL

	return c.SendText(ctx, message)
}

// SendMarkdown 发送 Markdown 消息
func (c *Client) SendMarkdown(ctx context.Context, markdown string) error {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取 access_token 失败: %w", err)
	}

	payload := map[string]any{
		"touser":  "@all",
		"msgtype": "markdown",
		"agentid": c.agentID,
		"markdown": map[string]string{
			"content": markdown,
		},
	}

	return c.sendMessage(ctx, token, payload)
}

// Send 发送通用消息
func (c *Client) Send(ctx context.Context, msg *notification.Message) error {
	switch msg.Type {
	case notification.NotificationTypeText:
		return c.SendText(ctx, msg.Content)
	case notification.NotificationTypeImage:
		return c.SendImage(ctx, msg.ImageURL, msg.Content)
	case notification.NotificationTypeFile:
		return c.SendFile(ctx, msg.FileURL, msg.FileName)
	case notification.NotificationTypeMarkdown:
		return c.SendMarkdown(ctx, msg.Content)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

// TestConnection 测试连接
func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("获取 access_token 失败: %w", err)
	}

	c.logger.Info("WeChat 企业微信连接测试成功")
	return nil
}

// getAccessToken 获取 access_token
func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	// 如果 token 未过期，直接返回
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		c.corpID, c.secret)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("API 返回错误: code=%d, msg=%s", result.ErrCode, result.ErrMsg)
	}

	// 缓存 token，提前 5 分钟过期
	c.accessToken = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)

	c.logger.Info("获取 access_token 成功", zap.Int("expires_in", result.ExpiresIn))
	return c.accessToken, nil
}

// sendMessage 发送消息
func (c *Client) sendMessage(ctx context.Context, token string, payload map[string]any) error {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("发送消息失败: code=%d, msg=%s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

// 编译期断言
var _ notification.Client = (*Client)(nil)
