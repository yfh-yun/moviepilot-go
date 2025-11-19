package tvdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/core/config"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/httpclient"
)

// Client TVDB API客户端
type Client struct {
	httpClient  *httpclient.Client
	baseURL     string
	apiKey      string
	pin         string
	accessToken string
	mu          sync.RWMutex
	logger      *logger.Logger
	cache       *Cache
}

// AuthResponse 认证响应
type AuthResponse struct {
	Status string `json:"status"`
	Data   struct {
		Token string `json:"token"`
	} `json:"data"`
}

// SeriesResponse 剧集响应
type SeriesResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Slug       string `json:"slug"`
		Status     string `json:"status"`
		FirstAired string `json:"firstAired"`
		LastAired  string `json:"lastAired"`
		Overview   string `json:"overview"`
		Runtime    int    `json:"runtime"`
		Network    string `json:"network"`
		Banner     string `json:"banner"`
		Poster     string `json:"poster"`
		Fanart     string `json:"fanart"`
		IMDBID     string `json:"imdbId"`
		Zap2itID   string `json:"zap2itId"`
	} `json:"data"`
}

// EpisodeResponse 剧集响应
type EpisodeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID          int    `json:"id"`
		SeriesID    int    `json:"seriesId"`
		Season      int    `json:"seasonNumber"`
		Episode     int    `json:"episodeNumber"`
		Title       string `json:"episodeName"`
		Overview    string `json:"overview"`
		FirstAired  string `json:"firstAired"`
		Runtime     int    `json:"runtime"`
		Thumbnail   string `json:"filename"`
		Rating      string `json:"siteRating"`
		RatingCount int    `json:"siteRatingCount"`
	} `json:"data"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Status string `json:"status"`
	Data   []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Slug       string `json:"slug"`
		Status     string `json:"status"`
		FirstAired string `json:"firstAired"`
		Network    string `json:"network"`
		IMDBID     string `json:"imdbId"`
		Overview   string `json:"overview"`
		Banner     string `json:"banner"`
		Poster     string `json:"poster"`
	} `json:"data"`
}

// Series 剧集信息
type Series struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	Status     string    `json:"status"`
	FirstAired string    `json:"firstAired"`
	LastAired  string    `json:"lastAired"`
	Overview   string    `json:"overview"`
	Runtime    int       `json:"runtime"`
	Network    string    `json:"network"`
	Banner     string    `json:"banner"`
	Poster     string    `json:"poster"`
	Fanart     string    `json:"fanart"`
	IMDBID     string    `json:"imdbId"`
	Zap2itID   string    `json:"zap2itId"`
	Episodes   []Episode `json:"episodes"`
}

// Episode 剧集信息
type Episode struct {
	ID          int    `json:"id"`
	SeriesID    int    `json:"seriesId"`
	Season      int    `json:"seasonNumber"`
	Episode     int    `json:"episodeNumber"`
	Title       string `json:"title"`
	Overview    string `json:"overview"`
	FirstAired  string `json:"firstAired"`
	Runtime     int    `json:"runtime"`
	Thumbnail   string `json:"thumbnail"`
	Rating      string `json:"rating"`
	RatingCount int    `json:"ratingCount"`
}

// NewClient 创建TVDB客户端
func NewClient(cfg *config.Config) *Client {
	httpClient := httpclient.NewClient(&httpclient.Config{
		BaseURL:   "https://api4.thetvdb.com/v4/",
		Timeout:   30 * time.Second,
		UserAgent: "MoviePilot/1.0",
	})

	client := &Client{
		httpClient: httpClient,
		baseURL:    "https://api4.thetvdb.com/v4/",
		apiKey:     cfg.TVDB.APIKey,
		pin:        cfg.TVDB.PIN,
		cache:      NewCache(24*time.Hour, 1000),
		logger:     logger.NewLogger("tvdb"),
	}

	// 异步初始化认证
	go client.authenticate()

	return client
}

// authenticate 认证并获取访问令牌
func (c *Client) authenticate() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" {
		return nil
	}

	payload := map[string]string{
		"apikey": c.apiKey,
		"pin":    c.pin,
	}

	resp, err := c.httpClient.Post(context.Background(), "login", nil, payload)
	if err != nil {
		return fmt.Errorf("认证请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("认证失败，状态码: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("解析认证响应失败: %w", err)
	}

	if authResp.Status != "success" {
		return fmt.Errorf("认证失败: %s", authResp.Status)
	}

	c.accessToken = authResp.Data.Token
	return nil
}

// getHeaders 获取认证头
func (c *Client) getHeaders() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if c.accessToken != "" {
		headers["Authorization"] = "Bearer " + c.accessToken
	}

	return headers
}

// GetSeries 获取剧集信息
func (c *Client) GetSeries(ctx context.Context, seriesID int) (*Series, error) {
	// 检查缓存
	if cached, found := c.cache.GetSeries(seriesID); found {
		return cached, nil
	}

	// 确保认证
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%sseries/%d", c.baseURL, seriesID)
	headers := c.getHeaders()

	resp, err := c.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("获取剧集信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, HandleHTTPError(resp.StatusCode)
	}

	var seriesResp SeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&seriesResp); err != nil {
		return nil, fmt.Errorf("解析剧集响应失败: %w", err)
	}

	series := c.parseSeries(seriesResp)

	// 获取剧集列表
	if episodes, err := c.GetEpisodes(ctx, seriesID); err == nil {
		series.Episodes = episodes
	}

	// 缓存结果
	c.cache.SetSeries(seriesID, series)

	return series, nil
}

// GetEpisodes 获取剧集列表
func (c *Client) GetEpisodes(ctx context.Context, seriesID int) ([]Episode, error) {
	// 检查缓存
	if cached, found := c.cache.GetEpisodes(seriesID); found {
		return cached, nil
	}

	// 确保认证
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%sseries/%d/episodes/default", c.baseURL, seriesID)
	headers := c.getHeaders()

	var episodes []Episode
	page := 1

	for {
		pageURL := fmt.Sprintf("%s?page=%d", url, page)
		resp, err := c.httpClient.Get(ctx, pageURL, headers)
		if err != nil {
			return nil, fmt.Errorf("获取剧集列表失败: %w", err)
		}

		var episodesResp struct {
			Status string `json:"status"`
			Data   []struct {
				ID          int    `json:"id"`
				SeriesID    int    `json:"seriesId"`
				Season      int    `json:"seasonNumber"`
				Episode     int    `json:"episodeNumber"`
				Title       string `json:"episodeName"`
				Overview    string `json:"overview"`
				FirstAired  string `json:"firstAired"`
				Runtime     int    `json:"runtime"`
				Thumbnail   string `json:"filename"`
				Rating      string `json:"siteRating"`
				RatingCount int    `json:"siteRatingCount"`
			} `json:"data"`
			Links struct {
				Next int `json:"next"`
				Last int `json:"last"`
			} `json:"links"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&episodesResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("解析剧集列表响应失败: %w", err)
		}
		resp.Body.Close()

		// 解析当前页的剧集
		for _, ep := range episodesResp.Data {
			episode := Episode{
				ID:          ep.ID,
				SeriesID:    ep.SeriesID,
				Season:      ep.Season,
				Episode:     ep.Episode,
				Title:       ep.Title,
				Overview:    ep.Overview,
				FirstAired:  ep.FirstAired,
				Runtime:     ep.Runtime,
				Thumbnail:   ep.Thumbnail,
				Rating:      ep.Rating,
				RatingCount: ep.RatingCount,
			}
			episodes = append(episodes, episode)
		}

		// 检查是否有下一页
		if episodesResp.Links.Next == 0 || page >= episodesResp.Links.Last {
			break
		}
		page = episodesResp.Links.Next
	}

	// 缓存结果
	c.cache.SetEpisodes(seriesID, episodes)

	return episodes, nil
}

