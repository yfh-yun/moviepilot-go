package mediaserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/integration/emby"
	"github.com/yfh-yun/moviepilot-go/internal/integration/jellyfin"
	"github.com/yfh-yun/moviepilot-go/internal/integration/plex"
	"github.com/yfh-yun/moviepilot-go/internal/repository"
)

// Service 表示媒体服务器服务
type Service struct {
	logger       *zap.Logger
	repo         repository.MediaServerRepository
	servers      map[MediaServerType]MediaServer
	serversMutex sync.RWMutex
	running      bool
	stopChan     chan struct{}
}

// NewService 创建新的媒体服务器服务
func NewService(
	logger *zap.Logger,
	repo repository.MediaServerRepository,
	config *Config,
) (*Service, error) {
	service := &Service{
		logger:   logger,
		repo:     repo,
		servers:  make(map[MediaServerType]MediaServer),
		stopChan: make(chan struct{}),
	}

	// 初始化媒体服务器
	if err := service.initializeServers(config); err != nil {
		return nil, errors.Wrap(err, "初始化媒体服务器失败")
	}

	// 启动健康检查
	go service.healthCheckLoop()

	service.running = true
	service.logger.Info("媒体服务器服务初始化成功")

	return service, nil
}

// initializeServers 初始化所有媒体服务器
func (s *Service) initializeServers(config *Config) error {
	s.serversMutex.Lock()
	defer s.serversMutex.Unlock()

	// 初始化Emby
	if config.Emby.Enabled && config.Emby.URL != "" && config.Emby.APIKey != "" {
		embyConfig := &emby.ClientConfig{
			URL:        config.Emby.URL,
			APIKey:     config.Emby.APIKey,
			Timeout:    config.Emby.Timeout,
			RetryCount: config.Emby.RetryCount,
			RetryDelay: config.Emby.RetryDelay,
			Enabled:    true,
		}

		embyClient := emby.NewClient(embyConfig, s.logger)
		s.servers[MediaServerTypeEmby] = &embyServerWrapper{
			client: embyClient,
			config: embyConfig,
			logger: s.logger,
		}
		s.logger.Info("Emby客户端初始化成功", zap.String("url", config.Emby.URL))
	}

	// 初始化Jellyfin
	if config.Jellyfin.Enabled && config.Jellyfin.URL != "" && config.Jellyfin.APIKey != "" {
		jellyfinConfig := &jellyfin.ClientConfig{
			URL:        config.Jellyfin.URL,
			APIKey:     config.Jellyfin.APIKey,
			Timeout:    config.Jellyfin.Timeout,
			RetryCount: config.Jellyfin.RetryCount,
			RetryDelay: config.Jellyfin.RetryDelay,
			Enabled:    true,
		}

		jellyfinClient := jellyfin.NewClient(jellyfinConfig, s.logger)
		s.servers[MediaServerTypeJellyfin] = &jellyfinServerWrapper{
			client: jellyfinClient,
			config: jellyfinConfig,
			logger: s.logger,
		}
		s.logger.Info("Jellyfin客户端初始化成功", zap.String("url", config.Jellyfin.URL))
	}

	// 初始化Plex
	if config.Plex.Enabled && config.Plex.URL != "" && config.Plex.Token != "" {
		plexConfig := &plex.ClientConfig{
			URL:        config.Plex.URL,
			Token:      config.Plex.Token,
			Timeout:    config.Plex.Timeout,
			RetryCount: config.Plex.RetryCount,
			RetryDelay: config.Plex.RetryDelay,
			Enabled:    true,
		}

		plexClient := plex.NewClient(plexConfig, s.logger)
		s.servers[MediaServerTypePlex] = &plexServerWrapper{
			client: plexClient,
			config: plexConfig,
			logger: s.logger,
		}
		s.logger.Info("Plex客户端初始化成功", zap.String("url", config.Plex.URL))
	}

	if len(s.servers) == 0 {
		return errors.New("没有可用的媒体服务器配置")
	}

	return nil
}

