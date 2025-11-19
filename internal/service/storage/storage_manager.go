// Package storage 提供存储管理服务
package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/model"
	"github.com/yfh-yun/moviepilot-go/internal/repository/interfaces"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"

	"go.uber.org/zap"
)

// StorageManager 存储管理器
type StorageManager struct {
	storageRepo interfaces.StorageRepository
	mediaRepo   interfaces.MediaRepository
	logger      *zap.Logger

	// 存储后端映射
	storageBackends map[string]StorageBackend

	// 缓存
	cache *StorageCache

	// 配置
	config *StorageConfig
}

// StorageBackend 存储后端接口
type StorageBackend interface {
	// 基础操作
	Init(ctx context.Context) error
	Test(ctx context.Context) error
	GetInfo() (*StorageInfo, error)

	// 文件操作
	UploadFile(ctx context.Context, req *UploadRequest) (*UploadResponse, error)
	DownloadFile(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error)
	DeleteFile(ctx context.Context, filePath string) error
	CopyFile(ctx context.Context, srcPath, dstPath string) error
	MoveFile(ctx context.Context, srcPath, dstPath string) error

	// 目录操作
	CreateDirectory(ctx context.Context, dirPath string) error
	DeleteDirectory(ctx context.Context, dirPath string) error
	ListDirectory(ctx context.Context, dirPath string) ([]*FileStat, error)

	// 信息查询
	GetFileInfo(ctx context.Context, filePath string) (*FileStat, error)
	GetDirectoryInfo(ctx context.Context, dirPath string) (*DirectoryInfo, error)
	GetStorageUsage(ctx context.Context) (*StorageUsage, error)

	// 高级功能
	SyncDirectory(ctx context.Context, req *SyncRequest) (*SyncResponse, error)
	SearchFiles(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

	// 清理和维护
	CleanupTempFiles(ctx context.Context) error
	OptimizeStorage(ctx context.Context) error
}

// StorageConfig 存储配置
type StorageConfig struct {
	// 默认存储
	DefaultStorage string `json:"default_storage"`

	// 存储后端配置
	StorageBackends map[string]*StorageBackendConfig `json:"storage_backends"`

	// 传输配置
	ChunkSize      int64         `json:"chunk_size"`      // 分块大小 (bytes)
	MaxConcurrency int           `json:"max_concurrency"` // 最大并发数
	Timeout        time.Duration `json:"timeout"`         // 操作超时
	RetryCount     int           `json:"retry_count"`     // 重试次数
	RetryInterval  time.Duration `json:"retry_interval"`  // 重试间隔

	// 缓存配置
	EnableCache bool          `json:"enable_cache"`
	CacheSize   int64         `json:"cache_size"`   // 缓存大小
	CacheExpire time.Duration `json:"cache_expire"` // 缓存过期时间

	// 同步配置
	AutoSync      bool          `json:"auto_sync"`       // 自动同步
	SyncInterval  time.Duration `json:"sync_interval"`   // 同步间隔
	SyncOnStartup bool          `json:"sync_on_startup"` // 启动时同步

	// 备份配置
	EnableBackup    bool          `json:"enable_backup"`
	BackupInterval  time.Duration `json:"backup_interval"`  // 备份间隔
	BackupRetention int           `json:"backup_retention"` // 备份保留天数

	// 安全配置
	Encryption    bool `json:"encryption"`     // 加密存储
	Compression   bool `json:"compression"`    // 压缩存储
	AccessControl bool `json:"access_control"` // 访问控制
}

// StorageBackendConfig 存储后端配置
type StorageBackendConfig struct {
	Type    string `json:"type"` // local, ftp, s3, webdav
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`

	// 连接配置
	Endpoint string `json:"endpoint"`
	Region   string `json:"region"`
	Bucket   string `json:"bucket"`

	// 认证配置
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	KeyID    string `json:"key_id"`
	Secret   string `json:"secret"`

	// 高级配置
	RootPath string            `json:"root_path"`
	TLS      bool              `json:"tls"`
	Debug    bool              `json:"debug"`
	Options  map[string]string `json:"options"`
}

// UploadRequest 上传请求
type UploadRequest struct {
	FilePath      string            `json:"file_path"`      // 文件路径
	FileName      string            `json:"file_name"`      // 文件名
	FileSize      int64             `json:"file_size"`      // 文件大小
	DataReader    io.Reader         `json:"-"`              // 数据读取器
	ContentType   string            `json:"content_type"`   // 内容类型
	Compression   bool              `json:"compression"`    // 是否压缩
	Encryption    bool              `json:"encryption"`     // 是否加密
	ChunkSize     int64             `json:"chunk_size"`     // 分块大小
	Overwrite     bool              `json:"overwrite"`      // 是否覆盖
	CreateDirs    bool              `json:"create_dirs"`    // 是否创建目录
	Metadata      map[string]string `json:"metadata"`       // 元数据
	Tags          []string          `json:"tags"`           // 标签
	AccessControl []string          `json:"access_control"` // 访问控制
}

// UploadResponse 上传响应
type UploadResponse struct {
	Success    bool              `json:"success"`
	FileID     string            `json:"file_id"`       // 文件ID
	FilePath   string            `json:"file_path"`     // 文件路径
	ETag       string            `json:"etag"`          // 文件ETag
	Checksum   string            `json:"checksum"`      // 文件校验和
	Size       int64             `json:"size"`          // 文件大小
	UploadedAt time.Time         `json:"uploaded_at"`   // 上传时间
	URL        string            `json:"url,omitempty"` // 访问URL
	Metadata   map[string]string `json:"metadata"`      // 元数据
	Message    string            `json:"message"`       // 消息
	ErrorCode  string            `json:"error_code"`    // 错误代码
}

// DownloadRequest 下载请求
type DownloadRequest struct {
	FilePath     string     `json:"file_path"`       // 文件路径
	Range        *FileRange `json:"range,omitempty"` // 下载范围
	Encryption   bool       `json:"encryption"`      // 是否解密
	Decompress   bool       `json:"decompress"`      // 是否解压
	Stream       bool       `json:"stream"`          // 是否流式下载
	CacheControl string     `json:"cache_control"`   // 缓存控制
}

// DownloadResponse 下载响应
type DownloadResponse struct {
	Success     bool              `json:"success"`
	DataReader  io.ReadCloser     `json:"-"`             // 数据读取器
	FilePath    string            `json:"file_path"`     // 文件路径
	FileName    string            `json:"file_name"`     // 文件名
	Size        int64             `json:"size"`          // 文件大小
	ContentType string            `json:"content_type"`  // 内容类型
	ETag        string            `json:"etag"`          // 文件ETag
	Checksum    string            `json:"checksum"`      // 文件校验和
	ModifiedAt  time.Time         `json:"modified_at"`   // 修改时间
	URL         string            `json:"url,omitempty"` // 访问URL
	Metadata    map[string]string `json:"metadata"`      // 元数据
	Message     string            `json:"message"`       // 消息
	ErrorCode   string            `json:"error_code"`    // 错误代码
}

// SyncRequest 同步请求
type SyncRequest struct {
	SourceStorage  string   `json:"source_storage"`  // 源存储
	TargetStorage  string   `json:"target_storage"`  // 目标存储
	SourcePath     string   `json:"source_path"`     // 源路径
	TargetPath     string   `json:"target_path"`     // 目标路径
	SyncMode       string   `json:"sync_mode"`       // 同步模式: full, incremental, mirror
	DeleteOrphaned bool     `json:"delete_orphaned"` // 删除孤立文件
	PreserveAttrs  bool     `json:"preserve_attrs"`  // 保留属性
	CheckIntegrity bool     `json:"check_integrity"` // 检查完整性
	DryRun         bool     `json:"dry_run"`         // 预演模式
	IncludePattern []string `json:"include_pattern"` // 包含模式
	ExcludePattern []string `json:"exclude_pattern"` // 排除模式
}

// SyncResponse 同步响应
type SyncResponse struct {
	Success      bool           `json:"success"`
	SyncedFiles  []*SyncedFile  `json:"synced_files"`  // 同步的文件
	FailedFiles  []*FailedFile  `json:"failed_files"`  // 失败的文件
	SkippedFiles []*SkippedFile `json:"skipped_files"` // 跳过的文件
	TotalFiles   int            `json:"total_files"`   // 总文件数
	SyncedSize   int64          `json:"synced_size"`   // 同步大小
	Duration     time.Duration  `json:"duration"`      // 执行时间
	Message      string         `json:"message"`       // 消息
}

// BackupRequest 备份请求
type BackupRequest struct {
	SourceStorage string   `json:"source_storage"` // 源存储
	TargetStorage string   `json:"target_storage"` // 目标存储
	SourcePath    string   `json:"source_path"`    // 源路径
	TargetPath    string   `json:"target_path"`    // 目标路径
	BackupMode    string   `json:"backup_mode"`    // 备份模式: full, incremental, differential
	Compression   bool     `json:"compression"`    // 是否压缩
	Encryption    bool     `json:"encryption"`     // 是否加密
	Schedule      string   `json:"schedule"`       // 调度设置
	Retention     int      `json:"retention"`      // 保留天数
	Tags          []string `json:"tags"`           // 标签
}

// RestoreRequest 恢复请求
type RestoreRequest struct {
	BackupID      string   `json:"backup_id"`      // 备份ID
	TargetStorage string   `json:"target_storage"` // 目标存储
	TargetPath    string   `json:"target_path"`    // 目标路径
	RestoreMode   string   `json:"restore_mode"`   // 恢复模式: full, selective
	Files         []string `json:"files"`          // 指定文件
	Overwrite     bool     `json:"overwrite"`      // 是否覆盖
	PreserveAttrs bool     `json:"preserve_attrs"` // 保留属性
}

// SyncedFile 同步文件
type SyncedFile struct {
	SourcePath string        `json:"source_path"`
	TargetPath string        `json:"target_path"`
	Size       int64         `json:"size"`
	Action     string        `json:"action"` // create, update, copy
	Checksum   string        `json:"checksum"`
	SyncedAt   time.Time     `json:"synced_at"`
	Duration   time.Duration `json:"duration"`
}

// NewStorageManager 创建存储管理器实例
func NewStorageManager(
	storageRepo interfaces.StorageRepository,
	mediaRepo interfaces.MediaRepository,
	config *StorageConfig,
) *StorageManager {
	manager := &StorageManager{
		storageRepo:     storageRepo,
		mediaRepo:       mediaRepo,
		logger:          logger.Logger,
		storageBackends: make(map[string]StorageBackend),
		config:          config,
		cache:           NewStorageCache(config),
	}

	// 初始化存储后端
	for name, backendConfig := range config.StorageBackends {
		if backendConfig.Enabled {
			backend := manager.createBackend(backendConfig)
			if backend != nil {
				manager.storageBackends[name] = backend
				manager.logger.Info("存储后端初始化成功",
					zap.String("name", name),
					zap.String("type", backendConfig.Type))
			}
		}
	}

	return manager
}

// createBackend 创建存储后端
func (sm *StorageManager) createBackend(config *StorageBackendConfig) StorageBackend {
	switch config.Type {
	case "local":
		return NewLocalStorageBackend(config)
	case "s3":
		return NewS3StorageBackend(config)
	case "ftp":
		return NewFTPStorageBackend(config)
	case "webdav":
		return NewWebDAVStorageBackend(config)
	default:
		sm.logger.Error("不支持的存储后端类型", zap.String("type", config.Type))
		return nil
	}
}

// UploadFile 上传文件
func (sm *StorageManager) UploadFile(ctx context.Context, storageName, filePath string, reader io.Reader, size int64) (*UploadResponse, error) {
	backend, exists := sm.storageBackends[storageName]
	if !exists {
		return nil, fmt.Errorf("存储后端不存在: %s", storageName)
	}

	sm.logger.Info("开始上传文件",
		zap.String("storage", storageName),
		zap.String("path", filePath),
		zap.Int64("size", size))

	req := &UploadRequest{
		FilePath:   filePath,
		FileName:   filepath.Base(filePath),
		FileSize:   size,
		DataReader: reader,
		ChunkSize:  sm.config.ChunkSize,
		Overwrite:  true,
		CreateDirs: true,
	}

	resp, err := backend.UploadFile(ctx, req)
	if err != nil {
		sm.logger.Error("文件上传失败", zap.Error(err))
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("上传失败: %s", resp.Message)
	}

	// 保存文件记录到数据库
	file := &model.File{
		ID:          utils.GenerateID(),
		Name:        resp.FileName,
		Path:        resp.FilePath,
		Size:        resp.Size,
		Storage:     storageName,
		Checksum:    resp.Checksum,
		ContentType: resp.ContentType,
		CreatedAt:   resp.UploadedAt,
	}

	if err := sm.storageRepo.Create(ctx, file); err != nil {
		sm.logger.Warn("保存文件记录失败", zap.Error(err))
	}

	sm.logger.Info("文件上传完成",
		zap.String("file_id", resp.FileID),
		zap.String("path", resp.FilePath))

	return resp, nil
}

// DownloadFile 下载文件
func (sm *StorageManager) DownloadFile(ctx context.Context, storageName, filePath string) (*DownloadResponse, error) {
	backend, exists := sm.storageBackends[storageName]
	if !exists {
		return nil, fmt.Errorf("存储后端不存在: %s", storageName)
	}

	sm.logger.Info("开始下载文件",
		zap.String("storage", storageName),
		zap.String("path", filePath))

	req := &DownloadRequest{
		FilePath: filePath,
		Stream:   true,
	}

	resp, err := backend.DownloadFile(ctx, req)
	if err != nil {
		sm.logger.Error("文件下载失败", zap.Error(err))
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("下载失败: %s", resp.Message)
	}

	sm.logger.Info("文件下载完成",
		zap.String("path", resp.FilePath),
		zap.Int64("size", resp.Size))

	return resp, nil
}

// ListDirectory 列出目录
func (sm *StorageManager) ListDirectory(ctx context.Context, storageName, dirPath string) ([]*FileStat, error) {
	backend, exists := sm.storageBackends[storageName]
	if !exists {
		return nil, fmt.Errorf("存储后端不存在: %s", storageName)
	}

	return backend.ListDirectory(ctx, dirPath)
}

// GetFileInfo 获取文件信息
func (sm *StorageManager) GetFileInfo(ctx context.Context, storageName, filePath string) (*FileStat, error) {
	backend, exists := sm.storageBackends[storageName]
	if !exists {
		return nil, fmt.Errorf("存储后端不存在: %s", storageName)
	}

	return backend.GetFileInfo(ctx, filePath)
}

// SyncStorage 同步存储
func (sm *StorageManager) SyncStorage(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	sourceBackend, exists := sm.storageBackends[req.SourceStorage]
	if !exists {
		return nil, fmt.Errorf("源存储后端不存在: %s", req.SourceStorage)
	}

	targetBackend, exists := sm.storageBackends[req.TargetStorage]
	if !exists {
		return nil, fmt.Errorf("目标存储后端不存在: %s", req.TargetStorage)
	}

	sm.logger.Info("开始同步存储",
		zap.String("source", req.SourceStorage),
		zap.String("target", req.TargetStorage),
		zap.String("source_path", req.SourcePath),
		zap.String("target_path", req.TargetPath))

	startTime := time.Now()

	// 执行同步逻辑
	resp := &SyncResponse{
		Success:      true,
		SyncedFiles:  []*SyncedFile{},
		FailedFiles:  []*FailedFile{},
		SkippedFiles: []*SkippedFile{},
		Duration:     time.Since(startTime),
		Message:      "同步完成",
	}

	// 获取源文件列表
	sourceFiles, err := sourceBackend.ListDirectory(ctx, req.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("获取源文件列表失败: %w", err)
	}

	// 遍历源文件进行同步
	for _, file := range sourceFiles {
		if file.IsDir {
			continue // 跳过目录
		}

		// 检查是否匹配包含/排除模式
		if sm.shouldSkipFile(file.Name, req.IncludePattern, req.ExcludePattern) {
			resp.SkippedFiles = append(resp.SkippedFiles, &SkippedFile{
				FilePath: file.Name,
				Reason:   "不匹配同步模式",
			})
			continue
		}

		// 执行文件同步
		syncedFile, err := sm.syncFile(ctx, sourceBackend, targetBackend, req, file)
		if err != nil {
			resp.FailedFiles = append(resp.FailedFiles, &FailedFile{
				FilePath: file.Name,
				Error:    err.Error(),
			})
			continue
		}

		if syncedFile != nil {
			resp.SyncedFiles = append(resp.SyncedFiles, syncedFile)
			resp.SyncedSize += syncedFile.Size
		}
	}

	resp.TotalFiles = len(resp.SyncedFiles) + len(resp.FailedFiles) + len(resp.SkippedFiles)

	sm.logger.Info("存储同步完成",
		zap.Int("synced", len(resp.SyncedFiles)),
		zap.Int("failed", len(resp.FailedFiles)),
		zap.Int64("size", resp.SyncedSize),
		zap.Duration("duration", resp.Duration))

	return resp, nil
}

// shouldSkipFile 检查是否应该跳过文件
func (sm *StorageManager) shouldSkipFile(filename string, includePatterns, excludePatterns []string) bool {
	// 检查排除模式
	for _, pattern := range excludePatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}

	// 检查包含模式
	if len(includePatterns) > 0 {
		for _, pattern := range includePatterns {
			if matched, _ := filepath.Match(pattern, filename); matched {
				return false
			}
		}
		return true // 有包含模式但都不匹配
	}

	return false
}

// syncFile 同步单个文件
func (sm *StorageManager) syncFile(ctx context.Context, sourceBackend, targetBackend StorageBackend, req *SyncRequest, file *FileStat) (*SyncedFile, error) {
	sourcePath := filepath.Join(req.SourcePath, file.Name)
	targetPath := filepath.Join(req.TargetPath, file.Name)

	// 检查目标文件是否已存在
	targetInfo, err := targetBackend.GetFileInfo(ctx, targetPath)
	if err == nil && targetInfo != nil {
		// 文件已存在，比较修改时间和大小
		if targetInfo.ModTime.Equal(file.ModTime) && targetInfo.Size == file.Size {
			return nil, nil // 文件相同，跳过
		}
	}

	if req.DryRun {
		// 预演模式，不实际执行
		return &SyncedFile{
			SourcePath: sourcePath,
			TargetPath: targetPath,
			Size:       file.Size,
			Action:     "copy",
		}, nil
	}

	// 执行文件下载
	downloadReq := &DownloadRequest{FilePath: sourcePath}
	downloadResp, err := sourceBackend.DownloadFile(ctx, downloadReq)
	if err != nil {
		return nil, fmt.Errorf("下载源文件失败: %w", err)
	}
	defer downloadResp.DataReader.Close()

	// 执行文件上传
	uploadReq := &UploadRequest{
		FilePath:   targetPath,
		FileName:   file.Name,
		FileSize:   file.Size,
		DataReader: downloadResp.DataReader,
		Overwrite:  true,
		CreateDirs: true,
	}

	uploadResp, err := targetBackend.UploadFile(ctx, uploadReq)
	if err != nil {
		return nil, fmt.Errorf("上传目标文件失败: %w", err)
	}

	if !uploadResp.Success {
		return nil, fmt.Errorf("上传失败: %s", uploadResp.Message)
	}

	return &SyncedFile{
		SourcePath: sourcePath,
		TargetPath: targetPath,
		Size:       file.Size,
		Action:     "create",
		Checksum:   uploadResp.Checksum,
		SyncedAt:   time.Now(),
	}, nil
}

// GetStorageUsage 获取存储使用情况
func (sm *StorageManager) GetStorageUsage(ctx context.Context, storageName string) (*StorageUsage, error) {
	backend, exists := sm.storageBackends[storageName]
	if !exists {
		return nil, fmt.Errorf("存储后端不存在: %s", storageName)
	}

	return backend.GetStorageUsage(ctx)
}

// DeleteFile 删除文件
func (sm *StorageManager) DeleteFile(ctx context.Context, storageName, filePath string) error {
	backend, exists := sm.storageBackends[storageName]
	if !exists {
		return fmt.Errorf("存储后端不存在: %s", storageName)
	}

	// 从数据库删除记录
	if err := sm.storageRepo.DeleteByPath(ctx, filePath, storageName); err != nil {
		sm.logger.Warn("删除数据库记录失败", zap.Error(err))
	}

	// 从存储后端删除文件
	return backend.DeleteFile(ctx, filePath)
}

// InitializeBackends 初始化所有存储后端
func (sm *StorageManager) InitializeBackends(ctx context.Context) error {
	for name, backend := range sm.storageBackends {
		if err := backend.Init(ctx); err != nil {
			return fmt.Errorf("初始化存储后端 %s 失败: %w", name, err)
		}

		// 测试连接
		if err := backend.Test(ctx); err != nil {
			return fmt.Errorf("测试存储后端 %s 连接失败: %w", name, err)
		}

		sm.logger.Info("存储后端初始化成功", zap.String("name", name))
	}

	// 如果启用了自动同步，启动时执行同步
	if sm.config.AutoSync && sm.config.SyncOnStartup {
		go sm.runAutoSync(ctx)
	}

	return nil
}

// runAutoSync 运行自动同步
func (sm *StorageManager) runAutoSync(ctx context.Context) {
	ticker := time.NewTicker(sm.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sm.logger.Info("执行自动同步")
			// 这里应该根据配置执行相应的同步任务
		}
	}
}

// Close 关闭存储管理器
func (sm *StorageManager) Close(ctx context.Context) error {
	sm.logger.Info("关闭存储管理器")

	// 清理缓存
	if sm.cache != nil {
		sm.cache.Clear()
	}

	return nil
}
