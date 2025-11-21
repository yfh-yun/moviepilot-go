package services

import (
	"context"
	"moviepilot-go/internal/business/domains"
	"moviepilot-go/internal/models"
	"time"
)

// UserService 用户服务接口
type UserService interface {
	CreateUser(ctx context.Context, user *domains.User) error
	GetUserByID(ctx context.Context, id uint) (*domains.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domains.User, error)
	UpdateUser(ctx context.Context, user *domains.User) error
	DeleteUser(ctx context.Context, id uint) error
	AuthenticateUser(ctx context.Context, username, password string) (*domains.User, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

// MediaService 媒体服务接口
type MediaService interface {
	CreateMedia(ctx context.Context, media *domains.Media) error
	GetMediaByID(ctx context.Context, id uint) (*domains.Media, error)
	GetMediaByTMDBID(ctx context.Context, tmdbID int, mediaType string) (*domains.Media, error)
	UpdateMedia(ctx context.Context, media *domains.Media) error
	DeleteMedia(ctx context.Context, id uint) error
	SearchMedia(ctx context.Context, criteria domains.SearchCriteria) ([]domains.Media, int, error)
	GetPopularMedia(ctx context.Context, mediaType string, limit int) ([]domains.Media, error)
	GetLatestMedia(ctx context.Context, mediaType string, limit int) ([]domains.Media, error)
}

// SubscribeService 订阅服务接口
type SubscribeService interface {
	CreateSubscribe(ctx context.Context, subscribe *domains.Subscribe) error
	GetSubscribeByID(ctx context.Context, id uint) (*domains.Subscribe, error)
	GetSubscribesByUserID(ctx context.Context, userID uint) ([]domains.Subscribe, error)
	UpdateSubscribe(ctx context.Context, subscribe *domains.Subscribe) error
	DeleteSubscribe(ctx context.Context, id uint) error
	GetActiveSubscribes(ctx context.Context) ([]domains.Subscribe, error)
	ProcessSubscriptions(ctx context.Context) error
	RenewSubscription(ctx context.Context, id uint) error
}

// TransferService 转移服务接口
type TransferService interface {
	CreateTransfer(ctx context.Context, transfer *domains.Transfer) error
	GetTransferByID(ctx context.Context, id uint) (*domains.Transfer, error)
	GetTransfersByUserID(ctx context.Context, userID uint) ([]domains.Transfer, error)
	UpdateTransfer(ctx context.Context, transfer *domains.Transfer) error
	DeleteTransfer(ctx context.Context, id uint) error
	GetActiveTransfers(ctx context.Context) ([]domains.Transfer, error)
	ProcessTransfer(ctx context.Context, transferID uint) error
	CancelTransfer(ctx context.Context, transferID uint) error
	GetTransferHistory(ctx context.Context, filters map[string]interface{}) ([]domains.Transfer, int, error)
}

// DownloadService 下载服务接口
type DownloadService interface {
	StartDownload(ctx context.Context, url string, config domains.DownloadConfig) (*domains.Transfer, error)
	PauseDownload(ctx context.Context, downloadID uint) error
	ResumeDownload(ctx context.Context, downloadID uint) error
	CancelDownload(ctx context.Context, downloadID uint) error
	GetDownloadStatus(ctx context.Context, downloadID uint) (*domains.Transfer, error)
	GetActiveDownloads(ctx context.Context) ([]domains.Transfer, error)
	CleanupCompletedDownloads(ctx context.Context) error
}

// NotificationService 通知服务接口
type NotificationService interface {
	SendNotification(ctx context.Context, userID uint, title, message string, config domains.NotificationConfig) error
	SendBulkNotification(ctx context.Context, userIDs []uint, title, message string, config domains.NotificationConfig) error
	GetNotificationHistory(ctx context.Context, userID uint, limit int) ([]models.Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID uint) error
	GetUnreadNotifications(ctx context.Context, userID uint) ([]models.Notification, error)
}

// SearchService 搜索服务接口
type SearchService interface {
	SearchMedia(ctx context.Context, query string, filters map[string]interface{}) ([]domains.Media, error)
	SearchTorrents(ctx context.Context, query string, filters map[string]interface{}) ([]models.TorrentInfo, error)
	GetSearchSuggestions(ctx context.Context, query string) ([]string, error)
	GetPopularSearches(ctx context.Context) ([]string, error)
	SaveSearchHistory(ctx context.Context, userID uint, query string) error
	GetSearchHistory(ctx context.Context, userID uint, limit int) ([]models.SearchHistory, error)
}

// PluginService 插件服务接口
type PluginService interface {
	LoadPlugin(ctx context.Context, pluginID string) error
	UnloadPlugin(ctx context.Context, pluginID string) error
	GetLoadedPlugins(ctx context.Context) ([]models.PluginInfo, error)
	GetPluginInfo(ctx context.Context, pluginID string) (*models.PluginInfo, error)
	ExecutePluginAction(ctx context.Context, pluginID, action string, params map[string]interface{}) (interface{}, error)
	GetPluginConfig(ctx context.Context, pluginID string) (map[string]interface{}, error)
	UpdatePluginConfig(ctx context.Context, pluginID string, config map[string]interface{}) error
}

// WorkflowService 工作流服务接口
type WorkflowService interface {
	CreateWorkflow(ctx context.Context, workflow *domains.Workflow) error
	GetWorkflowByID(ctx context.Context, id uint) (*domains.Workflow, error)
	GetWorkflowsByUserID(ctx context.Context, userID uint) ([]domains.Workflow, error)
	UpdateWorkflow(ctx context.Context, workflow *domains.Workflow) error
	DeleteWorkflow(ctx context.Context, id uint) error
	ExecuteWorkflow(ctx context.Context, workflowID uint, triggerData map[string]interface{}) error
	GetWorkflowExecutions(ctx context.Context, workflowID uint) ([]domains.WorkflowExecution, error)
	ScheduleWorkflow(ctx context.Context, workflowID uint, schedule string) error
}

// SystemService 系统服务接口
type SystemService interface {
	GetSystemInfo(ctx context.Context) (map[string]interface{}, error)
	GetSystemStats(ctx context.Context) (map[string]interface{}, error)
	GetSystemHealth(ctx context.Context) (map[string]interface{}, error)
	PerformSystemMaintenance(ctx context.Context, task string) error
	GetSystemLogs(ctx context.Context, level string, limit int) ([]models.SystemLog, error)
	CleanupSystem(ctx context.Context) error
	BackupSystem(ctx context.Context) (string, error)
	RestoreSystem(ctx context.Context, backupPath string) error
}

// StorageService 存储服务接口
type StorageService interface {
	GetStorageInfo(ctx context.Context, storageType string) (domains.StorageConfig, error)
	UploadFile(ctx context.Context, storageType string, filePath string, data []byte) error
	DownloadFile(ctx context.Context, storageType string, filePath string) ([]byte, error)
	DeleteFile(ctx context.Context, storageType string, filePath string) error
	ListFiles(ctx context.Context, storageType string, path string) ([]models.FileInfo, error)
	GetFileStats(ctx context.Context, storageType string) (map[string]interface{}, error)
	MigrateFiles(ctx context.Context, fromStorage, toStorage string) error
}

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, username, password string) (*domains.TokenPair, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (*domains.TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*domains.User, error)
	GetCurrentUser(ctx context.Context, token string) (*domains.User, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error
	EnableTwoFactor(ctx context.Context, userID uint) (string, error)
	VerifyTwoFactor(ctx context.Context, userID uint, code string) error
}

// HistoryService 历史服务接口
type HistoryService interface {
	CreateHistory(ctx context.Context, history *models.History) error
	GetHistoryByUserID(ctx context.Context, userID uint, filters map[string]interface{}) ([]models.History, error)
	GetHistoryByType(ctx context.Context, historyType string, limit int) ([]models.History, error)
	DeleteHistory(ctx context.Context, id uint) error
	ClearHistory(ctx context.Context, userID uint, historyType string) error
	GetHistoryStats(ctx context.Context, userID uint) (map[string]interface{}, error)
}

// ConfigService 配置服务接口
type ConfigService interface {
	GetConfig(ctx context.Context, key string) (interface{}, error)
	SetConfig(ctx context.Context, key string, value interface{}) error
	DeleteConfig(ctx context.Context, key string) error
	GetAllConfigs(ctx context.Context) (map[string]interface{}, error)
	ResetConfig(ctx context.Context, key string) error
	ExportConfigs(ctx context.Context) (map[string]interface{}, error)
	ImportConfigs(ctx context.Context, configs map[string]interface{}) error
	ValidateConfig(ctx context.Context, key string, value interface{}) error
}

// CacheService 缓存服务接口
type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
	SetMultiple(ctx context.Context, items map[string]interface{}, expiration time.Duration) error
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
}

// MetricsService 指标服务接口
type MetricsService interface {
	RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error
	IncrementCounter(ctx context.Context, name string, tags map[string]string) error
	SetGauge(ctx context.Context, name string, value float64, tags map[string]string) error
	RecordHistogram(ctx context.Context, name string, value float64, tags map[string]string) error
	GetMetrics(ctx context.Context, query string) (map[string]interface{}, error)
	GetMetricsSummary(ctx context.Context, timeRange string) (map[string]interface{}, error)
	ResetMetrics(ctx context.Context) error
}