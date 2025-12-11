package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// MediaRepository 媒体仓储接口
type MediaRepository interface {
	// Create 创建媒体记录
	Create(ctx context.Context, media *database.Media) error

	// GetByID 根据ID获取媒体
	GetByID(ctx context.Context, id string) (*database.Media, error)

	// Update 更新媒体
	Update(ctx context.Context, media *database.Media) error

	// Delete 删除媒体
	Delete(ctx context.Context, id string) error

	// List 获取媒体列表
	List(ctx context.Context, params ListMediaParams) ([]*database.Media, int64, error)

	// Search 搜索媒体
	Search(ctx context.Context, query string, params SearchMediaParams) ([]*database.Media, error)

	// GetByTMDBID 根据TMDB ID获取媒体
	GetByTMDBID(ctx context.Context, tmdbID int64, mediaType string) (*database.Media, error)

	// GetByTitle 根据标题获取媒体
	GetByTitle(ctx context.Context, title string, year int) (*database.Media, error)
}

// ListMediaParams 媒体列表查询参数
type ListMediaParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Type     string `json:"type"`
	Genre    string `json:"genre"`
	Year     int    `json:"year"`
}

// SearchMediaParams 媒体搜索参数
type SearchMediaParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Type     string `json:"type"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
}
