package emby

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

// Config Emby 客户端配置
// URL 和 APIKey 预计从配置系统注入

type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Client Emby 媒体服务器客户端骨架
type Client struct {
	baseURL *url.URL
	apiKey  string
	client  *http.Client
	logger  *zap.Logger
}

// embySystemInfo Emby /System/Info 响应结构（简化版）
type embySystemInfo struct {
	SystemID   string `json:"SystemId"`
	ServerName string `json:"ServerName"`
	Version    string `json:"Version"`
}

// embyLibraryItem Emby 媒体库条目
type embyLibraryItem struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	CollectionType string `json:"CollectionType"`
	ItemCount      int64  `json:"ItemCount"`
}

// embyLibraryResponse /Library/MediaFolders 响应
type embyLibraryResponse struct {
	Items []embyLibraryItem `json:"Items"`
}

// embyItem Emby 单个媒体条目（简化）
type embyItem struct {
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

// embyItemsResponse Emby /Items 或搜索响应
type embyItemsResponse struct {
	Items []embyItem `json:"Items"`
}

// NewClient 创建 Emby 客户端实例
// 目前只做参数检查与 http.Client 初始化，具体 API 后续逐步实现

func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("Emby BaseURL 不能为空")
	}
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 Emby BaseURL 失败: %w", err)
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
		return fmt.Errorf("Emby APIKey 未配置")
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

// TestConnection 实现 mediaserver.MediaServerClient.TestConnection
// 这里先做一个轻量级的 /System/Info Ping，占位实现

func (c *Client) TestConnection(ctx context.Context) error {
	if err := c.ensureAPIKey(); err != nil {
		return err
	}

	reqURL := c.buildURL("/emby/System/Info", nil)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Emby-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Emby 连接测试失败: status=%d", resp.StatusCode)
	}

	c.logger.Info("Emby 连接测试成功")
	return nil
}

// GetServerInfo 获取服务器基础信息（骨架实现）
func (c *Client) GetServerInfo(ctx context.Context) (*mediaserver.ServerInfo, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	reqURL := c.buildURL("/emby/System/Info", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Emby-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 Emby 服务器信息失败: status=%d", resp.StatusCode)
	}

	var info embySystemInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("解析 Emby 服务器信息失败: %w", err)
	}

	serverInfo := &mediaserver.ServerInfo{
		ID:      info.SystemID,
		Name:    info.ServerName,
		Version: info.Version,
		Type:    "emby",
		URL:     c.baseURL.String(),
	}

	return serverInfo, nil
}

// ListLibraries 列出媒体库（骨架实现）
func (c *Client) ListLibraries(ctx context.Context) ([]*mediaserver.MediaLibrary, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	reqURL := c.buildURL("/emby/Library/MediaFolders", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Emby-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 Emby 媒体库失败: status=%d", resp.StatusCode)
	}

	var lr embyLibraryResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("解析 Emby 媒体库响应失败: %w", err)
	}

	libraries := make([]*mediaserver.MediaLibrary, 0, len(lr.Items))
	for _, item := range lr.Items {
		libType := mediaserver.MediaTypeUnknown
		switch item.CollectionType {
		case "movies":
			libType = mediaserver.MediaTypeMovie
		case "tvshows":
			libType = mediaserver.MediaTypeSeries
		}

		libraries = append(libraries, &mediaserver.MediaLibrary{
			ID:        item.ID,
			Name:      item.Name,
			Type:      libType,
			ItemCount: item.ItemCount,
		})
	}

	return libraries, nil
}

// ScanLibrary 触发媒体库扫描（骨架实现）
func (c *Client) ScanLibrary(ctx context.Context, libraryID string) error {
	// TODO: 调用 /Library/Refresh 或相关 API
	c.logger.Info("触发 Emby 媒体库扫描(占位)", zap.String("library_id", libraryID))
	return nil
}

// GetItem 根据ID获取媒体条目（骨架实现）
func (c *Client) GetItem(ctx context.Context, id string) (*mediaserver.MediaItem, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("Fields", "ProviderIds,ProductionYear,ParentIndexNumber,IndexNumber,SeriesName,OriginalTitle")
	reqURL := c.buildURL("/emby/Items/"+id, q)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Emby-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 Emby 媒体条目失败: status=%d", resp.StatusCode)
	}

	var ei embyItem
	if err := json.NewDecoder(resp.Body).Decode(&ei); err != nil {
		return nil, fmt.Errorf("解析 Emby 媒体条目失败: %w", err)
	}

	item := mapEmbyItemToMediaItem(&ei)
	return item, nil
}

// SearchItems 搜索媒体条目（骨架实现）
func (c *Client) SearchItems(ctx context.Context, query mediaserver.SearchQuery) ([]*mediaserver.MediaItem, error) {
	if err := c.ensureAPIKey(); err != nil {
		return nil, err
	}

	q := url.Values{}
	if query.Keyword != "" {
		q.Set("SearchTerm", query.Keyword)
	}
	if query.Type != "" && query.Type != mediaserver.MediaTypeUnknown {
		// Emby 使用类型字符串，如 Movie, Series, Episode
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

	reqURL := c.buildURL("/emby/Items", q)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Emby-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Emby 搜索失败: status=%d", resp.StatusCode)
	}

	var ir embyItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return nil, fmt.Errorf("解析 Emby 搜索结果失败: %w", err)
	}

	results := make([]*mediaserver.MediaItem, 0, len(ir.Items))
	for i := range ir.Items {
		item := mapEmbyItemToMediaItem(&ir.Items[i])
		results = append(results, item)
	}

	return results, nil
}

// mapEmbyItemToMediaItem 将 Emby 的 item 映射为统一的 MediaItem
func mapEmbyItemToMediaItem(ei *embyItem) *mediaserver.MediaItem {
	item := &mediaserver.MediaItem{
		ID:           ei.ID,
		Name:         ei.Name,
		OriginalName: ei.OriginalTitle,
		Type:         mediaserver.MediaTypeUnknown,
		ExternalIDs:  mediaserver.ExternalID{},
	}

	// 类型映射
	switch ei.Type {
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
	if ei.ProductionYear != nil {
		year := *ei.ProductionYear
		item.Year = &year
	}

	// 季/集
	if ei.ParentIndexNumber != nil {
		season := *ei.ParentIndexNumber
		item.Season = &season
	}
	if ei.IndexNumber != nil {
		ep := *ei.IndexNumber
		item.Episode = &ep
	}

	// 外部ID映射
	if ei.ProviderIDs != nil {
		if v, ok := ei.ProviderIDs["Tmdb"]; ok && v != "" {
			if id, err := strconv.Atoi(v); err == nil {
				item.ExternalIDs.TMDBID = &id
			}
		}
		if v, ok := ei.ProviderIDs["Tvdb"]; ok && v != "" {
			if id, err := strconv.Atoi(v); err == nil {
				item.ExternalIDs.TVDBID = &id
			}
		}
		if v, ok := ei.ProviderIDs["Imdb"]; ok && v != "" {
			id := v
			item.ExternalIDs.IMDBID = &id
		}
		if v, ok := ei.ProviderIDs["Douban"]; ok && v != "" {
			id := v
			item.ExternalIDs.DoubanID = &id
		}
	}

	return item
}

// 编译期断言，确保 Client 实现了 MediaServerClient 接口
var _ mediaserver.MediaServerClient = (*Client)(nil)
