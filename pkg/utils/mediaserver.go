package utils

import (
	"fmt"
	"sync"
)

// MediaServerHelper 媒体服务器帮助类
type MediaServerHelper struct {
	services map[string]*MediaServerService
	mutex    sync.RWMutex
}

// MediaServerService 媒体服务器服务信息
type MediaServerService struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
}

// MediaServerConfig 媒体服务器配置
type MediaServerConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Host       string                 `json:"host"`
	Port       int                    `json:"port"`
	APIKey     string                 `json:"api_key"`
	Username   string                 `json:"username"`
	Password   string                 `json:"password"`
	Enabled    bool                   `json:"enabled"`
	Priority   int                    `json:"priority"`
	SSL        bool                   `json:"ssl"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// NewMediaServerHelper 创建媒体服务器助手实例
func NewMediaServerHelper() *MediaServerHelper {
	return &MediaServerHelper{
		services: make(map[string]*MediaServerService),
	}
}

// AddService 添加媒体服务器服务
func (msh *MediaServerHelper) AddService(service *MediaServerService) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if service.Type == "" {
		return fmt.Errorf("service type cannot be empty")
	}

	msh.mutex.Lock()
	defer msh.mutex.Unlock()

	msh.services[service.Name] = service
	return nil
}

// RemoveService 移除媒体服务器服务
func (msh *MediaServerHelper) RemoveService(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	msh.mutex.Lock()
	defer msh.mutex.Unlock()

	if _, exists := msh.services[name]; !exists {
		return fmt.Errorf("service not found: %s", name)
	}

	delete(msh.services, name)
	return nil
}

// GetService 获取媒体服务器服务
func (msh *MediaServerHelper) GetService(name string) (*MediaServerService, error) {
	if name == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	msh.mutex.RLock()
	defer msh.mutex.RUnlock()

	service, exists := msh.services[name]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", name)
	}

	return service, nil
}

// GetAllServices 获取所有媒体服务器服务
func (msh *MediaServerHelper) GetAllServices() []*MediaServerService {
	msh.mutex.RLock()
	defer msh.mutex.RUnlock()

	services := make([]*MediaServerService, 0, len(msh.services))
	for _, service := range msh.services {
		services = append(services, service)
	}

	return services
}

// GetEnabledServices 获取启用的媒体服务器服务
func (msh *MediaServerHelper) GetEnabledServices() []*MediaServerService {
	msh.mutex.RLock()
	defer msh.mutex.RUnlock()

	var enabledServices []*MediaServerService
	for _, service := range msh.services {
		if service.Enabled {
			enabledServices = append(enabledServices, service)
		}
	}

	return enabledServices
}

// IsMediaServer 判断是否为指定类型的媒体服务器
func (msh *MediaServerHelper) IsMediaServer(serviceType string, service *MediaServerService, name string) bool {
	if service == nil && name != "" {
		s, err := msh.GetService(name)
		if err != nil {
			return false
		}
		service = s
	}

	return service != nil && service.Type == serviceType
}

// IsPlex 判断是否为Plex媒体服务器
func (msh *MediaServerHelper) IsPlex(service *MediaServerService, name string) bool {
	return msh.IsMediaServer("plex", service, name)
}

// IsEmby 判断是否为Emby媒体服务器
func (msh *MediaServerHelper) IsEmby(service *MediaServerService, name string) bool {
	return msh.IsMediaServer("emby", service, name)
}

// IsJellyfin 判断是否为Jellyfin媒体服务器
func (msh *MediaServerHelper) IsJellyfin(service *MediaServerService, name string) bool {
	return msh.IsMediaServer("jellyfin", service, name)
}

// GetServiceByType 根据类型获取媒体服务器服务
func (msh *MediaServerHelper) GetServiceByType(serviceType string) []*MediaServerService {
	msh.mutex.RLock()
	defer msh.mutex.RUnlock()

	var services []*MediaServerService
	for _, service := range msh.services {
		if service.Type == serviceType {
			services = append(services, service)
		}
	}

	return services
}

// GetServiceByPriority 根据优先级获取媒体服务器服务
func (msh *MediaServerHelper) GetServiceByPriority() []*MediaServerService {
	services := msh.GetEnabledServices()
	
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

// UpdateService 更新媒体服务器服务
func (msh *MediaServerHelper) UpdateService(name string, updates map[string]interface{}) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	msh.mutex.Lock()
	defer msh.mutex.Unlock()

	service, exists := msh.services[name]
	if !exists {
		return fmt.Errorf("service not found: %s", name)
	}

	// 应用更新
	for key, value := range updates {
		switch key {
		case "name":
			if newName, ok := value.(string); ok {
				delete(msh.services, name)
				service.Name = newName
				msh.services[newName] = service
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

// EnableService 启用媒体服务器服务
func (msh *MediaServerHelper) EnableService(name string) error {
	return msh.UpdateService(name, map[string]interface{}{"enabled": true})
}

// DisableService 禁用媒体服务器服务
func (msh *MediaServerHelper) DisableService(name string) error {
	return msh.UpdateService(name, map[string]interface{}{"enabled": false})
}

// ValidateConfig 验证媒体服务器配置
func (msh *MediaServerHelper) ValidateConfig(config *MediaServerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Type == "" {
		return fmt.Errorf("media server type cannot be empty")
	}

	if config.Name == "" {
		return fmt.Errorf("media server name cannot be empty")
	}

	if config.Host == "" {
		return fmt.Errorf("media server host cannot be empty")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", config.Port)
	}

	// 根据类型进行特定验证
	switch config.Type {
	case "plex":
		return msh.validatePlexConfig(config)
	case "emby":
		return msh.validateEmbyConfig(config)
	case "jellyfin":
		return msh.validateJellyfinConfig(config)
	default:
		return fmt.Errorf("unsupported media server type: %s", config.Type)
	}
}

// validatePlexConfig 验证Plex配置
func (msh *MediaServerHelper) validatePlexConfig(config *MediaServerConfig) error {
	// Plex需要API Token
	if config.APIKey == "" {
		return fmt.Errorf("plex API token is required")
	}
	return nil
}

// validateEmbyConfig 验证Emby配置
func (msh *MediaServerHelper) validateEmbyConfig(config *MediaServerConfig) error {
	// Emby需要API Key
	if config.APIKey == "" {
		return fmt.Errorf("emby API key is required")
	}
	
	// 如果提供了用户名，则必须提供密码
	if config.Username != "" && config.Password == "" {
		return fmt.Errorf("password is required when username is provided")
	}
	
	return nil
}

// validateJellyfinConfig 验证Jellyfin配置
func (msh *MediaServerHelper) validateJellyfinConfig(config *MediaServerConfig) error {
	// Jellyfin需要API Key
	if config.APIKey == "" {
		return fmt.Errorf("jellyfin API key is required")
	}
	
	// 如果提供了用户名，则必须提供密码
	if config.Username != "" && config.Password == "" {
		return fmt.Errorf("password is required when username is provided")
	}
	
	return nil
}

// GetServiceCount 获取服务数量
func (msh *MediaServerHelper) GetServiceCount() int {
	msh.mutex.RLock()
	defer msh.mutex.RUnlock()

	return len(msh.services)
}

// GetEnabledServiceCount 获取启用服务数量
func (msh *MediaServerHelper) GetEnabledServiceCount() int {
	services := msh.GetEnabledServices()
	return len(services)
}

// ClearServices 清空所有服务
func (msh *MediaServerHelper) ClearServices() {
	msh.mutex.Lock()
	defer msh.mutex.Unlock()

	msh.services = make(map[string]*MediaServerService)
}

// ImportServices 导入服务配置
func (msh *MediaServerHelper) ImportServices(configs []*MediaServerConfig) error {
	if configs == nil {
		return fmt.Errorf("configs cannot be nil")
	}

	for _, config := range configs {
		if err := msh.ValidateConfig(config); err != nil {
			return fmt.Errorf("invalid config for %s: %v", config.Name, err)
		}

		service := &MediaServerService{
			Name:     config.Name,
			Type:     config.Type,
			Enabled:  config.Enabled,
			Priority: config.Priority,
			Config: map[string]interface{}{
				"host":     config.Host,
				"port":     config.Port,
				"api_key":  config.APIKey,
				"username": config.Username,
				"password": config.Password,
				"ssl":      config.SSL,
			},
		}

		// 添加额外配置
		for key, value := range config.Extra {
			service.Config[key] = value
		}

		if err := msh.AddService(service); err != nil {
			return fmt.Errorf("failed to add service %s: %v", config.Name, err)
		}
	}

	return nil
}

// ExportServices 导出服务配置
func (msh *MediaServerHelper) ExportServices() []*MediaServerConfig {
	services := msh.GetAllServices()
	configs := make([]*MediaServerConfig, 0, len(services))

	for _, service := range services {
		config := &MediaServerConfig{
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
		if apiKey, ok := service.Config["api_key"].(string); ok {
			config.APIKey = apiKey
		}
		if username, ok := service.Config["username"].(string); ok {
			config.Username = username
		}
		if password, ok := service.Config["password"].(string); ok {
			config.Password = password
		}
		if ssl, ok := service.Config["ssl"].(bool); ok {
			config.SSL = ssl
		}

		// 添加额外配置
		for key, value := range service.Config {
			if key != "host" && key != "port" && key != "api_key" && 
			   key != "username" && key != "password" && key != "ssl" {
				config.Extra[key] = value
			}
		}

		configs = append(configs, config)
	}

	return configs
}

// GetPrimaryService 获取主要媒体服务器（优先级最高的启用服务）
func (msh *MediaServerHelper) GetPrimaryService() (*MediaServerService, error) {
	services := msh.GetServiceByPriority()
	if len(services) == 0 {
		return nil, fmt.Errorf("no enabled media server found")
	}

	return services[0], nil
}

// GetServiceURL 获取媒体服务器URL
func (msh *MediaServerHelper) GetServiceURL(service *MediaServerService) (string, error) {
	if service == nil {
		return "", fmt.Errorf("service cannot be nil")
	}

	host, ok := service.Config["host"].(string)
	if !ok {
		return "", fmt.Errorf("host not found in service config")
	}

	port, ok := service.Config["port"].(int)
	if !ok {
		return "", fmt.Errorf("port not found in service config")
	}

	ssl, ok := service.Config["ssl"].(bool)
	if !ok {
		ssl = false
	}

	protocol := "http"
	if ssl {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s:%d", protocol, host, port), nil
}