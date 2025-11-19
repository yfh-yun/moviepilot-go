// Package system 系统管理API处理器
// 提供系统管理和配置相关的HTTP API端点，包括事件系统、监控和任务跟踪
package system

import (
	"net/http"
	"strconv"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/event"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/monitor"
	"github.com/yfh-yun/moviepilot-go/internal/service/system"
	"github.com/yfh-yun/moviepilot-go/internal/task"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 系统处理器
type Handler struct {
	systemService *system.Service
	eventManager  *event.Manager
	monitor       *monitor.SystemMonitor
	taskManager   *task.Manager
	logger        *zap.Logger
}

// NewHandler 创建系统处理器
func NewHandler(
	systemService *system.Service,
	eventManager *event.Manager,
	monitor *monitor.SystemMonitor,
	taskManager *task.Manager,
) *Handler {
	return &Handler{
		systemService: systemService,
		eventManager:  eventManager,
		monitor:       monitor,
		taskManager:   taskManager,
		logger:        logger.Logger,
	}
}

// GetSystemInfo 获取系统信息
// @Summary 获取系统信息
// @Description 获取系统基本信息
// @Tags 系统管理
// @Produce json
// @Success 200 {object} response.Response{data=system.InfoResponse}
// @Router /api/v1/system/info [get]
func (h *Handler) GetSystemInfo(c *gin.Context) {
	info, err := h.systemService.GetSystemInfo(c.Request.Context())
	if err != nil {
		h.logger.Error("Get system info failed",
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "获取系统信息失败")
		return
	}

	response.Success(c, info)
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 系统健康状态检查
// @Tags 系统管理
// @Produce json
// @Success 200 {object} response.Response{data=system.HealthResponse}
// @Router /api/v1/system/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	health, err := h.systemService.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.Error("Health check failed",
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "健康检查失败")
		return
	}

	response.Success(c, health)
}

// GetVersion 获取版本信息
// @Summary 获取版本信息
// @Description 获取应用版本信息
// @Tags 系统管理
// @Produce json
// @Success 200 {object} response.Response{data=system.VersionResponse}
// @Router /api/v1/system/version [get]
func (h *Handler) GetVersion(c *gin.Context) {
	version, err := h.systemService.GetVersion(c.Request.Context())
	if err != nil {
		h.logger.Error("Get version failed",
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "获取版本信息失败")
		return
	}

	response.Success(c, version)
}

// GetSystemConfigRequest 获取系统配置请求
type GetSystemConfigRequest struct {
	Namespace string `form:"namespace"`
}

// GetSystemConfig 获取系统配置
// @Summary 获取系统配置
// @Description 获取系统配置信息
// @Tags 系统管理
// @Produce json
// @Param namespace query string false "配置命名空间"
// @Success 200 {object} response.Response{data=system.ConfigResponse}
// @Router /api/v1/system/config [get]
func (h *Handler) GetSystemConfig(c *gin.Context) {
	var req GetSystemConfigRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.CodeInvalidParams, err.Error())
		return
	}

	config, err := h.systemService.GetSystemConfig(c.Request.Context(), req.Namespace)
	if err != nil {
		h.logger.Error("Get system config failed",
			zap.String("namespace", req.Namespace),
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "获取系统配置失败")
		return
	}

	response.Success(c, config)
}

// UpdateSystemConfigRequest 更新系统配置请求
type UpdateSystemConfigRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Config    map[string]interface{} `json:"config" binding:"required"`
}

// UpdateSystemConfig 更新系统配置
// @Summary 更新系统配置
// @Description 更新系统配置信息
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param request body UpdateSystemConfigRequest true "更新配置请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/system/config [put]
func (h *Handler) UpdateSystemConfig(c *gin.Context) {
	var req UpdateSystemConfigRequest

	if err := validator.BindAndValidate(c, &req); err != nil {
		response.Error(c, response.CodeInvalidParams, err.Error())
		return
	}

	if err := h.systemService.UpdateSystemConfig(c.Request.Context(), req.Namespace, req.Config); err != nil {
		h.logger.Error("Update system config failed",
			zap.String("namespace", req.Namespace),
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "更新系统配置失败")
		return
	}

	response.Success(c, nil)
}

// BackupConfig 备份配置
// @Summary 备份配置
// @Description 备份系统配置文件
// @Tags 系统管理
// @Produce json
// @Success 200 {object} response.Response{data=system.BackupResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/system/config/backup [post]
func (h *Handler) BackupConfig(c *gin.Context) {
	backup, err := h.systemService.BackupConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("Backup config failed",
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "备份配置失败")
		return
	}

	response.Success(c, backup)
}

// RestoreConfigRequest 恢复配置请求
type RestoreConfigRequest struct {
	BackupID string `json:"backup_id" binding:"required"`
}

