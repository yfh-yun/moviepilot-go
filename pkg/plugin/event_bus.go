package plugin

import (
	"fmt"
	"sync"
	"time"
)

// eventBus 事件总线实现
type eventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// NewEventBus 创建新的事件总线
func NewEventBus() EventBus {
	return &eventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe 订阅事件
func (eb *eventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = make([]EventHandler, 0)
	}

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Unsubscribe 取消订阅事件
func (eb *eventBus) Unsubscribe(eventType string, handler EventHandler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	handlers := eb.handlers[eventType]
	if handlers == nil {
		return
	}

	for i, h := range handlers {
		if h == handler {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish 发布事件
func (eb *eventBus) Publish(event Event) {
	eb.mutex.RLock()
	handlers := eb.handlers[event.Type]
	eb.mutex.RUnlock()

	if handlers == nil {
		return
	}

	// 异步处理事件
	go func() {
		for _, handler := range handlers {
			if err := handler.HandleEvent(event); err != nil {
				// 记录错误但不影响其他处理器
				fmt.Printf("Event handler error: %v\n", err)
			}
		}
	}()
}

// CreateEvent 创建事件
func CreateEvent(eventType, source string, data map[string]interface{}) Event {
	return Event{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      eventType,
		Source:    source,
		Data:      data,
		Timestamp: time.Now(),
	}
}