package subscribe

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	subscribebiz "moviepilot-go/internal/business/services/subscribe"
	"moviepilot-go/pkg/logger"
)

// AnalyticsHandler 订阅分析 API 处理器
type AnalyticsHandler struct {
	analyticsService subscribebiz.AnalyticsService
	historyService   subscribebiz.HistoryService
	logger           *zap.Logger
}

// NewAnalyticsHandler 创建分析处理器
func NewAnalyticsHandler(
	analyticsService subscribebiz.AnalyticsService,
	historyService subscribebiz.HistoryService,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		historyService:   historyService,
		logger:           logger.GetLogger(),
	}
}

// GetOverallStats 获取总体统计
// @Summary 获取订阅总体统计
// @Description 获取所有订阅的统计信息
// @Tags subscribe-analytics
// @Produce json
// @Success 200 {object} subscribebiz.OverallStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribe/stats/overall [get]
func (h *AnalyticsHandler) GetOverallStats(c *gin.Context) {
	stats, err := h.analyticsService.GetOverallStats(c.Request.Context())
	if err != nil {
		h.logger.Error("获取总体统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSubscribeStats 获取订阅统计
// @Summary 获取单个订阅统计
// @Description 获取指定订阅的详细统计信息
// @Tags subscribe-analytics
// @Produce json
// @Param id path int true "订阅ID"
// @Success 200 {object} subscribebiz.SubscribeStats
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribe/stats/{id} [get]
func (h *AnalyticsHandler) GetSubscribeStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	stats, err := h.analyticsService.GetSubscribeStats(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("获取订阅统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTrendData 获取趋势数据
// @Summary 获取订阅趋势数据
// @Description 获取指定天数的订阅趋势数据
// @Tags subscribe-analytics
// @Produce json
// @Param days query int false "天数" default(30)
// @Success 200 {array} subscribebiz.TrendPoint
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribe/stats/trend [get]
func (h *AnalyticsHandler) GetTrendData(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90 // 最多90天
	}

	trends, err := h.analyticsService.GetTrendData(c.Request.Context(), days)
	if err != nil {
		h.logger.Error("获取趋势数据失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// GetTopSubscribes 获取热门订阅
// @Summary 获取热门订阅
// @Description 获取下载量最多的订阅列表
// @Tags subscribe-analytics
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Success 200 {array} subscribebiz.TopSubscribe
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribe/stats/top [get]
func (h *AnalyticsHandler) GetTopSubscribes(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // 最多100条
	}

	tops, err := h.analyticsService.GetTopSubscribes(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("获取热门订阅失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tops)
}

// GetMatchHistory 获取匹配历史
// @Summary 获取订阅匹配历史
// @Description 获取指定订阅的匹配历史记录
// @Tags subscribe-analytics
// @Produce json
// @Param id path int true "订阅ID"
// @Param limit query int false "数量限制" default(50)
// @Success 200 {array} subscribebiz.MatchRecord
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribe/{id}/history [get]
func (h *AnalyticsHandler) GetMatchHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	history, err := h.historyService.GetMatchHistory(c.Request.Context(), uint(id), limit)
	if err != nil {
		h.logger.Error("获取匹配历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetMatchStats 获取匹配统计
// @Summary 获取订阅匹配统计
// @Description 获取指定订阅的匹配统计信息
// @Tags subscribe-analytics
// @Produce json
// @Param id path int true "订阅ID"
// @Success 200 {object} subscribebiz.MatchStats
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscribe/{id}/match-stats [get]
func (h *AnalyticsHandler) GetMatchStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	stats, err := h.historyService.GetMatchStats(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("获取匹配统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
