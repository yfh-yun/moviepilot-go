package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/repository/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WorkflowStatus 表示工作流状态
type WorkflowStatus string

const (
	WorkflowStatusDraft    WorkflowStatus = "draft"
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusInactive WorkflowStatus = "inactive"
	WorkflowStatusError    WorkflowStatus = "error"
)

// TriggerType 表示触发器类型
type TriggerType string

const (
	TriggerTypeSchedule TriggerType = "schedule"
	TriggerTypeEvent    TriggerType = "event"
	TriggerTypeManual   TriggerType = "manual"
)

// ActionType 表示动作类型
type ActionType string

const (
	ActionTypeAddSubscribe  ActionType = "add_subscribe"
	ActionTypeStartDownload ActionType = "start_download"
	ActionTypeSendMessage   ActionType = "send_message"
	ActionTypeCallPlugin    ActionType = "call_plugin"
	ActionTypeRefreshMedia  ActionType = "refresh_media"
	ActionTypeCustomScript  ActionType = "custom_script"
)

// ConditionType 表示条件类型
type ConditionType string

const (
	ConditionTypeMediaMatch  ConditionType = "media_match"
	ConditionTypeTimeRange   ConditionType = "time_range"
	ConditionTypePluginCheck ConditionType = "plugin_check"
	ConditionTypeCustom      ConditionType = "custom"
)

