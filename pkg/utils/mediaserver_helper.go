package utils

import (
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// MediaServerConf 媒体服务器配置
type MediaServerConf struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	APIKey   string `json:"api_key"`
	Username string `json:"username"`
	Password string `json:"password"`
	Enabled  bool   `json:"enabled"`
}

// MediaServerServiceInfo 服务信息
type MediaServerServiceInfo struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// MediaServerSystemConfigKey 媒体服务器系统配置键
type MediaServerSystemConfigKey string

// MediaServerModuleType 媒体服务器模块类型
type MediaServerModuleType string

// 常量定义
const (
	MediaServerSystemConfigKeyMediaServers MediaServerSystemConfigKey = "MediaServers"
	MediaServerModuleTypeMediaServer       MediaServerModuleType      = "MediaServer"
)

// MediaServerHelper 媒体服务器帮助类
type MediaServerHelper struct {
	configKey  MediaServerSystemConfigKey
	confType   string
	moduleType MediaServerModuleType
	services   []*MediaServerServiceInfo
	mutex      sync.RWMutex
	logger     *zap.Logger
}

// NewMediaServerHelper 创建媒体服务器帮助类
func NewMediaServerHelper() *MediaServerHelper {
	return &MediaServerHelper{
		configKey:  MediaServerSystemConfigKeyMediaServers,
		confType:   "MediaServerConf",
		moduleType: MediaServerModuleTypeMediaServer,
		services:   make([]*MediaServerServiceInfo, 0),
		logger:     logger.GetLogger(),
	}
}

// IsMediaServer 判断服务是否为指定类型的媒体服务器
func (h *MediaServerHelper) IsMediaServer(serviceType string, service *MediaServerServiceInfo, name string) bool {
	// 如果未提供service，则通过name获取服务
	if service == nil && name != "" {
		service = h.GetService(name)
	}

	// 判断服务类型是否为指定类型
	return service != nil && service.Type == serviceType
}

// GetService 通过名称获取服务
func (h *MediaServerHelper) GetService(name string) *MediaServerServiceInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, service := range h.services {
		if service.Name == name {
			return service
		}
	}

	return nil
}

// GetServices 获取所有服务
func (h *MediaServerHelper) GetServices() []*MediaServerServiceInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 返回服务的副本，避免外部修改
	services := make([]*MediaServerServiceInfo, len(h.services))
	copy(services, h.services)
	return services
}

// AddService 添加服务
func (h *MediaServerHelper) AddService(service *MediaServerServiceInfo) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.services = append(h.services, service)
}

// RemoveService 移除服务
func (h *MediaServerHelper) RemoveService(id uint) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for i, service := range h.services {
		if service.ID == id {
			h.services = append(h.services[:i], h.services[i+1:]...)
			break
		}
	}
}

// UpdateService 更新服务
func (h *MediaServerHelper) UpdateService(service *MediaServerServiceInfo) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for i, s := range h.services {
		if s.ID == service.ID {
			h.services[i] = service
			break
		}
	}
}
