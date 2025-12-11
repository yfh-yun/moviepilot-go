package plugin

import (
	"context"
	"sync"
)

// MockConfigStore 插件配置存储的模拟实现
type MockConfigStore struct {
	mu      sync.RWMutex
	configs map[string]map[string]any
}

// NewMockConfigStore 创建模拟配置存储
func NewMockConfigStore() *MockConfigStore {
	return &MockConfigStore{
		configs: make(map[string]map[string]any),
	}
}

// Get 获取插件配置
func (m *MockConfigStore) Get(ctx context.Context, pluginID string) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[pluginID]
	if !exists {
		return make(map[string]any), nil
	}

	// 返回配置的副本，防止外部修改
	result := make(map[string]any)
	for k, v := range config {
		result[k] = v
	}

	return result, nil
}

// Set 设置插件配置
func (m *MockConfigStore) Set(ctx context.Context, pluginID string, config map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 保存配置的副本，防止外部修改
	m.configs[pluginID] = make(map[string]any)
	for k, v := range config {
		m.configs[pluginID][k] = v
	}

	return nil
}

// Delete 删除插件配置
func (m *MockConfigStore) Delete(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.configs, pluginID)
	return nil
}

// MockDataStore 插件数据存储的模拟实现
type MockDataStore struct {
	mu   sync.RWMutex
	data map[string]map[string]any
}

// NewMockDataStore 创建模拟数据存储
func NewMockDataStore() *MockDataStore {
	return &MockDataStore{
		data: make(map[string]map[string]any),
	}
}

// Get 获取插件数据
func (m *MockDataStore) Get(ctx context.Context, pluginID, key string) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pluginData, exists := m.data[pluginID]
	if !exists {
		return nil, nil
	}

	value, exists := pluginData[key]
	if !exists {
		return nil, nil
	}

	return value, nil
}

// Set 设置插件数据
func (m *MockDataStore) Set(ctx context.Context, pluginID, key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[pluginID]; !exists {
		m.data[pluginID] = make(map[string]any)
	}

	m.data[pluginID][key] = value
	return nil
}

// Delete 删除插件数据
func (m *MockDataStore) Delete(ctx context.Context, pluginID, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pluginData, exists := m.data[pluginID]; exists {
		delete(pluginData, key)
	}

	return nil
}

// DeleteAll 删除插件所有数据
func (m *MockDataStore) DeleteAll(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, pluginID)
	return nil
}
