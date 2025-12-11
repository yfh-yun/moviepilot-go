package plex

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

// Config Plex 客户端配置
// Token 一般为 X-Plex-Token，从配置系统注入

type Config struct {
	BaseURL  string
	Token    string
	ClientID string // X-Plex-Client-Identifier，可选
	Timeout  time.Duration
}

// Client Plex 媒体服务器客户端骨架
type Client struct {
	baseURL  *url.URL
	token    string
	clientID string
	client   *http.Client
	logger   *zap.Logger
}

// plexServerResponse /servers 或 /identity 响应（简化）
type plexServerResponse struct {
	MediaContainer struct {
		MachineIdentifier string `json:"machineIdentifier"`
		FriendlyName      string `json:"friendlyName"`
		Version           string `json:"version"`
	} `json:"MediaContainer"`
}

// plexLibrarySection Plex 媒体库条目
type plexLibrarySection struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// plexLibrarySectionsResponse /library/sections 响应
type plexLibrarySectionsResponse struct {
	MediaContainer struct {
		Directory []plexLibrarySection `json:"Directory"`
	} `json:"MediaContainer"`
}

// plexMetadataItem 元数据条目（简化）
type plexMetadataItem struct {
	RatingKey     string `json:"ratingKey"`
	Title         string `json:"title"`
	OriginalTitle string `json:"originalTitle"`
	Type          string `json:"type"`
	Year          int    `json:"year"`
	ParentIndex   int    `json:"parentIndex"`
	Index         int    `json:"index"`
	GUID          string `json:"guid"`
}

// plexMetadataResponse /library/metadata/{id} 或 /library/sections/{key}/all 响应
type plexMetadataResponse struct {
	MediaContainer struct {
		Metadata []plexMetadataItem `json:"Metadata"`
	} `json:"MediaContainer"`
}

// NewClient 创建 Plex 客户端实例
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("Plex BaseURL 不能为空")
	}
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 Plex BaseURL 失败: %w", err)
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL:  parsed,
		token:    cfg.Token,
		clientID: cfg.ClientID,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: appLogger.GetLogger(),
	}, nil
}

// ensureToken 校验是否配置了 Token
func (c *Client) ensureToken() error {
	if c.token == "" {
		return fmt.Errorf("Plex Token 未配置")
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

// addHeaders 为请求添加 Plex 必要头部
func (c *Client) addHeaders(req *http.Request) {
	if c.token != "" {
		q := req.URL.Query()
		q.Set("X-Plex-Token", c.token)
		req.URL.RawQuery = q.Encode()
	}
	// 可选的标识头
	if c.clientID != "" {
		req.Header.Set("X-Plex-Client-Identifier", c.clientID)
	}
}

// TestConnection 测试连接是否正常
func (c *Client) TestConnection(ctx context.Context) error {
	if err := c.ensureToken(); err != nil {
		return err
	}

	reqURL := c.buildURL("/identity", nil)
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
		return fmt.Errorf("Plex 连接测试失败: status=%d", resp.StatusCode)
	}

	c.logger.Info("Plex 连接测试成功")
	return nil
}

// GetServerInfo 获取服务器基础信息（骨架实现）
func (c *Client) GetServerInfo(ctx context.Context) (*mediaserver.ServerInfo, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}

	reqURL := c.buildURL("/identity", nil)
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
		return nil, fmt.Errorf("获取 Plex 服务器信息失败: status=%d", resp.StatusCode)
	}

	var sr plexServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("解析 Plex 服务器信息失败: %w", err)
	}

	info := &mediaserver.ServerInfo{
		ID:      sr.MediaContainer.MachineIdentifier,
		Name:    sr.MediaContainer.FriendlyName,
		Version: sr.MediaContainer.Version,
		Type:    "plex",
		URL:     c.baseURL.String(),
	}

	return info, nil
}

// ListLibraries 列出媒体库（骨架实现）
func (c *Client) ListLibraries(ctx context.Context) ([]*mediaserver.MediaLibrary, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}

	reqURL := c.buildURL("/library/sections", nil)
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
		return nil, fmt.Errorf("获取 Plex 媒体库失败: status=%d", resp.StatusCode)
	}

	var lr plexLibrarySectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("解析 Plex 媒体库响应失败: %w", err)
	}

	libs := make([]*mediaserver.MediaLibrary, 0, len(lr.MediaContainer.Directory))
	for _, d := range lr.MediaContainer.Directory {
		mType := mediaserver.MediaTypeUnknown
		switch d.Type {
		case "movie":
			mType = mediaserver.MediaTypeMovie
		case "show":
			mType = mediaserver.MediaTypeSeries
		}

		libs = append(libs, &mediaserver.MediaLibrary{
			ID:   d.Key,
			Name: d.Title,
			Type: mType,
		})
	}

	return libs, nil
}

