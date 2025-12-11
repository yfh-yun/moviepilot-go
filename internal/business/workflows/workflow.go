package workflows

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// WorkflowStatus 定义工作流状态
const (
	WorkflowStatusPending   = "pending"   // 待执行
	WorkflowStatusRunning   = "running"   // 执行中
	WorkflowStatusPaused    = "paused"    // 已暂停
	WorkflowStatusCompleted = "completed" // 执行完成
	WorkflowStatusFailed    = "failed"    // 执行失败
	WorkflowStatusCancelled = "cancelled" // 已取消
)

// WorkflowPriority 定义工作流优先级
const (
	WorkflowPriorityLow    = "low"    // 低优先级
	WorkflowPriorityMedium = "medium" // 中优先级
	WorkflowPriorityHigh   = "high"   // 高优先级
)

// WorkflowType 定义工作流类型
const (
	WorkflowTypeSequential = "sequential" // 顺序执行
	WorkflowTypeParallel   = "parallel"   // 并行执行
	WorkflowTypeHybrid     = "hybrid"     // 混合执行
)

// Workflow 定义工作流接口
type Workflow interface {
	// GetID 获取工作流ID
	GetID() string

	// GetName 获取工作流名称
	GetName() string

	// GetDescription 获取工作流描述
	GetDescription() string

	// GetStatus 获取工作流状态
	GetStatus() string

	// GetPriority 获取工作流优先级
	GetPriority() string

	// GetType 获取工作流类型
	GetType() string

	// Initialize 初始化工作流
	Initialize(ctx WorkflowContext) error

	// Execute 执行工作流
	Execute(ctx WorkflowContext) (*WorkflowResult, error)

	// Pause 暂停工作流
	Pause() error

	// Resume 恢复工作流
	Resume() error

	// Cancel 取消工作流
	Cancel() error

	// Rollback 回滚工作流
	Rollback() error

	// GetSteps 获取工作流步骤列表
	GetSteps() []Step

	// AddStep 添加工作流步骤
	AddStep(step Step) error

	// RemoveStep 移除工作流步骤
	RemoveStep(stepID string) error

	// UpdateStep 更新工作流步骤
	UpdateStep(step Step) error
}

// Step 定义工作流步骤接口
type Step interface {
	// GetID 获取步骤ID
	GetID() string

	// GetName 获取步骤名称
	GetName() string

	// GetDescription 获取步骤描述
	GetDescription() string

	// GetStatus 获取步骤状态
	GetStatus() string

	// GetActionName 获取步骤对应的动作名称
	GetActionName() string

	// GetActionType 获取步骤对应的动作类型
	GetActionType() string

	// GetInput 获取步骤输入参数
	GetInput() map[string]any

	// GetOutput 获取步骤输出参数
	GetOutput() map[string]any

	// GetDependencies 获取步骤依赖的步骤ID列表
	GetDependencies() []string

	// Execute 执行步骤
	Execute(ctx StepContext) (*StepResult, error)

	// Rollback 回滚步骤
	Rollback() error
}

// WorkflowContext 定义工作流执行上下文
type WorkflowContext struct {
	context.Context
	Logger        *zap.Logger    // 日志记录器
	WorkflowID    string         // 工作流ID
	WorkflowName  string         // 工作流名称
	Input         map[string]any // 输入参数
	GlobalContext map[string]any // 全局上下文
	Services      map[string]any // 服务实例
	Priority      string         // 工作流优先级
	Type          string         // 工作流类型
	Timeout       time.Duration  // 工作流超时时间
	RetryCount    int            // 重试次数
	RetryInterval time.Duration  // 重试间隔
}

// StepContext 定义步骤执行上下文
type StepContext struct {
	WorkflowContext
	StepID       string         // 步骤ID
	StepName     string         // 步骤名称
	ActionName   string         // 动作名称
	ActionType   string         // 动作类型
	Input        map[string]any // 步骤输入参数
	Dependencies []string       // 依赖的步骤ID
}

