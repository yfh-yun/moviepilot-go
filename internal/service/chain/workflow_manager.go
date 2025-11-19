package chain

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/repository"
	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
)

// WorkflowManager 工作流管理器
type WorkflowManager struct {
	cache           *cache.Cache
	logger          *logger.Logger
	workflowRepo    *repository.WorkflowRepository
	workflowService *service.WorkflowService

	// 动作定义缓存
	actions     map[string]WorkflowAction
	actionsLock sync.RWMutex

	// 事件触发器映射
	eventWorkflows     map[string][]int64
	eventWorkflowsLock sync.RWMutex

	// 运行中的工作流状态
	runningWorkflows map[int64]*WorkflowExecutionContext
	runningLock      sync.RWMutex

	// 全局停止标志
	stoppedWorkflows map[int64]bool
	stopLock         sync.RWMutex
}

// WorkflowAction 工作流动作接口
type WorkflowAction interface {
	// GetID 获取动作ID
	GetID() string

	// GetName 获取动作名称
	GetName() string

	// GetDescription 获取动作描述
	GetDescription() string

	// GetDataSchema 获取数据模式
	GetDataSchema() map[string]interface{}

	// Execute 执行动作
	Execute(ctx context.Context, workflowID int64, data map[string]interface{},
		executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error)

	// IsDone 检查动作是否完成
	IsDone() bool

	// GetSuccess 获取执行结果
	GetSuccess() bool

	// GetMessage 获取执行消息
	GetMessage() string
}

// WorkflowExecutionContext 工作流执行上下文
type WorkflowExecutionContext struct {
	WorkflowID     int64                  `json:"workflow_id"`
	CurrentAction  string                 `json:"current_action"`
	Context        map[string]interface{} `json:"context"`
	Metadata       map[string]interface{} `json:"metadata"`
	StartTime      time.Time              `json:"start_time"`
	LastUpdateTime time.Time              `json:"last_update_time"`
	ExecutionCount int                    `json:"execution_count"`
	LoopCount      int                    `json:"loop_count"`
	IsCompleted    bool                   `json:"is_completed"`
	Success        bool                   `json:"success"`
	Message        string                 `json:"message"`
}

// WorkflowActionResult 工作流动作执行结果
type WorkflowActionResult struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data"`
	NextAction   string                 `json:"next_action"`
	ShouldLoop   bool                   `json:"should_loop"`
	LoopInterval time.Duration          `json:"loop_interval"`
	Context      map[string]interface{} `json:"context"`
}

// WorkflowEventTrigger 工作流事件触发器
type WorkflowEventTrigger struct {
	WorkflowID      int64                  `json:"workflow_id"`
	EventType       string                 `json:"event_type"`
	EventConditions map[string]interface{} `json:"event_conditions"`
	IsEnabled       bool                   `json:"is_enabled"`
	TriggerCount    int64                  `json:"trigger_count"`
	LastTriggered   time.Time              `json:"last_triggered"`
}

// NewWorkflowManager 创建工作流管理器
func NewWorkflowManager(cache *cache.Cache, logger *logger.Logger,
	workflowRepo *repository.WorkflowRepository) *WorkflowManager {

	manager := &WorkflowManager{
		cache:            cache,
		logger:           logger,
		workflowRepo:     workflowRepo,
		workflowService:  service.NewWorkflowService(workflowRepo, logger),
		actions:          make(map[string]WorkflowAction),
		eventWorkflows:   make(map[string][]int64),
		runningWorkflows: make(map[int64]*WorkflowExecutionContext),
		stoppedWorkflows: make(map[int64]bool),
	}

	// 初始化动作和事件
	if err := manager.Initialize(); err != nil {
		logger.Error("初始化工作流管理器失败", "error", err)
	}

	return manager
}

// Initialize 初始化工作流管理器
func (wm *WorkflowManager) Initialize() error {
	wm.logger.Info("初始化工作流管理器")

	// 加载所有动作
	if err := wm.loadActions(); err != nil {
		return fmt.Errorf("加载动作失败: %w", err)
	}

	// 加载事件触发器
	if err := wm.loadWorkflowEvents(); err != nil {
		return fmt.Errorf("加载事件触发器失败: %w", err)
	}

	wm.logger.Info("工作流管理器初始化完成",
		"actions_count", len(wm.actions),
		"event_workflows_count", len(wm.eventWorkflows))

	return nil
}

