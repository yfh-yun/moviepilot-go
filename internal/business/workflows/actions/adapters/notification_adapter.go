package adapters

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// NotificationChannel 定义通知渠道
type NotificationChannel string

// NotificationChannel 常量定义
const (
	NotificationChannelEmail      = NotificationChannel("email")      // 邮件
	NotificationChannelTelegram   = NotificationChannel("telegram")   // Telegram
	NotificationChannelWebhook    = NotificationChannel("webhook")    // Webhook
	NotificationChannelSlack      = NotificationChannel("slack")      // Slack
	NotificationChannelDiscord    = NotificationChannel("discord")    // Discord
	NotificationChannelWeChat     = NotificationChannel("wechat")     // 微信
	NotificationChannelSMS        = NotificationChannel("sms")        // SMS
	NotificationChannelPushbullet = NotificationChannel("pushbullet") // Pushbullet
	NotificationChannelPushover   = NotificationChannel("pushover")   // Pushover
	NotificationChannelOther      = NotificationChannel("other")      // 其他
)

// NotificationLevel 定义通知级别
type NotificationLevel string

// NotificationLevel 常量定义
const (
	NotificationLevelLow    = NotificationLevel("low")    // 低优先级
	NotificationLevelMedium = NotificationLevel("medium") // 中优先级
	NotificationLevelHigh   = NotificationLevel("high")   // 高优先级
)

