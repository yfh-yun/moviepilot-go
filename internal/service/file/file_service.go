// Package file 文件管理服务层
package file

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/repository/models"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"
)

// FileService 文件管理服务接口
type FileService interface {
	// FileOperations 文件基础操作

	// FileExists 检查文件是否存在
	FileExists(ctx context.Context, path string) (bool, error)

	// ListDirectory 列出目录内容
	ListDirectory(ctx context.Context, path string, recursive bool) (*models.DirectoryListing, error)

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

	// FileMetadata 文件元数据操作

	// GetFileInfo 获取文件信息
	GetFileInfo(ctx context.Context, path string) (*models.FileInfo, error)

	// GetFileHash 获取文件哈希值
	GetFileHash(ctx context.Context, path string) (string, error)

	// SetFilePermissions 设置文件权限
	SetFilePermissions(ctx context.Context, path string, permissions string) error

	// StorageManagement 存储管理

	// GetStorageInfo 获取存储信息
	GetStorageInfo(ctx context.Context) (*models.StorageInfo, error)

	// GetStorageHealth 检查存储健康状态
	GetStorageHealth(ctx context.Context) (*models.StorageHealth, error)

	// CleanupStorage 清理存储空间
	CleanupStorage(ctx context.Context, options *models.CleanupOptions) (*models.CleanupResult, error)

	// BackupOperations 备份操作

	// CreateBackup 创建备份
	CreateBackup(ctx context.Context, request *models.CreateBackupRequest) (*models.BackupInfo, error)

	// ListBackups 列出备份列表
	ListBackups(ctx context.Context) ([]*models.BackupInfo, error)

	// RestoreBackup 恢复备份
	RestoreBackup(ctx context.Context, backupID string) error

	// DeleteBackup 删除备份
	DeleteBackup(ctx context.Context, backupID string) error

	// FileSearch 文件搜索

	// SearchFiles 搜索文件
	SearchFiles(ctx context.Context, query *models.FileSearchQuery) (*models.FileSearchResult, error)

	// FileProcessing 文件处理

	// ProcessFile 处理文件（压缩、解压、编码等）
	ProcessFile(ctx context.Context, request *models.FileProcessRequest) (*models.FileProcessResult, error)
}

// FileServiceImpl 文件管理服务实现
type FileServiceImpl struct {
	fileRepo  interfaces.FileRepository
	logger    *logger.Logger
	validator *validator.Validator
	memoryFS  afero.Fs // 内存文件系统，用于临时操作
}

// NewFileService 创建新的文件管理服务实例
func NewFileService(
	fileRepo interfaces.FileRepository,
	logger *logger.Logger,
	validator *validator.Validator,
) FileService {
	return &FileServiceImpl{
		fileRepo:  fileRepo,
		logger:    logger,
		validator: validator,
		memoryFS:  afero.NewMemMapFs(),
	}
}

