// Package implementations 提供具体动作实现
package implementations

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/business/workflows/base"
	"github.com/yfh-yun/moviepilot-go/internal/business/workflows/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/business/workflows/types"

	"go.uber.org/zap"
)

// ScanAction 文件扫描动作
// 负责扫描指定目录下的媒体文件
type ScanAction struct {
	*base.BaseAction

	// 文件服务
	fileService interfaces.FileService

	// 配置
	config *ScanConfig

	// 验证器
	validator *ScanValidator
}

// ScanConfig 扫描动作配置
type ScanConfig struct {
	// 扫描路径
	Paths []string `json:"paths"`

	// 扫描选项
	Recursive     bool     `json:"recursive"`
	FollowSymlinks bool     `json:"follow_symlinks"`
	MaxDepth      int      `json:"max_depth"`

	// 文件过滤
	Extensions    []string `json:"extensions"`
	MinSize       int64    `json:"min_size"`       // 最小文件大小（字节）
	MaxSize       int64    `json:"max_size"`       // 最大文件大小（字节）
	IncludeRegex  string   `json:"include_regex"`  // 包含正则表达式
	ExcludeRegex  string   `json:"exclude_regex"`  // 排除正则表达式
	ExcludePaths  []string `json:"exclude_paths"`  // 排除路径

	// 媒体文件识别
	VideoExtensions []string `json:"video_extensions"`
	AudioExtensions []string `json:"audio_extensions"`
	SubtitleExtensions []string `json:"subtitle_extensions"`

	// 元数据提取
	ExtractMetadata bool `json:"extract_metadata"`
	GenerateHash    bool `json:"generate_hash"`
	ExtractThumb    bool `json:"extract_thumb"`

	// 性能选项
	MaxConcurrency int           `json:"max_concurrency"`
	Timeout        time.Duration `json:"timeout"`
	BatchSize      int           `json:"batch_size"`

	// 缓存选项
	EnableCache     bool          `json:"enable_cache"`
	CacheTTL        time.Duration `json:"cache_ttl"`
	SkipExisting    bool          `json:"skip_existing"`
	CompareModified bool          `json:"compare_modified"`

	// 通知选项
	NotifyOnComplete bool `json:"notify_on_complete"`
	NotifyOnNewFile  bool `json:"notify_on_new_file"`
	NotifyOnError    bool `json:"notify_on_error"`
}

// ScanValidator 扫描动作验证器
type ScanValidator struct {
	rules map[string]ValidationRule
}

// NewScanAction 创建扫描动作
func NewScanAction(actionID string, fileService interfaces.FileService, cache interfaces.Cache) *ScanAction {
	baseAction := base.NewBaseAction(actionID, cache)
	
	config := &ScanConfig{
		Recursive:      true,
		MaxDepth:       10,
		MinSize:        1024, // 1KB
		MaxSize:        100 * 1024 * 1024 * 1024, // 100GB
		VideoExtensions: []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg", ".3gp"},
		AudioExtensions: []string{".mp3", ".flac", ".aac", ".ogg", ".wav", ".wma", ".m4a"},
		SubtitleExtensions: []string{".srt", ".ass", ".ssa", ".sub", ".vtt"},
		ExtractMetadata: true,
		GenerateHash:    false,
		ExtractThumb:    false,
		MaxConcurrency:  5,
		Timeout:         30 * time.Minute,
		BatchSize:       100,
		EnableCache:     true,
		CacheTTL:        24 * time.Hour,
		SkipExisting:    true,
		CompareModified: true,
		NotifyOnComplete: true,
		NotifyOnNewFile:  true,
		NotifyOnError:    true,
	}

	validator := &ScanValidator{
		rules: map[string]ValidationRule{
			"paths":         {Required: true, Type: "array"},
			"recursive":     {Required: false, Type: "bool"},
			"max_depth":     {Required: false, Type: "int", Min: 1, Max: 100},
			"min_size":      {Required: false, Type: "int", Min: 0},
			"max_size":      {Required: false, Type: "int", Min: 0},
			"max_concurrency": {Required: false, Type: "int", Min: 1, Max: 50},
			"timeout":       {Required: false, Type: "duration", Min: time.Second, Max: time.Hour * 24},
			"batch_size":    {Required: false, Type: "int", Min: 1, Max: 1000},
		},
	}

	action := &ScanAction{
		BaseAction:   baseAction,
		fileService:  fileService,
		config:       config,
		validator:    validator,
	}

	// 设置配置和验证器
	baseAction.SetConfig(configToMap(config))
	baseAction.SetValidator(validator)

	return action
}

