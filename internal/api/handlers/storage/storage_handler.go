// Package storage Storage API处理器模块
package storage

import (
	"net/http"
	"strconv"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/service/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler Storage API处理器
type Handler struct {
	service storage.Service
	logger  *logger.Logger
}

// NewHandler 创建新的Storage处理器
func NewHandler(service storage.Service, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	storageGroup := router.Group("/storage")
	{
		// 存储管理
		storageGroup.GET("/", h.GetStorageList)
		storageGroup.GET("/:id", h.GetStorage)
		storageGroup.POST("/", h.CreateStorage)
		storageGroup.PUT("/:id", h.UpdateStorage)
		storageGroup.DELETE("/:id", h.DeleteStorage)
		storageGroup.POST("/:id/test", h.TestStorage)
		storageGroup.POST("/:id/scan", h.ScanStorage)

		// 存储统计
		storageGroup.GET("/:id/statistics", h.GetStorageStatistics)
		storageGroup.GET("/:id/health", h.GetStorageHealth)
		storageGroup.POST("/:id/cleanup", h.CleanupStorage)

		// 存储监控
		storageGroup.GET("/:id/monitor", h.GetStorageMonitor)
		storageGroup.POST("/:id/maintenance", h.MaintenanceStorage)
	}
}

// GetStorageList 获取存储列表
// @Summary 获取存储列表
// @Description 获取存储配置列表
// @Tags 存储
// @Produce json
// @Param type query string false "存储类型过滤"
// @Param enabled query bool false "是否启用过滤"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {object} Response{data=[]StorageInfo}
// @Router /storage [get]
func (h *Handler) GetStorageList(c *gin.Context) {
	storageType := c.Query("type")
	enabledParam := c.Query("enabled")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "20")

	var enabled *bool
	if enabledParam != "" {
		temp := enabledParam == "true"
		enabled = &temp
	}

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := storage.ListParams{
		Type:    storageType,
		Enabled: enabled,
		Page:    page,
		Count:   count,
	}

	storages, err := h.service.GetStorageList(params)
	if err != nil {
		h.logger.Error("Failed to get storage list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取存储列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    storages,
	})
}

// GetStorage 获取存储详情
// @Summary 获取存储详情
// @Description 获取存储配置详细信息
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Success 200 {object} Response{data=StorageDetail}
// @Router /storage/{id} [get]
func (h *Handler) GetStorage(c *gin.Context) {
	storageID := c.Param("id")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	storageDetail, err := h.service.GetStorage(storageID)
	if err != nil {
		h.logger.Error("Failed to get storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取存储详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    storageDetail,
	})
}

// CreateStorage 创建存储配置
// @Summary 创建存储配置
// @Description 创建新的存储配置
// @Tags 存储
// @Produce json
// @Param storage body CreateStorageRequest true "存储配置信息"
// @Success 200 {object} Response{data=StorageInfo}
// @Router /storage [post]
func (h *Handler) CreateStorage(c *gin.Context) {
	var request CreateStorageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	storageInfo, err := h.service.CreateStorage(request)
	if err != nil {
		h.logger.Error("Failed to create storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建存储配置失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    storageInfo,
		"message": "存储配置创建成功",
	})
}

// UpdateStorage 更新存储配置
// @Summary 更新存储配置
// @Description 更新存储配置信息
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Param storage body UpdateStorageRequest true "存储配置信息"
// @Success 200 {object} Response{data=StorageInfo}
// @Router /storage/{id} [put]
func (h *Handler) UpdateStorage(c *gin.Context) {
	storageID := c.Param("id")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	var request UpdateStorageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	storageInfo, err := h.service.UpdateStorage(storageID, request)
	if err != nil {
		h.logger.Error("Failed to update storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新存储配置失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    storageInfo,
		"message": "存储配置更新成功",
	})
}

// DeleteStorage 删除存储配置
// @Summary 删除存储配置
// @Description 删除存储配置
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Success 200 {object} Response
// @Router /storage/{id} [delete]
func (h *Handler) DeleteStorage(c *gin.Context) {
	storageID := c.Param("id")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	err := h.service.DeleteStorage(storageID)
	if err != nil {
		h.logger.Error("Failed to delete storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除存储配置失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "存储配置删除成功",
	})
}

