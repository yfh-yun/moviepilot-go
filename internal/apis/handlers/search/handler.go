package search

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	searchbiz "moviepilot-go/internal/business/services/search"
	"moviepilot-go/pkg/logger"
)

// Handler 搜索 API 处理器
type Handler struct {
	service searchbiz.Service
	logger  *zap.Logger
}

// NewHandler 创建搜索 API 处理器
func NewHandler(service searchbiz.Service) *Handler {
	return &Handler{
		service: service,
		logger:  logger.GetLogger(),
	}
}

// Search 搜索种子
// @Summary 搜索种子
// @Description 搜索种子资源
// @Tags search
// @Accept json
// @Produce json
// @Param request body searchbiz.SearchRequest true "搜索请求"
// @Success 200 {object} searchbiz.SearchResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search [post]
func (h *Handler) Search(c *gin.Context) {
	var req searchbiz.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Order == "" {
		req.Order = "desc"
	}

	response, err := h.service.Search(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("搜索失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchSimple 简单搜索
// @Summary 简单搜索
// @Description 通过 URL 参数进行简单搜索
// @Tags search
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param sites query string false "站点列表（逗号分隔）"
// @Param category query string false "分类"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" Enums(seeders, leechers, size, date)
// @Param order query string false "排序方向" Enums(asc, desc) default(desc)
// @Param min_seeders query int false "最小做种数"
// @Success 200 {object} searchbiz.SearchResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search [get]
func (h *Handler) SearchSimple(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "关键词不能为空"})
		return
	}

	req := searchbiz.SearchRequest{
		Keyword:  keyword,
		Category: c.Query("category"),
		SortBy:   c.Query("sort_by"),
		Order:    c.DefaultQuery("order", "desc"),
		UseCache: true,
	}

	// 解析站点列表
	if sitesStr := c.Query("sites"); sitesStr != "" {
		req.Sites = splitAndTrim(sitesStr, ",")
	}

	// 解析分页参数
	if page := c.Query("page"); page != "" {
		var p int
		if _, err := fmt.Sscanf(page, "%d", &p); err == nil {
			req.Page = p
		}
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		var ps int
		if _, err := fmt.Sscanf(pageSize, "%d", &ps); err == nil {
			req.PageSize = ps
		}
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 解析最小做种数
	if minSeeders := c.Query("min_seeders"); minSeeders != "" {
		var ms int
		if _, err := fmt.Sscanf(minSeeders, "%d", &ms); err == nil {
			req.MinSeeders = ms
		}
	}

	response, err := h.service.Search(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("搜索失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchMultiSite 多站点搜索
// @Summary 多站点搜索
// @Description 在多个站点搜索种子
// @Tags search
// @Accept json
// @Produce json
// @Param request body MultiSiteSearchRequest true "搜索请求"
// @Success 200 {object} searchbiz.MultiSiteSearchResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/multi [post]
func (h *Handler) SearchMultiSite(c *gin.Context) {
	var req MultiSiteSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("绑定请求参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opts := searchbiz.SearchOptions{
		Category:   req.Category,
		Page:       req.Page,
		PageSize:   req.PageSize,
		SortBy:     req.SortBy,
		Order:      req.Order,
		MinSeeders: req.MinSeeders,
		MaxResults: req.MaxResults,
	}

	response, err := h.service.SearchMultiSite(c.Request.Context(), req.Sites, req.Keyword, opts)
	if err != nil {
		h.logger.Error("多站点搜索失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSearchHistory 获取搜索历史
// @Summary 获取搜索历史
// @Description 获取用户的搜索历史
// @Tags search
// @Produce json
// @Param limit query int false "限制数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/history [get]
func (h *Handler) GetSearchHistory(c *gin.Context) {
	// TODO: 从上下文获取用户 ID
	userID := uint(1)

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		var l int
		if _, err := fmt.Sscanf(limitStr, "%d", &l); err == nil && l > 0 {
			limit = l
		}
	}

	history, err := h.service.GetSearchHistory(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("获取搜索历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  history,
		"total": len(history),
	})
}

// ClearSearchHistory 清除搜索历史
// @Summary 清除搜索历史
// @Description 清除用户的搜索历史
// @Tags search
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/history [delete]
func (h *Handler) ClearSearchHistory(c *gin.Context) {
	// TODO: 从上下文获取用户 ID
	userID := uint(1)

	if err := h.service.ClearSearchHistory(c.Request.Context(), userID); err != nil {
		h.logger.Error("清除搜索历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "搜索历史已清除"})
}

// LastSearchResults 获取最近搜索结果
// @Summary 查询搜索结果
// @Description 查询最近的搜索结果
// @Tags search
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/last [get]
func (h *Handler) LastSearchResults(c *gin.Context) {
	h.logger.Info("获取最近搜索结果请求")

	results, err := h.service.LastSearchResults(c.Request.Context())
	if err != nil {
		h.logger.Error("获取最近搜索结果失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// SearchByID 根据媒体ID搜索
// @Summary 精确搜索资源
// @Description 根据TMDBID/豆瓣ID精确搜索站点资源 tmdb:/douban:/bangumi:
// @Tags search
// @Produce json
// @Param mediaid path string true "媒体ID" format(mediaid)
// @Param mtype query string false "媒体类型"
// @Param area query string false "搜索区域" default(title)
// @Param title query string false "标题"
// @Param year query string false "年份"
// @Param season query string false "季号"
// @Param sites query string false "站点列表（逗号分隔）"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/media/{mediaid} [get]
func (h *Handler) SearchByID(c *gin.Context) {
	mediaID := c.Param("mediaid")
	if mediaID == "" {
		h.logger.Error("缺少媒体ID参数")
		c.JSON(http.StatusBadRequest, gin.H{"error": "媒体ID不能为空"})
		return
	}

	mediaType := c.Query("mtype")
	area := c.DefaultQuery("area", "title")
	title := c.Query("title")
	year := c.Query("year")
	season := c.Query("season")
	sites := c.Query("sites")

	h.logger.Info("根据媒体ID搜索请求",
		zap.String("mediaID", mediaID),
		zap.String("mediaType", mediaType),
		zap.String("area", area),
		zap.String("title", title),
		zap.String("year", year),
		zap.String("season", season),
		zap.String("sites", sites),
	)

	results, err := h.service.SearchByID(c.Request.Context(), mediaID, mediaType, area, title, year, season, sites)
	if err != nil {
		h.logger.Error("根据媒体ID搜索失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未搜索到任何资源",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// SearchByTitle 根据标题搜索
// @Summary 模糊搜索资源
// @Description 根据名称模糊搜索站点资源，支持分页，关键词为空是返回首页资源
// @Tags search
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" default(0)
// @Param sites query string false "站点列表（逗号分隔）"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/title [get]
func (h *Handler) SearchByTitle(c *gin.Context) {
	keyword := c.Query("keyword")
	pageStr := c.DefaultQuery("page", "0")
	page, _ := strconv.Atoi(pageStr)
	sites := c.Query("sites")

	h.logger.Info("根据标题搜索请求",
		zap.String("keyword", keyword),
		zap.Int("page", page),
		zap.String("sites", sites),
	)

	results, err := h.service.SearchByTitle(c.Request.Context(), keyword, page, sites)
	if err != nil {
		h.logger.Error("根据标题搜索失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未搜索到任何资源",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// MultiSiteSearchRequest 多站点搜索请求
type MultiSiteSearchRequest struct {
	Keyword    string   `json:"keyword" binding:"required"`
	Sites      []string `json:"sites" binding:"required"`
	Category   string   `json:"category"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	SortBy     string   `json:"sort_by"`
	Order      string   `json:"order"`
	MinSeeders int      `json:"min_seeders"`
	MaxResults int      `json:"max_results"`
}

// splitAndTrim 分割并去除空格
func splitAndTrim(s string, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
