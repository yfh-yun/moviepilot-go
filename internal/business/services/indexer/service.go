// Package indexer 索引器服务包
package indexer

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/integration/indexer/parsers"
	"github.com/yfh-yun/moviepilot-go/internal/integration/indexer/spiders"
)

// IndexerService 索引器服务
type IndexerService struct {
	spiderManager *spider.SpiderManager
	parserManager *parser.ParserManager
	mutex         sync.RWMutex
}

// NewIndexerService 创建索引器服务
func NewIndexerService() *IndexerService {
	return &IndexerService{
		spiderManager: spider.NewSpiderManager(),
		parserManager: parser.NewParserManager(),
	}
}

// Initialize 初始化服务
func (s *IndexerService) Initialize() error {
	logger.Info("初始化索引器服务")

	// 注册内置解析器
	if err := s.registerBuiltinParsers(); err != nil {
		return fmt.Errorf("注册内置解析器失败: %w", err)
	}

	logger.Info("索引器服务初始化完成")
	return nil
}

// registerBuiltinParsers 注册内置解析器
func (s *IndexerService) registerBuiltinParsers() error {
	// 注册NexusPHP解析器
	nexusPHPParser := parser.NewNexusPHPParser()
	if err := s.parserManager.RegisterParser("nexusphp", nexusPHPParser); err != nil {
		return fmt.Errorf("注册NexusPHP解析器失败: %w", err)
	}

	// TODO: 注册其他解析器
	// gazelleParser := parser.NewGazelleParser()
	// s.parserManager.RegisterParser("gazelle", gazelleParser)

	logger.Info("内置解析器注册完成", zap.Int("count", 1))
	return nil
}

// AddSite 添加站点
func (s *IndexerService) AddSite(siteID string, config *spider.SiteConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger.Info("添加站点",
		zap.String("site_id", siteID),
		zap.String("name", config.Name))

	// 添加到Spider管理器
	if err := s.spiderManager.AddSite(siteID, config); err != nil {
		return fmt.Errorf("添加到Spider管理器失败: %w", err)
	}

	// 创建对应的解析器配置
	parserConfig := &parser.SiteConfig{
		Name:    config.Name,
		Domain:  config.URL,
		Charset: "UTF-8",
		Enabled: config.Enabled,
	}

	// 注册解析器
	switch config.Type {
	case "nexusphp":
		nexusPHPParser := parser.NewNexusPHPParser()
		if err := s.parserManager.RegisterParser(siteID, nexusPHPParser); err != nil {
			return fmt.Errorf("注册NexusPHP解析器失败: %w", err)
		}
	default:
		// 对于其他类型，尝试通用解析器
		if err := s.parserManager.RegisterParser(siteID, parser.NewBaseParser(config.Name, config.URL)); err != nil {
			return fmt.Errorf("注册解析器失败: %w", err)
		}
	}

	// 更新解析器配置
	if err := s.parserManager.UpdateConfig(siteID, parserConfig); err != nil {
		logger.Warn("更新解析器配置失败", zap.Error(err))
	}

	return nil
}

// RemoveSite 移除站点
func (s *IndexerService) RemoveSite(siteID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger.Info("移除站点", zap.String("site_id", siteID))

	// 从Spider管理器移除
	if err := s.spiderManager.RemoveSite(siteID); err != nil {
		return fmt.Errorf("从Spider管理器移除失败: %w", err)
	}

	return nil
}

// SearchTorrents 搜索种子
func (s *IndexerService) SearchTorrents(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	logger.Info("搜索种子",
		zap.String("keyword", req.Keyword),
		zap.Strings("sites", req.SiteIDs),
		zap.Int("limit", req.Limit),
		zap.Int("page", req.Page))

	response := &SearchResponse{
		SiteResults: make(map[string]*SearchSiteResult),
		Metadata:     &SearchMetadata{},
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	// 如果没有指定站点，搜索所有启用的站点
	if len(req.SiteIDs) == 0 {
		sites := s.spiderManager.ListSites()
		for siteID, config := range sites {
			if config.Enabled {
				req.SiteIDs = append(req.SiteIDs, siteID)
			}
		}
	}

	// 并发搜索每个站点
	for _, siteID := range req.SiteIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			siteResult, err := s.searchSite(ctx, id, req)
			if err != nil {
				logger.Warn("站点搜索失败",
					zap.String("site_id", id),
					zap.Error(err))
				return
			}

			mutex.Lock()
			response.SiteResults[id] = siteResult
			response.TotalResults += len(siteResult.Torrents)
			mutex.Unlock()

			logger.Debug("站点搜索完成",
				zap.String("site_id", id),
				zap.Int("count", len(siteResult.Torrents)))
		}(siteID)
	}

	wg.Wait()

	logger.Info("搜索完成",
		zap.String("keyword", req.Keyword),
		zap.Int("total_results", response.TotalResults),
		zap.Int("successful_sites", len(response.SiteResults)))

	return response, nil
}

