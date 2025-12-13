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

// SiteRepositoryImpl 站点仓储实现
type SiteRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewSiteRepository 创建站点仓储实例
func NewSiteRepository(db *gorm.DB) interfaces.SiteRepository {
	return &SiteRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建站点
func (r *SiteRepositoryImpl) Create(ctx context.Context, site *database.Site) error {
	err := r.db.WithContext(ctx).Create(site).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site")
	}
	return err
}

// Update 更新站点
func (r *SiteRepositoryImpl) Update(ctx context.Context, site *database.Site) error {
	err := r.db.WithContext(ctx).Save(site).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site")
	}
	return err
}

// Delete 删除站点
func (r *SiteRepositoryImpl) Delete(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&database.Site{}, id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site")
	}
	return err
}

// GetByID 根据ID获取站点
func (r *SiteRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.Site, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site:id:%d", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site")
		if err == nil && hit {
			if site, ok := cachedValue.(*database.Site); ok {
				return site, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var site database.Site
	err := r.db.WithContext(ctx).First(&site, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &site, 3600*time.Second, "site")
	}

	return &site, nil
}

// GetByDomain 根据域名获取站点
func (r *SiteRepositoryImpl) GetByDomain(ctx context.Context, domain string) (*database.Site, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site:domain:%s", domain)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site")
		if err == nil && hit {
			if site, ok := cachedValue.(*database.Site); ok {
				return site, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var site database.Site
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&site).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &site, 3600*time.Second, "site")
	}

	return &site, nil
}

// List 获取站点列表
func (r *SiteRepositoryImpl) List(ctx context.Context, opts interfaces.ListOptions) ([]*database.Site, int64, error) {
	// 生成缓存键，包含分页和排序参数
	cacheKey := fmt.Sprintf("site:list:offset:%d:limit:%d:order:%s:%s",
		opts.GetOffset(), opts.GetLimit(), opts.OrderBy, opts.Order)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Sites []*database.Site
				Total int64
			}); ok {
				return cacheData.Sites, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var sites []*database.Site
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Site{})

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	offset := opts.GetOffset()
	limit := opts.GetLimit()
	query = query.Offset(offset).Limit(limit)

	// 应用排序
	orderBy := "pri DESC, created_at DESC"
	if opts.OrderBy != "" {
		order := opts.Order
		if order == "" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", opts.OrderBy, order)
	}
	query = query.Order(orderBy)

	// 执行查询
	err := query.Find(&sites).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		cacheData := struct {
			Sites []*database.Site
			Total int64
		}{
			Sites: sites,
			Total: total,
		}
		r.cache.Set(cacheKey, cacheData, 3600*time.Second, "site")
	}

	return sites, total, err
}

// ListActive 获取启用的站点列表
func (r *SiteRepositoryImpl) ListActive(ctx context.Context) ([]*database.Site, error) {
	// 生成缓存键
	cacheKey := "site:active_list"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site")
		if err == nil && hit {
			if sites, ok := cachedValue.([]*database.Site); ok {
				return sites, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var sites []*database.Site
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("pri DESC").Find(&sites).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, sites, 3600*time.Second, "site")
	}

	return sites, nil
}

// UpdateStatus 更新站点状态
func (r *SiteRepositoryImpl) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	err := r.db.WithContext(ctx).Model(&database.Site{}).Where("id = ?", id).Update("is_active", isActive).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site")
	}
	return err
}

// IncrementFailCount 增加失败次数
func (r *SiteRepositoryImpl) IncrementFailCount(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Model(&database.Site{}).Where("id = ?", id).UpdateColumn("fail_count", gorm.Expr("fail_count + ?", 1)).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site")
	}
	return err
}

// ResetFailCount 重置失败次数
func (r *SiteRepositoryImpl) ResetFailCount(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Model(&database.Site{}).Where("id = ?", id).UpdateColumn("fail_count", 0).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site")
	}
	return err
}

// UpdateStatistics 更新站点统计
func (r *SiteRepositoryImpl) UpdateStatistics(ctx context.Context, siteName string, success bool, seconds int) error {
	// TODO: 实现站点统计更新逻辑
	// 1. 查询站点
	// 2. 更新统计信息
	// 3. 保存到数据库
	return nil
}
