package media

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// Organizer 媒体整理器
type Organizer struct {
	recognizer *Recognizer
	logger     *zap.Logger
}

// NewOrganizer 创建媒体整理器
func NewOrganizer(recognizer *Recognizer, logger *zap.Logger) *Organizer {
	return &Organizer{
		recognizer: recognizer,
		logger:     logger,
	}
}

// OrganizeOptions 整理选项
type OrganizeOptions struct {
	SourcePath      string // 源路径
	DestPath        string // 目标路径
	UseHardLink     bool   // 使用硬链接
	CreateDirectory bool   // 创建目录
	RenameFile      bool   // 重命名文件
	Template        string // 命名模板
}

// OrganizeResult 整理结果
type OrganizeResult struct {
	SourcePath string
	DestPath   string
	Success    bool
	Error      error
}

// Organize 整理媒体文件
func (o *Organizer) Organize(ctx context.Context, opts OrganizeOptions) (*OrganizeResult, error) {
	result := &OrganizeResult{
		SourcePath: opts.SourcePath,
	}

	// 识别媒体信息
	info, err := o.recognizer.RecognizeFromPath(ctx, opts.SourcePath)
	if err != nil {
		result.Error = err
		return result, err
	}

	// 生成目标路径
	destPath, err := o.generateDestPath(opts, info)
	if err != nil {
		result.Error = err
		return result, err
	}

	result.DestPath = destPath

	// 创建目标目录
	if opts.CreateDirectory {
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			result.Error = fmt.Errorf("failed to create directory: %w", err)
			return result, result.Error
		}
	}

	// 复制或链接文件
	if opts.UseHardLink {
		if err := os.Link(opts.SourcePath, destPath); err != nil {
			result.Error = fmt.Errorf("failed to create hard link: %w", err)
			return result, result.Error
		}
	} else {
		if err := o.copyFile(opts.SourcePath, destPath); err != nil {
			result.Error = fmt.Errorf("failed to copy file: %w", err)
			return result, result.Error
		}
	}

	result.Success = true

	if o.logger != nil {
		o.logger.Info("media organized",
			zap.String("source", opts.SourcePath),
			zap.String("dest", destPath),
			zap.Bool("hardlink", opts.UseHardLink))
	}

	return result, nil
}

// generateDestPath 生成目标路径
func (o *Organizer) generateDestPath(opts OrganizeOptions, info *MediaInfo) (string, error) {
	var filename string

	if opts.RenameFile && opts.Template != "" {
		// 使用模板生成文件名
		filename = o.applyTemplate(opts.Template, info)
	} else {
		// 使用原始文件名
		filename = filepath.Base(opts.SourcePath)
	}

	// 构建完整路径
	var destPath string

	if info.Type == "tv" {
		// 电视剧: /dest/剧名/Season 01/剧名.S01E05.mkv
		seasonDir := fmt.Sprintf("Season %02d", info.Season)
		destPath = filepath.Join(opts.DestPath, info.Title, seasonDir, filename)
	} else {
		// 电影: /dest/电影名 (年份)/电影名.mkv
		movieDir := info.Title
		if info.Year > 0 {
			movieDir = fmt.Sprintf("%s (%d)", info.Title, info.Year)
		}
		destPath = filepath.Join(opts.DestPath, movieDir, filename)
	}

	return destPath, nil
}

// applyTemplate 应用命名模板
func (o *Organizer) applyTemplate(template string, info *MediaInfo) string {
	result := template

	// 替换变量
	replacements := map[string]string{
		"{title}":   info.Title,
		"{year}":    fmt.Sprintf("%d", info.Year),
		"{season}":  fmt.Sprintf("%02d", info.Season),
		"{episode}": fmt.Sprintf("%02d", info.Episode),
		"{quality}": info.Quality,
		"{source}":  info.Source,
		"{codec}":   info.Codec,
		"{audio}":   info.Audio,
		"{group}":   info.Group,
	}

	for key, value := range replacements {
		result = strings.ReplaceAll(result, key, value)
	}

	// 添加原始扩展名
	ext := filepath.Ext(info.OriginalName)
	if !strings.HasSuffix(result, ext) {
		result += ext
	}

	return result
}

// copyFile 复制文件
func (o *Organizer) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}

	// 复制文件权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
