package interfaces

import "github.com/yfh-yun/moviepilot-go/internal/model"

// SubscribeHistoryRepository 订阅历史仓储接口
type SubscribeHistoryRepository interface {
	// 基础CRUD
	Create(history *model.SubscribeHistory) error
	GetByID(id uint) (*model.SubscribeHistory, error)
	Update(history *model.SubscribeHistory) error
	Delete(id uint) error
	List(offset, limit int) ([]*model.SubscribeHistory, int64, error)
	
	// 按媒体ID获取
	GetByMediaID(mediaID string) ([]*model.SubscribeHistory, error)
	
	// 按TMDBID获取
	GetByTMDBID(tmdbID int, season *int) ([]*model.SubscribeHistory, error)
	
	// 按豆瓣ID获取
	GetByDoubanID(doubanID string) ([]*model.SubscribeHistory, error)
	
	// 按BangumiID获取
	GetByBangumiID(bangumiID int) ([]*model.SubscribeHistory, error)
	
	// 按标题获取
	GetByTitle(title string, season *int) ([]*model.SubscribeHistory, error)
	
	// 按状态获取
	GetByState(state string) ([]*model.SubscribeHistory, error)
	
	// 按日期获取
	GetByDate(date string) ([]*model.SubscribeHistory, error)
	
	// 检查是否存在
	ExistsByTMDBID(tmdbID int, season *int) bool
	ExistsByDoubanID(doubanID string) bool
	ExistsByBangumiID(bangumiID int) bool
	ExistsByMediaID(mediaID string) bool
	
	// 批量操作
	BatchCreate(histories []*model.SubscribeHistory) error
}