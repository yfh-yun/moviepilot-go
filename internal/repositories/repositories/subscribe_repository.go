package repositories

import (
	"errors"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"time"

	"gorm.io/gorm"
)

// subscribeRepository 订阅仓储实现
type subscribeRepository struct {
	db *gorm.DB
}

// NewSubscribeRepository 创建订阅仓储
func NewSubscribeRepository(db *gorm.DB) interfaces.SubscribeRepository {
	return &subscribeRepository{db: db}
}

// Create 创建订阅
func (r *subscribeRepository) Create(subscribe *model.Subscribe) error {
	return r.db.Create(subscribe).Error
}

// Update 更新订阅
func (r *subscribeRepository) Update(subscribe *model.Subscribe) error {
	return r.db.Save(subscribe).Error
}

// Delete 删除订阅
func (r *subscribeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Subscribe{}, id).Error
}

// GetByID 根据ID获取订阅
func (r *subscribeRepository) GetByID(id uint) (*model.Subscribe, error) {
	var subscribe model.Subscribe
	err := r.db.First(&subscribe, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// GetByTMDBID 根据TMDBID获取订阅
func (r *subscribeRepository) GetByTMDBID(tmdbID int, season *int) ([]*model.Subscribe, error) {
	var subscribes []*model.Subscribe
	query := r.db.Where("tmdbid = ?", tmdbID)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.Find(&subscribes).Error
	return subscribes, err
}

// GetByDoubanID 根据豆瓣ID获取订阅
func (r *subscribeRepository) GetByDoubanID(doubanID string) (*model.Subscribe, error) {
	var subscribe model.Subscribe
	err := r.db.Where("doubanid = ?", doubanID).First(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// GetByBangumiID 根据BangumiID获取订阅
func (r *subscribeRepository) GetByBangumiID(bangumiID int) (*model.Subscribe, error) {
	var subscribe model.Subscribe
	err := r.db.Where("bangumiid = ?", bangumiID).First(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// GetByMediaID 根据媒体ID获取订阅
func (r *subscribeRepository) GetByMediaID(mediaID string) (*model.Subscribe, error) {
	var subscribe model.Subscribe
	err := r.db.Where("mediaid = ?", mediaID).First(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// GetByTitle 根据标题获取订阅
func (r *subscribeRepository) GetByTitle(title string, season *int) (*model.Subscribe, error) {
	var subscribe model.Subscribe
	query := r.db.Where("name = ?", title)

	if season != nil {
		query = query.Where("season = ?", *season)
	}

	err := query.First(&subscribe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscribe, nil
}

// GetByState 根据状态获取订阅列表
func (r *subscribeRepository) GetByState(state string) ([]*model.Subscribe, error) {
	var subscribes []*model.Subscribe
	query := r.db

	if state != "" {
		query = query.Where("state IN ?", state)
	}

	err := query.Find(&subscribes).Error
	return subscribes, err
}

// ListByUsername 根据用户名获取订阅列表
func (r *subscribeRepository) ListByUsername(username string, state, mtype string) ([]*model.Subscribe, error) {
	var subscribes []*model.Subscribe
	query := r.db.Where("username = ?", username)

	if state != "" {
		query = query.Where("state = ?", state)
	}

	if mtype != "" {
		query = query.Where("type = ?", mtype)
	}

	err := query.Find(&subscribes).Error
	return subscribes, err
}

// ListByType 根据类型和时间获取订阅列表
func (r *subscribeRepository) ListByType(mtype string, days int) ([]*model.Subscribe, error) {
	var subscribes []*model.Subscribe
	since := time.Now().AddDate(0, 0, -days)

	err := r.db.Where("type = ? AND created_at >= ?", mtype, since).Find(&subscribes).Error
	return subscribes, err
}

// List 获取订阅列表
func (r *subscribeRepository) List(offset, limit int) ([]*model.Subscribe, int64, error) {
	var subscribes []*model.Subscribe
	var total int64

	err := r.db.Model(&model.Subscribe{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Find(&subscribes).Error
	return subscribes, total, err
}

// UpdateState 更新订阅状态
func (r *subscribeRepository) UpdateState(id uint, state string) error {
	return r.db.Model(&model.Subscribe{}).Where("id = ?", id).Update("state", state).Error
}

// UpdateLastUpdate 更新最后更新时间
func (r *subscribeRepository) UpdateLastUpdate(id uint, lastUpdate time.Time) error {
	return r.db.Model(&model.Subscribe{}).Where("id = ?", id).Update("last_update", lastUpdate).Error
}

// DeleteByTMDBID 根据TMDBID删除订阅
func (r *subscribeRepository) DeleteByTMDBID(tmdbID int, season int) error {
	return r.db.Where("tmdbid = ? AND season = ?", tmdbID, season).Delete(&model.Subscribe{}).Error
}

// DeleteByDoubanID 根据豆瓣ID删除订阅
func (r *subscribeRepository) DeleteByDoubanID(doubanID string) error {
	return r.db.Where("doubanid = ?", doubanID).Delete(&model.Subscribe{}).Error
}

// DeleteByMediaID 根据媒体ID删除订阅
func (r *subscribeRepository) DeleteByMediaID(mediaID string) error {
	return r.db.Where("mediaid = ?", mediaID).Delete(&model.Subscribe{}).Error
}

// Exists 检查订阅是否存在
func (r *subscribeRepository) Exists(tmdbID *int, doubanID *string, season *int) (bool, error) {
	var count int64
	query := r.db.Model(&model.Subscribe{})

	if tmdbID != nil {
		query = query.Where("tmdbid = ?", *tmdbID)
		if season != nil {
			query = query.Where("season = ?", *season)
		}
	} else if doubanID != nil {
		query = query.Where("doubanid = ?", *doubanID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// Count 统计订阅数量
func (r *subscribeRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Subscribe{}).Count(&count).Error
	return count, err
}

// CountByState 根据状态统计订阅数量
func (r *subscribeRepository) CountByState(state string) (int64, error) {
	var count int64
	query := r.db.Model(&model.Subscribe{})

	if state != "" {
		query = query.Where("state = ?", state)
	}

	err := query.Count(&count).Error
	return count, err
}
