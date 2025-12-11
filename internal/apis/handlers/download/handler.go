package download

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	downloadbiz "moviepilot-go/internal/business/services/download"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/models/types"
	"moviepilot-go/pkg/logger"
)

// Handler 下载核心功能 API 处理器
type Handler struct {
	downloadService *downloadbiz.DownloadService
	logger          *zap.Logger
}

// NewHandler 创建下载处理器
func NewHandler(
	downloadService *downloadbiz.DownloadService,
) *Handler {
	return &Handler{
		downloadService: downloadService,
		logger:          logger.GetLogger(),
	}
}

// DownloadingTorrent 正在下载的任务信息
type DownloadingTorrent struct {
	Hash       string `json:"hash"`
	Title      string `json:"title"`
	Size       int64  `json:"size"`
	Status     string `json:"status"`
	Downloader string `json:"downloader"`
}

// GetDownloading 获取正在下载的任务
// @Summary 获取正在下载的任务
// @Description 查询正在下载的任务
// @Tags download
// @Produce json
// @Param name query string false "任务名称"
// @Success 200 {array} DownloadingTorrent
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/ [get]
func (h *Handler) GetDownloading(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	name := c.Query("name")

	h.logger.Debug("GetDownloading called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("name", name),
	)

	// 调用下载服务获取正在下载的任务
	downloads, err := h.downloadService.Downloading(c.Request.Context())
	if err != nil {
		h.logger.Error("获取正在下载任务失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get downloading tasks"})
		return
	}

	// 转换为响应格式
	var downloadingTorrents []DownloadingTorrent
	for _, download := range downloads {
		// 过滤名称
		if name != "" && download.Title != name {
			continue
		}

		downloadingTorrents = append(downloadingTorrents, DownloadingTorrent{
			Hash:       download.Hash,
			Title:      download.Title,
			Size:       download.Size,
			Status:     download.Status,
			Downloader: download.Downloader,
		})
	}

	c.JSON(http.StatusOK, downloadingTorrents)
}

// Download 添加下载（含媒体信息）
// @Summary 添加下载（含媒体信息）
// @Description 添加下载任务（含媒体信息）
// @Tags download
// @Accept json
// @Produce json
// @Param media_in body dto.MediaInfo true "媒体信息"
// @Param torrent_in body dto.TorrentInfo true "种子信息"
// @Param downloader query string false "下载器"
// @Param save_path query string false "保存路径"
// @Success 200 {object} dto.Response
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/ [post]
func (h *Handler) Download(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	// 解析请求体
	var req struct {
		MediaInfo   dto.MediaInfo   `json:"media_in" binding:"required"`
		TorrentInfo dto.TorrentInfo `json:"torrent_in" binding:"required"`
		Downloader  string          `json:"downloader"`
		SavePath    string          `json:"save_path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Download invalid request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建上下文
	context := &dto.Context{
		MediaInfo:   &req.MediaInfo,
		TorrentInfo: &req.TorrentInfo,
	}

	// 调用下载服务添加下载
	task, err := h.downloadService.AddDownload(c.Request.Context(), context, req.Downloader)
	if err != nil {
		h.logger.Error("添加下载失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("title", req.TorrentInfo.Title),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add download"})
		return
	}

	h.logger.Info("添加下载成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("title", req.TorrentInfo.Title),
		zap.String("download_id", task.DownloadID),
	)

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: map[string]any{
			"download_id": task.DownloadID,
		},
	})
}

// AddDownload 添加下载（不含媒体信息）
// @Summary 添加下载（不含媒体信息）
// @Description 添加下载任务（不含媒体信息）
// @Tags download
// @Accept json
// @Produce json
// @Param torrent_in body dto.TorrentInfo true "种子信息"
// @Param downloader query string false "下载器"
// @Param save_path query string false "保存路径"
// @Success 200 {object} dto.Response
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/add [post]
func (h *Handler) AddDownload(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	// 解析请求体
	var req struct {
		TorrentInfo dto.TorrentInfo `json:"torrent_in" binding:"required"`
		Downloader  string          `json:"downloader"`
		SavePath    string          `json:"save_path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("AddDownload invalid request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 调用媒体服务识别媒体信息
	// 暂时创建一个空的媒体信息
	mediaInfo := &dto.MediaInfo{
		Type: string(types.MediaTypeMovie), // 默认类型，实际应该由媒体服务识别
	}

	// 创建上下文
	context := &dto.Context{
		MediaInfo:   mediaInfo,
		TorrentInfo: &req.TorrentInfo,
	}

	// 调用下载服务添加下载
	task, err := h.downloadService.AddDownload(c.Request.Context(), context, req.Downloader)
	if err != nil {
		h.logger.Error("添加下载失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("title", req.TorrentInfo.Title),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add download"})
		return
	}

	h.logger.Info("添加下载成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("title", req.TorrentInfo.Title),
		zap.String("download_id", task.DownloadID),
	)

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data: map[string]any{
			"download_id": task.DownloadID,
		},
	})
}

