package mediaserver

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	mediaserverbiz "moviepilot-go/internal/business/services/mediaserver"
	"moviepilot-go/pkg/logger"
)

// SyncHandler 媒体库同步 API 处理器
type SyncHandler struct {
	syncService mediaserverbiz.SyncService
	logger      *zap.Logger
}

// NewSyncHandler 创建同步处理器
func NewSyncHandler(syncService mediaserverbiz.SyncService) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		logger:      logger.GetLogger(),
	}
}

// SyncLibrary 同步媒体库
// @Summary 同步媒体库
// @Description 同步指定媒体服务器的媒体库
// @Tags mediaserver
// @Produce json
// @Param server_id path string true "服务器ID"
// @Param library_id path string true "媒体库ID"
// @Success 200 {object} mediaserverbiz.SyncResult
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/{server_id}/library/{library_id}/sync [post]
func (h *SyncHandler) SyncLibrary(c *gin.Context) {
	serverID := c.Param("server_id")
	libraryID := c.Param("library_id")

	result, err := h.syncService.SyncLibrary(c.Request.Context(), serverID, libraryID)
	if err != nil {
		h.logger.Error("同步媒体库失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SyncAll 同步所有媒体库
// @Summary 同步所有媒体库
// @Description 同步所有媒体服务器的所有媒体库
// @Tags mediaserver
// @Produce json
// @Success 200 {array} mediaserverbiz.SyncResult
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/sync/all [post]
func (h *SyncHandler) SyncAll(c *gin.Context) {
	results, err := h.syncService.SyncAll(c.Request.Context())
	if err != nil {
		h.logger.Error("同步所有媒体库失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetSyncHistory 获取同步历史
// @Summary 获取同步历史
// @Description 获取指定媒体服务器的同步历史记录
// @Tags mediaserver
// @Produce json
// @Param server_id path string true "服务器ID"
// @Param limit query int false "数量限制" default(50)
// @Success 200 {array} mediaserverbiz.SyncRecord
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/{server_id}/sync/history [get]
func (h *SyncHandler) GetSyncHistory(c *gin.Context) {
	serverID := c.Param("server_id")

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}

	history, err := h.syncService.GetSyncHistory(c.Request.Context(), serverID, limit)
	if err != nil {
		h.logger.Error("获取同步历史失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetSyncStats 获取同步统计
// @Summary 获取同步统计
// @Description 获取指定媒体服务器的同步统计信息
// @Tags mediaserver
// @Produce json
// @Param server_id path string true "服务器ID"
// @Success 200 {object} mediaserverbiz.SyncStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/{server_id}/sync/stats [get]
func (h *SyncHandler) GetSyncStats(c *gin.Context) {
	serverID := c.Param("server_id")

	stats, err := h.syncService.GetSyncStats(c.Request.Context(), serverID)
	if err != nil {
		h.logger.Error("获取同步统计失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