// WorkflowResult 定义工作流执行结果
type WorkflowResult struct {
	Success      bool           `json:"success"`                 // 执行是否成功
	ErrorMessage string         `json:"error_message,omitempty"` // 错误信息
	Output       map[string]any `json:"output"`                  // 输出数据
	Duration     time.Duration  `json:"duration"`                // 执行时长
	Status       string         `json:"status"`                  // 执行状态
	StepResults  []StepResult   `json:"step_results"`            // 步骤执行结果
}

// StepResult 定义步骤执行结果
type StepResult struct {
	StepID       string         `json:"step_id"`                 // 步骤ID
	StepName     string         `json:"step_name"`               // 步骤名称
	ActionName   string         `json:"action_name"`             // 动作名称
	ActionType   string         `json:"action_type"`             // 动作类型
	Success      bool           `json:"success"`                 // 执行是否成功
	ErrorMessage string         `json:"error_message,omitempty"` // 错误信息
	Output       map[string]any `json:"output"`                  // 输出数据
	Duration     time.Duration  `json:"duration"`                // 执行时长
	Status       string         `json:"status"`                  // 执行状态
	Dependencies []string       `json:"dependencies"`            // 依赖的步骤ID
}

// StepConfig 定义步骤配置
type StepConfig struct {
	ID           string         `json:"id"`           // 步骤ID
	Name         string         `json:"name"`         // 步骤名称
	Description  string         `json:"description"`  // 步骤描述
	ActionName   string         `json:"action_name"`  // 动作名称
	ActionType   string         `json:"action_type"`  // 动作类型
	Input        map[string]any `json:"input"`        // 输入参数
	Dependencies []string       `json:"dependencies"` // 依赖的步骤ID
}

// WorkflowConfig 定义工作流配置
type WorkflowConfig struct {
	ID            string        `json:"id"`             // 工作流ID
	Name          string        `json:"name"`           // 工作流名称
	Description   string        `json:"description"`    // 工作流描述
	Steps         []StepConfig  `json:"steps"`          // 步骤列表
	Priority      string        `json:"priority"`       // 工作流优先级
	Type          string        `json:"type"`           // 工作流类型
	Timeout       time.Duration `json:"timeout"`        // 超时时间
	RetryCount    int           `json:"retry_count"`    // 重试次数
	RetryInterval time.Duration `json:"retry_interval"` // 重试间隔
}

// BaseWorkflow 定义工作流基础实现
type BaseWorkflow struct {
	ID            string        // 工作流ID
	Name          string        // 工作流名称
	Description   string        // 工作流描述
	Steps         []Step        // 步骤列表
	Status        string        // 工作流状态
	Priority      string        // 工作流优先级
	Type          string        // 工作流类型
	Timeout       time.Duration // 超时时间
	RetryCount    int           // 重试次数
	RetryInterval time.Duration // 重试间隔
	Logger        *zap.Logger   // 日志记录器
	Initialized   bool          // 是否已初始化
	CreatedAt     time.Time     // 创建时间
	UpdatedAt     time.Time     // 更新时间
}

// Initialize 初始化工作流
func (w *BaseWorkflow) Initialize(ctx WorkflowContext) error {
	if w.Initialized {
		return nil
	}

	w.Logger = ctx.Logger
	w.Status = WorkflowStatusPending
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()
	w.Initialized = true

	// 如果未设置优先级，使用默认值
	if w.Priority == "" {
		w.Priority = WorkflowPriorityMedium
	}

	// 如果未设置类型，使用默认值
	if w.Type == "" {
		w.Type = WorkflowTypeSequential
	}

	// 如果未设置超时时间，使用默认值
	if w.Timeout == 0 {
		w.Timeout = time.Hour * 1
	}

	// 如果未设置重试次数，使用默认值
	if w.RetryCount < 0 {
		w.RetryCount = 0
	}

	// 如果未设置重试间隔，使用默认值
	if w.RetryInterval == 0 {
		w.RetryInterval = time.Second * 5
	}

	w.Logger.Info("Workflow initialized", zap.String("workflow_id", w.ID), zap.String("workflow_name", w.Name), zap.String("workflow_type", w.Type))
	return nil
}

// GetID 获取工作流ID
func (w *BaseWorkflow) GetID() string {
	return w.ID
}

