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

// UserConfigRepositoryImpl 用户配置仓储实现
type UserConfigRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewUserConfigRepository 创建用户配置仓储实例
func NewUserConfigRepository(db *gorm.DB) interfaces.UserConfigRepository {
	return &UserConfigRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Create 创建用户配置
func (r *UserConfigRepositoryImpl) Create(ctx context.Context, config *database.UserConfig) error {
	err := r.db.WithContext(ctx).Create(config).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("user_config")
	}
	return err
}

// GetByID 根据ID获取用户配置
func (r *UserConfigRepositoryImpl) GetByID(ctx context.Context, id string) (*database.UserConfig, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("user_config:id:%s", id)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user_config")
		if err == nil && hit {
			if config, ok := cachedValue.(*database.UserConfig); ok {
				return config, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var config database.UserConfig
	err := r.db.WithContext(ctx).First(&config, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &config, 3600*time.Second, "user_config")
	}

	return &config, nil
}

// GetByUserID 根据用户ID获取配置
func (r *UserConfigRepositoryImpl) GetByUserID(ctx context.Context, userID string) (*database.UserConfig, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("user_config:user_id:%s", userID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user_config")
		if err == nil && hit {
			if config, ok := cachedValue.(*database.UserConfig); ok {
				return config, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var config database.UserConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &config, 3600*time.Second, "user_config")
	}

	return &config, nil
}

// Update 更新用户配置
func (r *UserConfigRepositoryImpl) Update(ctx context.Context, config *database.UserConfig) error {
	err := r.db.WithContext(ctx).Save(config).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("user_config")
	}
	return err
}

// Delete 删除用户配置
func (r *UserConfigRepositoryImpl) Delete(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Delete(&database.UserConfig{}, "id = ?", id).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("user_config")
	}
	return err
}

// GetByKey 根据用户名和配置键获取配置
func (r *UserConfigRepositoryImpl) GetByKey(ctx context.Context, username string, key string) (*database.UserConfig, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("user_config:username:%s:key:%s", username, key)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user_config")
		if err == nil && hit {
			if config, ok := cachedValue.(*database.UserConfig); ok {
				return config, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var config database.UserConfig
	err := r.db.WithContext(ctx).Where("userid = ? AND key = ?", username, key).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &config, 3600*time.Second, "user_config")
	}

	return &config, nil
}

// DeleteByKey 根据用户名和配置键删除配置
func (r *UserConfigRepositoryImpl) DeleteByKey(ctx context.Context, username string, key string) error {
	// 获取配置
	userConfig, err := r.GetByKey(ctx, username, key)
	if err != nil {
		return err
	}

	// 如果配置存在，则删除
	if userConfig != nil {
		err = r.db.WithContext(ctx).Delete(userConfig).Error
		// 清除相关缓存
		if r.cache != nil {
			r.cache.Clear("user_config")
		}
	}

	return nil
}

// List 获取用户配置列表
func (r *UserConfigRepositoryImpl) List(ctx context.Context, params interfaces.ListUserConfigParams) ([]*database.UserConfig, int64, error) {
	// 生成缓存键，包含分页和过滤参数
	cacheKey := fmt.Sprintf("user_config:list:page:%d:page_size:%d:user_id:%s:key:%s",
		params.Page, params.PageSize, params.UserID, params.Key)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "user_config")
		if err == nil && hit {
			if cacheData, ok := cachedValue.(struct {
				Configs []*database.UserConfig
				Total   int64
			}); ok {
				return cacheData.Configs, cacheData.Total, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var configs []*database.UserConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&database.UserConfig{})

	// 添加过滤条件
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.Key != "" {
		query = query.Where("key = ?", params.Key)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&configs).Error
	if err != nil {
		return nil, 0, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, struct {
			Configs []*database.UserConfig
			Total   int64
		}{Configs: configs, Total: total}, 3600*time.Second, "user_config")
	}

	return configs, total, err
}
