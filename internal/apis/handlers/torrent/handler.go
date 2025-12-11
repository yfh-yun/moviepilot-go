package torrent

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	torrentservice "moviepilot-go/internal/business/services/torrent"
	"moviepilot-go/pkg/logger"
)

// Handler 种子API处理器
type Handler struct {
	torrentService torrentservice.Service
	logger         *zap.Logger
}

// NewHandler 创建种子API处理器
func NewHandler(torrentService torrentservice.Service) *Handler {
	return &Handler{
		torrentService: torrentService,
		logger:         logger.GetLogger(),
	}
}

// GetTorrentsCache 获取种子缓存
// @Summary 获取种子缓存
// @Description 获取当前种子缓存数据
// @Tags torrent
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/torrent/cache [get]
func (h *Handler) GetTorrentsCache(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("获取种子缓存",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 获取订阅模式（从配置中获取，目前默认使用rss）
	// TODO: 从配置中获取订阅模式
	subscribeMode := "rss"

	// 获取格式化的种子缓存
	response, err := h.torrentService.GetFormattedCache(c.Request.Context(), subscribeMode)
	if err != nil {
		h.logger.Error("获取种子缓存失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("获取种子缓存成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("count", response.Count),
		zap.Int("sites", response.Sites),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// DeleteTorrentCache 删除指定种子缓存
// @Summary 删除指定种子缓存
// @Description 删除指定站点的指定种子缓存
// @Tags torrent
// @Produce json
// @Param domain path string true "站点域名"
// @Param torrent_hash path string true "种子哈希值"
// @Success 200 {object} map[string]interface{}
// @Router /api/torrent/cache/{domain}/{torrent_hash} [delete]
func (h *Handler) DeleteTorrentCache(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	domain := c.Param("domain")
	torrentHash := c.Param("torrent_hash")

	h.logger.Info("删除指定种子缓存",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("domain", domain),
		zap.String("torrent_hash", torrentHash),
	)

	// 删除指定种子缓存
	err := h.torrentService.DeleteTorrentCache(c.Request.Context(), domain, torrentHash)
	if err != nil {
		h.logger.Error("删除指定种子缓存失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("domain", domain),
			zap.String("torrent_hash", torrentHash),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("删除指定种子缓存成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("domain", domain),
		zap.String("torrent_hash", torrentHash),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "种子删除成功",
	})
}

// ClearTorrentsCache 清理种子缓存
// @Summary 清理种子缓存
// @Description 清理所有种子缓存
// @Tags torrent
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/torrent/cache [delete]
func (h *Handler) ClearTorrentsCache(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("清理种子缓存",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 清理所有种子缓存
	err := h.torrentService.ClearTorrentsCache(c.Request.Context())
	if err != nil {
		h.logger.Error("清理种子缓存失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("清理种子缓存成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "种子缓存清理完成",
	})
}

// RefreshTorrentsCache 刷新种子缓存
// @Summary 刷新种子缓存
// @Description 刷新种子缓存
// @Tags torrent
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/torrent/cache/refresh [post]
func (h *Handler) RefreshTorrentsCache(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("刷新种子缓存",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 刷新种子缓存
	torrentsMap, err := h.torrentService.RefreshTorrentsCache(c.Request.Context())
	if err != nil {
		h.logger.Error("刷新种子缓存失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 统计刷新结果
	totalCount := 0
	for _, torrents := range torrentsMap {
		totalCount += len(torrents)
	}
	sitesCount := len(torrentsMap)

	h.logger.Info("刷新种子缓存成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("sites_count", sitesCount),
		zap.Int("total_count", totalCount),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("缓存刷新完成，共刷新 %d 个站点，%d 个种子", sitesCount, totalCount),
	})
}

// ReidentifyTorrent 重新识别种子
// @Summary 重新识别种子
// @Description 重新识别指定的种子
// @Tags torrent
// @Produce json
// @Param domain path string true "站点域名"
// @Param torrent_hash path string true "种子哈希值"
// @Param tmdbid query int false "TMDB ID"
// @Param doubanid query string false "豆瓣ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/torrent/cache/reidentify/{domain}/{torrent_hash} [post]
func (h *Handler) ReidentifyTorrent(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	domain := c.Param("domain")
	torrentHash := c.Param("torrent_hash")

	// 获取可选参数
	var tmdbID *int
	var doubanID *string

	tmdbIDStr := c.Query("tmdbid")
	if tmdbIDStr != "" {
		id, err := strconv.Atoi(tmdbIDStr)
		if err == nil {
			tmdbID = &id
		}
	}

	doubanIDStr := c.Query("doubanid")
	if doubanIDStr != "" {
		doubanID = &doubanIDStr
	}

	h.logger.Info("重新识别种子",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("domain", domain),
		zap.String("torrent_hash", torrentHash),
		zap.Intp("tmdb_id", tmdbID),
		zap.Stringp("douban_id", doubanID),
	)

	// 重新识别种子
	cacheItem, err := h.torrentService.ReidentifyTorrent(c.Request.Context(), domain, torrentHash, tmdbID, doubanID)
	if err != nil {
		h.logger.Error("重新识别种子失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("domain", domain),
			zap.String("torrent_hash", torrentHash),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("重新识别种子成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("domain", domain),
		zap.String("torrent_hash", torrentHash),
		zap.String("media_name", cacheItem.MediaName),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "重新识别完成",
		"data": map[string]any{
			"media_name": cacheItem.MediaName,
			"media_year": cacheItem.MediaYear,
			"media_type": cacheItem.MediaType,
		},
	})
}
