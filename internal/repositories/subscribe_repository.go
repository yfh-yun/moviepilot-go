package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/cache"
)

// subscribeRepository 订阅仓储实现
type subscribeRepository struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewSubscribeRepository 创建订阅仓储
func NewSubscribeRepository(db *gorm.DB) interfaces.SubscribeRepository {
	return &subscribeRepository{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建订阅
func (r *subscribeRepository) Create(ctx context.Context, subscribe *database.Subscribe) error {
	err := r.db.WithContext(ctx).Create(subscribe).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe")
	}
	return err
}

// Update 更新订阅
func (r *subscribeRepository) Update(ctx context.Context, subscribe *database.Subscribe) error {
	err := r.db.WithContext(ctx).Save(subscribe).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe")
	}
	return err
}

// Delete 删除订阅
func (r *subscribeRepository) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.Subscribe{}, id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("subscribe")
	}
	return err
}

// GetByID 根据 ID 查找订阅
func (r *subscribeRepository) GetByID(ctx context.Context, id string) (*database.Subscribe, error) {
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

	var subscribe database.Subscribe
	err := r.db.WithContext(ctx).First(&subscribe, id).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &subscribe, 3600*time.Second, "subscribe")
	}

	return &subscribe, nil
}

// List 获取订阅列表
func (r *subscribeRepository) List(ctx context.Context, params interfaces.ListSubscribeParams) ([]*database.Subscribe, int64, error) {
	// 生成缓存键
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

	// 应用过滤条件
	if params.Status != "" {
		query = query.Where("state = ?", params.Status)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.UserID != "" {
		query = query.Where("username = ?", params.UserID)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	query = query.Order("created_at DESC")

	// 应用分页
	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Limit(params.PageSize).Offset(offset)
	}

	err := query.Find(&subscribes).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		cacheData := struct {
			Subscribes []*database.Subscribe
			Total      int64
		}{
			Subscribes: subscribes,
			Total:      total,
		}
		r.cache.Set(cacheKey, cacheData, 3600*time.Second, "subscribe")
	}

	return subscribes, total, err
}

// GetActiveSubscriptions 获取活跃订阅
func (r *subscribeRepository) GetActiveSubscriptions(ctx context.Context) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := "subscribe:active"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).Where("state IN ?", []string{"N", "R"}).Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, err
}

// GetByUserID 根据用户ID获取订阅列表
func (r *subscribeRepository) GetByUserID(ctx context.Context, userID string) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:user:%s", userID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).Where("username = ?", userID).Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, err
}

// Exists 检查订阅是否存在
func (r *subscribeRepository) Exists(ctx context.Context, tmdbID *int, doubanID *string, season *int) (bool, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:exists:tmdb:%v:douban:%v:season:%v", tmdbID, doubanID, season)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if exists, ok := cachedValue.(bool); ok {
				return exists, nil
			}
		}
	}

	query := r.db.WithContext(ctx).Model(&database.Subscribe{})

	// 构建查询条件
	if tmdbID != nil {
		query = query.Where("tmdb_id = ?", *tmdbID)
	}
	if doubanID != nil {
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

	exists := count > 0

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, exists, 3600*time.Second, "subscribe")
	}

	return exists, err
}

// GetByTMDBID 根据TMDB ID获取订阅
func (r *subscribeRepository) GetByTMDBID(ctx context.Context, tmdbID int, season *int) (*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:tmdb:%d:season:%v", tmdbID, season)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribe, ok := cachedValue.(*database.Subscribe); ok {
				return subscribe, nil
			}
		}
	}

	var subscribe database.Subscribe
	query := r.db.WithContext(ctx).Where("tmdb_id = ?", tmdbID)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.First(&subscribe).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &subscribe, 3600*time.Second, "subscribe")
	}

	return &subscribe, nil
}

// GetByDoubanID 根据豆瓣ID获取订阅
func (r *subscribeRepository) GetByDoubanID(ctx context.Context, doubanID string, season *int) (*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:douban:%s:season:%v", doubanID, season)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribe, ok := cachedValue.(*database.Subscribe); ok {
				return subscribe, nil
			}
		}
	}

	var subscribe database.Subscribe
	query := r.db.WithContext(ctx).Where("douban_id = ?", doubanID)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.First(&subscribe).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &subscribe, 3600*time.Second, "subscribe")
	}

	return &subscribe, nil
}

// ListByState 根据状态查询订阅
func (r *subscribeRepository) ListByState(ctx context.Context, state string) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:state:%s", state)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("state = ?", state).
		Order("created_at DESC").
		Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, err
}

// ListActive 查询活跃订阅（状态为R的订阅）
func (r *subscribeRepository) ListActive(ctx context.Context) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := "subscribe:active:state:R"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("state = ?", "R").
		Order("created_at DESC").
		Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, err
}

// Statistics 统计订阅信息
func (r *subscribeRepository) Statistics(ctx context.Context) (map[string]int64, error) {
	// 生成缓存键
	cacheKey := "subscribe:statistics"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if stats, ok := cachedValue.(map[string]int64); ok {
				return stats, nil
			}
		}
	}

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
	if err := r.db.WithContext(ctx).
		Model(&database.Subscribe{}).
		Select("state, count(*) as count").
		Group("state").
		Scan(&stateStats).Error; err != nil {
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
	if err := r.db.WithContext(ctx).
		Model(&database.Subscribe{}).
		Select("type, count(*) as count").
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, err
	}
	for _, stat := range typeStats {
		stats["type_"+stat.Type] = stat.Count
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, stats, 3600*time.Second, "subscribe")
	}

	return stats, nil
}

// ListByType 根据类型查询订阅
func (r *subscribeRepository) ListByType(ctx context.Context, mediaType string) ([]*database.Subscribe, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("subscribe:type:%s", mediaType)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "subscribe")
		if err == nil && hit {
			if subscribes, ok := cachedValue.([]*database.Subscribe); ok {
				return subscribes, nil
			}
		}
	}

	var subscribes []*database.Subscribe
	err := r.db.WithContext(ctx).
		Where("type = ?", mediaType).
		Order("created_at DESC").
		Find(&subscribes).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, subscribes, 3600*time.Second, "subscribe")
	}

	return subscribes, err
}

// UpdateState 更新订阅状态
func (r *subscribeRepository) UpdateState(ctx context.Context, id uint, state string) error {
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
