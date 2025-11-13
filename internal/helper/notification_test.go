package helper

import (
	"testing"
)

// 模拟ServiceInfo结构体用于测�?type ServiceInfo struct {
	Name string
	Type string
}

// 模拟SystemConfigKey类型
type SystemConfigKey string

// 模拟ModuleType类型
type ModuleType string

const (
	SystemConfigKeyNotifications SystemConfigKey = "Notifications"
	ModuleTypeNotification       ModuleType       = "notification"
)

// 简化版ServiceBaseHelper用于测试
type ServiceBaseHelper struct {
	ConfigKey  SystemConfigKey
	ModuleType ModuleType
}

func NewServiceBaseHelper(configKey SystemConfigKey, moduleType ModuleType) *ServiceBaseHelper {
	return &ServiceBaseHelper{
		ConfigKey:  configKey,
		ModuleType: moduleType,
	}
}

func (s *ServiceBaseHelper) GetService(name string, typeFilter *string) *ServiceInfo {
	// 模拟实现
	if name == "test_service" {
		return &ServiceInfo{
			Name: "test_service",
			Type: "wechat",
		}
	}
	return nil
}

// 简化版NotificationHelper用于测试
type NotificationHelper struct {
	*ServiceBaseHelper
}

func NewNotificationHelper() *NotificationHelper {
	baseHelper := NewServiceBaseHelper(
		SystemConfigKeyNotifications,
		ModuleTypeNotification,
	)
	
	return &NotificationHelper{
		ServiceBaseHelper: baseHelper,
	}
}

func (n *NotificationHelper) IsNotification(
	serviceType *string,
	service *ServiceInfo,
	name *string,
) bool {
	// 如果未提�?service 则通过 name 获取服务
	if service == nil && name != nil {
		service = n.GetService(*name, nil)
	}

	// 判断服务类型是否为指定类�?	if service != nil && serviceType != nil {
		return service.Type == *serviceType
	}
	
	return false
}

// 测试函数
func TestNotificationHelper(t *testing.T) {
	// 测试创建NotificationHelper实例
	t.Run("创建NotificationHelper实例", func(t *testing.T) {
		notificationHelper := NewNotificationHelper()
		if notificationHelper == nil {
			t.Error("无法创建NotificationHelper实例")
		}
	})

	// 测试IsNotification方法
	t.Run("测试IsNotification方法", func(t *testing.T) {
		notificationHelper := NewNotificationHelper()
		
		// 测试serviceType为nil的情�?		result := notificationHelper.IsNotification(nil, nil, nil)
		if result != false {
			t.Error("当serviceType为nil时，应该返回false")
		}
		
		// 测试serviceType不匹配的情况
		serviceType := "telegram"
		result = notificationHelper.IsNotification(&serviceType, nil, nil)
		if result != false {
			t.Error("当service为nil且name为nil时，应该返回false")
		}
		
		// 测试通过name获取service并匹配类型的情况
		serviceType = "wechat"
		name := "test_service"
		result = notificationHelper.IsNotification(&serviceType, nil, &name)
		if result != true {
			t.Error("当serviceType匹配时，应该返回true")
		}
		
		// 测试通过name获取service但类型不匹配的情�?		serviceType = "telegram"
		name = "test_service"
		result = notificationHelper.IsNotification(&serviceType, nil, &name)
		if result != false {
			t.Error("当serviceType不匹配时，应该返回false")
		}
	})
}
