// Package models MoviePilot数据模型定义
package models

import (
	"time"

	"gorm.io/gorm"
)

// MediaType 媒体类型枚举
type MediaType string

const (
	MediaTypeMovie  MediaType = "movie"  // 电影
	MediaTypeTV     MediaType = "tv"     // 电视剧
	MediaTypeAnime  MediaType = "anime"  // 动漫
	MediaTypeMusic  MediaType = "music"  // 音乐
	MediaTypeGame   MediaType = "game"   // 游戏
)

// MediaStatus 媒体状态枚举
type MediaStatus string

const (
	MediaStatusReleased      MediaStatus = "released"      // 已上映
	MediaStatusUpcoming      MediaStatus = "upcoming"      // 即将上映
	MediaStatusInProduction  MediaStatus = "in_production" // 制作中
	MediaStatusCancelled     MediaStatus = "cancelled"     // 已取消
	MediaStatusUnknown       MediaStatus = "unknown"       // 未知
)

// MediaItem 媒体元数据模型
type MediaItem struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	UserID        uint         `gorm:"not null;index" json:"user_id"`
	TMDBID        int          `gorm:"not null;index" json:"tmdb_id"`
	IMDBID        string       `gorm:"size:20;index" json:"imdb_id,omitempty"`
	TVDBID        int          `gorm:"index" json:"tvdb_id,omitempty"`
	MediaTyp      MediaType    `gorm:"size:20;not null;index" json:"media_type"` // movie, tv, anime
	Title         string       `gorm:"size:200;not null" json:"title"`
	OriginalTitle string       `gorm:"size:200" json:"original_title,omitempty"`
	Year          int          `gorm:"index" json:"year,omitempty"`
	Overview      string       `gorm:"type:text" json:"overview,omitempty"`
	PosterPath    string       `gorm:"size:500" json:"poster_path,omitempty"`
	BackdropPath  string       `gorm:"size:500" json:"backdrop_path,omitempty"`
	Rating        float64      `gorm:"type:decimal(3,1)" json:"rating,omitempty"`
	ReleaseDate   *time.Time   `json:"release_date,omitempty"`
	Status        MediaStatus  `gorm:"size:20" json:"status,omitempty"`
	Genres        []string     `gorm:"type:text[]" json:"genres,omitempty"`
	Countries     []string     `gorm:"type:text[]" json:"countries,omitempty"`
	Languages     []string     `gorm:"type:text[]" json:"languages,omitempty"`
	Runtime       int          `json:"runtime,omitempty"`
	EpisodeCount  int          `json:"episode_count,omitempty"`
	SeasonCount   int          `json:"season_count,omitempty"`

	// 自定义字段
	CustomTitle   string       `gorm:"size:200" json:"custom_title,omitempty"`
	Tags          []string     `gorm:"type:text[]" json:"tags,omitempty"`
	Note          string       `gorm:"type:text" json:"note,omitempty"`

	// 状态字段
	Favorite      bool         `gorm:"default:false;not null;index" json:"favorite"`
	Watched       bool         `gorm:"default:false;not null;index" json:"watched"`
	InLibrary     bool         `gorm:"default:false;not null;index" json:"in_library"`

	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Versions      []MediaVersion `gorm:"foreignKey:MediaID" json:"versions,omitempty"`
	Files         []MediaFile     `gorm:"foreignKey:MediaID" json:"files,omitempty"`
	Subscriptions []Subscription   `gorm:"foreignKey:MediaID" json:"subscriptions,omitempty"`
}

// TableName 指定表名
func (MediaItem) TableName() string {
	return "media_items"
}

// BeforeCreate GORM钩子：创建前
func (m *MediaItem) BeforeCreate(tx *gorm.DB) error {
	// 如果年份为空且有发布日期，从发布日期提取年份
	if m.Year == 0 && m.ReleaseDate != nil {
		m.Year = m.ReleaseDate.Year()
	}
	return nil
}

// BeforeUpdate GORM钩子：更新前
func (m *MediaItem) BeforeUpdate(tx *gorm.DB) error {
	// 如果年份为空且有发布日期，从发布日期提取年份
	if m.Year == 0 && m.ReleaseDate != nil {
		m.Year = m.ReleaseDate.Year()
	}
	return nil
}

