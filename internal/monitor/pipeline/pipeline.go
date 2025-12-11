package pipeline

import (
	"moviepilot-go/internal/monitor/filewatch"
)

// EventPipeline 事件处理管道
type EventPipeline struct {
	// 这里可以添加各种处理器，如刮削处理器、整理处理器、入库处理器等
	// 实际实现中需要根据业务需求进行扩展
}

// NewEventPipeline 创建事件处理管道
func NewEventPipeline() *EventPipeline {
	return &EventPipeline{}
}

// HandleEvent 处理文件系统事件
func (ep *EventPipeline) HandleEvent(event filewatch.Event) {
	// 实际实现中需要根据事件类型和路径进行业务逻辑处理
	// 例如：
	// - 检查文件是否为媒体文件
	// - 触发刮削流程
	// - 触发整理流程
	// - 触发入库流程
	// - 触发通知流程

	// 这里简化处理，仅打印事件信息
	// 在实际实现中，应该将事件转换为领域事件，通过事件系统或工作流引擎触发后续处理
}

// RegisterHandler 注册事件处理器
// 实际实现中可以根据需要扩展，支持注册不同类型的处理器
func (ep *EventPipeline) RegisterHandler(handler any) {
	// 实际实现中需要根据handler类型进行注册
}
