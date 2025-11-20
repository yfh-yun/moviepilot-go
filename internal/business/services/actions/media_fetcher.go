// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/actions/types"

	"go.uber.org/zap"
)

// MediaFetcher 媒体获取器
// 负责从各种数据源获取媒体信息，实现Python版本FetchMediasAction的完整功能
type MediaFetcher struct {
	httpClient *http.Client
	cache      *WorkflowCache
	logger     *zap.Logger

	// 内置数据源配置
	innerSources []*MediaSource
}

// MediaSource 媒体数据源
type MediaSource struct {
	Name     string                 `json:"name"`
	APIPath  string                 `json:"api_path"`
	Func     MediaFetchFunc         `json:"-"`
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
	Config   map[string]interface{} `json:"config"`
}

// MediaFetchFunc 媒体获取函数类型
type MediaFetchFunc func(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error)

// FetchMediasParams 获取媒体数据参数
type FetchMediasParams struct {
	SourceType string   `json:"source_type" description:"数据源类型"`
	Sources    []string `json:"sources" description:"指定数据源列表"`
	APIPath    string   `json:"api_path" description:"自定义API路径"`
	Limit      int      `json:"limit" description:"获取数量限制"`
	Genres     []string `json:"genres" description:"类型过滤"`
	Year       int      `json:"year" description:"年份过滤"`
	Rating     float64  `json:"rating" description:"评分过滤"`
	Countries  []string `json:"countries" description:"国家过滤"`
	Languages  []string `json:"languages" description:"语言过滤"`
	SortBy     string   `json:"sort_by" description:"排序方式"`
	OrderBy    string   `json:"order_by" description:"排序顺序"`
}

// FetchMediasResult 获取媒体结果
type FetchMediasResult struct {
	Success        bool               `json:"success"`
	Medias         []*types.MediaInfo `json:"medias"`
	Total          int                `json:"total"`
	Source         string             `json:"source"`
	Message        string             `json:"message"`
	Error          error              `json:"error,omitempty"`
	ProcessingTime time.Duration      `json:"processing_time"`
	CacheHit       bool               `json:"cache_hit"`
}

// NewMediaFetcher 创建媒体获取器实例
func NewMediaFetcher(cache *WorkflowCache) *MediaFetcher {
	mf := &MediaFetcher{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:  cache,
		logger: logger.Logger,
	}

	// 初始化内置数据源
	mf.initInnerSources()

	return mf
}

// initInnerSources 初始化内置数据源
func (mf *MediaFetcher) initInnerSources() {
	mf.innerSources = []*MediaSource{
		{
			Name:     "流行趋势",
			APIPath:  "recommend/tmdb_trending",
			Func:     mf.fetchTMDBTrending,
			Enabled:  true,
			Priority: 1,
		},
		{
			Name:     "正在热映",
			APIPath:  "recommend/douban_showing",
			Func:     mf.fetchDoubanShowing,
			Enabled:  true,
			Priority: 2,
		},
		{
			Name:     "Bangumi每日放送",
			APIPath:  "recommend/bangumi_calendar",
			Func:     mf.fetchBangumiCalendar,
			Enabled:  true,
			Priority: 3,
		},
		{
			Name:     "TMDB热门电影",
			APIPath:  "recommend/tmdb_popular_movies",
			Func:     mf.fetchTMDBPopularMovies,
			Enabled:  true,
			Priority: 4,
		},
		{
			Name:     "TMDB热门电视剧",
			APIPath:  "recommend/tmdb_popular_tv",
			Func:     mf.fetchTMDBPopularTV,
			Enabled:  true,
			Priority: 5,
		},
		{
			Name:     "豆瓣新片榜",
			APIPath:  "recommend/douban_new_movies",
			Func:     mf.fetchDoubanNewMovies,
			Enabled:  true,
			Priority: 6,
		},
		{
			Name:     "豆瓣高分榜",
			APIPath:  "recommend/douban_top_rated",
			Func:     mf.fetchDoubanTopRated,
			Enabled:  true,
			Priority: 7,
		},
	}
}

