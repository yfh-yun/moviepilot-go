package repositories

import (
	"fmt"

	"github.com/yfh-yun/moviepilot-go/internal/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type siteStatisticRepository struct {
	db *gorm.DB
}

// NewSiteStatisticRepository 创建站点统计仓储
func NewSiteStatisticRepository() interfaces.SiteStatisticRepository {
	return &siteStatisticRepository{
		db: database.GetDB(),
	}
}

// Create 创建站点统计
func (r *siteStatisticRepository) Create(statistic *model.SiteStatistic) error {
	logger.Debug("Creating site statistic", 
		zap.String("site_name", statistic.SiteName))
	
	if err := r.db.Create(statistic).Error; err != nil {
		logger.Error("Failed to create site statistic", zap.Error(err))
		return fmt.Errorf("failed to create site statistic: %w", err)
	}
	
	logger.Info("Site statistic created successfully", zap.Uint("id", statistic.ID))
	return nil
}

// GetByID 根据ID获取站点统计
func (r *siteStatisticRepository) GetByID(id uint) (*model.SiteStatistic, error) {
	var statistic model.SiteStatistic
	if err := r.db.First(&statistic, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("site statistic with id %d not found", id)
		}
		logger.Error("Failed to get site statistic by ID", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get site statistic: %w", err)
	}
	
	return &statistic, nil
}

// Update 更新站点统计
func (r *siteStatisticRepository) Update(statistic *model.SiteStatistic) error {
	logger.Debug("Updating site statistic", zap.Uint("id", statistic.ID))
	
	if err := r.db.Save(statistic).Error; err != nil {
		logger.Error("Failed to update site statistic", zap.Uint("id", statistic.ID), zap.Error(err))
		return fmt.Errorf("failed to update site statistic: %w", err)
	}
	
	logger.Info("Site statistic updated successfully", zap.Uint("id", statistic.ID))
	return nil
}

