package models

import (
	"gorm.io/gorm"
	
	"moviepilot-go/pkg/models"
)

// SiteStatistic 站点统计表模型
type SiteStatistic struct {
	models.SiteStatistic
}

// GetByDomain 按域名获取站点统计信息
func (s *SiteStatistic) GetByDomain(db *gorm.DB, domain string) (*models.SiteStatistic, error) {
	var siteStatistic models.SiteStatistic
	err := db.Where("domain = ?", domain).First(&siteStatistic).Error
	if err != nil {
		return nil, err
	}
	return &siteStatistic, nil
}

// Reset 清空站点统计表
func (s *SiteStatistic) Reset(db *gorm.DB) error {
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.SiteStatistic{}).Error
}