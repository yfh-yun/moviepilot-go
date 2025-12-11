package events

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"

	domains "moviepilot-go/internal/business/domains/events"
	"moviepilot-go/pkg/logger"
)

// subscriber 广播事件订阅者
type subscriber struct {
	id      string
	handler HandlerFunc
}

// chainSubscriber 链式事件订阅者
type chainSubscriber struct {
	id       string
	priority int
	handler  HandlerFunc
}

// bus 事件总线实现
type bus struct {
	logger *zap.Logger

	// 广播事件订阅者映射
	broadcastSubs map[domains.EventType][]subscriber

	// 链式事件订阅者映射
	chainSubs map[domains.ChainEventType][]chainSubscriber

	// 禁用的处理器ID集合
	disabledHandlers map[string]struct{}

	// 禁用的处理器类集合
	disabledClasses map[string]struct{}

	// 广播事件队列
	broadcastQueue chan *domains.Event

	// 停止信号通道
	stopCh chan struct{}

	// 读写锁，保护共享资源
	mu sync.RWMutex
}

// NewBus 创建事件总线实例
func NewBus() Bus {
	b := &bus{
		logger:           logger.GetLogger(),
		broadcastSubs:    make(map[domains.EventType][]subscriber),
		chainSubs:        make(map[domains.ChainEventType][]chainSubscriber),
		disabledHandlers: make(map[string]struct{}),
		disabledClasses:  make(map[string]struct{}),
		broadcastQueue:   make(chan *domains.Event, 1024),
		stopCh:           make(chan struct{}),
	}

	// 启动广播事件消费者协程
	go b.runBroadcastConsumer()

	return b
}

// SubscribeBroadcast 订阅广播事件，返回处理器ID
func (b *bus) SubscribeBroadcast(eventType domains.EventType, handler HandlerFunc) string {
	id := handlerID(handler)
	b.mu.Lock()
	defer b.mu.Unlock()

	b.broadcastSubs[eventType] = append(b.broadcastSubs[eventType], subscriber{
		id:      id,
		handler: handler,
	})

	b.logger.Debug("subscribed to broadcast event", zap.String("event_type", string(eventType)), zap.String("handler_id", id))
	return id
}

// SubscribeChain 订阅链式事件，返回处理器ID
func (b *bus) SubscribeChain(eventType domains.ChainEventType, priority int, handler HandlerFunc) string {
	id := handlerID(handler)
	b.mu.Lock()
	defer b.mu.Unlock()

	// 添加订阅者
	b.chainSubs[eventType] = append(b.chainSubs[eventType], chainSubscriber{
		id:       id,
		priority: priority,
		handler:  handler,
	})

	// 按优先级排序（升序，优先级值越小越先执行）
	sort.Slice(b.chainSubs[eventType], func(i, j int) bool {
		return b.chainSubs[eventType][i].priority < b.chainSubs[eventType][j].priority
	})

	b.logger.Debug("subscribed to chain event", zap.String("event_type", string(eventType)), zap.String("handler_id", id), zap.Int("priority", priority))
	return id
}

// UnsubscribeBroadcast 取消订阅广播事件
func (b *bus) UnsubscribeBroadcast(eventType domains.EventType, handler HandlerFunc) {
	id := handlerID(handler)
	b.mu.Lock()
	defer b.mu.Unlock()

	// 查找并移除订阅者
	if subs, exists := b.broadcastSubs[eventType]; exists {
		for i, sub := range subs {
			if sub.id == id {
				b.broadcastSubs[eventType] = append(subs[:i], subs[i+1:]...)
				b.logger.Debug("unsubscribed from broadcast event", zap.String("event_type", string(eventType)), zap.String("handler_id", id))
				break
			}
		}

		// 如果没有订阅者了，删除该事件类型
		if len(b.broadcastSubs[eventType]) == 0 {
			delete(b.broadcastSubs, eventType)
		}
	}
}

// UnsubscribeChain 取消订阅链式事件
func (b *bus) UnsubscribeChain(eventType domains.ChainEventType, handler HandlerFunc) {
	id := handlerID(handler)
	b.mu.Lock()
	defer b.mu.Unlock()

	// 查找并移除订阅者
	if subs, exists := b.chainSubs[eventType]; exists {
		for i, sub := range subs {
			if sub.id == id {
				b.chainSubs[eventType] = append(subs[:i], subs[i+1:]...)
				b.logger.Debug("unsubscribed from chain event", zap.String("event_type", string(eventType)), zap.String("handler_id", id))
				break
			}
		}

		// 如果没有订阅者了，删除该事件类型
		if len(b.chainSubs[eventType]) == 0 {
			delete(b.chainSubs, eventType)
		}
	}
}

// PublishBroadcast 发布广播事件
func (b *bus) PublishBroadcast(ctx context.Context, eventType domains.EventType, data any, priority int) error {
	ev := domains.NewBroadcastEvent(eventType, data, priority)

	select {
	case b.broadcastQueue <- ev:
		b.logger.Debug("published broadcast event", zap.String("event_type", string(eventType)), zap.String("event_id", ev.ID))
		return nil
	default:
		b.logger.Warn("broadcast queue full, dropping event", zap.String("event_type", string(eventType)))
		return fmt.Errorf("broadcast queue full")
	}
}

