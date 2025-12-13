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

// SubscribeRepositoryImpl 订阅仓储实现
type SubscribeRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewSubscribeRepository 创建订阅仓储实例
func NewSubscribeRepository(db *gorm.DB) interfaces.SubscribeRepository {
	return &SubscribeRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建订阅
func (r *SubscribeRepositoryImpl) Create(ctx context.Context, subscribe *database.Subscribe) error {
	err := r.db.WithContext(ctx).Create(subscribe).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe")
	}
	return err
}

// GetByID 根据ID获取订阅
func (r *SubscribeRepositoryImpl) GetByID(ctx context.Context, id string) (*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribe, ok := cachedValue.(*database.Subscribe); ok {
				return subscribe, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var subscribe database.Subscribe
	err := r.db.WithContext(ctx).First(&subscribe, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &subscribe, 3600*time.Second, "subscribe")
	}

	return &subscribe, nil
}

// Update 更新订阅
func (r *SubscribeRepositoryImpl) Update(ctx context.Context, subscribe *database.Subscribe) error {
	err := r.db.WithContext(ctx).Save(subscribe).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe")
	}
	return err
}

// Delete 删除订阅
func (r *SubscribeRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.Subscribe{}, "id = ?", id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe")
	}
	return err
}

// List 获取订阅列表
func (r *SubscribeRepositoryImpl) List(ctx context.Context, params interfaces.ListSubscribeParams) ([]*database.Subscribe, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("subscribe:list:page:%d:page_size:%d:status:%s:type:%s:user_id:%s",
		params.Page, params.PageSize, params.Status, params.Type, params.UserID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Subscribes []*database.Subscribe
				Total      int64
			}); ok {
				return cacheData.Subscribes, cacheData.Total, nil
			}
		}
	}

	var subscribes []*database.Subscribe
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Subscribe{})

	// 添加过滤条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
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
	err := query.Offset(offset).Limit(params.PageSize).Find(&subscribes).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, struct {
			Subscribes []*database.Subscribe
			Total      int64
		}{Subscribes: subscribes, Total: total}, 3600*time.Second, "subscribe")
	}

	return subscribes, total, err
}

// GetByUserID 根据用户ID获取订阅列表
func (r *SubscribeRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:list:user:%s", userID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, nil
}

// GetActiveSubscriptions 获取活跃订阅
func (r *SubscribeRepositoryImpl) GetActiveSubscriptions(ctx context.Context) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := "subscribe:list:active_subscriptions"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, nil
}

// Exists 检查订阅是否存在
func (r *SubscribeRepositoryImpl) Exists(ctx context.Context, tmdbID *int, doubanID *string, season *int) (bool, error) {
	// 生成缓存键，根据不同的参数组合生成不同的缓存键
	var cacheKey string
	if tmdbID != nil {
		if season != nil {
			cacheKey = fmt.Sprintf("subscribe:exists:tmdb:%d:season:%d", *tmdbID, *season)
		} else {
			cacheKey = fmt.Sprintf("subscribe:exists:tmdb:%d", *tmdbID)
		}
	} else if doubanID != nil && *doubanID != "" {
		if season != nil {
			cacheKey = fmt.Sprintf("subscribe:exists:douban:%s:season:%d", *doubanID, *season)
		} else {
			cacheKey = fmt.Sprintf("subscribe:exists:douban:%s", *doubanID)
		}
	} else {
		// 如果没有提供任何ID，则无法缓存，直接查询数据库
		query := r.db.WithContext(ctx).Model(&database.Subscribe{})

		// 构建查询条件
		if tmdbID != nil {
			query = query.Where("tmdb_id = ?", *tmdbID)
		}
		if doubanID != nil && *doubanID != "" {
			query = query.Where("douban_id = ?", *doubanID)
		}
		if season != nil {
			query = query.Where("season = ?", *season)
		}

		var count int64
		err := query.Count(&count).Error
		return count > 0, err
	}

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if exists, ok := cachedValue.(bool); ok {
				return exists, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	query := r.db.WithContext(ctx).Model(&database.Subscribe{})

	// 构建查询条件
	if tmdbID != nil {
		query = query.Where("tmdb_id = ?", *tmdbID)
	}
	if doubanID != nil && *doubanID != "" {
		query = query.Where("douban_id = ?", *doubanID)
	}
	if season != nil {
		query = query.Where("season = ?", *season)
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, count > 0, 3600*time.Second, "subscribe")
	}

	return count > 0, nil
}

