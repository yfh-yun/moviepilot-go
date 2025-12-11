package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Engine 工作流引擎
type Engine interface {
	// Execute 执行工作流
	Execute(ctx context.Context, workflow *Workflow) (*ExecutionResult, error)

	// ExecuteAsync 异步执行工作流
	ExecuteAsync(ctx context.Context, workflow *Workflow) (string, error)

	// GetExecution 获取执行状态
	GetExecution(ctx context.Context, executionID string) (*Execution, error)

	// CancelExecution 取消执行
	CancelExecution(ctx context.Context, executionID string) error

	// PauseExecution 暂停执行
	PauseExecution(ctx context.Context, executionID string) error

	// ResumeExecution 恢复执行
	ResumeExecution(ctx context.Context, executionID string) error

	// RollbackExecution 回滚执行
	RollbackExecution(ctx context.Context, executionID string) error

	// ExecuteParallel 并行执行工作流
	ExecuteParallel(ctx context.Context, workflows []*Workflow) ([]*ExecutionResult, error)
}

// engine 工作流引擎实现
type engine struct {
	executions map[string]*Execution
	workflows  map[string]*Workflow // workflowID -> Workflow
	mutex      sync.RWMutex
	logger     *zap.Logger
}

// NewEngine 创建工作流引擎
func NewEngine() Engine {
	return &engine{
		executions: make(map[string]*Execution),
		workflows:  make(map[string]*Workflow),
		logger:     logger.GetLogger(),
	}
}

// Workflow 工作流定义
type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []Step         `json:"steps"`
	Variables   map[string]any `json:"variables"`
}

// Step 工作流步骤
type Step struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Type           StepType       `json:"type"`
	Action         string         `json:"action"`
	Parameters     map[string]any `json:"parameters"`
	Condition      string         `json:"condition"`       // 执行条件
	OnSuccess      string         `json:"on_success"`      // 成功后跳转
	OnFailure      string         `json:"on_failure"`      // 失败后跳转
	Timeout        int            `json:"timeout"`         // 超时时间（秒）
	RetryCount     int            `json:"retry_count"`     // 重试次数
	RollbackAction string         `json:"rollback_action"` // 回滚动作
}

// StepType 步骤类型
type StepType string

const (
	StepTypeAction       StepType = "action"       // 动作
	StepTypeCondition    StepType = "condition"    // 条件判断
	StepTypeLoop         StepType = "loop"         // 循环
	StepTypeParallel     StepType = "parallel"     // 并行执行
	StepTypeDelay        StepType = "delay"        // 延迟
	StepTypeNotification StepType = "notification" // 通知
)

// Execution 执行实例
type Execution struct {
	ID              string                 `json:"id"`
	WorkflowID      string                 `json:"workflow_id"`
	Status          ExecutionStatus        `json:"status"`
	CurrentStep     string                 `json:"current_step"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         *time.Time             `json:"end_time"`
	Duration        int64                  `json:"duration"` // 毫秒
	StepResults     map[string]*StepResult `json:"step_results"`
	Variables       map[string]any         `json:"variables"`
	ErrorMsg        string                 `json:"error_msg"`
	PauseCh         chan struct{}          // 暂停信号通道
	ResumeCh        chan struct{}          // 恢复信号通道
	Paused          bool                   // 是否暂停
	RollbackHistory []string               // 回滚历史
}

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
)

// StepResult 步骤执行结果
type StepResult struct {
	StepID     string          `json:"step_id"`
	Status     ExecutionStatus `json:"status"`
	StartTime  time.Time       `json:"start_time"`
	EndTime    *time.Time      `json:"end_time"`
	Duration   int64           `json:"duration"`
	Output     map[string]any  `json:"output"`
	ErrorMsg   string          `json:"error_msg"`
	RetryCount int             `json:"retry_count"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	ExecutionID string          `json:"execution_id"`
	Status      ExecutionStatus `json:"status"`
	Duration    int64           `json:"duration"`
	ErrorMsg    string          `json:"error_msg"`
}

