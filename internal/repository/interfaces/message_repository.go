package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/model"
)

// MessageRepository 消息仓储接口
type MessageRepository interface {
	// Create 创建消息
	Create(message *model.Message) error

	// Update 更新消息
	Update(message *model.Message) error

	// Delete 删除消息
	Delete(id uint) error

	// GetByID 根据ID获取消息
	GetByID(id uint) (*model.Message, error)

	// List 获取消息列表
	List(offset, limit int) ([]*model.Message, int64, error)

	// GetUnread 获取未读消息
	GetUnread(userID *int) ([]*model.Message, error)

	// CountUnread 统计未读消息数量
	CountUnread(userID *int) (int64, error)

	// GetByUserID 根据用户ID获取消息
	GetByUserID(userID uint, offset, limit int) ([]*model.Message, int64, error)

	// MarkAsRead 标记消息为已读
	MarkAsRead(id uint) error

	// MarkAllAsRead 标记所有消息为已读
	MarkAllAsRead(userID *int) error
}
