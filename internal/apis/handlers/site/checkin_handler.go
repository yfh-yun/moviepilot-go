package site

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	siteservice "moviepilot-go/internal/business/services/site"
	"moviepilot-go/pkg/logger"
)

// CheckinHandler 站点签到 API 处理器
type CheckinHandler struct {
	checkinService siteservice.CheckinService
	logger         *zap.Logger
}

// NewCheckinHandler 创建签到处理器
func NewCheckinHandler(checkinService siteservice.CheckinService) *CheckinHandler {
	return &CheckinHandler{
		checkinService: checkinService,
		logger:         logger.GetLogger(),
	}
}

// Checkin 执行签到
// @Summary 执行站点签到
// @Description 对指定站点执行签到操作
// @Tags site
// @Security BearerAuth
// @Produce json
// @Param site_id path int true "站点ID"
// @Success 200 {object} siteservice.CheckinResult
// @Failure 400 {object} map[string]interface{}
// @Router /api/sites/{site_id}/checkin [post]
func (h *CheckinHandler) Checkin(c *gin.Context) {
	siteIDStr := c.Param("site_id")
	siteID, err := strconv.ParseUint(siteIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	result, err := h.checkinService.Checkin(c.Request.Context(), uint(siteID))
	if err != nil {
		h.logger.Error("签到失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// BatchCheckin 批量签到
// @Summary 批量站点签到
// @Description 对多个站点执行批量签到操作
// @Tags site
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param site_ids body []uint true "站点ID列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/sites/checkin/batch [post]
func (h *CheckinHandler) BatchCheckin(c *gin.Context) {
	var req struct {
		SiteIDs []uint `json:"site_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 简化实现：逐个签到
	results := []any{}
	for _, siteID := range req.SiteIDs {
		result, err := h.checkinService.Checkin(c.Request.Context(), siteID)
		if err != nil {
			h.logger.Error("签到失败", zap.Uint("site_id", siteID), zap.Error(err))
			results = append(results, map[string]any{
				"site_id": siteID,
				"success": false,
				"message": err.Error(),
			})
		} else {
			results = append(results, map[string]any{
				"site_id": siteID,
				"success": true,
				"data":    result,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// GetCheckinHistory 获取签到历史
// @Summary 获取签到历史
// @Description 获取指定站点的签到历史记录
// @Tags site
// @Security BearerAuth
// @Produce json
// @Param site_id path int true "站点ID"
// @Param days query int false "天数" default(30)
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/sites/{site_id}/checkin/history [get]
func (h *CheckinHandler) GetCheckinHistory(c *gin.Context) {
	siteIDStr := c.Param("site_id")
	siteID, err := strconv.ParseUint(siteIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	// 使用 GetCheckinLogs 方法代替 GetCheckinHistory
	logs, total, err := h.checkinService.GetCheckinLogs(c.Request.Context(), uint(siteID), 1, 100)
	if err != nil {
		h.logger.Error("获取签到历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   total,
		"data":    logs,
	})
}

// GetCheckinStats 获取签到统计
// @Summary 获取签到统计
// @Description 获取指定站点的签到统计信息
// @Tags site
// @Security BearerAuth
// @Produce json
// @Param site_id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/sites/{site_id}/checkin/stats [get]
func (h *CheckinHandler) GetCheckinStats(c *gin.Context) {
	siteIDStr := c.Param("site_id")
	siteID, err := strconv.ParseUint(siteIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	// 简化实现：获取最近100条记录并统计
	logs, _, err := h.checkinService.GetCheckinLogs(c.Request.Context(), uint(siteID), 1, 100)
	if err != nil {
		h.logger.Error("获取签到统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 简单统计
	total := len(logs)
	success := 0
	for _, log := range logs {
		if log.Success {
			success++
		}
	}

	stats := map[string]any{
		"total":        total,
		"success":      success,
		"success_rate": float64(success) / float64(total) * 100,
	}

	c.JSON(http.StatusOK, stats)
}

// GetAllCheckinHistory 获取所有站点签到历史
// @Summary 获取所有站点签到历史
// @Description 获取所有站点的签到历史记录
// @Tags site
// @Security BearerAuth
// @Produce json
// @Param days query int false "天数" default(7)
// @Success 200 {object} map[string]interface{}
// @Router /api/sites/checkin/history [get]
func (h *CheckinHandler) GetAllCheckinHistory(c *gin.Context) {
	_ = c.DefaultQuery("days", "7") // days 参数预留，待实现时使用

	// TODO: 获取所有站点ID
	// 这里需要从站点服务获取所有站点列表
	// 然后逐个获取签到历史

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "功能待实现",
		"data":    []any{},
	})
}

// ScheduleCheckin 调度自动签到
// @Summary 调度自动签到
// @Description 触发自动签到调度任务
// @Tags site
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/sites/checkin/schedule [post]
func (h *CheckinHandler) ScheduleCheckin(c *gin.Context) {
	// 简化实现：调用 CheckinAll 方法签到所有站点
	if err := h.checkinService.CheckinAll(c.Request.Context()); err != nil {
		h.logger.Error("调度签到失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "签到任务已触发",
	})
}
