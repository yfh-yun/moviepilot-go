package tvdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"moviepilot-go/internal/integration/metadata"
)

// Config TVDB 配置
type Config struct {
	APIKey   string
	BaseURL  string // 默认 https://api4.thetvdb.com/v4
	Timeout  time.Duration
	JWTToken string // 部分 API 需要登录获取 token，这里预留
}

// Client TVDB 客户端骨架
type Client struct {
	apiKey  string
	baseURL *url.URL
	client  *http.Client
}

// buildURL 构造带路径的完整 URL
func (c *Client) buildURL(path string, q url.Values) string {
	u := *c.baseURL
	u.Path = path
	if q != nil {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// tvdbSearchResult TVDB 剧集搜索结果（简化）
type tvdbSearchResult struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Overview   string `json:"overview"`
	FirstAired string `json:"firstAired"`
}

// tvdbSearchResponse v4 /search 统一响应（简化）
type tvdbSearchResponse struct {
	Data []tvdbSearchResult `json:"data"`
}

// tvdbSeriesDetail 剧集详情（简化）
type tvdbSeriesDetail struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Overview   string `json:"overview"`
	FirstAired string `json:"firstAired"`
}

// NewClient 创建 TVDB 客户端
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api4.thetvdb.com/v4"
	}
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 TVDB BaseURL 失败: %w", err)
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: parsed,
		client:  &http.Client{Timeout: cfg.Timeout},
	}, nil
}

// Name 实现 MetadataProvider.Name
func (c *Client) Name() metadata.ProviderName {
	return metadata.ProviderTVDB
}

// SearchMovie TVDB 主要是剧集平台，这里返回空
func (c *Client) SearchMovie(ctx context.Context, keyword string, opts metadata.SearchOptions) ([]*metadata.MovieInfo, error) {
	// TODO: TVDB 电影支持可选实现
	return []*metadata.MovieInfo{}, nil
}

// SearchTV 搜索剧集（最小实现：调用 /search）
func (c *Client) SearchTV(ctx context.Context, keyword string, opts metadata.SearchOptions) ([]*metadata.TVShowInfo, error) {
	if keyword == "" {
		return []*metadata.TVShowInfo{}, nil
	}

	q := url.Values{}
	q.Set("q", keyword)
	if opts.Year > 0 {
		q.Set("year", strconv.Itoa(opts.Year))
	}

	reqURL := c.buildURL("/search", q)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	// v4 需要授权，这里仅留出 header 占位，后续可接入 JWT/token
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TVDB 搜索剧集失败: status=%d", resp.StatusCode)
	}

	var sr tvdbSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("解析 TVDB 搜索结果失败: %w", err)
	}

	shows := make([]*metadata.TVShowInfo, 0, len(sr.Data))
	for _, r := range sr.Data {
		shows = append(shows, mapTVDBSearchResultToTVShowInfo(&r))
	}

	if opts.Limit > 0 && len(shows) > opts.Limit {
		shows = shows[:opts.Limit]
	}

	return shows, nil
}

// GetMovieByTMDB 通过 TMDB ID 获取电影（占位实现）
func (c *Client) GetMovieByTMDB(ctx context.Context, tmdbID int, lang metadata.Language) (*metadata.MovieInfo, error) {
	return nil, fmt.Errorf("TVDB 暂不实现 GetMovieByTMDB")
}

// GetTVByTMDB 通过 TMDB ID 获取剧集（占位实现）
func (c *Client) GetTVByTMDB(ctx context.Context, tmdbID int, lang metadata.Language) (*metadata.TVShowInfo, error) {
	// 最小实现：尝试按 TMDB ID 搜索，再调用 GetTVByID
	if tmdbID <= 0 {
		return nil, fmt.Errorf("无效的 TMDB ID: %d", tmdbID)
	}

	// 这里只提供占位逻辑：直接返回 "未实现"，以免误导
	// 后续可接入 TVDB 对 TMDB 映射的专用 API
	return nil, fmt.Errorf("TVDB GetTVByTMDB 暂未实现映射逻辑")
}

// GetMovieByID 通过本方ID获取电影（占位实现）
func (c *Client) GetMovieByID(ctx context.Context, id string, lang metadata.Language) (*metadata.MovieInfo, error) {
	return nil, fmt.Errorf("TVDB 暂不实现 GetMovieByID")
}

// GetTVByID 通过本方ID获取剧集（最小实现：调用 /series/{id}）
func (c *Client) GetTVByID(ctx context.Context, id string, lang metadata.Language) (*metadata.TVShowInfo, error) {
	if id == "" {
		return nil, fmt.Errorf("TVDB 剧集ID 不能为空")
	}

	path := "/series/" + id
	reqURL := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TVDB 获取剧集详情失败: status=%d", resp.StatusCode)
	}

	var sd tvdbSeriesDetail
	if err := json.NewDecoder(resp.Body).Decode(&sd); err != nil {
		return nil, fmt.Errorf("解析 TVDB 剧集详情失败: %w", err)
	}

	show := mapTVDBSeriesDetailToTVShowInfo(&sd)
	return show, nil
}

// 编译期断言
var _ metadata.MetadataProvider = (*Client)(nil)

// mapTVDBSearchResultToTVShowInfo 将搜索结果映射为统一 TVShowInfo
func mapTVDBSearchResultToTVShowInfo(r *tvdbSearchResult) *metadata.TVShowInfo {
	info := &metadata.TVShowInfo{
		ID:       strconv.Itoa(r.ID),
		Provider: metadata.ProviderTVDB,
		Title:    r.Name,
		Overview: r.Overview,
	}
	if len(r.FirstAired) >= 4 {
		if year, err := strconv.Atoi(r.FirstAired[:4]); err == nil {
			info.Year = year
		}
	}
	if r.ID > 0 {
		id := r.ID
		info.TVDBID = &id
	}
	return info
}

// mapTVDBSeriesDetailToTVShowInfo 将详情映射为统一 TVShowInfo
func mapTVDBSeriesDetailToTVShowInfo(sd *tvdbSeriesDetail) *metadata.TVShowInfo {
	info := &metadata.TVShowInfo{
		ID:       strconv.Itoa(sd.ID),
		Provider: metadata.ProviderTVDB,
		Title:    sd.Name,
		Overview: sd.Overview,
	}
	if len(sd.FirstAired) >= 4 {
		if year, err := strconv.Atoi(sd.FirstAired[:4]); err == nil {
			info.Year = year
		}
	}
	if sd.ID > 0 {
		id := sd.ID
		info.TVDBID = &id
	}
	return info
}
