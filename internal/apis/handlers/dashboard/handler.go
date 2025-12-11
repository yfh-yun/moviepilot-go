package dashboard

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	"moviepilot-go/internal/business/services/dashboard"
	"moviepilot-go/internal/models/dto"
)

// Handler Dashboard API 处理器
type Handler struct {
	dashboardService *dashboard.DashboardService
	logger           *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(dashboardService *dashboard.DashboardService, logger *zap.Logger) *Handler {
	return &Handler{
		dashboardService: dashboardService,
		logger:           logger,
	}
}

// GetStatistic 获取媒体数量统计
// @Summary 媒体数量统计
// @Description 查询媒体数量统计信息
// @Tags dashboard
// @Produce json
// @Param name query string false "媒体服务器名称"
// @Success 200 {object} dto.Statistic
// @Router /api/dashboard/statistic [get]
func (h *Handler) GetStatistic(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	name := c.Query("name")
	stat, err := h.dashboardService.GetStatistic(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("获取统计信息失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusOK, &dto.Statistic{})
		return
	}

	h.logger.Info("获取统计信息成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("name", name),
	)
	c.JSON(http.StatusOK, stat)
}

// GetStatistic2 获取媒体数量统计（API_TOKEN）
// @Summary 媒体数量统计（API_TOKEN）
// @Description 查询媒体数量统计信息 API_TOKEN认证
// @Tags dashboard
// @Produce json
// @Param token query string true "API Token"
// @Success 200 {object} dto.Statistic
// @Router /api/dashboard/statistic2 [get]
func (h *Handler) GetStatistic2(c *gin.Context) {
	h.GetStatistic(c)
}

// GetStorage 获取本地存储空间
// @Summary 本地存储空间
// @Description 查询本地存储空间信息
// @Tags dashboard
// @Produce json
// @Success 200 {object} dto.Storage
// @Router /api/dashboard/storage [get]
func (h *Handler) GetStorage(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	storage, err := h.dashboardService.GetStorage(c.Request.Context())
	if err != nil {
		h.logger.Error("获取存储信息失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusOK, &dto.Storage{})
		return
	}

	h.logger.Info("获取存储信息成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, storage)
}

// GetStorage2 获取本地存储空间（API_TOKEN）
// @Summary 本地存储空间（API_TOKEN）
// @Description 查询本地存储空间信息 API_TOKEN认证
// @Tags dashboard
// @Produce json
// @Param token query string true "API Token"
// @Success 200 {object} dto.Storage
// @Router /api/dashboard/storage2 [get]
func (h *Handler) GetStorage2(c *gin.Context) {
	h.GetStorage(c)
}

// GetProcesses 获取进程信息
// @Summary 进程信息
// @Description 查询进程信息
// @Tags dashboard
// @Produce json
// @Success 200 {array} dto.ProcessInfo
// @Router /api/dashboard/processes [get]
func (h *Handler) GetProcesses(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	processes, err := h.dashboardService.GetProcesses(c.Request.Context())
	if err != nil {
		h.logger.Error("获取进程信息失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusOK, []dto.ProcessInfo{})
		return
	}

	h.logger.Info("获取进程信息成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, processes)
}

// GetDownloaderInfo 获取下载器信息
// @Summary 下载器信息
// @Description 查询下载器信息
// @Tags dashboard
// @Produce json
// @Param name query string false "下载器名称"
// @Success 200 {object} dto.DownloaderInfo
// @Router /api/dashboard/downloader [get]
func (h *Handler) GetDownloaderInfo(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	name := c.Query("name")
	info, err := h.dashboardService.GetDownloaderInfo(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("获取下载器信息失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("name", name),
		)
		c.JSON(http.StatusOK, &dto.DownloaderInfo{})
		return
	}

	h.logger.Info("获取下载器信息成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("name", name),
	)
	c.JSON(http.StatusOK, info)
}

// GetDownloaderInfo2 获取下载器信息（API_TOKEN）
// @Summary 下载器信息（API_TOKEN）
// @Description 查询下载器信息 API_TOKEN认证
// @Tags dashboard
// @Produce json
// @Param token query string true "API Token"
// @Param name query string false "下载器名称"
// @Success 200 {object} dto.DownloaderInfo
// @Router /api/dashboard/downloader2 [get]
func (h *Handler) GetDownloaderInfo2(c *gin.Context) {
	h.GetDownloaderInfo(c)
}

// GetScheduleInfo 获取定时任务信息
// @Summary 定时任务信息
// @Description 查询定时任务信息
// @Tags dashboard
// @Produce json
// @Success 200 {array} dto.ScheduleInfo
// @Router /api/dashboard/schedule [get]
func (h *Handler) GetScheduleInfo(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	schedules, err := h.dashboardService.GetScheduleInfo(c.Request.Context())
	if err != nil {
		h.logger.Error("获取定时任务信息失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusOK, []dto.ScheduleInfo{})
		return
	}

	h.logger.Info("获取定时任务信息成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, schedules)
}

// GetScheduleInfo2 获取定时任务信息（API_TOKEN）
// @Summary 定时任务信息（API_TOKEN）
// @Description 查询定时任务信息 API_TOKEN认证
// @Tags dashboard
// @Produce json
// @Param token query string true "API Token"
// @Success 200 {array} dto.ScheduleInfo
// @Router /api/dashboard/schedule2 [get]
func (h *Handler) GetScheduleInfo2(c *gin.Context) {
	h.GetScheduleInfo(c)
}

// GetCPUUsage 获取CPU使用率
// @Summary 获取当前CPU使用率
// @Description 获取当前系统CPU使用率
// @Tags dashboard
// @Produce json
// @Success 200 {integer} int
// @Router /api/dashboard/cpu [get]
func (h *Handler) GetCPUUsage(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	cpuUsage, err := h.dashboardService.GetCPUUsage(c.Request.Context())
	if err != nil {
		h.logger.Error("获取CPU使用率失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusOK, 0)
		return
	}

	h.logger.Info("获取CPU使用率成功",
		zap.Int("cpu_usage", cpuUsage),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, cpuUsage)
}

// GetCPUUsage2 获取CPU使用率（API_TOKEN）
// @Summary 获取当前CPU使用率（API_TOKEN）
// @Description 获取当前CPU使用率 API_TOKEN认证
// @Tags dashboard
// @Produce json
// @Param token query string true "API Token"
// @Success 200 {integer} int
// @Router /api/dashboard/cpu2 [get]
func (h *Handler) GetCPUUsage2(c *gin.Context) {
	h.GetCPUUsage(c)
}

// GetMemoryUsage 获取内存使用情况
// @Summary 获取当前内存使用量和使用率
// @Description 获取当前内存使用量和使用率
// @Tags dashboard
// @Produce json
// @Success 200 {array} int
// @Router /api/dashboard/memory [get]
func (h *Handler) GetMemoryUsage(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	memoryUsage, err := h.dashboardService.GetMemoryUsage(c.Request.Context())
	if err != nil {
		h.logger.Error("获取内存使用情况失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusOK, []int{0, 0, 0})
		return
	}

	h.logger.Info("获取内存使用情况成功",
		zap.Ints("memory_usage", memoryUsage),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, memoryUsage)
}

// GetMemoryUsage2 获取内存使用情况（API_TOKEN）
// @Summary 获取当前内存使用量和使用率（API_TOKEN）
// @Description 获取当前内存使用量和使用率 API_TOKEN认证
// @Tags dashboard
// @Produce json
// @Param token query string true "API Token"
// @Success 200 {array} int
// @Router /api/dashboard/memory2 [get]
func (h *Handler) GetMemoryUsage2(c *gin.Context) {
	h.GetMemoryUsage(c)
}

// GetNetworkUsage 获取网络流量
// @Summary 获取当前网络流量
// @Description 获取当前网络流量（上行和下行流量，单位：bytes/s）
// @Tags dashboard
// @Produce json
// @Success 200 {array} int
// @Router /api/dashboard/network [get]
func (h *Handler) GetNetworkUsage(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	networkUsage, err := h.dashboardService.GetNetworkUsage(c.Request.Context())
	if err != nil {
		h.logger.Error("获取网络流量失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusOK, []int{0, 0})
		return
	}

	h.logger.Info("获取网络流量成功",
		zap.Ints("network_usage", networkUsage),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, networkUsage)
}

// GetNetworkUsage2 获取网络流量（API_TOKEN）
// @Summary 获取当前网络流量（API_TOKEN）
// @Description 获取当前网络流量 API_TOKEN认证
// @Tags dashboard
// @Produce json
// @Param token query string true "API Token"
// @Success 200 {array} int
// @Router /api/dashboard/network2 [get]
func (h *Handler) GetNetworkUsage2(c *gin.Context) {
	h.GetNetworkUsage(c)
}

// GetTransferStatistic 获取文件整理统计
// @Summary 文件整理统计
// @Description 查询文件整理统计信息
// @Tags dashboard
// @Produce json
// @Param days query int false "统计天数" default(7)
// @Success 200 {array} int
// @Router /api/dashboard/transfer [get]
func (h *Handler) GetTransferStatistic(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	daysStr := c.DefaultQuery("days", "7")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 7
	}

	transferStat, err := h.dashboardService.GetTransferStatistic(c.Request.Context(), days)
	if err != nil {
		h.logger.Error("获取文件整理统计失败",
			zap.Error(err),
			zap.Int("days", days),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusOK, []int{})
		return
	}

	h.logger.Info("获取文件整理统计成功",
		zap.Int("days", days),
		zap.Ints("transfer_stat", transferStat),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, transferStat)
}
