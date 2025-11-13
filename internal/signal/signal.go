package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	
	"moviepilot-go/internal/logger"
)

// SetupSignalHandler 设置信号处理器，用于优雅关闭服务
func SetupSignalHandler() {
	loggerManager := logger.GetLoggerManager()
	
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	
	// 注册要监听的信号
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	
	// 启动goroutine处理信号
	go func() {
		sig := <-sigChan
		loggerManager.Info(fmt.Sprintf("收到信号 %v，开始优雅停止服�?..", sig))
		
		// 在这里可以执行清理操�?		// 例如关闭数据库连接、停止定时任务等
		
		// 退出程�?		os.Exit(0)
	}()
}