// searchSite 在单个站点搜索
func (s *IndexerService) searchSite(ctx context.Context, siteID string, req *SearchRequest) (*SearchSiteResult, error) {
	// 构建搜索过滤器
	filters := &spider.SearchFilters{
		Keyword:    req.Keyword,
		Page:       req.Page,
		Limit:      req.Limit,
		MinSeeders: req.MinSeeders,
		FreeLeech:  req.FreeLeech,
		HDR:        req.HDR,
	}

	// 使用Spider搜索
	spiders, err := s.spiderManager.Search(ctx, req.Keyword, filters)
	if err != nil {
		return nil, fmt.Errorf("Spider搜索失败: %w", err)
	}

	// 合并搜索结果
	var allTorrents []*spider.TorrentItem
	for _, items := range spiders {
		allTorrents = append(allTorrents, items...)
	}

	// 应用站点过滤器
	filteredTorrents := s.applyTorrentFilters(allTorrents, req)

	// 转换为搜索结果格式
	torrentInfos := make([]*TorrentInfo, 0, len(filteredTorrents))
	for _, torrent := range filteredTorrents {
		torrentInfo := s.convertTorrentItem(torrent)
		torrentInfos = append(torrentInfos, torrentInfo)
	}

	return &SearchSiteResult{
		SiteID:     siteID,
		Torrents:    torrentInfos,
		Count:       len(torrentInfos),
		HasMore:     len(torrentInfos) >= req.Limit,
		SearchTime:   "0s", // TODO: 计算实际搜索时间
	}, nil
}

// applyTorrentFilters 应用种子过滤器
func (s *IndexerService) applyTorrentFilters(torrents []*spider.TorrentItem, req *SearchRequest) []*spider.TorrentItem {
	var filtered []*spider.TorrentItem

	for _, torrent := range torrents {
		// 应用过滤器
		if req.MinSeeders > 0 && torrent.Seeders < req.MinSeeders {
			continue
		}

		if req.FreeLeech && !torrent.FreeLeech {
			continue
		}

		if req.HDR && !torrent.HDR {
			continue
		}

		// 关键词过滤
		if !s.matchKeyword(torrent.Title, req.Keyword) {
			continue
		}

		filtered = append(filtered, torrent)
	}

	return filtered
}

// matchKeyword 检查关键词匹配
func (s *IndexerService) matchKeyword(title, keyword string) bool {
	if keyword == "" {
		return true
	}

	// 简单的关键词匹配（实际实现可能需要更复杂的逻辑）
	return len(keyword) > 0 && len(title) > 0
}

// convertTorrentItem 转换种子项目
func (s *IndexerService) convertTorrentItem(item *spider.TorrentItem) *TorrentInfo {
	return &TorrentInfo{
		ID:              item.ID,
		Title:           item.Title,
		Category:        item.Category,
		SubCategory:     item.SubCategory,
		Size:            item.Size,
		Seeders:         item.Seeders,
		Leechers:        item.Leechers,
		Downloads:       item.Downloads,
		UploadedAt:      item.UploadedAt.Format("2006-01-02 15:04:05"),
		Promotional:     item.Promotional,
		FreeLeech:       item.FreeLeech,
		HDR:             item.HDR,
		Dubbed:          item.Dubbed,
		Subtitled:       item.Subtitled,
		UploadFactor:    item.UploadFactor,
		DownloadFactor:  item.DownloadFactor,
		Comments:        item.Comments,
		IMDBID:          item.IMDBID,
		TMDBID:          fmt.Sprintf("%d", item.TMDBID),
		ReleaseGroup:    item.ReleaseGroup,
		Meta:            item.Meta,
	}
}