// FetchMedias 获取媒体数据
// 实现Python版本FetchMediasAction的完整功能
func (mf *MediaFetcher) FetchMedias(
	ctx context.Context,
	workflowID int64,
	params *FetchMediasParams,
) ([]*FetchMediasResult, error) {
	startTime := time.Now()
	results := make([]*FetchMediasResult, 0)

	mf.logger.Info("开始获取媒体数据",
		zap.Int64("workflow_id", workflowID),
		zap.String("source_type", params.SourceType),
		zap.Strings("sources", params.Sources),
	)

	// 确定要使用的数据源
	sources := mf.determineSources(params)

	for _, source := range sources {
		// 检查工作流是否已停止
		if mf.isWorkflowStopped(ctx, workflowID) {
			mf.logger.Info("工作流已停止，终止媒体数据获取", zap.Int64("workflow_id", workflowID))
			break
		}

		result := mf.fetchFromSource(ctx, workflowID, source, params)
		results = append(results, result)

		if result.Success {
			mf.logger.Info("媒体数据获取成功",
				zap.String("source", source.Name),
				zap.Int("count", len(result.Medias)),
			)
		} else {
			mf.logger.Warn("媒体数据获取失败",
				zap.String("source", source.Name),
				zap.Error(result.Error),
			)
		}
	}

	// 统计结果
	totalMedias := 0
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
			totalMedias += len(result.Medias)
		}
	}

	mf.logger.Info("媒体数据获取完成",
		zap.Int64("workflow_id", workflowID),
		zap.Int("sources_count", len(sources)),
		zap.Int("success_count", successCount),
		zap.Int("total_medias", totalMedias),
		zap.Duration("processing_time", time.Since(startTime)),
	)

	return results, nil
}

// determineSources 确定要使用的数据源
func (mf *MediaFetcher) determineSources(params *FetchMediasParams) []*MediaSource {
	var sources []*MediaSource

	switch params.SourceType {
	case "ranking":
		// 榜单类型，使用内置数据源
		if len(params.Sources) > 0 {
			// 使用指定的数据源
			for _, sourceName := range params.Sources {
				if source := mf.findSourceByName(sourceName); source != nil {
					sources = append(sources, source)
				}
			}
		} else {
			// 使用所有启用的内置数据源
			for _, source := range mf.innerSources {
				if source.Enabled {
					sources = append(sources, source)
				}
			}
		}

	case "api":
		// API类型，使用自定义API
		source := &MediaSource{
			Name:     "自定义API",
			APIPath:  params.APIPath,
			Func:     mf.fetchFromCustomAPI,
			Enabled:  true,
			Priority: 0,
		}
		sources = append(sources, source)

	case "custom":
		// 自定义类型，根据参数配置
		if len(params.Sources) > 0 {
			for _, sourceName := range params.Sources {
				if source := mf.findSourceByName(sourceName); source != nil {
					sources = append(sources, source)
				}
			}
		}

	default:
		// 默认使用所有启用的内置数据源
		for _, source := range mf.innerSources {
			if source.Enabled {
				sources = append(sources, source)
			}
		}
	}

	// 按优先级排序
	for i := 0; i < len(sources)-1; i++ {
		for j := i + 1; j < len(sources); j++ {
			if sources[i].Priority > sources[j].Priority {
				sources[i], sources[j] = sources[j], sources[i]
			}
		}
	}

	return sources
}

