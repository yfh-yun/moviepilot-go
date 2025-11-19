// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// EventManager 事件管理器
// 提供事件发送、路由、队列管理和通知功能
type EventManager struct {
	messageService MessageService
	eventQueue     EventQueue
	logger         *zap.Logger
	handlers       map[string][]EventHandler
	mutex          sync.RWMutex
}

// EventQueue 事件队列接口
type EventQueue interface {
	Publish(ctx context.Context, event *Event) error
	Subscribe(ctx context.Context, eventType string, handler EventHandler) error
	Unsubscribe(ctx context.Context, eventType string, handler EventHandler) error
	Close() error
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	EventType() string
}

// MessageService 消息服务接口
type MessageService interface {
	SendMessage(ctx context.Context, title, content string, messageType string, userIDs []uint) error
}

// NewEventManager 创建事件管理器实例
func NewEventManager(
	messageService MessageService,
	eventQueue EventQueue,
) *EventManager {
	return &EventManager{
		messageService: messageService,
		eventQueue:     eventQueue,
		logger:         logger.NewLogger("event_manager"),
		handlers:       make(map[string][]EventHandler),
	}
}

// SendEventAction 发送事件动作
// 实现Python项目中的send_event.py功能
type SendEventAction struct {
	manager *EventManager
	logger  *zap.Logger
}

// NewSendEventAction 创建发送事件动作实例
func NewSendEventAction(manager *EventManager) *SendEventAction {
	return &SendEventAction{
		manager: manager,
		logger:  logger.NewLogger("send_event_action"),
	}
}

// Execute 执行发送事件动作
func (a *SendEventAction) Execute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionCtx *ActionContext,
) (*ActionContext, error) {
	// 解析参数
	eventType, ok := params["event_type"].(string)
	if !ok {
		return nil, fmt.Errorf("事件类型参数缺失")
	}

	eventData, ok := params["event_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("事件数据参数缺失")
	}

	targetUsers := []uint{}
	if users, exists := params["target_users"]; exists {
		if usersList, ok := users.([]interface{}); ok {
			for _, user := range usersList {
				if userID, ok := user.(float64); ok {
					targetUsers = append(targetUsers, uint(userID))
				}
			}
		}
	}

	// 创建事件
	event := &Event{
		ID:         a.generateEventID(),
		Type:       eventType,
		Data:       eventData,
		WorkflowID: workflowID,
		UserIDs:    targetUsers,
		CreatedAt:  time.Now(),
	}

	a.logger.Info("开始发送事件",
		zap.String("event_type", eventType),
		zap.String("event_id", event.ID),
		zap.Int64("workflow_id", workflowID),
		zap.Int("target_users", len(targetUsers)))

	// 发送事件到队列
	if err := a.manager.SendEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("发送事件失败: %w", err)
	}

	// 如果需要通知用户
	if len(targetUsers) > 0 {
		if err := a.notifyUsers(ctx, event, targetUsers); err != nil {
			a.logger.Warn("通知用户失败", zap.Error(err))
		}
	}

	// 记录事件到工作流上下文
	if actionCtx.Variables == nil {
		actionCtx.Variables = make(map[string]interface{})
	}

	sentEvents := []interface{}{}
	if existing, exists := actionCtx.Variables["sent_events"]; exists {
		if events, ok := existing.([]interface{}); ok {
			sentEvents = events
		}
	}

	sentEvents = append(sentEvents, map[string]interface{}{
		"event_id":   event.ID,
		"event_type": eventType,
		"sent_at":    event.CreatedAt,
	})

	actionCtx.Variables["sent_events"] = sentEvents
	actionCtx.Variables["last_sent_event_id"] = event.ID

	a.logger.Info("事件发送成功",
		zap.String("event_id", event.ID),
		zap.String("event_type", eventType))

	return actionCtx, nil
}

