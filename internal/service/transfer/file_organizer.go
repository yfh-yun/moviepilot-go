// Package transfer 提供文件整理服务
package transfer

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"

	"go.uber.org/zap"
)

// FileOrganizer 文件整理器
type FileOrganizer struct {
	mediaRepo    interfaces.MediaRepository
	transferRepo interfaces.TransferHistoryRepository
	logger       *zap.Logger

	// 整理配置
	config *OrganizerConfig

	// 命名规则
	namingRules []*NamingRule

	// 文件类型映射
	fileTypeMap map[string]FileType

	// 正则表达式缓存
	regexCache map[string]*regexp.Regexp
}

// OrganizerConfig 整理器配置
type OrganizerConfig struct {
	// 路径配置
	DownloadPath    string `json:"download_path" yaml:"download_path"`
	LibraryPath     string `json:"library_path" yaml:"library_path"`
	TVPath          string `json:"tv_path" yaml:"tv_path"`
	MoviePath       string `json:"movie_path" yaml:"movie_path"`
	AnimePath       string `json:"anime_path" yaml:"anime_path"`
	DocumentaryPath string `json:"documentary_path" yaml:"documentary_path"`

	// 目录结构配置
	EnableCategoryFolder   bool `json:"enable_category_folder" yaml:"enable_category_folder"`
	EnableYearFolder       bool `json:"enable_year_folder" yaml:"enable_year_folder"`
	EnableSeasonFolder     bool `json:"enable_season_folder" yaml:"enable_season_folder"`
	EnableQualityFolder    bool `json:"enable_quality_folder" yaml:"enable_quality_folder"`
	EnableResolutionFolder bool `json:"enable_resolution_folder" yaml:"enable_resolution_folder"`

	// 文件命名配置
	NamingFormat string `json:"naming_format" yaml:"naming_format"` // tv, movie, custom
	CustomNaming string `json:"custom_naming" yaml:"custom_naming"` // 自定义命名格式

	// 重命名配置
	EnableRename       bool `json:"enable_rename" yaml:"enable_rename"`
	ReplaceSpaces      bool `json:"replace_spaces" yaml:"replace_spaces"`
	ReplaceUnderscores bool `json:"replace_underscores" yaml:"replace_underscores"`
	RemoveSpecialChars bool `json:"remove_special_chars" yaml:"remove_special_chars"`

	// 重复文件处理
	DuplicateAction string `json:"duplicate_action" yaml:"duplicate_action"` // skip, replace, rename
	KeepBoth        bool   `json:"keep_both" yaml:"keep_both"`

	// 处理选项
	SkipExisting     bool `json:"skip_existing" yaml:"skip_existing"`
	CreateFolders    bool `json:"create_folders" yaml:"create_folders"`
	PreserveOriginal bool `json:"preserve_original" yaml:"preserve_original"`

	// 媒体类型映射
	CategoryMapping map[string]string `json:"category_mapping" yaml:"category_mapping"`
}

// FileType 文件类型
type FileType string

const (
	FileTypeVideo    FileType = "video"
	FileTypeAudio    FileType = "audio"
	FileTypeSubtitle FileType = "subtitle"
	FileTypeImage    FileType = "image"
	FileTypeNFO      FileType = "nfo"
	FileTypeOther    FileType = "other"
)

// NamingRule 命名规则
type NamingRule struct {
	Pattern    string            `json:"pattern"`
	Template   string            `json:"template"`
	Priority   int               `json:"priority"`
	Conditions map[string]string `json:"conditions"`
}

