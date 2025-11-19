package subscribe

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yfh-yun/moviepilot-go/internal/service/subscribe"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
)

// Handler 订阅处理器
type Handler struct {
	service *subscribe.SubscribeService
}

// NewHandler 创建订阅处理器
func NewHandler(service *subscribe.SubscribeService) *Handler {
	return &Handler{
		service: service,
	}
}

// Create 创建订阅
// @Summary 创建订阅
// @Description 创建一个新的媒体订阅
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param request body subscribe.CreateRequest true "订阅信息"
// @Success 200 {object} response.Response{data=subscribe.SubscribeResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	var req subscribe.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.CodeInvalidParam, err.Error())
		return
	}

	// 从上下文获取用户名
	if username, exists := c.Get("username"); exists {
		req.Username = username.(string)
	}

	result, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		if err == subscribe.ErrSubscribeExists {
			response.Error(c, response.CodeInvalidParam, "订阅已存在")
			return
		}
		if err == subscribe.ErrMediaInfoNotFound {
			response.Error(c, response.CodeNotFound, "未找到媒体信息")
			return
		}
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, result)
}

// Update 更新订阅
// @Summary 更新订阅
// @Description 更新订阅信息
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param request body subscribe.UpdateRequest true "更新信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe [put]
// @Security BearerAuth
func (h *Handler) Update(c *gin.Context) {
	var req subscribe.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.CodeInvalidParam, err.Error())
		return
	}

	if err := h.service.Update(c.Request.Context(), &req); err != nil {
		if err == subscribe.ErrSubscribeNotFound {
			response.Error(c, response.CodeNotFound, "订阅不存在")
			return
		}
		if err == subscribe.ErrInvalidState {
			response.Error(c, response.CodeInvalidParam, "无效的订阅状态")
			return
		}
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除订阅
// @Summary 删除订阅
// @Description 根据ID删除订阅
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path int true "订阅ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe/{id} [delete]
// @Security BearerAuth
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParam, "无效的订阅ID")
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		if err == subscribe.ErrSubscribeNotFound {
			response.Error(c, response.CodeNotFound, "订阅不存在")
			return
		}
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetByID 获取订阅详情
// @Summary 获取订阅详情
// @Description 根据ID获取订阅详细信息
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path int true "订阅ID"
// @Success 200 {object} response.Response{data=subscribe.SubscribeResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe/{id} [get]
// @Security BearerAuth
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParam, "无效的订阅ID")
		return
	}

	result, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == subscribe.ErrSubscribeNotFound {
			response.Error(c, response.CodeNotFound, "订阅不存在")
			return
		}
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, result)
}

// List 获取订阅列表
// @Summary 获取订阅列表
// @Description 获取所有订阅列表(分页)
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData}
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe [get]
// @Security BearerAuth
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	subscribes, total, err := h.service.List(c.Request.Context(), offset, pageSize)
	if err != nil {
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, response.PageData{
		Total: total,
		Items: subscribes,
		Page:  page,
		Size:  pageSize,
	})
}

// ListByUsername 获取用户订阅列表
// @Summary 获取用户订阅列表
// @Description 根据用户名获取订阅列表
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param username path string true "用户名"
// @Param state query string false "状态" Enums(N, R, P, S)
// @Param type query string false "类型" Enums(movie, tv)
// @Success 200 {object} response.Response{data=[]subscribe.SubscribeResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe/user/{username} [get]
// @Security BearerAuth
func (h *Handler) ListByUsername(c *gin.Context) {
	username := c.Param("username")
	state := c.Query("state")
	mtype := c.Query("type")

	subscribes, err := h.service.ListByUsername(c.Request.Context(), username, state, mtype)
	if err != nil {
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, subscribes)
}

// UpdateState 更新订阅状态
// @Summary 更新订阅状态
// @Description 更新订阅的状态
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path int true "订阅ID"
// @Param state path string true "状态" Enums(N, R, P, S)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe/{id}/state/{state} [put]
// @Security BearerAuth
func (h *Handler) UpdateState(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParam, "无效的订阅ID")
		return
	}

	state := c.Param("state")
	if err := h.service.UpdateState(c.Request.Context(), uint(id), state); err != nil {
		if err == subscribe.ErrSubscribeNotFound {
			response.Error(c, response.CodeNotFound, "订阅不存在")
			return
		}
		if err == subscribe.ErrInvalidState {
			response.Error(c, response.CodeInvalidParam, "无效的订阅状态")
			return
		}
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// Reset 重置订阅
// @Summary 重置订阅
// @Description 重置订阅的状态和进度
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param id path int true "订阅ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe/{id}/reset [post]
// @Security BearerAuth
func (h *Handler) Reset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.CodeInvalidParam, "无效的订阅ID")
		return
	}

	if err := h.service.Reset(c.Request.Context(), uint(id)); err != nil {
		if err == subscribe.ErrSubscribeNotFound {
			response.Error(c, response.CodeNotFound, "订阅不存在")
			return
		}
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetByMediaID 根据媒体ID获取订阅
// @Summary 根据媒体ID获取订阅
// @Description 根据TMDB/豆瓣/Bangumi ID获取订阅信息
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param media_id path string true "媒体ID (tmdb:123, douban:123456, bangumi:123)"
// @Param season query int false "季数"
// @Success 200 {object} response.Response{data=subscribe.SubscribeResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/subscribe/media/{media_id} [get]
// @Security BearerAuth
func (h *Handler) GetByMediaID(c *gin.Context) {
	mediaID := c.Param("media_id")
	var season *int
	if seasonStr := c.Query("season"); seasonStr != "" {
		s, err := strconv.Atoi(seasonStr)
		if err == nil {
			season = &s
		}
	}

	result, err := h.service.GetByMediaID(c.Request.Context(), mediaID, season)
	if err != nil {
		if err == subscribe.ErrSubscribeNotFound {
			response.Error(c, response.CodeNotFound, "订阅不存在")
			return
		}
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, result)
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	subscribe := router.Group("/subscribe")
	{
		subscribe.POST("", h.Create)                       // 创建订阅
		subscribe.GET("", h.List)                          // 获取订阅列表
		subscribe.PUT("", h.Update)                        // 更新订阅
		subscribe.GET("/:id", h.GetByID)                   // 获取订阅详情
		subscribe.DELETE("/:id", h.Delete)                 // 删除订阅
		subscribe.PUT("/:id/state/:state", h.UpdateState)  // 更新订阅状态
		subscribe.POST("/:id/reset", h.Reset)              // 重置订阅
		subscribe.GET("/user/:username", h.ListByUsername) // 获取用户订阅
		subscribe.GET("/media/:media_id", h.GetByMediaID)  // 根据媒体ID获取订阅
	}
}
