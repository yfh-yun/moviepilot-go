// Package repositories 文件管理仓储实现
package repositories

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"gorm.io/gorm"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/model"
)

// FileRepositoryImpl 文件管理仓储实现
type FileRepositoryImpl struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewFileRepository 创建新的文件管理仓储实例
func NewFileRepository(db *gorm.DB, logger *logger.Logger) interfaces.FileRepository {
	return &model.FileRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// FileExists 检查文件是否存在
func (r *FileRepositoryImpl) FileExists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsDirectory 检查路径是否为目录
func (r *FileRepositoryImpl) IsDirectory(ctx context.Context, path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// ListFiles 列出目录下的文件
func (r *FileRepositoryImpl) ListFiles(ctx context.Context, path string, recursive bool) ([]string, error) {
	if recursive {
		return r.listFilesRecursive(ctx, path)
	}
	return r.listFilesSimple(ctx, path)
}

// listFilesSimple 简单列出目录下的文件
func (r *FileRepositoryImpl) listFilesSimple(ctx context.Context, path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, filepath.Join(path, entry.Name()))
	}
	return files, nil
}

// listFilesRecursive 递归列出目录下的文件
func (r *FileRepositoryImpl) listFilesRecursive(ctx context.Context, path string) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, filePath)
		}
		return nil
	})

	return files, err
}

// GetFileInfo 获取文件信息
func (r *FileRepositoryImpl) GetFileInfo(ctx context.Context, path string) (*model.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	sysInfo := info.Sys()
	var uid, gid uint32
	var mode uint32

	if sysInfo != nil {
		if stat, ok := sysInfo.(*syscall.Stat_t); ok {
			uid = stat.Uid
			gid = stat.Gid
			mode = uint32(stat.Mode)
		}
	}

	return &model.FileInfo{
		Name:      info.Name(),
		Path:      path,
		Size:      info.Size(),
		Mode:      mode,
		ModTime:   info.ModTime(),
		IsDir:     info.IsDir(),
		UID:       uid,
		GID:       gid,
		Extension: filepath.Ext(path),
	}, nil
}

// ReadFile 读取文件内容
func (r *FileRepositoryImpl) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile 写入文件内容
func (r *FileRepositoryImpl) WriteFile(ctx context.Context, path string, data []byte) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// DeleteFile 删除文件
func (r *FileRepositoryImpl) DeleteFile(ctx context.Context, path string) error {
	return os.Remove(path)
}

// MoveFile 移动或重命名文件
func (r *FileRepositoryImpl) MoveFile(ctx context.Context, oldPath, newPath string) error {
	// 确保目标目录存在
	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.Rename(oldPath, newPath)
}

