// Package search 搜索管理API处理器
package search

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/response"
	"moviepilot-go/pkg/validator"
)

// SearchHandler 搜索管理处理器
// 提供媒体内容搜索、索引管理、搜索历史等功能
type SearchHandler struct {
	searchService service.SearchService
	logger        *zap.Logger
}

// NewSearchHandler 创建搜索管理处理器
func NewSearchHandler(searchService service.SearchService, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

// SearchRequest 搜索请求结构体
type SearchRequest struct {
	Query     string `form:"query" binding:"required,min=1,max=200"`
	Type      string `form:"type" binding:"omitempty,oneof=movie tv book music person"`
	Year      int    `form:"year" binding:"omitempty,min=1900,max=2030"`
	Genre     string `form:"genre" binding:"omitempty,min=1,max=50"`
	Language  string `form:"language" binding:"omitempty,min=2,max=10"`
	Country   string `form:"country" binding:"omitempty,min=2,max=10"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=50"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=relevance rating date title"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// SearchResultItem 搜索结果项结构体
type SearchResultItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Year        int       `json:"year"`
	Genres      []string  `json:"genres"`
	Rating      float64   `json:"rating"`
	Poster      string    `json:"poster"`
	Overview    string    `json:"overview"`
	ReleaseDate time.Time `json:"release_date"`
	Score       float64   `json:"score"`
	Source      string    `json:"source"`
}

// SearchResponse 搜索响应结构体
type SearchResponse struct {
	Results []SearchResultItem `json:"results"`
	Total   int64              `json:"total"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Query   string             `json:"query"`
}

// SearchHistoryItem 搜索历史项结构体
type SearchHistoryItem struct {
	ID          string    `json:"id"`
	Query       string    `json:"query"`
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	ResultCount int       `json:"result_count"`
}

// SearchHistoryResponse 搜索历史响应结构体
type SearchHistoryResponse struct {
	History []SearchHistoryItem `json:"history"`
	Total   int64               `json:"total"`
}

// Search 执行搜索
// @Summary 执行搜索
// @Description 根据关键词和筛选条件搜索媒体内容
// @Tags search
// @Accept json
// @Produce json
// @Param request query SearchRequest true "搜索参数"
// @Success 200 {object} response.SuccessResponse{data=SearchResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	var req SearchRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("搜索请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("搜索请求参数验证失败", zap.Error(err))
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
	if req.SortBy == "" {
		req.SortBy = "relevance"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	ctx := c.Request.Context()

	// 调用服务层执行搜索
	results, total, err := h.searchService.Search(ctx, service.SearchParams{
		Query:     req.Query,
		Type:      req.Type,
		Year:      req.Year,
		Genre:     req.Genre,
		Language:  req.Language,
		Country:   req.Country,
		Page:      req.Page,
		Limit:     req.Limit,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	})

	if err != nil {
		h.logger.Error("搜索失败", zap.Error(err), zap.String("query", req.Query))
		response.InternalServerError(c, "搜索失败")
		return
	}

	// 转换为响应格式
	var responseResults []SearchResultItem
	for _, result := range results {
		responseResults = append(responseResults, SearchResultItem{
			ID:          result.ID,
			Title:       result.Title,
			Type:        result.Type,
			Year:        result.Year,
			Genres:      result.Genres,
			Rating:      result.Rating,
			Poster:      result.Poster,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			Score:       result.Score,
			Source:      result.Source,
		})
	}

	response.Success(c, SearchResponse{
		Results: responseResults,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		Query:   req.Query,
	})
}

// GetSearchHistory 获取搜索历史
// @Summary 获取搜索历史
// @Description 获取用户的搜索历史记录
// @Tags search
// @Accept json
// @Produce json
// @Param limit query int false "返回数量限制"
// @Success 200 {object} response.SuccessResponse{data=SearchHistoryResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/search/history [get]
func (h *SearchHandler) GetSearchHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequest(c, "limit参数格式错误")
		return
	}

	if limit < 1 || limit > 100 {
		response.BadRequest(c, "limit参数必须在1-100之间")
		return
	}

	ctx := c.Request.Context()

	// 调用服务层获取搜索历史
	history, total, err := h.searchService.GetSearchHistory(ctx, limit)
	if err != nil {
		h.logger.Error("获取搜索历史失败", zap.Error(err))
		response.InternalServerError(c, "获取搜索历史失败")
		return
	}

	// 转换为响应格式
	var responseHistory []SearchHistoryItem
	for _, item := range history {
		responseHistory = append(responseHistory, SearchHistoryItem{
			ID:          item.ID,
			Query:       item.Query,
			Type:        item.Type,
			Timestamp:   item.Timestamp,
			ResultCount: item.ResultCount,
		})
	}

	response.Success(c, SearchHistoryResponse{
		History: responseHistory,
		Total:   total,
	})
}

