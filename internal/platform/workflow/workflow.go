package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// WorkflowStatus 工作流状态枚举
type WorkflowStatus string

const (
	// StatusPending 待执行
	StatusPending WorkflowStatus = "pending"
	// StatusRunning 运行中
	StatusRunning WorkflowStatus = "running"
	// StatusCompleted 已完成
	StatusCompleted WorkflowStatus = "completed"
	// StatusFailed 失败
	StatusFailed WorkflowStatus = "failed"
	// StatusCanceled 已取消
	StatusCanceled WorkflowStatus = "canceled"
	// StatusPaused 已暂停
	StatusPaused WorkflowStatus = "paused"
)

// TaskStatus 任务状态枚举
type TaskStatus string

const (
	// TaskPending 待执行
	TaskPending TaskStatus = "pending"
	// TaskRunning 运行中
	TaskRunning TaskStatus = "running"
	// TaskCompleted 已完成
	TaskCompleted TaskStatus = "completed"
	// TaskFailed 失败
	TaskFailed TaskStatus = "failed"
	// TaskSkipped 跳过
	TaskSkipped TaskStatus = "skipped"
)

// Error 定义工作流相关错误
var (
	ErrWorkflowNotFound      = fmt.Errorf("workflow not found")
	ErrWorkflowAlreadyExists = fmt.Errorf("workflow already exists")
	ErrWorkflowNotRunning    = fmt.Errorf("workflow not running")
	ErrTaskNotFound          = fmt.Errorf("task not found")
	ErrTaskAlreadyExists     = fmt.Errorf("task already exists")
	ErrInvalidWorkflowStatus = fmt.Errorf("invalid workflow status")
	ErrInvalidTaskStatus     = fmt.Errorf("invalid task status")
	ErrWorkflowCanceled      = fmt.Errorf("workflow canceled")
	ErrWorkflowTimeout       = fmt.Errorf("workflow timeout")
)

// TaskResult 任务执行结果
type TaskResult struct {
	Status    TaskStatus         `json:"status"`    // 任务状态
	Output    map[string]interface{} `json:"output"` // 任务输出
	Error     string             `json:"error"`     // 错误信息
	StartTime time.Time          `json:"start_time"` // 开始时间
	EndTime   time.Time          `json:"end_time"`   // 结束时间
	Duration  time.Duration      `json:"duration"`   // 执行时长
}

// WorkflowContext 工作流上下文，用于在任务之间传递数据

type WorkflowContext struct {
	Context   context.Context  // 基础上下文
	Variables map[string]interface{} // 全局变量
	Results   map[string]*TaskResult // 任务执行结果
	Logger    logger.Logger    // 日志记录器
	Workflow  *Workflow        // 工作流实例
}

// NewWorkflowContext 创建新的工作流上下文
func NewWorkflowContext(ctx context.Context, variables map[string]interface{}, logger logger.Logger, workflow *Workflow) *WorkflowContext {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	return &WorkflowContext{
		Context:   ctx,
		Variables: variables,
		Results:   make(map[string]*TaskResult),
		Logger:    logger,
		Workflow:  workflow,
	}
}

// Task 任务接口
type Task interface {
	// ID 获取任务ID
	ID() string

	// Name 获取任务名称
	Name() string

	// Description 获取任务描述
	Description() string

	// Dependencies 获取依赖的任务ID列表
	Dependencies() []string

	// Run 执行任务
	Run(ctx *WorkflowContext) (*TaskResult, error)

	// Status 获取任务状态
	Status() TaskStatus

	// SetStatus 设置任务状态
	SetStatus(status TaskStatus)

	// CanRun 检查任务是否可以执行
	CanRun(ctx *WorkflowContext) bool

	// Priority 获取任务优先级
	Priority() int
}

// Workflow 工作流定义
type Workflow struct {
	ID          string            // 工作流ID
	Name        string            // 工作流名称
	Description string            // 工作流描述
	Version     string            // 版本号
	CreatedAt   time.Time         // 创建时间
	UpdatedAt   time.Time         // 更新时间
	Status      WorkflowStatus    // 状态
	Tasks       map[string]Task   // 任务列表
	Variables   map[string]interface{} // 全局变量
	Metadata    map[string]string // 元数据
	Result      *TaskResult       // 工作流执行结果
	wg          sync.WaitGroup
	mutex       sync.RWMutex
	doneCh      chan struct{}
	timeout     time.Duration
	parent      *Workflow        // 父工作流（用于子工作流）
	children    []*Workflow      // 子工作流列表
}

