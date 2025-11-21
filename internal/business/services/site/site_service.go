package site

import (
	"encoding/json"
	"fmt"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/models"
	"time"

	"gorm.io/gorm"
)

// SiteService 站点服务
type SiteService struct {
	siteRepo          interfaces.SiteRepository
	siteUserDataRepo  interfaces.SiteUserDataRepository
	siteStatisticRepo interfaces.SiteStatisticRepository
	siteIconRepo      interfaces.SiteIconRepository
}

// NewSiteService 创建站点服务实例
func NewSiteService(
	siteRepo interfaces.SiteRepository,
	siteUserDataRepo interfaces.SiteUserDataRepository,
	siteStatisticRepo interfaces.SiteStatisticRepository,
	siteIconRepo interfaces.SiteIconRepository,
) *SiteService {
	return &SiteService{
		siteRepo:          siteRepo,
		siteUserDataRepo:  siteUserDataRepo,
		siteStatisticRepo: siteStatisticRepo,
		siteIconRepo:      siteIconRepo,
	}
}

// CreateSite 创建站点
func (s *SiteService) CreateSite(site *models.Site) error {
	// 检查站点名称是否已存在
	existingSite, err := s.siteRepo.GetByName(site.Name)
	if err == nil && existingSite != nil {
		return fmt.Errorf("站点名称已存在: %s", site.Name)
	}

	// 检查域名是否已存在
	if site.Domain != "" {
		existingSite, err = s.siteRepo.GetByDomain(site.Domain)
		if err == nil && existingSite != nil {
			return fmt.Errorf("站点域名已存在: %s", site.Domain)
		}
	}

	// 设置默认值
	if site.Settings == "" {
		site.Settings = "{}"
	}

	return s.siteRepo.Create(site)
}

// UpdateSite 更新站点
func (s *SiteService) UpdateSite(site *models.Site) error {
	// 检查站点是否存在
	existingSite, err := s.siteRepo.GetByID(site.ID)
	if err != nil {
		return fmt.Errorf("站点不存在: %d", site.ID)
	}

	// 检查新的站点名称是否已被其他站点使用
	if existingSite.Name != site.Name {
		siteWithSameName, err := s.siteRepo.GetByName(site.Name)
		if err == nil && siteWithSameName != nil && siteWithSameName.ID != site.ID {
			return fmt.Errorf("站点名称已存在: %s", site.Name)
		}
	}

	// 检查新的域名是否已被其他站点使用
	if existingSite.Domain != site.Domain && site.Domain != "" {
		siteWithSameDomain, err := s.siteRepo.GetByDomain(site.Domain)
		if err == nil && siteWithSameDomain != nil && siteWithSameDomain.ID != site.ID {
			return fmt.Errorf("站点域名已存在: %s", site.Domain)
		}
	}

	return s.siteRepo.Update(site)
}

// DeleteSite 删除站点
func (s *SiteService) DeleteSite(siteID uint) error {
	// 检查站点是否存在
	site, err := s.siteRepo.GetByID(siteID)
	if err != nil {
		return fmt.Errorf("站点不存在: %d", siteID)
	}

	// 删除相关的站点数据
	if err := s.siteUserDataRepo.DeleteBySiteName(site.Name); err != nil {
		return fmt.Errorf("删除站点用户数据失败: %v", err)
	}

	if err := s.siteStatisticRepo.DeleteBySiteName(site.Name); err != nil {
		return fmt.Errorf("删除站点统计信息失败: %v", err)
	}

	if err := s.siteIconRepo.DeleteBySiteName(site.Name); err != nil {
		return fmt.Errorf("删除站点图标失败: %v", err)
	}

	// 删除站点
	return s.siteRepo.Delete(siteID)
}

// GetSiteByID 根据ID获取站点
func (s *SiteService) GetSiteByID(siteID uint) (*models.Site, error) {
	site, err := s.siteRepo.GetByID(siteID)
	if err != nil {
		return nil, fmt.Errorf("获取站点失败: %v", err)
	}
	return site, nil
}

// GetSiteByName 根据名称获取站点
func (s *SiteService) GetSiteByName(name string) (*models.Site, error) {
	site, err := s.siteRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("获取站点失败: %v", err)
	}
	return site, nil
}

// ListSites 获取站点列表
func (s *SiteService) ListSites(offset, limit int) ([]*models.Site, int64, error) {
	sites, total, err := s.siteRepo.List(offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("获取站点列表失败: %v", err)
	}
	return sites, total, nil
}

// SearchSites 搜索站点
func (s *SiteService) SearchSites(keyword string, offset, limit int) ([]*models.Site, int64, error) {
	sites, total, err := s.siteRepo.Search(keyword, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索站点失败: %v", err)
	}
	return sites, total, nil
}

// GetActiveSites 获取活跃站点
func (s *SiteService) GetActiveSites() ([]*models.Site, error) {
	sites, err := s.siteRepo.GetActive()
	if err != nil {
		return nil, fmt.Errorf("获取活跃站点失败: %v", err)
	}
	return sites, nil
}

// GetRSSSites 获取RSS站点
func (s *SiteService) GetRSSSites() ([]*models.Site, error) {
	sites, err := s.siteRepo.GetRSSSites()
	if err != nil {
		return nil, fmt.Errorf("获取RSS站点失败: %v", err)
	}
	return sites, nil
}

// GetSearchSites 获取搜索站点
func (s *SiteService) GetSearchSites() ([]*models.Site, error) {
	sites, err := s.siteRepo.GetSearchSites()
	if err != nil {
		return nil, fmt.Errorf("获取搜索站点失败: %v", err)
	}
	return sites, nil
}

