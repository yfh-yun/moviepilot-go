// Package trimemedia Trimedia API客户端
package trimemedia

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/httpclient"

	"go.uber.org/zap"
)

// Category 媒体分类
type Category string

const (
	CategoryMovie  Category = "Movie"
	CategoryTV     Category = "TV"
	CategoryMix    Category = "Mix"
	CategoryOthers Category = "Others"
)

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeMovie   MediaType = "Movie"
	MediaTypeTV      MediaType = "TV"
	MediaTypeSeason  MediaType = "Season"
	MediaTypeEpisode MediaType = "Episode"
	MediaTypeVideo   MediaType = "Video"
	MediaTypeDir     MediaType = "Directory"
)

// User 用户信息
type User struct {
	GUID     string `json:"guid"`
	Username string `json:"username"`
	IsAdmin  int    `json:"is_admin"`
}

// MediaSummary 媒体汇总信息
type MediaSummary struct {
	Favorite int `json:"favorite"`
	Movie    int `json:"movie"`
	TV       int `json:"tv"`
	Video    int `json:"video"`
	Total    int `json:"total"`
}

// Version 版本信息
type Version struct {
	Frontend string `json:"frontend"`
	Backend  string `json:"mediasrvVersion"`
}

// MediaItem 媒体项目
type MediaItem struct {
	GUID          string    `json:"guid"`
	AncestorGUID  string    `json:"ancestor_guid"`
	Type          MediaType `json:"type"`
	TVTitle       string    `json:"tv_title"`
	ParentTitle   string    `json:"parent_title"`
	Title         string    `json:"title"`
	OriginalTitle string    `json:"original_title"`
	Overview      string    `json:"overview"`
	Poster        string    `json:"poster"`
	Backdrops     string    `json:"backdrops"`
	Posters       string    `json:"posters"`
	DoubanID      int       `json:"douban_id"`
	IMDBID        string    `json:"imdb_id"`
	TrimID        string    `json:"trim_id"`
	ReleaseDate   string    `json:"release_date"`
	AirDate       string    `json:"air_date"`
	VoteAverage   string    `json:"vote_average"`
	SeasonNumber  int       `json:"season_number"`
	EpisodeNumber int       `json:"episode_number"`
	Duration      int       `json:"duration"` // 片长(秒)
	TS            int       `json:"ts"`       // 已播放(秒)
	Watched       int       `json:"watched"`   // 1:已看完
}

// TMDBID 从TrimID获取TMDB ID
func (m *MediaItem) TMDBID() *int {
	if m.TrimID == "" {
		return nil
	}
	
	trimID := strings.TrimPrefix(m.TrimID, "tt")
	trimID = strings.TrimPrefix(trimID, "tm")
	
	if id, err := strconv.Atoi(trimID); err == nil {
		return &id
	}
	
	return nil
}

// APIResponse API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
}

// APIClient Trimedia API客户端
type APIClient struct {
	host     string
	apiKey   string
	token    string
	version  *Version
	client   *httpclient.Client
	logger   *zap.Logger
}

// NewAPIClient 创建API客户端
func NewAPIClient(host, apiKey string) *APIClient {
	if host == "" || apiKey == "" {
		return nil
	}

	// 规范化host
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}
	host = strings.TrimSuffix(host, "/")

	client := &APIClient{
		host:   host,
		apiKey: apiKey,
		client: httpclient.NewClient(httpclient.Options{
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "MoviePilot-Trimedia/1.0",
			},
		}),
		logger: logger.Logger,
	}

	return client
}

// Host 获取host
func (c *APIClient) Host() string {
	return c.host
}

// APIKey 获取API密钥
func (c *APIClient) APIKey() string {
	return c.apiKey
}

// Token 获取认证令牌
func (c *APIClient) Token() string {
	return c.token
}

// Version 获取版本信息
func (c *APIClient) Version() *Version {
	return c.version
}

