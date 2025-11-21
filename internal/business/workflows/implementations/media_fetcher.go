// Package implementations 提供动作系统的具体实现
package implementations

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/business/workflows/base"
	"moviepilot-go/internal/business/workflows/interfaces"
	"moviepilot-go/internal/business/workflows/types"
	"moviepilot-go/pkg/logger"
)

// MediaFetcher 媒体获取动作
// 对应Python版本的FetchMediasAction
type MediaFetcher struct {
	*base.Action
	config      *MediaFetcherConfig
	httpClient  *http.Client
	dataSources map[string]MediaDataSource
}

// MediaFetcherConfig 媒体获取器配置
type MediaFetcherConfig struct {
	Sources       []string          `json:"sources" description:"数据源列表"`
	DefaultLimit  int               `json:"default_limit" description:"默认获取数量"`
	Timeout       time.Duration     `json:"timeout" description:"请求超时时间"`
	EnableCache   bool              `json:"enable_cache" description:"启用缓存"`
	CacheTimeout  time.Duration     `json:"cache_timeout" description:"缓存超时时间"`
	UserAgent     string            `json:"user_agent" description:"用户代理"`
	RetryCount    int               `json:"retry_count" description:"重试次数"`
	FilterParams  map[string]string `json:"filter_params" description:"过滤参数"`
	CustomHeaders map[string]string `json:"custom_headers" description:"自定义请求头"`
}

// MediaDataSource 媒体数据源接口
type MediaDataSource interface {
	Name() string
	Fetch(ctx context.Context, params *FetchParams) ([]*types.MediaInfo, error)
	IsAvailable(ctx context.Context) bool
	GetPriority() int
}

// FetchParams 获取参数
type FetchParams struct {
	SourceType string                 `json:"source_type"`
	Sources    []string               `json:"sources"`
	Limit      int                    `json:"limit"`
	Genres     []string               `json:"genres"`
	Year       int                    `json:"year"`
	Rating     float64                `json:"rating"`
	Keywords   string                 `json:"keywords"`
	SortBy     string                 `json:"sort_by"`
	Extra      map[string]interface{} `json:"extra"`
}

// NewMediaFetcher 创建媒体获取器实例
func NewMediaFetcher() interfaces.Action {
	return &MediaFetcher{
		Action: base.NewAction("MediaFetcher", "媒体获取器，从各种数据源获取媒体信息"),
		config: &MediaFetcherConfig{
			Sources:      []string{"tmdb", "douban", "omdb"},
			DefaultLimit: 20,
			Timeout:      30 * time.Second,
			EnableCache:  true,
			CacheTimeout: 60 * time.Minute,
			UserAgent:    "MoviePilot-MediaFetcher/1.0",
			RetryCount:   3,
			FilterParams: map[string]string{
				"language": "zh-CN",
				"region":   "CN",
			},
			CustomHeaders: map[string]string{
				"Accept": "application/json",
			},
		},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		dataSources: make(map[string]MediaDataSource),
	}
}

// Execute 执行媒体获取
func (mf *MediaFetcher) Execute(ctx context.Context, workflowID int64, params map[string]interface{}, actionContext *types.ActionContext) (*types.ActionContext, error) {
	logger.Debug("MediaFetcher execution started", 
		"workflow_id", workflowID,
		"action", "MediaFetcher")

	// 检查缓存
	cacheKey := mf.generateCacheKey(params)
	if mf.CheckCache(ctx, workflowID, cacheKey) {
		logger.Info("Using cached media data", "workflow_id", workflowID)
		cachedData, err := mf.GetCache(ctx, workflowID, cacheKey)
		if err == nil {
			if medias, ok := cachedData.([]*types.MediaInfo); ok {
				mf.SetData("medias", medias)
				mf.SetData("count", len(medias))
				mf.SetDone(fmt.Sprintf("使用缓存数据: %d 个媒体", len(medias)))
				return actionContext, nil
			}
		}
	}

	// 解析参数
	fetchParams, err := mf.parseFetchParams(params)
	if err != nil {
		mf.SetError(fmt.Sprintf("参数解析失败: %v", err))
		return actionContext, err
	}

	// 执行获取
	medias, err := mf.fetchMedias(ctx, fetchParams)
	if err != nil {
		mf.SetError(fmt.Sprintf("媒体获取失败: %v", err))
		return actionContext, err
	}

	// 保存缓存
	if mf.config.EnableCache {
		if err := mf.SaveCache(ctx, workflowID, cacheKey, medias, mf.config.CacheTimeout); err != nil {
			logger.Warn("Failed to save media cache", "error", err)
		}
	}

	// 设置结果
	mf.SetData("medias", medias)
	mf.SetData("count", len(medias))
	mf.SetData("sources_used", fetchParams.Sources)
	mf.SetDone(fmt.Sprintf("成功获取 %d 个媒体", len(medias)))

	logger.Info("MediaFetcher execution completed", 
		"workflow_id", workflowID,
		"media_count", len(medias),
		"sources", fetchParams.Sources)

	return actionContext, nil
}

