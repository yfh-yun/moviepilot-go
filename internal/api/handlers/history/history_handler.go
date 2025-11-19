// Package history History API处理器模块
package history

import (
	"net/http"
	"strconv"

	"github.com/yfh-yun/moviepilot-go/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Handler History API处理器
type Handler struct {
	service service.HistoryService
	logger  *zap.Logger
}

// NewHandler 创建新的History处理器
func NewHandler(service service.HistoryService, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// parsePagination 解析并验证分页参数
func (h *Handler) parsePagination(c *gin.Context) (int, int, error) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, err := strconv.Atoi(pageParam)
	if err != nil {
		return 0, 0, err
	}
	if page <= 0 {
		page = 1
	}

	count, err := strconv.Atoi(countParam)
	if err != nil {
		return 0, 0, err
	}
	if count <= 0 {
		count = 30
	}
	if count > 1000 { // 防止过大的查询
		count = 1000
	}

	return page, count, nil
}

// sendSuccessResponse 发送成功响应
func (h *Handler) sendSuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// sendErrorResponse 发送错误响应
func (h *Handler) sendErrorResponse(c *gin.Context, message string, err error) {
	h.logger.Error(message, zap.Error(err))
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: message,
	})
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	historyGroup := router.Group("/history")
	{
		historyGroup.GET("/download", h.GetDownloadHistory)
		historyGroup.GET("/transfer", h.GetTransferHistory)
		historyGroup.GET("/subscribe", h.GetSubscribeHistory)
		historyGroup.GET("/system", h.GetSystemHistory)
	}
}

// GetDownloadHistory 获取下载历史
// @Summary 获取下载历史
// @Description 获取下载历史记录
// @Tags 历史记录
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Param status query string false "状态过滤 (success,failed,pending)"
// @Param media_type query string false "媒体类型过滤"
// @Success 200 {object} Response{data=[]DownloadHistory}
// @Router /history/download [get]
func (h *Handler) GetDownloadHistory(c *gin.Context) {
	page, count, err := h.parsePagination(c)
	if err != nil {
		h.sendErrorResponse(c, "分页参数格式错误", err)
		return
	}

	params := service.DownloadHistoryParams{
		Page:      page,
		Count:     count,
		Status:    c.Query("status"),
		MediaType: c.Query("media_type"),
	}

	histories, err := h.service.GetDownloadHistory(params)
	if err != nil {
		h.sendErrorResponse(c, "获取下载历史失败", err)
		return
	}

	h.sendSuccessResponse(c, histories)
}

// GetTransferHistory 获取转移历史
// @Summary 获取转移历史
// @Description 获取文件转移历史记录
// @Tags 历史记录
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Param status query string false "状态过滤 (success,failed,pending)"
// @Param media_type query string false "媒体类型过滤"
// @Success 200 {object} Response{data=[]TransferHistory}
// @Router /history/transfer [get]
func (h *Handler) GetTransferHistory(c *gin.Context) {
	page, count, err := h.parsePagination(c)
	if err != nil {
		h.sendErrorResponse(c, "分页参数格式错误", err)
		return
	}

	params := service.TransferHistoryParams{
		Page:      page,
		Count:     count,
		Status:    c.Query("status"),
		MediaType: c.Query("media_type"),
	}

	histories, err := h.service.GetTransferHistory(params)
	if err != nil {
		h.sendErrorResponse(c, "获取转移历史失败", err)
		return
	}

	h.sendSuccessResponse(c, histories)
}

// GetSubscribeHistory 获取订阅历史
// @Summary 获取订阅历史
// @Description 获取订阅历史记录
// @Tags 历史记录
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Param status query string false "状态过滤 (success,failed,pending)"
// @Param media_type query string false "媒体类型过滤"
// @Success 200 {object} Response{data=[]SubscribeHistory}
// @Router /history/subscribe [get]
func (h *Handler) GetSubscribeHistory(c *gin.Context) {
	page, count, err := h.parsePagination(c)
	if err != nil {
		h.sendErrorResponse(c, "分页参数格式错误", err)
		return
	}

	params := service.SubscribeHistoryParams{
		Page:      page,
		Count:     count,
		Status:    c.Query("status"),
		MediaType: c.Query("media_type"),
	}

	histories, err := h.service.GetSubscribeHistory(params)
	if err != nil {
		h.sendErrorResponse(c, "获取订阅历史失败", err)
		return
	}

	h.sendSuccessResponse(c, histories)
}

// GetSystemHistory 获取系统历史
// @Summary 获取系统历史
// @Description 获取系统操作历史记录
// @Tags 历史记录
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Param type query string false "类型过滤 (download,transfer,subscribe,plugin)"
// @Param level query string false "级别过滤 (info,warning,error)"
// @Success 200 {object} Response{data=[]SystemHistory}
// @Router /history/system [get]
func (h *Handler) GetSystemHistory(c *gin.Context) {
	page, count, err := h.parsePagination(c)
	if err != nil {
		h.sendErrorResponse(c, "分页参数格式错误", err)
		return
	}

	params := service.SystemHistoryParams{
		Page:  page,
		Count: count,
		Type:  c.Query("type"),
		Level: c.Query("level"),
	}

	histories, err := h.service.GetSystemHistory(params)
	if err != nil {
		h.sendErrorResponse(c, "获取系统历史失败", err)
		return
	}

	h.sendSuccessResponse(c, histories)
}
