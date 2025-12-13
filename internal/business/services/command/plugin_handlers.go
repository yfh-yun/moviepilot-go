package command

import (
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// PluginHandlerBase 插件命令处理器基类
type PluginHandlerBase struct {
	logger   *zap.Logger
	name     string
	desc     string
	category string
	eventType string
	data     map[string]interface{}
	show     bool
	pid      string
}

// NewPluginHandlerBase 创建插件命令处理器基类
func NewPluginHandlerBase(
	name string,
	desc string,
	category string,
	eventType string,
	data map[string]interface{},
	show bool,
	pid string,
) *PluginHandlerBase {
	return &PluginHandlerBase{
		logger:   logger.GetLogger(),
		name:     name,
		desc:     desc,
		category: category,
		eventType: eventType,
		data:     data,
		show:     show,
		pid:      pid,
	}
}

// Name 命令名称
func (h *PluginHandlerBase) Name() string {
	return h.name
}

// Description 命令描述
func (h *PluginHandlerBase) Description() string {
	return h.desc
}

// Category 命令分类
func (h *PluginHandlerBase) Category() string {
	return h.category
}

// EventType 事件类型
func (h *PluginHandlerBase) EventType() string {
	return h.eventType
}

// Data 事件数据
func (h *PluginHandlerBase) Data() map[string]interface{} {
	return h.data
}

// Show 是否显示
func (h *PluginHandlerBase) Show() bool {
	return h.show
}

// PID 插件ID
func (h *PluginHandlerBase) PID() string {
	return h.pid
}
