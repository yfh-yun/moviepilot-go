package interfaces

import (
	"context"
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// SiteRepository 站点仓储接口
type SiteRepository interface {
	// Create 创建站点
	Create(site *model.Site) error
	
	// GetByID 根据ID获取站点
	GetByID(id uint) (*model.Site, error)
	
	// GetByName 根据名称获取站点
	GetByName(name string) (*model.Site, error)
	
	// GetByDomain 根据域名获取站点
	GetByDomain(domain string) (*model.Site, error)
	
	// GetActive 获取活跃站点
	GetActive() ([]*model.Site, error)
	
	// GetRSSSites 获取RSS站点
	GetRSSSites() ([]*model.Site, error)
	
	// GetSearchSites 获取搜索站点
	GetSearchSites() ([]*model.Site, error)
	
	// Update 更新站点
	Update(site *model.Site) error
	
	// Delete 删除站点
	Delete(id uint) error
	
	// List 分页获取站点列表
	List(offset, limit int) ([]*model.Site, int64, error)
	
	// Search 搜索站点
	Search(keyword string, offset, limit int) ([]*model.Site, int64, error)
	
	// Count 统计站点数量
	Count() (int64, error)
	
	// 异步查询方法
	GetActiveAsync(ctx context.Context) ([]*model.Site, error)
	GetByDomainAsync(ctx context.Context, domain string) (*model.Site, error)
	ListAsync(ctx context.Context, offset, limit int) ([]*model.Site, int64, error)
	
	// 高级查询方法
	ListOrderByPriority() ([]*model.Site, error)
	ListOrderByPriorityAsync(ctx context.Context) ([]*model.Site, error)
	GetDomainsByIDs(ids []uint) ([]string, error)
	Exists(domain string) (bool, error)
	
	// 批量操作
	BatchCreate(sites []*model.Site) error
	BatchUpdate(sites []*model.Site) error
	BatchDelete(ids []uint) error
	
	// 统计方法
	CountByStatus(isActive bool) (int64, error)
	GetFailCountThreshold(threshold int) ([]*model.Site, error)
	
	// Cookie和认证
	UpdateCookie(domain, cookie string) error
	UpdateFailCount(domain string, increment bool) error
	ResetFailCount(domain string) error
}
