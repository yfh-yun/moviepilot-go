package utils

import (
	"fmt"
	"sync"
)

// DownloaderHelper 下载器帮助类
type DownloaderHelper struct {
	services map[string]*DownloaderService
	mutex    sync.RWMutex
}

// DownloaderService 下载器服务信息
type DownloaderService struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
}

// DownloaderConfig 下载器配置
type DownloaderConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Host       string                 `json:"host"`
	Port       int                    `json:"port"`
	Username   string                 `json:"username"`
	Password   string                 `json:"password"`
	Enabled    bool                   `json:"enabled"`
	Priority   int                    `json:"priority"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// NewDownloaderHelper 创建下载器助手实例
func NewDownloaderHelper() *DownloaderHelper {
	return &DownloaderHelper{
		services: make(map[string]*DownloaderService),
	}
}

// AddService 添加下载器服务
func (dh *DownloaderHelper) AddService(service *DownloaderService) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if service.Type == "" {
		return fmt.Errorf("service type cannot be empty")
	}

	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	dh.services[service.Name] = service
	return nil
}

// RemoveService 移除下载器服务
func (dh *DownloaderHelper) RemoveService(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	if _, exists := dh.services[name]; !exists {
		return fmt.Errorf("service not found: %s", name)
	}

	delete(dh.services, name)
	return nil
}

// GetService 获取下载器服务
func (dh *DownloaderHelper) GetService(name string) (*DownloaderService, error) {
	if name == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	dh.mutex.RLock()
	defer dh.mutex.RUnlock()

	service, exists := dh.services[name]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", name)
	}

	return service, nil
}

// GetAllServices 获取所有下载器服务
func (dh *DownloaderHelper) GetAllServices() []*DownloaderService {
	dh.mutex.RLock()
	defer dh.mutex.RUnlock()

	services := make([]*DownloaderService, 0, len(dh.services))
	for _, service := range dh.services {
		services = append(services, service)
	}

	return services
}

// GetEnabledServices 获取启用的下载器服务
func (dh *DownloaderHelper) GetEnabledServices() []*DownloaderService {
	dh.mutex.RLock()
	defer dh.mutex.RUnlock()

	var enabledServices []*DownloaderService
	for _, service := range dh.services {
		if service.Enabled {
			enabledServices = append(enabledServices, service)
		}
	}

	return enabledServices
}

// IsDownloader 判断是否为指定类型的下载器
func (dh *DownloaderHelper) IsDownloader(serviceType string, service *DownloaderService, name string) bool {
	if service == nil && name != "" {
		s, err := dh.GetService(name)
		if err != nil {
			return false
		}
		service = s
	}

	return service != nil && service.Type == serviceType
}

// IsQbittorrent 判断是否为qBittorrent下载器
func (dh *DownloaderHelper) IsQbittorrent(service *DownloaderService, name string) bool {
	return dh.IsDownloader("qbittorrent", service, name)
}

// IsTransmission 判断是否为Transmission下载器
func (dh *DownloaderHelper) IsTransmission(service *DownloaderService, name string) bool {
	return dh.IsDownloader("transmission", service, name)
}

// IsAria2 判断是否为Aria2下载器
func (dh *DownloaderHelper) IsAria2(service *DownloaderService, name string) bool {
	return dh.IsDownloader("aria2", service, name)
}

// GetServiceByType 根据类型获取下载器服务
func (dh *DownloaderHelper) GetServiceByType(serviceType string) []*DownloaderService {
	dh.mutex.RLock()
	defer dh.mutex.RUnlock()

	var services []*DownloaderService
	for _, service := range dh.services {
		if service.Type == serviceType {
			services = append(services, service)
		}
	}

	return services
}

// GetServiceByPriority 根据优先级获取下载器服务
func (dh *DownloaderHelper) GetServiceByPriority() []*DownloaderService {
	services := dh.GetEnabledServices()
	
	// 按优先级排序（优先级数字越小，优先级越高）
	for i := 0; i < len(services)-1; i++ {
		for j := i + 1; j < len(services); j++ {
			if services[i].Priority > services[j].Priority {
				services[i], services[j] = services[j], services[i]
			}
		}
	}

	return services
}

// UpdateService 更新下载器服务
func (dh *DownloaderHelper) UpdateService(name string, updates map[string]interface{}) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	service, exists := dh.services[name]
	if !exists {
		return fmt.Errorf("service not found: %s", name)
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "name":
			if newName, ok := value.(string); ok {
				delete(dh.services, name)
				service.Name = newName
				dh.services[newName] = service
			}
		case "type":
			if serviceType, ok := value.(string); ok {
				service.Type = serviceType
			}
		case "enabled":
			if enabled, ok := value.(bool); ok {
				service.Enabled = enabled
			}
		case "priority":
			if priority, ok := value.(int); ok {
				service.Priority = priority
			}
		case "config":
			if config, ok := value.(map[string]interface{}); ok {
				service.Config = config
			}
		}
	}

	return nil
}

// EnableService 启用下载器服务
func (dh *DownloaderHelper) EnableService(name string) error {
	return dh.UpdateService(name, map[string]interface{}{"enabled": true})
}

// DisableService 禁用下载器服务
func (dh *DownloaderHelper) DisableService(name string) error {
	return dh.UpdateService(name, map[string]interface{}{"enabled": false})
}

// ValidateConfig 验证下载器配置
func (dh *DownloaderHelper) ValidateConfig(config *DownloaderConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Type == "" {
		return fmt.Errorf("downloader type cannot be empty")
	}

	if config.Name == "" {
		return fmt.Errorf("downloader name cannot be empty")
	}

	if config.Host == "" {
		return fmt.Errorf("downloader host cannot be empty")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", config.Port)
	}

	// 根据类型进行特定验证
	switch config.Type {
	case "qbittorrent":
		return dh.validateQbittorrentConfig(config)
	case "transmission":
		return dh.validateTransmissionConfig(config)
	case "aria2":
		return dh.validateAria2Config(config)
	default:
		return fmt.Errorf("unsupported downloader type: %s", config.Type)
	}
}

// validateQbittorrentConfig 验证qBittorrent配置
func (dh *DownloaderHelper) validateQbittorrentConfig(config *DownloaderConfig) error {
	// qBittorrent特定验证逻辑
	if config.Username == "" && config.Password != "" {
		return fmt.Errorf("username is required when password is provided")
	}
	return nil
}

// validateTransmissionConfig 验证Transmission配置
func (dh *DownloaderHelper) validateTransmissionConfig(config *DownloaderConfig) error {
	// Transmission特定验证逻辑
	if config.Username == "" && config.Password != "" {
		return fmt.Errorf("username is required when password is provided")
	}
	return nil
}

// validateAria2Config 验证Aria2配置
func (dh *DownloaderHelper) validateAria2Config(config *DownloaderConfig) error {
	// Aria2特定验证逻辑
	// Aria2通常不需要认证，但如果提供了token，则验证格式
	if token, exists := config.Extra["token"]; exists {
		if tokenStr, ok := token.(string); !ok || tokenStr == "" {
			return fmt.Errorf("invalid aria2 token format")
		}
	}
	return nil
}

// GetServiceCount 获取服务数量
func (dh *DownloaderHelper) GetServiceCount() int {
	dh.mutex.RLock()
	defer dh.mutex.RUnlock()

	return len(dh.services)
}

// GetEnabledServiceCount 获取启用服务数量
func (dh *DownloaderHelper) GetEnabledServiceCount() int {
	services := dh.GetEnabledServices()
	return len(services)
}

// ClearServices 清空所有服务
func (dh *DownloaderHelper) ClearServices() {
	dh.mutex.Lock()
	defer dh.mutex.Unlock()

	dh.services = make(map[string]*DownloaderService)
}

// ImportServices 导入服务配置
func (dh *DownloaderHelper) ImportServices(configs []*DownloaderConfig) error {
	if configs == nil {
		return fmt.Errorf("configs cannot be nil")
	}

	for _, config := range configs {
		if err := dh.ValidateConfig(config); err != nil {
			return fmt.Errorf("invalid config for %s: %v", config.Name, err)
		}

		service := &DownloaderService{
			Name:     config.Name,
			Type:     config.Type,
			Enabled:  config.Enabled,
			Priority: config.Priority,
			Config: map[string]interface{}{
				"host":     config.Host,
				"port":     config.Port,
				"username": config.Username,
				"password": config.Password,
			},
		}

		// 添加额外配置
		for key, value := range config.Extra {
			service.Config[key] = value
		}

		if err := dh.AddService(service); err != nil {
			return fmt.Errorf("failed to add service %s: %v", config.Name, err)
		}
	}

	return nil
}

// ExportServices 导出服务配置
func (dh *DownloaderHelper) ExportServices() []*DownloaderConfig {
	services := dh.GetAllServices()
	configs := make([]*DownloaderConfig, 0, len(services))

	for _, service := range services {
		config := &DownloaderConfig{
			Type:     service.Type,
			Name:     service.Name,
			Enabled:  service.Enabled,
			Priority: service.Priority,
			Extra:    make(map[string]interface{}),
		}

		// 从配置中提取基本信息
		if host, ok := service.Config["host"].(string); ok {
			config.Host = host
		}
		if port, ok := service.Config["port"].(int); ok {
			config.Port = port
		}
		if username, ok := service.Config["username"].(string); ok {
			config.Username = username
		}
		if password, ok := service.Config["password"].(string); ok {
			config.Password = password
		}

		// 添加额外配置
		for key, value := range service.Config {
			if key != "host" && key != "port" && key != "username" && key != "password" {
				config.Extra[key] = value
			}
		}

		configs = append(configs, config)
	}

	return configs
}