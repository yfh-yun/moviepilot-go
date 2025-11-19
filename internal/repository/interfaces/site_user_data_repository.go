package interfaces

import "github.com/yfh-yun/moviepilot-go/internal/model"

// SiteUserDataRepository 站点用户数据仓储接口
type SiteUserDataRepository interface {
	// 基础CRUD
	Create(userData *model.SiteUserData) error
	GetByID(id uint) (*model.SiteUserData, error)
	Update(userData *model.SiteUserData) error
	Delete(id uint) error
	List(offset, limit int) ([]*model.SiteUserData, int64, error)
	
	// 按域名获取
	GetByDomain(domain string) ([]*model.SiteUserData, error)
	AsyncGetByDomain(domain string) ([]*model.SiteUserData, error)
	
	// 按域名和日期获取
	GetByDomainAndDate(domain, workdate *string) ([]*model.SiteUserData, error)
	
	// 按日期获取
	GetByDate(date string) ([]*model.SiteUserData, error)
	
	// 获取最新数据
	GetLatest() ([]*model.SiteUserData, error)
	
	// 按用户获取
	GetByUsername(username string) ([]*model.SiteUserData, error)
	
	// 更新用户数据
	UpdateUserData(domain, name string, payload map[string]interface{}) error
	
	// 批量操作
	BatchCreate(userDataList []*model.SiteUserData) error
}