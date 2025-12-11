package events

import (
	"time"

	"github.com/google/uuid"
)

// EventKind 事件类型枚举
type EventKind int

const (
	EventKindBroadcast EventKind = iota // 广播事件
	EventKindChain                      // 链式事件
)

// EventType 广播事件类型
type EventType string

// ChainEventType 链式事件类型
type ChainEventType string

// Event 统一事件对象
type Event struct {
	ID        string    // 事件唯一标识
	Kind      EventKind // 事件类型
	Type      any       // EventType 或 ChainEventType
	Data      any       // 事件数据
	Priority  int       // 事件优先级
	CreatedAt time.Time // 事件创建时间
}

// NewBroadcastEvent 创建广播事件
func NewBroadcastEvent(t EventType, data any, priority int) *Event {
	return &Event{
		ID:        uuid.NewString(),
		Kind:      EventKindBroadcast,
		Type:      t,
		Data:      data,
		Priority:  priority,
		CreatedAt: time.Now(),
	}
}

// NewChainEvent 创建链式事件
func NewChainEvent(t ChainEventType, data any, priority int) *Event {
	return &Event{
		ID:        uuid.NewString(),
		Kind:      EventKindChain,
		Type:      t,
		Data:      data,
		Priority:  priority,
		CreatedAt: time.Now(),
	}
}
