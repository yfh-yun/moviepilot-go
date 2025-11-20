package notification

import (
	"context"
	"time"
)

// NotificationLevel 通知级别
type NotificationLevel string

const (
	LevelInfo     NotificationLevel = "info"
	LevelSuccess  NotificationLevel = "success"
	LevelWarning  NotificationLevel = "warning"
	LevelError    NotificationLevel = "error"
	LevelCritical NotificationLevel = "critical"
)

// NotificationMessage 通知消息
type NotificationMessage struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Level     NotificationLevel       `json:"level"`
	Channel   string                 `json:"channel"`
	UserID    string                 `json:"user_id,omitempty"`
	ImageURL  string                 `json:"image_url,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationProvider 通知提供商接口
type NotificationProvider interface {
	// GetName 获取提供商名称
	GetName() string
	
	// GetType 获取提供商类型
	GetType() string
	
	// Send 发送通知消息
	Send(ctx context.Context, message *NotificationMessage) error
	
	// SendBatch 批量发送通知消息
	SendBatch(ctx context.Context, messages []*NotificationMessage) error
	
	// ValidateConfig 验证配置
	ValidateConfig(config map[string]interface{}) error
	
	// IsHealthy 健康检查
	IsHealthy(ctx context.Context) bool
	
	// GetConfig 获取当前配置
	GetConfig() map[string]interface{}
	
	// Initialize 初始化提供商
	Initialize(config map[string]interface{}) error
	
	// Close 关闭提供商
	Close() error
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Providers map[string]ProviderConfig `json:"providers"`
	Rules     []NotificationRule        `json:"rules"`
	Queue     QueueConfig              `json:"queue"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config"`
	Priority int                    `json:"priority"`
}

// NotificationRule 通知规则
type NotificationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     []RuleAction           `json:"actions"`
	Providers   []string               `json:"providers"`
	Priority    int                    `json:"priority"`
}

// RuleAction 规则动作
type RuleAction struct {
	Type      string                 `json:"type"`
	Template  string                 `json:"template"`
	Variables map[string]interface{} `json:"variables"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	Enabled     bool          `json:"enabled"`
	MaxSize     int           `json:"max_size"`
	BatchSize   int           `json:"batch_size"`
	RetryCount  int           `json:"retry_count"`
	RetryDelay  time.Duration `json:"retry_delay"`
	FlushPeriod time.Duration `json:"flush_period"`
}

// DeliveryStatus 投递状态
type DeliveryStatus struct {
	MessageID  string    `json:"message_id"`
	Provider   string    `json:"provider"`
	Status     string    `json:"status"` // pending, sent, failed, retry
	Error      string    `json:"error,omitempty"`
	SentAt     time.Time `json:"sent_at"`
	RetryCount int       `json:"retry_count"`
}

// NotificationHistory 通知历史
type NotificationHistory struct {
	ID        string              `json:"id"`
	Message   *NotificationMessage `json:"message"`
	Status    []*DeliveryStatus    `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}