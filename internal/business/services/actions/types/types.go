// Package types 定义动作系统中使用的数据类型
package types

import (
	"time"
)

// TorrentInfo 种子信息
type TorrentInfo struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	DownloadURL  string            `json:"download_url"`
	Hash         string            `json:"hash"`
	Size         int64             `json:"size"`
	SiteID       string            `json:"site_id"`
	SiteName     string            `json:"site_name"`
	Type         string            `json:"type"`         // movie, tv, documentary
	Year         int               `json:"year"`
	Season       int               `json:"season"`
	Episodes     []int             `json:"episodes"`
	Resolution   string            `json:"resolution"`    // 1080p, 720p, 4k
	Quality      string            `json:"quality"`       // bluray, webdl, hdtv
	Codec        string            `json:"codec"`         // h264, h265, xvid
	Audio        string            `json:"audio"`         // dts, ac3, aac
	Source       string            `json:"source"`        // 完整来源信息
	PubDate      time.Time         `json:"pub_date"`
	Seeders      int               `json:"seeders"`
	Leechers     int               `json:"leechers"`
	Completed    int               `json:"completed"`
	IMDBID       string            `json:"imdb_id"`
	TMDBID       int               `json:"tmdb_id"`
	DoubanID     string            `json:"douban_id"`
	BangumiID    int               `json:"bangumi_id"`
	Category     string            `json:"category"`
	Tags         []string          `json:"tags"`
	Metadata     map[string]string `json:"metadata"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	ID          int               `json:"id"`
	TMDBID      int               `json:"tmdb_id"`
	IMDBID      string            `json:"imdb_id"`
	DoubanID    string            `json:"douban_id"`
	BangumiID   int               `json:"bangumi_id"`
	Title       string            `json:"title"`
	OriginalTitle string          `json:"original_title"`
	Year        int               `json:"year"`
	Type        string            `json:"type"`         // movie, tv, documentary
	Season      int               `json:"season"`
	Episodes    []int             `json:"episodes"`
	Resolution  string            `json:"resolution"`
	Quality     string            `json:"quality"`
	Codec       string            `json:"codec"`
	Audio       string            `json:"audio"`
	Source      string            `json:"source"`
	Overview    string            `json:"overview"`
	Genres      []string          `json:"genres"`
	Countries   []string          `json:"countries"`
	Languages   []string          `json:"languages"`
	Directors   []string          `json:"directors"`
	Actors      []string          `json:"actors"`
	Rating      float64           `json:"rating"`
	Popularity  float64           `json:"popularity"`
	Poster      string            `json:"poster"`
	Backdrop    string            `json:"backdrop"`
	Trailer     string            `json:"trailer"`
	Homepage    string            `json:"homepage"`
	Status      string            `json:"status"`
	ReleaseDate time.Time         `json:"release_date"`
	Runtime     int               `json:"runtime"`      // 运行时间（分钟）
	Budget      int64             `json:"budget"`       // 预算
	Revenue     int64             `json:"revenue"`      // 收入
	Networks    []string          `json:"networks"`     // 电视网络
	Seasons     []SeasonInfo      `json:"seasons"`      // 季度信息
	Metadata    map[string]string `json:"metadata"`
}

// SeasonInfo 季度信息
type SeasonInfo struct {
	SeasonNumber int       `json:"season_number"`
	EpisodeCount int       `json:"episode_count"`
	AirDate      time.Time `json:"air_date"`
	Poster       string    `json:"poster"`
	Overview     string    `json:"overview"`
}

// Download 下载任务
type Download struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Hash        string    `json:"hash"`
	Size        int64     `json:"size"`
	Type        string    `json:"type"`
	Season      int       `json:"season"`
	Episodes    []int     `json:"episodes"`
	Downloader  string    `json:"downloader"`
	SavePath    string    `json:"save_path"`
	Labels      []string  `json:"labels"`
	Status      string    `json:"status"`      // pending, downloading, completed, failed, cancelled, paused
	Progress    float64   `json:"progress"`    // 0-100
	Speed       int64     `json:"speed"`       // bytes/s
	SiteID      string    `json:"site_id"`
	SiteName    string    `json:"site_name"`
	MediaID     int       `json:"media_id"`
	UserID      uint      `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
	RetryCount  int       `json:"retry_count"`
	Priority    int       `json:"priority"`    // 优先级 1-10
}

