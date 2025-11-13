package models

import (
	"gorm.io/gorm"
	
	"moviepilot-go/pkg/models"
)

// SiteUserData 站点用户数据表模型
type SiteUserData struct {
	models.SiteUserData
}

// GetByDomain 按域名获取站点用户数据
func (s *SiteUserData) GetByDomain(db *gorm.DB, domain string, workdate *string, worktime *string) ([]models.SiteUserData, error) {
	var siteUserDataList []models.SiteUserData
	query := db.Where("domain = ?", domain)
	
	if workdate != nil && worktime != nil {
		query = query.Where("updated_day = ? AND updated_time = ?", *workdate, *worktime)
	} else if workdate != nil {
		query = query.Where("updated_day = ?", *workdate)
	}
	
	err := query.Find(&siteUserDataList).Error
	return siteUserDataList, err
}

// GetByDate 按日期获取站点用户数据
func (s *SiteUserData) GetByDate(db *gorm.DB, date string) ([]models.SiteUserData, error) {
	var siteUserDataList []models.SiteUserData
	err := db.Where("updated_day = ?", date).Find(&siteUserDataList).Error
	return siteUserDataList, err
}

// GetLatest 获取各站点最新一天的数据
func (s *SiteUserData) GetLatest(db *gorm.DB) ([]models.SiteUserData, error) {
	var siteUserDataList []models.SiteUserData
	
	// 先获取各站点最新更新日期
	type latestUpdate struct {
		Domain          string
		LatestUpdateDay string
	}
	
	var latestUpdates []latestUpdate
	err := db.Model(&models.SiteUserData{}).
		Select("domain, MAX(updated_day) as latest_update_day").
		Group("domain").
		Where("(err_msg IS NULL OR err_msg = '')").
		Scan(&latestUpdates).Error
	
	if err != nil {
		return nil, err
	}
	
	// 构建查询条件
	if len(latestUpdates) > 0 {
		// 创建查询条件
		query := db.Where("(")
		for i, lu := range latestUpdates {
			if i > 0 {
				query = query.Or("(")
			} else {
				query = query.Where("(")
			}
			query = query.Where("domain = ? AND updated_day = ?", lu.Domain, lu.LatestUpdateDay)
			query = query.Where(")")
		}
		query = query.Where(")")
		
		// 按更新时间倒序排列
		err = query.Order("updated_time DESC").Find(&siteUserDataList).Error
	}
	
	return siteUserDataList, err
}