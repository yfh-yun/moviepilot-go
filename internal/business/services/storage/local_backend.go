// Package storage 实现本地存储后端
package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/utils"

	"go.uber.org/zap"
)

// LocalStorageBackend 本地存储后端
type LocalStorageBackend struct {
	config   *StorageBackendConfig
	logger   *zap.Logger
	rootPath string
}

// NewLocalStorageBackend 创建本地存储后端实例
func NewLocalStorageBackend(config *StorageBackendConfig) *LocalStorageBackend {
	rootPath := config.RootPath
	if rootPath == "" {
		rootPath = "./storage"
	}

	// 确保路径是绝对路径
	if !filepath.IsAbs(rootPath) {
		if absPath, err := filepath.Abs(rootPath); err == nil {
			rootPath = absPath
		}
	}

	return &LocalStorageBackend{
		config:   config,
		logger:   logger.Logger,
		rootPath: rootPath,
	}
}

// Init 初始化后端
func (lsb *LocalStorageBackend) Init(ctx context.Context) error {
	lsb.logger.Info("初始化本地存储后端", zap.String("root_path", lsb.rootPath))

	// 创建根目录
	if err := os.MkdirAll(lsb.rootPath, 0755); err != nil {
		return fmt.Errorf("创建根目录失败: %w", err)
	}

	// 测试目录是否可写
	testFile := filepath.Join(lsb.rootPath, ".test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("目录不可写: %w", err)
	}

	// 清理测试文件
	os.Remove(testFile)

	lsb.logger.Info("本地存储后端初始化成功")
	return nil
}

// Test 测试后端连接
func (lsb *LocalStorageBackend) Test(ctx context.Context) error {
	// 检查根目录是否存在
	if _, err := os.Stat(lsb.rootPath); os.IsNotExist(err) {
		return fmt.Errorf("根目录不存在: %s", lsb.rootPath)
	}

	// 检查读写权限
	testFile := filepath.Join(lsb.rootPath, ".test_connection")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("写入权限测试失败: %w", err)
	}

	// 测试读取
	if _, err := os.ReadFile(testFile); err != nil {
		return fmt.Errorf("读取权限测试失败: %w", err)
	}

	// 清理测试文件
	os.Remove(testFile)

	return nil
}

// GetInfo 获取存储信息
func (lsb *LocalStorageBackend) GetInfo() (*StorageInfo, error) {
	stat, err := os.Stat(lsb.rootPath)
	if err != nil {
		return nil, fmt.Errorf("获取根目录信息失败: %w", err)
	}

	// 获取磁盘使用情况
	usage, err := lsb.getDiskUsage()
	if err != nil {
		lsb.logger.Warn("获取磁盘使用情况失败", zap.Error(err))
		usage = &StorageUsage{Total: 0, Used: 0, Free: 0}
	}

	return &StorageInfo{
		Type:       "local",
		Name:       lsb.config.Name,
		Endpoint:   lsb.rootPath,
		TotalSpace: usage.Total,
		UsedSpace:  usage.Used,
		FreeSpace:  usage.Free,
		CreatedAt:  stat.ModTime(),
		IsWritable: lsb.isWritable(lsb.rootPath),
		Features:   []string{"file_operations", "directory_operations", "metadata", "permissions"},
	}, nil
}

