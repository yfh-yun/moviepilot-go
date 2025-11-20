// Package actions 文件抓取相关业务逻辑
package actions

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
)

// ScrapeFileAction 文件抓取动作
type ScrapeFileAction struct {
	fileRepo     repository.FileRepository
	mediaRepo    repository.MediaRepository
	pluginMgr    service.PluginManager
	fileScanner  *FileScanner
	logger       logger.Logger
}

// NewScrapeFileAction 创建文件抓取动作实例
func NewScrapeFileAction(
	fileRepo repository.FileRepository,
	mediaRepo repository.MediaRepository,
	pluginMgr service.PluginManager,
	fileScanner *FileScanner,
	logger logger.Logger,
) *ScrapeFileAction {
	return &ScrapeFileAction{
		fileRepo:    fileRepo,
		mediaRepo:   mediaRepo,
		pluginMgr:   pluginMgr,
		fileScanner: fileScanner,
		logger:      logger,
	}
}

// Execute 执行文件抓取动作
func (a *ScrapeFileAction) Execute(ctx context.Context, req *ScrapeFileRequest) (*ScrapeFileResponse, error) {
	a.logger.Info("开始执行文件抓取动作",
		logger.String("path", req.Path),
		logger.String("user_id", req.UserID),
		logger.Strings("extensions", req.Extensions),
	)

	// 1. 验证请求参数
	if err := a.validateRequest(req); err != nil {
		a.logger.Error("请求参数验证失败", logger.Error(err))
		return nil, fmt.Errorf("请求参数验证失败: %w", err)
	}

	// 2. 扫描文件
	scannedFiles, err := a.scanFiles(ctx, req)
	if err != nil {
		a.logger.Error("文件扫描失败", logger.Error(err))
		return nil, fmt.Errorf("文件扫描失败: %w", err)
	}

	// 3. 文件识别和分类
	processedFiles, err := a.processFiles(ctx, scannedFiles, req)
	if err != nil {
		a.logger.Error("文件处理失败", logger.Error(err))
		return nil, fmt.Errorf("文件处理失败: %w", err)
	}

	// 4. 保存文件信息
	if err := a.saveFileInfos(ctx, processedFiles); err != nil {
		a.logger.Error("保存文件信息失败", logger.Error(err))
		return nil, fmt.Errorf("保存文件信息失败: %w", err)
	}

	// 5. 媒体匹配
	mediaMatches, err := a.matchMedia(ctx, processedFiles)
	if err != nil {
		a.logger.Error("媒体匹配失败", logger.Error(err))
		return nil, fmt.Errorf("媒体匹配失败: %w", err)
	}

	// 6. 返回结果
	response := &ScrapeFileResponse{
		Success:       true,
		ProcessedFiles: processedFiles,
		MediaMatches:  mediaMatches,
		TotalFiles:    len(processedFiles),
		Message:       "文件抓取完成",
		ScrapedAt:     time.Now(),
	}

	a.logger.Info("文件抓取动作执行完成",
		logger.String("path", req.Path),
		logger.Int("total_files", response.TotalFiles),
		logger.Int("media_matches", len(mediaMatches)),
	)

	return response, nil
}

// ScrapeFileRequest 文件抓取请求
type ScrapeFileRequest struct {
	UserID     string   `json:"user_id" validate:"required"`
	Path       string   `json:"path" validate:"required"`
	Extensions []string `json:"extensions"`
	Recursive  bool     `json:"recursive"`
	MinSize    int64    `json:"min_size"`
	MaxSize    int64    `json:"max_size"`
	ExcludeDirs[]string `json:"exclude_dirs"`
	AutoMatch  bool     `json:"auto_match"`
}

// ScrapeFileResponse 文件抓取响应
type ScrapeFileResponse struct {
	Success       bool                    `json:"success"`
	ProcessedFiles []repository.File       `json:"processed_files"`
	MediaMatches  []MediaMatchInfo        `json:"media_matches"`
	TotalFiles    int                     `json:"total_files"`
	Message       string                  `json:"message"`
	ScrapedAt     time.Time               `json:"scraped_at"`
}

