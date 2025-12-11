package transfer

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/apis/middlewares"
	"moviepilot-go/internal/business/services/transfer"
	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/models/dto"
)

// Handler 转移处理器
type Handler struct {
	logger          *zap.Logger
	historyService  transfer.HistoryService
	transferService transfer.Service
}

// NewHandler 创建转移处理器
func NewHandler(historyService transfer.HistoryService, transferService transfer.Service, logger *zap.Logger) *Handler {
	return &Handler{
		logger:          logger,
		historyService:  historyService,
		transferService: transferService,
	}
}

// NewTransferHandler 创建转移处理器（路由使用，兼容旧代码）
func NewTransferHandler(historyService transfer.HistoryService, logger *zap.Logger) *Handler {
	return &Handler{
		logger:          logger,
		historyService:  historyService,
		transferService: nil, // 暂时使用nil，后续会完善
	}
}

// ManualTransfer 手动转移
// @Summary 手动转移文件
// @Description 手动触发文件转移任务
// @Tags transfer
// @Accept json
// @Produce json
// @Param request body dto.TransferManualRequest true "手动转移请求"
// @Success 200 {object} dto.TransferManualResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transfer/manual [post]
func (h *Handler) ManualTransfer(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	var req dto.TransferManualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid manual transfer request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Manual transfer requested",
		zap.String("source_path", req.SourcePath),
		zap.String("target_path", req.TargetPath),
		zap.String("media_id", req.MediaID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 基于源路径构造一个简单的 Media 信息（后续可与 TMDB/数据库联动增强）
	fileName := filepath.Base(req.SourcePath)
	media := database.Media{
		Title: fileName,
		Type:  "movie", // 先默认按电影处理，后续可根据实际类型扩展
	}

	// 构造转移任务，交由业务层执行并记录历史
	tasks := []transfer.Task{
		{
			Media:      media,
			SourcePath: req.SourcePath,
			TargetPath: req.TargetPath,
			Category:   "manual", // 标记为手动转移任务
		},
	}

	histories, err := h.historyService.ExecuteAndSave(c.Request.Context(), tasks)
	if err != nil {
		h.logger.Error("Manual transfer execute failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	taskID := "transfer_" + generateTaskID()
	if len(histories) > 0 {
		taskID = "history_" + strconv.FormatUint(uint64(histories[0].ID), 10)
	}

	response := dto.TransferManualResponse{
		TaskID:     taskID,
		SourcePath: req.SourcePath,
		TargetPath: req.TargetPath,
		MediaID:    req.MediaID,
		Status:     "created",
		Message:    "Transfer task created successfully",
	}

	h.logger.Info("Manual transfer task created",
		zap.String("task_id", response.TaskID),
		zap.String("source", response.SourcePath),
		zap.String("target", response.TargetPath),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, response)
}

// GetHistory 获取转移历史
// @Summary 获取转移历史
// @Description 获取文件转移的历史记录
// @Tags transfer
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param status query string false "状态过滤"
// @Success 200 {object} dto.TransferHistoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transfer/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status")

	h.logger.Info("Transfer history requested",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.String("status", status),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	resp, err := h.historyService.ListHistory(c.Request.Context(), page, size, status)
	if err != nil {
		h.logger.Error("Transfer history query failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Transfer history retrieved",
		zap.Int("page", resp.Page),
		zap.Int("total", resp.Total),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, resp)
}

// DeleteHistory 删除转移历史
// @Summary 删除转移历史
// @Description 删除指定的转移历史记录
// @Tags transfer
// @Accept json
// @Produce json
// @Param id path string true "历史记录ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transfer/history/{id} [delete]
func (h *Handler) DeleteHistory(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	historyID := c.Param("id")
	if historyID == "" {
		h.logger.Error("Missing history ID for deletion",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "History ID is required",
		})
		return
	}

	h.logger.Info("Transfer history deletion requested",
		zap.String("history_id", historyID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用业务服务删除历史记录
	idUint, err := strconv.ParseUint(historyID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid history ID"})
		return
	}

	if err := h.historyService.DeleteHistory(c.Request.Context(), uint(idUint)); err != nil {
		h.logger.Error("Delete transfer history failed",
			zap.Error(err),
			zap.String("history_id", historyID),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Transfer history deleted",
		zap.String("history_id", historyID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, gin.H{
		"history_id": historyID,
		"status":     "deleted",
		"message":    "Transfer history deleted successfully",
	})
}

// GetStatus 获取转移状态
// @Summary 获取转移状态
// @Description 获取当前正在进行的转移任务状态
// @Tags transfer
// @Accept json
// @Produce json
// @Param task_id query string false "任务ID"
// @Success 200 {object} dto.TransferStatusResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transfer/status [get]
func (h *Handler) GetStatus(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	taskID := c.Query("task_id")

	h.logger.Info("Transfer status requested",
		zap.String("task_id", taskID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取转移状态逻辑
	// 1. 查询指定任务状态
	// 2. 如果未指定task_id，返回所有活动任务
	// 3. 包含进度信息、速度、剩余时间等

	response := dto.TransferStatusResponse{
		TaskID:    taskID,
		Status:    "not_found",
		Progress:  0,
		Speed:     0,
		Remaining: 0,
		Message:   "No transfer task found",
	}

	h.logger.Info("Transfer status retrieved",
		zap.String("task_id", taskID),
		zap.String("status", response.Status),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, response)
}

// CancelTransfer 取消转移
// @Summary 取消转移任务
// @Description 取消正在进行的转移任务
// @Tags transfer
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transfer/{task_id}/cancel [post]
func (h *Handler) CancelTransfer(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	taskID := c.Param("task_id")
	if taskID == "" {
		h.logger.Error("Missing task ID for cancellation",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
		})
		return
	}

	h.logger.Info("Transfer cancellation requested",
		zap.String("task_id", taskID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现取消转移逻辑
	// 1. 查找活动任务
	// 2. 停止文件传输
	// 3. 清理临时文件
	// 4. 更新任务状态

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"status":  "cancelled",
		"message": "Transfer task cancelled successfully",
	})
}

// RetryTransfer 重试转移
// @Summary 重试转移任务
// @Description 重试失败的转移任务
// @Tags transfer
// @Accept json
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transfer/{task_id}/retry [post]
func (h *Handler) RetryTransfer(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	taskID := c.Param("task_id")
	if taskID == "" {
		h.logger.Error("Missing task ID for retry",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
		})
		return
	}

	h.logger.Info("Transfer retry requested",
		zap.String("task_id", taskID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现重试转移逻辑
	// 1. 查找失败任务
	// 2. 重置任务状态
	// 3. 重新启动转移流程
	// 4. 返回新任务ID

	newTaskID := "retry_" + generateTaskID()
	c.JSON(http.StatusOK, gin.H{
		"old_task_id": taskID,
		"new_task_id": newTaskID,
		"status":      "retrying",
		"message":     "Transfer retry started successfully",
	})
}

// QueryName 查询整理后的名称
// @Summary 查询整理后的名称
// @Description 查询整理后的名称
// @Tags transfer
// @Produce json
// @Param path query string true "文件路径"
// @Param filetype query string true "文件类型"
// @Success 200 {object} map[string]interface{}
// @Router /api/transfer/name [get]
func (h *Handler) QueryName(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	path := c.Query("path")
	filetype := c.Query("filetype")

	h.logger.Info("Querying renamed media name",
		zap.String("path", path),
		zap.String("filetype", filetype),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 检查transferService是否可用
	if h.transferService == nil {
		h.logger.Error("Transfer service not initialized",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Transfer service not initialized"})
		return
	}

	// 调用服务层查询整理后的名称
	newName, err := h.transferService.QueryName(path, filetype)
	if err != nil {
		h.logger.Error("Query renamed media name failed",
			zap.Error(err),
			zap.String("path", path),
			zap.String("filetype", filetype),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("Query renamed media name success",
		zap.String("path", path),
		zap.String("filetype", filetype),
		zap.String("new_name", newName),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]any{
			"name": newName,
		},
	})
}

// GetQueue 查询整理队列
// @Summary 查询整理队列
// @Description 查询整理队列
// @Tags transfer
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/transfer/queue [get]
func (h *Handler) GetQueue(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("Getting transfer queue",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 检查transferService是否可用
	if h.transferService == nil {
		h.logger.Error("Transfer service not initialized",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Transfer service not initialized"})
		return
	}

	// 调用服务层查询整理队列
	queue, err := h.transferService.GetQueue()
	if err != nil {
		h.logger.Error("Get transfer queue failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("Get transfer queue success",
		zap.Int("queue_length", len(queue)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, queue)
}

// RemoveQueue 从整理队列中删除任务
// @Summary 从整理队列中删除任务
// @Description 从整理队列中删除任务
// @Tags transfer
// @Accept json
// @Produce json
// @Param fileitem body map[string]interface{} true "文件项"
// @Success 200 {object} map[string]interface{}
// @Router /api/transfer/queue [delete]
func (h *Handler) RemoveQueue(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	var fileitem map[string]any
	if err := c.ShouldBindJSON(&fileitem); err != nil {
		h.logger.Error("Invalid fileitem format",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fileitem format"})
		return
	}

	// 获取path字段
	path, ok := fileitem["path"].(string)
	if !ok {
		h.logger.Error("Missing path in fileitem",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing path in fileitem"})
		return
	}

	h.logger.Info("Removing task from transfer queue",
		zap.String("path", path),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 检查transferService是否可用
	if h.transferService == nil {
		h.logger.Error("Transfer service not initialized",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Transfer service not initialized"})
		return
	}

	// 调用服务层从整理队列中删除任务
	err := h.transferService.RemoveFromQueue(path)
	if err != nil {
		h.logger.Error("Remove task from transfer queue failed",
			zap.Error(err),
			zap.String("path", path),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("Remove task from transfer queue success",
		zap.String("path", path),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Now 立即执行下载器文件整理
// @Summary 立即执行下载器文件整理
// @Description 立即执行下载器文件整理
// @Tags transfer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/transfer/now [get]
func (h *Handler) Now(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("Processing downloader files",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 检查transferService是否可用
	if h.transferService == nil {
		h.logger.Error("Transfer service not initialized",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Transfer service not initialized"})
		return
	}

	// 调用服务层立即执行下载器文件整理
	err := h.transferService.Process()
	if err != nil {
		h.logger.Error("Process downloader files failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("Process downloader files success",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 辅助函数
func generateTaskID() string {
	// TODO: 实现生成唯一任务ID的逻辑
	// 可以使用UUID或时间戳+随机数
	return "temp_task_id"
}
