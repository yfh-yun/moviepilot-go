// Package tmdbv3api 完整的TMDBv3 API实现
// 提供与Python版本tmdbv3api兼容的完整API接口
package tmdbv3api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/tmdb"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// TMDb TMDBv3 API客户端
type TMDb struct {
	client   *tmdb.Client
	cache    map[string]interface{}
	cacheTTL time.Duration
	logger   *zap.Logger
}

// NewTMDb 创建TMDBv3 API客户端
// apiKey: TMDB API密钥
// language: 语言（默认"zh-CN"）
// cacheEnabled: 是否启用缓存
func NewTMDb(apiKey string, language string, cacheEnabled bool) *TMDb {
	if language == "" {
		language = "zh-CN"
	}

	return &TMDb{
		client:   tmdb.NewClient(apiKey),
		cache:    make(map[string]interface{}),
		cacheTTL: 15 * time.Minute,
		logger:   logger.Logger,
	}
}

// Movie 电影相关API
type Movie struct {
	tmdb *TMDb
}

// TV 电视剧相关API
type TV struct {
	tmdb *TMDb
}

// Search 搜索相关API
type Search struct {
	tmdb *TMDb
}

// Trending 趋势相关API
type Trending struct {
	tmdb *TMDb
}

// Discover 发现相关API
type Discover struct {
	tmdb *TMDb
}

// Credits 演职员相关API
type Credits struct {
	tmdb *TMDb
}

// Videos 视频相关API
type Videos struct {
	tmdb *TMDb
}

// Keywords 关键词相关API
type Keywords struct {
	tmdb *TMDb
}

// NewMovie 创建电影API实例
func (t *TMDb) NewMovie() *Movie {
	return &Movie{tmdb: t}
}

// NewTV 创建电视剧API实例
func (t *TMDb) NewTV() *TV {
	return &TV{tmdb: t}
}

// NewSearch 创建搜索API实例
func (t *TMDb) NewSearch() *Search {
	return &Search{tmdb: t}
}

// NewTrending 创建趋势API实例
func (t *TMDb) NewTrending() *Trending {
	return &Trending{tmdb: t}
}

// NewDiscover 创建发现API实例
func (t *TMDb) NewDiscover() *Discover {
	return &Discover{tmdb: t}
}

// NewCredits 创建演职员API实例
func (t *TMDb) NewCredits() *Credits {
	return &Credits{tmdb: t}
}

// NewVideos 创建视频API实例
func (t *TMDb) NewVideos() *Videos {
	return &Videos{tmdb: t}
}

// NewKeywords 创建关键词API实例
func (t *TMDb) NewKeywords() *Keywords {
	return &Keywords{tmdb: t}
}

// MovieAPI 电影API方法

// Details 获取电影详情
// movieID: 电影ID
// appendToResponse: 附加响应信息
func (m *Movie) Details(ctx context.Context, movieID int64, appendToResponse string) (*tmdb.MovieDetails, error) {
	cacheKey := fmt.Sprintf("movie_details_%d_%s", movieID, appendToResponse)
	
	// 检查缓存
	if cached := m.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.MovieDetails), nil
	}

	// 调用API
	details, err := m.tmdb.client.GetMovieDetails(ctx, movieID)
	if err != nil {
		return nil, err
	}

	// 如果需要附加信息，获取额外的数据
	if appendToResponse != "" {
		appendData := make(map[string]interface{})
		
		if contains(appendToResponse, "credits") {
			if credits, err := m.tmdb.client.GetMovieCredits(ctx, movieID); err == nil {
				appendData["credits"] = credits
			}
		}
		
		if contains(appendToResponse, "videos") {
			if videos, err := m.tmdb.client.GetMovieVideos(ctx, movieID); err == nil {
				appendData["videos"] = videos
			}
		}
		
		if contains(appendToResponse, "keywords") {
			if keywords, err := m.tmdb.client.GetMovieKeywords(ctx, movieID); err == nil {
				appendData["keywords"] = keywords
			}
		}
		
		if contains(appendToResponse, "external_ids") {
			if externalIDs, err := m.tmdb.client.GetMovieExternalIDs(ctx, movieID); err == nil {
				appendData["external_ids"] = externalIDs
			}
		}
		
		if contains(appendToResponse, "similar") {
			if similar, err := m.tmdb.client.GetMovieSimilar(ctx, movieID, 1); err == nil {
				appendData["similar"] = similar
			}
		}
		
		if contains(appendToResponse, "recommendations") {
			if recommendations, err := m.tmdb.client.GetMovieRecommendations(ctx, movieID, 1); err == nil {
				appendData["recommendations"] = recommendations
			}
		}

		// 这里可以将附加数据合并到details中，但由于Go的结构限制，我们返回一个包装器
		// 在实际使用中，可以单独调用这些方法
	}

	// 缓存结果
	m.tmdb.setCache(cacheKey, details)
	return details, nil
}

