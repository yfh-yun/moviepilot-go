package mediaserver

import (
	"context"
	"fmt"
)

// Config 媒体服务器配置
type Config struct {
	Type   string
	URL    string
	APIKey string
}

// Manager 媒体服务器管理器
type Manager struct {
	servers map[string]Server
}

// NewManager 创建媒体服务器管理器
func NewManager() *Manager {
	return &Manager{
		servers: make(map[string]Server),
	}
}

// Get 获取媒体服务器实例
func (m *Manager) Get(serverID string) (Server, error) {
	server, ok := m.servers[serverID]
	if !ok {
		return nil, fmt.Errorf("媒体服务器不存在: %s", serverID)
	}
	return server, nil
}

// RefreshAll 刷新所有媒体服务器
func (m *Manager) RefreshAll(ctx context.Context) error {
	for _, server := range m.servers {
		if err := server.RefreshAll(ctx); err != nil {
			// 记录错误但继续处理其他服务器
			continue
		}
	}
	return nil
}

// CreateServer 创建媒体服务器实例
func CreateServer(config Config) (Server, error) {
	// TODO: 根据配置类型创建不同的服务器实例
	return nil, fmt.Errorf("未实现")
}

// Server 媒体服务器接口
type Server interface {
	// Test 测试连接
	Test(ctx context.Context) error
	// GetLibraries 获取媒体库列表
	GetLibraries(ctx context.Context) ([]Library, error)
	// RefreshLibrary 刷新指定媒体库
	RefreshLibrary(ctx context.Context, libraryID string) error
	// RefreshAll 刷新所有媒体库
	RefreshAll(ctx context.Context) error
}

// Library 媒体库
type Library struct {
	ID   string
	Name string
	Type string
}