// ScanLibrary 触发媒体库扫描（占位实现）
func (c *Client) ScanLibrary(ctx context.Context, libraryID string) error {
	// TODO: 调用 /library/sections/{key}/refresh
	c.logger.Info("触发 Plex 媒体库扫描(占位)", zap.String("library_id", libraryID))
	return nil
}

// GetItem 根据ID获取媒体条目（骨架实现）
func (c *Client) GetItem(ctx context.Context, id string) (*mediaserver.MediaItem, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}

	reqURL := c.buildURL("/library/metadata/"+id, nil)
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
		return nil, fmt.Errorf("获取 Plex 媒体条目失败: status=%d", resp.StatusCode)
	}

	var mr plexMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, fmt.Errorf("解析 Plex 媒体条目失败: %w", err)
	}
	if len(mr.MediaContainer.Metadata) == 0 {
		return nil, fmt.Errorf("Plex 媒体条目不存在: id=%s", id)
	}

	item := mapPlexMetadataToMediaItem(&mr.MediaContainer.Metadata[0])
	return item, nil
}

// SearchItems 搜索媒体条目（占位实现）
func (c *Client) SearchItems(ctx context.Context, query mediaserver.SearchQuery) ([]*mediaserver.MediaItem, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}

	q := url.Values{}
	if query.Keyword != "" {
		q.Set("query", query.Keyword)
	}
	// 使用 /search?query= 进行模糊搜索，然后从返回的 Metadata 里提取
	reqURL := c.buildURL("/search", q)
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
		return nil, fmt.Errorf("Plex 搜索失败: status=%d", resp.StatusCode)
	}

	var mr plexMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, fmt.Errorf("解析 Plex 搜索结果失败: %w", err)
	}

	items := make([]*mediaserver.MediaItem, 0, len(mr.MediaContainer.Metadata))
	for i := range mr.MediaContainer.Metadata {
		item := mapPlexMetadataToMediaItem(&mr.MediaContainer.Metadata[i])
		// 应用简单的类型/年份过滤（如果提供）
		if query.Type != "" && query.Type != mediaserver.MediaTypeUnknown {
			if item.Type != query.Type {
				continue
			}
		}
		if query.Year != nil && item.Year != nil && *item.Year != *query.Year {
			continue
		}
		items = append(items, item)
		if query.Limit > 0 && len(items) >= query.Limit {
			break
		}
	}

	return items, nil
}

// mapPlexMetadataToMediaItem 将 Plex Metadata 映射为统一的 MediaItem
func mapPlexMetadataToMediaItem(pm *plexMetadataItem) *mediaserver.MediaItem {
	item := &mediaserver.MediaItem{
		ID:           pm.RatingKey,
		Name:         pm.Title,
		OriginalName: pm.OriginalTitle,
		Type:         mediaserver.MediaTypeUnknown,
		ExternalIDs:  mediaserver.ExternalID{},
	}

	// 类型映射
	switch pm.Type {
	case "movie":
		item.Type = mediaserver.MediaTypeMovie
	case "show":
		item.Type = mediaserver.MediaTypeSeries
	case "episode":
		item.Type = mediaserver.MediaTypeEpisode
	}

	// 年份
	if pm.Year > 0 {
		year := pm.Year
		item.Year = &year
	}

	// 季/集
	if pm.ParentIndex > 0 {
		season := pm.ParentIndex
		item.Season = &season
	}
	if pm.Index > 0 {
		ep := pm.Index
		item.Episode = &ep
	}

	// GUID 示例："com.plexapp.agents.imdb://tt0133093?lang=en" 等
	// 这里简单解析常见的 imdb / tmdb / tvdb
	if pm.GUID != "" {
		if id, ok := parsePlexGUID(pm.GUID, "imdb"); ok {
			imdb := id
			item.ExternalIDs.IMDBID = &imdb
		}
		if id, ok := parsePlexGUID(pm.GUID, "tmdb"); ok {
			if n, err := strconv.Atoi(id); err == nil {
				item.ExternalIDs.TMDBID = &n
			}
		}
		if id, ok := parsePlexGUID(pm.GUID, "tvdb"); ok {
			if n, err := strconv.Atoi(id); err == nil {
				item.ExternalIDs.TVDBID = &n
			}
		}
	}

	return item
}

// parsePlexGUID 从 Plex GUID 中提取指定源的 ID（极简解析，占位实现）
func parsePlexGUID(guid, source string) (string, bool) {
	// 非严格解析：只做简单包含判断，真实实现可根据具体格式拆分
	if guid == "" {
		return "", false
	}
	// 例："com.plexapp.agents.imdb://tt0133093?lang=en"
	if source == "imdb" && (len(guid) > 0 && (strconv.IntSize > 0)) {
		// 这里先占位：真实实现可用 strings.Contains + 正则抽取
		// 当前仅返回 false，避免误解析
		return "", false
	}
	return "", false
}

// 编译期断言，确保 Client 实现了 MediaServerClient 接口
var _ mediaserver.MediaServerClient = (*Client)(nil)
