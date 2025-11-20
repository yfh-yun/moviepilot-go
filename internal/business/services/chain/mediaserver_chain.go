// Package chain 媒体服务器处理链
package chain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/jellyfin"
	"github.com/yfh-yun/moviepilot-go/internal/integration/emby"
	"github.com/yfh-yun/moviepilot-go/internal/integration/plex"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"

	"go.uber.org/zap"
)

// MediaServerChain 媒体服务器处理链
// 负责与各种媒体服务器的集成和同步
type MediaServerChain struct {
	*ChainBase
	
	// 媒体服务器客户端
	jellyfinClient *jellyfin.Client
	embyClient     *emby.Client
	plexClient     *plex.Client
	
	// 同步管理器
	syncManager    *SyncManager
	
	// 配置
	config         *MediaServerConfig
	
	// 缓存
	serverCache    map[string]*ServerInfo
	cacheMutex     sync.RWMutex
	
	logger         *zap.Logger
}

// MediaServerConfig 媒体服务器配置
type MediaServerConfig struct {
	Jellyfin *JellyfinConfig `json:"jellyfin"`
	Emby     *EmbyConfig     `json:"emby"`
	Plex     *PlexConfig     `json:"p"`
	
	// 同步配置
	SyncInterval time.Duration `json:"sync_interval"`
	AutoSync     bool           `json:"auto_sync"`
	
	// 性能配置
	MaxConcurrent int `json:"max_concurrent"`
	Timeout      time.Duration `json:"timeout"`
}

// JellyfinConfig Jellyfin配置
type JellyfinConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"api_key"`
	Enabled  bool   `json:"enabled"`
	Insecure bool   `json:"insecure"`
}

// EmbyConfig Emby配置
type EmbyConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"api_key"`
	Enabled  bool   `json:"enabled"`
	Insecure bool   `json:"insecure"`
}

