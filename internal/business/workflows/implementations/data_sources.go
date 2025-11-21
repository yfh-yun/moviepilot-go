// Package implementations 提供数据源实现
package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"moviepilot-go/internal/business/workflows/types"
	"moviepilot-go/pkg/logger"
)

// TMDBDataSource TMDB数据源
type TMDBDataSource struct {
	client    *http.Client
	apiKey    string
	baseURL   string
	language  string
	available bool
	lastCheck time.Time
}

// NewTMDBDataSource 创建TMDB数据源
func NewTMDBDataSource(client *http.Client) *TMDBDataSource {
	return &TMDBDataSource{
		client:   client,
		baseURL:  "https://api.themoviedb.org/3",
		apiKey:   "your_tmdb_api_key", // 应该从配置中获取
		language: "zh-CN",
	}
}

// Name 返回数据源名称
func (ds *TMDBDataSource) Name() string {
	return "tmdb"
}

// IsAvailable 检查数据源可用性
func (ds *TMDBDataSource) IsAvailable(ctx context.Context) bool {
	// 缓存检查结果5分钟
	if time.Since(ds.lastCheck) < 5*time.Minute {
		return ds.available
	}

	// 简单的健康检查
	req, err := http.NewRequestWithContext(ctx, "GET", ds.baseURL+"/configuration", nil)
	if err != nil {
		ds.available = false
		ds.lastCheck = time.Now()
		return false
	}

	q := req.URL.Query()
	q.Set("api_key", ds.apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := ds.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		ds.available = false
		ds.lastCheck = time.Now()
		return false
	}
	defer resp.Body.Close()

	ds.available = true
	ds.lastCheck = time.Now()
	return true
}

// GetPriority 返回数据源优先级
func (ds *TMDBDataSource) GetPriority() int {
	return 1 // 最高优先级
}

// Fetch 获取媒体数据
func (ds *TMDBDataSource) Fetch(ctx context.Context, params *FetchParams) ([]*types.MediaInfo, error) {
	var medias []*types.MediaInfo

	// 搜索电影
	if params.SourceType == "movie" || params.SourceType == "" {
		movies, err := ds.searchMovies(ctx, params)
		if err != nil {
			logger.Warn("Failed to search movies from TMDB", "error", err)
		} else {
			medias = append(medias, movies...)
		}
	}

	// 搜索电视剧
	if params.SourceType == "tv" || params.SourceType == "" {
		tvShows, err := ds.searchTVShows(ctx, params)
		if err != nil {
			logger.Warn("Failed to search TV shows from TMDB", "error", err)
		} else {
			medias = append(medias, tvShows...)
		}
	}

	return medias, nil
}

// searchMovies 搜索电影
func (ds *TMDBDataSource) searchMovies(ctx context.Context, params *FetchParams) ([]*types.MediaInfo, error) {
	endpoint := ds.baseURL + "/search/movie"
	reqParams := url.Values{
		"api_key":  {ds.apiKey},
		"language": {ds.language},
	}

	if params.Keywords != "" {
		reqParams.Set("query", params.Keywords)
	}
	if params.Year > 0 {
		reqParams.Set("year", strconv.Itoa(params.Year))
	}
	if params.Limit > 0 {
		reqParams.Set("limit", strconv.Itoa(params.Limit))
	}

	return ds.makeRequest(ctx, endpoint, reqParams, "movie")
}

// searchTVShows 搜索电视剧
func (ds *TMDBDataSource) searchTVShows(ctx context.Context, params *FetchParams) ([]*types.MediaInfo, error) {
	endpoint := ds.baseURL + "/search/tv"
	reqParams := url.Values{
		"api_key":  {ds.apiKey},
		"language": {ds.language},
	}

	if params.Keywords != "" {
		reqParams.Set("query", params.Keywords)
	}
	if params.Limit > 0 {
		reqParams.Set("limit", strconv.Itoa(params.Limit))
	}

	return ds.makeRequest(ctx, endpoint, reqParams, "tv")
}

