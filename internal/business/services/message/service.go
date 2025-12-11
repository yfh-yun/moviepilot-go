package message

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/pkg/logger"
)

// Service 消息服务接口
type Service interface {
	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, msg *Message) error

	// GetMessages 获取消息列表
	GetMessages(ctx context.Context, userID string, limit int) ([]*Message, error)

	// GetUnreadCount 获取未读消息数
	GetUnreadCount(ctx context.Context, userID string) (int64, error)

	// MarkAsRead 标记为已读
	MarkAsRead(ctx context.Context, messageID uint) error

	// DeleteMessage 删除消息
	DeleteMessage(ctx context.Context, messageID uint) error

	// ClearMessages 清空消息
	ClearMessages(ctx context.Context, userID string) error
}

// service 服务实现
type service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建服务
func NewService(db *gorm.DB) Service {
	return &service{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// Message 消息模型
type Message struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    string     `gorm:"size:100;index" json:"user_id"`
	Title     string     `gorm:"size:200" json:"title"`
	Content   string     `gorm:"type:text" json:"content"`
	Type      string     `gorm:"size:20" json:"type"` // info, success, warning, error
	Channel   string     `gorm:"size:50" json:"channel"`
	Link      string     `gorm:"size:500" json:"link"`
	Image     string     `gorm:"size:500" json:"image"`
	IsRead    bool       `gorm:"default:false;index" json:"is_read"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ReadAt    *time.Time `json:"read_at"`
}

// TableName 表名
func (Message) TableName() string {
	return "messages"
}

// CreateMessage 创建消息
func (s *service) CreateMessage(ctx context.Context, msg *Message) error {
	s.logger.Info("创建消息",
		zap.String("user_id", msg.UserID),
		zap.String("title", msg.Title),
	)

	return s.db.WithContext(ctx).Create(msg).Error
}

// GetMessages 获取消息列表
func (s *service) GetMessages(ctx context.Context, userID string, limit int) ([]*Message, error) {
	s.logger.Info("获取消息列表",
		zap.String("user_id", userID),
		zap.Int("limit", limit),
	)

	var messages []*Message
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

// GetUnreadCount 获取未读消息数
func (s *service) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&Message{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error

	return count, err
}

// MarkAsRead 标记为已读
func (s *service) MarkAsRead(ctx context.Context, messageID uint) error {
	s.logger.Info("标记消息为已读", zap.Uint("message_id", messageID))

	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&Message{}).
		Where("id = ?", messageID).
		Updates(map[string]any{
			"is_read": true,
			"read_at": now,
		}).Error
}

// DeleteMessage 删除消息
func (s *service) DeleteMessage(ctx context.Context, messageID uint) error {
	s.logger.Info("删除消息", zap.Uint("message_id", messageID))

	return s.db.WithContext(ctx).
		Delete(&Message{}, messageID).Error
}

// ClearMessages 清空消息
func (s *service) ClearMessages(ctx context.Context, userID string) error {
	s.logger.Info("清空消息", zap.String("user_id", userID))

	return s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&Message{}).Error
}