// Execute 执行工作流
func (e *engine) Execute(ctx context.Context, workflow *Workflow) (*ExecutionResult, error) {
	e.logger.Info("执行工作流",
		zap.String("workflow_id", workflow.ID),
		zap.String("workflow_name", workflow.Name),
	)

	executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano())
	startTime := time.Now()

	// 保存工作流定义
	e.mutex.Lock()
	e.workflows[workflow.ID] = workflow

	execution := &Execution{
		ID:              executionID,
		WorkflowID:      workflow.ID,
		Status:          StatusRunning,
		StartTime:       startTime,
		StepResults:     make(map[string]*StepResult),
		Variables:       workflow.Variables,
		PauseCh:         make(chan struct{}, 1),
		ResumeCh:        make(chan struct{}, 1),
		Paused:          false,
		RollbackHistory: make([]string, 0),
	}

	e.executions[executionID] = execution
	e.mutex.Unlock()

	// 执行步骤
	for _, step := range workflow.Steps {
		// 检查暂停信号
		select {
		case <-execution.PauseCh:
			// 进入暂停状态
			e.mutex.Lock()
			execution.Paused = true
			execution.Status = "paused"
			e.mutex.Unlock()

			e.logger.Info("工作流已暂停",
				zap.String("execution_id", executionID),
				zap.String("step_id", step.ID),
			)
		default:
			// 正常执行
		}

		if err := e.executeStep(ctx, execution, &step); err != nil {
			execution.Status = StatusFailed
			execution.ErrorMsg = err.Error()
			endTime := time.Now()
			execution.EndTime = &endTime
			execution.Duration = time.Since(startTime).Milliseconds()

			return &ExecutionResult{
				ExecutionID: executionID,
				Status:      StatusFailed,
				Duration:    execution.Duration,
				ErrorMsg:    err.Error(),
			}, err
		}
	}

	// 完成
	execution.Status = StatusCompleted
	endTime := time.Now()
	execution.EndTime = &endTime
	execution.Duration = time.Since(startTime).Milliseconds()

	e.logger.Info("工作流执行完成",
		zap.String("execution_id", executionID),
		zap.Int64("duration", execution.Duration),
	)

	return &ExecutionResult{
		ExecutionID: executionID,
		Status:      StatusCompleted,
		Duration:    execution.Duration,
	}, nil
}

// ExecuteParallel 并行执行工作流
func (e *engine) ExecuteParallel(ctx context.Context, workflows []*Workflow) ([]*ExecutionResult, error) {
	if len(workflows) == 0 {
		return nil, nil
	}

	results := make([]*ExecutionResult, len(workflows))
	errs := make([]error, len(workflows))
	var wg sync.WaitGroup

	e.logger.Info("开始并行执行工作流",
		zap.Int("workflow_count", len(workflows)),
	)

	// 并行执行每个工作流
	for i, workflow := range workflows {
		wg.Add(1)
		go func(idx int, wf *Workflow) {
			defer wg.Done()

			result, err := e.Execute(ctx, wf)
			results[idx] = result
			errs[idx] = err
		}(i, workflow)
	}

	// 等待所有工作流执行完成
	wg.Wait()

	// 收集错误
	var finalErr error
	errCount := 0
	for _, err := range errs {
		if err != nil {
			errCount++
			if finalErr == nil {
				finalErr = err
			}
		}
	}

	e.logger.Info("并行执行工作流完成",
		zap.Int("total_workflows", len(workflows)),
		zap.Int("failed_workflows", errCount),
		zap.Int("success_workflows", len(workflows)-errCount),
	)

	return results, finalErr
}

// ExecuteAsync 异步执行工作流
func (e *engine) ExecuteAsync(ctx context.Context, workflow *Workflow) (string, error) {
	executionID := fmt.Sprintf("exec_%d", time.Now().UnixNano())

	go func() {
		_, err := e.Execute(context.Background(), workflow)
		if err != nil {
			e.logger.Error("异步执行工作流失败",
				zap.String("execution_id", executionID),
				zap.Error(err),
			)
		}
	}()

	return executionID, nil
}

// GetExecution 获取执行状态
func (e *engine) GetExecution(ctx context.Context, executionID string) (*Execution, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	execution, ok := e.executions[executionID]
	if !ok {
		return nil, fmt.Errorf("执行不存在: %s", executionID)
	}

	return execution, nil
}

// PauseExecution 暂停执行
func (e *engine) PauseExecution(ctx context.Context, executionID string) error {
	e.mutex.Lock()
	execution, ok := e.executions[executionID]
	if !ok {
		e.mutex.Unlock()
		return fmt.Errorf("执行不存在: %s", executionID)
	}

	if execution.Status != StatusRunning {
		e.mutex.Unlock()
		return fmt.Errorf("执行状态错误，当前状态: %s", execution.Status)
	}

	// 发送暂停信号
	select {
	case execution.PauseCh <- struct{}{}:
		e.mutex.Unlock()
		e.logger.Info("工作流暂停请求已发送", zap.String("execution_id", executionID))
		return nil
	default:
		e.mutex.Unlock()
		return fmt.Errorf("无法发送暂停信号")
	}
}

