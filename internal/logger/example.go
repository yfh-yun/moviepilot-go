package logger

import (
	"time"
)

// Example 使用示例
func Example() {
	// 获取日志管理�?	manager := GetLoggerManager()
	
	// 记录不同级别的日�?	manager.Info("这是一条信息日�?)
	manager.Debug("这是一条调试日�?)
	manager.Warning("这是一条警告日�?)
	manager.Error("这是一条错误日�?)
	
	// 获取特定名称的logger
	logger := manager.GetLogger("test")
	logger.Info("使用特定logger记录日志")
	
	// 带参数的日志
	manager.Info("用户 %s 登录系统，IP地址: %s", "张三", "192.168.1.100")
	
	// 模拟一些处�?	time.Sleep(100 * time.Millisecond)
	
	// 关闭日志系统
	manager.Shutdown()
}
