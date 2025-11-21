package repositories

import (
	"context"
	"errors"
	
	"gorm.io/gorm"
	
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/models"
)

// MediaRepositoryImpl 媒体仓储实现
type MediaRepositoryImpl struct {
	db *gorm.DB
}

// NewMediaRepository 创建媒体仓储实例
func NewMediaRepository(db *gorm.DB) interfaces.MediaRepository {
	return &MediaRepositoryImpl{db: db}
}

// Create 创建媒体记录
func (r *MediaRepositoryImpl) Create(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

// GetByID 根据ID获取媒体
func (r *MediaRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Media, error) {
	var media models.Media
	err := r.db.WithContext(ctx).First(&media, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// Update 更新媒体
func (r *MediaRepositoryImpl) Update(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Save(media).Error
}

// Delete 删除媒体
func (r *MediaRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Media{}, "id = ?", id).Error
}

// List 获取媒体列表
func (r *MediaRepositoryImpl) List(ctx context.Context, params interfaces.ListMediaParams) ([]*models.Media, int64, error) {
	var medias []*models.Media
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Media{})
	
	// 添加过滤条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Genre != "" {
		query = query.Where("genre LIKE ?", "%"+params.Genre+"%")
	}
	if params.Year > 0 {
		query = query.Where("year = ?", params.Year)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Offset(offset).Limit(params.PageSize).Find(&medias).Error
	
	return medias, total, err
}

// Search 搜索媒体
func (r *MediaRepositoryImpl) Search(ctx context.Context, query string, params interfaces.SearchMediaParams) ([]*models.Media, error) {
	var medias []*models.Media
	
	dbQuery := r.db.WithContext(ctx).Where("title LIKE ? OR original_title LIKE ?", "%"+query+"%", "%"+query+"%")
	
	// 添加过滤条件
	if params.Type != "" {
		dbQuery = dbQuery.Where("type = ?", params.Type)
	}
	if params.Year > 0 {
		dbQuery = dbQuery.Where("year = ?", params.Year)
	}
	if params.Genre != "" {
		dbQuery = dbQuery.Where("genre LIKE ?", "%"+params.Genre+"%")
	}
	
	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := dbQuery.Offset(offset).Limit(params.PageSize).Find(&medias).Error
	
	return medias, err
}

// GetByTMDBID 根据TMDB ID获取媒体
func (r *MediaRepositoryImpl) GetByTMDBID(ctx context.Context, tmdbID int64, mediaType string) (*models.Media, error) {
	var media models.Media
	err := r.db.WithContext(ctx).Where("tmdb_id = ? AND type = ?", tmdbID, mediaType).First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// GetByTitle 根据标题获取媒体
func (r *MediaRepositoryImpl) GetByTitle(ctx context.Context, title string, year int) (*models.Media, error) {
	var media models.Media
	query := r.db.WithContext(ctx).Where("title = ?", title)
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	
	err := query.First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}