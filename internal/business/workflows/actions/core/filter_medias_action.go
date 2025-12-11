package core

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// FilterMediasAction 实现过滤媒体数据动作
type FilterMediasAction struct {
	*common.BaseAction
	medias []map[string]any // 过滤后的媒体列表
}

// NewFilterMediasAction 创建新的过滤媒体数据动作实例
func NewFilterMediasAction() base.Action {
	return &FilterMediasAction{
		BaseAction: common.NewBaseAction("filter_medias", base.ActionTypeFilter),
		medias:     []map[string]any{},
	}
}

// GetName 获取动作名称
func (a *FilterMediasAction) GetName() string {
	return "过滤媒体数据"
}

// GetDescription 获取动作描述
func (a *FilterMediasAction) GetDescription() string {
	return "对媒体数据列表进行过滤"
}

// GetData 获取动作参数模板
func (a *FilterMediasAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FilterMediasParams
	return map[string]any{
		"type": nil,
		"vote": 0,
		"year": nil,
	}
}

// Success 判断动作是否成功
func (a *FilterMediasAction) Success() bool {
	// 成功条件与Python一致：动作已完成
	return a.Done()
}

// execute 执行过滤媒体数据动作（核心逻辑）
func (a *FilterMediasAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取上下文
	medias, ok := ctx.GlobalContext["medias"].([]map[string]any)
	if !ok || len(medias) == 0 {
		// 没有媒体数据，直接返回
		return map[string]any{
			"success":         true,
			"medias_filtered": 0,
			"filtered_count":  0,
		}, nil
	}

	// 获取参数
	params := ctx.Input
	if params == nil {
		params = map[string]any{}
	}

	// 解析参数
	mediaType, _ := params["type"].(string)
	vote, _ := params["vote"].(int)
	year, _ := params["year"].(string)

	// 过滤媒体数据
	filteredMedias := []map[string]any{}

	for _, media := range medias {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 类型过滤
		if mediaType != "" {
			mediaTypeVal, _ := media["type"].(string)
			if mediaTypeVal != mediaType {
				continue
			}
		}

		// 评分过滤
		if vote > 0 {
			voteAverage, ok := media["vote_average"].(float64)
			if !ok || voteAverage < float64(vote) {
				continue
			}
		}

		// 年份过滤
		if year != "" {
			mediaYear, _ := media["year"].(string)
			if mediaYear != year {
				continue
			}
		}

		// 添加到过滤后的列表
		filteredMedias = append(filteredMedias, media)
	}

	// 更新上下文
	ctx.GlobalContext["medias"] = filteredMedias
	a.medias = filteredMedias

	// 记录日志
	filterCount := len(medias) - len(filteredMedias)
	ctx.Logger.Info(fmt.Sprintf("过滤后剩余 %d 条媒体数据，共过滤掉 %d 条", len(filteredMedias), filterCount))

	// 标记任务完成
	message := fmt.Sprintf("过滤后剩余 %d 条媒体数据", len(filteredMedias))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":         true,
		"medias_filtered": len(filteredMedias),
		"filtered_count":  filterCount,
		"message":         message,
	}

	return output, nil
}
