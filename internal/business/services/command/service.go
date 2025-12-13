package command

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	domains_events "moviepilot-go/internal/business/domains/events"
	"moviepilot-go/internal/infrastructure/events"
	"moviepilot-go/internal/integration/notification"
	"moviepilot-go/internal/schedulers"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/plugin"
)

// CommandType 命令类型
type CommandType string

const (
	// CommandTypeScheduler 调度器命令
	CommandTypeScheduler CommandType = "scheduler"
	// CommandTypeHandler 处理器命令
	CommandTypeHandler CommandType = "handler"
	// CommandTypePlugin 插件命令
	CommandTypePlugin CommandType = "plugin"
)

// CommandInfo 命令信息
type CommandInfo struct {
	Name        string                 `json:"name"`
	Type        CommandType            `json:"type"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	ID          string                 `json:"id,omitempty"`
	PID         string                 `json:"pid,omitempty"`
	Show        bool                   `json:"show,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// Handler 命令处理器接口
type Handler interface {
	// Name 命令名称（如 "subscribe"）
	Name() string
	// Description 命令描述
	Description() string
	// Category 命令分类
	Category() string
	// Execute 执行命令
	Execute(ctx context.Context, args []string) error
}

// SchedulerHandler 调度器命令处理器接口
type SchedulerHandler interface {
	// ID 调度器任务ID
	ID() string
	// Name 命令名称
	Name() string
	// Description 命令描述
	Description() string
	// Category 命令分类
	Category() string
}

// PluginHandler 插件命令处理器接口
type PluginHandler interface {
	// Name 命令名称
	Name() string
	// Description 命令描述
	Description() string
	// Category 命令分类
	Category() string
	// EventType 事件类型
	EventType() string
	// Data 事件数据
	Data() map[string]interface{}
	// Show 是否显示
	Show() bool
	// PID 插件ID
	PID() string
}

// Service 命令服务接口
type Service interface {
	// RegisterHandler 注册处理器命令
	RegisterHandler(cmd Handler) error
	// RegisterScheduler 注册调度器命令
	RegisterScheduler(cmd SchedulerHandler) error
	// RegisterPlugin 注册插件命令
	RegisterPlugin(cmd PluginHandler) error
	// Execute 执行命令
	Execute(ctx context.Context, input string) error
	// ExecuteEvent 从事件中执行命令
	ExecuteEvent(ctx context.Context, cmdStr string) error
	// List 列出所有命令
	List() []CommandInfo
	// Get 获取命令信息
	Get(cmdName string) (CommandInfo, bool)
}

// service 命令服务实现
type service struct {
	logger             *zap.Logger
	handlers           map[string]Handler
	schedulerHandlers  map[string]SchedulerHandler
	pluginHandlers     map[string]PluginHandler
	eventManager       events.Bus
	scheduler          schedulers.Scheduler
	pluginManager      plugin.Manager
	registeredCommands map[string]CommandInfo
	notificationRouter notification.Router
	mu                 sync.RWMutex
}

// ServiceOption 命令服务选项
type ServiceOption func(*service)

// WithEventManager 设置事件管理器
func WithEventManager(em events.Bus) ServiceOption {
	return func(s *service) {
		s.eventManager = em
	}
}

// WithScheduler 设置调度器
func WithScheduler(scheduler schedulers.Scheduler) ServiceOption {
	return func(s *service) {
		s.scheduler = scheduler
	}
}

// WithPluginManager 设置插件管理器
func WithPluginManager(pm plugin.Manager) ServiceOption {
	return func(s *service) {
		s.pluginManager = pm
	}
}

// WithNotificationRouter 设置通知路由器
func WithNotificationRouter(router *notification.Router) ServiceOption {
	return func(s *service) {
		s.notificationRouter = *router
	}
}

// NewService 创建命令服务
func NewService(options ...ServiceOption) Service {
	s := &service{
		logger:             logger.GetLogger(),
		handlers:           make(map[string]Handler),
		schedulerHandlers:  make(map[string]SchedulerHandler),
		pluginHandlers:     make(map[string]PluginHandler),
		registeredCommands: make(map[string]CommandInfo),
	}

	// 应用选项
	for _, opt := range options {
		opt(s)
	}

	// 初始化命令
	s.initCommands()

	return s
}

// initCommands 初始化命令
func (s *service) initCommands() {
	s.logger.Debug("Initializing commands...")

	// 注册预设命令
	if err := RegisterPresetCommands(s); err != nil {
		s.logger.Error("Failed to register preset commands", zap.Error(err))
	}

	// 构建插件命令
	pluginCommands := s.buildPluginCommands()
	for _, cmd := range pluginCommands {
		s.RegisterPlugin(cmd)
	}

	// 触发命令注册事件
	s.triggerCommandRegisterEvent()
}

// RegisterHandler 注册处理器命令
func (s *service) RegisterHandler(cmd Handler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := cmd.Name()
	if _, exists := s.handlers[name]; exists {
		return fmt.Errorf("command handler already registered: %s", name)
	}
	s.handlers[name] = cmd
	s.logger.Info("Command handler registered", zap.String("command", name))
	return nil
}

// RegisterScheduler 注册调度器命令
func (s *service) RegisterScheduler(cmd SchedulerHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := cmd.Name()
	if _, exists := s.schedulerHandlers[name]; exists {
		return fmt.Errorf("scheduler command already registered: %s", name)
	}
	s.schedulerHandlers[name] = cmd
	s.logger.Info("Scheduler command registered", zap.String("command", name))
	return nil
}

// RegisterPlugin 注册插件命令
func (s *service) RegisterPlugin(cmd PluginHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := cmd.Name()
	if _, exists := s.pluginHandlers[name]; exists {
		return fmt.Errorf("plugin command already registered: %s", name)
	}
	s.pluginHandlers[name] = cmd
	s.logger.Info("Plugin command registered", zap.String("command", name))
	return nil
}

// Execute 执行命令
func (s *service) Execute(ctx context.Context, input string) error {
	// 解析命令
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return fmt.Errorf("empty command input")
	}

	// 获取命令名称（去掉前缀斜杠）
	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	// 执行命令
	return s.executeCommand(ctx, cmdName, args)
}