// GetTorrentDetails 获取种子详情
func (s *IndexerService) GetTorrentDetails(ctx context.Context, siteID, torrentID string) (*TorrentDetail, error) {
	logger.Info("获取种子详情",
		zap.String("site_id", siteID),
		zap.String("torrent_id", torrentID))

	// 使用Spider获取详情
	spider, err := s.spiderManager.GetSpider(siteID)
	if err != nil {
		return nil, fmt.Errorf("获取Spider失败: %w", err)
	}

	detail, err := spider.GetTorrentDetails(ctx, torrentID)
	if err != nil {
		return nil, fmt.Errorf("获取种子详情失败: %w", err)
	}

	// 转换为统一的详情格式
	torrentDetail := s.convertTorrentDetail(detail)

	logger.Info("种子详情获取成功",
		zap.String("site_id", siteID),
		zap.String("torrent_id", torrentID),
		zap.String("title", torrentDetail.Title))

	return torrentDetail, nil
}

// convertTorrentDetail 转换种子详情
func (s *IndexerService) convertTorrentDetail(detail *spider.TorrentDetail) *TorrentDetail {
	torrentDetail := &TorrentDetail{
		ID:              detail.ID,
		Title:           detail.Title,
		Description:     detail.Description,
		Category:        detail.Category,
		SubCategory:     detail.SubCategory,
		Size:            detail.Size,
		Seeders:         detail.Seeders,
		Leechers:        detail.Leechers,
		Downloads:       detail.Downloads,
		UploadedAt:      detail.UploadedAt.Format("2006-01-02 15:04:05"),
		Promotional:     detail.Promotional,
		FreeLeech:       detail.FreeLeech,
		HDR:             detail.HDR,
		Dubbed:          detail.Dubbed,
		Subtitled:       detail.Subtitled,
		UploadFactor:    detail.UploadFactor,
		DownloadFactor:  detail.DownloadFactor,
		Comments:        detail.Comments,
		IMDBID:          detail.IMDBID,
		TMDBID:          fmt.Sprintf("%d", detail.TMDBID),
		TVDBID:          detail.TVDBID,
		ReleaseGroup:    detail.ReleaseGroup,
		Meta:            make(map[string]string),
		Files:           make([]*TorrentFile, 0, len(detail.Files)),
		Screenshots:      detail.Pictures,
		CommentsList:    detail.CommentsList,
	}

	// 复制元数据
	for k, v := range detail.Meta {
		if str, ok := v.(string); ok {
			torrentDetail.Meta[k] = str
		}
	}

	// 转换文件列表
	for _, file := range detail.Files {
		torrentFile := &TorrentFile{
			Path:      file.Path,
			Size:      file.Size,
			Extension: s.getFileExtension(file.Path),
		}
		torrentDetail.Files = append(torrentDetail.Files, torrentFile)
	}

	// 转换媒体信息
	if detail.MediaInfo != nil {
		torrentDetail.MediaInfo = &MediaInfo{
			Type:         detail.MediaInfo.Type,
			Title:        detail.MediaInfo.Title,
			OriginalTitle: detail.MediaInfo.OriginalTitle,
			Year:         detail.MediaInfo.Year,
			Season:       detail.MediaInfo.Season,
			Episode:      detail.MediaInfo.Episode,
			IMDBID:       detail.MediaInfo.IMDBID,
			TMDBID:       fmt.Sprintf("%d", detail.MediaInfo.TMDBID),
			TVDBID:       detail.MediaInfo.TVDBID,
			Overview:     detail.MediaInfo.Overview,
			Genres:       detail.MediaInfo.Genres,
			Poster:       detail.MediaInfo.Poster,
			Backdrop:     detail.MediaInfo.Backdrop,
			Rating:       detail.MediaInfo.Rating,
			Runtime:      detail.MediaInfo.Runtime,
			ReleaseDate:  detail.MediaInfo.ReleaseDate,
			Status:       detail.MediaInfo.Status,
			Network:      detail.MediaInfo.Network,
			Language:     detail.MediaInfo.Language,
			Country:      detail.MediaInfo.Country,
			Director:     detail.MediaInfo.Director,
			Writer:       detail.MediaInfo.Writer,
			Actors:       detail.MediaInfo.Actors,
			Subtitles:    detail.MediaInfo.Subtitles,
		}
	}

	// 转换技术信息
	if detail.Meta != nil {
		if _, ok := detail.Meta["container"]; ok {
			torrentDetail.TechnicalInfo = &TechnicalInfo{
				Container: detail.Meta["container"],
			}
		}
	}

	return torrentDetail
}

