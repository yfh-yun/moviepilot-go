package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
	"moviepilot-go/internal/logger"
)

func main() {
	fmt.Println("Resource Helper Example")
	
	// 创建资源帮助类实�?	// 这会自动检查更�?	resourceHelper := helper.NewResourceHelper()
	
	if resourceHelper == nil {
		logger.Error("Failed to create ResourceHelper")
		return
	}
	
	fmt.Println("ResourceHelper created successfully")
	
	// 手动调用检查方�?	fmt.Println("Checking for resource updates...")
	err := resourceHelper.Check()
	if err != nil {
		fmt.Printf("Error checking for updates: %v\n", err)
	} else {
		fmt.Println("Resource check completed")
	}
	
	// 显示资源帮助类的基本信息
	fmt.Printf("Repository URL: %s\n", resourceHelper.repo)
	fmt.Printf("Files API URL: %s\n", resourceHelper.filesAPI)
	fmt.Printf("Base directory: %s\n", resourceHelper.baseDir)
	
	fmt.Println("Example completed")
}
