package interfaces

import (
	"context"
	"moviepilot-go/internal/models"
)

// SiteIconRepository 站点图标仓储接口
type SiteIconRepository interface {
	// Create 创建站点图标
	Create(ctx context.Context, icon *models.SiteIcon) error
	
	// GetByID 根据ID获取站点图标
	GetByID(ctx context.Context, id string) (*models.SiteIcon, error)
	
	// GetByDomain 根据域名获取站点图标
	GetByDomain(ctx context.Context, domain string) (*models.SiteIcon, error)
	
	// Update 更新站点图标
	Update(ctx context.Context, icon *models.SiteIcon) error
	
	// Delete 删除站点图标
	Delete(ctx context.Context, id string) error
	
	// List 获取站点图标列表
	List(ctx context.Context, params ListSiteIconParams) ([]*models.SiteIcon, int64, error)
}

// ListSiteIconParams 站点图标列表查询参数
type ListSiteIconParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Domain   string `json:"domain"`
}