// makeRequest 发起HTTP请求
func (ds *TMDBDataSource) makeRequest(ctx context.Context, endpoint string, params url.Values, mediaType string) ([]*types.MediaInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = params.Encode()

	resp, err := ds.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API error: %d", resp.StatusCode)
	}

	var response struct {
		Results []struct {
			ID            int      `json:"id"`
			Title         string   `json:"title"`
			Name          string   `json:"name"`
			OriginalTitle string   `json:"original_title"`
			OriginalName  string   `json:"original_name"`
			Overview      string   `json:"overview"`
			ReleaseDate   string   `json:"release_date"`
			FirstAirDate  string   `json:"first_air_date"`
			VoteAverage   float64  `json:"vote_average"`
			VoteCount     int      `json:"vote_count"`
			Popularity    float64  `json:"popularity"`
			PosterPath    string   `json:"poster_path"`
			BackdropPath  string   `json:"backdrop_path"`
			GenreIDs      []int    `json:"genre_ids"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var medias []*types.MediaInfo
	for _, result := range response.Results {
		media := &types.MediaInfo{
			TMDBID:         result.ID,
			Type:           mediaType,
			Title:          ds.getTitle(result.Title, result.Name),
			OriginalTitle:  ds.getTitle(result.OriginalTitle, result.OriginalName),
			Overview:       result.Overview,
			Rating:         result.VoteAverage,
			VoteCount:      result.VoteCount,
			Popularity:     result.Popularity,
			PosterPath:     result.PosterPath,
			BackdropPath:   result.BackdropPath,
			GenreIDs:       result.GenreIDs,
			Source:         "tmdb",
		}

		// 解析年份
		date := ds.getDate(result.ReleaseDate, result.FirstAirDate)
		if len(date) >= 4 {
			if year, err := strconv.Atoi(date[:4]); err == nil {
				media.Year = year
			}
		}

		// 构建完整URL
		if media.PosterPath != "" {
			media.PosterURL = "https://image.tmdb.org/t/p/w500" + media.PosterPath
		}
		if media.BackdropPath != "" {
			media.BackdropURL = "https://image.tmdb.org/t/p/w1280" + media.BackdropPath
		}

		medias = append(medias, media)
	}

	return medias, nil
}

// getTitle 获取标题
func (ds *TMDBDataSource) getTitle(title, name string) string {
	if title != "" {
		return title
	}
	return name
}

// getDate 获取日期
func (ds *TMDBDataSource) getDate(releaseDate, firstAirDate string) string {
	if releaseDate != "" {
		return releaseDate
	}
	return firstAirDate
}

// DoubanDataSource 豆瓣数据源
type DoubanDataSource struct {
	client    *http.Client
	baseURL   string
	available bool
	lastCheck time.Time
}

// NewDoubanDataSource 创建豆瓣数据源
func NewDoubanDataSource(client *http.Client) *DoubanDataSource {
	return &DoubanDataSource{
		client:  client,
		baseURL: "https://api.douban.com/v2",
	}
}

// Name 返回数据源名称
func (ds *DoubanDataSource) Name() string {
	return "douban"
}

// IsAvailable 检查数据源可用性
func (ds *DoubanDataSource) IsAvailable(ctx context.Context) bool {
	if time.Since(ds.lastCheck) < 5*time.Minute {
		return ds.available
	}

	// 豆瓣API可能需要特殊处理，这里简化为总是可用
	ds.available = true
	ds.lastCheck = time.Now()
	return true
}

// GetPriority 返回数据源优先级
func (ds *DoubanDataSource) GetPriority() int {
	return 2
}

// Fetch 获取媒体数据
func (ds *DoubanDataSource) Fetch(ctx context.Context, params *FetchParams) ([]*types.MediaInfo, error) {
	// 豆瓣API实现
	// 注意：实际豆瓣API可能需要更复杂的认证和反爬虫处理
	return []*types.MediaInfo{}, nil
}

// OMDBDataSource OMDB数据源
type OMDBDataSource struct {
	client    *http.Client
	apiKey    string
	baseURL   string
	available bool
	lastCheck time.Time
}

// NewOMDBDataSource 创建OMDB数据源
func NewOMDBDataSource(client *http.Client) *OMDBDataSource {
	return &OMDBDataSource{
		client:  client,
		baseURL: "http://www.omdbapi.com",
		apiKey:  "your_omdb_api_key", // 应该从配置中获取
	}
}

// Name 返回数据源名称
func (ds *OMDBDataSource) Name() string {
	return "omdb"
}

// IsAvailable 检查数据源可用性
func (ds *OMDBDataSource) IsAvailable(ctx context.Context) bool {
	if time.Since(ds.lastCheck) < 5*time.Minute {
		return ds.available
	}

	reqParams := url.Values{
		"apikey": {ds.apiKey},
		"s":      {"test"}, // 测试搜索
	}

	req, err := http.NewRequestWithContext(ctx, "GET", ds.baseURL, nil)
	if err != nil {
		ds.available = false
		ds.lastCheck = time.Now()
		return false
	}

	req.URL.RawQuery = reqParams.Encode()

	resp, err := ds.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		ds.available = false
		ds.lastCheck = time.Now()
		return false
	}
	defer resp.Body.Close()

	ds.available = true
	ds.lastCheck = time.Now()
	return true
}

// GetPriority 返回数据源优先级
func (ds *OMDBDataSource) GetPriority() int {
	return 3
}

// Fetch 获取媒体数据
func (ds *OMDBDataSource) Fetch(ctx context.Context, params *FetchParams) ([]*types.MediaInfo, error) {
	reqParams := url.Values{
		"apikey": {ds.apiKey},
	}

	if params.Keywords != "" {
		reqParams.Set("s", params.Keywords)
	}
	if params.Year > 0 {
		reqParams.Set("y", strconv.Itoa(params.Year))
	}
	if params.Limit > 0 {
		reqParams.Set("page", "1")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", ds.baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = reqParams.Encode()

	resp, err := ds.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OMDB API error: %d", resp.StatusCode)
	}

	var response struct {
		Search []struct {
			Title  string `json:"Title"`
			Year   string `json:"Year"`
			imdbID string `json:"imdbID"`
			Type   string `json:"Type"`
			Poster string `json:"Poster"`
		} `json:"Search"`
		TotalResults string `json:"totalResults"`
		Response     string `json:"Response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.Response != "True" {
		return nil, fmt.Errorf("OMDB API returned false")
	}

	var medias []*types.MediaInfo
	for _, result := range response.Search {
		media := &types.MediaInfo{
			Title:     result.Title,
			IMDBID:    result.imdbID,
			Type:      strings.ToLower(result.Type),
			PosterURL: result.Poster,
			Source:    "omdb",
		}

		// 解析年份
		if year, err := strconv.Atoi(result.Year); err == nil {
			media.Year = year
		}

		medias = append(medias, media)
	}

	return medias, nil
}