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

// SiteIconRepositoryImpl 站点图标仓储实现
type SiteIconRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewSiteIconRepository 创建站点图标仓储实例
func NewSiteIconRepository(db *gorm.DB) interfaces.SiteIconRepository {
	return &SiteIconRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建站点图标
func (r *SiteIconRepositoryImpl) Create(ctx context.Context, icon *database.SiteIcon) error {
	err := r.db.WithContext(ctx).Create(icon).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_icon")
	}
	return err
}

// GetByID 根据ID获取站点图标
func (r *SiteIconRepositoryImpl) GetByID(ctx context.Context, id string) (*database.SiteIcon, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site_icon:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_icon")
		if err == nil && hit {
			if icon, ok := cachedValue.(*database.SiteIcon); ok {
				return icon, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var icon database.SiteIcon
	err := r.db.WithContext(ctx).First(&icon, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &icon, 3600*time.Second, "site_icon")
	}

	return &icon, nil
}

// GetBySiteName 根据站点名称获取站点图标
func (r *SiteIconRepositoryImpl) GetBySiteName(ctx context.Context, siteName string) (*database.SiteIcon, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site_icon:site_name:%s", siteName)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_icon")
		if err == nil && hit {
			if icon, ok := cachedValue.(*database.SiteIcon); ok {
				return icon, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var icon database.SiteIcon
	err := r.db.WithContext(ctx).Where("site_name = ?", siteName).First(&icon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &icon, 3600*time.Second, "site_icon")
	}

	return &icon, nil
}

// GetByDomain 根据域名获取站点图标
func (r *SiteIconRepositoryImpl) GetByDomain(ctx context.Context, domain string) (*database.SiteIcon, error) {
	// 调用GetBySiteName，因为数据库中使用的是site_name字段
	return r.GetBySiteName(ctx, domain)
}

// Update 更新站点图标
func (r *SiteIconRepositoryImpl) Update(ctx context.Context, icon *database.SiteIcon) error {
	err := r.db.WithContext(ctx).Save(icon).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_icon")
	}
	return err
}

// Delete 删除站点图标
func (r *SiteIconRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.SiteIcon{}, "id = ?", id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_icon")
	}
	return err
}

// List 获取站点图标列表
func (r *SiteIconRepositoryImpl) List(ctx context.Context, params interfaces.ListSiteIconParams) ([]*database.SiteIcon, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("site_icon:list:page:%d:page_size:%d:domain:%s",
		params.Page, params.PageSize, params.Domain)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_icon")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Icons []*database.SiteIcon
				Total int64
			}); ok {
				return cacheData.Icons, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var icons []*database.SiteIcon
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SiteIcon{})

	// 添加过滤条件
	if params.Domain != "" {
		query = query.Where("domain LIKE ?", "%"+params.Domain+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&icons).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		cacheData := struct {
			Icons []*database.SiteIcon
			Total int64
		}{
			Icons: icons,
			Total: total,
		}
		r.cache.Set(cacheKey, cacheData, 3600*time.Second, "site_icon")
	}

	return icons, total, err
}
