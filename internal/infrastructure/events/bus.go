package events

import (
	"context"

	domains "moviepilot-go/internal/business/domains/events"
)

// HandlerFunc 事件处理器函数类型
type HandlerFunc func(ctx context.Context, event *domains.Event) error

// Bus 事件总线接口
type Bus interface {
	// 订阅广播事件，返回处理器ID
	SubscribeBroadcast(eventType domains.EventType, handler HandlerFunc) string

	// 订阅链式事件，返回处理器ID
	SubscribeChain(eventType domains.ChainEventType, priority int, handler HandlerFunc) string

	// 取消订阅广播事件
	UnsubscribeBroadcast(eventType domains.EventType, handler HandlerFunc)

	// 取消订阅链式事件
	UnsubscribeChain(eventType domains.ChainEventType, handler HandlerFunc)

	// 发布广播事件
	PublishBroadcast(ctx context.Context, eventType domains.EventType, data any, priority int) error

	// 分发链式事件
	DispatchChain(ctx context.Context, eventType domains.ChainEventType, data any, priority int) (*domains.Event, error)

	// 禁用处理器
	DisableHandler(handlerID string)

	// 启用处理器
	EnableHandler(handlerID string)

	// 列出所有处理器信息
	ListHandlers() []HandlerInfo
}

// HandlerInfo 处理器信息，用于可视化
type HandlerInfo struct {
	EventType string // 事件类型
	HandlerID string // 处理器ID
	Priority  *int   // 优先级，仅链式事件有值
	Status    string // 状态：enabled / disabled
}
