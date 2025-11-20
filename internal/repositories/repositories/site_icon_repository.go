package repositories

import (
	"fmt"

	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type siteIconRepository struct {
	db *gorm.DB
}

// NewSiteIconRepository 创建站点图标仓储
func NewSiteIconRepository() interfaces.SiteIconRepository {
	return &siteIconRepository{
		db: database.GetDB(),
	}
}

// Create 创建站点图标
func (r *siteIconRepository) Create(icon *model.SiteIcon) error {
	logger.Debug("Creating site icon", 
		zap.String("site_name", icon.SiteName),
		zap.String("url", icon.URL))
	
	if err := r.db.Create(icon).Error; err != nil {
		logger.Error("Failed to create site icon", zap.Error(err))
		return fmt.Errorf("failed to create site icon: %w", err)
	}
	
	logger.Info("Site icon created successfully", zap.Uint("id", icon.ID))
	return nil
}

// GetByID 根据ID获取站点图标
func (r *siteIconRepository) GetByID(id uint) (*model.SiteIcon, error) {
	var icon model.SiteIcon
	if err := r.db.First(&icon, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("site icon with id %d not found", id)
		}
		logger.Error("Failed to get site icon by ID", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get site icon: %w", err)
	}
	
	return &icon, nil
}

// Update 更新站点图标
func (r *siteIconRepository) Update(icon *model.SiteIcon) error {
	logger.Debug("Updating site icon", zap.Uint("id", icon.ID))
	
	if err := r.db.Save(icon).Error; err != nil {
		logger.Error("Failed to update site icon", zap.Uint("id", icon.ID), zap.Error(err))
		return fmt.Errorf("failed to update site icon: %w", err)
	}
	
	logger.Info("Site icon updated successfully", zap.Uint("id", icon.ID))
	return nil
}

// Delete 删除站点图标
func (r *siteIconRepository) Delete(id uint) error {
	logger.Debug("Deleting site icon", zap.Uint("id", id))
	
	if err := r.db.Delete(&model.SiteIcon{}, id).Error; err != nil {
		logger.Error("Failed to delete site icon", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete site icon: %w", err)
	}
	
	logger.Info("Site icon deleted successfully", zap.Uint("id", id))
	return nil
}

// List 获取站点图标列表
func (r *siteIconRepository) List(offset, limit int) ([]*model.SiteIcon, int64, error) {
	var icons []*model.SiteIcon
	var total int64
	
	// 获取总数
	if err := r.db.Model(&model.SiteIcon{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count site icons", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count site icons: %w", err)
	}
	
	// 获取分页数据
	if err := r.db.Offset(offset).Limit(limit).Order("id DESC").Find(&icons).Error; err != nil {
		logger.Error("Failed to list site icons", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list site icons: %w", err)
	}
	
	return icons, total, nil
}

// GetByDomain 按域名获取站点图标
func (r *siteIconRepository) GetByDomain(domain string) (*model.SiteIcon, error) {
	var icon model.SiteIcon
	if err := r.db.Where("site_name = ?", domain).First(&icon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("Failed to get site icon by domain", zap.String("domain", domain), zap.Error(err))
		return nil, fmt.Errorf("failed to get site icon by domain: %w", err)
	}
	
	return &icon, nil
}

// AsyncGetByDomain 异步按域名获取站点图标（占位符实现）
func (r *siteIconRepository) AsyncGetByDomain(domain string) (*model.SiteIcon, error) {
	// Go中可以通过context实现异步，这里简化处理
	return r.GetByDomain(domain)
}

// UpdateIcon 更新图标
func (r *siteIconRepository) UpdateIcon(name, domain, iconURL, iconBase64 string) error {
	logger.Debug("Updating site icon", 
		zap.String("name", name),
		zap.String("domain", domain),
		zap.String("url", iconURL))
	
	// 查询现有图标
	existingIcon, err := r.GetByDomain(domain)
	if err != nil {
		return fmt.Errorf("failed to check existing site icon: %w", err)
	}
	
	iconBase64Full := ""
	if iconBase64 != "" {
		iconBase64Full = "data:image/ico;base64," + iconBase64
	}
	
	if existingIcon != nil {
		// 更新现有图标
		existingIcon.Icon = iconBase64Full
		existingIcon.URL = iconURL
		
		if err := r.Update(existingIcon); err != nil {
			return fmt.Errorf("failed to update site icon: %w", err)
		}
	} else {
		// 创建新图标
		newIcon := &model.SiteIcon{
			SiteName: name,
			Icon:     iconBase64Full,
			URL:      iconURL,
		}
		
		if err := r.Create(newIcon); err != nil {
			return fmt.Errorf("failed to create site icon: %w", err)
		}
	}
	
	logger.Info("Site icon updated successfully", zap.String("domain", domain))
	return nil
}

// GetBySiteName 按站点名称获取图标
func (r *siteIconRepository) GetBySiteName(siteName string) (*model.SiteIcon, error) {
	var icon model.SiteIcon
	if err := r.db.Where("site_name = ?", siteName).First(&icon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("Failed to get site icon by site name", zap.String("site_name", siteName), zap.Error(err))
		return nil, fmt.Errorf("failed to get site icon by site name: %w", err)
	}
	
	return &icon, nil
}

// Empty 清空图标
func (r *siteIconRepository) Empty() error {
	logger.Info("Emptying site icons")
	
	if err := r.db.Exec("DELETE FROM siteicons").Error; err != nil {
		logger.Error("Failed to empty site icons", zap.Error(err))
		return fmt.Errorf("failed to empty site icons: %w", err)
	}
	
	logger.Info("Site icons emptied successfully")
	return nil
}
