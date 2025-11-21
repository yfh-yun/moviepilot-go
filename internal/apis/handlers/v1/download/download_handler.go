// Package download 下载管理API处理器
package download

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/response"
	"moviepilot-go/pkg/validator"
)

// DownloadHandler 下载管理处理器
// 提供下载任务的管理、监控和控制功能
type DownloadHandler struct {
	downloadService service.DownloadService
	logger          *zap.Logger
}

// NewDownloadHandler 创建下载管理处理器
func NewDownloadHandler(downloadService service.DownloadService, logger *zap.Logger) *DownloadHandler {
	return &DownloadHandler{
		downloadService: downloadService,
		logger:          logger,
	}
}

// DownloadTaskResponse 下载任务响应结构体
type DownloadTaskResponse struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Progress   float64   `json:"progress"`
	FileSize   int64     `json:"file_size"`
	Downloaded int64     `json:"downloaded"`
	Speed      int64     `json:"speed"`
	ETA        string    `json:"eta"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ListDownloadsRequest 下载列表请求结构体
type ListDownloadsRequest struct {
	Status string `form:"status" validate:"omitempty,oneof=pending downloading completed failed paused"`
	Type   string `form:"type" validate:"omitempty,oneof=movie tv book music"`
	Page   int    `form:"page" validate:"omitempty,min=1"`
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
}

// ListDownloadsResponse 下载列表响应结构体
type ListDownloadsResponse struct {
	Tasks []DownloadTaskResponse `json:"tasks"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
}

// CreateDownloadRequest 创建下载任务请求结构体
type CreateDownloadRequest struct {
	Title    string `json:"title" binding:"required,min=1,max=200"`
	Type     string `json:"type" binding:"required,oneof=movie tv book music"`
	URL      string `json:"url" binding:"required,url"`
	SavePath string `json:"save_path" binding:"required"`
}

