package core

import (
	"fmt"
	"math/rand"
	"time"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// FetchTorrentsAction 实现搜索站点资源动作
type FetchTorrentsAction struct {
	*common.BaseAction
	torrents []map[string]any // 种子列表
}

// NewFetchTorrentsAction 创建新的搜索站点资源动作实例
func NewFetchTorrentsAction() base.Action {
	return &FetchTorrentsAction{
		BaseAction: common.NewBaseAction("fetch_torrents", base.ActionTypeCore),
		torrents:   []map[string]any{},
	}
}

// GetName 获取动作名称
func (a *FetchTorrentsAction) GetName() string {
	return "搜索站点资源"
}

// GetDescription 获取动作描述
func (a *FetchTorrentsAction) GetDescription() string {
	return "搜索站点种子资源列表"
}

// GetData 获取动作参数模板
func (a *FetchTorrentsAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FetchTorrentsParams
	return map[string]any{
		"search_type":  "keyword",
		"name":         nil,
		"year":         nil,
		"type":         nil,
		"season":       nil,
		"sites":        []int{},
		"match_media":  false,
	}
}

// Success 判断动作是否成功
func (a *FetchTorrentsAction) Success() bool {
	// 成功条件与Python一致：动作已完成
	return a.Done()
}

// execute 执行搜索站点资源动作（核心逻辑）
func (a *FetchTorrentsAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取参数
	params := ctx.Input
	if params == nil {
		params = map[string]any{}
	}

	// 解析参数
	searchType, ok := params["search_type"].(string)
	if !ok {
		searchType = "keyword"
	}

	name, _ := params["name"].(string)
	year, _ := params["year"].(string)
	mediaType, _ := params["type"].(string)
	season, _ := params["season"].(int)
	sites, _ := params["sites"].([]int)
	matchMedia, _ := params["match_media"].(bool)

	// 获取搜索服务实例
	searchService, _ := ctx.Services["SearchService"].(interface {
		// SearchByTitle 按关键字搜索
		SearchByTitle(ctx any, title string, sites []int) ([]map[string]any, error)
		// SearchByID 按ID搜索
		SearchByID(ctx any, params map[string]any) ([]map[string]any, error)
		// RecognizeMedia 识别媒体信息
		RecognizeMedia(ctx any, meta map[string]any) (map[string]any, error)
	})

	if searchService == nil {
		return map[string]any{
			"success":      false,
			"error_message": "搜索服务不可用",
		}, nil
	}

	// 初始化随机数生成器
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	if searchType == "keyword" {
		// 按关键字搜索
		if name == "" {
			return map[string]any{
				"success":      true,
				"torrents_found": 0,
			}, nil
		}

		// 调用搜索服务按标题搜索
		torrents, err := searchService.SearchByTitle(ctx, name, sites)
		if err != nil {
			ctx.Logger.Error(fmt.Sprintf("按关键字搜索失败: %s", err.Error()))
			return map[string]any{
				"success":      false,
				"error_message": err.Error(),
			}, err
		}

		// 过滤和处理种子
		for _, torrent := range torrents {
			// 检查工作流是否已停止
			if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
				ctx.Logger.Info("工作流已停止，终止执行")
				break
			}

			// 过滤条件
			// 年份过滤
			if year != "" {
				metaYear, _ := torrent["meta_info"].(map[string]any)["year"].(string)
				if metaYear != year {
					continue
				}
			}

			// 类型过滤
			if mediaType != "" {
				mediaInfo, ok := torrent["media_info"].(map[string]any)
				if ok {
					mt, _ := mediaInfo["type"].(string)
					if mt != mediaType {
						continue
					}
				}
			}

			// 季度过滤
			if season > 0 {
				beginSeason, _ := torrent["meta_info"].(map[string]any)["begin_season"].(int)
				if beginSeason != season {
					continue
				}
			}

			// 识别媒体信息
			if matchMedia {
				metaInfo, _ := torrent["meta_info"].(map[string]any)
				mediaInfo, err := searchService.RecognizeMedia(ctx, metaInfo)
				if err != nil {
					ctx.Logger.Error(fmt.Sprintf("识别媒体信息失败: %s", err.Error()))
					continue
				}
				if mediaInfo == nil {
					ctx.Logger.Warn(fmt.Sprintf("%s 未识别到媒体信息", torrent["torrent_info"].(map[string]any)["title"].(string)))
					continue
				}
				torrent["media_info"] = mediaInfo
			}

			// 添加到结果列表
			a.torrents = append(a.torrents, torrent)
		}
	} else {
		// 搜索媒体列表
		medias, ok := ctx.GlobalContext["medias"].([]map[string]any)
		if !ok || len(medias) == 0 {
			return map[string]any{
				"success":      true,
				"torrents_found": 0,
			}, nil
		}

		for _, media := range medias {
			// 检查工作流是否已停止
			if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
				ctx.Logger.Info("工作流已停止，终止执行")
				break
			}

			// 调用搜索服务按ID搜索
			searchParams := map[string]any{
				"tmdbid":   media["tmdb_id"],
				"doubanid": media["douban_id"],
				"mtype":    media["type"],
				"sites":    sites,
			}
			torrents, err := searchService.SearchByID(ctx, searchParams)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("按ID搜索失败: %s", err.Error()))
				continue
			}

			// 添加种子到结果列表
			for _, torrent := range torrents {
				a.torrents = append(a.torrents, torrent)
			}

			// 随机休眠 5-30秒
			sleepTime := r.Intn(26) + 5 // 5-30秒
			ctx.Logger.Info(fmt.Sprintf("随机休眠 %d 秒 ...", sleepTime))
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}

	// 更新上下文
	if len(a.torrents) > 0 {
		// 获取现有的torrents
		existingTorrents, ok := ctx.GlobalContext["torrents"].([]map[string]any)
		if !ok {
			existingTorrents = []map[string]any{}
		}

		// 添加新搜索到的种子
		existingTorrents = append(existingTorrents, a.torrents...)
		ctx.GlobalContext["torrents"] = existingTorrents
		ctx.Logger.Info(fmt.Sprintf("共搜索到 %d 条资源", len(a.torrents)))
	}

	// 标记任务完成
	message := fmt.Sprintf("搜索到 %d 个资源", len(a.torrents))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":       true,
		"torrents_found": len(a.torrents),
		"message":       message,
	}

	return output, nil
}