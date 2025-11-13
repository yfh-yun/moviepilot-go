package command

import (
	"moviepilot-go/internal/logger"
)

// CommandChain 命令链基�?type CommandChain struct {
	commandManager *CommandManager
}

// NewCommandChain 创建命令链实�?func NewCommandChain(cm *CommandManager) *CommandChain {
	return &CommandChain{
		commandManager: cm,
	}
}

// RegisterCommands 注册命令
func (cc *CommandChain) RegisterCommands(commands map[string]*Command) {
	logger.Log.Debugf("注册�?%d 个命�?, len(commands))
	// 在实际实现中，这里会通知其他组件有关命令变更的信�?}

// PostMessage 发送消息通知
func (cc *CommandChain) PostMessage(channel, source, title string, userID interface{}) {
	logger.Log.Infof("发送消息通知: %s (渠道: %s, 来源: %s, 用户: %v)", title, channel, source, userID)
	// 在实际实现中，这里会发送消息到指定渠道
}
