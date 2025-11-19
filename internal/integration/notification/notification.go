package notification

import (
	"context"
	"fmt"
	"time"
)

// NotificationLevel 定义通知级别
type NotificationLevel string

const (
	LevelInfo    NotificationLevel = "info"
	LevelWarning NotificationLevel = "warning"
	LevelError   NotificationLevel = "error"
	LevelSuccess NotificationLevel = "success"
)

// NotificationPriority 定义通知优先级
type NotificationPriority int

const (
	PriorityLow NotificationPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

// Message 定义通知消息结构
type Message struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Level      NotificationLevel      `json:"level"`
	Priority   NotificationPriority   `json:"priority"`
	Channel    string                 `json:"channel"`
	Category   string                 `json:"category"`
	Tags       []string               `json:"tags"`
	ImageURL   string                 `json:"image_url,omitempty"`
	LinkURL    string                 `json:"link_url,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	TTL        time.Duration          `json:"ttl,omitempty"`
	CustomData map[string]interface{} `json:"custom_data,omitempty"`
}

// Sender 定义通知发送器接口
type Sender interface {
	// Name 返回发送器名称
	Name() string

	// SupportedLevels 返回支持的通知级别
	SupportedLevels() []NotificationLevel

	// Send 发送通知消息
	Send(ctx context.Context, message *Message) error

	// Validate 验证消息是否适合此发送器
	Validate(message *Message) error

	// HealthCheck 检查发送器健康状态
	HealthCheck(ctx context.Context) error

	// Close 关闭发送器
	Close() error
}

// Channel 定义通知渠道
type Channel struct {
	Name        string            `json:"name"`
	Enabled     bool              `json:"enabled"`
	Sender      Sender            `json:"-"`
	Config      map[string]string `json:"config"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NotificationResult 定义通知发送结果
type NotificationResult struct {
	Channel   string    `json:"channel"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	SentAt    time.Time `json:"sent_at"`
	MessageID string    `json:"message_id,omitempty"`
}

// Error 定义通知错误类型
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 定义错误码
var (
	ErrChannelDisabled = Error{
		Code:    "CHANNEL_DISABLED",
		Message: "notification channel is disabled",
	}

	ErrUnsupportedLevel = Error{
		Code:    "UNSUPPORTED_LEVEL",
		Message: "notification level is not supported by this channel",
	}

	ErrInvalidMessage = Error{
		Code:    "INVALID_MESSAGE",
		Message: "message validation failed",
	}

	ErrChannelUnhealthy = Error{
		Code:    "CHANNEL_UNHEALTHY",
		Message: "notification channel is unhealthy",
	}
)

// MessageBuilder 提供链式API构建消息
type MessageBuilder struct {
	message *Message
}

// NewMessageBuilder 创建新的消息构建器
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		message: &Message{
			ID:         generateMessageID(),
			Timestamp:  time.Now(),
			Tags:       make([]string, 0),
			CustomData: make(map[string]interface{}),
			Level:      LevelInfo,
			Priority:   PriorityNormal,
		},
	}
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// WithTitle 设置消息标题
func (b *MessageBuilder) WithTitle(title string) *MessageBuilder {
	b.message.Title = title
	return b
}

// WithContent 设置消息内容
func (b *MessageBuilder) WithContent(content string) *MessageBuilder {
	b.message.Content = content
	return b
}

// WithLevel 设置消息级别
func (b *MessageBuilder) WithLevel(level NotificationLevel) *MessageBuilder {
	b.message.Level = level
	return b
}

// WithPriority 设置消息优先级
func (b *MessageBuilder) WithPriority(priority NotificationPriority) *MessageBuilder {
	b.message.Priority = priority
	return b
}

// WithChannel 设置目标渠道
func (b *MessageBuilder) WithChannel(channel string) *MessageBuilder {
	b.message.Channel = channel
	return b
}

// WithCategory 设置消息分类
func (b *MessageBuilder) WithCategory(category string) *MessageBuilder {
	b.message.Category = category
	return b
}

// WithTags 设置消息标签
func (b *MessageBuilder) WithTags(tags ...string) *MessageBuilder {
	b.message.Tags = append(b.message.Tags, tags...)
	return b
}

// WithImageURL 设置图片URL
func (b *MessageBuilder) WithImageURL(url string) *MessageBuilder {
	b.message.ImageURL = url
	return b
}

// WithLinkURL 设置链接URL
func (b *MessageBuilder) WithLinkURL(url string) *MessageBuilder {
	b.message.LinkURL = url
	return b
}

// WithTTL 设置消息TTL
func (b *MessageBuilder) WithTTL(ttl time.Duration) *MessageBuilder {
	b.message.TTL = ttl
	return b
}

// WithCustomData 设置自定义数据
func (b *MessageBuilder) WithCustomData(key string, value interface{}) *MessageBuilder {
	if b.message.CustomData == nil {
		b.message.CustomData = make(map[string]interface{})
	}
	b.message.CustomData[key] = value
	return b
}

// Build 构建消息
func (b *MessageBuilder) Build() *Message {
	return b.message
}

// Validate 验证消息是否有效
func (m *Message) Validate() error {
	if m.Title == "" && m.Content == "" {
		return ErrInvalidMessage
	}

	if m.Channel == "" {
		return ErrInvalidMessage
	}

	return nil
}

// IsExpired 检查消息是否已过期
func (m *Message) IsExpired() bool {
	if m.TTL == 0 {
		return false
	}
	return time.Since(m.Timestamp) > m.TTL
}

// GetSummary 获取消息摘要
func (m *Message) GetSummary() string {
	if len(m.Content) > 100 {
		return m.Content[:100] + "..."
	}
	return m.Content
}