// NewWorkflow 创建新的工作流
func NewWorkflow(id, name, description, version string) *Workflow {
	now := time.Now()
	return &Workflow{
		ID:          id,
		Name:        name,
		Description: description,
		Version:     version,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      StatusPending,
		Tasks:       make(map[string]Task),
		Variables:   make(map[string]interface{}),
		Metadata:    make(map[string]string),
		doneCh:      make(chan struct{}),
		children:    make([]*Workflow, 0),
	}
}

// AddTask 添加任务到工作流
func (w *Workflow) AddTask(task Task) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, exists := w.Tasks[task.ID()]; exists {
		return fmt.Errorf("%w: task %s", ErrTaskAlreadyExists, task.ID())
	}

	w.Tasks[task.ID()] = task
	w.UpdatedAt = time.Now()
	return nil
}

// GetTask 获取任务
func (w *Workflow) GetTask(id string) (Task, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	task, exists := w.Tasks[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	return task, nil
}

// RemoveTask 移除任务
func (w *Workflow) RemoveTask(id string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, exists := w.Tasks[id]; !exists {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}

	delete(w.Tasks, id)
	w.UpdatedAt = time.Now()
	return nil
}

// SetVariable 设置全局变量
func (w *Workflow) SetVariable(key string, value interface{}) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.Variables[key] = value
	w.UpdatedAt = time.Now()
}

// GetVariable 获取全局变量
func (w *Workflow) GetVariable(key string) (interface{}, bool) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	value, exists := w.Variables[key]
	return value, exists
}

// SetMetadata 设置元数据
func (w *Workflow) SetMetadata(key, value string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.Metadata[key] = value
	w.UpdatedAt = time.Now()
}

// GetMetadata 获取元数据
func (w *Workflow) GetMetadata(key string) (string, bool) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	value, exists := w.Metadata[key]
	return value, exists
}

// AddChild 添加子工作流
func (w *Workflow) AddChild(child *Workflow) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	child.parent = w
	w.children = append(w.children, child)
	w.UpdatedAt = time.Now()
}

// GetChildren 获取子工作流列表
func (w *Workflow) GetChildren() []*Workflow {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	children := make([]*Workflow, len(w.children))
	copy(children, w.children)
	return children
}

// SetTimeout 设置超时时间
func (w *Workflow) SetTimeout(timeout time.Duration) {
	w.timeout = timeout
}

// WorkflowManager 工作流管理器接口
type WorkflowManager interface {
	// RegisterWorkflow 注册工作流
	RegisterWorkflow(workflow *Workflow) error

	// UnregisterWorkflow 注销工作流
	UnregisterWorkflow(id string) error

	// GetWorkflow 获取工作流
	GetWorkflow(id string) (*Workflow, error)

	// ListWorkflows 列出所有工作流
	ListWorkflows() []*Workflow

	// RunWorkflow 运行工作流
	RunWorkflow(id string, ctx context.Context, variables map[string]interface{}) error

	// StopWorkflow 停止工作流
	StopWorkflow(id string) error

	// PauseWorkflow 暂停工作流
	PauseWorkflow(id string) error

	// ResumeWorkflow 恢复工作流
	ResumeWorkflow(id string) error

	// GetWorkflowStatus 获取工作流状态
	GetWorkflowStatus(id string) (WorkflowStatus, error)

	// GetWorkflowResult 获取工作流执行结果
	GetWorkflowResult(id string) (*TaskResult, error)

	// WaitForWorkflow 等待工作流完成
	WaitForWorkflow(id string) error
}

// WorkflowExecutor 工作流执行器接口
type WorkflowExecutor interface {
	// Execute 执行工作流
	Execute(workflow *Workflow, ctx context.Context, variables map[string]interface{}) error

	// Cancel 取消执行
	Cancel() error
}

// WorkflowObserver 工作流观察者接口
type WorkflowObserver interface {
	// OnWorkflowStarted 工作流开始时回调
	OnWorkflowStarted(workflow *Workflow)

	// OnWorkflowCompleted 工作流完成时回调
	OnWorkflowCompleted(workflow *Workflow, result *TaskResult)

	// OnWorkflowFailed 工作流失败时回调
	OnWorkflowFailed(workflow *Workflow, err error)

	// OnWorkflowCanceled 工作流取消时回调
	OnWorkflowCanceled(workflow *Workflow)

	// OnWorkflowPaused 工作流暂停时回调
	OnWorkflowPaused(workflow *Workflow)

	// OnWorkflowResumed 工作流恢复时回调
	OnWorkflowResumed(workflow *Workflow)

	// OnTaskStarted 任务开始时回调
	OnTaskStarted(workflow *Workflow, task Task)

	// OnTaskCompleted 任务完成时回调
	OnTaskCompleted(workflow *Workflow, task Task, result *TaskResult)

	// OnTaskFailed 任务失败时回调
	OnTaskFailed(workflow *Workflow, task Task, err error)

	// OnTaskSkipped 任务跳过时回调
	OnTaskSkipped(workflow *Workflow, task Task)
}

