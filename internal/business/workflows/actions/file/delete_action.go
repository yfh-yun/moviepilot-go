package file

import (
	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// DeleteAction 实现文件删除动作
type DeleteAction struct {
	*common.BaseAction
}

// NewDeleteAction 创建新的文件删除动作实例
func NewDeleteAction() base.Action {
	return &DeleteAction{
		BaseAction: common.NewBaseAction("delete", base.ActionTypeFile),
	}
}

// execute 执行文件删除动作（核心逻辑）
func (a *DeleteAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 从上下文中获取输入参数
	path, _ := ctx.Input["path"].(string)
	recursive, _ := ctx.Input["recursive"].(bool)
	force, _ := ctx.Input["force"].(bool)
	verify, _ := ctx.Input["verify"].(bool)

	// 模拟删除结果
	output := map[string]any{
		"path":      path,
		"recursive": recursive,
		"force":     force,
		"verify":    verify,
		"deleted":   true,
		"success":   true,
	}

	return output, nil
}
