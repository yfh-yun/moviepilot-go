package service

import (
	"moviepilot-go/internal/config"
	"moviepilot-go/pkg/models"
)

// ServiceBaseHelper 服务基类帮助�?type ServiceBaseHelper struct {
	ConfigKey  string
	ConfType   interface{}
	ModuleType string
	config     *config.AppConfig
}

// NewServiceBaseHelper 创建服务基类帮助类实�?func NewServiceBaseHelper() *ServiceBaseHelper {
	return &ServiceBaseHelper{
		config: config.GetConfig(),
	}
}

// GetService 获取服务信息
func (sbh *ServiceBaseHelper) GetService(name string) *models.ServiceInfo {
	// TODO: 实现获取服务的逻辑
	// 这需要根据具体的配置和模块类型来实现
	return nil
}

// GetServices 获取所有服务信�?func (sbh *ServiceBaseHelper) GetServices() []*models.ServiceInfo {
	// TODO: 实现获取所有服务的逻辑
	return []*models.ServiceInfo{}
}

// GetConfig 获取配置信息
func (sbh *ServiceBaseHelper) GetConfig() *config.AppConfig {
	return sbh.config
}
