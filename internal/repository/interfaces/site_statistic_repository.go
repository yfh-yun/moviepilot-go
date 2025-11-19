package interfaces

import "github.com/yfh-yun/moviepilot-go/internal/model"

// SiteStatisticRepository 站点统计仓储接口
type SiteStatisticRepository interface {
	// 基础CRUD
	Create(statistic *model.SiteStatistic) error
	GetByID(id uint) (*model.SiteStatistic, error)
	Update(statistic *model.SiteStatistic) error
	Delete(id uint) error
	List(offset, limit int) ([]*model.SiteStatistic, int64, error)
	
	// 按域名获取
	GetByDomain(domain string) (*model.SiteStatistic, error)
	AsyncGetByDomain(domain string) (*model.SiteStatistic, error)
	
	// 获取所有统计
	ListAll() ([]*model.SiteStatistic, error)
	
	// 更新统计
	UpdateSuccess(domain string, seconds *int) error
	UpdateFail(domain string) error
	AsyncUpdateSuccess(domain string, seconds *int) error
	AsyncUpdateFail(domain string) error
	
	// 清空统计
	Reset() error
	AsyncReset() error
	
	// 按日期获取
	GetByDate(date string) ([]*model.SiteStatistic, error)
}