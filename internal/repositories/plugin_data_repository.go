package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/pkg/plugin"
)

// PluginDataRepository 插件数据存储实现
type PluginDataRepository struct {
	db *gorm.DB
}

// NewPluginDataRepository 创建插件数据存储
func NewPluginDataRepository(db *gorm.DB) plugin.DataStore {
	return &PluginDataRepository{
		db: db,
	}
}

// Get 获取插件数据
func (r *PluginDataRepository) Get(ctx context.Context, pluginID, key string) (any, error) {
	var pluginData database.PluginData
	result := r.db.Where("plugin_key = ? AND data_key = ?", pluginID, key).First(&pluginData)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取插件数据失败: %w", result.Error)
	}

	return pluginData.DataValue, nil
}

// Set 设置插件数据
func (r *PluginDataRepository) Set(ctx context.Context, pluginID, key string, value any) error {
	// 查找或创建插件数据
	var pluginData database.PluginData
	result := r.db.Where("plugin_key = ? AND data_key = ?", pluginID, key).FirstOrCreate(&pluginData, database.PluginData{
		PluginKey: pluginID,
		DataKey:   key,
		DataValue: fmt.Sprintf("%v", value),
	})
	if result.Error != nil {
		return fmt.Errorf("创建插件数据失败: %w", result.Error)
	}

	// 更新数据
	pluginData.DataValue = fmt.Sprintf("%v", value)
	if err := r.db.Save(&pluginData).Error; err != nil {
		return fmt.Errorf("保存插件数据失败: %w", err)
	}

	return nil
}

// Delete 删除插件数据
func (r *PluginDataRepository) Delete(ctx context.Context, pluginID, key string) error {
	result := r.db.Where("plugin_key = ? AND data_key = ?", pluginID, key).Delete(&database.PluginData{})
	if result.Error != nil {
		return fmt.Errorf("删除插件数据失败: %w", result.Error)
	}
	return nil
}

// DeleteAll 删除插件的所有数据
func (r *PluginDataRepository) DeleteAll(ctx context.Context, pluginID string) error {
	result := r.db.Where("plugin_key = ?", pluginID).Delete(&database.PluginData{})
	if result.Error != nil {
		return fmt.Errorf("删除插件所有数据失败: %w", result.Error)
	}
	return nil
}
