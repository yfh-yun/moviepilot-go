package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/cache"
)

// MessageRepositoryImpl 消息仓储实现
type MessageRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewMessageRepository 创建消息仓储实例
func NewMessageRepository(db *gorm.DB) interfaces.MessageRepository {
	// 初始化缓存，使用TTL缓存，1000个条目，3600秒过期时间
	cacheBackend := cache.Cache("ttl", 1000, 3600)
	return &MessageRepositoryImpl{
		db:    db,
		cache: cacheBackend,
	}
}

// Create 创建消息
func (r *MessageRepositoryImpl) Create(ctx context.Context, message *database.Message) error {
	err := r.db.WithContext(ctx).Create(message).Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除所有消息缓存
		r.cache.Clear("message")
	}

	return nil
}

// GetByID 根据ID获取消息
func (r *MessageRepositoryImpl) GetByID(ctx context.Context, id string) (*database.Message, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("message:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "message")
		if err == nil && hit {
			if message, ok := cachedValue.(*database.Message); ok {
				return message, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var message database.Message
	err := r.db.WithContext(ctx).First(&message, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 存储到缓存
	if r.cache != nil {
		_ = r.cache.Set(cacheKey, &message, 0, "message")
	}

	return &message, nil
}

// Update 更新消息
func (r *MessageRepositoryImpl) Update(ctx context.Context, message *database.Message) error {
	err := r.db.WithContext(ctx).Save(message).Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除所有消息缓存
		r.cache.Clear("message")
	}

	return nil
}

// Delete 删除消息
func (r *MessageRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.Message{}, "id = ?", id).Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除所有消息缓存
		r.cache.Clear("message")
	}

	return nil
}

// List 获取消息列表
func (r *MessageRepositoryImpl) List(ctx context.Context, params interfaces.ListMessageParams) ([]*database.Message, int64, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("message:list:page:%d:page_size:%d:type:%s:status:%s:user_id:%s",
		params.Page, params.PageSize, params.Type, params.Status, params.UserID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "message")
		if err == nil && hit {
			if cachedResult, ok := cachedValue.(struct {
				Messages []*database.Message
				Total    int64
			}); ok {
				return cachedResult.Messages, cachedResult.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var messages []*database.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Message{})

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
	if err != nil {
		return nil, 0, err
	}

	// 存储到缓存
	if r.cache != nil {
		// 忽略缓存设置失败的错误
		_ = r.cache.Set(cacheKey, struct {
			Messages []*database.Message
			Total    int64
		}{
			Messages: messages,
			Total:    total,
		}, 0, "message")
	}

	return messages, total, err
}

// GetByUserID 根据用户ID获取消息列表
func (r *MessageRepositoryImpl) GetByUserID(ctx context.Context, userID string, params interfaces.ListMessageParams) ([]*database.Message, int64, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("message:user:%s:page:%d:page_size:%d:type:%s:status:%s",
		userID, params.Page, params.PageSize, params.Type, params.Status)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "message")
		if err == nil && hit {
			if cachedResult, ok := cachedValue.(struct {
				Messages []*database.Message
				Total    int64
			}); ok {
				return cachedResult.Messages, cachedResult.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var messages []*database.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Message{}).Where("user_id = ?", userID)

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
	if err != nil {
		return nil, 0, err
	}

	// 存储到缓存
	if r.cache != nil {
		// 忽略缓存设置失败的错误
		_ = r.cache.Set(cacheKey, struct {
			Messages []*database.Message
			Total    int64
		}{
			Messages: messages,
			Total:    total,
		}, 0, "message")
	}

	return messages, total, err
}

// MarkAsRead 标记为已读
func (r *MessageRepositoryImpl) MarkAsRead(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Model(&database.Message{}).Where("id = ?", id).Update("status", "read").Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除所有消息缓存
		r.cache.Clear("message")
	}

	return nil
}

// MarkAllAsRead 标记所有消息为已读
func (r *MessageRepositoryImpl) MarkAllAsRead(ctx context.Context, userID string) error {
	err := r.db.WithContext(ctx).Model(&database.Message{}).Where("user_id = ? AND status != ?", userID, "read").Update("status", "read").Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除所有消息缓存
		r.cache.Clear("message")
	}

	return nil
}

// GetUnreadCount 获取未读消息数量
func (r *MessageRepositoryImpl) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("message:unread_count:%s", userID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "message")
		if err == nil && hit {
			if count, ok := cachedValue.(int64); ok {
				return count, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var count int64
	err := r.db.WithContext(ctx).Model(&database.Message{}).Where("user_id = ? AND status = ?", userID, "unread").Count(&count).Error
	if err != nil {
		return 0, err
	}

	// 存储到缓存
	if r.cache != nil {
		// 忽略缓存设置失败的错误
		_ = r.cache.Set(cacheKey, count, 0, "message")
	}

	return count, nil
}
