package jackett

import (
	"context"
	"encoding/xml"
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

// Config Jackett 配置
type Config struct {
	// BaseURL Jackett 服务地址
	BaseURL string
	// APIKey API 密钥
	APIKey string
	// Timeout 请求超时时间
	Timeout time.Duration
}

// Client Jackett 客户端
type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
	logger  *zap.Logger
}

// NewClient 创建 Jackett 客户端
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
	return "jackett"
}

// Search 搜索种子
func (c *Client) Search(ctx context.Context, opts indexer.SearchOptions) ([]*indexer.Torrent, error) {
	// 构建搜索 URL
	searchURL := fmt.Sprintf("%s/api/v2.0/indexers/all/results/torznab/api", c.baseURL)

	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("t", "search")

	if opts.Query != "" {
		params.Set("q", opts.Query)
	}

	if opts.IMDBID != "" {
		params.Set("imdbid", opts.IMDBID)
	}

	if opts.Category != "" {
		params.Set("cat", c.mapCategory(opts.Category))
	}

	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}

	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}

	fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())

	c.logger.Debug("Jackett 搜索请求", zap.String("url", fullURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jackett API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 解析 Torznab XML 响应
	torrents, err := c.parseTorznabResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 应用最小做种数过滤
	if opts.MinSeeders > 0 {
		filtered := make([]*indexer.Torrent, 0, len(torrents))
		for _, t := range torrents {
			if t.Seeders >= opts.MinSeeders {
				filtered = append(filtered, t)
			}
		}
		torrents = filtered
	}

	c.logger.Info("Jackett 搜索完成",
		zap.String("query", opts.Query),
		zap.Int("count", len(torrents)))

	return torrents, nil
}

// TestConnection 测试连接
func (c *Client) TestConnection(ctx context.Context) error {
	testURL := fmt.Sprintf("%s/api/v2.0/indexers/all/results/torznab/api?apikey=%s&t=caps",
		c.baseURL, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Jackett API 返回错误: status=%d, body=%s", resp.StatusCode, string(body))
	}

	c.logger.Info("Jackett 连接测试成功")
	return nil
}

// GetCapabilities 获取索引器能力
func (c *Client) GetCapabilities(ctx context.Context) (*indexer.Capabilities, error) {
	// Jackett 支持的基本能力
	return &indexer.Capabilities{
		SupportedCategories: []indexer.TorrentCategory{
			indexer.CategoryMovie,
			indexer.CategoryTV,
			indexer.CategoryAnime,
			indexer.CategoryMusic,
			indexer.CategoryOther,
		},
		SupportIMDBSearch: true,
		SupportTMDBSearch: false, // Jackett 不直接支持 TMDB
		MaxResults:        100,
	}, nil
}

// parseTorznabResponse 解析 Torznab XML 响应
func (c *Client) parseTorznabResponse(body io.Reader) ([]*indexer.Torrent, error) {
	var rss TorznabRSS
	if err := xml.NewDecoder(body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("解析 XML 失败: %w", err)
	}

	torrents := make([]*indexer.Torrent, 0, len(rss.Channel.Items))
	for _, item := range rss.Channel.Items {
		torrent := c.mapItemToTorrent(&item)
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

// mapItemToTorrent 将 Torznab Item 映射为 Torrent
func (c *Client) mapItemToTorrent(item *TorznabItem) *indexer.Torrent {
	torrent := &indexer.Torrent{
		Title:       item.Title,
		Description: item.Description,
		Link:        item.Link,
		IndexerName: "jackett",
	}

	// 解析发布时间
	if item.PubDate != "" {
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			torrent.PublishDate = t
		}
	}

	// 解析 Torznab 属性
	for _, attr := range item.Attributes {
		switch attr.Name {
		case "size":
			if size, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
				torrent.Size = size
			}
		case "seeders":
			if seeders, err := strconv.Atoi(attr.Value); err == nil {
				torrent.Seeders = seeders
			}
		case "peers":
			if leechers, err := strconv.Atoi(attr.Value); err == nil {
				torrent.Leechers = leechers
			}
		case "magneturl":
			torrent.MagnetURL = attr.Value
		case "imdbid":
			torrent.IMDBID = attr.Value
		}
	}

	// 如果没有磁力链接，使用 Link
	if torrent.MagnetURL == "" && torrent.Link != "" {
		torrent.MagnetURL = torrent.Link
	}

	return torrent
}

// mapCategory 映射分类到 Torznab 分类 ID
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

// TorznabRSS Torznab RSS 响应结构
type TorznabRSS struct {
	XMLName xml.Name       `xml:"rss"`
	Channel TorznabChannel `xml:"channel"`
}

// TorznabChannel Torznab Channel
type TorznabChannel struct {
	Title string        `xml:"title"`
	Items []TorznabItem `xml:"item"`
}

// TorznabItem Torznab Item
type TorznabItem struct {
	Title       string             `xml:"title"`
	Description string             `xml:"description"`
	Link        string             `xml:"link"`
	PubDate     string             `xml:"pubDate"`
	Attributes  []TorznabAttribute `xml:"attr"`
}

// TorznabAttribute Torznab 属性
type TorznabAttribute struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// 编译期断言
var _ indexer.Client = (*Client)(nil)