// loadActions 加载所有工作流动作
func (wm *WorkflowManager) loadActions() error {
	wm.actionsLock.Lock()
	defer wm.actionsLock.Unlock()

	// 注册内置动作
	builtInActions := []WorkflowAction{
		&AddDownloadAction{},
		&AddSubscribeAction{},
		&FetchMediasAction{},
		&FilterMediasAction{},
		&SendMessageAction{},
		&TransferFileAction{},
		&ScanFileAction{},
		&ScrapeFileAction{},
		&InvokePluginAction{},
		&SendEventAction{},
	}

	for _, action := range builtInActions {
		wm.actions[action.GetID()] = action
		wm.logger.Debug("注册工作流动作",
			"action_id", action.GetID(),
			"name", action.GetName())
	}

	return nil
}

// loadWorkflowEvents 加载工作流事件触发器
func (wm *WorkflowManager) loadWorkflowEvents() error {
	wm.eventWorkflowsLock.Lock()
	defer wm.eventWorkflowsLock.Unlock()

	// 清空现有事件映射
	wm.eventWorkflows = make(map[string][]int64)

	// 获取所有事件触发的工作流
	workflows, err := wm.workflowRepo.GetEventTriggeredWorkflows(context.Background())
	if err != nil {
		return fmt.Errorf("获取事件触发工作流失败: %w", err)
	}

	for _, workflow := range workflows {
		if workflow.TriggerType == "event" && workflow.State != "P" {
			// 注册事件触发器
			wm.registerWorkflowEvent(workflow.ID, workflow.EventType)
		}
	}

	return nil
}

// ExecuteAction 执行工作流动作
func (wm *WorkflowManager) ExecuteAction(ctx context.Context, workflowID int64,
	action *model.WorkflowAction, executeContext *WorkflowExecutionContext) (*WorkflowActionResult, error) {

	if executeContext == nil {
		executeContext = &WorkflowExecutionContext{
			WorkflowID:     workflowID,
			Context:        make(map[string]interface{}),
			Metadata:       make(map[string]interface{}),
			StartTime:      time.Now(),
			LastUpdateTime: time.Now(),
		}
	}

	wm.actionsLock.RLock()
	actionObj, exists := wm.actions[action.Type]
	wm.actionsLock.RUnlock()

	if !exists {
		return nil, fmt.Errorf("未找到动作类型: %s", action.Type)
	}

	wm.logger.Info("执行工作流动作",
		"workflow_id", workflowID,
		"action_id", action.ID,
		"action_type", action.Type,
		"action_name", action.Name)

	// 实例化动作对象（注意：这里需要根据实际的action创建新的实例）
	// 简化起见，直接使用已注册的动作对象

	// 执行动作
	result, err := actionObj.Execute(ctx, workflowID, action.Data, executeContext)
	if err != nil {
		wm.logger.Error("工作流动作执行失败",
			"workflow_id", workflowID,
			"action_id", action.ID,
			"error", err)
		return nil, err
	}

	// 检查是否需要循环执行
	if result.ShouldLoop && result.LoopInterval > 0 {
		wm.logger.Info("工作流动作进入循环模式",
			"workflow_id", workflowID,
			"action_id", action.ID,
			"loop_interval", result.LoopInterval)

		for !actionObj.IsDone() {
			// 检查是否被停止
			if wm.IsWorkflowStopped(workflowID) {
				wm.logger.Info("工作流被停止，退出循环",
					"workflow_id", workflowID,
					"action_id", action.ID)
				break
			}

			// 等待循环间隔
			time.Sleep(result.LoopInterval)

			// 继续执行
			wm.logger.Info("继续执行循环动作",
				"workflow_id", workflowID,
				"action_id", action.ID)

			loopResult, err := actionObj.Execute(ctx, workflowID, action.Data, executeContext)
			if err != nil {
				wm.logger.Error("循环动作执行失败",
					"workflow_id", workflowID,
					"action_id", action.ID,
					"error", err)
				break
			}

			result = loopResult
		}
	}

	// 更新执行上下文
	if result.Context != nil {
		for key, value := range result.Context {
			executeContext.Context[key] = value
		}
	}
	executeContext.LastUpdateTime = time.Now()
	executeContext.ExecutionCount++

	if actionObj.GetSuccess() {
		wm.logger.Info("工作流动作执行成功",
			"workflow_id", workflowID,
			"action_id", action.ID,
			"message", actionObj.GetMessage())
	} else {
		wm.logger.Error("工作流动作执行失败",
			"workflow_id", workflowID,
			"action_id", action.ID,
			"message", actionObj.GetMessage())
	}

	return result, nil
}

