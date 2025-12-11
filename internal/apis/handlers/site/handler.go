package site

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/apis/middleware"
	siteservice "moviepilot-go/internal/business/services/site"
	"moviepilot-go/pkg/logger"
)

// Handler 站点 API 处理器
type Handler struct {
	logger      *zap.Logger
	siteService siteservice.SiteService
}

// NewHandler 创建站点 API 处理器
func NewHandler(siteService siteservice.SiteService) *Handler {
	return &Handler{
		logger:      logger.GetLogger(),
		siteService: siteService,
	}
}

// CreateSite 创建站点
// @Summary 创建站点
// @Description 创建新的站点
// @Tags sites
// @Accept json
// @Produce json
// @Param site body siteservice.CreateSiteRequest true "站点信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites [post]
func (h *Handler) CreateSite(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	h.logger.Debug("CreateSite called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	// 检查siteService是否为空
	if h.siteService == nil {
		h.logger.Error("站点服务未初始化",
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "站点服务未初始化",
		})
		return
	}

	var req siteservice.CreateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDUint, _ := strconv.ParseUint(userID, 10, 32)

	// 创建站点
	site, err := h.siteService.Create(c.Request.Context(), uint(userIDUint), &req)
	if err != nil {
		h.logger.Error("创建站点失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建站点失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "站点创建成功",
		"data":    site,
	})
}

// UpdateSite 更新站点
// @Summary 更新站点
// @Description 更新站点信息
// @Tags sites
// @Accept json
// @Produce json
// @Param id path int true "站点 ID"
// @Param site body siteservice.UpdateSiteRequest true "站点信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites/{id} [put]
func (h *Handler) UpdateSite(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	// 检查siteService是否为空
	if h.siteService == nil {
		h.logger.Error("站点服务未初始化",
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "站点服务未初始化",
		})
		return
	}

	var req siteservice.UpdateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新站点
	site, err := h.siteService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("更新站点失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新站点失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "站点更新成功",
		"data":    site,
	})
}

// DeleteSite 删除站点
// @Summary 删除站点
// @Description 删除站点
// @Tags sites
// @Produce json
// @Param id path int true "站点 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites/{id} [delete]
func (h *Handler) DeleteSite(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	// 检查siteService是否为空
	if h.siteService == nil {
		h.logger.Error("站点服务未初始化",
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "站点服务未初始化",
		})
		return
	}

	// 删除站点
	err = h.siteService.Delete(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("删除站点失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除站点失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "站点删除成功",
		"site_id": id,
	})
}

// GetSite 获取站点详情
// @Summary 获取站点详情
// @Description 获取站点详细信息
// @Tags sites
// @Produce json
// @Param id path int true "站点 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites/{id} [get]
func (h *Handler) GetSite(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	// 检查siteService是否为空
	if h.siteService == nil {
		h.logger.Error("站点服务未初始化",
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "站点服务未初始化",
		})
		return
	}

	// 获取站点详情
	site, err := h.siteService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("获取站点详情失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "站点不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    site,
	})
}

// ListSites 获取站点列表
// @Summary 获取站点列表
// @Description 获取所有站点，按优先级排序
// @Tags sites
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites [get]
func (h *Handler) ListSites(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	userIDUint, _ := strconv.ParseUint(userID, 10, 32)

	// 检查siteService是否为空
	if h.siteService == nil {
		h.logger.Error("站点服务未初始化",
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "站点服务未初始化",
		})
		return
	}

	// 获取所有站点，按优先级排序
	sites, err := h.siteService.ListByPriority(c.Request.Context(), uint(userIDUint))
	if err != nil {
		h.logger.Error("获取站点列表失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取站点列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sites,
	})
}

// EnableSite 启用站点
// @Summary 启用站点
// @Description 启用站点
// @Tags sites
// @Produce json
// @Param id path int true "站点 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites/{id}/enable [post]
func (h *Handler) EnableSite(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	// 简化实现：直接返回成功
	h.logger.Info("站点启用功能待实现",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.Uint64("site_id", id),
	)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "站点启用功能待实现",
		"site_id": id,
	})
}

// DisableSite 禁用站点
// @Summary 禁用站点
// @Description 禁用站点
// @Tags sites
// @Produce json
// @Param id path int true "站点 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites/{id}/disable [post]
func (h *Handler) DisableSite(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	// 简化实现：直接返回成功
	h.logger.Info("站点禁用功能待实现",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.Uint64("site_id", id),
	)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "站点禁用功能待实现",
		"site_id": id,
	})
}

// UpdateSitesPriority 批量更新站点优先级
// @Summary 批量更新站点优先级
// @Description 批量更新站点优先级
// @Tags sites
// @Accept json
// @Produce json
// @Param priorities body []map[string]int true "站点优先级列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites/priorities [post]
func (h *Handler) UpdateSitesPriority(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	h.logger.Debug("UpdateSitesPriority called",
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
	)

	// 检查siteService是否为空
	if h.siteService == nil {
		h.logger.Error("站点服务未初始化",
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "站点服务未初始化",
		})
		return
	}

	var priorities []map[string]int
	if err := c.ShouldBindJSON(&priorities); err != nil {
		h.logger.Error("绑定请求参数失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为map[uint]int格式
	priorityMap := make(map[uint]int)
	for _, p := range priorities {
		id := uint(p["id"])
		priority := p["pri"]
		priorityMap[id] = priority
	}

	// 批量更新优先级
	err := h.siteService.UpdatePriorities(c.Request.Context(), priorityMap)
	if err != nil {
		h.logger.Error("批量更新站点优先级失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "批量更新站点优先级失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "批量更新站点优先级成功",
	})
}

// TestSite 测试站点连接
// @Summary 测试站点连接
// @Description 测试站点连接是否正常
// @Tags sites
// @Produce json
// @Param id path int true "站点 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/sites/{id}/test [post]
func (h *Handler) TestSite(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	// 检查siteService是否为空
	if h.siteService == nil {
		h.logger.Error("站点服务未初始化",
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "站点服务未初始化",
		})
		return
	}

	// 验证站点Cookie
	isValid, err := h.siteService.ValidateCookie(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("测试站点连接失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.String("user_id", userID),
			zap.Uint64("site_id", id),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "测试站点连接失败",
		})
		return
	}

	if isValid {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "站点连接正常",
			"site_id": id,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "站点连接失败，Cookie可能无效",
			"site_id": id,
		})
	}
}
