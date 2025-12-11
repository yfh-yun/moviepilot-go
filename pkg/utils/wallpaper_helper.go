package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// WallpaperSource 壁纸来源类型
type WallpaperSource string

const (
	WallpaperSourceBing        WallpaperSource = "bing"
	WallpaperSourceMediaServer WallpaperSource = "mediaserver"
	WallpaperSourceCustomize   WallpaperSource = "customize"
	WallpaperSourceTMDB        WallpaperSource = "tmdb"
)

// WallpaperHelper 壁纸帮助类
type WallpaperHelper struct {
	logger           *zap.Logger
	config           Config
	cache            Cache
	mutex            sync.RWMutex
	mediaServerChain MediaServerChain
	tmdbChain        TmdbChain
}

// Config 壁纸配置
type Config struct {
	Source                WallpaperSource `json:"source"`
	CustomizeAPIURL       string          `json:"customize_api_url"`
	SecurityImageSuffixes []string        `json:"security_image_suffixes"`
}

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(key string, defaultValue any) any
	// Set 设置缓存值
	Set(key string, value any, ttl time.Duration)
}

// MediaServerChain 媒体服务器链接口
type MediaServerChain interface {
	// GetLatestWallpaper 获取最新壁纸
	GetLatestWallpaper() string
	// GetLatestWallpapers 获取最新壁纸列表
	GetLatestWallpapers(count int) []string
}

// TmdbChain TMDB链接口
type TmdbChain interface {
	// GetRandomWallpaper 获取随机壁纸
	GetRandomWallpaper() string
	// GetTrendingWallpapers 获取热门壁纸
	GetTrendingWallpapers(num int) []string
}

// NewWallpaperHelper 创建壁纸帮助类实例
func NewWallpaperHelper(config Config, cache Cache, mediaServerChain MediaServerChain, tmdbChain TmdbChain) *WallpaperHelper {
	return &WallpaperHelper{
		logger:           logger.GetLogger(),
		config:           config,
		cache:            cache,
		mediaServerChain: mediaServerChain,
		tmdbChain:        tmdbChain,
	}
}

// GetWallpaper 获取登录页面壁纸
func (h *WallpaperHelper) GetWallpaper() string {
	switch h.config.Source {
	case WallpaperSourceBing:
		return h.GetBingWallpaper()
	case WallpaperSourceMediaServer:
		return h.GetMediaServerWallpaper()
	case WallpaperSourceCustomize:
		return h.GetCustomizeWallpaper()
	case WallpaperSourceTMDB:
		return h.GetTMDbWallpaper()
	default:
		return ""
	}
}

// GetWallpapers 获取登录页面壁纸列表
func (h *WallpaperHelper) GetWallpapers(num int) []string {
	if num <= 0 {
		num = 10
	}

	switch h.config.Source {
	case WallpaperSourceBing:
		return h.GetBingWallpapers(num)
	case WallpaperSourceMediaServer:
		return h.GetMediaServerWallpapers(num)
	case WallpaperSourceCustomize:
		return h.GetCustomizeWallpapers()
	case WallpaperSourceTMDB:
		return h.GetTMDbWallpapers(num)
	default:
		return []string{}
	}
}

// GetBingWallpaper 获取Bing每日壁纸
func (h *WallpaperHelper) GetBingWallpaper() string {
	// 检查缓存
	cacheKey := "wallpaper_bing"
	if cached, ok := h.cache.Get(cacheKey, "").(string); ok && cached != "" {
		return cached
	}

	// 请求Bing壁纸API
	url := "https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1"
	resp, err := http.Get(url)
	if err != nil {
		h.logger.Error("请求Bing壁纸API失败", zap.Error(err))
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Bing壁纸API返回错误状态码", zap.Int("status_code", resp.StatusCode))
		return ""
	}

	// 解析响应
	var result struct {
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("读取Bing壁纸API响应失败", zap.Error(err))
		return ""
	}

	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.Error("解析Bing壁纸API响应失败", zap.Error(err))
		return ""
	}

	// 获取壁纸URL
	var wallpaperURL string
	if len(result.Images) > 0 {
		wallpaperURL = fmt.Sprintf("https://cn.bing.com%s", result.Images[0].URL)
	}

	// 缓存结果，有效期1小时
	if wallpaperURL != "" {
		h.cache.Set(cacheKey, wallpaperURL, time.Hour)
	}

	return wallpaperURL
}

