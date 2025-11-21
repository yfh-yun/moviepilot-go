// Package actions 提供下载参数验证器的实现
package actions

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"moviepilot-go/internal/business/services/actions/types"
	"moviepilot-go/pkg/logger"
)

// ValidationError 验证错误类型
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("字段 '%s': %s", e.Field, e.Message)
}

// DownloadValidator 下载参数验证器
type DownloadValidator struct {
	logger logger.Logger
	// 支持的下载器类型
	supportedDownloaders map[string]bool
	// 支持的URL模式
	magnetPattern *regexp.Regexp
	torrentPattern *regexp.Regexp
}

// NewDownloadValidator 创建下载验证器实例
func NewDownloadValidator() *DownloadValidator {
	return &DownloadValidator{
		logger:               logger.NewLogger("download_validator"),
		supportedDownloaders: map[string]bool{"qbittorrent": true, "transmission": true},
		magnetPattern:        regexp.MustCompile(`^magnet:\?xt=urn:btih:[a-zA-Z0-9]+`),
		torrentPattern:       regexp.MustCompile(`\.torrent$`),
	}
}

// ValidateDownloadParams 验证下载参数
func (v *DownloadValidator) ValidateDownloadParams(params *AddDownloadParams) error {
	// 验证下载器类型
	if err := v.validateDownloader(params.Downloader); err != nil {
		return err
	}

	// 验证保存路径
	if params.SavePath != "" {
		if err := v.validateSavePath(params.SavePath); err != nil {
			return err
		}
	}

	// 验证标签
	if err := v.validateLabels(params.Labels); err != nil {
		return err
	}

	// 验证质量和分辨率
	if err := v.validateQuality(params.Quality); err != nil {
		return err
	}

	if err := v.validateResolution(params.Resolution); err != nil {
		return err
	}

	return nil
}

// ValidateTorrent 验证种子信息
func (v *DownloadValidator) ValidateTorrent(torrent *types.Torrent) error {
	// 验证URL
	if err := v.validateTorrentURL(torrent.URL); err != nil {
		return fmt.Errorf("无效的种子URL '%s': %w", torrent.URL, err)
	}

	// 验证标题
	if err := v.validateTorrentTitle(torrent.Title); err != nil {
		return fmt.Errorf("无效的种子标题: %w", err)
	}

	// 验证大小（可选）
	if torrent.Size < 0 {
		return fmt.Errorf("无效的种子大小: %d", torrent.Size)
	}

	return nil
}

// validateDownloader 验证下载器类型
func (v *DownloadValidator) validateDownloader(downloader string) error {
	if downloader == "" {
		return &ValidationError{Field: "downloader", Message: "下载器类型不能为空"}
	}

	if !v.supportedDownloaders[strings.ToLower(downloader)] {
		supportedList := make([]string, 0, len(v.supportedDownloaders))
		for d := range v.supportedDownloaders {
			supportedList = append(supportedList, d)
		}
		return &ValidationError{
			Field:   "downloader",
			Message: fmt.Sprintf("不支持的下载器类型，支持的类型: %s", strings.Join(supportedList, ", ")),
		}
	}

	return nil
}

// validateSavePath 验证保存路径
func (v *DownloadValidator) validateSavePath(path string) error {
	// 检查路径格式
	if filepath.Base(path) == path {
		return &ValidationError{Field: "save_path", Message: "保存路径必须是绝对路径"}
	}

	// 检查路径安全性
	if strings.Contains(path, "..") {
		return &ValidationError{Field: "save_path", Message: "保存路径不能包含相对路径"}
	}

	// 检查路径字符
	if matched, _ := regexp.MatchString(`[<>"|?*]`, path); matched {
		return &ValidationError{Field: "save_path", Message: "保存路径包含非法字符"}
	}

	return nil
}

// validateLabels 验证标签
func (v *DownloadValidator) validateLabels(labels []string) error {
	for i, label := range labels {
		if strings.TrimSpace(label) == "" {
			return &ValidationError{Field: fmt.Sprintf("labels[%d]", i), Message: "标签不能为空字符串"}
		}

		if len(label) > 50 {
			return &ValidationError{Field: fmt.Sprintf("labels[%d]", i), Message: "标签长度不能超过50个字符"}
		}

		// 检查标签字符
		if matched, _ := regexp.MatchString(`[\/\\]`, label); matched {
			return &ValidationError{Field: fmt.Sprintf("labels[%d]", i), Message: "标签不能包含斜杠字符"}
		}
	}

	return nil
}

