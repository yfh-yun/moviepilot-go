package command

import (
	"reflect"
	"strings"
	"sync"
	"github.com/google/uuid"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/scheduler"
	"moviepilot-go/internal/core"
)

// Command 命令结构�?type Command struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "scheduler" �?"command"
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Function    func(args ...interface{}) error `json:"-"`
	Data        map[string]interface{} `json:"data"`
	Show        bool                   `json:"show"`
	PID         string                 `json:"pid,omitempty"`
}

// CommandManager 命令管理�?type CommandManager struct {
	// 注册的命令集�?	registeredCommands map[string]*Command
	// 所有命令集�?	commands map[string]*Command
	// 内建命令集合
	presetCommands map[string]*Command
	// 插件命令集合
	pluginCommands map[string]*Command
	// 其他命令集合
	otherCommands map[string]*Command
	// 互斥�?	mutex sync.RWMutex
	// 定时服务管理�?	scheduler *scheduler.Scheduler
	// 事件总线
	eventBus *core.EventBus
	// 命令�?	commandChain *CommandChain
}

// NewCommandManager 创建新的命令管理器实�?func NewCommandManager(s *scheduler.Scheduler, eb *core.EventBus) *CommandManager {
	cm := &CommandManager{
		registeredCommands: make(map[string]*Command),
		commands:          make(map[string]*Command),
		presetCommands:    make(map[string]*Command),
		pluginCommands:    make(map[string]*Command),
		otherCommands:     make(map[string]*Command),
		scheduler:         s,
		eventBus:          eb,
	}
	
	// 初始化命令链
	cm.commandChain = NewCommandChain(cm)
	
	// 初始化内建命�?	cm.initPresetCommands()
	
	// 注册事件监听�?	cm.registerEventListeners()
	
	return cm
}

// initPresetCommands 初始化内建命�?func (cm *CommandManager) initPresetCommands() {
	cm.presetCommands["/cookiecloud"] = &Command{
		ID:          "cookiecloud",
		Type:        "scheduler",
		Description: "同步站点",
		Category:    "站点",
		Show:        true,
	}
	
	cm.presetCommands["/sites"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "查询站点",
		Category:    "站点",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行查询站点命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/site_cookie"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "更新站点Cookie",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行更新站点Cookie命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/site_statistic"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "站点数据统计",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行站点数据统计命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/site_enable"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "启用站点",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行启用站点命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/site_disable"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "禁用站点",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行禁用站点命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/mediaserver_sync"] = &Command{
		ID:          "mediaserver_sync",
		Type:        "scheduler",
		Description: "同步媒体服务�?,
		Category:    "管理",
		Show:        true,
	}
	
	cm.presetCommands["/subscribes"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "查询订阅",
		Category:    "订阅",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行查询订阅命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/subscribe_refresh"] = &Command{
		ID:          "subscribe_refresh",
		Type:        "scheduler",
		Description: "刷新订阅",
		Category:    "订阅",
		Show:        true,
	}
	
	cm.presetCommands["/subscribe_search"] = &Command{
		ID:          "subscribe_search",
		Type:        "scheduler",
		Description: "搜索订阅",
		Category:    "订阅",
		Show:        true,
	}
	
	cm.presetCommands["/subscribe_delete"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "删除订阅",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行删除订阅命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/subscribe_tmdb"] = &Command{
		ID:          "subscribe_tmdb",
		Type:        "scheduler",
		Description: "订阅元数据更�?,
		Show:        true,
	}
	
	cm.presetCommands["/downloading"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "正在下载",
		Category:    "管理",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行正在下载命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/transfer"] = &Command{
		ID:          "transfer",
		Type:        "scheduler",
		Description: "下载文件整理",
		Category:    "管理",
		Show:        true,
	}
	
	cm.presetCommands["/redo"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "手动整理",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行手动整理命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/clear_cache"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "清理缓存",
		Category:    "管理",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行清理缓存命令")
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/restart"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "重启系统",
		Category:    "管理",
		Function: func(args ...interface{}) error {
			logger.Log.Info("执行重启系统命令")
			// 实际重启逻辑将在后续实现
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
	
	cm.presetCommands["/version"] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: "当前版本",
		Category:    "管理",
		Function: func(args ...interface{}) error {
			logger.Log.Infof("当前版本: %s", config.Version)
			return nil
		},
		Data: make(map[string]interface{}),
		Show: true,
	}
}

