package notification

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	notificationbiz "moviepilot-go/internal/business/services/notification"
	"moviepilot-go/pkg/logger"
)

// Handler 通知 API 处理器
type Handler struct {
	service         notificationbiz.Service
	templateService notificationbiz.TemplateService
	logger          *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(
	service notificationbiz.Service,
	templateService notificationbiz.TemplateService,
) *Handler {
	return &Handler{
		service:         service,
		templateService: templateService,
		logger:          logger.GetLogger(),
	}
}

// Send 发送通知
// @Summary 发送通知
// @Description 发送一条通知消息
// @Tags notification
// @Accept json
// @Produce json
// @Param notification body notificationbiz.Notification true "通知内容"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/notification/send [post]
func (h *Handler) Send(c *gin.Context) {
	var notification notificationbiz.Notification

	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Send(c.Request.Context(), &notification); err != nil {
		h.logger.Error("发送通知失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知发送成功"})
}

// SendToChannel 发送到指定渠道
// @Summary 发送到指定渠道
// @Description 发送通知到指定的渠道
// @Tags notification
// @Accept json
// @Produce json
// @Param channel path string true "渠道名称"
// @Param notification body notificationbiz.Notification true "通知内容"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/notification/send/{channel} [post]
func (h *Handler) SendToChannel(c *gin.Context) {
	channel := notificationbiz.Channel(c.Param("channel"))

	var notification notificationbiz.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SendToChannel(c.Request.Context(), channel, &notification); err != nil {
		h.logger.Error("发送到渠道失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知发送成功"})
}

// SendBatch 批量发送
// @Summary 批量发送通知
// @Description 批量发送多条通知
// @Tags notification
// @Accept json
// @Produce json
// @Param notifications body []notificationbiz.Notification true "通知列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/notification/send/batch [post]
func (h *Handler) SendBatch(c *gin.Context) {
	var notifications []*notificationbiz.Notification

	if err := c.ShouldBindJSON(&notifications); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SendBatch(c.Request.Context(), notifications); err != nil {
		h.logger.Error("批量发送失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "批量发送成功"})
}

// GetHistory 获取通知历史
// @Summary 获取通知历史
// @Description 获取通知发送历史记录
// @Tags notification
// @Produce json
// @Param limit query int false "数量限制" default(50)
// @Success 200 {array} notificationbiz.NotificationRecord
// @Failure 500 {object} map[string]interface{}
// @Router /api/notification/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		var parsed int
		if _, err := fmt.Sscanf(limitStr, "%d", &parsed); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.service.GetHistory(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("获取通知历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetStats 获取通知统计
// @Summary 获取通知统计
// @Description 获取通知发送统计信息
// @Tags notification
// @Produce json
// @Success 200 {object} notificationbiz.NotificationStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/notification/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("获取通知统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SendFromTemplate 从模板发送
// @Summary 从模板发送通知
// @Description 使用模板发送通知
// @Tags notification
// @Accept json
// @Produce json
// @Param request body TemplateRequest true "模板请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/notification/send/template [post]
func (h *Handler) SendFromTemplate(c *gin.Context) {
	var req TemplateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从模板构建通知
	notification, err := notificationbiz.BuildNotificationFromTemplate(
		h.templateService,
		req.TemplateName,
		req.Data,
		req.Type,
	)
	if err != nil {
		h.logger.Error("构建通知失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置渠道
	notification.Channels = req.Channels

	// 发送通知
	if err := h.service.Send(c.Request.Context(), notification); err != nil {
		h.logger.Error("发送通知失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知发送成功"})
}

// TestChannel 测试渠道
// @Summary 测试通知渠道
// @Description 测试指定通知渠道的连接
// @Tags notification
// @Produce json
// @Param channel path string true "渠道名称"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/notification/test/{channel} [post]
func (h *Handler) TestChannel(c *gin.Context) {
	channel := c.Param("channel")

	// 发送测试通知
	testNotification := &notificationbiz.Notification{
		Title:    "测试通知",
		Content:  "这是一条测试消息，用于验证通知渠道是否正常工作。",
		Type:     notificationbiz.TypeInfo,
		Priority: notificationbiz.PriorityNormal,
		Channels: []notificationbiz.Channel{notificationbiz.Channel(channel)},
	}

	if err := h.service.Send(c.Request.Context(), testNotification); err != nil {
		h.logger.Error("测试渠道失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "测试成功"})
}

// TemplateRequest 模板请求
type TemplateRequest struct {
	TemplateName string                           `json:"template_name" binding:"required"`
	Data         any                              `json:"data" binding:"required"`
	Type         notificationbiz.NotificationType `json:"type"`
	Channels     []notificationbiz.Channel        `json:"channels"`
}
