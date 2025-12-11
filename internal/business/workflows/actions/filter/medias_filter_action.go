package filter

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// MediasFilterAction 实现媒体过滤动作
type MediasFilterAction struct {
	*common.BaseAction

	medias []map[string]any // 过滤后的媒体列表
}

// NewMediasFilterAction 创建新的媒体过滤动作实例
func NewMediasFilterAction() base.Action {
	return &MediasFilterAction{
		BaseAction: common.NewBaseAction("medias_filter", base.ActionTypeFilter),
		medias:     []map[string]any{},
	}
}

// GetDescription 获取动作描述
func (a *MediasFilterAction) GetDescription() string {
	return "对媒体数据列表进行过滤"
}

// GetData 获取动作参数模板
func (a *MediasFilterAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FilterMediasParams
	return map[string]any{
		"type": map[string]any{
			"type":        "string",
			"description": "媒体类型 (电影/电视剧)",
			"default":     nil,
		},
		"vote": map[string]any{
			"type":        "integer",
			"description": "评分",
			"default":     0,
		},
		"year": map[string]any{
			"type":        "string",
			"description": "年份",
			"default":     nil,
		},
	}
}

// Success 判断动作是否成功
func (a *MediasFilterAction) Success() bool {
	// 动作是否成功取决于是否完成，对应Python中的success属性
	return a.Done()
}

// execute 执行媒体过滤动作（核心逻辑）
func (a *MediasFilterAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	mediaType, _ := ctx.Input["type"].(string)
	vote, _ := ctx.Input["vote"].(int)
	year, _ := ctx.Input["year"].(string)

	// 从上下文中获取medias
	medias, ok := ctx.GlobalContext["medias"].([]map[string]any)
	if !ok || len(medias) == 0 {
		return map[string]any{"success": true, "medias_filtered": 0}, nil
	}

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("媒体过滤失败: %v", r))
		}
	}()

	// 初始化过滤后的媒体列表
	a.medias = []map[string]any{}

	// 遍历处理每个media
	for _, media := range medias {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 过滤媒体类型
		if mediaType != "" {
			mType, _ := media["type"].(string)
			if mType != mediaType {
				continue
			}
		}

		// 过滤评分
		if vote > 0 {
			voteAverage, _ := media["vote_average"].(float64)
			if voteAverage < float64(vote) {
				continue
			}
		}

		// 过滤年份
		if year != "" {
			mYear, _ := media["year"].(string)
			if mYear != year {
				continue
			}
		}

		// 添加到过滤后的媒体列表
		a.medias = append(a.medias, media)
	}

	// 更新上下文
	ctx.Logger.Info(fmt.Sprintf("过滤后剩余 %d 条媒体数据", len(a.medias)))
	ctx.GlobalContext["medias"] = a.medias

	// 标记任务完成
	message := fmt.Sprintf("过滤后剩余 %d 条媒体数据", len(a.medias))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":        true,
		"medias_filtered": len(a.medias),
		"message":        message,
	}

	return output, nil
}