// TransferTask 整理任务
type TransferTask struct {
	ID          string            `json:"id"`
	FileItem    *model.FileItem   `json:"file_item"`
	MediaInfo   *model.MediaInfo  `json:"media_info,omitempty"`
	MetaInfo    *model.MetaInfo   `json:"meta_info,omitempty"`
	Options     *TransferOptions  `json:"options"`
	Progress    *TransferProgress `json:"progress"`
	Status      string            `json:"status"` // pending, running, completed, failed
	Error       string            `json:"error,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// TransferOptions 整理选项
type TransferOptions struct {
	SourcePath   string `json:"source_path"`
	TargetPath   string `json:"target_path"`
	Category     string `json:"category"`
	LibraryPath  string `json:"library_path"`
	DownloadPath string `json:"download_path"`

	// 重命名选项
	EnableRename   bool   `json:"enable_rename"`
	NamingFormat   string `json:"naming_format"`
	CustomTemplate string `json:"custom_template"`

	// 路径选项
	CreateFolders    bool `json:"create_folders"`
	PreserveOriginal bool `json:"preserve_original"`

	// 处理选项
	SkipExisting    bool   `json:"skip_existing"`
	DuplicateAction string `json:"duplicate_action"`

	// 文件过滤
	IncludePatterns []string `json:"include_patterns"`
	ExcludePatterns []string `json:"exclude_patterns"`
	MaxFileSize     int64    `json:"max_file_size"`
	MinFileSize     int64    `json:"min_file_size"`

	// 元数据
	Metadata map[string]string `json:"metadata"`
}

// TransferProgress 整理进度
type TransferProgress struct {
	TotalFiles     int       `json:"total_files"`
	ProcessedFiles int       `json:"processed_files"`
	FailedFiles    int       `json:"failed_files"`
	SkippedFiles   int       `json:"skipped_files"`
	TotalSize      int64     `json:"total_size"`
	ProcessedSize  int64     `json:"processed_size"`
	CurrentFile    string    `json:"current_file"`
	Progress       float64   `json:"progress"`
	StartTime      time.Time `json:"start_time"`
	EstimatedTime  time.Time `json:"estimated_time"`
}

// TransferResult 整理结果
type TransferResult struct {
	Success         bool             `json:"success"`
	TaskID          string           `json:"task_id"`
	Message         string           `json:"message"`
	ProcessedFiles  []ProcessedFile  `json:"processed_files"`
	SkippedFiles    []SkippedFile    `json:"skipped_files"`
	FailedFiles     []FailedFile     `json:"failed_files"`
	TransferSummary *TransferSummary `json:"transfer_summary"`
	Duration        time.Duration    `json:"duration"`
}

// ProcessedFile 已处理文件
type ProcessedFile struct {
	SourcePath     string            `json:"source_path"`
	TargetPath     string            `json:"target_path"`
	FileSize       int64             `json:"file_size"`
	FileType       FileType          `json:"file_type"`
	Action         string            `json:"action"` // copy, move, rename
	ProcessingTime time.Duration     `json:"processing_time"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// SkippedFile 跳过的文件
type SkippedFile struct {
	SourcePath string `json:"source_path"`
	FileSize   int64  `json:"file_size"`
	Reason     string `json:"reason"`
}

// FailedFile 失败的文件
type FailedFile struct {
	SourcePath string `json:"source_path"`
	FileSize   int64  `json:"file_size"`
	Error      string `json:"error"`
}

// TransferSummary 整理摘要
type TransferSummary struct {
	TotalFiles     int              `json:"total_files"`
	ProcessedFiles int              `json:"processed_files"`
	FailedFiles    int              `json:"failed_files"`
	SkippedFiles   int              `json:"skipped_files"`
	TotalSize      int64            `json:"total_size"`
	ProcessedSize  int64            `json:"processed_size"`
	FileTypeStats  map[FileType]int `json:"file_type_stats"`
	CategoryStats  map[string]int   `json:"category_stats"`
	Duration       time.Duration    `json:"duration"`
	SuccessRate    float64          `json:"success_rate"`
}

// NewFileOrganizer 创建文件整理器实例
func NewFileOrganizer(
	mediaRepo interfaces.MediaRepository,
	transferRepo interfaces.TransferHistoryRepository,
	config *OrganizerConfig,
) *FileOrganizer {
	organizer := &FileOrganizer{
		mediaRepo:    mediaRepo,
		transferRepo: transferRepo,
		logger:       logger.Logger,
		config:       config,
		fileTypeMap:  initFileTypeMap(),
		regexCache:   make(map[string]*regexp.Regexp),
	}

	// 初始化命名规则
	organizer.initNamingRules()

	return organizer
}

// OrganizeFiles 整理文件
func (fo *FileOrganizer) OrganizeFiles(ctx context.Context, task *TransferTask) (*TransferResult, error) {
	fo.logger.Info("开始文件整理",
		zap.String("task_id", task.ID),
		zap.String("source_path", task.FileItem.Path),
		zap.String("category", task.Options.Category),
	)

	startTime := time.Now()

	// 初始化进度
	task.Progress = &TransferProgress{
		TotalFiles: 1, // 单文件任务
		StartTime:  startTime,
		Progress:   0,
	}

	task.Status = "running"
	task.StartedAt = &startTime

	// 识别媒体信息（如果未提供）
	if task.MediaInfo == nil {
		mediaInfo, err := fo.identifyMedia(ctx, task.FileItem)
		if err != nil {
			return fo.createErrorResult(task, fmt.Sprintf("媒体识别失败: %v", err), time.Since(startTime))
		}
		task.MediaInfo = mediaInfo
	}

	// 构建目标路径
	targetPath, err := fo.buildTargetPath(ctx, task)
	if err != nil {
		return fo.createErrorResult(task, fmt.Sprintf("构建目标路径失败: %v", err), time.Since(startTime))
	}

	// 检查文件是否已存在
	if fo.fileExists(targetPath) {
		if task.Options.SkipExisting {
			skipped := &SkippedFile{
				SourcePath: task.FileItem.Path,
				FileSize:   task.FileItem.Size,
				Reason:     "文件已存在，跳过",
			}
			return fo.createSkipResult(task, []*SkippedFile{skipped}, time.Since(startTime))
		}

		// 处理重复文件
		targetPath, err = fo.handleDuplicate(ctx, task, targetPath)
		if err != nil {
			return fo.createErrorResult(task, fmt.Sprintf("处理重复文件失败: %v", err), time.Since(startTime))
		}
	}

	// 创建目标目录
	if task.Options.CreateFolders {
		if err := fo.createDirectory(filepath.Dir(targetPath)); err != nil {
			return fo.createErrorResult(task, fmt.Sprintf("创建目标目录失败: %v", err), time.Since(startTime))
		}
	}

	// 执行文件操作
	action := "copy"
	if !task.Options.PreserveOriginal {
		action = "move"
	}

	processingStart := time.Now()
	if err := fo.performFileOperation(task.FileItem.Path, targetPath, action); err != nil {
		return fo.createErrorResult(task, fmt.Sprintf("文件操作失败: %v", err), time.Since(startTime))
	}
	processingTime := time.Since(processingStart)

	// 更新进度
	task.Progress.ProcessedFiles = 1
	task.Progress.ProcessedSize = task.FileItem.Size
	task.Progress.Progress = 100

	// 创建整理记录
	record := &model.TransferHistory{
		ID:         utils.GenerateID(),
		SourcePath: task.FileItem.Path,
		TargetPath: targetPath,
		Action:     action,
		FileSize:   task.FileItem.Size,
		FileType:   string(fo.getFileType(task.FileItem.Name)),
		Category:   task.Options.Category,
		Status:     "completed",
		MediaID:    task.MediaInfo.ID,
		CreatedAt:  time.Now(),
	}

	if err := fo.transferRepo.Create(ctx, record); err != nil {
		fo.logger.Warn("保存整理记录失败", zap.Error(err))
	}

	// 记录处理文件
	processedFile := &ProcessedFile{
		SourcePath:     task.FileItem.Path,
		TargetPath:     targetPath,
		FileSize:       task.FileItem.Size,
		FileType:       fo.getFileType(task.FileItem.Name),
		Action:         action,
		ProcessingTime: processingTime,
		Metadata: map[string]string{
			"media_title": task.MediaInfo.Title,
			"media_year":  fmt.Sprintf("%d", task.MediaInfo.Year),
			"media_type":  task.MediaInfo.Type,
		},
	}

	// 更新任务状态
	task.Status = "completed"
	now := time.Now()
	task.CompletedAt = &now

	fo.logger.Info("文件整理完成",
		zap.String("task_id", task.ID),
		zap.String("source", task.FileItem.Path),
		zap.String("target", targetPath),
		zap.String("action", action),
		zap.Duration("duration", time.Since(startTime)),
	)

	return &TransferResult{
		Success:         true,
		TaskID:          task.ID,
		Message:         "文件整理完成",
		ProcessedFiles:  []*ProcessedFile{processedFile},
		TransferSummary: fo.createSummary([]*ProcessedFile{processedFile}, nil, nil, time.Since(startTime)),
		Duration:        time.Since(startTime),
	}, nil
}

// identifyMedia 识别媒体信息
func (fo *FileOrganizer) identifyMedia(ctx context.Context, fileItem *model.FileItem) (*model.MediaInfo, error) {
	fo.logger.Info("识别媒体信息", zap.String("filename", fileItem.Name))

	// 从文件名提取元数据
	metaInfo := fo.extractMetaFromFileName(fileItem.Name)
	if metaInfo == nil {
		return nil, fmt.Errorf("无法从文件名提取媒体信息")
	}

	// 查询媒体数据库
	mediaInfo, err := fo.mediaRepo.GetByTitleAndYear(ctx, metaInfo.Title, metaInfo.Year)
	if err != nil {
		return nil, fmt.Errorf("查询媒体信息失败: %w", err)
	}

	if mediaInfo == nil {
		// 创建基础媒体信息
		mediaInfo = &model.MediaInfo{
			ID:        utils.GenerateID(),
			Title:     metaInfo.Title,
			Year:      metaInfo.Year,
			Type:      metaInfo.Type,
			Season:    metaInfo.Season,
			Episodes:  metaInfo.Episodes,
			Category:  fo.mapCategory(metaInfo.Type),
			CreatedAt: time.Now(),
		}
	}

	fo.logger.Info("媒体信息识别成功",
		zap.String("title", mediaInfo.Title),
		zap.String("type", mediaInfo.Type),
		zap.Int("year", mediaInfo.Year),
	)

	return mediaInfo, nil
}

// buildTargetPath 构建目标路径
func (fo *FileOrganizer) buildTargetPath(ctx context.Context, task *TransferTask) (string, error) {
	mediaInfo := task.MediaInfo
	options := task.Options

	// 确定基础路径
	basePath := fo.getBasePath(mediaInfo.Type, options.LibraryPath)

	var pathComponents []string

	// 添加分类文件夹
	if fo.config.EnableCategoryFolder && options.Category != "" {
		pathComponents = append(pathComponents, options.Category)
	}

	// 添加年份文件夹
	if fo.config.EnableYearFolder && mediaInfo.Year > 0 {
		pathComponents = append(pathComponents, fmt.Sprintf("%d", mediaInfo.Year))
	}

	// 构建媒体名称
	mediaName := fo.generateMediaName(mediaInfo, options.NamingFormat)

	// 添加季文件夹（电视剧）
	if mediaInfo.Type == "tv" && fo.config.EnableSeasonFolder && mediaInfo.Season > 0 {
		pathComponents = append(pathComponents, mediaName)
		pathComponents = append(pathComponents, fmt.Sprintf("Season %02d", mediaInfo.Season))
	} else {
		pathComponents = append(pathComponents, mediaName)
	}

	// 构建完整目标路径
	targetDir := filepath.Join(pathComponents...)
	targetPath := filepath.Join(basePath, targetDir, task.FileItem.Name)

	// 如果启用重命名
	if options.EnableRename {
		newFileName := fo.generateFileName(mediaInfo, task.FileItem, options)
		targetPath = filepath.Join(filepath.Dir(targetPath), newFileName)
	}

	return targetPath, nil
}

// generateMediaName 生成媒体名称
func (fo *FileOrganizer) generateMediaName(mediaInfo *model.MediaInfo, namingFormat string) string {
	switch namingFormat {
	case "title":
		return mediaInfo.Title
	case "title_year":
		return fmt.Sprintf("%s (%d)", mediaInfo.Title, mediaInfo.Year)
	case "original":
		return mediaInfo.TitleYear
	default:
		// 默认格式
		if mediaInfo.Year > 0 {
			return fmt.Sprintf("%s (%d)", mediaInfo.Title, mediaInfo.Year)
		}
		return mediaInfo.Title
	}
}

// generateFileName 生成文件名
func (fo *FileOrganizer) generateFileName(mediaInfo *model.MediaInfo, fileItem *model.FileItem, options *string) string {
	ext := filepath.Ext(fileItem.Name)
	baseName := strings.TrimSuffix(fileItem.Name, ext)

	// 如果有自定义模板
	if options != nil && *options != "" {
		return fo.applyCustomTemplate(mediaInfo, baseName, ext, *options)
	}

	// 根据媒体类型生成文件名
	switch mediaInfo.Type {
	case "tv":
		if mediaInfo.Season > 0 && len(mediaInfo.Episodes) > 0 {
			// 电视剧格式: Title - S01E01
			return fmt.Sprintf("%s - S%02dE%02d%s",
				fo.cleanString(mediaInfo.Title),
				mediaInfo.Season,
				mediaInfo.Episodes[0],
				ext)
		}
		// 季度包格式: Title - Season 01
		return fmt.Sprintf("%s - Season %02d%s",
			fo.cleanString(mediaInfo.Title),
			mediaInfo.Season,
			ext)

	case "movie":
		// 电影格式: Title (Year)
		if mediaInfo.Year > 0 {
			return fmt.Sprintf("%s (%d)%s",
				fo.cleanString(mediaInfo.Title),
				mediaInfo.Year,
				ext)
		}
		return fmt.Sprintf("%s%s", fo.cleanString(mediaInfo.Title), ext)

	default:
		// 默认保持原文件名
		return fileItem.Name
	}
}

// applyCustomTemplate 应用自定义模板
func (fo *FileOrganizer) applyCustomTemplate(mediaInfo *model.MediaInfo, originalName, ext, template string) string {
	result := template

	// 替换模板变量
	replacements := map[string]string{
		"{title}":    fo.cleanString(mediaInfo.Title),
		"{year}":     fmt.Sprintf("%d", mediaInfo.Year),
		"{type}":     mediaInfo.Type,
		"{season}":   fmt.Sprintf("%02d", mediaInfo.Season),
		"{episode}":  fmt.Sprintf("%02d", fo.getFirstEpisode(mediaInfo.Episodes)),
		"{episodes}": fo.formatEpisodes(mediaInfo.Episodes),
		"{original}": originalName,
		"{category}": mediaInfo.Category,
	}

	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result + ext
}

// getBasePath 获取基础路径
func (fo *FileOrganizer) getBasePath(mediaType, libraryPath string) string {
	if libraryPath != "" {
		return libraryPath
	}

	switch mediaType {
	case "tv":
		if fo.config.TVPath != "" {
			return fo.config.TVPath
		}
	case "movie":
		if fo.config.MoviePath != "" {
			return fo.config.MoviePath
		}
	case "anime":
		if fo.config.AnimePath != "" {
			return fo.config.AnimePath
		}
	case "documentary":
		if fo.config.DocumentaryPath != "" {
			return fo.config.DocumentaryPath
		}
	}

	// 默认使用通用库路径
	if fo.config.LibraryPath != "" {
		return fo.config.LibraryPath
	}

	return fo.config.DownloadPath
}

// mapCategory 映射分类
func (fo *FileOrganizer) mapCategory(mediaType string) string {
	if fo.config.CategoryMapping != nil {
		if mapped, exists := fo.config.CategoryMapping[mediaType]; exists {
			return mapped
		}
	}
	return mediaType
}

// getFileType 获取文件类型
func (fo *FileOrganizer) getFileType(filename string) FileType {
	ext := strings.ToLower(filepath.Ext(filename))
	if fileType, exists := fo.fileTypeMap[ext]; exists {
		return fileType
	}
	return FileTypeOther
}

// extractMetaFromFileName 从文件名提取元数据
func (fo *FileOrganizer) extractMetaFromFileName(filename string) *model.MetaInfo {
	// 简化实现，实际应该使用复杂的正则表达式
	filename = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))

	// 尝试匹配电视剧格式 "Title S01E01"
	tvPattern := fo.getRegex(`(.+?)\s*[Ss](\d{1,2})[Ee](\d{1,2})`)
	if matches := tvPattern.FindStringSubmatch(filename); len(matches) >= 4 {
		return &model.MetaInfo{
			Title:    fo.cleanString(matches[1]),
			Season:   utils.ParseInt(matches[2]),
			Episodes: []int{utils.ParseInt(matches[3])},
			Type:     "tv",
		}
	}

	// 尝试匹配电影格式 "Title (Year)"
	moviePattern := fo.getRegex(`(.+?)\s*\((\d{4})\)`)
	if matches := moviePattern.FindStringSubmatch(filename); len(matches) >= 3 {
		return &model.MetaInfo{
			Title: fo.cleanString(matches[1]),
			Year:  utils.ParseInt(matches[2]),
			Type:  "movie",
		}
	}

	// 默认解析
	return &model.MetaInfo{
		Title: fo.cleanString(filename),
		Type:  "unknown",
	}
}

