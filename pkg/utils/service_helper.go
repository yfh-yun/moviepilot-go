package utils

import (
	"sync"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// ServiceServiceInfo 服务信息
type ServiceServiceInfo struct {
	Name     string `json:"name"`
	Instance any    `json:"instance"`
	Module   any    `json:"module"`
	Type     string `json:"type"`
	Config   any    `json:"config"`
}

// ConfigHelper 服务配置帮助类
type ConfigHelper struct {
	logger      *zap.Logger
	configStore ServiceConfigStore
}

// ServiceConfigStore 服务配置存储接口
type ServiceConfigStore interface {
	// Get 获取配置
	Get(key string, defaultValue any) any
	// Set 设置配置
	Set(key string, value any)
}

// ServiceModuleManager 模块管理器接口
type ServiceModuleManager interface {
	// GetRunningTypeModules 获取指定类型的运行中模块
	GetRunningTypeModules(moduleType string) []ServiceModule
}

// ServiceModule 模块接口
type ServiceModule interface {
	// GetInstances 获取模块实例
	GetInstances() map[string]any
}

// ServiceSystemConfigKey 系统配置键
type ServiceSystemConfigKey string

// NewConfigHelper 创建配置帮助类实例
func NewConfigHelper(configStore ServiceConfigStore) *ConfigHelper {
	return &ConfigHelper{
		logger:      logger.GetLogger(),
		configStore: configStore,
	}
}

// GetConfigs 获取配置列表
func (h *ConfigHelper) GetConfigs(configKey ServiceSystemConfigKey) []map[string]any {
	// 从配置存储中获取配置
	configs := h.configStore.Get(string(configKey), []map[string]any{}).([]map[string]any)
	if configs == nil {
		return []map[string]any{}
	}

	return configs
}

// ServiceBaseHelper 通用服务帮助类
type ServiceBaseHelper struct {
	logger        *zap.Logger
	configKey     ServiceSystemConfigKey
	moduleType    string
	configHelper  *ConfigHelper
	moduleManager ServiceModuleManager
	mutex         sync.RWMutex
}

// NewServiceBaseHelper 创建通用服务帮助类实例
func NewServiceBaseHelper(configKey ServiceSystemConfigKey, moduleType string, configHelper *ConfigHelper, moduleManager ServiceModuleManager) *ServiceBaseHelper {
	return &ServiceBaseHelper{
		logger:        logger.GetLogger(),
		configKey:     configKey,
		moduleType:    moduleType,
		configHelper:  configHelper,
		moduleManager: moduleManager,
	}
}

// GetConfigs 获取配置列表
func (h *ServiceBaseHelper) GetConfigs(includeDisabled bool) map[string]map[string]any {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	configs := h.configHelper.GetConfigs(h.configKey)
	result := make(map[string]map[string]any)

	for _, config := range configs {
		name, ok := config["name"].(string)
		if !ok || name == "" {
			continue
		}

		enabled, _ := config["enabled"].(bool)
		typeStr, _ := config["type"].(string)

		if (enabled && typeStr != "") || includeDisabled {
			result[name] = config
		}
	}

	return result
}

// GetConfig 获取指定名称的配置
func (h *ServiceBaseHelper) GetConfig(name string) map[string]any {
	if name == "" {
		return nil
	}

	configs := h.GetConfigs(false)
	return configs[name]
}

// IterateModuleInstances 迭代模块实例
func (h *ServiceBaseHelper) IterateModuleInstances() []*ServiceServiceInfo {
	configs := h.GetConfigs(false)
	modules := h.moduleManager.GetRunningTypeModules(h.moduleType)

	var serviceInfos []*ServiceServiceInfo

	for _, module := range modules {
		if module == nil {
			continue
		}

		instances := module.GetInstances()
		if instances == nil {
			continue
		}

		for name, instance := range instances {
			if instance == nil {
				continue
			}

			config := configs[name]
			if config == nil {
				continue
			}

			typeStr, _ := config["type"].(string)

			serviceInfo := &ServiceServiceInfo{
				Name:     name,
				Instance: instance,
				Module:   module,
				Type:     typeStr,
				Config:   config,
			}

			serviceInfos = append(serviceInfos, serviceInfo)
		}
	}

	return serviceInfos
}

// GetServices 获取服务信息列表
func (h *ServiceBaseHelper) GetServices(typeFilter string, nameFilters []string) map[string]*ServiceServiceInfo {
	serviceInfos := h.IterateModuleInstances()
	result := make(map[string]*ServiceServiceInfo)

	// 构建名称过滤器集合
	nameFilterSet := make(map[string]bool)
	for _, name := range nameFilters {
		nameFilterSet[name] = true
	}

	for _, serviceInfo := range serviceInfos {
		// 名称过滤
		if len(nameFilterSet) > 0 && !nameFilterSet[serviceInfo.Name] {
			continue
		}

		// 类型过滤
		if typeFilter != "" && serviceInfo.Type != typeFilter {
			continue
		}

		result[serviceInfo.Name] = serviceInfo
	}

	return result
}

// GetService 获取指定名称的服务信息
func (h *ServiceBaseHelper) GetService(name, typeFilter string) *ServiceServiceInfo {
	if name == "" {
		return nil
	}

	serviceInfos := h.IterateModuleInstances()
	for _, serviceInfo := range serviceInfos {
		if serviceInfo.Name == name {
			if typeFilter == "" || serviceInfo.Type == typeFilter {
				return serviceInfo
			}
		}
	}

	return nil
}

// DownloaderConf 下载器配置
type DownloaderConf struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Enabled bool           `json:"enabled"`
	Config  map[string]any `json:"config"`
}

// GetDownloaderConfigs 获取下载器配置
func (h *ConfigHelper) GetDownloaderConfigs() []*DownloaderConf {
	return make([]*DownloaderConf, 0)
}

// GetMediaServerConfigs 获取媒体服务器配置
func (h *ConfigHelper) GetMediaServerConfigs() []*MediaServerConf {
	return make([]*MediaServerConf, 0)
}

// ServiceSystemConfigKey 系统配置键常量
const (
	ServiceSystemConfigKeyDownloaders         ServiceSystemConfigKey = "Downloaders"
	ServiceSystemConfigKeyNotifications       ServiceSystemConfigKey = "Notifications"
	ServiceSystemConfigKeyNotificationSwitchs ServiceSystemConfigKey = "NotificationSwitchs"
)

// ServiceModuleType 模块类型

type ServiceModuleType string

// ServiceModuleType 模块类型常量
const (
	ServiceModuleTypeDownloader   ServiceModuleType = "downloader"
	ServiceModuleTypeNotification ServiceModuleType = "notification"
)