// GetListActions 获取所有可用动作列表
func (wm *WorkflowManager) GetListActions() []map[string]interface{} {
	wm.actionsLock.RLock()
	defer wm.actionsLock.RUnlock()

	var actions []map[string]interface{}
	for id, action := range wm.actions {
		actions = append(actions, map[string]interface{}{
			"type":        id,
			"name":        action.GetName(),
			"description": action.GetDescription(),
			"data": map[string]interface{}{
				"label":  action.GetName(),
				"schema": action.GetDataSchema(),
			},
		})
	}

	return actions
}

// UpdateWorkflowEvent 更新工作流事件触发器
func (wm *WorkflowManager) UpdateWorkflowEvent(workflow *model.Workflow) error {
	// 先移除旧的事件监听器
	wm.removeWorkflowEvent(workflow.ID, workflow.EventType)

	// 如果工作流是事件触发类型且未被禁用
	if workflow.TriggerType == "event" && workflow.State != "P" {
		// 注册事件触发器
		return wm.registerWorkflowEvent(workflow.ID, workflow.EventType)
	}

	return nil
}

// loadWorkflowEvents 加载工作流事件（支持单个工作流）
func (wm *WorkflowManager) LoadWorkflowEvents(workflowID *int64) error {
	var workflows []*model.Workflow
	var err error

	if workflowID != nil {
		// 加载指定工作流
		workflow, err := wm.workflowRepo.GetWorkflowByID(context.Background(), *workflowID)
		if err != nil {
			return fmt.Errorf("获取工作流失败: %w", err)
		}
		if workflow != nil {
			workflows = []*model.Workflow{workflow}
		}
	} else {
		// 加载所有事件触发的工作流
		workflows, err = wm.workflowRepo.GetEventTriggeredWorkflows(context.Background())
		if err != nil {
			return fmt.Errorf("获取事件触发工作流失败: %w", err)
		}
	}

	for _, workflow := range workflows {
		if err := wm.UpdateWorkflowEvent(workflow); err != nil {
			wm.logger.Error("更新工作流事件失败",
				"workflow_id", workflow.ID,
				"error", err)
		}
	}

	return nil
}

// registerWorkflowEvent 注册工作流事件触发器
func (wm *WorkflowManager) registerWorkflowEvent(workflowID int64, eventType string) error {
	wm.eventWorkflowsLock.Lock()
	defer wm.eventWorkflowsLock.Unlock()

	// 确保先移除旧的监听器
	wm.removeWorkflowEventUnsafe(workflowID, eventType)

	// 添加新的事件监听器
	if wm.eventWorkflows[eventType] == nil {
		wm.eventWorkflows[eventType] = []int64{}
	}

	wm.eventWorkflows[eventType] = append(wm.eventWorkflows[eventType], workflowID)

	wm.logger.Info("已注册工作流事件触发器",
		"workflow_id", workflowID,
		"event_type", eventType)

	return nil
}

// removeWorkflowEvent 移除工作流事件触发器
func (wm *WorkflowManager) removeWorkflowEvent(workflowID int64, eventType string) {
	wm.eventWorkflowsLock.Lock()
	defer wm.eventWorkflowsLock.Unlock()
	wm.removeWorkflowEventUnsafe(workflowID, eventType)
}

// removeWorkflowEventUnsafe 非线程安全的移除工作流事件触发器
func (wm *WorkflowManager) removeWorkflowEventUnsafe(workflowID int64, eventType string) {
	if workflows, exists := wm.eventWorkflows[eventType]; exists {
		for i, id := range workflows {
			if id == workflowID {
				// 移除工作流ID
				wm.eventWorkflows[eventType] = append(workflows[:i], workflows[i+1:]...)
				break
			}
		}

		// 如果没有工作流使用此事件，移除事件类型
		if len(wm.eventWorkflows[eventType]) == 0 {
			delete(wm.eventWorkflows, eventType)
		}
	}

	wm.logger.Info("已移除工作流事件触发器",
		"workflow_id", workflowID,
		"event_type", eventType)
}

