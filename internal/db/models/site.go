package models

import (
	"gorm.io/gorm"
	
	"moviepilot-go/pkg/models"
)

// Site 站点表模型
type Site struct {
	models.Site
}

// GetByDomain 按域名获取站点
func (s *Site) GetByDomain(db *gorm.DB, domain string) (*models.Site, error) {
	var site models.Site
	err := db.Where("domain = ?", domain).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

// GetActives 获取启用的站点列表
func (s *Site) GetActives(db *gorm.DB) ([]models.Site, error) {
	var sites []models.Site
	err := db.Where("is_active = ?", true).Find(&sites).Error
	return sites, err
}

// ListOrderByPri 按优先级获取站点列表
func (s *Site) ListOrderByPri(db *gorm.DB) ([]models.Site, error) {
	var sites []models.Site
	err := db.Order("pri").Find(&sites).Error
	return sites, err
}

// GetDomainsByIds 按ID获取站点域名
func (s *Site) GetDomainsByIds(db *gorm.DB, ids []int) ([]string, error) {
	var domains []string
	err := db.Model(&models.Site{}).Where("id IN ?", ids).Pluck("domain", &domains).Error
	return domains, err
}

// Reset 清空站点表
func (s *Site) Reset(db *gorm.DB) error {
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Site{}).Error
}