// parseFetchParams 解析获取参数
func (mf *MediaFetcher) parseFetchParams(params map[string]interface{}) (*FetchParams, error) {
	fetchParams := &FetchParams{
		Limit:   mf.config.DefaultLimit,
		Sources: mf.config.Sources,
	}

	if sourceType, ok := params["source_type"].(string); ok {
		fetchParams.SourceType = sourceType
	}

	if sources, ok := params["sources"].([]string); ok {
		fetchParams.Sources = sources
	}

	if limit, ok := params["limit"].(float64); ok {
		fetchParams.Limit = int(limit)
	}

	if genres, ok := params["genres"].([]string); ok {
		fetchParams.Genres = genres
	}

	if year, ok := params["year"].(float64); ok {
		fetchParams.Year = int(year)
	}

	if rating, ok := params["rating"].(float64); ok {
		fetchParams.Rating = rating
	}

	if keywords, ok := params["keywords"].(string); ok {
		fetchParams.Keywords = keywords
	}

	if sortBy, ok := params["sort_by"].(string); ok {
		fetchParams.SortBy = sortBy
	}

	if extra, ok := params["extra"].(map[string]interface{}); ok {
		fetchParams.Extra = extra
	}

	return fetchParams, nil
}

// fetchMedias 执行媒体获取
func (mf *MediaFetcher) fetchMedias(ctx context.Context, params *FetchParams) ([]*types.MediaInfo, error) {
	var allMedias []*types.MediaInfo
	var errors []error

	// 初始化数据源
	if err := mf.initializeDataSources(ctx); err != nil {
		return nil, fmt.Errorf("初始化数据源失败: %v", err)
	}

	// 按优先级排序数据源
	sortedSources := mf.getSortedSources(params.Sources)

	// 从每个数据源获取数据
	for _, sourceName := range sortedSources {
		source, exists := mf.dataSources[sourceName]
		if !exists {
			logger.Warn("Data source not found", "source", sourceName)
			continue
		}

		// 检查数据源可用性
		if !source.IsAvailable(ctx) {
			logger.Warn("Data source unavailable", "source", sourceName)
			continue
		}

		// 获取媒体数据
		medias, err := source.Fetch(ctx, params)
		if err != nil {
			logger.Error("Failed to fetch from source", "source", sourceName, "error", err)
			errors = append(errors, err)
			continue
		}

		// 过滤和合并结果
		filteredMedias := mf.filterMedias(medias, params)
		allMedias = append(allMedias, filteredMedias...)

		logger.Info("Fetched media from source", 
			"source", sourceName, 
			"count", len(filteredMedias))

		// 如果已达到限制数量，停止获取
		if len(allMedias) >= params.Limit {
			break
		}
	}

	// 去重和排序
	allMedias = mf.deduplicateMedias(allMedias)
	allMedias = mf.sortMedias(allMedias, params.SortBy)

	// 限制数量
	if len(allMedias) > params.Limit {
		allMedias = allMedias[:params.Limit]
	}

	if len(allMedias) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("所有数据源获取失败: %v", errors)
	}

	return allMedias, nil
}

// initializeDataSources 初始化数据源
func (mf *MediaFetcher) initializeDataSources(ctx context.Context) error {
	if len(mf.dataSources) > 0 {
		return nil // 已初始化
	}

	// 注册内置数据源
	mf.registerDataSource(NewTMDBDataSource(mf.httpClient))
	mf.registerDataSource(NewDoubanDataSource(mf.httpClient))
	mf.registerDataSource(NewOMDBDataSource(mf.httpClient))

	logger.Info("Data sources initialized", "count", len(mf.dataSources))
	return nil
}

// registerDataSource 注册数据源
func (mf *MediaFetcher) registerDataSource(source MediaDataSource) {
	mf.dataSources[source.Name()] = source
}