// registerEventListeners 注册事件监听�?func (cm *CommandManager) registerEventListeners() {
	// 注册命令执行事件监听�?	cm.eventBus.Subscribe("CommandExecute", cm.commandEventListener)
	
	// 注册模块重载事件监听�?	cm.eventBus.Subscribe("ModuleReload", cm.moduleReloadEventListener)
}

// commandEventListener 命令执行事件监听�?func (cm *CommandManager) commandEventListener(e *core.Event) error {
	eventData := e.Data
	
	// 命令参数
	eventStr, _ := eventData["cmd"].(string)
	// 消息渠道
	eventChannel, _ := eventData["channel"].(string)
	// 消息来源
	eventSource, _ := eventData["source"].(string)
	// 消息用户
	eventUser := eventData["user"]
	
	if eventStr != "" {
		// 分离命令和参�?		parts := strings.Fields(eventStr)
		cmd := parts[0]
		args := ""
		if len(parts) > 1 {
			args = strings.Join(parts[1:], " ")
		}
		
		if _, exists := cm.Get(cmd); exists {
			cm.Execute(cmd, args, eventChannel, eventSource, eventUser)
		}
	}
	
	return nil
}

// moduleReloadEventListener 模块重载事件监听�?func (cm *CommandManager) moduleReloadEventListener(e *core.Event) error {
	// 发生模块重载时，重新注册命令
	cm.InitCommands("")
	return nil
}

// InitCommands 初始化菜单命�?func (cm *CommandManager) InitCommands(pid string) {
	// 使用goroutine后台初始化命令，避免阻塞
	go cm.initCommandsBackground(pid)
}

// initCommandsBackground 后台初始化菜单命�?func (cm *CommandManager) initCommandsBackground(pid string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	logger.Log.Debug("获取锁以在后台初始化命令")
	
	// 构建插件命令
	cm.buildPluginCommands(pid)
	
	// 合并所有命�?	cm.commands = make(map[string]*Command)
	for k, v := range cm.presetCommands {
		cm.commands[k] = v
	}
	for k, v := range cm.pluginCommands {
		cm.commands[k] = v
	}
	for k, v := range cm.otherCommands {
		cm.commands[k] = v
	}
	
	// 强制触发注册
	forceRegister := false
	// 触发事件允许可以拦截和调整命�?	event, initialCommands := cm.triggerRegisterCommandsEvent()
	
	// 检查事件是否被取消
	if event != nil && event.Data != nil {
		// 如果事件返回有效�?event_data，使用事件中调整后的命令
		eventData := event.Data
		// 如果事件被取消，跳过命令注册
		if cancel, exists := eventData["cancel"].(bool); exists && cancel {
			if source, exists := eventData["source"].(string); exists {
				logger.Log.Debugf("Command initialization canceled by event: %s", source)
			}
			return
		}
		// 如果拦截源与插件标识一致时，这里认为需要强制触发注�?		if pid != "" {
			if source, exists := eventData["source"].(string); exists && pid == source {
				forceRegister = true
			}
		}
		if cmds, exists := eventData["commands"].(map[string]*Command); exists {
			initialCommands = cmds
		}
		logger.Log.Debugf("Registering command count from event: %d", len(initialCommands))
	} else {
		logger.Log.Debugf("Registering initial command count: %d", len(initialCommands))
	}
	
	// initial_commands 必须�?cm.commands 的子�?	filteredInitialCommands := cm.filterKeysToSubset(initialCommands, cm.commands)
	// 如果 filtered_initial_commands 为空，则跳过注册
	if len(filteredInitialCommands) == 0 && !forceRegister {
		logger.Log.Debug("Filtered commands are empty, skipping registration.")
		return
	}
	
	// 对比调整后的命令与当前命�?	if !cm.commandsEqual(filteredInitialCommands, cm.registeredCommands) || forceRegister {
		logger.Log.Debug("Command set has changed or force registration is enabled.")
		cm.registeredCommands = filteredInitialCommands
		cm.commandChain.RegisterCommands(filteredInitialCommands)
	} else {
		logger.Log.Debug("Command set unchanged, skipping broadcast registration.")
	}
}

