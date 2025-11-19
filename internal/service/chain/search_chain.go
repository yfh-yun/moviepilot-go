package chain

import (
	"fmt"
	"sort"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repository"
	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/internal/utils"
)

// SearchChain 搜索处理链
type SearchChain struct {
	logger         *utils.Logger
	searchRepo     *repository.SearchRepository
	torrentService *service.TorrentService
	siteService    *service.SiteService
	mediaService   *service.MediaService
	searchCache    *utils.Cache
}

// NewSearchChain 创建搜索业务链实例
func NewSearchChain(
	logger *utils.Logger,
	searchRepo *repository.SearchRepository,
	torrentService *service.TorrentService,
	siteService *service.SiteService,
	mediaService *service.MediaService,
) *SearchChain {
	return &SearchChain{
		logger:         logger,
		searchRepo:     searchRepo,
		torrentService: torrentService,
		siteService:    siteService,
		mediaService:   mediaService,
		searchCache:    utils.NewCache(30 * time.Minute), // 30分钟缓存
	}
}

// SearchTorrents 搜索种子
func (s *SearchChain) SearchTorrents(query string, mediaType string, siteNames []string, filters map[string]interface{}) (*models.SearchResult, error) {
	s.logger.Info("开始搜索种子", "query", query, "type", mediaType, "sites", siteNames)

	// 检查缓存
	cacheKey := s.generateCacheKey(query, mediaType, siteNames, filters)
	if cachedResult, found := s.searchCache.Get(cacheKey); found {
		s.logger.Debug("从缓存中获取搜索结果")
		return cachedResult.(*models.SearchResult), nil
	}

	// 获取要搜索的站点
	sites, err := s.getSearchSites(siteNames)
	if err != nil {
		return nil, fmt.Errorf("获取搜索站点失败: %v", err)
	}

	// 并行搜索所有站点
	results := make(chan *models.SiteSearchResult, len(sites))
	errors := make(chan error, len(sites))

	for _, site := range sites {
		go s.searchSite(site, query, mediaType, filters, results, errors)
	}

	// 收集结果
	var siteResults []*models.SiteSearchResult
	var searchErrors []error

	for i := 0; i < len(sites); i++ {
		select {
		case result := <-results:
			if result != nil {
				siteResults = append(siteResults, result)
			}
		case err := <-errors:
			searchErrors = append(searchErrors, err)
		}
	}

	// 合并结果
	mergedResults := s.mergeSearchResults(siteResults)

	// 排序结果
	sortedResults := s.sortSearchResults(mergedResults, filters)

	// 过滤结果
	filteredResults := s.filterSearchResults(sortedResults, filters)

	searchResult := &models.SearchResult{
		Query:      query,
		MediaType:  mediaType,
		Sites:      siteNames,
		TotalCount: len(filteredResults),
		Torrents:   filteredResults,
		Errors:     searchErrors,
		SearchTime: time.Now(),
	}

	// 保存到缓存
	s.searchCache.Set(cacheKey, searchResult, 30*time.Minute)

	// 保存搜索记录
	s.saveSearchHistory(query, mediaType, len(filteredResults), searchErrors)

	s.logger.Info("搜索完成", "query", query, "count", len(filteredResults), "errors", len(searchErrors))
	return searchResult, nil
}

// SearchByMedia 根据媒体信息搜索种子
func (s *SearchChain) SearchByMedia(mediaInfo *models.MediaInfo, siteNames []string, priority bool) (*models.SearchResult, error) {
	s.logger.Info("根据媒体信息搜索种子", "title", mediaInfo.Title, "year", mediaInfo.Year)

	// 构建搜索关键词
	searchQueries := s.buildSearchQueries(mediaInfo)

	var allResults []*models.TorrentInfo
	var searchErrors []error

	// 逐个搜索关键词
	for _, query := range searchQueries {
		result, err := s.SearchTorrents(query, mediaInfo.MediaType, siteNames, nil)
		if err != nil {
			searchErrors = append(searchErrors, err)
			continue
		}

		// 过滤结果，确保与媒体信息匹配
		filtered := s.filterTorrentsByMedia(result.Torrents, mediaInfo)
		allResults = append(allResults, filtered...)

		// 如果找到足够的结果，可以提前结束
		if len(allResults) >= 10 && priority {
			break
		}
	}

	// 去重
	deduplicated := s.deduplicateTorrents(allResults)

	// 排序
	sorted := s.sortTorrentsByQuality(deduplicated)

	searchResult := &models.SearchResult{
		Query:      mediaInfo.Title,
		MediaType:  mediaInfo.MediaType,
		Sites:      siteNames,
		TotalCount: len(sorted),
		Torrents:   sorted,
		Errors:     searchErrors,
		SearchTime: time.Now(),
	}

	s.logger.Info("媒体搜索完成", "title", mediaInfo.Title, "count", len(sorted))
	return searchResult, nil
}

