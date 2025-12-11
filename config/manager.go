// Package config 配置管理
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"moviepilot-go/config/providers"
	"moviepilot-go/config/providers/file"
)

// Manager 配置管理器
type Manager struct {
	providers map[string]providers.ConfigProvider
	env       string
}

// NewManager 创建配置管理器
func NewManager(env string) *Manager {
	return &Manager{
		providers: make(map[string]providers.ConfigProvider),
		env:       env,
	}
}

// RegisterProvider 注册配置提供者
func (m *Manager) RegisterProvider(name string, provider providers.ConfigProvider) {
	m.providers[name] = provider
}

// Load 加载配置
func (m *Manager) Load(ctx context.Context, path string) (map[string]any, error) {
	// 尝试从环境特定配置加载
	envPath := filepath.Join("environments", m.env, path)

	if provider, exists := m.providers["file"]; exists {
		config, err := provider.Load(ctx, envPath)
		if err == nil {
			return config, nil
		}
	}

	// 回退到默认配置
	if provider, exists := m.providers["file"]; exists {
		return provider.Load(ctx, path)
	}

	return nil, fmt.Errorf("no suitable provider found")
}

// GetEnv 获取当前环境
func (m *Manager) GetEnv() string {
	if m.env == "" {
		m.env = os.Getenv("APP_ENV")
		if m.env == "" {
			m.env = "development"
		}
	}
	return m.env
}

// Init 初始化配置管理器
func Init() (*Manager, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	manager := NewManager(env)

	// 注册文件提供者
	fileProvider := file.NewFileProvider("config")
	manager.RegisterProvider("file", fileProvider)

	return manager, nil
}
