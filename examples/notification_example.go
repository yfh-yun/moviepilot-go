package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
)

func main() {
	// 创建NotificationHelper实例
	notificationHelper := helper.NewNotificationHelper()
	
	// 检查服务是否为指定类型
	serviceType := "wechat"
	serviceName := "my_wechat_service"
	
	isWechat := notificationHelper.IsNotification(&serviceType, nil, &serviceName)
	
	fmt.Printf("服务 %s 是否为微信通知类型: %t\n", serviceName, isWechat)
	
	// 检查另一个类�?	serviceType = "telegram"
	isTelegram := notificationHelper.IsNotification(&serviceType, nil, &serviceName)
	
	fmt.Printf("服务 %s 是否为Telegram通知类型: %t\n", serviceName, isTelegram)
}
