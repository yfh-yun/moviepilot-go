package helper

import (
	"moviepilot-go/internal/db"
	"moviepilot-go/pkg/models"
)

// ServiceConfigHelper 配置帮助类，获取不同类型的服务配�?type ServiceConfigHelper struct{}

// NewServiceConfigHelper 创建ServiceConfigHelper实例
func NewServiceConfigHelper() *ServiceConfigHelper {
	return &ServiceConfigHelper{}
}

// GetConfigs 通用获取配置的方法，根据 configKey 获取相应的配置并返回指定类型的配置列�?func (s *ServiceConfigHelper) GetConfigs(configKey models.SystemConfigKey) []map[string]interface{} {
	/*
	 * 通用获取配置的方法，根据 config_key 获取相应的配置并返回指定类型的配置列�?	 * :param configKey: 系统配置�?key
	 * :return: 配置对象列表
	 */
	configData := db.NewSystemConfigOper().Get(configKey)
	if configData == nil {
		return []map[string]interface{}{}
	}
	
	// 类型断言获取配置数据
	if configs, ok := configData.([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(configs))
		for _, config := range configs {
			if configMap, ok := config.(map[string]interface{}); ok {
				result = append(result, configMap)
			}
		}
		return result
	}
	
	return []map[string]interface{}{}
}

// GetDownloaderConfigs 获取下载器的配置
func (s *ServiceConfigHelper) GetDownloaderConfigs() []models.DownloaderConf {
	/*
	 * 获取下载器的配置
	 */
	configs := s.GetConfigs(models.SystemConfigKeyDownloaders)
	result := make([]models.DownloaderConf, 0, len(configs))
	
	for _, config := range configs {
		downloaderConf := models.DownloaderConf{}
		
		if name, exists := config["name"]; exists {
			if nameStr, ok := name.(string); ok {
				downloaderConf.Name = nameStr
			}
		}
		
		if configType, exists := config["type"]; exists {
			if typeStr, ok := configType.(string); ok {
				downloaderConf.Type = typeStr
			}
		}
		
		if defaultVal, exists := config["default"]; exists {
			if defaultBool, ok := defaultVal.(bool); ok {
				downloaderConf.Default = defaultBool
			}
		}
		
		if enabled, exists := config["enabled"]; exists {
			if enabledBool, ok := enabled.(bool); ok {
				downloaderConf.Enabled = enabledBool
			}
		}
		
		if configData, exists := config["config"]; exists {
			if configMap, ok := configData.(map[string]interface{}); ok {
				downloaderConf.Config = configMap
			}
		}
		
		result = append(result, downloaderConf)
	}
	
	return result
}

// GetMediaserverConfigs 获取媒体服务器的配置
func (s *ServiceConfigHelper) GetMediaserverConfigs() []models.MediaServerConf {
	/*
	 * 获取媒体服务器的配置
	 */
	configs := s.GetConfigs(models.SystemConfigKeyMediaServers)
	result := make([]models.MediaServerConf, 0, len(configs))
	
	for _, config := range configs {
		mediaserverConf := models.MediaServerConf{}
		
		if name, exists := config["name"]; exists {
			if nameStr, ok := name.(string); ok {
				mediaserverConf.Name = nameStr
			}
		}
		
		if configType, exists := config["type"]; exists {
			if typeStr, ok := configType.(string); ok {
				mediaserverConf.Type = typeStr
			}
		}
		
		if enabled, exists := config["enabled"]; exists {
			if enabledBool, ok := enabled.(bool); ok {
				mediaserverConf.Enabled = enabledBool
			}
		}
		
		if configData, exists := config["config"]; exists {
			if configMap, ok := configData.(map[string]interface{}); ok {
				mediaserverConf.Config = configMap
			}
		}
		
		if syncLibraries, exists := config["sync_libraries"]; exists {
			if syncLibs, ok := syncLibraries.([]interface{}); ok {
				mediaserverConf.SyncLibraries = syncLibs
			}
		}
		
		result = append(result, mediaserverConf)
	}
	
	return result
}

