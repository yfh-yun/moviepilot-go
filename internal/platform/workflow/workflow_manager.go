package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// WorkflowManagerImpl 工作流管理器实现
type WorkflowManagerImpl struct {
	workflows      map[string]*Workflow
	executors      map[string]WorkflowExecutor
	observers      []WorkflowObserver
	workflowMutex  sync.RWMutex
	executorMutex  sync.RWMutex
	observerMutex  sync.RWMutex
	logger         logger.Logger
	wg             sync.WaitGroup
	shutdownCh     chan struct{}
}

// NewWorkflowManager 创建工作流管理器实例
func NewWorkflowManager(logger logger.Logger) *WorkflowManagerImpl {
	return &WorkflowManagerImpl{
		workflows:    make(map[string]*Workflow),
		executors:    make(map[string]WorkflowExecutor),
		observers:    make([]WorkflowObserver, 0),
		logger:       logger,
		shutdownCh:   make(chan struct{}),
	}
}

// Initialize 初始化工作流管理器
func (wm *WorkflowManagerImpl) Initialize() error {
	wm.logger.Info("Initializing workflow manager")

	// 可以在这里添加初始化逻辑
	// 例如：加载预定义的工作流模板等

	return nil
}

// Start 启动工作流管理器
func (wm *WorkflowManagerImpl) Start() error {
	wm.logger.Info("Starting workflow manager")

	// 启动监控工作协程
	wm.wg.Add(1)
	go wm.monitorWorkflowStatus()

	return nil
}

// Stop 停止工作流管理器
func (wm *WorkflowManagerImpl) Stop() error {
	wm.logger.Info("Stopping workflow manager")

	// 发送关闭信号
	close(wm.shutdownCh)

	// 停止所有正在运行的工作流
	wm.workflowMutex.RLock()
	workflowIDs := make([]string, 0, len(wm.workflows))
	for id := range wm.workflows {
		workflowIDs = append(workflowIDs, id)
	}
	wm.workflowMutex.RUnlock()

	for _, id := range workflowIDs {
		if err := wm.StopWorkflow(id); err != nil {
			wm.logger.Error("Failed to stop workflow", "workflow_id", id, "error", err.Error())
		}
	}

	// 等待所有工作协程结束
	wm.wg.Wait()

	return nil
}

// RegisterWorkflow 注册工作流
func (wm *WorkflowManagerImpl) RegisterWorkflow(workflow *Workflow) error {
	if workflow == nil {
		return fmt.Errorf("workflow cannot be nil")
	}

	// 验证工作流
	if err := ValidateWorkflow(workflow); err != nil {
		return fmt.Errorf("invalid workflow: %w", err)
	}

	wm.workflowMutex.Lock()
	defer wm.workflowMutex.Unlock()

	if _, exists := wm.workflows[workflow.ID]; exists {
		return fmt.Errorf("%w: %s", ErrWorkflowAlreadyExists, workflow.ID)
	}

	wm.workflows[workflow.ID] = workflow
	wm.logger.Info("Workflow registered successfully", "workflow_id", workflow.ID, "workflow_name", workflow.Name)

	return nil
}

// UnregisterWorkflow 注销工作流
func (wm *WorkflowManagerImpl) UnregisterWorkflow(id string) error {
	wm.workflowMutex.Lock()
	defer wm.workflowMutex.Unlock()

	workflow, exists := wm.workflows[id]
	if !exists {
		return fmt.Errorf("%w: %s", ErrWorkflowNotFound, id)
	}

	// 检查工作流是否正在运行
	if workflow.Status == StatusRunning {
		return fmt.Errorf("cannot unregister running workflow: %s", id)
	}

	// 移除工作流
	delete(wm.workflows, id)
	
	// 移除对应的执行器
	if executor, exists := wm.executors[id]; exists {
		executor.Cancel()
		delete(wm.executors, id)
	}

	wm.logger.Info("Workflow unregistered successfully", "workflow_id", id)
	return nil
}

// GetWorkflow 获取工作流
func (wm *WorkflowManagerImpl) GetWorkflow(id string) (*Workflow, error) {
	wm.workflowMutex.RLock()
	defer wm.workflowMutex.RUnlock()

	workflow, exists := wm.workflows[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrWorkflowNotFound, id)
	}

	return workflow, nil
}

// ListWorkflows 列出所有工作流
func (wm *WorkflowManagerImpl) ListWorkflows() []*Workflow {
	wm.workflowMutex.RLock()
	defer wm.workflowMutex.RUnlock()

	workflows := make([]*Workflow, 0, len(wm.workflows))
	for _, workflow := range wm.workflows {
		workflows = append(workflows, workflow)
	}

	return workflows
}

