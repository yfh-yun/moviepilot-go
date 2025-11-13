package plugins

import (
	"moviepilot-go/pkg/models"
)

// PluginChain 插件处理�?type PluginChain struct {
	// 可以添加处理链相关的字段
}

// NewPluginChain 创建一个新的插件处理链实例
func NewPluginChain() *PluginChain {
	return &PluginChain{}
}

// PostMessage 发送消�?func (pc *PluginChain) PostMessage(notification *models.Notification) {
	// 实现发送消息的逻辑
	// 这里应该调用消息处理系统
}