// MediaVersion 媒体版本模型（用于电视剧的季、集）
type MediaVersion struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	MediaID       uint         `gorm:"not null;index" json:"media_id"`
	Season        int          `gorm:"index" json:"season,omitempty"`
	Episode       int          `gorm:"index" json:"episode,omitempty"`
	Title         string       `gorm:"size:200" json:"title,omitempty"`
	OriginalTitle string       `gorm:"size:200" json:"original_title,omitempty"`
	Overview      string       `gorm:"type:text" json:"overview,omitempty"`
	PosterPath    string       `gorm:"size:500" json:"poster_path,omitempty"`
	StillPath     string       `gorm:"size:500" json:"still_path,omitempty"`
	Rating        float64      `gorm:"type:decimal(3,1)" json:"rating,omitempty"`
	ReleaseDate   *time.Time   `json:"release_date,omitempty"`
	Runtime       int          `json:"runtime,omitempty"`
	EpisodeCount  int          `json:"episode_count,omitempty"`

	// 自定义字段
	CustomTitle   string       `gorm:"size:200" json:"custom_title,omitempty"`
	Tags          []string     `gorm:"type:text[]" json:"tags,omitempty"`
	Note          string       `gorm:"type:text" json:"note,omitempty"`

	// 状态字段
	Watched       bool         `gorm:"default:false;not null;index" json:"watched"`
	InLibrary     bool         `gorm:"default:false;not null;index" json:"in_library"`

	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Media         MediaItem     `gorm:"foreignKey:MediaID" json:"media,omitempty"`
	Files         []MediaFile    `gorm:"foreignKey:VersionID" json:"files,omitempty"`
}

// TableName 指定表名
func (MediaVersion) TableName() string {
	return "media_versions"
}

// MediaFile 媒体文件模型
type MediaFile struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	MediaID         *uint          `gorm:"index" json:"media_id,omitempty"`
	VersionID       *uint          `gorm:"index" json:"version_id,omitempty"`
	FilePath        string         `gorm:"size:1000;not null;uniqueIndex" json:"file_path"`
	FileName        string         `gorm:"size:500;not null" json:"file_name"`
	FileSize        int64          `gorm:"not null" json:"file_size"`
	FileHash        string         `gorm:"size:64;index" json:"file_hash,omitempty"` // SHA1或MD5
	VideoCodec      string         `gorm:"size:50" json:"video_codec,omitempty"`
	AudioCodec      string         `gorm:"size:50" json:"audio_codec,omitempty"`
	Resolution      string         `gorm:"size:20;index" json:"resolution,omitempty"` // 1080p, 4K
	Duration        int            `json:"duration,omitempty"`                        // 秒
	Width           int            `json:"width,omitempty"`                           // 视频宽度
	Height          int            `json:"height,omitempty"`                          // 视频高度
	BitRate         int            `json:"bit_rate,omitempty"`                       // 码率（bps）
	FrameRate       float64        `gorm:"type:decimal(4,2)" json:"frame_rate,omitempty"` // 帧率
	AudioChannels   int            `json:"audio_channels,omitempty"`                 // 音频声道数
	SubtitleLanguages []string     `gorm:"type:text[]" json:"subtitle_languages,omitempty"`
	AudioLanguages   []string     `gorm:"type:text[]" json:"audio_languages,omitempty"`

	// 状态字段
	Processed       bool           `gorm:"default:false;not null;index" json:"processed"`
	InLibrary       bool           `gorm:"default:false;not null;index" json:"in_library"`
	IsDuplicate     bool           `gorm:"default:false;not null;index" json:"is_duplicate"`

	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Media           *MediaItem     `gorm:"foreignKey:MediaID" json:"media,omitempty"`
	Version         *MediaVersion  `gorm:"foreignKey:VersionID" json:"version,omitempty"`
}

// TableName 指定表名
func (MediaFile) TableName() string {
	return "media_files"
}

// MetadataScrape 元数据刮削记录模型
type MetadataScrape struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	MediaID     uint         `gorm:"not null;index" json:"media_id"`
	Source      string       `gorm:"size:20;not null;index" json:"source"` // tmdb, tvdb, imdb, themoviedb, douban
	Data        string       `gorm:"type:jsonb" json:"data"`              // 刮削到的原始数据
	Status      string       `gorm:"size:20;not null" json:"status"`     // success, failed, partial
	ErrorMessage string      `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`

	// 关联
	Media       MediaItem    `gorm:"foreignKey:MediaID" json:"media,omitempty"`
}

// TableName 指定表名
func (MetadataScrape) TableName() string {
	return "metadata_scrapes"
}