// GetSearchHistory 获取搜索历史
func (s *SearchChain) GetSearchHistory(limit int) ([]*models.SearchHistory, error) {
	s.logger.Debug("获取搜索历史", "limit", limit)

	history, err := s.searchRepo.GetHistory(limit)
	if err != nil {
		s.logger.Error("获取搜索历史失败", "error", err)
		return nil, err
	}

	s.logger.Debug("搜索历史获取完成", "count", len(history))
	return history, nil
}

// ClearSearchHistory 清除搜索历史
func (s *SearchChain) ClearSearchHistory() error {
	s.logger.Info("清除搜索历史")

	err := s.searchRepo.ClearHistory()
	if err != nil {
		s.logger.Error("清除搜索历史失败", "error", err)
		return err
	}

	s.logger.Info("搜索历史已清除")
	return nil
}

// GetPopularSearches 获取热门搜索
func (s *SearchChain) GetPopularSearches(limit int) ([]*models.PopularSearch, error) {
	s.logger.Debug("获取热门搜索", "limit", limit)

	popular, err := s.searchRepo.GetPopularSearches(limit)
	if err != nil {
		s.logger.Error("获取热门搜索失败", "error", err)
		return nil, err
	}

	s.logger.Debug("热门搜索获取完成", "count", len(popular))
	return popular, nil
}

// AutoComplete 搜索自动补全
func (s *SearchChain) AutoComplete(query string, limit int) ([]string, error) {
	s.logger.Debug("搜索自动补全", "query", query, "limit", limit)

	// 从搜索历史中获取补全建议
	suggestions, err := s.searchRepo.GetAutoComplete(query, limit)
	if err != nil {
		s.logger.Warn("获取自动补全失败", "error", err)
		return nil, err
	}

	// 从媒体库中获取补全建议
	mediaSuggestions, err := s.mediaService.AutoComplete(query, limit)
	if err == nil {
		suggestions = append(suggestions, mediaSuggestions...)
	}

	// 去重和限制数量
	suggestions = s.deduplicateStrings(suggestions)
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	s.logger.Debug("自动补全完成", "count", len(suggestions))
	return suggestions, nil
}

// SearchSites 获取可搜索的站点列表
func (s *SearchChain) SearchSites() ([]*models.SearchSite, error) {
	s.logger.Debug("获取可搜索站点列表")

	sites, err := s.siteService.GetSearchableSites()
	if err != nil {
		s.logger.Error("获取搜索站点失败", "error", err)
		return nil, err
	}

	searchSites := make([]*models.SearchSite, 0, len(sites))
	for _, site := range sites {
		searchSite := &models.SearchSite{
			Name:    site.Name,
			Domain:  site.Domain,
			Enabled: site.Enabled,
			Type:    site.Type,
		}
		searchSites = append(searchSites, searchSite)
	}

	s.logger.Debug("搜索站点列表获取完成", "count", len(searchSites))
	return searchSites, nil
}

// 内部辅助方法

// generateCacheKey 生成缓存键
func (s *SearchChain) generateCacheKey(query string, mediaType string, siteNames []string, filters map[string]interface{}) string {
	key := fmt.Sprintf("search:%s:%s:%v", query, mediaType, siteNames)
	if filters != nil {
		key += fmt.Sprintf(":%v", filters)
	}
	return key
}

// getSearchSites 获取搜索站点
func (s *SearchChain) getSearchSites(siteNames []string) ([]*models.Site, error) {
	if len(siteNames) == 0 {
		// 获取所有启用的站点
		return s.siteService.GetEnabledSites()
	}

	var sites []*models.Site
	for _, siteName := range siteNames {
		site, err := s.siteService.GetSiteByName(siteName)
		if err != nil {
			s.logger.Warn("获取站点失败", "site", siteName, "error", err)
			continue
		}
		if site.Enabled {
			sites = append(sites, site)
		}
	}

	if len(sites) == 0 {
		return nil, fmt.Errorf("没有可用的搜索站点")
	}

	return sites, nil
}

// searchSite 在单个站点搜索
func (s *SearchChain) searchSite(site *models.Site, query string, mediaType string, filters map[string]interface{}, results chan<- *models.SiteSearchResult, errors chan<- error) {
	s.logger.Debug("在站点搜索", "site", site.Name, "query", query)

	torrents, err := s.torrentService.Search(site, query, mediaType, filters)
	if err != nil {
		s.logger.Warn("站点搜索失败", "site", site.Name, "error", err)
		errors <- fmt.Errorf("%s: %v", site.Name, err)
		return
	}

	result := &models.SiteSearchResult{
		SiteName: site.Name,
		Torrents: torrents,
		Count:    len(torrents),
	}

	results <- result
	s.logger.Debug("站点搜索完成", "site", site.Name, "count", len(torrents))
}

