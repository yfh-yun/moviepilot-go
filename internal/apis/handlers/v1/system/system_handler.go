// Package system 系统管理API处理器
package system

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/response"
	"moviepilot-go/pkg/validator"
)

// SystemHandler 系统管理处理器
// 提供系统配置、健康检查、日志管理、备份恢复等系统管理功能
type SystemHandler struct {
	systemService service.SystemService
	logger        *zap.Logger
}

// NewSystemHandler 创建系统管理处理器
func NewSystemHandler(systemService service.SystemService, logger *zap.Logger) *SystemHandler {
	return &SystemHandler{
		systemService: systemService,
		logger:        logger,
	}
}

// SystemInfoResponse 系统信息响应结构体
type SystemInfoResponse struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	GoVersion string            `json:"go_version"`
	BuildTime string            `json:"build_time"`
	GitCommit string            `json:"git_commit"`
	Runtime   RuntimeInfo       `json:"runtime"`
	System    SystemStats       `json:"system"`
	Services  map[string]string `json:"services"`
}

// RuntimeInfo 运行时信息结构体
type RuntimeInfo struct {
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	MemoryUsage  string `json:"memory_usage"`
}

// SystemStats 系统统计信息结构体
type SystemStats struct {
	Uptime       string       `json:"uptime"`
	CPUUsage     string       `json:"cpu_usage"`
	MemoryUsage  string       `json:"memory_usage"`
	DiskUsage    string       `json:"disk_usage"`
	NetworkStats NetworkStats `json:"network_stats"`
}

// NetworkStats 网络统计信息结构体
type NetworkStats struct {
	BytesReceived     int64 `json:"bytes_received"`
	BytesSent         int64 `json:"bytes_sent"`
	Requests          int64 `json:"requests"`
	ActiveConnections int   `json:"active_connections"`
}

// GetSystemInfo 获取系统信息
// @Summary 获取系统信息
// @Description 获取系统基本信息、运行时信息和系统统计
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=SystemInfoResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/info [get]
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取系统基本信息
	systemInfo, err := h.systemService.GetSystemInfo(ctx)
	if err != nil {
		h.logger.Error("获取系统信息失败", zap.Error(err))
		response.InternalServerError(c, "获取系统信息失败")
		return
	}

	// 获取运行时信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	runtimeInfo := RuntimeInfo{
		GOOS:         runtime.GOOS,
		GOARCH:       runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemoryUsage:  fmt.Sprintf("%dMB", memStats.Alloc/1024/1024),
	}

	// 获取系统统计信息
	systemStats, err := h.systemService.GetSystemStats(ctx)
	if err != nil {
		h.logger.Warn("获取系统统计信息失败", zap.Error(err))
		// 不返回错误，使用默认值
		systemStats = service.SystemStats{}
	}

	responseInfo := SystemInfoResponse{
		Name:      systemInfo.Name,
		Version:   systemInfo.Version,
		GoVersion: systemInfo.GoVersion,
		BuildTime: systemInfo.BuildTime,
		GitCommit: systemInfo.GitCommit,
		Runtime:   runtimeInfo,
		System: SystemStats{
			Uptime:      systemStats.Uptime,
			CPUUsage:    systemStats.CPUUsage,
			MemoryUsage: systemStats.MemoryUsage,
			DiskUsage:   systemStats.DiskUsage,
			NetworkStats: NetworkStats{
				BytesReceived:     systemStats.NetworkStats.BytesReceived,
				BytesSent:         systemStats.NetworkStats.BytesSent,
				Requests:          systemStats.NetworkStats.Requests,
				ActiveConnections: systemStats.NetworkStats.ActiveConnections,
			},
		},
		Services: systemInfo.Services,
	}

	response.Success(c, responseInfo)
}