// GetServer 获取指定类型的媒体服务器
func (s *Service) GetServer(serverType MediaServerType) (MediaServer, error) {
	s.serversMutex.RLock()
	defer s.serversMutex.RUnlock()

	server, exists := s.servers[serverType]
	if !exists {
		return nil, errors.Wrapf(ErrServerNotFound, "服务器类型: %s", serverType)
	}

	return server, nil
}

// ListServers 获取所有可用的媒体服务器
func (s *Service) ListServers() []MediaServerType {
	s.serversMutex.RLock()
	defer s.serversMutex.RUnlock()

	serverTypes := make([]MediaServerType, 0, len(s.servers))
	for serverType := range s.servers {
		serverTypes = append(serverTypes, serverType)
	}

	return serverTypes
}

// HealthCheckAll 检查所有媒体服务器的健康状态
func (s *Service) HealthCheckAll() map[MediaServerType]error {
	s.serversMutex.RLock()
	defer s.serversMutex.RUnlock()

	results := make(map[MediaServerType]error)

	for serverType, server := range s.servers {
		if err := server.HealthCheck(); err != nil {
			results[serverType] = err
		} else {
			results[serverType] = nil
		}
	}

	return results
}

// RefreshLibrary 刷新指定媒体库
func (s *Service) RefreshLibrary(ctx context.Context, serverType MediaServerType, libraryID string) error {
	server, err := s.GetServer(serverType)
	if err != nil {
		return errors.Wrap(err, "获取媒体服务器失败")
	}

	if err := server.RefreshLibrary(libraryID); err != nil {
		return errors.Wrap(err, "刷新媒体库失败")
	}

	s.logger.Info("媒体库刷新成功",
		zap.String("server_type", string(serverType)),
		zap.String("library_id", libraryID))

	return nil
}

// GetPlaybackSessions 获取播放会话
func (s *Service) GetPlaybackSessions(ctx context.Context, serverType MediaServerType) ([]PlaybackSession, error) {
	server, err := s.GetServer(serverType)
	if err != nil {
		return nil, errors.Wrap(err, "获取媒体服务器失败")
	}

	sessions, err := server.GetPlaybackSessions()
	if err != nil {
		return nil, errors.Wrap(err, "获取播放会话失败")
	}

	return sessions, nil
}

// SyncMediaLibraries 同步媒体库
func (s *Service) SyncMediaLibraries(ctx context.Context, serverType MediaServerType) error {
	server, err := s.GetServer(serverType)
	if err != nil {
		return errors.Wrap(err, "获取媒体服务器失败")
	}

	libraries, err := server.GetLibraries()
	if err != nil {
		return errors.Wrap(err, "获取媒体库列表失败")
	}

	// 同步到数据库
	for _, library := range libraries {
		err := s.repo.SaveMediaLibrary(ctx, library)
		if err != nil {
			s.logger.Error("保存媒体库失败",
				zap.String("server_type", string(serverType)),
				zap.String("library_id", library.ID),
				zap.Error(err))
		}
	}

	s.logger.Info("媒体库同步完成",
		zap.String("server_type", string(serverType)),
		zap.Int("library_count", len(libraries)))

	return nil
}

// healthCheckLoop 健康检查循环
func (s *Service) healthCheckLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performHealthCheck()
		case <-s.stopChan:
			s.logger.Info("健康检查循环停止")
			return
		}
	}
}

// performHealthCheck 执行健康检查
func (s *Service) performHealthCheck() {
	s.serversMutex.RLock()
	defer s.serversMutex.RUnlock()

	for serverType, server := range s.servers {
		if err := server.HealthCheck(); err != nil {
			s.logger.Warn("媒体服务器健康检查失败",
				zap.String("server_type", string(serverType)),
				zap.Error(err))
		} else {
			s.logger.Debug("媒体服务器健康检查成功",
				zap.String("server_type", string(serverType)))
		}
	}
}

// Stop 停止服务
func (s *Service) Stop() {
	if s.running {
		s.running = false
		close(s.stopChan)
		s.logger.Info("媒体服务器服务已停止")
	}
}

// 服务器包装器实现

type embyServerWrapper struct {
	client *emby.Client
	config *emby.ClientConfig
	logger *zap.Logger
}