// HandleEvent 处理事件，触发相应的工作流
func (wm *WorkflowManager) HandleEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error {
	wm.eventWorkflowsLock.RLock()
	workflows, exists := wm.eventWorkflows[eventType]
	wm.eventWorkflowsLock.RUnlock()

	if !exists {
		return nil // 没有注册此事件的工作流
	}

	// 复制工作流列表，避免并发问题
	workflowIDs := make([]int64, len(workflows))
	copy(workflowIDs, workflows)

	for _, workflowID := range workflowIDs {
		if err := wm.triggerWorkflow(ctx, workflowID, eventType, eventData); err != nil {
			wm.logger.Error("触发工作流失败",
				"workflow_id", workflowID,
				"event_type", eventType,
				"error", err)
		}
	}

	return nil
}

// triggerWorkflow 触发工作流执行
func (wm *WorkflowManager) triggerWorkflow(ctx context.Context, workflowID int64,
	eventType string, eventData map[string]interface{}) error {

	// 检查工作流是否存在且启用
	workflow, err := wm.workflowRepo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("获取工作流失败: %w", err)
	}

	if workflow == nil || workflow.State == "P" {
		wm.logger.Debug("工作流不存在或已禁用", "workflow_id", workflowID)
		return nil
	}

	// 检查事件条件
	if !wm.checkEventConditions(workflow.EventConditions, eventData) {
		wm.logger.Debug("工作流事件条件不匹配，跳过执行",
			"workflow_id", workflowID,
			"workflow_name", workflow.Name)
		return nil
	}

	// 检查工作流是否正在运行
	wm.runningLock.RLock()
	_, isRunning := wm.runningWorkflows[workflowID]
	wm.runningLock.RUnlock()

	if isRunning {
		wm.logger.Warning("工作流正在运行中，跳过重复触发",
			"workflow_id", workflowID,
			"workflow_name", workflow.Name)
		return nil
	}

	wm.logger.Info("事件触发工作流执行",
		"event_type", eventType,
		"workflow_id", workflowID,
		"workflow_name", workflow.Name)

	// 发送工作流执行事件
	go func() {
		// 这里可以通过事件总线发送工作流执行事件
		// 或者直接调用工作流执行方法
		_, err := wm.ExecuteWorkflow(ctx, workflowID, false)
		if err != nil {
			wm.logger.Error("事件触发工作流执行失败",
				"workflow_id", workflowID,
				"error", err)
		}
	}()

	return nil
}

// checkEventConditions 检查事件条件
func (wm *WorkflowManager) checkEventConditions(conditions map[string]interface{},
	eventData map[string]interface{}) bool {

	if conditions == nil || len(conditions) == 0 {
		return true
	}

	for field, expectedValue := range conditions {
		actualValue, exists := eventData[field]
		if !exists {
			return false
		}

		// 支持多种条件匹配方式
		if expectedValueMap, ok := expectedValue.(map[string]interface{}); ok {
			// 复杂条件匹配
			if !wm.checkComplexCondition(actualValue, expectedValueMap) {
				return false
			}
		} else {
			// 简单值匹配
			if actualValue != expectedValue {
				return false
			}
		}
	}

	return true
}

