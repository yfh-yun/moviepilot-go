package interfaces

import (
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"time"
)

// SubscribeRepository 订阅仓储接口
type SubscribeRepository interface {
	// Create 创建订阅
	Create(subscribe *model.Subscribe) error

	// Update 更新订阅
	Update(subscribe *model.Subscribe) error

	// Delete 删除订阅
	Delete(id uint) error

	// GetByID 根据ID获取订阅
	GetByID(id uint) (*model.Subscribe, error)

	// GetByTMDBID 根据TMDBID获取订阅
	GetByTMDBID(tmdbID int, season *int) ([]*model.Subscribe, error)

	// GetByDoubanID 根据豆瓣ID获取订阅
	GetByDoubanID(doubanID string) (*model.Subscribe, error)

	// GetByBangumiID 根据BangumiID获取订阅
	GetByBangumiID(bangumiID int) (*model.Subscribe, error)

	// GetByMediaID 根据媒体ID获取订阅
	GetByMediaID(mediaID string) (*model.Subscribe, error)

	// GetByTitle 根据标题获取订阅
	GetByTitle(title string, season *int) (*model.Subscribe, error)

	// GetByState 根据状态获取订阅列表
	GetByState(state string) ([]*model.Subscribe, error)

	// ListByUsername 根据用户名获取订阅列表
	ListByUsername(username string, state, mtype string) ([]*model.Subscribe, error)

	// ListByType 根据类型和时间获取订阅列表
	ListByType(mtype string, days int) ([]*model.Subscribe, error)

	// List 获取订阅列表
	List(offset, limit int) ([]*model.Subscribe, int64, error)

	// UpdateState 更新订阅状态
	UpdateState(id uint, state string) error

	// UpdateLastUpdate 更新最后更新时间
	UpdateLastUpdate(id uint, lastUpdate time.Time) error

	// DeleteByTMDBID 根据TMDBID删除订阅
	DeleteByTMDBID(tmdbID int, season int) error

	// DeleteByDoubanID 根据豆瓣ID删除订阅
	DeleteByDoubanID(doubanID string) error

	// DeleteByMediaID 根据媒体ID删除订阅
	DeleteByMediaID(mediaID string) error

	// Exists 检查订阅是否存在
	Exists(tmdbID *int, doubanID *string, season *int) (bool, error)

	// Count 统计订阅数量
	Count() (int64, error)

	// CountByState 根据状态统计订阅数量
	CountByState(state string) (int64, error)
}