// Conditional 条件接口
type Conditional interface {
	// Evaluate 评估条件
	Evaluate(ctx *WorkflowContext) bool
}

// RetryStrategy 重试策略接口
type RetryStrategy interface {
	// ShouldRetry 判断是否应该重试
	ShouldRetry(attempt int, err error) bool

	// GetDelay 获取重试延迟
	GetDelay(attempt int) time.Duration

	// MaxAttempts 获取最大重试次数
	MaxAttempts() int
}

// BasicTask 基础任务实现
type BasicTask struct {
	id          string
	name        string
	description string
	dependencies []string
	status      TaskStatus
	priority    int
	retryStrategy RetryStrategy
}

// NewBasicTask 创建基础任务
func NewBasicTask(id, name, description string, dependencies []string, priority int) *BasicTask {
	return &BasicTask{
		id:           id,
		name:         name,
		description:  description,
		dependencies: dependencies,
		status:       TaskPending,
		priority:     priority,
	}
}

// ID 获取任务ID
func (t *BasicTask) ID() string {
	return t.id
}

// Name 获取任务名称
func (t *BasicTask) Name() string {
	return t.name
}

// Description 获取任务描述
func (t *BasicTask) Description() string {
	return t.description
}

// Dependencies 获取依赖的任务ID列表
func (t *BasicTask) Dependencies() []string {
	return t.dependencies
}

// Status 获取任务状态
func (t *BasicTask) Status() TaskStatus {
	return t.status
}

// SetStatus 设置任务状态
func (t *BasicTask) SetStatus(status TaskStatus) {
	t.status = status
}

// Priority 获取任务优先级
func (t *BasicTask) Priority() int {
	return t.priority
}

// SetRetryStrategy 设置重试策略
func (t *BasicTask) SetRetryStrategy(strategy RetryStrategy) {
	t.retryStrategy = strategy
}

// GetRetryStrategy 获取重试策略
func (t *BasicTask) GetRetryStrategy() RetryStrategy {
	return t.retryStrategy
}

// CanRun 检查任务是否可以执行
func (t *BasicTask) CanRun(ctx *WorkflowContext) bool {
	// 检查任务状态
	if t.Status() != TaskPending {
		return false
	}

	// 检查所有依赖任务是否已完成
	for _, depID := range t.Dependencies() {
		result, exists := ctx.Results[depID]
		if !exists || result.Status != TaskCompleted {
			return false
		}
	}

	return true
}

// Run 执行任务（需要被子类实现）
func (t *BasicTask) Run(ctx *WorkflowContext) (*TaskResult, error) {
	return nil, fmt.Errorf("method Run not implemented for task %s", t.ID())
}

// ValidateWorkflow 验证工作流的正确性
func ValidateWorkflow(workflow *Workflow) error {
	if workflow == nil {
		return fmt.Errorf("workflow is nil")
	}

	if workflow.ID == "" {
		return fmt.Errorf("workflow ID cannot be empty")
	}

	if len(workflow.Tasks) == 0 {
		return fmt.Errorf("workflow must contain at least one task")
	}

	// 检查任务依赖是否存在
	taskIDs := make(map[string]bool)
	for id := range workflow.Tasks {
		taskIDs[id] = true
	}

	for _, task := range workflow.Tasks {
		for _, depID := range task.Dependencies() {
			if !taskIDs[depID] {
				return fmt.Errorf("task %s has invalid dependency %s", task.ID(), depID)
			}
		}

		// 检查循环依赖
		if hasCycle := checkCycleDependency(workflow, task.ID(), make(map[string]bool)); hasCycle {
			return fmt.Errorf("workflow contains circular dependency involving task %s", task.ID())
		}
	}

	return nil
}

// 检查任务之间是否存在循环依赖
func checkCycleDependency(workflow *Workflow, taskID string, visited map[string]bool) bool {
	if visited[taskID] {
		return true // 检测到循环
	}

	task, exists := workflow.Tasks[taskID]
	if !exists {
		return false
	}

	visited[taskID] = true
	defer delete(visited, taskID)

	for _, depID := range task.Dependencies() {
		if checkCycleDependency(workflow, depID, visited) {
			return true
		}
	}

	return false
}

// MergeVariables 合并变量
func MergeVariables(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 复制基础变量
	for k, v := range base {
		result[k] = v
	}

	// 覆盖变量
	for k, v := range override {
		result[k] = v
	}

	return result
}