package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// SiteUserDataRepository 站点用户数据仓储接口
type SiteUserDataRepository interface {
	// Create 创建站点用户数据
	Create(ctx context.Context, userData *database.SiteUserData) error

	// GetByID 根据ID获取站点用户数据
	GetByID(ctx context.Context, id string) (*database.SiteUserData, error)

	// GetByUserID 根据用户ID获取站点用户数据
	GetByUserID(ctx context.Context, userID string) ([]*database.SiteUserData, error)

	// GetBySiteID 根据站点ID获取用户数据
	GetBySiteID(ctx context.Context, siteID string) ([]*database.SiteUserData, error)

	// GetByUserAndSite 根据用户ID和站点ID获取数据
	GetByUserAndSite(ctx context.Context, userID, siteID string) (*database.SiteUserData, error)

	// Update 更新站点用户数据
	Update(ctx context.Context, userData *database.SiteUserData) error

	// Delete 删除站点用户数据
	Delete(ctx context.Context, id string) error

	// List 获取站点用户数据列表
	List(ctx context.Context, params ListSiteUserDataParams) ([]*database.SiteUserData, int64, error)
}

// ListSiteUserDataParams 站点用户数据列表查询参数
type ListSiteUserDataParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	UserID   string `json:"user_id"`
	SiteID   string `json:"site_id"`
}
