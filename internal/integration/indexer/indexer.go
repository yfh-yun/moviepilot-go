// Package indexer 索引模块
// 提供站点索引和搜索功能
package indexer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"

	"go.uber.org/zap"
)

// ModuleType 模块类型
type ModuleType string

const (
	ModuleTypeIndexer ModuleType = "Indexer"
)

// OtherModulesType 其他模块类型
type OtherModulesType string

const (
	OtherModulesTypeIndexer OtherModulesType = "Indexer"
)

// IndexerModule 索引模块
type IndexerModule struct {
	logger      *zap.Logger
	siteParsers map[string]SiteParser
	siteSpiders map[string]SiteSpider
	sitesHelper SitesHelper
	mutex       sync.RWMutex
}

// SiteParser 站点解析器接口
type SiteParser interface {
	GetSchema() string
	ParseSite(html string) (*SiteInfo, error)
	ParseTorrentList(html string) ([]*TorrentInfo, error)
	ParseTorrentDetail(html string) (*TorrentDetail, error)
	ParseUser(html string) (*UserInfo, error)
}

// SiteSpider 站点爬虫接口
type SiteSpider interface {
	GetName() string
	GetDomain() string
	Search(ctx context.Context, keyword string, mediaType string) ([]*TorrentInfo, error)
	GetTorrentDetail(ctx context.Context, id string) (*TorrentDetail, error)
	GetUserInfo(ctx context.Context) (*UserInfo, error)
	Download(ctx context.Context, id string) ([]byte, error)
}

// SitesHelper 站点助手接口
type SitesHelper interface {
	GetIndexers() []*SiteInfo
	Check(domain string) (bool, string)
	GetSiteUserData(siteID string) *models.SiteUserData
}

// SiteInfo 站点信息
type SiteInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Schema      string            `json:"schema"`
	URL         string            `json:"url"`
	Language    string            `json:"language"`
	Enabled     bool              `json:"enabled"`
	Priority    int               `json:"priority"`
	Settings    map[string]string `json:"settings"`
	LastCheck   time.Time         `json:"last_check"`
	Status      string            `json:"status"`
}

// TorrentInfo 种子信息
type TorrentInfo struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Size         int64             `json:"size"`
	Seeders      int               `json:"seeders"`
	Leechers     int               `json:"leechers"`
	Completed    int               `json:"completed"`
	UploadDate   time.Time         `json:"upload_date"`
	DownloadURL  string            `json:"download_url"`
	DetailURL    string            `json:"detail_url"`
	Category     string            `json:"category"`
	Tags         []string          `json:"tags"`
	Uploader     string            `json:"uploader"`
	IMDBID       string            `json:"imdb_id"`
	TMDBID       int               `json:"tmdb_id"`
	FreeTorrent bool              `json:"free_torrent"`
	DoubleUpload bool             `json:"double_upload"`
	HDR          bool              `json:"hdr"`
	UHD          bool              `json:"uhd"`
	Source       string            `json:"source"`
	Codec        string            `json:"codec"`
	Resolution   string            `json:"resolution"`
	Container    string            `json:"container"`
	Audio        string            `json:"audio"`
	Subtitles    []string          `json:"subtitles"`
	Metadata     map[string]string `json:"metadata"`
}

// TorrentDetail 种子详情
type TorrentDetail struct {
	*TorrentInfo
	Files        []*TorrentFile `json:"files"`
	Comments     []string       `json:"comments"`
	Grabs        int            `json:"grabs"`
	LastActivity time.Time      `json:"last_activity"`
	Peers        []*PeerInfo    `json:"peers"`
}

// TorrentFile 种子文件
type TorrentFile struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

