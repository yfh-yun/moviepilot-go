package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// WallpaperHelper 壁纸帮助类
type WallpaperHelper struct {
	wallpaperType string
	httpClient    *http.Client
	cacheDir      string
	mutex         sync.RWMutex
}

// WallpaperType 壁纸类型
type WallpaperType string

const (
	WallpaperTypeBing        WallpaperType = "bing"
	WallpaperTypeMediaServer WallpaperType = "mediaserver"
	WallpaperTypeCustomize   WallpaperType = "customize"
	WallpaperTypeTMDB        WallpaperType = "tmdb"
)

// BingWallpaperResponse Bing壁纸响应
type BingWallpaperResponse struct {
	Images []BingImage `json:"images"`
}

// BingImage Bing图片信息
type BingImage struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	Copyright     string `json:"copyright"`
	CopyrightLink string `json:"copyrightlink"`
	Date          string `json:"date"`
}

// TMDBImage TMDB图片信息
type TMDBImage struct {
	FilePath string  `json:"file_path"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	VoteAvg  float64 `json:"vote_average"`
	VoteCnt  int     `json:"vote_count"`
}

// NewWallpaperHelper 创建壁纸助手实例
func NewWallpaperHelper() *WallpaperHelper {
	return &WallpaperHelper{
		wallpaperType: "customize",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cacheDir: filepath.Join(os.TempDir(), "wallpaper_cache"),
	}
}

// SetWallpaperType 设置壁纸类型
func (wh *WallpaperHelper) SetWallpaperType(wallpaperType string) {
	wh.mutex.Lock()
	defer wh.mutex.Unlock()
	wh.wallpaperType = wallpaperType
}

// GetWallpaperType 获取壁纸类型
func (wh *WallpaperHelper) GetWallpaperType() string {
	wh.mutex.RLock()
	defer wh.mutex.RUnlock()
	return wh.wallpaperType
}

// GetWallpaper 获取登录页面壁纸
func (wh *WallpaperHelper) GetWallpaper() (string, error) {
	wh.mutex.RLock()
	wallpaperType := wh.wallpaperType
	wh.mutex.RUnlock()

	switch wallpaperType {
	case string(WallpaperTypeBing):
		return wh.getBingWallpaper()
	case string(WallpaperTypeMediaServer):
		return wh.getMediaserverWallpaper()
	case string(WallpaperTypeCustomize):
		return wh.getCustomizeWallpaper()
	case string(WallpaperTypeTMDB):
		return wh.getTMDBWallpaper()
	default:
		return "", nil
	}
}

// GetWallpapers 获取登录页面壁纸列表
func (wh *WallpaperHelper) GetWallpapers(num int) ([]string, error) {
	if num <= 0 {
		num = 10
	}

	wh.mutex.RLock()
	wallpaperType := wh.wallpaperType
	wh.mutex.RUnlock()

	switch wallpaperType {
	case string(WallpaperTypeBing):
		return wh.getBingWallpapers(num)
	case string(WallpaperTypeMediaServer):
		return wh.getMediaserverWallpapers(num)
	case string(WallpaperTypeCustomize):
		return wh.getCustomizeWallpapers()
	case string(WallpaperTypeTMDB):
		return wh.getTMDBWallpapers(num)
	default:
		return []string{}, nil
	}
}

// getBingWallpaper 获取Bing壁纸
func (wh *WallpaperHelper) getBingWallpaper() (string, error) {
	// 缓存1小时
	cacheKey := "bing_wallpaper"
	if cached := wh.getCachedWallpaper(cacheKey, time.Hour); cached != "" {
		return cached, nil
	}

	// 获取Bing壁纸
	url := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=zh-CN"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := wh.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Bing wallpaper: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Bing API returned status: %d", resp.StatusCode)
	}

	var response BingWallpaperResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode Bing response: %v", err)
	}

	if len(response.Images) == 0 {
		return "", fmt.Errorf("no Bing wallpaper found")
	}

	image := response.Images[0]
	imageURL := "https://www.bing.com" + image.URL

	// 缓存结果
	wh.setCachedWallpaper(cacheKey, imageURL)

	return imageURL, nil
}

// getBingWallpapers 获取Bing壁纸列表
func (wh *WallpaperHelper) getBingWallpapers(num int) ([]string, error) {
	// 缓存1小时
	cacheKey := fmt.Sprintf("bing_wallpapers_%d", num)
	if cached := wh.getCachedWallpaperList(cacheKey, time.Hour); len(cached) > 0 {
		return cached, nil
	}

	// 获取多张Bing壁纸
	url := fmt.Sprintf("https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=%d&mkt=zh-CN", num)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := wh.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Bing wallpapers: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bing API returned status: %d", resp.StatusCode)
	}

	var response BingWallpaperResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Bing response: %v", err)
	}

	var wallpapers []string
	for _, image := range response.Images {
		imageURL := "https://www.bing.com" + image.URL
		wallpapers = append(wallpapers, imageURL)
	}

	// 缓存结果
	wh.setCachedWallpaperList(cacheKey, wallpapers)

	return wallpapers, nil
}

// getMediaserverWallpaper 获取媒体服务器壁纸
func (wh *WallpaperHelper) getMediaserverWallpaper() (string, error) {
	// 这里应该从媒体服务器获取随机壁纸
	// 简化实现，返回一个示例URL
	return "https://via.placeholder.com/1920x1080/4285f4/ffffff?text=MediaServer+Wallpaper", nil
}

// getMediaserverWallpapers 获取媒体服务器壁纸列表
func (wh *WallpaperHelper) getMediaserverWallpapers(num int) ([]string, error) {
	// 这里应该从媒体服务器获取壁纸列表
	// 简化实现，返回示例URLs
	wallpapers := make([]string, num)
	for i := 0; i < num; i++ {
		wallpapers[i] = fmt.Sprintf("https://via.placeholder.com/1920x1080/4285f4/ffffff?text=MediaServer+Wallpaper+%d", i+1)
	}
	return wallpapers, nil
}

// getCustomizeWallpaper 获取自定义壁纸
func (wh *WallpaperHelper) getCustomizeWallpaper() (string, error) {
	// 这里应该从自定义配置获取壁纸
	// 简化实现，返回一个示例URL
	return "https://via.placeholder.com/1920x1080/34a853/ffffff?text=Custom+Wallpaper", nil
}

// getCustomizeWallpapers 获取自定义壁纸列表
func (wh *WallpaperHelper) getCustomizeWallpapers() ([]string, error) {
	// 这里应该从自定义配置获取壁纸列表
	// 简化实现，返回示例URLs
	return []string{
		"https://via.placeholder.com/1920x1080/34a853/ffffff?text=Custom+Wallpaper+1",
		"https://via.placeholder.com/1920x1080/4285f4/ffffff?text=Custom+Wallpaper+2",
		"https://via.placeholder.com/1920x1080/fbbc04/ffffff?text=Custom+Wallpaper+3",
	}, nil
}

// getTMDBWallpaper 获取TMDB壁纸
func (wh *WallpaperHelper) getTMDBWallpaper() (string, error) {
	// 缓存1小时
	cacheKey := "tmdb_wallpaper"
	if cached := wh.getCachedWallpaper(cacheKey, time.Hour); cached != "" {
		return cached, nil
	}

	// 这里应该从TMDB API获取随机壁纸
	// 简化实现，返回一个示例URL
	imageURL := "https://via.placeholder.com/1920x1080/ea4335/ffffff?text=TMDB+Wallpaper"

	// 缓存结果
	wh.setCachedWallpaper(cacheKey, imageURL)

	return imageURL, nil
}

// getTMDBWallpapers 获取TMDB壁纸列表
func (wh *WallpaperHelper) getTMDBWallpapers(num int) ([]string, error) {
	// 缓存1小时
	cacheKey := fmt.Sprintf("tmdb_wallpapers_%d", num)
	if cached := wh.getCachedWallpaperList(cacheKey, time.Hour); len(cached) > 0 {
		return cached, nil
	}

	// 这里应该从TMDB API获取壁纸列表
	// 简化实现，返回示例URLs
	wallpapers := make([]string, num)
	for i := 0; i < num; i++ {
		wallpapers[i] = fmt.Sprintf("https://via.placeholder.com/1920x1080/ea4335/ffffff?text=TMDB+Wallpaper+%d", i+1)
	}

	// 缓存结果
	wh.setCachedWallpaperList(cacheKey, wallpapers)

	return wallpapers, nil
}

// getCachedWallpaper 获取缓存的壁纸
func (wh *WallpaperHelper) getCachedWallpaper(cacheKey string, ttl time.Duration) string {
	cacheFile := filepath.Join(wh.cacheDir, cacheKey+".txt")

	file, err := os.Stat(cacheFile)
	if err != nil {
		return ""
	}

	if time.Since(file.ModTime()) > ttl {
		return ""
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// setCachedWallpaper 设置缓存的壁纸
func (wh *WallpaperHelper) setCachedWallpaper(cacheKey, imageURL string) error {
	if err := os.MkdirAll(wh.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	cacheFile := filepath.Join(wh.cacheDir, cacheKey+".txt")
	return os.WriteFile(cacheFile, []byte(imageURL), 0644)
}

// getCachedWallpaperList 获取缓存的壁纸列表
func (wh *WallpaperHelper) getCachedWallpaperList(cacheKey string, ttl time.Duration) []string {
	cacheFile := filepath.Join(wh.cacheDir, cacheKey+".json")

	file, err := os.Stat(cacheFile)
	if err != nil {
		return nil
	}

	if time.Since(file.ModTime()) > ttl {
		return nil
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	var wallpapers []string
	if err := json.Unmarshal(data, &wallpapers); err != nil {
		return nil
	}

	return wallpapers
}

// setCachedWallpaperList 设置缓存的壁纸列表
func (wh *WallpaperHelper) setCachedWallpaperList(cacheKey string, wallpapers []string) error {
	if err := os.MkdirAll(wh.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	data, err := json.Marshal(wallpapers)
	if err != nil {
		return fmt.Errorf("failed to marshal wallpapers: %v", err)
	}

	cacheFile := filepath.Join(wh.cacheDir, cacheKey+".json")
	return os.WriteFile(cacheFile, data, 0644)
}

// DownloadWallpaper 下载壁纸
func (wh *WallpaperHelper) DownloadWallpaper(imageURL string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("image URL cannot be empty")
	}

	// 生成文件名
	ext := filepath.Ext(imageURL)
	if ext == "" {
		ext = ".jpg"
	}

	filename := fmt.Sprintf("wallpaper_%d%s", time.Now().Unix(), ext)
	filePath := filepath.Join(wh.cacheDir, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	// 下载图片
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := wh.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download wallpaper: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 确保目录存在
	if err := os.MkdirAll(wh.cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %v", err)
	}

	// 保存文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return filePath, nil
}

// GetRandomWallpaper 获取随机壁纸
func (wh *WallpaperHelper) GetRandomWallpaper() (string, error) {
	wallpapers, err := wh.GetWallpapers(20)
	if err != nil {
		return "", err
	}

	if len(wallpapers) == 0 {
		return "", fmt.Errorf("no wallpapers available")
	}

	// 随机选择一张
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(wallpapers))

	return wallpapers[index], nil
}

// ClearCache 清空缓存
func (wh *WallpaperHelper) ClearCache() error {
	if _, err := os.Stat(wh.cacheDir); os.IsNotExist(err) {
		return nil
	}

	return os.RemoveAll(wh.cacheDir)
}

// GetCacheSize 获取缓存大小
func (wh *WallpaperHelper) GetCacheSize() (int64, error) {
	var size int64

	err := filepath.Walk(wh.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// SetCacheDir 设置缓存目录
func (wh *WallpaperHelper) SetCacheDir(dir string) {
	wh.mutex.Lock()
	defer wh.mutex.Unlock()
	wh.cacheDir = dir
}

// GetCacheDir 获取缓存目录
func (wh *WallpaperHelper) GetCacheDir() string {
	wh.mutex.RLock()
	defer wh.mutex.RUnlock()
	return wh.cacheDir
}

// SetTimeout 设置HTTP客户端超时
func (wh *WallpaperHelper) SetTimeout(timeout time.Duration) {
	wh.mutex.Lock()
	defer wh.mutex.Unlock()
	wh.httpClient.Timeout = timeout
}

// GetTimeout 获取HTTP客户端超时
func (wh *WallpaperHelper) GetTimeout() time.Duration {
	wh.mutex.RLock()
	defer wh.mutex.RUnlock()
	return wh.httpClient.Timeout
}

// ValidateWallpaperURL 验证壁纸URL
func (wh *WallpaperHelper) ValidateWallpaperURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	// 简单的URL格式验证
	if !strings.Contains(url, ".") {
		return fmt.Errorf("invalid URL format")
	}

	return nil
}

// GetSupportedTypes 获取支持的壁纸类型
func (wh *WallpaperHelper) GetSupportedTypes() []string {
	return []string{
		string(WallpaperTypeBing),
		string(WallpaperTypeMediaServer),
		string(WallpaperTypeCustomize),
		string(WallpaperTypeTMDB),
	}
}

// IsTypeSupported 检查类型是否支持
func (wh *WallpaperHelper) IsTypeSupported(wallpaperType string) bool {
	supportedTypes := wh.GetSupportedTypes()
	for _, supportedType := range supportedTypes {
		if supportedType == wallpaperType {
			return true
		}
	}
	return false
}

// GetStats 获取统计信息
func (wh *WallpaperHelper) GetStats() map[string]interface{} {
	wh.mutex.RLock()
	defer wh.mutex.RUnlock()

	stats := map[string]interface{}{
		"wallpaper_type": wh.wallpaperType,
		"cache_dir":      wh.cacheDir,
		"timeout":        wh.httpClient.Timeout.String(),
	}

	// 添加缓存统计
	if cacheSize, err := wh.GetCacheSize(); err == nil {
		stats["cache_size"] = cacheSize
	}

	// 添加缓存文件数量
	if fileCount := wh.getCacheFileCount(); fileCount >= 0 {
		stats["cache_files"] = fileCount
	}

	return stats
}

// getCacheFileCount 获取缓存文件数量
func (wh *WallpaperHelper) getCacheFileCount() int {
	count := 0

	filepath.Walk(wh.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})

	return count
}
