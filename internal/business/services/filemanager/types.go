package filemanager

import (
	"context"
	"io"
	"time"
)

// FileInfo 文件信息
type FileInfo struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	Size         int64             `json:"size"`
	IsDir        bool              `json:"is_dir"`
	ModifiedTime time.Time         `json:"modified_time"`
	CreatedTime  time.Time         `json:"created_time"`
	ContentType  string            `json:"content_type"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// UploadProgress 上传进度
type UploadProgress struct {
	TotalBytes     int64   `json:"total_bytes"`
	UploadedBytes  int64   `json:"uploaded_bytes"`
	Percentage    float64 `json:"percentage"`
	Speed         int64    `json:"speed"` // bytes/second
	EstimatedTime int64    `json:"estimated_time"` // seconds
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	TotalBytes       int64   `json:"total_bytes"`
	DownloadedBytes  int64   `json:"downloaded_bytes"`
	Percentage       float64 `json:"percentage"`
	Speed            int64    `json:"speed"` // bytes/second
	EstimatedTime    int64    `json:"estimated_time"` // seconds
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Enabled    bool                   `json:"enabled"`
	Config     map[string]interface{} `json:"config"`
	Priority   int                    `json:"priority"`
	ReadOnly   bool                   `json:"read_only"`
	TempDir    string                 `json:"temp_dir,omitempty"`
	MaxRetries int                    `json:"max_retries"`
	Timeout    int                    `json:"timeout"`
}

// StorageProvider 存储提供商接口
type StorageProvider interface {
	// GetName 获取存储名称
	GetName() string
	
	// GetType 获取存储类型
	GetType() string
	
	// ValidateConfig 验证配置
	ValidateConfig(config map[string]interface{}) error
	
	// Initialize 初始化存储
	Initialize(config map[string]interface{}) error
	
	// ListFiles 列出文件
	ListFiles(ctx context.Context, path string) ([]*FileInfo, error)
	
	// UploadFile 上传文件
	UploadFile(ctx context.Context, localPath, remotePath string, progress *UploadProgress) error
	
	// UploadStream 上传流
	UploadStream(ctx context.Context, stream io.Reader, remotePath string, size int64, progress *UploadProgress) error
	
	// DownloadFile 下载文件
	DownloadFile(ctx context.Context, remotePath, localPath string, progress *DownloadProgress) error
	
	// DownloadStream 下载流
	DownloadStream(ctx context.Context, remotePath string) (io.ReadCloser, error)
	
	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, path string) error
	
	// CreateDirectory 创建目录
	CreateDirectory(ctx context.Context, path string) error
	
	// DeleteDirectory 删除目录
	DeleteDirectory(ctx context.Context, path string, recursive bool) error
	
	// RenameFile 重命名文件
	RenameFile(ctx context.Context, oldPath, newPath string) error
	
	// CopyFile 复制文件
	CopyFile(ctx context.Context, sourcePath, targetPath string) error
	
	// MoveFile 移动文件
	MoveFile(ctx context.Context, sourcePath, targetPath string) error
	
	// GetFileInfo 获取文件信息
	GetFileInfo(ctx context.Context, path string) (*FileInfo, error)
	
	// GetURL 获取文件访问URL
	GetURL(ctx context.Context, path string, expires time.Duration) (string, error)
	
	// IsHealthy 健康检查
	IsHealthy(ctx context.Context) bool
	
	// GetConfig 获取配置
	GetConfig() map[string]interface{}
	
	// Close 关闭存储
	Close() error
}

// TransferType 传输类型
type TransferType string

const (
	TransferTypeUpload   TransferType = "upload"
	TransferTypeDownload TransferType = "download"
)

// TransferTask 传输任务
type TransferTask struct {
	ID            string           `json:"id"`
	Type          TransferType     `json:"type"`
	StorageName   string           `json:"storage_name"`
	SourcePath    string           `json:"source_path"`
	TargetPath    string           `json:"target_path"`
	Status        string           `json:"status"` // pending, running, completed, failed, paused
	Progress      float64          `json:"progress"`
	Speed         int64            `json:"speed"`
	TotalBytes    int64            `json:"total_bytes"`
	Transferred   int64            `json:"transferred"`
	Error         string           `json:"error,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	StartedAt     *time.Time       `json:"started_at,omitempty"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
	RetryCount    int              `json:"retry_count"`
	MaxRetries    int              `json:"max_retries"`
	Priority      int              `json:"priority"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// FileManagerService 文件管理服务
type FileManagerService interface {
	// AddStorage 添加存储
	AddStorage(config *StorageConfig) error
	
	// RemoveStorage 移除存储
	RemoveStorage(name string) error
	
	// GetStorage 获取存储
	GetStorage(name string) (StorageProvider, error)
	
	// ListStorages 列出所有存储
	ListStorages() []string
	
	// ListFiles 在指定存储中列出文件
	ListFiles(ctx context.Context, storageName, path string) ([]*FileInfo, error)
	
	// UploadFile 上传文件
	UploadFile(ctx context.Context, storageName, localPath, remotePath string, progress *UploadProgress) error
	
	// DownloadFile 下载文件
	DownloadFile(ctx context.Context, storageName, remotePath, localPath string, progress *DownloadProgress) error
	
	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, storageName, path string) error
	
	// CopyFilesBetweenStorages 在存储间复制文件
	CopyFilesBetweenStorages(ctx context.Context, sourceStorage, sourcePath, targetStorage, targetPath string) error
	
	// MoveFilesBetweenStorages 在存储间移动文件
	MoveFilesBetweenStorages(ctx context.Context, sourceStorage, sourcePath, targetStorage, targetPath string) error
	
	// SyncFiles 同步文件
	SyncFiles(ctx context.Context, sourceStorage, sourcePath, targetStorage, targetPath string, delete bool) error
	
	// CreateTransferTask 创建传输任务
	CreateTransferTask(task *TransferTask) error
	
	// GetTransferTask 获取传输任务
	GetTransferTask(taskID string) (*TransferTask, error)
	
	// ListTransferTasks 列出传输任务
	ListTransferTasks(status string) ([]*TransferTask, error)
	
	// PauseTransferTask 暂停传输任务
	PauseTransferTask(taskID string) error
	
	// ResumeTransferTask 恢复传输任务
	ResumeTransferTask(taskID string) error
	
	// CancelTransferTask 取消传输任务
	CancelTransferTask(taskID string) error
	
	// RetryTransferTask 重试传输任务
	RetryTransferTask(taskID string) error
	
	// GetTransferProgress 获取传输进度
	GetTransferProgress(taskID string) (*TransferProgress, error)
}

// TransferProgress 传输进度
type TransferProgress struct {
	TaskID        string    `json:"task_id"`
	Status        string    `json:"status"`
	Progress      float64   `json:"progress"`
	Speed         int64     `json:"speed"`
	TotalBytes    int64     `json:"total_bytes"`
	Transferred   int64     `json:"transferred"`
	EstimatedTime int64     `json:"estimated_time"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SyncResult 同步结果
type SyncResult struct {
	SourceFiles      []*FileInfo `json:"source_files"`
	TargetFiles      []*FileInfo `json:"target_files"`
	AddedFiles       []*FileInfo `json:"added_files"`
	UpdatedFiles     []*FileInfo `json:"updated_files"`
	DeletedFiles     []*FileInfo `json:"deleted_files"`
	SkippedFiles    []*FileInfo `json:"skipped_files"`
	TotalSize        int64       `json:"total_size"`
	TransferredSize  int64       `json:"transferred_size"`
	ProcessedCount   int         `json:"processed_count"`
	SuccessCount     int         `json:"success_count"`
	ErrorCount       int         `json:"error_count"`
	StartTime        time.Time   `json:"start_time"`
	EndTime          time.Time   `json:"end_time"`
	Duration         time.Duration `json:"duration"`
	Errors           []string    `json:"errors,omitempty"`
}