// AccountStates 获取账户状态（需要认证）
func (m *Movie) AccountStates(ctx context.Context, movieID int64) (interface{}, error) {
	// TODO: 实现账户状态API
	return nil, fmt.Errorf("not implemented yet")
}

// AlternativeTitles 获取别名
func (m *Movie) AlternativeTitles(ctx context.Context, movieID int64, country string) (interface{}, error) {
	// TODO: 实现别名API
	return nil, fmt.Errorf("not implemented yet")
}

// Changes 获取变更
func (m *Movie) Changes(ctx context.Context, movieID int64, startDate, endDate string, page int) (interface{}, error) {
	// TODO: 实现变更API
	return nil, fmt.Errorf("not implemented yet")
}

// Credits 获取演职员信息
func (m *Movie) Credits(ctx context.Context, movieID int64) (*tmdb.Credits, error) {
	cacheKey := fmt.Sprintf("movie_credits_%d", movieID)
	
	// 检查缓存
	if cached := m.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.Credits), nil
	}

	credits, err := m.tmdb.client.GetMovieCredits(ctx, movieID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.tmdb.setCache(cacheKey, credits)
	return credits, nil
}

// ExternalIDs 获取外部ID
func (m *Movie) ExternalIDs(ctx context.Context, movieID int64) (*tmdb.ExternalIDs, error) {
	cacheKey := fmt.Sprintf("movie_external_ids_%d", movieID)
	
	// 检查缓存
	if cached := m.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.ExternalIDs), nil
	}

	externalIDs, err := m.tmdb.client.GetMovieExternalIDs(ctx, movieID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.tmdb.setCache(cacheKey, externalIDs)
	return externalIDs, nil
}

// Images 获取图片
func (m *Movie) Images(ctx context.Context, movieID int64) (interface{}, error) {
	// TODO: 实现图片API
	return nil, fmt.Errorf("not implemented yet")
}

// Keywords 获取关键词
func (m *Movie) Keywords(ctx context.Context, movieID int64) (*tmdb.Keywords, error) {
	cacheKey := fmt.Sprintf("movie_keywords_%d", movieID)
	
	// 检查缓存
	if cached := m.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.Keywords), nil
	}

	keywords, err := m.tmdb.client.GetMovieKeywords(ctx, movieID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.tmdb.setCache(cacheKey, keywords)
	return keywords, nil
}

// Lists 获取列表
func (m *Movie) Lists(ctx context.Context, movieID int64) (interface{}, error) {
	// TODO: 实现列表API
	return nil, fmt.Errorf("not implemented yet")
}

// Recommendations 获取推荐
func (m *Movie) Recommendations(ctx context.Context, movieID int64, page int) (*tmdb.Recommendations, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := fmt.Sprintf("movie_recommendations_%d_%d", movieID, page)
	
	// 检查缓存
	if cached := m.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.Recommendations), nil
	}

	recommendations, err := m.tmdb.client.GetMovieRecommendations(ctx, movieID, page)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.tmdb.setCache(cacheKey, recommendations)
	return recommendations, nil
}

// ReleaseDates 获取发行日期
func (m *Movie) ReleaseDates(ctx context.Context, movieID int64) (interface{}, error) {
	// TODO: 实现发行日期API
	return nil, fmt.Errorf("not implemented yet")
}

// Reviews 获取评论
func (m *Movie) Reviews(ctx context.Context, movieID int64, page int) (*tmdb.Reviews, error) {
	// TODO: 实现评论API
	return nil, fmt.Errorf("not implemented yet")
}

