package core

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	TraceID   string                 `json:"trace_id"` // 用于追踪事件链路
}

// EventHandler 事件处理�?type EventHandler func(event *Event) error

// EventBus 事件总线
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewEventBus 创建新的事件总线
func NewEventBus(logger *zap.Logger) *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish 发布事件
func (eb *EventBus) Publish(eventType string, data map[string]interface{}) error {
	// 创建事件
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		TraceID:   uuid.New().String(), // 简化实现，实际应从上下文获�?	}
	
	// 获取处理�?	eb.mu.RLock()
	handlers, ok := eb.handlers[eventType]
	eb.mu.RUnlock()
	
	if !ok || len(handlers) == 0 {
		// 没有处理器订阅该事件
		return nil
	}
	
	// 调用所有处理器
	for _, handler := range handlers {
		if err := handler(event); err != nil {
			eb.logger.Error("处理事件失败",
				zap.String("event_type", eventType),
				zap.String("event_id", event.ID),
				zap.Error(err))
		}
	}
	
	return nil
}

// PublishWithTraceID 使用指定TraceID发布事件
func (eb *EventBus) PublishWithTraceID(eventType string, data map[string]interface{}, traceID string) error {
	// 创建事件
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		TraceID:   traceID,
	}
	
	// 获取处理�?	eb.mu.RLock()
	handlers, ok := eb.handlers[eventType]
	eb.mu.RUnlock()
	
	if !ok || len(handlers) == 0 {
		// 没有处理器订阅该事件
		return nil
	}
	
	// 调用所有处理器
	for _, handler := range handlers {
		if err := handler(event); err != nil {
			eb.logger.Error("处理事件失败",
				zap.String("event_type", eventType),
				zap.String("event_id", event.ID),
				zap.String("trace_id", traceID),
				zap.Error(err))
		}
	}
	
	return nil
}