// ExecuteEvent 从事件中执行命令
func (s *service) ExecuteEvent(ctx context.Context, cmdStr string) error {
	// 解析命令
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command input")
	}

	// 获取命令名称（去掉前缀斜杠）
	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	// 执行命令
	return s.executeCommand(ctx, cmdName, args)
}

// executeCommand 执行具体命令
func (s *service) executeCommand(ctx context.Context, cmdName string, args []string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 查找命令类型
	if handler, exists := s.handlers[cmdName]; exists {
		// 记录命令开始执行
		s.logger.Info("Command execution started",
			zap.String("command", cmdName),
			zap.String("command_type", string(CommandTypeHandler)),
			zap.Strings("args", args),
			zap.String("description", handler.Description()),
			zap.String("category", handler.Category()))

		// 发送开始通知
		s.sendNotification(ctx, fmt.Sprintf("开始执行：%s ...", handler.Description()))

		// 执行命令，添加重试机制
		err := s.executeWithRetry(ctx, func() error {
			return handler.Execute(ctx, args)
		}, 1) // 处理器命令通常只执行一次

		// 记录命令执行结果
		if err != nil {
			s.logger.Error("Command execution failed",
				zap.String("command", cmdName),
				zap.String("command_type", string(CommandTypeHandler)),
				zap.Error(err),
				zap.Strings("args", args))
			// 发送失败通知
			s.sendNotification(ctx, fmt.Sprintf("执行命令 %s 出错：%s", cmdName, err.Error()))
		} else {
			s.logger.Info("Command execution completed successfully",
				zap.String("command", cmdName),
				zap.String("command_type", string(CommandTypeHandler)))
			// 发送成功通知
			s.sendNotification(ctx, fmt.Sprintf("%s 执行完成", handler.Description()))
		}

		return err
	}

	if schedulerHandler, exists := s.schedulerHandlers[cmdName]; exists {
		// 记录命令开始执行
		s.logger.Info("Command execution started",
			zap.String("command", cmdName),
			zap.String("command_type", string(CommandTypeScheduler)),
			zap.String("job_id", schedulerHandler.ID()),
			zap.String("description", schedulerHandler.Description()),
			zap.String("category", schedulerHandler.Category()))

		// 发送开始通知
		s.sendNotification(ctx, fmt.Sprintf("开始执行：%s ...", schedulerHandler.Description()))

		// 执行调度器任务，添加重试机制
		err := s.executeWithRetry(ctx, func() error {
			return s.scheduler.RunJob(schedulerHandler.ID())
		}, 2) // 调度器任务可以重试一次

		// 记录命令执行结果
		if err != nil {
			s.logger.Error("Scheduler command execution failed",
				zap.String("command", cmdName),
				zap.String("job_id", schedulerHandler.ID()),
				zap.Error(err))
			// 发送失败通知
			s.sendNotification(ctx, fmt.Sprintf("执行调度命令 %s 出错：%s", cmdName, err.Error()))
		} else {
			s.logger.Info("Scheduler command execution completed successfully",
				zap.String("command", cmdName),
				zap.String("job_id", schedulerHandler.ID()))
			// 发送成功通知
			s.sendNotification(ctx, fmt.Sprintf("%s 执行完成", schedulerHandler.Description()))
		}

		return err
	}

	if pluginHandler, exists := s.pluginHandlers[cmdName]; exists {
		// 记录命令开始执行
		s.logger.Info("Command execution started",
			zap.String("command", cmdName),
			zap.String("command_type", string(CommandTypePlugin)),
			zap.String("event_type", pluginHandler.EventType()),
			zap.String("pid", pluginHandler.PID()),
			zap.String("description", pluginHandler.Description()),
			zap.String("category", pluginHandler.Category()))

		// 发送开始通知
		s.sendNotification(ctx, fmt.Sprintf("开始执行：%s ...", pluginHandler.Description()))

		// 发送插件事件，添加重试机制
		var err error
		if s.eventManager != nil {
			err = s.executeWithRetry(ctx, func() error {
				return s.eventManager.PublishBroadcast(ctx, domains_events.EventType(pluginHandler.EventType()), pluginHandler.Data(), 0)
			}, 3) // 插件事件可以重试两次

			// 记录命令执行结果
			if err != nil {
				s.logger.Error("Plugin command execution failed",
					zap.String("command", cmdName),
					zap.String("event_type", pluginHandler.EventType()),
					zap.Error(err))
				// 发送失败通知
				s.sendNotification(ctx, fmt.Sprintf("执行插件命令 %s 出错：%s", cmdName, err.Error()))
			} else {
				s.logger.Info("Plugin command execution completed successfully",
					zap.String("command", cmdName),
					zap.String("event_type", pluginHandler.EventType()))
				// 发送成功通知
				s.sendNotification(ctx, fmt.Sprintf("%s 执行完成", pluginHandler.Description()))
			}
		} else {
			err = fmt.Errorf("event manager not available for plugin command")
			s.logger.Error("Plugin command execution failed",
				zap.String("command", cmdName),
				zap.Error(err))
			// 发送失败通知
			s.sendNotification(ctx, fmt.Sprintf("执行插件命令 %s 出错：%s", cmdName, err.Error()))
		}

		return err
	}

	err := fmt.Errorf("command not found: %s", cmdName)
	s.logger.Error("Command execution failed",
		zap.String("command", cmdName),
		zap.Error(err),
		zap.Strings("args", args))
	// 发送失败通知
	s.sendNotification(ctx, fmt.Sprintf("执行命令 %s 出错：命令不存在", cmdName))
	return err
}

