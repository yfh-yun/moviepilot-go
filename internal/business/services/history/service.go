package history

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/pkg/logger"
)

// Service 历史记录服务接口
type Service interface {
	// GetDownloadHistory 获取下载历史
	GetDownloadHistory(ctx context.Context, userID string, limit int) ([]*database.DownloadHistory, error)

	// GetDownloadHistoryByPage 分页获取下载历史
	GetDownloadHistoryByPage(ctx context.Context, page int, count int) ([]*database.DownloadHistory, error)

	// DeleteDownloadHistory 删除下载历史
	DeleteDownloadHistory(ctx context.Context, historyID int) error

	// GetTransferHistory 分页获取整理记录
	GetTransferHistory(ctx context.Context, title string, page int, count int, status *bool) ([]*database.TransferHistory, int64, error)

	// DeleteTransferHistory 删除整理记录
	DeleteTransferHistory(ctx context.Context, historyID int, deleteSrc bool, deleteDest bool) error

	// EmptyTransferHistory 清空整理记录
	EmptyTransferHistory(ctx context.Context) error

	// GetOperationHistory 获取操作历史
	GetOperationHistory(ctx context.Context, userID string, limit int) ([]*OperationRecord, error)

	// RecordOperation 记录操作
	RecordOperation(ctx context.Context, record *OperationRecord) error

	// ClearHistory 清空历史
	ClearHistory(ctx context.Context, userID string, historyType string) error

	// GetHistoryStats 获取历史统计
	GetHistoryStats(ctx context.Context, userID string) (*HistoryStats, error)
}

