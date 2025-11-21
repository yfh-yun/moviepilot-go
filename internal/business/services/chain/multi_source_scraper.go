package chain

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"moviepilot-go/internal/models"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/utils"
	"go.uber.org/zap"
)

// MultiSourceScraper 多数据源刮削器
type MultiSourceScraper struct {
	logger    *zap.Logger
	sources   map[string]DataSource
	scheduler *ScraperScheduler
	cacher    *ScraperCache
	aggregator *ResultAggregator
}

// DataSource 数据源接口
type DataSource interface {
	// Name 数据源名称
	Name() string
	
	// Priority 数据源优先级（数值越小优先级越高）
	Priority() int
	
	// IsAvailable 检查数据源是否可用
	IsAvailable(ctx context.Context) bool
	
	// SearchByTitle 根据标题搜索
	SearchByTitle(ctx context.Context, req *SearchRequest) ([]*model.MediaInfo, error)
	
	// SearchByIMDBID 根据IMDB ID搜索
	SearchByIMDBID(ctx context.Context, imdbID string) (*model.MediaInfo, error)
	
	// SearchByTMDBID 根据TMDB ID搜索
	SearchByTMDBID(ctx context.Context, tmdbID int) (*model.MediaInfo, error)
	
	// GetSimilar 获取相似内容
	GetSimilar(ctx context.Context, mediaID string, limit int) ([]*model.MediaInfo, error)
	
	// GetRecommendations 获取推荐内容
	GetRecommendations(ctx context.Context, mediaID string, limit int) ([]*model.MediaInfo, error)
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Title       string `json:"title"`
	Year        *int   `json:"year"`
	Type        string `json:"type"`        // movie, tv
	Language    string `json:"language"`    // 语言偏好
	Country     string `json:"country"`     // 国家偏好
	Adult       bool   `json:"adult"`       // 是否包含成人内容
	MinRating   *float64 `json:"min_rating"` // 最低评分
	Limit       int    `json:"limit"`        // 结果数量限制
}

// ScrapeResult 刮削结果
type ScrapeResult struct {
	Source    string              `json:"source"`
	Data      *model.MediaInfo     `json:"data"`
	Confidence float64            `json:"confidence"`
	Metadata  map[string]interface{} `json:"metadata"`
	Duration  time.Duration       `json:"duration"`
	Error     error               `json:"error,omitempty"`
}

// ResultAggregator 结果聚合器
type ResultAggregator struct {
	logger      *zap.Logger
	weights     map[string]float64
	confidenceCalculator *ConfidenceCalculator
}

// ConfidenceCalculator 置信度计算器
type ConfidenceCalculator struct {
	titleMatcher    *utils.TitleMatcher
	yearComparator  *utils.YearComparator
	ratingAnalyzer *utils.RatingAnalyzer
}

// ScraperCache 刮削缓存
type ScraperCache struct {
	cache       map[string]*CacheEntry
	mutex       sync.RWMutex
	ttl         time.Duration
	maxSize     int
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Data      *model.MediaInfo
	ExpiresAt time.Time
	Metadata  map[string]interface{}
}

// ScraperScheduler 刮削调度器
type ScraperScheduler struct {
	workers   int
	timeout   time.Duration
	semaphore chan struct{}
}

// NewMultiSourceScraper 创建多数据源刮削器
func NewMultiSourceScraper(logger *zap.Logger) *MultiSourceScraper {
	scraper := &MultiSourceScraper{
		logger:     logger,
		sources:    make(map[string]DataSource),
		cacher:     NewScraperCache(time.Hour*2, 1000),
		aggregator: NewResultAggregator(logger),
	}
	
	scraper.scheduler = NewScraperScheduler(10, time.Second*30)
	
	return scraper
}

// RegisterSource 注册数据源
func (s *MultiSourceScraper) RegisterSource(source DataSource) {
	s.sources[source.Name()] = source
	s.logger.Info("注册数据源", 
		zap.String("name", source.Name()),
		zap.Int("priority", source.Priority()))
}

