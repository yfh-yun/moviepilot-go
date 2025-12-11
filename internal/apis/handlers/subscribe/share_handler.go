package subscribe

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	"moviepilot-go/internal/business/services/subscribe"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/pkg/logger"
)

// ShareHandler 订阅分享 API 处理器
type ShareHandler struct {
	shareService subscribe.ShareService
	logger       *zap.Logger
}

// NewShareHandler 创建分享处理器
func NewShareHandler(shareService subscribe.ShareService) *ShareHandler {
	return &ShareHandler{
		shareService: shareService,
		logger:       logger.GetLogger(),
	}
}

// ShareSubscribe 分享订阅
// @Summary 分享订阅
// @Description 将订阅分享给其他用户
// @Tags subscribe
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ShareSubscribeRequest true "分享请求"
// @Success 200 {object} dto.SubscribeShare
// @Failure 400 {object} map[string]interface{}
// @Router /api/subscribe/share [post]
func (h *ShareHandler) ShareSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	var req dto.ShareSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("ShareSubscribe invalid request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	share, err := h.shareService.ShareSubscribe(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("分享订阅失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    share,
	})
}

// DeleteShare 删除分享
// @Summary 删除分享
// @Description 删除已分享的订阅
// @Tags subscribe
// @Security BearerAuth
// @Produce json
// @Param share_id path int true "分享ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/subscribe/share/{share_id} [delete]
func (h *ShareHandler) DeleteShare(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	shareIDStr := c.Param("share_id")
	shareID, err := strconv.Atoi(shareIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分享ID"})
		return
	}

	if err := h.shareService.DeleteShare(c.Request.Context(), shareID); err != nil {
		h.logger.Error("删除分享失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "分享已删除",
	})
}

// ForkSubscribe 复用订阅
// @Summary 复用订阅
// @Description 从分享中复用订阅
// @Tags subscribe
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param share_id body int true "分享ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/subscribe/fork [post]
func (h *ShareHandler) ForkSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	var req struct {
		ShareID int `json:"share_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.shareService.ForkSubscribe(c.Request.Context(), req.ShareID); err != nil {
		h.logger.Error("复用订阅失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订阅已复用",
	})
}

// GetShares 获取分享列表
// @Summary 获取分享列表
// @Description 获取订阅分享列表
// @Tags subscribe
// @Security BearerAuth
// @Produce json
// @Param name query string false "分享名称"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.SubscribeShare
// @Router /api/subscribe/shares [get]
func (h *ShareHandler) GetShares(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	name := c.Query("name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	shares, err := h.shareService.GetShares(c.Request.Context(), name, page, count)
	if err != nil {
		h.logger.Error("获取分享列表失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("page", page),
			zap.Int("count", count),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, shares)
}

// GetShareStatistics 获取分享统计
// @Summary 获取分享统计
// @Description 获取用户的分享统计信息
// @Tags subscribe
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.ShareStatistics
// @Router /api/subscribe/share/statistics [get]
func (h *ShareHandler) GetShareStatistics(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	stats, err := h.shareService.GetShareStatistics(c.Request.Context(), userIDStr)
	if err != nil {
		h.logger.Error("获取分享统计失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// FollowUser 关注用户
// @Summary 关注用户
// @Description 关注订阅分享人
// @Tags subscribe
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param share_uid body string true "分享人UID"
// @Success 200 {object} map[string]interface{}
// @Router /api/subscribe/follow [post]
func (h *ShareHandler) FollowUser(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	var req struct {
		ShareUID string `json:"share_uid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := strconv.FormatUint(uint64(userID), 10)

	if err := h.shareService.FollowUser(c.Request.Context(), userIDStr, req.ShareUID); err != nil {
		h.logger.Error("关注用户失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "关注成功",
	})
}

// UnfollowUser 取消关注用户
// @Summary 取消关注用户
// @Description 取消关注订阅分享人
// @Tags subscribe
// @Security BearerAuth
// @Produce json
// @Param share_uid query string true "分享人UID"
// @Success 200 {object} map[string]interface{}
// @Router /api/subscribe/follow [delete]
func (h *ShareHandler) UnfollowUser(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	shareUID := c.Query("share_uid")
	if shareUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分享人UID不能为空"})
		return
	}

	userIDStr := strconv.FormatUint(uint64(userID), 10)

	if err := h.shareService.UnfollowUser(c.Request.Context(), userIDStr, shareUID); err != nil {
		h.logger.Error("取消关注失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已取消关注",
	})
}

// GetFollowedUsers 获取关注列表
// @Summary 获取关注列表
// @Description 获取已关注的订阅分享人列表
// @Tags subscribe
// @Security BearerAuth
// @Produce json
// @Success 200 {array} string
// @Router /api/subscribe/follow [get]
func (h *ShareHandler) GetFollowedUsers(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	users, err := h.shareService.GetFollowedUsers(c.Request.Context(), userIDStr)
	if err != nil {
		h.logger.Error("获取关注列表失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetPopularShares 获取热门分享
// @Summary 获取热门分享
// @Description 获取热门订阅分享
// @Tags subscribe
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Success 200 {array} dto.SubscribeShare
// @Router /api/subscribe/popular [get]
func (h *ShareHandler) GetPopularShares(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	shares, err := h.shareService.GetPopularShares(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("获取热门分享失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("limit", limit),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, shares)
}
