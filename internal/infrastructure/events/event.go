// Package event 事件系统模块
package event

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType 事件类型
type EventType string

const (
	// 系统事件
	EventSystemStart    EventType = "system.start"
	EventSystemStop     EventType = "system.stop"
	EventSystemRestart  EventType = "system.restart"
	EventSystemError    EventType = "system.error"

	// 用户事件
	EventUserLogin      EventType = "user.login"
	EventUserLogout     EventType = "user.logout"
	EventUserRegister   EventType = "user.register"
	EventUserUpdate     EventType = "user.update"

	// 媒体事件
	EventMediaAdded     EventType = "media.added"
	EventMediaUpdated   EventType = "media.updated"
	EventMediaDeleted   EventType = "media.deleted"
	EventMediaPlayed    EventType = "media.played"

	// 下载事件
	EventDownloadStart  EventType = "download.start"
	EventDownloadPause  EventType = "download.pause"
	EventDownloadResume EventType = "download.resume"
	EventDownloadComplete EventType = "download.complete"
	EventDownloadFailed EventType = "download.failed"

	// 订阅事件
	EventSubscribeCreated EventType = "subscribe.created"
	EventSubscribeUpdated EventType = "subscribe.updated"
	EventSubscribeDeleted EventType = "subscribe.deleted"
	EventSubscribeTriggered EventType = "subscribe.triggered"

	// 种子事件
	EventTorrentFound   EventType = "torrent.found"
	EventTorrentAdded   EventType = "torrent.added"
	EventTorrentDownloaded EventType = "torrent.downloaded"
	EventTorrentFailed  EventType = "torrent.failed"

	// 工作流事件
	EventWorkflowStart  EventType = "workflow.start"
	EventWorkflowComplete EventType = "workflow.complete"
	EventWorkflowFailed EventType = "workflow.failed"

	// 插件事件
	EventPluginLoaded   EventType = "plugin.loaded"
	EventPluginUnloaded EventType = "plugin.unloaded"
	EventPluginError    EventType = "plugin.error"
)

// EventPriority 事件优先级
type EventPriority int