// SendEvent 发送事件
func (m *EventManager) SendEvent(ctx context.Context, event *Event) error {
	// 发布到事件队列
	if err := m.eventQueue.Publish(ctx, event); err != nil {
		return fmt.Errorf("发布事件到队列失败: %w", err)
	}

	// 触发本地处理器
	if err := m.triggerHandlers(ctx, event); err != nil {
		m.logger.Warn("触发本地处理器失败", zap.Error(err))
	}

	m.logger.Info("事件发送成功",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.Type))

	return nil
}

// RegisterEventHandler 注册事件处理器
func (m *EventManager) RegisterEventHandler(eventType string, handler EventHandler) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.handlers[eventType] = append(m.handlers[eventType], handler)

	// 订阅事件队列
	if err := m.eventQueue.Subscribe(context.Background(), eventType, handler); err != nil {
		return fmt.Errorf("订阅事件队列失败: %w", err)
	}

	m.logger.Info("事件处理器注册成功",
		zap.String("event_type", eventType),
		zap.String("handler", fmt.Sprintf("%T", handler)))

	return nil
}

// UnregisterEventHandler 取消注册事件处理器
func (m *EventManager) UnregisterEventHandler(eventType string, handler EventHandler) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	handlers := m.handlers[eventType]
	for i, h := range handlers {
		// 简单的引用比较
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			m.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	// 取消订阅事件队列
	if err := m.eventQueue.Subscribe(context.Background(), eventType, handler); err != nil {
		return fmt.Errorf("取消订阅事件队列失败: %w", err)
	}

	return nil
}

// triggerHandlers 触发事件处理器
func (m *EventManager) triggerHandlers(ctx context.Context, event *Event) error {
	m.mutex.RLock()
	handlers := m.handlers[event.Type]
	m.mutex.RUnlock()

	if len(handlers) == 0 {
		return nil // 没有处理器
	}

	var wg sync.WaitGroup
	errors := make([]error, 0, len(handlers))

	for _, handler := range handlers {
		if handler.EventType() != event.Type {
			continue // 类型不匹配
		}

		wg.Add(1)
		go func(h EventHandler) {
			defer wg.Done()
			if err := h.Handle(ctx, event); err != nil {
				m.logger.Error("事件处理器执行失败",
					zap.String("event_type", event.Type),
					zap.String("handler", fmt.Sprintf("%T", h)),
					zap.Error(err))
				errors = append(errors, err)
			}
		}(handler)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("部分事件处理器执行失败: %v", errors)
	}

	return nil
}

// notifyUsers 通知用户
func (a *SendEventAction) notifyUsers(ctx context.Context, event *Event, userIDs []uint) error {
	title := fmt.Sprintf("工作流事件通知: %s", event.Type)

	content := fmt.Sprintf("事件ID: %s\n工作流ID: %d\n事件类型: %s\n发生时间: %s",
		event.ID, event.WorkflowID, event.Type, event.CreatedAt.Format("2006-01-02 15:04:05"))

	// 添加事件数据
	if event.Data != nil {
		dataJSON, _ := json.MarshalIndent(event.Data, "", "  ")
		content += fmt.Sprintf("\n\n事件数据:\n%s", string(dataJSON))
	}

	messageType := "system_event"
	if event.Type == "error" || event.Type == "failure" {
		messageType = "system_alert"
	}

	return a.manager.messageService.SendMessage(ctx, title, content, messageType, userIDs)
}

// generateEventID 生成事件ID
func (a *SendEventAction) generateEventID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("evt_%d_%d", timestamp, time.Now().Unix()%1000)
}

// GetEventStatistics 获取事件统计信息
func (m *EventManager) GetEventStatistics() *EventStatistics {
	m.mutex.RLock()
	handlerCount := 0
	for _, handlers := range m.handlers {
		handlerCount += len(handlers)
	}
	m.mutex.RUnlock()

	return &EventStatistics{
		RegisteredHandlers: handlerCount,
		EventTypes:         m.getEventTypes(),
		LastUpdateTime:     time.Now(),
	}
}