// RunWorkflow 运行工作流
func (wm *WorkflowManagerImpl) RunWorkflow(id string, ctx context.Context, variables map[string]interface{}) error {
	// 获取工作流
	workflow, err := wm.GetWorkflow(id)
	if err != nil {
		return err
	}

	// 检查工作流状态
	if workflow.Status != StatusPending && workflow.Status != StatusCompleted && workflow.Status != StatusFailed && workflow.Status != StatusCanceled {
		return fmt.Errorf("cannot run workflow in status: %s", workflow.Status)
	}

	// 创建执行器
	executor := NewDefaultWorkflowExecutor(wm, workflow, wm.logger)

	// 保存执行器
	wm.executorMutex.Lock()
	wm.executors[id] = executor
	wm.executorMutex.Unlock()

	// 重置工作流状态
	workflow.mutex.Lock()
	workflow.Status = StatusPending
	workflow.Result = nil
	workflow.UpdatedAt = time.Now()
	
	// 重置所有任务状态
	for _, task := range workflow.Tasks {
		task.SetStatus(TaskPending)
	}
	workflow.mutex.Unlock()

	// 通知观察者工作流开始
	wm.notifyObservers(func(observer WorkflowObserver) {
		observer.OnWorkflowStarted(workflow)
	})

	// 执行工作流
	go func() {
		defer func() {
			// 清理执行器
			wm.executorMutex.Lock()
			delete(wm.executors, id)
			wm.executorMutex.Unlock()
		}()

		if err := executor.Execute(workflow, ctx, variables); err != nil {
			workflow.mutex.Lock()
			workflow.Status = StatusFailed
			workflow.UpdatedAt = time.Now()
			workflow.mutex.Unlock()

			wm.logger.Error("Workflow execution failed", "workflow_id", id, "error", err.Error())

			// 通知观察者工作流失败
			wm.notifyObservers(func(observer WorkflowObserver) {
				observer.OnWorkflowFailed(workflow, err)
			})
		} else {
			// 执行成功，更新工作流状态
			workflow.mutex.Lock()
			workflow.Status = StatusCompleted
			workflow.UpdatedAt = time.Now()
			workflow.mutex.Unlock()

			// 通知观察者工作流完成
			wm.notifyObservers(func(observer WorkflowObserver) {
				observer.OnWorkflowCompleted(workflow, workflow.Result)
			})
		}
	}()

	return nil
}

// StopWorkflow 停止工作流
func (wm *WorkflowManagerImpl) StopWorkflow(id string) error {
	// 获取工作流
	workflow, err := wm.GetWorkflow(id)
	if err != nil {
		return err
	}

	// 检查工作流状态
	if workflow.Status != StatusRunning {
		return fmt.Errorf("%w: %s", ErrWorkflowNotRunning, id)
	}

	// 获取执行器
	executor, exists := func() (WorkflowExecutor, bool) {
		wm.executorMutex.RLock()
		defer wm.executorMutex.RUnlock()
		return wm.executors[id]
	}()

	if !exists {
		return fmt.Errorf("no active executor found for workflow: %s", id)
	}

	// 取消执行
	if err := executor.Cancel(); err != nil {
		return err
	}

	// 更新工作流状态
	workflow.mutex.Lock()
	workflow.Status = StatusCanceled
	workflow.UpdatedAt = time.Now()
	workflow.mutex.Unlock()

	wm.logger.Info("Workflow stopped", "workflow_id", id)

	// 通知观察者工作流取消
	wm.notifyObservers(func(observer WorkflowObserver) {
		observer.OnWorkflowCanceled(workflow)
	})

	return nil
}

// PauseWorkflow 暂停工作流
func (wm *WorkflowManagerImpl) PauseWorkflow(id string) error {
	// 获取工作流
	workflow, err := wm.GetWorkflow(id)
	if err != nil {
		return err
	}

	// 检查工作流状态
	if workflow.Status != StatusRunning {
		return fmt.Errorf("only running workflows can be paused")
	}

	// 更新工作流状态
	workflow.mutex.Lock()
	workflow.Status = StatusPaused
	workflow.UpdatedAt = time.Now()
	workflow.mutex.Unlock()

	wm.logger.Info("Workflow paused", "workflow_id", id)

	// 通知观察者工作流暂停
	wm.notifyObservers(func(observer WorkflowObserver) {
		observer.OnWorkflowPaused(workflow)
	})

	return nil
}

// ResumeWorkflow 恢复工作流
func (wm *WorkflowManagerImpl) ResumeWorkflow(id string) error {
	// 获取工作流
	workflow, err := wm.GetWorkflow(id)
	if err != nil {
		return err
	}

	// 检查工作流状态
	if workflow.Status != StatusPaused {
		return fmt.Errorf("only paused workflows can be resumed")
	}

	// 更新工作流状态
	workflow.mutex.Lock()
	workflow.Status = StatusRunning
	workflow.UpdatedAt = time.Now()
	workflow.mutex.Unlock()

	wm.logger.Info("Workflow resumed", "workflow_id", id)

	// 通知观察者工作流恢复
	wm.notifyObservers(func(observer WorkflowObserver) {
		observer.OnWorkflowResumed(workflow)
	})

	return nil
}