// PeerInfo 连接信息
type PeerInfo struct {
	IP        string    `json:"ip"`
	Port      int       `json:"port"`
	Client    string    `json:"client"`
	Progress  float64   `json:"progress"`
	Upload    int64     `json:"upload"`
	Download  int64     `json:"download"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Uploaded     int64     `json:"uploaded"`
	Downloaded   int64     `json:"downloaded"`
	Ratio        float64   `json:"ratio"`
	Uploads      int       `json:"uploads"`
	Downloads    int       `json:"downloads"`
	Seeding      int       `json:"seeding"`
	Leeching     int       `json:"leeching"`
	Rank         string    `json:"rank"`
	JoinDate     time.Time `json:"join_date"`
	LastActive   time.Time `json:"last_active"`
	Upload       int64     `json:"upload"`
	Download     int64     `json:"download"`
	BonusPoints  float64   `json:"bonus_points"`
}

// Comment 评论信息
type Comment struct {
	ID       string    `json:"id"`
	Author   string    `json:"author"`
	Content  string    `json:"content"`
	PostDate time.Time `json:"post_date"`
}

// TorrentDetail 种子详情
type TorrentDetail struct {
	*TorrentInfo
	Files        []*TorrentFile `json:"files"`
	Comments     []*Comment     `json:"comments"`
	Grabs        int            `json:"grabs"`
	LastActivity time.Time      `json:"last_activity"`
	Peers        []*PeerInfo    `json:"peers"`
}
	JoinDate     time.Time `json:"join_date"`
	LastActive   time.Time `json:"last_active"`
	Class        string    `json:"class"`
	BonusPoints  float64   `json:"bonus_points"`
	Invites      int       `json:"invites"`
	SeedTime     int64     `json:"seed_time"`
	SeedBonus    float64   `json:"seed_bonus"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword    string            `json:"keyword"`
	MediaType  string            `json:"media_type"`
	SiteIDs    []string          `json:"site_ids"`
	Category   string            `json:"category"`
	Tags       []string          `json:"tags"`
	MinSize    int64             `json:"min_size"`
	MaxSize    int64             `json:"max_size"`
	MinSeeders int               `json:"min_seeders"`
	Freeleech  bool              `json:"freeleech"`
	HDR        bool              `json:"hdr"`
	UHD        bool              `json:"uhd"`
	IMDBID     string            `json:"imdb_id"`
	TMDBID     int               `json:"tmdb_id"`
	Filters    map[string]string `json:"filters"`
	Limit      int               `json:"limit"`
	Page       int               `json:"page"`
}

// SearchResult 搜索结果
type SearchResult struct {
	SiteName    string          `json:"site_name"`
	SiteID      string          `json:"site_id"`
	Torrents    []*TorrentInfo  `json:"torrents"`
	Total       int             `json:"total"`
	Page        int             `json:"page"`
	HasMore     bool            `json:"has_more"`
	SearchTime  time.Duration   `json:"search_time"`
	Error       string          `json:"error,omitempty"`
}

// NewIndexerModule 创建索引模块
func NewIndexerModule(sitesHelper SitesHelper) *IndexerModule {
	return &IndexerModule{
		logger:      logger.Logger,
		siteParsers: make(map[string]SiteParser),
		siteSpiders: make(map[string]SiteSpider),
		sitesHelper: sitesHelper,
		mutex:       sync.RWMutex{},
	}
}

// InitModule 初始化模块
func (im *IndexerModule) InitModule(ctx context.Context) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	// 注册内置解析器
	im.registerBuiltinParsers()

	// 注册内置爬虫
	im.registerBuiltinSpiders()

	im.logger.Info("Indexer module initialized",
		zap.Int("parsers_count", len(im.siteParsers)),
		zap.Int("spiders_count", len(im.siteSpiders)))

	return nil
}

// GetName 获取模块名称
func (im *IndexerModule) GetName() string {
	return "站点索引"
}

// GetType 获取模块类型
func (im *IndexerModule) GetType() ModuleType {
	return ModuleTypeIndexer
}

// GetSubType 获取模块子类型
func (im *IndexerModule) GetSubType() OtherModulesType {
	return OtherModulesTypeIndexer
}

// GetPriority 获取模块优先级
func (im *IndexerModule) GetPriority() int {
	return 0
}

// Stop 停止模块
func (im *IndexerModule) Stop(ctx context.Context) error {
	im.logger.Info("Indexer module stopped")
	return nil
}

// Test 测试模块
func (im *IndexerModule) Test(ctx context.Context) error {
	sites := im.sitesHelper.GetIndexers()
	if len(sites) == 0 {
		return fmt.Errorf("未配置站点或未通过用户认证")
	}

	// 测试每个站点
	for _, site := range sites {
		if !site.Enabled {
			continue
		}

		if spider, exists := im.siteSpiders[site.Schema]; exists {
			if _, err := spider.GetUserInfo(ctx); err != nil {
				im.logger.Warn("Site test failed",
					zap.String("site", site.Name),
					zap.Error(err))
			}
		}
	}

	return nil
}

