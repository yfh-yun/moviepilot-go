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

// PluginDataRepositoryImpl 插件数据仓储实现
type PluginDataRepositoryImpl struct {
	db    *gorm.DB
	cache cache.CacheBackend
}

// NewPluginDataRepository 创建插件数据仓储实例
func NewPluginDataRepository(db *gorm.DB) interfaces.PluginDataRepository {
	// 初始化缓存，使用TTL缓存，1000个条目，3600秒过期时间
	cacheBackend := cache.Cache("ttl", 1000, 3600)
	return &PluginDataRepositoryImpl{
		db:    db,
		cache: cacheBackend,
	}
}

// Get 根据插件ID和数据key获取数据
func (r *PluginDataRepositoryImpl) Get(ctx context.Context, pluginID string, key string) (*database.PluginData, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("plugin_data:get:%s:%s", pluginID, key)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "plugin_data")
		if err == nil && hit {
			if data, ok := cachedValue.(*database.PluginData); ok {
				return data, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var data database.PluginData
	err := r.db.WithContext(ctx).
		Where("plugin_key = ? AND data_key = ?", pluginID, key).
		First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 存储到缓存
	if r.cache != nil {
		_ = r.cache.Set(cacheKey, &data, 0, "plugin_data")
	}

	return &data, nil
}

// Set 设置插件数据
func (r *PluginDataRepositoryImpl) Set(ctx context.Context, pluginID string, key string, value string) error {
	// 先查询是否存在
	var data database.PluginData
	err := r.db.WithContext(ctx).
		Where("plugin_key = ? AND data_key = ?", pluginID, key).
		First(&data).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建新记录
			data = database.PluginData{
				PluginKey: pluginID,
				DataKey:   key,
				DataValue: value,
			}
			err = r.db.WithContext(ctx).Create(&data).Error
		} else {
			return err
		}
	} else {
		// 存在，更新记录
		data.DataValue = value
		err = r.db.WithContext(ctx).Save(&data).Error
	}

	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除该插件的所有数据缓存
		r.cache.Clear("plugin_data")
	}

	return nil
}

// Delete 删除插件数据
func (r *PluginDataRepositoryImpl) Delete(ctx context.Context, pluginID string, key string) error {
	err := r.db.WithContext(ctx).
		Where("plugin_key = ? AND data_key = ?", pluginID, key).
		Delete(&database.PluginData{}).Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除该插件的所有数据缓存
		r.cache.Clear("plugin_data")
	}

	return nil
}

// ListByPlugin 获取插件的所有数据
func (r *PluginDataRepositoryImpl) ListByPlugin(ctx context.Context, pluginID string) ([]*database.PluginData, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("plugin_data:list:%s", pluginID)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "plugin_data")
		if err == nil && hit {
			if dataList, ok := cachedValue.([]*database.PluginData); ok {
				return dataList, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var dataList []*database.PluginData
	err := r.db.WithContext(ctx).
		Where("plugin_key = ?", pluginID).
		Find(&dataList).Error
	if err != nil {
		return nil, err
	}

	// 存储到缓存
	if r.cache != nil {
		_ = r.cache.Set(cacheKey, dataList, 0, "plugin_data")
	}

	return dataList, nil
}

// DeleteByPlugin 删除插件的所有数据
func (r *PluginDataRepositoryImpl) DeleteByPlugin(ctx context.Context, pluginID string) error {
	err := r.db.WithContext(ctx).
		Where("plugin_key = ?", pluginID).
		Delete(&database.PluginData{}).Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除该插件的所有数据缓存
		r.cache.Clear("plugin_data")
	}

	return nil
}

// BatchSet 批量设置插件数据
func (r *PluginDataRepositoryImpl) BatchSet(ctx context.Context, pluginID string, data map[string]string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range data {
			var pluginData database.PluginData
			err := tx.Where("plugin_key = ? AND data_key = ?", pluginID, key).First(&pluginData).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 不存在，创建新记录
					pluginData = database.PluginData{
						PluginKey: pluginID,
						DataKey:   key,
						DataValue: value,
					}
					if err := tx.Create(&pluginData).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}

			// 存在，更新记录
			pluginData.DataValue = value
			if err := tx.Save(&pluginData).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 清除相关缓存
	if r.cache != nil {
		// 清除该插件的所有数据缓存
		r.cache.Clear("plugin_data")
	}

	return nil
}

// BatchGet 批量获取插件数据
func (r *PluginDataRepositoryImpl) BatchGet(ctx context.Context, pluginID string, keys []string) (map[string]string, error) {
	// 生成缓存键，对keys进行排序以确保缓存键的唯一性
	cacheKey := fmt.Sprintf("plugin_data:batch_get:%s:%v", pluginID, keys)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "plugin_data")
		if err == nil && hit {
			if result, ok := cachedValue.(map[string]string); ok {
				return result, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var dataList []*database.PluginData
	err := r.db.WithContext(ctx).
		Where("plugin_key = ? AND data_key IN ?", pluginID, keys).
		Find(&dataList).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(dataList))
	for _, data := range dataList {
		result[data.DataKey] = data.DataValue
	}

	// 存储到缓存
	if r.cache != nil {
		_ = r.cache.Set(cacheKey, result, 0, "plugin_data")
	}

	return result, nil
}

// Exists 检查插件数据是否存在
func (r *PluginDataRepositoryImpl) Exists(ctx context.Context, pluginID string, key string) (bool, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("plugin_data:exists:%s:%s", pluginID, key)

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "plugin_data")
		if err == nil && hit {
			if exists, ok := cachedValue.(bool); ok {
				return exists, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.PluginData{}).
		Where("plugin_key = ? AND data_key = ?", pluginID, key).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	exists := count > 0

	// 存储到缓存
	if r.cache != nil {
		_ = r.cache.Set(cacheKey, exists, 0, "plugin_data")
	}

	return exists, nil
}

// ListAllPlugins 获取所有插件ID列表
func (r *PluginDataRepositoryImpl) ListAllPlugins(ctx context.Context) ([]string, error) {
	// 生成缓存键
	cacheKey := "plugin_data:list_all_plugins"

	// 检查缓存
	if r.cache != nil {
		cachedValue, hit, err := r.cache.Get(cacheKey, "plugin_data")
		if err == nil && hit {
			if plugins, ok := cachedValue.([]string); ok {
				return plugins, nil
			}
		}
	}

	// 缓存未命中，查询数据库
	var plugins []string
	err := r.db.WithContext(ctx).
		Model(&database.PluginData{}).
		Distinct("plugin_key").
		Pluck("plugin_key", &plugins).Error
	if err != nil {
		return nil, err
	}

	// 存储到缓存
	if r.cache != nil {
		_ = r.cache.Set(cacheKey, plugins, 0, "plugin_data")
	}

	return plugins, nil
}
