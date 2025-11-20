// Package notification 通知管理API处理器
package notification

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/business/services"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"
)

// NotificationHandler 通知管理处理器
// 提供通知发送、通知列表、通知状态管理等功能
type NotificationHandler struct {
	notificationService service.NotificationService
	logger              *zap.Logger
}

// NewNotificationHandler 创建通知管理处理器
func NewNotificationHandler(notificationService service.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// NotificationItemResponse 通知项响应结构体
type NotificationItemResponse struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	Level     string                 `json:"level"`
	Read      bool                   `json:"read"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ListNotificationsRequest 通知列表请求结构体
type ListNotificationsRequest struct {
	Type  string `form:"type" validate:"omitempty,oneof=system media download plugin user"`
	Level string `form:"level" validate:"omitempty,oneof=info warning error success"`
	Read  *bool  `form:"read"`
	Page  int    `form:"page" validate:"omitempty,min=1"`
	Limit int    `form:"limit" validate:"omitempty,min=1,max=100"`
}

// ListNotificationsResponse 通知列表响应结构体
type ListNotificationsResponse struct {
	Notifications []NotificationItemResponse `json:"notifications"`
	Total         int64                      `json:"total"`
	UnreadCount   int                        `json:"unread_count"`
	Page          int                        `json:"page"`
	Limit         int                        `json:"limit"`
}

// SendNotificationRequest 发送通知请求结构体
type SendNotificationRequest struct {
	Title    string                 `json:"title" binding:"required,min=1,max=200"`
	Message  string                 `json:"message" binding:"required,min=1,max=1000"`
	Type     string                 `json:"type" binding:"required,oneof=system media download plugin user"`
	Level    string                 `json:"level" binding:"required,oneof=info warning error success"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ListNotifications 获取通知列表
// @Summary 获取通知列表
// @Description 获取通知列表，支持按类型、级别和已读状态筛选
// @Tags notification
// @Accept json
// @Produce json
// @Param request query ListNotificationsRequest false "查询参数"
// @Success 200 {object} response.SuccessResponse{data=ListNotificationsResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	var req ListNotificationsRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("获取通知列表请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("获取通知列表请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
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

	// 调用服务层获取通知列表
	notifications, total, unreadCount, err := h.notificationService.ListNotifications(ctx, service.ListNotificationsParams{
		Type:  req.Type,
		Level: req.Level,
		Read:  req.Read,
		Page:  req.Page,
		Limit: req.Limit,
	})

	if err != nil {
		h.logger.Error("获取通知列表失败", zap.Error(err))
		response.InternalServerError(c, "获取通知列表失败")
		return
	}

	// 转换为响应格式
	var responseNotifications []NotificationItemResponse
	for _, notification := range notifications {
		responseNotifications = append(responseNotifications, NotificationItemResponse{
			ID:        notification.ID,
			Title:     notification.Title,
			Message:   notification.Message,
			Type:      notification.Type,
			Level:     notification.Level,
			Read:      notification.Read,
			CreatedAt: notification.CreatedAt,
			Metadata:  notification.Metadata,
		})
	}

	response.Success(c, ListNotificationsResponse{
		Notifications: responseNotifications,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          req.Page,
		Limit:         req.Limit,
	})
}

// GetNotificationDetail 获取通知详情
// @Summary 获取通知详情
// @Description 根据通知ID获取通知详细信息
// @Tags notification
// @Accept json
// @Produce json
// @Param id path string true "通知ID"
// @Success 200 {object} response.SuccessResponse{data=NotificationItemResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/{id} [get]
func (h *NotificationHandler) GetNotificationDetail(c *gin.Context) {
	notificationID := c.Param("id")

	if notificationID == "" {
		response.BadRequest(c, "通知ID不能为空")
		return
	}

	ctx := c.Request.Context()

	notification, err := h.notificationService.GetNotificationDetail(ctx, notificationID)
	if err != nil {
		if err == service.ErrNotificationNotFound {
			response.NotFound(c, "通知不存在")
			return
		}
		h.logger.Error("获取通知详情失败", zap.Error(err), zap.String("notification_id", notificationID))
		response.InternalServerError(c, "获取通知详情失败")
		return
	}

	responseNotification := NotificationItemResponse{
		ID:        notification.ID,
		Title:     notification.Title,
		Message:   notification.Message,
		Type:      notification.Type,
		Level:     notification.Level,
		Read:      notification.Read,
		CreatedAt: notification.CreatedAt,
		Metadata:  notification.Metadata,
	}

	response.Success(c, responseNotification)
}

