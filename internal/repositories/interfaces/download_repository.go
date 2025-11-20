package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// DownloadRepository 下载历史仓储接口
type DownloadRepository interface {
	// Create 创建下载历史记录
	Create(history *model.DownloadHistory) error

	// Update 更新下载历史记录
	Update(history *model.DownloadHistory) error

	// Delete 删除下载历史记录
	Delete(id uint) error

	// GetByID 根据ID获取下载历史
	GetByID(id uint) (*model.DownloadHistory, error)

	// GetByHash 根据下载Hash获取下载历史
	GetByHash(downloadHash string) (*model.DownloadHistory, error)

	// GetByPath 根据路径获取下载历史
	GetByPath(path string) (*model.DownloadHistory, error)

	// GetByMediaID 根据媒体ID获取下载历史
	GetByMediaID(tmdbID *int, doubanID *string) ([]*model.DownloadHistory, error)

	// GetLast 获取最后的下载记录
	GetLast(mtype, title, year, season, episode *string, tmdbID *int) ([]*model.DownloadHistory, error)

	// ListByPage 分页获取下载历史
	ListByPage(page, count int) ([]*model.DownloadHistory, int64, error)

	// ListByUserDate 根据用户和时间获取下载历史
	ListByUserDate(date string, username string) ([]*model.DownloadHistory, error)

	// ListByDate 根据日期获取下载历史
	ListByDate(date string, mtype string, tmdbID string, seasons *string) ([]*model.DownloadHistory, error)

	// ListByType 根据类型获取下载历史
	ListByType(mtype string, days int) ([]*model.DownloadHistory, error)

	// Count 统计下载历史数量
	Count() (int64, error)

	// CountByUser 统计用户下载数量
	CountByUser(username string) (int64, error)
}

// DownloadFilesRepository 下载文件仓储接口
type DownloadFilesRepository interface {
	// Create 创建下载文件记录
	Create(files *model.DownloadFiles) error

	// Update 更新下载文件记录
	Update(files *model.DownloadFiles) error

	// Delete 删除下载文件记录
	Delete(id uint) error

	// GetByHash 根据下载Hash获取文件列表
	GetByHash(downloadHash string, state *int) ([]*model.DownloadFiles, error)

	// GetByFullPath 根据完整路径获取文件
	GetByFullPath(fullPath string, allFiles bool) ([]*model.DownloadFiles, error)

	// GetBySavePath 根据保存路径获取文件列表
	GetBySavePath(savePath string) ([]*model.DownloadFiles, error)

	// DeleteByFullPath 根据完整路径删除文件
	DeleteByFullPath(fullPath string) error

	// UpdateState 更新文件状态
	UpdateState(id uint, state int) error

	// Count 统计文件数量
	Count() (int64, error)
}