// getFileExtension 获取文件扩展名
func (s *IndexerService) getFileExtension(path string) string {
	if lastDot := len(path) - 1; lastDot >= 0 && path[lastDot] == '.' {
		return path[lastDot+1:]
	}
	return ""
}

// GetSiteInfo 获取站点信息
func (s *IndexerService) GetSiteInfo(siteID string) (*SiteInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 从Spider管理器获取站点配置
	config, err := s.spiderManager.GetConfig(siteID)
	if err != nil {
		return nil, fmt.Errorf("获取站点配置失败: %w", err)
	}

	// 从Parser管理器获取站点信息
	parser, err := s.parserManager.GetParser(siteID)
	if err != nil {
		return nil, fmt.Errorf("获取解析器失败: %w", err)
	}

	siteInfo := parser.GetSiteInfo()
	if siteInfo == nil {
		return nil, fmt.Errorf("站点信息为空")
	}

	// 合并信息
	combinedInfo := &SiteInfo{
		Name:         siteInfo.Name,
		Domain:       siteInfo.Domain,
		Charset:      siteInfo.Charset,
		Description:  siteInfo.Description,
		Features:     siteInfo.Features,
		Tags:         siteInfo.Tags,
		Categories:   siteInfo.Categories,
		Resolutions:  siteInfo.Resolutions,
		VideoCodecs:  siteInfo.VideoCodecs,
		AudioCodecs:  siteInfo.AudioCodecs,
		Containers:   siteInfo.Containers,
		Sources:      siteInfo.Sources,
		Enabled:      config.Enabled,
		Username:     config.Username,
	}

	return combinedInfo, nil
}

// ListSites 列出所有站点
func (s *IndexerService) ListSites() map[string]*SiteInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sites := make(map[string]*SiteInfo)

	// 从Spider管理器获取所有站点配置
	siteConfigs := s.spiderManager.ListSites()
	for siteID, config := range siteConfigs {
		// 获取解析器信息
		parser, err := s.parserManager.GetParser(siteID)
		if err != nil {
			logger.Warn("获取解析器失败", 
				zap.String("site_id", siteID),
				zap.Error(err))
			continue
		}

		siteInfo := parser.GetSiteInfo()
		if siteInfo == nil {
			continue
		}

		// 合并信息
		sites[siteID] = &SiteInfo{
			Name:         siteInfo.Name,
			Domain:       siteInfo.Domain,
			Charset:      siteInfo.Charset,
			Description:  siteInfo.Description,
			Features:     siteInfo.Features,
			Tags:         siteInfo.Tags,
			Categories:   siteInfo.Categories,
			Resolutions:  siteInfo.Resolutions,
			VideoCodecs:  siteInfo.VideoCodecs,
			AudioCodecs:  siteInfo.AudioCodecs,
			Containers:   siteInfo.Containers,
			Sources:      siteInfo.Sources,
			Enabled:      config.Enabled,
			Username:     config.Username,
		}
	}

	return sites
}

// HealthCheck 健康检查
func (s *IndexerService) HealthCheck(ctx context.Context) map[string]error {
	logger.Info("开始索引器服务健康检查")

	// Spider健康检查
	spiderResults := s.spiderManager.HealthCheck(ctx)

	// Parser健康检查
	parserConfigs := s.parserManager.ListParsers()
	parserResults := make(map[string]error)
	for siteID := range parserConfigs {
		// 简单检查解析器是否存在
		if _, err := s.parserManager.GetParser(siteID); err != nil {
			parserResults[siteID] = err
		}
	}

	// 合并结果
	combinedResults := make(map[string]error)
	for siteID, err := range spiderResults {
		combinedResults["spider_"+siteID] = err
	}
	for siteID, err := range parserResults {
		combinedResults["parser_"+siteID] = err
	}

	healthyCount := 0
	unhealthyCount := 0
	for _, err := range combinedResults {
		if err == nil {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	logger.Info("健康检查完成",
		zap.Int("healthy", healthyCount),
		zap.Int("unhealthy", unhealthyCount))

	return combinedResults
}

// GetStatistics 获取统计信息
func (s *IndexerService) GetStatistics() *ServiceStatistics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	spiderStats := s.spiderManager.GetStatistics()
	parserStats := s.parserManager.GetStatistics()

	return &ServiceStatistics{
		SpiderStatistics: spiderStats,
		ParserStatistics: parserStats,
		TotalSites:       len(spiderStats.SitesByType),
		LastUpdated:      "now", // TODO: 实际时间
	}
}

// Close 关闭服务
func (s *IndexerService) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger.Info("关闭索引器服务")

	// 关闭Spider管理器
	if err := s.spiderManager.Close(); err != nil {
		logger.Warn("关闭Spider管理器失败", zap.Error(err))
	}

	// 关闭Parser管理器
	s.parserManager.Close()

	return nil
}