// PlexConfig Plex配置
type PlexConfig struct {
	URL      string `json:"url"`
	Token    string `json:"token"`
	Enabled  bool   `json:"enabled"`
	Insecure bool   `json:"insecure"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Type     string    `json:"type"` // jellyfin, emby, plex
	URL      string    `json:"url"`
	Status   string    `json:"status"` // online, offline, error
	LastSync time.Time `json:"last_sync"`
	
	// 统计信息
	Libraries   int `json:"libraries"`
	Items       int `json:"items"`
	Users       int `json:"users"`
}

// Library 媒体库
type Library struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // movies, tvshows, music
	Path        string    `json:"path"`
	ItemCount   int       `json:"item_count"`
	LastScanned time.Time `json:"last_scanned"`
	RefreshedAt time.Time `json:"refreshed_at"`
}

// MediaItem 媒体项
type MediaItem struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Type         string                 `json:"type"` // movie, episode, music
	LibraryID    string                 `json:"library_id"`
	Path         string                 `json:"path"`
	Size         int64                  `json:"size"`
	Duration     int64                  `json:"duration"` // seconds
	Quality      string                 `json:"quality"`
	Year         int                    `json:"year"`
	Genres       []string               `json:"genres"`
	Rating       float64                `json:"rating"`
	Played       bool                   `json:"played"`
	PlayCount    int                    `json:"play_count"`
	LastPlayed   *time.Time             `json:"last_played"`
	AddedAt      time.Time              `json:"added_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// PlayState 播放状态
type PlayState struct {
	ItemID      string    `json:"item_id"`
	UserID      string    `json:"user_id"`
	Position    int64     `json:"position"`    // milliseconds
	Duration    int64     `json:"duration"`    // milliseconds
	IsPlaying   bool      `json:"is_playing"`
	LastUpdated time.Time `json:"last_updated"`
}

// SyncManager 同步管理器
type SyncManager struct {
	syncQueue    chan *SyncTask
	activeSyncs  map[string]*SyncTask
	syncMutex    sync.RWMutex
	
	// 配置
	maxConcurrent int
	timeout      time.Duration
	
	logger       *zap.Logger
}

// SyncTask 同步任务
type SyncTask struct {
	ID        string              `json:"id"`
	ServerID  string              `json:"server_id"`
	Type      string              `json:"type"` // full, library, item
	TargetID  string              `json:"target_id"` // library_id or item_id
	Status    string              `json:"status"` // pending, running, completed, failed
	CreatedAt time.Time           `json:"created_at"`
	StartedAt *time.Time          `json:"started_at"`
	EndedAt   *time.Time          `json:"ended_at"`
	Error     string              `json:"error"`
	Progress  *SyncProgress       `json:"progress"`
}

// SyncProgress 同步进度
type SyncProgress struct {
	Total       int     `json:"total"`
	Processed   int     `json:"processed"`
	Failed      int     `json:"failed"`
	Percentage  float64 `json:"percentage"`
	CurrentItem string  `json:"current_item"`
}

// NewMediaServerChain 创建媒体服务器处理链实例
func NewMediaServerChain(
	base *ChainBase,
	config *MediaServerConfig,
	logger *zap.Logger,
) *MediaServerChain {
	chain := &MediaServerChain{
		ChainBase:    base,
		config:       config,
		serverCache:  make(map[string]*ServerInfo),
		logger:       logger.With(zap.String("module", "chain.mediaserver")),
	}
	
	// 初始化同步管理器
	chain.syncManager = NewSyncManager(
		config.MaxConcurrent,
		config.Timeout,
		chain.logger.With(zap.String("component", "sync_manager")),
	)
	
	// 初始化客户端
	if config.Jellyfin != nil && config.Jellyfin.Enabled {
		chain.jellyfinClient = jellyfin.NewClient(&jellyfin.Config{
			URL:      config.Jellyfin.URL,
			APIKey:   config.Jellyfin.APIKey,
			Insecure: config.Jellyfin.Insecure,
		})
	}
	
	if config.Emby != nil && config.Emby.Enabled {
		chain.embyClient = emby.NewClient(&emby.Config{
			URL:      config.Emby.URL,
			APIKey:   config.Emby.APIKey,
			Insecure: config.Emby.Insecure,
		})
	}
	
	if config.Plex != nil && config.Plex.Enabled {
		chain.plexClient = plex.NewClient(&plex.Config{
			URL:      config.Plex.URL,
			Token:    config.Plex.Token,
			Insecure: config.Plex.Insecure,
		})
	}
	
	return chain
}

// GetServerInfo 获取服务器信息
func (msc *MediaServerChain) GetServerInfo(ctx context.Context, serverType string) (*ServerInfo, error) {
	msc.cacheMutex.RLock()
	if cached, exists := msc.serverCache[serverType]; exists {
		// 检查缓存是否过期
		if time.Since(cached.LastSync) < 5*time.Minute {
			msc.cacheMutex.RUnlock()
			return cached, nil
		}
	}
	msc.cacheMutex.RUnlock()
	
	var serverInfo *ServerInfo
	var err error
	
	switch serverType {
	case "jellyfin":
		serverInfo, err = msc.getJellyfinInfo(ctx)
	case "emby":
		serverInfo, err = msc.getEmbyInfo(ctx)
	case "plex":
		serverInfo, err = msc.getPlexInfo(ctx)
	default:
		return nil, fmt.Errorf("不支持的媒体服务器类型: %s", serverType)
	}
	
	if err != nil {
		return nil, fmt.Errorf("获取%s服务器信息失败: %w", serverType, err)
	}
	
	// 缓存结果
	msc.cacheMutex.Lock()
	msc.serverCache[serverType] = serverInfo
	msc.cacheMutex.Unlock()
	
	return serverInfo, nil
}

// GetLibraries 获取媒体库列表
func (msc *MediaServerChain) GetLibraries(ctx context.Context, serverType string) ([]*Library, error) {
	var libraries []*Library
	var err error
	
	switch serverType {
	case "jellyfin":
		libraries, err = msc.getJellyfinLibraries(ctx)
	case "emby":
		libraries, err = msc.getEmbyLibraries(ctx)
	case "plex":
		libraries, err = msc.getPlexLibraries(ctx)
	default:
		return nil, fmt.Errorf("不支持的媒体服务器类型: %s", serverType)
	}
	
	if err != nil {
		return nil, fmt.Errorf("获取%s媒体库失败: %w", serverType, err)
	}
	
	return libraries, nil
}

// ScanLibrary 扫描媒体库
func (msc *MediaServerChain) ScanLibrary(ctx context.Context, serverType, libraryID string) error {
	task := &SyncTask{
		ID:       fmt.Sprintf("scan_%s_%s_%d", serverType, libraryID, time.Now().Unix()),
		ServerID: serverType,
		Type:     "library",
		TargetID: libraryID,
		Status:   "pending",
		CreatedAt: time.Now(),
	}
	
	return msc.syncManager.EnqueueTask(ctx, task)
}

// GetMediaItems 获取媒体项
func (msc *MediaServerChain) GetMediaItems(ctx context.Context, serverType string, filter *MediaItemFilter) ([]*MediaItem, error) {
	var items []*MediaItem
	var err error
	
	switch serverType {
	case "jellyfin":
		items, err = msc.getJellyfinItems(ctx, filter)
	case "emby":
		items, err = msc.getEmbyItems(ctx, filter)
	case "plex":
		items, err = msc.getPlexItems(ctx, filter)
	default:
		return nil, fmt.Errorf("不支持的媒体服务器类型: %s", serverType)
	}
	
	if err != nil {
		return nil, fmt.Errorf("获取%s媒体项失败: %w", serverType, err)
	}
	
	return items, nil
}

// GetPlayState 获取播放状态
func (msc *MediaServerChain) GetPlayState(ctx context.Context, serverType, itemID, userID string) (*PlayState, error) {
	var playState *PlayState
	var err error
	
	switch serverType {
	case "jellyfin":
		playState, err = msc.getJellyfinPlayState(ctx, itemID, userID)
	case "emby":
		playState, err = msc.getEmbyPlayState(ctx, itemID, userID)
	case "plex":
		playState, err = msc.getPlexPlayState(ctx, itemID, userID)
	default:
		return nil, fmt.Errorf("不支持的媒体服务器类型: %s", serverType)
	}
	
	if err != nil {
		return nil, fmt.Errorf("获取%s播放状态失败: %w", serverType, err)
	}
	
	return playState, nil
}

// SyncAllServers 同步所有服务器
func (msc *MediaServerChain) SyncAllServers(ctx context.Context) error {
	serverTypes := []string{}
	
	if msc.config.Jellyfin != nil && msc.config.Jellyfin.Enabled {
		serverTypes = append(serverTypes, "jellyfin")
	}
	if msc.config.Emby != nil && msc.config.Emby.Enabled {
		serverTypes = append(serverTypes, "emby")
	}
	if msc.config.Plex != nil && msc.config.Plex.Enabled {
		serverTypes = append(serverTypes, "plex")
	}
	
	for _, serverType := range serverTypes {
		task := &SyncTask{
			ID:       fmt.Sprintf("sync_%s_%d", serverType, time.Now().Unix()),
			ServerID: serverType,
			Type:     "full",
			Status:   "pending",
			CreatedAt: time.Now(),
		}
		
		if err := msc.syncManager.EnqueueTask(ctx, task); err != nil {
			msc.logger.Error("入队同步任务失败",
				zap.String("server_type", serverType),
				zap.Error(err))
		}
	}
	
	return nil
}

// getJellyfinInfo 获取Jellyfin服务器信息
func (msc *MediaServerChain) getJellyfinInfo(ctx context.Context) (*ServerInfo, error) {
	if msc.jellyfinClient == nil {
		return nil, fmt.Errorf("Jellyfin客户端未初始化")
	}
	
	info, err := msc.jellyfinClient.GetPublicSystemInfo(ctx)
	if err != nil {
		return nil, err
	}
	
	return &ServerInfo{
		ID:       info.ID,
		Name:     info.ServerName,
		Type:     "jellyfin",
		URL:      msc.config.Jellyfin.URL,
		Status:   "online",
		LastSync: time.Now(),
	}, nil
}

// getEmbyInfo 获取Emby服务器信息
func (msc *MediaServerChain) getEmbyInfo(ctx context.Context) (*ServerInfo, error) {
	if msc.embyClient == nil {
		return nil, fmt.Errorf("Emby客户端未初始化")
	}
	
	info, err := msc.embyClient.GetSystemInfo(ctx)
	if err != nil {
		return nil, err
	}
	
	return &ServerInfo{
		ID:       info.ID,
		Name:     info.ServerName,
		Type:     "emby",
		URL:      msc.config.Emby.URL,
		Status:   "online",
		LastSync: time.Now(),
	}, nil
}

// getPlexInfo 获取Plex服务器信息
func (msc *MediaServerChain) getPlexInfo(ctx context.Context) (*ServerInfo, error) {
	if msc.plexClient == nil {
		return nil, fmt.Errorf("Plex客户端未初始化")
	}
	
	info, err := msc.plexClient.GetServerInfo(ctx)
	if err != nil {
		return nil, err
	}
	
	return &ServerInfo{
		ID:       info.MachineIdentifier,
		Name:     info.FriendlyName,
		Type:     "plex",
		URL:      msc.config.Plex.URL,
		Status:   "online",
		LastSync: time.Now(),
	}, nil
}

// MediaItemFilter 媒体项过滤器
type MediaItemFilter struct {
	LibraryID   string   `json:"library_id"`
	Type        string   `json:"type"`
	Genres      []string `json:"genres"`
	Year        int      `json:"year"`
	MinRating   float64  `json:"min_rating"`
	MaxRating   float64  `json:"max_rating"`
	Played      *bool    `json:"played"`
	Limit       int      `json:"limit"`
	Offset      int      `json:"offset"`
	SortBy      string   `json:"sort_by"`
	SortOrder   string   `json:"sort_order"` // asc, desc
}

// 以下是各平台具体实现方法的占位符，实际需要根据各平台API实现
func (msc *MediaServerChain) getJellyfinLibraries(ctx context.Context) ([]*Library, error) {
	// TODO: 实现Jellyfin库列表获取
	return []*Library{}, nil
}

func (msc *MediaServerChain) getEmbyLibraries(ctx context.Context) ([]*Library, error) {
	// TODO: 实现Emby库列表获取
	return []*Library{}, nil
}

func (msc *MediaServerChain) getPlexLibraries(ctx context.Context) ([]*Library, error) {
	// TODO: 实现Plex库列表获取
	return []*Library{}, nil
}

func (msc *MediaServerChain) getJellyfinItems(ctx context.Context, filter *MediaItemFilter) ([]*MediaItem, error) {
	// TODO: 实现Jellyfin媒体项获取
	return []*MediaItem{}, nil
}

func (msc *MediaServerChain) getEmbyItems(ctx context.Context, filter *MediaItemFilter) ([]*MediaItem, error) {
	// TODO: 实现Emby媒体项获取
	return []*MediaItem{}, nil
}

func (msc *MediaServerChain) getPlexItems(ctx context.Context, filter *MediaItemFilter) ([]*MediaItem, error) {
	// TODO: 实现Plex媒体项获取
	return []*MediaItem{}, nil
}

func (msc *MediaServerChain) getJellyfinPlayState(ctx context.Context, itemID, userID string) (*PlayState, error) {
	// TODO: 实现Jellyfin播放状态获取
	return nil, fmt.Errorf("未实现")
}

func (msc *MediaServerChain) getEmbyPlayState(ctx context.Context, itemID, userID string) (*PlayState, error) {
	// TODO: 实现Emby播放状态获取
	return nil, fmt.Errorf("未实现")
}

func (msc *MediaServerChain) getPlexPlayState(ctx context.Context, itemID, userID string) (*PlayState, error) {
	// TODO: 实现Plex播放状态获取
	return nil, fmt.Errorf("未实现")
}