// mergeSearchResults 合并搜索结果
func (s *SearchChain) mergeSearchResults(siteResults []*models.SiteSearchResult) []*models.TorrentInfo {
	var allTorrents []*models.TorrentInfo

	for _, siteResult := range siteResults {
		allTorrents = append(allTorrents, siteResult.Torrents...)
	}

	return allTorrents
}

// sortSearchResults 排序搜索结果
func (s *SearchChain) sortSearchResults(torrents []*models.TorrentInfo, filters map[string]interface{}) []*models.TorrentInfo {
	sort.Slice(torrents, func(i, j int) bool {
		// 按种子数量排序
		if torrents[i].Seeders != torrents[j].Seeders {
			return torrents[i].Seeders > torrents[j].Seeders
		}

		// 按文件大小排序
		return torrents[i].Size > torrents[j].Size
	})

	return torrents
}

// filterSearchResults 过滤搜索结果
func (s *SearchChain) filterSearchResults(torrents []*models.TorrentInfo, filters map[string]interface{}) []*models.TorrentInfo {
	if filters == nil {
		return torrents
	}

	var filtered []*models.TorrentInfo

	for _, torrent := range torrents {
		if s.passFilter(torrent, filters) {
			filtered = append(filtered, torrent)
		}
	}

	return filtered
}

// passFilter 检查种子是否通过过滤条件
func (s *SearchChain) passFilter(torrent *models.TorrentInfo, filters map[string]interface{}) bool {
	// 这里可以实现各种过滤逻辑
	// 例如：文件大小、种子数量、发布时间等

	return true
}

// buildSearchQueries 构建搜索关键词
func (s *SearchChain) buildSearchQueries(mediaInfo *models.MediaInfo) []string {
	var queries []string

	// 基础搜索词
	queries = append(queries, mediaInfo.Title)

	// 带年份的搜索词
	if mediaInfo.Year > 0 {
		queries = append(queries, fmt.Sprintf("%s %d", mediaInfo.Title, mediaInfo.Year))
	}

	// 其他可能的搜索词变体
	// ...

	return queries
}

// filterTorrentsByMedia 根据媒体信息过滤种子
func (s *SearchChain) filterTorrentsByMedia(torrents []*models.TorrentInfo, mediaInfo *models.MediaInfo) []*models.TorrentInfo {
	var filtered []*models.TorrentInfo

	for _, torrent := range torrents {
		if s.torrentMatchesMedia(torrent, mediaInfo) {
			filtered = append(filtered, torrent)
		}
	}

	return filtered
}

// torrentMatchesMedia 检查种子是否匹配媒体信息
func (s *SearchChain) torrentMatchesMedia(torrent *models.TorrentInfo, mediaInfo *models.MediaInfo) bool {
	// 这里可以实现匹配逻辑
	// 例如：标题匹配、年份匹配等

	return true
}

// deduplicateTorrents 去重种子
func (s *SearchChain) deduplicateTorrents(torrents []*models.TorrentInfo) []*models.TorrentInfo {
	seen := make(map[string]bool)
	var unique []*models.TorrentInfo

	for _, torrent := range torrents {
		key := fmt.Sprintf("%s:%s", torrent.Title, torrent.Hash)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, torrent)
		}
	}

	return unique
}

// sortTorrentsByQuality 按质量排序种子
func (s *SearchChain) sortTorrentsByQuality(torrents []*models.TorrentInfo) []*models.TorrentInfo {
	sort.Slice(torrents, func(i, j int) bool {
		// 按种子质量排序（种子数量、文件大小等）
		return torrents[i].Seeders > torrents[j].Seeders
	})

	return torrents
}

// saveSearchHistory 保存搜索历史
func (s *SearchChain) saveSearchHistory(query string, mediaType string, resultCount int, errors []error) {
	history := &models.SearchHistory{
		Query:       query,
		MediaType:   mediaType,
		ResultCount: resultCount,
		ErrorCount:  len(errors),
		SearchTime:  time.Now(),
	}

	err := s.searchRepo.SaveHistory(history)
	if err != nil {
		s.logger.Warn("保存搜索历史失败", "error", err)
	}
}

// deduplicateStrings 去重字符串
func (s *SearchChain) deduplicateStrings(strs []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			unique = append(unique, str)
		}
	}

	return unique
}
