package site

import (
	"fmt"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/business/services/site"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SiteHandler 站点处理器
type SiteHandler struct {
	siteService *site.SiteService
}

// NewSiteHandler 创建站点处理器实例
func NewSiteHandler(siteService *site.SiteService) *SiteHandler {
	return &SiteHandler{
		siteService: siteService,
	}
}

// CreateSite 创建站点
// @Summary 创建站点
// @Tags 站点管理
// @Accept json
// @Produce json
// @Param site body models.Site true "站点信息"
// @Success 201 {object} models.Site "创建成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites [post]
func (h *SiteHandler) CreateSite(c *gin.Context) {
	var siteReq models.Site
	if err := c.ShouldBindJSON(&siteReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	if err := h.siteService.CreateSite(&siteReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建站点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, siteReq)
}

// GetSiteByID 根据ID获取站点
// @Summary 根据ID获取站点
// @Tags 站点管理
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} models.Site "获取成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 404 {object} map[string]string "站点不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/{id} [get]
func (h *SiteHandler) GetSiteByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	site, err := h.siteService.GetSiteByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "站点不存在: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, site)
}

// GetSiteByName 根据名称获取站点
// @Summary 根据名称获取站点
// @Tags 站点管理
// @Produce json
// @Param name path string true "站点名称"
// @Success 200 {object} models.Site "获取成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 404 {object} map[string]string "站点不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/name/{name} [get]
func (h *SiteHandler) GetSiteByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "站点名称不能为空"})
		return
	}

	site, err := h.siteService.GetSiteByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "站点不存在: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, site)
}

// UpdateSite 更新站点
// @Summary 更新站点
// @Tags 站点管理
// @Accept json
// @Produce json
// @Param id path int true "站点ID"
// @Param site body models.Site true "站点信息"
// @Success 200 {object} models.Site "更新成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 404 {object} map[string]string "站点不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/{id} [put]
func (h *SiteHandler) UpdateSite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	var siteReq models.Site
	if err := c.ShouldBindJSON(&siteReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	siteReq.ID = uint(id)

	if err := h.siteService.UpdateSite(&siteReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新站点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, siteReq)
}

// DeleteSite 删除站点
// @Summary 删除站点
// @Tags 站点管理
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]string "删除成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 404 {object} map[string]string "站点不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/{id} [delete]
func (h *SiteHandler) DeleteSite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	if err := h.siteService.DeleteSite(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除站点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "站点删除成功"})
}

// ListSites 获取站点列表
// @Summary 获取站点列表
// @Tags 站点管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites [get]
func (h *SiteHandler) ListSites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	sites, total, err := h.siteService.ListSites(offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取站点列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sites":       sites,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// SearchSites 搜索站点
// @Summary 搜索站点
// @Tags 站点管理
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "搜索成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/search [get]
func (h *SiteHandler) SearchSites(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	sites, total, err := h.siteService.SearchSites(keyword, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索站点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sites":       sites,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		"keyword":     keyword,
	})
}

// GetActiveSites 获取活跃站点
// @Summary 获取活跃站点
// @Tags 站点管理
// @Produce json
// @Success 200 {object} []models.Site "获取成功"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/active [get]
func (h *SiteHandler) GetActiveSites(c *gin.Context) {
	sites, err := h.siteService.GetActiveSites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取活跃站点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, sites)
}

// GetRSSSites 获取RSS站点
// @Summary 获取RSS站点
// @Tags 站点管理
// @Produce json
// @Success 200 {object} []models.Site "获取成功"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/rss [get]
func (h *SiteHandler) GetRSSSites(c *gin.Context) {
	sites, err := h.siteService.GetRSSSites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取RSS站点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, sites)
}

// GetSearchSites 获取搜索站点
// @Summary 获取搜索站点
// @Tags 站点管理
// @Produce json
// @Success 200 {object} []models.Site "获取成功"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/search-enabled [get]
func (h *SiteHandler) GetSearchSites(c *gin.Context) {
	sites, err := h.siteService.GetSearchSites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取搜索站点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, sites)
}

// UpdateSiteCookie 更新站点Cookie
// @Summary 更新站点Cookie
// @Tags 站点管理
// @Accept json
// @Produce json
// @Param id path int true "站点ID"
// @Param cookie body map[string]string true "Cookie信息"
// @Success 200 {object} map[string]string "更新成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 404 {object} map[string]string "站点不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/{id}/cookie [put]
func (h *SiteHandler) UpdateSiteCookie(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	var cookieReq struct {
		Cookie string `json:"cookie" binding:"required"`
	}

	if err := c.ShouldBindJSON(&cookieReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	if err := h.siteService.UpdateSiteCookie(uint(id), cookieReq.Cookie); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新Cookie失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cookie更新成功"})
}

// UpdateSiteSettings 更新站点设置
// @Summary 更新站点设置
// @Tags 站点管理
// @Accept json
// @Produce json
// @Param id path int true "站点ID"
// @Param settings body map[string]interface{} true "设置信息"
// @Success 200 {object} map[string]string "更新成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 404 {object} map[string]string "站点不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/{id}/settings [put]
func (h *SiteHandler) UpdateSiteSettings(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	if err := h.siteService.UpdateSiteSettings(uint(id), settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置更新成功"})
}

// ToggleSiteActive 切换站点激活状态
// @Summary 切换站点激活状态
// @Tags 站点管理
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} models.Site "切换成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 404 {object} map[string]string "站点不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/{id}/toggle-active [put]
func (h *SiteHandler) ToggleSiteActive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	site, err := h.siteService.ToggleSiteActive(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "切换状态失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, site)
}

// GetSiteStatistics 获取站点统计信息
// @Summary 获取站点统计信息
// @Tags 站点管理
// @Produce json
// @Param name path string true "站点名称"
// @Success 200 {object} []models.SiteStatistic "获取成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/{name}/statistics [get]
func (h *SiteHandler) GetSiteStatistics(c *gin.Context) {
	siteName := c.Param("name")
	if siteName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "站点名称不能为空"})
		return
	}

	statistics, err := h.siteService.GetSiteStatistics(siteName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, statistics)
}

// GetSiteUserData 获取站点用户数据
// @Summary 获取站点用户数据
// @Tags 站点管理
// @Produce json
// @Param name path string true "站点名称"
// @Success 200 {object} []models.SiteUserData "获取成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/{name}/user-data [get]
func (h *SiteHandler) GetSiteUserData(c *gin.Context) {
	siteName := c.Param("name")
	if siteName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "站点名称不能为空"})
		return
	}

	userData, err := h.siteService.GetSiteUserData(siteName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户数据失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, userData)
}

// ImportSites 导入站点数据
// @Summary 导入站点数据
// @Tags 站点管理
// @Accept json
// @Produce json
// @Param sites body []models.Site true "站点列表"
// @Success 200 {object} map[string]string "导入成功"
// @Failure 400 {object} map[string]string "参数错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/sites/import [post]
func (h *SiteHandler) ImportSites(c *gin.Context) {
	var sites []models.Site
	if err := c.ShouldBindJSON(&sites); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	if len(sites) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "站点列表不能为空"})
		return
	}

	// 转换为指针数组
	sitePtrs := make([]*models.Site, len(sites))
	for i := range sites {
		sitePtrs[i] = &sites[i]
	}

	if err := h.siteService.ImportSiteData(sitePtrs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入站点失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("成功导入 %d 个站点", len(sites))})
}