// ScrapeMedia 刮削媒体信息
func (s *MultiSourceScraper) ScrapeMedia(ctx context.Context, req *SearchRequest) (*model.MediaInfo, error) {
	s.logger.Info("开始多数据源刮削",
		zap.String("title", req.Title),
		zap.String("type", req.Type),
		zap.Strings("available_sources", s.getSourceNames()))

	// 1. 生成缓存键
	cacheKey := s.generateCacheKey(req)
	
	// 2. 尝试从缓存获取
	if cached := s.cacher.Get(cacheKey); cached != nil {
		s.logger.Debug("从缓存获取刮削结果", zap.String("cache_key", cacheKey))
		return cached, nil
	}

	// 3. 并发刮削多个数据源
	results := s.scrapeMultipleSources(ctx, req)
	
	// 4. 聚合结果
	mediaInfo, err := s.aggregator.AggregateResults(results, req)
	if err != nil {
		return nil, fmt.Errorf("聚合刮削结果失败: %w", err)
	}
	
	// 5. 缓存结果
	if mediaInfo != nil {
		s.cacher.Set(cacheKey, mediaInfo, map[string]interface{}{
			"sources_used": s.getUsedSources(results),
			"scraped_at":   time.Now(),
		})
	}

	s.logger.Info("多数据源刮削完成",
		zap.String("title", req.Title),
		zap.String("best_source", mediaInfo.Source),
		zap.Float64("confidence", mediaInfo.Confidence),
		zap.Int("sources_used", len(s.getUsedSources(results))))

	return mediaInfo, nil
}

