package adapters

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// DownloadItem 定义下载项
type DownloadItem struct {
	ID         string         `json:"id"`          // 下载ID
	Name       string         `json:"name"`        // 下载名称
	URL        string         `json:"url"`         // 下载URL
	Status     string         `json:"status"`      // 下载状态
	Size       int64          `json:"size"`        // 文件大小
	Downloaded int64          `json:"downloaded"`  // 已下载大小
	Progress   float64        `json:"progress"`    // 下载进度
	Speed      int64          `json:"speed"`       // 下载速度
	Eta        time.Duration  `json:"eta"`         // 预计剩余时间
	Category   string         `json:"category"`    // 下载分类
	SavePath   string         `json:"save_path"`   // 保存路径
	ClientName string         `json:"client_name"` // 客户端名称
	Tags       []string       `json:"tags"`        // 标签
	CreatedAt  time.Time      `json:"created_at"`  // 创建时间
	UpdatedAt  time.Time      `json:"updated_at"`  // 更新时间
	Metadata   map[string]any `json:"metadata"`    // 元数据
}

// DownloadStatus 定义下载状态
const (
	DownloadStatusPending   = "pending"   // 待开始
	DownloadStatusRunning   = "running"   // 下载中
	DownloadStatusPaused    = "paused"    // 已暂停
	DownloadStatusCompleted = "completed" // 已完成
	DownloadStatusFailed    = "failed"    // 下载失败
	DownloadStatusCancelled = "cancelled" // 已取消
)

// DownloadService 定义下载服务接口
type DownloadService interface {
	// AddDownload 添加下载任务
	AddDownload(ctx context.Context, params AddDownloadParams) (string, error)

	// GetDownloads 获取下载列表
	GetDownloads(ctx context.Context, params GetDownloadsParams) ([]DownloadItem, error)

	// GetDownload 获取单个下载详情
	GetDownload(ctx context.Context, downloadID string) (*DownloadItem, error)

	// PauseDownload 暂停下载
	PauseDownload(ctx context.Context, downloadID string) error

	// ResumeDownload 恢复下载
	ResumeDownload(ctx context.Context, downloadID string) error

	// CancelDownload 取消下载
	CancelDownload(ctx context.Context, downloadID string) error

	// DeleteDownload 删除下载
	DeleteDownload(ctx context.Context, downloadID string, deleteFiles bool) error
}

// AddDownloadParams 添加下载参数
type AddDownloadParams struct {
	URLs       []string       `json:"urls"`        // 下载URL列表
	ClientName string         `json:"client_name"` // 客户端名称
	Category   string         `json:"category"`    // 下载分类
	SavePath   string         `json:"save_path"`   // 保存路径
	Paused     bool           `json:"paused"`      // 是否暂停下载
	Tags       []string       `json:"tags"`        // 标签
	Metadata   map[string]any `json:"metadata"`    // 元数据
}

// GetDownloadsParams 获取下载列表参数
type GetDownloadsParams struct {
	ClientName string `json:"client_name"` // 客户端名称
	Status     string `json:"status"`      // 下载状态过滤
	Category   string `json:"category"`    // 下载分类过滤
	Limit      int    `json:"limit"`       // 返回结果数量限制
	SortBy     string `json:"sort_by"`     // 排序字段
	SortOrder  string `json:"sort_order"`  // 排序顺序
	StartAfter string `json:"start_after"` // 起始ID
}

// DownloadServiceAdapter 下载服务适配器实现
type DownloadServiceAdapter struct {
	logger *zap.Logger
	// 实际的下载服务客户端可以在这里注入
}

// NewDownloadServiceAdapter 创建新的下载服务适配器实例
func NewDownloadServiceAdapter(logger *zap.Logger) *DownloadServiceAdapter {
	return &DownloadServiceAdapter{
		logger: logger,
	}
}

// AddDownload 添加下载任务
func (a *DownloadServiceAdapter) AddDownload(ctx context.Context, params AddDownloadParams) (string, error) {
	// 实际实现中，这里应该调用核心业务服务的下载API
	// 这里使用模拟实现，返回一个随机生成的ID
	a.logger.Info("Adding download", zap.String("client_name", params.ClientName), zap.Strings("urls", params.URLs))
	return "download-" + time.Now().Format("20060102150405"), nil
}

// GetDownloads 获取下载列表
func (a *DownloadServiceAdapter) GetDownloads(ctx context.Context, params GetDownloadsParams) ([]DownloadItem, error) {
	// 实际实现中，这里应该调用核心业务服务的API获取下载列表
	// 这里使用模拟实现，返回一个空列表
	a.logger.Info("Getting downloads", zap.String("client_name", params.ClientName), zap.String("status", params.Status))
	return []DownloadItem{}, nil
}

// GetDownload 获取单个下载详情
func (a *DownloadServiceAdapter) GetDownload(ctx context.Context, downloadID string) (*DownloadItem, error) {
	// 实际实现中，这里应该调用核心业务服务的API获取单个下载详情
	// 这里使用模拟实现，返回nil
	a.logger.Info("Getting download", zap.String("download_id", downloadID))
	return nil, nil
}

// PauseDownload 暂停下载
func (a *DownloadServiceAdapter) PauseDownload(ctx context.Context, downloadID string) error {
	// 实际实现中，这里应该调用核心业务服务的API暂停下载
	// 这里使用模拟实现，返回nil
	a.logger.Info("Pausing download", zap.String("download_id", downloadID))
	return nil
}

