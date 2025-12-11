package core

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// FetchRssAction 实现获取RSS资源列表动作
type FetchRssAction struct {
	*common.BaseAction
	rssTorrents []map[string]any // RSS种子列表
	hasError    bool             // 是否有错误
}

// NewFetchRssAction 创建新的获取RSS资源列表动作实例
func NewFetchRssAction() base.Action {
	return &FetchRssAction{
		BaseAction:  common.NewBaseAction("fetch_rss", base.ActionTypeCore),
		rssTorrents: []map[string]any{},
		hasError:    false,
	}
}

// GetName 获取动作名称
func (a *FetchRssAction) GetName() string {
	return "获取RSS资源"
}

// GetDescription 获取动作描述
func (a *FetchRssAction) GetDescription() string {
	return "订阅RSS地址获取资源"
}

// GetData 获取动作参数模板
func (a *FetchRssAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FetchRssParams
	return map[string]any{
		"url":          "",
		"proxy":        false,
		"timeout":      15,
		"content_type": nil,
		"referer":      nil,
		"ua":           nil,
		"match_media":  nil,
	}
}

// Success 判断动作是否成功
func (a *FetchRssAction) Success() bool {
	return !a.hasError
}

// execute 执行获取RSS资源列表动作（核心逻辑）
func (a *FetchRssAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取参数
	params := ctx.Input
	if params == nil {
		params = map[string]any{}
	}

	// 解析参数
	url, ok := params["url"].(string)
	if !ok || url == "" {
		// URL为空，直接返回
		return map[string]any{
			"success":   true,
			"rss_items": 0,
		}, nil
	}

	proxy, _ := params["proxy"].(bool)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 15 // 默认超时时间
	}

	contentType, _ := params["content_type"].(string)
	referer, _ := params["referer"].(string)
	ua, _ := params["ua"].(string)
	matchMedia, _ := params["match_media"].(string)

	// 构建headers
	headers := map[string]string{}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	if referer != "" {
		headers["Referer"] = referer
	}
	if ua != "" {
		headers["User-Agent"] = ua
	}

	// 获取RSS服务实例
	rssService, _ := ctx.Services["RSSService"].(interface {
		// ParseRSS 解析RSS源
		ParseRSS(ctx any, url string, options map[string]any) ([]map[string]any, error)
	})

	// 获取ActionChain实例（用于识别媒体信息）
	actionChain, _ := ctx.Services["ActionChain"].(interface {
		// RecognizeMedia 识别媒体信息
		RecognizeMedia(ctx any, meta map[string]any) (map[string]any, error)
	})

	// 获取RSS数据
	rssItems := []map[string]any{}
	var err error
	if rssService != nil {
		// 调用RSS服务解析RSS
		rssOptions := map[string]any{
			"proxy":   proxy,
			"timeout": timeout,
			"headers": headers,
		}
		rssItems, err = rssService.ParseRSS(ctx, url, rssOptions)
		if err != nil {
			ctx.Logger.Error(fmt.Sprintf("RSS地址 %s 请求失败: %s", url, err.Error()))
			a.hasError = true
			return map[string]any{
				"success":       false,
				"error_message": err.Error(),
			}, err
		}
	}

	if len(rssItems) == 0 {
		ctx.Logger.Error(fmt.Sprintf("RSS地址 %s 未获取到RSS数据！", url))
		return map[string]any{
			"success":   true,
			"rss_items": 0,
		}, nil
	}

	// 组装种子
	for _, item := range rssItems {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		// 获取标题
		title, ok := item["title"].(string)
		if !ok || title == "" {
			continue
		}

		// 构建torrentinfo
		torrentInfo := map[string]any{
			"title":       title,
			"enclosure":   item["enclosure"],
			"page_url":    item["link"],
			"size":        item["size"],
			"description": item["description"],
		}

		// 处理发布时间
		if pubdate, ok := item["pubdate"].(string); ok {
			torrentInfo["pubdate"] = pubdate
		}

		// 构建meta信息
		meta := map[string]any{
			"title":    torrentInfo["title"],
			"subtitle": torrentInfo["description"],
		}

		// 识别媒体信息（如果需要）
		var mediaInfo map[string]any
		if matchMedia != "" && actionChain != nil {
			mediaInfo, err = actionChain.RecognizeMedia(ctx, meta)
			if err != nil {
				ctx.Logger.Error(fmt.Sprintf("识别媒体信息失败: %s", err.Error()))
				continue
			}
			if mediaInfo == nil {
				ctx.Logger.Warn(fmt.Sprintf("%s 未识别到媒体信息", title))
				continue
			}
		}

		// 构建上下文对象
		contextObj := map[string]any{
			"meta_info":    meta,
			"media_info":   mediaInfo,
			"torrent_info": torrentInfo,
		}

		// 添加到RSS种子列表
		a.rssTorrents = append(a.rssTorrents, contextObj)
	}

	// 更新上下文
	if len(a.rssTorrents) > 0 {
		// 获取现有的torrents
		existingTorrents, ok := ctx.GlobalContext["torrents"].([]map[string]any)
		if !ok {
			existingTorrents = []map[string]any{}
		}

		// 添加新获取的RSS种子
		existingTorrents = append(existingTorrents, a.rssTorrents...)
		ctx.GlobalContext["torrents"] = existingTorrents
	}

	// 标记任务完成
	message := fmt.Sprintf("获取到 %d 个RSS资源", len(a.rssTorrents))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":   !a.hasError,
		"rss_items": len(a.rssTorrents),
		"message":   message,
	}

	return output, nil
}
