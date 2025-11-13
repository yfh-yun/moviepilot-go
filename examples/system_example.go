package main

import (
	"fmt"
	"time"

	"moviepilot-go/internal/helper"
	"moviepilot-go/internal/utils"
)

func main() {
	fmt.Println("System Helper Example")
	
	// 创建系统帮助类实�?
	systemHelper := helper.NewSystemHelper()
	
	if systemHelper == nil {
		fmt.Println("Failed to create SystemHelper")
		return
	}
	
	fmt.Println("SystemHelper created successfully")
	
	// 显示系统帮助类的基本信息
	fmt.Printf("System flag file: %s\n", systemHelper.systemFlagFile)
	
	// 测试处理配置变更事件
	fmt.Println("\n=== 处理配置变更事件 ===")
	eventData := map[string]interface{}{
		"key": "DEBUG",
	}
	systemHelper.HandleConfigChanged(eventData)
	fmt.Println("配置变更事件处理完成")
	
	// 测试判断是否可以内部重启
	fmt.Println("\n=== 判断是否可以内部重启 ===")
	canRestart := systemHelper.CanRestart()
	fmt.Printf("可以内部重启: %v\n", canRestart)
	
	// 测试获取当前容器ID
	fmt.Println("\n=== 获取当前容器ID ===")
	containerID := systemHelper.getContainerID()
	if containerID != nil {
		fmt.Printf("容器ID: %s\n", *containerID)
	} else {
		fmt.Println("未获取到容器ID")
	}
	
	// 测试检查当前容器是否配置了自动重启策略
	fmt.Println("\n=== 检查当前容器是否配置了自动重启策略 ===")
	hasPolicy := systemHelper.checkRestartPolicy()
	fmt.Printf("容器配置了自动重启策�? %v\n", hasPolicy)
	
	// 测试设置系统已修改标�?
	fmt.Println("\n=== 设置系统已修改标�?===")
	systemHelper.SetSystemModified()
	fmt.Println("系统已修改标志设置完�?)
	
	// 测试检查系统是否已被重�?
	fmt.Println("\n=== 检查系统是否已被重�?===")
	isReset := systemHelper.IsSystemReset()
	fmt.Printf("系统已被重置: %v\n", isReset)
	
	// 测试设置信号处理�?
	fmt.Println("\n=== 设置信号处理�?===")
	systemHelper.SetupSignalHandler()
	fmt.Println("信号处理器设置完�?)
	
	// 演示系统工具类的其他功能
	fmt.Println("\n=== 系统工具类功能演�?===")
	sysUtils := utils.NewSystemUtils()
	
	fmt.Printf("是否为Docker环境: %v\n", sysUtils.IsDocker())
	fmt.Printf("是否为Windows系统: %v\n", sysUtils.IsWindows())
	fmt.Printf("是否为MacOS系统: %v\n", sysUtils.IsMacos())
	fmt.Printf("是否为ARM64架构: %v\n", sysUtils.IsAarch64())
	fmt.Printf("是否为AMD64架构: %v\n", sysUtils.IsX8664())
	fmt.Printf("系统平台: %s\n", sysUtils.Platform())
	fmt.Printf("CPU架构: %s\n", sysUtils.CPUArch())
	
	// 等待一段时间以观察信号处理
	fmt.Println("\n等待5秒以观察信号处理...")
	time.Sleep(5 * time.Second)
	
	fmt.Println("\nExample completed")
}
