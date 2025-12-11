package database

import (
	"time"
)

// WriteFileRequest 写入文件请求
type WriteFileRequest struct {
	Path string `json:"path" binding:"required"`
	Data []byte `json:"data" binding:"required"`
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	OldPath string `json:"old_path" binding:"required"`
	NewPath string `json:"new_path" binding:"required"`
}

// CopyFileRequest 复制文件请求
type CopyFileRequest struct {
	SrcPath string `json:"src_path" binding:"required"`
	DstPath string `json:"dst_path" binding:"required"`
}

// CreateDirectoryRequest 创建目录请求
type CreateDirectoryRequest struct {
	Path string `json:"path" binding:"required"`
}

// SetPermissionsRequest 设置权限请求
type SetPermissionsRequest struct {
	Path        string `json:"path" binding:"required"`
	Permissions string `json:"permissions" binding:"required"`
}

// CleanupOptions 清理选项
type CleanupOptions struct {
	CleanTempFiles  bool `json:"clean_temp_files"`
	RemoveEmptyDirs bool `json:"remove_empty_dirs"`
	MaxAgeDays      int  `json:"max_age_days"`
}

// CleanupResult 清理结果
type CleanupResult struct {
	FilesDeleted    int64 `json:"files_deleted"`
	DirsDeleted     int64 `json:"dirs_deleted"`
	SpaceFreedBytes int64 `json:"space_freed_bytes"`
}

// CreateBackupRequest 创建备份请求
type CreateBackupRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Paths       []string `json:"paths"`
	Exclude     []string `json:"exclude"`
}

// BackupInfo 备份信息
type BackupInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}

// FileSearchQuery 文件搜索查询
type FileSearchQuery struct {
	Path      string `json:"path"`
	Pattern   string `json:"pattern"`
	Recursive bool   `json:"recursive"`
	FileType  string `json:"file_type"`
	MaxDepth  int    `json:"max_depth"`
}

// FileSearchResult 文件搜索结果
type FileSearchResult struct {
	Files []FileInfo `json:"files"`
	Total int        `json:"total"`
}

// FileProcessRequest 文件处理请求
type FileProcessRequest struct {
	Path      string         `json:"path" binding:"required"`
	Operation string         `json:"operation" binding:"required"`
	Options   map[string]any `json:"options"`
}

// FileProcessResult 文件处理结果
type FileProcessResult struct {
	Success  bool           `json:"success"`
	Message  string         `json:"message"`
	Data     map[string]any `json:"data,omitempty"`
	Error    string         `json:"error,omitempty"`
	Duration time.Duration  `json:"duration"`
}

// FileHandlerInfo 文件处理器信息
type FileHandlerInfo struct {
	HandlerName      string   `json:"handler_name"`
	Version          string   `json:"version"`
	RegisteredRoutes []string `json:"registered_routes"`
}

// DirectoryListing 目录列表
type DirectoryListing struct {
	Path    string     `json:"path"`
	Files   []FileInfo `json:"files"`
	Dirs    []FileInfo `json:"dirs"`
	Total   int        `json:"total"`
	HasMore bool       `json:"has_more"`
}

// FileInfo 文件信息
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	IsDir       bool      `json:"is_dir"`
	ModTime     time.Time `json:"mod_time"`
	Mode        string    `json:"mode"`
	Hash        string    `json:"hash,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
	Permissions string    `json:"permissions"`
	Owner       string    `json:"owner,omitempty"`
	Group       string    `json:"group,omitempty"`
}

// StorageInfo 存储信息
type StorageInfo struct {
	TotalBytes   int64   `json:"total_bytes"`
	FreeBytes    int64   `json:"free_bytes"`
	UsedBytes    int64   `json:"used_bytes"`
	TotalFiles   int64   `json:"total_files"`
	TotalDirs    int64   `json:"total_dirs"`
	UsagePercent float64 `json:"usage_percent"`
	MountPoint   string  `json:"mount_point"`
	Filesystem   string  `json:"filesystem"`
}

// StorageHealth 存储健康状态
type StorageHealth struct {
	Status      string            `json:"status"`
	Message     string            `json:"message"`
	Checks      map[string]string `json:"checks"`
	LastChecked time.Time         `json:"last_checked"`
}
