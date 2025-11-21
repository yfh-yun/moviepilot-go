package actions

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/errors"
)

var handlerLog = logger.NewLogger("clean_torrents_handler")

// CleanTorrentsHandler 种子清理处理器接口
type CleanTorrentsHandler interface {
	// CleanTorrents 清理种子
	CleanTorrents(c *gin.Context)
	
	// GetCleanTask 获取清理任务状态
	GetCleanTask(c *gin.Context)
	
	// ListCleanTasks 列出清理任务
	ListCleanTasks(c *gin.Context)
	
	// CancelCleanTask 取消清理任务
	CancelCleanTask(c *gin.Context)
	
	// GetCleanStats 获取清理统计信息
	GetCleanStats(c *gin.Context)
}

// CleanTorrentsHandlerImpl 种子清理处理器实现
type CleanTorrentsHandlerImpl struct {
	cleanTorrentsService CleanTorrentsService
	validator           *CleanTorrentsValidator
}

// NewCleanTorrentsHandler 创建种子清理处理器实例
func NewCleanTorrentsHandler(cleanTorrentsService CleanTorrentsService) CleanTorrentsHandler {
	return &CleanTorrentsHandlerImpl{
		cleanTorrentsService: cleanTorrentsService,
		validator:           NewCleanTorrentsValidator(),
	}
}

// CleanTorrents 清理种子
// @Summary 清理种子
// @Description 根据指定策略清理下载器中的种子
// @Tags actions
// @Accept json
// @Produce json
// @Param request body CleanTorrentRequest true "清理请求参数"
// @Success 200 {object} CleanTorrentResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/actions/clean-torrents [post]
func (h *CleanTorrentsHandlerImpl) CleanTorrents(c *gin.Context) {
	var req CleanTorrentRequest
	
	// 绑定请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		handlerLog.Error("Failed to bind request body", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	
	// 验证请求参数
	if err := h.validator.ValidateCleanRequest(req); err != nil {
		handlerLog.Error("Invalid clean request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 调用服务层清理种子
	resp, err := h.cleanTorrentsService.CleanTorrents(c.Request.Context(), req)
	if err != nil {
		handlerLog.Error("Failed to clean torrents", "error", err.Error())
		statusCode := getStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	
	handlerLog.Info("Clean torrents task started", "task_id", resp.TaskID)
	c.JSON(http.StatusOK, resp)
}

// GetCleanTask 获取清理任务状态
// @Summary 获取清理任务状态
// @Description 根据任务ID获取清理任务的详细状态
// @Tags actions
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} CleanTorrentTask
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/actions/clean-torrents/{task_id} [get]
func (h *CleanTorrentsHandlerImpl) GetCleanTask(c *gin.Context) {
	// 获取任务ID
	taskID := c.Param("task_id")
	
	// 验证任务ID
	if err := h.validator.ValidateTaskID(taskID); err != nil {
		handlerLog.Error("Invalid task ID", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 调用服务层获取任务状态
	task, err := h.cleanTorrentsService.GetCleanTask(c.Request.Context(), taskID)
	if err != nil {
		handlerLog.Error("Failed to get clean task", "task_id", taskID, "error", err.Error())
		statusCode := getStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	
	handlerLog.Info("Retrieved clean task status", "task_id", taskID, "status", task.Status)
	c.JSON(http.StatusOK, task)
}

// ListCleanTasks 列出清理任务
// @Summary 列出清理任务
// @Description 获取清理任务列表，支持分页
// @Tags actions
// @Accept json
// @Produce json
// @Param limit query int false "每页数量" default(10)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} CleanTorrentTask
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/actions/clean-torrents [get]
func (h *CleanTorrentsHandlerImpl) ListCleanTasks(c *gin.Context) {
	// 获取分页参数
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	// 验证分页参数
	if err := h.validator.ValidatePagination(limit, offset); err != nil {
		handlerLog.Error("Invalid pagination parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 调用服务层获取任务列表
	tasks, err := h.cleanTorrentsService.ListCleanTasks(c.Request.Context(), limit, offset)
	if err != nil {
		handlerLog.Error("Failed to list clean tasks", "error", err.Error())
		statusCode := getStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	
	handlerLog.Info("Retrieved clean tasks list", "count", len(tasks))
	c.JSON(http.StatusOK, tasks)
}

// CancelCleanTask 取消清理任务
// @Summary 取消清理任务
// @Description 取消正在进行的清理任务
// @Tags actions
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/actions/clean-torrents/{task_id}/cancel [post]
func (h *CleanTorrentsHandlerImpl) CancelCleanTask(c *gin.Context) {
	// 获取任务ID
	taskID := c.Param("task_id")
	
	// 验证任务ID
	if err := h.validator.ValidateTaskID(taskID); err != nil {
		handlerLog.Error("Invalid task ID", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 调用服务层取消任务
	err := h.cleanTorrentsService.CancelCleanTask(c.Request.Context(), taskID)
	if err != nil {
		handlerLog.Error("Failed to cancel clean task", "task_id", taskID, "error", err.Error())
		statusCode := getStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	
	handlerLog.Info("Clean task cancellation requested", "task_id", taskID)
	c.JSON(http.StatusOK, gin.H{"message": "Task cancellation requested"})
}

// GetCleanStats 获取清理统计信息
// @Summary 获取清理统计信息
// @Description 获取种子清理的统计数据
// @Tags actions
// @Accept json
// @Produce json
// @Success 200 {object} CleanTorrentStats
// @Failure 500 {object} map[string]string
// @Router /api/v1/actions/clean-torrents/stats [get]
func (h *CleanTorrentsHandlerImpl) GetCleanStats(c *gin.Context) {
	// 调用服务层获取统计信息
	stats, err := h.cleanTorrentsService.GetCleanStats(c.Request.Context())
	if err != nil {
		handlerLog.Error("Failed to get clean stats", "error", err.Error())
		statusCode := getStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	
	handlerLog.Info("Retrieved clean stats")
	c.JSON(http.StatusOK, stats)
}

// getStatusCode 根据错误类型获取HTTP状态码
func getStatusCode(err error) int {
	switch err.(type) {
	case *errors.ValidationError:
		return http.StatusBadRequest
	case *errors.NotFoundError:
		return http.StatusNotFound
	case *errors.InvalidOperationError:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
