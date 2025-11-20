package search

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/bangumi"
	"github.com/yfh-yun/moviepilot-go/internal/integration/douban"
	"github.com/yfh-yun/moviepilot-go/internal/integration/tmdb"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

type SearchService struct {
	tmdbClient        *tmdb.Client
	doubanClient      *douban.Client
	bangumiClient     *bangumi.Client
	siteRepo          interfaces.SiteRepository
	mediaRepo         interfaces.MediaRepository
	searchHistoryRepo interfaces.SearchHistoryRepository
	logger            *logger.Logger
}

func NewSearchService(
	tmdbClient *tmdb.Client,
	doubanClient *douban.Client,
	bangumiClient *bangumi.Client,
	siteRepo interfaces.SiteRepository,
	mediaRepo interfaces.MediaRepository,
	searchHistoryRepo interfaces.SearchHistoryRepository,
	logger *logger.Logger,
) *SearchService {
	return &SearchService{
		tmdbClient:        tmdbClient,
		doubanClient:      doubanClient,
		bangumiClient:     bangumiClient,
		siteRepo:          siteRepo,
		mediaRepo:         mediaRepo,
		searchHistoryRepo: searchHistoryRepo,
		logger:            logger,
	}
}

type SearchRequest struct {
	Query    string `json:"query"`
	Type     string `json:"type,omitempty"`
	Year     string `json:"year,omitempty"`
	Language string `json:"language,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Category string `json:"category,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`
	Order    string `json:"order,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	HasMore  bool           `json:"has_more"`
}

type SearchResult struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Type      string  `json:"type"`
	Year      string  `json:"year,omitempty"`
	Poster    string  `json:"poster,omitempty"`
	Rating    float64 `json:"rating,omitempty"`
	Overview  string  `json:"overview,omitempty"`
	Source    string  `json:"source"`
	Relevance float64 `json:"relevance"`
	Size      int64   `json:"size,omitempty"`
	Seeds     int     `json:"seeds,omitempty"`
	Peers     int     `json:"peers,omitempty"`
	Domain    string  `json:"domain,omitempty"`
	Status    string  `json:"status,omitempty"`
}

// MediaSearch 媒体搜索
func (s *SearchService) MediaSearch(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	s.logger.Info("开始媒体搜索", "query", req.Query, "type", req.Type)

	// 记录搜索历史
	if req.UserID != "" {
		s.recordSearchHistory(ctx, req.UserID, req.Query, "media")
	}

	var results []SearchResult

	// 并发搜索多个数据源
	resultChan := make(chan []SearchResult, 3)
	errorChan := make(chan error, 3)

	go func() {
		if tmdbResults, err := s.searchTMDB(ctx, req); err == nil {
			resultChan <- tmdbResults
		} else {
			errorChan <- err
		}
	}()

	go func() {
		if doubanResults, err := s.searchDouban(ctx, req); err == nil {
			resultChan <- doubanResults
		} else {
			errorChan <- err
		}
	}()

	go func() {
		if bangumiResults, err := s.searchBangumi(ctx, req); err == nil {
			resultChan <- bangumiResults
		} else {
			errorChan <- err
		}
	}()

	// 收集结果
	for i := 0; i < 3; i++ {
		select {
		case result := <-resultChan:
			results = append(results, result...)
		case err := <-errorChan:
			s.logger.Warn("搜索数据源失败", "error", err)
		case <-time.After(5 * time.Second):
			s.logger.Warn("搜索数据源超时")
		}
	}

	// 去重和排序
	results = s.deduplicateAndSort(results, req.Query)

	// 分页处理
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if end > len(results) {
		end = len(results)
	}

	if start >= len(results) {
		return &SearchResponse{
			Results:  []SearchResult{},
			Total:    len(results),
			Page:     req.Page,
			PageSize: req.PageSize,
			HasMore:  false,
		}, nil
	}

	return &SearchResponse{
		Results:  results[start:end],
		Total:    len(results),
		Page:     req.Page,
		PageSize: req.PageSize,
		HasMore:  end < len(results),
	}, nil
}

// TorrentSearch 种子搜索
func (s *SearchService) TorrentSearch(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	s.logger.Info("开始种子搜索", "query", req.Query, "category", req.Category)

	// 记录搜索历史
	if req.UserID != "" {
		s.recordSearchHistory(ctx, req.UserID, req.Query, "torrent")
	}

	// 获取站点列表
	sites, err := s.siteRepo.ListSites(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取站点列表失败: %w", err)
	}

	var results []SearchResult

	// 简化实现：这里应该通过RSS或其他方式搜索种子
	// 目前返回空结果，表示功能待实现
	s.logger.Warn("种子搜索功能待实现", "query", req.Query)

	return &SearchResponse{
		Results:  results,
		Total:    0,
		Page:     req.Page,
		PageSize: req.PageSize,
		HasMore:  false,
	}, nil
}

// SiteSearch 站点搜索
func (s *SearchService) SiteSearch(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	s.logger.Info("开始站点搜索", "query", req.Query)

	// 记录搜索历史
	if req.UserID != "" {
		s.recordSearchHistory(ctx, req.UserID, req.Query, "site")
	}

	// 获取所有站点
	sites, err := s.siteRepo.ListSites(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取站点列表失败: %w", err)
	}

	var results []SearchResult

	// 过滤站点
	for _, site := range sites {
		if s.matchSite(site, req.Query) {
			results = append(results, SearchResult{
				ID:        fmt.Sprintf("%d", site.ID),
				Title:     site.Name,
				Type:      "site",
				Domain:    site.Domain,
				Status:    site.Status,
				Source:    "internal",
				Relevance: s.calculateRelevance(site.Name, req.Query),
			})
		}
	}

	// 按相关性排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance > results[j].Relevance
	})

	// 分页处理
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if end > len(results) {
		end = len(results)
	}

	if start >= len(results) {
		return &SearchResponse{
			Results:  []SearchResult{},
			Total:    len(results),
			Page:     req.Page,
			PageSize: req.PageSize,
			HasMore:  false,
		}, nil
	}

	return &SearchResponse{
		Results:  results[start:end],
		Total:    len(results),
		Page:     req.Page,
		PageSize: req.PageSize,
		HasMore:  end < len(results),
	}, nil
}

// GetSearchHistory 获取搜索历史
func (s *SearchService) GetSearchHistory(ctx context.Context, userID string, page, pageSize int) (*models.PaginatedResponse, error) {
	history, total, err := s.searchHistoryRepo.GetUserSearchHistory(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取搜索历史失败: %w", err)
	}

	return &models.PaginatedResponse{
		Items:    history,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  (page * pageSize) < total,
	}, nil
}

// ClearSearchHistory 清空搜索历史
func (s *SearchService) ClearSearchHistory(ctx context.Context, userID string) error {
	return s.searchHistoryRepo.ClearUserSearchHistory(ctx, userID)
}

// 辅助方法

func (s *SearchService) searchTMDB(ctx context.Context, req SearchRequest) ([]SearchResult, error) {
	// 简化实现，实际应该调用TMDB API
	s.logger.Debug("TMDB搜索", "query", req.Query)
	return []SearchResult{}, nil
}

func (s *SearchService) searchDouban(ctx context.Context, req SearchRequest) ([]SearchResult, error) {
	// 简化实现，实际应该调用豆瓣API
	s.logger.Debug("豆瓣搜索", "query", req.Query)
	return []SearchResult{}, nil
}

func (s *SearchService) searchBangumi(ctx context.Context, req SearchRequest) ([]SearchResult, error) {
	// 简化实现，实际应该调用Bangumi API
	s.logger.Debug("Bangumi搜索", "query", req.Query)
	return []SearchResult{}, nil
}

func (s *SearchService) deduplicateAndSort(results []SearchResult, query string) []SearchResult {
	// 去重逻辑
	seen := make(map[string]bool)
	var uniqueResults []SearchResult

	for _, result := range results {
		key := result.ID + "_" + result.Source
		if !seen[key] {
			seen[key] = true
			uniqueResults = append(uniqueResults, result)
		}
	}

	// 按相关性排序
	sort.Slice(uniqueResults, func(i, j int) bool {
		return uniqueResults[i].Relevance > uniqueResults[j].Relevance
	})

	return uniqueResults
}

func (s *SearchService) matchSite(site models.Site, query string) bool {
	query = strings.ToLower(query)
	name := strings.ToLower(site.Name)
	domain := strings.ToLower(site.Domain)

	return strings.Contains(name, query) || strings.Contains(domain, query)
}

func (s *SearchService) calculateRelevance(title, query string) float64 {
	titleLower := strings.ToLower(title)
	queryLower := strings.ToLower(query)

	if titleLower == queryLower {
		return 1.0
	}
	if strings.Contains(titleLower, queryLower) {
		return 0.8
	}
	if strings.Contains(titleLower, string(queryLower[0])) {
		return 0.3
	}
	return 0.1
}

func (s *SearchService) recordSearchHistory(ctx context.Context, userID, query, searchType string) {
	history := &models.SearchHistory{
		UserID:     userID,
		Query:      query,
		SearchType: searchType,
		CreatedAt:  time.Now(),
	}

	if err := s.searchHistoryRepo.CreateSearchHistory(ctx, history); err != nil {
		s.logger.Warn("记录搜索历史失败", "error", err)
	}
}
