package core

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// MessageAction 实现发送消息动作
type MessageAction struct {
	*common.BaseAction
}

// NewMessageAction 创建新的发送消息动作实例
func NewMessageAction() base.Action {
	return &MessageAction{
		BaseAction: common.NewBaseAction("send_message", base.ActionTypeCore),
	}
}

// GetName 获取动作名称
func (a *MessageAction) GetName() string {
	return "发送消息"
}

// GetDescription 获取动作描述
func (a *MessageAction) GetDescription() string {
	return "发送任务执行消息"
}

// GetData 获取动作参数模板
func (a *MessageAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的SendMessageParams
	return map[string]any{
		"client": map[string]any{
			"type":        "array",
			"description": "消息渠道",
			"default":     []string{},
		},
		"userid": map[string]any{
			"type":        "string",
			"description": "用户ID",
			"default":     nil,
		},
	}
}

// Success 判断动作是否成功
func (a *MessageAction) Success() bool {
	// 动作是否成功，对应Python中的success属性
	return a.Done()
}

// execute 执行发送消息动作（核心逻辑）
func (a *MessageAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	client, _ := ctx.Input["client"].([]string)
	userid, _ := ctx.Input["userid"].(string)

	// 从上下文中获取进度
	progress, _ := ctx.GlobalContext["progress"].(float64)

	// 构建消息文本
	msgText := fmt.Sprintf("当前进度：%.1f%%", progress)

	// 从上下文中获取执行历史
	executeHistory, ok := ctx.GlobalContext["execute_history"].([]map[string]any)
	if ok && len(executeHistory) > 0 {
		index := 1
		for _, history := range executeHistory {
			message, ok := history["message"].(string)
			if !ok || message == "" {
				continue
			}
			action, _ := history["action"].(string)
			msgText += fmt.Sprintf("\n%d. %s：%s", index, action, message)
			index++
		}

		// 发送消息
		if len(client) == 0 {
			client = []string{""}
		}

		// 从上下文中获取系统配置服务
		systemConfigService, _ := ctx.Services["SystemConfigService"].(interface {
			GetConfig(ctx base.ActionContext, key string) (string, error)
		})

		// 获取MP_DOMAIN配置
		mpDomain := ""
		if systemConfigService != nil {
			mpDomain, _ = systemConfigService.GetConfig(ctx, "MP_DOMAIN")
		}

		// 从上下文中获取ActionChain服务
		actionChain, _ := ctx.Services["ActionChainService"].(interface {
			PostMessage(ctx base.ActionContext, notification map[string]any) error
		})

		// 遍历client列表，发送消息
		for _, c := range client {
			// 构建通知对象
			notification := map[string]any{
				"source": c,
				"userid": userid,
				"title":  "【工作流执行结果】",
				"text":   msgText,
				"link":   fmt.Sprintf("%s#/workflow", mpDomain),
			}

			// 发送消息
			if actionChain != nil {
				err := actionChain.PostMessage(ctx, notification)
				if err != nil {
					ctx.Logger.Error(fmt.Sprintf("发送消息失败：%v", err))
				}
			} else {
				// 如果ActionChain服务不可用，记录日志
				ctx.Logger.Info(fmt.Sprintf("发送消息：%s", msgText))
			}
		}
	}

	// 标记任务完成
	a.JobDone("")

	// 输出结果
	output := map[string]any{
		"success": true,
		"message": "消息发送成功",
	}

	return output, nil
}
