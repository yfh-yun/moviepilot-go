package bootstrap

import (
	"moviepilot-go/pkg/logger"
)

// initLogger 初始化日志系统
func initLogger() error {
	return logger.Init()
}
