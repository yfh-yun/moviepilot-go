// Package dashboard Dashboard API处理器模块
package dashboard

import (
	"net/http"
	"strconv"

	"github.com/yfh-yun/moviepilot-go/internal/service/dashboard"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler Dashboard API处理器
type Handler struct {
	service dashboard.Service
	logger  *zap.Logger
}

// NewHandler 创建新的Dashboard处理器
func NewHandler(service dashboard.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	dashboardGroup := router.Group("/dashboard")
	{
		dashboardGroup.GET("/overview", h.GetOverview)
		dashboardGroup.GET("/statistics", h.GetStatistics)
		dashboardGroup.GET("/recent", h.GetRecentActivities)
		dashboardGroup.GET("/charts", h.GetChartsData)
	}
}

// GetOverview 获取仪表板总览
// @Summary 获取仪表板总览
// @Description 获取仪表板总览统计信息
// @Tags 仪表板
// @Produce json
// @Success 200 {object} Response{data=DashboardOverview}
// @Router /dashboard/overview [get]
func (h *Handler) GetOverview(c *gin.Context) {
	overview, err := h.service.GetOverview()
	if err != nil {
		h.logger.Error("Failed to get dashboard overview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取仪表板总览失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    overview,
	})
}

// GetStatistics 获取详细统计
// @Summary 获取详细统计
// @Description 获取系统详细统计信息
// @Tags 仪表板
// @Produce json
// @Param period query string false "统计周期 (day,week,month,year)" default(week)
// @Param type query string false "统计类型 (download,transfer,subscribe,media)" default(download)
// @Success 200 {object} Response{data=DashboardStatistics}
// @Router /dashboard/statistics [get]
func (h *Handler) GetStatistics(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	statType := c.DefaultQuery("type", "download")

	statistics, err := h.service.GetStatistics(statType, period)
	if err != nil {
		h.logger.Error("Failed to get dashboard statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取详细统计失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    statistics,
	})
}

// GetRecentActivities 获取最近活动
// @Summary 获取最近活动
// @Description 获取最近的活动记录
// @Tags 仪表板
// @Produce json
// @Param limit query int false "数量限制" default(20)
// @Success 200 {object} Response{data=[]RecentActivity}
// @Router /dashboard/recent [get]
func (h *Handler) GetRecentActivities(c *gin.Context) {
	limitParam := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitParam)

	activities, err := h.service.GetRecentActivities(limit)
	if err != nil {
		h.logger.Error("Failed to get recent activities", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取最近活动失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}

// GetChartsData 获取图表数据
// @Summary 获取图表数据
// @Description 获取用于图表展示的数据
// @Tags 仪表板
// @Produce json
// @Param chart_type query string false "图表类型 (download,transfer,media)" default(download)
// @Param period query string false "统计周期 (day,week,month,year)" default(month)
// @Success 200 {object} Response{data=ChartData}
// @Router /dashboard/charts [get]
func (h *Handler) GetChartsData(c *gin.Context) {
	chartType := c.DefaultQuery("chart_type", "download")
	period := c.DefaultQuery("period", "month")

	chartData, err := h.service.GetChartsData(chartType, period)
	if err != nil {
		h.logger.Error("Failed to get charts data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取图表数据失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    chartData,
	})
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// DashboardOverview 仪表板总览结构
type DashboardOverview struct {
	TotalDownloads  int64 `json:"total_downloads"`
	ActiveDownloads int64 `json:"active_downloads"`
	CompletedToday  int64 `json:"completed_today"`
	FailedDownloads int64 `json:"failed_downloads"`

	TotalTransfers     int64 `json:"total_transfers"`
	ActiveTransfers    int64 `json:"active_transfers"`
	CompletedTransfers int64 `json:"completed_transfers"`
	FailedTransfers    int64 `json:"failed_transfers"`

	TotalMedia   int64 `json:"total_media"`
	MoviesCount  int64 `json:"movies_count"`
	TVShowsCount int64 `json:"tv_shows_count"`
	AnimeCount   int64 `json:"anime_count"`

	SystemInfo  SystemOverview `json:"system_info"`
	RecentStats RecentStats    `json:"recent_stats"`
}

// SystemOverview 系统总览
type SystemOverview struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkUsage float64 `json:"network_usage"`
	Uptime       string  `json:"uptime"`
}

// RecentStats 最近统计
type RecentStats struct {
	DownloadsLast24h  int64 `json:"downloads_last_24h"`
	TransfersLast24h  int64 `json:"transfers_last_24h"`
	MediaAddedLast24h int64 `json:"media_added_last_24h"`
}

// DashboardStatistics 详细统计结构
type DashboardStatistics struct {
	Period string `json:"period"`
	Type   string `json:"type"`

	// 下载统计
	DownloadStats DownloadStats `json:"download_stats"`

	// 转移统计
	TransferStats TransferStats `json:"transfer_stats"`

	// 订阅统计
	SubscribeStats SubscribeStats `json:"subscribe_stats"`

	// 媒体统计
	MediaStats MediaStats `json:"media_stats"`
}

// DownloadStats 下载统计
type DownloadStats struct {
	Total        int64   `json:"total"`
	Completed    int64   `json:"completed"`
	Failed       int64   `json:"failed"`
	Active       int64   `json:"active"`
	SuccessRate  float64 `json:"success_rate"`
	AverageSpeed int64   `json:"average_speed"`
	TotalSize    int64   `json:"total_size"`
}

// TransferStats 转移统计
type TransferStats struct {
	Total        int64   `json:"total"`
	Completed    int64   `json:"completed"`
	Failed       int64   `json:"failed"`
	Active       int64   `json:"active"`
	SuccessRate  float64 `json:"success_rate"`
	AverageSpeed int64   `json:"average_speed"`
	TotalSize    int64   `json:"total_size"`
}

// SubscribeStats 订阅统计
type SubscribeStats struct {
	Total           int64   `json:"total"`
	Active          int64   `json:"active"`
	Completed       int64   `json:"completed"`
	Failed          int64   `json:"failed"`
	SuccessRate     float64 `json:"success_rate"`
	AverageInterval int64   `json:"average_interval"`
}

// MediaStats 媒体统计
type MediaStats struct {
	TotalMovies     int64 `json:"total_movies"`
	TotalTVShows    int64 `json:"total_tv_shows"`
	TotalEpisodes   int64 `json:"total_episodes"`
	TotalAnime      int64 `json:"total_anime"`
	MediaAddedToday int64 `json:"media_added_today"`
	MediaAddedWeek  int64 `json:"media_added_week"`
	MediaAddedMonth int64 `json:"media_added_month"`
}

// RecentActivity 最近活动结构
type RecentActivity struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

// ChartData 图表数据结构
type ChartData struct {
	Type     string         `json:"type"`
	Period   string         `json:"period"`
	Labels   []string       `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

// ChartDataset 图表数据集
type ChartDataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor string    `json:"backgroundColor"`
	BorderColor     string    `json:"borderColor"`
}
