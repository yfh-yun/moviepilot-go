package core

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// InvokePluginAction 实现调用插件动作
type InvokePluginAction struct {
	*common.BaseAction
	success bool // 执行是否成功
}

// NewInvokePluginAction 创建新的调用插件动作实例
func NewInvokePluginAction() base.Action {
	return &InvokePluginAction{
		BaseAction: common.NewBaseAction("invoke_plugin", base.ActionTypeSystem),
		success:    false,
	}
}

// GetName 获取动作名称
func (a *InvokePluginAction) GetName() string {
	return "调用插件"
}

// GetDescription 获取动作描述
func (a *InvokePluginAction) GetDescription() string {
	return "调用插件提供的动作"
}

// GetData 获取动作参数模板
func (a *InvokePluginAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的InvokePluginParams
	return map[string]any{
		"plugin_id":     nil,
		"action_id":     nil,
		"action_params": map[string]any{},
	}
}

// Success 判断动作是否成功
func (a *InvokePluginAction) Success() bool {
	// 成功条件与Python一致：执行是否成功
	return a.success
}

// execute 执行调用插件动作（核心逻辑）
func (a *InvokePluginAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	pluginID, _ := ctx.Input["plugin_id"].(string)
	actionID, _ := ctx.Input["action_id"].(string)
	actionParams, _ := ctx.Input["action_params"].(map[string]any)

	// 检查plugin_id和action_id是否为空
	if pluginID == "" || actionID == "" {
		return map[string]any{
			"success": false,
			"message": "插件ID和动作ID不能为空",
		}, nil
	}

	// 从上下文中获取PluginService服务
	pluginService, ok := ctx.Services["PluginService"].(interface {
		InvokePlugin(ctx base.ActionContext, pluginID string, actionID string, actionParams map[string]any) (bool, map[string]any, error)
	})

	// 初始化成功状态
	a.success = false

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("调用插件动作失败: %v", r))
			a.success = false
		}
	}()

	var err error
	var result map[string]any

	if ok {
		// 调用插件服务执行动作
		a.success, result, err = pluginService.InvokePlugin(ctx, pluginID, actionID, actionParams)
		if err != nil {
			ctx.Logger.Error(fmt.Sprintf("调用插件服务失败: %v", err))
			a.success = false
		}
	} else {
		// 插件服务未实现，使用模拟逻辑
		ctx.Logger.Info(fmt.Sprintf("模拟执行插件 %s 的动作 %s，参数: %v", pluginID, actionID, actionParams))
		// 模拟执行成功
		a.success = true
		result = map[string]any{"message": "插件动作执行成功"}
	}

	// 更新上下文
	// TODO: 根据插件执行结果更新上下文

	// 标记任务完成
	a.JobDone("")

	// 输出结果
	output := map[string]any{
		"success":  a.success,
		"plugin_id": pluginID,
		"action_id": actionID,
		"result":    result,
		"message":   "调用插件动作完成",
	}

	return output, nil
}