// TestStorage 测试存储连接
// @Summary 测试存储连接
// @Description 测试存储配置的连接和可用性
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Success 200 {object} Response{data=TestResult}
// @Router /storage/{id}/test [post]
func (h *Handler) TestStorage(c *gin.Context) {
	storageID := c.Param("id")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	testResult, err := h.service.TestStorage(storageID)
	if err != nil {
		h.logger.Error("Failed to test storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "测试存储连接失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    testResult,
		"message": "存储连接测试完成",
	})
}

// ScanStorage 扫描存储内容
// @Summary 扫描存储内容
// @Description 扫描存储设备中的媒体内容
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Param deep query bool false "深度扫描" default(false)
// @Success 200 {object} Response{data=ScanResult}
// @Router /storage/{id}/scan [post]
func (h *Handler) ScanStorage(c *gin.Context) {
	storageID := c.Param("id")
	deepParam := c.DefaultQuery("deep", "false")

	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	deep := deepParam == "true"

	scanResult, err := h.service.ScanStorage(storageID, deep)
	if err != nil {
		h.logger.Error("Failed to scan storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "扫描存储内容失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    scanResult,
		"message": "存储内容扫描完成",
	})
}

// GetStorageStatistics 获取存储统计信息
// @Summary 获取存储统计信息
// @Description 获取存储设备的统计信息
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Success 200 {object} Response{data=StorageStatistics}
// @Router /storage/{id}/statistics [get]
func (h *Handler) GetStorageStatistics(c *gin.Context) {
	storageID := c.Param("id")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	statistics, err := h.service.GetStorageStatistics(storageID)
	if err != nil {
		h.logger.Error("Failed to get storage statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取存储统计信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    statistics,
	})
}

// GetStorageHealth 获取存储健康状态
// @Summary 获取存储健康状态
// @Description 获取存储设备的健康状态信息
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Success 200 {object} Response{data=StorageHealth}
// @Router /storage/{id}/health [get]
func (h *Handler) GetStorageHealth(c *gin.Context) {
	storageID := c.Param("id")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	health, err := h.service.GetStorageHealth(storageID)
	if err != nil {
		h.logger.Error("Failed to get storage health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取存储健康状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}

// CleanupStorage 清理存储空间
// @Summary 清理存储空间
// @Description 清理存储设备的无效文件和缓存
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Param type query string false "清理类型 (cache,temp,trash)" default(cache)
// @Success 200 {object} Response{data=CleanupResult}
// @Router /storage/{id}/cleanup [post]
func (h *Handler) CleanupStorage(c *gin.Context) {
	storageID := c.Param("id")
	cleanupType := c.DefaultQuery("type", "cache")

	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	result, err := h.service.CleanupStorage(storageID, cleanupType)
	if err != nil {
		h.logger.Error("Failed to cleanup storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "清理存储空间失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "存储空间清理完成",
	})
}

// GetStorageMonitor 获取存储监控信息
// @Summary 获取存储监控信息
// @Description 获取存储设备的实时监控信息
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Param duration query string false "监控时长 (1h,24h,7d)" default(1h)
// @Success 200 {object} Response{data=StorageMonitor}
// @Router /storage/{id}/monitor [get]
func (h *Handler) GetStorageMonitor(c *gin.Context) {
	storageID := c.Param("id")
	duration := c.DefaultQuery("duration", "1h")

	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	monitor, err := h.service.GetStorageMonitor(storageID, duration)
	if err != nil {
		h.logger.Error("Failed to get storage monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取存储监控信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    monitor,
	})
}

// MaintenanceStorage 维护存储设备
// @Summary 维护存储设备
// @Description 对存储设备进行维护操作
// @Tags 存储
// @Produce json
// @Param id path string true "存储ID"
// @Param operation body MaintenanceRequest true "维护操作"
// @Success 200 {object} Response{data=MaintenanceResult}
// @Router /storage/{id}/maintenance [post]
func (h *Handler) MaintenanceStorage(c *gin.Context) {
	storageID := c.Param("id")
	if storageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "存储ID不能为空",
		})
		return
	}

	var request MaintenanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	result, err := h.service.MaintenanceStorage(storageID, request)
	if err != nil {
		h.logger.Error("Failed to maintenance storage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "存储设备维护失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "存储设备维护完成",
	})
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// StorageInfo 存储信息结构
type StorageInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Path       string `json:"path"`
	Enabled    bool   `json:"enabled"`
	Status     string `json:"status"`
	TotalSize  int64  `json:"total_size"`
	UsedSize   int64  `json:"used_size"`
	FreeSize   int64  `json:"free_size"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

// StorageDetail 存储详情结构
type StorageDetail struct {
	StorageInfo
	Config      map[string]interface{} `json:"config"`
	Permissions map[string]interface{} `json:"permissions"`
	Health      string                 `json:"health"`
	LastScan    string                 `json:"last_scan"`
	ScannedSize int64                  `json:"scanned_size"`
	FileCount   int64                  `json:"file_count"`
	MediaCount  int64                  `json:"media_count"`
}

// CreateStorageRequest 创建存储请求结构
type CreateStorageRequest struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Path   string                 `json:"path"`
	Config map[string]interface{} `json:"config"`
}

// UpdateStorageRequest 更新存储请求结构
type UpdateStorageRequest struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Path    string                 `json:"path"`
	Config  map[string]interface{} `json:"config"`
	Enabled *bool                  `json:"enabled"`
}

// TestResult 测试结果结构
type TestResult struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Latency   int64  `json:"latency"`
	Available bool   `json:"available"`
	Capacity  int64  `json:"capacity"`
	Used      int64  `json:"used"`
	Free      int64  `json:"free"`
}

// ScanResult 扫描结果结构
type ScanResult struct {
	Status    string `json:"status"`
	Scanned   int64  `json:"scanned"`
	NewFiles  int64  `json:"new_files"`
	Updated   int64  `json:"updated"`
	Deleted   int64  `json:"deleted"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Duration  string `json:"duration"`
}

