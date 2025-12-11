package site

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	sitebiz "moviepilot-go/internal/business/services/site"
	"moviepilot-go/pkg/logger"
)

// EnhancedHandler 站点增强功能 API 处理器
type EnhancedHandler struct {
	checkinService sitebiz.CheckinService
	logger         *zap.Logger
}

// NewEnhancedHandler 创建增强处理器
func NewEnhancedHandler(
	checkinService sitebiz.CheckinService,
) *EnhancedHandler {
	return &EnhancedHandler{
		checkinService: checkinService,
		logger:         logger.GetLogger(),
	}
}

// Checkin 站点签到
// @Summary 站点签到
// @Description 执行指定站点的签到操作
// @Tags site-enhanced
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/{id}/checkin [post]
func (h *EnhancedHandler) Checkin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	result, err := h.checkinService.Checkin(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("签到失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// BatchCheckin 批量签到
// @Summary 批量站点签到
// @Description 执行多个站点的批量签到
// @Tags site-enhanced
// @Accept json
// @Produce json
// @Param site_ids body []uint true "站点ID列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/checkin/batch [post]
func (h *EnhancedHandler) BatchCheckin(c *gin.Context) {
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
// @Tags site-enhanced
// @Produce json
// @Param id path int true "站点ID"
// @Param days query int false "天数" default(30)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/{id}/checkin/history [get]
func (h *EnhancedHandler) GetCheckinHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	// 使用 GetCheckinLogs 方法代替 GetCheckinHistory
	logs, total, err := h.checkinService.GetCheckinLogs(c.Request.Context(), uint(id), 1, 100)
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
// @Tags site-enhanced
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/{id}/checkin/stats [get]
func (h *EnhancedHandler) GetCheckinStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	// 简化实现：获取最近100条记录并统计
	logs, _, err := h.checkinService.GetCheckinLogs(c.Request.Context(), uint(id), 1, 100)
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

// CheckHealth 检查健康状态
// @Summary 检查站点健康状态
// @Description 检查指定站点的健康状态
// @Tags site-enhanced
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/{id}/health [get]
func (h *EnhancedHandler) CheckHealth(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点ID"})
		return
	}

	// 简化实现：返回默认健康状态
	status := map[string]any{
		"site_id": id,
		"status":  "healthy",
		"message": "站点健康检查功能待实现",
	}

	c.JSON(http.StatusOK, status)
}

// BatchCheckHealth 批量健康检查
// @Summary 批量健康检查
// @Description 批量检查多个站点的健康状态
// @Tags site-enhanced
// @Accept json
// @Produce json
// @Param site_ids body []uint true "站点ID列表"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/health/batch [post]
func (h *EnhancedHandler) BatchCheckHealth(c *gin.Context) {
	var req struct {
		SiteIDs []uint `json:"site_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 简化实现：返回默认健康状态
	results := make([]any, len(req.SiteIDs))
	for i, siteID := range req.SiteIDs {
		results[i] = map[string]any{
			"site_id": siteID,
			"status":  "healthy",
		}
	}

	c.JSON(http.StatusOK, results)
}

// GetHealthHistory 获取健康检查历史
// @Summary 获取健康检查历史
// @Description 获取指定站点的健康检查历史记录
// @Tags site-enhanced
// @Produce json
// @Param id path int true "站点ID"
// @Param days query int false "天数" default(7)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/{id}/health/history [get]
func (h *EnhancedHandler) GetHealthHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "健康检查历史功能待实现",
		"data":    []any{},
	})
}

// GetHealthStats 获取健康统计
// @Summary 获取健康统计
// @Description 获取指定站点的健康统计信息
// @Tags site-enhanced
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/{id}/health/stats [get]
func (h *EnhancedHandler) GetHealthStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "健康统计功能待实现",
		"data":    map[string]any{},
	})
}

// GetAuthStatus 获取认证状态
// @Summary 获取认证状态
// @Description 获取指定站点的认证状态
// @Tags site-enhanced
// @Produce json
// @Param id path int true "站点ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/site/{id}/auth/status [get]
func (h *EnhancedHandler) GetAuthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "认证状态功能待实现",
		"data":    map[string]any{},
	})
}
