package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/db"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/signal"
	"moviepilot-go/internal/tray"
	"moviepilot-go/internal/web"
)

func main() {
	fmt.Println("MoviePilot-Go - 自动化媒体管理工�?)
	fmt.Println("项目启动�?..")
	
	// 初始化配�?	appConfig := config.GetConfig()
	
	// 初始化日�?	logManager := logger.GetLoggerManager()
	
	// 设置进程�?	// Go中没有直接设置进程名的方法，但可以通过其他方式实现
	
	// 注册信号处理器，用于优雅关闭服务
	signal.SetupSignalHandler()
	
	// 启动托盘图标
	tray.StartTray()
	
	// 初始化数据库
	if err := db.InitDB(); err != nil {
		logManager.Error(fmt.Sprintf("初始化数据库失败: %v", err))
		return
	}
	
	// 更新数据�?	if err := db.UpdateDB(); err != nil {
		logManager.Error(fmt.Sprintf("更新数据库失�? %v", err))
		return
	}
	
	// 创建生命周期管理�?	lifespan := web.NewLifespan()
	
	// 启动应用
	if err := lifespan.Startup(context.Background()); err != nil {
		logManager.Error(fmt.Sprintf("启动应用失败: %v", err))
		return
	}
	
	// 初始化Web服务�?	server := web.NewServer()
	
	// 启动Web服务器（在goroutine中）
	go func() {
		if err := server.Start(); err != nil {
			logManager.Error(fmt.Sprintf("启动Web服务器失�? %v", err))
		}
	}()
	
	fmt.Printf("MoviePilot-Go 启动成功！监听端�? %d\n", appConfig.Port)
	fmt.Printf("项目名称: %s\n", appConfig.ProjectName)
	fmt.Printf("API版本: %s\n", appConfig.APIVersion)
	
	// 等待中断信号以优雅地关闭服务�?	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	// 关闭服务�?	if err := server.Stop(); err != nil {
		logManager.Error(fmt.Sprintf("关闭Web服务器失�? %v", err))
	}
	
	// 关闭应用
	if err := lifespan.Shutdown(context.Background()); err != nil {
		logManager.Error(fmt.Sprintf("关闭应用失败: %v", err))
	}
	
	fmt.Println("MoviePilot-Go 已停�?)
}