// executeWithRetry 带重试机制的命令执行
func (s *service) executeWithRetry(ctx context.Context, fn func() error, maxRetries int) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			s.logger.Debug("Retrying command execution",
				zap.Int("retry", i),
				zap.Int("max_retries", maxRetries),
				zap.Error(lastErr))
			// 重试间隔
			time.Sleep(time.Second * time.Duration(i))
		}

		// 捕获可能的panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("Command execution panicked",
						zap.Any("panic", r),
						zap.Stack("stack_trace"))
					lastErr = fmt.Errorf("command execution panicked: %v", r)
				}
			}()

			lastErr = fn()
		}()

		// 如果执行成功，直接返回
		if lastErr == nil {
			return nil
		}
	}

	return lastErr
}

// sendNotification 发送通知消息
func (s *service) sendNotification(ctx context.Context, message string) {
	// 创建通知消息
	msg := &notification.Message{
		Type:    notification.NotificationTypeText,
		Content: message,
	}

	// 使用defer-recover模式处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			s.logger.Warn("Notification send failed due to panic", zap.Any("panic", r))
		}
	}()

	// 发送通知
	if err := s.notificationRouter.Broadcast(ctx, msg); err != nil {
		s.logger.Warn("Failed to send notification", zap.Error(err))
	}
}

// List 列出所有命令
func (s *service) List() []CommandInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]CommandInfo, 0, len(s.handlers)+len(s.schedulerHandlers)+len(s.pluginHandlers))

	// 添加处理器命令
	for name, handler := range s.handlers {
		infos = append(infos, CommandInfo{
			Name:        name,
			Type:        CommandTypeHandler,
			Description: handler.Description(),
			Category:    handler.Category(),
			Show:        true,
		})
	}

	// 添加调度器命令
	for name, handler := range s.schedulerHandlers {
		infos = append(infos, CommandInfo{
			Name:        name,
			Type:        CommandTypeScheduler,
			Description: handler.Description(),
			Category:    handler.Category(),
			ID:          handler.ID(),
			Show:        true,
		})
	}

	// 添加插件命令
	for name, handler := range s.pluginHandlers {
		infos = append(infos, CommandInfo{
			Name:        name,
			Type:        CommandTypePlugin,
			Description: handler.Description(),
			Category:    handler.Category(),
			PID:         handler.PID(),
			Show:        handler.Show(),
			Data:        handler.Data(),
		})
	}

	return infos
}

