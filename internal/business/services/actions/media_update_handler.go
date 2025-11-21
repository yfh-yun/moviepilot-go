package actions

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/response"
)

// MediaUpdateHandler 媒体更新处理器接口
type MediaUpdateHandler interface {
	// UpdateMedia 更新单个媒体
	UpdateMedia(c *gin.Context)
	// BatchUpdateMedias 批量更新媒体
	BatchUpdateMedias(c *gin.Context)
	// GetUpdateTask 获取更新任务状态
	GetUpdateTask(c *gin.Context)
	// ListUpdateTasks 列出更新任务
	ListUpdateTasks(c *gin.Context)
	// CancelUpdateTask 取消更新任务
	CancelUpdateTask(c *gin.Context)
	// GetUpdateStats 获取更新统计信息
	GetUpdateStats(c *gin.Context)
}

// mediaUpdateHandler 媒体更新处理器实现
type mediaUpdateHandler struct {
	mediaUpdater    MediaUpdater
	mediaValidator  MediaUpdateValidator
	logger          logger.Logger
}

// NewMediaUpdateHandler 创建媒体更新处理器实例
func NewMediaUpdateHandler(
	mediaUpdater MediaUpdater,
	mediaValidator MediaUpdateValidator,
	logger logger.Logger,
) MediaUpdateHandler {
	return &mediaUpdateHandler{
		mediaUpdater:   mediaUpdater,
		mediaValidator: mediaValidator,
		logger:         logger,
	}
}

// UpdateMedia 更新单个媒体
// @Summary 更新单个媒体
// @Description 更新指定ID的媒体信息
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param media_id path string true "媒体ID"
// @Param request body MediaUpdateRequest true "更新请求参数"
// @Success 200 {object} response.JSONResponse{data=MediaUpdateResult}
// @Failure 400 {object} response.JSONResponse{data=response.ErrorResponse}
// @Failure 500 {object} response.JSONResponse{data=response.ErrorResponse}
// @Router /api/v1/medias/{media_id}/update [post]
func (h *mediaUpdaterHandler) UpdateMedia(c *gin.Context) {
	ctx := c.Request.Context()
	mediaID := c.Param("media_id")

	// 验证媒体ID
	if mediaID == "" {
		h.logger.Warn("媒体ID不能为空")
		response.JSONError(c, http.StatusBadRequest, "媒体ID不能为空")
		return
	}

	// 绑定请求参数
	var req MediaUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("绑定请求参数失败", "error", err.Error())
		response.JSONError(c, http.StatusBadRequest, "请求参数格式错误: "+err.Error())
		return
	}

	// 设置媒体ID
	req.MediaID = mediaID

	// 验证请求参数
	if err := h.mediaValidator.ValidateUpdateRequest(c, &req); err != nil {
		h.logger.Warn("验证请求参数失败", "error", err.Error())
		response.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 执行更新
	result, err := h.mediaUpdater.UpdateMedia(ctx, &req)
	if err != nil {
		h.logger.Error("更新媒体失败", "media_id", mediaID, "error", err.Error())
		response.JSONError(c, http.StatusInternalServerError, "更新媒体失败: "+err.Error())
		return
	}

	h.logger.Info("更新媒体成功", "media_id", mediaID)
	response.JSONSuccess(c, http.StatusOK, result)
}

// BatchUpdateMedias 批量更新媒体
// @Summary 批量更新媒体
// @Description 根据条件批量更新媒体信息
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param request body MediaUpdateRequest true "批量更新请求参数"
// @Success 202 {object} response.JSONResponse{data=MediaUpdateResponse}
// @Failure 400 {object} response.JSONResponse{data=response.ErrorResponse}
// @Failure 500 {object} response.JSONResponse{data=response.ErrorResponse}
// @Router /api/v1/medias/batch/update [post]
func (h *mediaUpdaterHandler) BatchUpdateMedias(c *gin.Context) {
	ctx := c.Request.Context()

	// 绑定请求参数
	var req MediaUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("绑定请求参数失败", "error", err.Error())
		response.JSONError(c, http.StatusBadRequest, "请求参数格式错误: "+err.Error())
		return
	}

	// 验证请求参数
	if err := h.mediaValidator.ValidateBatchUpdateRequest(c, &req); err != nil {
		h.logger.Warn("验证请求参数失败", "error", err.Error())
		response.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 执行批量更新
	result, err := h.mediaUpdater.BatchUpdateMedias(ctx, &req)
	if err != nil {
		h.logger.Error("批量更新媒体失败", "error", err.Error())
		response.JSONError(c, http.StatusInternalServerError, "批量更新媒体失败: "+err.Error())
		return
	}

	h.logger.Info("批量更新媒体任务创建成功", "task_id", result.TaskID)
	response.JSONSuccess(c, http.StatusAccepted, result)
}

