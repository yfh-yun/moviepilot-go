package media

import (
	"fmt"
	"strings"

	"go.uber.org/zap"

	"moviepilot-go/internal/models/database"
)

// CompositeService 按需路由到不同的媒体识别实现（Simple/TMDB/站点等）。
type CompositeService struct {
	logger   *zap.Logger
	services map[string]Service
	fallback Service
}

// NewCompositeService 创建组合识别服务。
// services 使用 source -> Service 的映射；fallback 为默认降级策略。
func NewCompositeService(logger *zap.Logger, services map[string]Service, fallback Service) *CompositeService {
	// 复制一份 map，避免外部修改
	copied := make(map[string]Service)
	for k, v := range services {
		if v != nil {
			copied[strings.ToLower(k)] = v
		}
	}

	// 若未显式指定 fallback，尝试使用 simple 实现
	if fallback == nil {
		if simple, ok := copied["simple"]; ok {
			fallback = simple
		}
	}

	return &CompositeService{
		logger:   logger,
		services: copied,
		fallback: fallback,
	}
}

// NewDefaultCompositeService 使用 SimpleService 作为默认实现，并预留扩展空间。
func NewDefaultCompositeService(logger *zap.Logger) *CompositeService {
	services := map[string]Service{
		"simple": NewSimpleService(logger),
	}
	return NewCompositeService(logger, services, services["simple"])
}

// Register 动态注册新的识别实现。
func (c *CompositeService) Register(source string, service Service) {
	if c.services == nil {
		c.services = make(map[string]Service)
	}
	key := strings.ToLower(strings.TrimSpace(source))
	if key == "" || service == nil {
		return
	}
	c.services[key] = service
}

func (c *CompositeService) Identify(files []FileItem, opts IdentifyOptions) ([]database.Media, error) {
	source := strings.ToLower(strings.TrimSpace(opts.Source))

	// 首先尝试显式指定的 source
	if source != "" {
		if service, ok := c.services[source]; ok {
			return service.Identify(files, opts)
		}
		if c.logger != nil {
			c.logger.Warn("media identify source not found, fallback to default", zap.String("source", source))
		}
	}

	// 未指定 source 或未命中映射时，使用 fallback
	if c.fallback != nil {
		return c.fallback.Identify(files, opts)
	}

	// 若没有 fallback，则尝试服务列表中的第一个实现
	for name, service := range c.services {
		if service != nil {
			if c.logger != nil {
				c.logger.Info("using first available media identify service", zap.String("source", name))
			}
			return service.Identify(files, opts)
		}
	}

	return nil, fmt.Errorf("no media identify service available")
}
