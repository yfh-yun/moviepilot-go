package plex

import (
	"time"
)

// MediaServerInfo 表示Plex服务器基本信息
type MediaServerInfo struct {
	MachineIdentifier string `json:"machineIdentifier"`
	Name              string `json:"friendlyName"`
	Version           string `json:"version"`
	Platform          string `json:"platform"`
}

// UserInfo 表示Plex用户信息
type UserInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Thumb       string `json:"thumb"`
	HasPassword bool   `json:"hasPassword"`
}

// LibraryInfo 表示媒体库信息
type LibraryInfo struct {
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Agent     string    `json:"agent"`
	Scanner   string    `json:"scanner"`
	Language  string    `json:"language"`
	UUID      string    `json:"uuid"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
	ScannedAt time.Time `json:"scannedAt"`
	Content   bool      `json:"content"`
	Directory bool      `json:"directory"`
	Hidden    bool      `json:"hidden"`
	Refresh   bool      `json:"refreshing"`
}

// MediaItem 表示媒体项信息
type MediaItem struct {
	RatingKey      string    `json:"ratingKey"`
	Key            string    `json:"key"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	MediaType      string    `json:"mediaType"`
	Year           int       `json:"year"`
	Summary        string    `json:"summary"`
	Thumb          string    `json:"thumb"`
	Art            string    `json:"art"`
	Duration       int64     `json:"duration"`
	ViewCount      int       `json:"viewCount"`
	LastViewedAt   time.Time `json:"lastViewedAt"`
	AddedAt        time.Time `json:"addedAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Guid           string    `json:"guid"`
	Rating         float64   `json:"rating"`
	AudienceRating float64   `json:"audienceRating"`
}

// PlaybackInfo 表示播放信息
type PlaybackInfo struct {
	SessionKey   string    `json:"sessionKey"`
	RatingKey    string    `json:"ratingKey"`
	Key          string    `json:"key"`
	Title        string    `json:"title"`
	Type         string    `json:"type"`
	ViewOffset   int64     `json:"viewOffset"`
	Duration     int64     `json:"duration"`
	State        string    `json:"state"`
	User         string    `json:"user"`
	Client       string    `json:"client"`
	Player       string    `json:"player"`
	LastViewedAt time.Time `json:"lastViewedAt"`
}

// ClientConfig 表示Plex客户端配置
type ClientConfig struct {
	URL        string        `json:"url"`
	Token      string        `json:"token"`
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	RetryDelay time.Duration `json:"retry_delay"`
	Enabled    bool          `json:"enabled"`
}