// Name 获取动作名称
func (sa *ScanAction) Name() string {
	return "ScanAction"
}

// Description 获取动作描述
func (sa *ScanAction) Description() string {
	return "扫描指定目录下的媒体文件"
}

// Data 获取动作数据
func (sa *ScanAction) Data() map[string]interface{} {
	data := sa.BaseAction.Data()
	data["config"] = sa.config
	data["file_service"] = sa.fileService != nil
	return data
}

// doExecute 执行扫描动作
func (sa *ScanAction) doExecute(ctx context.Context, workflowID int64, params types.ActionParams, context *types.ActionContext) (*types.ActionContext, error) {
	sa.logger.Info("开始执行扫描动作",
		zap.Int64("workflow_id", workflowID),
		zap.Any("params", params))

	// 提取参数
	paths, ok := params["paths"].([]string)
	if !ok {
		if len(sa.config.Paths) > 0 {
			paths = sa.config.Paths
		} else {
			return context, fmt.Errorf("缺少必需参数: paths")
		}
	}

	// 合并配置
	config := sa.mergeConfig(params)

	// 执行扫描
	files, err := sa.scanFiles(ctx, paths, config)
	if err != nil {
		sa.SetError(fmt.Sprintf("文件扫描失败: %s", err.Error()))
		return context, err
	}

	// 添加到上下文
	for _, file := range files {
		context.AddFile(file)
	}

	// 添加日志
	context.AddLog("info", fmt.Sprintf("扫描完成，发现 %d 个文件", len(files)), "ScanAction", map[string]interface{}{
		"paths":     paths,
		"file_count": len(files),
		"config":    config,
	})

	// 设置完成状态
	sa.SetDone(fmt.Sprintf("扫描完成，发现 %d 个文件", len(files)))

	// 更新上下文进度
	context.UpdateProgress(100, "扫描动作完成")

	sa.logger.Info("扫描动作执行完成",
		zap.Strings("paths", paths),
		zap.Int("file_count", len(files)))

	return context, nil
}

// mergeConfig 合并配置
func (sa *ScanAction) mergeConfig(params types.ActionParams) *ScanConfig {
	config := *sa.config // 复制配置

	// 合并参数
	if recursive, ok := params["recursive"].(bool); ok {
		config.Recursive = recursive
	}
	if maxDepth, ok := params["max_depth"].(float64); ok {
		config.MaxDepth = int(maxDepth)
	}
	if minSize, ok := params["min_size"].(float64); ok {
		config.MinSize = int64(minSize)
	}
	if maxSize, ok := params["max_size"].(float64); ok {
		config.MaxSize = int64(maxSize)
	}
	if extensions, ok := params["extensions"].([]string); ok {
		config.Extensions = extensions
	}

	return &config
}

