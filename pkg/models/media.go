package models

import (
	"time"
)

// MediaType 媒体类型枚举
type MediaType string

const (
	Movie    MediaType = "movie"
	TV       MediaType = "tv"
	Unknown  MediaType = "unknown"
)

// MediaInfo 媒体信息结构�?type MediaInfo struct {
	ID           int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	TMDBID       int        `json:"tmdb_id" gorm:"index"`
	IMDBID       string     `json:"imdb_id" gorm:"index"`
	DoubanID     string     `json:"douban_id" gorm:"index"`
	Title        string     `json:"title"`
	OriginalTitle string    `json:"original_title"`
	Year         int        `json:"year"`
	Type         MediaType  `json:"media_type"`
	Genres       []string   `json:"genres" gorm:"serializer:json"`
	Overview     string     `json:"overview"`
	Runtime      int        `json:"runtime"`
	ReleaseDate  time.Time  `json:"release_date"`
	Rating       float64    `json:"rating"`
	PosterPath   string     `json:"poster_path"`
	BackdropPath string     `json:"backdrop_path"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	TitleYear    string     `json:"title_year,omitempty"`
}

// TableName 设置表名
func (m *MediaInfo) TableName() string {
	return "media_info"
}

// Season 季信�?type Season struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	MediaID      int64     `json:"media_id" gorm:"index"`
	SeasonNumber int       `json:"season_number"`
	Name         string    `json:"name"`
	Overview     string    `json:"overview"`
	PosterPath   string    `json:"poster_path"`
	EpisodeCount int       `json:"episode_count"`
	AirDate      time.Time `json:"air_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 设置表名
func (s *Season) TableName() string {
	return "seasons"
}

// Episode 剧集信息
type Episode struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	SeasonID     int64     `json:"season_id" gorm:"index"`
	EpisodeNumber int       `json:"episode_number"`
	Name         string    `json:"name"`
	Overview     string    `json:"overview"`
	AirDate      time.Time `json:"air_date"`
	Runtime      int       `json:"runtime"`
	StillPath    string    `json:"still_path"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 设置表名
func (e *Episode) TableName() string {
	return "episodes"
}
