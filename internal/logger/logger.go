package logger

import (
	"go.uber.org/zap"
)

// Logger 日志记录器包�?type Logger struct {
	*zap.Logger
	manager *LoggerManager
}

// NewLogger 创建新的日志记录�?func NewLogger() (*Logger, error) {
	manager := GetLoggerManager()
	
	// 创建默认logger
	logger := manager.GetLogger("moviepilot")
	
	return &Logger{
		Logger:  logger,
		manager: manager,
	}, nil
}

// GetLogger 获取指定名称的logger
func (l *Logger) GetLogger(name string) *Logger {
	zapLogger := l.manager.GetLogger(name)
	return &Logger{
		Logger:  zapLogger,
		manager: l.manager,
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info 记录信息日志
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn 记录警告日志
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error 记录错误日志
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal 记录致命错误日志
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// Sync 同步日志缓冲�?func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
