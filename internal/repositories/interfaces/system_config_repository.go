package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// SystemConfigRepository 系统配置仓储接口
type SystemConfigRepository interface {
	// Get 根据key获取配置
	Get(ctx context.Context, key string) (*database.SystemConfig, error)

	// Set 设置配置
	Set(ctx context.Context, key string, value string) error

	// Delete 删除配置
	Delete(ctx context.Context, key string) error

	// List 获取所有配置
	List(ctx context.Context) ([]*database.SystemConfig, error)

	// BatchSet 批量设置配置
	BatchSet(ctx context.Context, configs map[string]string) error

	// BatchGet 批量获取配置
	BatchGet(ctx context.Context, keys []string) (map[string]string, error)

	// Exists 检查配置是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// GetByPrefix 根据前缀获取配置
	GetByPrefix(ctx context.Context, prefix string) ([]*database.SystemConfig, error)

	// DeleteByPrefix 根据前缀删除配置
	DeleteByPrefix(ctx context.Context, prefix string) error

	// GetType 获取配置类型
	GetType(ctx context.Context, key string) (string, error)

	// SetWithType 设置配置和类型
	SetWithType(ctx context.Context, key string, value string, configType string) error

	// BatchSetWithTypes 批量设置配置和类型
	BatchSetWithTypes(ctx context.Context, configs map[string]ConfigItem) error

	// BatchDelete 批量删除配置
	BatchDelete(ctx context.Context, keys []string) error
}

// ConfigItem 配置项辅助类型
type ConfigItem struct {
	Value string
	Type  string
}
