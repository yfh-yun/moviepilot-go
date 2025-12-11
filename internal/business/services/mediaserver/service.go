package mediaserver

import (
	"context"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/mediaserver"
)

// Service 媒体服务器服务接口
type Service interface {
	// TestConnection 测试连接
	TestConnection(ctx context.Context, config mediaserver.Config) error

	// GetLibraries 获取媒体库列表
	GetLibraries(ctx context.Context, serverID string) ([]mediaserver.Library, error)

	// RefreshLibrary 刷新媒体库
	RefreshLibrary(ctx context.Context, serverID string, libraryID string) error

	// RefreshAll 刷新所有媒体库
	RefreshAll(ctx context.Context) error
}

// service 服务实现
type service struct {
	manager *mediaserver.Manager
	logger  *zap.Logger
}

// NewService 创建服务
func NewService() Service {
	return &service{
		manager: mediaserver.NewManager(),
		logger:  logger.GetLogger(),
	}
}

// TestConnection 测试连接
func (s *service) TestConnection(ctx context.Context, config mediaserver.Config) error {
	s.logger.Info("测试媒体服务器连接", zap.String("type", config.Type))

	server, err := mediaserver.CreateServer(config)
	if err != nil {
		return err
	}

	return server.Test(ctx)
}

// GetLibraries 获取媒体库列表
func (s *service) GetLibraries(ctx context.Context, serverID string) ([]mediaserver.Library, error) {
	s.logger.Info("获取媒体库列表", zap.String("serverID", serverID))

	server, err := s.manager.Get(serverID)
	if err != nil {
		return nil, err
	}

	return server.GetLibraries(ctx)
}

// RefreshLibrary 刷新媒体库
func (s *service) RefreshLibrary(ctx context.Context, serverID string, libraryID string) error {
	s.logger.Info("刷新媒体库",
		zap.String("serverID", serverID),
		zap.String("libraryID", libraryID))

	server, err := s.manager.Get(serverID)
	if err != nil {
		return err
	}

	return server.RefreshLibrary(ctx, libraryID)
}

// RefreshAll 刷新所有媒体库
func (s *service) RefreshAll(ctx context.Context) error {
	s.logger.Info("刷新所有媒体库")
	return s.manager.RefreshAll(ctx)
}