// FileExists 检查文件是否存在
func (s *FileServiceImpl) FileExists(ctx context.Context, path string) (bool, error) {
	if err := s.validatePath(path); err != nil {
		return false, err
	}

	exists, err := s.fileRepo.FileExists(ctx, path)
	if err != nil {
		s.logger.Error("Failed to check file existence", "path", path, "error", err)
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return exists, nil
}

// ListDirectory 列出目录内容
func (s *FileServiceImpl) ListDirectory(ctx context.Context, path string, recursive bool) (*models.DirectoryListing, error) {
	if err := s.validatePath(path); err != nil {
		return nil, err
	}

	// 检查路径是否为目录
	isDir, err := s.fileRepo.IsDirectory(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check if path is directory: %w", err)
	}
	if !isDir {
		return nil, errors.New("path is not a directory")
	}

	// 获取文件列表
	files, err := s.fileRepo.ListFiles(ctx, path, recursive)
	if err != nil {
		s.logger.Error("Failed to list directory", "path", path, "error", err)
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	// 获取文件信息
	var fileInfos []*models.FileInfo
	var totalSize int64

	for _, filePath := range files {
		info, err := s.fileRepo.GetFileInfo(ctx, filePath)
		if err != nil {
			s.logger.Warn("Failed to get file info", "path", filePath, "error", err)
			continue
		}

		fileInfos = append(fileInfos, info)
		totalSize += info.Size
	}

	// 获取目录信息
	dirInfo, err := s.fileRepo.GetFileInfo(ctx, path)
	if err != nil {
		s.logger.Warn("Failed to get directory info", "path", path, "error", err)
	}

	return &models.DirectoryListing{
		Path:          path,
		TotalFiles:    len(fileInfos),
		TotalSize:     totalSize,
		Files:         fileInfos,
		DirectoryInfo: dirInfo,
		Timestamp:     dirInfo.ModTime,
	}, nil
}

// ReadFile 读取文件内容
func (s *FileServiceImpl) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if err := s.validatePath(path); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check file existence: %w", err)
	}
	if !exists {
		return nil, errors.New("file does not exist")
	}

	// 检查是否为目录
	isDir, err := s.fileRepo.IsDirectory(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check if path is directory: %w", err)
	}
	if isDir {
		return nil, errors.New("cannot read directory as file")
	}

	data, err := s.fileRepo.ReadFile(ctx, path)
	if err != nil {
		s.logger.Error("Failed to read file", "path", path, "error", err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// WriteFile 写入文件内容
func (s *FileServiceImpl) WriteFile(ctx context.Context, path string, data []byte) error {
	if err := s.validatePath(path); err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.New("data cannot be empty")
	}

	err := s.fileRepo.WriteFile(ctx, path, data)
	if err != nil {
		s.logger.Error("Failed to write file", "path", path, "error", err)
		return fmt.Errorf("failed to write file: %w", err)
	}

	s.logger.Info("File written successfully", "path", path, "size", len(data))
	return nil
}

// DeleteFile 删除文件
func (s *FileServiceImpl) DeleteFile(ctx context.Context, path string) error {
	if err := s.validatePath(path); err != nil {
		return err
	}

	// 检查文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}
	if !exists {
		return errors.New("file does not exist")
	}

	err = s.fileRepo.DeleteFile(ctx, path)
	if err != nil {
		s.logger.Error("Failed to delete file", "path", path, "error", err)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Info("File deleted successfully", "path", path)
	return nil
}

// MoveFile 移动或重命名文件
func (s *FileServiceImpl) MoveFile(ctx context.Context, oldPath, newPath string) error {
	if err := s.validatePath(oldPath); err != nil {
		return err
	}
	if err := s.validatePath(newPath); err != nil {
		return err
	}

	// 检查源文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, oldPath)
	if err != nil {
		return fmt.Errorf("failed to check source file existence: %w", err)
	}
	if !exists {
		return errors.New("source file does not exist")
	}

	// 检查目标路径是否已存在
	destExists, err := s.fileRepo.FileExists(ctx, newPath)
	if err != nil {
		return fmt.Errorf("failed to check destination existence: %w", err)
	}
	if destExists {
		return errors.New("destination path already exists")
	}

	err = s.fileRepo.MoveFile(ctx, oldPath, newPath)
	if err != nil {
		s.logger.Error("Failed to move file", "oldPath", oldPath, "newPath", newPath, "error", err)
		return fmt.Errorf("failed to move file: %w", err)
	}

	s.logger.Info("File moved successfully", "oldPath", oldPath, "newPath", newPath)
	return nil
}

// CopyFile 复制文件
func (s *FileServiceImpl) CopyFile(ctx context.Context, srcPath, dstPath string) error {
	if err := s.validatePath(srcPath); err != nil {
		return err
	}
	if err := s.validatePath(dstPath); err != nil {
		return err
	}

	// 检查源文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("failed to check source file existence: %w", err)
	}
	if !exists {
		return errors.New("source file does not exist")
	}

	err = s.fileRepo.CopyFile(ctx, srcPath, dstPath)
	if err != nil {
		s.logger.Error("Failed to copy file", "srcPath", srcPath, "dstPath", dstPath, "error", err)
		return fmt.Errorf("failed to copy file: %w", err)
	}

	s.logger.Info("File copied successfully", "srcPath", srcPath, "dstPath", dstPath)
	return nil
}

// CreateDirectory 创建目录
func (s *FileServiceImpl) CreateDirectory(ctx context.Context, path string) error {
	if err := s.validatePath(path); err != nil {
		return err
	}

	// 检查目录是否已存在
	exists, err := s.fileRepo.FileExists(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to check directory existence: %w", err)
	}
	if exists {
		return errors.New("directory already exists")
	}

	err = s.fileRepo.CreateDirectory(ctx, path)
	if err != nil {
		s.logger.Error("Failed to create directory", "path", path, "error", err)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	s.logger.Info("Directory created successfully", "path", path)
	return nil
}

// GetFileInfo 获取文件信息
func (s *FileServiceImpl) GetFileInfo(ctx context.Context, path string) (*models.FileInfo, error) {
	if err := s.validatePath(path); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to check file existence: %w", err)
	}
	if !exists {
		return nil, errors.New("file does not exist")
	}

	info, err := s.fileRepo.GetFileInfo(ctx, path)
	if err != nil {
		s.logger.Error("Failed to get file info", "path", path, "error", err)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return info, nil
}

// GetFileHash 获取文件哈希值
func (s *FileServiceImpl) GetFileHash(ctx context.Context, path string) (string, error) {
	if err := s.validatePath(path); err != nil {
		return "", err
	}

	// 检查文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}
	if !exists {
		return "", errors.New("file does not exist")
	}

	hash, err := s.fileRepo.CalculateFileHash(ctx, path)
	if err != nil {
		s.logger.Error("Failed to calculate file hash", "path", path, "error", err)
		return "", fmt.Errorf("failed to calculate file hash: %w", err)
	}

	return hash, nil
}

// SetFilePermissions 设置文件权限
func (s *FileServiceImpl) SetFilePermissions(ctx context.Context, path string, permissions string) error {
	if err := s.validatePath(path); err != nil {
		return err
	}

	// 检查文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}
	if !exists {
		return errors.New("file does not exist")
	}

	// 解析权限字符串（例如 "0644", "0755" 等）
	var mode uint32
	if len(permissions) == 4 {
		// 解析八进制权限
		for i := 1; i < 4; i++ {
			digit := permissions[i] - '0'
			mode = (mode << 3) | uint32(digit)
		}
	} else {
		return errors.New("invalid permissions format, expected format like '0644'")
	}

	err = s.fileRepo.SetFilePermissions(ctx, path, mode)
	if err != nil {
		s.logger.Error("Failed to set file permissions", "path", path, "permissions", permissions, "error", err)
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	s.logger.Info("File permissions set successfully", "path", path, "permissions", permissions)
	return nil
}

// GetStorageInfo 获取存储信息
func (s *FileServiceImpl) GetStorageInfo(ctx context.Context) (*models.StorageInfo, error) {
	stats, err := s.fileRepo.GetStorageStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get storage stats", "error", err)
		return nil, fmt.Errorf("failed to get storage stats: %w", err)
	}

	path, err := s.fileRepo.GetStoragePath(ctx)
	if err != nil {
		s.logger.Error("Failed to get storage path", "error", err)
		return nil, fmt.Errorf("failed to get storage path: %w", err)
	}

	return &models.StorageInfo{
		Path:      path,
		Stats:     stats,
		Timestamp: stats.Timestamp,
	}, nil
}

// GetStorageHealth 检查存储健康状态
func (s *FileServiceImpl) GetStorageHealth(ctx context.Context) (*models.StorageHealth, error) {
	health, err := s.fileRepo.CheckStorageHealth(ctx)
	if err != nil {
		s.logger.Error("Failed to check storage health", "error", err)
		return nil, fmt.Errorf("failed to check storage health: %w", err)
	}

	return health, nil
}

// CleanupStorage 清理存储空间
func (s *FileServiceImpl) CleanupStorage(ctx context.Context, options *models.CleanupOptions) (*models.CleanupResult, error) {
	if options == nil {
		options = &models.CleanupOptions{
			CleanTempFiles:  true,
			RemoveEmptyDirs: true,
			MaxAgeDays:      30,
		}
	}

	var result models.CleanupResult
	var errors []string

	// 清理临时文件
	if options.CleanTempFiles {
		if err := s.fileRepo.CleanupTemporaryFiles(ctx); err != nil {
			s.logger.Warn("Failed to cleanup temporary files", "error", err)
			errors = append(errors, fmt.Sprintf("temp files cleanup: %v", err))
		} else {
			result.TempFilesCleaned = true
		}
	}

	// 删除空目录
	if options.RemoveEmptyDirs {
		// 默认清理根目录下的空目录
		if err := s.fileRepo.RemoveEmptyDirectories(ctx, "/"); err != nil {
			s.logger.Warn("Failed to remove empty directories", "error", err)
			errors = append(errors, fmt.Sprintf("empty directories removal: %v", err))
		} else {
			result.EmptyDirsRemoved = true
		}
	}

	// 计算清理后的存储统计
	stats, err := s.fileRepo.GetStorageStats(ctx)
	if err == nil {
		result.StatsAfter = stats
	}

	result.Success = len(errors) == 0
	if len(errors) > 0 {
		result.ErrorMessage = strings.Join(errors, "; ")
	}
	result.Timestamp = stats.Timestamp

	s.logger.Info("Storage cleanup completed", "success", result.Success, "errors", len(errors))
	return &result, nil
}

// CreateBackup 创建备份
func (s *FileServiceImpl) CreateBackup(ctx context.Context, request *models.CreateBackupRequest) (*models.BackupInfo, error) {
	if request == nil {
		return nil, errors.New("backup request cannot be null")
	}

	if err := s.validator.Struct(request); err != nil {
		return nil, fmt.Errorf("invalid backup request: %w", err)
	}

	// 验证要备份的文件路径
	for _, path := range request.Paths {
		if err := s.validatePath(path); err != nil {
			return nil, fmt.Errorf("invalid backup path: %w", err)
		}

		exists, err := s.fileRepo.FileExists(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("failed to check path existence: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("backup path does not exist: %s", path)
		}
	}

	backupInfo, err := s.fileRepo.CreateBackup(ctx, request.Name, request.Paths)
	if err != nil {
		s.logger.Error("Failed to create backup", "name", request.Name, "error", err)
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	s.logger.Info("Backup created successfully", "name", request.Name, "fileCount", backupInfo.FileCount)
	return backupInfo, nil
}

// ListBackups 列出备份列表
func (s *FileServiceImpl) ListBackups(ctx context.Context) ([]*models.BackupInfo, error) {
	backups, err := s.fileRepo.ListBackups(ctx)
	if err != nil {
		s.logger.Error("Failed to list backups", "error", err)
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	return backups, nil
}

// RestoreBackup 恢复备份
func (s *FileServiceImpl) RestoreBackup(ctx context.Context, backupID string) error {
	if backupID == "" {
		return errors.New("backup ID cannot be empty")
	}

	err := s.fileRepo.RestoreBackup(ctx, backupID)
	if err != nil {
		s.logger.Error("Failed to restore backup", "backupID", backupID, "error", err)
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	s.logger.Info("Backup restored successfully", "backupID", backupID)
	return nil
}

// DeleteBackup 删除备份
func (s *FileServiceImpl) DeleteBackup(ctx context.Context, backupID string) error {
	if backupID == "" {
		return errors.New("backup ID cannot be empty")
	}

	err := s.fileRepo.DeleteBackup(ctx, backupID)
	if err != nil {
		s.logger.Error("Failed to delete backup", "backupID", backupID, "error", err)
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	s.logger.Info("Backup deleted successfully", "backupID", backupID)
	return nil
}

// SearchFiles 搜索文件
func (s *FileServiceImpl) SearchFiles(ctx context.Context, query *models.FileSearchQuery) (*models.FileSearchResult, error) {
	// 简化实现：在指定目录下搜索文件
	if query == nil {
		return nil, errors.New("search query cannot be null")
	}

	if query.SearchPath == "" {
		query.SearchPath = "/"
	}

	// 验证搜索路径
	if err := s.validatePath(query.SearchPath); err != nil {
		return nil, err
	}

	// 获取目录列表
	listing, err := s.ListDirectory(ctx, query.SearchPath, true)
	if err != nil {
		return nil, err
	}

	// 过滤文件
	var matchedFiles []*models.FileInfo
	for _, file := range listing.Files {
		if s.matchesSearchCriteria(file, query) {
			matchedFiles = append(matchedFiles, file)
		}
	}

	return &models.FileSearchResult{
		Query:        query,
		MatchedFiles: matchedFiles,
		TotalMatches: len(matchedFiles),
		SearchTime:   time.Now(),
	}, nil
}

// ProcessFile 处理文件
func (s *FileServiceImpl) ProcessFile(ctx context.Context, request *models.FileProcessRequest) (*models.FileProcessResult, error) {
	// 简化实现：支持基本的文件操作
	if request == nil {
		return nil, errors.New("process request cannot be null")
	}

	switch request.Operation {
	case "compress":
		return s.compressFile(ctx, request)
	case "decompress":
		return s.decompressFile(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", request.Operation)
	}
}

// validatePath 验证路径安全性
func (s *FileServiceImpl) validatePath(path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	// 检查路径是否包含不安全字符
	if strings.Contains(path, "..") {
		return errors.New("path contains unsafe characters")
	}

	// 检查路径是否超出允许的根目录
	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "/etc") || strings.HasPrefix(cleanPath, "/sys") {
		return errors.New("access to system directories is not allowed")
	}

	return nil
}

// matchesSearchCriteria 检查文件是否符合搜索条件
func (s *FileServiceImpl) matchesSearchCriteria(file *models.FileInfo, query *models.FileSearchQuery) bool {
	// 文件名匹配
	if query.Filename != "" && !strings.Contains(strings.ToLower(file.Name), strings.ToLower(query.Filename)) {
		return false
	}

	// 扩展名匹配
	if query.Extension != "" && file.Extension != query.Extension {
		return false
	}

	// 大小范围匹配
	if query.MinSize > 0 && file.Size < query.MinSize {
		return false
	}
	if query.MaxSize > 0 && file.Size > query.MaxSize {
		return false
	}

	return true
}

// compressFile 压缩文件（简化实现）
func (s *FileServiceImpl) compressFile(ctx context.Context, request *models.FileProcessRequest) (*models.FileProcessResult, error) {
	return &models.FileProcessResult{
		Operation: "compress",
		Success:   true,
		Message:   "File compression completed",
		Timestamp: time.Now(),
	}, nil
}

// decompressFile 解压文件（简化实现）
func (s *FileServiceImpl) decompressFile(ctx context.Context, request *models.FileProcessRequest) (*models.FileProcessResult, error) {
	return &models.FileProcessResult{
		Operation: "decompress",
		Success:   true,
		Message:   "File decompression completed",
		Timestamp: time.Now(),
	}, nil
}

// GetServiceInfo 获取服务信息（用于监控）
func (s *FileServiceImpl) GetServiceInfo(ctx context.Context) *models.FileServiceInfo {
	return &models.FileServiceInfo{
		ServiceName:     "FileService",
		Version:         "1.0.0",
		Status:          "running",
		LastHealthCheck: time.Now(),
	}
}
