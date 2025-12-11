package history

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	historybiz "moviepilot-go/internal/business/services/history"
	"moviepilot-go/pkg/logger"
)

// Handler 历史记录 API 处理器
type Handler struct {
	historyService historybiz.Service
	logger         *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(historyService historybiz.Service) *Handler {
	return &Handler{
		historyService: historyService,
		logger:         logger.GetLogger(),
	}
}

// GetDownloadHistory 获取下载历史
// @Summary 获取下载历史
// @Description 获取用户的下载历史记录
// @Tags history
// @Produce json
// @Param user_id query string true "用户ID"
// @Param limit query int false "数量限制" default(50)
// @Success 200 {array} database.DownloadHistory
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/download [get]
func (h *Handler) GetDownloadHistory(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	history, err := h.historyService.GetDownloadHistory(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("获取下载历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetOperationHistory 获取操作历史
// @Summary 获取操作历史
// @Description 获取用户的操作历史记录
// @Tags history
// @Produce json
// @Param user_id query string true "用户ID"
// @Param limit query int false "数量限制" default(50)
// @Success 200 {array} historybiz.OperationRecord
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/operation [get]
func (h *Handler) GetOperationHistory(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	records, err := h.historyService.GetOperationHistory(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("获取操作历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

// RecordOperation 记录操作
// @Summary 记录操作
// @Description 记录用户操作
// @Tags history
// @Accept json
// @Produce json
// @Param record body historybiz.OperationRecord true "操作记录"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/operation [post]
func (h *Handler) RecordOperation(c *gin.Context) {
	var record historybiz.OperationRecord

	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.historyService.RecordOperation(c.Request.Context(), &record); err != nil {
		h.logger.Error("记录操作失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "记录成功"})
}

// ClearHistory 清空历史
// @Summary 清空历史
// @Description 清空用户的历史记录
// @Tags history
// @Produce json
// @Param user_id query string true "用户ID"
// @Param type query string false "历史类型" Enums(download, operation, all) default(all)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/clear [post]
func (h *Handler) ClearHistory(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	historyType := c.DefaultQuery("type", "all")

	if err := h.historyService.ClearHistory(c.Request.Context(), userID, historyType); err != nil {
		h.logger.Error("清空历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "清空成功"})
}

// GetHistoryStats 获取历史统计
// @Summary 获取历史统计
// @Description 获取用户的历史统计信息
// @Tags history
// @Produce json
// @Param user_id query string true "用户ID"
// @Success 200 {object} historybiz.HistoryStats
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/stats [get]
func (h *Handler) GetHistoryStats(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	stats, err := h.historyService.GetHistoryStats(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("获取历史统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// DownloadHistory 获取下载历史记录
// @Summary 获取下载历史记录
// @Description 分页查询下载历史记录
// @Tags history
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {array} database.DownloadHistory
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/download [get]
func (h *Handler) DownloadHistory(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	countStr := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageStr)
	count, _ := strconv.Atoi(countStr)

	histories, err := h.historyService.GetDownloadHistoryByPage(c.Request.Context(), page, count)
	if err != nil {
		h.logger.Error("获取下载历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, histories)
}

// DeleteDownloadHistory 删除下载历史记录
// @Summary 删除下载历史记录
// @Description 删除指定的下载历史记录
// @Tags history
// @Produce json
// @Param id query int true "历史记录ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/download [delete]
func (h *Handler) DeleteDownloadHistory(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID不能为空"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.historyService.DeleteDownloadHistory(c.Request.Context(), id); err != nil {
		h.logger.Error("删除下载历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TransferHistory 获取整理记录
// @Summary 获取整理记录
// @Description 分页查询整理记录，支持标题和状态过滤
// @Tags history
// @Produce json
// @Param title query string false "标题"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Param status query bool false "状态（true成功，false失败）"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/transfer [get]
func (h *Handler) TransferHistory(c *gin.Context) {
	title := c.Query("title")
	pageStr := c.DefaultQuery("page", "1")
	countStr := c.DefaultQuery("count", "30")
	statusStr := c.Query("status")

	page, _ := strconv.Atoi(pageStr)
	count, _ := strconv.Atoi(countStr)

	var status *bool
	if statusStr != "" {
		statusBool, err := strconv.ParseBool(statusStr)
		if err == nil {
			status = &statusBool
		}
	}

	histories, total, err := h.historyService.GetTransferHistory(c.Request.Context(), title, page, count, status)
	if err != nil {
		h.logger.Error("获取整理记录失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]any{
			"list":  histories,
			"total": total,
		},
	})
}

// DeleteTransferHistory 删除整理记录
// @Summary 删除整理记录
// @Description 删除指定的整理记录，支持同时删除源文件和目标文件
// @Tags history
// @Produce json
// @Param id query int true "历史记录ID"
// @Param deletesrc query bool false "是否删除源文件"
// @Param deletedest query bool false "是否删除目标文件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/transfer [delete]
func (h *Handler) DeleteTransferHistory(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID不能为空"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	deleteSrc, _ := strconv.ParseBool(c.DefaultQuery("deletesrc", "false"))
	deleteDest, _ := strconv.ParseBool(c.DefaultQuery("deletedest", "false"))

	if err := h.historyService.DeleteTransferHistory(c.Request.Context(), id, deleteSrc, deleteDest); err != nil {
		h.logger.Error("删除整理记录失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// EmptyTransferHistory 清空整理记录
// @Summary 清空整理记录
// @Description 清空所有整理记录
// @Tags history
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/history/empty/transfer [get]
func (h *Handler) EmptyTransferHistory(c *gin.Context) {
	if err := h.historyService.EmptyTransferHistory(c.Request.Context()); err != nil {
		h.logger.Error("清空整理记录失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