// Delete 删除站点统计
func (r *siteStatisticRepository) Delete(id uint) error {
	logger.Debug("Deleting site statistic", zap.Uint("id", id))
	
	if err := r.db.Delete(&model.SiteStatistic{}, id).Error; err != nil {
		logger.Error("Failed to delete site statistic", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to delete site statistic: %w", err)
	}
	
	logger.Info("Site statistic deleted successfully", zap.Uint("id", id))
	return nil
}

// List 获取站点统计列表
func (r *siteStatisticRepository) List(offset, limit int) ([]*model.SiteStatistic, int64, error) {
	var statistics []*model.SiteStatistic
	var total int64
	
	if err := r.db.Model(&model.SiteStatistic{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count site statistics", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count site statistics: %w", err)
	}
	
	if err := r.db.Offset(offset).Limit(limit).Order("id DESC").Find(&statistics).Error; err != nil {
		logger.Error("Failed to list site statistics", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list site statistics: %w", err)
	}
	
	return statistics, total, nil
}

// GetByDomain 按域名获取站点统计
func (r *siteStatisticRepository) GetByDomain(domain string) (*model.SiteStatistic, error) {
	var statistic model.SiteStatistic
	if err := r.db.Where("site_name = ?", domain).First(&statistic).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("Failed to get site statistic by domain", zap.String("domain", domain), zap.Error(err))
		return nil, fmt.Errorf("failed to get site statistic by domain: %w", err)
	}
	
	return &statistic, nil
}

// AsyncGetByDomain 异步按域名获取站点统计（占位符实现）
func (r *siteStatisticRepository) AsyncGetByDomain(domain string) (*model.SiteStatistic, error) {
	return r.GetByDomain(domain)
}

// ListAll 获取所有站点统计
func (r *siteStatisticRepository) ListAll() ([]*model.SiteStatistic, error) {
	var statistics []*model.SiteStatistic
	
	if err := r.db.Order("id DESC").Find(&statistics).Error; err != nil {
		logger.Error("Failed to list all site statistics", zap.Error(err))
		return nil, fmt.Errorf("failed to list all site statistics: %w", err)
	}
	
	return statistics, nil
}

// UpdateSuccess 更新站点访问成功
func (r *siteStatisticRepository) UpdateSuccess(domain string, seconds *int) error {
	logger.Debug("Updating site statistic success", zap.String("domain", domain))
	
	statistic, err := r.GetByDomain(domain)
	if err != nil {
		return fmt.Errorf("failed to check existing site statistic: %w", err)
	}
	
	if statistic != nil {
		// 更新现有记录
		statistic.Success += 1
		if seconds != nil {
			statistic.Seconds = *seconds
		}
		statistic.LstState = 0
		statistic.LstModDate = "当前时间" // 应该使用实际时间
		
		if err := r.Update(statistic); err != nil {
			return fmt.Errorf("failed to update site statistic: %w", err)
		}
	} else {
		// 创建新记录
		newStatistic := &model.SiteStatistic{
			SiteName: domain,
			Success:  1,
			Fail: 0,
			LstState: 0,
			LstModDate: "当前时间",
			Note: "",
		}
		
		if seconds != nil {
			newStatistic.Seconds = *seconds
		} else {
			newStatistic.Seconds = 1
		}
		
		if err := r.Create(newStatistic); err != nil {
			return fmt.Errorf("failed to create site statistic: %w", err)
		}
	}
	
	logger.Info("Site statistic success updated successfully", zap.String("domain", domain))
	return nil
}

// UpdateFail 更新站点访问失败
func (r *siteStatisticRepository) UpdateFail(domain string) error {
	logger.Debug("Updating site statistic fail", zap.String("domain", domain))
	
	statistic, err := r.GetByDomain(domain)
	if err != nil {
		return fmt.Errorf("failed to check existing site statistic: %w", err)
	}
	
	if statistic != nil {
		// 更新现有记录
		statistic.Fail += 1
		statistic.LstState = 1
		statistic.LstModDate = "当前时间"
		
		if err := r.Update(statistic); err != nil {
			return fmt.Errorf("failed to update site statistic: %w", err)
		}
	} else {
		// 创建新记录
		newStatistic := &model.SiteStatistic{
			SiteName: domain,
			Success: 0,
			Fail:    1,
			LstState: 1,
			LstModDate: "当前时间",
		}
		
		if err := r.Create(newStatistic); err != nil {
			return fmt.Errorf("failed to create site statistic: %w", err)
		}
	}
	
	logger.Info("Site statistic fail updated successfully", zap.String("domain", domain))
	return nil
}

// AsyncUpdateSuccess 异步更新站点访问成功（占位符实现）
func (r *siteStatisticRepository) AsyncUpdateSuccess(domain string, seconds *int) error {
	return r.UpdateSuccess(domain, seconds)
}

// AsyncUpdateFail 异步更新站点访问失败（占位符实现）
func (r *siteStatisticRepository) AsyncUpdateFail(domain string) error {
	return r.UpdateFail(domain)
}

// Reset 重置站点统计
func (r *siteStatisticRepository) Reset() error {
	logger.Info("Resetting site statistics")
	
	if err := r.db.Exec("DELETE FROM sitestatistics").Error; err != nil {
		logger.Error("Failed to reset site statistics", zap.Error(err))
		return fmt.Errorf("failed to reset site statistics: %w", err)
	}
	
	logger.Info("Site statistics reset successfully")
	return nil
}

// AsyncReset 异步重置站点统计（占位符实现）
func (r *siteStatisticRepository) AsyncReset() error {
	return r.Reset()
}

// GetByDate 按日期获取站点统计
func (r *siteStatisticRepository) GetByDate(date string) ([]*model.SiteStatistic, error) {
	var statistics []*model.SiteStatistic
	
	if err := r.db.Where("lst_mod_date LIKE ?", date+"%").Order("id DESC").Find(&statistics).Error; err != nil {
		logger.Error("Failed to get site statistics by date", 
			zap.String("date", date), zap.Error(err))
		return nil, fmt.Errorf("failed to get site statistics by date: %w", err)
	}
	
	return statistics, nil
}
