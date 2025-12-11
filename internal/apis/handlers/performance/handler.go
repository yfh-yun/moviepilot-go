package performance

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	performancebiz "moviepilot-go/internal/business/services/performance"
	"moviepilot-go/pkg/logger"
)

// Handler 性能监控 API 处理器
type Handler struct {
	monitorService performancebiz.MonitorService
	logger         *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(monitorService performancebiz.MonitorService) *Handler {
	return &Handler{
		monitorService: monitorService,
		logger:         logger.GetLogger(),
	}
}

// GetMetrics 获取性能指标
// @Summary 获取性能指标
// @Description 获取当前系统性能指标
// @Tags performance
// @Produce json
// @Success 200 {object} performancebiz.Metrics
// @Failure 500 {object} map[string]interface{}
// @Router /api/performance/metrics [get]
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics, err := h.monitorService.GetMetrics(c.Request.Context())
	if err != nil {
		h.logger.Error("获取性能指标失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetHistory 获取历史数据
// @Summary 获取历史数据
// @Description 获取指定时间范围内的性能历史数据
// @Tags performance
// @Produce json
// @Param duration query string false "时间范围" default("1h")
// @Success 200 {array} performancebiz.MetricsSnapshot
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/performance/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	durationStr := c.DefaultQuery("duration", "1h")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的时间范围"})
		return
	}

	history, err := h.monitorService.GetHistory(c.Request.Context(), duration)
	if err != nil {
		h.logger.Error("获取历史数据失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// StartMonitoring 开始监控
// @Summary 开始监控
// @Description 启动性能监控
// @Tags performance
// @Produce json
// @Param interval query string false "监控间隔" default("10s")
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/performance/monitor/start [post]
func (h *Handler) StartMonitoring(c *gin.Context) {
	intervalStr := c.DefaultQuery("interval", "10s")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的监控间隔"})
		return
	}

	h.monitorService.StartMonitoring(c.Request.Context(), interval)

	c.JSON(http.StatusOK, gin.H{"message": "监控已启动"})
}

// StopMonitoring 停止监控
// @Summary 停止监控
// @Description 停止性能监控
// @Tags performance
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/performance/monitor/stop [post]
func (h *Handler) StopMonitoring(c *gin.Context) {
	h.monitorService.StopMonitoring()

	c.JSON(http.StatusOK, gin.H{"message": "监控已停止"})
}