// GetWorkflowStatus 获取工作流状态
func (wm *WorkflowManagerImpl) GetWorkflowStatus(id string) (WorkflowStatus, error) {
	workflow, err := wm.GetWorkflow(id)
	if err != nil {
		return "", err
	}

	return workflow.Status, nil
}

// GetWorkflowResult 获取工作流执行结果
func (wm *WorkflowManagerImpl) GetWorkflowResult(id string) (*TaskResult, error) {
	workflow, err := wm.GetWorkflow(id)
	if err != nil {
		return nil, err
	}

	return workflow.Result, nil
}

// WaitForWorkflow 等待工作流完成
func (wm *WorkflowManagerImpl) WaitForWorkflow(id string) error {
	workflow, err := wm.GetWorkflow(id)
	if err != nil {
		return err
	}

	// 如果工作流已经完成，直接返回
	if workflow.Status == StatusCompleted || workflow.Status == StatusFailed || workflow.Status == StatusCanceled {
		return nil
	}

	// 等待工作流完成
	doneCh := workflow.doneCh
	select {
	case <-doneCh:
		return nil
	case <-wm.shutdownCh:
		return fmt.Errorf("workflow manager shutting down")
	}
}

// RegisterObserver 注册工作流观察者
func (wm *WorkflowManagerImpl) RegisterObserver(observer WorkflowObserver) {
	wm.observerMutex.Lock()
	defer wm.observerMutex.Unlock()

	wm.observers = append(wm.observers, observer)
	wm.logger.Info("Workflow observer registered")
}

// UnregisterObserver 注销工作流观察者
func (wm *WorkflowManagerImpl) UnregisterObserver(observer WorkflowObserver) {
	wm.observerMutex.Lock()
	defer wm.observerMutex.Unlock()

	for i, obs := range wm.observers {
		if obs == observer {
			wm.observers = append(wm.observers[:i], wm.observers[i+1:]...)
			wm.logger.Info("Workflow observer unregistered")
			break
		}
	}
}

// notifyObservers 通知所有观察者
func (wm *WorkflowManagerImpl) notifyObservers(callback func(observer WorkflowObserver)) {
	wm.observerMutex.RLock()
	observers := make([]WorkflowObserver, len(wm.observers))
	copy(observers, wm.observers)
	wm.observerMutex.RUnlock()

	for _, observer := range observers {
		go callback(observer)
	}
}

// monitorWorkflowStatus 监控工作流状态的协程
func (wm *WorkflowManagerImpl) monitorWorkflowStatus() {
	defer wm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.checkRunningWorkflows()
		case <-wm.shutdownCh:
			return
		}
	}
}

// checkRunningWorkflows 检查运行中的工作流
func (wm *WorkflowManagerImpl) checkRunningWorkflows() {
	wm.workflowMutex.RLock()
	workflows := make([]*Workflow, 0)
	for _, workflow := range wm.workflows {
		if workflow.Status == StatusRunning {
			workflows = append(workflows, workflow)
		}
	}
	wm.workflowMutex.RUnlock()

	for _, workflow := range workflows {
		// 检查工作流是否超时
		if workflow.timeout > 0 {
			if time.Since(workflow.CreatedAt) > workflow.timeout {
				wm.logger.Warn("Workflow timeout detected", "workflow_id", workflow.ID)
				_ = wm.StopWorkflow(workflow.ID)
			}
		}

		// 记录工作流运行状态
		wm.logger.Debug("Monitoring workflow", 
			"workflow_id", workflow.ID, 
			"status", workflow.Status,
			"running_time", time.Since(workflow.CreatedAt).String(),
		)
	}
}

// GetActiveWorkflows 获取活跃的工作流数量
func (wm *WorkflowManagerImpl) GetActiveWorkflows() int {
	wm.workflowMutex.RLock()
	defer wm.workflowMutex.RUnlock()

	count := 0
	for _, workflow := range wm.workflows {
		if workflow.Status == StatusRunning || workflow.Status == StatusPaused {
			count++
		}
	}

	return count
}

// CreateWorkflow 从模板创建工作流
func (wm *WorkflowManagerImpl) CreateWorkflow(id, name, description, version string) *Workflow {
	workflow := NewWorkflow(id, name, description, version)
	return workflow
}

// GetWorkflowStats 获取工作流统计信息
func (wm *WorkflowManagerImpl) GetWorkflowStats() map[string]interface{} {
	wm.workflowMutex.RLock()
	defer wm.workflowMutex.RUnlock()

	stats := map[string]interface{}{
		"total_workflows": len(wm.workflows),
		"status_counts":   make(map[string]int),
		"active_tasks":    0,
	}

	statusCounts := stats["status_counts"].(map[string]int)

	for _, workflow := range wm.workflows {
		statusCounts[string(workflow.Status)]++

		// 计算活跃任务数
		activeTasks := 0
		for _, task := range workflow.Tasks {
			if task.Status() == TaskRunning || task.Status() == TaskPending {
				activeTasks++
			}
		}
		stats["active_tasks"] = stats["active_tasks"].(int) + activeTasks
	}

	return stats
}