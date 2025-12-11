package search

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	searchbiz "moviepilot-go/internal/business/services/search"
	"moviepilot-go/pkg/logger"
)

// EnhancedHandler 搜索增强功能 API 处理器
type EnhancedHandler struct {
	historyService searchbiz.HistoryService
	cacheService   searchbiz.CacheService
	ranker         searchbiz.Ranker
	logger         *zap.Logger
}

// NewEnhancedHandler 创建增强处理器
func NewEnhancedHandler(
	historyService searchbiz.HistoryService,
	cacheService searchbiz.CacheService,
	ranker searchbiz.Ranker,
) *EnhancedHandler {
	return &EnhancedHandler{
		historyService: historyService,
		cacheService:   cacheService,
		ranker:         ranker,
		logger:         logger.GetLogger(),
	}
}

// GetSearchHistory 获取搜索历史
// @Summary 获取搜索历史
// @Description 获取用户的搜索历史记录
// @Tags search-enhanced
// @Produce json
// @Param user_id query string true "用户ID"
// @Param limit query int false "数量限制" default(50)
// @Success 200 {array} searchbiz.SearchRecord
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/history [get]
func (h *EnhancedHandler) GetSearchHistory(c *gin.Context) {
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

	history, err := h.historyService.GetSearchHistory(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("获取搜索历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetPopularSearches 获取热门搜索
// @Summary 获取热门搜索
// @Description 获取最近一段时间的热门搜索关键词
// @Tags search-enhanced
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Param days query int false "天数" default(7)
// @Success 200 {array} searchbiz.PopularSearch
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/popular [get]
func (h *EnhancedHandler) GetPopularSearches(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	daysStr := c.DefaultQuery("days", "7")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	popular, err := h.historyService.GetPopularSearches(c.Request.Context(), limit, days)
	if err != nil {
		h.logger.Error("获取热门搜索失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, popular)
}

// GetSearchStats 获取搜索统计
// @Summary 获取搜索统计
// @Description 获取用户的搜索统计信息
// @Tags search-enhanced
// @Produce json
// @Param user_id query string true "用户ID"
// @Success 200 {object} searchbiz.SearchStats
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/stats [get]
func (h *EnhancedHandler) GetSearchStats(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	stats, err := h.historyService.GetSearchStats(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("获取搜索统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RecordSearch 记录搜索
// @Summary 记录搜索
// @Description 记录一次搜索操作
// @Tags search-enhanced
// @Accept json
// @Produce json
// @Param record body searchbiz.SearchRecord true "搜索记录"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/record [post]
func (h *EnhancedHandler) RecordSearch(c *gin.Context) {
	var record searchbiz.SearchRecord

	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.historyService.RecordSearch(c.Request.Context(), &record); err != nil {
		h.logger.Error("记录搜索失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "记录成功"})
}

// ClearCache 清空搜索缓存
// @Summary 清空搜索缓存
// @Description 清空所有搜索缓存
// @Tags search-enhanced
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/cache/clear [post]
func (h *EnhancedHandler) ClearCache(c *gin.Context) {
	if err := h.cacheService.Clear(c.Request.Context()); err != nil {
		h.logger.Error("清空缓存失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "缓存已清空"})
}

// GetCacheStats 获取缓存统计
// @Summary 获取缓存统计
// @Description 获取搜索缓存的统计信息
// @Tags search-enhanced
// @Produce json
// @Success 200 {object} searchbiz.CacheStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/cache/stats [get]
func (h *EnhancedHandler) GetCacheStats(c *gin.Context) {
	stats, err := h.cacheService.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("获取缓存统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RankResults 对搜索结果排序
// @Summary 对搜索结果排序
// @Description 使用智能算法对搜索结果进行排序
// @Tags search-enhanced
// @Accept json
// @Produce json
// @Param request body RankRequest true "排序请求"
// @Success 200 {array} searchbiz.RankedResult
// @Failure 400 {object} map[string]interface{}
// @Router /api/search/rank [post]
func (h *EnhancedHandler) RankResults(c *gin.Context) {
	var req RankRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ranked := h.ranker.Rank(req.Results, req.Query)

	c.JSON(http.StatusOK, ranked)
}

// RankRequest 排序请求
type RankRequest struct {
	Results []*searchbiz.SearchResult `json:"results" binding:"required"`
	Query   *searchbiz.SearchQuery    `json:"query" binding:"required"`
}

// CleanOldHistory 清理旧历史
// @Summary 清理旧历史
// @Description 清理指定天数之前的搜索历史
// @Tags search-enhanced
// @Produce json
// @Param days query int false "天数" default(90)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/history/clean [post]
func (h *EnhancedHandler) CleanOldHistory(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "90")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 90
	}

	if err := h.historyService.CleanOldRecords(c.Request.Context(), days); err != nil {
		h.logger.Error("清理旧历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "清理完成"})
}

// SearchWithCache 带缓存的搜索
// @Summary 带缓存的搜索
// @Description 执行搜索并使用缓存优化
// @Tags search-enhanced
// @Accept json
// @Produce json
// @Param query body searchbiz.SearchQuery true "搜索查询"
// @Success 200 {object} SearchResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search/cached [post]
func (h *EnhancedHandler) SearchWithCache(c *gin.Context) {
	var query searchbiz.SearchQuery

	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// 生成缓存键
	cacheKey := h.cacheService.GenerateKey(&query)

	// 尝试从缓存获取
	results, err := h.cacheService.Get(c.Request.Context(), cacheKey)
	fromCache := false

	if err != nil {
		h.logger.Warn("获取缓存失败", zap.Error(err))
	} else if results != nil {
		fromCache = true
		h.logger.Info("使用缓存结果", zap.String("key", cacheKey))
	} else {
		// TODO: 这里应该调用实际的搜索服务
		// results, err = h.searchService.Search(c.Request.Context(), &query)
		results = []*searchbiz.SearchResult{} // 临时返回空结果
	}

	duration := time.Since(startTime).Milliseconds()

	// 对结果排序
	ranked := h.ranker.Rank(results, &query)

	response := SearchResponse{
		Results:   ranked,
		Total:     len(ranked),
		Duration:  duration,
		FromCache: fromCache,
		CacheKey:  cacheKey,
	}

	c.JSON(http.StatusOK, response)
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results   []*searchbiz.RankedResult `json:"results"`
	Total     int                       `json:"total"`
	Duration  int64                     `json:"duration"`
	FromCache bool                      `json:"from_cache"`
	CacheKey  string                    `json:"cache_key"`
}