// handleDuplicate 处理重复文件
func (fo *FileOrganizer) handleDuplicate(ctx context.Context, task *TransferTask, targetPath string) (string, error) {
	switch task.Options.DuplicateAction {
	case "skip":
		return "", fmt.Errorf("文件已存在，跳过处理")
	case "replace":
		// 删除现有文件
		if err := utils.RemoveFile(targetPath); err != nil {
			return "", fmt.Errorf("删除现有文件失败: %w", err)
		}
		return targetPath, nil
	case "rename":
		// 生成新文件名
		ext := filepath.Ext(targetPath)
		base := strings.TrimSuffix(targetPath, ext)
		counter := 1

		for {
			newPath := fmt.Sprintf("%s (%d)%s", base, counter, ext)
			if !fo.fileExists(newPath) {
				return newPath, nil
			}
			counter++
		}
	default:
		return "", fmt.Errorf("未知的重复文件处理方式: %s", task.Options.DuplicateAction)
	}
}

// performFileOperation 执行文件操作
func (fo *FileOrganizer) performFileOperation(sourcePath, targetPath, action string) error {
	switch action {
	case "copy":
		return utils.CopyFile(sourcePath, targetPath)
	case "move":
		return utils.MoveFile(sourcePath, targetPath)
	default:
		return fmt.Errorf("不支持的文件操作: %s", action)
	}
}

