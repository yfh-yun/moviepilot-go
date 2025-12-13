package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/repositories/interfaces"
)

// DownloadHistoryRepositoryImpl 下载历史仓储实现
type DownloadHistoryRepositoryImpl struct {
	db *gorm.DB
}

// NewDownloadHistoryRepository 创建下载历史仓储实例
func NewDownloadHistoryRepository(db *gorm.DB) interfaces.DownloadHistoryRepository {
	return &DownloadHistoryRepositoryImpl{db: db}
}

// Create 创建下载历史
func (r *DownloadHistoryRepositoryImpl) Create(ctx context.Context, history *database.DownloadHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetByID 根据ID获取下载历史
func (r *DownloadHistoryRepositoryImpl) GetByID(ctx context.Context, id uint) (*database.DownloadHistory, error) {
	var history database.DownloadHistory
	err := r.db.WithContext(ctx).First(&history, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// Update 更新下载历史
func (r *DownloadHistoryRepositoryImpl) Update(ctx context.Context, history *database.DownloadHistory) error {
	return r.db.WithContext(ctx).Save(history).Error
}

// Delete 删除下载历史
func (r *DownloadHistoryRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&database.DownloadHistory{}, id).Error
}

// GetByPath 根据路径查询下载历史
func (r *DownloadHistoryRepositoryImpl) GetByPath(ctx context.Context, path string) (*database.DownloadHistory, error) {
	var history database.DownloadHistory
	err := r.db.WithContext(ctx).Where("path = ?", path).First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetByHash 根据hash查询下载历史
func (r *DownloadHistoryRepositoryImpl) GetByHash(ctx context.Context, hash string) (*database.DownloadHistory, error) {
	var history database.DownloadHistory
	err := r.db.WithContext(ctx).Where("download_hash = ?", hash).First(&history).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetByMediaID 根据媒体ID查询下载历史
func (r *DownloadHistoryRepositoryImpl) GetByMediaID(ctx context.Context, tmdbID *int, doubanID *string) ([]*database.DownloadHistory, error) {
	var histories []*database.DownloadHistory
	query := r.db.WithContext(ctx).Model(&database.DownloadHistory{})

	if tmdbID != nil {
		query = query.Where("tmdb_id = ?", *tmdbID)
	}
	if doubanID != nil && *doubanID != "" {
		query = query.Where("douban_id = ?", *doubanID)
	}

	err := query.Order("date DESC").Find(&histories).Error
	return histories, err
}

// ListByPage 分页查询下载历史
func (r *DownloadHistoryRepositoryImpl) ListByPage(ctx context.Context, params interfaces.ListDownloadHistoryParams) ([]*database.DownloadHistory, int64, error) {
	var histories []*database.DownloadHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&database.DownloadHistory{})

	// 添加过滤条件
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.State != "" {
		query = query.Where("state = ?", params.State)
	}
	if params.Title != "" {
		query = query.Where("title LIKE ?", "%"+params.Title+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.
		Offset(offset).
		Limit(params.PageSize).
		Order("date DESC").
		Find(&histories).Error

	return histories, total, err
}

// AddFiles 批量添加下载文件
func (r *DownloadHistoryRepositoryImpl) AddFiles(ctx context.Context, files []*database.DownloadFile) error {
	if len(files) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(files, 100).Error
}

// GetFilesByHash 根据hash查询下载文件
func (r *DownloadHistoryRepositoryImpl) GetFilesByHash(ctx context.Context, hash string, state *int) ([]*database.DownloadFile, error) {
	var files []*database.DownloadFile
	query := r.db.WithContext(ctx).Where("download_hash = ?", hash)

	if state != nil {
		query = query.Where("state = ?", *state)
	}

	err := query.Find(&files).Error
	return files, err
}

// GetFileByFullPath 根据完整路径查询单个下载文件
func (r *DownloadHistoryRepositoryImpl) GetFileByFullPath(ctx context.Context, fullPath string) (*database.DownloadFile, error) {
	var file database.DownloadFile
	err := r.db.WithContext(ctx).Where("fullpath = ?", fullPath).First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// GetFilesByFullPath 根据完整路径查询所有下载文件
func (r *DownloadHistoryRepositoryImpl) GetFilesByFullPath(ctx context.Context, fullPath string) ([]*database.DownloadFile, error) {
	var files []*database.DownloadFile
	err := r.db.WithContext(ctx).Where("fullpath LIKE ?", fullPath+"%").Find(&files).Error
	return files, err
}

// GetFilesBySavePath 根据保存路径查询下载文件
func (r *DownloadHistoryRepositoryImpl) GetFilesBySavePath(ctx context.Context, savePath string) ([]*database.DownloadFile, error) {
	var files []*database.DownloadFile
	err := r.db.WithContext(ctx).Where("savepath LIKE ?", savePath+"%").Find(&files).Error
	return files, err
}

// UpdateFileState 更新文件状态
func (r *DownloadHistoryRepositoryImpl) UpdateFileState(ctx context.Context, fullPath string, state int) error {
	return r.db.WithContext(ctx).
		Model(&database.DownloadFile{}).
		Where("fullpath = ?", fullPath).
		Update("state", state).Error
}

// GetLastBy 根据tmdbid、season、season_episode查询下载记录
// tmdbid + mtype 或 title + year
func (r *DownloadHistoryRepositoryImpl) GetLastBy(ctx context.Context, mtype, title, year, season, episode *string, tmdbID *int) ([]*database.DownloadHistory, error) {
	var histories []*database.DownloadHistory
	query := r.db.WithContext(ctx).Model(&database.DownloadHistory{}).Order("id DESC")

	// TMDBID + 类型
	if tmdbID != nil && mtype != nil && *mtype != "" {
		query = query.Where("tmdbid = ? AND type = ?", *tmdbID, *mtype)

		// 电视剧某季某集
		if season != nil && *season != "" && episode != nil && *episode != "" {
			query = query.Where("seasons = ? AND episodes = ?", *season, *episode)
			// 电视剧某季
		} else if season != nil && *season != "" {
			query = query.Where("seasons = ?", *season)
		}
		// 标题 + 年份
	} else if title != nil && *title != "" && year != nil && *year != "" {
		query = query.Where("title = ? AND year = ?", *title, *year)

		// 电视剧某季某集
		if season != nil && *season != "" && episode != nil && *episode != "" {
			query = query.Where("seasons = ? AND episodes = ?", *season, *episode)
			// 电视剧某季
		} else if season != nil && *season != "" {
			query = query.Where("seasons = ?", *season)
		}
	} else {
		// 不满足条件，返回空列表
		return []*database.DownloadHistory{}, nil
	}

	err := query.Find(&histories).Error
	return histories, err
}

// ListByUserDate 查询某用户某时间之后的下载历史
func (r *DownloadHistoryRepositoryImpl) ListByUserDate(ctx context.Context, date, username string) ([]*database.DownloadHistory, error) {
	var histories []*database.DownloadHistory
	query := r.db.WithContext(ctx).Model(&database.DownloadHistory{}).Order("id DESC")

	query = query.Where("date < ?", date)
	if username != "" {
		query = query.Where("username = ?", username)
	}

	err := query.Find(&histories).Error
	return histories, err
}

// ListByDate 查询某时间之后的下载历史
func (r *DownloadHistoryRepositoryImpl) ListByDate(ctx context.Context, date, mtype string, tmdbID int, seasons *string) ([]*database.DownloadHistory, error) {
	var histories []*database.DownloadHistory
	query := r.db.WithContext(ctx).Model(&database.DownloadHistory{}).Order("id DESC")

	query = query.Where("date > ? AND type = ? AND tmdbid = ?", date, mtype, tmdbID)
	if seasons != nil && *seasons != "" {
		query = query.Where("seasons = ?", *seasons)
	}

	err := query.Find(&histories).Error
	return histories, err
}

// ListByType 根据类型和天数查询下载历史
func (r *DownloadHistoryRepositoryImpl) ListByType(ctx context.Context, mtype string, days int) ([]*database.DownloadHistory, error) {
	var histories []*database.DownloadHistory
	query := r.db.WithContext(ctx).Model(&database.DownloadHistory{})

	// 计算天数前的时间
	query = query.Where("type = ? AND date >= DATE_SUB(NOW(), INTERVAL ? DAY)", mtype, days)

	err := query.Find(&histories).Error
	return histories, err
}

// TruncateFiles 清空下载文件表
func (r *DownloadHistoryRepositoryImpl) TruncateFiles(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM downloadfiles").Error
}
