package models

import (
	"gorm.io/gorm"
	
	"moviepilot-go/pkg/models"
)

// SiteIcon 站点图标表模型
type SiteIcon struct {
	models.SiteIcon
}

// GetByDomain 按域名获取站点图标
func (s *SiteIcon) GetByDomain(db *gorm.DB, domain string) (*models.SiteIcon, error) {
	var siteIcon models.SiteIcon
	err := db.Where("domain = ?", domain).First(&siteIcon).Error
	if err != nil {
		return nil, err
	}
	return &siteIcon, nil
}