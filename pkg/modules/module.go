package modules

import (
	"moviepilot-go/pkg/models"
)

// Module 模块接口定义
type Module interface {
	// InitModule 模块初始�?	InitModule() error
	
	// InitSetting 模块开关设置，返回开关名和开关值，开关值为true时代表有值即打开，不实现该方法或返回nil代表不使用开�?	// 部分模块支持同时开启多个，此时设置项以,分隔，开关值使用contains判断
	InitSetting() (string, interface{})
	
	// GetName 获取模块名称
	GetName() string
	
	// GetType 获取模块类型
	GetType() models.ModuleType
	
	// GetSubType 获取模块子类型（下载器、媒体服务器、消息通道、存储类型、其他杂项模块类型）
	GetSubType() interface{}
	
	// GetPriority 获取模块优先级，数字越小优先级越高，只有同一接口下优先级才生�?	GetPriority() int
	
	// Stop 如果关闭时模块有服务需要停止，需要实现此方法
	Stop() error
	
	// Test 模块测试, 返回测试结果和错误信�?	Test() (bool, string)
}

// Service 服务接口定义
type Service interface {
	// InitService 初始化服�?	InitService(serviceName string, serviceType interface{}) error
	
	// GetInstances 获取服务实例列表
	GetInstances() map[string]interface{}
	
	// GetInstance 获取指定名称的服务实�?	GetInstance(name *string) interface{}
	
	// GetConfigs 获取已启用的服务配置字典
	GetConfigs() map[string]interface{}
	
	// GetConfig 获取指定名称的服务配�?	GetConfig(name *string) interface{}
	
	// GetDefaultConfigName 获取默认服务配置的名�?	GetDefaultConfigName() *string
}
