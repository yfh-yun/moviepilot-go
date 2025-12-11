package base

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// ActionType 定义动作类型
const (
	ActionTypeFile     = "file"     // 文件处理动作
	ActionTypeResource = "resource" // 资源获取动作
	ActionTypeFilter   = "filter"   // 过滤动作
	ActionTypeCore     = "core"     // 核心业务动作
	ActionTypeSystem   = "system"   // 系统功能动作
)

// ActionStatus 定义动作状态
const (
	ActionStatusPending   = "pending"   // 待执行
	ActionStatusRunning   = "running"   // 执行中
	ActionStatusCompleted = "completed" // 执行完成
	ActionStatusFailed    = "failed"    // 执行失败
	ActionStatusCancelled = "cancelled" // 已取消
)

// ActionResult 定义动作执行结果
type ActionResult struct {
	Success      bool           `json:"success"`                 // 执行是否成功
	ErrorMessage string         `json:"error_message,omitempty"` // 错误信息
	Output       map[string]any `json:"output"`                  // 输出数据
	Duration     time.Duration  `json:"duration"`                // 执行时长
	Status       string         `json:"status"`                  // 执行状态
}

// ActionContext 定义动作执行上下文
type ActionContext struct {
	context.Context
	Logger        *zap.Logger    // 日志记录器
	WorkflowID    string         // 所属工作流ID
	ActionID      string         // 动作ID
	ActionName    string         // 动作名称
	ActionType    string         // 动作类型
	Input         map[string]any // 输入参数
	GlobalContext map[string]any // 全局上下文
	Services      map[string]any // 服务实例
	IsAsync       bool           // 是否异步执行
}

// Action 定义动作接口
type Action interface {
	// GetName 获取动作名称
	GetName() string

	// GetType 获取动作类型
	GetType() string

	// GetDescription 获取动作描述
	GetDescription() string

	// GetData 获取动作参数模板
	GetData() map[string]any

	// Initialize 初始化动作
	Initialize(ctx ActionContext) error

	// IsInitialized 检查动作是否已初始化
	IsInitialized() bool

	// Execute 执行动作
	Execute(ctx ActionContext) (*ActionResult, error)

	// GetStatus 获取动作状态
	GetStatus() string

	// Cancel 取消动作执行
	Cancel() error

	// GetActionID 获取动作ID
	GetActionID() string

	// Success 判断动作是否成功
	Success() bool

	// Done 判断动作是否完成
	Done() bool

	// Message 获取执行信息
	Message() string

	// JobDone 标记动作完成
	JobDone(message string)

	// CheckCache 检查是否处理过
	CheckCache(workflowID, key string) bool

	// SaveCache 保存缓存
	SaveCache(workflowID string, data interface{}) error
}