// Subscribe 订阅
type Subscribe struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Year        int       `json:"year"`
	Season      int       `json:"season"`
	TMDBID      int       `json:"tmdb_id"`
	IMDBID      string    `json:"imdb_id"`
	DoubanID    string    `json:"douban_id"`
	BangumiID   int       `json:"bangumi_id"`
	Username    string    `json:"username"`
	Status      string    `json:"status"`      // active, paused, completed, failed
	Keyword     string    `json:"keyword"`
	Quality     string    `json:"quality"`
	Resolution  string    `json:"resolution"`
	Filter      string    `json:"filter"`
	Sites       []string  `json:"sites"`
	Downloaders []string  `json:"downloaders"`
	SavePath    string    `json:"save_path"`
	TotalEpisodes int     `json:"total_episodes"`
	DownloadedEpisodes []int `json:"downloaded_episodes"`
	LatestSeason int      `json:"latest_season"`
	LatestEpisode int      `json:"latest_episode"`
	LastUpdate  time.Time `json:"last_update"`
	NextCheck   time.Time `json:"next_check"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Note        string    `json:"note"`
	Priority    int       `json:"priority"`
	AutoSearch  bool      `json:"auto_search"`
	AutoDownload bool     `json:"auto_download"`
}

// Media 媒体库条目
type Media struct {
	ID             int       `json:"id"`
	TMDBID         int       `json:"tmdb_id"`
	IMDBID         string    `json:"imdb_id"`
	DoubanID       string    `json:"douban_id"`
	BangumiID      int       `json:"bangumi_id"`
	Title          string    `json:"title"`
	OriginalTitle  string    `json:"original_title"`
	Year           int       `json:"year"`
	Type           string    `json:"type"`
	Season         int       `json:"season"`
	Episodes       []int     `json:"episodes"`
	Resolution     string    `json:"resolution"`
	Quality        string    `json:"quality"`
	Codec          string    `json:"codec"`
	Audio          string    `json:"audio"`
	Source         string    `json:"source"`
	Overview       string    `json:"overview"`
	Genres         []string  `json:"genres"`
	Countries      []string  `json:"countries"`
	Languages      []string  `json:"languages"`
	Directors      []string  `json:"directors"`
	Actors         []string  `json:"actors"`
	Rating         float64   `json:"rating"`
	Popularity     float64   `json:"popularity"`
	Poster         string    `json:"poster"`
	Backdrop       string    `json:"backdrop"`
	Trailer        string    `json:"trailer"`
	Homepage       string    `json:"homepage"`
	Status         string    `json:"status"`
	ReleaseDate    time.Time `json:"release_date"`
	Runtime        int       `json:"runtime"`
	Budget         int64     `json:"budget"`
	Revenue        int64     `json:"revenue"`
	Networks       []string  `json:"networks"`
	Seasons        []SeasonInfo `json:"seasons"`
	Path           string    `json:"path"`
	Size           int64     `json:"size"`
	Format         string    `json:"format"`
	VideoCodec     string    `json:"video_codec"`
	AudioCodec     string    `json:"audio_codec"`
	ResolutionReal string    `json:"resolution_real"`
	Hash           string    `json:"hash"`
	Note           string    `json:"note"`
	UserID         uint      `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastScraped    *time.Time `json:"last_scraped,omitempty"`
}

// File 文件信息
type File struct {
	ID           int       `json:"id"`
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Size         int64     `json:"size"`
	Type         string    `json:"type"`         // video, audio, subtitle, image, document
	MediaType    string    `json:"media_type"`   // movie, tv, documentary
	Title        string    `json:"title"`
	Year         int       `json:"year"`
	Season       int       `json:"season"`
	Episodes     []int     `json:"episodes"`
	Resolution   string    `json:"resolution"`
	Quality      string    `json:"quality"`
	Codec        string    `json:"codec"`
	Audio        string    `json:"audio"`
	Source       string    `json:"source"`
	TMDBID       int       `json:"tmdb_id"`
	IMDBID       string    `json:"imdb_id"`
	DoubanID     string    `json:"douban_id"`
	BangumiID    int       `json:"bangumi_id"`
	Hash         string    `json:"hash"`
	ParentPath   string    `json:"parent_path"`
	StoragePath  string    `json:"storage_path"`
	TransferPath string    `json:"transfer_path,omitempty"`
	Status       string    `json:"status"`      // local, transferred, ignored, error
	Note         string    `json:"note"`
	UserID       uint      `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ScrapedAt    *time.Time `json:"scraped_at,omitempty"`
	TransferredAt *time.Time `json:"transferred_at,omitempty"`
}

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	UserID    uint                   `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`     // info, warning, error, debug
	Tags      []string               `json:"tags"`
}

// Message 消息
type Message struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`      // info, warning, error, success
	Channel   string    `json:"channel"`   // wechat, telegram, email, webhook
	UserID    uint      `json:"user_id"`
	Status    string    `json:"status"`    // unread, read, sent, failed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	ErrorMsg  string    `json:"error_msg,omitempty"`
	Metadata  map[string]string `json:"metadata"`
}

