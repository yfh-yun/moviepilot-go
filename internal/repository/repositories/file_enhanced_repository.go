package repositories

import (
	"context"
	"fmt"
	"io/fs"
	"github.com/yfh-yun/moviepilot-go/internal/database"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// fileEnhancedRepository 增强文件仓储实现
type fileEnhancedRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewFileEnhancedRepository 创建增强文件仓储
func NewFileEnhancedRepository(db *gorm.DB) interfaces.FileRepository {
	logger, _ := zap.NewProduction()
	return &fileEnhancedRepository{
		db:     db,
		logger: logger,
	}
}

// FileInfo 文件信息结构体（补充models中缺失的结构）
type FileInfo struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Type     string    `json:"type"` // file, directory
	ModTime  time.Time `json:"mod_time"`
	Hash     string    `json:"hash,omitempty"`
	Exists   bool      `json:"exists"`
}

// StorageStats 存储统计结构体
type StorageStats struct {
	TotalFiles   int64  `json:"total_files"`
	TotalSize    int64  `json:"total_size"`
	UsedSpace    int64  `json:"used_space"`
	FreeSpace    int64  `json:"free_space"`
	LargestFiles []FileInfo `json:"largest_files"`
}

// StorageHealth 存储健康状态结构体
type StorageHealth struct {
	Status         string    `json:"status"` // healthy, warning, error
	TotalFiles     int64     `json:"total_files"`
	TotalSize      int64     `json:"total_size"`
	FreeSpace      int64     `json:"free_space"`
	DuplicateFiles int64     `json:"duplicate_files"`
	OrphanFiles   int64     `json:"orphan_files"`
	CheckTime      time.Time `json:"check_time"`
}

