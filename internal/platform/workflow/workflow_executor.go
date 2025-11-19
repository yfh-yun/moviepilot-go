package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// DefaultWorkflowExecutor 默认工作流执行器
type DefaultWorkflowExecutor struct {
	manager   *WorkflowManagerImpl
	workflow  *Workflow
	logger    logger.Logger
	cancelCh  chan struct{}
	mutex     sync.RWMutex
	wg        sync.WaitGroup
	completed bool
	err       error
}

// NewDefaultWorkflowExecutor 创建默认工作流执行器
func NewDefaultWorkflowExecutor(manager *WorkflowManagerImpl, workflow *Workflow, logger logger.Logger) *DefaultWorkflowExecutor {
	return &DefaultWorkflowExecutor{
		manager:  manager,
		workflow: workflow,
		logger:   logger,
		cancelCh: make(chan struct{}),
	}
}

// Execute 执行工作流
func (e *DefaultWorkflowExecutor) Execute(workflow *Workflow, ctx context.Context, variables map[string]interface{}) error {
	startTime := time.Now()
	var result *TaskResult

	// 合并变量
	totalVars := MergeVariables(workflow.Variables, variables)

	// 创建工作流上下文
	workflowCtx := NewWorkflowContext(ctx, totalVars, e.logger, workflow)

	// 更新工作流状态
	e.updateWorkflowStatus(StatusRunning)

	// 检查工作流是否被取消
	select {
	case <-e.cancelCh:
		return ErrWorkflowCanceled
	default:
	}

	// 执行工作流任务
	if err := e.executeTasks(workflowCtx); err != nil {
		result = &TaskResult{
			Status:    TaskFailed,
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   time.Now(),
			Duration:  time.Since(startTime),
		}

		e.err = err
		e.updateWorkflowResult(result)
		return err
	}

	// 执行成功
	result = &TaskResult{
		Status:    TaskCompleted,
		Output:    workflowCtx.Variables,
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  time.Since(startTime),
	}

	e.updateWorkflowResult(result)
	e.completed = true

	return nil
}

// Cancel 取消执行
func (e *DefaultWorkflowExecutor) Cancel() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.completed {
		return nil
	}

	// 关闭取消通道
	close(e.cancelCh)

	// 等待所有任务完成
	e.wg.Wait()

	return nil
}