// MediaMatchInfo 媒体匹配信息
type MediaMatchInfo struct {
	FileID     string    `json:"file_id"`
	FileTitle  string    `json:"file_title"`
	MediaID    string    `json:"media_id"`
	MediaTitle string    `json:"media_title"`
	MatchType  string    `json:"match_type"`
	Confidence float64   `json:"confidence"`
	MatchedAt  time.Time `json:"matched_at"`
}

// validateRequest 验证请求参数
func (a *ScrapeFileAction) validateRequest(req *ScrapeFileRequest) error {
	if req.Path == "" {
		return fmt.Errorf("抓取路径不能为空")
	}

	if len(req.Extensions) == 0 {
		// 使用默认的媒体文件扩展名
		req.Extensions = []string{
			".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm",
			".mp3", ".flac", ".wav", ".aac", ".ogg",
		}
	}

	if req.MinSize <= 0 {
		req.MinSize = 1024 * 1024 // 1MB
	}

	return nil
}

// scanFiles 扫描文件
func (a *ScrapeFileAction) scanFiles(ctx context.Context, req *ScrapeFileRequest) ([]string, error) {
	scanOptions := &ScanOptions{
		Path:         req.Path,
		Extensions:   req.Extensions,
		Recursive:    req.Recursive,
		MinSize:      req.MinSize,
		MaxSize:      req.MaxSize,
		ExcludeDirs:  req.ExcludeDirs,
	}

	files, err := a.fileScanner.ScanFiles(ctx, scanOptions)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// processFiles 处理文件
func (a *ScrapeFileAction) processFiles(ctx context.Context, files []string, req *ScrapeFileRequest) ([]repository.File, error) {
	var processedFiles []repository.File

	for _, filePath := range files {
		fileInfo, err := a.processFile(ctx, filePath, req)
		if err != nil {
			a.logger.Error("处理文件失败", 
				logger.String("file_path", filePath),
				logger.Error(err))
			continue
		}

		processedFiles = append(processedFiles, *fileInfo)
	}

	return processedFiles, nil
}

// processFile 处理单个文件
func (a *ScrapeFileAction) processFile(ctx context.Context, filePath string, req *ScrapeFileRequest) (*repository.File, error) {
	// 获取文件基本信息
	fileInfo, err := a.fileScanner.GetFileInfo(ctx, filePath)
	if err != nil {
		return nil, err
	}

	// 文件类型识别
	fileType := a.identifyFileType(filePath, fileInfo.Extension)

	// 提取文件元数据
	metadata, err := a.extractFileMetadata(ctx, filePath, fileType)
	if err != nil {
		a.logger.Error("提取文件元数据失败", 
			logger.String("file_path", filePath),
			logger.Error(err))
		metadata = make(map[string]interface{})
	}

	// 计算文件哈希
	hash, err := a.fileScanner.CalculateFileHash(ctx, filePath)
	if err != nil {
		a.logger.Error("计算文件哈希失败", 
			logger.String("file_path", filePath),
			logger.Error(err))
		hash = ""
	}

	// 创建文件记录
	file := &repository.File{
		Path:         filePath,
		Name:         fileInfo.Name,
		Extension:    fileInfo.Extension,
		Size:         fileInfo.Size,
		Type:         fileType,
		Hash:         hash,
		Metadata:     metadata,
		ScrapedBy:    req.UserID,
		ScrapedAt:    time.Now(),
		LastModified: fileInfo.ModTime,
	}

	return file, nil
}

// identifyFileType 识别文件类型
func (a *ScrapeFileAction) identifyFileType(filePath, extension string) string {
	extension = strings.ToLower(extension)

	videoExts := map[string]bool{
		".mp4": true, ".mkv": true, ".avi": true, ".mov": true,
		".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
		".3gp": true, ".ogv": true, ".ts": true, ".mts": true,
	}

	audioExts := map[string]bool{
		".mp3": true, ".flac": true, ".wav": true, ".aac": true,
		".ogg": true, ".wma": true, ".m4a": true, ".opus": true,
	}

	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".tiff": true, ".webp": true, ".svg": true,
	}

	switch {
	case videoExts[extension]:
		return "video"
	case audioExts[extension]:
		return "audio"
	case imageExts[extension]:
		return "image"
	default:
		return "other"
	}
}