// createDirectory 创建目录
func (fo *FileOrganizer) createDirectory(dirPath string) error {
	return utils.CreateDirectory(dirPath, 0755)
}

// fileExists 检查文件是否存在
func (fo *FileOrganizer) fileExists(filePath string) bool {
	return utils.FileExists(filePath)
}

// cleanString 清理字符串
func (fo *FileOrganizer) cleanString(s string) string {
	if fo.config.RemoveSpecialChars {
		// 移除特殊字符
		s = regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(s, "")
	}

	if fo.config.ReplaceSpaces {
		s = strings.ReplaceAll(s, " ", ".")
	}

	if fo.config.ReplaceUnderscores {
		s = strings.ReplaceAll(s, "_", " ")
	}

	return strings.TrimSpace(s)
}

// getRegex 获取缓存的正则表达式
func (fo *FileOrganizer) getRegex(pattern string) *regexp.Regexp {
	if regex, exists := fo.regexCache[pattern]; exists {
		return regex
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		fo.logger.Warn("正则表达式编译失败", zap.String("pattern", pattern), zap.Error(err))
		return regexp.MustCompile(`.*`) // 返回匹配所有
	}

	fo.regexCache[pattern] = regex
	return regex
}

// getFirstEpisode 获取第一集
func (fo *FileOrganizer) getFirstEpisode(episodes []int) int {
	if len(episodes) == 0 {
		return 0
	}
	sort.Ints(episodes)
	return episodes[0]
}

