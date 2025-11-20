package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// 事件类型常量
const (
	// 系统事件
	EventSystemStart       = "system_start"
	EventSystemStop        = "system_stop"
	EventSystemUpdate      = "system_update"

	// 下载事件
	EventDownloadAdd       = "download_add"
	EventDownloadStart     = "download_start"
	EventDownloadProgress  = "download_progress"
	EventDownloadComplete  = "download_complete"
	EventDownloadError     = "download_error"
	EventDownloadPause     = "download_pause"
	EventDownloadResume    = "download_resume"
	EventDownloadRemove    = "download_remove"

	// 媒体事件
	EventMediaIdentify     = "media_identify"
	EventMediaMatched      = "media_matched"
	EventMediaRename       = "media_rename"
	EventMediaMove         = "media_move"
	EventMediaScan         = "media_scan"
	EventMediaUpdate       = "media_update"

	// 任务事件
	EventTaskStart         = "task_start"
	EventTaskProgress      = "task_progress"
	EventTaskComplete      = "task_complete"
	EventTaskError         = "task_error"

	// 插件事件
	EventPluginLoad        = "plugin_load"
	EventPluginUnload      = "plugin_unload"
	EventPluginUpdate      = "plugin_update"

	// 通知事件
	EventNotification      = "notification"
)

// Event 事件接口
type Event interface {
	GetID() string
	GetType() string
	GetData() map[string]interface{}
	GetTimestamp() time.Time
	GetSource() string
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
}

// GetID 获取事件ID
func (e *BaseEvent) GetID() string {
	return e.ID
}

// GetType 获取事件类型
func (e *BaseEvent) GetType() string {
	return e.EventType
}

// GetData 获取事件数据
func (e *BaseEvent) GetData() map[string]interface{} {
	return e.Data
}

// GetTimestamp 获取事件时间戳
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetSource 获取事件源
func (e *BaseEvent) GetSource() string {
	return e.Source
}

// EventHandler 事件处理器函数类型
type EventHandler func(context.Context, Event) error

// EventSubscriber 事件订阅者
type EventSubscriber struct {
	ID        string
	Handler   EventHandler
	Priority  int           // 优先级，值越小优先级越高
	Once      bool          // 是否只执行一次
	Filter    func(Event) bool // 过滤函数
}

// EventManager 事件管理器
type EventManager struct {
	logger       *logger.Logger
	subscribers  map[string][]EventSubscriber
	mutex        sync.RWMutex
	wg           sync.WaitGroup
	shutdown     chan struct{}
	maxWorkers   int
	workerPool   chan struct{}
}

// NewEventManager 创建事件管理器
func NewEventManager(log *logger.Logger, maxWorkers int) *EventManager {
	if maxWorkers <= 0 {
		maxWorkers = 10
	}

	em := &EventManager{
		logger:      log,
		subscribers: make(map[string][]EventSubscriber),
		shutdown:    make(chan struct{}),
		maxWorkers:  maxWorkers,
		workerPool:  make(chan struct{}, maxWorkers),
	}

	log.Info("Event manager initialized", "max_workers", maxWorkers)
	return em
}

// Subscribe 订阅事件
func (em *EventManager) Subscribe(eventType string, subscriber EventSubscriber) string {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if _, exists := em.subscribers[eventType]; !exists {
		em.subscribers[eventType] = make([]EventSubscriber, 0)
	}

	// 如果没有指定ID，生成一个默认ID
	if subscriber.ID == "" {
		subscriber.ID = generateEventID()
	}

	// 添加订阅者
	em.subscribers[eventType] = append(em.subscribers[eventType], subscriber)

	// 按优先级排序
	em.sortSubscribers(eventType)

	em.logger.Info("Event subscribed", "event_type", eventType, "subscriber_id", subscriber.ID)
	return subscriber.ID
}

// Unsubscribe 取消订阅
func (em *EventManager) Unsubscribe(eventType string, subscriberID string) bool {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if subscribers, exists := em.subscribers[eventType]; exists {
		for i, subscriber := range subscribers {
			if subscriber.ID == subscriberID {
				// 删除订阅者
				em.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
				em.logger.Info("Event unsubscribed", "event_type", eventType, "subscriber_id", subscriberID)

				// 如果没有订阅者了，删除事件类型
				if len(em.subscribers[eventType]) == 0 {
					delete(em.subscribers, eventType)
				}

				return true
			}
		}
	}

	em.logger.Warn("Failed to unsubscribe event", "event_type", eventType, "subscriber_id", subscriberID)
	return false
}

// UnsubscribeAll 取消所有订阅
func (em *EventManager) UnsubscribeAll(subscriberID string) int {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	count := 0
	for eventType, subscribers := range em.subscribers {
		newSubscribers := make([]EventSubscriber, 0)
		for _, subscriber := range subscribers {
			if subscriber.ID != subscriberID {
				newSubscribers = append(newSubscribers, subscriber)
			} else {
				count++
			}
		}

		if len(newSubscribers) > 0 {
			em.subscribers[eventType] = newSubscribers
		} else {
			delete(em.subscribers, eventType)
		}
	}

	em.logger.Info("All event subscriptions removed", "subscriber_id", subscriberID, "count", count)
	return count
}

