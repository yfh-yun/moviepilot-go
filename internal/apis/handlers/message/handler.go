package message

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	messagebiz "moviepilot-go/internal/business/services/message"
	"moviepilot-go/pkg/logger"
)

// Handler 消息 API 处理器
type Handler struct {
	messageService messagebiz.Service
	logger         *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(messageService messagebiz.Service) *Handler {
	return &Handler{
		messageService: messageService,
		logger:         logger.GetLogger(),
	}
}

// CreateMessage 创建消息
// @Summary 创建消息
// @Description 创建新消息
// @Tags message
// @Accept json
// @Produce json
// @Param message body messagebiz.Message true "消息内容"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message [post]
func (h *Handler) CreateMessage(c *gin.Context) {
	var msg messagebiz.Message

	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.messageService.CreateMessage(c.Request.Context(), &msg); err != nil {
		h.logger.Error("创建消息失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "创建成功", "id": msg.ID})
}

// GetMessages 获取消息列表
// @Summary 获取消息列表
// @Description 获取用户的消息列表
// @Tags message
// @Produce json
// @Param user_id query string true "用户ID"
// @Param limit query int false "数量限制" default(50)
// @Success 200 {array} messagebiz.Message
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message [get]
func (h *Handler) GetMessages(c *gin.Context) {
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

	messages, err := h.messageService.GetMessages(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("获取消息列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetUnreadCount 获取未读消息数
// @Summary 获取未读消息数
// @Description 获取用户的未读消息数量
// @Tags message
// @Produce json
// @Param user_id query string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message/unread [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	count, err := h.messageService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("获取未读消息数失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkAsRead 标记为已读
// @Summary 标记为已读
// @Description 标记消息为已读
// @Tags message
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message/{id}/read [post]
func (h *Handler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的消息ID"})
		return
	}

	if err := h.messageService.MarkAsRead(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("标记消息为已读失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "标记成功"})
}

// DeleteMessage 删除消息
// @Summary 删除消息
// @Description 删除指定消息
// @Tags message
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message/{id} [delete]
func (h *Handler) DeleteMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的消息ID"})
		return
	}

	if err := h.messageService.DeleteMessage(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("删除消息失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ClearMessages 清空消息
// @Summary 清空消息
// @Description 清空用户的所有消息
// @Tags message
// @Produce json
// @Param user_id query string true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message/clear [post]
func (h *Handler) ClearMessages(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	if err := h.messageService.ClearMessages(c.Request.Context(), userID); err != nil {
		h.logger.Error("清空消息失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "清空成功"})
}

// UserMessage 接收用户消息
// @Summary 接收用户消息
// @Description 用户消息响应，配置请求中需要添加参数：token=API_TOKEN&source=消息配置名
// @Tags message
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message [post]
func (h *Handler) UserMessage(c *gin.Context) {
	// 获取请求信息
	body, _ := c.GetRawData()
	form, _ := c.MultipartForm()
	args := c.Request.URL.Query()

	h.logger.Info("Received user message",
		zap.String("body", string(body)),
		zap.Any("form", form),
		zap.Any("args", args),
	)

	// TODO: 实现消息处理逻辑
	// 1. 启动消息链处理
	// 2. 返回成功响应

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// WebMessage 接收WEB消息
// @Summary 接收WEB消息
// @Description WEB消息响应
// @Tags message
// @Produce json
// @Param text query string true "消息内容"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message/web [post]
func (h *Handler) WebMessage(c *gin.Context) {
	text := c.Query("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息内容不能为空"})
		return
	}

	// TODO: 实现WEB消息处理逻辑
	// 1. 获取当前用户
	// 2. 处理WEB消息
	// 3. 返回成功响应

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetWebMessages 获取WEB消息列表
// @Summary 获取WEB消息列表
// @Description 获取WEB消息列表
// @Tags message
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message/web [get]
func (h *Handler) GetWebMessages(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	countStr := c.DefaultQuery("count", "20")

	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}
	count, _ := strconv.Atoi(countStr)
	if count <= 0 {
		count = 20
	}
	if count > 100 {
		count = 100
	}

	// TODO: 实现获取WEB消息列表逻辑
	// 1. 分页获取消息
	// 2. 返回消息列表

	c.JSON(http.StatusOK, []map[string]any{})
}

// IncomingVerify 回调请求验证
// @Summary 回调请求验证
// @Description 微信/VoceChat等验证响应
// @Tags message
// @Produce json
// @Param token query string false "Token"
// @Param echostr query string false "Echo字符串"
// @Param msg_signature query string false "消息签名"
// @Param timestamp query string false "时间戳"
// @Param nonce query string false "随机数"
// @Param source query string false "消息源"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message [get]
func (h *Handler) IncomingVerify(c *gin.Context) {
	token := c.Query("token")
	echostr := c.Query("echostr")
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	source := c.Query("source")

	h.logger.Info("Received verification request",
		zap.String("token", token),
		zap.String("echostr", echostr),
		zap.String("msg_signature", msgSignature),
		zap.String("timestamp", timestamp),
		zap.String("nonce", nonce),
		zap.String("source", source),
	)

	// 微信验证
	if echostr != "" && msgSignature != "" && timestamp != "" && nonce != "" {
		// TODO: 实现微信验证逻辑
		// 1. 验证签名
		// 2. 返回echostr
		return
	}

	// VoceChat验证
	c.JSON(http.StatusOK, map[string]any{
		"status": "OK",
	})
}

// WebPushSubscribe 客户端webpush通知订阅
// @Summary 客户端webpush通知订阅
// @Description 客户端webpush通知订阅
// @Tags message
// @Accept json
// @Produce json
// @Param subscription body map[string]interface{} true "订阅信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message/webpush/subscribe [post]
func (h *Handler) WebPushSubscribe(c *gin.Context) {
	var subscription map[string]any
	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现webpush订阅逻辑
	// 1. 保存订阅信息
	// 2. 返回成功响应

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// WebPushSend 发送webpush通知
// @Summary 发送webpush通知
// @Description 发送webpush通知
// @Tags message
// @Accept json
// @Produce json
// @Param payload body map[string]interface{} true "通知内容"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/message/webpush/send [post]
func (h *Handler) WebPushSend(c *gin.Context) {
	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 实现发送webpush通知逻辑
	// 1. 获取所有订阅
	// 2. 发送通知
	// 3. 返回成功响应

	c.JSON(http.StatusOK, gin.H{"success": true})
}
