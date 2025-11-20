package interfaces

import "github.com/yfh-yun/moviepilot-go/internal/models"

// SiteIconRepository 站点图标仓储接口
type SiteIconRepository interface {
	// 基础CRUD
	Create(icon *model.SiteIcon) error
	GetByID(id uint) (*model.SiteIcon, error)
	Update(icon *model.SiteIcon) error
	Delete(id uint) error
	List(offset, limit int) ([]*model.SiteIcon, int64, error)
	
	// 按域名获取
	GetByDomain(domain string) (*model.SiteIcon, error)
	AsyncGetByDomain(domain string) (*model.SiteIcon, error)
	
	// 更新图标
	UpdateIcon(name, domain, iconURL, iconBase64 string) error
	
	// 按站点名称获取
	GetBySiteName(siteName string) (*model.SiteIcon, error)
	
	// 清空图标
	Empty() error
}