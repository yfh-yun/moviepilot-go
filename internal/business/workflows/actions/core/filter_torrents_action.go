package core

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// FilterTorrentsAction 实现过滤资源数据动作
type FilterTorrentsAction struct {
	*common.BaseAction
	torrents []map[string]any // 过滤后的资源列表
}

// NewFilterTorrentsAction 创建新的过滤资源数据动作实例
func NewFilterTorrentsAction() base.Action {
	return &FilterTorrentsAction{
		BaseAction: common.NewBaseAction("filter_torrents", base.ActionTypeFilter),
		torrents:   []map[string]any{},
	}
}

// GetDescription 获取动作描述
func (a *FilterTorrentsAction) GetDescription() string {
	return "对资源列表数据进行过滤"
}

// GetData 获取动作参数模板
func (a *FilterTorrentsAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FilterTorrentsParams
	return map[string]any{
		"rule_groups": []string{},
		"quality":     nil,
		"resolution":  nil,
		"effect":      nil,
		"include":     nil,
		"exclude":     nil,
		"size":        nil,
	}
}

// execute 执行过滤资源数据动作（核心逻辑）
func (a *FilterTorrentsAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取上下文
	torrents, ok := ctx.GlobalContext["torrents"].([]map[string]any)
	if !ok || len(torrents) == 0 {
		// 没有资源数据，直接返回
		return map[string]any{
			"success":          true,
			"torrents_filtered": 0,
			"filtered_count":   0,
		}, nil
	}

	// 获取参数
	params := ctx.Input
	if params == nil {
		params = map[string]any{}
	}

	// 解析参数
	ruleGroups, _ := params["rule_groups"].([]string)
	quality, _ := params["quality"].(string)
	resolution, _ := params["resolution"].(string)
	effect, _ := params["effect"].(string)
	include, _ := params["include"].(string)
	exclude, _ := params["exclude"].(string)
	size, _ := params["size"].(string)

	// 过滤资源数据
	filteredTorrents := []map[string]any{}

	for _, torrent := range torrents {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 获取torrent_info
		torrentInfo, _ := torrent["torrent_info"].(map[string]any)

		// 构建过滤参数
		filterParams := map[string]any{
			"quality":    quality,
			"resolution": resolution,
			"effect":     effect,
			"include":    include,
			"exclude":    exclude,
			"size":       size,
		}

		// 从上下文中获取TorrentHelper服务 - 目前未实现，使用模拟逻辑
		// torrentHelper, _ := ctx.Services["TorrentHelper"].(subscribe.TorrentHelper)
		// 模拟TorrentHelper.FilterTorrent逻辑
		passedFirstFilter := true
		ctx.Logger.Debug(fmt.Sprintf("使用filterParams %v 过滤种子 %v", filterParams, torrentInfo))

		// 从上下文中获取ActionChain服务 - 目前未实现，使用模拟逻辑
		// actionChain, _ := ctx.Services["ActionChain"].(*actions.ActionChain)
		// 模拟ActionChain.FilterTorrents逻辑
		passedSecondFilter := true
		ctx.Logger.Debug(fmt.Sprintf("使用ruleGroups %v 过滤种子 %v", ruleGroups, torrentInfo))

		// 如果通过两个过滤，添加到过滤后的列表
		if passedFirstFilter && passedSecondFilter {
			filteredTorrents = append(filteredTorrents, torrent)
		}
	}

	// 更新上下文
	ctx.GlobalContext["torrents"] = filteredTorrents
	a.torrents = filteredTorrents

	// 记录日志
	filterCount := len(torrents) - len(filteredTorrents)
	ctx.Logger.Info(fmt.Sprintf("过滤后剩余 %d 个资源，共过滤掉 %d 个", len(filteredTorrents), filterCount))

	// 标记任务完成
	message := fmt.Sprintf("过滤后剩余 %d 个资源", len(filteredTorrents))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":          true,
		"torrents_filtered": len(filteredTorrents),
		"filtered_count":   filterCount,
		"message":          message,
	}

	return output, nil
}