// MarkAsRead 标记通知为已读
// @Summary 标记通知为已读
// @Description 标记单个通知为已读状态
// @Tags notification
// @Accept json
// @Produce json
// @Param id path string true "通知ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	notificationID := c.Param("id")

	if notificationID == "" {
		response.BadRequest(c, "通知ID不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.notificationService.MarkAsRead(ctx, notificationID)
	if err != nil {
		if err == service.ErrNotificationNotFound {
			response.NotFound(c, "通知不存在")
			return
		}
		h.logger.Error("标记通知为已读失败", zap.Error(err), zap.String("notification_id", notificationID))
		response.InternalServerError(c, "标记通知为已读失败")
		return
	}

	logger.Info("通知标记为已读成功", zap.String("notification_id", notificationID))
	response.Success(c, gin.H{
		"message":         "通知标记为已读成功",
		"notification_id": notificationID,
	})
}

// MarkAllAsRead 标记所有通知为已读
// @Summary 标记所有通知为已读
// @Description 标记所有未读通知为已读状态
// @Tags notification
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.notificationService.MarkAllAsRead(ctx)
	if err != nil {
		h.logger.Error("标记所有通知为已读失败", zap.Error(err))
		response.InternalServerError(c, "标记所有通知为已读失败")
		return
	}

	logger.Info("所有通知标记为已读成功")
	response.Success(c, gin.H{
		"message": "所有通知标记为已读成功",
	})
}

// SendNotification 发送通知
// @Summary 发送通知
// @Description 发送新的系统通知
// @Tags notification
// @Accept json
// @Produce json
// @Param request body SendNotificationRequest true "通知信息"
// @Success 201 {object} response.SuccessResponse{data=NotificationItemResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications [post]
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req SendNotificationRequest

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("发送通知请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("发送通知请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}

	ctx := c.Request.Context()

	// 调用服务层发送通知
	notification, err := h.notificationService.SendNotification(ctx, service.SendNotificationParams{
		Title:    req.Title,
		Message:  req.Message,
		Type:     req.Type,
		Level:    req.Level,
		Metadata: req.Metadata,
	})

	if err != nil {
		h.logger.Error("发送通知失败", zap.Error(err))
		response.InternalServerError(c, "发送通知失败")
		return
	}

	responseNotification := NotificationItemResponse{
		ID:        notification.ID,
		Title:     notification.Title,
		Message:   notification.Message,
		Type:      notification.Type,
		Level:     notification.Level,
		Read:      notification.Read,
		CreatedAt: notification.CreatedAt,
		Metadata:  notification.Metadata,
	}

	logger.Info("通知发送成功", zap.String("notification_id", notification.ID), zap.String("title", notification.Title))
	response.Created(c, responseNotification)
}

