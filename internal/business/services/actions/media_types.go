package actions

import (
	"time"
)

// MediaType 媒体类型枚举
type MediaType string

const (
	MediaTypeMovie      MediaType = "movie"
	MediaTypeSeries     MediaType = "series"
	MediaTypeTVShow     MediaType = "tvshow"
	MediaTypeAnime      MediaType = "anime"
	MediaTypeDocumentary MediaType = "documentary"
)

// MediaSource 媒体来源枚举
type MediaSource string

const (
	MediaSourceTMDB   MediaSource = "tmdb"
	MediaSourceTVDB   MediaSource = "tvdb"
	MediaSourceIMDB   MediaSource = "imdb"
	MediaSourceLocal  MediaSource = "local"
	MediaSourceRSS    MediaSource = "rss"
)

// MediaStatus 媒体状态枚举
type MediaStatus string

const (
	MediaStatusWatching  MediaStatus = "watching"
	MediaStatusCompleted MediaStatus = "completed"
	MediaStatusPaused    MediaStatus = "paused"
	MediaStatusPlan      MediaStatus = "plan"
)

// MediaItem 媒体项目核心结构
type MediaItem struct {
	// 基础信息
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	OriginalTitle string     `json:"original_title,omitempty"`
	Type        MediaType   `json:"type"`
	Source      MediaSource `json:"source"`
	SourceID    string      `json:"source_id"` // 来源系统中的ID
	
	// 元数据
	Overview    string      `json:"overview,omitempty"`
	Poster      string      `json:"poster,omitempty"`
	Backdrop    string      `json:"backdrop,omitempty"`
	Rating      float64     `json:"rating,omitempty"`
	ReleaseDate time.Time   `json:"release_date,omitempty"`
	
	// 系列特有信息
	SeasonCount  int         `json:"season_count,omitempty"`
	EpisodeCount int         `json:"episode_count,omitempty"`
	Seasons      []Season    `json:"seasons,omitempty"`
	
	// 本地状态
	Status       MediaStatus `json:"status,omitempty"`
	Progress     float64     `json:"progress,omitempty"` // 观看进度百分比
	LastWatched  time.Time   `json:"last_watched,omitempty"`
	
	// 下载信息
	IsDownloaded bool        `json:"is_downloaded"`
	DownloadPath string      `json:"download_path,omitempty"`
	
	// 扩展信息
	Genres      []string    `json:"genres,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Season 季信息结构
type Season struct {
	ID           int       `json:"id"`
	SeasonNumber int       `json:"season_number"`
	Title        string    `json:"title,omitempty"`
	Overview     string    `json:"overview,omitempty"`
	Poster       string    `json:"poster,omitempty"`
	EpisodeCount int       `json:"episode_count"`
	Episodes     []Episode `json:"episodes,omitempty"`
	IsDownloaded bool      `json:"is_downloaded"`
}

// Episode 剧集信息结构
type Episode struct {
	ID            int       `json:"id"`
	EpisodeNumber int       `json:"episode_number"`
	SeasonNumber  int       `json:"season_number"`
	Title         string    `json:"title,omitempty"`
	Overview      string    `json:"overview,omitempty"`
	AirDate       time.Time `json:"air_date,omitempty"`
	IsWatched     bool      `json:"is_watched"`
	IsDownloaded  bool      `json:"is_downloaded"`
	FileSize      int64     `json:"file_size,omitempty"`
	FilePath      string    `json:"file_path,omitempty"`
}

// MediaSearchParams 媒体搜索参数
type MediaSearchParams struct {
	Query        string     `json:"query" form:"query"`
	Type         MediaType  `json:"type" form:"type"`
	Year         int        `json:"year" form:"year"`
	Page         int        `json:"page" form:"page"`
	Limit        int        `json:"limit" form:"limit"`
	SortBy       string     `json:"sort_by" form:"sort_by"`
	SortOrder    string     `json:"sort_order" form:"sort_order"` // asc or desc
	WithGenres   []string   `json:"with_genres" form:"with_genres"`
	WithTags     []string   `json:"with_tags" form:"with_tags"`
	Status       MediaStatus `json:"status" form:"status"`
	IsDownloaded *bool      `json:"is_downloaded" form:"is_downloaded"`
}

// MediaResponse 媒体响应结构
type MediaResponse struct {
	Items       []MediaItem `json:"items"`
	Total       int64       `json:"total"`
	Page        int         `json:"page"`
	PageSize    int         `json:"page_size"`
	TotalPages  int         `json:"total_pages"`
}

// MediaDetailParams 媒体详情参数
type MediaDetailParams struct {
	MediaID    string     `json:"media_id" form:"media_id" binding:"required"`
	IncludeEpisodes bool   `json:"include_episodes" form:"include_episodes"`
	SeasonNumber int      `json:"season_number" form:"season_number"`
}

// MediaUpdateParams 媒体更新参数
type MediaUpdateParams struct {
	MediaID    string                 `json:"media_id" binding:"required"`
	Status     MediaStatus            `json:"status,omitempty"`
	Progress   *float64               `json:"progress,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// EpisodeUpdateParams 剧集更新参数
type EpisodeUpdateParams struct {
	MediaID      string `json:"media_id" binding:"required"`
	SeasonNumber int    `json:"season_number" binding:"required"`
	EpisodeNumber int   `json:"episode_number" binding:"required"`
	IsWatched    *bool  `json:"is_watched,omitempty"`
}