// Get 获取命令信息
func (s *service) Get(cmdName string) (CommandInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 查找处理器命令
	if handler, exists := s.handlers[cmdName]; exists {
		return CommandInfo{
			Name:        cmdName,
			Type:        CommandTypeHandler,
			Description: handler.Description(),
			Category:    handler.Category(),
			Show:        true,
		}, true
	}

	// 查找调度器命令
	if schedulerHandler, exists := s.schedulerHandlers[cmdName]; exists {
		return CommandInfo{
			Name:        cmdName,
			Type:        CommandTypeScheduler,
			Description: schedulerHandler.Description(),
			Category:    schedulerHandler.Category(),
			ID:          schedulerHandler.ID(),
			Show:        true,
		}, true
	}

	// 查找插件命令
	if pluginHandler, exists := s.pluginHandlers[cmdName]; exists {
		return CommandInfo{
			Name:        cmdName,
			Type:        CommandTypePlugin,
			Description: pluginHandler.Description(),
			Category:    pluginHandler.Category(),
			PID:         pluginHandler.PID(),
			Show:        pluginHandler.Show(),
			Data:        pluginHandler.Data(),
		}, true
	}

	return CommandInfo{}, false
}

// handleCommandEvent 处理命令执行事件
func (s *service) handleCommandEvent(event *domains_events.Event) error {
	// 解析事件数据
	if event.Data == nil {
		return fmt.Errorf("empty event data")
	}

	// 获取命令字符串
	eventDataMap, ok := event.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data type")
	}

	cmd, ok := eventDataMap["cmd"].(string)
	if !ok {
		return fmt.Errorf("invalid cmd in event data")
	}

	// 获取其他事件数据
	channel, _ := eventDataMap["channel"].(string)
	source, _ := eventDataMap["source"].(string)
	user, _ := eventDataMap["user"].(string)

	s.logger.Info("Command execution event received",
		zap.String("cmd", cmd),
		zap.String("channel", channel),
		zap.String("source", source),
		zap.String("user", user))

	// 执行命令
	return s.ExecuteEvent(context.Background(), cmd)
}

// buildPluginCommands 构建插件命令
func (s *service) buildPluginCommands() []PluginHandler {
	var pluginHandlers []PluginHandler

	// 如果插件管理器不可用，返回空列表
	if s.pluginManager == nil {
		s.logger.Debug("Plugin manager not available, skipping plugin commands")
		return pluginHandlers
	}

	// 从插件管理器获取插件命令
	ctx := context.Background()
	_, err := s.pluginManager.GetPlugins(ctx)
	if err != nil {
		s.logger.Error("Failed to get plugins for commands", zap.Error(err))
		return pluginHandlers
	}

	// TODO: 根据实际的Plugin接口实现插件命令构建
	// 目前返回空列表

	s.logger.Info("Plugin commands built", zap.Int("count", len(pluginHandlers)))
	return pluginHandlers
}

// triggerCommandRegisterEvent 触发命令注册事件
func (s *service) triggerCommandRegisterEvent() {
	if s.eventManager == nil {
		s.logger.Debug("Event manager not available, skipping command register event")
		return
	}

	// 收集所有命令信息
	commands := s.List()

	// 转换为事件数据
	commandData := make(map[string]CommandInfo)
	for _, cmd := range commands {
		// 只包含显示的命令
		if cmd.Show {
			commandData[cmd.Name] = cmd
		}
	}

	// 触发事件
	ctx := context.Background()
	eventType := domains_events.ChainEventType("CommandRegister")

	s.logger.Info("Triggering command register event",
		zap.Int("command_count", len(commandData)),
		zap.String("event_type", string(eventType)))

	// 发送事件
	_, err := s.eventManager.DispatchChain(ctx, eventType, commandData, 0)
	if err != nil {
		s.logger.Error("Failed to dispatch command register event", zap.Error(err))
		return
	}

	// 使用原始数据更新注册的命令
	s.updateRegisteredCommands(commandData)
}

// updateRegisteredCommands 更新注册的命令
func (s *service) updateRegisteredCommands(commands map[string]CommandInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新注册的命令
	s.registeredCommands = commands

	s.logger.Info("Registered commands updated",
		zap.Int("command_count", len(commands)))
}