// extractFileMetadata 提取文件元数据
func (a *ScrapeFileAction) extractFileMetadata(ctx context.Context, filePath, fileType string) (map[string]interface{}, error) {
	switch fileType {
	case "video":
		return a.extractVideoMetadata(ctx, filePath)
	case "audio":
		return a.extractAudioMetadata(ctx, filePath)
	default:
		return make(map[string]interface{}), nil
	}
}

// extractVideoMetadata 提取视频元数据
func (a *ScrapeFileAction) extractVideoMetadata(ctx context.Context, filePath string) (map[string]interface{}, error) {
	result, err := a.pluginMgr.CallPlugin(ctx, "media", "extract_video_metadata", map[string]interface{}{
		"file_path": filePath,
	})

	if err != nil {
		return nil, err
	}

	metadata, ok := result["metadata"].(map[string]interface{})
	if !ok {
		return make(map[string]interface{}), nil
	}

	return metadata, nil
}

// extractAudioMetadata 提取音频元数据
func (a *ScrapeFileAction) extractAudioMetadata(ctx context.Context, filePath string) (map[string]interface{}, error) {
	result, err := a.pluginMgr.CallPlugin(ctx, "media", "extract_audio_metadata", map[string]interface{}{
		"file_path": filePath,
	})

	if err != nil {
		return nil, err
	}

	metadata, ok := result["metadata"].(map[string]interface{})
	if !ok {
		return make(map[string]interface{}), nil
	}

	return metadata, nil
}

// saveFileInfos 保存文件信息
func (a *ScrapeFileAction) saveFileInfos(ctx context.Context, files []repository.File) error {
	for _, file := range files {
		// 检查文件是否已存在（基于哈希）
		existingFile, err := a.fileRepo.GetByHash(ctx, file.Hash)
		if err == nil && existingFile != nil {
			// 更新现有记录
			file.ID = existingFile.ID
			if err := a.fileRepo.Update(ctx, &file); err != nil {
				a.logger.Error("更新文件记录失败", 
					logger.String("file_id", file.ID),
					logger.Error(err))
			}
		} else {
			// 创建新记录
			if err := a.fileRepo.Create(ctx, &file); err != nil {
				a.logger.Error("创建文件记录失败", 
					logger.String("file_path", file.Path),
					logger.Error(err))
			}
		}
	}

	return nil
}

// matchMedia 匹配媒体
func (a *ScrapeFileAction) matchMedia(ctx context.Context, files []repository.File) ([]MediaMatchInfo, error) {
	var mediaMatches []MediaMatchInfo

	for _, file := range files {
		if file.Type != "video" {
			continue
		}

		// 调用插件进行媒体匹配
		result, err := a.pluginMgr.CallPlugin(ctx, "media", "match_media", map[string]interface{}{
			"file_path": file.Path,
			"file_name": file.Name,
			"metadata":  file.Metadata,
		})

		if err != nil {
			a.logger.Error("媒体匹配失败", 
				logger.String("file_path", file.Path),
				logger.Error(err))
			continue
		}

		match, ok := result["match"].(map[string]interface{})
		if !ok || result["confidence"].(float64) < 0.7 {
			continue
		}

		mediaMatch := MediaMatchInfo{
			FileID:     file.ID,
			FileTitle:  file.Name,
			MediaID:    getString(match, "media_id"),
			MediaTitle: getString(match, "media_title"),
			MatchType:  getString(match, "match_type"),
			Confidence: result["confidence"].(float64),
			MatchedAt:  time.Now(),
		}

		mediaMatches = append(mediaMatches, mediaMatch)
	}

	return mediaMatches, nil
}