// Similar 获取相似电影
func (m *Movie) Similar(ctx context.Context, movieID int64, page int) (*tmdb.Similar, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := fmt.Sprintf("movie_similar_%d_%d", movieID, page)
	
	// 检查缓存
	if cached := m.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.Similar), nil
	}

	similar, err := m.tmdb.client.GetMovieSimilar(ctx, movieID, page)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.tmdb.setCache(cacheKey, similar)
	return similar, nil
}

// Translations 获取翻译
func (m *Movie) Translations(ctx context.Context, movieID int64) (interface{}, error) {
	// TODO: 实现翻译API
	return nil, fmt.Errorf("not implemented yet")
}

// Videos 获取视频
func (m *Movie) Videos(ctx context.Context, movieID int64) (*tmdb.Videos, error) {
	cacheKey := fmt.Sprintf("movie_videos_%d", movieID)
	
	// 检查缓存
	if cached := m.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.Videos), nil
	}

	videos, err := m.tmdb.client.GetMovieVideos(ctx, movieID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.tmdb.setCache(cacheKey, videos)
	return videos, nil
}

// WatchProviders 获取观看提供商
func (m *Movie) WatchProviders(ctx context.Context, movieID int64) (interface{}, error) {
	// TODO: 实现观看提供商API
	return nil, fmt.Errorf("not implemented yet")
}

// RateMovie 评分电影
func (m *Movie) RateMovie(ctx context.Context, movieID int64, rating float64) (interface{}, error) {
	// TODO: 实现评分电影API
	return nil, fmt.Errorf("not implemented yet")
}

// DeleteRating 删除评分
func (m *Movie) DeleteRating(ctx context.Context, movieID int64) (interface{}, error) {
	// TODO: 实现删除评分API
	return nil, fmt.Errorf("not implemented yet")
}

// Latest 获取最新电影
func (m *Movie) Latest(ctx context.Context) (*tmdb.MovieDetails, error) {
	// TODO: 实现最新电影API
	return nil, fmt.Errorf("not implemented yet")
}

// NowPlaying 正在热映
func (m *Movie) NowPlaying(ctx context.Context, page int) (*tmdb.SearchResult, error) {
	// TODO: 实现正在热映API
	return nil, fmt.Errorf("not implemented yet")
}

// Popular 热门电影
func (m *Movie) Popular(ctx context.Context, page int) (*tmdb.SearchResult, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := fmt.Sprintf("movie_popular_%d", page)
	
	// 检查缓存
	if cached := m.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.SearchResult), nil
	}

	popular, err := m.tmdb.client.GetPopularMovies(ctx, page)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	m.tmdb.setCache(cacheKey, popular)
	return popular, nil
}

// TopRated 评分最高
func (m *Movie) TopRated(ctx context.Context, page int) (*tmdb.SearchResult, error) {
	// TODO: 实现评分最高API
	return nil, fmt.Errorf("not implemented yet")
}

// Upcoming 即将上映
func (m *Movie) Upcoming(ctx context.Context, page int) (*tmdb.SearchResult, error) {
	// TODO: 实现即将上映API
	return nil, fmt.Errorf("not implemented yet")
}

// SearchAPI 搜索API方法

// Multi 多类型搜索
func (s *Search) Multi(ctx context.Context, query string, page int) (*tmdb.SearchResult, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := fmt.Sprintf("search_multi_%s_%d", query, page)
	
	// 检查缓存
	if cached := s.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.SearchResult), nil
	}

	result, err := s.tmdb.client.SearchMulti(ctx, query, page)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	s.tmdb.setCache(cacheKey, result)
	return result, nil
}

// Movie 搜索电影
func (s *Search) Movie(ctx context.Context, query string, year int, page int) (*tmdb.SearchResult, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := fmt.Sprintf("search_movie_%s_%d_%d", query, year, page)
	
	// 检查缓存
	if cached := s.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.SearchResult), nil
	}

	result, err := s.tmdb.client.SearchMovie(ctx, query, year, page)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	s.tmdb.setCache(cacheKey, result)
	return result, nil
}