// validateQuality 验证质量
func (v *DownloadValidator) validateQuality(quality string) error {
	if quality == "" {
		return nil // 质量是可选的
	}

	// 支持的质量选项
	supportedQualities := []string{"HD", "SD", "1080p", "720p", "4K", "8K", "DVD", "BluRay", "WEB-DL"}
	valid := false
	for _, q := range supportedQualities {
		if strings.EqualFold(q, quality) {
			valid = true
			break
		}
	}

	if !valid {
		return &ValidationError{Field: "quality", Message: fmt.Sprintf("不支持的质量选项: %s", quality)}
	}

	return nil
}

// validateResolution 验证分辨率
func (v *DownloadValidator) validateResolution(resolution string) error {
	if resolution == "" {
		return nil // 分辨率是可选的
	}

	// 检查分辨率格式
	pattern := regexp.MustCompile(`^\d+p$|^\d+\sx\s\d+$`)
	if !pattern.MatchString(resolution) {
		return &ValidationError{Field: "resolution", Message: fmt.Sprintf("无效的分辨率格式: %s，支持格式如 '1080p' 或 '1920 x 1080'", resolution)}
	}

	return nil
}

// validateTorrentURL 验证种子URL
func (v *DownloadValidator) validateTorrentURL(urlStr string) error {
	if urlStr == "" {
		return &ValidationError{Field: "url", Message: "种子URL不能为空"}
	}

	// 检查是否是磁力链接
	if v.magnetPattern.MatchString(urlStr) {
		return nil // 磁力链接格式正确
	}

	// 检查是否是.torrent文件链接
	if v.torrentPattern.MatchString(urlStr) {
		// 检查URL格式
		if _, err := url.Parse(urlStr); err != nil {
			return &ValidationError{Field: "url", Message: "无效的torrent文件URL"}
		}
		return nil
	}

	// 尝试作为普通URL解析
	if _, err := url.Parse(urlStr); err != nil {
		return &ValidationError{Field: "url", Message: "无效的URL格式"}
	}

	return nil
}

// validateTorrentTitle 验证种子标题
func (v *DownloadValidator) validateTorrentTitle(title string) error {
	if title == "" {
		return &ValidationError{Field: "title", Message: "种子标题不能为空"}
	}

	if len(title) > 500 {
		return &ValidationError{Field: "title", Message: "种子标题长度不能超过500个字符"}
	}

	return nil
}

// ValidateBatchDownloads 批量验证多个种子
func (v *DownloadValidator) ValidateBatchDownloads(torrents []*types.Torrent) []ValidationResult {
	results := make([]ValidationResult, 0, len(torrents))

	for i, torrent := range torrents {
		err := v.ValidateTorrent(torrent)
		results = append(results, ValidationResult{
			Index:   i,
			Torrent: torrent,
			Valid:   err == nil,
			Error:   err,
		})
	}

	return results
}

// GetValidTorrents 从验证结果中获取有效的种子
func (v *DownloadValidator) GetValidTorrents(results []ValidationResult) []*types.Torrent {
	validTorrents := make([]*types.Torrent, 0)
	for _, result := range results {
		if result.Valid {
			validTorrents = append(validTorrents, result.Torrent)
		}
	}
	return validTorrents
}

// ValidationResult 验证结果
type ValidationResult struct {
	Index   int              `json:"index"`
	Torrent *types.Torrent   `json:"torrent,omitempty"`
	Valid   bool             `json:"valid"`
	Error   error            `json:"error,omitempty"`
}

// IsValidAll 检查是否全部验证通过
func (v *DownloadValidator) IsValidAll(results []ValidationResult) bool {
	for _, result := range results {
		if !result.Valid {
			return false
		}
	}
	return true
}

// GetValidationErrors 获取所有验证错误
func (v *DownloadValidator) GetValidationErrors(results []ValidationResult) []string {
	errors := make([]string, 0)
	for _, result := range results {
		if !result.Valid && result.Error != nil {
			errors = append(errors, fmt.Sprintf("索引 %d: %v", result.Index, result.Error))
		}
	}
	return errors
}
