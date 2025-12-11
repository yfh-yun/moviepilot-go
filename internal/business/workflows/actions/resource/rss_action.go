package resource

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// RSSAction 实现RSS资源获取动作
type RSSAction struct {
	*common.BaseAction

	rssTorrents []map[string]any // RSS资源列表
	hasError    bool             // 是否有错误
}

// NewRSSAction 创建新的RSS资源获取动作实例
func NewRSSAction() base.Action {
	return &RSSAction{
		BaseAction:  common.NewBaseAction("rss", base.ActionTypeResource),
		rssTorrents: []map[string]any{},
		hasError:    false,
	}
}

// GetDescription 获取动作描述
func (a *RSSAction) GetDescription() string {
	return "订阅RSS地址获取资源"
}

// GetData 获取动作参数模板
func (a *RSSAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FetchRssParams
	return map[string]any{
		"url": map[string]any{
			"type":        "string",
			"description": "RSS地址",
			"default":     nil,
		},
		"proxy": map[string]any{
			"type":        "boolean",
			"description": "是否使用代理",
			"default":     false,
		},
		"timeout": map[string]any{
			"type":        "integer",
			"description": "超时时间",
			"default":     15,
		},
		"content_type": map[string]any{
			"type":        "string",
			"description": "Content-Type",
			"default":     nil,
		},
		"referer": map[string]any{
			"type":        "string",
			"description": "Referer",
			"default":     nil,
		},
		"ua": map[string]any{
			"type":        "string",
			"description": "User-Agent",
			"default":     nil,
		},
		"match_media": map[string]any{
			"type":        "string",
			"description": "匹配媒体信息",
			"default":     nil,
		},
	}
}

// Success 判断动作是否成功
func (a *RSSAction) Success() bool {
	// 动作是否成功取决于是否有错误，对应Python中的success属性
	return !a.hasError
}

// execute 执行RSS资源获取动作（核心逻辑）
func (a *RSSAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	url, _ := ctx.Input["url"].(string)
	if url == "" {
		return map[string]any{"success": true, "rss_items_loaded": 0}, nil
	}

	// proxy, _ := ctx.Input["proxy"].(bool) // 暂时未使用，后续实现时取消注释
	timeout, _ := ctx.Input["timeout"].(int)
	if timeout == 0 {
		timeout = 15
	}
	contentType, _ := ctx.Input["content_type"].(string)
	referer, _ := ctx.Input["referer"].(string)
	ua, _ := ctx.Input["ua"].(string)
	matchMedia, _ := ctx.Input["match_media"].(string)

	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// rssService, _ := ctx.Services["RSSService"].(interface{})

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("RSS资源获取失败: %v", r))
			a.hasError = true
		}
	}()

	// 构建请求头
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

	// 请求RSS地址获取数据
	// TODO: 实现请求RSS地址获取数据的逻辑
	// rssItems := RssHelper().parse(url=params.url, proxy=settings.PROXY if params.proxy else None, timeout=params.timeout, headers=headers)
	rssItems := []map[string]any{}
	ctx.Logger.Info(fmt.Sprintf("请求RSS地址 %s 获取数据...", url))

	if rssItems == nil {
		ctx.Logger.Error(fmt.Sprintf("RSS地址 %s 请求失败！", url))
		a.hasError = true
		return map[string]any{
			"success":          false,
			"rss_items_loaded": 0,
			"error_message":    fmt.Sprintf("RSS地址 %s 请求失败！", url),
		}, nil
	}

	if len(rssItems) == 0 {
		ctx.Logger.Error(fmt.Sprintf("RSS地址 %s 未获取到RSS数据！", url))
		return map[string]any{
			"success":          true,
			"rss_items_loaded": 0,
		}, nil
	}

	// 组装种子信息
	a.rssTorrents = []map[string]any{}
	for _, item := range rssItems {
		// 检查工作流是否已停止
		if stop, _ := ctx.GlobalContext["stopped"].(bool); stop {
			ctx.Logger.Info("工作流已停止，终止执行")
			break
		}

		title, ok := item["title"].(string)
		if !ok || title == "" {
			continue
		}

		// 构建种子信息
		torrentInfo := map[string]any{
			"title":     title,
			"enclosure": item["enclosure"],
			"page_url":  item["link"],
			"size":      item["size"],
			"pubdate":   item["pubdate"],
		}

		// 构建元数据
		metaInfo := map[string]any{
			"title":    title,
			"subtitle": item["description"],
		}

		// 构建上下文
		context := map[string]any{
			"meta_info":    metaInfo,
			"torrent_info": torrentInfo,
		}

		// 识别媒体信息（如果需要）
		if matchMedia != "" {
			// TODO: 实现识别媒体信息的逻辑
			// mediainfo = ActionChain().recognize_media(meta)
			ctx.Logger.Debug(fmt.Sprintf("识别 %s 的媒体信息", title))
		}

		// 添加到RSS资源列表
		a.rssTorrents = append(a.rssTorrents, context)
	}

	// 更新上下文
	if len(a.rssTorrents) > 0 {
		ctx.Logger.Info(fmt.Sprintf("获取到 %d 个RSS资源", len(a.rssTorrents)))
		// 将获取到的RSS资源添加到上下文中
		// 获取现有torrents
		torrents, ok := ctx.GlobalContext["torrents"].([]map[string]any)
		if !ok {
			torrents = []map[string]any{}
		}
		// 合并RSS资源到torrents
		torrents = append(torrents, a.rssTorrents...)
		// 更新上下文
		ctx.GlobalContext["torrents"] = torrents
	}

	// 标记任务完成
	message := fmt.Sprintf("获取到 %d 个资源", len(a.rssTorrents))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":          !a.hasError,
		"rss_items_loaded": len(a.rssTorrents),
		"message":          message,
	}

	return output, nil
}