// getSortedSources 获取按优先级排序的数据源
func (mf *MediaFetcher) getSortedSources(requestedSources []string) []string {
	var sources []string
	seen := make(map[string]bool)

	// 添加请求的数据源
	for _, name := range requestedSources {
		if source, exists := mf.dataSources[name]; exists && !seen[name] {
			sources = append(sources, name)
			seen[name] = true
		}
	}

	// 添加其他可用数据源
	for name, source := range mf.dataSources {
		if !seen[name] && source.IsAvailable(context.Background()) {
			sources = append(sources, name)
			seen[name] = true
		}
	}

	return sources
}

// filterMedias 过滤媒体数据
func (mf *MediaFetcher) filterMedias(medias []*types.MediaInfo, params *FetchParams) []*types.MediaInfo {
	var filtered []*types.MediaInfo

	for _, media := range medias {
		// 类型过滤
		if len(params.Genres) > 0 {
			if !mf.containsGenre(media.Genres, params.Genres) {
				continue
			}
		}

		// 年份过滤
		if params.Year > 0 && media.Year != params.Year {
			continue
		}

		// 评分过滤
		if params.Rating > 0 && media.Rating < params.Rating {
			continue
		}

		// 关键词过滤
		if params.Keywords != "" && !mf.containsKeywords(media, params.Keywords) {
			continue
		}

		filtered = append(filtered, media)
	}

	return filtered
}

// containsGenre 检查是否包含指定类型
func (mf *MediaFetcher) containsGenre(mediaGenres []string, filterGenres []string) bool {
	for _, filter := range filterGenres {
		for _, genre := range mediaGenres {
			if strings.EqualFold(genre, filter) {
				return true
			}
		}
	}
	return false
}

// containsKeywords 检查是否包含关键词
func (mf *MediaFetcher) containsKeywords(media *types.MediaInfo, keywords string) bool {
	searchText := strings.ToLower(fmt.Sprintf("%s %s %s", 
		media.Title, media.OriginalTitle, strings.Join(media.Genres, " ")))
	
	keywordList := strings.Split(strings.ToLower(keywords), " ")
	for _, keyword := range keywordList {
		if !strings.Contains(searchText, keyword) {
			return false
		}
	}
	return true
}

// deduplicateMedias 去重媒体数据
func (mf *MediaFetcher) deduplicateMedias(medias []*types.MediaInfo) []*types.MediaInfo {
	seen := make(map[string]bool)
	var deduped []*types.MediaInfo

	for _, media := range medias {
		key := mf.generateMediaKey(media)
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, media)
		}
	}

	return deduped
}

// generateMediaKey 生成媒体唯一键
func (mf *MediaFetcher) generateMediaKey(media *types.MediaInfo) string {
	if media.IMDBID != "" {
		return "imdb:" + media.IMDBID
	}
	if media.TMDBID != 0 {
		return "tmdb:" + strconv.Itoa(media.TMDBID)
	}
	return fmt.Sprintf("%s:%s:%d", media.Title, media.Year, media.Type)
}

// sortMedias 排序媒体数据
func (mf *MediaFetcher) sortMedias(medias []*types.MediaInfo, sortBy string) []*types.MediaInfo {
	switch strings.ToLower(sortBy) {
	case "rating":
		// 按评分降序
		for i := 0; i < len(medias)-1; i++ {
			for j := i + 1; j < len(medias); j++ {
				if medias[i].Rating < medias[j].Rating {
					medias[i], medias[j] = medias[j], medias[i]
				}
			}
		}
	case "year":
		// 按年份降序
		for i := 0; i < len(medias)-1; i++ {
			for j := i + 1; j < len(medias); j++ {
				if medias[i].Year < medias[j].Year {
					medias[i], medias[j] = medias[j], medias[i]
				}
			}
		}
	case "title":
		// 按标题升序
		for i := 0; i < len(medias)-1; i++ {
			for j := i + 1; j < len(medias); j++ {
				if medias[i].Title > medias[j].Title {
					medias[i], medias[j] = medias[j], medias[i]
				}
			}
		}
	}

	return medias
}

// generateCacheKey 生成缓存键
func (mf *MediaFetcher) generateCacheKey(params map[string]interface{}) string {
	keyData, _ := json.Marshal(params)
	return fmt.Sprintf("media_fetch:%x", md5.Sum(keyData))
}

// Initialize 初始化媒体获取器
func (mf *MediaFetcher) Initialize() error {
	logger.Info("Initializing MediaFetcher", 
		"sources", mf.config.Sources,
		"default_limit", mf.config.DefaultLimit)
	return nil
}

// Cleanup 清理资源
func (mf *MediaFetcher) Cleanup() error {
	logger.Info("Cleaning up MediaFetcher")
	if mf.httpClient != nil {
		mf.httpClient.CloseIdleConnections()
	}
	return nil
}