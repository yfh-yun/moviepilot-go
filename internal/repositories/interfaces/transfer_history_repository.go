package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"time"
)

// TransferHistoryRepository 转移历史仓储接口
type TransferHistoryRepository interface {
	// Create 创建转移历史
	Create(transferHistory *model.TransferHistory) error
	
	// BatchCreate 批量创建转移历史
	BatchCreate(histories []*model.TransferHistory) error
	
	// GetByID 根据ID获取转移历史
	GetByID(id uint) (*model.TransferHistory, error)
	
	// GetByDownloadHash 根据下载Hash获取转移历史
	GetByDownloadHash(downloadHash string) ([]*model.TransferHistory, error)
	
	// GetByDateRange 根据日期范围获取转移历史
	GetByDateRange(startDate, endDate time.Time) ([]*model.TransferHistory, error)
	
	// GetByUsername 根据用户名获取转移历史
	GetByUsername(username string) ([]*model.TransferHistory, error)
	
	// GetByStatus 根据状态获取转移历史
	GetByStatus(status string) ([]*model.TransferHistory, error)
	
	// GetFailed 获取失败的转移历史
	GetFailed() ([]*model.TransferHistory, error)
	
	// GetByTarget 根据目标获取转移历史
	GetByTarget(target string) ([]*model.TransferHistory, error)
	
	// Update 更新转移历史
	Update(transferHistory *model.TransferHistory) error
	
	// UpdateStatus 更新转移状态
	UpdateStatus(id uint, status string, failReason string) error
	
	// Delete 删除转移历史
	Delete(id uint) error
	
	// DeleteByDownloadHash 根据下载Hash删除转移历史
	DeleteByDownloadHash(downloadHash string) error
	
	// List 分页获取转移历史列表
	List(offset, limit int) ([]*model.TransferHistory, int64, error)
	
	// Search 搜索转移历史
	Search(keyword string, offset, limit int) ([]*model.TransferHistory, int64, error)
	
	// Count 统计转移历史数量
	Count() (int64, error)
	
	// CountByStatus 根据状态统计数量
	CountByStatus(status string) (int64, error)
	
	// GetByType 根据类型获取转移历史
	GetByType(transferType string) ([]*model.TransferHistory, error)
}