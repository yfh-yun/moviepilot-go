package search

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/pkg/cache"
	"moviepilot-go/pkg/logger"
)

// SearchOptions 搜索选项
type SearchOptions struct {
	Category   string
	Page       int
	PageSize   int
	SortBy     string
	Order      string
	MinSeeders int
	MaxResults int
}

// Torrent 种子信息
type Torrent struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Seeders     int       `json:"seeders"`
	Leechers    int       `json:"leechers"`
	Size        int64     `json:"size"`
	PublishDate time.Time `json:"publish_date"`
	Site        string    `json:"site"`
	Category    string    `json:"category"`
}

// Service 搜索服务接口
type Service interface {
	// Search 搜索种子
	Search(ctx context.Context, req SearchRequest) (*SearchResponse, error)

	// SearchMultiSite 多站点搜索
	SearchMultiSite(ctx context.Context, sites []string, keyword string, opts SearchOptions) (*MultiSiteSearchResponse, error)

	// GetSearchHistory 获取搜索历史
	GetSearchHistory(ctx context.Context, userID uint, limit int) ([]*database.Search, error)

	// SaveSearchHistory 保存搜索历史
	SaveSearchHistory(ctx context.Context, history *database.Search) error

	// ClearSearchHistory 清除搜索历史
	ClearSearchHistory(ctx context.Context, userID uint) error

	// LastSearchResults 获取最近搜索结果
	LastSearchResults(ctx context.Context) ([]any, error)

	// SearchByID 根据媒体ID搜索
	SearchByID(ctx context.Context, mediaID string, mediaType string, area string, title string, year string, season string, sites string) ([]any, error)

	// SearchByTitle 根据标题搜索
	SearchByTitle(ctx context.Context, keyword string, page int, sites string) ([]any, error)
}

// service 搜索服务实现
type service struct {
	cache  cache.Cache
	logger *zap.Logger
}