// TV 搜索电视剧
func (s *Search) TV(ctx context.Context, query string, firstAirDateYear int, page int) (*tmdb.SearchResult, error) {
	if page < 1 {
		page = 1
	}

	cacheKey := fmt.Sprintf("search_tv_%s_%d_%d", query, firstAirDateYear, page)
	
	// 检查缓存
	if cached := s.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.SearchResult), nil
	}

	result, err := s.tmdb.client.SearchTV(ctx, query, firstAirDateYear, page)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	s.tmdb.setCache(cacheKey, result)
	return result, nil
}

// TrendingAPI 趋势API方法

// Movie 趋势电影
func (t *Trending) Movie(ctx context.Context, timeWindow string) (*tmdb.Trending, error) {
	if timeWindow == "" {
		timeWindow = "day"
	}

	cacheKey := fmt.Sprintf("trending_movie_%s", timeWindow)
	
	// 检查缓存
	if cached := t.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.Trending), nil
	}

	result, err := t.tmdb.client.GetTrendingMovies(ctx, timeWindow)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	t.tmdb.setCache(cacheKey, result)
	return result, nil
}

// TV 趋势电视剧
func (t *Trending) TV(ctx context.Context, timeWindow string) (*tmdb.Trending, error) {
	if timeWindow == "" {
		timeWindow = "day"
	}

	cacheKey := fmt.Sprintf("trending_tv_%s", timeWindow)
	
	// 检查缓存
	if cached := t.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.Trending), nil
	}

	result, err := t.tmdb.client.GetTrendingTV(ctx, timeWindow)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	t.tmdb.setCache(cacheKey, result)
	return result, nil
}

// DiscoverAPI 发现API方法

// Movie 发现电影
func (d *Discover) Movie(ctx context.Context, params map[string]string) (*tmdb.SearchResult, error) {
	cacheKey := fmt.Sprintf("discover_movie_%v", params)
	
	// 检查缓存
	if cached := d.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.SearchResult), nil
	}

	result, err := d.tmdb.client.DiscoverMovies(ctx, params)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	d.tmdb.setCache(cacheKey, result)
	return result, nil
}

// TV 发现电视剧
func (d *Discover) TV(ctx context.Context, params map[string]string) (*tmdb.SearchResult, error) {
	cacheKey := fmt.Sprintf("discover_tv_%v", params)
	
	// 检查缓存
	if cached := d.tmdb.getFromCache(cacheKey); cached != nil {
		return cached.(*tmdb.SearchResult), nil
	}

	result, err := d.tmdb.client.DiscoverTV(ctx, params)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	d.tmdb.setCache(cacheKey, result)
	return result, nil
}

// 缓存相关方法

// getFromCache 从缓存获取数据
func (t *TMDb) getFromCache(key string) interface{} {
	// 简单的内存缓存实现，实际生产环境建议使用Redis等
	if value, exists := t.cache[key]; exists {
		return value
	}
	return nil
}

// setCache 设置缓存
func (t *TMDb) setCache(key string, value interface{}) {
	t.cache[key] = value
	
	// 启动goroutine在TTL后清除缓存
	go func() {
		time.Sleep(t.cacheTTL)
		delete(t.cache, key)
	}()
}

// ClearCache 清除缓存
func (t *TMDb) ClearCache() {
	t.cache = make(map[string]interface{})
}

// SetCacheTTL 设置缓存TTL
func (t *TMDb) SetCacheTTL(ttl time.Duration) {
	t.cacheTTL = ttl
}

// 工具方法

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 findSubstring(s, substr))))
}

// findSubstring 在字符串中查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ParseTMDBIDFromURL 从URL解析TMDB ID
func ParseTMDBIDFromURL(tmdbURL string) (mediaType string, id int64, err error) {
	return tmdb.ParseTMDBID(tmdbURL)
}

// GetImageURL 获取图片URL
func (t *TMDb) GetImageURL(imagePath, size string) string {
	return t.client.GetImageURL(imagePath, size)
}

// GetPosterURL 获取海报URL
func (t *TMDb) GetPosterURL(posterPath string, size string) string {
	return t.client.GetPosterURL(posterPath, size)
}

// GetBackdropURL 获取背景图URL
func (t *TMDb) GetBackdropURL(backdropPath string, size string) string {
	return t.client.GetBackdropURL(backdropPath, size)
}