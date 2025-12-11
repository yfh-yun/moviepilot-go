package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

// PluginDataRepositoryImpl 插件数据仓储实现
type PluginDataRepositoryImpl struct {
	db *gorm.DB
}

// NewPluginDataRepository 创建插件数据仓储实例
func NewPluginDataRepository(db *gorm.DB) interfaces.PluginDataRepository {
	return &PluginDataRepositoryImpl{db: db}
}

// Get 根据插件ID和数据key获取数据
func (r *PluginDataRepositoryImpl) Get(ctx context.Context, pluginID string, key string) (*database.PluginData, error) {
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
			return r.db.WithContext(ctx).Create(&data).Error
		}
		return err
	}

	// 存在，更新记录
	data.DataValue = value
	return r.db.WithContext(ctx).Save(&data).Error
}

// Delete 删除插件数据
func (r *PluginDataRepositoryImpl) Delete(ctx context.Context, pluginID string, key string) error {
	return r.db.WithContext(ctx).
		Where("plugin_key = ? AND data_key = ?", pluginID, key).
		Delete(&database.PluginData{}).Error
}

// ListByPlugin 获取插件的所有数据
func (r *PluginDataRepositoryImpl) ListByPlugin(ctx context.Context, pluginID string) ([]*database.PluginData, error) {
	var dataList []*database.PluginData
	err := r.db.WithContext(ctx).
		Where("plugin_key = ?", pluginID).
		Find(&dataList).Error
	return dataList, err
}

// DeleteByPlugin 删除插件的所有数据
func (r *PluginDataRepositoryImpl) DeleteByPlugin(ctx context.Context, pluginID string) error {
	return r.db.WithContext(ctx).
		Where("plugin_key = ?", pluginID).
		Delete(&database.PluginData{}).Error
}

// BatchSet 批量设置插件数据
func (r *PluginDataRepositoryImpl) BatchSet(ctx context.Context, pluginID string, data map[string]string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
}

// BatchGet 批量获取插件数据
func (r *PluginDataRepositoryImpl) BatchGet(ctx context.Context, pluginID string, keys []string) (map[string]string, error) {
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
	return result, nil
}

// Exists 检查插件数据是否存在
func (r *PluginDataRepositoryImpl) Exists(ctx context.Context, pluginID string, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.PluginData{}).
		Where("plugin_key = ? AND data_key = ?", pluginID, key).
		Count(&count).Error
	return count > 0, err
}

// ListAllPlugins 获取所有插件ID列表
func (r *PluginDataRepositoryImpl) ListAllPlugins(ctx context.Context) ([]string, error) {
	var plugins []string
	err := r.db.WithContext(ctx).
		Model(&database.PluginData{}).
		Distinct("plugin_key").
		Pluck("plugin_key", &plugins).Error
	return plugins, err
}
