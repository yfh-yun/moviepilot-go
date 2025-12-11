package telegram

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

// Config Telegram 配置
type Config struct {
	// BotToken Telegram Bot Token
	BotToken string
	// ChatID 目标聊天ID（可以是用户ID或群组ID）
	ChatID string
	// Timeout 请求超时时间
	Timeout time.Duration
	// ParseMode 消息解析模式（Markdown, HTML, 或空）
	ParseMode string
}

// Client Telegram 通知客户端
type Client struct {
	botToken  string
	chatID    string
	parseMode string
	client    *http.Client
	logger    *zap.Logger
}

// NewClient 创建 Telegram 客户端
func NewClient(cfg Config) (*Client, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}
	if cfg.ChatID == "" {
		return nil, fmt.Errorf("chat ID is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.ParseMode == "" {
		cfg.ParseMode = "Markdown"
	}

	return &Client{
		botToken:  cfg.BotToken,
		chatID:    cfg.ChatID,
		parseMode: cfg.ParseMode,
		client:    &http.Client{Timeout: cfg.Timeout},
		logger:    logger.GetLogger(),
	}, nil
}

// Name 返回通知渠道名称
func (c *Client) Name() string {
	return "telegram"
}

// SendText 发送文本消息
func (c *Client) SendText(ctx context.Context, message string) error {
	return c.sendMessage(ctx, map[string]any{
		"chat_id":    c.chatID,
		"text":       message,
		"parse_mode": c.parseMode,
	})
}

// SendImage 发送图片消息
func (c *Client) SendImage(ctx context.Context, imageURL string, caption string) error {
	payload := map[string]any{
		"chat_id": c.chatID,
		"photo":   imageURL,
	}
	if caption != "" {
		payload["caption"] = caption
		payload["parse_mode"] = c.parseMode
	}

	return c.sendPhoto(ctx, payload)
}

// SendFile 发送文件消息
func (c *Client) SendFile(ctx context.Context, fileURL string, filename string) error {
	payload := map[string]any{
		"chat_id":  c.chatID,
		"document": fileURL,
	}
	if filename != "" {
		payload["caption"] = filename
	}

	return c.sendDocument(ctx, payload)
}

// SendMarkdown 发送 Markdown 消息
func (c *Client) SendMarkdown(ctx context.Context, markdown string) error {
	return c.sendMessage(ctx, map[string]any{
		"chat_id":    c.chatID,
		"text":       markdown,
		"parse_mode": "Markdown",
	})
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
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", c.botToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("Telegram API 返回 ok=false")
	}

	c.logger.Info("Telegram 连接测试成功", zap.String("bot_username", result.Result.Username))
	return nil
}

// sendMessage 发送消息到 Telegram
func (c *Client) sendMessage(ctx context.Context, payload map[string]any) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)
	return c.doRequest(ctx, url, payload)
}

// sendPhoto 发送图片到 Telegram
func (c *Client) sendPhoto(ctx context.Context, payload map[string]any) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", c.botToken)
	return c.doRequest(ctx, url, payload)
}

// sendDocument 发送文件到 Telegram
func (c *Client) sendDocument(ctx context.Context, payload map[string]any) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", c.botToken)
	return c.doRequest(ctx, url, payload)
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(ctx context.Context, url string, payload map[string]any) error {
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
		return fmt.Errorf("Telegram API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("Telegram API 返回错误: %s", result.Description)
	}

	return nil
}

// 编译期断言
var _ notification.Client = (*Client)(nil)