// HealthCheckResponse 健康检查响应结构体
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp int64             `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	Database  string            `json:"database"`
	Cache     string            `json:"cache"`
	Storage   string            `json:"storage"`
}

// HealthCheck 系统健康检查
// @Summary 系统健康检查
// @Description 检查系统各个组件的健康状态
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=HealthCheckResponse}
// @Failure 503 {object} response.ErrorResponse
// @Router /api/v1/system/health [get]
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	// 检查系统健康状态
	healthStatus, err := h.systemService.CheckHealth(ctx)
	if err != nil {
		h.logger.Error("系统健康检查失败", zap.Error(err))
		response.ServiceUnavailable(c, "系统服务不可用")
		return
	}

	responseData := HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Version:   "2.8.1",
		Services:  healthStatus.Services,
		Database:  healthStatus.Database,
		Cache:     healthStatus.Cache,
		Storage:   healthStatus.Storage,
	}

	// 如果有服务不健康，返回503
	if healthStatus.Database != "healthy" || healthStatus.Cache != "healthy" {
		responseData.Status = "unhealthy"
		response.ServiceUnavailable(c, responseData)
		return
	}

	response.Success(c, responseData)
}

// GetSystemConfig 获取系统配置
// @Summary 获取系统配置
// @Description 获取当前系统的配置信息
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=map[string]interface{}}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/config [get]
func (h *SystemHandler) GetSystemConfig(c *gin.Context) {
	ctx := c.Request.Context()

	config, err := h.systemService.GetSystemConfig(ctx)
	if err != nil {
		h.logger.Error("获取系统配置失败", zap.Error(err))
		response.InternalServerError(c, "获取系统配置失败")
		return
	}

	response.Success(c, config)
}

// UpdateSystemConfigRequest 更新系统配置请求结构体
type UpdateSystemConfigRequest struct {
	Config map[string]interface{} `json:"config" binding:"required"`
}

// UpdateSystemConfig 更新系统配置
// @Summary 更新系统配置
// @Description 更新系统配置信息
// @Tags system
// @Accept json
// @Produce json
// @Param request body UpdateSystemConfigRequest true "配置数据"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/config [put]
func (h *SystemHandler) UpdateSystemConfig(c *gin.Context) {
	var req UpdateSystemConfigRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("更新系统配置请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("更新系统配置请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}

	ctx := c.Request.Context()

	err := h.systemService.UpdateSystemConfig(ctx, req.Config)
	if err != nil {
		h.logger.Error("更新系统配置失败", zap.Error(err))
		response.InternalServerError(c, "更新系统配置失败")
		return
	}

	logger.Info("系统配置更新成功", zap.Any("config", req.Config))
	response.Success(c, gin.H{
		"message": "系统配置更新成功",
	})
}

// BackupConfig 备份系统配置
// @Summary 备份系统配置
// @Description 备份当前系统配置到指定位置
// @Tags system
// @Accept json
// @Produce json
// @Param path query string false "备份路径"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/backup [post]
func (h *SystemHandler) BackupConfig(c *gin.Context) {
	backupPath := c.Query("path")
	ctx := c.Request.Context()

	backupFile, err := h.systemService.BackupConfig(ctx, backupPath)
	if err != nil {
		h.logger.Error("备份系统配置失败", zap.Error(err))
		response.InternalServerError(c, "备份系统配置失败")
		return
	}

	logger.Info("系统配置备份成功", zap.String("backup_file", backupFile))
	response.Success(c, gin.H{
		"message":     "系统配置备份成功",
		"backup_file": backupFile,
	})
}

// RestoreConfigRequest 恢复配置请求结构体
type RestoreConfigRequest struct {
	BackupFile string `json:"backup_file" binding:"required"`
}

// RestoreConfig 恢复系统配置
// @Summary 恢复系统配置
// @Description 从备份文件恢复系统配置
// @Tags system
// @Accept json
// @Produce json
// @Param request body RestoreConfigRequest true "恢复配置数据"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/restore [post]
func (h *SystemHandler) RestoreConfig(c *gin.Context) {
	var req RestoreConfigRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("恢复系统配置请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("恢复系统配置请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}

	ctx := c.Request.Context()

	err := h.systemService.RestoreConfig(ctx, req.BackupFile)
	if err != nil {
		h.logger.Error("恢复系统配置失败", zap.Error(err), zap.String("backup_file", req.BackupFile))
		response.InternalServerError(c, "恢复系统配置失败")
		return
	}

	logger.Info("系统配置恢复成功", zap.String("backup_file", req.BackupFile))
	response.Success(c, gin.H{
		"message": "系统配置恢复成功",
	})
}

// ResetConfig 重置系统配置
// @Summary 重置系统配置
// @Description 重置系统配置为默认值
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/reset [post]
func (h *SystemHandler) ResetConfig(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.systemService.ResetConfig(ctx)
	if err != nil {
		h.logger.Error("重置系统配置失败", zap.Error(err))
		response.InternalServerError(c, "重置系统配置失败")
		return
	}

	logger.Info("系统配置重置成功")
	response.Success(c, gin.H{
		"message": "系统配置重置成功",
	})
}

// GetLogsRequest 获取日志请求结构体
type GetLogsRequest struct {
	Level   string `form:"level" validate:"omitempty,oneof=debug info warn error"`
	Service string `form:"service" validate:"omitempty,max=50"`
	Limit   int    `form:"limit" validate:"omitempty,min=1,max=1000"`
	Offset  int    `form:"offset" validate:"omitempty,min=0"`
}

// LogEntry 日志条目结构体
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Message   string `json:"message"`
	Context   string `json:"context"`
}

// GetLogsResponse 获取日志响应结构体
type GetLogsResponse struct {
	Logs   []LogEntry `json:"logs"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// GetLogs 获取系统日志
