package dashboard

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	dashboardbiz "moviepilot-go/internal/business/services/dashboard"
	"moviepilot-go/pkg/logger"
)

// EnhancedHandler Dashboard 增强 API 处理器
type EnhancedHandler struct {
	dashboardService dashboardbiz.Service
	logger           *zap.Logger
}

// NewEnhancedHandler 创建增强处理器
func NewEnhancedHandler(dashboardService dashboardbiz.Service) *EnhancedHandler {
	return &EnhancedHandler{
		dashboardService: dashboardService,
		logger:           logger.GetLogger(),
	}
}

// GetDashboardData 获取仪表板数据
// @Summary 获取仪表板数据
// @Description 获取完整的仪表板数据，包括统计、活动、图表等
// @Tags dashboard
// @Produce json
// @Success 200 {object} dashboardbiz.DashboardData
// @Failure 500 {object} map[string]interface{}
// @Router /api/dashboard [get]
func (h *EnhancedHandler) GetDashboardData(c *gin.Context) {
	data, err := h.dashboardService.GetDashboardData(c.Request.Context())
	if err != nil {
		h.logger.Error("获取仪表板数据失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetStatistics 获取统计信息
// @Summary 获取统计信息
// @Description 获取系统统计信息
// @Tags dashboard
// @Produce json
// @Success 200 {object} dashboardbiz.Statistics
// @Failure 500 {object} map[string]interface{}
// @Router /api/dashboard/statistics [get]
func (h *EnhancedHandler) GetStatistics(c *gin.Context) {
	stats, err := h.dashboardService.GetStatistics(c.Request.Context())
	if err != nil {
		h.logger.Error("获取统计信息失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentActivity 获取最近活动
// @Summary 获取最近活动
// @Description 获取系统最近的活动记录
// @Tags dashboard
// @Produce json
// @Param limit query int false "数量限制" default(20)
// @Success 200 {array} dashboardbiz.Activity
// @Failure 500 {object} map[string]interface{}
// @Router /api/dashboard/activity [get]
func (h *EnhancedHandler) GetRecentActivity(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	activities, err := h.dashboardService.GetRecentActivity(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("获取最近活动失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, activities)
}

// GetChartData 获取图表数据
// @Summary 获取图表数据
// @Description 获取指定类型的图表数据
// @Tags dashboard
// @Produce json
// @Param type query string true "图表类型" Enums(downloads, subscribes, storage)
// @Param days query int false "天数" default(7)
// @Success 200 {object} dashboardbiz.ChartData
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/dashboard/chart [get]
func (h *EnhancedHandler) GetChartData(c *gin.Context) {
	chartType := c.Query("type")
	if chartType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图表类型不能为空"})
		return
	}

	daysStr := c.DefaultQuery("days", "7")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	chart, err := h.dashboardService.GetChartData(c.Request.Context(), chartType, days)
	if err != nil {
		h.logger.Error("获取图表数据失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chart)
}
