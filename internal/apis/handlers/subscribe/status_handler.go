package subscribe

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// StatusHandler 订阅状态管理处理器
type StatusHandler struct {
	logger *zap.Logger
}

// NewStatusHandler 创建状态处理器
func NewStatusHandler() *StatusHandler {
	return &StatusHandler{
		logger: logger.GetLogger(),
	}
}

// UpdateSubscribeStatus 更新订阅状态
// @Summary 更新订阅状态
// @Description 更新订阅的启用/禁用状态
// @Tags subscribe
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param sub_id path int true "订阅ID"
// @Param status body object true "状态信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/subscribes/status/{sub_id} [put]
func (h *StatusHandler) UpdateSubscribeStatus(c *gin.Context) {
	subIDStr := c.Param("sub_id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("更新订阅状态",
		zap.Int("sub_id", subID),
		zap.Bool("enabled", req.Enabled))

	// TODO: 实现状态更新逻辑
	// 1. 查询订阅
	// 2. 更新状态
	// 3. 保存到数据库

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订阅状态已更新",
	})
}

// GetSubscribeHistory 获取订阅历史
// @Summary 获取订阅历史
// @Description 获取指定类型的订阅历史记录
// @Tags subscribe
// @Security BearerAuth
// @Produce json
// @Param mtype path string true "媒体类型" Enums(movie, tv)
// @Success 200 {array} map[string]interface{}
// @Router /api/subscribes/history/{mtype} [get]
func (h *StatusHandler) GetSubscribeHistory(c *gin.Context) {
	mtype := c.Param("mtype")

	h.logger.Info("获取订阅历史", zap.String("mtype", mtype))

	// TODO: 实现历史查询逻辑
	// 1. 根据媒体类型查询
	// 2. 返回历史记录

	history := []map[string]any{
		{
			"id":          1,
			"name":        "示例订阅",
			"type":        mtype,
			"status":      "completed",
			"create_time": "2025-11-23T10:00:00Z",
		},
	}

	c.JSON(http.StatusOK, history)
}

// GetMediaSubscribe 根据媒体ID获取订阅
// @Summary 根据媒体ID获取订阅
// @Description 获取指定媒体的订阅信息
// @Tags subscribe
// @Security BearerAuth
// @Produce json
// @Param media_id path string true "媒体ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/subscribes/media/{media_id} [get]
func (h *StatusHandler) GetMediaSubscribe(c *gin.Context) {
	mediaID := c.Param("media_id")

	h.logger.Info("获取媒体订阅", zap.String("media_id", mediaID))

	// TODO: 实现查询逻辑
	subscribe := gin.H{
		"id":       1,
		"media_id": mediaID,
		"name":     "示例订阅",
		"status":   "active",
	}

	c.JSON(http.StatusOK, subscribe)
}

// DeleteMediaSubscribe 根据媒体ID删除订阅
// @Summary 根据媒体ID删除订阅
// @Description 删除指定媒体的订阅
// @Tags subscribe
// @Security BearerAuth
// @Produce json
// @Param media_id path string true "媒体ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/subscribes/media/{media_id} [delete]
func (h *StatusHandler) DeleteMediaSubscribe(c *gin.Context) {
	mediaID := c.Param("media_id")

	h.logger.Info("删除媒体订阅", zap.String("media_id", mediaID))

	// TODO: 实现删除逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订阅已删除",
	})
}

// OverSeerrNotify OverSeerr/JellySeerr通知订阅
// @Summary OverSeerr通知
// @Description 接收OverSeerr/JellySeerr的订阅通知
// @Tags subscribe
// @Accept json
// @Produce json
// @Param notification body map[string]interface{} true "通知信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/subscribes/seerr [post]
func (h *StatusHandler) OverSeerrNotify(c *gin.Context) {
	var notification map[string]any
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("收到OverSeerr通知", zap.Any("notification", notification))

	// TODO: 实现OverSeerr通知处理逻辑
	// 1. 解析通知内容
	// 2. 创建订阅
	// 3. 触发搜索

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "通知已处理",
	})
}
