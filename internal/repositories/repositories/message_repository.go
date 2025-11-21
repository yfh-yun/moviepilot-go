package repositories

import (
	"context"
	"errors"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/models"
)

// MessageRepositoryImpl 消息仓储实现
type MessageRepositoryImpl struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓储实例
func NewMessageRepository(db *gorm.DB) interfaces.MessageRepository {
	return &MessageRepositoryImpl{db: db}
}

// Create 创建消息
func (r *MessageRepositoryImpl) Create(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetByID 根据ID获取消息
func (r *MessageRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).First(&message, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// Update 更新消息
func (r *MessageRepositoryImpl) Update(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

// Delete 删除消息
func (r *MessageRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Message{}, "id = ?", id).Error
}

// List 获取消息列表
func (r *MessageRepositoryImpl) List(ctx context.Context, params interfaces.ListMessageParams) ([]*models.Message, int64, error) {
	var messages []*models.Message
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Message{})
	
	// 添加过滤条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&messages).Error
	
	return messages, total, err
}

// GetByUserID 根据用户ID获取消息列表
func (r *MessageRepositoryImpl) GetByUserID(ctx context.Context, userID string, params interfaces.ListMessageParams) ([]*models.Message, int64, error) {
	var messages []*models.Message
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Message{}).Where("user_id = ?", userID)
	
	// 添加过滤条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&messages).Error
	
	return messages, total, err
}

// MarkAsRead 标记为已读
func (r *MessageRepositoryImpl) MarkAsRead(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.Message{}).Where("id = ?", id).Update("status", "read").Error
}

// MarkAllAsRead 标记所有消息为已读
func (r *MessageRepositoryImpl) MarkAllAsRead(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&models.Message{}).Where("user_id = ? AND status != ?", userID, "read").Update("status", "read").Error
}

// GetUnreadCount 获取未读消息数量
func (r *MessageRepositoryImpl) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Message{}).Where("user_id = ? AND status = ?", userID, "unread").Count(&count).Error
	return count, err
}