// UpdateSiteCookie 更新站点Cookie
func (s *SiteService) UpdateSiteCookie(siteID uint, cookie string) error {
	site, err := s.siteRepo.GetByID(siteID)
	if err != nil {
		return fmt.Errorf("站点不存在: %d", siteID)
	}

	site.Cookie = cookie
	site.UpdatedAt = time.Now()

	return s.siteRepo.Update(site)
}

// UpdateSiteSettings 更新站点设置
func (s *SiteService) UpdateSiteSettings(siteID uint, settings map[string]interface{}) error {
	site, err := s.siteRepo.GetByID(siteID)
	if err != nil {
		return fmt.Errorf("站点不存在: %d", siteID)
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("序列化设置失败: %v", err)
	}

	site.Settings = string(settingsJSON)
	site.UpdatedAt = time.Now()

	return s.siteRepo.Update(site)
}

// ToggleSiteActive 切换站点激活状态
func (s *SiteService) ToggleSiteActive(siteID uint) (*models.Site, error) {
	site, err := s.siteRepo.GetByID(siteID)
	if err != nil {
		return nil, fmt.Errorf("站点不存在: %d", siteID)
	}

	site.IsActive = !site.IsActive
	site.UpdatedAt = time.Now()

	err = s.siteRepo.Update(site)
	if err != nil {
		return nil, fmt.Errorf("更新站点状态失败: %v", err)
	}

	return site, nil
}

// GetSiteStatistics 获取站点统计信息
func (s *SiteService) GetSiteStatistics(siteName string) ([]*models.SiteStatistic, error) {
	statistics, err := s.siteStatisticRepo.GetBySiteName(siteName)
	if err != nil {
		return nil, fmt.Errorf("获取站点统计信息失败: %v", err)
	}
	return statistics, nil
}

// GetSiteUserData 获取站点用户数据
func (s *SiteService) GetSiteUserData(siteName string) ([]*models.SiteUserData, error) {
	userData, err := s.siteUserDataRepo.GetBySiteName(siteName)
	if err != nil {
		return nil, fmt.Errorf("获取站点用户数据失败: %v", err)
	}
	return userData, nil
}

// UpdateSiteUserData 更新站点用户数据
func (s *SiteService) UpdateSiteUserData(siteName, username string, data *models.SiteUserData) error {
	existingData, err := s.siteUserDataRepo.GetBySiteNameAndUsername(siteName, username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("查询用户数据失败: %v", err)
	}

	if existingData == nil {
		// 创建新记录
		data.SiteName = siteName
		data.Username = username
		return s.siteUserDataRepo.Create(data)
	}

	// 更新现有记录
	existingData.Uploaded = data.Uploaded
	existingData.Downloaded = data.Downloaded
	existingData.Ratio = data.Ratio
	existingData.Seeding = data.Seeding
	existingData.Leeching = data.Leeching
	existingData.Bonus = data.Bonus
	existingData.Invites = data.Invites
	existingData.UserLevel = data.UserLevel
	existingData.UpdatedAt = time.Now()

	return s.siteUserDataRepo.Update(existingData)
}

// GetSiteIcon 获取站点图标
func (s *SiteService) GetSiteIcon(siteName string) (*models.SiteIcon, error) {
	siteIcon, err := s.siteIconRepo.GetBySiteName(siteName)
	if err != nil {
		return nil, fmt.Errorf("获取站点图标失败: %v", err)
	}
	return siteIcon, nil
}

// UpdateSiteIcon 更新站点图标
func (s *SiteService) UpdateSiteIcon(siteName string, icon []byte, url string) error {
	existingIcon, err := s.siteIconRepo.GetBySiteName(siteName)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("查询站点图标失败: %v", err)
	}

	if existingIcon == nil {
		// 创建新记录
		siteIcon := &models.SiteIcon{
			SiteName:  siteName,
			Icon:      string(icon),
			URL:       url,
			UpdatedAt: &time.Time{},
		}
		return s.siteIconRepo.Create(siteIcon)
	}

	// 更新现有记录
	existingIcon.Icon = string(icon)
	existingIcon.URL = url
	existingIcon.UpdatedAt = &time.Time{}

	return s.siteIconRepo.Update(existingIcon)
}

// ImportSiteData 导入站点数据
func (s *SiteService) ImportSiteData(sites []*models.Site) error {
	tx := repository.DB.Begin()

	for _, site := range sites {
		// 检查站点是否存在
		existingSite, err := s.siteRepo.GetByName(site.Name)
		if err != nil && err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return fmt.Errorf("检查站点失败: %v", err)
		}

		if existingSite != nil {
			// 更新现有站点
			existingSite.Domain = site.Domain
			existingSite.URL = site.URL
			existingSite.SignURL = site.SignURL
			existingSite.LoginPage = site.LoginPage
			existingSite.UserAgent = site.UserAgent
			existingSite.Proxy = site.Proxy
			existingSite.Render = site.Render
			existingSite.Public = site.Public
			existingSite.IsRSS = site.IsRSS
			existingSite.RSS = site.RSS
			existingSite.Subscribed = site.Subscribed
			existingSite.FailLimit = site.FailLimit
			existingSite.FailCount = site.FailCount
			existingSite.IsActive = site.IsActive
			existingSite.Priority = site.Priority
			existingSite.Settings = site.Settings

			if err := s.siteRepo.Update(existingSite); err != nil {
				tx.Rollback()
				return fmt.Errorf("更新站点失败: %v", err)
			}
		} else {
			// 创建新站点
			if err := s.siteRepo.Create(site); err != nil {
				tx.Rollback()
				return fmt.Errorf("创建站点失败: %v", err)
			}
		}
	}

	return tx.Commit().Error
}
