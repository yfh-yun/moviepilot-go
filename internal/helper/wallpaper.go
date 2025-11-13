package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"moviepilot-go/internal/chain"
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/core/cache"
	"moviepilot-go/internal/utils"
)

// WallpaperHelper 壁纸帮助�?type WallpaperHelper struct {
}

// 使用单例模式
var (
	wallpaperHelperInstance *WallpaperHelper
	once                    sync.Once
)

// GetWallpaperHelper 获取壁纸帮助类单例实�?func GetWallpaperHelper() *WallpaperHelper {
	once.Do(func() {
		wallpaperHelperInstance = &WallpaperHelper{}
	})
	return wallpaperHelperInstance
}

// GetWallpaper 获取登录页面壁纸
func (w *WallpaperHelper) GetWallpaper() string {
	switch config.GlobalConfig.Wallpaper {
	case "bing":
		return w.GetBingWallpaper()
	case "mediaserver":
		return w.GetMediaserverWallpaper()
	case "customize":
		return w.GetCustomizeWallpaper()
	case "tmdb":
		return w.GetTmdbWallpaper()
	default:
		return ""
	}
}

// GetWallpapers 获取登录页面壁纸列表
func (w *WallpaperHelper) GetWallpapers(num int) []string {
	switch config.GlobalConfig.Wallpaper {
	case "bing":
		return w.GetBingWallpapers(num)
	case "mediaserver":
		return w.GetMediaserverWallpapers(num)
	case "customize":
		return w.GetCustomizeWallpapers()
	case "tmdb":
		return w.GetTmdbWallpapers(num)
	default:
		return []string{}
	}
}

// GetTmdbWallpaper 获取TMDB每日壁纸
func (w *WallpaperHelper) GetTmdbWallpaper() string {
	// 使用缓存
	cacheKey := "tmdb_wallpaper"
	if val, exists := cache.Get(cacheKey); exists {
		if str, ok := val.(string); ok {
			return str
		}
	}

	// 获取新的壁纸
	tmdbChain := chain.GetTmdbChain()
	wallpaper := tmdbChain.GetRandomWallpager()

	// 缓存结果 (1小时)
	cache.Set(cacheKey, wallpaper, 1*time.Hour)
	return wallpaper
}

// GetTmdbWallpapers 获取指定数量的TMDB壁纸
func (w *WallpaperHelper) GetTmdbWallpapers(num int) []string {
	// 使用缓存
	cacheKey := fmt.Sprintf("tmdb_wallpapers_%d", num)
	if val, exists := cache.Get(cacheKey); exists {
		if wallpapers, ok := val.([]string); ok {
			return wallpapers
		}
	}

	// 获取新的壁纸列表
	tmdbChain := chain.GetTmdbChain()
	wallpapers := tmdbChain.GetTrendingWallpapers(num)

	// 缓存结果 (1小时)
	cache.Set(cacheKey, wallpapers, 1*time.Hour)
	return wallpapers
}

// GetBingWallpaper 获取Bing每日壁纸
func (w *WallpaperHelper) GetBingWallpaper() string {
	// 使用缓存
	cacheKey := "bing_wallpaper"
	if val, exists := cache.Get(cacheKey); exists {
		if str, ok := val.(string); ok {
			return str
		}
	}

	url := "https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1"
	resp, err := utils.RequestUtils{}.GetRes(url, 5*time.Second)
	if err != nil || resp.StatusCode != http.StatusOK {
		cache.Set(cacheKey, "", 1*time.Hour)
		return ""
	}

	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		cache.Set(cacheKey, "", 1*time.Hour)
		return ""
	}

	images, ok := result["images"].([]interface{})
	if !ok || len(images) == 0 {
		cache.Set(cacheKey, "", 1*time.Hour)
		return ""
	}

	image, ok := images[0].(map[string]interface{})
	if !ok {
		cache.Set(cacheKey, "", 1*time.Hour)
		return ""
	}

	imageURL, ok := image["url"].(string)
	if !ok {
		cache.Set(cacheKey, "", 1*time.Hour)
		return ""
	}

	wallpaperURL := "https://cn.bing.com" + imageURL
	// 缓存结果 (1小时)
	cache.Set(cacheKey, wallpaperURL, 1*time.Hour)
	return wallpaperURL
}

// GetBingWallpapers 获取指定数量的Bing每日壁纸
func (w *WallpaperHelper) GetBingWallpapers(num int) []string {
	// 使用缓存
	cacheKey := fmt.Sprintf("bing_wallpapers_%d", num)
	if val, exists := cache.Get(cacheKey); exists {
		if wallpapers, ok := val.([]string); ok {
			return wallpapers
		}
	}

	url := fmt.Sprintf("https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=%d", num)
	resp, err := utils.RequestUtils{}.GetRes(url, 5*time.Second)
	if err != nil || resp.StatusCode != http.StatusOK {
		cache.Set(cacheKey, []string{}, 1*time.Hour)
		return []string{}
	}

	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		cache.Set(cacheKey, []string{}, 1*time.Hour)
		return []string{}
	}

	images, ok := result["images"].([]interface{})
	if !ok {
		cache.Set(cacheKey, []string{}, 1*time.Hour)
		return []string{}
	}

	wallpapers := make([]string, 0)
	for _, img := range images {
		image, ok := img.(map[string]interface{})
		if !ok {
			continue
		}

		imageURL, ok := image["url"].(string)
		if !ok {
			continue
		}

		wallpaperURL := "https://cn.bing.com" + imageURL
		wallpapers = append(wallpapers, wallpaperURL)
	}

	// 缓存结果 (1小时)
	cache.Set(cacheKey, wallpapers, 1*time.Hour)
	return wallpapers
}

