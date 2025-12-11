package core

import (
	"fmt"
	"net/http"
	"time"

	"moviepilot-go/internal/business/workflows/actions/base"
	"moviepilot-go/internal/business/workflows/actions/common"
)

// sourceItem 定义数据源项
type sourceItem struct {
	Func    func() []map[string]any `json:"func,omitempty"` // 数据源函数
	Name    string                  `json:"name"`           // 数据源名称
	APIPath string                  `json:"api_path"`       // API路径
}

// FetchMediasAction 实现获取媒体数据动作
type FetchMediasAction struct {
	*common.BaseAction
	medias       []map[string]any // 媒体列表
	hasError     bool             // 是否有错误
	innerSources []sourceItem     // 内部数据源列表
}

// NewFetchMediasAction 创建新的获取媒体数据动作实例
func NewFetchMediasAction() base.Action {
	action := &FetchMediasAction{
		BaseAction: common.NewBaseAction("fetch_medias", base.ActionTypeCore),
		medias:     []map[string]any{},
		hasError:   false,
		innerSources: []sourceItem{
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

	// TODO: 实现事件管理，获取额外数据源
	// 目前先使用默认数据源

	return action
}

// GetName 获取动作名称
func (a *FetchMediasAction) GetName() string {
	return "获取媒体数据"
}

// GetDescription 获取动作描述
func (a *FetchMediasAction) GetDescription() string {
	return "获取榜单等媒体数据列表"
}

// GetData 获取动作参数模板
func (a *FetchMediasAction) GetData() map[string]any {
	// 返回参数模板，对应Python中的FetchMediasParams
	return map[string]any{
		"source_type": "ranking",
		"sources":     []string{},
		"api_path":    nil,
	}
}

// Success 判断动作是否成功
func (a *FetchMediasAction) Success() bool {
	return !a.hasError
}

// getSource 获取数据源
func (a *FetchMediasAction) getSource(apiPath string) *sourceItem {
	for _, s := range a.innerSources {
		if s.APIPath == apiPath {
			return &s
		}
	}
	return nil
}

// execute 执行获取媒体数据动作（核心逻辑）
func (a *FetchMediasAction) execute(ctx base.ActionContext) (map[string]any, error) {
	// 获取参数
	params := ctx.Input
	if params == nil {
		params = map[string]any{}
	}

	// 解析参数
	sourceType, ok := params["source_type"].(string)
	if !ok {
		sourceType = "ranking"
	}

	sources, ok := params["sources"].([]string)
	if !ok {
		sources = []string{}
	}

	apiPath, _ := params["api_path"].(string)

	// 获取推荐服务实例
	recommendService, _ := ctx.Services["RecommendService"].(interface {
		// GetMediasFromSource 从指定数据源获取媒体数据
		GetMediasFromSource(ctx any, apiPath string) ([]map[string]any, error)
	})

	// 获取配置信息
	port, _ := ctx.GlobalContext["port"].(int)
	if port == 0 {
		port = 3001 // 默认端口
	}

	token, _ := ctx.GlobalContext["api_token"].(string)
	if token == "" {
		token = "" // 默认空token
	}

	if sourceType == "ranking" {
		// 从榜单获取数据
		for _, apiPath := range sources {
			// 检查工作流是否已停止
			if stop, _ := ctx.GlobalContext["workflow_stopped"].(bool); stop {
				ctx.Logger.Info("工作流已停止，终止执行")
				break
			}

			source := a.getSource(apiPath)
			if source == nil {
				continue
			}

			ctx.Logger.Info(fmt.Sprintf("获取媒体数据 %s ...", source.Name))

			results := []map[string]any{}

			if recommendService != nil {
				// 使用推荐服务获取数据
				var err error
				results, err = recommendService.GetMediasFromSource(ctx, apiPath)
				if err != nil {
					ctx.Logger.Error(fmt.Sprintf("获取媒体数据失败: %s", err.Error()))
					continue
				}
			} else {
				// 调用内部API获取数据
				internalURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/%s?token=%s", port, apiPath, token)
				client := &http.Client{
					Timeout: 15 * time.Second,
				}
				resp, err := client.Post(internalURL, "application/json", nil)
				if err != nil {
					ctx.Logger.Error(fmt.Sprintf("调用API失败: %s", err.Error()))
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					// TODO: 解析JSON响应
					// 这里简化处理，实际应该使用json.Decoder解析
					results = []map[string]any{}
				}
			}

			if len(results) > 0 {
				ctx.Logger.Info(fmt.Sprintf("%s 获取到 %d 条数据", source.Name, len(results)))
				a.medias = append(a.medias, results...)
			} else {
				ctx.Logger.Error(fmt.Sprintf("%s 获取数据失败", source.Name))
			}
		}
	} else {
		// 从指定API路径获取数据
		if apiPath != "" {
			ctx.Logger.Info(fmt.Sprintf("获取媒体数据 %s ...", apiPath))

			results := []map[string]any{}

			if recommendService != nil {
				// 使用推荐服务获取数据
				var err error
				results, err = recommendService.GetMediasFromSource(ctx, apiPath)
				if err != nil {
					ctx.Logger.Error(fmt.Sprintf("获取媒体数据失败: %s", err.Error()))
					results = []map[string]any{}
				}
			} else {
				// 调用内部API获取数据
				internalURL := fmt.Sprintf("http://127.0.0.1:%d%s?token=%s", port, apiPath, token)
				client := &http.Client{
					Timeout: 15 * time.Second,
				}
				resp, err := client.Post(internalURL, "application/json", nil)
				if err != nil {
					ctx.Logger.Error(fmt.Sprintf("调用API失败: %s", err.Error()))
					results = []map[string]any{}
				} else {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						// TODO: 解析JSON响应
						results = []map[string]any{}
					} else {
						results = []map[string]any{}
					}
				}
			}

			if len(results) > 0 {
				ctx.Logger.Info(fmt.Sprintf("%s 获取到 %d 条数据", apiPath, len(results)))
				a.medias = append(a.medias, results...)
			} else {
				ctx.Logger.Error(fmt.Sprintf("%s 获取数据失败", apiPath))
			}
		}
	}

	// 更新上下文
	if len(a.medias) > 0 {
		// 获取现有的medias
		existingMedias, ok := ctx.GlobalContext["medias"].([]map[string]any)
		if !ok {
			existingMedias = []map[string]any{}
		}

		// 添加新获取的媒体数据
		existingMedias = append(existingMedias, a.medias...)
		ctx.GlobalContext["medias"] = existingMedias
	}

	// 标记任务完成
	message := fmt.Sprintf("获取到 %d 条媒体数据", len(a.medias))
	a.JobDone(message)

	// 输出结果
	output := map[string]any{
		"success":      !a.hasError,
		"medias_added": len(a.medias),
		"message":      message,
	}

	return output, nil
}
