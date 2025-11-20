package repositories

import (
	"context"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// transferHistoryRepository 转移历史仓储实现
type transferHistoryRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTransferHistoryRepository 创建转移历史仓储
func NewTransferHistoryRepository(db *gorm.DB, logger *zap.Logger) interfaces.TransferHistoryRepository {
	return &transferHistoryRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建转移历史记录
func (r *transferHistoryRepository) Create(history *model.TransferHistory) error {
	return r.db.Create(history).Error
}

// Update 更新转移历史记录
func (r *transferHistoryRepository) Update(history *model.TransferHistory) error {
	return r.db.Save(history).Error
}

// Delete 删除转移历史记录
func (r *transferHistoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.TransferHistory{}, id).Error
}

// GetByID 根据ID获取转移历史
func (r *transferHistoryRepository) GetByID(id uint) (*model.TransferHistory, error) {
	var history model.TransferHistory
	err := r.db.First(&history, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// List 获取转移历史列表
func (r *transferHistoryRepository) List(limit, offset int) ([]*model.TransferHistory, int64, error) {
	var histories []*model.TransferHistory
	var total int64

	// 获取总数
	err := r.db.Model(&model.TransferHistory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = r.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&histories).Error
	return histories, total, err
}

// ListByStatus 根据状态获取转移历史
func (r *transferHistoryRepository) ListByStatus(status TransferStatus, limit, offset int) ([]*model.TransferHistory, int64, error) {
	var histories []*model.TransferHistory
	var total int64

	query := r.db.Model(&model.TransferHistory{}).Where("status = ?", status)

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&histories).Error
	return histories, total, err
}

// ListByMediaID 根据媒体ID获取转移历史
func (r *transferHistoryRepository) ListByMediaID(mediaID uint) ([]*model.TransferHistory, error) {
	var histories []*model.TransferHistory
	err := r.db.Where("media_id = ?", mediaID).Order("created_at DESC").Find(&histories).Error
	return histories, err
}

// ListByDateRange 根据日期范围获取转移历史
func (r *transferHistoryRepository) ListByDateRange(start, end time.Time, limit, offset int) ([]*model.TransferHistory, int64, error) {
	var histories []*model.TransferHistory
	var total int64

	query := r.db.Model(&model.TransferHistory{}).Where("created_at BETWEEN ? AND ?", start, end)

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&histories).Error
	return histories, total, err
}

// GetBySourcePath 根据源路径获取转移历史
func (r *transferHistoryRepository) GetBySourcePath(sourcePath string) (*model.TransferHistory, error) {
	var history model.TransferHistory
	err := r.db.Where("source_path = ?", sourcePath).First(&history).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// GetByTargetPath 根据目标路径获取转移历史
func (r *transferHistoryRepository) GetByTargetPath(targetPath string) (*model.TransferHistory, error) {
	var history model.TransferHistory
	err := r.db.Where("target_path = ?", targetPath).First(&history).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

// UpdateStatus 更新转移状态
func (r *transferHistoryRepository) UpdateStatus(id uint, status TransferStatus, message string) error {
	return r.db.Model(&model.TransferHistory{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"status_msg":   message,
			"updated_at":   time.Now(),
		}).Error
}

// SetProgress 设置转移进度
func (r *transferHistoryRepository) SetProgress(id uint, progress int) error {
	return r.db.Model(&model.TransferHistory{}).
		Where("id = ?", id).
		Update("progress", progress).Error
}

// SetDuration 设置转移耗时
func (r *transferHistoryRepository) SetDuration(id uint, duration time.Duration) error {
	return r.db.Model(&model.TransferHistory{}).
		Where("id = ?", id).
		Update("duration", duration.Milliseconds()).Error
}

// GetStatistics 获取转移统计
func (r *transferHistoryRepository) GetStatistics(ctx context.Context, days int) (*model.TransferStatistics, error) {
	var stats model.TransferStatistics

	startDate := time.Now().AddDate(0, 0, -days)

	// 总转移数
	err := r.db.Model(&model.TransferHistory{}).
		Where("created_at >= ?", startDate).
		Count(&stats.TotalTransfers).Error
	if err != nil {
		return nil, err
	}

	// 成功转移数
	err = r.db.Model(&model.TransferHistory{}).
		Where("created_at >= ? AND status = ?", startDate, TransferStatusSuccess).
		Count(&stats.SuccessfulTransfers).Error
	if err != nil {
		return nil, err
	}

	// 失败转移数
	err = r.db.Model(&model.TransferHistory{}).
		Where("created_at >= ? AND status = ?", startDate, TransferStatusFailed).
		Count(&stats.FailedTransfers).Error
	if err != nil {
		return nil, err
	}

	// 总转移大小
	err = r.db.Model(&model.TransferHistory{}).
		Where("created_at >= ? AND status = ?", startDate, TransferStatusSuccess).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&stats.TotalSize).Error
	if err != nil {
		return nil, err
	}

	// 平均转移时间
	err = r.db.Model(&model.TransferHistory{}).
		Where("created_at >= ? AND status = ? AND duration > 0", startDate, TransferStatusSuccess).
		Select("COALESCE(AVG(duration), 0)").
		Scan(&stats.AverageDuration).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// DeleteOldRecords 删除旧记录
func (r *transferHistoryRepository) DeleteOldRecords(olderThan time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", olderThan).
		Delete(&model.TransferHistory{})
	return result.RowsAffected, result.Error
}

// BatchCreate 批量创建转移历史
func (r *transferHistoryRepository) BatchCreate(histories []*model.TransferHistory) error {
	return r.db.CreateInBatches(histories, 100).Error
}

// BatchUpdateStatus 批量更新状态
func (r *transferHistoryRepository) BatchUpdateStatus(ids []uint, status TransferStatus, message string) error {
	return r.db.Model(&model.TransferHistory{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":       status,
			"status_msg":   message,
			"updated_at":   time.Now(),
		}).Error
}

// GetFailedTransfers 获取失败的转移记录
func (r *transferHistoryRepository) GetFailedTransfers(limit int) ([]*model.TransferHistory, error) {
	var histories []*model.TransferHistory
	err := r.db.Where("status = ?", TransferStatusFailed).
		Order("created_at DESC").
		Limit(limit).
		Find(&histories).Error
	return histories, err
}

// RetryFailedTransfer 重试失败的转移
func (r *transferHistoryRepository) RetryFailedTransfer(id uint) error {
	return r.db.Model(&model.TransferHistory{}).
		Where("id = ? AND status = ?", id, TransferStatusFailed).
		Updates(map[string]interface{}{
			"status":       TransferStatusPending,
			"status_msg":   "准备重试",
			"progress":     0,
			"retry_count":  gorm.Expr("retry_count + 1"),
			"updated_at":   time.Now(),
		}).Error
}