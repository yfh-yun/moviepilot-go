package file

import (
	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// CopyAction 实现文件复制动作
type CopyAction struct {
	*common.BaseAction
}

// NewCopyAction 创建新的文件复制动作实例
func NewCopyAction() base.Action {
	return &CopyAction{
		BaseAction: common.NewBaseAction("copy", base.ActionTypeFile),
	}
}

// execute 执行文件复制动作（核心逻辑）
func (a *CopyAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 从上下文中获取输入参数
	sourcePath, _ := ctx.Input["source_path"].(string)
	destPath, _ := ctx.Input["destination_path"].(string)
	overwrite, _ := ctx.Input["overwrite"].(bool)
	verify, _ := ctx.Input["verify"].(bool)

	// 模拟复制结果
	output := map[string]any{
		"source_path":      sourcePath,
		"destination_path": destPath,
		"overwrite":        overwrite,
		"verify":           verify,
		"copied":           true,
		"success":          true,
	}

	return output, nil
}
