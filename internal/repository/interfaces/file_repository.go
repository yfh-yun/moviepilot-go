// Package interfaces 定义文件管理相关的仓储接口
package interfaces

import (
	"context"
	"github.com/yfh-yun/moviepilot-go/internal/model"
)

// FileRepository 定义文件管理相关的仓储操作接口
type FileRepository interface {
	// FileSystem 文件系统操作

	// FileExists 检查文件是否存在
	FileExists(ctx context.Context, path string) (bool, error)

	// IsDirectory 检查路径是否为目录
	IsDirectory(ctx context.Context, path string) (bool, error)

	// ListFiles 列出目录下的文件
	ListFiles(ctx context.Context, path string, recursive bool) ([]string, error)

	// GetFileInfo 获取文件信息
	GetFileInfo(ctx context.Context, path string) (*model.FileInfo, error)

	// ReadFile 读取文件内容
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// WriteFile 写入文件内容
	WriteFile(ctx context.Context, path string, data []byte) error

	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, path string) error

	// MoveFile 移动或重命名文件
	MoveFile(ctx context.Context, oldPath, newPath string) error

	// CopyFile 复制文件
	CopyFile(ctx context.Context, srcPath, dstPath string) error

	// CreateDirectory 创建目录
	CreateDirectory(ctx context.Context, path string) error

	// FileOperations 文件操作

	// CalculateFileHash 计算文件哈希
	CalculateFileHash(ctx context.Context, path string) (string, error)

	// GetFileSize 获取文件大小
	GetFileSize(ctx context.Context, path string) (int64, error)

	// SetFilePermissions 设置文件权限
	SetFilePermissions(ctx context.Context, path string, mode uint32) error

	// StorageManagement 存储管理

	// GetStorageStats 获取存储统计信息
	GetStorageStats(ctx context.Context) (*model.StorageStats, error)

	// GetStoragePath 获取存储路径信息
	GetStoragePath(ctx context.Context) (string, error)

	// CheckStorageHealth 检查存储健康状态
	CheckStorageHealth(ctx context.Context) (*model.StorageHealth, error)

	// BackupOperations 备份操作

	// CreateBackup 创建备份
	CreateBackup(ctx context.Context, name string, paths []string) (*model.BackupInfo, error)

	// ListBackups 列出备份列表
	ListBackups(ctx context.Context) ([]*model.BackupInfo, error)

	// RestoreBackup 恢复备份
	RestoreBackup(ctx context.Context, backupID string) error

	// DeleteBackup 删除备份
	DeleteBackup(ctx context.Context, backupID string) error

	// Cleanup 清理操作

	// CleanupTemporaryFiles 清理临时文件
	CleanupTemporaryFiles(ctx context.Context) error

	// RemoveEmptyDirectories 删除空目录
	RemoveEmptyDirectories(ctx context.Context, path string) error
}
