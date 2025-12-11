package actions

// 从base包导入并重新导出所有类型，避免循环导入
import base "moviepilot-go/internal/business/workflows/actions/base"

// 动作类型常量
const (
	ActionTypeFile     = base.ActionTypeFile     // 文件处理动作
	ActionTypeResource = base.ActionTypeResource // 资源获取动作
	ActionTypeFilter   = base.ActionTypeFilter   // 过滤动作
	ActionTypeCore     = base.ActionTypeCore     // 核心业务动作
	ActionTypeSystem   = base.ActionTypeSystem   // 系统功能动作
)

// 动作状态常量
const (
	ActionStatusPending   = base.ActionStatusPending   // 待执行
	ActionStatusRunning   = base.ActionStatusRunning   // 执行中
	ActionStatusCompleted = base.ActionStatusCompleted // 执行完成
	ActionStatusFailed    = base.ActionStatusFailed    // 执行失败
	ActionStatusCancelled = base.ActionStatusCancelled // 已取消
)

// ActionResult 定义动作执行结果
type ActionResult = base.ActionResult

// ActionContext 定义动作执行上下文
type ActionContext = base.ActionContext

// Action 定义动作接口
type Action = base.Action
