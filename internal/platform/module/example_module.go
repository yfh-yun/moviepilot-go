package module

import (
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// ExampleModule 示例模块实现
type ExampleModule struct {
	*ModuleBase
	initialized bool
	running     bool
	config      map[string]interface{}
	taskDone    chan bool
}

// NewExampleModule 创建示例模块
func NewExampleModule() *ExampleModule {
	// 创建基础模块
	base := NewModuleBase(
		"example_module",
		"示例模块",
		"一个用于演示模块系统的示例模块",
		"1.0.0",
		"MoviePilot Team",
		100,  // 优先级
		[]string{}, // 依赖项
	)

	return &ExampleModule{
		ModuleBase:  base,
		initialized: false,
		running:     false,
		taskDone:    make(chan bool),
	}
}

// Initialize 初始化模块
func (em *ExampleModule) Initialize(config map[string]interface{}, logger logger.Logger) error {
	// 调用基础类初始化
	if err := em.ModuleBase.Initialize(config, logger); err != nil {
		return err
	}

	// 模块特定的初始化逻辑
	em.logger.Info("Initializing example module")
	em.config = em.GetConfig()
	em.initialized = true

	// 初始化默认配置
	if em.config["interval"] == nil {
		em.config["interval"] = 30
	}

	if em.config["enabled"] == nil {
		em.config["enabled"] = true
	}

	em.logger.Info("Example module initialized successfully", 
		"interval", em.config["interval"], 
		"enabled", em.config["enabled"])

	return nil
}

// Start 启动模块
func (em *ExampleModule) Start() error {
	// 调用基础类的Start方法
	if err := em.ModuleBase.Start(); err != nil {
		return err
	}

	// 检查是否已初始化
	if !em.initialized {
		em.SetError(fmt.Errorf("module not initialized"))
		return fmt.Errorf("module not initialized")
	}

	// 检查是否启用
	enabled, ok := em.config["enabled"].(bool)
	if !ok || !enabled {
		em.SetError(fmt.Errorf("module is disabled"))
		return fmt.Errorf("module is disabled")
	}

	em.logger.Info("Starting example module")
	em.running = true
	em.SetStatus(StatusRunning)

	// 启动后台任务
	go em.backgroundTask()

	em.logger.Info("Example module started successfully")
	return nil
}

// Stop 停止模块
func (em *ExampleModule) Stop() error {
	// 调用基础类的Stop方法
	if err := em.ModuleBase.Stop(); err != nil {
		return err
	}

	em.logger.Info("Stopping example module")

	// 停止后台任务
	if em.running {
		em.running = false
		select {
		case em.taskDone <- true:
			// 任务已停止
		case <-time.After(2 * time.Second):
			// 超时，强制停止
			em.logger.Warn("Background task timed out during shutdown")
		}
	}

	em.SetStatus(StatusStopped)
	em.logger.Info("Example module stopped successfully")
	return nil
}

// UpdateConfig 更新模块配置
func (em *ExampleModule) UpdateConfig(config map[string]interface{}) error {
	// 调用基础类更新配置
	if err := em.ModuleBase.UpdateConfig(config); err != nil {
		return err
	}

	// 更新本地配置副本
	em.config = em.GetConfig()
	em.logger.Info("Example module config updated", "config", em.config)

	// 如果模块正在运行，可能需要重新配置后台任务
	if em.IsRunning() {
		em.logger.Info("Config changed, reconfiguring background task")
		// 这里可以重新配置后台任务
	}

	return nil
}

// backgroundTask 后台任务示例
func (em *ExampleModule) backgroundTask() {
	interval, ok := em.config["interval"].(int)
	if !ok {
		interval = 30 // 默认30秒
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	em.logger.Info("Background task started", "interval", interval)

	for {
		select {
		case <-ticker.C:
			em.runTask()
		case <-em.taskDone:
			em.logger.Info("Background task stopped")
			return
		}
	}
}

// runTask 运行周期性任务
func (em *ExampleModule) runTask() {
	if !em.running {
		return
	}

	em.logger.Debug("Running periodic task")

	// 模拟任务执行
	taskResult := fmt.Sprintf("Task executed at %s", time.Now().Format(time.RFC3339))

	em.logger.Info("Task completed", "result", taskResult)
}

// GetStats 获取模块统计信息
func (em *ExampleModule) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"initialized": em.initialized,
		"running":     em.running,
		"config":      em.config,
		"status":      em.GetStatus(),
	}
}

// ModuleEventHandler 模块事件处理器接口
type ModuleEventHandler interface {
	// HandleModuleEvent 处理模块事件
	HandleModuleEvent(event *ModuleEvent)

	// GetEventTypes 获取处理器支持的事件类型
	GetEventTypes() []string
}

// ModuleEvent 模块事件
type ModuleEvent struct {
	EventType   string      // 事件类型
	ModuleID    string      // 模块ID
	Timestamp   time.Time   // 事件时间戳
	Data        interface{} // 事件数据
	Error       error       // 错误信息（如果有）
}

// NewModuleEvent 创建新的模块事件
func NewModuleEvent(eventType, moduleID string, data interface{}, err error) *ModuleEvent {
	return &ModuleEvent{
		EventType: eventType,
		ModuleID:  moduleID,
		Timestamp: time.Now(),
		Data:      data,
		Error:     err,
	}
}

// EventManager 模块事件管理器
type EventManager struct {
	handlers    []ModuleEventHandler
	eventTypes  map[string][]ModuleEventHandler
	mutex       sync.RWMutex
	logger      logger.Logger
}

