package main

import (
	"fmt"
	
	"moviepilot-go/internal/helper"
	"moviepilot-go/pkg/models"
)

func main() {
	fmt.Println("Service Helper Example")
	
	// 创建服务配置帮助类实�?	serviceConfigHelper := helper.NewServiceConfigHelper()
	
	if serviceConfigHelper == nil {
		fmt.Println("Failed to create ServiceConfigHelper")
		return
	}
	
	fmt.Println("ServiceConfigHelper created successfully")
	
	// 获取下载器配�?	fmt.Println("\n=== 获取下载器配�?===")
	downloaderConfigs := serviceConfigHelper.GetDownloaderConfigs()
	fmt.Printf("找到 %d 个下载器配置\n", len(downloaderConfigs))
	
	// 显示下载器配置信�?	for i, config := range downloaderConfigs {
		fmt.Printf("  %d. 名称: %s, 类型: %s, 默认: %v, 启用: %v\n", 
			i+1, config.Name, config.Type, config.Default, config.Enabled)
		if i >= 4 { // 只显示前5�?			fmt.Println("  ...")
			break
		}
	}
	
	// 获取媒体服务器配�?	fmt.Println("\n=== 获取媒体服务器配�?===")
	mediaserverConfigs := serviceConfigHelper.GetMediaserverConfigs()
	fmt.Printf("找到 %d 个媒体服务器配置\n", len(mediaserverConfigs))
	
	// 显示媒体服务器配置信�?	for i, config := range mediaserverConfigs {
		fmt.Printf("  %d. 名称: %s, 类型: %s, 启用: %v\n", 
			i+1, config.Name, config.Type, config.Enabled)
		if i >= 4 { // 只显示前5�?			fmt.Println("  ...")
			break
		}
	}
	
	// 获取通知配置
	fmt.Println("\n=== 获取通知配置 ===")
	notificationConfigs := serviceConfigHelper.GetNotificationConfigs()
	fmt.Printf("找到 %d 个通知配置\n", len(notificationConfigs))
	
	// 显示通知配置信息
	for i, config := range notificationConfigs {
		fmt.Printf("  %d. 名称: %s, 类型: %s, 启用: %v\n", 
			i+1, config.Name, config.Type, config.Enabled)
		if i >= 4 { // 只显示前5�?			fmt.Println("  ...")
			break
		}
	}
	
	// 获取通知开关配�?	fmt.Println("\n=== 获取通知开关配�?===")
	notificationSwitches := serviceConfigHelper.GetNotificationSwitches()
	fmt.Printf("找到 %d 个通知开关配置\n", len(notificationSwitches))
	
	// 显示通知开关配置信�?	for i, switchConf := range notificationSwitches {
		fmt.Printf("  %d. 类型: %s, 动作: %s\n", 
			i+1, switchConf.Type, switchConf.Action)
		if i >= 4 { // 只显示前5�?			fmt.Println("  ...")
			break
		}
	}
	
	// 获取特定类型的通知开�?	fmt.Println("\n=== 获取特定类型的通知开�?===")
	switchAction := serviceConfigHelper.GetNotificationSwitch("download.start")
	if switchAction == nil {
		fmt.Println("未找到指定类型的通知开�?)
	} else {
		fmt.Printf("找到通知开关，动作: %s\n", *switchAction)
	}
	
	// 创建服务基础帮助类实�?	fmt.Println("\n=== 创建服务基础帮助�?===")
	serviceBaseHelper := helper.NewServiceBaseHelper(
		models.SystemConfigKeyDownloaders,
		models.DownloaderConf{},
		models.ModuleTypeDownload,
	)
	
	if serviceBaseHelper == nil {
		fmt.Println("Failed to create ServiceBaseHelper")
		return
	}
	
	fmt.Println("ServiceBaseHelper created successfully")
	
	// 获取配置
	fmt.Println("\n=== 获取配置 ===")
	configs := serviceBaseHelper.GetConfigs(false)
	fmt.Printf("找到 %d 个配置\n", len(configs))
	
	// 获取特定配置
	fmt.Println("\n=== 获取特定配置 ===")
	config := serviceBaseHelper.GetConfig("non-existent-config")
	if config == nil {
		fmt.Println("未找到指定配�?)
	} else {
		fmt.Printf("找到配置: %v\n", config)
	}
	
	// 获取服务
	fmt.Println("\n=== 获取服务 ===")
	services := serviceBaseHelper.GetServices(nil, nil)
	fmt.Printf("找到 %d 个服务\n", len(services))
	
	// 获取特定服务
	fmt.Println("\n=== 获取特定服务 ===")
	service := serviceBaseHelper.GetService("non-existent-service", nil)
	if service == nil {
		fmt.Println("未找到指定服�?)
	} else {
		fmt.Printf("找到服务: %s\n", service.Name)
	}
	
	fmt.Println("\nExample completed")
}
