package repositories

import (
	"errors"
	"fmt"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/model"

	"gorm.io/gorm"
)

// searchRepository 搜索仓储实现
type searchRepository struct {
	db *gorm.DB
}

// NewSearchRepository 创建搜索仓储
func NewSearchRepository(db *gorm.DB) interfaces.SearchRepository {
	return &searchRepository{db: db}
}

// Create 创建搜索记录
func (r *searchRepository) Create(search *model.Search) error {
	if search == nil {
		return errors.New("search cannot be nil")
	}
	return r.db.Create(search).Error
}

// GetByID 根据ID获取搜索记录
func (r *searchRepository) GetByID(id uint) (*model.Search, error) {
	var search model.Search
	err := r.db.First(&search, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &search, nil
}

// GetByKeyword 根据关键词获取搜索记录
func (r *searchRepository) GetByKeyword(keyword string) ([]*model.Search, error) {
	var searches []*model.Search
	err := r.db.Where("keyword = ?", keyword).Order("created_at DESC").Find(&searches).Error
	return searches, err
}

// GetByUserID 根据用户ID获取搜索记录
func (r *searchRepository) GetByUserID(userID uint) ([]*model.Search, error) {
	var searches []*model.Search
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&searches).Error
	return searches, err
}

// GetByType 根据类型获取搜索记录
func (r *searchRepository) GetByType(searchType string) ([]*model.Search, error) {
	var searches []*model.Search
	err := r.db.Where("type = ?", searchType).Order("created_at DESC").Find(&searches).Error
	return searches, err
}

// GetByDateRange 根据日期范围获取搜索记录
func (r *searchRepository) GetByDateRange(startDate, endDate interface{}) ([]*model.Search, error) {
	var searches []*model.Search
	err := r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").Find(&searches).Error
	return searches, err
}

// Update 更新搜索记录
func (r *searchRepository) Update(search *model.Search) error {
	if search == nil {
		return errors.New("search cannot be nil")
	}
	return r.db.Save(search).Error
}

// Delete 删除搜索记录
func (r *searchRepository) Delete(id uint) error {
	return r.db.Delete(&model.Search{}, id).Error
}

// DeleteByUserID 根据用户ID删除搜索记录
func (r *searchRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.Search{}).Error
}

// DeleteByDateRange 根据日期范围删除搜索记录
func (r *searchRepository) DeleteByDateRange(startDate, endDate interface{}) error {
	return r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).Delete(&model.Search{}).Error
}

// List 分页获取搜索记录列表
func (r *searchRepository) List(offset, limit int) ([]*model.Search, int64, error) {
	var searches []*model.Search
	var total int64

	err := r.db.Model(&model.Search{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&searches).Error
	return searches, total, err
}

// Count 统计搜索记录数量
func (r *searchRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Search{}).Count(&count).Error
	return count, err
}

// Search 搜索搜索记录（支持多条件）
func (r *searchRepository) Search(keyword, searchType string, userID *uint, offset, limit int) ([]*model.Search, int64, error) {
	var searches []*model.Search
	var total int64

	query := r.db.Model(&model.Search{})

	// 添加搜索条件
	if keyword != "" {
		query = query.Where("keyword LIKE ?", "%"+keyword+"%")
	}
	if searchType != "" {
		query = query.Where("type = ?", searchType)
	}
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&searches).Error; err != nil {
		return nil, 0, err
	}

	return searches, total, err
}

// GetPopularKeywords 获取热门关键词
func (r *searchRepository) GetPopularKeywords(limit int) ([]*model.PopularKeyword, error) {
	var results []*model.PopularKeyword
	
	err := r.db.Model(&model.Search{}).
		Select("keyword, COUNT(*) as count").
		Group("keyword").
		Order("count DESC").
		Limit(limit).
		Find(&results).Error
	
	return results, err
}

// GetSearchStatistics 获取搜索统计信息
func (r *searchRepository) GetSearchStatistics() (*model.SearchStatistics, error) {
	var stats model.SearchStatistics
	
	// 总搜索次数
	if err := r.db.Model(&model.Search{}).Count(&stats.TotalSearches).Error; err != nil {
		return nil, err
	}
	
	// 今日搜索次数
	if err := r.db.Model(&model.Search{}).
		Where("DATE(created_at) = DATE('now')").
		Count(&stats.TodaySearches).Error; err != nil {
		return nil, err
	}
	
	// 本周搜索次数
	if err := r.db.Model(&model.Search{}).
		Where("created_at >= DATE('now', '-7 days')").
		Count(&stats.WeeklySearches).Error; err != nil {
		return nil, err
	}
	
	// 本月搜索次数
	if err := r.db.Model(&model.Search{}).
		Where("created_at >= DATE('now', '-30 days')").
		Count(&stats.MonthlySearches).Error; err != nil {
		return nil, err
	}
	
	return &stats, nil
}

// ClearOldSearches 清理旧的搜索记录
func (r *searchRepository) ClearOldSearches(days int) error {
	return r.db.Where("created_at < DATE('now', '-%d days')", days).Delete(&model.Search{}).Error
}