// StorageStatistics 存储统计结构
type StorageStatistics struct {
	TotalFiles    int64 `json:"total_files"`
	MediaFiles    int64 `json:"media_files"`
	CacheFiles    int64 `json:"cache_files"`
	TempFiles     int64 `json:"temp_files"`
	Duplicates    int64 `json:"duplicates"`
	TotalSize     int64 `json:"total_size"`
	MediaSize     int64 `json:"media_size"`
	CacheSize     int64 `json:"cache_size"`
	TempSize      int64 `json:"temp_size"`
	DuplicateSize int64 `json:"duplicate_size"`
}

// StorageHealth 存储健康结构
type StorageHealth struct {
	Status        string   `json:"status"`
	HealthScore   int      `json:"health_score"`
	DiskUsage     int      `json:"disk_usage"`
	IOPerformance int      `json:"io_performance"`
	NetworkStatus string   `json:"network_status"`
	LastCheck     string   `json:"last_check"`
	Issues        []string `json:"issues"`
}

// CleanupResult 清理结果结构
type CleanupResult struct {
	CleanedFiles int64    `json:"cleaned_files"`
	FreedSpace   int64    `json:"freed_space"`
	Duration     string   `json:"duration"`
	CleanedTypes []string `json:"cleaned_types"`
}

// StorageMonitor 存储监控结构
type StorageMonitor struct {
	IOUsage     []IOUsagePoint     `json:"io_usage"`
	DiskUsage   []DiskUsagePoint   `json:"disk_usage"`
	Temperature []TemperaturePoint `json:"temperature"`
	Network     []NetworkPoint     `json:"network"`
}

// MaintenanceRequest 维护请求结构
type MaintenanceRequest struct {
	Operation string            `json:"operation"`
	Params    map[string]string `json:"params"`
}

// MaintenanceResult 维护结果结构
type MaintenanceResult struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Duration  string `json:"duration"`
	Optimized bool   `json:"optimized"`
}

// IOUsagePoint IO使用率数据点
type IOUsagePoint struct {
	Timestamp string `json:"timestamp"`
	Read      int64  `json:"read"`
	Write     int64  `json:"write"`
}

// DiskUsagePoint 磁盘使用率数据点
type DiskUsagePoint struct {
	Timestamp string `json:"timestamp"`
	Used      int64  `json:"used"`
	Free      int64  `json:"free"`
}

// TemperaturePoint 温度数据点
type TemperaturePoint struct {
	Timestamp string `json:"timestamp"`
	Temp      int    `json:"temp"`
}

// NetworkPoint 网络数据点
type NetworkPoint struct {
	Timestamp string `json:"timestamp"`
	Upload    int64  `json:"upload"`
	Download  int64  `json:"download"`
}