// GetName 获取工作流名称
func (w *BaseWorkflow) GetName() string {
	return w.Name
}

// GetDescription 获取工作流描述
func (w *BaseWorkflow) GetDescription() string {
	return w.Description
}

// GetStatus 获取工作流状态
func (w *BaseWorkflow) GetStatus() string {
	return w.Status
}

// GetPriority 获取工作流优先级
func (w *BaseWorkflow) GetPriority() string {
	return w.Priority
}

// GetType 获取工作流类型
func (w *BaseWorkflow) GetType() string {
	return w.Type
}

// GetSteps 获取工作流步骤列表
func (w *BaseWorkflow) GetSteps() []Step {
	return w.Steps
}

// AddStep 添加工作流步骤
func (w *BaseWorkflow) AddStep(step Step) error {
	// 添加步骤到列表
	w.Steps = append(w.Steps, step)

	// 更新最后更新时间
	w.UpdatedAt = time.Now()

	w.Logger.Info("Step added to workflow", zap.String("workflow_id", w.ID), zap.String("step_id", step.GetID()), zap.String("step_name", step.GetName()))
	return nil
}

// RemoveStep 移除工作流步骤
func (w *BaseWorkflow) RemoveStep(stepID string) error {
	// 查找并移除步骤
	for i, step := range w.Steps {
		if step.GetID() == stepID {
			w.Steps = append(w.Steps[:i], w.Steps[i+1:]...)
			// 更新最后更新时间
			w.UpdatedAt = time.Now()
			w.Logger.Info("Step removed from workflow", zap.String("workflow_id", w.ID), zap.String("step_id", stepID))
			return nil
		}
	}

	return nil
}

// UpdateStep 更新工作流步骤
func (w *BaseWorkflow) UpdateStep(step Step) error {
	// 查找并更新步骤
	for i, s := range w.Steps {
		if s.GetID() == step.GetID() {
			w.Steps[i] = step
			// 更新最后更新时间
			w.UpdatedAt = time.Now()
			w.Logger.Info("Step updated in workflow", zap.String("workflow_id", w.ID), zap.String("step_id", step.GetID()), zap.String("step_name", step.GetName()))
			return nil
		}
	}

	// 如果步骤不存在，添加到列表
	return w.AddStep(step)
}

// Pause 暂停工作流
func (w *BaseWorkflow) Pause() error {
	// 更新工作流状态为暂停
	w.Status = WorkflowStatusPaused
	// 更新最后更新时间
	w.UpdatedAt = time.Now()

	w.Logger.Info("Workflow paused", zap.String("workflow_id", w.ID), zap.String("workflow_name", w.Name))
	return nil
}

// Resume 恢复工作流
func (w *BaseWorkflow) Resume() error {
	// 更新工作流状态为运行中
	w.Status = WorkflowStatusRunning
	// 更新最后更新时间
	w.UpdatedAt = time.Now()

	w.Logger.Info("Workflow resumed", zap.String("workflow_id", w.ID), zap.String("workflow_name", w.Name))
	return nil
}

// Cancel 取消工作流
func (w *BaseWorkflow) Cancel() error {
	// 更新工作流状态为取消
	w.Status = WorkflowStatusCancelled
	// 更新最后更新时间
	w.UpdatedAt = time.Now()

	w.Logger.Info("Workflow cancelled", zap.String("workflow_id", w.ID), zap.String("workflow_name", w.Name))
	return nil
}

// Rollback 回滚工作流
func (w *BaseWorkflow) Rollback() error {
	// TODO: 实现工作流回滚逻辑
	// 这里需要回滚所有已执行的步骤

	w.Logger.Info("Workflow rollback initiated", zap.String("workflow_id", w.ID), zap.String("workflow_name", w.Name))

	// 记录回滚过程
	for i := len(w.Steps) - 1; i >= 0; i-- {
		step := w.Steps[i]
		// 回滚单个步骤
		if err := step.Rollback(); err != nil {
			w.Logger.Error("Step rollback failed", zap.String("workflow_id", w.ID), zap.String("step_id", step.GetID()), zap.String("error", err.Error()))
			continue
		}
		w.Logger.Info("Step rolled back successfully", zap.String("workflow_id", w.ID), zap.String("step_id", step.GetID()))
	}

	// 更新工作流状态
	w.Status = WorkflowStatusFailed
	// 更新最后更新时间
	w.UpdatedAt = time.Now()

	w.Logger.Info("Workflow rollback completed", zap.String("workflow_id", w.ID), zap.String("workflow_name", w.Name))
	return nil
}