func (e *embyServerWrapper) GetType() MediaServerType {
	return MediaServerTypeEmby
}

func (e *embyServerWrapper) GetName() string {
	return "Emby"
}

func (e *embyServerWrapper) HealthCheck() error {
	return e.client.HealthCheck(context.Background())
}

func (e *embyServerWrapper) GetServerInfo() (*ServerInfo, error) {
	info, err := e.client.GetServerInfo(context.Background())
	if err != nil {
		return nil, err
	}

	return &ServerInfo{
		ID:           info.ID,
		Name:         info.Name,
		Version:      info.Version,
		Type:         MediaServerTypeEmby,
		LocalAddress: info.LocalAddress,
		WanAddress:   info.WanAddress,
		LastSeen:     time.Now(),
	}, nil
}

func (e *embyServerWrapper) GetUsers() ([]User, error) {
	embyUsers, err := e.client.GetUsers(context.Background())
	if err != nil {
		return nil, err
	}

	users := make([]User, len(embyUsers))
	for i, embyUser := range embyUsers {
		users[i] = User{
			ID:         embyUser.ID,
			Name:       embyUser.Name,
			IsAdmin:    embyUser.Policy.IsAdministrator,
			IsDisabled: embyUser.Policy.IsDisabled,
			LastLogin:  embyUser.LastLogin,
		}
	}

	return users, nil
}

func (e *embyServerWrapper) GetLibraries() ([]Library, error) {
	embyLibraries, err := e.client.GetLibraries(context.Background())
	if err != nil {
		return nil, err
	}

	libraries := make([]Library, len(embyLibraries))
	for i, embyLibrary := range embyLibraries {
		libraries[i] = Library{
			ID:   embyLibrary.ID,
			Name: embyLibrary.Name,
			Type: embyLibrary.Type,
			Path: embyLibrary.Path,
		}
	}

	return libraries, nil
}

func (e *embyServerWrapper) GetLibraryItems(libraryID string, params map[string]string) ([]MediaItem, error) {
	embyItems, err := e.client.GetLibraryItems(context.Background(), libraryID, params)
	if err != nil {
		return nil, err
	}

	items := make([]MediaItem, len(embyItems))
	for i, embyItem := range embyItems {
		items[i] = MediaItem{
			ID:          embyItem.ID,
			Title:       embyItem.Name,
			Type:        embyItem.Type,
			MediaType:   embyItem.MediaType,
			Path:        embyItem.Path,
			Size:        embyItem.Size,
			ReleaseDate: embyItem.PremiereDate,
			AddedAt:     embyItem.DateCreated,
			ProviderIDs: ProviderIDs{
				Tmdb: embyItem.ProviderIDs.Tmdb,
				Imdb: embyItem.ProviderIDs.Imdb,
				Tvdb: embyItem.ProviderIDs.Tvdb,
			},
		}
	}

	return items, nil
}

func (e *embyServerWrapper) RefreshLibrary(libraryID string) error {
	return e.client.RefreshLibrary(context.Background(), libraryID, nil)
}

func (e *embyServerWrapper) GetPlaybackSessions() ([]PlaybackSession, error) {
	embySessions, err := e.client.GetPlaybackSessions(context.Background())
	if err != nil {
		return nil, err
	}

	sessions := make([]PlaybackSession, len(embySessions))
	for i, embySession := range embySessions {
		sessions[i] = PlaybackSession{
			ID:           embySession.ID,
			ItemID:       embySession.ItemID,
			Title:        embySession.UserName,
			UserName:     embySession.UserName,
			Client:       embySession.Client,
			Position:     embySession.Position / 10000000, // 转换为秒
			Duration:     embySession.Duration / 10000000, // 转换为秒
			IsPaused:     embySession.IsPaused,
			LastActivity: embySession.LastActivity,
		}
	}

	return sessions, nil
}

// jellyfinServerWrapper 和 plexServerWrapper 实现类似，这里省略重复代码

// 为简洁起见，这里省略jellyfinServerWrapper和plexServerWrapper的完整实现
// 它们遵循与embyServerWrapper相同的模式