// Search 搜索种子
func (im *IndexerModule) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
	sites := im.sitesHelper.GetIndexers()
	if len(sites) == 0 {
		return nil, fmt.Errorf("no sites configured")
	}

	// 过滤站点
	targetSites := im.filterSites(sites, req)
	if len(targetSites) == 0 {
		return nil, fmt.Errorf("no available sites for search")
	}

	// 并发搜索
	results := make([]*SearchResult, 0)
	resultChan := make(chan *SearchResult, len(targetSites))
	errorChan := make(chan error, len(targetSites))

	var wg sync.WaitGroup
	for _, site := range targetSites {
		wg.Add(1)
		go func(site *SiteInfo) {
			defer wg.Done()
			
			start := time.Now()
			result, err := im.searchSite(ctx, site, req)
			if err != nil {
				errorChan <- fmt.Errorf("site %s search failed: %w", site.Name, err)
				return
			}
			
			result.SearchTime = time.Since(start)
			resultChan <- result
		}(site)
	}

	// 等待所有搜索完成
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// 收集结果
	for result := range resultChan {
		results = append(results, result)
	}

	// 收集错误
	var errors []string
	for err := range errorChan {
		errors = append(errors, err.Error())
		im.logger.Warn("Search error", zap.String("error", err))
	}

	if len(results) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all searches failed: %s", strings.Join(errors, "; "))
	}

	return results, nil
}

// GetTorrentDetail 获取种子详情
func (im *IndexerModule) GetTorrentDetail(ctx context.Context, siteID, torrentID string) (*TorrentDetail, error) {
	site := im.getSiteByID(siteID)
	if site == nil {
		return nil, fmt.Errorf("site not found: %s", siteID)
	}

	spider, exists := im.siteSpiders[site.Schema]
	if !exists {
		return nil, fmt.Errorf("spider not found for schema: %s", site.Schema)
	}

	return spider.GetTorrentDetail(ctx, torrentID)
}

// DownloadTorrent 下载种子
func (im *IndexerModule) DownloadTorrent(ctx context.Context, siteID, torrentID string) ([]byte, error) {
	site := im.getSiteByID(siteID)
	if site == nil {
		return nil, fmt.Errorf("site not found: %s", siteID)
	}

	spider, exists := im.siteSpiders[site.Schema]
	if !exists {
		return nil, fmt.Errorf("spider not found for schema: %s", site.Schema)
	}

	return spider.Download(ctx, torrentID)
}

// GetUserInfo 获取用户信息
func (im *IndexerModule) GetUserInfo(ctx context.Context, siteID string) (*UserInfo, error) {
	site := im.getSiteByID(siteID)
	if site == nil {
		return nil, fmt.Errorf("site not found: %s", siteID)
	}

	spider, exists := im.siteSpiders[site.Schema]
	if !exists {
		return nil, fmt.Errorf("spider not found for schema: %s", site.Schema)
	}

	return spider.GetUserInfo(ctx)
}

// RegisterParser 注册解析器
func (im *IndexerModule) RegisterParser(schema string, parser SiteParser) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.siteParsers[schema] = parser
	im.logger.Info("Parser registered", zap.String("schema", schema))
}

// RegisterSpider 注册爬虫
func (im *IndexerModule) RegisterSpider(spider SiteSpider) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.siteSpiders[spider.GetDomain()] = spider
	im.logger.Info("Spider registered", zap.String("domain", spider.GetDomain()))
}

// GetParser 获取解析器
func (im *IndexerModule) GetParser(schema string) (SiteParser, bool) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	parser, exists := im.siteParsers[schema]
	return parser, exists
}

// GetSpider 获取爬虫
func (im *IndexerModule) GetSpider(domain string) (SiteSpider, bool) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	spider, exists := im.siteSpiders[domain]
	return spider, exists
}

// ListParsers 列出所有解析器
func (im *IndexerModule) ListParsers() map[string]SiteParser {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	result := make(map[string]SiteParser)
	for schema, parser := range im.siteParsers {
		result[schema] = parser
	}

	return result
}

// ListSpiders 列出所有爬虫
func (im *IndexerModule) ListSpiders() map[string]SiteSpider {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	result := make(map[string]SiteSpider)
	for domain, spider := range im.siteSpiders {
		result[domain] = spider
	}

	return result
}