// CopyFile 复制文件
func (r *FileRepositoryImpl) CopyFile(ctx context.Context, srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	// 确保目标目录存在
	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// CreateDirectory 创建目录
func (r *FileRepositoryImpl) CreateDirectory(ctx context.Context, path string) error {
	return os.MkdirAll(path, 0755)
}

// CalculateFileHash 计算文件哈希
func (r *FileRepositoryImpl) CalculateFileHash(ctx context.Context, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetFileSize 获取文件大小
func (r *FileRepositoryImpl) GetFileSize(ctx context.Context, path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// SetFilePermissions 设置文件权限
func (r *FileRepositoryImpl) SetFilePermissions(ctx context.Context, path string, mode uint32) error {
	return os.Chmod(path, os.FileMode(mode))
}

// GetStorageStats 获取存储统计信息
func (r *FileRepositoryImpl) GetStorageStats(ctx context.Context) (*model.StorageStats, error) {
	usage, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &model.StorageStats{
		Total:       usage.Total,
		Free:        usage.Free,
		Used:        usage.Used,
		UsedPercent: usage.UsedPercent,
		InodesTotal: usage.InodesTotal,
		InodesFree:  usage.InodesFree,
		InodesUsed:  usage.InodesUsed,
		Timestamp:   time.Now(),
	}, nil
}

// GetStoragePath 获取存储路径信息
func (r *FileRepositoryImpl) GetStoragePath(ctx context.Context) (string, error) {
	return os.Getwd()
}

// CheckStorageHealth 检查存储健康状态
func (r *FileRepositoryImpl) CheckStorageHealth(ctx context.Context) (*model.StorageHealth, error) {
	stats, err := r.GetStorageStats(ctx)
	if err != nil {
		return nil, err
	}

	health := &model.StorageHealth{
		Stats:       stats,
		IsHealthy:   stats.UsedPercent < 90,
		IsWriteable: true,
		Timestamp:   time.Now(),
	}

	// 测试写入权限
	testFile := "/tmp/storage_health_test"
	if err := r.WriteFile(ctx, testFile, []byte("test")); err != nil {
		health.IsWriteable = false
	} else {
		_ = r.DeleteFile(ctx, testFile)
	}

	return health, nil
}

// CreateBackup 创建备份
func (r *FileRepositoryImpl) CreateBackup(ctx context.Context, name string, paths []string) (*model.BackupInfo, error) {
	backupDir := "/backup/" + name
	if err := r.CreateDirectory(ctx, backupDir); err != nil {
		return nil, err
	}

	var backedUpFiles []string
	for _, path := range paths {
		dstPath := filepath.Join(backupDir, filepath.Base(path))
		if err := r.CopyFile(ctx, path, dstPath); err != nil {
			r.logger.Warn("Failed to backup file", "path", path, "error", err)
			continue
		}
		backedUpFiles = append(backedUpFiles, dstPath)
	}

	return &model.BackupInfo{
		ID:        name,
		Name:      name,
		Path:      backupDir,
		FileCount: len(backedUpFiles),
		TotalSize: r.calculateTotalSize(ctx, backedUpFiles),
		CreatedAt: time.Now(),
		Status:    "completed",
	}, nil
}

// ListBackups 列出备份列表
func (r *FileRepositoryImpl) ListBackups(ctx context.Context) ([]*model.BackupInfo, error) {
	backupDir := "/backup"
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.BackupInfo{}, nil
		}
		return nil, err
	}

	var backups []*model.BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			backupPath := filepath.Join(backupDir, entry.Name())
			files, _ := r.ListFiles(ctx, backupPath, true)

			info, _ := entry.Info()
			backups = append(backups, &model.BackupInfo{
				ID:        entry.Name(),
				Name:      entry.Name(),
				Path:      backupPath,
				FileCount: len(files),
				CreatedAt: info.ModTime(),
				Status:    "completed",
			})
		}
	}

	return backups, nil
}

// RestoreBackup 恢复备份
func (r *FileRepositoryImpl) RestoreBackup(ctx context.Context, backupID string) error {
	backupDir := "/backup/" + backupID
	files, err := r.ListFiles(ctx, backupDir, false)
	if err != nil {
		return err
	}

	for _, file := range files {
		dstPath := "/" + filepath.Base(file)
		if err := r.CopyFile(ctx, file, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// DeleteBackup 删除备份
func (r *FileRepositoryImpl) DeleteBackup(ctx context.Context, backupID string) error {
	backupDir := "/backup/" + backupID
	return os.RemoveAll(backupDir)
}

// CleanupTemporaryFiles 清理临时文件
func (r *FileRepositoryImpl) CleanupTemporaryFiles(ctx context.Context) error {
	tempDir := "/tmp"
	files, err := r.ListFiles(ctx, tempDir, false)
	if err != nil {
		return err
	}

	var cleanedCount int
	for _, file := range files {
		// 清理超过1天的临时文件
		info, err := r.GetFileInfo(ctx, file)
		if err != nil {
			continue
		}

		if time.Since(info.ModTime) > 24*time.Hour {
			if err := r.DeleteFile(ctx, file); err == nil {
				cleanedCount++
			}
		}
	}

	r.logger.Info("Cleaned temporary files", "count", cleanedCount)
	return nil
}

// RemoveEmptyDirectories 删除空目录
func (r *FileRepositoryImpl) RemoveEmptyDirectories(ctx context.Context, path string) error {
	return filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			entries, err := os.ReadDir(currentPath)
			if err != nil {
				return nil
			}

			if len(entries) == 0 {
				return os.Remove(currentPath)
			}
		}

		return nil
	})
}

// calculateTotalSize 计算文件总大小
func (r *FileRepositoryImpl) calculateTotalSize(ctx context.Context, files []string) int64 {
	var total int64
	for _, file := range files {
		if size, err := r.GetFileSize(ctx, file); err == nil {
			total += size
		}
	}
	return total
}