// scanFiles 扫描文件
func (sa *ScanAction) scanFiles(ctx context.Context, paths []string, config *ScanConfig) ([]*types.File, error) {
	var allFiles []*types.File

	for _, path := range paths {
		files, err := sa.scanPath(ctx, path, config)
		if err != nil {
			sa.logger.Warn("扫描路径失败",
				zap.String("path", path),
				zap.Error(err))
			continue
		}
		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

// scanPath 扫描单个路径
func (sa *ScanAction) scanPath(ctx context.Context, path string, config *ScanConfig) ([]*types.File, error) {
	sa.logger.Debug("开始扫描路径", zap.String("path", path))

	var files []*types.File

	// 检查路径是否存在
	if !sa.pathExists(path) {
		return nil, fmt.Errorf("路径不存在: %s", path)
	}

	// 扫描文件
	err := filepath.Walk(path, func(filePath string, info interface{}, err error) error {
		if err != nil {
			return err
		}

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 跳过目录
		fileInfo, ok := info.(interface{ IsDir() bool })
		if ok && fileInfo.IsDir() {
			return nil
		}

		// 检查是否应该包含此文件
		if !sa.shouldIncludeFile(filePath, config) {
			return nil
		}

		// 创建文件对象
		file, err := sa.createFile(filePath, config)
		if err != nil {
			sa.logger.Warn("创建文件对象失败",
				zap.String("path", filePath),
				zap.Error(err))
			return nil
		}

		files = append(files, file)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("扫描路径失败: %w", err)
	}

	sa.logger.Debug("路径扫描完成",
		zap.String("path", path),
		zap.Int("file_count", len(files)))

	return files, nil
}

// shouldIncludeFile 检查是否应该包含文件
func (sa *ScanAction) shouldIncludeFile(filePath string, config *ScanConfig) bool {
	// 检查扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	if len(config.Extensions) > 0 {
		found := false
		for _, allowedExt := range config.Extensions {
			if ext == allowedExt {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	} else {
		// 默认只包含媒体文件
		isMedia := false
		for _, mediaExt := range config.VideoExtensions {
			if ext == mediaExt {
				isMedia = true
				break
			}
		}
		if !isMedia {
			for _, mediaExt := range config.AudioExtensions {
				if ext == mediaExt {
					isMedia = true
					break
				}
			}
		}
		if !isMedia {
			for _, mediaExt := range config.SubtitleExtensions {
				if ext == mediaExt {
					isMedia = true
					break
				}
			}
		}
		if !isMedia {
			return false
		}
	}

	// 检查排除路径
	for _, excludePath := range config.ExcludePaths {
		if strings.Contains(filePath, excludePath) {
			return false
		}
	}

	return true
}

// createFile 创建文件对象
func (sa *ScanAction) createFile(filePath string, config *ScanConfig) (*types.File, error) {
	// 获取文件信息
	fileInfo, err := sa.getFileInfo(filePath)
	if err != nil {
		return nil, err
	}

	// 检查文件大小
	if config.MinSize > 0 && fileInfo.Size < config.MinSize {
		return nil, fmt.Errorf("文件太小: %d", fileInfo.Size)
	}
	if config.MaxSize > 0 && fileInfo.Size > config.MaxSize {
		return nil, fmt.Errorf("文件太大: %d", fileInfo.Size)
	}

	file := &types.File{
		Path:       filePath,
		Name:       filepath.Base(filePath),
		Extension:  strings.ToLower(filepath.Ext(filePath)),
		Size:       fileInfo.Size,
		Type:       sa.getFileType(filePath, config),
		Status:     "local",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 提取元数据
	if config.ExtractMetadata {
		sa.extractMetadata(file, config)
	}

	return file, nil
}

// getFileType 获取文件类型
func (sa *ScanAction) getFileType(filePath string, config *ScanConfig) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	for _, videoExt := range config.VideoExtensions {
		if ext == videoExt {
			return "video"
		}
	}

	for _, audioExt := range config.AudioExtensions {
		if ext == audioExt {
			return "audio"
		}
	}

	for _, subtitleExt := range config.SubtitleExtensions {
		if ext == subtitleExt {
			return "subtitle"
		}
	}

	return "unknown"
}

// extractMetadata 提取元数据
func (sa *ScanAction) extractMetadata(file *types.File, config *ScanConfig) {
	// 这里可以实现具体的元数据提取逻辑
	// 例如：从文件名解析标题、年份、季集信息等
	
	// 简化实现：从文件名提取基本信息
	fileName := strings.TrimSuffix(file.Name, file.Extension)
	
	// 尝试解析年份
	if year := sa.extractYear(fileName); year > 0 {
		file.Year = year
	}
	
	// 尝试解析季集信息
	if season, episodes := sa.extractSeasonEpisodes(fileName); season > 0 {
		file.Season = season
		file.Episodes = episodes
	}
}

// extractYear 提取年份
func (sa *ScanAction) extractYear(fileName string) int {
	// 简化的年份提取逻辑
	// 实际实现应该使用正则表达式
	return 0
}

// extractSeasonEpisodes 提取季集信息
func (sa *ScanAction) extractSeasonEpisodes(fileName string) (int, []int) {
	// 简化的季集提取逻辑
	// 实际实现应该使用正则表达式
	return 0, nil
}

// pathExists 检查路径是否存在
func (sa *ScanAction) pathExists(path string) bool {
	// 简化实现
	return true
}

// getFileInfo 获取文件信息
func (sa *ScanAction) getFileInfo(filePath string) (*FileInfo, error) {
	// 简化实现
	return &FileInfo{
		Size:    1024 * 1024, // 1MB
		ModTime: time.Now(),
	}, nil
}

// FileInfo 文件信息
type FileInfo struct {
	Size    int64
	ModTime time.Time
}

// ValidateConfig 验证配置
func (sv *ScanValidator) ValidateConfig(config map[string]interface{}) error {
	// 验证最大深度
	if maxDepth, ok := config["max_depth"].(float64); ok {
		if maxDepth < 1 || maxDepth > 100 {
			return fmt.Errorf("最大深度必须在1-100之间: %f", maxDepth)
		}
	}

	// 验证文件大小
	if minSize, ok := config["min_size"].(float64); ok {
		if minSize < 0 {
			return fmt.Errorf("最小文件大小不能为负数: %f", minSize)
		}
	}

	if maxSize, ok := config["max_size"].(float64); ok {
		if maxSize < 0 {
			return fmt.Errorf("最大文件大小不能为负数: %f", maxSize)
		}
	}

	return nil
}

// ValidateParams 验证参数
func (sv *ScanValidator) ValidateParams(params types.ActionParams) error {
	for paramName, rule := range sv.rules {
		value, exists := params[paramName]
		if !exists {
			if rule.Required {
				return fmt.Errorf("缺少必需参数: %s", paramName)
			}
			continue
		}

		// 类型验证
		switch rule.Type {
		case "string":
			if _, ok := value.(string); !ok {
				return fmt.Errorf("参数 %s 必须是字符串", paramName)
			}
		case "int":
			if _, ok := value.(int); !ok {
				if _, ok := value.(float64); !ok {
					return fmt.Errorf("参数 %s 必须是整数", paramName)
				}
			}
		case "bool":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("参数 %s 必须是布尔值", paramName)
			}
		case "array":
			if _, ok := value.([]interface{}); !ok {
				return fmt.Errorf("参数 %s 必须是数组", paramName)
			}
		case "duration":
			// 简化实现
		}

		// 范围验证
		if rule.Min != nil || rule.Max != nil {
			var numValue float64
			switch v := value.(type) {
			case int:
				numValue = float64(v)
			case float64:
				numValue = v
			default:
				continue
			}

			if rule.Min != nil && numValue < toFloat64(rule.Min) {
				return fmt.Errorf("参数 %s 不能小于 %v", paramName, rule.Min)
			}
			if rule.Max != nil && numValue > toFloat64(rule.Max) {
				return fmt.Errorf("参数 %s 不能大于 %v", paramName, rule.Max)
			}
		}
	}

	return nil
}

// ValidateContext 验证上下文
func (sv *ScanValidator) ValidateContext(context *types.ActionContext) error {
	if context == nil {
		return fmt.Errorf("上下文不能为空")
	}

	if context.WorkflowID <= 0 {
		return fmt.Errorf("无效的工作流ID: %d", context.WorkflowID)
	}

	return nil
}

// 辅助函数
func configToMap(config *ScanConfig) map[string]interface{} {
	return map[string]interface{}{
		"paths":              config.Paths,
		"recursive":          config.Recursive,
		"follow_symlinks":    config.FollowSymlinks,
		"max_depth":          config.MaxDepth,
		"extensions":         config.Extensions,
		"min_size":           config.MinSize,
		"max_size":           config.MaxSize,
		"include_regex":      config.IncludeRegex,
		"exclude_regex":      config.ExcludeRegex,
		"exclude_paths":      config.ExcludePaths,
		"video_extensions":   config.VideoExtensions,
		"audio_extensions":   config.AudioExtensions,
		"subtitle_extensions": config.SubtitleExtensions,
		"extract_metadata":   config.ExtractMetadata,
		"generate_hash":      config.GenerateHash,
		"extract_thumb":      config.ExtractThumb,
		"max_concurrency":    config.MaxConcurrency,
		"timeout":            config.Timeout,
		"batch_size":         config.BatchSize,
		"enable_cache":       config.EnableCache,
		"cache_ttl":          config.CacheTTL,
		"skip_existing":      config.SkipExisting,
		"compare_modified":   config.CompareModified,
		"notify_on_complete": config.NotifyOnComplete,
		"notify_on_new_file": config.NotifyOnNewFile,
		"notify_on_error":    config.NotifyOnError,
	}
}

// 确保实现接口
var _ interfaces.Action = (*ScanAction)(nil)
var _ interfaces.ActionValidator = (*ScanValidator)(nil)