// filterSites 过滤站点
func (im *IndexerModule) filterSites(sites []*SiteInfo, req *SearchRequest) []*SiteInfo {
	var filtered []*SiteInfo

	for _, site := range sites {
		if !site.Enabled {
			continue
		}

		// 检查站点ID过滤
		if len(req.SiteIDs) > 0 {
			found := false
			for _, siteID := range req.SiteIDs {
				if site.ID == siteID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查搜索条件
		if !im.canSearch(site, req.Keyword) {
			continue
		}

		// 检查站点流控
		if blocked, msg := im.sitesHelper.Check(site.Domain); blocked {
			im.logger.Warn("Site blocked", zap.String("site", site.Name), zap.String("msg", msg))
			continue
		}

		filtered = append(filtered, site)
	}

	return filtered
}

// canSearch 检查是否可以搜索
func (im *IndexerModule) canSearch(site *SiteInfo, keyword string) bool {
	// 检查语言支持
	if keyword != "" && site.Language == "en" && im.isChinese(keyword) {
		im.logger.Warn("Site does not support Chinese search", zap.String("site", site.Name))
		return false
	}

	return true
}

// isChinese 检查是否包含中文
func (im *IndexerModule) isChinese(text string) bool {
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

// searchSite 搜索单个站点
func (im *IndexerModule) searchSite(ctx context.Context, site *SiteInfo, req *SearchRequest) (*SearchResult, error) {
	spider, exists := im.siteSpiders[site.Schema]
	if !exists {
		return nil, fmt.Errorf("spider not found for schema: %s", site.Schema)
	}

	torrents, err := spider.Search(ctx, req.Keyword, req.MediaType)
	if err != nil {
		return nil, err
	}

	// 应用过滤条件
	filteredTorrents := im.filterTorrents(torrents, req)

	return &SearchResult{
		SiteName: site.Name,
		SiteID:   site.ID,
		Torrents: filteredTorrents,
		Total:    len(filteredTorrents),
		Page:     req.Page,
		HasMore:  len(filteredTorrents) >= req.Limit,
	}, nil
}

// filterTorrents 过滤种子
func (im *IndexerModule) filterTorrents(torrents []*TorrentInfo, req *SearchRequest) []*TorrentInfo {
	var filtered []*TorrentInfo

	for _, torrent := range torrents {
		// 大小过滤
		if req.MinSize > 0 && torrent.Size < req.MinSize {
			continue
		}
		if req.MaxSize > 0 && torrent.Size > req.MaxSize {
			continue
		}

		// 做种数过滤
		if req.MinSeeders > 0 && torrent.Seeders < req.MinSeeders {
			continue
		}

		// 免费种子过滤
		if req.Freeleech && !torrent.FreeTorrent {
			continue
		}

		// HDR过滤
		if req.HDR && !torrent.HDR {
			continue
		}

		// UHD过滤
		if req.UHD && !torrent.UHD {
			continue
		}

		// 分类过滤
		if req.Category != "" && torrent.Category != req.Category {
			continue
		}

		// 标签过滤
		if len(req.Tags) > 0 {
			hasTag := false
			for _, tag := range req.Tags {
				for _, torrentTag := range torrent.Tags {
					if strings.EqualFold(torrentTag, tag) {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		filtered = append(filtered, torrent)
	}

	// 限制结果数量
	if req.Limit > 0 && len(filtered) > req.Limit {
		filtered = filtered[:req.Limit]
	}

	return filtered
}

// getSiteByID 根据ID获取站点
func (im *IndexerModule) getSiteByID(siteID string) *SiteInfo {
	sites := im.sitesHelper.GetIndexers()
	for _, site := range sites {
		if site.ID == siteID {
			return site
		}
	}
	return nil
}

// clearSearchText 清理搜索文本
func (im *IndexerModule) clearSearchText(text string) string {
	if text == "" {
		return ""
	}

	// 移除特殊字符
	text = strings.ReplaceAll(text, ".", " ")
	text = strings.ReplaceAll(text, "_", " ")
	text = strings.ReplaceAll(text, "-", " ")
	
	// 移除多余空格
	text = strings.Join(strings.Fields(text), " ")
	
	return text
}

// registerBuiltinParsers 注册内置解析器
func (im *IndexerModule) registerBuiltinParsers() {
	// 这里注册各种站点解析器
	// 例如：NexusPHP、Gazelle、Unit3D等
}

// registerBuiltinSpiders 注册内置爬虫
func (im *IndexerModule) registerBuiltinSpiders() {
	// 这里注册各种站点爬虫
	// 例如：HaiDan、HDDolby、MTorrent等
}