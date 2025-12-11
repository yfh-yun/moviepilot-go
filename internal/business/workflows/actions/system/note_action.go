package system

import (
	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// NoteAction 实现备注动作
type NoteAction struct {
	*common.BaseAction
}

// NewNoteAction 创建新的备注动作实例
func NewNoteAction() base.Action {
	return &NoteAction{
		BaseAction: common.NewBaseAction("note", base.ActionTypeSystem),
	}
}

// GetName 获取动作名称
func (a *NoteAction) GetName() string {
	return "备注"
}

// GetDescription 获取动作描述
func (a *NoteAction) GetDescription() string {
	return "给工作流添加备注"
}

// GetData 获取动作参数模板
func (a *NoteAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的data属性
	return map[string]any{}
}

// Success 判断动作是否成功
func (a *NoteAction) Success() bool {
	// 备注动作总是成功，对应Python中的success属性
	return true
}

// execute 执行备注动作（核心逻辑）
func (a *NoteAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 备注动作直接返回结果，不做任何操作，对应Python中的execute方法
	output := map[string]any{
		"success": true,
		"message": "备注动作执行成功",
	}

	return output, nil
}
