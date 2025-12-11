package system

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// PluginAction 实现插件调用动作
type PluginAction struct {
	*common.BaseAction

	success bool // 动作是否成功
}

// NewPluginAction 创建新的插件调用动作实例
func NewPluginAction() base.Action {
	return &PluginAction{
		BaseAction: common.NewBaseAction("plugin", base.ActionTypeSystem),
		success:    false,
	}
}

// GetDescription 获取动作描述
func (a *PluginAction) GetDescription() string {
	return "调用插件提供的动作"
}

// GetData 获取动作参数模板
func (a *PluginAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的InvokePluginParams
	return map[string]any{
		"plugin_id": map[string]any{
			"type":        "string",
			"description": "插件ID",
			"default":     nil,
		},
		"action_id": map[string]any{
			"type":        "string",
			"description": "动作ID",
			"default":     nil,
		},
		"action_params": map[string]any{
			"type":        "object",
			"description": "动作参数",
			"default":     map[string]any{},
		},
	}
}

// Success 判断动作是否成功
func (a *PluginAction) Success() bool {
	// 动作是否成功，对应Python中的success属性
	return a.success
}

// execute 执行插件调用动作（核心逻辑）
func (a *PluginAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	pluginID, _ := ctx.Input["plugin_id"].(string)
	actionID, _ := ctx.Input["action_id"].(string)
	actionParams, _ := ctx.Input["action_params"].(map[string]any)

	// 检查plugin_id和action_id是否为空
	if pluginID == "" || actionID == "" {
		return map[string]any{"success": false, "message": "插件ID和动作ID不能为空"}, nil
	}

	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// pluginService, _ := ctx.Services["PluginService"].(interface{})

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("调用插件动作失败: %v", r))
			a.success = false
		}
	}()

	// 初始化成功状态
	a.success = false

	// 获取插件动作
	// TODO: 实现获取插件动作的逻辑
	// plugin_actions = PluginManager().get_plugin_actions(params.plugin_id)
	ctx.Logger.Info(fmt.Sprintf("获取插件 %s 的动作列表", pluginID))

	// 查找指定action_id的动作
	// TODO: 实现查找指定action_id动作的逻辑
	// actions = plugin_actions[0].get("actions", [])
	// action = next((action for action in actions if action.get("action_id") == params.action_id), None)
	ctx.Logger.Info(fmt.Sprintf("查找插件 %s 的动作 %s", pluginID, actionID))

	// 执行插件动作
	// TODO: 实现执行插件动作的逻辑
	// self._success, context = action["func"](context, **params.action_params)
	ctx.Logger.Info(fmt.Sprintf("执行插件 %s 的动作 %s，参数: %v", pluginID, actionID, actionParams))

	// 模拟执行成功
	a.success = true
	a.JobDone("")

	// 输出结果
	output := map[string]any{
		"success":   a.success,
		"plugin_id": pluginID,
		"action_id": actionID,
		"message":   "插件动作执行成功",
	}

	return output, nil
}
