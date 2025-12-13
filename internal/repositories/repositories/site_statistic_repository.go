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

// SiteStatisticRepositoryImpl 站点统计仓储实现
type SiteStatisticRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewSiteStatisticRepository 创建站点统计仓储实例
func NewSiteStatisticRepository(db *gorm.DB) interfaces.SiteStatisticRepository {
	return &SiteStatisticRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建站点统计
func (r *SiteStatisticRepositoryImpl) Create(ctx context.Context, statistic *database.SiteStatistic) error {
	err := r.db.WithContext(ctx).Create(statistic).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_statistic")
	}
	return err
}

// GetByID 根据ID获取站点统计
func (r *SiteStatisticRepositoryImpl) GetByID(ctx context.Context, id string) (*database.SiteStatistic, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site_statistic:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_statistic")
		if err == nil && hit {
			if statistic, ok := cachedValue.(*database.SiteStatistic); ok {
				return statistic, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var statistic database.SiteStatistic
	err := r.db.WithContext(ctx).First(&statistic, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &statistic, 3600*time.Second, "site_statistic")
	}

	return &statistic, nil
}

// GetBySiteID 根据站点ID获取统计
func (r *SiteStatisticRepositoryImpl) GetBySiteID(ctx context.Context, siteID string) (*database.SiteStatistic, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site_statistic:site_id:%s", siteID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_statistic")
		if err == nil && hit {
			if statistic, ok := cachedValue.(*database.SiteStatistic); ok {
				return statistic, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var statistic database.SiteStatistic
	err := r.db.WithContext(ctx).Where("site_id = ?", siteID).First(&statistic).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &statistic, 3600*time.Second, "site_statistic")
	}

	return &statistic, nil
}

// Update 更新站点统计
func (r *SiteStatisticRepositoryImpl) Update(ctx context.Context, statistic *database.SiteStatistic) error {
	err := r.db.WithContext(ctx).Save(statistic).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_statistic")
	}
	return err
}

// Delete 删除站点统计
func (r *SiteStatisticRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.SiteStatistic{}, "id = ?", id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_statistic")
	}
	return err
}

// List 获取站点统计列表
func (r *SiteStatisticRepositoryImpl) List(ctx context.Context, params interfaces.ListSiteStatisticParams) ([]*database.SiteStatistic, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("site_statistic:list:page:%d:page_size:%d:site_id:%s:date_from:%s:date_to:%s",
		params.Page, params.PageSize, params.SiteID, params.DateFrom, params.DateTo)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_statistic")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Statistics []*database.SiteStatistic
				Total      int64
			}); ok {
				return cacheData.Statistics, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var statistics []*database.SiteStatistic
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SiteStatistic{})

	// 添加过滤条件
	if params.SiteID != "" {
		query = query.Where("site_id = ?", params.SiteID)
	}
	if params.DateFrom != "" {
		query = query.Where("date >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("date <= ?", params.DateTo)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Order("date DESC").Find(&statistics).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		cacheData := struct {
			Statistics []*database.SiteStatistic
			Total      int64
		}{
			Statistics: statistics,
			Total:      total,
		}
		r.cache.Set(cacheKey, cacheData, 3600*time.Second, "site_statistic")
	}

	return statistics, total, err
}

// UpdateStatistics 更新统计数据
func (r *SiteStatisticRepositoryImpl) UpdateStatistics(ctx context.Context, siteID string, increment map[string]int64) error {
	updates := make(map[string]any)
	for key, value := range increment {
		updates[key] = gorm.Expr(key+" + ?", value)
	}

	err := r.db.WithContext(ctx).Model(&database.SiteStatistic{}).Where("site_id = ?", siteID).Updates(updates).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_statistic")
	}
	return err
}