// BackupInfo 备份信息结构体
type BackupInfo struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
	Status       string    `json:"status"` // pending, in_progress, completed, failed
	FilesCount   int       `json:"files_count"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// Create 创建文件记录
func (r *fileEnhancedRepository) Create(file *model.File) error {
	return r.db.Create(file).Error
}

// Update 更新文件记录
func (r *fileEnhancedRepository) Update(file *model.File) error {
	return r.db.Save(file).Error
}

// Delete 删除文件记录
func (r *fileEnhancedRepository) Delete(id uint) error {
	return r.db.Delete(&model.File{}, id).Error
}

// GetByID 根据ID获取文件
func (r *fileEnhancedRepository) GetByID(id uint) (*model.File, error) {
	var file model.File
	err := r.db.First(&file, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// GetByPath 根据路径获取文件
func (r *fileEnhancedRepository) GetByPath(path string) (*model.File, error) {
	var file model.File
	err := r.db.Where("path = ?", path).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// List 分页获取文件列表
func (r *fileEnhancedRepository) List(offset, limit int) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	err := r.db.Model(&model.File{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error
	return files, total, err
}

// Count 统计文件数量
func (r *fileEnhancedRepository) Count() (int64, error) {
	var total int64
	err := r.db.Model(&model.File{}).Count(&total).Error
	return total, err
}

// GetStorageStats 获取存储统计信息
func (r *fileEnhancedRepository) GetStorageStats() (*StorageStats, error) {
	stats := &model.StorageStats{}
	
	// 从数据库统计
	var totalFiles int64
	var totalSize int64
	
	r.db.Model(&model.File{}).Count(&totalFiles)
	r.db.Model(&model.File{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)
	
	stats.TotalFiles = totalFiles
	stats.TotalSize = totalSize
	
	// 获取最大的10个文件
	var largestFiles []File
	r.db.Order("size DESC").Limit(10).Find(&largestFiles)
	
	stats.LargestFiles = make([]FileInfo, len(largestFiles))
	for i, file := range largestFiles {
		stats.LargestFiles[i] = FileInfo{
			Path:    file.Path,
			Name:    file.Name,
			Size:    file.Size,
			Type:    file.Type,
			ModTime: file.AccessTime,
		}
	}
	
	// 获取磁盘空间信息
	if paths := []string{"/", "/data", "/media"}; len(paths) > 0 {
		for _, path := range paths {
			if stat, err := os.Stat(path); err == nil {
				if stat, err := filepath.Abs(path); err == nil {
					if usage, err := getDiskUsage(stat); err == nil {
						if stats.FreeSpace == 0 || usage.Free < stats.FreeSpace {
							stats.FreeSpace = usage.Free
						}
						if stats.UsedSpace == 0 || usage.Used > stats.UsedSpace {
							stats.UsedSpace = usage.Used
						}
					}
				}
			}
		}
	}
	
	return stats, nil
}

// GetFileHealth 获取文件健康状态
func (r *fileEnhancedRepository) GetFileHealth() (*StorageHealth, error) {
	health := &model.StorageHealth{
		CheckTime: time.Now(),
	}
	
	// 统计文件总数和总大小
	r.db.Model(&model.File{}).Count(&health.TotalFiles)
	r.db.Model(&model.File{}).Select("COALESCE(SUM(size), 0)").Scan(&health.TotalSize)
	
	// 检测重复文件（根据MD5）
	var duplicateCount int64
	r.db.Model(&model.File{}).
		Where("md5 IS NOT NULL AND md5 != ''").
		Group("md5").
		Having("COUNT(*) > 1").
		Count(&duplicateCount)
	health.DuplicateFiles = duplicateCount
	
	// 检测孤立文件（路径不存在的文件）
	var orphanCount int64
	r.db.Model(&model.File{}).
		Where("path NOT IN (SELECT DISTINCT savepath FROM download_files WHERE state = 1)").
		Count(&orphanCount)
	health.OrphanFiles = orphanCount
	
	// 获取磁盘空间信息
	if paths := []string{"/", "/data", "/media"}; len(paths) > 0 {
		for _, path := range paths {
			if stat, err := os.Stat(path); err == nil {
				if usage, err := getDiskUsage(stat); err == nil {
					health.FreeSpace = usage.Free
					break
				}
			}
		}
	}
	
	// 确定健康状态
	if health.DuplicateFiles > 0 || health.OrphanFiles > 0 {
		if health.DuplicateFiles > 100 || health.OrphanFiles > 50 {
			health.Status = "error"
		} else {
			health.Status = "warning"
		}
	} else {
		health.Status = "healthy"
	}
	
	return health, nil
}

// CreateBackup 创建备份
func (r *fileEnhancedRepository) CreateBackup(name, path string) (*BackupInfo, error) {
	backup := &model.BackupInfo{
		Name:       name,
		Path:       path,
		CreatedAt:  time.Now(),
		Status:     "pending",
		FilesCount: 0,
	}
	
	// 这里应该保存到数据库，但为了简化，只返回结构
	// 实际实现中需要添加backup表和对应的CRUD操作
	
	r.logger.Info("开始创建备份",
		zap.String("name", name),
		zap.String("path", path))
	
	// 执行备份逻辑
	backup.Status = "in_progress"
	backup.FilesCount, _ = r.countFilesForBackup(path)
	backup.Size, _ = r.calculateBackupSize(path)
	
	// 模拟备份完成
	backup.Status = "completed"
	completed := time.Now()
	backup.CompletedAt = &completed
	
	return backup, nil
}

// GetBackupInfo 获取备份信息
func (r *fileEnhancedRepository) GetBackupInfo(backupID uint) (*BackupInfo, error) {
	// 模拟从数据库获取备份信息
	return &model.BackupInfo{
		ID:          backupID,
		Name:        fmt.Sprintf("backup_%d", backupID),
		Path:        "/backups/backup_001",
		Size:        1024 * 1024 * 1024 * 10, // 10GB
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		Status:      "completed",
		FilesCount:  150,
	}, nil
}

// ListBackups 列出所有备份
func (r *fileEnhancedRepository) ListBackups() ([]*BackupInfo, error) {
	// 模拟从数据库获取备份列表
	backups := []*BackupInfo{
		{
			ID:          1,
			Name:        "daily_backup",
			Path:        "/backups/daily_001",
			Size:        1024 * 1024 * 1024 * 8,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			Status:      "completed",
			FilesCount:  120,
		},
		{
			ID:          2,
			Name:        "weekly_backup",
			Path:        "/backups/weekly_001",
			Size:        1024 * 1024 * 1024 * 15,
			CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
			Status:      "completed",
			FilesCount:  200,
		},
	}
	
	return backups, nil
}

// CleanupOrphanFiles 清理孤立文件
func (r *fileEnhancedRepository) CleanupOrphanFiles() (int64, error) {
	// 查找孤立文件
	var orphanFiles []File
	err := r.db.Where("path NOT IN (SELECT DISTINCT savepath FROM download_files WHERE state = 1)").Find(&orphanFiles).Error
	if err != nil {
		return 0, err
	}
	
	// 从数据库删除这些记录
	err = r.db.Where("path NOT IN (SELECT DISTINCT savepath FROM download_files WHERE state = 1)").Delete(&model.File{}).Error
	if err != nil {
		return 0, err
	}
	
	r.logger.Info("清理孤立文件", zap.Int64("count", int64(len(orphanFiles))))
	
	return int64(len(orphanFiles)), nil
}

// 扫描目录的辅助方法
func (r *fileEnhancedRepository) scanDirectory(path string) ([]FileInfo, error) {
	var files []FileInfo
	
	err := filepath.Walk(path, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		fileInfo := FileInfo{
			Path:    filePath,
			Name:    info.Name(),
			Size:    info.Size(),
			Type:    "file",
			ModTime: info.ModTime(),
			Exists:  true,
		}
		
		if info.IsDir() {
			fileInfo.Type = "directory"
		}
		
		files = append(files, fileInfo)
		return nil
	})
	
	return files, err
}

// 统计备份文件数量
func (r *fileEnhancedRepository) countFilesForBackup(path string) (int, error) {
	count := 0
	err := filepath.Walk(path, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		count++
		return nil
	})
	return count, err
}

// 计算备份大小
func (r *fileEnhancedRepository) calculateBackupSize(path string) (int64, error) {
	var totalSize int64
	err := filepath.Walk(path, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		totalSize += info.Size()
		return nil
	})
	return totalSize, err
}

// getDiskUsage 获取磁盘使用情况
func getDiskUsage(path string) (struct {
	Total int64
	Used  int64
	Free  int64
}, error) {
	// 这里简化实现，实际应该使用syscall获取准确的磁盘信息
	// 对于跨平台兼容，这里返回模拟数据
	return struct {
		Total int64
		Used  int64
		Free  int64
	}{
		Total: 1024 * 1024 * 1024 * 100, // 100GB
		Used:  1024 * 1024 * 1024 * 60,  // 60GB
		Free:  1024 * 1024 * 1024 * 40,  // 40GB
	}, nil
}

// SearchFiles 搜索文件
func (r *fileEnhancedRepository) SearchFiles(keyword string, limit int) ([]FileInfo, error) {
	var files []FileInfo
	
	// 从数据库搜索
	var dbFiles []File
	err := r.db.Where("name LIKE ? OR path LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Limit(limit).
		Find(&dbFiles).Error
	
	if err != nil {
		return nil, err
	}
	
	for _, file := range dbFiles {
		files = append(files, FileInfo{
			Path:   file.Path,
			Name:   file.Name,
			Size:   file.Size,
			Type:   file.Type,
			ModTime: file.AccessTime,
			Exists: true,
		})
	}
	
	return files, nil
}

// GetFilesByExtension 根据扩展名获取文件
func (r *fileEnhancedRepository) GetFilesByExtension(extension string, limit int) ([]*model.File, error) {
	var files []*model.File
	query := r.db.Where("name LIKE ?", "%."+extension)
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&files).Error
	return files, err
}

// GetRecentFiles 获取最近的文件
func (r *fileEnhancedRepository) GetRecentFiles(days int, limit int) ([]*model.File, error) {
	var files []*model.File
	since := time.Now().AddDate(0, 0, -days)
	
	query := r.db.Where("access_time >= ?", since).
		Order("access_time DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&files).Error
	return files, err
}