// @Summary 获取系统日志
// @Description 获取系统日志，支持按级别、服务、分页查询
// @Tags system
// @Accept json
// @Produce json
// @Param request query GetLogsRequest false "日志查询参数"
// @Success 200 {object} response.SuccessResponse{data=GetLogsResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/logs [get]
func (h *SystemHandler) GetLogs(c *gin.Context) {
	var req GetLogsRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("获取日志请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("获取日志请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}

	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	ctx := c.Request.Context()

	// 调用服务层获取日志
	logEntries, total, err := h.systemService.GetLogs(ctx, service.GetLogsParams{
		Level:   req.Level,
		Service: req.Service,
		Limit:   req.Limit,
		Offset:  req.Offset,
	})

	if err != nil {
		h.logger.Error("获取系统日志失败", zap.Error(err))
		response.InternalServerError(c, "获取系统日志失败")
		return
	}

	// 转换为响应格式
	var responseLogs []LogEntry
	for _, entry := range logEntries {
		responseLogs = append(responseLogs, LogEntry{
			Timestamp: entry.Timestamp.Format("2006-01-02 15:04:05"),
			Level:     entry.Level,
			Service:   entry.Service,
			Message:   entry.Message,
			Context:   entry.Context,
		})
	}

	response.Success(c, GetLogsResponse{
		Logs:   responseLogs,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	})
}

// ClearLogs 清空系统日志
// @Summary 清空系统日志
// @Description 清空系统日志文件
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/logs/clear [delete]
func (h *SystemHandler) ClearLogs(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.systemService.ClearLogs(ctx)
	if err != nil {
		h.logger.Error("清空系统日志失败", zap.Error(err))
		response.InternalServerError(c, "清空系统日志失败")
		return
	}

	logger.Info("系统日志清空成功")
	response.Success(c, gin.H{
		"message": "系统日志清空成功",
	})
}

