// Package providers 配置提供者接口
package providers

import "context"

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	// Load 加载配置
	Load(ctx context.Context, path string) (map[string]any, error)

	// Watch 监听配置变化
	Watch(ctx context.Context, path string, callback func(map[string]any)) error

	// Save 保存配置
	Save(ctx context.Context, path string, config map[string]any) error

	// Validate 验证配置
	Validate(ctx context.Context, config map[string]any, rules map[string]any) error
}

// ProviderConfig 提供者配置
type ProviderConfig struct {
	Type     string         `yaml:"type"`
	Settings map[string]any `yaml:"settings"`
}
