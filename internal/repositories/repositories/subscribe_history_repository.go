package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/cache"
)

// SubscribeHistoryRepositoryImpl 订阅历史仓储实现
type SubscribeHistoryRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewSubscribeHistoryRepository 创建订阅历史仓储实例
func NewSubscribeHistoryRepository(db *gorm.DB) interfaces.SubscribeHistoryRepository {
	return &SubscribeHistoryRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建订阅历史
func (r *SubscribeHistoryRepositoryImpl) Create(ctx context.Context, history *database.SubscribeHistory) error {
	err := r.db.WithContext(ctx).Create(history).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe_history")
	}
	return err
}

// GetByID 根据ID获取订阅历史
func (r *SubscribeHistoryRepositoryImpl) GetByID(ctx context.Context, id string) (*database.SubscribeHistory, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe_history:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe_history")
		if err == nil && hit {
			if history, ok := cachedValue.(*database.SubscribeHistory); ok {
				return history, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var history database.SubscribeHistory
	err := r.db.WithContext(ctx).First(&history, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &history, 3600*time.Second, "subscribe_history")
	}

	return &history, nil
}

// GetBySubscribeID 根据订阅ID获取历史
func (r *SubscribeHistoryRepositoryImpl) GetBySubscribeID(ctx context.Context, subscribeID string, params interfaces.ListSubscribeHistoryParams) ([]*database.SubscribeHistory, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("subscribe_history:subscribe_id:%s:page:%d:page_size:%d:status:%s:date_from:%s:date_to:%s",
		subscribeID, params.Page, params.PageSize, params.Status, params.DateFrom, params.DateTo)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe_history")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Histories []*database.SubscribeHistory
				Total     int64
			}); ok {
				return cacheData.Histories, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var histories []*database.SubscribeHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SubscribeHistory{}).Where("subscribe_id = ?", subscribeID)

	// 添加过滤条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&histories).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		cacheData := struct {
			Histories []*database.SubscribeHistory
			Total     int64
		}{
			Histories: histories,
			Total:     total,
		}
		r.cache.Set(cacheKey, cacheData, 3600*time.Second, "subscribe_history")
	}

	return histories, total, err
}

// Update 更新订阅历史
func (r *SubscribeHistoryRepositoryImpl) Update(ctx context.Context, history *database.SubscribeHistory) error {
	err := r.db.WithContext(ctx).Save(history).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe_history")
	}
	return err
}

// Delete 删除订阅历史
func (r *SubscribeHistoryRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.SubscribeHistory{}, "id = ?", id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe_history")
	}
	return err
}

// List 获取订阅历史列表
func (r *SubscribeHistoryRepositoryImpl) List(ctx context.Context, params interfaces.ListSubscribeHistoryParams) ([]*database.SubscribeHistory, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("subscribe_history:list:page:%d:page_size:%d:subscribe_id:%s:status:%s:date_from:%s:date_to:%s",
		params.Page, params.PageSize, params.SubscribeID, params.Status, params.DateFrom, params.DateTo)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe_history")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Histories []*database.SubscribeHistory
				Total     int64
			}); ok {
				return cacheData.Histories, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var histories []*database.SubscribeHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SubscribeHistory{})

	// 添加过滤条件
	if params.SubscribeID != "" {
		query = query.Where("subscribe_id = ?", params.SubscribeID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&histories).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		cacheData := struct {
			Histories []*database.SubscribeHistory
			Total     int64
		}{
			Histories: histories,
			Total:     total,
		}
		r.cache.Set(cacheKey, cacheData, 3600*time.Second, "subscribe_history")
	}

	return histories, total, err
}

// GetLatestBySubscribeID 获取订阅的最新历史
func (r *SubscribeHistoryRepositoryImpl) GetLatestBySubscribeID(ctx context.Context, subscribeID string) (*database.SubscribeHistory, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe_history:latest:%s", subscribeID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe_history")
		if err == nil && hit {
			if history, ok := cachedValue.(*database.SubscribeHistory); ok {
				return history, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var history database.SubscribeHistory
	err := r.db.WithContext(ctx).Where("subscribe_id = ?", subscribeID).Order("created_at DESC").First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &history, 3600*time.Second, "subscribe_history")
	}

	return &history, nil
}