// ListDownloads 获取下载任务列表
// @Summary 获取下载任务列表
// @Description 获取下载任务列表，支持按状态、类型和分页查询
// @Tags download
// @Accept json
// @Produce json
// @Param request query ListDownloadsRequest false "查询参数"
// @Success 200 {object} response.SuccessResponse{data=ListDownloadsResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads [get]
func (h *DownloadHandler) ListDownloads(c *gin.Context) {
	var req ListDownloadsRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("获取下载列表请求参数绑定失败", zap.Error(err))
		response.InvalidParams(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.ValidateStruct(req); err != nil {
		h.logger.Warn("获取下载列表请求参数验证失败", zap.Error(err))
		response.ValidateError(c, err)
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	ctx := c.Request.Context()

	// 调用服务层获取下载任务列表
	tasks, total, err := h.downloadService.ListDownloads(ctx, service.ListDownloadsParams{
		Status: req.Status,
		Type:   req.Type,
		Page:   req.Page,
		Limit:  req.Limit,
	})

	if err != nil {
		h.logger.Error("获取下载任务列表失败", zap.Error(err))
		response.ServerError(c, "获取下载任务列表失败")
		return
	}

	// 转换为响应格式
	var responseTasks []DownloadTaskResponse
	for _, task := range tasks {
		responseTasks = append(responseTasks, DownloadTaskResponse{
			ID:         task.ID,
			Title:      task.Title,
			Type:       task.Type,
			Status:     task.Status,
			Progress:   task.Progress,
			FileSize:   task.FileSize,
			Downloaded: task.Downloaded,
			Speed:      task.Speed,
			ETA:        task.ETA,
			CreatedAt:  task.CreatedAt,
			UpdatedAt:  task.UpdatedAt,
		})
	}

	response.Success(c, ListDownloadsResponse{
		Tasks: responseTasks,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	})
}

// GetDownloadDetail 获取下载任务详情
// @Summary 获取下载任务详情
// @Description 根据任务ID获取下载任务详细信息
// @Tags download
// @Accept json
// @Produce json
// @Param id path string true "下载任务ID"
// @Success 200 {object} response.SuccessResponse{data=DownloadTaskResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads/{id} [get]
func (h *DownloadHandler) GetDownloadDetail(c *gin.Context) {
	taskID := c.Param("id")

	if taskID == "" {
		response.InvalidParams(c, "下载任务ID不能为空")
		return
	}

	ctx := c.Request.Context()

	task, err := h.downloadService.GetDownloadDetail(ctx, taskID)
	if err != nil {
		if err == service.ErrDownloadNotFound {
			response.NotFound(c, "下载任务不存在")
			return
		}
		h.logger.Error("获取下载任务详情失败", zap.Error(err), zap.String("task_id", taskID))
		response.ServerError(c, "获取下载任务详情失败")
		return
	}

	responseTask := DownloadTaskResponse{
		ID:         task.ID,
		Title:      task.Title,
		Type:       task.Type,
		Status:     task.Status,
		Progress:   task.Progress,
		FileSize:   task.FileSize,
		Downloaded: task.Downloaded,
		Speed:      task.Speed,
		ETA:        task.ETA,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
	}

	response.Success(c, responseTask)
}

// CreateDownload 创建下载任务
// @Summary 创建下载任务
// @Description 创建新的下载任务
// @Tags download
// @Accept json
// @Produce json
// @Param request body CreateDownloadRequest true "下载任务信息"
// @Success 201 {object} response.SuccessResponse{data=DownloadTaskResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads [post]
func (h *DownloadHandler) CreateDownload(c *gin.Context) {
	var req CreateDownloadRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("创建下载任务请求参数绑定失败", zap.Error(err))
		response.InvalidParams(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.ValidateStruct(req); err != nil {
		h.logger.Warn("创建下载任务请求参数验证失败", zap.Error(err))
		response.ValidateError(c, err)
		return
	}

	ctx := c.Request.Context()

	// 调用服务层创建下载任务
	task, err := h.downloadService.CreateDownload(ctx, service.CreateDownloadParams{
		Title:    req.Title,
		Type:     req.Type,
		URL:      req.URL,
		SavePath: req.SavePath,
	})

	if err != nil {
		h.logger.Error("创建下载任务失败", zap.Error(err))
		response.ServerError(c, "创建下载任务失败")
		return
	}

	responseTask := DownloadTaskResponse{
		ID:         task.ID,
		Title:      task.Title,
		Type:       task.Type,
		Status:     task.Status,
		Progress:   task.Progress,
		FileSize:   task.FileSize,
		Downloaded: task.Downloaded,
		Speed:      task.Speed,
		ETA:        task.ETA,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
	}

	h.logger.Info("下载任务创建成功", zap.String("task_id", task.ID), zap.String("title", task.Title))
	response.Success(c, responseTask)
}

// DeleteDownload 删除下载任务
// @Summary 删除下载任务
// @Description 根据任务ID删除下载任务
// @Tags download
// @Accept json
// @Produce json
// @Param id path string true "下载任务ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads/{id} [delete]
func (h *DownloadHandler) DeleteDownload(c *gin.Context) {
	taskID := c.Param("id")

	if taskID == "" {
		response.InvalidParams(c, "下载任务ID不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.downloadService.DeleteDownload(ctx, taskID)
	if err != nil {
		if err == service.ErrDownloadNotFound {
			response.NotFound(c, "下载任务不存在")
			return
		}
		h.logger.Error("删除下载任务失败", zap.Error(err), zap.String("task_id", taskID))
		response.ServerError(c, "删除下载任务失败")
		return
	}

	h.logger.Info("下载任务删除成功", zap.String("task_id", taskID))
	response.Success(c, gin.H{
		"message": "下载任务删除成功",
		"task_id": taskID,
	})
}

// PauseDownload 暂停下载任务
// @Summary 暂停下载任务
// @Description 暂停指定的下载任务
// @Tags download
// @Accept json
// @Produce json
// @Param id path string true "下载任务ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads/{id}/pause [post]
func (h *DownloadHandler) PauseDownload(c *gin.Context) {
	taskID := c.Param("id")

	if taskID == "" {
		response.InvalidParams(c, "下载任务ID不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.downloadService.PauseDownload(ctx, taskID)
	if err != nil {
		if err == service.ErrDownloadNotFound {
			response.NotFound(c, "下载任务不存在")
			return
		}
		h.logger.Error("暂停下载任务失败", zap.Error(err), zap.String("task_id", taskID))
		response.ServerError(c, "暂停下载任务失败")
		return
	}

	h.logger.Info("下载任务暂停成功", zap.String("task_id", taskID))
	response.Success(c, gin.H{
		"message": "下载任务暂停成功",
		"task_id": taskID,
	})
}

// ResumeDownload 恢复下载任务
// @Summary 恢复下载任务
// @Description 恢复暂停的下载任务
// @Tags download
// @Accept json
// @Produce json
// @Param id path string true "下载任务ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads/{id}/resume [post]
func (h *DownloadHandler) ResumeDownload(c *gin.Context) {
	taskID := c.Param("id")

	if taskID == "" {
		response.InvalidParams(c, "下载任务ID不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.downloadService.ResumeDownload(ctx, taskID)
	if err != nil {
		if err == service.ErrDownloadNotFound {
			response.NotFound(c, "下载任务不存在")
			return
		}
		h.logger.Error("恢复下载任务失败", zap.Error(err), zap.String("task_id", taskID))
		response.ServerError(c, "恢复下载任务失败")
		return
	}

	h.logger.Info("下载任务恢复成功", zap.String("task_id", taskID))
	response.Success(c, gin.H{
		"message": "下载任务恢复成功",
		"task_id": taskID,
	})
}

// GetDownloadStats 获取下载统计信息
// @Summary 获取下载统计信息
// @Description 获取下载相关的统计信息
// @Tags download
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=service.DownloadStats}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads/stats [get]
func (h *DownloadHandler) GetDownloadStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.downloadService.GetDownloadStats(ctx)
	if err != nil {
		h.logger.Error("获取下载统计信息失败", zap.Error(err))
		response.ServerError(c, "获取下载统计信息失败")
		return
	}

	response.Success(c, stats)
}

// GetDownloadSpeed 获取下载速度
// @Summary 获取下载速度
// @Description 获取当前下载任务的实时速度信息
// @Tags download
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=map[string]int64}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads/speed [get]
func (h *DownloadHandler) GetDownloadSpeed(c *gin.Context) {
	ctx := c.Request.Context()

	speed, err := h.downloadService.GetDownloadSpeed(ctx)
	if err != nil {
		h.logger.Error("获取下载速度失败", zap.Error(err))
		response.ServerError(c, "获取下载速度失败")
		return
	}

	response.Success(c, speed)
}

// ClearCompletedDownloads 清空已完成下载
// @Summary 清空已完成下载
// @Description 清空所有已完成状态的下载任务
// @Tags download
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads/clear-completed [delete]
func (h *DownloadHandler) ClearCompletedDownloads(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.downloadService.ClearCompletedDownloads(ctx)
	if err != nil {
		h.logger.Error("清空已完成下载失败", zap.Error(err))
		response.ServerError(c, "清空已完成下载失败")
		return
	}

	h.logger.Info("已完成下载任务清空成功")
	response.Success(c, gin.H{
		"message": "已完成下载任务清空成功",
	})
}

// BatchDeleteDownloads 批量删除下载任务
// @Summary 批量删除下载任务
// @Description 批量删除指定的下载任务
// @Tags download
// @Accept json
// @Produce json
// @Param task_ids body []string true "下载任务ID列表"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/downloads/batch-delete [delete]
func (h *DownloadHandler) BatchDeleteDownloads(c *gin.Context) {
	var taskIDs []string

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&taskIDs); err != nil {
		h.logger.Warn("批量删除下载任务请求参数绑定失败", zap.Error(err))
		response.InvalidParams(c, "请求参数格式错误")
		return
	}

	if len(taskIDs) == 0 {
		response.InvalidParams(c, "任务ID列表不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.downloadService.BatchDeleteDownloads(ctx, taskIDs)
	if err != nil {
		h.logger.Error("批量删除下载任务失败", zap.Error(err))
		response.ServerError(c, "批量删除下载任务失败")
		return
	}

	h.logger.Info("批量删除下载任务成功", zap.Int("count", len(taskIDs)))
	response.Success(c, gin.H{
		"message": "批量删除下载任务成功",
		"count":   len(taskIDs),
	})
}
