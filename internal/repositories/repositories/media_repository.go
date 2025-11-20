package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mediaRepository 媒体仓储实现
type mediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository 创建媒体仓储
func NewMediaRepository(db *gorm.DB) interfaces.MediaRepository {
	return &mediaRepository{db: db}
}

// Create 创建媒体记录
func (r *mediaRepository) Create(media *model.Media) error {
	return r.db.Create(media).Error
}

// Update 更新媒体记录
func (r *mediaRepository) Update(media *model.Media) error {
	return r.db.Save(media).Error
}

// Delete 删除媒体记录
func (r *mediaRepository) Delete(id uint) error {
	return r.db.Delete(&model.Media{}, id).Error
}

// GetByID 根据ID获取媒体
func (r *mediaRepository) GetByID(id uint) (*model.Media, error) {
	var media model.Media
	err := r.db.First(&media, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// GetByTMDBID 根据TMDBID获取媒体
func (r *mediaRepository) GetByTMDBID(tmdbID int) (*model.Media, error) {
	var media model.Media
	err := r.db.Where("tmdbid = ?", tmdbID).First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// GetByIMDBID 根据IMDBID获取媒体
func (r *mediaRepository) GetByIMDBID(imdbID string) (*model.Media, error) {
	var media model.Media
	err := r.db.Where("imdbid = ?", imdbID).First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// GetByDoubanID 根据豆瓣ID获取媒体
func (r *mediaRepository) GetByDoubanID(doubanID string) (*model.Media, error) {
	var media model.Media
	err := r.db.Where("doubanid = ?", doubanID).First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// GetByBangumiID 根据BangumiID获取媒体
func (r *mediaRepository) GetByBangumiID(bangumiID int) (*model.Media, error) {
	var media model.Media
	err := r.db.Where("bangumiid = ?", bangumiID).First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// GetByTitle 根据标题获取媒体
func (r *mediaRepository) GetByTitle(title string, year *string) ([]*model.Media, error) {
	var medias []*model.Media
	query := r.db.Where("title = ?", title)

	if year != nil {
		query = query.Where("year = ?", *year)
	}

	err := query.Find(&medias).Error
	return medias, err
}

// List 根据条件获取媒体列表
func (r *mediaRepository) List(mtype string, offset, limit int) ([]*model.Media, int64, error) {
	var medias []*model.Media
	var total int64
	query := r.db.Model(&model.Media{})

	if mtype != "" {
		query = query.Where("type = ?", mtype)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Find(&medias).Error
	return medias, total, err
}

// Search 搜索媒体
func (r *mediaRepository) Search(keyword string, mtype string, offset, limit int) ([]*model.Media, int64, error) {
	var medias []*model.Media
	var total int64
	query := r.db.Model(&model.Media{}).Where("title LIKE ? OR original_title LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	if mtype != "" {
		query = query.Where("type = ?", mtype)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Find(&medias).Error
	return medias, total, err
}

// UpdateStatus 更新状态
func (r *mediaRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Media{}).Where("id = ?", id).Update("state", status).Error
}

// UpdatePoster 更新海报
func (r *mediaRepository) UpdatePoster(id uint, poster string) error {
	return r.db.Model(&model.Media{}).Where("id = ?", id).Update("poster", poster).Error
}

// UpdateBackdrop 更新背景图
func (r *mediaRepository) UpdateBackdrop(id uint, backdrop string) error {
	return r.db.Model(&model.Media{}).Where("id = ?", id).Update("backdrop", backdrop).Error
}

// Exists 检查媒体是否存在
func (r *mediaRepository) Exists(tmdbID *int, doubanID *string) (bool, error) {
	var count int64
	query := r.db.Model(&model.Media{})

	if tmdbID != nil {
		query = query.Where("tmdbid = ?", *tmdbID)
	} else if doubanID != nil {
		query = query.Where("doubanid = ?", *doubanID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// Count 统计媒体数量
func (r *mediaRepository) Count(mtype string) (int64, error) {
	var count int64
	query := r.db.Model(&model.Media{})

	if mtype != "" {
		query = query.Where("type = ?", mtype)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetRecent 获取最近的媒体记录
func (r *mediaRepository) GetRecent(days int, limit int) ([]*model.Media, error) {
	var medias []*model.Media
	since := time.Now().AddDate(0, 0, -days)

	query := r.db.Where("created_at >= ?", since).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&medias).Error
	return medias, err
}