// filterKeysToSubset 过滤命令键为子集
func (cm *CommandManager) filterKeysToSubset(source, target map[string]*Command) map[string]*Command {
	result := make(map[string]*Command)
	for key := range target {
		if cmd, exists := source[key]; exists {
			result[key] = cmd
		}
	}
	return result
}

// commandsEqual 比较两个命令集合是否相等
func (cm *CommandManager) commandsEqual(a, b map[string]*Command) bool {
	if len(a) != len(b) {
		return false
	}
	
	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}
		if valueA.ID != valueB.ID || valueA.Type != valueB.Type || 
		   valueA.Description != valueB.Description || valueA.Category != valueB.Category {
			return false
		}
	}
	return true
}

// triggerRegisterCommandsEvent 触发命令注册事件
func (cm *CommandManager) triggerRegisterCommandsEvent() (*core.Event, map[string]*Command) {
	// 构建初始命令集合
	commands := make(map[string]*Command)
	
	// 添加内建命令
	for cmd, command := range cm.presetCommands {
		if command.Show {
			commands[cmd] = &Command{
				Type:        "preset",
				Description: command.Description,
				Category:    command.Category,
				PID:         command.PID,
			}
		}
	}
	
	// 添加插件命令
	for cmd, command := range cm.pluginCommands {
		if command.Show {
			commands[cmd] = &Command{
				Type:        "plugin",
				Description: command.Description,
				Category:    command.Category,
				PID:         command.PID,
			}
		}
	}
	
	// 添加其他命令
	for cmd, command := range cm.otherCommands {
		if command.Show {
			commands[cmd] = &Command{
				Type:        "other",
				Description: command.Description,
				Category:    command.Category,
			}
		}
	}
	
	// 触发事件允许调整命令数据
	eventData := map[string]interface{}{
		"commands": commands,
		"origin":   "CommandChain",
		"service":  nil,
	}
	
	event := &core.Event{
		Type: "CommandRegister",
		Data: eventData,
	}
	
	// 发布事件
	cm.eventBus.Publish("CommandRegister", eventData)
	
	return event, commands
}

// buildPluginCommands 构建插件命令
func (cm *CommandManager) buildPluginCommands(pid string) {
	// TODO: 实现插件命令构建逻辑
	// 这里暂时留空，后续根据插件系统实�?	// 为了保证命令顺序的一致性，目前这里没有直接使用 pid 获取单一插件命令，后续如果存在性能问题，可以考虑优化这里的逻辑
	cm.pluginCommands = make(map[string]*Command)
	
	// 模拟从插件管理器获取命令
	// pluginCommands := cm.pluginManager.GetPluginCommands()
	// for _, command := range pluginCommands {
	// 	cmd := command.Get("cmd")
	// 	if cmd != "" {
	// 		cm.pluginCommands[cmd] = &Command{
	// 			ID:          uuid.New().String(),
	// 			PID:         command.Get("pid"),
	// 			Type:        "command",
	// 			Function:    cm.SendPluginEvent,
	// 			Description: command.Get("desc"),
	// 			Category:    command.Get("category"),
	// 			Show:        command.Get("show", true),
	// 			Data: map[string]interface{}{
	// 				"etype": command.Get("event"),
	// 				"data":  command.Get("data"),
	// 			},
	// 		}
	// 	}
	// }
}