// GetByTMDBID 根据TMDB ID获取订阅
func (r *SubscribeRepositoryImpl) GetByTMDBID(ctx context.Context, tmdbID int, season *int) (*database.Subscribe, error) {
	// 生成缓存键，包含season参数
	var cacheKey string
	if season != nil {
		cacheKey = fmt.Sprintf("subscribe:tmdb:%d:season:%d", tmdbID, *season)
	} else {
		cacheKey = fmt.Sprintf("subscribe:tmdb:%d", tmdbID)
	}

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribe, ok := cachedValue.(*database.Subscribe); ok {
				return subscribe, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var subscribe database.Subscribe
	query := r.db.WithContext(ctx).Where("tmdb_id = ?", tmdbID)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.First(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &subscribe, 3600*time.Second, "subscribe")
	}

	return &subscribe, nil
}

// GetByDoubanID 根据豆瓣ID获取订阅
func (r *SubscribeRepositoryImpl) GetByDoubanID(ctx context.Context, doubanID string, season *int) (*database.Subscribe, error) {
	// 生成缓存键，包含season参数
	var cacheKey string
	if season != nil {
		cacheKey = fmt.Sprintf("subscribe:douban:%s:season:%d", doubanID, *season)
	} else {
		cacheKey = fmt.Sprintf("subscribe:douban:%s", doubanID)
	}

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribe, ok := cachedValue.(*database.Subscribe); ok {
				return subscribe, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var subscribe database.Subscribe
	query := r.db.WithContext(ctx).Where("douban_id = ?", doubanID)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.First(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &subscribe, 3600*time.Second, "subscribe")
	}

	return &subscribe, nil
}

// ListByState 根据状态查询订阅
func (r *SubscribeRepositoryImpl) ListByState(ctx context.Context, state string) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:list:state:%s", state)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("state = ?", state).
		Order("last_update DESC").
		Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, nil
}

// ListActive 查询活跃订阅（状态为R的订阅）
func (r *SubscribeRepositoryImpl) ListActive(ctx context.Context) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := "subscribe:list:active"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("state = ?", "R").
		Order("last_update DESC").
		Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, nil
}

// Statistics 统计订阅信息
func (r *SubscribeRepositoryImpl) Statistics(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 统计总数
	var total int64
	if err := r.db.WithContext(ctx).Model(&database.Subscribe{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 按状态统计
	var stateStats []struct {
		State string
		Count int64
	}
	err := r.db.WithContext(ctx).
		Model(&database.Subscribe{}).
		Select("state, COUNT(*) as count").
		Group("state").
		Scan(&stateStats).Error
	if err != nil {
		return nil, err
	}

	for _, stat := range stateStats {
		stats["state_"+stat.State] = stat.Count
	}

	// 按类型统计
	var typeStats []struct {
		Type  string
		Count int64
	}
	err = r.db.WithContext(ctx).
		Model(&database.Subscribe{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats).Error
	if err != nil {
		return nil, err
	}

	for _, stat := range typeStats {
		stats["type_"+stat.Type] = stat.Count
	}

	return stats, nil
}

// ListByType 根据类型查询订阅
func (r *SubscribeRepositoryImpl) ListByType(ctx context.Context, mediaType string) ([]*database.Subscribe, error) {
	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("type = ?", mediaType).
		Order("last_update DESC").
		Find(&subscribes).Error
	return subscribes, err
}

// UpdateState 更新订阅状态
func (r *SubscribeRepositoryImpl) UpdateState(ctx context.Context, id uint, state string) error {
	err := r.db.WithContext(ctx).
		Model(&database.Subscribe{}).
		Where("id = ?", id).
		Update("state", state).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe")
	}
	return err
}
