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

// SystemConfigRepositoryImpl 系统配置仓储实现
type SystemConfigRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewSystemConfigRepository 创建系统配置仓储实例
func NewSystemConfigRepository(db *gorm.DB) interfaces.SystemConfigRepository {
	return &SystemConfigRepositoryImpl{
		db:    db,
		cache: cache.Cache("ttl", 1000, 3600),
	}
}

// Get 根据key获取配置
func (r *SystemConfigRepositoryImpl) Get(ctx context.Context, key string) (*database.SystemConfig, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("system_config:%s", key)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "system_config")
		if err == nil && hit {
			if config, ok := cachedValue.(*database.SystemConfig); ok {
				return config, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var config database.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, &config, 3600*time.Second, "system_config")
	}

	return &config, nil
}

// Set 设置配置
func (r *SystemConfigRepositoryImpl) Set(ctx context.Context, key string, value string) error {
	// 先查询是否存在
	var config database.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建新记录
			config = database.SystemConfig{
				Key:   key,
				Value: value,
			}
			err = r.db.WithContext(ctx).Create(&config).Error
		} else {
			return err
		}
	} else {
		// 存在，更新记录
		config.Value = value
		err = r.db.WithContext(ctx).Save(&config).Error
	}

	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("system_config")
	}
	return err
}

// Delete 删除配置
func (r *SystemConfigRepositoryImpl) Delete(ctx context.Context, key string) error {
	err := r.db.WithContext(ctx).Where("key = ?", key).Delete(&database.SystemConfig{}).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("system_config")
	}
	return err
}

// List 获取所有配置
func (r *SystemConfigRepositoryImpl) List(ctx context.Context) ([]*database.SystemConfig, error) {
	// 生成缓存键
	cacheKey := "system_config:list"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "system_config")
		if err == nil && hit {
			if configs, ok := cachedValue.([]*database.SystemConfig); ok {
				return configs, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var configs []*database.SystemConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, configs, 3600*time.Second, "system_config")
	}

	return configs, nil
}

// BatchSet 批量设置配置
func (r *SystemConfigRepositoryImpl) BatchSet(ctx context.Context, configs map[string]string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range configs {
			var config database.SystemConfig
			err := tx.Where("key = ?", key).First(&config).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 不存在，创建新记录
					config = database.SystemConfig{
						Key:   key,
						Value: value,
					}
					if err := tx.Create(&config).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}

			// 存在，更新记录
			config.Value = value
			if err := tx.Save(&config).Error; err != nil {
				return err
			}
		}
		return nil
	})

	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("system_config")
	}
	return err
}

// BatchGet 批量获取配置
func (r *SystemConfigRepositoryImpl) BatchGet(ctx context.Context, keys []string) (map[string]string, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("system_config:batch:%v", keys)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "system_config")
		if err == nil && hit {
			if result, ok := cachedValue.(map[string]string); ok {
				return result, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var configs []*database.SystemConfig
	err := r.db.WithContext(ctx).Where("key IN ?", keys).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(configs))
	for _, config := range configs {
		result[config.Key] = config.Value
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, result, 3600*time.Second, "system_config")
	}

	return result, nil
}

// Exists 检查配置是否存在
func (r *SystemConfigRepositoryImpl) Exists(ctx context.Context, key string) (bool, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("system_config:exists:%s", key)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "system_config")
		if err == nil && hit {
			if exists, ok := cachedValue.(bool); ok {
				return exists, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var count int64
	err := r.db.WithContext(ctx).Model(&database.SystemConfig{}).Where("key = ?", key).Count(&count).Error
	if err != nil {
		return false, err
	}

	exists := count > 0

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, exists, 3600*time.Second, "system_config")
	}

	return exists, nil
}

// GetByPrefix 根据前缀获取配置
func (r *SystemConfigRepositoryImpl) GetByPrefix(ctx context.Context, prefix string) ([]*database.SystemConfig, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("system_config:prefix:%s", prefix)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "system_config")
		if err == nil && hit {
			if configs, ok := cachedValue.([]*database.SystemConfig); ok {
				return configs, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var configs []*database.SystemConfig
	err := r.db.WithContext(ctx).Where("key LIKE ?", prefix+"%").Find(&configs).Error
	if err != nil {
		return nil, err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, configs, 3600*time.Second, "system_config")
	}

	return configs, nil
}

// DeleteByPrefix 根据前缀删除配置
func (r *SystemConfigRepositoryImpl) DeleteByPrefix(ctx context.Context, prefix string) error {
	err := r.db.WithContext(ctx).Where("key LIKE ?", prefix+"%").Delete(&database.SystemConfig{}).Error
	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("system_config")
	}
	return err
}

// GetType 获取配置类型
func (r *SystemConfigRepositoryImpl) GetType(ctx context.Context, key string) (string, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("system_config:type:%s", key)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "system_config")
		if err == nil && hit {
			if configType, ok := cachedValue.(string); ok {
				return configType, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var config database.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}

	// 将结果存入缓存
	if r.cache != nil {
		r.cache.Set(cacheKey, config.Type, 3600*time.Second, "system_config")
	}

	return config.Type, nil
}

// SetWithType 设置配置和类型
func (r *SystemConfigRepositoryImpl) SetWithType(ctx context.Context, key string, value string, configType string) error {
	// 先查询是否存在
	var config database.SystemConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建新记录
			config = database.SystemConfig{
				Key:   key,
				Value: value,
				Type:  configType,
			}
			err = r.db.WithContext(ctx).Create(&config).Error
		} else {
			return err
		}
	} else {
		// 存在，更新记录
		config.Value = value
		config.Type = configType
		err = r.db.WithContext(ctx).Save(&config).Error
	}

	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("system_config")
	}
	return err
}

// BatchSetWithTypes 批量设置配置和类型
func (r *SystemConfigRepositoryImpl) BatchSetWithTypes(ctx context.Context, configs map[string]interfaces.ConfigItem) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, item := range configs {
			var config database.SystemConfig
			err := tx.Where("key = ?", key).First(&config).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 不存在，创建新记录
					config = database.SystemConfig{
						Key:   key,
						Value: item.Value,
						Type:  item.Type,
					}
					if err := tx.Create(&config).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}

			// 存在，更新记录
			config.Value = item.Value
			config.Type = item.Type
			if err := tx.Save(&config).Error; err != nil {
				return err
			}
		}
		return nil
	})

	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("system_config")
	}
	return err
}

// 批量删除配置
func (r *SystemConfigRepositoryImpl) BatchDelete(ctx context.Context, keys []string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, key := range keys {
			if err := tx.Where("key = ?", key).Delete(&database.SystemConfig{}).Error; err != nil {
				return err
			}
		}
		return nil
	})

	// 清除相关缓存
	if r.cache != nil {
		r.cache.Clear("system_config")
	}
	return err
}
