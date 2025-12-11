package filter

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// TorrentsFilterAction 实现种子过滤动作
type TorrentsFilterAction struct {
	*common.BaseAction

	torrents []map[string]any // 过滤后的种子列表
}

// NewTorrentsFilterAction 创建新的种子过滤动作实例
func NewTorrentsFilterAction() base.Action {
	return &TorrentsFilterAction{
		BaseAction: common.NewBaseAction("torrents_filter", base.ActionTypeFilter),
		torrents:   []map[string]any{},
	}
}

// GetDescription 获取动作描述
func (a *TorrentsFilterAction) GetDescription() string {
	return "对资源列表数据进行过滤"
}

// GetData 获取动作参数模板
func (a *TorrentsFilterAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FilterTorrentsParams
	return map[string]any{
		"rule_groups": map[string]any{
			"type":        "array",
			"description": "规则组",
			"default":     []string{},
		},
		"quality": map[string]any{
			"type":        "string",
			"description": "资源质量",
			"default":     nil,
		},
		"resolution": map[string]any{
			"type":        "string",
			"description": "资源分辨率",
			"default":     nil,
		},
		"effect": map[string]any{
			"type":        "string",
			"description": "特效",
			"default":     nil,
		},
		"include": map[string]any{
			"type":        "string",
			"description": "包含规则",
			"default":     nil,
		},
		"exclude": map[string]any{
			"type":        "string",
			"description": "排除规则",
			"default":     nil,
		},
		"size": map[string]any{
			"type":        "string",
			"description": "资源大小范围（MB）",
			"default":     nil,
		},
	}
}

// Success 判断动作是否成功
func (a *TorrentsFilterAction) Success() bool {
	// 动作是否成功取决于是否完成，对应Python中的success属性
	return a.Done()
}

// execute 执行种子过滤动作（核心逻辑）
func (a *TorrentsFilterAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	ruleGroups, _ := ctx.Input["rule_groups"].([]string)
	quality, _ := ctx.Input["quality"].(string)
	resolution, _ := ctx.Input["resolution"].(string)
	effect, _ := ctx.Input["effect"].(string)
	include, _ := ctx.Input["include"].(string)
	exclude, _ := ctx.Input["exclude"].(string)
	size, _ := ctx.Input["size"].(string)

	// 从上下文中获取torrents
	torrents, ok := ctx.GlobalContext["torrents"].([]map[string]any)
	if !ok || len(torrents) == 0 {
		return map[string]any{"success": true, "torrents_filtered": 0}, nil
	}

	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// torrentService, _ := ctx.Services["TorrentService"].(interface{})
	// actionChainService, _ := ctx.Services["ActionChainService"].(interface{})

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("种子过滤失败: %v", r))
		}
	}()

	// 初始化过滤后的种子列表
	a.torrents = []map[string]any{}

	// 遍历处理每个torrent
	for _, torrent := range torrents {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 构建过滤参数
		filterParams := map[string]any{
			"quality":    quality,
			"resolution": resolution,
			"effect":     effect,
			"include":    include,
			"exclude":    exclude,
			"size":       size,
		}

		// 使用TorrentHelper().filter_torrent进行过滤
		// TODO: 实现TorrentHelper().filter_torrent的逻辑
		passedFirstFilter := true
		ctx.Logger.Debug(fmt.Sprintf("使用filterParams %v 过滤种子 %v", filterParams, torrent["torrent_info"]))

		// 使用ActionChain().filter_torrents进行进一步过滤
		// TODO: 实现ActionChain().filter_torrents的逻辑
		passedSecondFilter := true
		ctx.Logger.Debug(fmt.Sprintf("使用ruleGroups %v 过滤种子 %v", ruleGroups, torrent["torrent_info"]))

		// 如果通过两个过滤，添加到过滤后的种子列表
		if passedFirstFilter && passedSecondFilter {
			a.torrents = append(a.torrents, torrent)
		}
	}

	// 更新上下文
	ctx.Logger.Info(fmt.Sprintf("过滤后剩余 %d 个资源", len(a.torrents)))
	ctx.GlobalContext["torrents"] = a.torrents

	// 标记任务完成
	message := fmt.Sprintf("过滤后剩余 %d 个资源", len(a.torrents))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":           true,
		"torrents_filtered": len(a.torrents),
		"message":           message,
	}

	return output, nil
}