// NewEventManager 创建事件管理器
func NewEventManager(logger logger.Logger) *EventManager {
	return &EventManager{
		handlers:   make([]ModuleEventHandler, 0),
		eventTypes: make(map[string][]ModuleEventHandler),
		logger:     logger,
	}
}

// RegisterEventHandler 注册事件处理器
func (em *EventManager) RegisterEventHandler(handler ModuleEventHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.handlers = append(em.handlers, handler)

	// 注册事件类型
	for _, eventType := range handler.GetEventTypes() {
		if _, exists := em.eventTypes[eventType]; !exists {
			em.eventTypes[eventType] = make([]ModuleEventHandler, 0)
		}
		em.eventTypes[eventType] = append(em.eventTypes[eventType], handler)
	}

	em.logger.Debug("Event handler registered", 
		"handler", fmt.Sprintf("%T", handler), 
		"events", handler.GetEventTypes())
}

// UnregisterEventHandler 注销事件处理器
func (em *EventManager) UnregisterEventHandler(handler ModuleEventHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	// 从处理器列表中移除
	for i, h := range em.handlers {
		if h == handler {
			em.handlers = append(em.handlers[:i], em.handlers[i+1:]...)
			break
		}
	}

	// 从事件类型映射中移除
	for eventType, handlers := range em.eventTypes {
		for i, h := range handlers {
			if h == handler {
				em.eventTypes[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
		// 如果该事件类型没有处理器了，删除该类型
		if len(em.eventTypes[eventType]) == 0 {
			delete(em.eventTypes, eventType)
		}
	}

	em.logger.Debug("Event handler unregistered", 
		"handler", fmt.Sprintf("%T", handler))
}

// TriggerEvent 触发事件
func (em *EventManager) TriggerEvent(event *ModuleEvent) {
	em.mutex.RLock()
	handlers, exists := em.eventTypes[event.EventType]
	em.mutex.RUnlock()

	if !exists || len(handlers) == 0 {
		em.logger.Debug("No handlers for event type", 
			"event_type", event.EventType,
			"module_id", event.ModuleID)
		return
	}

	em.logger.Debug("Triggering event", 
		"event_type", event.EventType,
		"module_id", event.ModuleID,
		"handler_count", len(handlers))

	// 异步处理事件
	for _, handler := range handlers {
		go func(h ModuleEventHandler) {
			defer func() {
				if r := recover(); r != nil {
					em.logger.Error("Panic in event handler", 
						"handler", fmt.Sprintf("%T", h),
						"error", r)
				}
			}()

			h.HandleModuleEvent(event)
		}(handler)
	}
}

// GetRegisteredEventTypes 获取已注册的事件类型
func (em *EventManager) GetRegisteredEventTypes() []string {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	types := make([]string, 0, len(em.eventTypes))
	for eventType := range em.eventTypes {
		types = append(types, eventType)
	}

	return types
}

// ModuleEventTypes 预定义的模块事件类型
const (
	EventModuleInitialized = "module.initialized" // 模块初始化完成
	EventModuleStarting    = "module.starting"    // 模块开始启动
	EventModuleStarted     = "module.started"     // 模块启动完成
	EventModuleStopping    = "module.stopping"    // 模块开始停止
	EventModuleStopped     = "module.stopped"     // 模块停止完成
	EventModuleError       = "module.error"       // 模块错误
	EventModuleConfigUpdated = "module.config_updated" // 模块配置更新
	EventModuleDependencyFailed = "module.dependency_failed" // 模块依赖失败
)

// DefaultModuleEventHandler 默认的模块事件处理器实现
type DefaultModuleEventHandler struct {
	logger logger.Logger
}

// NewDefaultModuleEventHandler 创建默认的模块事件处理器
func NewDefaultModuleEventHandler(logger logger.Logger) *DefaultModuleEventHandler {
	return &DefaultModuleEventHandler{
		logger: logger,
	}
}

// GetEventTypes 获取支持的事件类型
func (h *DefaultModuleEventHandler) GetEventTypes() []string {
	return []string{
		EventModuleInitialized,
		EventModuleStarting,
		EventModuleStarted,
		EventModuleStopping,
		EventModuleStopped,
		EventModuleError,
		EventModuleConfigUpdated,
		EventModuleDependencyFailed,
	}
}

// HandleModuleEvent 处理模块事件
func (h *DefaultModuleEventHandler) HandleModuleEvent(event *ModuleEvent) {
	switch event.EventType {
	case EventModuleInitialized:
		h.logger.Info("Module initialized", "module", event.ModuleID)
	case EventModuleStarting:
		h.logger.Info("Module starting", "module", event.ModuleID)
	case EventModuleStarted:
		h.logger.Info("Module started", "module", event.ModuleID)
	case EventModuleStopping:
		h.logger.Info("Module stopping", "module", event.ModuleID)
	case EventModuleStopped:
		h.logger.Info("Module stopped", "module", event.ModuleID)
	case EventModuleError:
		h.logger.Error("Module error", "module", event.ModuleID, "error", event.Error)
	case EventModuleConfigUpdated:
		h.logger.Info("Module config updated", "module", event.ModuleID)
	case EventModuleDependencyFailed:
		h.logger.Error("Module dependency failed", "module", event.ModuleID, "error", event.Error)
	default:
		h.logger.Debug("Unhandled module event", "type", event.EventType, "module", event.ModuleID)
	}
}