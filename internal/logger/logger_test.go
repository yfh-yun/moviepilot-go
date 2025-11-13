package logger

import (
	"testing"
	"time"
)

// TestLogger 测试日志功能
func TestLogger(t *testing.T) {
	// 获取日志管理�?	manager := GetLoggerManager()
	
	// 测试各种级别的日�?	manager.Info("这是一条信息日�?)
	manager.Debug("这是一条调试日�?)
	manager.Warning("这是一条警告日�?)
	manager.Error("这是一条错误日�?)
	
	// 测试带参数的日志
	manager.Info("用户 %s 登录系统，IP地址: %s", "张三", "192.168.1.100")
	
	// 获取特定名称的logger
	logger := manager.GetLogger("test")
	logger.Info("使用特定logger记录日志")
	
	// 等待日志写入完成
	time.Sleep(100 * time.Millisecond)
	
	// 更新日志配置
	settings := GetLogSettings()
	settings.LogLevel = "DEBUG"
	manager.UpdateLoggers()
	
	// 测试更新后的日志级别
	manager.Debug("更新日志级别后的一条调试日�?)
	
	// 关闭日志系统
	manager.Shutdown()
	
	t.Log("日志测试完成")
}
