package workflows

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Engine 定义工作流引擎接口
type Engine interface {
	// RegisterWorkflow 注册工作流
	RegisterWorkflow(workflow Workflow) error

	// GetWorkflow 获取工作流
	GetWorkflow(workflowID string) (Workflow, error)

	// ExecuteWorkflow 执行工作流
	ExecuteWorkflow(ctx WorkflowContext) (*WorkflowResult, error)

	// ExecuteWorkflowAsync 异步执行工作流
	ExecuteWorkflowAsync(ctx WorkflowContext) (string, error)

	// PauseWorkflow 暂停工作流
	PauseWorkflow(workflowID string) error

	// ResumeWorkflow 恢复工作流
	ResumeWorkflow(workflowID string) error

	// CancelWorkflow 取消工作流
	CancelWorkflow(workflowID string) error

	// RollbackWorkflow 回滚工作流
	RollbackWorkflow(workflowID string) error

	// GetWorkflowStatus 获取工作流状态
	GetWorkflowStatus(workflowID string) (string, error)

	// GetWorkflowResults 获取工作流执行结果
	GetWorkflowResults(workflowID string) (*WorkflowResult, error)

	// ListWorkflows 列出所有工作流
	ListWorkflows(params ListWorkflowsParams) ([]Workflow, error)
}

// ListWorkflowsParams 定义列出工作流的参数
type ListWorkflowsParams struct {
	Status   string // 工作流状态过滤
	Priority string // 工作流优先级过滤
	Type     string // 工作流类型过滤
	Limit    int    // 返回结果数量限制
	Offset   int    // 偏移量
}

// DefaultWorkflowEngine 实现默认的工作流引擎
type DefaultWorkflowEngine struct {
	// workflows 已注册的工作流列表
	workflows map[string]Workflow

	// results 工作流执行结果
	results map[string]*WorkflowResult

	// mutex 互斥锁，用于保护workflows和results的并发访问
	mutex sync.RWMutex

	// logger 日志记录器
	logger *zap.Logger

	// runningWorkflows 正在运行的工作流
	runningWorkflows map[string]context.CancelFunc

	// runningMutex 互斥锁，用于保护runningWorkflows的并发访问
	runningMutex sync.Mutex
}

// NewDefaultWorkflowEngine 创建新的工作流引擎实例
func NewDefaultWorkflowEngine(logger *zap.Logger) *DefaultWorkflowEngine {
	return &DefaultWorkflowEngine{
		workflows:        make(map[string]Workflow),
		results:          make(map[string]*WorkflowResult),
		logger:           logger,
		runningWorkflows: make(map[string]context.CancelFunc),
	}
}

// RegisterWorkflow 注册工作流
func (e *DefaultWorkflowEngine) RegisterWorkflow(workflow Workflow) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 检查工作流是否已注册
	workflowID := workflow.GetID()
	if _, exists := e.workflows[workflowID]; exists {
		return nil // 工作流已注册，无需重复注册
	}

	// 注册工作流
	e.workflows[workflowID] = workflow
	e.logger.Info("Workflow registered", zap.String("workflow_id", workflowID), zap.String("workflow_name", workflow.GetName()), zap.String("workflow_type", workflow.GetType()))

	return nil
}

// GetWorkflow 获取工作流
func (e *DefaultWorkflowEngine) GetWorkflow(workflowID string) (Workflow, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	workflow, exists := e.workflows[workflowID]
	if !exists {
		return nil, nil // 工作流不存在
	}

	return workflow, nil
}

// ExecuteWorkflow 执行工作流
func (e *DefaultWorkflowEngine) ExecuteWorkflow(ctx WorkflowContext) (*WorkflowResult, error) {
	// 记录工作流执行开始
	e.logger.Info("Workflow execution started", 
		zap.String("workflow_id", ctx.WorkflowID), 
		zap.String("workflow_name", ctx.WorkflowName),
		zap.String("workflow_type", ctx.Type),
		zap.String("priority", ctx.Priority))

	// 获取工作流
	workflow, err := e.GetWorkflow(ctx.WorkflowID)
	if err != nil {
		e.logger.Error("Failed to get workflow", 
			zap.String("workflow_id", ctx.WorkflowID), 
			zap.String("error", err.Error()))
		return nil, err
	}

	if workflow == nil {
		e.logger.Error("Workflow not found", zap.String("workflow_id", ctx.WorkflowID))
		return nil, nil // 工作流不存在
	}

	// 初始化工作流
	if initErr := workflow.Initialize(ctx); initErr != nil {
		e.logger.Error("Failed to initialize workflow", 
			zap.String("workflow_id", ctx.WorkflowID), 
			zap.String("workflow_name", workflow.GetName()),
			zap.String("error", initErr.Error()))
		return nil, initErr
	}

	// 执行工作流
	result, err := workflow.Execute(ctx)

	// 记录执行结果
	if err != nil || !result.Success {
		e.logger.Error("Workflow execution failed", 
			zap.String("workflow_id", ctx.WorkflowID), 
			zap.String("workflow_name", workflow.GetName()),
			zap.String("error", err.Error()),
			zap.String("result_status", result.Status),
			zap.String("error_message", result.ErrorMessage))
	} else {
		e.logger.Info("Workflow execution completed successfully", 
			zap.String("workflow_id", ctx.WorkflowID), 
			zap.String("workflow_name", workflow.GetName()),
			zap.Duration("duration", result.Duration),
			zap.Int("steps_count", len(result.StepResults)))
	}

	// 保存执行结果
	e.mutex.Lock()
	e.results[ctx.WorkflowID] = result
	e.mutex.Unlock()

	return result, err
}

