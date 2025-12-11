package douban

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

// Config 豆瓣配置（可能通过自建代理，因为官方 API 受限）
type Config struct {
	BaseURL string // 可能是自建代理地址
	Timeout time.Duration
}

// Client 豆瓣元数据客户端骨架
type Client struct {
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

// doubanMovieSearchResult 豆瓣电影搜索结果（简化，假设代理返回）
type doubanMovieSearchResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Year  string `json:"year"`
	Rate  string `json:"rate"`
}

// doubanSearchResponse 豆瓣搜索响应（简化）
type doubanSearchResponse struct {
	Subjects []doubanMovieSearchResult `json:"subjects"`
}

// doubanMovieDetail 豆瓣电影详情（简化）
type doubanMovieDetail struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Year    string `json:"year"`
	Summary string `json:"summary"`
	Rating  struct {
		Average float64 `json:"average"`
	} `json:"rating"`
}

// NewClient 创建豆瓣客户端
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("Douban BaseURL 不能为空")
	}
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 Douban BaseURL 失败: %w", err)
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	return &Client{
		baseURL: parsed,
		client:  &http.Client{Timeout: cfg.Timeout},
	}, nil
}

// Name 实现 MetadataProvider.Name
func (c *Client) Name() metadata.ProviderName {
	return metadata.ProviderDouban
}

// SearchMovie 搜索电影（最小实现：假设代理提供 /search/movie）
func (c *Client) SearchMovie(ctx context.Context, keyword string, opts metadata.SearchOptions) ([]*metadata.MovieInfo, error) {
	if keyword == "" {
		return []*metadata.MovieInfo{}, nil
	}

	q := url.Values{}
	q.Set("q", keyword)
	if opts.Limit > 0 {
		q.Set("count", strconv.Itoa(opts.Limit))
	}

	reqURL := c.buildURL("/search/movie", q)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("豆瓣搜索电影失败: status=%d", resp.StatusCode)
	}

	var sr doubanSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("解析豆瓣搜索结果失败: %w", err)
	}

	movies := make([]*metadata.MovieInfo, 0, len(sr.Subjects))
	for _, r := range sr.Subjects {
		movies = append(movies, mapDoubanSearchResultToMovieInfo(&r))
	}

	return movies, nil
}

// SearchTV 搜索剧集（占位实现）
func (c *Client) SearchTV(ctx context.Context, keyword string, opts metadata.SearchOptions) ([]*metadata.TVShowInfo, error) {
	// TODO: 调用代理包装的豆瓣搜索 API
	return []*metadata.TVShowInfo{}, nil
}

// GetMovieByTMDB 通过 TMDB ID 获取电影（占位实现）
func (c *Client) GetMovieByTMDB(ctx context.Context, tmdbID int, lang metadata.Language) (*metadata.MovieInfo, error) {
	return nil, fmt.Errorf("Douban 不支持通过 TMDB 直接查询")
}

// GetTVByTMDB 通过 TMDB ID 获取剧集（占位实现）
func (c *Client) GetTVByTMDB(ctx context.Context, tmdbID int, lang metadata.Language) (*metadata.TVShowInfo, error) {
	return nil, fmt.Errorf("Douban 不支持通过 TMDB 直接查询")
}

// GetMovieByID 通过本方ID获取电影（最小实现：假设代理提供 /movie/{id}）
func (c *Client) GetMovieByID(ctx context.Context, id string, lang metadata.Language) (*metadata.MovieInfo, error) {
	if id == "" {
		return nil, fmt.Errorf("豆瓣电影ID 不能为空")
	}

	path := "/movie/" + id
	reqURL := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("豆瓣获取电影详情失败: status=%d", resp.StatusCode)
	}

	var detail doubanMovieDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("解析豆瓣电影详情失败: %w", err)
	}

	movie := mapDoubanDetailToMovieInfo(&detail)
	return movie, nil
}

// GetTVByID 通过本方ID获取剧集（占位实现）
func (c *Client) GetTVByID(ctx context.Context, id string, lang metadata.Language) (*metadata.TVShowInfo, error) {
	// TODO: 调用 /tv/{id} 或代理接口
	return nil, fmt.Errorf("GetTVByID 未实现")
}

// 编译期断言
var _ metadata.MetadataProvider = (*Client)(nil)

// mapDoubanSearchResultToMovieInfo 将搜索结果映射为统一 MovieInfo
func mapDoubanSearchResultToMovieInfo(r *doubanMovieSearchResult) *metadata.MovieInfo {
	info := &metadata.MovieInfo{
		ID:       r.ID,
		Provider: metadata.ProviderDouban,
		Title:    r.Title,
	}
	if r.Year != "" {
		if year, err := strconv.Atoi(r.Year); err == nil {
			info.Year = year
		}
	}
	// 豆瓣评分暂不映射（MovieInfo 中无 Rating 字段）
	_ = r.Rate
	return info
}

// mapDoubanDetailToMovieInfo 将详情映射为统一 MovieInfo
func mapDoubanDetailToMovieInfo(d *doubanMovieDetail) *metadata.MovieInfo {
	info := &metadata.MovieInfo{
		ID:       d.ID,
		Provider: metadata.ProviderDouban,
		Title:    d.Title,
		Overview: d.Summary,
	}
	// 豆瓣评分暂不映射（MovieInfo 中无 Rating 字段）
	_ = d.Rating.Average
	if d.Year != "" {
		if year, err := strconv.Atoi(d.Year); err == nil {
			info.Year = year
		}
	}
	return info
}