// GetUpdateTask 获取更新任务状态
// @Summary 获取更新任务状态
// @Description 获取指定任务ID的更新任务状态
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.JSONResponse{data=MediaUpdateTask}
// @Failure 400 {object} response.JSONResponse{data=response.ErrorResponse}
// @Failure 404 {object} response.JSONResponse{data=response.ErrorResponse}
// @Failure 500 {object} response.JSONResponse{data=response.ErrorResponse}
// @Router /api/v1/medias/update/tasks/{task_id} [get]
func (h *mediaUpdaterHandler) GetUpdateTask(c *gin.Context) {
	ctx := c.Request.Context()
	taskID := c.Param("task_id")

	// 验证任务ID
	if taskID == "" {
		h.logger.Warn("任务ID不能为空")
		response.JSONError(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// 获取任务
	task, err := h.mediaUpdater.GetUpdateTask(ctx, taskID)
	if err != nil {
		h.logger.Warn("获取任务失败", "task_id", taskID, "error", err.Error())
		response.JSONError(c, http.StatusNotFound, "任务不存在")
		return
	}

	h.logger.Info("获取任务成功", "task_id", taskID)
	response.JSONSuccess(c, http.StatusOK, task)
}

// ListUpdateTasks 列出更新任务
// @Summary 列出更新任务
// @Description 分页列出更新任务
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页数量，默认为20"
// @Param status query string false "任务状态"
// @Param media_type query string false "媒体类型"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} response.JSONResponse{data=MediaUpdateTaskListResponse}
// @Failure 400 {object} response.JSONResponse{data=response.ErrorResponse}
// @Failure 500 {object} response.JSONResponse{data=response.ErrorResponse}
// @Router /api/v1/medias/update/tasks [get]
func (h *mediaUpdaterHandler) ListUpdateTasks(c *gin.Context) {
	ctx := c.Request.Context()

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 创建查询参数
	query := &MediaUpdateTaskQuery{
		Page:       page,
		PageSize:   pageSize,
		Status:     c.Query("status"),
		MediaType:  c.Query("media_type"),
		StartTime:  c.Query("start_time"),
		EndTime:    c.Query("end_time"),
	}

	// 验证查询参数
	if err := h.mediaValidator.ValidateTaskQuery(c, query); err != nil {
		h.logger.Warn("验证查询参数失败", "error", err.Error())
		response.JSONError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取任务列表
	tasks, err := h.mediaUpdater.ListUpdateTasks(ctx, query)
	if err != nil {
		h.logger.Error("获取任务列表失败", "error", err.Error())
		response.JSONError(c, http.StatusInternalServerError, "获取任务列表失败: "+err.Error())
		return
	}

	h.logger.Info("获取任务列表成功", "total", tasks.Total, "page", tasks.Page)
	response.JSONSuccess(c, http.StatusOK, tasks)
}

// CancelUpdateTask 取消更新任务
// @Summary 取消更新任务
// @Description 取消指定任务ID的更新任务
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} response.JSONResponse{data=response.SuccessResponse}
// @Failure 400 {object} response.JSONResponse{data=response.ErrorResponse}
// @Failure 404 {object} response.JSONResponse{data=response.ErrorResponse}
// @Failure 500 {object} response.JSONResponse{data=response.ErrorResponse}
// @Router /api/v1/medias/update/tasks/{task_id}/cancel [post]
func (h *mediaUpdaterHandler) CancelUpdateTask(c *gin.Context) {
	ctx := c.Request.Context()
	taskID := c.Param("task_id")

	// 验证任务ID
	if taskID == "" {
		h.logger.Warn("任务ID不能为空")
		response.JSONError(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	// 取消任务
	err := h.mediaUpdater.CancelUpdateTask(ctx, taskID)
	if err != nil {
		h.logger.Warn("取消任务失败", "task_id", taskID, "error", err.Error())
		response.JSONError(c, http.StatusNotFound, "任务不存在或无法取消")
		return
	}

	h.logger.Info("取消任务成功", "task_id", taskID)
	response.JSONSuccess(c, http.StatusOK, response.SuccessResponse{Message: "任务已取消"})
}

// GetUpdateStats 获取更新统计信息
// @Summary 获取更新统计信息
// @Description 获取媒体更新的统计信息
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JSONResponse{data=MediaUpdateStats}
// @Failure 500 {object} response.JSONResponse{data=response.ErrorResponse}
// @Router /api/v1/medias/update/stats [get]
func (h *mediaUpdaterHandler) GetUpdateStats(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取统计信息
	stats, err := h.mediaUpdater.GetUpdateStats(ctx)
	if err != nil {
		h.logger.Error("获取统计信息失败", "error", err.Error())
		response.JSONError(c, http.StatusInternalServerError, "获取统计信息失败: "+err.Error())
		return
	}

	h.logger.Info("获取统计信息成功")
	response.JSONSuccess(c, http.StatusOK, stats)
}
