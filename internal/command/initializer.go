package command

import (
	"moviepilot-go/internal/core"
	"moviepilot-go/internal/scheduler"
)

// InitCommand 初始化命�?func InitCommand() *CommandManager {
	// 创建并返回一个新的命令管理器实例
	s := scheduler.GetScheduler()
	eb := core.NewEventBus(nil) // 在实际使用中应该传入真实的logger实例
	return NewCommandManager(s, eb)
}

// StopCommand 停止命令
func StopCommand() {
	// 空实现，与Python版本保持一�?}

// RestartCommand 重启命令
func RestartCommand() {
	// 创建命令管理器实例并初始化命�?	s := scheduler.GetScheduler()
	eb := core.NewEventBus(nil) // 在实际使用中应该传入真实的logger实例
	cm := NewCommandManager(s, eb)
	cm.InitCommands("")
}
