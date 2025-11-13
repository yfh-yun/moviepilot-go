package main

import (
	"fmt"
	
	"moviepilot-go/internal/helper"
)

func main() {
	fmt.Println("Storage Helper Example")
	
	// 创建存储帮助类实�?	storageHelper := helper.NewStorageHelper()
	
	if storageHelper == nil {
		fmt.Println("Failed to create StorageHelper")
		return
	}
	
	fmt.Println("StorageHelper created successfully")
	
	// 获取所有存储设�?	fmt.Println("\n=== 获取所有存储设�?===")
	storagies := storageHelper.GetStoragies()
	fmt.Printf("找到 %d 个存储设置\n", len(storagies))
	
	// 显示存储设置信息
	for i, storage := range storagies {
		fmt.Printf("  %d. 类型: %s, 名称: %s, 配置项数�? %d\n", 
			i+1, storage.Type, storage.Name, len(storage.Config))
		if i >= 4 { // 只显示前5�?			fmt.Println("  ...")
			break
		}
	}
	
	// 获取指定存储配置
	fmt.Println("\n=== 获取指定存储配置 ===")
	storage := storageHelper.GetStorage("non-existent-storage")
	if storage == nil {
		fmt.Println("未找到指定存储配�?)
	} else {
		fmt.Printf("找到存储配置: %s (%s)\n", storage.Name, storage.Type)
	}
	
	// 添加存储配置
	fmt.Println("\n=== 添加存储配置 ===")
	localConf := map[string]interface{}{
		"path": "/data/local",
		"readonly": false,
	}
	
	storageHelper.AddStorage("local", "本地存储", localConf)
	fmt.Println("已添加本地存储配�?)
	
	// 验证添加的存储配�?	localStorage := storageHelper.GetStorage("local")
	if localStorage != nil {
		fmt.Printf("本地存储配置: %s (%s)\n", localStorage.Name, localStorage.Type)
		fmt.Printf("  路径: %s\n", localStorage.Config["path"])
		fmt.Printf("  只读: %v\n", localStorage.Config["readonly"])
	}
	
	// 设置存储配置
	fmt.Println("\n=== 设置存储配置 ===")
	alipanConf := map[string]interface{}{
		"access_key": "test_access_key",
		"secret_key": "test_secret_key",
		"drive_id":   "test_drive_id",
	}
	
	storageHelper.SetStorage("alipan", alipanConf)
	fmt.Println("已设置阿里云盘存储配�?)
	
	// 验证设置的存储配�?	alipanStorage := storageHelper.GetStorage("alipan")
	if alipanStorage != nil {
		fmt.Printf("阿里云盘存储配置类型: %s\n", alipanStorage.Type)
		fmt.Printf("  Access Key: %s\n", alipanStorage.Config["access_key"])
		fmt.Printf("  Secret Key: %s\n", alipanStorage.Config["secret_key"])
		fmt.Printf("  Drive ID: %s\n", alipanStorage.Config["drive_id"])
	}
	
	// 重置存储配置
	fmt.Println("\n=== 重置存储配置 ===")
	storageHelper.ResetStorage("alipan")
	fmt.Println("已重置阿里云盘存储配�?)
	
	// 验证重置的存储配�?	resetStorage := storageHelper.GetStorage("alipan")
	if resetStorage != nil {
		fmt.Printf("重置后的阿里云盘存储配置类型: %s\n", resetStorage.Type)
		fmt.Printf("  配置项数�? %d\n", len(resetStorage.Config))
	}
	
	fmt.Println("\nExample completed")
}