// ScrapeByIMDBID 根据IMDB ID刮削
func (s *MultiSourceScraper) ScrapeByIMDBID(ctx context.Context, imdbID string) (*model.MediaInfo, error) {
	s.logger.Info("根据IMDB ID刮削", zap.String("imdb_id", imdbID))

	// 检查缓存
	cacheKey := fmt.Sprintf("imdb:%s", imdbID)
	if cached := s.cacher.Get(cacheKey); cached != nil {
		return cached, nil
	}

	var results []*ScrapeResult
	var mu sync.Mutex

	// 并发查询所有数据源
	var wg sync.WaitGroup
	for name, source := range s.sources {
		if !source.IsAvailable(ctx) {
			continue
		}

		wg.Add(1)
		go func(name string, source DataSource) {
			defer wg.Done()
			
			s.scheduler.Acquire()
			defer s.scheduler.Release()

			start := time.Now()
			mediaInfo, err := source.SearchByIMDBID(ctx, imdbID)
			duration := time.Since(start)

			mu.Lock()
			results = append(results, &ScrapeResult{
				Source:    name,
				Data:      mediaInfo,
				Duration:  duration,
				Error:     err,
			})
			mu.Unlock()

		}(name, source)
	}
	wg.Wait()

	// 聚合结果
	mediaInfo, err := s.aggregator.AggregateIMDBResults(results, imdbID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if mediaInfo != nil {
		s.cacher.Set(cacheKey, mediaInfo, map[string]interface{}{
			"sources_used": s.getUsedSources(results),
			"scraped_at":   time.Now(),
		})
	}

	return mediaInfo, nil
}

// ScrapeByTMDBID 根据TMDB ID刮削
func (s *MultiSourceScraper) ScrapeByTMDBID(ctx context.Context, tmdbID int) (*model.MediaInfo, error) {
	s.logger.Info("根据TMDB ID刮削", zap.Int("tmdb_id", tmdbID))

	// 检查缓存
	cacheKey := fmt.Sprintf("tmdb:%d", tmdbID)
	if cached := s.cacher.Get(cacheKey); cached != nil {
		return cached, nil
	}

	var results []*ScrapeResult
	var mu sync.Mutex

	// 并发查询所有数据源
	var wg sync.WaitGroup
	for name, source := range s.sources {
		if !source.IsAvailable(ctx) {
			continue
		}

		wg.Add(1)
		go func(name string, source DataSource) {
			defer wg.Done()
			
			s.scheduler.Acquire()
			defer s.scheduler.Release()

			start := time.Now()
			mediaInfo, err := source.SearchByTMDBID(ctx, tmdbID)
			duration := time.Since(start)

			mu.Lock()
			results = append(results, &ScrapeResult{
				Source:    name,
				Data:      mediaInfo,
				Duration:  duration,
				Error:     err,
			})
			mu.Unlock()

		}(name, source)
	}
	wg.Wait()

	// 聚合结果
	mediaInfo, err := s.aggregator.AggregateTMDBResults(results, tmdbID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if mediaInfo != nil {
		s.cacher.Set(cacheKey, mediaInfo, map[string]interface{}{
			"sources_used": s.getUsedSources(results),
			"scraped_at":   time.Now(),
		})
	}

	return mediaInfo, nil
}

// GetSimilar 获取相似内容
func (s *MultiSourceScraper) GetSimilar(ctx context.Context, mediaID string, limit int) ([]*model.MediaInfo, error) {
	s.logger.Info("获取相似内容", zap.String("media_id", mediaID))

	var allResults []*model.MediaInfo
	var mu sync.Mutex

	// 并发从多个数据源获取相似内容
	var wg sync.WaitGroup
	for name, source := range s.sources {
		if !source.IsAvailable(ctx) {
			continue
		}

		wg.Add(1)
		go func(name string, source DataSource) {
			defer wg.Done()
			
			results, err := source.GetSimilar(ctx, mediaID, limit)
			if err != nil {
				s.logger.Warn("获取相似内容失败",
					zap.String("source", name),
					zap.String("media_id", mediaID),
					zap.Error(err))
				return
			}

			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()

		}(name, source)
	}
	wg.Wait()

	// 去重和排序
	return s.deduplicateAndSort(allResults, limit), nil
}

// scrapeMultipleSources 从多个数据源刮削
func (s *MultiSourceScraper) scrapeMultipleSources(ctx context.Context, req *SearchRequest) []*ScrapeResult {
	var results []*ScrapeResult
	var mu sync.Mutex

	// 并发查询所有数据源
	var wg sync.WaitGroup
	for name, source := range s.sources {
		if !source.IsAvailable(ctx) {
			s.logger.Debug("数据源不可用，跳过", zap.String("source", name))
			continue
		}

		wg.Add(1)
		go func(name string, source DataSource) {
			defer wg.Done()
			
			s.scheduler.Acquire()
			defer s.scheduler.Release()

			start := time.Now()
			mediaInfos, err := source.SearchByTitle(ctx, req)
			duration := time.Since(start)

			if err != nil {
				s.logger.Warn("数据源刮削失败",
					zap.String("source", name),
					zap.String("title", req.Title),
					zap.Error(err))
				
				mu.Lock()
				results = append(results, &ScrapeResult{
					Source:   name,
					Data:     nil,
					Duration: duration,
					Error:    err,
				})
				mu.Unlock()
				return
			}

			// 为每个结果计算置信度
			for _, mediaInfo := range mediaInfos {
				confidence := s.calculateConfidence(mediaInfo, req)
				
				mu.Lock()
				results = append(results, &ScrapeResult{
					Source:     name,
					Data:       mediaInfo,
					Confidence: confidence,
					Duration:   duration,
				})
				mu.Unlock()
			}

		}(name, source)
	}
	wg.Wait()

	return results
}

// calculateConfidence 计算刮削结果的置信度
func (s *MultiSourceScraper) calculateConfidence(mediaInfo *model.MediaInfo, req *SearchRequest) float64 {
	var score float64 = 0.0

	// 标题匹配度
	titleScore := utils.CalculateTitleSimilarity(mediaInfo.Title, req.Title)
	score += titleScore * 0.4

	// 年份匹配度
	if req.Year != nil && mediaInfo.Year != nil {
		yearDiff := abs(*req.Year - *mediaInfo.Year)
		if yearDiff == 0 {
			score += 0.2
		} else if yearDiff <= 1 {
			score += 0.1
		}
	}

	// 数据源权重
	sourceWeight := s.aggregator.weights[mediaInfo.Source]
	score *= sourceWeight

	// 评分权重
	if mediaInfo.Vote != nil && *mediaInfo.Vote > 0 {
		ratingScore := min(*mediaInfo.Vote/10.0, 1.0)
		score += ratingScore * 0.1
	}

	return min(score, 1.0)
}

// generateCacheKey 生成缓存键
func (s *MultiSourceScraper) generateCacheKey(req *SearchRequest) string {
	parts := []string{req.Title, req.Type}
	if req.Year != nil {
		parts = append(parts, fmt.Sprintf("%d", *req.Year))
	}
	if req.Language != "" {
		parts = append(parts, req.Language)
	}
	return strings.Join(parts, ":")
}

// getSourceNames 获取数据源名称列表
func (s *MultiSourceScraper) getSourceNames() []string {
	var names []string
	for name := range s.sources {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getUsedSources 获取已使用的数据源
func (s *MultiSourceScraper) getUsedSources(results []*ScrapeResult) []string {
	used := make(map[string]bool)
	for _, result := range results {
		if result.Data != nil && result.Error == nil {
			used[result.Source] = true
		}
	}
	
	var names []string
	for name := range used {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// deduplicateAndSort 去重和排序
func (s *MultiSourceScraper) deduplicateAndSort(results []*model.MediaInfo, limit int) []*model.MediaInfo {
	seen := make(map[string]bool)
	var deduplicated []*model.MediaInfo

	for _, item := range results {
		key := fmt.Sprintf("%s:%s", item.Type, item.TMDBID)
		if !seen[key] {
			seen[key] = true
			deduplicated = append(deduplicated, item)
		}
	}

	// 按评分排序
	sort.Slice(deduplicated, func(i, j int) bool {
		if deduplicated[i].Vote == nil && deduplicated[j].Vote == nil {
			return false
		}
		if deduplicated[i].Vote == nil {
			return false
		}
		if deduplicated[j].Vote == nil {
			return true
		}
		return *deduplicated[i].Vote > *deduplicated[j].Vote
	})

	if limit > 0 && len(deduplicated) > limit {
		deduplicated = deduplicated[:limit]
	}

	return deduplicated
}

// 辅助函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(x, y float64) float64 {
	if x < y {
		return x
	}
	return y
}

// NewScraperScheduler 创建刮削调度器
func NewScraperScheduler(workers int, timeout time.Duration) *ScraperScheduler {
	return &ScraperScheduler{
		workers:   workers,
		timeout:   timeout,
		semaphore: make(chan struct{}, workers),
	}
}

// Acquire 获取工作槽位
func (s *ScraperScheduler) Acquire() {
	s.semaphore <- struct{}{}
}

// Release 释放工作槽位
func (s *ScraperScheduler) Release() {
	<-s.semaphore
}

// NewScraperCache 创建刮削缓存
func NewScraperCache(ttl time.Duration, maxSize int) *ScraperCache {
	return &ScraperCache{
		cache:   make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get 获取缓存
func (c *ScraperCache) Get(key string) *model.MediaInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil
	}
	return entry.Data
}

// Set 设置缓存
func (c *ScraperCache) Set(key string, data *model.MediaInfo, metadata map[string]interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查缓存大小限制
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
		Metadata:  metadata,
	}
}

// evictOldest 淘汰最旧的缓存
func (c *ScraperCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}