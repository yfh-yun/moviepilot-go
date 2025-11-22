package repositories

import (
	"fmt"

	"gorm.io/gorm"

	"moviepilot-go/internal/models"
)

// FindOptions 查询选项
type FindOptions struct {
	Filter  map[string]interface{} // 过滤条件
	OrderBy string                 // 排序字段
	Limit   int                    // 限制数量
	Offset  int                    // 偏移量
}

// SubscribeRepository 订阅仓储接口
type SubscribeRepository interface {
	Create(subscribe *models.Subscribe) error
	Update(subscribe *models.Subscribe) error
	Delete(id uint) error
	FindByID(id uint) (*models.Subscribe, error)
	FindAll(opts FindOptions) ([]models.Subscribe, int64, error)
	FindActive() ([]models.Subscribe, error)
	FindByTMDBID(tmdbID int, mediaType string, season *int) (*models.Subscribe, error)
	UpdateState(id uint, state string) error
	UpdateEpisode(id uint, episode int) error
}

// subscribeRepository 订阅仓储实现
type subscribeRepository struct {
	db *gorm.DB
}

// NewSubscribeRepository 创建订阅仓储
func NewSubscribeRepository(db *gorm.DB) SubscribeRepository {
	return &subscribeRepository{db: db}
}

// Create 创建订阅
func (r *subscribeRepository) Create(subscribe *models.Subscribe) error {
	return r.db.Create(subscribe).Error
}

// Update 更新订阅
func (r *subscribeRepository) Update(subscribe *models.Subscribe) error {
	return r.db.Save(subscribe).Error
}

// Delete 删除订阅
func (r *subscribeRepository) Delete(id uint) error {
	return r.db.Delete(&models.Subscribe{}, id).Error
}

// FindByID 根据 ID 查找订阅
func (r *subscribeRepository) FindByID(id uint) (*models.Subscribe, error) {
	var subscribe models.Subscribe
	err := r.db.First(&subscribe, id).Error
	if err != nil {
		return nil, err
	}
	return &subscribe, nil
}

// FindAll 查找所有订阅
func (r *subscribeRepository) FindAll(opts FindOptions) ([]models.Subscribe, int64, error) {
	var subscribes []models.Subscribe
	var total int64

	query := r.db.Model(&models.Subscribe{})

	// 应用过滤条件
	if opts.Filter != nil {
		for key, value := range opts.Filter {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if opts.OrderBy != "" {
		query = query.Order(opts.OrderBy)
	} else {
		query = query.Order("created_at DESC")
	}

	// 应用分页
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	err := query.Find(&subscribes).Error
	return subscribes, total, err
}

// FindActive 查找所有活跃订阅
func (r *subscribeRepository) FindActive() ([]models.Subscribe, error) {
	var subscribes []models.Subscribe
	err := r.db.Where("state IN ?", []string{"N", "R"}).Find(&subscribes).Error
	return subscribes, err
}

// FindByTMDBID 根据 TMDB ID 查找订阅
func (r *subscribeRepository) FindByTMDBID(tmdbID int, mediaType string, season *int) (*models.Subscribe, error) {
	var subscribe models.Subscribe
	query := r.db.Where("tmdb_id = ? AND type = ?", tmdbID, mediaType)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.First(&subscribe).Error
	if err != nil {
		return nil, err
	}
	return &subscribe, nil
}

// UpdateState 更新订阅状态
func (r *subscribeRepository) UpdateState(id uint, state string) error {
	return r.db.Model(&models.Subscribe{}).Where("id = ?", id).Update("state", state).Error
}

// UpdateEpisode 更新订阅集数
func (r *subscribeRepository) UpdateEpisode(id uint, episode int) error {
	return r.db.Model(&models.Subscribe{}).Where("id = ?", id).Updates(map[string]interface{}{
		"start_episode": episode,
		"lack_episode":  gorm.Expr("GREATEST(0, total_episode - ?)", episode),
	}).Error
}