// request 发送请求
func (c *APIClient) request(ctx context.Context, method, path string, data interface{}) (*APIResponse, error) {
	apiPath := c.host + "/api/v1" + path

	var req *http.Request
	var err error

	if method == "GET" {
		req, err = http.NewRequestWithContext(ctx, method, apiPath, nil)
	} else {
		var body []byte
		if data != nil {
			body, err = json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("marshal request data failed: %w", err)
			}
		}
		
		req, err = http.NewRequestWithContext(ctx, method, apiPath, strings.NewReader(string(body)))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置认证头
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	// 设置认证X头
	authX := c.getAuthX(path, data)
	if authX != "" {
		req.Header.Set("authx", authX)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	// 如果需要重新认证
	if apiResp.Code == 401 {
		c.token = ""
	}

	return &apiResp, nil
}

// getAuthX 生成认证X头
func (c *APIClient) getAuthX(apiPath string, data interface{}) string {
	if data == nil {
		return ""
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	timestamp := time.Now().Unix()
	hash := md5.Sum([]byte(fmt.Sprintf("%s%s%d%s%s",
		jsonData, apiPath, timestamp, c.apiKey, c.token)))

	return fmt.Sprintf("%d,%s,%s", timestamp, c.apiKey, hex.EncodeToString(hash[:]))
}

// SysVersion 获取系统版本
func (c *APIClient) SysVersion(ctx context.Context) (*Version, error) {
	resp, err := c.request(ctx, "GET", "/sys/version", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("get version failed: %s", resp.Message)
	}

	if resp.Data != nil {
		dataMap, ok := resp.Data.(map[string]interface{})
		if ok {
			c.version = &Version{
				Frontend: getString(dataMap, "version"),
				Backend:  getString(dataMap, "mediasrvVersion"),
			}
			return c.version
		}
	}

	return nil, nil
}

// Login 登录
func (c *APIClient) Login(ctx context.Context, username, password string) (string, error) {
	data := map[string]interface{}{
		"username": username,
		"password": password,
	}

	resp, err := c.request(ctx, "POST", "/login", data)
	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", fmt.Errorf("login failed: %s", resp.Message)
	}

	if resp.Data != nil {
		dataMap, ok := resp.Data.(map[string]interface{})
		if ok {
			if token := getString(dataMap, "token"); token != "" {
				c.token = token
				c.logger.Info("Login successful",
					zap.String("username", username))
				return token, nil
			}
		}
	}

	return "", fmt.Errorf("no token in response")
}

// IsAuthenticated 检查是否已认证
func (c *APIClient) IsAuthenticated() bool {
	return c.token != ""
}

// GetMediaSummary 获取媒体汇总
func (c *APIClient) GetMediaSummary(ctx context.Context) (*MediaSummary, error) {
	resp, err := c.request(ctx, "GET", "/summary", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("get summary failed: %s", resp.Message)
	}

	if resp.Data != nil {
		dataMap, ok := resp.Data.(map[string]interface{})
		if ok {
			summary := &MediaSummary{
				Favorite: getInt(dataMap, "favorite"),
				Movie:    getInt(dataMap, "movie"),
				TV:       getInt(dataMap, "tv"),
				Video:    getInt(dataMap, "video"),
			}
			summary.Total = summary.Movie + summary.TV + summary.Video
			return summary, nil
		}
	}

	return nil, nil
}

// SearchMedia 搜索媒体
func (c *APIClient) SearchMedia(ctx context.Context, keyword string, mediaType Category, page int) ([]*MediaItem, error) {
	params := url.Values{}
	params.Set("keyword", keyword)
	if mediaType != "" {
		params.Set("category", string(mediaType))
	}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	path := "/search?" + params.Encode()

	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("search failed: %s", resp.Message)
	}

	if resp.Data != nil {
		if items, ok := resp.Data.([]interface{}); ok {
			return c.parseMediaItems(items)
		}
	}

	return []*MediaItem{}, nil
}

// GetMediaDetail 获取媒体详情
func (c *APIClient) GetMediaDetail(ctx context.Context, guid string) (*MediaItem, error) {
	path := fmt.Sprintf("/detail/%s", guid)

	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("get detail failed: %s", resp.Message)
	}

	if resp.Data != nil {
		dataMap, ok := resp.Data.(map[string]interface{})
		if ok {
			return c.parseMediaItem(dataMap), nil
		}
	}

	return nil, nil
}