// checkComplexCondition 检查复杂条件匹配
func (wm *WorkflowManager) checkComplexCondition(actualValue interface{},
	condition map[string]interface{}) bool {

	for operator, expectedValue := range condition {
		switch operator {
		case "equals":
			if actualValue != expectedValue {
				return false
			}
		case "not_equals":
			if actualValue == expectedValue {
				return false
			}
		case "contains":
			if actualValueStr, ok := actualValue.(string); ok {
				if expectedValueStr, ok := expectedValue.(string); ok {
					if !contains(actualValueStr, expectedValueStr) {
						return false
					}
				}
			} else {
				// 转换为字符串比较
				actualStr := fmt.Sprintf("%v", actualValue)
				expectedStr := fmt.Sprintf("%v", expectedValue)
				if !contains(actualStr, expectedStr) {
					return false
				}
			}
		case "not_contains":
			if actualValueStr, ok := actualValue.(string); ok {
				if expectedValueStr, ok := expectedValue.(string); ok {
					if contains(actualValueStr, expectedValueStr) {
						return false
					}
				}
			} else {
				// 转换为字符串比较
				actualStr := fmt.Sprintf("%v", actualValue)
				expectedStr := fmt.Sprintf("%v", expectedValue)
				if contains(actualStr, expectedStr) {
					return false
				}
			}
		case "in":
			if expectedSlice, ok := expectedValue.([]interface{}); ok {
				found := false
				for _, val := range expectedSlice {
					if actualValue == val {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		case "not_in":
			if expectedSlice, ok := expectedValue.([]interface{}); ok {
				for _, val := range expectedSlice {
					if actualValue == val {
						return false
					}
				}
			}
		case "regex":
			// 这里可以实现正则表达式匹配
			// 简化起见，暂时跳过
		}
	}

	return true
}

// GetEventWorkflows 获取所有事件触发的工作流
func (wm *WorkflowManager) GetEventWorkflows() map[string][]int64 {
	wm.eventWorkflowsLock.RLock()
	defer wm.eventWorkflowsLock.RUnlock()

	result := make(map[string][]int64)
	for eventType, workflowIDs := range wm.eventWorkflows {
		result[eventType] = make([]int64, len(workflowIDs))
		copy(result[eventType], workflowIDs)
	}

	return result
}

// StopWorkflow 停止工作流
func (wm *WorkflowManager) StopWorkflow(workflowID int64) {
	wm.stopLock.Lock()
	defer wm.stopLock.Unlock()

	wm.stoppedWorkflows[workflowID] = true
	wm.logger.Info("设置工作流停止标志", "workflow_id", workflowID)
}

// IsWorkflowStopped 检查工作流是否被停止
func (wm *WorkflowManager) IsWorkflowStopped(workflowID int64) bool {
	wm.stopLock.RLock()
	defer wm.stopLock.RUnlock()

	return wm.stoppedWorkflows[workflowID]
}

// ClearWorkflowStopFlag 清除工作流停止标志
func (wm *WorkflowManager) ClearWorkflowStopFlag(workflowID int64) {
	wm.stopLock.Lock()
	defer wm.stopLock.Unlock()

	delete(wm.stoppedWorkflows, workflowID)
}

// GetRunningWorkflows 获取运行中的工作流
func (wm *WorkflowManager) GetRunningWorkflows() map[int64]*WorkflowExecutionContext {
	wm.runningLock.RLock()
	defer wm.runningLock.RUnlock()

	result := make(map[int64]*WorkflowExecutionContext)
	for workflowID, context := range wm.runningWorkflows {
		// 深拷贝上下文
		contextCopy := *context
		result[workflowID] = &contextCopy
	}

	return result
}

// SetRunningWorkflow 设置运行中的工作流
func (wm *WorkflowManager) SetRunningWorkflow(workflowID int64, context *WorkflowExecutionContext) {
	wm.runningLock.Lock()
	defer wm.runningLock.Unlock()

	wm.runningWorkflows[workflowID] = context
}

// RemoveRunningWorkflow 移除运行中的工作流
func (wm *WorkflowManager) RemoveRunningWorkflow(workflowID int64) {
	wm.runningLock.Lock()
	defer wm.runningLock.Unlock()

	delete(wm.runningWorkflows, workflowID)
}

// ExecuteWorkflow 执行完整的工作流
func (wm *WorkflowManager) ExecuteWorkflow(ctx context.Context, workflowID int64, fromBeginning bool) (*model.WorkflowExecutionResult, error) {
	// 获取工作流信息
	workflow, err := wm.workflowRepo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("获取工作流失败: %w", err)
	}

	if workflow == nil {
		return nil, errors.New("工作流不存在")
	}

	// 验证工作流
	if err := wm.validateWorkflow(workflow); err != nil {
		return nil, err
	}

	// 创建执行上下文
	executeContext := &WorkflowExecutionContext{
		WorkflowID:     workflowID,
		Context:        make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
	}

	// 如果需要从头开始，重置工作流状态
	if fromBeginning {
		if err := wm.workflowService.ResetWorkflow(ctx, workflowID); err != nil {
			return nil, fmt.Errorf("重置工作流失败: %w", err)
		}
	}

	// 设置工作流为运行状态
	wm.SetRunningWorkflow(workflowID, executeContext)
	defer wm.RemoveRunningWorkflow(workflowID)

	// 清除停止标志
	defer wm.ClearWorkflowStopFlag(workflowID)

	// 执行工作流
	result, err := wm.executeWorkflowInternal(ctx, workflow, executeContext)
	if err != nil {
		wm.logger.Error("执行工作流失败",
			"workflow_id", workflowID,
			"error", err)
		return nil, err
	}

	wm.logger.Info("工作流执行完成",
		"workflow_id", workflowID,
		"success", result.Success,
		"message", result.Message)

	return result, nil
}

// executeWorkflowInternal 内部工作流执行逻辑
func (wm *WorkflowManager) executeWorkflowInternal(ctx context.Context, workflow *model.Workflow,
	executeContext *WorkflowExecutionContext) (*model.WorkflowExecutionResult, error) {

	// 更新工作流状态为运行中
	if err := wm.workflowRepo.UpdateWorkflowState(ctx, workflow.ID, "R"); err != nil {
		wm.logger.Error("更新工作流状态失败", "error", err)
	}

	// 这里应该根据工作流的flows定义来执行动作序列
	// 简化起见，我们假设flows定义了动作的执行顺序

	var lastActionID string
	var allSuccess = true
	var executionMessage string

	// 遍历所有动作（这里需要根据实际的flows结构来实现）
	for _, action := range workflow.Actions {
		// 检查是否被停止
		if wm.IsWorkflowStopped(workflow.ID) {
			executionMessage = "工作流被停止"
			allSuccess = false
			break
		}

		// 执行动作
		result, err := wm.ExecuteAction(ctx, workflow.ID, &action, executeContext)
		if err != nil {
			executionMessage = fmt.Sprintf("动作执行失败: %v", err)
			allSuccess = false
			break
		}

		if !result.Success {
			executionMessage = result.Message
			allSuccess = false
			break
		}

		lastActionID = action.ID

		// 如果有下一个动作，继续执行
		if result.NextAction != "" {
			// 这里可以实现跳转到指定动作的逻辑
		}
	}

	// 更新工作流最终状态
	finalState := "F" // 失败
	if allSuccess {
		finalState = "S" // 成功
		executionMessage = "工作流执行成功"
	}

	if err := wm.workflowRepo.UpdateWorkflowState(ctx, workflow.ID, finalState); err != nil {
		wm.logger.Error("更新工作流最终状态失败", "error", err)
	}

	return &model.WorkflowExecutionResult{
		WorkflowID:    workflow.ID,
		Success:       allSuccess,
		Message:       executionMessage,
		ExecutionTime: time.Since(executeContext.StartTime),
		LastActionID:  lastActionID,
		Context:       executeContext.Context,
	}, nil
}

// validateWorkflow 验证工作流
func (wm *WorkflowManager) validateWorkflow(workflow *model.Workflow) error {
	if workflow == nil {
		return errors.New("工作流不能为空")
	}

	if len(workflow.Actions) == 0 {
		return errors.New("工作流无动作")
	}

	if len(workflow.Flows) == 0 {
		return errors.New("工作流无流程")
	}

	return nil
}

// Stop 停止工作流管理器
func (wm *WorkflowManager) Stop() {
	wm.logger.Info("停止工作流管理器")

	wm.actionsLock.Lock()
	wm.actions = make(map[string]WorkflowAction)
	wm.actionsLock.Unlock()

	wm.eventWorkflowsLock.Lock()
	wm.eventWorkflows = make(map[string][]int64)
	wm.eventWorkflowsLock.Unlock()

	wm.runningLock.Lock()
	wm.runningWorkflows = make(map[int64]*WorkflowExecutionContext)
	wm.runningLock.Unlock()

	wm.stopLock.Lock()
	wm.stoppedWorkflows = make(map[int64]bool)
	wm.stopLock.Unlock()
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