// getEventTypes 获取注册的事件类型
func (m *EventManager) getEventTypes() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	types := make([]string, 0, len(m.handlers))
	for eventType := range m.handlers {
		types = append(types, eventType)
	}

	return types
}

// 数据结构定义

// Event 事件
type Event struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	WorkflowID  int64                  `json:"workflow_id"`
	UserIDs     []uint                 `json:"user_ids"`
	CreatedAt   time.Time              `json:"created_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
}

// EventStatistics 事件统计信息
type EventStatistics struct {
	RegisteredHandlers int       `json:"registered_handlers"`
	EventTypes         []string  `json:"event_types"`
	TotalEvents        int64     `json:"total_events"`
	ProcessedEvents    int64     `json:"processed_events"`
	FailedEvents       int64     `json:"failed_events"`
	LastUpdateTime     time.Time `json:"last_update_time"`
}

// WorkflowCompletedHandler 工作流完成事件处理器
type WorkflowCompletedHandler struct {
	messageService MessageService
	logger         *zap.Logger
}

func NewWorkflowCompletedHandler(messageService MessageService) *WorkflowCompletedHandler {
	return &WorkflowCompletedHandler{
		messageService: messageService,
		logger:         logger.NewLogger("workflow_completed_handler"),
	}
}

func (h *WorkflowCompletedHandler) EventType() string {
	return "workflow_completed"
}

func (h *WorkflowCompletedHandler) Handle(ctx context.Context, event *Event) error {
	h.logger.Info("处理工作流完成事件",
		zap.String("event_id", event.ID),
		zap.Int64("workflow_id", event.WorkflowID))

	// 这里可以实现工作流完成后的清理工作
	// 例如：清理缓存、发送完成通知等
	return nil
}

// DownloadCompletedHandler 下载完成事件处理器
type DownloadCompletedHandler struct {
	messageService MessageService
	logger         *zap.Logger
}

func NewDownloadCompletedHandler(messageService MessageService) *DownloadCompletedHandler {
	return &DownloadCompletedHandler{
		messageService: messageService,
		logger:         logger.NewLogger("download_completed_handler"),
	}
}

func (h *DownloadCompletedHandler) EventType() string {
	return "download_completed"
}

func (h *DownloadCompletedHandler) Handle(ctx context.Context, event *Event) error {
	h.logger.Info("处理下载完成事件",
		zap.String("event_id", event.ID),
		zap.Int64("workflow_id", event.WorkflowID))

	// 这里可以实现下载完成后的处理工作
	// 例如：文件转码、元数据更新等
	return nil
}

// MemoryEventQueue 内存事件队列实现
type MemoryEventQueue struct {
	subscribers map[string][]EventHandler
	events      []*Event
	mutex       sync.RWMutex
	closed      bool
}

func NewMemoryEventQueue() *MemoryEventQueue {
	return &MemoryEventQueue{
		subscribers: make(map[string][]EventHandler),
		events:      make([]*Event, 0),
	}
}

func (q *MemoryEventQueue) Publish(ctx context.Context, event *Event) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.closed {
		return fmt.Errorf("事件队列已关闭")
	}

	q.events = append(q.events, event)
	return nil
}

func (q *MemoryEventQueue) Subscribe(ctx context.Context, eventType string, handler EventHandler) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.closed {
		return fmt.Errorf("事件队列已关闭")
	}

	q.subscribers[eventType] = append(q.subscribers[eventType], handler)
	return nil
}

func (q *MemoryEventQueue) Unsubscribe(ctx context.Context, eventType string, handler EventHandler) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	handlers := q.subscribers[eventType]
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			q.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	return nil
}

func (q *MemoryEventQueue) Close() error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.closed = true
	q.events = nil
	q.subscribers = nil
	return nil
}
