// Package interfaces 定义动作管理器的接口
package interfaces

import (
	"context"

	"moviepilot-go/internal/business/workflows/types"
)

// Manager 动作管理器接口
// 负责管理所有动作的生命周期、执行和监控
type Manager interface {
	// 动作管理
	RegisterAction(action interfaces.Action) error
	UnregisterAction(actionID string) error
	GetAction(actionID string) (interfaces.Action, bool)
	ListActions() []interfaces.Action
	GetActionsByType(actionType string) []interfaces.Action

	// 动作链管理
	CreateChain(chainID string) interfaces.ActionChain
	GetChain(chainID string) (interfaces.ActionChain, bool)
	DeleteChain(chainID string) error
	ListChains() []string

	// 执行管理
	ExecuteAction(ctx context.Context, actionID string, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error)
	ExecuteChain(ctx context.Context, chainID string, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error)
	ExecuteActionAsync(ctx context.Context, actionID string, workflowID int64, params types.ActionParams, context *types.ActionContext) (<-chan *types.ActionContext, <-chan error)
	ExecuteChainAsync(ctx context.Context, chainID string, workflowID int64, params types.ActionParams, context *types.ActionContext) (<-chan *types.ActionContext, <-chan error)

	// 状态管理
	GetExecutionStatus(workflowID int64) (*ExecutionStatus, error)
	StopExecution(workflowID int64) error
	PauseExecution(workflowID int64) error
	ResumeExecution(workflowID int64) error
	ListExecutions() []ExecutionInfo

	// 配置管理
	SetGlobalConfig(config *ManagerConfig) error
	GetGlobalConfig() *ManagerConfig
	SetActionConfig(actionID string, config map[string]interface{}) error
	GetActionConfig(actionID string) (map[string]interface{}, error)

	// 观察者模式
	AddObserver(observer interfaces.ActionObserver) error
	RemoveObserver(observer interfaces.ActionObserver) error
	NotifyObservers(event ObserverEvent)

	// 工厂模式
	SetFactory(factory interfaces.ActionFactory) error
	GetFactory() interfaces.ActionFactory

	// 缓存管理
	SetCache(cache Cache) error
	GetCache() Cache

	// 生命周期
	Initialize() error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error

	// 健康检查
	HealthCheck() *HealthStatus
	GetMetrics() *ManagerMetrics
}

// ExecutionStatus 执行状态
type ExecutionStatus struct {
	WorkflowID      int64     `json:"workflow_id"`
	Status          string    `json:"status"`          // running, completed, failed, cancelled, paused
	StartTime       *int64    `json:"start_time"`
	EndTime         *int64    `json:"end_time"`
	Duration        *int64    `json:"duration"`
	CurrentAction   string    `json:"current_action"`
	CompletedCount  int       `json:"completed_count"`
	TotalCount      int       `json:"total_count"`
	Progress        float64   `json:"progress"`
	Message         string    `json:"message"`
	Error           string    `json:"error,omitempty"`
	RetryCount      int       `json:"retry_count"`
	LastUpdateTime  int64     `json:"last_update_time"`
}

// ExecutionInfo 执行信息
type ExecutionInfo struct {
	WorkflowID      int64  `json:"workflow_id"`
	ActionID        string `json:"action_id"`
	ChainID         string `json:"chain_id,omitempty"`
	Status          string `json:"status"`
	StartTime       int64  `json:"start_time"`
	EndTime         int64  `json:"end_time"`
	Duration        int64  `json:"duration"`
	Success         bool   `json:"success"`
	Message         string `json:"message"`
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	// 执行配置
	DefaultTimeout    int64 `json:"default_timeout"`    // 默认超时时间（秒）
	MaxConcurrency    int   `json:"max_concurrency"`    // 最大并发数
	EnableRetry       bool  `json:"enable_retry"`       // 是否启用重试
	DefaultRetryCount int   `json:"default_retry_count"` // 默认重试次数

	// 缓存配置
	CacheEnabled      bool          `json:"cache_enabled"`       // 是否启用缓存
	CacheTTL          int64         `json:"cache_ttl"`           // 缓存TTL（秒）
	CleanupInterval   int64         `json:"cleanup_interval"`    // 清理间隔（秒）
	CacheSize         int64         `json:"cache_size"`          // 缓存大小限制

	// 监控配置
	MetricsEnabled    bool          `json:"metrics_enabled"`     // 是否启用指标
	TracingEnabled    bool          `json:"tracing_enabled"`     // 是否启用追踪
	LogLevel          string        `json:"log_level"`           // 日志级别

	// 性能配置
	WorkerPoolSize    int           `json:"worker_pool_size"`    // 工作池大小
	QueueSize         int           `json:"queue_size"`          // 队列大小
	BatchSize         int           `json:"batch_size"`          // 批处理大小
	BatchTimeout      int64         `json:"batch_timeout"`       // 批处理超时（秒）

	// 安全配置
	EnableWhitelist   bool          `json:"enable_whitelist"`    // 是否启用白名单
	AllowedActionTypes []string     `json:"allowed_action_types"` // 允许的动作类型
	MaxExecutionTime  int64         `json:"max_execution_time"`  // 最大执行时间（秒）
}

