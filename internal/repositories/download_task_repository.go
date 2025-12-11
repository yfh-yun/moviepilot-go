package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"moviepilot-go/internal/models/entity"
)

// DownloadTaskRepository 下载任务仓储接口
type DownloadTaskRepository interface {
	// Create 创建下载任务
	Create(ctx context.Context, task *entity.DownloadTask) error
	// GetByID 根据ID获取下载任务
	GetByID(ctx context.Context, id uint) (*entity.DownloadTask, error)
	// GetByHash 根据哈希获取下载任务
	GetByHash(ctx context.Context, hash string) (*entity.DownloadTask, error)
	// GetByUserID 根据用户ID获取下载任务列表
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entity.DownloadTask, error)
	// GetByStatus 根据状态获取下载任务列表
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entity.DownloadTask, error)
	// Update 更新下载任务
	Update(ctx context.Context, task *entity.DownloadTask) error
	// UpdateStatus 更新下载任务状态
	UpdateStatus(ctx context.Context, id uint, status string) error
	// UpdateProgress 更新下载进度
	UpdateProgress(ctx context.Context, id uint, progress float64, downloadSpeed, uploadSpeed int64) error
	// Delete 删除下载任务
	Delete(ctx context.Context, id uint) error
	// Count 统计下载任务数量
	Count(ctx context.Context, userID uint, status string) (int64, error)
}

// downloadTaskRepository 下载任务仓储实现
type downloadTaskRepository struct {
	db *gorm.DB
}

// NewDownloadTaskRepository 创建下载任务仓储
func NewDownloadTaskRepository(db *gorm.DB) DownloadTaskRepository {
	return &downloadTaskRepository{db: db}
}

// Create 创建下载任务
func (r *downloadTaskRepository) Create(ctx context.Context, task *entity.DownloadTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取下载任务
func (r *downloadTaskRepository) GetByID(ctx context.Context, id uint) (*entity.DownloadTask, error) {
	var task entity.DownloadTask
	err := r.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByHash 根据哈希获取下载任务
func (r *downloadTaskRepository) GetByHash(ctx context.Context, hash string) (*entity.DownloadTask, error) {
	var task entity.DownloadTask
	err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByUserID 根据用户ID获取下载任务列表
func (r *downloadTaskRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*entity.DownloadTask, error) {
	var tasks []*entity.DownloadTask
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&tasks).Error
	return tasks, err
}

// GetByStatus 根据状态获取下载任务列表
func (r *downloadTaskRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entity.DownloadTask, error) {
	var tasks []*entity.DownloadTask
	query := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&tasks).Error
	return tasks, err
}

// Update 更新下载任务
func (r *downloadTaskRepository) Update(ctx context.Context, task *entity.DownloadTask) error {
	task.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(task).Error
}

// UpdateStatus 更新下载任务状态
func (r *downloadTaskRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&entity.DownloadTask{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// UpdateProgress 更新下载进度
func (r *downloadTaskRepository) UpdateProgress(ctx context.Context, id uint, progress float64, downloadSpeed, uploadSpeed int64) error {
	return r.db.WithContext(ctx).Model(&entity.DownloadTask{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"progress":       progress,
			"download_speed": downloadSpeed,
			"upload_speed":   uploadSpeed,
			"updated_at":     time.Now(),
		}).Error
}

// Delete 删除下载任务
func (r *downloadTaskRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.DownloadTask{}, id).Error
}

// Count 统计下载任务数量
func (r *downloadTaskRepository) Count(ctx context.Context, userID uint, status string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entity.DownloadTask{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&count).Error
	return count, err
}
