package mediaserver

import (
	"time"

	"github.com/pkg/errors"
)

// MediaServerType 表示媒体服务器类型
type MediaServerType string

const (
	MediaServerTypeEmby     MediaServerType = "emby"
	MediaServerTypeJellyfin MediaServerType = "jellyfin"
	MediaServerTypePlex     MediaServerType = "plex"
)

// MediaServer 表示媒体服务器通用接口
type MediaServer interface {
	// GetType 获取服务器类型
	GetType() MediaServerType

	// GetName 获取服务器名称
	GetName() string

	// HealthCheck 健康检查
	HealthCheck() error

	// GetServerInfo 获取服务器信息
	GetServerInfo() (*ServerInfo, error)

	// GetUsers 获取用户列表
	GetUsers() ([]User, error)

	// GetLibraries 获取媒体库列表
	GetLibraries() ([]Library, error)

	// GetLibraryItems 获取媒体库项目
	GetLibraryItems(libraryID string, params map[string]string) ([]MediaItem, error)

	// RefreshLibrary 刷新媒体库
	RefreshLibrary(libraryID string) error

	// GetPlaybackSessions 获取播放会话
	GetPlaybackSessions() ([]PlaybackSession, error)
}

// ServerInfo 表示服务器信息
type ServerInfo struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Type         MediaServerType `json:"type"`
	LocalAddress string          `json:"local_address"`
	WanAddress   string          `json:"wan_address"`
	LastSeen     time.Time       `json:"last_seen"`
}

// User 表示用户信息
type User struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	IsAdmin    bool      `json:"is_admin"`
	IsDisabled bool      `json:"is_disabled"`
	LastLogin  time.Time `json:"last_login"`
}

// Library 表示媒体库信息
type Library struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Path        string    `json:"path"`
	ItemCount   int       `json:"item_count"`
	LastRefresh time.Time `json:"last_refresh"`
}

// MediaItem 表示媒体项信息
type MediaItem struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	MediaType   string      `json:"media_type"`
	Path        string      `json:"path"`
	Size        int64       `json:"size"`
	ReleaseDate time.Time   `json:"release_date"`
	AddedAt     time.Time   `json:"added_at"`
	ProviderIDs ProviderIDs `json:"provider_ids"`
}

// ProviderIDs 表示提供商ID
type ProviderIDs struct {
	Tmdb string `json:"tmdb"`
	Imdb string `json:"imdb"`
	Tvdb string `json:"tvdb"`
}

// PlaybackSession 表示播放会话
type PlaybackSession struct {
	ID           string    `json:"id"`
	ItemID       string    `json:"item_id"`
	Title        string    `json:"title"`
	UserName     string    `json:"user_name"`
	Client       string    `json:"client"`
	Position     int64     `json:"position"` // 播放位置(秒)
	Duration     int64     `json:"duration"` // 总时长(秒)
	IsPaused     bool      `json:"is_paused"`
	LastActivity time.Time `json:"last_activity"`
}

// RefreshRequest 表示刷新请求
type RefreshRequest struct {
	LibraryID    string `json:"library_id"`
	ForceFull    bool   `json:"force_full"`
	UpdateImages bool   `json:"update_images"`
}

// Error 定义媒体服务器错误类型
var (
	ErrServerNotFound     = errors.New("媒体服务器未找到")
	ErrServerNotConnected = errors.New("媒体服务器未连接")
	ErrInvalidLibraryID   = errors.New("无效的媒体库ID")
	ErrRefreshInProgress  = errors.New("媒体库刷新中")
	ErrOperationFailed    = errors.New("操作失败")
)

// Config 表示媒体服务器配置
type Config struct {
	Emby     EmbyConfig     `json:"emby"`
	Jellyfin JellyfinConfig `json:"jellyfin"`
	Plex     PlexConfig     `json:"plex"`
}

// EmbyConfig 表示Emby配置
type EmbyConfig struct {
	Enabled    bool          `json:"enabled"`
	URL        string        `json:"url"`
	APIKey     string        `json:"api_key"`
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	RetryDelay time.Duration `json:"retry_delay"`
}

// JellyfinConfig 表示Jellyfin配置
type JellyfinConfig struct {
	Enabled    bool          `json:"enabled"`
	URL        string        `json:"url"`
	APIKey     string        `json:"api_key"`
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	RetryDelay time.Duration `json:"retry_delay"`
}

// PlexConfig 表示Plex配置
type PlexConfig struct {
	Enabled    bool          `json:"enabled"`
	URL        string        `json:"url"`
	Token      string        `json:"token"`
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	RetryDelay time.Duration `json:"retry_delay"`
}