// GetCommands 获取命令列表
func (cm *CommandManager) GetCommands() map[string]*Command {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	// 返回命令副本以防止并发修�?	commands := make(map[string]*Command)
	for k, v := range cm.commands {
		commands[k] = v
	}
	return commands
}

// Get 获取指定命令
func (cm *CommandManager) Get(cmd string) (*Command, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	command, exists := cm.commands[cmd]
	return command, exists
}

// Register 注册单个命令
func (cm *CommandManager) Register(cmd string, function func(args ...interface{}) error, 
	data map[string]interface{}, desc, category string, show bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// 单独调用的，统一注册到其他命令中
	cm.otherCommands[cmd] = &Command{
		ID:          uuid.New().String(),
		Type:        "command",
		Description: desc,
		Category:    category,
		Function:    function,
		Data:        data,
		Show:        show,
	}
}

// Execute 执行命令
func (cm *CommandManager) Execute(cmd, dataStr, channel, source string, userID interface{}) {
	command, exists := cm.Get(cmd)
	if !exists {
		logger.Log.Errorf("命令不存�? %s", cmd)
		return
	}
	
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string) // 简化处理，实际可能需要类型断言
		logger.Log.Infof("用户 %s 开始执行：%s ...", userIDStr, command.Description)
	} else {
		logger.Log.Infof("开始执行：%s ...", command.Description)
	}
	
	// 执行命令
	err := cm.runCommand(command, dataStr, channel, source, userID)
	if err != nil {
		logger.Log.Errorf("执行命令 %s 出错�?s", cmd, err.Error())
		// TODO: 发送错误消�?		return
	}
	
	if userID != nil {
		logger.Log.Infof("用户 %s %s 执行完成", userIDStr, command.Description)
	} else {
		logger.Log.Infof("%s 执行完成", command.Description)
	}
}

// runCommand 运行命令
func (cm *CommandManager) runCommand(command *Command, dataStr, channel, source string, userID interface{}) error {
	if command.Type == "scheduler" {
		// 定时服务
		if userID != nil {
			cm.commandChain.PostMessage(channel, source, "开始执�?"+command.Description+" ...", userID)
		}
		
		// 执行定时任务
		cm.scheduler.Start(command.ID)
		
		if userID != nil {
			cm.commandChain.PostMessage(channel, source, command.Description+" 执行完成", userID)
		}
	} else {
		// 普通命�?		var args []interface{}
		
		if len(command.Data) > 0 {
			// 有内置参数直接使用内置参�?			data := make(map[string]interface{})
			for k, v := range command.Data {
				data[k] = v
			}
			
			data["channel"] = channel
			data["source"] = source
			data["user"] = userID
			
			if dataStr != "" {
				data["arg_str"] = dataStr
			}
			
			args = append(args, data)
		} else if command.Function != nil {
			// 根据函数签名传递不同参�?			if dataStr != "" || channel != "" || source != "" || userID != nil {
				// 检查函数参数数�?				funcType := reflect.TypeOf(command.Function)
				if funcType.Kind() == reflect.Func {
					numIn := funcType.NumIn()
					if numIn >= 4 {
						args = append(args, dataStr, channel, userID, source)
					} else if numIn == 3 {
						args = append(args, channel, userID, source)
					} else if numIn > 0 {
						args = append(args, dataStr)
					}
				}
			}
		}
		
		// 执行命令函数
		if command.Function != nil {
			return command.Function(args...)
		}
	}
	
	return nil
}

// SendPluginEvent 发送插件命令事�?func (cm *CommandManager) SendPluginEvent(eventType string, data map[string]interface{}) {
	cm.eventBus.Publish(eventType, data)
}
