package subtitle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/utils"
)

// SubtitleManager 字幕管理器
type SubtitleManager struct {
	providers map[string]Provider
	config    *Config
}

// Provider 字幕提供者接口
type Provider interface {
	Name() string
	Search(ctx context.Context, req *SearchRequest) ([]*Subtitle, error)
	Download(ctx context.Context, subtitle *Subtitle) ([]byte, error)
	Test(ctx context.Context) error
}

// Config 字幕配置
type Config struct {
	EnabledProviders []string          `json:"enabled_providers"`
	Languages        []string          `json:"languages"`
	DownloadPath     string            `json:"download_path"`
	AutoDownload     bool              `json:"auto_download"`
	MaxResults       int               `json:"max_results"`
	Timeout          time.Duration     `json:"timeout"`
	ProviderConfigs  map[string]interface{} `json:"provider_configs"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Title      string    `json:"title"`
	Year       int       `json:"year"`
	Season     int       `json:"season"`
	Episode    int       `json:"episode"`
	Language   string    `json:"language"`
	FileHash   string    `json:"file_hash"`
	FileSize   int64     `json:"file_size"`
	IMDBID     string    `json:"imdb_id"`
}

// Subtitle 字幕信息
type Subtitle struct {
	ID          string            `json:"id"`
	Provider    string            `json:"provider"`
	Title       string            `json:"title"`
	Year        int               `json:"year"`
	Season      int               `json:"season"`
	Episode     int               `json:"episode"`
	Language    string            `json:"language"`
	Format      string            `json:"format"`
	Rate        float64           `json:"rate"`
	Downloads   int               `json:"downloads"`
	UploadDate  time.Time         `json:"upload_date"`
	FileURL     string            `json:"file_url"`
	FileSize    int64             `json:"file_size"`
	Encoding    string            `json:"encoding"`
	Extra       map[string]string `json:"extra"`
}

// NewSubtitleManager 创建字幕管理器
func NewSubtitleManager(config *Config) *SubtitleManager {
	if config == nil {
		config = &Config{
			EnabledProviders: []string{"opensubtitles", "subhd"},
			Languages:        []string{"zh", "en"},
			AutoDownload:     false,
			MaxResults:       20,
			Timeout:          30 * time.Second,
		}
	}

	sm := &SubtitleManager{
		providers: make(map[string]Provider),
		config:    config,
	}

	// 注册提供者
	sm.registerProviders()

	return sm
}

// registerProviders 注册字幕提供者
func (sm *SubtitleManager) registerProviders() {
	// OpenSubtitles
	if sm.isProviderEnabled("opensubtitles") {
		sm.providers["opensubtitles"] = NewOpenSubtitlesProvider(sm.config.ProviderConfigs["opensubtitles"])
	}

	// SubHD
	if sm.isProviderEnabled("subhd") {
		sm.providers["subhd"] = NewSubHDProvider(sm.config.ProviderConfigs["subhd"])
	}

	// Shooter
	if sm.isProviderEnabled("shooter") {
		sm.providers["shooter"] = NewShooterProvider(sm.config.ProviderConfigs["shooter"])
	}

	// Xunlei
	if sm.isProviderEnabled("xunlei") {
		sm.providers["xunlei"] = NewXunleiProvider(sm.config.ProviderConfigs["xunlei"])
	}
}

// isProviderEnabled 检查提供者是否启用
func (sm *SubtitleManager) isProviderEnabled(name string) bool {
	for _, provider := range sm.config.EnabledProviders {
		if provider == name {
			return true
		}
	}
	return false
}

// Search 搜索字幕
func (sm *SubtitleManager) Search(ctx context.Context, req *SearchRequest) ([]*Subtitle, error) {
	var allSubtitles []*Subtitle

	// 并发搜索所有启用的提供者
	resultChan := make(chan []*Subtitle, len(sm.providers))
	errorChan := make(chan error, len(sm.providers))

	for name, provider := range sm.providers {
		go func(name string, provider Provider) {
			subtitles, err := provider.Search(ctx, req)
			if err != nil {
				logger.Error("Failed to search subtitles from %s: %v", name, err)
				errorChan <- err
				return
			}
			resultChan <- subtitles
		}(name, provider)
	}

	// 收集结果
	for i := 0; i < len(sm.providers); i++ {
		select {
		case subtitles := <-resultChan:
			allSubtitles = append(allSubtitles, subtitles...)
		case err := <-errorChan:
			logger.Warn("Provider search error: %v", err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 去重和排序
	allSubtitles = sm.deduplicateAndSort(allSubtitles)

	// 限制结果数量
	if sm.config.MaxResults > 0 && len(allSubtitles) > sm.config.MaxResults {
		allSubtitles = allSubtitles[:sm.config.MaxResults]
	}

	return allSubtitles, nil
}

// Download 下载字幕
func (sm *SubtitleManager) Download(ctx context.Context, subtitle *Subtitle) (string, error) {
	provider, exists := sm.providers[subtitle.Provider]
	if !exists {
		return "", fmt.Errorf("provider %s not found", subtitle.Provider)
	}

	// 下载字幕数据
	data, err := provider.Download(ctx, subtitle)
	if err != nil {
		return "", fmt.Errorf("failed to download subtitle: %w", err)
	}

	// 生成文件名
	filename := sm.generateFilename(subtitle)
	filepath := filepath.Join(sm.config.DownloadPath, filename)

	// 保存文件
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save subtitle file: %w", err)
	}

	logger.Info("Subtitle downloaded: %s", filepath)
	return filepath, nil
}

// AutoDownload 自动下载字幕
func (sm *SubtitleManager) AutoDownload(ctx context.Context, mediaFile string) ([]string, error) {
	if !sm.config.AutoDownload {
		return nil, fmt.Errorf("auto download is disabled")
	}

	// 解析媒体文件信息
	mediaInfo, err := sm.parseMediaFile(mediaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse media file: %w", err)
	}

	var downloadedFiles []string

	// 为每种语言搜索字幕
	for _, language := range sm.config.Languages {
		req := &SearchRequest{
			Title:    mediaInfo.Title,
			Year:     mediaInfo.Year,
			Season:   mediaInfo.Season,
			Episode:  mediaInfo.Episode,
			Language: language,
			FileHash: mediaInfo.Hash,
			FileSize: mediaInfo.Size,
		}

		subtitles, err := sm.Search(ctx, req)
		if err != nil {
			logger.Warn("Failed to search subtitles for language %s: %v", language, err)
			continue
		}

		// 下载最佳匹配的字幕
		if len(subtitles) > 0 {
			bestSubtitle := sm.selectBestSubtitle(subtitles, mediaInfo)
			filepath, err := sm.Download(ctx, bestSubtitle)
			if err != nil {
				logger.Warn("Failed to download subtitle: %v", err)
				continue
			}
			downloadedFiles = append(downloadedFiles, filepath)
		}
	}

	return downloadedFiles, nil
}

// deduplicateAndSort 去重和排序
func (sm *SubtitleManager) deduplicateAndSort(subtitles []*Subtitle) []*Subtitle {
	seen := make(map[string]bool)
	var result []*Subtitle

	for _, subtitle := range subtitles {
		key := fmt.Sprintf("%s_%s_%s", subtitle.Provider, subtitle.ID, subtitle.Language)
		if !seen[key] {
			seen[key] = true
			result = append(result, subtitle)
		}
	}

	// 按评分和下载量排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if sm.compareSubtitles(result[i], result[j]) < 0 {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// compareSubtitles 比较两个字幕
func (sm *SubtitleManager) compareSubtitles(a, b *Subtitle) int {
	// 优先比较评分
	if a.Rate != b.Rate {
		if a.Rate > b.Rate {
			return 1
		}
		return -1
	}

	// 其次比较下载量
	if a.Downloads != b.Downloads {
		if a.Downloads > b.Downloads {
			return 1
		}
		return -1
	}

	// 最后比较上传时间
	if a.UploadDate.After(b.UploadDate) {
		return 1
	}
	return -1
}

// selectBestSubtitle 选择最佳字幕
func (sm *SubtitleManager) selectBestSubtitle(subtitles []*Subtitle, mediaInfo *MediaInfo) *Subtitle {
	// 简单选择第一个，实际可以根据更多条件选择
	if len(subtitles) > 0 {
		return subtitles[0]
	}
	return nil
}

// generateFilename 生成字幕文件名
func (sm *SubtitleManager) generateFilename(subtitle *Subtitle) string {
	filename := subtitle.Title

	if subtitle.Year > 0 {
		filename += fmt.Sprintf(".%d", subtitle.Year)
	}

	if subtitle.Season > 0 && subtitle.Episode > 0 {
		filename += fmt.Sprintf(".S%02dE%02d", subtitle.Season, subtitle.Episode)
	}

	filename += fmt.Sprintf(".%s.%s", subtitle.Language, subtitle.Format)

	return filename
}

// parseMediaFile 解析媒体文件
func (sm *SubtitleManager) parseMediaFile(filepath string) (*MediaInfo, error) {
	filename := filepath.Base(filepath)
	
	// 提取文件信息
	info := &MediaInfo{
		Title: utils.ExtractTitle(filename),
		Year:  utils.ExtractYear(filename),
		Season: utils.ExtractSeason(filename),
		Episode: utils.ExtractEpisode(filename),
	}

	// 获取文件大小和哈希
	if stat, err := os.Stat(filepath); err == nil {
		info.Size = stat.Size()
		info.Hash = utils.CalculateFileHash(filepath)
	}

	return info, nil
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Title  string `json:"title"`
	Year   int    `json:"year"`
	Season int    `json:"season"`
	Episode int   `json:"episode"`
	Size   int64  `json:"size"`
	Hash   string `json:"hash"`
}

// GetProviders 获取所有提供者
func (sm *SubtitleManager) GetProviders() map[string]Provider {
	return sm.providers
}

// TestProvider 测试提供者
func (sm *SubtitleManager) TestProvider(ctx context.Context, providerName string) error {
	provider, exists := sm.providers[providerName]
	if !exists {
		return fmt.Errorf("provider %s not found", providerName)
	}

	return provider.Test(ctx)
}

// TestAllProviders 测试所有提供者
func (sm *SubtitleManager) TestAllProviders(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for name, provider := range sm.providers {
		err := provider.Test(ctx)
		results[name] = err
		if err != nil {
			logger.Error("Provider %s test failed: %v", name, err)
		} else {
			logger.Info("Provider %s test passed", name)
		}
	}

	return results
}