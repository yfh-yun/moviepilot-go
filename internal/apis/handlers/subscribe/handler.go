package subscribe

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	"moviepilot-go/internal/business/services/subscribe"
	"moviepilot-go/pkg/logger"
)

// Handler 订阅 API 处理器
type Handler struct {
	service *subscribe.Service
	logger  *zap.Logger
}

// NewHandler 创建订阅处理器
func NewHandler(service *subscribe.Service, log *zap.Logger) *Handler {
	if log == nil {
		log = logger.GetLogger()
	}
	return &Handler{
		service: service,
		logger:  log,
	}
}

// CreateSubscribe 创建订阅
// @Summary 创建订阅
// @Description 创建新的媒体订阅
// @Tags 订阅
// @Accept json
// @Produce json
// @Param subscribe body subscribe.CreateSubscribeRequest true "订阅信息"
// @Success 201 {object} database.Subscribe
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes [post]
func (h *Handler) CreateSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	h.logger.Debug("CreateSubscribe called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	var req subscribe.AddSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.Add(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create subscribe",
			zap.String("title", req.Title),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "create_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribe created successfully",
		zap.Int("id", result.SubscribeID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusCreated, result)
}

// GetSubscribe 获取订阅详情
// @Summary 获取订阅详情
// @Description 根据 ID 获取订阅详情
// @Tags 订阅
// @Produce json
// @Param id path int true "订阅 ID"
// @Success 200 {object} database.Subscribe
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/subscribes/{id} [get]
func (h *Handler) GetSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	result, err := h.service.GetSubscribe(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Warn("subscribe not found",
			zap.Uint("id", uint(id)),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "订阅不存在",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListSubscribes 获取订阅列表
// @Summary 获取订阅列表
// @Description 获取订阅列表，支持分页和筛选
// @Tags 订阅
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param state query string false "状态筛选" Enums(N, R, P, S)
// @Param type query string false "类型筛选" Enums(movie, tv)
// @Success 200 {object} map[string]interface{}
// @Router /api/subscribes [get]
func (h *Handler) ListSubscribes(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	state := c.Query("state")
	mediaType := c.Query("type")

	opts := subscribe.ListOptions{
		Page:     page,
		PageSize: pageSize,
		State:    state,
		Type:     mediaType,
	}

	subscribes, total, err := h.service.ListSubscribes(c.Request.Context(), opts)
	if err != nil {
		h.logger.Error("failed to list subscribes",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "list_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     subscribes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateSubscribe 更新订阅
// @Summary 更新订阅
// @Description 更新订阅信息
// @Tags 订阅
// @Accept json
// @Produce json
// @Param id path int true "订阅 ID"
// @Param subscribe body subscribe.UpdateSubscribeRequest true "更新信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/subscribes/{id} [put]
func (h *Handler) UpdateSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	var req subscribe.UpdateSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.UpdateSubscribe(c.Request.Context(), uint(id), req); err != nil {
		h.logger.Error("failed to update subscribe",
			zap.Uint("id", uint(id)),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribe updated successfully",
		zap.Uint("id", uint(id)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅更新成功",
	})
}

// DeleteSubscribe 删除订阅
// @Summary 删除订阅
// @Description 删除指定订阅
// @Tags 订阅
// @Produce json
// @Param id path int true "订阅 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/subscribes/{id} [delete]
func (h *Handler) DeleteSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	if err := h.service.DeleteSubscribe(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("failed to delete subscribe",
			zap.Uint("id", uint(id)),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "delete_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribe deleted successfully",
		zap.Uint("id", uint(id)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅删除成功",
	})
}

// PauseSubscribe 暂停订阅
// @Summary 暂停订阅
// @Description 暂停指定订阅
// @Tags 订阅
// @Produce json
// @Param id path int true "订阅 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/subscribes/{id}/pause [post]
func (h *Handler) PauseSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	if err := h.service.PauseSubscribe(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("failed to pause subscribe",
			zap.Uint("id", uint(id)),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "pause_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribe paused successfully",
		zap.Uint("id", uint(id)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅已暂停",
	})
}

// ResumeSubscribe 恢复订阅
// @Summary 恢复订阅
// @Description 恢复指定订阅
// @Tags 订阅
// @Produce json
// @Param id path int true "订阅 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/subscribes/{id}/resume [post]
func (h *Handler) ResumeSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	if err := h.service.ResumeSubscribe(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("failed to resume subscribe",
			zap.Uint("id", uint(id)),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "resume_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribe resumed successfully",
		zap.Uint("id", uint(id)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅已恢复",
	})
}

// UpdateSubscribeStatus 更新订阅状态
// @Summary 更新订阅状态
// @Description 更新指定订阅的状态
// @Tags 订阅
// @Produce json
// @Param subid path int true "订阅 ID"
// @Param state query string true "状态" Enums(R, P, S)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/subscribes/status/{subid} [put]
func (h *Handler) UpdateSubscribeStatus(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	subidStr := c.Param("subid")
	subid, err := strconv.ParseUint(subidStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	state := c.Query("state")
	validStates := map[string]bool{"R": true, "P": true, "S": true}
	if !validStates[state] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": "无效的订阅状态",
		})
		return
	}

	// 使用 UpdateSubscribe 方法更新状态
	var req subscribe.UpdateSubscribeRequest
	req.State = &state
	if err := h.service.UpdateSubscribe(c.Request.Context(), uint(subid), req); err != nil {
		h.logger.Error("failed to update subscribe status",
			zap.Uint("id", uint(subid)),
			zap.String("state", state),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_status_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribe status updated successfully",
		zap.Uint("id", uint(subid)),
		zap.String("state", state),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅状态更新成功",
	})
}

// GetSubscribeByMediaID 根据媒体ID获取订阅
// @Summary 根据媒体ID获取订阅
// @Description 根据TMDBID/豆瓣ID/BangumiID获取订阅
// @Tags 订阅
// @Produce json
// @Param mediaid path string true "媒体ID (tmdb:/douban:/bangumi:)"
// @Param season query int false "季号"
// @Param title query string false "标题"
// @Success 200 {object} subscribe.Subscribe
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/subscribes/media/{mediaid} [get]
func (h *Handler) GetSubscribeByMediaID(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	mediaid := c.Param("mediaid")
	title := c.Query("title")

	var season *int
	seasonStr := c.Query("season")
	if seasonStr != "" {
		seasonVal, err := strconv.Atoi(seasonStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_season",
				"message": "季号必须是数字",
			})
			return
		}
		season = &seasonVal
	}

	subscribe, err := h.service.GetSubscribeByMediaID(c.Request.Context(), mediaid, season, title)
	if err != nil {
		h.logger.Error("failed to get subscribe by media id",
			zap.String("mediaid", mediaid),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "get_subscribe_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subscribe)
}

// RefreshSubscribes 刷新所有订阅
// @Summary 刷新所有订阅
// @Description 触发刷新所有订阅
// @Tags 订阅
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/refresh [get]
func (h *Handler) RefreshSubscribes(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	if err := h.service.RefreshSubscribes(c.Request.Context()); err != nil {
		h.logger.Error("failed to refresh subscribes",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "refresh_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribes refresh triggered",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅刷新已触发",
	})
}

// ResetSubscribe 重置订阅
// @Summary 重置订阅
// @Description 重置指定订阅的状态和计数
// @Tags 订阅
// @Produce json
// @Param subid path int true "订阅 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/reset/{subid} [get]
func (h *Handler) ResetSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	subidStr := c.Param("subid")
	subid, err := strconv.ParseUint(subidStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	if err := h.service.ResetSubscribe(c.Request.Context(), uint(subid)); err != nil {
		h.logger.Error("failed to reset subscribe",
			zap.Uint("id", uint(subid)),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "reset_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribe reset successfully",
		zap.Uint("id", uint(subid)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅重置成功",
	})
}

// CheckSubscribes 刷新订阅TMDB信息
// @Summary 刷新订阅TMDB信息
// @Description 触发刷新所有订阅的TMDB信息
// @Tags 订阅
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/check [get]
func (h *Handler) CheckSubscribes(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	if err := h.service.CheckSubscribes(c.Request.Context()); err != nil {
		h.logger.Error("failed to check subscribes",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "check_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribes check triggered",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅TMDB信息刷新已触发",
	})
}

// SearchAllSubscribes 搜索所有订阅
// @Summary 搜索所有订阅
// @Description 触发搜索所有订阅
// @Tags 订阅
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/search [get]
func (h *Handler) SearchAllSubscribes(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	if err := h.service.SearchAllSubscribes(c.Request.Context()); err != nil {
		h.logger.Error("failed to search all subscribes",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("search all subscribes triggered",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "所有订阅搜索已触发",
	})
}

// SearchSubscribe 搜索指定订阅
// @Summary 搜索指定订阅
// @Description 触发搜索指定订阅
// @Tags 订阅
// @Produce json
// @Param subscribe_id path int true "订阅 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/search/{subscribe_id} [get]
func (h *Handler) SearchSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	subscribeIDStr := c.Param("subscribe_id")
	subscribeID, err := strconv.ParseUint(subscribeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	if err := h.service.SearchSubscribe(c.Request.Context(), uint(subscribeID)); err != nil {
		h.logger.Error("failed to search subscribe",
			zap.Uint("id", uint(subscribeID)),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("search subscribe triggered",
		zap.Uint("id", uint(subscribeID)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅搜索已触发",
	})
}

// DeleteSubscribeByMediaID 根据媒体ID删除订阅
// @Summary 根据媒体ID删除订阅
// @Description 根据TMDBID/豆瓣ID/BangumiID删除订阅
// @Tags 订阅
// @Produce json
// @Param mediaid path string true "媒体ID (tmdb:/douban:/bangumi:)"
// @Param season query int false "季号"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/media/{mediaid} [delete]
func (h *Handler) DeleteSubscribeByMediaID(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	mediaid := c.Param("mediaid")

	var season *int
	seasonStr := c.Query("season")
	if seasonStr != "" {
		seasonVal, err := strconv.Atoi(seasonStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_season",
				"message": "季号必须是数字",
			})
			return
		}
		season = &seasonVal
	}

	if err := h.service.DeleteSubscribeByMediaID(c.Request.Context(), mediaid, season); err != nil {
		h.logger.Error("failed to delete subscribe by media id",
			zap.String("mediaid", mediaid),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "delete_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("subscribe deleted by media id",
		zap.String("mediaid", mediaid),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "订阅删除成功",
	})
}

// SeerrSubscribe OverSeerr/JellySeerr通知订阅
// @Summary OverSeerr/JellySeerr通知订阅
// @Description 处理OverSeerr/JellySeerr webhook通知
// @Tags 订阅
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Webhook请求体"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/seerr [post]
func (h *Handler) SeerrSubscribe(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "无效的请求体",
		})
		return
	}

	if err := h.service.SeerrSubscribe(c.Request.Context(), req); err != nil {
		h.logger.Error("failed to process seerr webhook",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "seerr_failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info("seerr webhook processed successfully",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "OverSeerr/JellySeerr通知处理成功",
	})
}

// GetPopularSubscribes 获取热门订阅
// @Summary 获取热门订阅
// @Description 获取基于用户共享数据的热门订阅
// @Tags 订阅
// @Produce json
// @Param stype query string true "类型 (movie/tv)"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {array} subscribe.Subscribe
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/popular [get]
func (h *Handler) GetPopularSubscribes(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	stype := c.Query("stype")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "30"))

	subscribes, err := h.service.GetPopularSubscribes(c.Request.Context(), stype, page, count)
	if err != nil {
		h.logger.Error("failed to get popular subscribes",
			zap.String("type", stype),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "get_popular_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subscribes)
}

// GetUserSubscribes 获取用户订阅
// @Summary 获取用户订阅
// @Description 获取指定用户的订阅列表
// @Tags 订阅
// @Produce json
// @Param username path string true "用户名"
// @Success 200 {array} subscribe.Subscribe
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/user/{username} [get]
func (h *Handler) GetUserSubscribes(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	username := c.Param("username")

	subscribes, err := h.service.GetUserSubscribes(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("failed to get user subscribes",
			zap.String("username", username),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "get_user_subscribes_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, subscribes)
}

// GetSubscribeFiles 获取订阅相关文件
// @Summary 获取订阅相关文件
// @Description 获取指定订阅的相关文件信息
// @Tags 订阅
// @Produce json
// @Param subscribe_id path int true "订阅 ID"
// @Success 200 {object} subscribe.SubscribeInfo
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribes/files/{subscribe_id} [get]
func (h *Handler) GetSubscribeFiles(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	subscribeIDStr := c.Param("subscribe_id")
	subscribeID, err := strconv.ParseUint(subscribeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "订阅 ID 必须是数字",
		})
		return
	}

	filesInfo, err := h.service.GetSubscribeFiles(c.Request.Context(), uint(subscribeID))
	if err != nil {
		h.logger.Error("failed to get subscribe files",
			zap.Uint("id", uint(subscribeID)),
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "get_files_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, filesInfo)
}
