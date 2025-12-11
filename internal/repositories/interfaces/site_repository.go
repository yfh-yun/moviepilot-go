package interfaces

import (
	"context"
	"moviepilot-go/internal/models/database"
)

// SiteRepository 站点仓库接口
type SiteRepository interface {
	// Create 创建站点
	Create(ctx context.Context, site *database.Site) error

	// Update 更新站点
	Update(ctx context.Context, site *database.Site) error

	// Delete 删除站点
	Delete(ctx context.Context, id uint) error

	// GetByID 根据 ID 获取站点
	GetByID(ctx context.Context, id uint) (*database.Site, error)

	// GetByDomain 根据域名获取站点
	GetByDomain(ctx context.Context, domain string) (*database.Site, error)

	// List 获取站点列表
	List(ctx context.Context, opts ListOptions) ([]*database.Site, int64, error)

	// ListActive 获取启用的站点列表
	ListActive(ctx context.Context) ([]*database.Site, error)

	// UpdateStatus 更新站点状态
	UpdateStatus(ctx context.Context, id uint, isActive bool) error

	// IncrementFailCount 增加失败次数
	IncrementFailCount(ctx context.Context, id uint) error

	// ResetFailCount 重置失败次数
	ResetFailCount(ctx context.Context, id uint) error

	// UpdateStatistics 更新站点统计
	UpdateStatistics(ctx context.Context, siteName string, success bool, seconds int) error
}
