package storage

import (
	"context"
	"io"
	"time"
)

// StorageProvider 存储提供商接口
type StorageProvider interface {
	// 基本信息
	Name() string
	Type() string
	IsConnected() bool

	// 文件操作
	Upload(ctx context.Context, filePath string, reader io.Reader, size int64) error
	Download(ctx context.Context, filePath string) (io.ReadCloser, error)
	Delete(ctx context.Context, filePath string) error
	Exists(ctx context.Context, filePath string) (bool, error)
	Move(ctx context.Context, srcPath, dstPath string) error
	Copy(ctx context.Context, srcPath, dstPath string) error

	// 目录操作
	List(ctx context.Context, path string) ([]FileInfo, error)
	Mkdir(ctx context.Context, path string) error
	Rmdir(ctx context.Context, path string) error

	// 统计信息
	GetQuota(ctx context.Context) (*QuotaInfo, error)
	GetFileInfo(ctx context.Context, path string) (*FileInfo, error)

	// 连接管理
	Connect(ctx context.Context) error
	Disconnect() error
}

// FileInfo 文件信息
type FileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	IsDir        bool      `json:"isDir"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modifiedTime"`
	MimeType     string    `json:"mimeType"`
	Checksum     string    `json:"checksum"`
}

// QuotaInfo 配额信息
type QuotaInfo struct {
	Total     int64 `json:"total"`     // 总容量
	Used      int64 `json:"used"`      // 已使用
	Available int64 `json:"available"` // 可用空间
	Files     int64 `json:"files"`     // 文件数量
	Folders   int64 `json:"folders"`   // 文件夹数量
}

// TransferProgress 传输进度
type TransferProgress struct {
	FilePath      string `json:"filePath"`
	BytesSent     int64  `json:"bytesSent"`
	TotalBytes    int64  `json:"totalBytes"`
	Percentage    int    `json:"percentage"`
	Speed         int64  `json:"speed"`         // bytes/sec
	TimeElapsed   int64  `json:"timeElapsed"`   // seconds
	TimeRemaining int64  `json:"timeRemaining"` // seconds
}

// TransferCallback 传输回调函数
type TransferCallback func(progress *TransferProgress)

// Config 存储配置
type Config struct {
	Provider   string                 `json:"provider"`
	Name       string                 `json:"name"`
	Enabled    bool                   `json:"enabled"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ProviderType 存储提供商类型
const (
	ProviderLocal  = "local"
	ProviderSMB    = "smb"
	ProviderAliPan = "alipan"
	ProviderU115   = "u115"
	ProviderRClone = "rclone"
)

// Error 存储错误
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("存储错误[%s]: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("存储错误[%s]: %s", e.Code, e.Message)
}

// 预定义错误
var (
	ErrNotImplemented   = &Error{Code: "NOT_IMPLEMENTED", Message: "功能未实现"}
	ErrNotConnected     = &Error{Code: "NOT_CONNECTED", Message: "存储未连接"}
	ErrFileNotFound     = &Error{Code: "FILE_NOT_FOUND", Message: "文件未找到"}
	ErrPermissionDenied = &Error{Code: "PERMISSION_DENIED", Message: "权限不足"}
	ErrQuotaExceeded    = &Error{Code: "QUOTA_EXCEEDED", Message: "存储配额不足"}
	ErrNetworkError     = &Error{Code: "NETWORK_ERROR", Message: "网络错误"}
	ErrInvalidPath      = &Error{Code: "INVALID_PATH", Message: "无效路径"}
)

// NewError 创建新的错误
func NewError(code, message string, details ...string) error {
	err := &Error{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 {
		err.Details = details[0]
	}

	return err
}

// IsNotFoundError 检查是否为"未找到"错误
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	if storageErr, ok := err.(*Error); ok {
		return storageErr.Code == "FILE_NOT_FOUND"
	}

	return false
}

// IsPermissionError 检查是否为权限错误
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}

	if storageErr, ok := err.(*Error); ok {
		return storageErr.Code == "PERMISSION_DENIED"
	}

	return false
}

// IsQuotaError 检查是否为配额错误
func IsQuotaError(err error) bool {
	if err == nil {
		return false
	}

	if storageErr, ok := err.(*Error); ok {
		return storageErr.Code == "QUOTA_EXCEEDED"
	}

	return false
}

// StorageManager 存储管理器
type StorageManager struct {
	providers map[string]StorageProvider
	mu        sync.RWMutex
}

// NewStorageManager 创建存储管理器
func NewStorageManager() *StorageManager {
	return &StorageManager{
		providers: make(map[string]StorageProvider),
	}
}

// RegisterProvider 注册存储提供商
func (sm *StorageManager) RegisterProvider(name string, provider StorageProvider) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.providers[name]; exists {
		return NewError("PROVIDER_EXISTS", "存储提供商已存在", name)
	}

	sm.providers[name] = provider
	return nil
}

// GetProvider 获取存储提供商
func (sm *StorageManager) GetProvider(name string) (StorageProvider, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	provider, exists := sm.providers[name]
	if !exists {
		return nil, NewError("PROVIDER_NOT_FOUND", "存储提供商未找到", name)
	}

	return provider, nil
}

// ListProviders 列出所有存储提供商
func (sm *StorageManager) ListProviders() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var providers []string
	for name := range sm.providers {
		providers = append(providers, name)
	}

	return providers
}

// RemoveProvider 移除存储提供商
func (sm *StorageManager) RemoveProvider(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	provider, exists := sm.providers[name]
	if !exists {
		return NewError("PROVIDER_NOT_FOUND", "存储提供商未找到", name)
	}

	// 断开连接
	if err := provider.Disconnect(); err != nil {
		return fmt.Errorf("断开连接失败: %w", err)
	}

	delete(sm.providers, name)
	return nil
}
