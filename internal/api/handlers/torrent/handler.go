// Package torrent Torrent API处理器模块
package torrent

import (
	"net/http"
	"strconv"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/service/torrent"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler Torrent API处理器
type Handler struct {
	service torrent.Service
	logger  *logger.Logger
}

// NewHandler 创建新的Torrent处理器
func NewHandler(service torrent.Service, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	torrentGroup := router.Group("/torrent")
	{
		torrentGroup.GET("/", h.GetTorrentList)
		torrentGroup.GET("/:id", h.GetTorrent)
		torrentGroup.POST("/add", h.AddTorrent)
		torrentGroup.POST("/:id/pause", h.PauseTorrent)
		torrentGroup.POST("/:id/resume", h.ResumeTorrent)
		torrentGroup.DELETE("/:id", h.DeleteTorrent)
		torrentGroup.GET("/:id/files", h.GetTorrentFiles)
		torrentGroup.POST("/:id/files/priority", h.SetFilePriority)
	}
}

// GetTorrentList 获取种子列表
// @Summary 获取种子列表
// @Description 获取当前活动的种子列表
// @Tags 种子
// @Produce json
// @Param status query string false "状态过滤 (downloading,completed,paused)"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]TorrentInfo}
// @Router /torrent [get]
func (h *Handler) GetTorrentList(c *gin.Context) {
	status := c.Query("status")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := torrent.ListParams{
		Status: status,
		Page:   page,
		Count:  count,
	}

	torrents, err := h.service.GetTorrentList(params)
	if err != nil {
		h.logger.Error("Failed to get torrent list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取种子列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    torrents,
	})
}

// GetTorrent 获取种子详情
// @Summary 获取种子详情
// @Description 获取种子详细信息
// @Tags 种子
// @Produce json
// @Param id path string true "种子ID"
// @Success 200 {object} Response{data=TorrentDetail}
// @Router /torrent/{id} [get]
func (h *Handler) GetTorrent(c *gin.Context) {
	torrentID := c.Param("id")
	if torrentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "种子ID不能为空",
		})
		return
	}

	torrentDetail, err := h.service.GetTorrent(torrentID)
	if err != nil {
		h.logger.Error("Failed to get torrent", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取种子详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    torrentDetail,
	})
}

// AddTorrent 添加种子
// @Summary 添加种子
// @Description 添加新的种子到下载器
// @Tags 种子
// @Produce json
// @Param torrent body AddTorrentRequest true "种子信息"
// @Success 200 {object} Response{data=TorrentInfo}
// @Router /torrent/add [post]
func (h *Handler) AddTorrent(c *gin.Context) {
	var request AddTorrentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	torrentInfo, err := h.service.AddTorrent(request)
	if err != nil {
		h.logger.Error("Failed to add torrent", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "添加种子失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    torrentInfo,
		"message": "种子添加成功",
	})
}

// PauseTorrent 暂停种子
// @Summary 暂停种子
// @Description 暂停指定的种子下载
// @Tags 种子
// @Produce json
// @Param id path string true "种子ID"
// @Success 200 {object} Response
// @Router /torrent/{id}/pause [post]
func (h *Handler) PauseTorrent(c *gin.Context) {
	torrentID := c.Param("id")
	if torrentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "种子ID不能为空",
		})
		return
	}

	err := h.service.PauseTorrent(torrentID)
	if err != nil {
		h.logger.Error("Failed to pause torrent", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "暂停种子失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "种子已暂停",
	})
}

// ResumeTorrent 恢复种子
// @Summary 恢复种子
// @Description 恢复暂停的种子下载
// @Tags 种子
// @Produce json
// @Param id path string true "种子ID"
// @Success 200 {object} Response
// @Router /torrent/{id}/resume [post]
func (h *Handler) ResumeTorrent(c *gin.Context) {
	torrentID := c.Param("id")
	if torrentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "种子ID不能为空",
		})
		return
	}

	err := h.service.ResumeTorrent(torrentID)
	if err != nil {
		h.logger.Error("Failed to resume torrent", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "恢复种子失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "种子已恢复",
	})
}

