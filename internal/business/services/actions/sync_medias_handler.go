package actions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/errors"
)

// MediaSyncHandler 媒体同步处理器接口
type MediaSyncHandler interface {
	// SyncMedia 同步单个媒体
	SyncMedia(c *gin.Context)
	// BatchSyncMedias 批量同步媒体
	BatchSyncMedias(c *gin.Context)
	// GetSyncTask 获取同步任务信息
	GetSyncTask(c *gin.Context)
	// ListSyncTasks 列出同步任务
	ListSyncTasks(c *gin.Context)
	// CancelSyncTask 取消同步任务
	CancelSyncTask(c *gin.Context)
	// GetSyncStats 获取同步统计信息
	GetSyncStats(c *gin.Context)
	// ResolveConflict 解决同步冲突
	ResolveConflict(c *gin.Context)
}

// mediaSyncHandler 媒体同步处理器实现
type mediaSyncHandler struct {
	syncer    MediaSyncer
	validator MediaSyncValidator
	logger    *logger.Logger
}

// NewMediaSyncHandler 创建媒体同步处理器实例
func NewMediaSyncHandler(
	syncer MediaSyncer,
	validator MediaSyncValidator,
	logger *logger.Logger,
) MediaSyncHandler {
	return &mediaSyncHandler{
		syncer:    syncer,
		validator: validator,
		logger:    logger,
	}
}

// @Summary 同步单个媒体
// @Description 同步指定的单个媒体信息
// @Tags media-sync
// @Accept json
// @Produce json
// @Param request body MediaSyncRequest true "同步请求参数"
// @Success 200 {object} MediaSyncResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/sync [post]
func (h *mediaSyncHandler) SyncMedia(c *gin.Context) {
	h.logger.Debug("Received sync media request")
	
	var request MediaSyncRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数格式错误",
			"details": err.Error(),
		})
		return
	}
	
	// 验证请求参数
	if err := h.validator.ValidateSyncRequest(request); err != nil {
		h.logger.Error("Invalid sync request parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数验证失败",
			"details": err.Error(),
		})
		return
	}
	
	// 调用媒体同步器
	response, err := h.syncer.SyncMedia(c.Request.Context(), &request)
	if err != nil {
		h.handleError(c, err, "同步媒体失败")
		return
	}
	
	h.logger.Info("Media sync request processed successfully", "task_id", response.TaskID)
	c.JSON(http.StatusOK, response)
}

// @Summary 批量同步媒体
// @Description 批量同步多个媒体信息
// @Tags media-sync
// @Accept json
// @Produce json
// @Param request body MediaSyncRequest true "批量同步请求参数"
// @Success 200 {object} MediaSyncResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/sync/batch [post]
func (h *mediaSyncHandler) BatchSyncMedias(c *gin.Context) {
	h.logger.Debug("Received batch sync media request")
	
	var request MediaSyncRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数格式错误",
			"details": err.Error(),
		})
		return
	}
	
	// 验证请求参数
	if err := h.validator.ValidateBatchSyncRequest(request); err != nil {
		h.logger.Error("Invalid batch sync request parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数验证失败",
			"details": err.Error(),
		})
		return
	}
	
	// 调用媒体同步器
	response, err := h.syncer.BatchSyncMedias(c.Request.Context(), &request)
	if err != nil {
		h.handleError(c, err, "批量同步媒体失败")
		return
	}
	
	h.logger.Info("Batch media sync request processed successfully", "task_id", response.TaskID)
	c.JSON(http.StatusOK, response)
}

// @Summary 获取同步任务信息
// @Description 根据任务ID获取同步任务的详细信息
// @Tags media-sync
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} MediaSyncTask
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/sync/task/{task_id} [get]
func (h *mediaSyncHandler) GetSyncTask(c *gin.Context) {
	taskID := c.Param("task_id")
	h.logger.Debug("Received get sync task request", "task_id", taskID)
	
	// 验证任务ID
	if err := h.validator.ValidateTaskID(taskID); err != nil {
		h.logger.Error("Invalid task ID", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
			"details": err.Error(),
		})
		return
	}
	
	// 获取任务信息
	task, err := h.syncer.GetSyncTask(c.Request.Context(), taskID)
	if err != nil {
		h.handleError(c, err, "获取同步任务失败")
		return
	}
	
	h.logger.Info("Sync task retrieved successfully", "task_id", taskID)
	c.JSON(http.StatusOK, task)
}