// Execute 执行工作流
func (w *BaseWorkflow) Execute(ctx WorkflowContext) (*WorkflowResult, error) {
	startTime := time.Now()

	// 检查工作流是否已初始化
	if !w.Initialized {
		if err := w.Initialize(ctx); err != nil {
			w.Logger.Error("Failed to initialize workflow during execution", zap.String("workflow_id", w.ID), zap.String("error", err.Error()))
			return &WorkflowResult{
				Success:      false,
				ErrorMessage: "Failed to initialize workflow: " + err.Error(),
				Output:       make(map[string]any),
				Duration:     time.Since(startTime),
				Status:       WorkflowStatusFailed,
				StepResults:  []StepResult{},
			}, err
		}
	}

	// 更新工作流状态为执行中
	w.Status = WorkflowStatusRunning
	w.UpdatedAt = time.Now()

	w.Logger.Info("Workflow execution started", zap.String("workflow_id", w.ID), zap.String("workflow_name", w.Name), zap.Int("steps_count", len(w.Steps)))

	// 创建步骤上下文和执行步骤
	stepResults := make([]StepResult, 0, len(w.Steps))
	globalOutput := make(map[string]any)

	// 根据工作流类型执行步骤
	switch w.Type {
	case WorkflowTypeSequential:
		// 顺序执行所有步骤
		for i, step := range w.Steps {
			stepCtx := StepContext{
				WorkflowContext: ctx,
				StepID:          step.GetID(),
				StepName:        step.GetName(),
				ActionName:      step.GetActionName(),
				ActionType:      step.GetActionType(),
				Input:           step.GetInput(),
				Dependencies:    step.GetDependencies(),
			}

			// 执行步骤
			result, err := step.Execute(stepCtx)
			if err != nil || !result.Success {
				w.Logger.Error("Step execution failed", zap.String("workflow_id", w.ID), zap.String("step_id", step.GetID()), zap.String("error", err.Error()))

				// 如果步骤执行失败，回滚工作流
				if rollbackErr := w.Rollback(); rollbackErr != nil {
					w.Logger.Error("Workflow rollback failed", zap.String("workflow_id", w.ID), zap.String("error", rollbackErr.Error()))
				}

				// 更新工作流状态为失败
				w.Status = WorkflowStatusFailed
				w.UpdatedAt = time.Now()

				// 返回失败结果
				return &WorkflowResult{
					Success:      false,
					ErrorMessage: result.ErrorMessage,
					Output:       globalOutput,
					Duration:     time.Since(startTime),
					Status:       WorkflowStatusFailed,
					StepResults:  append(stepResults, *result),
				}, err
			}

			// 保存步骤结果
			stepResults = append(stepResults, *result)

			// 将步骤输出合并到全局输出
			for k, v := range result.Output {
				globalOutput[k] = v
			}

			w.Logger.Info("Step executed successfully", zap.String("workflow_id", w.ID), zap.String("step_id", step.GetID()), zap.String("step_name", step.GetName()), zap.Int("step_index", i+1), zap.Int("total_steps", len(w.Steps)))
		}

	case WorkflowTypeParallel:
		// 并行执行所有步骤
		stepResultsChan := make(chan StepResult, len(w.Steps))

		for _, step := range w.Steps {
			go func(step Step) {
				stepCtx := StepContext{
					WorkflowContext: ctx,
					StepID:          step.GetID(),
					StepName:        step.GetName(),
					ActionName:      step.GetActionName(),
					ActionType:      step.GetActionType(),
					Input:           step.GetInput(),
					Dependencies:    step.GetDependencies(),
				}

				result, _ := step.Execute(stepCtx)
				stepResultsChan <- *result
			}(step)
		}

		// 收集所有步骤结果
		allSuccess := true
		for i := 0; i < len(w.Steps); i++ {
			result := <-stepResultsChan
			stepResults = append(stepResults, result)

			if !result.Success {
				allSuccess = false
			}

			// 将步骤输出合并到全局输出
			for k, v := range result.Output {
				globalOutput[k] = v
			}
		}

		// 如果有任何步骤失败，回滚工作流
		if !allSuccess {
			w.Logger.Error("Some steps failed", zap.String("workflow_id", w.ID))

			if err := w.Rollback(); err != nil {
				w.Logger.Error("Workflow rollback failed", zap.String("workflow_id", w.ID), zap.String("error", err.Error()))
			}

			w.Status = WorkflowStatusFailed
			w.UpdatedAt = time.Now()

			return &WorkflowResult{
				Success:      false,
				ErrorMessage: "Some steps failed",
				Output:       globalOutput,
				Duration:     time.Since(startTime),
				Status:       WorkflowStatusFailed,
				StepResults:  stepResults,
			}, nil
		}

	default:
		// 默认使用顺序执行
		w.Logger.Warn("Unknown workflow type, using sequential execution", zap.String("workflow_id", w.ID), zap.String("workflow_type", w.Type))

		for i, step := range w.Steps {
			stepCtx := StepContext{
				WorkflowContext: ctx,
				StepID:          step.GetID(),
				StepName:        step.GetName(),
				ActionName:      step.GetActionName(),
				ActionType:      step.GetActionType(),
				Input:           step.GetInput(),
				Dependencies:    step.GetDependencies(),
			}

			result, err := step.Execute(stepCtx)
			if err != nil || !result.Success {
				w.Logger.Error("Step execution failed", zap.String("workflow_id", w.ID), zap.String("step_id", step.GetID()), zap.String("error", err.Error()))

				if rollbackErr := w.Rollback(); rollbackErr != nil {
					w.Logger.Error("Workflow rollback failed", zap.String("workflow_id", w.ID), zap.String("error", rollbackErr.Error()))
				}

				w.Status = WorkflowStatusFailed
				w.UpdatedAt = time.Now()

				return &WorkflowResult{
					Success:      false,
					ErrorMessage: result.ErrorMessage,
					Output:       globalOutput,
					Duration:     time.Since(startTime),
					Status:       WorkflowStatusFailed,
					StepResults:  append(stepResults, *result),
				}, err
			}

			stepResults = append(stepResults, *result)

			for k, v := range result.Output {
				globalOutput[k] = v
			}

			w.Logger.Info("Step executed successfully", zap.String("workflow_id", w.ID), zap.String("step_id", step.GetID()), zap.String("step_name", step.GetName()), zap.Int("step_index", i+1), zap.Int("total_steps", len(w.Steps)))
		}
	}

	// 更新工作流状态为执行完成
	w.Status = WorkflowStatusCompleted
	w.UpdatedAt = time.Now()

	// 创建工作流执行结果
	duration := time.Since(startTime)
	result := &WorkflowResult{
		Success:      true,
		ErrorMessage: "",
		Output:       globalOutput,
		Duration:     duration,
		Status:       WorkflowStatusCompleted,
		StepResults:  stepResults,
	}

	w.Logger.Info("Workflow execution completed successfully", zap.String("workflow_id", w.ID), zap.String("workflow_name", w.Name), zap.Duration("duration", duration), zap.Int("steps_count", len(stepResults)))

	return result, nil
}

// NewWorkflow 创建新的工作流实例
func NewWorkflow(config WorkflowConfig) (Workflow, error) {
	// TODO: 根据配置创建具体的工作流实例
	// 这里返回BaseWorkflow实例，实际应该返回具体的工作流实现

	return &BaseWorkflow{
		ID:            config.ID,
		Name:          config.Name,
		Description:   config.Description,
		Steps:         []Step{},
		Status:        WorkflowStatusPending,
		Priority:      config.Priority,
		Type:          config.Type,
		Timeout:       config.Timeout,
		RetryCount:    config.RetryCount,
		RetryInterval: config.RetryInterval,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}