// ObserverEvent 观察者事件
type ObserverEvent struct {
	Type      string                 `json:"type"`       // action_start, action_complete, action_error, chain_start, chain_complete
	ActionID  string                 `json:"action_id"`
	ChainID   string                 `json:"chain_id,omitempty"`
	WorkflowID int64                 `json:"workflow_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl int64) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Exists(ctx context.Context, key string) (bool, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Size(ctx context.Context) (int64, error)
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status     string            `json:"status"`     // healthy, unhealthy, degraded
	Timestamp int64             `json:"timestamp"`
	Checks     map[string]Check `json:"checks"`
	Summary    string            `json:"summary"`
}

// Check 健康检查项
type Check struct {
	Status  string `json:"status"`  // pass, fail, warn
	Message string `json:"message"`
}

// ManagerMetrics 管理器指标
type ManagerMetrics struct {
	// 动作指标
	TotalActions       int64 `json:"total_actions"`
	ActiveActions      int64 `json:"active_actions"`
	CompletedActions   int64 `json:"completed_actions"`
	FailedActions      int64 `json:"failed_actions"`

	// 链指标
	TotalChains        int64 `json:"total_chains"`
	ActiveChains       int64 `json:"active_chains"`
	CompletedChains    int64 `json:"completed_chains"`
	FailedChains       int64 `json:"failed_chains"`

	// 执行指标
	TotalExecutions    int64 `json:"total_executions"`
	ActiveExecutions   int64 `json:"active_executions"`
	CompletedExecutions int64 `json:"completed_executions"`
	FailedExecutions   int64 `json:"failed_executions"`

	// 性能指标
	AverageExecutionTime float64 `json:"average_execution_time"`
	P95ExecutionTime    float64 `json:"p95_execution_time"`
	P99ExecutionTime    float64 `json:"p99_execution_time"`

	// 缓存指标
	CacheHitRate       float64 `json:"cache_hit_rate"`
	CacheSize          int64   `json:"cache_size"`
	CacheEvictions     int64   `json:"cache_evictions"`

	// 资源指标
	MemoryUsage        int64   `json:"memory_usage"`
	CPUUsage           float64 `json:"cpu_usage"`
	GoroutineCount     int64   `json:"goroutine_count"`

	// 时间戳
	LastUpdateTime     int64   `json:"last_update_time"`
}

// WorkflowManager 工作流管理器接口
type WorkflowManager interface {
	// 工作流管理
	CreateWorkflow(workflow *types.Workflow) error
	UpdateWorkflow(workflowID int, workflow *types.Workflow) error
	DeleteWorkflow(workflowID int) error
	GetWorkflow(workflowID int) (*types.Workflow, error)
	ListWorkflows() ([]*types.Workflow, error)
	ListWorkflowsByStatus(status string) ([]*types.Workflow, error)

	// 工作流执行
	ExecuteWorkflow(ctx context.Context, workflowID int, trigger string, params types.ActionParams) (*types.WorkflowExecution, error)
	StopWorkflow(workflowExecutionID int64) error
	PauseWorkflow(workflowExecutionID int64) error
	ResumeWorkflow(workflowExecutionID int64) error

	// 执行历史
	GetWorkflowExecution(executionID int64) (*types.WorkflowExecution, error)
	ListWorkflowExecutions(workflowID int) ([]*types.WorkflowExecution, error)
	ListWorkflowExecutionsByStatus(status string) ([]*types.WorkflowExecution, error)

	// 调度管理
	ScheduleWorkflow(workflowID int, schedule *types.WorkflowSchedule) error
	UnscheduleWorkflow(workflowID int) error
	GetWorkflowSchedule(workflowID int) (*types.WorkflowSchedule, error)
	ListScheduledWorkflows() ([]*types.WorkflowSchedule, error)

	// 生命周期
	Initialize() error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}