// DeleteTorrent 删除种子
// @Summary 删除种子
// @Description 删除种子及其下载文件
// @Tags 种子
// @Produce json
// @Param id path string true "种子ID"
// @Param delete_files query bool false "是否删除文件" default(false)
// @Success 200 {object} Response
// @Router /torrent/{id} [delete]
func (h *Handler) DeleteTorrent(c *gin.Context) {
	torrentID := c.Param("id")
	deleteFilesParam := c.DefaultQuery("delete_files", "false")

	if torrentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "种子ID不能为空",
		})
		return
	}

	deleteFiles := deleteFilesParam == "true"

	err := h.service.DeleteTorrent(torrentID, deleteFiles)
	if err != nil {
		h.logger.Error("Failed to delete torrent", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除种子失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "种子删除成功",
	})
}

// GetTorrentFiles 获取种子文件列表
// @Summary 获取种子文件列表
// @Description 获取种子的文件列表
// @Tags 种子
// @Produce json
// @Param id path string true "种子ID"
// @Success 200 {object} Response{data=[]TorrentFile}
// @Router /torrent/{id}/files [get]
func (h *Handler) GetTorrentFiles(c *gin.Context) {
	torrentID := c.Param("id")
	if torrentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "种子ID不能为空",
		})
		return
	}

	files, err := h.service.GetTorrentFiles(torrentID)
	if err != nil {
		h.logger.Error("Failed to get torrent files", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取种子文件列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    files,
	})
}

// SetFilePriority 设置文件优先级
// @Summary 设置文件优先级
// @Description 设置种子文件的下载优先级
// @Tags 种子
// @Produce json
// @Param id path string true "种子ID"
// @Param file body FilePriorityRequest true "文件优先级设置"
// @Success 200 {object} Response
// @Router /torrent/{id}/files/priority [post]
func (h *Handler) SetFilePriority(c *gin.Context) {
	torrentID := c.Param("id")
	if torrentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "种子ID不能为空",
		})
		return
	}

	var request FilePriorityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	err := h.service.SetFilePriority(torrentID, request)
	if err != nil {
		h.logger.Error("Failed to set file priority", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "设置文件优先级失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件优先级设置成功",
	})
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// TorrentInfo 种子信息结构
type TorrentInfo struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Size          int64   `json:"size"`
	Downloaded    int64   `json:"downloaded"`
	Uploaded      int64   `json:"uploaded"`
	Status        string  `json:"status"`
	Progress      float64 `json:"progress"`
	DownloadSpeed int64   `json:"download_speed"`
	UploadSpeed   int64   `json:"upload_speed"`
	Ratio         float64 `json:"ratio"`
	CreateTime    string  `json:"create_time"`
	UpdateTime    string  `json:"update_time"`
}

// TorrentDetail 种子详情结构
type TorrentDetail struct {
	TorrentInfo
	Files         []TorrentFile `json:"files"`
	Trackers      []Tracker     `json:"trackers"`
	Peers         []Peer        `json:"peers"`
	Seeders       int           `json:"seeders"`
	Leechers      int           `json:"leechers"`
	DownloadLimit int64         `json:"download_limit"`
	UploadLimit   int64         `json:"upload_limit"`
	SavePath      string        `json:"save_path"`
	Comment       string        `json:"comment"`
	Hash          string        `json:"hash"`
}

// AddTorrentRequest 添加种子请求结构
type AddTorrentRequest struct {
	MagnetURL   string `json:"magnet_url"`
	TorrentFile string `json:"torrent_file"`
	SavePath    string `json:"save_path"`
	Label       string `json:"label"`
	Paused      bool   `json:"paused"`
}

// TorrentFile 种子文件结构
type TorrentFile struct {
	Index      int     `json:"index"`
	Name       string  `json:"name"`
	Size       int64   `json:"size"`
	Downloaded int64   `json:"downloaded"`
	Priority   string  `json:"priority"`
	Progress   float64 `json:"progress"`
}

// Tracker Tracker信息结构
type Tracker struct {
	URL      string `json:"url"`
	Status   string `json:"status"`
	Seeders  int    `json:"seeders"`
	Leechers int    `json:"leechers"`
}

// Peer Peer信息结构
type Peer struct {
	IP            string  `json:"ip"`
	Client        string  `json:"client"`
	Country       string  `json:"country"`
	DownloadSpeed int64   `json:"download_speed"`
	UploadSpeed   int64   `json:"upload_speed"`
	Progress      float64 `json:"progress"`
}

// FilePriorityRequest 文件优先级请求结构
type FilePriorityRequest struct {
	FileIndex int    `json:"file_index"`
	Priority  string `json:"priority"`
}
