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

// SiteUserDataRepositoryImpl 站点用户数据仓储实现
type SiteUserDataRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewSiteUserDataRepository 创建站点用户数据仓储实例
func NewSiteUserDataRepository(db *gorm.DB) interfaces.SiteUserDataRepository {
	return &SiteUserDataRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建站点用户数据
func (r *SiteUserDataRepositoryImpl) Create(ctx context.Context, userData *database.SiteUserData) error {
	err := r.db.WithContext(ctx).Create(userData).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_user_data")
	}
	return err
}

// GetByID 根据ID获取站点用户数据
func (r *SiteUserDataRepositoryImpl) GetByID(ctx context.Context, id string) (*database.SiteUserData, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site_user_data:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_user_data")
		if err == nil && hit {
			if userData, ok := cachedValue.(*database.SiteUserData); ok {
				return userData, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var userData database.SiteUserData
	err := r.db.WithContext(ctx).First(&userData, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &userData, 3600*time.Second, "site_user_data")
	}

	return &userData, nil
}

// GetByUserID 根据用户ID获取站点用户数据
func (r *SiteUserDataRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*database.SiteUserData, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site_user_data:user_id:%s", userID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_user_data")
		if err == nil && hit {
			if userDataList, ok := cachedValue.([]*database.SiteUserData); ok {
				return userDataList, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var userDataList []*database.SiteUserData
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userDataList).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, userDataList, 3600*time.Second, "site_user_data")
	}

	return userDataList, nil
}

// GetBySiteID 根据站点ID获取用户数据
func (r *SiteUserDataRepositoryImpl) GetBySiteID(ctx context.Context, siteID string) ([]*database.SiteUserData, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site_user_data:site_id:%s", siteID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_user_data")
		if err == nil && hit {
			if userDataList, ok := cachedValue.([]*database.SiteUserData); ok {
				return userDataList, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var userDataList []*database.SiteUserData
	err := r.db.WithContext(ctx).Where("site_id = ?", siteID).Find(&userDataList).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, userDataList, 3600*time.Second, "site_user_data")
	}

	return userDataList, nil
}

// GetByUserAndSite 根据用户ID和站点ID获取数据
func (r *SiteUserDataRepositoryImpl) GetByUserAndSite(ctx context.Context, userID, siteID string) (*database.SiteUserData, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("site_user_data:user_site:%s:%s", userID, siteID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_user_data")
		if err == nil && hit {
			if userData, ok := cachedValue.(*database.SiteUserData); ok {
				return userData, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var userData database.SiteUserData
	err := r.db.WithContext(ctx).Where("user_id = ? AND site_id = ?", userID, siteID).First(&userData).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &userData, 3600*time.Second, "site_user_data")
	}

	return &userData, nil
}

// Update 更新站点用户数据
func (r *SiteUserDataRepositoryImpl) Update(ctx context.Context, userData *database.SiteUserData) error {
	err := r.db.WithContext(ctx).Save(userData).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_user_data")
	}
	return err
}

// Delete 删除站点用户数据
func (r *SiteUserDataRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.SiteUserData{}, "id = ?", id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("site_user_data")
	}
	return err
}

// List 获取站点用户数据列表
func (r *SiteUserDataRepositoryImpl) List(ctx context.Context, params interfaces.ListSiteUserDataParams) ([]*database.SiteUserData, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("site_user_data:list:page:%d:page_size:%d:user_id:%s:site_id:%s",
		params.Page, params.PageSize, params.UserID, params.SiteID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "site_user_data")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				UserDataList []*database.SiteUserData
				Total        int64
			}); ok {
				return cacheData.UserDataList, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var userDataList []*database.SiteUserData
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SiteUserData{})

	// 添加过滤条件
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.SiteID != "" {
		query = query.Where("site_id = ?", params.SiteID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&userDataList).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		cacheData := struct {
			UserDataList []*database.SiteUserData
			Total        int64
		}{
			UserDataList: userDataList,
			Total:        total,
		}
		r.cache.Set(cacheKey, cacheData, 3600*time.Second, "site_user_data")
	}

	return userDataList, total, err
}
