package utils

import (
	"fmt"
	"sync"
)

// ServiceConfigHelper 服务配置帮助类
type ServiceConfigHelper struct {
	configs map[string][]ServiceConfig
	mutex   sync.RWMutex
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config"`
	Priority int                    `json:"priority"`
}

// NewServiceConfigHelper 创建服务配置助手实例
func NewServiceConfigHelper() *ServiceConfigHelper {
	return &ServiceConfigHelper{
		configs: make(map[string][]ServiceConfig),
	}
}

// GetConfigs 获取配置
func (sch *ServiceConfigHelper) GetConfigs(configType string) []ServiceConfig {
	sch.mutex.RLock()
	defer sch.mutex.RUnlock()

	configs, exists := sch.configs[configType]
	if !exists {
		return []ServiceConfig{}
	}

	// 返回副本
	result := make([]ServiceConfig, len(configs))
	copy(result, configs)
	return result
}

// SetConfigs 设置配置
func (sch *ServiceConfigHelper) SetConfigs(configType string, configs []ServiceConfig) {
	sch.mutex.Lock()
	defer sch.mutex.Unlock()

	sch.configs[configType] = configs
}

// AddConfig 添加配置
func (sch *ServiceConfigHelper) AddConfig(configType string, config ServiceConfig) error {
	if configType == "" {
		return fmt.Errorf("config type cannot be empty")
	}

	if config.Name == "" {
		return fmt.Errorf("config name cannot be empty")
	}

	sch.mutex.Lock()
	defer sch.mutex.Unlock()

	configs := sch.configs[configType]
	configs = append(configs, config)
	sch.configs[configType] = configs

	return nil
}

