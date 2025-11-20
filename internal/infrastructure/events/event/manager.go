// Package event 事件管理器
package event

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventManager 事件管理器
type EventManager struct {
	subscriptions map[string][]*EventSubscription
	queue        *EventQueue
	statistics   *EventStatistics
	retryPolicy  *RetryPolicy
	
	// 配置
	workerCount    int
	batchSize      int
	batchTimeout   time.Duration
	stopTimeout    time.Duration
	
	// 控制
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	mutex      sync.RWMutex
	
	// 错误处理
	errorHandler func(event *Event, err error)
}

// EventManagerConfig 事件管理器配置
type EventManagerConfig struct {
	QueueSize     int           `json:"queue_size"`
	WorkerCount   int           `json:"worker_count"`
	BatchSize     int           `json:"batch_size"`
	BatchTimeout  time.Duration `json:"batch_timeout"`
	StopTimeout   time.Duration `json:"stop_timeout"`
	RetryPolicy   *RetryPolicy  `json:"retry_policy"`
}

// DefaultEventManagerConfig 默认配置
func DefaultEventManagerConfig() *EventManagerConfig {
	return &EventManagerConfig{
		QueueSize:    1000,
		WorkerCount:  runtime.NumCPU(),
		BatchSize:    10,
		BatchTimeout: 100 * time.Millisecond,
		StopTimeout:  5 * time.Second,
		RetryPolicy:  DefaultRetryPolicy(),
	}
}

