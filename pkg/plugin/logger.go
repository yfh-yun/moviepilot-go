package plugin

import (
	"fmt"
	"log"
	"os"
)

// pluginLogger 插件日志器实现
type pluginLogger struct {
	debug bool
}

// NewLogger 创建日志器
func NewLogger() Logger {
	debug := os.Getenv("PLUGIN_DEBUG") == "true"
	return &pluginLogger{debug: debug}
}

// Debug 调试日志
func (l *pluginLogger) Debug(msg string, args ...interface{}) {
	if l.debug {
		l.log("DEBUG", msg, args...)
	}
}

// Info 信息日志
func (l *pluginLogger) Info(msg string, args ...interface{}) {
	l.log("INFO", msg, args...)
}

// Warn 警告日志
func (l *pluginLogger) Warn(msg string, args ...interface{}) {
	l.log("WARN", msg, args...)
}

// Error 错误日志
func (l *pluginLogger) Error(msg string, args ...interface{}) {
	l.log("ERROR", msg, args...)
}

// Fatal 致命错误日志
func (l *pluginLogger) Fatal(msg string, args ...interface{}) {
	l.log("FATAL", msg, args...)
	os.Exit(1)
}

// log 内部日志方法
func (l *pluginLogger) log(level, msg string, args ...interface{}) {
	// 格式化参数
	formatted := msg
	if len(args) > 0 {
		formatted = fmt.Sprintf(msg, args...)
	}
	
	// 输出日志
	logMessage := fmt.Sprintf("[%s] %s", level, formatted)
	log.Println(logMessage)
}