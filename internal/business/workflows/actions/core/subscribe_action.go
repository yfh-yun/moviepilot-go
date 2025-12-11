package core

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// SubscribeAction 实现添加订阅动作
type SubscribeAction struct {
	*common.BaseAction
	addedSubscribes []string // 已添加的订阅ID列表
	hasError        bool     // 是否有错误
}

// NewSubscribeAction 创建新的添加订阅动作实例
func NewSubscribeAction() base.Action {
	return &SubscribeAction{
		BaseAction: common.NewBaseAction("subscribe", base.ActionTypeCore),
	}
}

// GetName 获取动作名称
func (a *SubscribeAction) GetName() string {
	return "添加订阅"
}

// GetDescription 获取动作描述
func (a *SubscribeAction) GetDescription() string {
	return "根据媒体列表添加订阅"
}

// GetData 获取动作参数模板
func (a *SubscribeAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的AddSubscribeParams
	return map[string]any{}
}

// Success 判断动作是否成功
func (a *SubscribeAction) Success() bool {
	// 根据执行结果判断动作是否成功
	return !a.hasError
}

// execute 执行添加订阅动作（核心逻辑）
func (a *SubscribeAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 从上下文中获取medias
	medias, ok := ctx.GlobalContext["medias"].([]map[string]any)
	if !ok || len(medias) == 0 {
		return map[string]any{"success": true, "subscribes_added": 0}, nil
	}

	// 获取服务实例 - 使用actions包中定义的接口
	subscribeService, _ := ctx.Services["SubscribeService"].(interface {
		Exists(ctx any, mediainfo map[string]any, meta map[string]any) (bool, error)
		CreateSubscribe(ctx any, subscribe map[string]any) (string, string, error)
		GetSubscribe(ctx any, subscribeID string) (map[string]any, error)
	})

	// 初始化结果跟踪
	a.addedSubscribes = []string{}
	a.hasError = false
	started := false

	// 遍历处理每个media
	for _, media := range medias {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 检查缓存
		mediaType, _ := media["type"].(string)
		title, _ := media["title"].(string)
		year, _ := media["year"].(string)
		seasonStr := ""
		season := 0
		if seasonVal, ok := media["season"].(int); ok {
			season = seasonVal
			seasonStr = fmt.Sprintf("%d", seasonVal)
		} else if seasonStrVal, ok := media["season"].(string); ok {
			seasonStr = seasonStrVal
		}
		cacheKey := fmt.Sprintf("%s-%s-%s-%s", mediaType, title, year, seasonStr)
		if a.CheckCache(ctx.WorkflowID, cacheKey) {
			ctx.Logger.Info(fmt.Sprintf("%s %s 已添加过订阅，跳过", title, year))
			continue
		}

		// 检查订阅是否已存在
		exists := false
		if subscribeService != nil {
			// 创建meta信息
			meta := map[string]any{
				"begin_season": season,
			}

			// 调用Exists方法检查订阅是否已存在
			var err error
			exists, err = subscribeService.Exists(ctx, media, meta)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("检查订阅存在状态失败: %s", err.Error()))
				a.hasError = true
				continue
			}

			if exists {
				ctx.Logger.Info(fmt.Sprintf("%s %s 已存在订阅，跳过", title, year))
				// 保存到缓存，避免重复检查
				a.SaveCache(ctx.WorkflowID, cacheKey)
				continue
			}

			ctx.Logger.Debug(fmt.Sprintf("检查 %s %s 不存在订阅，准备添加", title, year))
		}

		// 添加订阅
		started = true
		if subscribeService != nil {
			// 调用订阅服务创建订阅
			subscribeID, message, err := subscribeService.CreateSubscribe(ctx, media)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("添加订阅失败: %s", err.Error()))
				a.hasError = true
				continue
			}

			if subscribeID == "" {
				ctx.Logger.Error(fmt.Sprintf("添加订阅失败: %s", message))
				a.hasError = true
				continue
			}

			a.addedSubscribes = append(a.addedSubscribes, subscribeID)

			// 保存缓存
			a.SaveCache(ctx.WorkflowID, cacheKey)
			ctx.Logger.Info(fmt.Sprintf("已添加订阅: %s %s (ID: %s)", title, year, subscribeID))
		}
	}

	// 更新上下文
	if len(a.addedSubscribes) > 0 {
		ctx.Logger.Info(fmt.Sprintf("已添加 %d 个订阅", len(a.addedSubscribes)))

		// 获取现有的订阅列表
		existingSubscribes, ok := ctx.GlobalContext["subscribes"].([]any)
		if !ok {
			existingSubscribes = []any{}
		}

		// 添加完整订阅对象到上下文
		for _, sid := range a.addedSubscribes {
			if subscribeService != nil {
				// 获取完整订阅对象
				subscribe, err := subscribeService.GetSubscribe(ctx, sid)
				if err != nil {
					ctx.Logger.Error(fmt.Sprintf("获取订阅详情失败: %s", err.Error()))
					continue
				}
				if subscribe != nil {
					existingSubscribes = append(existingSubscribes, subscribe)
				}
			} else {
				// 如果没有订阅服务，只添加订阅ID
				existingSubscribes = append(existingSubscribes, map[string]any{
					"id": sid,
				})
			}
		}

		// 更新上下文
		ctx.GlobalContext["subscribes"] = existingSubscribes
	} else if started {
		// 如果已开始处理但没有添加任何订阅，标记为错误
		a.hasError = true
	}

	// 标记任务完成
	message := fmt.Sprintf("已添加 %d 个订阅", len(a.addedSubscribes))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":          !a.hasError,
		"subscribes_added": len(a.addedSubscribes),
		"message":          message,
	}

	return output, nil
}