// Notification 定义通知内容
type Notification struct {
	ID          string              `json:"id"`          // 通知ID
	Title       string              `json:"title"`       // 通知标题
	Content     string              `json:"content"`     // 通知内容
	Channel     NotificationChannel `json:"channel"`     // 通知渠道
	Level       NotificationLevel   `json:"level"`       // 通知级别
	Recipients  []string            `json:"recipients"`  // 接收者
	Attachments []string            `json:"attachments"` // 附件列表
	Metadata    map[string]any      `json:"metadata"`    // 元数据
	Status      string              `json:"status"`      // 通知状态
	SentAt      time.Time           `json:"sent_at"`     // 发送时间
	CreatedAt   time.Time           `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time           `json:"updated_at"`  // 更新时间
}

// NotificationStatus 定义通知状态
const (
	NotificationStatusPending   = "pending"   // 待发送
	NotificationStatusSent      = "sent"      // 已发送
	NotificationStatusFailed    = "failed"    // 发送失败
	NotificationStatusCancelled = "cancelled" // 已取消
)

// NotificationService 定义通知服务接口
type NotificationService interface {
	// SendNotification 发送通知
	SendNotification(ctx context.Context, notification Notification) (string, error)

	// SendNotificationBatch 批量发送通知
	SendNotificationBatch(ctx context.Context, notifications []Notification) ([]string, error)

	// GetNotification 获取单个通知
	GetNotification(ctx context.Context, notificationID string) (*Notification, error)

	// GetNotifications 获取通知列表
	GetNotifications(ctx context.Context, params GetNotificationsParams) ([]Notification, error)

	// GetChannels 获取支持的通知渠道列表
	GetChannels(ctx context.Context) ([]NotificationChannel, error)

	// ValidateChannel 验证通知渠道配置
	ValidateChannel(ctx context.Context, channel NotificationChannel, config map[string]any) error
}

// GetNotificationsParams 获取通知列表参数
type GetNotificationsParams struct {
	Channel   NotificationChannel `json:"channel"`    // 通知渠道过滤
	Level     NotificationLevel   `json:"level"`      // 通知级别过滤
	Status    string              `json:"status"`     // 通知状态过滤
	Limit     int                 `json:"limit"`      // 返回结果数量限制
	Offset    int                 `json:"offset"`     // 偏移量
	SortBy    string              `json:"sort_by"`    // 排序字段
	SortOrder string              `json:"sort_order"` // 排序顺序
	StartDate time.Time           `json:"start_date"` // 开始日期
	EndDate   time.Time           `json:"end_date"`   // 结束日期
}

// NotificationServiceAdapter 通知服务适配器实现
type NotificationServiceAdapter struct {
	logger *zap.Logger
	// 实际的通知服务客户端可以在这里注入
}

// NewNotificationServiceAdapter 创建新的通知服务适配器实例
func NewNotificationServiceAdapter(logger *zap.Logger) *NotificationServiceAdapter {
	return &NotificationServiceAdapter{
		logger: logger,
	}
}

// SendNotification 发送通知
func (a *NotificationServiceAdapter) SendNotification(ctx context.Context, notification Notification) (string, error) {
	// 实际实现中，这里应该调用核心业务服务的通知API
	// 这里使用模拟实现，返回一个随机生成的通知ID
	a.logger.Info("Sending notification", zap.String("channel", string(notification.Channel)), zap.String("level", string(notification.Level)), zap.String("title", notification.Title))
	return "notification-" + time.Now().Format("20060102150405"), nil
}

// SendNotificationBatch 批量发送通知
func (a *NotificationServiceAdapter) SendNotificationBatch(ctx context.Context, notifications []Notification) ([]string, error) {
	// 实际实现中，这里应该调用核心业务服务的通知API
	// 这里使用模拟实现，返回随机生成的通知ID列表
	a.logger.Info("Sending notification batch", zap.Int("notification_count", len(notifications)))

	var notificationIDs []string
	for range notifications {
		notificationIDs = append(notificationIDs, "notification-"+time.Now().Format("20060102150405"))
	}

	return notificationIDs, nil
}

// GetNotification 获取单个通知
func (a *NotificationServiceAdapter) GetNotification(ctx context.Context, notificationID string) (*Notification, error) {
	// 实际实现中，这里应该调用核心业务服务的通知API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Getting notification", zap.String("notification_id", notificationID))
	return nil, nil
}

// GetNotifications 获取通知列表
func (a *NotificationServiceAdapter) GetNotifications(ctx context.Context, params GetNotificationsParams) ([]Notification, error) {
	// 实际实现中，这里应该调用核心业务服务的通知API
	// 这里使用模拟实现，返回空列表
	a.logger.Info("Getting notifications", zap.String("channel", string(params.Channel)), zap.String("level", string(params.Level)), zap.String("status", params.Status))
	return []Notification{}, nil
}

// GetChannels 获取支持的通知渠道列表
func (a *NotificationServiceAdapter) GetChannels(ctx context.Context) ([]NotificationChannel, error) {
	// 实际实现中，这里应该调用核心业务服务的通知API
	// 这里使用模拟实现，返回所有支持的通知渠道
	a.logger.Info("Getting notification channels")
	return []NotificationChannel{
		NotificationChannelEmail,
		NotificationChannelTelegram,
		NotificationChannelWebhook,
		NotificationChannelSlack,
		NotificationChannelDiscord,
		NotificationChannelWeChat,
		NotificationChannelSMS,
		NotificationChannelPushbullet,
		NotificationChannelPushover,
		NotificationChannelOther,
	}, nil
}

// ValidateChannel 验证通知渠道配置
func (a *NotificationServiceAdapter) ValidateChannel(ctx context.Context, channel NotificationChannel, config map[string]any) error {
	// 实际实现中，这里应该调用核心业务服务的通知API
	// 这里使用模拟实现，返回nil
	a.logger.Info("Validating notification channel", zap.String("channel", string(channel)))
	return nil
}

// MockNotificationService 模拟通知服务实现，用于测试
type MockNotificationService struct {
	logger        *zap.Logger
	notifications map[string]Notification
}

// NewMockNotificationService 创建新的模拟通知服务实例
func NewMockNotificationService(logger *zap.Logger) *MockNotificationService {
	return &MockNotificationService{
		logger:        logger,
		notifications: make(map[string]Notification),
	}
}

// SendNotification 发送通知（模拟实现）
func (m *MockNotificationService) SendNotification(ctx context.Context, notification Notification) (string, error) {
	m.logger.Info("Mock sending notification", zap.String("channel", string(notification.Channel)), zap.String("level", string(notification.Level)), zap.String("title", notification.Title))

	// 创建模拟通知
	notificationID := "mock-notification-" + time.Now().Format("20060102150405")
	newNotification := Notification{
		ID:          notificationID,
		Title:       notification.Title,
		Content:     notification.Content,
		Channel:     notification.Channel,
		Level:       notification.Level,
		Recipients:  notification.Recipients,
		Attachments: notification.Attachments,
		Metadata:    notification.Metadata,
		Status:      NotificationStatusSent,
		SentAt:      time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 如果未设置通知级别，使用默认值
	if newNotification.Level == "" {
		newNotification.Level = NotificationLevelMedium
	}

	m.notifications[notificationID] = newNotification
	return notificationID, nil
}

// SendNotificationBatch 批量发送通知（模拟实现）
func (m *MockNotificationService) SendNotificationBatch(ctx context.Context, notifications []Notification) ([]string, error) {
	m.logger.Info("Mock sending notification batch", zap.Int("notification_count", len(notifications)))

	var notificationIDs []string
	for _, notification := range notifications {
		notificationID, _ := m.SendNotification(ctx, notification)
		notificationIDs = append(notificationIDs, notificationID)
	}

	return notificationIDs, nil
}

// GetNotification 获取单个通知（模拟实现）
func (m *MockNotificationService) GetNotification(ctx context.Context, notificationID string) (*Notification, error) {
	m.logger.Info("Mock getting notification", zap.String("notification_id", notificationID))

	notification, exists := m.notifications[notificationID]
	if !exists {
		return nil, nil
	}

	return &notification, nil
}

// GetNotifications 获取通知列表（模拟实现）
func (m *MockNotificationService) GetNotifications(ctx context.Context, params GetNotificationsParams) ([]Notification, error) {
	m.logger.Info("Mock getting notifications", zap.String("channel", string(params.Channel)), zap.String("level", string(params.Level)), zap.String("status", params.Status))

	var notifications []Notification
	for _, notification := range m.notifications {
		if (params.Channel == "" || notification.Channel == params.Channel) &&
			(params.Level == "" || notification.Level == params.Level) &&
			(params.Status == "" || notification.Status == params.Status) {
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
}

// GetChannels 获取支持的通知渠道列表（模拟实现）
func (m *MockNotificationService) GetChannels(ctx context.Context) ([]NotificationChannel, error) {
	m.logger.Info("Mock getting notification channels")
	return []NotificationChannel{
		NotificationChannelEmail,
		NotificationChannelTelegram,
		NotificationChannelWebhook,
		NotificationChannelSlack,
		NotificationChannelDiscord,
		NotificationChannelWeChat,
		NotificationChannelSMS,
		NotificationChannelPushbullet,
		NotificationChannelPushover,
		NotificationChannelOther,
	}, nil
}

// ValidateChannel 验证通知渠道配置（模拟实现）
func (m *MockNotificationService) ValidateChannel(ctx context.Context, channel NotificationChannel, config map[string]any) error {
	m.logger.Info("Mock validating notification channel", zap.String("channel", string(channel)))
	return nil
}