// ExecuteWorkflowAsync 异步执行工作流
func (e *DefaultWorkflowEngine) ExecuteWorkflowAsync(ctx WorkflowContext) (string, error) {
	// 创建带有取消功能的上下文
	cancelCtx, cancel := context.WithCancel(ctx)

	// 将取消函数保存到runningWorkflows中
	e.runningMutex.Lock()
	e.runningWorkflows[ctx.WorkflowID] = cancel
	e.runningMutex.Unlock()

	// 异步执行工作流
	go func() {
		defer func() {
			// 执行完成后，从runningWorkflows中移除
			e.runningMutex.Lock()
			delete(e.runningWorkflows, ctx.WorkflowID)
			e.runningMutex.Unlock()
		}()

		// 执行工作流
		_, err := e.ExecuteWorkflow(WorkflowContext{
			Context:       cancelCtx,
			Logger:        ctx.Logger,
			WorkflowID:    ctx.WorkflowID,
			WorkflowName:  ctx.WorkflowName,
			Input:         ctx.Input,
			GlobalContext: ctx.GlobalContext,
			Services:      ctx.Services,
			Priority:      ctx.Priority,
			Type:          ctx.Type,
			Timeout:       ctx.Timeout,
			RetryCount:    ctx.RetryCount,
			RetryInterval: ctx.RetryInterval,
		})

		if err != nil {
			e.logger.Error("Failed to execute workflow asynchronously", zap.String("workflow_id", ctx.WorkflowID), zap.String("error", err.Error()))
		}
	}()

	return ctx.WorkflowID, nil
}

// PauseWorkflow 暂停工作流
func (e *DefaultWorkflowEngine) PauseWorkflow(workflowID string) error {
	// 获取工作流
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		e.logger.Error("Failed to get workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return err
	}

	if workflow == nil {
		return nil // 工作流不存在
	}

	// 暂停工作流
	if err := workflow.Pause(); err != nil {
		e.logger.Error("Failed to pause workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return err
	}

	e.logger.Info("Workflow paused", zap.String("workflow_id", workflowID))
	return nil
}

// ResumeWorkflow 恢复工作流
func (e *DefaultWorkflowEngine) ResumeWorkflow(workflowID string) error {
	// 获取工作流
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		e.logger.Error("Failed to get workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return err
	}

	if workflow == nil {
		return nil // 工作流不存在
	}

	// 恢复工作流
	if err := workflow.Resume(); err != nil {
		e.logger.Error("Failed to resume workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return err
	}

	e.logger.Info("Workflow resumed", zap.String("workflow_id", workflowID))
	return nil
}

// CancelWorkflow 取消工作流
func (e *DefaultWorkflowEngine) CancelWorkflow(workflowID string) error {
	// 从runningWorkflows中获取取消函数
	e.runningMutex.Lock()
	cancel, exists := e.runningWorkflows[workflowID]
	if exists {
		// 调用取消函数
		cancel()
		// 从runningWorkflows中移除
		delete(e.runningWorkflows, workflowID)
	}
	e.runningMutex.Unlock()

	// 获取工作流
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		e.logger.Error("Failed to get workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return err
	}

	if workflow == nil {
		return nil // 工作流不存在
	}

	// 取消工作流
	if err := workflow.Cancel(); err != nil {
		e.logger.Error("Failed to cancel workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return err
	}

	e.logger.Info("Workflow cancelled", zap.String("workflow_id", workflowID))
	return nil
}

// RollbackWorkflow 回滚工作流
func (e *DefaultWorkflowEngine) RollbackWorkflow(workflowID string) error {
	// 获取工作流
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		e.logger.Error("Failed to get workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return err
	}

	if workflow == nil {
		return nil // 工作流不存在
	}

	// 回滚工作流
	if err := workflow.Rollback(); err != nil {
		e.logger.Error("Failed to rollback workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return err
	}

	e.logger.Info("Workflow rolled back", zap.String("workflow_id", workflowID))
	return nil
}

// GetWorkflowStatus 获取工作流状态
func (e *DefaultWorkflowEngine) GetWorkflowStatus(workflowID string) (string, error) {
	// 获取工作流
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		e.logger.Error("Failed to get workflow", zap.String("workflow_id", workflowID), zap.String("error", err.Error()))
		return "", err
	}

	if workflow == nil {
		return "", nil // 工作流不存在
	}

	return workflow.GetStatus(), nil
}

// GetWorkflowResults 获取工作流执行结果
func (e *DefaultWorkflowEngine) GetWorkflowResults(workflowID string) (*WorkflowResult, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	result, exists := e.results[workflowID]
	if !exists {
		return nil, nil // 结果不存在
	}

	return result, nil
}

// ListWorkflows 列出所有工作流
func (e *DefaultWorkflowEngine) ListWorkflows(params ListWorkflowsParams) ([]Workflow, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	var workflows []Workflow

	// 遍历所有工作流
	for _, workflow := range e.workflows {
		// 应用过滤条件
		if (params.Status == "" || workflow.GetStatus() == params.Status) &&
			(params.Priority == "" || workflow.GetPriority() == params.Priority) &&
			(params.Type == "" || workflow.GetType() == params.Type) {
			workflows = append(workflows, workflow)
		}
	}

	// 应用分页
	if params.Limit > 0 {
		end := params.Offset + params.Limit
		if end > len(workflows) {
			end = len(workflows)
		}

		if params.Offset < len(workflows) {
			workflows = workflows[params.Offset:end]
		} else {
			workflows = []Workflow{}
		}
	}

	return workflows, nil
}