const (
	PriorityLow    EventPriority = 1
	PriorityNormal EventPriority = 5
	PriorityHigh   EventPriority = 10
	PriorityUrgent EventPriority = 20
)

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Priority  EventPriority          `json:"priority"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Context   context.Context        `json:"-"`
}

// NewEvent 创建新事件
func NewEvent(eventType EventType, data map[string]interface{}) *Event {
	return &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Data:      data,
		Priority:  PriorityNormal,
		Source:    "system",
		Timestamp: time.Now(),
		Context:   context.Background(),
	}
}

// NewEventWithPriority 创建带优先级的事件
func NewEventWithPriority(eventType EventType, data map[string]interface{}, priority EventPriority) *Event {
	event := NewEvent(eventType, data)
	event.Priority = priority
	return event
}

// NewEventWithSource 创建带来源的事件
func NewEventWithSource(eventType EventType, data map[string]interface{}, source string) *Event {
	event := NewEvent(eventType, data)
	event.Source = source
	return event
}

// SetPriority 设置优先级
func (e *Event) SetPriority(priority EventPriority) {
	e.Priority = priority
}

// SetSource 设置来源
func (e *Event) SetSource(source string) {
	e.Source = source
}

// SetContext 设置上下文
func (e *Event) SetContext(ctx context.Context) {
	e.Context = ctx
}

// GetData 获取数据
func (e *Event) GetData(key string) (interface{}, bool) {
	value, exists := e.Data[key]
	return value, exists
}

// SetData 设置数据
func (e *Event) SetData(key string, value interface{}) {
	if e.Data == nil {
		e.Data = make(map[string]interface{})
	}
	e.Data[key] = value
}

// EventHandler 事件处理器
type EventHandler interface {
	Handle(event *Event) error
	GetName() string
	GetPriority() EventPriority
	IsAsync() bool
}

// EventHandlerFunc 事件处理器函数
type EventHandlerFunc struct {
	Name     string
	Priority EventPriority
	Async    bool
	Func     func(event *Event) error
}

// Handle 处理事件
func (h *EventHandlerFunc) Handle(event *Event) error {
	return h.Func(event)
}

// GetName 获取处理器名称
func (h *EventHandlerFunc) GetName() string {
	return h.Name
}

// GetPriority 获取处理器优先级
func (h *EventHandlerFunc) GetPriority() EventPriority {
	return h.Priority
}

// IsAsync 是否异步处理
func (h *EventHandlerFunc) IsAsync() bool {
	return h.Async
}

// NewEventHandler 创建事件处理器
func NewEventHandler(name string, priority EventPriority, async bool, handler func(event *Event) error) EventHandler {
	return &EventHandlerFunc{
		Name:     name,
		Priority: priority,
		Async:    async,
		Func:     handler,
	}
}

// EventFilter 事件过滤器
type EventFilter interface {
	Match(event *Event) bool
	GetName() string
}

// EventTypeFilter 事件类型过滤器
type EventTypeFilter struct {
	Name  string
	Types []EventType
}

// Match 匹配事件
func (f *EventTypeFilter) Match(event *Event) bool {
	for _, eventType := range f.Types {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

// GetName 获取过滤器名称
func (f *EventTypeFilter) GetName() string {
	return f.Name
}

// NewEventTypeFilter 创建事件类型过滤器
func NewEventTypeFilter(name string, types ...EventType) EventFilter {
	return &EventTypeFilter{
		Name:  name,
		Types: types,
	}
}

// EventSourceFilter 事件来源过滤器
type EventSourceFilter struct {
	Name   string
	Sources []string
}

// Match 匹配事件
func (f *EventSourceFilter) Match(event *Event) bool {
	for _, source := range f.Sources {
		if event.Source == source {
			return true
		}
	}
	return false
}

// GetName 获取过滤器名称
func (f *EventSourceFilter) GetName() string {
	return f.Name
}

// NewEventSourceFilter 创建事件来源过滤器
func NewEventSourceFilter(name string, sources ...string) EventFilter {
	return &EventSourceFilter{
		Name:   name,
		Sources: sources,
	}
}

// EventSubscription 事件订阅
type EventSubscription struct {
	ID       string
	Filter   EventFilter
	Handler  EventHandler
	Active   bool
	Created  time.Time
}

// NewEventSubscription 创建事件订阅
func NewEventSubscription(filter EventFilter, handler EventHandler) *EventSubscription {
	return &EventSubscription{
		ID:      uuid.New().String(),
		Filter:  filter,
		Handler: handler,
		Active:  true,
		Created: time.Now(),
	}
}

// EventQueue 事件队列
type EventQueue struct {
	events chan *Event
	size   int
}

// NewEventQueue 创建事件队列
func NewEventQueue(size int) *EventQueue {
	return &EventQueue{
		events: make(chan *Event, size),
		size:   size,
	}
}

// Push 推入事件
func (eq *EventQueue) Push(event *Event) error {
	select {
	case eq.events <- event:
		return nil
	default:
		return fmt.Errorf("event queue is full")
	}
}

// Pop 弹出事件
func (eq *EventQueue) Pop() (*Event, bool) {
	select {
	case event := <-eq.events:
		return event, true
	default:
		return nil, false
	}
}

// Size 获取队列大小
func (eq *EventQueue) Size() int {
	return len(eq.events)
}

// IsFull 检查队列是否已满
func (eq *EventQueue) IsFull() bool {
	return len(eq.events) >= eq.size
}

// IsEmpty 检查队列是否为空
func (eq *EventQueue) IsEmpty() bool {
	return len(eq.events) == 0
}

// EventStatistics 事件统计
type EventStatistics struct {
	TotalEvents     int64     `json:"total_events"`
	ProcessedEvents int64     `json:"processed_events"`
	FailedEvents    int64     `json:"failed_events"`
	QueueSize       int       `json:"queue_size"`
	LastEventTime   time.Time `json:"last_event_time"`
	ProcessRate     float64   `json:"process_rate"`
	ErrorRate       float64   `json:"error_rate"`
	mutex           sync.RWMutex
}

// NewEventStatistics 创建事件统计
func NewEventStatistics() *EventStatistics {
	return &EventStatistics{
		LastEventTime: time.Now(),
	}
}

// IncrementTotal 增加总事件数
func (es *EventStatistics) IncrementTotal() {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	es.TotalEvents++
	es.LastEventTime = time.Now()
}

// IncrementProcessed 增加已处理事件数
func (es *EventStatistics) IncrementProcessed() {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	es.ProcessedEvents++
	es.calculateRates()
}

// IncrementFailed 增加失败事件数
func (es *EventStatistics) IncrementFailed() {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	es.FailedEvents++
	es.calculateRates()
}

// UpdateQueueSize 更新队列大小
func (es *EventStatistics) UpdateQueueSize(size int) {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	es.QueueSize = size
}

// GetSnapshot 获取统计快照
func (es *EventStatistics) GetSnapshot() map[string]interface{} {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	return map[string]interface{}{
		"total_events":     es.TotalEvents,
		"processed_events": es.ProcessedEvents,
		"failed_events":    es.FailedEvents,
		"queue_size":       es.QueueSize,
		"last_event_time":  es.LastEventTime,
		"process_rate":     es.ProcessRate,
		"error_rate":       es.ErrorRate,
	}
}

// calculateRates 计算处理率和错误率
func (es *EventStatistics) calculateRates() {
	if es.TotalEvents > 0 {
		es.ProcessRate = float64(es.ProcessedEvents) / float64(es.TotalEvents)
		es.ErrorRate = float64(es.FailedEvents) / float64(es.TotalEvents)
	}
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// DefaultRetryPolicy 默认重试策略
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  time.Second,
		MaxDelay:      time.Minute,
		BackoffFactor: 2.0,
	}
}

// GetDelay 获取重试延迟
func (rp *RetryPolicy) GetDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	delay := time.Duration(float64(rp.InitialDelay) * 
		(1 << (attempt - 1)) * rp.BackoffFactor)
	
	if delay > rp.MaxDelay {
		delay = rp.MaxDelay
	}
	
	return delay
}

// ShouldRetry 是否应该重试
func (rp *RetryPolicy) ShouldRetry(attempt int) bool {
	return attempt < rp.MaxRetries
}