// GetWatchHistory 获取观看历史
func (c *APIClient) GetWatchHistory(ctx context.Context, mediaType Category, page int) ([]*MediaItem, error) {
	params := url.Values{}
	if mediaType != "" {
		params.Set("category", string(mediaType))
	}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	path := "/history?" + params.Encode()

	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("get history failed: %s", resp.Message)
	}

	if resp.Data != nil {
		if items, ok := resp.Data.([]interface{}); ok {
			return c.parseMediaItems(items)
		}
	}

	return []*MediaItem{}, nil
}

// MarkAsWatched 标记为已观看
func (c *APIClient) MarkAsWatched(ctx context.Context, guid string) error {
	data := map[string]interface{}{
		"guid": guid,
	}

	resp, err := c.request(ctx, "POST", "/watched", data)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("mark as watched failed: %s", resp.Message)
	}

	return nil
}

// MarkAsUnwatched 标记为未观看
func (c *APIClient) MarkAsUnwatched(ctx context.Context, guid string) error {
	data := map[string]interface{}{
		"guid": guid,
	}

	resp, err := c.request(ctx, "POST", "/unwatched", data)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("mark as unwatched failed: %s", resp.Message)
	}

	return nil
}

// AddToFavorite 添加到收藏
func (c *APIClient) AddToFavorite(ctx context.Context, guid string) error {
	data := map[string]interface{}{
		"guid": guid,
	}

	resp, err := c.request(ctx, "POST", "/favorite", data)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("add to favorite failed: %s", resp.Message)
	}

	return nil
}

// RemoveFromFavorite 从收藏移除
func (c *APIClient) RemoveFromFavorite(ctx context.Context, guid string) error {
	path := fmt.Sprintf("/favorite/%s", guid)

	resp, err := c.request(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("remove from favorite failed: %s", resp.Message)
	}

	return nil
}

// GetFavorites 获取收藏列表
func (c *APIClient) GetFavorites(ctx context.Context, mediaType Category, page int) ([]*MediaItem, error) {
	params := url.Values{}
	if mediaType != "" {
		params.Set("category", string(mediaType))
	}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	path := "/favorite?" + params.Encode()

	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("get favorites failed: %s", resp.Message)
	}

	if resp.Data != nil {
		if items, ok := resp.Data.([]interface{}); ok {
			return c.parseMediaItems(items)
		}
	}

	return []*MediaItem{}, nil
}

// 辅助方法

// getString 从map获取字符串
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getInt 从map获取整数
func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
	}
	return 0
}

// parseMediaItems 解析媒体项目列表
func (c *APIClient) parseMediaItems(items []interface{}) []*MediaItem {
	var mediaItems []*MediaItem

	for _, item := range items {
		if dataMap, ok := item.(map[string]interface{}); ok {
			mediaItems = append(mediaItems, c.parseMediaItem(dataMap))
		}
	}

	return mediaItems
}

// parseMediaItem 解析单个媒体项目
func (c *APIClient) parseMediaItem(dataMap map[string]interface{}) *MediaItem {
	return &MediaItem{
		GUID:          getString(dataMap, "guid"),
		AncestorGUID:  getString(dataMap, "ancestor_guid"),
		Type:          MediaType(getString(dataMap, "type")),
		TVTitle:       getString(dataMap, "tv_title"),
		ParentTitle:   getString(dataMap, "parent_title"),
		Title:         getString(dataMap, "title"),
		OriginalTitle: getString(dataMap, "original_title"),
		Overview:      getString(dataMap, "overview"),
		Poster:        getString(dataMap, "poster"),
		Backdrops:     getString(dataMap, "backdrops"),
		Posters:       getString(dataMap, "posters"),
		DoubanID:      getInt(dataMap, "douban_id"),
		IMDBID:        getString(dataMap, "imdb_id"),
		TrimID:        getString(dataMap, "trim_id"),
		ReleaseDate:   getString(dataMap, "release_date"),
		AirDate:       getString(dataMap, "air_date"),
		VoteAverage:   getString(dataMap, "vote_average"),
		SeasonNumber:  getInt(dataMap, "season_number"),
		EpisodeNumber: getInt(dataMap, "episode_number"),
		Duration:      getInt(dataMap, "duration"),
		TS:            getInt(dataMap, "ts"),
		Watched:       getInt(dataMap, "watched"),
	}
}

// HealthCheck 健康检查
func (c *APIClient) HealthCheck(ctx context.Context) error {
	_, err := c.request(ctx, "GET", "/health", nil)
	return err
}