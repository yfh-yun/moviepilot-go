// Package interfaces 定义动作系统的接口
package interfaces

import (
	"context"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/actions/types"
)

// Action 动作接口
// 定义动作的基本契约，所有具体动作都必须实现此接口
type Action interface {
	// 基本属性
	Name() string
	Description() string
	Data() map[string]interface{}

	// 状态管理
	IsDone() bool
	IsSuccess() bool
	GetMessage() string
	SetDone(message string)
	SetError(message string)

	// 执行方法
	Execute(ctx context.Context, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error)

	// 缓存管理
	CheckCache(ctx context.Context, workflowID int64, key string) bool
	SaveCache(ctx context.Context, workflowID int64, data interface{}) error
	ClearCache(ctx context.Context, workflowID int64) error

	// 生命周期
	Initialize() error
	Cleanup() error

	// 统计信息
	GetStats() *ActionStats
	ResetStats()

	// 元数据
	GetActionID() string
	SetActionID(actionID string)
	Clone() Action
	ToJSON() map[string]interface{}
}

// ActionStats 动作统计信息
type ActionStats struct {
	ExecuteCount    int64         `json:"execute_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	LastExecute     time.Time     `json:"last_execute"`
	LastSuccess     time.Time     `json:"last_success"`
	LastError       time.Time     `json:"last_error"`
}

// ActionChain 动作链接口
type ActionChain interface {
	// 链管理
	AddAction(action Action) ActionChain
	InsertAction(index int, action Action) ActionChain
	RemoveAction(actionID string) ActionChain
	GetAction(actionID string) Action
	GetActions() []Action
	Count() int

	// 执行控制
	Execute(ctx context.Context, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error)
	ExecuteAsync(ctx context.Context, workflowID int64, params types.ActionParams, context *types.ActionContext) (<-chan *types.ActionContext, <-chan error)
	Stop() error
	IsRunning() bool

	// 状态管理
	GetProgress() float64
	GetCurrentAction() Action
	GetExecutionHistory() []ExecutionRecord

	// 配置
	SetParallel(parallel bool) ActionChain
	SetMaxConcurrency(max int) ActionChain
	SetTimeout(timeout time.Duration) ActionChain
	SetRetryPolicy(policy *RetryPolicy) ActionChain
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	ActionID    string        `json:"action_id"`
	ActionName  string        `json:"action_name"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	Success     bool          `json:"success"`
	Message     string        `json:"message"`
	Error       string        `json:"error,omitempty"`
	RetryCount  int           `json:"retry_count"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries int           `json:"max_retries"`
	Delay      time.Duration `json:"delay"`
	Backoff    BackoffType   `json:"backoff"`
}

// BackoffType 退避策略类型
type BackoffType string

const (
	BackoffFixed    BackoffType = "fixed"
	BackoffLinear   BackoffType = "linear"
	BackoffExponential BackoffType = "exponential"
)

// ActionValidator 动作验证器接口
type ActionValidator interface {
	ValidateConfig(config map[string]interface{}) error
	ValidateParams(params types.ActionParams) error
	ValidateContext(context *types.ActionContext) error
}

// ActionObserver 动作观察者接口
type ActionObserver interface {
	OnActionStart(ctx context.Context, action Action, workflowID int64)
	OnActionComplete(ctx context.Context, action Action, workflowID int64, result *types.ActionContext, err error)
	OnActionError(ctx context.Context, action Action, workflowID int64, err error)
	OnChainStart(ctx context.Context, chain ActionChain, workflowID int64)
	OnChainComplete(ctx context.Context, chain ActionChain, workflowID int64, result *types.ActionContext, err error)
}

// ActionFactory 动作工厂接口
type ActionFactory interface {
	CreateAction(actionType string, config map[string]interface{}) (Action, error)
	GetSupportedActionTypes() []string
	IsActionTypeSupported(actionType string) bool
	RegisterActionType(actionType string, creator ActionCreator) error
	UnregisterActionType(actionType string) error
}

// ActionCreator 动作创建器函数类型
type ActionCreator func(config map[string]interface{}) (Action, error)

// ActionRegistry 动作注册表接口
type ActionRegistry interface {
	// 注册管理
	Register(actionID string, action Action) error
	Unregister(actionID string) error
	Get(actionID string) (Action, bool)
	List() []string
	Count() int
	Clear() error

	// 查找功能
	FindByType(actionType string) []Action
	FindByTag(tag string) []Action
	Search(keyword string) []Action

	// 批量操作
	RegisterBatch(actions map[string]Action) error
	UnregisterBatch(actionIDs []string) error
}