// DeleteNotification 删除通知
// @Summary 删除通知
// @Description 删除指定的通知
// @Tags notification
// @Accept json
// @Produce json
// @Param id path string true "通知ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	notificationID := c.Param("id")

	if notificationID == "" {
		response.BadRequest(c, "通知ID不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.notificationService.DeleteNotification(ctx, notificationID)
	if err != nil {
		if err == service.ErrNotificationNotFound {
			response.NotFound(c, "通知不存在")
			return
		}
		h.logger.Error("删除通知失败", zap.Error(err), zap.String("notification_id", notificationID))
		response.InternalServerError(c, "删除通知失败")
		return
	}

	logger.Info("通知删除成功", zap.String("notification_id", notificationID))
	response.Success(c, gin.H{
		"message":         "通知删除成功",
		"notification_id": notificationID,
	})
}

// ClearAllNotifications 清空所有通知
// @Summary 清空所有通知
// @Description 清空所有通知记录
// @Tags notification
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications [delete]
func (h *NotificationHandler) ClearAllNotifications(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.notificationService.ClearAllNotifications(ctx)
	if err != nil {
		h.logger.Error("清空所有通知失败", zap.Error(err))
		response.InternalServerError(c, "清空所有通知失败")
		return
	}

	logger.Info("所有通知清空成功")
	response.Success(c, gin.H{
		"message": "所有通知清空成功",
	})
}

// GetNotificationStats 获取通知统计信息
// @Summary 获取通知统计信息
// @Description 获取通知的统计信息，包括总数、未读数量等
// @Tags notification
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=service.NotificationStats}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/stats [get]
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.notificationService.GetNotificationStats(ctx)
	if err != nil {
		h.logger.Error("获取通知统计信息失败", zap.Error(err))
		response.InternalServerError(c, "获取通知统计信息失败")
		return
	}

	response.Success(c, stats)
}

// BatchDeleteNotifications 批量删除通知
// @Summary 批量删除通知
// @Description 批量删除指定的通知
// @Tags notification
// @Accept json
// @Produce json
// @Param notification_ids body []string true "通知ID列表"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/batch-delete [delete]
func (h *NotificationHandler) BatchDeleteNotifications(c *gin.Context) {
	var notificationIDs []string

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&notificationIDs); err != nil {
		h.logger.Warn("批量删除通知请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	if len(notificationIDs) == 0 {
		response.BadRequest(c, "通知ID列表不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.notificationService.BatchDeleteNotifications(ctx, notificationIDs)
	if err != nil {
		h.logger.Error("批量删除通知失败", zap.Error(err))
		response.InternalServerError(c, "批量删除通知失败")
		return
	}

	logger.Info("批量删除通知成功", zap.Int("count", len(notificationIDs)))
	response.Success(c, gin.H{
		"message": "批量删除通知成功",
		"count":   len(notificationIDs),
	})
}

// GetNotificationPreferences 获取通知偏好设置
// @Summary 获取通知偏好设置
// @Description 获取用户的通知偏好设置
// @Tags notification
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=service.NotificationPreferences}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/preferences [get]
func (h *NotificationHandler) GetNotificationPreferences(c *gin.Context) {
	ctx := c.Request.Context()

	preferences, err := h.notificationService.GetNotificationPreferences(ctx)
	if err != nil {
		h.logger.Error("获取通知偏好设置失败", zap.Error(err))
		response.InternalServerError(c, "获取通知偏好设置失败")
		return
	}

	response.Success(c, preferences)
}

// UpdateNotificationPreferences 更新通知偏好设置
// @Summary 更新通知偏好设置
// @Description 更新用户的通知偏好设置
// @Tags notification
// @Accept json
// @Produce json
// @Param preferences body service.NotificationPreferences true "偏好设置"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/notifications/preferences [put]
func (h *NotificationHandler) UpdateNotificationPreferences(c *gin.Context) {
	var preferences service.NotificationPreferences

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&preferences); err != nil {
		h.logger.Warn("更新通知偏好设置请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(preferences); err != nil {
		h.logger.Warn("更新通知偏好设置请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}

	ctx := c.Request.Context()

	err := h.notificationService.UpdateNotificationPreferences(ctx, preferences)
	if err != nil {
		h.logger.Error("更新通知偏好设置失败", zap.Error(err))
		response.InternalServerError(c, "更新通知偏好设置失败")
		return
	}

	logger.Info("通知偏好设置更新成功")
	response.Success(c, gin.H{
		"message": "通知偏好设置更新成功",
	})
}