// WorkflowInstance 表示工作流实例
type WorkflowInstance struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Name        string                 `json:"name"`
	Status      WorkflowStatus         `json:"status"`
	CurrentStep int                    `json:"current_step"`
	TotalSteps  int                    `json:"total_steps"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    *time.Duration         `json:"duration,omitempty"`
	Results     []StepResult           `json:"results"`
	Error       *string                `json:"error,omitempty"`
	Context     map[string]interface{} `json:"context"`
}

// StepResult 表示步骤执行结果
type StepResult struct {
	StepNumber int            `json:"step_number"`
	Action     ActionType     `json:"action"`
	Status     string         `json:"status"`
	StartTime  time.Time      `json:"start_time"`
	EndTime    *time.Time     `json:"end_time,omitempty"`
	Duration   *time.Duration `json:"duration,omitempty"`
	Error      *string        `json:"error,omitempty"`
	Result     interface{}    `json:"result,omitempty"`
}

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
	logger            *zap.Logger
	instances         map[string]*WorkflowInstance
	actionHandlers    map[ActionType]ActionHandler
	conditionHandlers map[ConditionType]ConditionHandler
}

// ActionHandler 动作处理器接口
type ActionHandler func(ctx context.Context, params map[string]interface{}, context map[string]interface{}) (interface{}, error)

// ConditionHandler 条件处理器接口
type ConditionHandler func(ctx context.Context, params map[string]interface{}, context map[string]interface{}) (bool, error)

// NewWorkflowEngine 创建新的工作流引擎
func NewWorkflowEngine(logger *zap.Logger) *WorkflowEngine {
	return &WorkflowEngine{
		logger:            logger,
		instances:         make(map[string]*WorkflowInstance),
		actionHandlers:    make(map[ActionType]ActionHandler),
		conditionHandlers: make(map[ConditionType]ConditionHandler),
	}
}

// RegisterActionHandler 注册动作处理器
func (e *WorkflowEngine) RegisterActionHandler(actionType ActionType, handler ActionHandler) {
	e.actionHandlers[actionType] = handler
	e.logger.Info("注册动作处理器", zap.String("type", string(actionType)))
}

// RegisterConditionHandler 注册条件处理器
func (e *WorkflowEngine) RegisterConditionHandler(conditionType ConditionType, handler ConditionHandler) {
	e.conditionHandlers[conditionType] = handler
	e.logger.Info("注册条件处理器", zap.String("type", string(conditionType)))
}

// ExecuteWorkflow 执行工作流
func (e *WorkflowEngine) ExecuteWorkflow(ctx context.Context, workflow *models.Workflow, triggerData map[string]interface{}) (*WorkflowInstance, error) {
	if workflow.Status == string(WorkflowStatusInactive) {
		return nil, fmt.Errorf("workflow is inactive: %s", workflow.Name)
	}

	instance := &WorkflowInstance{
		ID:          uuid.New().String(),
		WorkflowID:  workflow.ID.String(),
		Name:        workflow.Name,
		Status:      WorkflowStatusActive,
		CurrentStep: 0,
		TotalSteps:  len(workflow.Steps),
		StartTime:   time.Now(),
		Context:     make(map[string]interface{}),
	}

	// 合并触发数据到上下文
	for k, v := range triggerData {
		instance.Context[k] = v
	}

	e.instances[instance.ID] = instance
	e.logger.Info("开始执行工作流",
		zap.String("instance_id", instance.ID),
		zap.String("workflow_id", workflow.ID.String()),
		zap.String("workflow_name", workflow.Name))

	// 异步执行工作流
	go e.executeWorkflowSteps(instance, workflow)

	return instance, nil
}

// executeWorkflowSteps 执行工作流步骤
func (e *WorkflowEngine) executeWorkflowSteps(instance *WorkflowInstance, workflow *models.Workflow) {
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("workflow execution panicked: %v", r)
			e.updateInstanceStatus(instance, WorkflowStatusError, &errorMsg)
			e.logger.Error("工作流执行异常",
				zap.String("instance_id", instance.ID),
				zap.String("workflow_name", workflow.Name),
				zap.Any("panic", r))
		}
	}()

	for i, step := range workflow.Steps {
		instance.CurrentStep = i + 1

		// 检查条件
		if step.Conditions != nil && len(step.Conditions) > 0 {
			if !e.evaluateConditions(instance.Context, step.Conditions) {
				e.logger.Info("条件不满足，跳过步骤",
					zap.String("instance_id", instance.ID),
					zap.Int("step", i+1))
				continue
			}
		}

		// 执行动作
		result := e.executeStep(instance.Context, step.Action, step.Params)
		instance.Results = append(instance.Results, result)

		if result.Error != nil {
			e.updateInstanceStatus(instance, WorkflowStatusError, result.Error)
			return
		}

		// 更新上下文
		if result.Result != nil {
			instance.Context[fmt.Sprintf("step_%d_result", i+1)] = result.Result
		}
	}

	e.updateInstanceStatus(instance, WorkflowStatusActive, nil)
	e.logger.Info("工作流执行完成",
		zap.String("instance_id", instance.ID),
		zap.String("workflow_name", workflow.Name))
}

// evaluateConditions 评估条件
func (e *WorkflowEngine) evaluateConditions(context map[string]interface{}, conditions []map[string]interface{}) bool {
	for _, condition := range conditions {
		conditionType, ok := condition["type"].(string)
		if !ok {
			e.logger.Warn("条件类型无效", zap.Any("condition", condition))
			return false
		}

		handler, exists := e.conditionHandlers[ConditionType(conditionType)]
		if !exists {
			e.logger.Warn("未找到条件处理器", zap.String("type", conditionType))
			return false
		}

		params, ok := condition["params"].(map[string]interface{})
		if !ok {
			params = make(map[string]interface{})
		}

		met, err := handler(context, params, context)
		if err != nil {
			e.logger.Error("条件评估失败",
				zap.String("type", conditionType),
				zap.Error(err))
			return false
		}

		if !met {
			return false
		}
	}

	return true
}

// executeStep 执行步骤
func (e *WorkflowEngine) executeStep(context map[string]interface{}, actionType ActionType, params map[string]interface{}) StepResult {
	result := StepResult{
		StepNumber: 0, // 将在调用时设置
		Action:     actionType,
		Status:     "running",
		StartTime:  time.Now(),
	}

	handler, exists := e.actionHandlers[actionType]
	if !exists {
		errorMsg := fmt.Sprintf("action handler not found: %s", actionType)
		result.Status = "failed"
		result.Error = &errorMsg
		result.EndTime = &time.Now()
		duration := result.EndTime.Sub(result.StartTime)
		result.Duration = &duration
		return result
	}

	resultData, err := handler(context, params, context)
	if err != nil {
		errorMsg := err.Error()
		result.Status = "failed"
		result.Error = &errorMsg
	} else {
		result.Status = "completed"
		result.Result = resultData
	}

	result.EndTime = &time.Now()
	duration := result.EndTime.Sub(result.StartTime)
	result.Duration = &duration

	return result
}

// updateInstanceStatus 更新实例状态
func (e *WorkflowEngine) updateInstanceStatus(instance *WorkflowInstance, status WorkflowStatus, errorMsg *string) {
	endTime := time.Now()
	duration := endTime.Sub(instance.StartTime)

	instance.Status = status
	instance.EndTime = &endTime
	instance.Duration = &duration
	instance.Error = errorMsg

	e.logger.Info("工作流实例状态更新",
		zap.String("instance_id", instance.ID),
		zap.String("status", string(status)),
		zap.Duration("duration", duration))
}

// GetWorkflowInstance 获取工作流实例
func (e *WorkflowEngine) GetWorkflowInstance(instanceID string) (*WorkflowInstance, error) {
	instance, exists := e.instances[instanceID]
	if !exists {
		return nil, fmt.Errorf("workflow instance not found: %s", instanceID)
	}
	return instance, nil
}

// ListWorkflowInstances 列出工作流实例
func (e *WorkflowEngine) ListWorkflowInstances() []*WorkflowInstance {
	instances := make([]*WorkflowInstance, 0, len(e.instances))
	for _, instance := range e.instances {
		instances = append(instances, instance)
	}
	return instances
}

// InitializeDefaultHandlers 初始化默认处理器
func (e *WorkflowEngine) InitializeDefaultHandlers() {
	// 注册默认动作处理器
	e.RegisterActionHandler(ActionTypeAddSubscribe, e.addSubscribeAction)
	e.RegisterActionHandler(ActionTypeStartDownload, e.startDownloadAction)
	e.RegisterActionHandler(ActionTypeSendMessage, e.sendMessageAction)
	e.RegisterActionHandler(ActionTypeCallPlugin, e.callPluginAction)

	// 注册默认条件处理器
	e.RegisterConditionHandler(ConditionTypeMediaMatch, e.mediaMatchCondition)
	e.RegisterConditionHandler(ConditionTypeTimeRange, e.timeRangeCondition)

	e.logger.Info("默认工作流处理器初始化完成")
}

// 默认动作处理器实现
func (e *WorkflowEngine) addSubscribeAction(ctx context.Context, params map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	e.logger.Info("执行添加订阅动作", zap.Any("params", params))
	// 实际实现需要调用订阅服务
	return map[string]interface{}{"result": "subscribe_added"}, nil
}

func (e *WorkflowEngine) startDownloadAction(ctx context.Context, params map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	e.logger.Info("执行开始下载动作", zap.Any("params", params))
	// 实际实现需要调用下载服务
	return map[string]interface{}{"result": "download_started"}, nil
}

func (e *WorkflowEngine) sendMessageAction(ctx context.Context, params map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	e.logger.Info("执行发送消息动作", zap.Any("params", params))
	// 实际实现需要调用消息服务
	return map[string]interface{}{"result": "message_sent"}, nil
}

func (e *WorkflowEngine) callPluginAction(ctx context.Context, params map[string]interface{}, context map[string]interface{}) (interface{}, error) {
	e.logger.Info("执行调用插件动作", zap.Any("params", params))
	// 实际实现需要调用插件系统
	return map[string]interface{}{"result": "plugin_called"}, nil
}

// 默认条件处理器实现
func (e *WorkflowEngine) mediaMatchCondition(ctx context.Context, params map[string]interface{}, context map[string]interface{}) (bool, error) {
	e.logger.Debug("评估媒体匹配条件", zap.Any("params", params))
	// 实际实现需要检查媒体匹配条件
	return true, nil
}

func (e *WorkflowEngine) timeRangeCondition(ctx context.Context, params map[string]interface{}, context map[string]interface{}) (bool, error) {
	e.logger.Debug("评估时间范围条件", zap.Any("params", params))
	// 实际实现需要检查时间范围
	return true, nil
}