// UploadFile 上传文件
func (lsb *LocalStorageBackend) UploadFile(ctx context.Context, req *UploadRequest) (*UploadResponse, error) {
	startTime := time.Now()

	// 构建完整文件路径
	fullPath := filepath.Join(lsb.rootPath, req.FilePath)

	lsb.logger.Debug("开始上传文件",
		zap.String("path", req.FilePath),
		zap.Int64("size", req.FileSize))

	// 检查目录是否存在，不存在则创建
	dir := filepath.Dir(fullPath)
	if req.CreateDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return &UploadResponse{
				Success:   false,
				Message:   fmt.Sprintf("创建目录失败: %v", err),
				ErrorCode: "CREATE_DIR_FAILED",
			}, nil
		}
	}

	// 检查文件是否已存在
	if _, err := os.Stat(fullPath); err == nil && !req.Overwrite {
		return &UploadResponse{
			Success:   false,
			Message:   "文件已存在且不允许覆盖",
			ErrorCode: "FILE_EXISTS",
		}, nil
	}

	// 创建目标文件
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return &UploadResponse{
			Success:   false,
			Message:   fmt.Sprintf("创建文件失败: %v", err),
			ErrorCode: "CREATE_FILE_FAILED",
		}, nil
	}
	defer file.Close()

	// 计算校验和
	hasher := sha256.New()
	multiWriter := io.MultiWriter(file, hasher)

	// 复制数据
	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 64 * 1024 // 默认64KB
	}

	written, err := io.CopyBuffer(multiWriter, req.DataReader, make([]byte, chunkSize))
	if err != nil {
		os.Remove(fullPath) // 删除部分写入的文件
		return &UploadResponse{
			Success:   false,
			Message:   fmt.Sprintf("写入文件失败: %v", err),
			ErrorCode: "WRITE_FILE_FAILED",
		}, nil
	}

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return &UploadResponse{
			Success:   false,
			Message:   fmt.Sprintf("获取文件信息失败: %v", err),
			ErrorCode: "GET_FILE_INFO_FAILED",
		}, nil
	}

	// 生成响应
	resp := &UploadResponse{
		Success:     true,
		FileID:      utils.GenerateID(),
		FilePath:    req.FilePath,
		FileName:    req.FileName,
		Size:        fileInfo.Size(),
		UploadedAt:  time.Now(),
		ETag:        fmt.Sprintf(`"%x"`, hasher.Sum(nil)),
		Checksum:    fmt.Sprintf("%x", hasher.Sum(nil)),
		ContentType: lsb.getContentType(req.FileName),
		Metadata:    req.Metadata,
		Message:     "上传成功",
	}

	lsb.logger.Info("文件上传完成",
		zap.String("path", resp.FilePath),
		zap.Int64("size", resp.Size),
		zap.Duration("duration", time.Since(startTime)))

	return resp, nil
}

// DownloadFile 下载文件
func (lsb *LocalStorageBackend) DownloadFile(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	// 构建完整文件路径
	fullPath := filepath.Join(lsb.rootPath, req.FilePath)

	lsb.logger.Debug("开始下载文件", zap.String("path", req.FilePath))

	// 检查文件是否存在
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return &DownloadResponse{
			Success:   false,
			Message:   "文件不存在",
			ErrorCode: "FILE_NOT_FOUND",
		}, nil
	}
	if err != nil {
		return &DownloadResponse{
			Success:   false,
			Message:   fmt.Sprintf("获取文件信息失败: %v", err),
			ErrorCode: "GET_FILE_INFO_FAILED",
		}, nil
	}

	// 检查是否为目录
	if fileInfo.IsDir() {
		return &DownloadResponse{
			Success:   false,
			Message:   "路径指向目录而非文件",
			ErrorCode: "IS_DIRECTORY",
		}, nil
	}

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		return &DownloadResponse{
			Success:   false,
			Message:   fmt.Sprintf("打开文件失败: %v", err),
			ErrorCode: "OPEN_FILE_FAILED",
		}, nil
	}

	// 处理范围请求
	if req.Range != nil {
		if req.Range.Start > 0 || req.Range.End > 0 {
			// 跳转到起始位置
			_, err = file.Seek(req.Range.Start, io.SeekStart)
			if err != nil {
				file.Close()
				return &DownloadResponse{
					Success:   false,
					Message:   fmt.Sprintf("文件定位失败: %v", err),
					ErrorCode: "SEEK_FAILED",
				}, nil
			}
		}
	}

	// 生成响应
	resp := &DownloadResponse{
		Success:     true,
		DataReader:  file,
		FilePath:    req.FilePath,
		FileName:    filepath.Base(fullPath),
		Size:        fileInfo.Size(),
		ContentType: lsb.getContentType(fullPath),
		ModifiedAt:  fileInfo.ModTime(),
		Message:     "下载成功",
	}

	return resp, nil
}

