package tmdb

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

// Config TMDB 配置
type Config struct {
	APIKey  string
	BaseURL string // 可选，默认 https://api.themoviedb.org/3
	Timeout time.Duration
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

// Client TMDB 客户端骨架
type Client struct {
	apiKey  string
	baseURL *url.URL
	client  *http.Client
}

// tmdbMovieResult /search/movie 的单条结果
type tmdbMovieResult struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	ReleaseDate   string  `json:"release_date"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	VoteAverage   float64 `json:"vote_average"`
}

// tmdbSearchMovieResponse /search/movie 响应
type tmdbSearchMovieResponse struct {
	Page         int               `json:"page"`
	TotalResults int               `json:"total_results"`
	TotalPages   int               `json:"total_pages"`
	Results      []tmdbMovieResult `json:"results"`
}

// tmdbMovieDetail /movie/{id} 详情
type tmdbMovieDetail struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title"`
	Overview      string `json:"overview"`
	ReleaseDate   string `json:"release_date"`
	PosterPath    string `json:"poster_path"`
	BackdropPath  string `json:"backdrop_path"`
	IMDBID        string `json:"imdb_id"`
}

// tmdbTVResult /search/tv 的单条结果
type tmdbTVResult struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	Overview     string `json:"overview"`
	FirstAirDate string `json:"first_air_date"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

// tmdbSearchTVResponse /search/tv 响应
type tmdbSearchTVResponse struct {
	Page         int            `json:"page"`
	TotalResults int            `json:"total_results"`
	TotalPages   int            `json:"total_pages"`
	Results      []tmdbTVResult `json:"results"`
}

// tmdbTVDetail /tv/{id} 详情（简化）
type tmdbTVDetail struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	Overview     string `json:"overview"`
	FirstAirDate string `json:"first_air_date"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

// NewClient 创建 TMDB 客户端
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("TMDB APIKey 不能为空")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.themoviedb.org/3"
	}
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("解析 TMDB BaseURL 失败: %w", err)
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
	return metadata.ProviderTMDB
}

// SearchMovie 搜索电影（调用 /search/movie）
func (c *Client) SearchMovie(ctx context.Context, keyword string, opts metadata.SearchOptions) ([]*metadata.MovieInfo, error) {
	if keyword == "" {
		return []*metadata.MovieInfo{}, nil
	}

	q := url.Values{}
	q.Set("api_key", c.apiKey)
	q.Set("query", keyword)
	if opts.Language != "" {
		q.Set("language", string(opts.Language))
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Year > 0 {
		q.Set("year", strconv.Itoa(opts.Year))
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
		return nil, fmt.Errorf("TMDB 搜索电影失败: status=%d", resp.StatusCode)
	}

	var sr tmdbSearchMovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("解析 TMDB 搜索结果失败: %w", err)
	}

	results := make([]*metadata.MovieInfo, 0, len(sr.Results))
	for _, r := range sr.Results {
		results = append(results, mapTMDBMovieResultToMovieInfo(&r))
	}

	// 根据 opts.Limit 做简单截断
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// SearchTV 搜索剧集（占位实现）
func (c *Client) SearchTV(ctx context.Context, keyword string, opts metadata.SearchOptions) ([]*metadata.TVShowInfo, error) {
	if keyword == "" {
		return []*metadata.TVShowInfo{}, nil
	}

	q := url.Values{}
	q.Set("api_key", c.apiKey)
	q.Set("query", keyword)
	if opts.Language != "" {
		q.Set("language", string(opts.Language))
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Year > 0 {
		q.Set("first_air_date_year", strconv.Itoa(opts.Year))
	}

	reqURL := c.buildURL("/search/tv", q)
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
		return nil, fmt.Errorf("TMDB 搜索剧集失败: status=%d", resp.StatusCode)
	}

	var sr tmdbSearchTVResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("解析 TMDB 剧集搜索结果失败: %w", err)
	}

	results := make([]*metadata.TVShowInfo, 0, len(sr.Results))
	for _, r := range sr.Results {
		results = append(results, mapTMDBTVResultToTVShowInfo(&r))
	}

	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// GetMovieByTMDB 通过 TMDB ID 获取电影（占位实现）
func (c *Client) GetMovieByTMDB(ctx context.Context, tmdbID int, lang metadata.Language) (*metadata.MovieInfo, error) {
	if tmdbID <= 0 {
		return nil, fmt.Errorf("无效的 TMDB ID: %d", tmdbID)
	}

	q := url.Values{}
	q.Set("api_key", c.apiKey)
	if lang != "" {
		q.Set("language", string(lang))
	}

	path := "/movie/" + strconv.Itoa(tmdbID)
	reqURL := c.buildURL(path, q)
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
		return nil, fmt.Errorf("TMDB 获取电影详情失败: status=%d", resp.StatusCode)
	}

	var md tmdbMovieDetail
	if err := json.NewDecoder(resp.Body).Decode(&md); err != nil {
		return nil, fmt.Errorf("解析 TMDB 电影详情失败: %w", err)
	}

	info := mapTMDBMovieDetailToMovieInfo(&md)
	return info, nil
}

