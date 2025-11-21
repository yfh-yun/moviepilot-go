package actions

import (
	"time"
)

// RSSType RSS类型枚举
type RSSType string

const (
	RSSTypeMovie      RSSType = "movie"
	RSSTypeSeries     RSSType = "series"
	RSSTypeAnimation  RSSType = "animation"
	RSSTypeDocumentary RSSType = "documentary"
	RSSTypeAll        RSSType = "all"
)

// RSSFormat RSS格式枚举
type RSSFormat string

const (
	RSSFormatXML      RSSFormat = "xml"
	RSSFormatJSON     RSSFormat = "json"
	RSSFormatTorrent  RSSFormat = "torrent"
	RSSFormatCustom   RSSFormat = "custom"
)

// RSSErrorType RSS错误类型枚举
type RSSErrorType string

const (
	RSSErrorTypeNetwork    RSSErrorType = "network"
	RSSErrorTypeParse      RSSErrorType = "parse"
	RSSErrorTypeTimeout    RSSErrorType = "timeout"
	RSSErrorTypeAuth       RSSErrorType = "auth"
	RSSErrorTypeValidation RSSErrorType = "validation"
	RSSErrorTypeUnknown    RSSErrorType = "unknown"
)

// RSSEntry RSS条目结构
type RSSEntry struct {
	// 基础信息
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description,omitempty"`
	PublishedAt time.Time `json:"published_at"`

	// 媒体相关信息
	MediaType   RSSType   `json:"media_type,omitempty"`
	Category    string    `json:"category,omitempty"`
	Tags        []string  `json:"tags,omitempty"`

	// 下载相关信息
	TorrentURL  string    `json:"torrent_url,omitempty"`
	MagnetURL   string    `json:"magnet_url,omitempty"`
	FileSize    int64     `json:"file_size,omitempty"`
	Seeders     int       `json:"seeders,omitempty"`
	Leechers    int       `json:"leechers,omitempty"`

	// 扩展信息
	Enclosure   *Enclosure `json:"enclosure,omitempty"`
	Author      string     `json:"author,omitempty"`
	GUID        string     `json:"guid,omitempty"`
	Comments    string     `json:"comments,omitempty"`
	MediaInfo   *MediaInfo `json:"media_info,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Enclosure RSS附件结构
type Enclosure struct {
	URL    string `json:"url"`
	Length int64  `json:"length"`
	Type   string `json:"type"`
}

// MediaInfo RSS媒体信息结构
type MediaInfo struct {
	Title    string   `json:"title,omitempty"`
	Thumbnail string  `json:"thumbnail,omitempty"`
	Content  string   `json:"content,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}

// RSSFeed RSS源结构
type RSSFeed struct {
	// 源信息
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description,omitempty"`
	Language    string    `json:"language,omitempty"`
	Author      string    `json:"author,omitempty"`
	LastBuild   time.Time `json:"last_build,omitempty"`
	Generator   string    `json:"generator,omitempty"`

	// 配置信息
	FeedURL     string    `json:"feed_url"`
	Format      RSSFormat `json:"format"`
	Type        RSSType   `json:"type"`
	Interval    int       `json:"interval"` // 刷新间隔（分钟）
	Enabled     bool      `json:"enabled"`

	// 认证信息
	Username    string    `json:"username,omitempty"`
	Password    string    `json:"password,omitempty"`
	Cookies     string    `json:"cookies,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`

	// 状态信息
	LastFetch   time.Time `json:"last_fetch,omitempty"`
	LastSuccess time.Time `json:"last_success,omitempty"`
	ErrorCount  int       `json:"error_count,omitempty"`
	LastError   string    `json:"last_error,omitempty"`

	// 缓存配置
	CacheEnabled bool      `json:"cache_enabled"`
	CacheTTL     int       `json:"cache_ttl"` // 缓存时间（分钟）

	// 过滤配置
	Filters     *RSSFilters `json:"filters,omitempty"`

	// 扩展信息
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RSSFilters RSS过滤条件结构
type RSSFilters struct {
	IncludeTitle     []string `json:"include_title,omitempty"`
	ExcludeTitle     []string `json:"exclude_title,omitempty"`
	IncludeKeywords  []string `json:"include_keywords,omitempty"`
	ExcludeKeywords  []string `json:"exclude_keywords,omitempty"`
	MinSize          int64    `json:"min_size,omitempty"`
	MaxSize          int64    `json:"max_size,omitempty"`
	MinSeeders       int      `json:"min_seeders,omitempty"`
	Categories       []string `json:"categories,omitempty"`
	MediaTypes       []RSSType `json:"media_types,omitempty"`
	Resolution       []string `json:"resolution,omitempty"`
	Codecs           []string `json:"codecs,omitempty"`
	Sources          []string `json:"sources,omitempty"`
}

// FetchRSSParams 获取RSS参数结构
type FetchRSSParams struct {
	// 源配置
	FeedURL     string    `json:"feed_url" binding:"required"`
	Format      RSSFormat `json:"format"`
	Type        RSSType   `json:"type"`

	// 请求配置
	Timeout     int       `json:"timeout"` // 超时时间（秒）
	Retries     int       `json:"retries"`
	Delay       int       `json:"delay"`   // 重试延迟（秒）
	UserAgent   string    `json:"user_agent,omitempty"`
	Username    string    `json:"username,omitempty"`
	Password    string    `json:"password,omitempty"`
	Cookies     string    `json:"cookies,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`

	// 过滤配置
	Filters     *RSSFilters `json:"filters,omitempty"`

	// 处理配置
	Limit       int       `json:"limit,omitempty"`
	CacheEnabled bool      `json:"cache_enabled"`
	CacheTTL    int       `json:"cache_ttl"`
	ParseTorrent bool      `json:"parse_torrent"` // 是否解析种子信息
	UseProxy    bool      `json:"use_proxy"`
}

// RSSResponse RSS响应结构
type RSSResponse struct {
	Success      bool        `json:"success"`
	Feed         *RSSFeed    `json:"feed,omitempty"`
	Entries      []RSSEntry  `json:"entries"`
	Total        int         `json:"total"`
	Filtered     int         `json:"filtered,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
	CacheHit     bool        `json:"cache_hit"`
	Error        *RSSError   `json:"error,omitempty"`
}

// RSSError RSS错误结构
type RSSError struct {
	Type        RSSErrorType `json:"type"`
	Message     string       `json:"message"`
	Code        int          `json:"code,omitempty"`
	Details     string       `json:"details,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
}

// RSSConfig RSS配置结构（用于批量管理）
type RSSConfig struct {
	Feeds       []RSSFeed    `json:"feeds"`
	GlobalFilters *RSSFilters `json:"global_filters,omitempty"`
	DefaultInterval int       `json:"default_interval"`
	DefaultTimeout  int       `json:"default_timeout"`
	MaxConcurrent   int       `json:"max_concurrent"`
}

// RSSStats RSS统计信息
type RSSStats struct {
	TotalFeeds    int     `json:"total_feeds"`
	EnabledFeeds  int     `json:"enabled_feeds"`
	DisabledFeeds int     `json:"disabled_feeds"`
	TotalEntries  int64   `json:"total_entries"`
	ErrorFeeds    int     `json:"error_feeds"`
	LastUpdated   time.Time `json:"last_updated"`
	SuccessRate   float64 `json:"success_rate"`
}
