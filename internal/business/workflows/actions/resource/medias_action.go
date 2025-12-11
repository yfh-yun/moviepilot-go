package resource

import (
	"fmt"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// Source 定义数据源结构
type Source struct {
	Name    string      `json:"name"`
	APIPath string      `json:"api_path"`
	Func    interface{} `json:"func,omitempty"`
}

// MediasAction 实现媒体资源获取动作
type MediasAction struct {
	*common.BaseAction

	medias       []map[string]any // 媒体列表
	hasError     bool             // 是否有错误
	innerSources []Source         // 内置数据源
}

// NewMediasAction 创建新的媒体资源获取动作实例
func NewMediasAction() base.Action {
	action := &MediasAction{
		BaseAction: common.NewBaseAction("medias", base.ActionTypeResource),
		medias:     []map[string]any{},
		hasError:   false,
		innerSources: []Source{
			{
				Name:    "流行趋势",
				APIPath: "recommend/tmdb_trending",
			},
			{
				Name:    "正在热映",
				APIPath: "recommend/douban_showing",
			},
			{
				Name:    "Bangumi每日放送",
				APIPath: "recommend/bangumi_calendar",
			},
			{
				Name:    "TMDB热门电影",
				APIPath: "recommend/tmdb_movies",
			},
			{
				Name:    "TMDB热门电视剧",
				APIPath: "recommend/tmdb_tvs?with_original_language=zh|en|ja|ko",
			},
			{
				Name:    "豆瓣热门电影",
				APIPath: "recommend/douban_movie_hot",
			},
			{
				Name:    "豆瓣热门电视剧",
				APIPath: "recommend/douban_tv_hot",
			},
			{
				Name:    "豆瓣热门动漫",
				APIPath: "recommend/douban_tv_animation",
			},
			{
				Name:    "豆瓣最新电影",
				APIPath: "recommend/douban_movies",
			},
			{
				Name:    "豆瓣最新电视剧",
				APIPath: "recommend/douban_tvs",
			},
			{
				Name:    "豆瓣电影TOP250",
				APIPath: "recommend/douban_movie_top250",
			},
			{
				Name:    "豆瓣国产剧集榜",
				APIPath: "recommend/douban_tv_weekly_chinese",
			},
			{
				Name:    "豆瓣全球剧集榜",
				APIPath: "recommend/douban_tv_weekly_global",
			},
		},
	}

	//  TODO: 实现事件广播，获取额外数据源
	// 这里可以添加事件广播逻辑，获取额外的数据源

	return action
}

// GetDescription 获取动作描述
func (a *MediasAction) GetDescription() string {
	return "获取榜单等媒体数据列表"
}

// GetData 获取动作参数模板
func (a *MediasAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FetchMediasParams
	return map[string]any{
		"source_type": map[string]any{
			"type":        "string",
			"description": "来源",
			"default":     "ranking",
		},
		"sources": map[string]any{
			"type":        "array",
			"description": "榜单",
			"default":     []string{},
		},
		"api_path": map[string]any{
			"type":        "string",
			"description": "API路径",
			"default":     nil,
		},
	}
}

// Success 判断动作是否成功
func (a *MediasAction) Success() bool {
	// 动作是否成功取决于是否有错误，对应Python中的success属性
	return !a.hasError
}

// getSource 获取数据源
func (a *MediasAction) getSource(apiPath string) *Source {
	for i, s := range a.innerSources {
		if s.APIPath == apiPath {
			return &a.innerSources[i]
		}
	}
	return nil
}

// execute 执行媒体资源获取动作（核心逻辑）
func (a *MediasAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取输入参数
	sourceType, _ := ctx.Input["source_type"].(string)
	if sourceType == "" {
		sourceType = "ranking"
	}
	sources, _ := ctx.Input["sources"].([]string)
	apiPath, _ := ctx.Input["api_path"].(string)

	// 获取服务实例 - 暂时未使用，后续实现时取消注释
	// recommendService, _ := ctx.Services["RecommendService"].(interface{})

	// 使用defer-recover来处理可能的panic
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error(fmt.Sprintf("获取媒体数据失败: %v", r))
			a.hasError = true
		}
	}()

	if sourceType == "ranking" {
		// 处理排行榜类型
		for _, apiPath := range sources {
			// 检查工作流是否已停止
			if stop, _ := ctx.GlobalContext["stopped"].(bool); stop {
				ctx.Logger.Info("工作流已停止，终止执行")
				break
			}

			source := a.getSource(apiPath)
			if source == nil {
				continue
			}

			ctx.Logger.Info(fmt.Sprintf("获取媒体数据 %s ...", source.Name))

			// 获取媒体数据
			results := []map[string]any{}
			// TODO: 实现获取媒体数据的逻辑
			// 这里模拟获取媒体数据，实际应该调用recommendService获取
			// if source.Func != nil {
			//     results = source.Func()
			// } else {
			//     // 调用内部API获取数据
			//     api_url = f"http://127.0.0.1:{settings.PORT}/api/v1/{source.APIPath}?token={settings.API_TOKEN}"
			//     res = RequestUtils(timeout=15).post_res(api_url)
			//     if res:
			//         results = res.json()
			// }

			if len(results) > 0 {
				ctx.Logger.Info(fmt.Sprintf("%s 获取到 %d 条数据", source.Name, len(results)))
				a.medias = append(a.medias, results...)
			} else {
				ctx.Logger.Error(fmt.Sprintf("%s 获取数据失败", source.Name))
				a.hasError = true
			}
		}
	} else {
		// 处理自定义API路径
		if apiPath != "" {
			ctx.Logger.Info(fmt.Sprintf("获取媒体数据 %s ...", apiPath))

			// 调用内部API获取数据
			// TODO: 实现调用内部API获取数据的逻辑
			results := []map[string]any{}
			// api_url = f"http://127.0.0.1:{settings.PORT}{apiPath}?token={settings.API_TOKEN}"
			// res = RequestUtils(timeout=15).post_res(api_url)
			// if res:
			//     results = res.json()

			if len(results) > 0 {
				ctx.Logger.Info(fmt.Sprintf("%s 获取到 %d 条数据", apiPath, len(results)))
				a.medias = append(a.medias, results...)
			} else {
				ctx.Logger.Error(fmt.Sprintf("%s 获取数据失败", apiPath))
				a.hasError = true
			}
		}
	}

	// 更新上下文中的medias
	if len(a.medias) > 0 {
		ctx.GlobalContext["medias"] = a.medias
	}

	// 标记任务完成
	message := fmt.Sprintf("获取到 %d 条媒数据", len(a.medias))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":      !a.hasError,
		"medias_added": len(a.medias),
		"message":      message,
	}

	return output, nil
}
