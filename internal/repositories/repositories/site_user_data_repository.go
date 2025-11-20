package repositories

import (
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"

	"gorm.io/gorm"
)

// SiteUserDataRepository 站点用户数据仓储实现
type SiteUserDataRepository struct {
	db *gorm.DB
}

// NewSiteUserDataRepository 创建站点用户数据仓储实例
func NewSiteUserDataRepository(db *gorm.DB) interfaces.SiteUserDataRepository {
	return &model.SiteUserDataRepository{db: db}
}

// Create 创建站点用户数据
func (r *SiteUserDataRepository) Create(siteUserData *model.SiteUserData) error {
	if siteUserData == nil {
		return errors.New("site user data cannot be nil")
	}

	// 检查站点名称和用户名是否已存在
	var existingData model.SiteUserData
	err := r.db.Where("site_name = ? AND username = ?",
		siteUserData.SiteName, siteUserData.Username).First(&existingData).Error
	if err == nil {
		return fmt.Errorf("site user data for site '%s' and user '%s' already exists",
			siteUserData.SiteName, siteUserData.Username)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check site user data: %w", err)
	}

	return r.db.Create(siteUserData).Error
}

// GetByID 根据ID获取站点用户数据
func (r *SiteUserDataRepository) GetByID(id uint) (*model.SiteUserData, error) {
	var siteUserData model.SiteUserData
	err := r.db.First(&siteUserData, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get site user data by id %d: %w", id, err)
	}
	return &siteUserData, nil
}

// GetBySiteName 根据站点名称获取站点用户数据列表
func (r *SiteUserDataRepository) GetBySiteName(siteName string) ([]*model.SiteUserData, error) {
	var siteUserDataList []*model.SiteUserData
	err := r.db.Where("site_name = ?", siteName).Find(&siteUserDataList).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get site user data by site name '%s': %w", siteName, err)
	}
	return siteUserDataList, nil
}

// GetBySiteNameAndUsername 根据站点名称和用户名获取站点用户数据
func (r *SiteUserDataRepository) GetBySiteNameAndUsername(siteName, username string) (*model.SiteUserData, error) {
	var siteUserData model.SiteUserData
	err := r.db.Where("site_name = ? AND username = ?", siteName, username).First(&siteUserData).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get site user data by site name '%s' and username '%s': %w",
			siteName, username, err)
	}
	return &siteUserData, nil
}

// Update 更新站点用户数据
func (r *SiteUserDataRepository) Update(siteUserData *model.SiteUserData) error {
	if siteUserData == nil {
		return errors.New("site user data cannot be nil")
	}

	// 检查站点名称和用户名是否与其他记录冲突
	var existingData model.SiteUserData
	err := r.db.Where("site_name = ? AND username = ? AND id != ?",
		siteUserData.SiteName, siteUserData.Username, siteUserData.ID).First(&existingData).Error
	if err == nil {
		return fmt.Errorf("site user data for site '%s' and user '%s' already exists",
			siteUserData.SiteName, siteUserData.Username)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check site user data: %w", err)
	}

	return r.db.Save(siteUserData).Error
}

// Delete 删除站点用户数据
func (r *SiteUserDataRepository) Delete(id uint) error {
	result := r.db.Delete(&model.SiteUserData{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete site user data %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("site user data with id %d not found", id)
	}
	return nil
}

// DeleteBySiteName 根据站点名称删除站点用户数据
func (r *SiteUserDataRepository) DeleteBySiteName(siteName string) error {
	result := r.db.Where("site_name = ?", siteName).Delete(&model.SiteUserData{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete site user data by site name '%s': %w", siteName, result.Error)
	}
	return nil
}

// List 分页获取站点用户数据列表
func (r *SiteUserDataRepository) List(offset, limit int) ([]*model.SiteUserData, int64, error) {
	var siteUserDataList []*model.SiteUserData
	var total int64

	// 获取总数
	if err := r.db.Model(&model.SiteUserData{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count site user data: %w", err)
	}

	// 获取分页数据
	if err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&siteUserDataList).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list site user data: %w", err)
	}

	return siteUserDataList, total, nil
}

// Count 统计站点用户数据数量
func (r *SiteUserDataRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.SiteUserData{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count site user data: %w", err)
	}
	return count, nil
}