// RestoreConfig 恢复配置
// @Summary 恢复配置
// @Description 从备份恢复系统配置
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param request body RestoreConfigRequest true "恢复配置请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/system/config/restore [post]
func (h *Handler) RestoreConfig(c *gin.Context) {
	var req RestoreConfigRequest

	if err := validator.BindAndValidate(c, &req); err != nil {
		response.Error(c, response.CodeInvalidParams, err.Error())
		return
	}

	if err := h.systemService.RestoreConfig(c.Request.Context(), req.BackupID); err != nil {
		h.logger.Error("Restore config failed",
			zap.String("backup_id", req.BackupID),
			zap.Error(err),
		)

		if err == system.ErrBackupNotFound {
			response.Error(c, response.CodeNotFound, "备份文件不存在")
		} else {
			response.Error(c, response.CodeServerError, "恢复配置失败")
		}
		return
	}

	response.Success(c, nil)
}

// ResetConfigRequest 重置配置请求
type ResetConfigRequest struct {
	Namespace string `json:"namespace" binding:"required"`
}

// ResetConfig 重置配置
// @Summary 重置配置
// @Description 重置系统配置到默认值
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param request body ResetConfigRequest true "重置配置请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/system/config/reset [post]
func (h *Handler) ResetConfig(c *gin.Context) {
	var req ResetConfigRequest

	if err := validator.BindAndValidate(c, &req); err != nil {
		response.Error(c, response.CodeInvalidParams, err.Error())
		return
	}

	if err := h.systemService.ResetConfig(c.Request.Context(), req.Namespace); err != nil {
		h.logger.Error("Reset config failed",
			zap.String("namespace", req.Namespace),
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "重置配置失败")
		return
	}

	response.Success(c, nil)
}

// PublishEventRequest 发布事件请求
type PublishEventRequest struct {
	EventType string                 `json:"event_type" binding:"required"`
	Data      map[string]interface{} `json:"data"`
	Priority  int                    `json:"priority" default:"1"`
}

// PublishEvent 发布系统事件
// @Summary 发布系统事件
// @Description 发布系统事件，用于事件驱动架构
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param request body PublishEventRequest true "发布事件请求"
// @Success 200 {object} response.Response{data=event.PublishResult}
// @Router /api/v1/system/events/publish [post]
func (h *Handler) PublishEvent(c *gin.Context) {
	var req PublishEventRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		response.Error(c, response.CodeInvalidParams, err.Error())
		return
	}

	eventData := event.EventData{
		Type:     req.EventType,
		Data:     req.Data,
		Priority: req.Priority,
		Time:     time.Now(),
	}

	result, err := h.eventManager.Publish(eventData)
	if err != nil {
		h.logger.Error("Publish event failed",
			zap.String("event_type", req.EventType),
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "发布事件失败")
		return
	}

	response.Success(c, result)
}

// GetSystemMetrics 获取系统监控指标
// @Summary 获取系统监控指标
// @Description 获取系统运行时的监控指标数据
// @Tags 系统管理
// @Produce json
// @Param metric query string false "指标类型 (cpu,memory,disk,network)"
// @Param duration query string false "时间范围 (1h,24h,7d)"
// @Success 200 {object} response.Response{data=monitor.SystemMetrics}
// @Router /api/v1/system/metrics [get]
func (h *Handler) GetSystemMetrics(c *gin.Context) {
	metricType := c.DefaultQuery("metric", "")
	duration := c.DefaultQuery("duration", "1h")

	metrics, err := h.monitor.GetMetrics(metricType, duration)
	if err != nil {
		h.logger.Error("Get system metrics failed",
			zap.String("metric", metricType),
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "获取监控指标失败")
		return
	}

	response.Success(c, metrics)
}

// GetSystemAlerts 获取系统告警
// @Summary 获取系统告警
// @Description 获取系统运行时的告警信息
// @Tags 系统管理
// @Produce json
// @Param status query string false "告警状态 (active,resolved)"
// @Param severity query string false "严重程度 (critical,warning,info)"
// @Success 200 {object} response.Response{data=[]monitor.Alert}
// @Router /api/v1/system/alerts [get]
func (h *Handler) GetSystemAlerts(c *gin.Context) {
	status := c.Query("status")
	severity := c.Query("severity")

	alerts, err := h.monitor.GetAlerts(status, severity)
	if err != nil {
		h.logger.Error("Get system alerts failed",
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "获取系统告警失败")
		return
	}

	response.Success(c, alerts)
}

// GetTaskListRequest 获取任务列表请求
type GetTaskListRequest struct {
	Status   string `form:"status"` // 任务状态: pending,running,completed,failed
	Page     int    `form:"page" default:"1"`
	PageSize int    `form:"page_size" default:"20"`
}

// GetTaskList 获取任务列表
// @Summary 获取任务列表
// @Description 获取系统异步任务执行列表
// @Tags 系统管理
// @Produce json
// @Param status query string false "任务状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=task.ListResponse}
// @Router /api/v1/system/tasks [get]
func (h *Handler) GetTaskList(c *gin.Context) {
	var req GetTaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.CodeInvalidParams, err.Error())
		return
	}

	params := task.ListParams{
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	tasks, err := h.taskManager.GetTaskList(params)
	if err != nil {
		h.logger.Error("Get task list failed",
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "获取任务列表失败")
		return
	}

	response.Success(c, tasks)
}

