package notification

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Service 通知服务接口
type Service interface {
	// Send 发送通知
	Send(ctx context.Context, notification *Notification) error

	// SendToChannel 发送到指定渠道
	SendToChannel(ctx context.Context, channel Channel, notification *Notification) error

	// SendBatch 批量发送
	SendBatch(ctx context.Context, notifications []*Notification) error

	// GetHistory 获取通知历史
	GetHistory(ctx context.Context, limit int) ([]*NotificationRecord, error)

	// GetStats 获取通知统计
	GetStats(ctx context.Context) (*NotificationStats, error)
}

// service 通知服务实现
type service struct {
	channels map[Channel]Notifier
	logger   *zap.Logger
}

// NewService 创建通知服务
func NewService() Service {
	return &service{
		channels: make(map[Channel]Notifier),
		logger:   logger.GetLogger(),
	}
}

// RegisterChannel 注册通知渠道
func (s *service) RegisterChannel(channel Channel, notifier Notifier) {
	s.channels[channel] = notifier
	s.logger.Info("注册通知渠道", zap.String("channel", string(channel)))
}

// Notification 通知消息
type Notification struct {
	Title    string           `json:"title"`
	Content  string           `json:"content"`
	Type     NotificationType `json:"type"`
	Priority Priority         `json:"priority"`
	Channels []Channel        `json:"channels"` // 指定发送渠道
	Data     map[string]any   `json:"data"`     // 额外数据
}

// NotificationType 通知类型
type NotificationType string

const (
	TypeInfo      NotificationType = "info"
	TypeSuccess   NotificationType = "success"
	TypeWarning   NotificationType = "warning"
	TypeError     NotificationType = "error"
	TypeDownload  NotificationType = "download"
	TypeSubscribe NotificationType = "subscribe"
)

// Priority 优先级
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 5
	PriorityHigh   Priority = 8
	PriorityUrgent Priority = 10
)

// Channel 通知渠道
type Channel string

const (
	ChannelTelegram Channel = "telegram"
	ChannelWeChat   Channel = "wechat"
	ChannelEmail    Channel = "email"
	ChannelWebhook  Channel = "webhook"
	ChannelWebUI    Channel = "webui"
)

// NotificationRecord 通知记录
type NotificationRecord struct {
	ID        uint             `json:"id"`
	Title     string           `json:"title"`
	Content   string           `json:"content"`
	Type      NotificationType `json:"type"`
	Priority  Priority         `json:"priority"`
	Channel   Channel          `json:"channel"`
	Success   bool             `json:"success"`
	ErrorMsg  string           `json:"error_msg"`
	CreatedAt time.Time        `json:"created_at"`
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalSent   int64                      `json:"total_sent"`
	SuccessSent int64                      `json:"success_sent"`
	FailedSent  int64                      `json:"failed_sent"`
	SuccessRate float64                    `json:"success_rate"`
	ByChannel   map[Channel]int64          `json:"by_channel"`
	ByType      map[NotificationType]int64 `json:"by_type"`
}

// Send 发送通知
func (s *service) Send(ctx context.Context, notification *Notification) error {
	s.logger.Info("发送通知",
		zap.String("title", notification.Title),
		zap.String("type", string(notification.Type)),
		zap.Int("priority", int(notification.Priority)),
	)

	// 如果未指定渠道，使用所有已注册渠道
	channels := notification.Channels
	if len(channels) == 0 {
		for channel := range s.channels {
			channels = append(channels, channel)
		}
	}

	// 发送到各个渠道
	var lastErr error
	for _, channel := range channels {
		if err := s.SendToChannel(ctx, channel, notification); err != nil {
			s.logger.Error("发送到渠道失败",
				zap.String("channel", string(channel)),
				zap.Error(err),
			)
			lastErr = err
		}
	}

	return lastErr
}

// SendToChannel 发送到指定渠道
func (s *service) SendToChannel(ctx context.Context, channel Channel, notification *Notification) error {
	notifier, ok := s.channels[channel]
	if !ok {
		return fmt.Errorf("渠道未注册: %s", channel)
	}

	s.logger.Debug("发送到渠道",
		zap.String("channel", string(channel)),
		zap.String("title", notification.Title),
	)

	return notifier.Send(ctx, notification)
}

// SendBatch 批量发送
func (s *service) SendBatch(ctx context.Context, notifications []*Notification) error {
	s.logger.Info("批量发送通知", zap.Int("count", len(notifications)))

	for _, notification := range notifications {
		if err := s.Send(ctx, notification); err != nil {
			s.logger.Error("批量发送失败",
				zap.String("title", notification.Title),
				zap.Error(err),
			)
		}
	}

	return nil
}

// GetHistory 获取通知历史
func (s *service) GetHistory(ctx context.Context, limit int) ([]*NotificationRecord, error) {
	// TODO: 从数据库获取历史记录
	return []*NotificationRecord{}, nil
}

// GetStats 获取通知统计
func (s *service) GetStats(ctx context.Context) (*NotificationStats, error) {
	// TODO: 从数据库获取统计数据
	return &NotificationStats{
		ByChannel: make(map[Channel]int64),
		ByType:    make(map[NotificationType]int64),
	}, nil
}

// Notifier 通知器接口
type Notifier interface {
	// Send 发送通知
	Send(ctx context.Context, notification *Notification) error

	// Test 测试连接
	Test(ctx context.Context) error

	// Name 获取名称
	Name() string
}
