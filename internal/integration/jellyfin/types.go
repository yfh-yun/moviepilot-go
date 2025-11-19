package jellyfin

import (
	"time"
)

// MediaServerInfo 表示Jellyfin服务器基本信息
type MediaServerInfo struct {
	ID           string `json:"Id"`
	Name         string `json:"ServerName"`
	Version      string `json:"Version"`
	LocalAddress string `json:"LocalAddress"`
	WanAddress   string `json:"WanAddress"`
}

// UserInfo 表示Jellyfin用户信息
type UserInfo struct {
	ID        string     `json:"Id"`
	Name      string     `json:"Name"`
	Policy    UserPolicy `json:"Policy"`
	LastLogin time.Time  `json:"LastLoginDate"`
}

// UserPolicy 表示用户权限策略
type UserPolicy struct {
	IsAdministrator bool `json:"IsAdministrator"`
	IsHidden        bool `json:"IsHidden"`
	IsDisabled      bool `json:"IsDisabled"`
}

// LibraryInfo 表示媒体库信息
type LibraryInfo struct {
	ID        string `json:"Id"`
	Name      string `json:"Name"`
	Type      string `json:"CollectionType"`
	Path      string `json:"Path"`
	ItemCount int    `json:"ItemCount"`
}

// MediaItem 表示媒体项信息
type MediaItem struct {
	ID           string      `json:"Id"`
	Name         string      `json:"Name"`
	Type         string      `json:"Type"`
	MediaType    string      `json:"MediaType"`
	Path         string      `json:"Path"`
	Size         int64       `json:"Size"`
	PremiereDate time.Time   `json:"PremiereDate"`
	DateCreated  time.Time   `json:"DateCreated"`
	ProviderIDs  ProviderIDs `json:"ProviderIds"`
}

// ProviderIDs 表示媒体提供商ID
type ProviderIDs struct {
	Tmdb string `json:"Tmdb"`
	Imdb string `json:"Imdb"`
	Tvdb string `json:"Tvdb"`
}

// PlaybackInfo 表示播放信息
type PlaybackInfo struct {
	ID           string    `json:"Id"`
	ItemID       string    `json:"ItemId"`
	SessionID    string    `json:"SessionId"`
	UserName     string    `json:"UserName"`
	Client       string    `json:"Client"`
	Position     int64     `json:"PositionTicks"`
	Duration     int64     `json:"PlaybackDurationTicks"`
	IsPaused     bool      `json:"IsPaused"`
	LastActivity time.Time `json:"LastActivityDate"`
}

// ClientConfig 表示Jellyfin客户端配置
type ClientConfig struct {
	URL        string        `json:"url"`
	APIKey     string        `json:"api_key"`
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	RetryDelay time.Duration `json:"retry_delay"`
	Enabled    bool          `json:"enabled"`
}
