package modules

import (
	"moviepilot-go/pkg/models"
)

// ServiceBase 抽象服务基类，负责服务的初始化、获取实例和配置管理
type ServiceBase struct {
	// 服务配置映射
	Configs map[string]interface{}
	// 服务实例映射
	Instances map[string]interface{}
	// 服务名称
	ServiceName string
}

// NewServiceBase 创建一个新的服务基类实�?func NewServiceBase() *ServiceBase {
	return &ServiceBase{
		Configs:   make(map[string]interface{}),
		Instances: make(map[string]interface{}),
	}
}

// InitService 初始化服务，获取配置并实例化对应服务
func (s *ServiceBase) InitService(serviceName string, serviceType interface{}) error {
	if serviceName == "" {
		return &ServiceError{"service_name is null"}
	}
	
	s.ServiceName = serviceName
	configs := s.GetConfigs()
	if configs == nil {
		return nil
	}
	
	s.Configs = configs
	s.Instances = make(map[string]interface{})
	
	if serviceType == nil {
		return nil
	}
	
	// 注意：Go中无法像Python那样动态实例化类型，这里需要在具体实现中处�?	// 子类需要根据具体的配置类型来实现实例化逻辑
	
	return nil
}

// GetInstances 获取服务实例列表
func (s *ServiceBase) GetInstances() map[string]interface{} {
	if s.Instances == nil {
		return make(map[string]interface{})
	}
	return s.Instances
}

// GetInstance 获取指定名称的服务实�?func (s *ServiceBase) GetInstance(name *string) interface{} {
	if s.Instances == nil {
		return nil
	}
	
	if name != nil && *name != "" {
		if instance, exists := s.Instances[*name]; exists {
			return instance
		}
		return nil
	}
	
	// 获取默认实例
	defaultName := s.GetDefaultConfigName()
	if defaultName != nil && *defaultName != "" {
		if instance, exists := s.Instances[*defaultName]; exists {
			return instance
		}
	}
	
	return nil
}

// GetConfigs 获取已启用的服务配置字典
func (s *ServiceBase) GetConfigs() map[string]interface{} {
	// 子类需要实现此方法
	return nil
}

// GetConfig 获取指定名称的服务配�?func (s *ServiceBase) GetConfig(name *string) interface{} {
	if s.Configs == nil {
		return nil
	}
	
	if name != nil && *name != "" {
		if config, exists := s.Configs[*name]; exists {
			return config
		}
		return nil
	}
	
	// 获取默认配置
	defaultName := s.GetDefaultConfigName()
	if defaultName != nil && *defaultName != "" {
		if config, exists := s.Configs[*defaultName]; exists {
			return config
		}
	}
	
	return nil
}

// GetDefaultConfigName 获取默认服务配置的名�?func (s *ServiceBase) GetDefaultConfigName() *string {
	// 默认使用第一个配置的名称
	if s.Configs == nil || len(s.Configs) == 0 {
		return nil
	}
	
	// 获取第一个配置的名称
	for name := range s.Configs {
		return &name
	}
	
	return nil
}

// ServiceError 服务错误类型
type ServiceError struct {
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}
