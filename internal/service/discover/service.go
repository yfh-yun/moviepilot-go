// Package discover Discover服务模块
package discover

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Service Discover服务接口
type Service interface {
	GetSources() ([]MediaSource, error)
	DiscoverBangumi(params DiscoverParams) ([]MediaInfo, error)
	DiscoverDoubanMovies(params DiscoverParams) ([]MediaInfo, error)
	DiscoverDoubanTVs(params DiscoverParams) ([]MediaInfo, error)
	DiscoverTMDbMovies(params TMDbDiscoverParams) ([]MediaInfo, error)
	DiscoverTMDbTVs(params TMDbDiscoverParams) ([]MediaInfo, error)
	GetTrendingMedia(params TrendingParams) ([]MediaInfo, error)
	GetPopularMedia(params PopularParams) ([]MediaInfo, error)
	GetNewMedia(params NewMediaParams) ([]MediaInfo, error)
	GetRandomMedia(params RandomMediaParams) ([]MediaInfo, error)
}

// DiscoverParams 探索参数
type DiscoverParams struct {
	Type     int    `json:"type"`     // 媒体类型: 1=电影, 2=剧集, 3=动画
	Category string `json:"category"` // 分类ID
	Sort     string `json:"sort"`     // 排序方式
	Year     string `json:"year"`     // 年份
	Page     int    `json:"page"`     // 页码
	Count    int    `json:"count"`    // 每页数量
	Tags     string `json:"tags"`     // 标签（豆瓣用）
}

// TMDbDiscoverParams TMDB探索参数
type TMDbDiscoverParams struct {
	SortBy               string  `json:"sort_by"`
	WithGenres           string  `json:"with_genres"`
	WithOriginalLanguage string  `json:"with_original_language"`
	WithKeywords         string  `json:"with_keywords"`
	WithWatchProviders   string  `json:"with_watch_providers"`
	VoteAverage          float64 `json:"vote_average"`
	VoteCount            int     `json:"vote_count"`
	ReleaseDate          string  `json:"release_date"`
	Page                 int     `json:"page"`
}

// TrendingParams 热门内容参数
type TrendingParams struct {
	Type       string `json:"type"`        // 媒体类型: all, movie, tv, anime
	Page       int    `json:"page"`        // 页码
	Count      int    `json:"count"`       // 每页数量
	TimeWindow string `json:"time_window"` // 时间窗口: day, week, month
}

// PopularParams 流行内容参数
type PopularParams struct {
	Type   string `json:"type"`   // 媒体类型: all, movie, tv
	Page   int    `json:"page"`   // 页码
	Count  int    `json:"count"`  // 每页数量
	Region string `json:"region"` // 地区代码
}

// NewMediaParams 最新内容参数
type NewMediaParams struct {
	Type               string `json:"type"`                 // 媒体类型: all, movie, tv
	Page               int    `json:"page"`                 // 页码
	Count              int    `json:"count"`                // 每页数量
	Region             string `json:"region"`               // 地区代码
	PrimaryReleaseYear string `json:"primary_release_year"` // 电影发行年份
	FirstAirDateYear   string `json:"first_air_date_year"`   // 电视剧首播年份
}

// RandomMediaParams 随机内容参数
type RandomMediaParams struct {
	Type      string  `json:"type"`       // 媒体类型: all, movie, tv
	Count     int     `json:"count"`      // 数量
	Genre     string  `json:"genre"`      // 类型ID
	MinRating float64 `json:"min_rating"` // 最低评分
	MinVotes  int     `json:"min_votes"`  // 最低评分人数
}