// service 服务实现
type service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建服务
func NewService(db *gorm.DB) Service {
	return &service{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// OperationRecord 操作记录
type OperationRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"size:100;index" json:"user_id"`
	Operation   string    `gorm:"size:50" json:"operation"` // subscribe, download, transfer, etc.
	Target      string    `gorm:"size:200" json:"target"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"size:20" json:"status"` // success, failed
	Details     string    `gorm:"type:json" json:"details"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// TableName 表名
func (OperationRecord) TableName() string {
	return "operation_history"
}

// HistoryStats 历史统计
type HistoryStats struct {
	TotalDownloads  int64 `json:"total_downloads"`
	TotalOperations int64 `json:"total_operations"`
	TodayDownloads  int64 `json:"today_downloads"`
	TodayOperations int64 `json:"today_operations"`
	WeekDownloads   int64 `json:"week_downloads"`
	WeekOperations  int64 `json:"week_operations"`
	MonthDownloads  int64 `json:"month_downloads"`
	MonthOperations int64 `json:"month_operations"`
}

// GetDownloadHistory 获取下载历史
func (s *service) GetDownloadHistory(ctx context.Context, userID string, limit int) ([]*database.DownloadHistory, error) {
	s.logger.Info("获取下载历史",
		zap.String("user_id", userID),
		zap.Int("limit", limit),
	)

	var history []*database.DownloadHistory
	err := s.db.WithContext(ctx).
		Where("username = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error

	return history, err
}

// GetOperationHistory 获取操作历史
func (s *service) GetOperationHistory(ctx context.Context, userID string, limit int) ([]*OperationRecord, error) {
	s.logger.Info("获取操作历史",
		zap.String("user_id", userID),
		zap.Int("limit", limit),
	)

	var records []*OperationRecord
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error

	return records, err
}

// RecordOperation 记录操作
func (s *service) RecordOperation(ctx context.Context, record *OperationRecord) error {
	s.logger.Info("记录操作",
		zap.String("user_id", record.UserID),
		zap.String("operation", record.Operation),
	)

	return s.db.WithContext(ctx).Create(record).Error
}

// ClearHistory 清空历史
func (s *service) ClearHistory(ctx context.Context, userID string, historyType string) error {
	s.logger.Info("清空历史",
		zap.String("user_id", userID),
		zap.String("type", historyType),
	)

	switch historyType {
	case "download":
		return s.db.WithContext(ctx).
			Where("username = ?", userID).
			Delete(&database.DownloadHistory{}).Error
	case "operation":
		return s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Delete(&OperationRecord{}).Error
	default:
		// 清空所有
		s.db.WithContext(ctx).Where("username = ?", userID).Delete(&database.DownloadHistory{})
		return s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&OperationRecord{}).Error
	}
}

// GetHistoryStats 获取历史统计
func (s *service) GetHistoryStats(ctx context.Context, userID string) (*HistoryStats, error) {
	s.logger.Info("获取历史统计", zap.String("user_id", userID))

	stats := &HistoryStats{}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	// 总下载数
	s.db.WithContext(ctx).
		Model(&database.DownloadHistory{}).
		Where("username = ?", userID).
		Count(&stats.TotalDownloads)

	// 今日下载数
	s.db.WithContext(ctx).
		Model(&database.DownloadHistory{}).
		Where("username = ? AND created_at >= ?", userID, today).
		Count(&stats.TodayDownloads)

	// 本周下载数
	s.db.WithContext(ctx).
		Model(&database.DownloadHistory{}).
		Where("username = ? AND created_at >= ?", userID, weekAgo).
		Count(&stats.WeekDownloads)

	// 本月下载数
	s.db.WithContext(ctx).
		Model(&database.DownloadHistory{}).
		Where("username = ? AND created_at >= ?", userID, monthAgo).
		Count(&stats.MonthDownloads)

	// 总操作数
	s.db.WithContext(ctx).
		Model(&OperationRecord{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalOperations)

	// 今日操作数
	s.db.WithContext(ctx).
		Model(&OperationRecord{}).
		Where("user_id = ? AND created_at >= ?", userID, today).
		Count(&stats.TodayOperations)

	// 本周操作数
	s.db.WithContext(ctx).
		Model(&OperationRecord{}).
		Where("user_id = ? AND created_at >= ?", userID, weekAgo).
		Count(&stats.WeekOperations)

	// 本月操作数
	s.db.WithContext(ctx).
		Model(&OperationRecord{}).
		Where("user_id = ? AND created_at >= ?", userID, monthAgo).
		Count(&stats.MonthOperations)

	return stats, nil
}

// GetDownloadHistoryByPage 分页获取下载历史
func (s *service) GetDownloadHistoryByPage(ctx context.Context, page int, count int) ([]*database.DownloadHistory, error) {
	s.logger.Info("分页获取下载历史", zap.Int("page", page), zap.Int("count", count))

	if page <= 0 {
		page = 1
	}
	if count <= 0 {
		count = 30
	}

	offset := (page - 1) * count

	var histories []*database.DownloadHistory
	err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(count).
		Offset(offset).
		Find(&histories).Error

	return histories, err
}

// DeleteDownloadHistory 删除下载历史
func (s *service) DeleteDownloadHistory(ctx context.Context, historyID int) error {
	s.logger.Info("删除下载历史", zap.Int("history_id", historyID))

	return s.db.WithContext(ctx).
		Where("id = ?", historyID).
		Delete(&database.DownloadHistory{}).Error
}

// GetTransferHistory 分页获取整理记录
func (s *service) GetTransferHistory(ctx context.Context, title string, page int, count int, status *bool) ([]*database.TransferHistory, int64, error) {
	s.logger.Info("分页获取整理记录", zap.String("title", title), zap.Int("page", page), zap.Int("count", count), zap.Any("status", status))

	if page <= 0 {
		page = 1
	}
	if count <= 0 {
		count = 30
	}

	offset := (page - 1) * count
	query := s.db.WithContext(ctx).Model(&database.TransferHistory{})

	// 处理标题搜索
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}

	// 处理状态过滤
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	var histories []*database.TransferHistory
	err := query.Order("created_at DESC").
		Limit(count).
		Offset(offset).
		Find(&histories).Error

	return histories, total, err
}

// DeleteTransferHistory 删除整理记录
func (s *service) DeleteTransferHistory(ctx context.Context, historyID int, deleteSrc bool, deleteDest bool) error {
	s.logger.Info("删除整理记录", zap.Int("history_id", historyID), zap.Bool("delete_src", deleteSrc), zap.Bool("delete_dest", deleteDest))

	// TODO: 实现删除源文件和目标文件的逻辑
	// 当前仅实现删除数据库记录

	return s.db.WithContext(ctx).
		Where("id = ?", historyID).
		Delete(&database.TransferHistory{}).Error
}

// EmptyTransferHistory 清空整理记录
func (s *service) EmptyTransferHistory(ctx context.Context) error {
	s.logger.Info("清空整理记录")

	return s.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&database.TransferHistory{}).Error
}
