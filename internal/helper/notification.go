package helper

import (
	"moviepilot-go/pkg/models"
)

// NotificationHelper 消息通知帮助�?type NotificationHelper struct {
	*ServiceBaseHelper
}

// NewNotificationHelper 创建NotificationHelper实例
func NewNotificationHelper() *NotificationHelper {
	baseHelper := NewServiceBaseHelper(
		models.SystemConfigKeyNotifications,
		models.NotificationConf{}, // conf_type
		models.ModuleTypeNotification, // module_type
	)
	
	return &NotificationHelper{
		ServiceBaseHelper: baseHelper,
	}
}

// IsNotification 通用的消息通知服务类型判断方法
func (n *NotificationHelper) IsNotification(
	serviceType *string,
	service *models.ServiceInfo,
	name *string,
) bool {
	/*
	 * 通用的消息通知服务类型判断方法
	 * :param serviceType: 消息通知服务的类型名称（�?'wechat', 'voicechat', 'telegram', 等）
	 * :param service: 要判断的服务信息
	 * :param name: 服务的名�?	 * :return: 如果服务类型或实例为指定类型，返�?True；否则返�?False
	 */
	
	// 如果未提�?service 则通过 name 获取服务
	if service == nil && name != nil {
		service = n.GetService(*name, nil)
	}

	// 判断服务类型是否为指定类�?	if service != nil && serviceType != nil {
		return service.Type == *serviceType
	}
	
	return false
}
