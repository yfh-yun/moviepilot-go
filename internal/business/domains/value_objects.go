package domains

import (
	"time"
)

// Quality 媒体质量值对象
type Quality struct {
	Resolution string // 1080p, 720p, 4K
	Source     string // BluRay, WEB-DL, HDTV
	Codec      string // H.264, H.265, AV1
	Audio      string // AAC, AC3, DTS
}

// String 返回质量的字符串表示
func (q Quality) String() string {
	result := q.Resolution
	if q.Source != "" {
		result += "." + q.Source
	}
	if q.Codec != "" {
		result += "." + q.Codec
	}
	if q.Audio != "" {
		result += "." + q.Audio
	}
	return result
}

// IsBetterThan 比较质量是否优于另一个
func (q Quality) IsBetterThan(other Quality) bool {
	// 简单的质量比较逻辑
	resolutionOrder := map[string]int{
		"480p": 1,
		"720p": 2,
		"1080p": 3,
		"4K": 4,
		"8K": 5,
	}
	
	sourceOrder := map[string]int{
		"CAM": 1,
		"TS": 2,
		"TC": 3,
		"DVDSCR": 4,
		"DVD": 5,
		"HDTV": 6,
		"WEB-DL": 7,
		"BluRay": 8,
		"Remux": 9,
	}
	
	if resolutionOrder[q.Resolution] > resolutionOrder[other.Resolution] {
		return true
	}
	
	if resolutionOrder[q.Resolution] == resolutionOrder[other.Resolution] {
		return sourceOrder[q.Source] >= sourceOrder[other.Source]
	}
	
	return false
}

// MediaFilter 媒体过滤器值对象
type MediaFilter struct {
	Genres     []string `json:"genres"`
	MinRating  float64  `json:"min_rating"`
	MaxRating  float64  `json:"max_rating"`
	YearFrom   int      `json:"year_from"`
	YearTo     int      `json:"year_to"`
	Quality    Quality  `json:"quality"`
	Keywords   []string `json:"keywords"`
	ExcludeKeywords []string `json:"exclude_keywords"`
}

// Matches 检查媒体是否匹配过滤器
func (f MediaFilter) Matches(media Media) bool {
	// 检查年份
	if f.YearFrom > 0 && media.ReleaseDate.Year() < f.YearFrom {
		return false
	}
	if f.YearTo > 0 && media.ReleaseDate.Year() > f.YearTo {
		return false
	}
	
	// 检查评分（假设有评分字段）
	// 这里需要根据实际的评分字段进行调整
	
	// 检查关键词匹配
	// 这里可以实现更复杂的匹配逻辑
	
	return true
}

// DownloadConfig 下载配置值对象
type DownloadConfig struct {
	MaxConcurrent int           `json:"max_concurrent"`
	Timeout       time.Duration `json:"timeout"`
	RetryCount    int           `json:"retry_count"`
	RetryDelay    time.Duration `json:"retry_delay"`
	UserAgent     string        `json:"user_agent"`
	Headers       map[string]string `json:"headers"`
}

// DefaultDownloadConfig 返回默认下载配置
func DefaultDownloadConfig() DownloadConfig {
	return DownloadConfig{
		MaxConcurrent: 3,
		Timeout:       30 * time.Minute,
		RetryCount:    3,
		RetryDelay:    5 * time.Second,
		UserAgent:     "MoviePilot-Go/1.0",
		Headers:       make(map[string]string),
	}
}

// NotificationConfig 通知配置值对象
type NotificationConfig struct {
	Enabled  bool     `json:"enabled"`
	Channels []string `json:"channels"` // telegram, email, webhook
	Events   []string `json:"events"`   // download_complete, transfer_complete, error
	Template string   `json:"template"`
}

// ShouldNotify 检查是否应该发送通知
func (n NotificationConfig) ShouldNotify(event string) bool {
	if !n.Enabled {
		return false
	}
	
	for _, e := range n.Events {
		if e == event || e == "all" {
			return true
		}
	}
	
	return false
}

// StorageConfig 存储配置值对象
type StorageConfig struct {
	Type       string            `json:"type"`       // local, aliyun, onedrive
	Path       string            `json:"path"`
	Config     map[string]interface{} `json:"config"`
	Metadata   map[string]string `json:"metadata"`
}

// IsLocal 检查是否为本地存储
func (s StorageConfig) IsLocal() bool {
	return s.Type == "local"
}

// IsCloud 检查是否为云存储
func (s StorageConfig) IsCloud() bool {
	return s.Type != "local"
}

// SearchCriteria 搜索条件值对象
type SearchCriteria struct {
	Query      string    `json:"query"`
	MediaType  string    `json:"media_type"`  // movie, tv, all
	Quality    Quality   `json:"quality"`
	Year       int       `json:"year"`
	Genres     []string  `json:"genres"`
	SortBy     string    `json:"sort_by"`     // rating, release_date, popularity
	SortOrder  string    `json:"sort_order"`  // asc, desc
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

// DefaultSearchCriteria 返回默认搜索条件
func DefaultSearchCriteria() SearchCriteria {
	return SearchCriteria{
		MediaType: "all",
		SortBy:    "popularity",
		SortOrder: "desc",
		Page:      1,
		PageSize:  20,
	}
}

// Validate 验证搜索条件
func (s SearchCriteria) Validate() error {
	if s.Query == "" {
		return nil // 空查询是允许的，表示获取所有
	}
	
	if s.Page < 1 {
		s.Page = 1
	}
	
	if s.PageSize < 1 || s.PageSize > 100 {
		s.PageSize = 20
	}
	
	return nil
}