// formatEpisodes 格式化集数
func (fo *FileOrganizer) formatEpisodes(episodes []int) string {
	if len(episodes) == 0 {
		return ""
	}

	sort.Ints(episodes)
	var episodeStrs []string
	for _, ep := range episodes {
		episodeStrs = append(episodeStrs, fmt.Sprintf("%02d", ep))
	}

	if len(episodeStrs) == 1 {
		return episodeStrs[0]
	}

	return fmt.Sprintf("%s-%s", episodeStrs[0], episodeStrs[len(episodeStrs)-1])
}

// initFileTypeMap 初始化文件类型映射
func initFileTypeMap() map[string]FileType {
	return map[string]FileType{
		// 视频文件
		".mp4":  FileTypeVideo,
		".mkv":  FileTypeVideo,
		".avi":  FileTypeVideo,
		".mov":  FileTypeVideo,
		".wmv":  FileTypeVideo,
		".flv":  FileTypeVideo,
		".webm": FileTypeVideo,
		".m4v":  FileTypeVideo,
		".3gp":  FileTypeVideo,

		// 音频文件
		".mp3":  FileTypeAudio,
		".flac": FileTypeAudio,
		".aac":  FileTypeAudio,
		".ogg":  FileTypeAudio,
		".wav":  FileTypeAudio,
		".m4a":  FileTypeAudio,
		".wma":  FileTypeAudio,

		// 字幕文件
		".srt": FileTypeSubtitle,
		".ass": FileTypeSubtitle,
		".sub": FileTypeSubtitle,
		".vtt": FileTypeSubtitle,

		// 图片文件
		".jpg":  FileTypeImage,
		".jpeg": FileTypeImage,
		".png":  FileTypeImage,
		".gif":  FileTypeImage,
		".bmp":  FileTypeImage,
		".tiff": FileTypeImage,

		// NFO文件
		".nfo": FileTypeNFO,
	}
}