// RemoveConfig 移除配置
func (sch *ServiceConfigHelper) RemoveConfig(configType, configName string) error {
	if configType == "" {
		return fmt.Errorf("config type cannot be empty")
	}

	if configName == "" {
		return fmt.Errorf("config name cannot be empty")
	}

	sch.mutex.Lock()
	defer sch.mutex.Unlock()

	configs, exists := sch.configs[configType]
	if !exists {
		return fmt.Errorf("config type not found: %s", configType)
	}

	for i, config := range configs {
		if config.Name == configName {
			sch.configs[configType] = append(configs[:i], configs[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("config not found: %s", configName)
}

// UpdateConfig 更新配置
func (sch *ServiceConfigHelper) UpdateConfig(configType, configName string, updates map[string]interface{}) error {
	if configType == "" {
		return fmt.Errorf("config type cannot be empty")
	}

	if configName == "" {
		return fmt.Errorf("config name cannot be empty")
	}

	sch.mutex.Lock()
	defer sch.mutex.Unlock()

	configs, exists := sch.configs[configType]
	if !exists {
		return fmt.Errorf("config type not found: %s", configType)
	}

	for i, config := range configs {
		if config.Name == configName {
			// 更新配置
			if newConfig, ok := updates["config"].(map[string]interface{}); ok {
				configs[i].Config = newConfig
			}
			
			// 更新其他字段
			if name, ok := updates["name"].(string); ok {
				configs[i].Name = name
			}
			
			if serviceType, ok := updates["type"].(string); ok {
				configs[i].Type = serviceType
			}

			sch.configs[configType] = configs
			return nil
		}
	}

	return fmt.Errorf("config not found: %s", configName)
}

// GetConfig 获取单个配置
func (sch *ServiceConfigHelper) GetConfig(configType, configName string) (*ServiceConfig, error) {
	if configType == "" {
		return nil, fmt.Errorf("config type cannot be empty")
	}

	if configName == "" {
		return nil, fmt.Errorf("config name cannot be empty")
	}

	sch.mutex.RLock()
	defer sch.mutex.RUnlock()

	configs, exists := sch.configs[configType]
	if !exists {
		return nil, fmt.Errorf("config type not found: %s", configType)
	}

	for _, config := range configs {
		if config.Name == configName {
			// 返回副本
			return &ServiceConfig{
				Type:   config.Type,
				Name:   config.Name,
				Config: config.Config,
			}, nil
		}
	}

	return nil, fmt.Errorf("config not found: %s", configName)
}

// GetDownloaderConfigs 获取下载器配置
func (sch *ServiceConfigHelper) GetDownloaderConfigs() []ServiceConfig {
	return sch.GetConfigs("downloaders")
}

// GetMediaserverConfigs 获取媒体服务器配置
func (sch *ServiceConfigHelper) GetMediaserverConfigs() []ServiceConfig {
	return sch.GetConfigs("mediaservers")
}

// GetNotificationConfigs 获取消息通知配置
func (sch *ServiceConfigHelper) GetNotificationConfigs() []ServiceConfig {
	return sch.GetConfigs("notifications")
}

// GetStorageConfigs 获取存储配置
func (sch *ServiceConfigHelper) GetStorageConfigs() []ServiceConfig {
	return sch.GetConfigs("storages")
}

// SetDownloaderConfigs 设置下载器配置
func (sch *ServiceConfigHelper) SetDownloaderConfigs(configs []ServiceConfig) {
	sch.SetConfigs("downloaders", configs)
}

// SetMediaserverConfigs 设置媒体服务器配置
func (sch *ServiceConfigHelper) SetMediaserverConfigs(configs []ServiceConfig) {
	sch.SetConfigs("mediaservers", configs)
}

// SetNotificationConfigs 设置消息通知配置
func (sch *ServiceConfigHelper) SetNotificationConfigs(configs []ServiceConfig) {
	sch.SetConfigs("notifications", configs)
}

// SetStorageConfigs 设置存储配置
func (sch *ServiceConfigHelper) SetStorageConfigs(configs []ServiceConfig) {
	sch.SetConfigs("storages", configs)
}

// ValidateConfig 验证配置
func (sch *ServiceConfigHelper) ValidateConfig(configType string, config ServiceConfig) error {
	if configType == "" {
		return fmt.Errorf("config type cannot be empty")
	}

	if config.Name == "" {
		return fmt.Errorf("config name cannot be empty")
	}

	if config.Type == "" {
		return fmt.Errorf("config type cannot be empty")
	}

	if config.Config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 根据配置类型进行特定验证
	switch configType {
	case "downloaders":
		return sch.validateDownloaderConfig(config)
	case "mediaservers":
		return sch.validateMediaserverConfig(config)
	case "notifications":
		return sch.validateNotificationConfig(config)
	case "storages":
		return sch.validateStorageConfig(config)
	default:
		return fmt.Errorf("unknown config type: %s", configType)
	}
}

// validateDownloaderConfig 验证下载器配置
func (sch *ServiceConfigHelper) validateDownloaderConfig(config ServiceConfig) error {
	// 检查必需字段
	if host, exists := config.Config["host"]; !exists || host.(string) == "" {
		return fmt.Errorf("downloader host is required")
	}

	if port, exists := config.Config["port"]; exists {
		if portInt, ok := port.(int); ok && (portInt <= 0 || portInt > 65535) {
			return fmt.Errorf("invalid port number: %d", portInt)
		}
	}

	return nil
}

// validateMediaserverConfig 验证媒体服务器配置
func (sch *ServiceConfigHelper) validateMediaserverConfig(config ServiceConfig) error {
	// 检查必需字段
	if host, exists := config.Config["host"]; !exists || host.(string) == "" {
		return fmt.Errorf("mediaserver host is required")
	}

	if apiKey, exists := config.Config["api_key"]; !exists || apiKey.(string) == "" {
		return fmt.Errorf("mediaserver api key is required")
	}

	return nil
}

// validateNotificationConfig 验证消息通知配置
func (sch *ServiceConfigHelper) validateNotificationConfig(config ServiceConfig) error {
	// 根据通知类型进行验证
	switch config.Type {
	case "telegram":
		if token, exists := config.Config["token"]; !exists || token.(string) == "" {
			return fmt.Errorf("telegram bot token is required")
		}
		if chatID, exists := config.Config["chat_id"]; !exists || chatID.(string) == "" {
			return fmt.Errorf("telegram chat ID is required")
		}
	case "email":
		if username, exists := config.Config["username"]; !exists || username.(string) == "" {
			return fmt.Errorf("email username is required")
		}
		if password, exists := config.Config["password"]; !exists || password.(string) == "" {
			return fmt.Errorf("email password is required")
		}
	}

	return nil
}

// validateStorageConfig 验证存储配置
func (sch *ServiceConfigHelper) validateStorageConfig(config ServiceConfig) error {
	// 根据存储类型进行验证
	switch config.Type {
	case "local":
		// 本地存储通常不需要特殊验证
	case "sftp":
		if host, exists := config.Config["host"]; !exists || host.(string) == "" {
			return fmt.Errorf("sftp host is required")
		}
		if username, exists := config.Config["username"]; !exists || username.(string) == "" {
			return fmt.Errorf("sftp username is required")
		}
	case "s3":
		if endpoint, exists := config.Config["endpoint"]; !exists || endpoint.(string) == "" {
			return fmt.Errorf("s3 endpoint is required")
		}
		if bucket, exists := config.Config["bucket"]; !exists || bucket.(string) == "" {
			return fmt.Errorf("s3 bucket is required")
		}
	}

	return nil
}

// GetConfigCount 获取配置数量
func (sch *ServiceConfigHelper) GetConfigCount(configType string) int {
	configs := sch.GetConfigs(configType)
	return len(configs)
}

// GetAllConfigTypes 获取所有配置类型
func (sch *ServiceConfigHelper) GetAllConfigTypes() []string {
	sch.mutex.RLock()
	defer sch.mutex.RUnlock()

	types := make([]string, 0, len(sch.configs))
	for configType := range sch.configs {
		types = append(types, configType)
	}

	return types
}

// ClearConfigs 清空配置
func (sch *ServiceConfigHelper) ClearConfigs(configType string) {
	sch.mutex.Lock()
	defer sch.mutex.Unlock()

	delete(sch.configs, configType)
}

// ClearAllConfigs 清空所有配置
func (sch *ServiceConfigHelper) ClearAllConfigs() {
	sch.mutex.Lock()
	defer sch.mutex.Unlock()

	sch.configs = make(map[string][]ServiceConfig)
}

// ExportConfigs 导出配置
func (sch *ServiceConfigHelper) ExportConfigs() map[string][]ServiceConfig {
	sch.mutex.RLock()
	defer sch.mutex.RUnlock()

	// 返回副本
	export := make(map[string][]ServiceConfig)
	for configType, configs := range sch.configs {
		export[configType] = append([]ServiceConfig{}, configs...)
	}

	return export
}

// ImportConfigs 导入配置
func (sch *ServiceConfigHelper) ImportConfigs(configs map[string][]ServiceConfig) error {
	if configs == nil {
		return fmt.Errorf("configs cannot be nil")
	}

	sch.mutex.Lock()
	defer sch.mutex.Unlock()

	// 验证所有配置
	for configType, configList := range configs {
		for _, config := range configList {
			if err := sch.validateDownloaderConfig(config); err != nil {
				return fmt.Errorf("invalid config for %s.%s: %v", configType, config.Name, err)
			}
		}
	}

	// 导入配置
	sch.configs = make(map[string][]ServiceConfig)
	for configType, configList := range configs {
		sch.configs[configType] = append([]ServiceConfig{}, configList...)
	}

	return nil
}

// GetServiceInfo 获取服务信息
func (sch *ServiceConfigHelper) GetServiceInfo(configType, configName string) (*ServiceInfo, error) {
	config, err := sch.GetConfig(configType, configName)
	if err != nil {
		return nil, err
	}

	// 转换为ServiceInfo
	info := &ServiceInfo{
		Type:   config.Type,
		Name:   config.Name,
		Config: config.Config,
	}

	// 从配置中提取其他信息
	if enabled, exists := config.Config["enabled"]; exists {
		if enabledBool, ok := enabled.(bool); ok {
			info.Enabled = enabledBool
		}
	}

	if priority, exists := config.Config["priority"]; exists {
		if priorityInt, ok := priority.(int); ok {
			info.Priority = priorityInt
		}
	}

	return info, nil
}

// GetEnabledConfigs 获取启用的配置
func (sch *ServiceConfigHelper) GetEnabledConfigs(configType string) []ServiceConfig {
	configs := sch.GetConfigs(configType)
	var result []ServiceConfig

	for _, config := range configs {
		if enabled, exists := config.Config["enabled"]; exists {
			if enabledBool, ok := enabled.(bool); ok && enabledBool {
				result = append(result, config)
			}
		} else {
			// 默认启用
			result = append(result, config)
		}
	}

	return result
}

// GetConfigsByType 根据服务类型获取配置
func (sch *ServiceConfigHelper) GetConfigsByType(configType, serviceType string) []ServiceConfig {
	configs := sch.GetConfigs(configType)
	var filtered []ServiceConfig

	for _, config := range configs {
		if config.Type == serviceType {
			filtered = append(filtered, config)
		}
	}

	return filtered
}

// GetConfigNames 获取配置名称列表
func (sch *ServiceConfigHelper) GetConfigNames(configType string) []string {
	configs := sch.GetConfigs(configType)
	names := make([]string, len(configs))

	for i, config := range configs {
		names[i] = config.Name
	}

	return names
}

// HasConfig 检查配置是否存在
func (sch *ServiceConfigHelper) HasConfig(configType, configName string) bool {
	_, err := sch.GetConfig(configType, configName)
	return err == nil
}

// CloneConfig 克隆配置
func (sch *ServiceConfigHelper) CloneConfig(configType, configName, newName string) error {
	config, err := sch.GetConfig(configType, configName)
	if err != nil {
		return err
	}

	// 创建副本
	clonedConfig := ServiceConfig{
		Type:   config.Type,
		Name:   newName,
		Config: cloneMap(config.Config),
	}

	return sch.AddConfig(configType, clonedConfig)
}

// cloneMap 克隆map
func cloneMap(original map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{})
	for key, value := range original {
		cloned[key] = value
	}
	return cloned
}