// ResumeExecution 恢复执行
func (e *engine) ResumeExecution(ctx context.Context, executionID string) error {
	e.mutex.Lock()
	execution, ok := e.executions[executionID]
	if !ok {
		e.mutex.Unlock()
		return fmt.Errorf("执行不存在: %s", executionID)
	}

	if !execution.Paused {
		e.mutex.Unlock()
		return fmt.Errorf("执行未暂停")
	}

	// 发送恢复信号
	select {
	case execution.ResumeCh <- struct{}{}:
		e.mutex.Unlock()
		e.logger.Info("工作流恢复请求已发送", zap.String("execution_id", executionID))
		return nil
	default:
		e.mutex.Unlock()
		return fmt.Errorf("无法发送恢复信号")
	}
}

// RollbackExecution 回滚执行
func (e *engine) RollbackExecution(ctx context.Context, executionID string) error {
	e.mutex.RLock()
	execution, ok := e.executions[executionID]
	if !ok {
		e.mutex.RUnlock()
		return fmt.Errorf("执行不存在: %s", executionID)
	}

	workflow, ok := e.workflows[execution.WorkflowID]
	if !ok {
		e.mutex.RUnlock()
		return fmt.Errorf("工作流不存在: %s", execution.WorkflowID)
	}
	e.mutex.RUnlock()

	if execution.Status != StatusCompleted && execution.Status != StatusFailed {
		return fmt.Errorf("执行状态错误，当前状态: %s", execution.Status)
	}

	e.logger.Info("开始回滚工作流执行",
		zap.String("execution_id", executionID),
		zap.String("workflow_id", execution.WorkflowID),
	)

	// 按相反顺序回滚已执行的步骤
	executedSteps := make([]Step, 0)
	for _, step := range workflow.Steps {
		if _, exists := execution.StepResults[step.ID]; exists {
			executedSteps = append(executedSteps, step)
		}
	}

	// 反转步骤顺序，从最后执行的步骤开始回滚
	for i, j := 0, len(executedSteps)-1; i < j; i, j = i+1, j-1 {
		executedSteps[i], executedSteps[j] = executedSteps[j], executedSteps[i]
	}

	// 执行回滚
	for _, step := range executedSteps {
		// 这里简化处理，实际应该根据步骤类型执行不同的回滚逻辑
		e.logger.Info("回滚步骤",
			zap.String("execution_id", executionID),
			zap.String("step_id", step.ID),
			zap.String("step_name", step.Name),
		)

		// 记录回滚历史
		e.mutex.Lock()
		execution.RollbackHistory = append(execution.RollbackHistory, step.ID)
		e.mutex.Unlock()

		// 这里可以根据步骤类型执行具体的回滚逻辑
		// 例如：删除创建的文件、撤销数据库操作等
	}

	e.logger.Info("工作流回滚完成",
		zap.String("execution_id", executionID),
		zap.Int("rolled_back_steps", len(executedSteps)),
	)

	return nil
}

// CancelExecution 取消执行
func (e *engine) CancelExecution(ctx context.Context, executionID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	execution, ok := e.executions[executionID]
	if !ok {
		return fmt.Errorf("执行不存在: %s", executionID)
	}

	if execution.Status == StatusRunning {
		execution.Status = StatusCancelled
		endTime := time.Now()
		execution.EndTime = &endTime
		execution.Duration = time.Since(execution.StartTime).Milliseconds()
	}

	return nil
}

// executeStep 执行步骤，增强调试信息
func (e *engine) executeStep(_ context.Context, execution *Execution, step *Step) error {
	e.logger.Info("执行步骤",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID),
		zap.String("step_name", step.Name),
		zap.String("step_type", string(step.Type)),
		zap.String("action", step.Action),
	)

	startTime := time.Now()
	result := &StepResult{
		StepID:     step.ID,
		Status:     StatusRunning,
		StartTime:  startTime,
		Output:     make(map[string]any),
		RetryCount: 0,
	}

	execution.CurrentStep = step.ID
	execution.StepResults[step.ID] = result

	// TODO: 根据步骤类型执行不同的逻辑
	// 这里简化处理，但添加更详细的调试信息
	time.Sleep(100 * time.Millisecond)

	// 添加步骤执行的详细输出
	result.Output = map[string]any{
		"step_id":      step.ID,
		"step_name":    step.Name,
		"step_type":    string(step.Type),
		"action":       step.Action,
		"parameters":   step.Parameters,
		"execution_id": execution.ID,
		"workflow_id":  execution.WorkflowID,
	}

	// 模拟一些执行细节
	result.Output["execution_details"] = map[string]any{
		"timestamp": startTime,
		"status":    "completed",
		"message":   "Step executed successfully",
	}

	result.Status = StatusCompleted
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = time.Since(startTime).Milliseconds()

	return nil
}