// MediaSource 媒体数据源
type MediaSource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	OriginalTitle string   `json:"original_title,omitempty"`
	Type          string   `json:"type"`
	Year          int      `json:"year,omitempty"`
	Rating        float64  `json:"rating,omitempty"`
	VoteCount     int      `json:"vote_count,omitempty"`
	Poster        string   `json:"poster,omitempty"`
	Backdrop      string   `json:"backdrop,omitempty"`
	Overview      string   `json:"overview,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Countries     []string `json:"countries,omitempty"`
	Languages     []string `json:"languages,omitempty"`
	ReleaseDate   string   `json:"release_date,omitempty"`
	Runtime       int      `json:"runtime,omitempty"`
	Popularity    float64  `json:"popularity,omitempty"`
	Source        string   `json:"source"`
	SourceURL     string   `json:"source_url,omitempty"`
}

// service Discover服务实现
type service struct {
	logger *zap.Logger
	cache  Cache
}

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// NewService 创建新的Discover服务
func NewService(
	logger *zap.Logger,
	cache Cache,
) Service {
	return &service{
		logger: logger,
		cache:  cache,
	}
}

// GetSources 获取可用的数据源列表
func (s *service) GetSources() ([]MediaSource, error) {
	sources := []MediaSource{
		{
			ID:      "bangumi",
			Name:    "Bangumi",
			Type:    "anime",
			Enabled: false, // TODO: 实现后启用
		},
		{
			ID:      "douban",
			Name:    "豆瓣",
			Type:    "movie_tv",
			Enabled: false, // TODO: 实现后启用
		},
		{
			ID:      "tmdb",
			Name:    "TMDB",
			Type:    "movie_tv",
			Enabled: false, // TODO: 实现后启用
		},
	}

	return sources, nil
}

// DiscoverBangumi 从Bangumi探索动漫内容
func (s *service) DiscoverBangumi(params DiscoverParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("discover:bangumi:%d:%s:%s:%s:%d:%d",
		params.Type, params.Category, params.Sort, params.Year, params.Page, params.Count)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for bangumi discover", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的Bangumi API调用
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:      1,
			Title:   "示例动漫 1",
			Type:    "anime",
			Year:    2023,
			Rating:  8.5,
			Source:  "bangumi",
		},
		{
			ID:      2,
			Title:   "示例动漫 2",
			Type:    "anime",
			Year:    2024,
			Rating:  9.0,
			Source:  "bangumi",
		},
	}

	// 缓存结果
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 1*time.Hour)
		}
	}

	return medias, nil
}

// DiscoverDoubanMovies 从豆瓣探索电影内容
func (s *service) DiscoverDoubanMovies(params DiscoverParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("discover:douban:movies:%s:%s:%d:%d",
		params.Sort, params.Tags, params.Page, params.Count)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for douban movies discover", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的豆瓣API调用
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:      3,
			Title:   "示例电影 1",
			Type:    "movie",
			Year:    2023,
			Rating:  8.2,
			Source:  "douban",
		},
		{
			ID:      4,
			Title:   "示例电影 2",
			Type:    "movie",
			Year:    2024,
			Rating:  7.8,
			Source:  "douban",
		},
	}

	// 缓存结果
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 30*time.Minute)
		}
	}

	return medias, nil
}

// DiscoverDoubanTVs 从豆瓣探索剧集内容
func (s *service) DiscoverDoubanTVs(params DiscoverParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("discover:douban:tvs:%s:%s:%d:%d",
		params.Sort, params.Tags, params.Page, params.Count)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for douban TVs discover", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的豆瓣API调用
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:      5,
			Title:   "示例剧集 1",
			Type:    "tv",
			Year:    2023,
			Rating:  8.7,
			Source:  "douban",
		},
		{
			ID:      6,
			Title:   "示例剧集 2",
			Type:    "tv",
			Year:    2024,
			Rating:  8.1,
			Source:  "douban",
		},
	}

	// 缓存结果
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 30*time.Minute)
		}
	}

	return medias, nil
}

// DiscoverTMDbMovies 从TMDB探索电影内容
func (s *service) DiscoverTMDbMovies(params TMDbDiscoverParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("discover:tmdb:movies:%s:%s:%s:%s:%s:%.1f:%d:%s:%d",
		params.SortBy, params.WithGenres, params.WithOriginalLanguage,
		params.WithKeywords, params.WithWatchProviders, params.VoteAverage,
		params.VoteCount, params.ReleaseDate, params.Page)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for TMDB movies discover", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的TMDB API调用
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:      7,
			Title:   "示例TMDB电影 1",
			Type:    "movie",
			Year:    2023,
			Rating:  8.3,
			Source:  "tmdb",
		},
		{
			ID:      8,
			Title:   "示例TMDB电影 2",
			Type:    "movie",
			Year:    2024,
			Rating:  7.9,
			Source:  "tmdb",
		},
	}

	// 缓存结果
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 30*time.Minute)
		}
	}

	return medias, nil
}

// DiscoverTMDbTVs 从TMDB探索剧集内容
func (s *service) DiscoverTMDbTVs(params TMDbDiscoverParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("discover:tmdb:tvs:%s:%s:%s:%s:%s:%.1f:%d:%s:%d",
		params.SortBy, params.WithGenres, params.WithOriginalLanguage,
		params.WithKeywords, params.WithWatchProviders, params.VoteAverage,
		params.VoteCount, params.ReleaseDate, params.Page)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for TMDB TVs discover", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的TMDB API调用
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:      9,
			Title:   "示例TMDB剧集 1",
			Type:    "tv",
			Year:    2023,
			Rating:  8.6,
			Source:  "tmdb",
		},
		{
			ID:      10,
			Title:   "示例TMDB剧集 2",
			Type:    "tv",
			Year:    2024,
			Rating:  8.0,
			Source:  "tmdb",
		},
	}

	// 缓存结果
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 30*time.Minute)
		}
	}

	return medias, nil
}

// helper function to convert string to int
func stringToInt(s string) int {
	if s == "" {
		return 0
	}
	i, _ := strconv.Atoi(s)
	return i
}

// helper function to convert string to float64
func stringToFloat(s string) float64 {
	if s == "" {
		return 0.0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// GetTrendingMedia 获取热门媒体内容
func (s *service) GetTrendingMedia(params TrendingParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("trending:%s:%s:%d:%d",
		params.Type, params.TimeWindow, params.Page, params.Count)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for trending media", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的热门内容获取逻辑
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:      1,
			Title:   "热门电影 1",
			Type:    "movie",
			Rating:  8.5,
			Source:  "tmdb",
		},
		{
			ID:      2,
			Title:   "热门剧集 1",
			Type:    "tv",
			Rating:  9.0,
			Source:  "tmdb",
		},
	}

	// 缓存结果
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 30*time.Minute)
		}
	}

	return medias, nil
}

// GetPopularMedia 获取流行媒体内容
func (s *service) GetPopularMedia(params PopularParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("popular:%s:%s:%d:%d",
		params.Type, params.Region, params.Page, params.Count)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for popular media", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的流行内容获取逻辑
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:      3,
			Title:   "流行电影 1",
			Type:    "movie",
			Rating:  8.2,
			Source:  "tmdb",
		},
		{
			ID:      4,
			Title:   "流行剧集 1",
			Type:    "tv",
			Rating:  8.8,
			Source:  "tmdb",
		},
	}

	// 缓存结果
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 1*time.Hour)
		}
	}

	return medias, nil
}

// GetNewMedia 获取最新媒体内容
func (s *service) GetNewMedia(params NewMediaParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("new:%s:%s:%s:%s:%s:%d:%d",
		params.Type, params.Region, params.PrimaryReleaseYear,
		params.FirstAirDateYear, params.Page, params.Count)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for new media", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的最新内容获取逻辑
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:          5,
			Title:       "最新电影 1",
			Type:        "movie",
			ReleaseDate: "2024-01-15",
			Rating:      7.8,
			Source:      "tmdb",
		},
		{
			ID:          6,
			Title:       "最新剧集 1",
			Type:        "tv",
			ReleaseDate: "2024-01-10",
			Rating:      8.1,
			Source:      "tmdb",
		},
	}

	// 缓存结果
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 45*time.Minute)
		}
	}

	return medias, nil
}

// GetRandomMedia 获取随机媒体内容
func (s *service) GetRandomMedia(params RandomMediaParams) ([]MediaInfo, error) {
	// 构建缓存键
	cacheKey := fmt.Sprintf("random:%s:%s:%.1f:%d:%d",
		params.Type, params.Genre, params.MinRating, params.MinVotes, params.Count)

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			var medias []MediaInfo
			if err := json.Unmarshal(cached, &medias); err == nil {
				s.logger.Debug("Cache hit for random media", zap.String("key", cacheKey))
				return medias, nil
			}
		}
	}

	// TODO: 实现实际的随机内容获取逻辑
	// 这里返回模拟数据
	medias := []MediaInfo{
		{
			ID:      7,
			Title:   "随机推荐 1",
			Type:    "movie",
			Rating:  7.5,
			Source:  "tmdb",
		},
		{
			ID:      8,
			Title:   "随机推荐 2",
			Type:    "tv",
			Rating:  8.3,
			Source:  "tmdb",
		},
	}

	// 缓存结果（随机内容缓存时间较短）
	if s.cache != nil {
		if data, err := json.Marshal(medias); err == nil {
			s.cache.Set(context.Background(), cacheKey, data, 15*time.Minute)
		}
	}

	return medias, nil
}

// helper function to format date
func tryParseDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	// 尝试解析常见日期格式
	formats := []string{"2006-01-02", "2006/01/02", "02-01-2006", "02/01/2006"}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return dateStr // 无法解析则返回原字符串
}