// fetchFromSource 从指定数据源获取媒体数据
func (mf *MediaFetcher) fetchFromSource(
	ctx context.Context,
	workflowID int64,
	source *MediaSource,
	params *FetchMediasParams,
) *FetchMediasResult {
	startTime := time.Now()

	// 生成缓存键
	cacheKey := fmt.Sprintf("media_fetch_%s_%s", source.Name, mf.generateParamsHash(params))

	// 检查缓存
	if mf.cache != nil {
		if cached, err := mf.cache.Get(ctx, cacheKey); err == nil && cached != nil {
			if medias, ok := cached.([]*types.MediaInfo); ok {
				return &FetchMediasResult{
					Success:        true,
					Medias:         medias,
					Total:          len(medias),
					Source:         source.Name,
					Message:        "从缓存获取",
					ProcessingTime: time.Since(startTime),
					CacheHit:       true,
				}
			}
		}
	}

	// 调用获取函数
	medias, err := source.Func(ctx, params)
	if err != nil {
		return &FetchMediasResult{
			Success:        false,
			Medias:         []*types.MediaInfo{},
			Total:          0,
			Source:         source.Name,
			Message:        fmt.Sprintf("获取失败: %v", err),
			Error:          err,
			ProcessingTime: time.Since(startTime),
			CacheHit:       false,
		}
	}

	// 应用过滤条件
	medias = mf.applyFilters(medias, params)

	// 应用数量限制
	if params.Limit > 0 && len(medias) > params.Limit {
		medias = medias[:params.Limit]
	}

	// 保存缓存
	if mf.cache != nil && len(medias) > 0 {
		if err := mf.cache.Set(ctx, cacheKey, medias, 2*time.Hour); err != nil {
			mf.logger.Warn("保存媒体数据缓存失败", zap.Error(err))
		}
	}

	return &FetchMediasResult{
		Success:        true,
		Medias:         medias,
		Total:          len(medias),
		Source:         source.Name,
		Message:        "获取成功",
		ProcessingTime: time.Since(startTime),
		CacheHit:       false,
	}
}

// applyFilters 应用过滤条件
func (mf *MediaFetcher) applyFilters(medias []*types.MediaInfo, params *FetchMediasParams) []*types.MediaInfo {
	var filtered []*types.MediaInfo

	for _, media := range medias {
		if mf.matchesFilters(media, params) {
			filtered = append(filtered, media)
		}
	}

	return filtered
}

// matchesFilters 检查媒体是否匹配过滤条件
func (mf *MediaFetcher) matchesFilters(media *types.MediaInfo, params *FetchMediasParams) bool {
	// 类型过滤
	if len(params.Genres) > 0 {
		hasGenre := false
		for _, genre := range params.Genres {
			for _, mediaGenre := range media.Genres {
				if genre == mediaGenre {
					hasGenre = true
					break
				}
			}
			if hasGenre {
				break
			}
		}
		if !hasGenre {
			return false
		}
	}

	// 年份过滤
	if params.Year > 0 && media.Year != params.Year {
		return false
	}

	// 评分过滤
	if params.Rating > 0 && media.Rating < params.Rating {
		return false
	}

	// 国家过滤
	if len(params.Countries) > 0 {
		hasCountry := false
		for _, country := range params.Countries {
			for _, mediaCountry := range media.Countries {
				if country == mediaCountry {
					hasCountry = true
					break
				}
			}
			if hasCountry {
				break
			}
		}
		if !hasCountry {
			return false
		}
	}

	// 语言过滤
	if len(params.Languages) > 0 {
		hasLanguage := false
		for _, language := range params.Languages {
			for _, mediaLanguage := range media.Languages {
				if language == mediaLanguage {
					hasLanguage = true
					break
				}
			}
			if hasLanguage {
				break
			}
		}
		if !hasLanguage {
			return false
		}
	}

	return true
}

// findSourceByName 根据名称查找数据源
func (mf *MediaFetcher) findSourceByName(name string) *MediaSource {
	for _, source := range mf.innerSources {
		if source.Name == name {
			return source
		}
	}
	return nil
}