// GetBingWallpapers 获取Bing壁纸列表
func (h *WallpaperHelper) GetBingWallpapers(num int) []string {
	if num <= 0 {
		num = 10
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("wallpaper_bing_list_%d", num)
	if cached, ok := h.cache.Get(cacheKey, []string{}).([]string); ok && len(cached) > 0 {
		return cached
	}

	// 请求Bing壁纸API
	url := fmt.Sprintf("https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=%d", num)
	resp, err := http.Get(url)
	if err != nil {
		h.logger.Error("请求Bing壁纸API失败", zap.Error(err))
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Bing壁纸API返回错误状态码", zap.Int("status_code", resp.StatusCode))
		return []string{}
	}

	// 解析响应
	var result struct {
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("读取Bing壁纸API响应失败", zap.Error(err))
		return []string{}
	}

	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.Error("解析Bing壁纸API响应失败", zap.Error(err))
		return []string{}
	}

	// 构建壁纸URL列表
	wallpapers := make([]string, 0, len(result.Images))
	for _, image := range result.Images {
		wallpapers = append(wallpapers, fmt.Sprintf("https://cn.bing.com%s", image.URL))
	}

	// 缓存结果，有效期1小时
	if len(wallpapers) > 0 {
		h.cache.Set(cacheKey, wallpapers, time.Hour)
	}

	return wallpapers
}

// GetMediaServerWallpaper 获取媒体服务器壁纸
func (h *WallpaperHelper) GetMediaServerWallpaper() string {
	// 检查缓存
	cacheKey := "wallpaper_mediaserver"
	if cached, ok := h.cache.Get(cacheKey, "").(string); ok && cached != "" {
		return cached
	}

	if h.mediaServerChain == nil {
		return ""
	}

	// 调用媒体服务器链获取壁纸
	wallpaperURL := h.mediaServerChain.GetLatestWallpaper()

	// 缓存结果，有效期1小时
	if wallpaperURL != "" {
		h.cache.Set(cacheKey, wallpaperURL, time.Hour)
	}

	return wallpaperURL
}

