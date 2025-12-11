package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// PluginDataRepository 插件数据仓储接口
type PluginDataRepository interface {
	// Get 根据插件ID和数据key获取数据
	Get(ctx context.Context, pluginID string, key string) (*database.PluginData, error)

	// Set 设置插件数据
	Set(ctx context.Context, pluginID string, key string, value string) error

	// Delete 删除插件数据
	Delete(ctx context.Context, pluginID string, key string) error

	// ListByPlugin 获取插件的所有数据
	ListByPlugin(ctx context.Context, pluginID string) ([]*database.PluginData, error)

	// DeleteByPlugin 删除插件的所有数据
	DeleteByPlugin(ctx context.Context, pluginID string) error

	// BatchSet 批量设置插件数据
	BatchSet(ctx context.Context, pluginID string, data map[string]string) error

	// BatchGet 批量获取插件数据
	BatchGet(ctx context.Context, pluginID string, keys []string) (map[string]string, error)

	// Exists 检查插件数据是否存在
	Exists(ctx context.Context, pluginID string, key string) (bool, error)

	// ListAllPlugins 获取所有插件ID列表
	ListAllPlugins(ctx context.Context) ([]string, error)
}
