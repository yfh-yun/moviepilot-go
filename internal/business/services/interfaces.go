package services

import (
	"context"
	"errors"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/models"
)

// SubscribeService 订阅服务接口
type SubscribeService interface {
	// 基础接口方法可以根据需要定义
	// 这里先定义空接口，由具体实现类实现
}

// DownloadService 下载服务接口
type DownloadService interface {
	// ListDownloads 获取下载任务列表
	ListDownloads(ctx context.Context, params ListDownloadsParams) ([]*DownloadTask, int64, error)
	
	// GetDownloadDetail 获取下载任务详情
	GetDownloadDetail(ctx context.Context, taskID string) (*DownloadTask, error)
	
	// CreateDownload 创建下载任务
	CreateDownload(ctx context.Context, params CreateDownloadParams) (*DownloadTask, error)
	
	// DeleteDownload 删除下载任务
	DeleteDownload(ctx context.Context, taskID string) error
	
	// PauseDownload 暂停下载任务
	PauseDownload(ctx context.Context, taskID string) error
	
	// ResumeDownload 恢复下载任务
	ResumeDownload(ctx context.Context, taskID string) error
	
	// GetDownloadStats 获取下载统计信息
	GetDownloadStats(ctx context.Context) (*DownloadStats, error)
	
	// GetDownloadSpeed 获取下载速度
	GetDownloadSpeed(ctx context.Context) (*DownloadSpeed, error)
	
	// ClearCompletedDownloads 清理已完成的下载任务
	ClearCompletedDownloads(ctx context.Context) error
	
	// BatchDeleteDownloads 批量删除下载任务
	BatchDeleteDownloads(ctx context.Context, taskIDs []string) error
}

// DownloadTask 下载任务
type DownloadTask struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Progress   float64   `json:"progress"`
	FileSize   int64     `json:"file_size"`
	Downloaded int64     `json:"downloaded"`
	Speed      int64     `json:"speed"`
	ETA        string    `json:"eta"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ListDownloadsParams 获取下载列表参数
type ListDownloadsParams struct {
	Status string `json:"status"`
	Type   string `json:"type"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// CreateDownloadParams 创建下载任务参数