// Publish 发布事件
func (em *EventManager) Publish(ctx context.Context, event Event) {
	em.mutex.RLock()
	// 获取该事件类型的订阅者（复制一份以避免锁竞争）
	eventType := event.GetType()
	var subscribers []EventSubscriber
	if subs, exists := em.subscribers[eventType]; exists {
		subscribers = make([]EventSubscriber, len(subs))
		copy(subscribers, subs)
	}
	// 获取通配符订阅者
	var wildcardSubscribers []EventSubscriber
	if subs, exists := em.subscribers["*"]; exists {
		wildcardSubscribers = make([]EventSubscriber, len(subs))
		copy(wildcardSubscribers, subs)
	}
	em.mutex.RUnlock()

	// 记录事件发布
	em.logger.Debug("Event published", "type", eventType, "id", event.GetID(), "source", event.GetSource())

	// 处理事件订阅者
	totalSubscribers := append(subscribers, wildcardSubscribers...)
	for _, subscriber := range totalSubscribers {
		// 检查过滤器
		if subscriber.Filter != nil && !subscriber.Filter(event) {
			continue
		}

		// 提交到工作池处理
		em.wg.Add(1)
		go em.handleEvent(ctx, event, subscriber)
	}
}

// PublishSync 同步发布事件（阻塞直到所有处理器完成）
func (em *EventManager) PublishSync(ctx context.Context, event Event) {
	em.Publish(ctx, event)
	em.wg.Wait()
}

// handleEvent 处理事件
func (em *EventManager) handleEvent(ctx context.Context, event Event, subscriber EventSubscriber) {
	defer em.wg.Done()

	// 检查是否关闭
	select {
	case <-em.shutdown:
		return
	default:
	}

	// 获取工作池槽位
	select {
	case em.workerPool <- struct{}{}:
		defer func() {
			<-em.workerPool
		}()
	default:
		em.logger.Warn("Worker pool full, skipping event handler", "event_type", event.GetType(), "subscriber_id", subscriber.ID)
		return
	}

	// 执行处理器
	start := time.Now()
	err := subscriber.Handler(ctx, event)
	duration := time.Since(start)

	// 记录处理结果
	if err != nil {
		em.logger.Error("Event handler failed", 
			"event_type", event.GetType(),
			"subscriber_id", subscriber.ID,
			"error", err.Error(),
			"duration", duration,
		)
	} else {
		em.logger.Debug("Event handler completed",
			"event_type", event.GetType(),
			"subscriber_id", subscriber.ID,
			"duration", duration,
		)
	}

	// 如果是一次性订阅者，取消订阅
	if subscriber.Once {
		em.Unsubscribe(event.GetType(), subscriber.ID)
	}
}

// ListSubscribers 列出指定事件类型的订阅者
func (em *EventManager) ListSubscribers(eventType string) []EventSubscriber {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if subscribers, exists := em.subscribers[eventType]; exists {
		result := make([]EventSubscriber, len(subscribers))
		copy(result, subscribers)
		return result
	}
	return []EventSubscriber{}
}

// GetStatistics 获取事件管理器统计信息
func (em *EventManager) GetStatistics() map[string]interface{} {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	eventCount := len(em.subscribers)
	totalSubscribers := 0
	subscribersByEvent := make(map[string]int)

	for eventType, subscribers := range em.subscribers {
		subscribersByEvent[eventType] = len(subscribers)
		totalSubscribers += len(subscribers)
	}

	return map[string]interface{}{
		"event_types":      eventCount,
		"total_subscribers": totalSubscribers,
		"subscribers_by_event": subscribersByEvent,
		"max_workers":      em.maxWorkers,
		"active_workers":   len(em.workerPool),
	}
}

// Stop 停止事件管理器
func (em *EventManager) Stop() {
	close(em.shutdown)
	em.wg.Wait()
	em.logger.Info("Event manager stopped")
}

// CreateEvent 创建新事件
func (em *EventManager) CreateEvent(eventType string, source string, data map[string]interface{}) Event {
	return &BaseEvent{
		ID:        generateEventID(),
		EventType: eventType,
		Source:    source,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// sortSubscribers 按优先级排序订阅者
func (em *EventManager) sortSubscribers(eventType string) {
	subscribers := em.subscribers[eventType]
	// 冒泡排序，按优先级从小到大排序
	for i := 0; i < len(subscribers)-1; i++ {
		for j := 0; j < len(subscribers)-i-1; j++ {
			if subscribers[j].Priority > subscribers[j+1].Priority {
				subscribers[j], subscribers[j+1] = subscribers[j+1], subscribers[j]
			}
		}
	}
	em.subscribers[eventType] = subscribers
}

// generateEventID 生成事件ID
func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

// AddEventListener 添加事件监听器的简化方法
func (em *EventManager) AddEventListener(eventType string, handler EventHandler) string {
	subscriber := EventSubscriber{
		Handler:  handler,
		Priority: 0,
		Once:     false,
	}
	return em.Subscribe(eventType, subscriber)
}

// AddOneTimeEventListener 添加一次性事件监听器
func (em *EventManager) AddOneTimeEventListener(eventType string, handler EventHandler) string {
	subscriber := EventSubscriber{
		Handler:  handler,
		Priority: 0,
		Once:     true,
	}
	return em.Subscribe(eventType, subscriber)
}

// AddFilteredEventListener 添加带过滤条件的事件监听器
func (em *EventManager) AddFilteredEventListener(eventType string, handler EventHandler, filter func(Event) bool) string {
	subscriber := EventSubscriber{
		Handler: handler,
		Priority: 0,
		Once:    false,
		Filter:  filter,
	}
	return em.Subscribe(eventType, subscriber)
}
