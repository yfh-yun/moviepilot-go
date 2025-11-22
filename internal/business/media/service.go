package media

import (
	"strings"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/internal/models"
)

// Service 定义媒体相关操作。
type Service interface {
	Identify(files []FileItem, opts IdentifyOptions) ([]models.Media, error)
}

// IdentifyOptions 控制识别行为。
type IdentifyOptions struct {
	ForceRefresh bool
	Source       string
}

// FileItem 是 Action 层传递的轻量文件信息。
type FileItem struct {
	Path string
}

// IdentifyResult 是占位结构，可扩展额外字段。
type IdentifyResult struct {
	Media models.Media
}

// SimpleService 是一个占位实现，使用文件名猜测媒体信息。
type SimpleService struct {
	logger *zap.Logger
}

// NewSimpleService 创建服务实例。
func NewSimpleService(logger *zap.Logger) *SimpleService {
	return &SimpleService{logger: logger}
}

// Identify 生成基础媒体信息。
func (s *SimpleService) Identify(files []FileItem, opts IdentifyOptions) ([]models.Media, error) {
	medias := make([]models.Media, 0, len(files))
	now := time.Now()

	for _, file := range files {
		title := sanitize(file.Path)
		mediaType := guessType(file.Path)
		medias = append(medias, models.Media{
			BaseModel: models.BaseModel{
				CreatedAt: now,
				UpdatedAt: now,
			},
			Title:       title,
			Type:        mediaType,
			Description: "placeholder media info",
		})
	}

	if s.logger != nil {
		s.logger.Debug("media identify placeholder", zap.Int("count", len(medias)), zap.Bool("force", opts.ForceRefresh))
	}

	return medias, nil
}

func sanitize(path string) string {
	base := path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		base = path[idx+1:]
	}
	base = strings.TrimSuffix(base, filepathExt(base))
	base = strings.ReplaceAll(base, ".", " ")
	return strings.TrimSpace(base)
}

func guessType(path string) string {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "s0") && strings.Contains(lower, "e") {
		return "tv"
	}
	return "movie"
}

func filepathExt(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '/' {
			break
		}
	}
	return ""
}