// ClearSearchHistory 清空搜索历史
// @Summary 清空搜索历史
// @Description 清空用户的搜索历史记录
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/search/history [delete]
func (h *SearchHandler) ClearSearchHistory(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.searchService.ClearSearchHistory(ctx)
	if err != nil {
		h.logger.Error("清空搜索历史失败", zap.Error(err))
		response.InternalServerError(c, "清空搜索历史失败")
		return
	}

	logger.Info("搜索历史清空成功")
	response.Success(c, gin.H{
		"message": "搜索历史清空成功",
	})
}

// DeleteSearchHistoryItem 删除搜索历史项
// @Summary 删除搜索历史项
// @Description 删除指定的搜索历史记录
// @Tags search
// @Accept json
// @Produce json
// @Param id path string true "搜索历史项ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/search/history/{id} [delete]
func (h *SearchHandler) DeleteSearchHistoryItem(c *gin.Context) {
	itemID := c.Param("id")

	if itemID == "" {
		response.BadRequest(c, "搜索历史项ID不能为空")
		return
	}

	ctx := c.Request.Context()

	err := h.searchService.DeleteSearchHistoryItem(ctx, itemID)
	if err != nil {
		if err == service.ErrSearchHistoryItemNotFound {
			response.NotFound(c, "搜索历史项不存在")
			return
		}
		h.logger.Error("删除搜索历史项失败", zap.Error(err), zap.String("item_id", itemID))
		response.InternalServerError(c, "删除搜索历史项失败")
		return
	}

	logger.Info("搜索历史项删除成功", zap.String("item_id", itemID))
	response.Success(c, gin.H{
		"message": "搜索历史项删除成功",
		"item_id": itemID,
	})
}

// GetSearchSuggestions 获取搜索建议
// @Summary 获取搜索建议
// @Description 根据输入的关键词获取搜索建议
// @Tags search
// @Accept json
// @Produce json
// @Param query query string true "搜索关键词"
// @Success 200 {object} response.SuccessResponse{data=[]string}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/search/suggestions [get]
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("query")

	if query == "" {
		response.BadRequest(c, "搜索关键词不能为空")
		return
	}

	ctx := c.Request.Context()

	suggestions, err := h.searchService.GetSearchSuggestions(ctx, query)
	if err != nil {
		h.logger.Error("获取搜索建议失败", zap.Error(err), zap.String("query", query))
		response.InternalServerError(c, "获取搜索建议失败")
		return
	}

	response.Success(c, suggestions)
}

// RebuildSearchIndex 重建搜索索引
// @Summary 重建搜索索引
// @Description 重建搜索索引，用于更新搜索数据
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/search/reindex [post]
func (h *SearchHandler) RebuildSearchIndex(c *gin.Context) {
	ctx := c.Request.Context()

	err := h.searchService.RebuildSearchIndex(ctx)
	if err != nil {
		h.logger.Error("重建搜索索引失败", zap.Error(err))
		response.InternalServerError(c, "重建搜索索引失败")
		return
	}

	logger.Info("搜索索引重建成功")
	response.Success(c, gin.H{
		"message": "搜索索引重建成功",
	})
}

// GetSearchStats 获取搜索统计信息
// @Summary 获取搜索统计信息
// @Description 获取搜索相关的统计信息
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=service.SearchStats}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/search/stats [get]
func (h *SearchHandler) GetSearchStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.searchService.GetSearchStats(ctx)
	if err != nil {
		h.logger.Error("获取搜索统计信息失败", zap.Error(err))
		response.InternalServerError(c, "获取搜索统计信息失败")
		return
	}

	response.Success(c, stats)
}

// GetPopularSearches 获取热门搜索
// @Summary 获取热门搜索
// @Description 获取热门搜索关键词
// @Tags search
// @Accept json
// @Produce json
// @Param limit query int false "返回数量限制"
// @Success 200 {object} response.SuccessResponse{data=[]service.PopularSearch}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/search/popular [get]
func (h *SearchHandler) GetPopularSearches(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequest(c, "limit参数格式错误")
		return
	}

	if limit < 1 || limit > 50 {
		response.BadRequest(c, "limit参数必须在1-50之间")
		return
	}

	ctx := c.Request.Context()

	popularSearches, err := h.searchService.GetPopularSearches(ctx, limit)
	if err != nil {
		h.logger.Error("获取热门搜索失败", zap.Error(err))
		response.InternalServerError(c, "获取热门搜索失败")
		return
	}

	response.Success(c, popularSearches)
}