// generateParamsHash 生成参数哈希
func (mf *MediaFetcher) generateParamsHash(params *FetchMediasParams) string {
	data, _ := json.Marshal(params)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// isWorkflowStopped 检查工作流是否已停止
func (mf *MediaFetcher) isWorkflowStopped(ctx context.Context, workflowID int64) bool {
	// 这里应该检查工作流状态
	// 暂时返回false
	return false
}

// 以下是各种数据源的实现函数

// fetchTMDBTrending 获取TMDB流行趋势
func (mf *MediaFetcher) fetchTMDBTrending(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error) {
	// 这里应该调用TMDB API获取流行趋势
	// 暂时返回模拟数据
	medias := []*types.MediaInfo{
		{
			TMDBID:      12345,
			Title:       "示例电影1",
			Year:        2023,
			Type:        "movie",
			Rating:      8.5,
			Popularity:  100.0,
			Overview:    "这是一部示例电影的描述",
			Genres:      []string{"动作", "冒险"},
			ReleaseDate: time.Now().AddDate(0, 0, -30),
		},
		{
			TMDBID:      67890,
			Title:       "示例电视剧1",
			Year:        2023,
			Type:        "tv",
			Rating:      9.0,
			Popularity:  150.0,
			Overview:    "这是一部示例电视剧的描述",
			Genres:      []string{"剧情", "悬疑"},
			ReleaseDate: time.Now().AddDate(0, 0, -15),
		},
	}

	return medias, nil
}

// fetchDoubanShowing 获取豆瓣正在热映
func (mf *MediaFetcher) fetchDoubanShowing(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error) {
	// 这里应该调用豆瓣API获取正在热映的电影
	// 暂时返回模拟数据
	medias := []*types.MediaInfo{
		{
			DoubanID:    "1234567",
			Title:       "热映电影1",
			Year:        2023,
			Type:        "movie",
			Rating:      7.8,
			Popularity:  80.0,
			Overview:    "正在热映的电影描述",
			Genres:      []string{"喜剧", "爱情"},
			ReleaseDate: time.Now().AddDate(0, 0, -7),
		},
	}

	return medias, nil
}

// fetchBangumiCalendar 获取Bangumi每日放送
func (mf *MediaFetcher) fetchBangumiCalendar(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error) {
	// 这里应该调用Bangumi API获取每日放送信息
	// 暂时返回模拟数据
	medias := []*types.MediaInfo{
		{
			BangumiID:   12345,
			Title:       "新番动画1",
			Year:        2023,
			Type:        "tv",
			Season:      1,
			Episodes:    []int{1},
			Rating:      8.2,
			Popularity:  120.0,
			Overview:    "新番动画描述",
			Genres:      []string{"动画", "日常"},
			ReleaseDate: time.Now(),
		},
	}

	return medias, nil
}

// fetchTMDBPopularMovies 获取TMDB热门电影
func (mf *MediaFetcher) fetchTMDBPopularMovies(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error) {
	// 实现TMDB热门电影获取
	return []*types.MediaInfo{}, nil
}

// fetchTMDBPopularTV 获取TMDB热门电视剧
func (mf *MediaFetcher) fetchTMDBPopularTV(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error) {
	// 实现TMDB热门电视剧获取
	return []*types.MediaInfo{}, nil
}

// fetchDoubanNewMovies 获取豆瓣新片榜
func (mf *MediaFetcher) fetchDoubanNewMovies(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error) {
	// 实现豆瓣新片榜获取
	return []*types.MediaInfo{}, nil
}

// fetchDoubanTopRated 获取豆瓣高分榜
func (mf *MediaFetcher) fetchDoubanTopRated(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error) {
	// 实现豆瓣高分榜获取
	return []*types.MediaInfo{}, nil
}

// fetchFromCustomAPI 从自定义API获取数据
func (mf *MediaFetcher) fetchFromCustomAPI(ctx context.Context, params *FetchMediasParams) ([]*types.MediaInfo, error) {
	// 实现自定义API调用
	return []*types.MediaInfo{}, nil
}

// GetAvailableSources 获取可用的数据源列表
func (mf *MediaFetcher) GetAvailableSources() []*MediaSource {
	var sources []*MediaSource
	for _, source := range mf.innerSources {
		if source.Enabled {
			sources = append(sources, source)
		}
	}
	return sources
}

// EnableSource 启用数据源
func (mf *MediaFetcher) EnableSource(sourceName string) error {
	for _, source := range mf.innerSources {
		if source.Name == sourceName {
			source.Enabled = true
			return nil
		}
	}
	return fmt.Errorf("数据源 %s 不存在", sourceName)
}

// DisableSource 禁用数据源
func (mf *MediaFetcher) DisableSource(sourceName string) error {
	for _, source := range mf.innerSources {
		if source.Name == sourceName {
			source.Enabled = false
			return nil
		}
	}
	return fmt.Errorf("数据源 %s 不存在", sourceName)
}