// DeleteFile 删除文件
func (lsb *LocalStorageBackend) DeleteFile(ctx context.Context, filePath string) error {
	fullPath := filepath.Join(lsb.rootPath, filePath)

	lsb.logger.Debug("删除文件", zap.String("path", filePath))

	// 检查文件是否存在
	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil // 文件不存在，视为删除成功
	}
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 根据类型删除
	if fileInfo.IsDir() {
		return os.RemoveAll(fullPath)
	}

	return os.Remove(fullPath)
}

// CopyFile 复制文件
func (lsb *LocalStorageBackend) CopyFile(ctx context.Context, srcPath, dstPath string) error {
	srcFullPath := filepath.Join(lsb.rootPath, srcPath)
	dstFullPath := filepath.Join(lsb.rootPath, dstPath)

	lsb.logger.Debug("复制文件",
		zap.String("src", srcPath),
		zap.String("dst", dstPath))

	// 确保目标目录存在
	dstDir := filepath.Dir(dstFullPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 执行复制
	return utils.CopyFile(srcFullPath, dstFullPath)
}

// MoveFile 移动文件
func (lsb *LocalStorageBackend) MoveFile(ctx context.Context, srcPath, dstPath string) error {
	srcFullPath := filepath.Join(lsb.rootPath, srcPath)
	dstFullPath := filepath.Join(lsb.rootPath, dstPath)

	lsb.logger.Debug("移动文件",
		zap.String("src", srcPath),
		zap.String("dst", dstPath))

	// 确保目标目录存在
	dstDir := filepath.Dir(dstFullPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 执行移动
	return utils.MoveFile(srcFullPath, dstFullPath)
}

// CreateDirectory 创建目录
func (lsb *LocalStorageBackend) CreateDirectory(ctx context.Context, dirPath string) error {
	fullPath := filepath.Join(lsb.rootPath, dirPath)

	lsb.logger.Debug("创建目录", zap.String("path", dirPath))

	return os.MkdirAll(fullPath, 0755)
}

// DeleteDirectory 删除目录
func (lsb *LocalStorageBackend) DeleteDirectory(ctx context.Context, dirPath string) error {
	fullPath := filepath.Join(lsb.rootPath, dirPath)

	lsb.logger.Debug("删除目录", zap.String("path", dirPath))

	return os.RemoveAll(fullPath)
}

// ListDirectory 列出目录
func (lsb *LocalStorageBackend) ListDirectory(ctx context.Context, dirPath string) ([]*FileStat, error) {
	fullPath := filepath.Join(lsb.rootPath, dirPath)

	lsb.logger.Debug("列出目录", zap.String("path", dirPath))

	// 读取目录
	entries, err := os.ReadDir(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", dirPath)
	}
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	var files []*FileStat
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			lsb.logger.Warn("获取文件信息失败",
				zap.String("name", entry.Name()),
				zap.Error(err))
			continue
		}

		fileStat := &FileStat{
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
			Mode:    info.Mode(),
		}

		// 计算相对路径
		relPath := filepath.Join(dirPath, entry.Name())
		fileStat.Path = relPath

		files = append(files, fileStat)
	}

	return files, nil
}

// GetFileInfo 获取文件信息
func (lsb *LocalStorageBackend) GetFileInfo(ctx context.Context, filePath string) (*FileStat, error) {
	fullPath := filepath.Join(lsb.rootPath, filePath)

	lsb.logger.Debug("获取文件信息", zap.String("path", filePath))

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	return &FileStat{
		Path:    filePath,
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
		Mode:    info.Mode(),
	}, nil
}