// Plugin 插件
type Plugin struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Identifier  string    `json:"identifier"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Homepage    string    `json:"homepage"`
	Icon        string    `json:"icon"`
	Status      string    `json:"status"`      // active, inactive, error
	Config      string    `json:"config"`      // JSON配置
	Data        string    `json:"data"`        // JSON数据
	UserID      uint      `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
}

// Note 备注
type Note struct {
	ID        int       `json:"id"`
	TargetID  int       `json:"target_id"`   // 目标ID（媒体ID、订阅ID等）
	TargetType string   `json:"target_type"` // 目标类型（media、subscribe等）
	Content   string    `json:"content"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Site 站点
type Site struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	URL         string    `json:"url"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	Type        string    `json:"type"`        // public, private, semi-private`
	Rank        int       `json:"rank"`        // 站点等级
	Level       string    `json:"level"`       // 用户等级
	Cookie      string    `json:"cookie"`
	Headers     string    `json:"headers"`     // JSON格式的请求头
	Proxy       string    `json:"proxy"`       // 代理设置
	Limit       int       `json:"limit"`       // 下载限制
	Interval    int       `json:"interval"`    // 搜索间隔
	Search      bool      `json:"search"`      // 是否启用搜索
	Download    bool      `json:"download"`    // 是否启用下载`
	Seeder      bool      `json:"seeder"`      // 是否做种`
	BR          bool      `json:"br"`          // 是否启用浏览器渲染`
	IMAX        bool      `json:"imax"`        // 是否支持IMAX`
	Dolby       bool      `json:"dolby"`       // 是否支持杜比`
	HDR         bool      `json:"hdr"`         // 是否支持HDR`
	UHD         bool      `json:"uhd"`         // 是否支持4K`
	Free        bool      `json:"free"`        // 是否免费`
	2XFree      bool      `json:"2xfree"`      // 是否2X免费`
	UserID      uint      `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastTest    *time.Time `json:"last_test,omitempty"`
}

// Workflow 工作流
type Workflow struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Trigger     string                 `json:"trigger"`     // manual, schedule, event
	Config      string                 `json:"config"`      // JSON配置
	Actions     []WorkflowAction       `json:"actions"`
	Status      string                 `json:"status"`      // active, inactive, error
	UserID      uint                   `json:"user_id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     *time.Time             `json:"next_run,omitempty"`
	RunCount    int                    `json:"run_count"`
	SuccessCount int                   `json:"success_count"`
	ErrorCount  int                    `json:"error_count"`
}

// WorkflowAction 工作流动作
type WorkflowAction struct {
	ID       int                    `json:"id"`
	Type     string                 `json:"type"`       // action类型
	Name     string                 `json:"name"`       // 动作名称
	Config   map[string]interface{} `json:"config"`     // 动作配置
	Order    int                    `json:"order"`      // 执行顺序
	Enabled  bool                   `json:"enabled"`    // 是否启用
	Timeout  int                    `json:"timeout"`    // 超时时间（秒）
	Retry    int                    `json:"retry"`      // 重试次数
}

// ActionContext 动作上下文
type ActionContext struct {
	WorkflowID int64          `json:"workflow_id"`
	Progress   int            `json:"progress"`
	Message    string         `json:"message"`
	Data       map[string]interface{} `json:"data"`
	
	// 媒体相关
	Medias     []*MediaInfo   `json:"medias"`
	Torrents   []*TorrentInfo `json:"torrents"`
	
	// 下载相关
	Downloads  []*Download    `json:"downloads"`
	
	// 订阅相关
	Subscribes []*Subscribe   `json:"subscribes"`
	
	// 文件相关
	Files      []*File        `json:"files"`
	
	// 消息相关
	Messages   []*Message     `json:"messages"`
	
	// 事件相关
	Events     []*Event       `json:"events"`
	
	// 插件相关
	Plugins    []*Plugin      `json:"plugins"`
	
	// 备注相关
	Notes      []*Note        `json:"notes"`
	
	// 站点相关
	Sites      []*Site        `json:"sites"`
	
	// 元数据
	Metadata   map[string]interface{} `json:"metadata"`
	
	// 时间戳
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// ActionParams 动作参数
type ActionParams map[string]interface{}

// EventData 事件数据
type EventData map[string]interface{}