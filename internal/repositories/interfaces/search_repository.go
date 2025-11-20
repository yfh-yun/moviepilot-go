package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// SearchRepository 搜索仓储接口
type SearchRepository interface {
	// Create 创建搜索记录
	Create(search *model.Search) error

	// GetByID 根据ID获取搜索记录
	GetByID(id uint) (*model.Search, error)

	// GetByKeyword 根据关键词获取搜索记录
	GetByKeyword(keyword string) ([]*model.Search, error)

	// GetByUserID 根据用户ID获取搜索记录
	GetByUserID(userID uint) ([]*model.Search, error)

	// GetByType 根据类型获取搜索记录
	GetByType(searchType string) ([]*model.Search, error)

	// GetByDateRange 根据日期范围获取搜索记录
	GetByDateRange(startDate, endDate interface{}) ([]*model.Search, error)

	// Update 更新搜索记录
	Update(search *model.Search) error

	// Delete 删除搜索记录
	Delete(id uint) error

	// DeleteByUserID 根据用户ID删除搜索记录
	DeleteByUserID(userID uint) error

	// DeleteByDateRange 根据日期范围删除搜索记录
	DeleteByDateRange(startDate, endDate interface{}) error

	// List 分页获取搜索记录列表
	List(offset, limit int) ([]*model.Search, int64, error)

	// Count 统计搜索记录数量
	Count() (int64, error)

	// Search 搜索搜索记录（支持多条件）
	Search(keyword, searchType string, userID *uint, offset, limit int) ([]*model.Search, int64, error)

	// GetPopularKeywords 获取热门关键词
	GetPopularKeywords(limit int) ([]*model.PopularKeyword, error)

	// GetSearchStatistics 获取搜索统计信息
	GetSearchStatistics() (*model.SearchStatistics, error)

	// ClearOldSearches 清理旧的搜索记录
	ClearOldSearches(days int) error
}