// StartDownload 开始下载任务
// @Summary 开始下载任务
// @Description 开始指定的下载任务
// @Tags download
// @Produce json
// @Param hashString path string true "任务哈希值"
// @Param name query string false "任务名称"
// @Success 200 {object} dto.Response
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/start/{hashString} [get]
func (h *Handler) StartDownload(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	hashString := c.Param("hashString")
	name := c.Query("name")

	h.logger.Debug("StartDownload called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("hash_string", hashString),
		zap.String("name", name),
	)

	// 调用下载服务开始任务
	err := h.downloadService.SetDownloading(c.Request.Context(), hashString, "start")
	if err != nil {
		h.logger.Error("开始下载任务失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("hash_string", hashString),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start download"})
		return
	}

	h.logger.Info("开始下载任务成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("hash_string", hashString),
	)

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
	})
}

// StopDownload 暂停下载任务
// @Summary 暂停下载任务
// @Description 暂停指定的下载任务
// @Tags download
// @Produce json
// @Param hashString path string true "任务哈希值"
// @Param name query string false "任务名称"
// @Success 200 {object} dto.Response
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/stop/{hashString} [get]
func (h *Handler) StopDownload(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	hashString := c.Param("hashString")
	name := c.Query("name")

	h.logger.Debug("StopDownload called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("hash_string", hashString),
		zap.String("name", name),
	)

	// 调用下载服务暂停任务
	err := h.downloadService.SetDownloading(c.Request.Context(), hashString, "stop")
	if err != nil {
		h.logger.Error("暂停下载任务失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("hash_string", hashString),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop download"})
		return
	}

	h.logger.Info("暂停下载任务成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("hash_string", hashString),
	)

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
	})
}

// GetClients 获取可用下载器
// @Summary 获取可用下载器
// @Description 查询可用的下载器列表
// @Tags download
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/download/clients [get]
func (h *Handler) GetClients(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Debug("GetClients called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 从配置中获取可用下载器列表
	// 暂时返回模拟数据
	clients := []map[string]any{
		{
			"name": "qBittorrent",
			"type": "qbittorrent",
		},
		{
			"name": "Transmission",
			"type": "transmission",
		},
		{
			"name": "Aria2",
			"type": "aria2",
		},
	}

	h.logger.Info("获取可用下载器成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("client_count", len(clients)),
	)

	c.JSON(http.StatusOK, clients)
}

// DeleteDownload 删除下载任务
// @Summary 删除下载任务
// @Description 删除指定的下载任务
// @Tags download
// @Produce json
// @Param hashString path string true "任务哈希值"
// @Param name query string false "任务名称"
// @Success 200 {object} dto.Response
// @Failure 500 {object} map[string]interface{}
// @Router /api/download/{hashString} [delete]
func (h *Handler) DeleteDownload(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	hashString := c.Param("hashString")
	name := c.Query("name")

	h.logger.Debug("DeleteDownload called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("hash_string", hashString),
		zap.String("name", name),
	)

	// 调用下载服务删除任务
	err := h.downloadService.RemoveDownloading(c.Request.Context(), hashString, false)
	if err != nil {
		h.logger.Error("删除下载任务失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("hash_string", hashString),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete download"})
		return
	}

	h.logger.Info("删除下载任务成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("hash_string", hashString),
	)

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
	})
}