// ResumeDownload 恢复下载
func (a *DownloadServiceAdapter) ResumeDownload(ctx context.Context, downloadID string) error {
	// 实际实现中，这里应该调用核心业务服务的API恢复下载
	// 这里使用模拟实现，返回nil
	a.logger.Info("Resuming download", zap.String("download_id", downloadID))
	return nil
}

// CancelDownload 取消下载
func (a *DownloadServiceAdapter) CancelDownload(ctx context.Context, downloadID string) error {
	// 实际实现中，这里应该调用核心业务服务的API取消下载
	// 这里使用模拟实现，返回nil
	a.logger.Info("Cancelling download", zap.String("download_id", downloadID))
	return nil
}

// DeleteDownload 删除下载
func (a *DownloadServiceAdapter) DeleteDownload(ctx context.Context, downloadID string, deleteFiles bool) error {
	// 实际实现中，这里应该调用核心业务服务的API删除下载
	// 这里使用模拟实现，返回nil
	a.logger.Info("Deleting download", zap.String("download_id", downloadID), zap.Bool("delete_files", deleteFiles))
	return nil
}

// MockDownloadService 模拟下载服务实现，用于测试
type MockDownloadService struct {
	logger    *zap.Logger
	downloads map[string]DownloadItem
}

// NewMockDownloadService 创建新的模拟下载服务实例
func NewMockDownloadService(logger *zap.Logger) *MockDownloadService {
	return &MockDownloadService{
		logger:    logger,
		downloads: make(map[string]DownloadItem),
	}
}

// AddDownload 添加下载任务（模拟实现）
func (m *MockDownloadService) AddDownload(ctx context.Context, params AddDownloadParams) (string, error) {
	m.logger.Info("Mock adding download", zap.String("client_name", params.ClientName), zap.Strings("urls", params.URLs))
	downloadID := "mock-download-" + time.Now().Format("20060102150405")

	// 创建模拟下载项
	download := DownloadItem{
		ID:         downloadID,
		Name:       params.URLs[0],
		URL:        params.URLs[0],
		Status:     DownloadStatusPending,
		Size:       0,
		Downloaded: 0,
		Progress:   0,
		Speed:      0,
		Eta:        0,
		Category:   params.Category,
		SavePath:   params.SavePath,
		ClientName: params.ClientName,
		Tags:       params.Tags,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   params.Metadata,
	}

	m.downloads[downloadID] = download
	return downloadID, nil
}

// GetDownloads 获取下载列表（模拟实现）
func (m *MockDownloadService) GetDownloads(ctx context.Context, params GetDownloadsParams) ([]DownloadItem, error) {
	m.logger.Info("Mock getting downloads", zap.String("client_name", params.ClientName), zap.String("status", params.Status))

	// 返回模拟下载列表
	var downloads []DownloadItem
	for _, download := range m.downloads {
		if params.Status == "" || download.Status == params.Status {
			downloads = append(downloads, download)
		}
	}

	return downloads, nil
}

// GetDownload 获取单个下载详情（模拟实现）
func (m *MockDownloadService) GetDownload(ctx context.Context, downloadID string) (*DownloadItem, error) {
	m.logger.Info("Mock getting download", zap.String("download_id", downloadID))

	// 返回模拟下载详情
	download, exists := m.downloads[downloadID]
	if !exists {
		return nil, nil
	}

	return &download, nil
}

// PauseDownload 暂停下载（模拟实现）
func (m *MockDownloadService) PauseDownload(ctx context.Context, downloadID string) error {
	m.logger.Info("Mock pausing download", zap.String("download_id", downloadID))

	// 更新模拟下载状态
	if download, exists := m.downloads[downloadID]; exists {
		download.Status = DownloadStatusPaused
		download.UpdatedAt = time.Now()
		m.downloads[downloadID] = download
	}

	return nil
}

// ResumeDownload 恢复下载（模拟实现）
func (m *MockDownloadService) ResumeDownload(ctx context.Context, downloadID string) error {
	m.logger.Info("Mock resuming download", zap.String("download_id", downloadID))

	// 更新模拟下载状态
	if download, exists := m.downloads[downloadID]; exists {
		download.Status = DownloadStatusRunning
		download.UpdatedAt = time.Now()
		m.downloads[downloadID] = download
	}

	return nil
}

// CancelDownload 取消下载（模拟实现）
func (m *MockDownloadService) CancelDownload(ctx context.Context, downloadID string) error {
	m.logger.Info("Mock cancelling download", zap.String("download_id", downloadID))

	// 更新模拟下载状态
	if download, exists := m.downloads[downloadID]; exists {
		download.Status = DownloadStatusCancelled
		download.UpdatedAt = time.Now()
		m.downloads[downloadID] = download
	}

	return nil
}

// DeleteDownload 删除下载（模拟实现）
func (m *MockDownloadService) DeleteDownload(ctx context.Context, downloadID string, deleteFiles bool) error {
	m.logger.Info("Mock deleting download", zap.String("download_id", downloadID), zap.Bool("delete_files", deleteFiles))

	// 从模拟下载列表中删除
	delete(m.downloads, downloadID)
	return nil
}