// NewService 创建搜索服务
func NewService(cache cache.Cache) Service {
	return &service{
		cache:  cache,
		logger: logger.GetLogger(),
	}
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword    string   `json:"keyword" binding:"required"`
	Sites      []string `json:"sites"`
	Category   string   `json:"category"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	SortBy     string   `json:"sort_by"`
	Order      string   `json:"order"`
	MinSeeders int      `json:"min_seeders"`
	MaxResults int      `json:"max_results"`
	UseCache   bool     `json:"use_cache"`
	UserID     uint     `json:"user_id"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Torrents    []Torrent     `json:"torrents"`
	Total       int           `json:"total"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
	Duration    time.Duration `json:"duration"`
	FromCache   bool          `json:"from_cache"`
	SiteCount   int           `json:"site_count"`
	FailedSites []string      `json:"failed_sites,omitempty"`
}

// MultiSiteSearchResponse 多站点搜索响应
type MultiSiteSearchResponse struct {
	Results      []SearchResult `json:"results"`
	Total        int            `json:"total"`
	Duration     time.Duration  `json:"duration"`
	SuccessSites int            `json:"success_sites"`
	FailedSites  []string       `json:"failed_sites"`
}

// Search 搜索种子
func (s *service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	s.logger.Info("搜索种子",
		zap.String("keyword", req.Keyword),
		zap.Strings("sites", req.Sites))

	start := time.Now()

	// 检查缓存
	if req.UseCache {
		cacheKey := s.buildCacheKey(req)
		var cached SearchResponse
		if err := s.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			s.logger.Debug("从缓存获取搜索结果", zap.String("keyword", req.Keyword))
			cached.FromCache = true
			return &cached, nil
		}
	}

	// 执行搜索 - 简化实现，返回空结果
	// TODO: 实现实际的搜索逻辑
	_ = req.Category
	_ = req.SortBy
	_ = req.Order
	_ = req.MinSeeders
	_ = req.MaxResults

	// 合并结果 - 简化实现
	torrents := []Torrent{}

	// 统计失败的站点
	var failedSites []string
	successCount := 0

	response := &SearchResponse{
		Torrents:    torrents,
		Total:       len(torrents),
		Page:        req.Page,
		PageSize:    req.PageSize,
		Duration:    time.Since(start),
		FromCache:   false,
		SiteCount:   successCount,
		FailedSites: failedSites,
	}

	// 保存到缓存
	if req.UseCache && len(torrents) > 0 {
		cacheKey := s.buildCacheKey(req)
		if err := s.cache.SetJSON(ctx, cacheKey, response, 5*time.Minute); err != nil {
			s.logger.Warn("保存搜索结果到缓存失败", zap.Error(err))
		}
	}

	// TODO: 保存搜索历史
	// 需要根据实际的 Search 模型字段调整
	_ = req.UserID

	s.logger.Info("搜索完成",
		zap.String("keyword", req.Keyword),
		zap.Int("total", len(torrents)),
		zap.Duration("duration", response.Duration))

	return response, nil
}

// SearchMultiSite 多站点搜索
func (s *service) SearchMultiSite(ctx context.Context, sites []string, keyword string, opts SearchOptions) (*MultiSiteSearchResponse, error) {
	s.logger.Info("多站点搜索",
		zap.String("keyword", keyword),
		zap.Int("sites", len(sites)))

	start := time.Now()

	// 执行搜索 - 简化实现，返回空结果
	// TODO: 实现实际的多站点搜索逻辑
	results := []SearchResult{}

	// 统计
	total := 0
	successCount := 0
	var failedSites []string

	// 简化实现，直接返回空结果
	total = 0
	successCount = 0

	response := &MultiSiteSearchResponse{
		Results:      results,
		Total:        total,
		Duration:     time.Since(start),
		SuccessSites: successCount,
		FailedSites:  failedSites,
	}

	s.logger.Info("多站点搜索完成",
		zap.String("keyword", keyword),
		zap.Int("total", total),
		zap.Int("success", successCount),
		zap.Duration("duration", response.Duration))

	return response, nil
}

// GetSearchHistory 获取搜索历史
func (s *service) GetSearchHistory(ctx context.Context, userID uint, limit int) ([]*database.Search, error) {
	// TODO: 实现从数据库获取搜索历史
	return nil, nil
}

// SaveSearchHistory 保存搜索历史
func (s *service) SaveSearchHistory(ctx context.Context, history *database.Search) error {
	// TODO: 实现保存搜索历史到数据库
	return nil
}

// ClearSearchHistory 清除搜索历史
func (s *service) ClearSearchHistory(ctx context.Context, userID uint) error {
	// TODO: 实现清除搜索历史
	return nil
}

// LastSearchResults 获取最近搜索结果
func (s *service) LastSearchResults(ctx context.Context) ([]any, error) {
	s.logger.Info("获取最近搜索结果")

	// TODO: 实现获取最近搜索结果逻辑
	// 1. 从缓存或数据库获取最近搜索结果
	// 2. 转换为需要的格式
	// 3. 返回结果

	// 目前返回空数组
	return []any{}, nil
}

// SearchByID 根据媒体ID搜索
func (s *service) SearchByID(ctx context.Context, mediaID string, mediaType string, area string, title string, year string, season string, sites string) ([]any, error) {
	s.logger.Info("根据媒体ID搜索",
		zap.String("mediaID", mediaID),
		zap.String("mediaType", mediaType),
		zap.String("area", area),
		zap.String("title", title),
		zap.String("year", year),
		zap.String("season", season),
		zap.String("sites", sites),
	)

	// TODO: 实现根据媒体ID搜索逻辑
	// 1. 根据mediaID前缀识别媒体类型（tmdb:/douban:/bangumi:）
	// 2. 根据配置的RECOGNIZE_SOURCE进行转换
	// 3. 调用search_chain.async_search_by_id进行搜索
	// 4. 转换结果格式
	// 5. 返回结果

	// 目前返回空数组
	return []any{}, nil
}

// SearchByTitle 根据标题搜索
func (s *service) SearchByTitle(ctx context.Context, keyword string, page int, sites string) ([]any, error) {
	s.logger.Info("根据标题搜索",
		zap.String("keyword", keyword),
		zap.Int("page", page),
		zap.String("sites", sites),
	)

	// TODO: 实现根据标题搜索逻辑
	// 1. 解析sites参数
	// 2. 调用search_chain.async_search_by_title进行搜索
	// 3. 转换结果格式
	// 4. 返回结果

	// 目前返回空数组
	return []any{}, nil
}

// buildCacheKey 构建缓存键
func (s *service) buildCacheKey(req SearchRequest) string {
	sites := strings.Join(req.Sites, ",")
	if sites == "" {
		sites = "all"
	}
	return fmt.Sprintf("search:%s:%s:%s:%d:%d",
		req.Keyword, sites, req.Category, req.Page, req.PageSize)
}

// filterByMinSeeders 按最小做种数过滤
func (s *service) filterByMinSeeders(torrents []Torrent, minSeeders int) []Torrent {
	var filtered []Torrent
	for _, t := range torrents {
		if t.Seeders >= minSeeders {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// sortTorrents 排序种子
func (s *service) sortTorrents(torrents []Torrent, sortBy string, order string) []Torrent {
	sort.Slice(torrents, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "seeders":
			less = torrents[i].Seeders < torrents[j].Seeders
		case "leechers":
			less = torrents[i].Leechers < torrents[j].Leechers
		case "size":
			less = torrents[i].Size < torrents[j].Size
		case "date":
			less = torrents[i].PublishDate.Before(torrents[j].PublishDate)
		default:
			less = torrents[i].Seeders < torrents[j].Seeders
		}

		if order == "asc" {
			return less
		}
		return !less
	})

	return torrents
}

// MergeAndDeduplicate 合并并去重
func (s *service) MergeAndDeduplicate(results []SearchResult) []Torrent {
	// 简化实现，直接返回空列表
	return []Torrent{}
}

// EnrichResults 丰富搜索结果
func (s *service) EnrichResults(ctx context.Context, torrents []Torrent) []EnrichedTorrent {
	enriched := make([]EnrichedTorrent, len(torrents))

	for i, torrent := range torrents {
		enriched[i] = EnrichedTorrent{
			Torrent: torrent,
			Score:   s.calculateScore(torrent),
		}
	}

	return enriched
}

// EnrichedTorrent 丰富的种子信息
type EnrichedTorrent struct {
	Torrent
	Score float64 `json:"score"`
}

// calculateScore 计算种子评分
func (s *service) calculateScore(torrent Torrent) float64 {
	// 简单的评分算法
	// 做种数权重: 0.5
	// 下载数权重: 0.3
	// 大小权重: 0.2

	seedersScore := float64(torrent.Seeders) * 0.5
	leechersScore := float64(torrent.Leechers) * 0.3
	sizeScore := float64(torrent.Size) / (1024 * 1024 * 1024) * 0.2 // GB

	return seedersScore + leechersScore + sizeScore
}
