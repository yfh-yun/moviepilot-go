package repositories

import (
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/model"

	"gorm.io/gorm"
)

// messageRepository 消息仓储实现
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓储
func NewMessageRepository(db *gorm.DB) interfaces.MessageRepository {
	return &messageRepository{db: db}
}

// Create 创建消息
func (r *messageRepository) Create(message *model.Message) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}
	return r.db.Create(message).Error
}

// Update 更新消息
func (r *messageRepository) Update(message *model.Message) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}
	return r.db.Save(message).Error
}

// Delete 删除消息
func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&model.Message{}, id).Error
}

// GetByID 根据ID获取消息
func (r *messageRepository) GetByID(id uint) (*model.Message, error) {
	var message model.Message
	err := r.db.First(&message, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// List 获取消息列表
func (r *messageRepository) List(offset, limit int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	err := r.db.Model(&model.Message{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&messages).Error
	return messages, total, err
}

// GetUnread 获取未读消息
func (r *messageRepository) GetUnread(userID *int) ([]*model.Message, error) {
	var messages []*model.Message
	query := r.db.Where("read = ?", false)
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	
	err := query.Order("created_at DESC").Find(&messages).Error
	return messages, err
}

// CountUnread 统计未读消息数量
func (r *messageRepository) CountUnread(userID *int) (int64, error) {
	var count int64
	query := r.db.Model(&model.Message{}).Where("read = ?", false)
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	
	err := query.Count(&count).Error
	return count, err
}

// GetByUserID 根据用户ID获取消息
func (r *messageRepository) GetByUserID(userID uint, offset, limit int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	err := r.db.Model(&model.Message{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&messages).Error
	return messages, total, err
}

// MarkAsRead 标记消息为已读
func (r *messageRepository) MarkAsRead(id uint) error {
	result := r.db.Model(&model.Message{}).Where("id = ?", id).Update("read", true)
	if result.Error != nil {
		return fmt.Errorf("failed to mark message as read: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("message with id %d not found", id)
	}
	return nil
}

// MarkAllAsRead 标记所有消息为已读
func (r *messageRepository) MarkAllAsRead(userID *int) error {
	query := r.db.Model(&model.Message{}).Where("read = ?", false)
	
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	
	return query.Update("read", true).Error
}