type CreateDownloadParams struct {
	Title    string `json:"title"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	SavePath string `json:"save_path"`
}

// DownloadStats 下载统计信息
type DownloadStats struct {
	TotalTasks       int64   `json:"total_tasks"`
	ActiveTasks      int64   `json:"active_tasks"`
	CompletedTasks   int64   `json:"completed_tasks"`
	FailedTasks      int64   `json:"failed_tasks"`
	TotalDownloaded  int64   `json:"total_downloaded"`
	TotalSpeed       float64 `json:"total_speed"`
}

// DownloadSpeed 下载速度信息
type DownloadSpeed struct {
	CurrentSpeed int64   `json:"current_speed"`
	AverageSpeed float64 `json:"average_speed"`
	PeakSpeed    int64   `json:"peak_speed"`
}

// MessageService 消息服务接口
type MessageService interface {
	SendMessage(ctx context.Context, title, content string, messageType string, userIDs []uint) error
	GetMessages(ctx context.Context, userID uint, page, size int) ([]*models.Message, int64, error)
	MarkAsRead(ctx context.Context, messageID, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error
	DeleteMessage(ctx context.Context, messageID, userID uint) error
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
}

// PluginService 插件服务接口
type PluginService interface {
	InstallPlugin(ctx context.Context, pluginID string) error
	UninstallPlugin(ctx context.Context, pluginID string) error
	EnablePlugin(ctx context.Context, pluginID string) error
	DisablePlugin(ctx context.Context, pluginID string) error
	GetPlugin(ctx context.Context, pluginID string) (*models.Plugin, error)
	ListPlugins(ctx context.Context, enabledOnly bool) ([]*models.Plugin, error)
	UpdatePluginConfig(ctx context.Context, pluginID string, config map[string]interface{}) error
	GetPluginConfig(ctx context.Context, pluginID string) (map[string]interface{}, error)
	ExecutePlugin(ctx context.Context, pluginID string, data map[string]interface{}) (map[string]interface{}, error)
}

// Note 备注信息
type Note struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	UserID    string `json:"user_id"`
	TargetID  string `json:"target_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// 下载相关错误
var (
	ErrDownloadNotFound = errors.New("下载任务不存在")
	ErrDownloadExists   = errors.New("下载任务已存在")
	ErrInvalidStatus     = errors.New("无效的下载状态")
)

// 新增的业务逻辑接口

// DownloadStatusMonitor 下载状态监控接口
type DownloadStatusMonitor interface {
	MonitorDownloadStatus(ctx context.Context, request *DownloadStatusRequest) (*DownloadStatusResponse, error)
}

// WorkflowManager 工作流管理接口
type WorkflowManager interface {
	CacheWorkflow(ctx context.Context, workflowID, userID string) error
	GetCachedWorkflow(ctx context.Context, workflowID, userID string) (interface{}, bool)
	InvalidateWorkflowCache(ctx context.Context, workflowID, userID string) error
}

// SubscribeManager 订阅管理接口
type SubscribeManager interface {
	CacheSubscription(ctx context.Context, sub *repository.Subscribe) error
	GetCachedSubscription(ctx context.Context, subID string) (*repository.Subscribe, bool)
	InvalidateSubscriptionCache(ctx context.Context, subID string) error
	GetSubscriptionStatus(ctx context.Context, subID string) (string, error)
}

// EventManager 事件管理接口
type EventManager interface {
	SendEvent(ctx context.Context, event *Event) error
	GetEventHistory(ctx context.Context, filter *EventFilter) ([]*Event, error)
	ProcessEvent(ctx context.Context, eventID string) error
}

// TorrentFetcher 种子获取接口
type TorrentFetcher interface {
	FetchTorrents(ctx context.Context, request *TorrentFetchRequest) (*TorrentFetchResponse, error)
}

// FileScraper 文件抓取接口
type FileScraper interface {
	ScrapeFiles(ctx context.Context, request *FileScrapeRequest) (*FileScrapeResponse, error)
}

// NoteManager 备注管理接口
type NoteManager interface {
	ManageNote(ctx context.Context, request *NoteRequest) (*NoteResponse, error)
}

// 请求和响应类型定义

// DownloadStatusRequest 下载状态监控请求
type DownloadStatusRequest struct {
	UserID      string `json:"user_id"`
	DownloadIDs []string `json:"download_ids"`
	Status      string `json:"status"`
}

// DownloadStatusResponse 下载状态监控响应
type DownloadStatusResponse struct {
	Success bool      `json:"success"`
	Results []string  `json:"results"`
	Message string    `json:"message"`
}

// TorrentFetchRequest 种子获取请求
type TorrentFetchRequest struct {
	UserID  string `json:"user_id"`
	Site    string `json:"site"`
	Keyword string `json:"keyword"`
	Category string `json:"category"`
	SizeMin int64  `json:"size_min"`
	SizeMax int64  `json:"size_max"`
}

// TorrentFetchResponse 种子获取响应
type TorrentFetchResponse struct {
	Success  bool        `json:"success"`
	Torrents interface{} `json:"torrents"`
	Total    int         `json:"total"`
	Message  string      `json:"message"`
}

// FileScrapeRequest 文件抓取请求
type FileScrapeRequest struct {
	UserID     string   `json:"user_id"`
	Path       string   `json:"path"`
	Extensions []string `json:"extensions"`
	Recursive  bool     `json:"recursive"`
	AutoMatch  bool     `json:"auto_match"`
}

// FileScrapeResponse 文件抓取响应
type FileScrapeResponse struct {
	Success  bool      `json:"success"`
	Files    []string  `json:"files"`
	Total    int       `json:"total"`
	Message  string    `json:"message"`
}

// NoteRequest 备注管理请求
type NoteRequest struct {
	Action     string     `json:"action"`
	UserID     string     `json:"user_id"`
	TargetType string     `json:"target_type"`
	TargetID   string     `json:"target_id"`
	Note       *Note      `json:"note,omitempty"`
	NoteID     string     `json:"note_id,omitempty"`
}

// NoteResponse 备注管理响应
type NoteResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	UserID    string                 `json:"user_id"`
	CreatedAt string                 `json:"created_at"`
}

// EventFilter 事件过滤器
type EventFilter struct {
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}