// GetMediaserverWallpaper 获取媒体服务器壁�?func (w *WallpaperHelper) GetMediaserverWallpaper() string {
	// 使用缓存
	cacheKey := "mediaserver_wallpaper"
	if val, exists := cache.Get(cacheKey); exists {
		if str, ok := val.(string); ok {
			return str
		}
	}

	mediaServerChain := chain.GetMediaServerChain()
	wallpaper := mediaServerChain.GetLatestWallpaper()

	// 缓存结果 (1小时)
	cache.Set(cacheKey, wallpaper, 1*time.Hour)
	return wallpaper
}

// GetMediaserverWallpapers 获取指定数量的媒体服务器壁纸
func (w *WallpaperHelper) GetMediaserverWallpapers(num int) []string {
	// 使用缓存
	cacheKey := fmt.Sprintf("mediaserver_wallpapers_%d", num)
	if val, exists := cache.Get(cacheKey); exists {
		if wallpapers, ok := val.([]string); ok {
			return wallpapers
		}
	}

	mediaServerChain := chain.GetMediaServerChain()
	wallpapers := mediaServerChain.GetLatestWallpapers(num)

	// 缓存结果 (1小时)
	cache.Set(cacheKey, wallpapers, 1*time.Hour)
	return wallpapers
}

// findFilesWithSuffixes 递归查找对象中所有包含特定后缀的文件，返回匹配的字符串列表
func (w *WallpaperHelper) findFilesWithSuffixes(obj interface{}, suffixes []string) []string {
	result := make([]string, 0)

	// 处理字符�?	if str, ok := obj.(string); ok {
		for _, suffix := range suffixes {
			if strings.HasSuffix(str, suffix) {
				result = append(result, str)
				break
			}
		}
		return result
	}

	// 处理字典(map[string]interface{})
	if dict, ok := obj.(map[string]interface{}); ok {
		for _, value := range dict {
			result = append(result, w.findFilesWithSuffixes(value, suffixes)...)
		}
		return result
	}

	// 处理列表([]interface{})
	if list, ok := obj.([]interface{}); ok {
		for _, item := range list {
			result = append(result, w.findFilesWithSuffixes(item, suffixes)...)
		}
		return result
	}

	return result
}

// GetCustomizeWallpaper 获取自定义壁纸api壁纸
func (w *WallpaperHelper) GetCustomizeWallpaper() string {
	// 使用缓存
	cacheKey := "customize_wallpaper"
	if val, exists := cache.Get(cacheKey); exists {
		if str, ok := val.(string); ok {
			return str
		}
	}

	wallpaperList := w.GetCustomizeWallpapers()
	if len(wallpaperList) > 0 {
		// 缓存结果 (1小时)
		cache.Set(cacheKey, wallpaperList[0], 1*time.Hour)
		return wallpaperList[0]
	}

	// 缓存空结�?(1小时)
	cache.Set(cacheKey, "", 1*time.Hour)
	return ""
}

// GetCustomizeWallpapers 获取自定义壁纸api壁纸列表
func (w *WallpaperHelper) GetCustomizeWallpapers() []string {
	// 使用缓存
	cacheKey := "customize_wallpapers"
	if val, exists := cache.Get(cacheKey); exists {
		if wallpapers, ok := val.([]string); ok {
			return wallpapers
		}
	}

	// 判断是否存在自定义壁纸api
	if config.GlobalConfig.CustomizeWallpaperApiUrl == "" {
		cache.Set(cacheKey, []string{}, 1*time.Hour)
		return []string{}
	}

	wallpaperList := make([]string, 0)
	resp, err := utils.RequestUtils{}.GetRes(config.GlobalConfig.CustomizeWallpaperApiUrl, 15*time.Second)
	if err != nil || resp == nil {
		cache.Set(cacheKey, []string{}, 1*time.Hour)
		return []string{}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cache.Set(cacheKey, []string{}, 1*time.Hour)
		return []string{}
	}

	// 如果返回的是图片格式
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && strings.HasPrefix(strings.ToLower(contentType), "image/") {
		wallpaperList = append(wallpaperList, config.GlobalConfig.CustomizeWallpaperApiUrl)
		// 缓存结果 (1小时)
		cache.Set(cacheKey, wallpaperList, 1*time.Hour)
		return wallpaperList
	}

	// 尝试解析JSON
	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		cache.Set(cacheKey, []string{}, 1*time.Hour)
		return []string{}
	}

	wallpaperList = w.findFilesWithSuffixes(result, config.GlobalConfig.SecurityImageSuffixes)
	// 缓存结果 (1小时)
	cache.Set(cacheKey, wallpaperList, 1*time.Hour)
	return wallpaperList
}
