package resource

import (
	"fmt"
	"math/rand"
	"time"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// TorrentsAction 实现种子资源获取动作
type TorrentsAction struct {
	*common.BaseAction

	torrents []map[string]any // 种子列表
}

// NewTorrentsAction 创建新的种子资源获取动作实例
func NewTorrentsAction() base.Action {
	return &TorrentsAction{
		BaseAction: common.NewBaseAction("torrents", base.ActionTypeResource),
		torrents:   []map[string]any{},
	}
}

// GetDescription 获取动作描述
func (a *TorrentsAction) GetDescription() string {
	return "搜索站点种子资源列表"
}

// GetData 获取动作参数模板
func (a *TorrentsAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FetchTorrentsParams
	return map[string]any{
		"search_type": map[string]any{
			"type":        "string",
			"description": "搜索类型",
			"default":     "keyword",
		},
		"name": map[string]any{
			"type":        "string",
			"description": "资源名称",
			"default":     nil,
		},
		"year": map[string]any{
			"type":        "string",
			"description": "年份",
			"default":     nil,
		},
		"type": map[string]any{
			"type":        "string",
			"description": "资源类型 (电影/电视剧)",
			"default":     nil,
		},
		"season": map[string]any{
			"type":        "integer",
			"description": "季度",
			"default":     nil,
		},
		"sites": map[string]any{
			"type":        "array",
			"description": "站点列表",
			"default":     []int{},
		},
		"match_media": map[string]any{
			"type":        "boolean",
			"description": "匹配媒体信息",
			"default":     false,
		},
	}
}

// Success 判断动作是否成功
func (a *TorrentsAction) Success() bool {
	// 动作是否成功取决于是否完成，对应Python中的success属性
	return a.Done()
}

// execute 执行种子资源获取动作（核心逻辑）
func (a *TorrentsAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	searchType, _ := ctx.Input["search_type"].(string)
	if searchType == "" {
		searchType = "keyword"
	}
	name, _ := ctx.Input["name"].(string)
	// year, _ := ctx.Input["year"].(string) // 暂时未使用，后续实现时取消注释
	// mediaType, _ := ctx.Input["type"].(string) // 暂时未使用，后续实现时取消注释
	// season, _ := ctx.Input["season"].(int) // 暂时未使用，后续实现时取消注释
	// sites, _ := ctx.Input["sites"].([]int) // 暂时未使用，后续实现时取消注释
	// matchMedia, _ := ctx.Input["match_media"].(bool) // 暂时未使用，后续实现时取消注释

	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// searchService, _ := ctx.Services["SearchService"].(interface{})

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("种子资源获取失败: %v", r))
		}
	}()

	// 初始化种子列表
	a.torrents = []map[string]any{}

	if searchType == "keyword" {
		// 按关键字搜索
		ctx.Logger.Info(fmt.Sprintf("按关键字搜索: %s", name))

		// TODO: 实现按关键字搜索的逻辑
		// torrents = searchchain.search_by_title(title=params.name, sites=params.sites)

		// 遍历处理每个torrent
		// for torrent in torrents:
		//     if global_vars.is_workflow_stopped(workflow_id):
		//         break
		//     if params.year and torrent.meta_info.year != params.year:
		//         continue
		//     if params.type and torrent.media_info and torrent.media_info.type != MediaType(params.type):
		//         continue
		//     if params.season and torrent.meta_info.begin_season != params.season:
		//         continue
		//     # 识别媒体信息
		//     if params.match_media:
		//         torrent.media_info = searchchain.recognize_media(torrent.meta_info)
		//         if not torrent.media_info:
		//             logger.warning(f"{torrent.torrent_info.title} 未识别到媒体信息")
		//             continue
		//     self._torrents.append(torrent)
	} else {
		// 搜索媒体列表
		medias, ok := ctx.GlobalContext["medias"].([]map[string]any)
		if !ok || len(medias) == 0 {
			return map[string]any{"success": true, "torrents_added": 0}, nil
		}

		for _, media := range medias {
			// 检查工作流是否已停止
			if stop, _ := ctx.GlobalContext["stopped"].(bool); stop {
				ctx.Logger.Info("工作流已停止，终止执行")
				break
			}

			tmdbID, _ := media["tmdb_id"].(string)
			doubanID, _ := media["douban_id"].(string)
			// mediaType, _ := media["type"].(string) // 暂时未使用，后续实现时取消注释

			ctx.Logger.Info(fmt.Sprintf("搜索媒体: %s (TMDB: %s, Douban: %s)", media["title"], tmdbID, doubanID))

			// TODO: 实现按ID搜索的逻辑
			// torrents = searchchain.search_by_id(tmdbid=media.tmdb_id,
			//                                     doubanid=media.douban_id,
			//                                     mtype=MediaType(media.type),
			//                                     sites=params.sites)

			// for torrent in torrents:
			//     self._torrents.append(torrent)

			// 随机休眠 5-30秒
			sleepTime := rand.Intn(26) + 5 // 5-30秒
			ctx.Logger.Info(fmt.Sprintf("随机休眠 %d 秒 ...", sleepTime))
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}

	// 更新上下文
	if len(a.torrents) > 0 {
		ctx.Logger.Info(fmt.Sprintf("共搜索到 %d 条资源", len(a.torrents)))
		// 将获取到的种子添加到上下文中
		// 获取现有torrents
		torrents, ok := ctx.GlobalContext["torrents"].([]map[string]any)
		if !ok {
			torrents = []map[string]any{}
		}
		// 合并种子列表
		torrents = append(torrents, a.torrents...)
		// 更新上下文
		ctx.GlobalContext["torrents"] = torrents
	}

	// 标记任务完成
	message := fmt.Sprintf("搜索到 %d 个资源", len(a.torrents))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":        true,
		"torrents_added": len(a.torrents),
		"message":        message,
	}

	return output, nil
}