// 数据结构定义

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword     string   `json:"keyword"`      // 搜索关键词
	SiteIDs     []string `json:"site_ids"`     // 站点ID列表
	Page        int      `json:"page"`         // 页码
	Limit       int      `json:"limit"`        // 每页数量
	MinSeeders  int      `json:"min_seeders"` // 最少做种数
	FreeLeech   bool     `json:"free_leech"`  // 是否免费
	HDR         bool     `json:"hdr"`         // 是否HDR
	Category    string   `json:"category"`    // 分类
	Resolution  string   `json:"resolution"`  // 分辨率
	SortBy      string   `json:"sort_by"`     // 排序方式
	SortOrder   string   `json:"sort_order"`   // 排序顺序
}

// SearchResponse 搜索响应
type SearchResponse struct {
	SiteResults   map[string]*SearchSiteResult `json:"site_results"`   // 各站点搜索结果
	TotalResults  int                         `json:"total_results"` // 总结果数
	Metadata      *SearchMetadata              `json:"metadata"`      // 搜索元数据
}

// SearchSiteResult 单站点搜索结果
type SearchSiteResult struct {
	SiteID     string        `json:"site_id"`     // 站点ID
	Torrents    []*TorrentInfo `json:"torrents"`    // 种子列表
	Count       int           `json:"count"`       // 结果数量
	HasMore     bool          `json:"has_more"`    // 是否还有更多
	SearchTime  string        `json:"search_time"` // 搜索耗时
	Error       string        `json:"error"`       // 错误信息
}

// SearchMetadata 搜索元数据
type SearchMetadata struct {
	Keyword     string    `json:"keyword"`     // 搜索关键词
	SearchTime  string    `json:"search_time"` // 搜索总耗时
	SitesCount  int       `json:"sites_count"` // 搜索站点数
	TotalSites  int       `json:"total_sites"` // 总站点数
	Timestamp   string    `json:"timestamp"`   // 搜索时间
}

// SiteInfo 站点信息（合并版）
type SiteInfo struct {
	Name         string   `json:"name"`         // 站点名称
	Domain       string   `json:"domain"`       // 站点域名
	Charset      string   `json:"charset"`      // 字符编码
	Description  string   `json:"description"`  // 描述
	Features     []string `json:"features"`     // 特性
	Tags         []string `json:"tags"`         // 标签
	Categories   []string `json:"categories"`   // 分类
	Resolutions  []string `json:"resolutions"`  // 分辨率
	VideoCodecs  []string `json:"video_codecs"`  // 视频编码
	AudioCodecs  []string `json:"audio_codecs"`  // 音频编码
	Containers   []string `json:"containers"`    // 容器格式
	Sources      []string `json:"sources"`       // 视频来源
	Enabled      bool     `json:"enabled"`      // 是否启用
	Username     string   `json:"username"`     // 用户名
}

// ServiceStatistics 服务统计信息
type ServiceStatistics struct {
	SpiderStatistics *spider.Statistics     `json:"spider_statistics"` // Spider统计
	ParserStatistics *parser.Statistics    `json:"parser_statistics"` // Parser统计
	TotalSites       int                  `json:"total_sites"`       // 总站点数
	LastUpdated      string               `json:"last_updated"`     // 最后更新时间
}

// 类型别名以避免冲突
type (
	TorrentInfo    = parser.TorrentInfo
	TorrentDetail  = parser.TorrentDetail
	TorrentFile    = parser.TorrentFile
	MediaInfo      = parser.MediaInfo
	TechnicalInfo  = parser.TechnicalInfo
)