// SystemStatsResponse 系统统计响应结构体
type SystemStatsResponse struct {
	Uptime        string                  `json:"uptime"`
	CPUUsage      string                  `json:"cpu_usage"`
	MemoryUsage   string                  `json:"memory_usage"`
	DiskUsage     string                  `json:"disk_usage"`
	NetworkStats  NetworkStats            `json:"network_stats"`
	ServiceStats  map[string]ServiceStats `json:"service_stats"`
	DatabaseStats DatabaseStats           `json:"database_stats"`
	CacheStats    CacheStats              `json:"cache_stats"`
}

// ServiceStats 服务统计结构体
type ServiceStats struct {
	Requests     int64  `json:"requests"`
	Errors       int64  `json:"errors"`
	ResponseTime string `json:"response_time"`
	Uptime       string `json:"uptime"`
}

// DatabaseStats 数据库统计结构体
type DatabaseStats struct {
	Connections int    `json:"connections"`
	Queries     int64  `json:"queries"`
	SlowQueries int64  `json:"slow_queries"`
	Status      string `json:"status"`
}

// CacheStats 缓存统计结构体
type CacheStats struct {
	Hits        int64  `json:"hits"`
	Misses      int64  `json:"misses"`
	HitRate     string `json:"hit_rate"`
	MemoryUsage string `json:"memory_usage"`
}

// GetSystemStats 获取系统统计信息
// @Summary 获取系统统计信息
// @Description 获取详细的系统统计信息，包括性能指标和使用情况
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=SystemStatsResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/system/stats [get]
func (h *SystemHandler) GetSystemStats(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取系统统计信息
	systemStats, err := h.systemService.GetSystemStats(ctx)
	if err != nil {
		h.logger.Error("获取系统统计信息失败", zap.Error(err))
		response.InternalServerError(c, "获取系统统计信息失败")
		return
	}

	// 获取服务统计信息
	serviceStats, err := h.systemService.GetServiceStats(ctx)
	if err != nil {
		h.logger.Warn("获取服务统计信息失败", zap.Error(err))
		serviceStats = make(map[string]service.ServiceStats)
	}

	// 获取数据库统计信息
	databaseStats, err := h.systemService.GetDatabaseStats(ctx)
	if err != nil {
		h.logger.Warn("获取数据库统计信息失败", zap.Error(err))
		databaseStats = service.DatabaseStats{}
	}

	// 获取缓存统计信息
	cacheStats, err := h.systemService.GetCacheStats(ctx)
	if err != nil {
		h.logger.Warn("获取缓存统计信息失败", zap.Error(err))
		cacheStats = service.CacheStats{}
	}

	// 转换为响应格式
	responseServiceStats := make(map[string]ServiceStats)
	for serviceName, stats := range serviceStats {
		responseServiceStats[serviceName] = ServiceStats{
			Requests:     stats.Requests,
			Errors:       stats.Errors,
			ResponseTime: stats.ResponseTime,
			Uptime:       stats.Uptime,
		}
	}

	responseData := SystemStatsResponse{
		Uptime:      systemStats.Uptime,
		CPUUsage:    systemStats.CPUUsage,
		MemoryUsage: systemStats.MemoryUsage,
		DiskUsage:   systemStats.DiskUsage,
		NetworkStats: NetworkStats{
			BytesReceived:     systemStats.NetworkStats.BytesReceived,
			BytesSent:         systemStats.NetworkStats.BytesSent,
			Requests:          systemStats.NetworkStats.Requests,
			ActiveConnections: systemStats.NetworkStats.ActiveConnections,
		},
		ServiceStats: responseServiceStats,
		DatabaseStats: DatabaseStats{
			Connections: databaseStats.Connections,
			Queries:     databaseStats.Queries,
			SlowQueries: databaseStats.SlowQueries,
			Status:      databaseStats.Status,
		},
		CacheStats: CacheStats{
			Hits:        cacheStats.Hits,
			Misses:      cacheStats.Misses,
			HitRate:     cacheStats.HitRate,
			MemoryUsage: cacheStats.MemoryUsage,
		},
	}

	response.Success(c, responseData)
}
