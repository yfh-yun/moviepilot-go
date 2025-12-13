package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/pkg/plugin"
)

// PluginConfigRepository 插件配置存储实现
type PluginConfigRepository struct {
	db *gorm.DB
}

// NewPluginConfigRepository 创建插件配置存储
func NewPluginConfigRepository(db *gorm.DB) plugin.ConfigStore {
	return &PluginConfigRepository{
		db: db,
	}
}

// Get 获取插件配置
func (r *PluginConfigRepository) Get(ctx context.Context, pluginID string) (map[string]any, error) {
	// 使用SystemConfig表存储插件配置
	var config database.SystemConfig
	result := r.db.Where("key = ?", "plugin_config_"+pluginID).First(&config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("获取插件配置失败: %w", result.Error)
	}

	var configData map[string]any
	if err := json.Unmarshal([]byte(config.Value), &configData); err != nil {
		return nil, fmt.Errorf("解析插件配置失败: %w", err)
	}

	return configData, nil
}

// Set 设置插件配置
func (r *PluginConfigRepository) Set(ctx context.Context, pluginID string, config map[string]any) error {
	// 将配置转换为JSON字符串
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化插件配置失败: %w", err)
	}

	// 查找或创建插件配置
	var sysConfig database.SystemConfig
	result := r.db.Where("key = ?", "plugin_config_"+pluginID).FirstOrCreate(&sysConfig, database.SystemConfig{
		Key:    "plugin_config_" + pluginID,
		Value:  string(configJSON),
		Type:   "json",
		Remark: "Plugin config for " + pluginID,
	})
	if result.Error != nil {
		return fmt.Errorf("创建插件配置失败: %w", result.Error)
	}

	// 更新配置
	sysConfig.Value = string(configJSON)
	if err := r.db.Save(&sysConfig).Error; err != nil {
		return fmt.Errorf("保存插件配置失败: %w", err)
	}

	return nil
}

// Delete 删除插件配置
func (r *PluginConfigRepository) Delete(ctx context.Context, pluginID string) error {
	result := r.db.Where("key = ?", "plugin_config_"+pluginID).Delete(&database.SystemConfig{})
	if result.Error != nil {
		return fmt.Errorf("删除插件配置失败: %w", result.Error)
	}
	return nil
}