// GetTVByTMDB 通过 TMDB ID 获取剧集（占位实现）
func (c *Client) GetTVByTMDB(ctx context.Context, tmdbID int, lang metadata.Language) (*metadata.TVShowInfo, error) {
	if tmdbID <= 0 {
		return nil, fmt.Errorf("无效的 TMDB ID: %d", tmdbID)
	}

	q := url.Values{}
	q.Set("api_key", c.apiKey)
	if lang != "" {
		q.Set("language", string(lang))
	}

	path := "/tv/" + strconv.Itoa(tmdbID)
	reqURL := c.buildURL(path, q)
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
		return nil, fmt.Errorf("TMDB 获取剧集详情失败: status=%d", resp.StatusCode)
	}

	var td tmdbTVDetail
	if err := json.NewDecoder(resp.Body).Decode(&td); err != nil {
		return nil, fmt.Errorf("解析 TMDB 剧集详情失败: %w", err)
	}

	info := mapTMDBTVDetailToTVShowInfo(&td)
	return info, nil
}

// GetMovieByID 通过本方ID获取电影（占位实现）
func (c *Client) GetMovieByID(ctx context.Context, id string, lang metadata.Language) (*metadata.MovieInfo, error) {
	// 对于 TMDB，通常 id 即 TMDB ID，可复用 GetMovieByTMDB
	tmdbID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("无效的 TMDB 电影ID: %s", id)
	}
	return c.GetMovieByTMDB(ctx, tmdbID, lang)
}

// GetTVByID 通过本方ID获取剧集（占位实现）
func (c *Client) GetTVByID(ctx context.Context, id string, lang metadata.Language) (*metadata.TVShowInfo, error) {
	// 对于 TMDB，通常 id 即 TMDB ID，可复用 GetTVByTMDB
	tmdbID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("无效的 TMDB 剧集ID: %s", id)
	}
	return c.GetTVByTMDB(ctx, tmdbID, lang)
}

// mapTMDBMovieResultToMovieInfo 将搜索结果映射为 MovieInfo
func mapTMDBMovieResultToMovieInfo(r *tmdbMovieResult) *metadata.MovieInfo {
	info := &metadata.MovieInfo{
		ID:       strconv.Itoa(r.ID),
		Provider: metadata.ProviderTMDB,
		Title:    r.Title,
		Original: r.OriginalTitle,
		Overview: r.Overview,
	}

	// 年份从 release_date 截取前四位
	if len(r.ReleaseDate) >= 4 {
		if year, err := strconv.Atoi(r.ReleaseDate[:4]); err == nil {
			info.Year = year
		}
	}

	// 图片 URL 只保留 path，完整 URL 由上层拼接（CDN基址可能配置化）
	info.PosterURL = r.PosterPath
	info.BackdropURL = r.BackdropPath

	// TMDB 自身 ID
	if r.ID > 0 {
		id := r.ID
		info.TMDBID = &id
	}

	return info
}

// mapTMDBTVResultToTVShowInfo 将剧集搜索结果映射为 TVShowInfo
func mapTMDBTVResultToTVShowInfo(r *tmdbTVResult) *metadata.TVShowInfo {
	info := &metadata.TVShowInfo{
		ID:       strconv.Itoa(r.ID),
		Provider: metadata.ProviderTMDB,
		Title:    r.Name,
		Original: r.OriginalName,
		Overview: r.Overview,
	}
	if len(r.FirstAirDate) >= 4 {
		if year, err := strconv.Atoi(r.FirstAirDate[:4]); err == nil {
			info.Year = year
		}
	}
	info.PosterURL = r.PosterPath
	info.PosterURL = r.PosterPath
	info.PosterURL = r.PosterPath
	if r.ID > 0 {
		id := r.ID
		info.TMDBID = &id
	}
	return info
}

// mapTMDBTVDetailToTVShowInfo 将剧集详情映射为 TVShowInfo
func mapTMDBTVDetailToTVShowInfo(td *tmdbTVDetail) *metadata.TVShowInfo {
	info := &metadata.TVShowInfo{
		ID:       strconv.Itoa(td.ID),
		Provider: metadata.ProviderTMDB,
		Title:    td.Name,
		Original: td.OriginalName,
		Overview: td.Overview,
	}
	if len(td.FirstAirDate) >= 4 {
		if year, err := strconv.Atoi(td.FirstAirDate[:4]); err == nil {
			info.Year = year
		}
	}
	info.PosterURL = td.PosterPath
	info.PosterURL = td.PosterPath
	info.PosterURL = td.PosterPath
	if td.ID > 0 {
		id := td.ID
		info.TMDBID = &id
	}
	return info
}

// mapTMDBMovieDetailToMovieInfo 将详情映射为 MovieInfo
func mapTMDBMovieDetailToMovieInfo(md *tmdbMovieDetail) *metadata.MovieInfo {
	info := &metadata.MovieInfo{
		ID:       strconv.Itoa(md.ID),
		Provider: metadata.ProviderTMDB,
		Title:    md.Title,
		Original: md.OriginalTitle,
		Overview: md.Overview,
	}

	if len(md.ReleaseDate) >= 4 {
		if year, err := strconv.Atoi(md.ReleaseDate[:4]); err == nil {
			info.Year = year
		}
	}

	info.PosterURL = md.PosterPath
	info.BackdropURL = md.BackdropPath

	if md.ID > 0 {
		id := md.ID
		info.TMDBID = &id
	}
	if md.IMDBID != "" {
		id := md.IMDBID
		info.IMDBID = &id
	}

	return info
}

// 编译期断言，确保实现 MetadataProvider 接口
var _ metadata.MetadataProvider = (*Client)(nil)