// GetDirectoryInfo 获取目录信息
func (lsb *LocalStorageBackend) GetDirectoryInfo(ctx context.Context, dirPath string) (*DirectoryInfo, error) {
	fullPath := filepath.Join(lsb.rootPath, dirPath)

	lsb.logger.Debug("获取目录信息", zap.String("path", dirPath))

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", dirPath)
	}
	if err != nil {
		return nil, fmt.Errorf("获取目录信息失败: %w", err)
	}

	// 统计目录内容
	var fileCount, dirCount int64
	var totalSize int64

	filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == fullPath {
			return nil // 跳过根目录本身
		}

		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
			totalSize += info.Size()
		}

		return nil
	})

	return &DirectoryInfo{
		Path:       dirPath,
		Name:       info.Name(),
		ModTime:    info.ModTime(),
		FileCount:  fileCount,
		DirCount:   dirCount,
		TotalSize:  totalSize,
		IsWritable: lsb.isWritable(fullPath),
	}, nil
}

// GetStorageUsage 获取存储使用情况
func (lsb *LocalStorageBackend) GetStorageUsage(ctx context.Context) (*StorageUsage, error) {
	return lsb.getDiskUsage()
}

// SyncDirectory 同步目录（本地存储暂不支持）
func (lsb *LocalStorageBackend) SyncDirectory(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	return &SyncResponse{
		Success: false,
		Message: "本地存储不支持目录同步",
	}, nil
}

// SearchFiles 搜索文件
func (lsb *LocalStorageBackend) SearchFiles(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	lsb.logger.Debug("搜索文件",
		zap.String("pattern", req.Pattern),
		zap.String("path", req.Path))

	var results []*FileStat
	searchPath := filepath.Join(lsb.rootPath, req.Path)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// 跳过目录本身
		if path == searchPath {
			return nil
		}

		// 检查文件名是否匹配模式
		matched, _ := filepath.Match(req.Pattern, info.Name())
		if !matched {
			return nil
		}

		// 获取相对路径
		relPath, _ := filepath.Rel(lsb.rootPath, path)

		result := &FileStat{
			Path:    relPath,
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
			Mode:    info.Mode(),
		}

		results = append(results, result)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("搜索文件失败: %w", err)
	}

	return &SearchResponse{
		Success: true,
		Files:   results,
		Total:   len(results),
		Message: "搜索完成",
	}, nil
}

// CleanupTempFiles 清理临时文件
func (lsb *LocalStorageBackend) CleanupTempFiles(ctx context.Context) error {
	lsb.logger.Info("清理临时文件")

	tempDir := filepath.Join(lsb.rootPath, "temp")
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		return nil // 临时目录不存在
	}

	// 清理超过1天的临时文件
	cutoff := time.Now().Add(-24 * time.Hour)

	return filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.ModTime.Before(cutoff) {
			os.Remove(path)
			lsb.logger.Debug("删除过期临时文件", zap.String("path", path))
		}

		return nil
	})
}

// OptimizeStorage 优化存储（本地存储暂不支持）
func (lsb *LocalStorageBackend) OptimizeStorage(ctx context.Context) error {
	lsb.logger.Info("存储优化", zap.String("message", "本地存储无需优化"))
	return nil
}

// 辅助方法

// getDiskUsage 获取磁盘使用情况
func (lsb *LocalStorageBackend) getDiskUsage() (*StorageUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(lsb.rootPath, &stat)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘信息失败: %w", err)
	}

	// 计算磁盘空间
	blockSize := uint64(stat.Bsize)
	totalSpace := uint64(stat.Blocks) * blockSize
	freeSpace := uint64(stat.Bavail) * blockSize
	usedSpace := totalSpace - freeSpace

	return &StorageUsage{
		Total: int64(totalSpace),
		Used:  int64(usedSpace),
		Free:  int64(freeSpace),
		Usage: float64(usedSpace) / float64(totalSpace) * 100,
	}, nil
}

// isWritable 检查目录是否可写
func (lsb *LocalStorageBackend) isWritable(dirPath string) bool {
	testFile := filepath.Join(dirPath, ".writable_test")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		return false
	}
	os.Remove(testFile)
	return true
}

// getContentType 获取文件内容类型
func (lsb *LocalStorageBackend) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}
