package site

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"moviepilot-go/internal/apis/middlewares"
	siteservice "moviepilot-go/internal/business/services/site"
)

// SiteHandler 站点处理器
type SiteHandler struct {
	siteService    siteservice.SiteService
	checkinService siteservice.CheckinService
}

// NewSiteHandler 创建站点处理器
func NewSiteHandler(
	siteService siteservice.SiteService,
	checkinService siteservice.CheckinService,
) *SiteHandler {
	return &SiteHandler{
		siteService:    siteService,
		checkinService: checkinService,
	}
}

// Create 创建站点
// @Summary 创建站点
// @Description 创建新站点
// @Tags 站点
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body siteservice.CreateSiteRequest true "站点信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/sites [post]
func (h *SiteHandler) Create(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	var req siteservice.CreateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	site, err := h.siteService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "站点创建成功",
		"site":    site,
	})
}

// List 获取站点列表
// @Summary 获取站点列表
// @Description 获取当前用户的站点列表
// @Tags 站点
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sites [get]
func (h *SiteHandler) List(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sites, total, err := h.siteService.List(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sites": sites,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Get 获取站点详情
// @Summary 获取站点详情
// @Description 获取指定站点的详细信息
// @Tags 站点
// @Security BearerAuth
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sites/{id} [get]
func (h *SiteHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	site, err := h.siteService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"site": site,
	})
}

// Update 更新站点
// @Summary 更新站点
// @Description 更新站点信息
// @Tags 站点
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "站点ID"
// @Param request body siteservice.UpdateSiteRequest true "站点信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sites/{id} [put]
func (h *SiteHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req siteservice.UpdateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	site, err := h.siteService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "站点更新成功",
		"site":    site,
	})
}

// Delete 删除站点
// @Summary 删除站点
// @Description 删除指定站点
// @Tags 站点
// @Security BearerAuth
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sites/{id} [delete]
func (h *SiteHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.siteService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "站点删除成功",
	})
}

// Checkin 执行签到
// @Summary 站点签到
// @Description 对指定站点执行签到
// @Tags 站点
// @Security BearerAuth
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sites/{id}/checkin [post]
func (h *SiteHandler) Checkin(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	log, err := h.checkinService.Checkin(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "签到成功",
		"log":     log,
	})
}
