package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/model"
)

// MediaRepository 媒体仓储接口
type MediaRepository interface {
	// Create 创建媒体记录
	Create(media *model.Media) error

	// Update 更新媒体记录
	Update(media *model.Media) error

	// Delete 删除媒体记录
	Delete(id uint) error

	// GetByID 根据ID获取媒体
	GetByID(id uint) (*model.Media, error)

	// GetByTMDBID 根据TMDBID获取媒体
	GetByTMDBID(tmdbID int) (*model.Media, error)

	// GetByIMDBID 根据IMDBID获取媒体
	GetByIMDBID(imdbID string) (*model.Media, error)

	// GetByDoubanID 根据豆瓣ID获取媒体
	GetByDoubanID(doubanID string) (*model.Media, error)

	// GetByBangumiID 根据BangumiID获取媒体
	GetByBangumiID(bangumiID int) (*model.Media, error)

	// GetByTitle 根据标题获取媒体
	GetByTitle(title string, year *string) ([]*model.Media, error)

	// List 根据条件获取媒体列表
	List(mtype string, offset, limit int) ([]*model.Media, int64, error)

	// Search 搜索媒体
	Search(keyword string, mtype string, offset, limit int) ([]*model.Media, int64, error)

	// UpdateStatus 更新状态
	UpdateStatus(id uint, status string) error

	// UpdatePoster 更新海报
	UpdatePoster(id uint, poster string) error

	// UpdateBackdrop 更新背景图
	UpdateBackdrop(id uint, backdrop string) error

	// Exists 检查媒体是否存在
	Exists(tmdbID *int, doubanID *string) (bool, error)

	// Count 统计媒体数量
	Count(mtype string) (int64, error)

	// GetRecent 获取最近的媒体记录
	GetRecent(days int, limit int) ([]*model.Media, error)
}
