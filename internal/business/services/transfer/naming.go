package transfer

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/pkg/utils/naming"
)

// NamingStrategy 命名策略接口
type NamingStrategy interface {
	// GeneratePath 生成目标路径
	GeneratePath(media *database.Media, sourcePath string, metadata FileMetadata) (string, error)
}

// FileMetadata 文件元数据(从文件名解析得到)
type FileMetadata struct {
	Resolution string
	Source     string
	Codec      string
	Audio      string
	Subtitle   string
	Group      string
}

// TemplateNamingStrategy 基于模板的命名策略
type TemplateNamingStrategy struct {
	template *naming.Template
	logger   *zap.Logger
}

// NewTemplateNamingStrategy 创建模板命名策略
func NewTemplateNamingStrategy(templateStr string, logger *zap.Logger) (*TemplateNamingStrategy, error) {
	tmpl, err := naming.ParseTemplate(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &TemplateNamingStrategy{
		template: tmpl,
		logger:   logger,
	}, nil
}

// GeneratePath 生成目标路径
func (s *TemplateNamingStrategy) GeneratePath(media *database.Media, sourcePath string, metadata FileMetadata) (string, error) {
	// 从 Media 模型获取基础变量
	vars := naming.MediaToVars(media, sourcePath)

	// 添加文件元数据
	vars["Resolution"] = metadata.Resolution
	vars["Source"] = metadata.Source
	vars["Codec"] = metadata.Codec
	vars["Audio"] = metadata.Audio
	vars["Subtitle"] = metadata.Subtitle
	vars["Group"] = metadata.Group

	// 渲染模板
	path, err := s.template.Render(vars)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	if s.logger != nil {
		s.logger.Debug("generated path from template",
			zap.String("source", sourcePath),
			zap.String("target", path),
			zap.String("template", s.template.Raw()))
	}

	return path, nil
}

// MovieNamingStrategy 电影命名策略
type MovieNamingStrategy struct {
	*TemplateNamingStrategy
}

// NewMovieNamingStrategy 创建电影命名策略
func NewMovieNamingStrategy(templateStr string, logger *zap.Logger) (*MovieNamingStrategy, error) {
	if templateStr == "" {
		templateStr = naming.GetDefaultTemplate("movie")
	}

	base, err := NewTemplateNamingStrategy(templateStr, logger)
	if err != nil {
		return nil, err
	}

	return &MovieNamingStrategy{
		TemplateNamingStrategy: base,
	}, nil
}

// TVNamingStrategy 电视剧命名策略
type TVNamingStrategy struct {
	*TemplateNamingStrategy
}

// NewTVNamingStrategy 创建电视剧命名策略
func NewTVNamingStrategy(templateStr string, logger *zap.Logger) (*TVNamingStrategy, error) {
	if templateStr == "" {
		templateStr = naming.GetDefaultTemplate("tv")
	}

	base, err := NewTemplateNamingStrategy(templateStr, logger)
	if err != nil {
		return nil, err
	}

	return &TVNamingStrategy{
		TemplateNamingStrategy: base,
	}, nil
}

// AnimeNamingStrategy 动漫命名策略
type AnimeNamingStrategy struct {
	*TemplateNamingStrategy
}

// NewAnimeNamingStrategy 创建动漫命名策略
func NewAnimeNamingStrategy(templateStr string, logger *zap.Logger) (*AnimeNamingStrategy, error) {
	if templateStr == "" {
		templateStr = naming.GetDefaultTemplate("anime")
	}

	base, err := NewTemplateNamingStrategy(templateStr, logger)
	if err != nil {
		return nil, err
	}

	return &AnimeNamingStrategy{
		TemplateNamingStrategy: base,
	}, nil
}

// NamingManager 命名管理器
type NamingManager struct {
	strategies map[string]NamingStrategy
	logger     *zap.Logger
}

// NamingConfig 命名配置
type NamingConfig struct {
	MovieTemplate string
	TVTemplate    string
	AnimeTemplate string
}

// NewNamingManager 创建命名管理器
func NewNamingManager(config NamingConfig, logger *zap.Logger) (*NamingManager, error) {
	strategies := make(map[string]NamingStrategy)

	// 创建电影策略
	movieStrategy, err := NewMovieNamingStrategy(config.MovieTemplate, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create movie strategy: %w", err)
	}
	strategies["movie"] = movieStrategy

	// 创建电视剧策略
	tvStrategy, err := NewTVNamingStrategy(config.TVTemplate, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create tv strategy: %w", err)
	}
	strategies["tv"] = tvStrategy

	// 创建动漫策略
	animeStrategy, err := NewAnimeNamingStrategy(config.AnimeTemplate, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create anime strategy: %w", err)
	}
	strategies["anime"] = animeStrategy

	return &NamingManager{
		strategies: strategies,
		logger:     logger,
	}, nil
}

// GeneratePath 根据媒体类型生成路径
func (m *NamingManager) GeneratePath(media *database.Media, sourcePath string, metadata FileMetadata) (string, error) {
	mediaType := media.Type
	if mediaType == "" {
		mediaType = "movie"
	}

	strategy, ok := m.strategies[mediaType]
	if !ok {
		// 回退到电影策略
		strategy = m.strategies["movie"]
		if m.logger != nil {
			m.logger.Warn("unknown media type, using movie strategy",
				zap.String("type", mediaType))
		}
	}

	return strategy.GeneratePath(media, sourcePath, metadata)
}

// SetStrategy 设置特定类型的策略
func (m *NamingManager) SetStrategy(mediaType string, strategy NamingStrategy) {
	m.strategies[mediaType] = strategy
}

// GetStrategy 获取特定类型的策略
func (m *NamingManager) GetStrategy(mediaType string) (NamingStrategy, bool) {
	strategy, ok := m.strategies[mediaType]
	return strategy, ok
}

// ParseFileMetadata 从文件路径解析元数据
// 这个函数应该与 media.ParseFileName 配合使用
func ParseFileMetadata(filePath string) FileMetadata {
	// 简化实现,实际应该调用 media.ParseFileName
	// 这里只提取基本信息

	filename := filepath.Base(filePath)

	metadata := FileMetadata{}

	// 简单的正则匹配(实际应该使用更完善的解析器)
	if contains(filename, "1080p") {
		metadata.Resolution = "1080p"
	} else if contains(filename, "2160p") || contains(filename, "4K") {
		metadata.Resolution = "2160p"
	} else if contains(filename, "720p") {
		metadata.Resolution = "720p"
	}

	if contains(filename, "BluRay") || contains(filename, "Blu-ray") {
		metadata.Source = "BluRay"
	} else if contains(filename, "WEB-DL") || contains(filename, "WEBDL") {
		metadata.Source = "WEB-DL"
	} else if contains(filename, "WEBRip") {
		metadata.Source = "WEBRip"
	}

	if contains(filename, "x264") || contains(filename, "H.264") {
		metadata.Codec = "x264"
	} else if contains(filename, "x265") || contains(filename, "H.265") || contains(filename, "HEVC") {
		metadata.Codec = "x265"
	}

	return metadata
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
