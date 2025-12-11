package prowlarr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/integration/indexer"
	"moviepilot-go/pkg/logger"
)

// Config Prowlarr 配置
type Config struct {
	// BaseURL Prowlarr 服务地址
	BaseURL string
	// APIKey API 密钥
	APIKey string
	// Timeout 请求超时时间
	Timeout time.Duration
}

// Client Prowlarr 客户端
type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
	logger  *zap.Logger
}

// NewClient 创建 Prowlarr 客户端
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		client:  &http.Client{Timeout: cfg.Timeout},
		logger:  logger.GetLogger(),
	}, nil
}

// Name 返回索引器名称
func (c *Client) Name() string {
	return "prowlarr"
}

// Search 搜索种子
func (c *Client) Search(ctx context.Context, opts indexer.SearchOptions) ([]*indexer.Torrent, error) {
	// 构建搜索 URL
	searchURL := fmt.Sprintf("%s/api/v1/search", c.baseURL)

	params := url.Values{}
	if opts.Query != "" {
		params.Set("query", opts.Query)
	}

	if opts.IMDBID != "" {
		params.Set("imdbId", opts.IMDBID)
	}

	if opts.TMDBID > 0 {
		params.Set("tmdbId", strconv.Itoa(opts.TMDBID))
	}

	if opts.Category != "" {
		params.Set("categories", c.mapCategory(opts.Category))
	}

	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}

	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}

	fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())

	c.logger.Debug("Prowlarr 搜索请求", zap.String("url", fullURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Prowlarr API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 解析 JSON 响应
	var results []ProwlarrResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为 Torrent
	torrents := make([]*indexer.Torrent, 0, len(results))
	for _, result := range results {
		torrent := c.mapResultToTorrent(&result)

		// 应用最小做种数过滤
		if opts.MinSeeders > 0 && torrent.Seeders < opts.MinSeeders {
			continue
		}

		torrents = append(torrents, torrent)
	}

	c.logger.Info("Prowlarr 搜索完成",
		zap.String("query", opts.Query),
		zap.Int("count", len(torrents)))

	return torrents, nil
}

// TestConnection 测试连接
func (c *Client) TestConnection(ctx context.Context) error {
	testURL := fmt.Sprintf("%s/api/v1/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Prowlarr API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	c.logger.Info("Prowlarr 连接测试成功")
	return nil
}

// GetCapabilities 获取索引器能力
func (c *Client) GetCapabilities(ctx context.Context) (*indexer.Capabilities, error) {
	// Prowlarr 支持的基本能力
	return &indexer.Capabilities{
		SupportedCategories: []indexer.TorrentCategory{
			indexer.CategoryMovie,
			indexer.CategoryTV,
			indexer.CategoryAnime,
			indexer.CategoryMusic,
			indexer.CategoryOther,
		},
		SupportIMDBSearch: true,
		SupportTMDBSearch: true, // Prowlarr 支持 TMDB
		MaxResults:        100,
	}, nil
}

// mapResultToTorrent 将 Prowlarr Result 映射为 Torrent
func (c *Client) mapResultToTorrent(result *ProwlarrResult) *indexer.Torrent {
	torrent := &indexer.Torrent{
		Title:       result.Title,
		Link:        result.DownloadUrl,
		MagnetURL:   result.MagnetUrl,
		Size:        result.Size,
		Seeders:     result.Seeders,
		Leechers:    result.Leechers,
		IndexerName: "prowlarr",
	}

	// 解析发布时间
	if result.PublishDate != "" {
		if t, err := time.Parse(time.RFC3339, result.PublishDate); err == nil {
			torrent.PublishDate = t
		}
	}

	// 设置 IMDB ID
	if result.ImdbId != "" {
		torrent.IMDBID = result.ImdbId
	}

	// 设置 TMDB ID
	if result.TmdbId > 0 {
		torrent.TMDBID = result.TmdbId
	}

	// 映射分类
	torrent.Category = c.mapCategoryFromID(result.Categories)

	return torrent
}

// mapCategory 映射分类到 Prowlarr 分类 ID
func (c *Client) mapCategory(category indexer.TorrentCategory) string {
	switch category {
	case indexer.CategoryMovie:
		return "2000" // Movies
	case indexer.CategoryTV:
		return "5000" // TV
	case indexer.CategoryAnime:
		return "5070" // Anime
	case indexer.CategoryMusic:
		return "3000" // Audio
	default:
		return ""
	}
}

// mapCategoryFromID 从分类 ID 映射到分类
func (c *Client) mapCategoryFromID(categories []int) indexer.TorrentCategory {
	if len(categories) == 0 {
		return indexer.CategoryOther
	}

	// 取第一个分类
	catID := categories[0]
	switch {
	case catID >= 2000 && catID < 3000:
		return indexer.CategoryMovie
	case catID >= 5000 && catID < 6000:
		return indexer.CategoryTV
	case catID == 5070:
		return indexer.CategoryAnime
	case catID >= 3000 && catID < 4000:
		return indexer.CategoryMusic
	default:
		return indexer.CategoryOther
	}
}

// ProwlarrResult Prowlarr 搜索结果
type ProwlarrResult struct {
	Title       string `json:"title"`
	DownloadUrl string `json:"downloadUrl"`
	MagnetUrl   string `json:"magnetUrl"`
	Size        int64  `json:"size"`
	Seeders     int    `json:"seeders"`
	Leechers    int    `json:"leechers"`
	PublishDate string `json:"publishDate"`
	ImdbId      string `json:"imdbId"`
	TmdbId      int    `json:"tmdbId"`
	Categories  []int  `json:"categories"`
}

// 编译期断言
var _ indexer.Client = (*Client)(nil)
