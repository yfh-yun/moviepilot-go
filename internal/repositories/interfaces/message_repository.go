package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// MessageRepository 消息仓储接口
type MessageRepository interface {
	// Create 创建消息
	Create(ctx context.Context, message *database.Message) error

	// GetByID 根据ID获取消息
	GetByID(ctx context.Context, id string) (*database.Message, error)

	// Update 更新消息
	Update(ctx context.Context, message *database.Message) error

	// Delete 删除消息
	Delete(ctx context.Context, id string) error

	// List 获取消息列表
	List(ctx context.Context, params ListMessageParams) ([]*database.Message, int64, error)

	// GetByUserID 根据用户ID获取消息列表
	GetByUserID(ctx context.Context, userID string, params ListMessageParams) ([]*database.Message, int64, error)

	// MarkAsRead 标记为已读
	MarkAsRead(ctx context.Context, id string) error

	// MarkAllAsRead 标记所有消息为已读
	MarkAllAsRead(ctx context.Context, userID string) error

	// GetUnreadCount 获取未读消息数量
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
}

// ListMessageParams 消息列表查询参数
type ListMessageParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	UserID   string `json:"user_id"`
}
