package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Event 事件
// 原Python: Event in app/core/event.py
type Event struct {
	ID       string         `json:"id"`       // 事件ID
	Type     string         `json:"type"`     // 事件类型
	Data     map[string]any `json:"data"`     // 事件数据
	Priority int            `json:"priority"` // 优先级（数字越小优先级越高）
	Time     time.Time      `json:"time"`     // 事件时间
}

// NewEvent 创建事件
func NewEvent(eventType string, data map[string]any) *Event {
	return &Event{
		ID:       uuid.New().String(),
		Type:     eventType,
		Data:     data,
		Priority: 10, // 默认优先级
		Time:     time.Now(),
	}
}

// NewEventWithPriority 创建带优先级的事件
func NewEventWithPriority(eventType string, data map[string]any, priority int) *Event {
	return &Event{
		ID:       uuid.New().String(),
		Type:     eventType,
		Data:     data,
		Priority: priority,
		Time:     time.Now(),
	}
}

// Handler 事件处理器函数类型
type Handler func(event *Event) error

// handlerInfo 处理器信息
type handlerInfo struct {
	ID      string
	Handler Handler
}

// Manager 事件管理器
// 原Python: EventManager in app/core/event.py
type Manager struct {
	subscribers map[string][]handlerInfo // 事件类型 -> 处理器列表
	queue       chan *Event              // 事件队列
	mu          sync.RWMutex             // 读写锁
	logger      *zap.Logger              // 日志
	ctx         context.Context          // 上下文
	cancel      context.CancelFunc       // 取消函数
	wg          sync.WaitGroup           // 等待组
}

// NewManager 创建事件管理器
func NewManager(logger *zap.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		subscribers: make(map[string][]handlerInfo),
		queue:       make(chan *Event, 1000), // 缓冲1000个事件
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
	}

	// 启动事件处理goroutine
	m.start()

	return m
}

// SendEvent 发送事件
// 原Python: send_event(event_type, data)
func (m *Manager) SendEvent(eventType string, data map[string]any) error {
	event := NewEvent(eventType, data)

	select {
	case m.queue <- event:
		m.logger.Debug("事件已发送",
			zap.String("event_id", event.ID),
			zap.String("event_type", event.Type))
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("发送事件超时")
	}
}

// SendEventWithPriority 发送带优先级的事件
func (m *Manager) SendEventWithPriority(eventType string, data map[string]any, priority int) error {
	event := NewEventWithPriority(eventType, data, priority)

	select {
	case m.queue <- event:
		m.logger.Debug("事件已发送",
			zap.String("event_id", event.ID),
			zap.String("event_type", event.Type),
			zap.Int("priority", priority))
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("发送事件超时")
	}
}

// Subscribe 订阅事件
// 原Python: subscribe(event_type, handler)
// 返回处理器ID，用于取消订阅
func (m *Manager) Subscribe(eventType string, handler Handler) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	handlerID := uuid.New().String()

	info := handlerInfo{
		ID:      handlerID,
		Handler: handler,
	}

	m.subscribers[eventType] = append(m.subscribers[eventType], info)

	m.logger.Debug("订阅事件",
		zap.String("event_type", eventType),
		zap.String("handler_id", handlerID))

	return handlerID
}

// Unsubscribe 取消订阅
// 原Python: unsubscribe(event_type, handler_id)
func (m *Manager) Unsubscribe(eventType string, handlerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	handlers, exists := m.subscribers[eventType]
	if !exists {
		return fmt.Errorf("事件类型不存在: %s", eventType)
	}

	// 查找并删除处理器
	for i, info := range handlers {
		if info.ID == handlerID {
			m.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			m.logger.Debug("取消订阅",
				zap.String("event_type", eventType),
				zap.String("handler_id", handlerID))
			return nil
		}
	}

	return fmt.Errorf("处理器不存在: %s", handlerID)
}

// start 启动事件处理
func (m *Manager) start() {
	m.wg.Add(1)
	go m.processEvents()
}

// processEvents 处理事件
func (m *Manager) processEvents() {
	defer m.wg.Done()

	for {
		select {
		case event := <-m.queue:
			m.handleEvent(event)
		case <-m.ctx.Done():
			m.logger.Info("事件管理器已停止")
			return
		}
	}
}

// handleEvent 处理单个事件
func (m *Manager) handleEvent(event *Event) {
	m.mu.RLock()
	handlers, exists := m.subscribers[event.Type]
	m.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		m.logger.Debug("没有订阅者",
			zap.String("event_type", event.Type))
		return
	}

	m.logger.Debug("处理事件",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.Type),
		zap.Int("handlers", len(handlers)))

	// 并发执行所有处理器
	var wg sync.WaitGroup
	for _, info := range handlers {
		wg.Add(1)
		go func(h handlerInfo) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					m.logger.Error("事件处理器panic",
						zap.String("event_type", event.Type),
						zap.String("handler_id", h.ID),
						zap.Any("panic", r))
				}
			}()

			if err := h.Handler(event); err != nil {
				m.logger.Error("事件处理失败",
					zap.String("event_type", event.Type),
					zap.String("handler_id", h.ID),
					zap.Error(err))
			}
		}(info)
	}

	wg.Wait()
}

// Stop 停止事件管理器
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
	close(m.queue)
	m.logger.Info("事件管理器已关闭")
}

// GetSubscriberCount 获取订阅者数量
func (m *Manager) GetSubscriberCount(eventType string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.subscribers[eventType])
}
