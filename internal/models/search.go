package models

import (
	"time"
)

// SearchHistory 搜索历史记录模型
type SearchHistory struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index"`
	Query       string    `json:"query"`
	SearchType  string    `json:"search_type"` // media, torrent, site
	ResultCount int       `json:"result_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// SearchStatistics 搜索统计信息
type SearchStatistics struct {
	TotalSearches   int64     `json:"total_searches"`
	MediaSearches   int64     `json:"media_searches"`
	TorrentSearches int64     `json:"torrent_searches"`
	SiteSearches    int64     `json:"site_searches"`
	LastSearchTime  time.Time `json:"last_search_time"`
	PopularQueries  []string  `json:"popular_queries"`
}

// SearchResult 搜索结果模型
type SearchResult struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Type      string  `json:"type"`
	Year      string  `json:"year,omitempty"`
	Poster    string  `json:"poster,omitempty"`
	Rating    float64 `json:"rating,omitempty"`
	Overview  string  `json:"overview,omitempty"`
	Source    string  `json:"source"`
	Relevance float64 `json:"relevance"`
}

// SearchRequest 搜索请求参数
type SearchRequest struct {
	Query    string `json:"query"`
	Type     string `json:"type,omitempty"`
	Year     string `json:"year,omitempty"`
	Language string `json:"language,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	HasMore  bool           `json:"has_more"`
}
