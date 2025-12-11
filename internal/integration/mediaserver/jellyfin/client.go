package jellyfin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/integration/mediaserver"
	appLogger "moviepilot-go/pkg/logger"
)

// Config Jellyfin 客户端配置
// Jellyfin 与 Emby API 结构非常相似，采用 BaseURL + APIKey 方式访问

type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Client Jellyfin 媒体服务器客户端骨架
type Client struct {
	baseURL *url.URL
	apiKey  string
	client  *http.Client
	logger  *zap.Logger
}

// jellyfinSystemInfo /System/Info 响应（与 Emby 兼容）
type jellyfinSystemInfo struct {
	SystemID   string `json:"SystemId"`
	ServerName string `json:"ServerName"`
	Version    string `json:"Version"`
}

// jellyfinLibraryItem 媒体库条目（与 Emby 接近）
type jellyfinLibraryItem struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	CollectionType string `json:"CollectionType"`
	ItemCount      int64  `json:"ItemCount"`
}

// jellyfinLibraryResponse /Library/MediaFolders 响应
type jellyfinLibraryResponse struct {
	Items []jellyfinLibraryItem `json:"Items"`
}

// jellyfinItem 单个媒体条目（简化）
type jellyfinItem struct {
	ID                string            `json:"Id"`
	Name              string            `json:"Name"`
	OriginalTitle     string            `json:"OriginalTitle"`
	Type              string            `json:"Type"`
	ProductionYear    *int              `json:"ProductionYear"`
	ParentIndexNumber *int              `json:"ParentIndexNumber"`
	IndexNumber       *int              `json:"IndexNumber"`
	SeriesName        string            `json:"SeriesName"`
	ProviderIDs       map[string]string `json:"ProviderIds"`
}

// jellyfinItemsResponse /Items 搜索结果
type jellyfinItemsResponse struct {
	Items []jellyfinItem `json:"Items"`
}

// NewClient 创建 Jellyfin 客户端实例
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("Jellyfin BaseURL 不能为空")
	}
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 Jellyfin BaseURL 失败: %w", err)
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: parsed,
		apiKey:  cfg.APIKey,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: appLogger.GetLogger(),
	}, nil
}

// ensureAPIKey 校验是否配置了 API Key
func (c *Client) ensureAPIKey() error {
	if c.apiKey == "" {
		return fmt.Errorf("Jellyfin APIKey 未配置")
	}
	return nil
}

// buildURL 构造带路径的完整URL
func (c *Client) buildURL(path string, q url.Values) string {
	u := *c.baseURL
	u.Path = path
	if q != nil {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// addHeaders 添加 Jellyfin 必要头部
func (c *Client) addHeaders(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("X-MediaBrowser-Token", c.apiKey)
	}
}

// TestConnection 测试 Jellyfin 连接
func (c *Client) TestConnection(ctx context.Context) error {
	if err := c.ensureAPIKey(); err != nil {
		return err
	}

	reqURL := c.buildURL("/System/Info", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Jellyfin 连接测试失败: status=%d", resp.StatusCode)
	}

	c.logger.Info("Jellyfin 连接测试成功")
	return nil
}

// GetServerInfo 获取服务器基础信息
func (c *Client) GetServerInfo(ctx context.Context) (*mediaserver.ServerInfo, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	reqURL := c.buildURL("/System/Info", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 Jellyfin 服务器信息失败: status=%d", resp.StatusCode)
	}

	var info jellyfinSystemInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("解析 Jellyfin 服务器信息失败: %w", err)
	}

	server := &mediaserver.ServerInfo{
		ID:      info.SystemID,
		Name:    info.ServerName,
		Version: info.Version,
		Type:    "jellyfin",
		URL:     c.baseURL.String(),
	}
	return server, nil
}

// ListLibraries 列出媒体库
func (c *Client) ListLibraries(ctx context.Context) ([]*mediaserver.MediaLibrary, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	reqURL := c.buildURL("/Library/MediaFolders", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 Jellyfin 媒体库失败: status=%d", resp.StatusCode)
	}

	var lr jellyfinLibraryResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("解析 Jellyfin 媒体库响应失败: %w", err)
	}

	libs := make([]*mediaserver.MediaLibrary, 0, len(lr.Items))
	for _, item := range lr.Items {
		mType := mediaserver.MediaTypeUnknown
		switch item.CollectionType {
		case "movies":
			mType = mediaserver.MediaTypeMovie
		case "tvshows":
			mType = mediaserver.MediaTypeSeries
		}

		libs = append(libs, &mediaserver.MediaLibrary{
			ID:        item.ID,
			Name:      item.Name,
			Type:      mType,
			ItemCount: item.ItemCount,
		})
	}

	return libs, nil
}