// GetMediaServerWallpapers 获取媒体服务器壁纸列表
func (h *WallpaperHelper) GetMediaServerWallpapers(num int) []string {
	if num <= 0 {
		num = 10
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("wallpaper_mediaserver_list_%d", num)
	if cached, ok := h.cache.Get(cacheKey, []string{}).([]string); ok && len(cached) > 0 {
		return cached
	}

	if h.mediaServerChain == nil {
		return []string{}
	}

	// 调用媒体服务器链获取壁纸列表
	wallpapers := h.mediaServerChain.GetLatestWallpapers(num)

	// 缓存结果，有效期1小时
	if len(wallpapers) > 0 {
		h.cache.Set(cacheKey, wallpapers, time.Hour)
	}

	return wallpapers
}

// GetTMDbWallpaper 获取TMDB壁纸
func (h *WallpaperHelper) GetTMDbWallpaper() string {
	// 检查缓存
	cacheKey := "wallpaper_tmdb"
	if cached, ok := h.cache.Get(cacheKey, "").(string); ok && cached != "" {
		return cached
	}

	if h.tmdbChain == nil {
		return ""
	}

	// 调用TMDB链获取壁纸
	wallpaperURL := h.tmdbChain.GetRandomWallpaper()

	// 缓存结果，有效期1小时
	if wallpaperURL != "" {
		h.cache.Set(cacheKey, wallpaperURL, time.Hour)
	}

	return wallpaperURL
}

// GetTMDbWallpapers 获取TMDB壁纸列表
func (h *WallpaperHelper) GetTMDbWallpapers(num int) []string {
	if num <= 0 {
		num = 10
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("wallpaper_tmdb_list_%d", num)
	if cached, ok := h.cache.Get(cacheKey, []string{}).([]string); ok && len(cached) > 0 {
		return cached
	}

	if h.tmdbChain == nil {
		return []string{}
	}

	// 调用TMDB链获取壁纸列表
	wallpapers := h.tmdbChain.GetTrendingWallpapers(num)

	// 缓存结果，有效期1小时
	if len(wallpapers) > 0 {
		h.cache.Set(cacheKey, wallpapers, time.Hour)
	}

	return wallpapers
}

// GetCustomizeWallpaper 获取自定义壁纸
func (h *WallpaperHelper) GetCustomizeWallpaper() string {
	// 检查缓存
	cacheKey := "wallpaper_customize"
	if cached, ok := h.cache.Get(cacheKey, "").(string); ok && cached != "" {
		return cached
	}

	// 获取自定义壁纸列表
	wallpapers := h.GetCustomizeWallpapers()
	var wallpaperURL string
	if len(wallpapers) > 0 {
		wallpaperURL = wallpapers[0]
	}

	// 缓存结果，有效期1小时
	if wallpaperURL != "" {
		h.cache.Set(cacheKey, wallpaperURL, time.Hour)
	}

	return wallpaperURL
}

// GetCustomizeWallpapers 获取自定义壁纸列表
func (h *WallpaperHelper) GetCustomizeWallpapers() []string {
	// 检查缓存
	cacheKey := "wallpaper_customize_list"
	if cached, ok := h.cache.Get(cacheKey, []string{}).([]string); ok && len(cached) > 0 {
		return cached
	}

	// 检查自定义壁纸API URL
	if h.config.CustomizeAPIURL == "" {
		return []string{}
	}

	// 请求自定义壁纸API
	resp, err := http.Get(h.config.CustomizeAPIURL)
	if err != nil {
		h.logger.Error("请求自定义壁纸API失败", zap.Error(err))
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("自定义壁纸API返回错误状态码", zap.Int("status_code", resp.StatusCode))
		return []string{}
	}

	// 检查响应类型
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && strings.HasPrefix(contentType, "image/") {
		// 如果是图片直接返回URL
		return []string{h.config.CustomizeAPIURL}
	}

	// 解析响应
	var result any
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("读取自定义壁纸API响应失败", zap.Error(err))
		return []string{}
	}

	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.Error("解析自定义壁纸API响应失败", zap.Error(err))
		return []string{}
	}

	// 查找图片URL
	wallpapers := h.findImageURLs(result)

	// 缓存结果，有效期1小时
	if len(wallpapers) > 0 {
		h.cache.Set(cacheKey, wallpapers, time.Hour)
	}

	return wallpapers
}

// findImageURLs 递归查找图片URL
func (h *WallpaperHelper) findImageURLs(obj any) []string {
	result := []string{}

	// 处理字符串
	if str, ok := obj.(string); ok {
		if h.isImageURL(str) {
			result = append(result, str)
		}
		return result
	}

	// 处理字典
	if dict, ok := obj.(map[string]any); ok {
		for _, value := range dict {
			result = append(result, h.findImageURLs(value)...)
		}
		return result
	}

	// 处理数组
	if arr, ok := obj.([]any); ok {
		for _, item := range arr {
			result = append(result, h.findImageURLs(item)...)
		}
		return result
	}

	return result
}

// isImageURL 检查是否为图片URL
func (h *WallpaperHelper) isImageURL(url string) bool {
	// 检查后缀
	url = strings.ToLower(url)
	for _, suffix := range h.config.SecurityImageSuffixes {
		if strings.HasSuffix(url, suffix) {
			return true
		}
	}
	return false
}

// SetSource 设置壁纸来源
func (h *WallpaperHelper) SetSource(source WallpaperSource) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.config.Source = source
}

// GetSource 获取壁纸来源
func (h *WallpaperHelper) GetSource() WallpaperSource {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.config.Source
}