// DispatchChain 分发链式事件
func (b *bus) DispatchChain(ctx context.Context, eventType domains.ChainEventType, data any, priority int) (*domains.Event, error) {
	ev := domains.NewChainEvent(eventType, data, priority)

	b.mu.RLock()
	subs := append([]chainSubscriber(nil), b.chainSubs[eventType]...)
	b.mu.RUnlock()

	if len(subs) == 0 {
		b.logger.Debug("no handlers for chain event", zap.String("event_type", string(eventType)))
		return nil, nil
	}

	b.logger.Debug("chain event started", zap.String("event_type", string(eventType)), zap.String("event_id", ev.ID))

	for _, s := range subs {
		if !b.isHandlerEnabled(s.id) {
			b.logger.Debug("handler disabled, skipping", zap.String("handler_id", s.id))
			continue
		}

		start := time.Now()
		if err := b.safeInvokeHandler(ctx, s.handler, ev); err != nil {
			b.handleEventError(ev, s.id, err)
		}
		b.logger.Debug("chain handler completed",
			zap.String("handler_id", s.id),
			zap.Int("priority", s.priority),
			zap.Duration("duration", time.Since(start)))
	}

	b.logger.Debug("chain event completed", zap.String("event_type", string(eventType)), zap.String("event_id", ev.ID))
	return ev, nil
}

// DisableHandler 禁用处理器
func (b *bus) DisableHandler(handlerID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.disabledHandlers[handlerID] = struct{}{}
	b.logger.Debug("disabled handler", zap.String("handler_id", handlerID))
}

// EnableHandler 启用处理器
func (b *bus) EnableHandler(handlerID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.disabledHandlers, handlerID)
	b.logger.Debug("enabled handler", zap.String("handler_id", handlerID))
}

// ListHandlers 列出所有处理器信息
func (b *bus) ListHandlers() []HandlerInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var handlers []HandlerInfo

	// 添加广播事件处理器
	for eventType, subs := range b.broadcastSubs {
		for _, sub := range subs {
			status := "enabled"
			if _, disabled := b.disabledHandlers[sub.id]; disabled {
				status = "disabled"
			}

			handlers = append(handlers, HandlerInfo{
				EventType: string(eventType),
				HandlerID: sub.id,
				Priority:  nil,
				Status:    status,
			})
		}
	}

	// 添加链式事件处理器
	for eventType, subs := range b.chainSubs {
		for _, sub := range subs {
			status := "enabled"
			if _, disabled := b.disabledHandlers[sub.id]; disabled {
				status = "disabled"
			}

			priority := sub.priority
			handlers = append(handlers, HandlerInfo{
				EventType: string(eventType),
				HandlerID: sub.id,
				Priority:  &priority,
				Status:    status,
			})
		}
	}

	return handlers
}

// runBroadcastConsumer 运行广播事件消费者协程
func (b *bus) runBroadcastConsumer() {
	for {
		select {
		case ev := <-b.broadcastQueue:
			if ev == nil {
				continue
			}
			b.dispatchBroadcast(context.Background(), ev)
		case <-b.stopCh:
			b.logger.Info("broadcast consumer stopped")
			return
		}
	}
}

// dispatchBroadcast 分发广播事件
func (b *bus) dispatchBroadcast(ctx context.Context, ev *domains.Event) {
	b.mu.RLock()
	subs := append([]subscriber(nil), b.broadcastSubs[ev.Type.(domains.EventType)]...)
	b.mu.RUnlock()

	if len(subs) == 0 {
		b.logger.Debug("no handlers for broadcast event", zap.String("event_type", string(ev.Type.(domains.EventType))))
		return
	}

	b.logger.Debug("dispatching broadcast event", zap.String("event_type", string(ev.Type.(domains.EventType))), zap.String("event_id", ev.ID), zap.Int("handler_count", len(subs)))

	// 每个处理器单独协程执行
	for _, s := range subs {
		if !b.isHandlerEnabled(s.id) {
			b.logger.Debug("handler disabled, skipping", zap.String("handler_id", s.id))
			continue
		}

		go func(sub subscriber) {
			start := time.Now()
			if err := b.safeInvokeHandler(ctx, sub.handler, ev); err != nil {
				b.handleEventError(ev, sub.id, err)
			}
			b.logger.Debug("broadcast handler completed",
				zap.String("handler_id", sub.id),
				zap.Duration("duration", time.Since(start)))
		}(s)
	}
}

// safeInvokeHandler 安全调用处理器，捕获并处理错误
func (b *bus) safeInvokeHandler(ctx context.Context, handler HandlerFunc, ev *domains.Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stackBuf := make([]byte, 4096)
			stackLen := runtime.Stack(stackBuf, false)
			err = fmt.Errorf("handler panic: %v\n%s", r, stackBuf[:stackLen])
		}
	}()

	return handler(ctx, ev)
}

// handleEventError 处理事件处理过程中的错误
func (b *bus) handleEventError(ev *domains.Event, handlerID string, err error) {
	b.logger.Error("event handler error",
		zap.String("event_id", ev.ID),
		zap.String("handler_id", handlerID),
		zap.Error(err))

	// 这里可以添加更多错误处理逻辑，如发送系统错误事件等
}

// isHandlerEnabled 检查处理器是否启用
func (b *bus) isHandlerEnabled(handlerID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 检查处理器是否直接被禁用
	if _, disabled := b.disabledHandlers[handlerID]; disabled {
		return false
	}

	// 检查处理器类是否被禁用
	// 从处理器ID中提取类名（如果有）
	// 这里简化实现，实际可以根据需要提取类名
	return true
}

// handlerID 获取处理器ID
func handlerID(handler HandlerFunc) string {
	pc := reflect.ValueOf(handler).Pointer()
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// handlerIDFromTarget 从目标对象获取处理器ID
func handlerIDFromTarget(target any) string {
	switch t := target.(type) {
	case HandlerFunc:
		return handlerID(t)
	default:
		// 对于其他类型，使用其类型名称
		return reflect.TypeOf(target).String()
	}
}
