package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// SiteStatisticRepository 站点统计仓储接口
type SiteStatisticRepository interface {
	// Create 创建站点统计
	Create(ctx context.Context, statistic *database.SiteStatistic) error

	// GetByID 根据ID获取站点统计
	GetByID(ctx context.Context, id string) (*database.SiteStatistic, error)

	// GetBySiteID 根据站点ID获取统计
	GetBySiteID(ctx context.Context, siteID string) (*database.SiteStatistic, error)

	// Update 更新站点统计
	Update(ctx context.Context, statistic *database.SiteStatistic) error

	// Delete 删除站点统计
	Delete(ctx context.Context, id string) error

	// List 获取站点统计列表
	List(ctx context.Context, params ListSiteStatisticParams) ([]*database.SiteStatistic, int64, error)

	// UpdateStatistics 更新统计数据
	UpdateStatistics(ctx context.Context, siteID string, increment map[string]int64) error
}

// ListSiteStatisticParams 站点统计列表查询参数
type ListSiteStatisticParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	SiteID   string `json:"site_id"`
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}