// GetNotificationConfigs 获取消息通知渠道的配�?func (s *ServiceConfigHelper) GetNotificationConfigs() []models.NotificationConf {
	/*
	 * 获取消息通知渠道的配�?	 */
	configs := s.GetConfigs(models.SystemConfigKeyNotifications)
	result := make([]models.NotificationConf, 0, len(configs))
	
	for _, config := range configs {
		notificationConf := models.NotificationConf{}
		
		if name, exists := config["name"]; exists {
			if nameStr, ok := name.(string); ok {
				notificationConf.Name = nameStr
			}
		}
		
		if configType, exists := config["type"]; exists {
			if typeStr, ok := configType.(string); ok {
				notificationConf.Type = typeStr
			}
		}
		
		if enabled, exists := config["enabled"]; exists {
			if enabledBool, ok := enabled.(bool); ok {
				notificationConf.Enabled = enabledBool
			}
		}
		
		if configData, exists := config["config"]; exists {
			if configMap, ok := configData.(map[string]interface{}); ok {
				notificationConf.Config = configMap
			}
		}
		
		if switchs, exists := config["switchs"]; exists {
			if switchList, ok := switchs.([]interface{}); ok {
				notificationConf.Switchs = switchList
			}
		}
		
		result = append(result, notificationConf)
	}
	
	return result
}

// GetNotificationSwitches 获取消息通知场景的开�?func (s *ServiceConfigHelper) GetNotificationSwitches() []models.NotificationSwitchConf {
	/*
	 * 获取消息通知场景的开�?	 */
	configs := s.GetConfigs(models.SystemConfigKeyNotificationSwitchs)
	result := make([]models.NotificationSwitchConf, 0, len(configs))
	
	for _, config := range configs {
		switchConf := models.NotificationSwitchConf{}
		
		if switchType, exists := config["type"]; exists {
			if typeStr, ok := switchType.(string); ok {
				switchConf.Type = typeStr
			}
		}
		
		if action, exists := config["action"]; exists {
			if actionStr, ok := action.(string); ok {
				switchConf.Action = actionStr
			}
		}
		
		result = append(result, switchConf)
	}
	
	return result
}

// GetNotificationSwitch 获取指定类型的消息通知场景的开�?func (s *ServiceConfigHelper) GetNotificationSwitch(mtype string) *string {
	/*
	 * 获取指定类型的消息通知场景的开�?	 */
	switches := s.GetNotificationSwitches()
	for _, switchConf := range switches {
		if switchConf.Type == mtype {
			return &switchConf.Action
		}
	}
	return nil
}

// ServiceBaseHelper 通用服务帮助类，抽象获取配置和服务实例的通用逻辑
type ServiceBaseHelper struct {
	ModuleManager *ModuleHelper
	ConfigKey     models.SystemConfigKey
	ConfType      interface{}
	ModuleType    models.ModuleType
}

// NewServiceBaseHelper 创建ServiceBaseHelper实例
func NewServiceBaseHelper(configKey models.SystemConfigKey, confType interface{}, moduleType models.ModuleType) *ServiceBaseHelper {
	return &ServiceBaseHelper{
		ModuleManager: NewModuleHelper(),
		ConfigKey:     configKey,
		ConfType:      confType,
		ModuleType:    moduleType,
	}
}

// GetConfigs 获取配置列表
func (s *ServiceBaseHelper) GetConfigs(includeDisabled bool) map[string]interface{} {
	/*
	 * 获取配置列表
	 * :param includeDisabled: 是否包含禁用的配置，默认 False（仅返回启用的配置）
	 * :return: 配置字典
	 */
	configs := NewServiceConfigHelper().GetConfigs(s.ConfigKey)
	result := make(map[string]interface{})
	
	for _, config := range configs {
		// 提取配置名称
		var name string
		if nameVal, exists := config["name"]; exists {
			if nameStr, ok := nameVal.(string); ok {
				name = nameStr
			}
		}
		
		// 提取配置类型
		var configType string
		if typeVal, exists := config["type"]; exists {
			if typeStr, ok := typeVal.(string); ok {
				configType = typeStr
			}
		}
		
		// 提取启用状�?		enabled := true
		if enabledVal, exists := config["enabled"]; exists {
			if enabledBool, ok := enabledVal.(bool); ok {
				enabled = enabledBool
			}
		}
		
		// 根据条件过滤配置
		if (name != "" && configType != "" && enabled) || includeDisabled {
			result[name] = config
		}
	}
	
	return result
}