// @Summary 列出同步任务
// @Description 分页列出同步任务
// @Tags media-sync
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "任务状态"
// @Param source query string false "同步源"
// @Param media_type query string false "媒体类型"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} MediaSyncTaskListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/sync/tasks [get]
func (h *mediaSyncHandler) ListSyncTasks(c *gin.Context) {
	h.logger.Debug("Received list sync tasks request")
	
	var query MediaSyncTaskQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.logger.Error("Failed to bind query parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查询参数格式错误",
			"details": err.Error(),
		})
		return
	}
	
	// 验证查询参数
	if err := h.validator.ValidateTaskQuery(query); err != nil {
		h.logger.Error("Invalid task query parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查询参数验证失败",
			"details": err.Error(),
		})
		return
	}
	
	// 获取任务列表
	response, err := h.syncer.ListSyncTasks(c.Request.Context(), &query)
	if err != nil {
		h.handleError(c, err, "列出同步任务失败")
		return
	}
	
	h.logger.Info("Sync tasks listed successfully", "page", query.Page, "total", response.Total)
	c.JSON(http.StatusOK, response)
}

// @Summary 取消同步任务
// @Description 根据任务ID取消正在执行的同步任务
// @Tags media-sync
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/sync/task/{task_id}/cancel [post]
func (h *mediaSyncHandler) CancelSyncTask(c *gin.Context) {
	taskID := c.Param("task_id")
	h.logger.Debug("Received cancel sync task request", "task_id", taskID)
	
	// 验证任务ID
	if err := h.validator.ValidateTaskID(taskID); err != nil {
		h.logger.Error("Invalid task ID", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
			"details": err.Error(),
		})
		return
	}
	
	// 取消任务
	if err := h.syncer.CancelSyncTask(c.Request.Context(), taskID); err != nil {
		h.handleError(c, err, "取消同步任务失败")
		return
	}
	
	h.logger.Info("Sync task cancelled successfully", "task_id", taskID)
	c.JSON(http.StatusOK, gin.H{
		"message": "任务已成功取消",
		"task_id": taskID,
	})
}

// @Summary 获取同步统计信息
// @Description 获取媒体同步的统计信息
// @Tags media-sync
// @Accept json
// @Produce json
// @Success 200 {object} MediaSyncStats
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/sync/stats [get]
func (h *mediaSyncHandler) GetSyncStats(c *gin.Context) {
	h.logger.Debug("Received get sync stats request")
	
	// 获取统计信息
	stats, err := h.syncer.GetSyncStats(c.Request.Context())
	if err != nil {
		h.handleError(c, err, "获取同步统计信息失败")
		return
	}
	
	h.logger.Info("Sync statistics retrieved successfully")
	c.JSON(http.StatusOK, stats)
}

// @Summary 解决同步冲突
// @Description 根据冲突解决方案处理同步冲突
// @Tags media-sync
// @Accept json
// @Produce json
// @Param request body SyncConflictResolution true "冲突解决方案"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/media/sync/conflict/resolve [post]
func (h *mediaSyncHandler) ResolveConflict(c *gin.Context) {
	h.logger.Debug("Received resolve conflict request")
	
	var resolution SyncConflictResolution
	if err := c.ShouldBindJSON(&resolution); err != nil {
		h.logger.Error("Failed to bind request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数格式错误",
			"details": err.Error(),
		})
		return
	}
	
	// 验证解决方案
	if err := h.validator.ValidateConflictResolution(resolution); err != nil {
		h.logger.Error("Invalid conflict resolution", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "冲突解决方案验证失败",
			"details": err.Error(),
		})
		return
	}
	
	// 解决冲突
	if err := h.syncer.ResolveConflict(c.Request.Context(), &resolution); err != nil {
		h.handleError(c, err, "解决同步冲突失败")
		return
	}
	
	h.logger.Info("Conflict resolved successfully", "conflict_id", resolution.ConflictID)
	c.JSON(http.StatusOK, gin.H{
		"message":     "冲突已成功解决",
		"conflict_id": resolution.ConflictID,
	})
}

// 错误处理辅助方法
func (h *mediaSyncHandler) handleError(c *gin.Context, err error, defaultMessage string) {
	// 获取错误代码
	errCode := errors.GetErrorCode(err)
	
	// 日志记录
	h.logger.Error(defaultMessage, "error", err.Error(), "code", errCode)
	
	// 根据错误代码返回不同的HTTP状态码
	switch errCode {
	case errors.ErrCodeInvalidInput:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   defaultMessage,
			"details": err.Error(),
		})
	case errors.ErrCodeNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"error":   defaultMessage,
			"details": err.Error(),
		})
	case errors.ErrCodeInvalidState:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   defaultMessage,
			"details": err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   defaultMessage,
			"details": "内部服务器错误",
		})
	}
}

// 响应类型定义（为了Swagger文档）

// ErrorResponse 错误响应
// swagger:model ErrorResponse
type ErrorResponse struct {
	// 错误信息
	Error string `json:"error"`
	// 详细错误信息
	Details string `json:"details,omitempty"`
}

// SuccessResponse 成功响应
// swagger:model SuccessResponse
type SuccessResponse struct {
	// 成功消息
	Message string      `json:"message"`
	// 附加数据
	Data    interface{} `json:"data,omitempty"`
}
