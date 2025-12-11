package dashboard

import (
	"context"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Service 仪表板服务接口
type Service interface {
	// GetDashboardData 获取仪表板数据
	GetDashboardData(ctx context.Context) (*DashboardData, error)

	// GetStatistics 获取统计信息
	GetStatistics(ctx context.Context) (*Statistics, error)

	// GetRecentActivity 获取最近活动
	GetRecentActivity(ctx context.Context, limit int) ([]*Activity, error)

	// GetChartData 获取图表数据
	GetChartData(ctx context.Context, chartType string, days int) (*ChartData, error)
}

// service 服务实现
type service struct {
	logger *zap.Logger
}

// NewService 创建服务
func NewService() Service {
	return &service{
		logger: logger.GetLogger(),
	}
}

// DashboardData 仪表板数据
type DashboardData struct {
	Statistics     *Statistics           `json:"statistics"`
	RecentActivity []*Activity           `json:"recent_activity"`
	Charts         map[string]*ChartData `json:"charts"`
	SystemInfo     *SystemInfo           `json:"system_info"`
}

// Statistics 统计信息
type Statistics struct {
	TotalSubscribes  int64 `json:"total_subscribes"`
	ActiveSubscribes int64 `json:"active_subscribes"`
	TotalDownloads   int64 `json:"total_downloads"`
	ActiveDownloads  int64 `json:"active_downloads"`
	TotalSites       int64 `json:"total_sites"`
	ActiveSites      int64 `json:"active_sites"`
	TotalStorage     int64 `json:"total_storage"`  // 字节
	UsedStorage      int64 `json:"used_storage"`   // 字节
	DownloadSpeed    int64 `json:"download_speed"` // 字节/秒
	UploadSpeed      int64 `json:"upload_speed"`   // 字节/秒
	TodayDownloads   int64 `json:"today_downloads"`
	WeekDownloads    int64 `json:"week_downloads"`
	MonthDownloads   int64 `json:"month_downloads"`
}

// Activity 活动记录
type Activity struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"` // subscribe, download, transfer, etc.
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Icon        string         `json:"icon"`
	Time        time.Time      `json:"time"`
	Extra       map[string]any `json:"extra"`
}

// ChartData 图表数据
type ChartData struct {
	Type   string           `json:"type"` // line, bar, pie
	Labels []string         `json:"labels"`
	Data   []map[string]any `json:"data"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Version     string  `json:"version"`
	Uptime      int64   `json:"uptime"`       // 秒
	CPUUsage    float64 `json:"cpu_usage"`    // 百分比
	MemoryUsage float64 `json:"memory_usage"` // 百分比
	DiskUsage   float64 `json:"disk_usage"`   // 百分比
	NetworkIn   int64   `json:"network_in"`   // 字节
	NetworkOut  int64   `json:"network_out"`  // 字节
}

// GetDashboardData 获取仪表板数据
func (s *service) GetDashboardData(ctx context.Context) (*DashboardData, error) {
	s.logger.Info("获取仪表板数据")

	// TODO: 实现实际的数据获取逻辑
	data := &DashboardData{
		Statistics: &Statistics{
			TotalSubscribes:  0,
			ActiveSubscribes: 0,
			TotalDownloads:   0,
			ActiveDownloads:  0,
		},
		RecentActivity: make([]*Activity, 0),
		Charts:         make(map[string]*ChartData),
		SystemInfo: &SystemInfo{
			Version: "1.0.0",
		},
	}

	return data, nil
}

// GetStatistics 获取统计信息
func (s *service) GetStatistics(ctx context.Context) (*Statistics, error) {
	s.logger.Info("获取统计信息")

	// TODO: 从数据库获取实际统计数据
	stats := &Statistics{
		TotalSubscribes:  0,
		ActiveSubscribes: 0,
		TotalDownloads:   0,
		ActiveDownloads:  0,
	}

	return stats, nil
}

// GetRecentActivity 获取最近活动
func (s *service) GetRecentActivity(ctx context.Context, limit int) ([]*Activity, error) {
	s.logger.Info("获取最近活动", zap.Int("limit", limit))

	// TODO: 从数据库获取实际活动记录
	activities := make([]*Activity, 0)

	return activities, nil
}

// GetChartData 获取图表数据
func (s *service) GetChartData(ctx context.Context, chartType string, days int) (*ChartData, error) {
	s.logger.Info("获取图表数据",
		zap.String("type", chartType),
		zap.Int("days", days),
	)

	// TODO: 根据类型生成不同的图表数据
	chart := &ChartData{
		Type:   chartType,
		Labels: make([]string, 0),
		Data:   make([]map[string]any, 0),
	}

	return chart, nil
}
