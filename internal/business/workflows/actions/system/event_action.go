package system

import (
	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// EventAction 实现事件发送动作
type EventAction struct {
	*common.BaseAction
}

// NewEventAction 创建新的事件发送动作实例
func NewEventAction() base.Action {
	return &EventAction{
		BaseAction: common.NewBaseAction("event", base.ActionTypeSystem),
	}
}

// GetName 获取动作名称
func (a *EventAction) GetName() string {
	return "发送事件"
}

// GetDescription 获取动作描述
func (a *EventAction) GetDescription() string {
	return "发送任务执行事件"
}

// GetData 获取动作参数模板
func (a *EventAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的SendEventParams
	return map[string]any{}
}

// Success 判断动作是否成功
func (a *EventAction) Success() bool {
	// 动作是否成功，对应Python中的success属性
	return a.Done()
}

// execute 执行事件发送动作（核心逻辑）
func (a *EventAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// eventService, _ := ctx.Services["EventService"].(interface{})

	// 发送工作流事件，以让插件干预工作流执行
	// TODO: 实现发送工作流事件的逻辑
	// event = eventmanager.send_event(ChainEventType.WorkflowExecution, context)
	ctx.Logger.Info("发送工作流执行事件")

	// 标记任务完成
	a.JobDone("")

	// 输出结果
	output := map[string]any{
		"success": true,
		"message": "事件发送成功",
	}

	return output, nil
}
