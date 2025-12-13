package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// eventManager 事件管理器实现
type eventManager struct {
	mu            sync.RWMutex               // 读写锁，保证线程安全
	subscriptions map[string]*EventSubscription // 订阅信息，key为订阅ID
	eventSubscriptions map[EventType]map[string]*EventSubscription // 按事件类型分组的订阅信息
	closed        bool                       // 是否已关闭
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewEventManager 创建事件管理器
func NewEventManager(ctx context.Context) EventManager {
	ctx, cancel := context.WithCancel(ctx)
	return &eventManager{
		subscriptions:     make(map[string]*EventSubscription),
		eventSubscriptions: make(map[EventType]map[string]*EventSubscription),
		ctx:               ctx,
		cancel:            cancel,
	}
}

// SubscribeEvent 订阅事件
func (em *eventManager) SubscribeEvent(eventType EventType, handler EventHandler, filter EventFilter) (string, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.closed {
		return "", fmt.Errorf("event manager is closed")
	}

	// 生成唯一订阅ID
	subscriptionID := uuid.New().String()

	// 创建订阅
	subscription := &EventSubscription{
		ID:        subscriptionID,
		EventType: eventType,
		Handler:   handler,
		Filter:    filter,
	}

	// 存储订阅
	em.subscriptions[subscriptionID] = subscription

	// 按事件类型分组存储
	if _, ok := em.eventSubscriptions[eventType]; !ok {
		em.eventSubscriptions[eventType] = make(map[string]*EventSubscription)
	}
	em.eventSubscriptions[eventType][subscriptionID] = subscription

	return subscriptionID, nil
}

// UnsubscribeEvent 取消订阅
func (em *eventManager) UnsubscribeEvent(subscriptionID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.closed {
		return fmt.Errorf("event manager is closed")
	}

	// 获取订阅
	subscription, ok := em.subscriptions[subscriptionID]
	if !ok {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	// 从按事件类型分组的存储中删除
	if eventSubs, ok := em.eventSubscriptions[subscription.EventType]; ok {
		delete(eventSubs, subscriptionID)
		// 如果该事件类型没有订阅了，删除该事件类型的存储
		if len(eventSubs) == 0 {
			delete(em.eventSubscriptions, subscription.EventType)
		}
	}

	// 从主存储中删除
	delete(em.subscriptions, subscriptionID)

	return nil
}

// SubscribeMultipleEvents 订阅多个事件
func (em *eventManager) SubscribeMultipleEvents(eventTypes []EventType, handler EventHandler, filter EventFilter) ([]string, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.closed {
		return nil, fmt.Errorf("event manager is closed")
	}

	subscriptionIDs := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		// 生成唯一订阅ID
		subscriptionID := uuid.New().String()

		// 创建订阅
		subscription := &EventSubscription{
			ID:        subscriptionID,
			EventType: eventType,
			Handler:   handler,
			Filter:    filter,
		}

		// 存储订阅
		em.subscriptions[subscriptionID] = subscription

		// 按事件类型分组存储
		if _, ok := em.eventSubscriptions[eventType]; !ok {
			em.eventSubscriptions[eventType] = make(map[string]*EventSubscription)
		}
		em.eventSubscriptions[eventType][subscriptionID] = subscription

		subscriptionIDs = append(subscriptionIDs, subscriptionID)
	}

	return subscriptionIDs, nil
}

// UnsubscribeAllEvents 取消所有订阅
func (em *eventManager) UnsubscribeAllEvents() error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.closed {
		return fmt.Errorf("event manager is closed")
	}

	// 清空所有订阅
	em.subscriptions = make(map[string]*EventSubscription)
	em.eventSubscriptions = make(map[EventType]map[string]*EventSubscription)

	return nil
}

// PublishEvent 发布事件（同步）
func (em *eventManager) PublishEvent(ctx context.Context, event *Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// 确保事件有ID和时间戳
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 获取该事件类型的所有订阅
	em.mu.RLock()
	if em.closed {
		em.mu.RUnlock()
		return fmt.Errorf("event manager is closed")
	}

	// 复制订阅列表，避免在处理事件时修改订阅列表导致死锁
	eventSubs := make([]*EventSubscription, 0)
	if subs, ok := em.eventSubscriptions[event.Type]; ok {
		for _, sub := range subs {
			eventSubs = append(eventSubs, sub)
		}
	}
	em.mu.RUnlock()

	// 处理所有匹配的订阅
	for _, sub := range eventSubs {
		// 检查过滤器
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}

		// 执行事件处理函数
		err := sub.Handler(ctx, event)
		if err != nil {
			// 记录错误，但不中断其他处理器
			fmt.Printf("error handling event %s: %v\n", event.ID, err)
		}
	}

	return nil
}

// PublishEventAsync 发布事件（异步）
func (em *eventManager) PublishEventAsync(event *Event) {
	if event == nil {
		return
	}

	// 确保事件有ID和时间戳
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 获取该事件类型的所有订阅
	em.mu.RLock()
	if em.closed {
		em.mu.RUnlock()
		return
	}

	// 复制订阅列表，避免在处理事件时修改订阅列表导致死锁
	eventSubs := make([]*EventSubscription, 0)
	if subs, ok := em.eventSubscriptions[event.Type]; ok {
		for _, sub := range subs {
			eventSubs = append(eventSubs, sub)
		}
	}
	em.mu.RUnlock()

	// 使用goroutine异步处理所有匹配的订阅
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()

		// 创建一个带有超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, sub := range eventSubs {
			// 检查过滤器
			if sub.Filter != nil && !sub.Filter(event) {
				continue
			}

			// 执行事件处理函数
			err := sub.Handler(ctx, event)
			if err != nil {
				// 记录错误，但不中断其他处理器
				fmt.Printf("error handling event %s: %v\n", event.ID, err)
			}
		}
	}()
}

// GetSubscriptions 获取所有订阅
func (em *eventManager) GetSubscriptions() []*EventSubscription {
	em.mu.RLock()
	defer em.mu.RUnlock()

	subscriptions := make([]*EventSubscription, 0, len(em.subscriptions))
	for _, sub := range em.subscriptions {
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions
}

// GetSubscriptionsByEventType 获取指定事件类型的订阅
func (em *eventManager) GetSubscriptionsByEventType(eventType EventType) []*EventSubscription {
	em.mu.RLock()
	defer em.mu.RUnlock()

	subscriptions := make([]*EventSubscription, 0)
	if subs, ok := em.eventSubscriptions[eventType]; ok {
		for _, sub := range subs {
			subscriptions = append(subscriptions, sub)
		}
	}

	return subscriptions
}

// Close 关闭事件管理器
func (em *eventManager) Close() error {
	em.mu.Lock()
	if em.closed {
		em.mu.Unlock()
		return nil
	}

	em.closed = true
	em.mu.Unlock()

	// 取消上下文
	em.cancel()

	// 等待所有异步事件处理完成
	em.wg.Wait()

	// 清空所有订阅
	em.mu.Lock()
	defer em.mu.Unlock()
	em.subscriptions = nil
	em.eventSubscriptions = nil

	return nil
}