// executeTasks 执行所有任务
func (e *DefaultWorkflowExecutor) executeTasks(ctx *WorkflowContext) error {
	e.logger.Info("Starting workflow execution", "workflow_id", e.workflow.ID, "task_count", len(e.workflow.Tasks))

	// 初始化任务状态映射
	taskStatus := make(map[string]TaskStatus)
	for _, task := range e.workflow.Tasks {
		taskStatus[task.ID()] = task.Status()
	}

	// 创建任务通道，用于控制并发
	concurrentTasks := e.getConcurrentTaskLimit()
	taskChan := make(chan Task, concurrentTasks)
	errChan := make(chan error, len(e.workflow.Tasks))
	doneChan := make(chan bool, len(e.workflow.Tasks))

	// 启动工作协程来处理任务
	for i := 0; i < concurrentTasks; i++ {
		e.wg.Add(1)
		go e.taskWorker(taskChan, ctx, errChan, doneChan)
	}

	// 任务调度器
	go e.taskScheduler(taskChan, taskStatus)

	// 等待所有任务完成或出错
	completedCount := 0
	for {
		select {
		case err := <-errChan:
			if err != nil {
				// 取消所有任务
				e.Cancel()
				return err
			}
		case <-doneChan:
			completedCount++
			if completedCount == len(e.workflow.Tasks) {
				// 所有任务完成
				close(taskChan)
				return nil
			}
		case <-e.cancelCh:
			return ErrWorkflowCanceled
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// taskScheduler 任务调度器
func (e *DefaultWorkflowExecutor) taskScheduler(taskChan chan<- Task, taskStatus map[string]TaskStatus) {
	defer func() {
		// 确保在结束时关闭任务通道
		// 注意：这只会在所有任务都被处理后执行
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// 继续尝试调度任务，直到所有任务都完成
	for {
		// 检查是否所有任务都已完成
		allCompleted := true
		for _, status := range taskStatus {
			if status != TaskCompleted && status != TaskFailed && status != TaskSkipped {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			return
		}

		// 尝试调度可执行的任务
		for _, task := range e.workflow.Tasks {
			// 检查任务是否已处理
			if taskStatus[task.ID()] != TaskPending {
				continue
			}

			// 检查任务是否可以执行
			workflowCtx := NewWorkflowContext(context.Background(), e.workflow.Variables, e.logger, e.workflow)
			workflowCtx.Results = make(map[string]*TaskResult)

			// 从taskStatus和ctx.Results构建当前状态
			for taskID, status := range taskStatus {
				if status == TaskCompleted {
					workflowCtx.Results[taskID] = &TaskResult{Status: TaskCompleted}
				}
			}

			if task.CanRun(workflowCtx) {
				// 检查是否取消
				select {
				case <-e.cancelCh:
					return
				default:
				}

				// 调度任务
				taskStatus[task.ID()] = TaskRunning
				task.SetStatus(TaskRunning)

				// 通知观察者任务开始
				e.manager.notifyObservers(func(observer WorkflowObserver) {
					observer.OnTaskStarted(e.workflow, task)
				})

				taskChan <- task
			}
		}

		select {
		case <-ticker.C:
		case <-e.cancelCh:
			return
		}
	}
}

// taskWorker 任务工作协程
func (e *DefaultWorkflowExecutor) taskWorker(taskChan <-chan Task, ctx *WorkflowContext, errChan chan<- error, doneChan chan<- bool) {
	defer e.wg.Done()

	for task := range taskChan {
		// 执行任务
		result, err := e.executeTask(task, ctx)

		if err != nil {
			errChan <- err
			return
		}

		// 记录任务结果
		ctx.Results[task.ID()] = result

		// 标记任务完成
		doneChan <- true
	}
}

// executeTask 执行单个任务
func (e *DefaultWorkflowExecutor) executeTask(task Task, ctx *WorkflowContext) (*TaskResult, error) {
	startTime := time.Now()
	var result *TaskResult
	var err error

	// 如果有重试策略，应用重试逻辑
	if retryStrategy := e.getTaskRetryStrategy(task); retryStrategy != nil {
		result, err = e.executeTaskWithRetry(task, ctx, retryStrategy)
	} else {
		result, err = e.executeTaskOnce(task, ctx)
	}

	// 更新任务状态
	task.SetStatus(result.Status)

	// 更新执行时间
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)

	// 通知观察者
	if result.Status == TaskCompleted {
		e.manager.notifyObservers(func(observer WorkflowObserver) {
			observer.OnTaskCompleted(e.workflow, task, result)
		})
	} else if result.Status == TaskFailed {
		e.manager.notifyObservers(func(observer WorkflowObserver) {
			observer.OnTaskFailed(e.workflow, task, fmt.Errorf(result.Error))
		})
	} else if result.Status == TaskSkipped {
		e.manager.notifyObservers(func(observer WorkflowObserver) {
			observer.OnTaskSkipped(e.workflow, task)
		})
	}

	return result, err
}

// executeTaskOnce 执行单个任务一次
func (e *DefaultWorkflowExecutor) executeTaskOnce(task Task, ctx *WorkflowContext) (*TaskResult, error) {
	// 检查是否取消
	select {
	case <-e.cancelCh:
		return &TaskResult{Status: TaskFailed, Error: ErrWorkflowCanceled.Error()}, ErrWorkflowCanceled
	default:
	}

	// 检查工作流状态
	if e.workflow.Status == StatusPaused {
		// 如果工作流暂停，等待恢复
		e.waitForResume()
	}

	e.logger.Info("Executing task", "task_id", task.ID(), "task_name", task.Name())

	// 执行任务
	result, err := task.Run(ctx)
	if err != nil {
		e.logger.Error("Task execution failed", "task_id", task.ID(), "error", err.Error())
		return &TaskResult{
			Status:    TaskFailed,
			Error:     err.Error(),
			StartTime: time.Now(),
		}, nil
	}

	if result == nil {
		result = &TaskResult{
			Status:    TaskCompleted,
			Output:    make(map[string]interface{}),
			StartTime: time.Now(),
		}
	}

	e.logger.Info("Task executed successfully", "task_id", task.ID())
	return result, nil
}

// executeTaskWithRetry 执行任务并应用重试策略
func (e *DefaultWorkflowExecutor) executeTaskWithRetry(task Task, ctx *WorkflowContext, strategy RetryStrategy) (*TaskResult, error) {
	var lastErr error

	for attempt := 0; attempt <= strategy.MaxAttempts(); attempt++ {
		result, err := e.executeTaskOnce(task, ctx)

		if err == nil && result.Status == TaskCompleted {
			return result, nil
		}

		lastErr = err
		if err == nil && result.Status != TaskFailed {
			// 如果不是失败状态，不重试
			return result, nil
		}

		// 检查是否应该重试
		if !strategy.ShouldRetry(attempt, lastErr) {
			e.logger.Info("Task retry limit reached", "task_id", task.ID(), "attempts", attempt+1)
			return result, nil
		}

		// 等待重试延迟
		delay := strategy.GetDelay(attempt)
		e.logger.Info("Retrying task", "task_id", task.ID(), "attempt", attempt+1, "delay", delay)

		select {
		case <-time.After(delay):
			// 继续重试
		case <-e.cancelCh:
			// 取消重试
			return &TaskResult{Status: TaskFailed, Error: ErrWorkflowCanceled.Error()}, ErrWorkflowCanceled
		}
	}

	return &TaskResult{Status: TaskFailed, Error: lastErr.Error()}, lastErr
}

// waitForResume 等待工作流恢复
func (e *DefaultWorkflowExecutor) waitForResume() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	e.logger.Info("Task execution paused", "workflow_id", e.workflow.ID)

	for {
		select {
		case <-ticker.C:
			if e.workflow.Status == StatusRunning {
				e.logger.Info("Task execution resumed", "workflow_id", e.workflow.ID)
				return
			}
		case <-e.cancelCh:
			return
		}
	}
}

// getConcurrentTaskLimit 获取并发任务限制
func (e *DefaultWorkflowExecutor) getConcurrentTaskLimit() int {
	// 默认并发任务数为工作流任务数的一半，但至少为2
	limit := len(e.workflow.Tasks) / 2
	if limit < 2 {
		limit = 2
	}
	if limit > 10 { // 最大不超过10
		limit = 10
	}

	return limit
}

// getTaskRetryStrategy 获取任务的重试策略
func (e *DefaultWorkflowExecutor) getTaskRetryStrategy(task Task) RetryStrategy {
	// 如果任务是BasicTask类型，尝试获取重试策略
	if basicTask, ok := task.(interface{ GetRetryStrategy() RetryStrategy }); ok {
		return basicTask.GetRetryStrategy()
	}

	return nil
}

// updateWorkflowStatus 更新工作流状态
func (e *DefaultWorkflowExecutor) updateWorkflowStatus(status WorkflowStatus) {
	e.workflow.mutex.Lock()
	e.workflow.Status = status
	e.workflow.UpdatedAt = time.Now()
	e.workflow.mutex.Unlock()
}

// updateWorkflowResult 更新工作流结果
func (e *DefaultWorkflowExecutor) updateWorkflowResult(result *TaskResult) {
	e.workflow.mutex.Lock()
	e.workflow.Result = result
	e.workflow.UpdatedAt = time.Now()
	
	// 如果工作流完成，关闭done通道
	if result.Status == TaskCompleted || result.Status == TaskFailed {
		close(e.workflow.doneCh)
	}
	
	e.workflow.mutex.Unlock()
}

// SequentialWorkflowExecutor 顺序执行器，按顺序执行任务
type SequentialWorkflowExecutor struct {
	manager  *WorkflowManagerImpl
	workflow *Workflow
	logger   logger.Logger
	cancelCh chan struct{}
}

// NewSequentialWorkflowExecutor 创建顺序执行器
func NewSequentialWorkflowExecutor(manager *WorkflowManagerImpl, workflow *Workflow, logger logger.Logger) *SequentialWorkflowExecutor {
	return &SequentialWorkflowExecutor{
		manager:  manager,
		workflow: workflow,
		logger:   logger,
		cancelCh: make(chan struct{}),
	}
}

// Execute 执行工作流（顺序执行）
func (e *SequentialWorkflowExecutor) Execute(workflow *Workflow, ctx context.Context, variables map[string]interface{}) error {
	startTime := time.Now()
	var result *TaskResult

	// 合并变量
	totalVars := MergeVariables(workflow.Variables, variables)

	// 创建工作流上下文
	workflowCtx := NewWorkflowContext(ctx, totalVars, e.logger, workflow)

	// 更新工作流状态
	workflow.Status = StatusRunning

	// 按顺序执行所有任务
	for _, task := range workflow.Tasks {
		// 检查是否取消
		select {
		case <-e.cancelCh:
			return ErrWorkflowCanceled
		default:
		}

		// 执行任务
		taskResult, err := e.executeTask(task, workflowCtx)
		if err != nil {
			result = &TaskResult{
				Status:    TaskFailed,
				Error:     err.Error(),
				StartTime: startTime,
				EndTime:   time.Now(),
				Duration:  time.Since(startTime),
			}

			workflow.Result = result
			workflow.Status = StatusFailed
			return err
		}

		// 记录任务结果
		workflowCtx.Results[task.ID()] = taskResult

		// 如果任务失败，停止执行
		if taskResult.Status != TaskCompleted {
			result = &TaskResult{
				Status:    TaskFailed,
				Error:     fmt.Sprintf("task %s failed with status %s", task.ID(), taskResult.Status),
				StartTime: startTime,
				EndTime:   time.Now(),
				Duration:  time.Since(startTime),
			}

			workflow.Result = result
			workflow.Status = StatusFailed
			return fmt.Errorf("task %s failed", task.ID())
		}
	}

	// 执行成功
	result = &TaskResult{
		Status:    TaskCompleted,
		Output:    workflowCtx.Variables,
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  time.Since(startTime),
	}

	workflow.Result = result
	workflow.Status = StatusCompleted

	return nil
}

// Cancel 取消执行
func (e *SequentialWorkflowExecutor) Cancel() error {
	close(e.cancelCh)
	return nil
}

// executeTask 执行单个任务（顺序执行器）
func (e *SequentialWorkflowExecutor) executeTask(task Task, ctx *WorkflowContext) (*TaskResult, error) {
	// 检查是否取消
	select {
	case <-e.cancelCh:
		return &TaskResult{Status: TaskFailed, Error: ErrWorkflowCanceled.Error()}, ErrWorkflowCanceled
	default:
	}

	e.logger.Info("Executing task sequentially", "task_id", task.ID(), "task_name", task.Name())

	// 通知观察者任务开始
	e.manager.notifyObservers(func(observer WorkflowObserver) {
		observer.OnTaskStarted(e.workflow, task)
	})

	// 执行任务
	result, err := task.Run(ctx)
	if err != nil {
		e.logger.Error("Task execution failed", "task_id", task.ID(), "error", err.Error())

		// 通知观察者任务失败
		e.manager.notifyObservers(func(observer WorkflowObserver) {
			observer.OnTaskFailed(e.workflow, task, err)
		})

		return &TaskResult{
			Status:    TaskFailed,
			Error:     err.Error(),
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  0,
		}, nil
	}

	if result == nil {
		result = &TaskResult{
			Status:    TaskCompleted,
			Output:    make(map[string]interface{}),
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  0,
		}
	}

	// 通知观察者任务完成
	e.manager.notifyObservers(func(observer WorkflowObserver) {
		observer.OnTaskCompleted(e.workflow, task, result)
	})

	e.logger.Info("Task executed successfully", "task_id", task.ID())
	return result, nil
}