// NewEventManager 创建事件管理器
func NewEventManager(config *EventManagerConfig) *EventManager {
	if config == nil {
		config = DefaultEventManagerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &EventManager{
		subscriptions: make(map[string][]*EventSubscription),
		queue:        NewEventQueue(config.QueueSize),
		statistics:   NewEventStatistics(),
		retryPolicy:  config.RetryPolicy,
		workerCount:  config.WorkerCount,
		batchSize:    config.BatchSize,
		batchTimeout: config.BatchTimeout,
		stopTimeout:  config.StopTimeout,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start 启动事件管理器
func (em *EventManager) Start() error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if em.running {
		return fmt.Errorf("event manager is already running")
	}

	em.running = true

	// 启动工作协程
	for i := 0; i < em.workerCount; i++ {
		em.wg.Add(1)
		go em.worker(i)
	}

	// 启动统计协程
	em.wg.Add(1)
	go em.statisticsUpdater()

	return nil
}

// Stop 停止事件管理器
func (em *EventManager) Stop() error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if !em.running {
		return nil
	}

	em.cancel()
	em.running = false

	// 等待所有工作协程完成
	done := make(chan struct{})
	go func() {
		em.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(em.stopTimeout):
		return fmt.Errorf("timeout waiting for event manager to stop")
	}
}

// Publish 发布事件
func (em *EventManager) Publish(event *Event) error {
	if !em.running {
		return fmt.Errorf("event manager is not running")
	}

	em.statistics.IncrementTotal()

	// 设置事件上下文
	if event.Context == nil {
		event.Context = em.ctx
	}

	// 推入队列
	if err := em.queue.Push(event); err != nil {
		em.statistics.IncrementFailed()
		return fmt.Errorf("failed to push event to queue: %w", err)
	}

	return nil
}

// Subscribe 订阅事件
func (em *EventManager) Subscribe(eventType EventType, filter EventFilter, handler EventHandler) (string, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	subscription := NewEventSubscription(filter, handler)
	
	eventTypeStr := string(eventType)
	em.subscriptions[eventTypeStr] = append(em.subscriptions[eventTypeStr], subscription)

	return subscription.ID, nil
}

// SubscribeAll 订阅所有事件
func (em *EventManager) SubscribeAll(filter EventFilter, handler EventHandler) (string, error) {
	return em.Subscribe("*", filter, handler)
}

// Unsubscribe 取消订阅
func (em *EventManager) Unsubscribe(subscriptionID string) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	for eventType, subscriptions := range em.subscriptions {
		for i, sub := range subscriptions {
			if sub.ID == subscriptionID {
				// 标记为非活跃
				sub.Active = false
				// 从列表中移除
				em.subscriptions[eventType] = append(subscriptions[:i], subscriptions[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("subscription not found: %s", subscriptionID)
}

// GetStatistics 获取统计信息
func (em *EventManager) GetStatistics() map[string]interface{} {
	em.statistics.UpdateQueueSize(em.queue.Size())
	return em.statistics.GetSnapshot()
}

// SetErrorHandler 设置错误处理器
func (em *EventManager) SetErrorHandler(handler func(event *Event, err error)) {
	em.errorHandler = handler
}

// worker 工作协程
func (em *EventManager) worker(id int) {
	defer em.wg.Done()

	for {
		select {
		case <-em.ctx.Done():
			return
		default:
			em.processBatch()
		}
	}
}

// processBatch 处理事件批次
func (em *EventManager) processBatch() {
	batch := make([]*Event, 0, em.batchSize)
	
	// 收集批次
	timeout := time.After(em.batchTimeout)
	
	for len(batch) < em.batchSize {
		select {
		case <-timeout:
			if len(batch) > 0 {
				em.processEvents(batch)
			}
			return
		case <-em.ctx.Done():
			return
		default:
			if event, ok := em.queue.Pop(); ok {
				batch = append(batch, event)
			} else {
				// 队列为空，等待一下
				time.Sleep(10 * time.Millisecond)
				if len(batch) > 0 {
					em.processEvents(batch)
				}
				return
			}
		}
	}

	// 处理批次
	if len(batch) > 0 {
		em.processEvents(batch)
	}
}

// processEvents 处理事件列表
func (em *EventManager) processEvents(events []*Event) {
	for _, event := range events {
		em.processEvent(event)
	}
}

// processEvent 处理单个事件
func (em *EventManager) processEvent(event *Event) {
	defer func() {
		if r := recover(); r != nil {
			em.statistics.IncrementFailed()
			if em.errorHandler != nil {
				em.errorHandler(event, fmt.Errorf("panic in event handler: %v", r))
			}
		}
	}()

	// 获取订阅者
	subscriptions := em.getSubscriptions(event)
	if len(subscriptions) == 0 {
		em.statistics.IncrementProcessed()
		return
	}

	// 处理每个订阅者
	for _, subscription := range subscriptions {
		if !subscription.Active {
			continue
		}

		err := em.handleEventWithRetry(event, subscription.Handler)
		if err != nil {
			em.statistics.IncrementFailed()
			if em.errorHandler != nil {
				em.errorHandler(event, err)
			}
		}
	}

	em.statistics.IncrementProcessed()
}

// handleEventWithRetry 带重试的事件处理
func (em *EventManager) handleEventWithRetry(event *Event, handler EventHandler) error {
	var lastErr error
	
	for attempt := 0; attempt <= em.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := em.retryPolicy.GetDelay(attempt)
			select {
			case <-time.After(delay):
			case <-event.Context.Done():
				return event.Context.Err()
			}
		}

		err := handler.Handle(event)
		if err == nil {
			return nil
		}

		lastErr = err
		
		if !em.retryPolicy.ShouldRetry(attempt) {
			break
		}
	}

	return fmt.Errorf("event handling failed after %d attempts: %w", em.retryPolicy.MaxRetries, lastErr)
}

// getSubscriptions 获取事件的订阅者
func (em *EventManager) getSubscriptions(event *Event) []*EventSubscription {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	var subscriptions []*EventSubscription

	// 获取特定类型的订阅
	if subs, exists := em.subscriptions[string(event.Type)]; exists {
		for _, sub := range subs {
			if sub.Active && sub.Filter.Match(event) {
				subscriptions = append(subscriptions, sub)
			}
		}
	}

	// 获取订阅所有事件的订阅
	if subs, exists := em.subscriptions["*"]; exists {
		for _, sub := range subs {
			if sub.Active && sub.Filter.Match(event) {
				subscriptions = append(subscriptions, sub)
			}
		}
	}

	// 按优先级排序
	for i := 0; i < len(subscriptions)-1; i++ {
		for j := i + 1; j < len(subscriptions); j++ {
			if subscriptions[i].Handler.GetPriority() < subscriptions[j].Handler.GetPriority() {
				subscriptions[i], subscriptions[j] = subscriptions[j], subscriptions[i]
			}
		}
	}

	return subscriptions
}

// statisticsUpdater 统计更新协程
func (em *EventManager) statisticsUpdater() {
	defer em.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-em.ctx.Done():
			return
		case <-ticker.C:
			em.statistics.UpdateQueueSize(em.queue.Size())
		}
	}
}

// GetQueueSize 获取队列大小
func (em *EventManager) GetQueueSize() int {
	return em.queue.Size()
}

// IsRunning 检查是否运行中
func (em *EventManager) IsRunning() bool {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	return em.running
}

// GetSubscriptionCount 获取订阅数量
func (em *EventManager) GetSubscriptionCount() int {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	count := 0
	for _, subscriptions := range em.subscriptions {
		count += len(subscriptions)
	}
	return count
}