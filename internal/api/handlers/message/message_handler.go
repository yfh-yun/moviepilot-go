// Package message Message API处理器模块
package message

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/internal/api/validator"
	
	"github.com/yfh-yun/moviepilot-go/internal/service"
)

// Handler Message API处理器
type Handler struct {
	service service.MessageService
	logger  *zap.Logger
}

// NewHandler 创建新的Message处理器
func NewHandler(service service.MessageService, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	messageGroup := router.Group("/message")
	{
		messageGroup.GET("/", h.GetMessageList)
		messageGroup.POST("/", h.SendMessage)
		messageGroup.PUT("/:id/read", h.MarkMessageRead)
		messageGroup.PUT("/read-all", h.MarkAllAsRead)
		messageGroup.DELETE("/:id", h.DeleteMessage)
		messageGroup.GET("/unread-count", h.GetUnreadCount)
	}
}

// MessageInfo 消息信息结构
type MessageInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

// SendMessageRequest 发送消息请求结构
type SendMessageRequest struct {
	Title     string   `json:"title" binding:"required"`
	Content   string   `json:"content" binding:"required"`
	MessageType string `json:"message_type" binding:"required"`
	UserIDs   []string `json:"user_ids" binding:"required"`
}

// GetMessageList 获取消息列表
// @Summary 获取消息列表
// @Description 获取用户消息列表
// @Tags 消息
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {object} response.APIResponse{data=[]MessageInfo}
// @Router /message [get]
func (h *Handler) GetMessageList(c *gin.Context) {
	// 获取用户ID（从JWT或中间件设置）
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权访问")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		response.InternalServerError(c, "用户ID格式错误")
		return
	}

	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "20")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	// 规范化分页参数
	page, count = validator.NormalizePage(page, count)

	// 这里需要从user_id字符串转换为uint，实际实现中可能需要调整
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	ctx := c.Request.Context()
	messages, total, err := h.service.GetMessages(ctx, uint(userIDUint), page, count)
	if err != nil {
		h.logger.Error("获取消息列表失败", zap.Error(err))
		response.InternalServerError(c, "获取消息列表失败")
		return
	}

	// 转换为响应格式
	var messageInfos []MessageInfo
	for _, msg := range messages {
		messageInfos = append(messageInfos, MessageInfo{
			ID:        strconv.FormatUint(uint64(msg.ID), 10),
			Title:     msg.Title,
			Content:   msg.Content,
			Type:      msg.Type,
			Read:      msg.IsRead,
			CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	response.SuccessWithPagination(c, messageInfos, page, count, total)
}

// SendMessage 发送消息
// @Summary 发送消息
// @Description 发送消息给用户
// @Tags 消息
// @Accept json
// @Produce json
// @Param request body SendMessageRequest true "发送消息请求"
// @Success 200 {object} response.APIResponse
// @Router /message [post]
func (h *Handler) SendMessage(c *gin.Context) {
	var req SendMessageRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("发送消息请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.NewValidator(h.logger).Validate(req); err != nil {
		h.logger.Warn("发送消息请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.NewValidator(h.logger).TranslateError(err))
		return
	}

	// 转换用户ID字符串为uint
	var userIDs []uint
	for _, idStr := range req.UserIDs {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "用户ID格式错误")
			return
		}
		userIDs = append(userIDs, uint(id))
	}

	ctx := c.Request.Context()
	err := h.service.SendMessage(ctx, req.Title, req.Content, req.MessageType, userIDs)
	if err != nil {
		h.logger.Error("发送消息失败", zap.Error(err))
		response.InternalServerError(c, "发送消息失败")
		return
	}

	response.SuccessWithMessage(c, "消息发送成功", nil)
}

// MarkMessageRead 标记消息已读
// @Summary 标记消息已读
// @Description 标记消息为已读状态
// @Tags 消息
// @Produce json
// @Param id path string true "消息ID"
// @Success 200 {object} response.APIResponse
// @Router /message/{id}/read [put]
func (h *Handler) MarkMessageRead(c *gin.Context) {
	// 获取用户ID（从JWT或中间件设置）
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权访问")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		response.InternalServerError(c, "用户ID格式错误")
		return
	}

	messageIDStr := c.Param("id")
	if messageIDStr == "" {
		response.BadRequest(c, "消息ID不能为空")
		return
	}

	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "消息ID格式错误")
		return
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	ctx := c.Request.Context()
	err = h.service.MarkAsRead(ctx, uint(messageID), uint(userIDUint))
	if err != nil {
		h.logger.Error("标记消息已读失败", zap.Error(err))
		response.InternalServerError(c, "标记消息已读失败")
		return
	}

	response.SuccessWithMessage(c, "消息已标记为已读", nil)
}

// MarkAllAsRead 标记所有消息已读
// @Summary 标记所有消息已读
// @Description 标记用户所有消息为已读状态
// @Tags 消息
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /message/read-all [put]
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	// 获取用户ID（从JWT或中间件设置）
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权访问")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		response.InternalServerError(c, "用户ID格式错误")
		return
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	ctx := c.Request.Context()
	err = h.service.MarkAllAsRead(ctx, uint(userIDUint))
	if err != nil {
		h.logger.Error("标记所有消息已读失败", zap.Error(err))
		response.InternalServerError(c, "标记所有消息已读失败")
		return
	}

	response.SuccessWithMessage(c, "所有消息已标记为已读", nil)
}

// DeleteMessage 删除消息
// @Summary 删除消息
// @Description 删除消息
// @Tags 消息
// @Produce json
// @Param id path string true "消息ID"
// @Success 200 {object} response.APIResponse
// @Router /message/{id} [delete]
func (h *Handler) DeleteMessage(c *gin.Context) {
	// 获取用户ID（从JWT或中间件设置）
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权访问")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		response.InternalServerError(c, "用户ID格式错误")
		return
	}

	messageIDStr := c.Param("id")
	if messageIDStr == "" {
		response.BadRequest(c, "消息ID不能为空")
		return
	}

	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "消息ID格式错误")
		return
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	ctx := c.Request.Context()
	err = h.service.DeleteMessage(ctx, uint(messageID), uint(userIDUint))
	if err != nil {
		h.logger.Error("删除消息失败", zap.Error(err))
		response.InternalServerError(c, "删除消息失败")
		return
	}

	response.SuccessWithMessage(c, "消息删除成功", nil)
}

// GetUnreadCount 获取未读消息数量
// @Summary 获取未读消息数量
// @Description 获取用户未读消息数量
// @Tags 消息
// @Produce json
// @Success 200 {object} response.APIResponse{data=int64}
// @Router /message/unread-count [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	// 获取用户ID（从JWT或中间件设置）
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权访问")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		response.InternalServerError(c, "用户ID格式错误")
		return
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		response.BadRequest(c, "用户ID格式错误")
		return
	}

	ctx := c.Request.Context()
	count, err := h.service.GetUnreadCount(ctx, uint(userIDUint))
	if err != nil {
		h.logger.Error("获取未读消息数量失败", zap.Error(err))
		response.InternalServerError(c, "获取未读消息数量失败")
		return
	}

	response.Success(c, count)
}