// SearchSeries 搜索剧集
func (c *Client) SearchSeries(ctx context.Context, query string) ([]Series, error) {
	// 检查缓存
	if cached, found := c.cache.GetSearch(query); found {
		return cached, nil
	}

	// 确保认证
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%ssearch/series?name=%s", c.baseURL, query)
	headers := c.getHeaders()

	resp, err := c.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("搜索剧集失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, HandleHTTPError(resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("解析搜索响应失败: %w", err)
	}

	series := c.parseSearchResults(searchResp)

	// 缓存结果
	c.cache.SetSearch(query, series)

	return series, nil
}

// parseSeries 解析剧集响应
func (c *Client) parseSeries(resp SeriesResponse) *Series {
	return &Series{
		ID:         resp.Data.ID,
		Name:       resp.Data.Name,
		Slug:       resp.Data.Slug,
		Status:     resp.Data.Status,
		FirstAired: resp.Data.FirstAired,
		LastAired:  resp.Data.LastAired,
		Overview:   resp.Data.Overview,
		Runtime:    resp.Data.Runtime,
		Network:    resp.Data.Network,
		Banner:     resp.Data.Banner,
		Poster:     resp.Data.Poster,
		Fanart:     resp.Data.Fanart,
		IMDBID:     resp.Data.IMDBID,
		Zap2itID:   resp.Data.Zap2itID,
	}
}

// parseSearchResults 解析搜索结果
func (c *Client) parseSearchResults(resp SearchResponse) []Series {
	var series []Series

	for _, item := range resp.Data {
		series = append(series, Series{
			ID:         item.ID,
			Name:       item.Name,
			Slug:       item.Slug,
			Status:     item.Status,
			FirstAired: item.FirstAired,
			Network:    item.Network,
			Banner:     item.Banner,
			Poster:     item.Poster,
			IMDBID:     item.IMDBID,
			Overview:   item.Overview,
		})
	}

	return series
}

// ensureAuth 确保认证有效
func (c *Client) ensureAuth() error {
	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	if token == "" {
		return c.authenticate()
	}

	// 可以添加令牌刷新逻辑
	return nil
}

// ClearCache 清空缓存
func (c *Client) ClearCache() {
	c.cache.Clear()
}

// GetCacheStats 获取缓存统计
func (c *Client) GetCacheStats() (seriesCount, episodesCount, searchCount int) {
	return c.cache.Size()
}