// ScanLibrary 触发媒体库扫描（占位实现）
func (c *Client) ScanLibrary(ctx context.Context, libraryID string) error {
	// TODO: 调用 Jellyfin 对应库刷新 API
	c.logger.Info("触发 Jellyfin 媒体库扫描(占位)", zap.String("library_id", libraryID))
	return nil
}

// GetItem 根据ID获取媒体条目
func (c *Client) GetItem(ctx context.Context, id string) (*mediaserver.MediaItem, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("Fields", "ProviderIds,ProductionYear,ParentIndexNumber,IndexNumber,SeriesName,OriginalTitle")
	reqURL := c.buildURL("/Items/"+id, q)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 Jellyfin 媒体条目失败: status=%d", resp.StatusCode)
	}

	var ji jellyfinItem
	if err := json.NewDecoder(resp.Body).Decode(&ji); err != nil {
		return nil, fmt.Errorf("解析 Jellyfin 媒体条目失败: %w", err)
	}

	item := mapJellyfinItemToMediaItem(&ji)
	return item, nil
}

// SearchItems 搜索媒体条目
func (c *Client) SearchItems(ctx context.Context, query mediaserver.SearchQuery) ([]*mediaserver.MediaItem, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	q := url.Values{}
	if query.Keyword != "" {
		q.Set("SearchTerm", query.Keyword)
	}
	if query.Type != "" && query.Type != mediaserver.MediaTypeUnknown {
		// Jellyfin 与 Emby 类似，使用 Movie/Series/Episode 等类型
		switch query.Type {
		case mediaserver.MediaTypeMovie:
			q.Set("IncludeItemTypes", "Movie")
		case mediaserver.MediaTypeSeries:
			q.Set("IncludeItemTypes", "Series")
		case mediaserver.MediaTypeEpisode:
			q.Set("IncludeItemTypes", "Episode")
		}
	}
	if query.Year != nil {
		q.Set("Years", strconv.Itoa(*query.Year))
	}
	if query.Limit > 0 {
		q.Set("Limit", strconv.Itoa(query.Limit))
	} else {
		q.Set("Limit", "20")
	}
	q.Set("Recursive", "true")
	q.Set("IncludeExternalContent", "true")

	reqURL := c.buildURL("/Items", q)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jellyfin 搜索失败: status=%d", resp.StatusCode)
	}

	var ir jellyfinItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return nil, fmt.Errorf("解析 Jellyfin 搜索结果失败: %w", err)
	}

	results := make([]*mediaserver.MediaItem, 0, len(ir.Items))
	for i := range ir.Items {
		item := mapJellyfinItemToMediaItem(&ir.Items[i])
		results = append(results, item)
	}

	return results, nil
}

// mapJellyfinItemToMediaItem 将 Jellyfin 条目映射为统一 MediaItem
func mapJellyfinItemToMediaItem(ji *jellyfinItem) *mediaserver.MediaItem {
	item := &mediaserver.MediaItem{
		ID:           ji.ID,
		Name:         ji.Name,
		OriginalName: ji.OriginalTitle,
		Type:         mediaserver.MediaTypeUnknown,
		ExternalIDs:  mediaserver.ExternalID{},
	}

	// 类型映射
	switch ji.Type {
	case "Movie":
		item.Type = mediaserver.MediaTypeMovie
	case "Series":
		item.Type = mediaserver.MediaTypeSeries
	case "Episode":
		item.Type = mediaserver.MediaTypeEpisode
	case "Season":
		item.Type = mediaserver.MediaTypeSeason
	}

	// 年份
	if ji.ProductionYear != nil {
		year := *ji.ProductionYear
		item.Year = &year
	}

	// 季/集
	if ji.ParentIndexNumber != nil {
		season := *ji.ParentIndexNumber
		item.Season = &season
	}
	if ji.IndexNumber != nil {
		ep := *ji.IndexNumber
		item.Episode = &ep
	}

	// 外部ID映射
	if ji.ProviderIDs != nil {
		if v, ok := ji.ProviderIDs["Tmdb"]; ok && v != "" {
			if id, err := strconv.Atoi(v); err == nil {
				item.ExternalIDs.TMDBID = &id
			}
		}
		if v, ok := ji.ProviderIDs["Tvdb"]; ok && v != "" {
			if id, err := strconv.Atoi(v); err == nil {
				item.ExternalIDs.TVDBID = &id
			}
		}
		if v, ok := ji.ProviderIDs["Imdb"]; ok && v != "" {
			id := v
			item.ExternalIDs.IMDBID = &id
		}
		if v, ok := ji.ProviderIDs["Douban"]; ok && v != "" {
			id := v
			item.ExternalIDs.DoubanID = &id
		}
	}

	return item
}

// 编译期断言，确保 Client 实现了 MediaServerClient 接口
var _ mediaserver.MediaServerClient = (*Client)(nil)