// GetConfig 获取指定名称配置
func (s *ServiceBaseHelper) GetConfig(name string) interface{} {
	/*
	 * 获取指定名称配置
	 */
	if name == "" {
		return nil
	}
	
	configs := s.GetConfigs(false)
	if config, exists := configs[name]; exists {
		return config
	}
	
	return nil
}

// IterateModuleInstances 迭代所有模块的实例及其对应的配置，返回 ServiceInfo 实例
func (s *ServiceBaseHelper) IterateModuleInstances() []*models.ServiceInfo {
	/*
	 * 迭代所有模块的实例及其对应的配置，返回 ServiceInfo 实例
	 */
	configs := s.GetConfigs(false)
	result := make([]*models.ServiceInfo, 0)
	
	// TODO: 实现模块管理器获取运行类型模块的逻辑
	// 由于模块系统尚未完全实现，此处暂时返回空列表
	/*
	modules := s.ModuleManager.GetRunningTypeModules(s.ModuleType)
	for _, module := range modules {
		if module == nil {
			continue
		}
		
		moduleInstances := module.GetInstances()
		if moduleInstances == nil {
			continue
		}
		
		for name, instance := range moduleInstances {
			if instance == nil {
				continue
			}
			
			config := configs[name]
			serviceInfo := models.ServiceInfo{
				Name:     name,
				Instance: instance,
				Module:   module,
				Type:     "", // 需要从config中提取类�?				Config:   config,
			}
			
			// 如果config存在且有type字段，则设置serviceInfo.Type
			if config != nil {
				if configMap, ok := config.(map[string]interface{}); ok {
					if typeVal, exists := configMap["type"]; exists {
						if typeStr, ok := typeVal.(string); ok {
							serviceInfo.Type = typeStr
						}
					}
				}
			}
			
			result = append(result, &serviceInfo)
		}
	}
	*/
	
	return result
}

// GetServices 获取服务信息列表，并根据类型和名称列表进行过�?func (s *ServiceBaseHelper) GetServices(typeFilter *string, nameFilters []string) map[string]*models.ServiceInfo {
	/*
	 * 获取服务信息列表，并根据类型和名称列表进行过�?	 * :param typeFilter: 需要过滤的服务类型
	 * :param nameFilters: 需要过滤的服务名称列表
	 * :return: 过滤后的服务信息字典
	 */
	// 创建名称过滤器集�?	nameFiltersSet := make(map[string]bool)
	if nameFilters != nil {
		for _, name := range nameFilters {
			nameFiltersSet[name] = true
		}
	}
	
	result := make(map[string]*models.ServiceInfo)
	
	// 获取服务信息
	serviceInfos := s.IterateModuleInstances()
	for _, serviceInfo := range serviceInfos {
		// 检查配置是否存�?		if serviceInfo.Config == nil {
			continue
		}
		
		// 检查类型过滤器
		if typeFilter != nil && serviceInfo.Type != *typeFilter {
			continue
		}
		
		// 检查名称过滤器
		if nameFilters != nil && !nameFiltersSet[serviceInfo.Name] {
			continue
		}
		
		result[serviceInfo.Name] = serviceInfo
	}
	
	return result
}

// GetService 获取指定名称的服务信息，并根据类型过�?func (s *ServiceBaseHelper) GetService(name string, typeFilter *string) *models.ServiceInfo {
	/*
	 * 获取指定名称的服务信息，并根据类型过�?	 * :param name: 服务名称
	 * :param typeFilter: 需要过滤的服务类型
	 * :return: 对应的服务信息，若不存在或类型不匹配则返�?None
	 */
	if name == "" {
		return nil
	}
	
	// 获取服务信息
	serviceInfos := s.IterateModuleInstances()
	for _, serviceInfo := range serviceInfos {
		if serviceInfo.Name == name {
			// 检查配置是否存�?			if serviceInfo.Config != nil {
				// 检查类型过滤器
				if typeFilter == nil || serviceInfo.Type == *typeFilter {
					return serviceInfo
				}
			}
		}
	}
	
	return nil
}