// GetTaskDetail 获取任务详情
// @Summary 获取任务详情
// @Description 获取指定任务的详细信息
// @Tags 系统管理
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response{data=task.TaskDetail}
// @Router /api/v1/system/tasks/{task_id} [get]
func (h *Handler) GetTaskDetail(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, response.CodeInvalidParams, "任务ID不能为空")
		return
	}

	taskDetail, err := h.taskManager.GetTaskDetail(taskID)
	if err != nil {
		h.logger.Error("Get task detail failed",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		if err == task.ErrTaskNotFound {
			response.Error(c, response.CodeNotFound, "任务不存在")
		} else {
			response.Error(c, response.CodeServerError, "获取任务详情失败")
		}
		return
	}

	response.Success(c, taskDetail)
}

// CancelTask 取消任务
// @Summary 取消任务
// @Description 取消正在执行的任务
// @Tags 系统管理
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.Response
// @Router /api/v1/system/tasks/{task_id}/cancel [post]
func (h *Handler) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.Error(c, response.CodeInvalidParams, "任务ID不能为空")
		return
	}

	err := h.taskManager.CancelTask(taskID)
	if err != nil {
		h.logger.Error("Cancel task failed",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		if err == task.ErrTaskNotFound {
			response.Error(c, response.CodeNotFound, "任务不存在")
		} else if err == task.ErrTaskNotRunning {
			response.Error(c, response.CodeBadRequest, "任务不在运行状态")
		} else {
			response.Error(c, response.CodeServerError, "取消任务失败")
		}
		return
	}

	response.Success(c, nil)
}

// GetSystemStatistics 获取系统统计信息
// @Summary 获取系统统计信息
// @Description 获取系统运行统计信息
// @Tags 系统管理
// @Produce json
// @Success 200 {object} response.Response{data=system.StatisticsResponse}
// @Router /api/v1/system/statistics [get]
func (h *Handler) GetSystemStatistics(c *gin.Context) {
	stats, err := h.systemService.GetSystemStatistics(c.Request.Context())
	if err != nil {
		h.logger.Error("Get system statistics failed",
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "获取系统统计失败")
		return
	}

	response.Success(c, stats)
}

// UpdateSystemConfigWithEvent 更新配置并触发事件
// @Summary 更新配置并触发事件
// @Description 更新系统配置并触发配置变更事件
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param request body UpdateSystemConfigRequest true "更新配置请求"
// @Success 200 {object} response.Response
// @Router /api/v1/system/config/update-with-event [post]
func (h *Handler) UpdateSystemConfigWithEvent(c *gin.Context) {
	var req UpdateSystemConfigRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		response.Error(c, response.CodeInvalidParams, err.Error())
		return
	}

	// 更新配置
	if err := h.systemService.UpdateSystemConfig(c.Request.Context(), req.Namespace, req.Config); err != nil {
		h.logger.Error("Update system config failed",
			zap.String("namespace", req.Namespace),
			zap.Error(err),
		)
		response.Error(c, response.CodeServerError, "更新系统配置失败")
		return
	}

	// 触发配置变更事件
	eventData := event.EventData{
		Type: "config_updated",
		Data: map[string]interface{}{
			"namespace": req.Namespace,
			"config":    req.Config,
		},
		Priority: 2,
		Time:     time.Now(),
	}

	_, err := h.eventManager.Publish(eventData)
	if err != nil {
		h.logger.Warn("Failed to publish config update event",
			zap.Error(err),
		)
	}

	response.Success(c, nil)
}

// RegisterRoutes 注册系统相关路由
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	systemGroup := r.Group("/system")
	{
		// 公开路由
		systemGroup.GET("/info", h.GetSystemInfo)
		systemGroup.GET("/health", h.HealthCheck)
		systemGroup.GET("/version", h.GetVersion)

		// 需要认证的路由
		systemGroup.GET("/config", h.GetSystemConfig)
		systemGroup.PUT("/config", h.UpdateSystemConfig)
		systemGroup.POST("/config/backup", h.BackupConfig)
		systemGroup.POST("/config/restore", h.RestoreConfig)
		systemGroup.POST("/config/reset", h.ResetConfig)
		systemGroup.POST("/config/update-with-event", h.UpdateSystemConfigWithEvent)

		// 事件系统路由
		systemGroup.POST("/events/publish", h.PublishEvent)

		// 监控系统路由
		systemGroup.GET("/metrics", h.GetSystemMetrics)
		systemGroup.GET("/alerts", h.GetSystemAlerts)

		// 任务管理路由
		systemGroup.GET("/tasks", h.GetTaskList)
		systemGroup.GET("/tasks/:task_id", h.GetTaskDetail)
		systemGroup.POST("/tasks/:task_id/cancel", h.CancelTask)

		// 统计信息路由
		systemGroup.GET("/statistics", h.GetSystemStatistics)
	}
}