// initNamingRules 初始化命名规则
func (fo *FileOrganizer) initNamingRules() {
	fo.namingRules = []*NamingRule{
		{
			Pattern:  `(.+?)\s*[Ss](\d{1,2})[Ee](\d{1,2})`,
			Template: "{title} - S{season}E{episode}",
			Priority: 1,
		},
		{
			Pattern:  `(.+?)\s*\((\d{4})\)`,
			Template: "{title} ({year})",
			Priority: 2,
		},
	}
}

// createErrorResult 创建错误结果
func (fo *FileOrganizer) createErrorResult(task *TransferTask, message string, duration time.Duration) *TransferResult {
	task.Status = "failed"
	task.Error = message

	return &TransferResult{
		Success:  false,
		TaskID:   task.ID,
		Message:  message,
		Duration: duration,
	}
}

// createSkipResult 创建跳过结果
func (fo *FileOrganizer) createSkipResult(task *TransferTask, skippedFiles []*SkippedFile, duration time.Duration) *TransferResult {
	task.Status = "completed"

	return &TransferResult{
		Success:         true,
		TaskID:          task.ID,
		Message:         "文件跳过处理",
		SkippedFiles:    skippedFiles,
		TransferSummary: fo.createSummary(nil, skippedFiles, nil, duration),
		Duration:        duration,
	}
}

// createSummary 创建摘要
func (fo *FileOrganizer) createSummary(processed []*ProcessedFile, skipped []*SkippedFile, failed []*FailedFile, duration time.Duration) *TransferSummary {
	summary := &TransferSummary{
		ProcessedFiles: len(processed),
		SkippedFiles:   len(skipped),
		FailedFiles:    len(failed),
		FileTypeStats:  make(map[FileType]int),
		CategoryStats:  make(map[string]int),
		Duration:       duration,
	}

	summary.TotalFiles = summary.ProcessedFiles + summary.SkippedFiles + summary.FailedFiles
	summary.SuccessRate = float64(summary.ProcessedFiles) / float64(summary.TotalFiles) * 100

	// 统计文件类型
	for _, file := range processed {
		summary.ProcessedSize += file.FileSize
		summary.FileTypeStats[file.FileType]++
	}

	// 统计分类
	for _, file := range processed {
		if category, exists := file.Metadata["media_category"]; exists {
			summary.CategoryStats[category]++
		}
	}

	summary.TotalSize = summary.ProcessedSize

	return summary
}
