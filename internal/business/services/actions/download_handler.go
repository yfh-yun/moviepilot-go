// Package actions 提供下载处理器的实现
package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"moviepilot-go/internal/business/services/actions/types"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"
)

// DownloadHandler 下载处理器
type DownloadHandler struct {
	downloadManager *DownloadManagerImpl
	logger          logger.Logger
}

// NewDownloadHandler 创建下载处理器实例
func NewDownloadHandler(
	downloadRepo interfaces.DownloadRepository,
	mediaRepo interfaces.MediaRepository,
	cache *WorkflowCache,
) *DownloadHandler {
	return &DownloadHandler{
		downloadManager: NewDownloadManager(downloadRepo, mediaRepo, cache),
		logger:          logger.NewLogger("download_handler"),
	}
}

// CreateDownload 创建下载任务
// @Summary 创建下载任务
// @Description 根据提供的参数创建下载任务
// @Tags downloads
// @Accept json
// @Produce json
// @Param request body CreateDownloadRequest true "下载请求参数"
// @Success 200 {object} CreateDownloadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/downloads [post]
func (h *DownloadHandler) CreateDownload(c *gin.Context) {
	var req CreateDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("请求参数无效", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "请求参数无效",
			Message: err.Error(),
		})
		return
	}

	// 验证参数
	if err := h.validateCreateRequest(&req); err != nil {
		h.logger.Error("参数验证失败", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "参数验证失败",
			Message: err.Error(),
		})
		return
	}

	// 构建下载参数
	downloadParams := &AddDownloadParams{
		Downloader: req.Downloader,
		SavePath:   req.SavePath,
		Labels:     req.Labels,
		OnlyLack:   req.OnlyLack,
		Sites:      req.Sites,
		Quality:    req.Quality,
		Resolution: req.Resolution,
	}

	// 构建种子列表
	torrents := make([]*types.Torrent, 0, len(req.Torrents))
	for _, t := range req.Torrents {
		torrents = append(torrents, &types.Torrent{
			ID:     t.ID,
			Title:  t.Title,
			URL:    t.URL,
			SiteID: t.SiteID,
			Size:   t.Size,
		})
	}

	// 执行下载添加
	results, err := h.downloadManager.AddDownload(c.Request.Context(), req.WorkflowID, downloadParams, torrents)
	if err != nil {
		h.logger.Error("添加下载任务失败", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "添加下载任务失败",
			Message: err.Error(),
		})
		return
	}

	// 构建响应
	response := CreateDownloadResponse{
		Total:    len(results),
		Success:  h.getSuccessCount(results),
		Failures: h.getFailureCount(results),
		Results:  make([]DownloadResult, 0, len(results)),
	}

	for _, r := range results {
		response.Results = append(response.Results, DownloadResult{
			Success:    r.Success,
			DownloadID: r.DownloadID,
			Message:    r.Message,
			TorrentID:  r.Torrent.ID,
			TorrentTitle: r.Torrent.Title,
		})
	}

	h.logger.Info("下载任务创建完成", "total", response.Total, "success", response.Success)
	c.JSON(http.StatusOK, response)
}

// GetDownloadStatus 获取下载状态
// @Summary 获取下载状态
// @Description 根据下载ID获取下载状态信息
// @Tags downloads
// @Accept json
// @Produce json
// @Param id path string true "下载ID"
// @Success 200 {object} DownloadStatusResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/downloads/{id}/status [get]
func (h *DownloadHandler) GetDownloadStatus(c *gin.Context) {
	downloadID := c.Param("id")
	if downloadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "下载ID不能为空",
			Message: "请提供有效的下载ID",
		})
		return
	}

	// 这里应该调用下载管理器的获取状态方法
	// 暂时返回模拟数据
	status := &types.DownloadStatus{
		DownloadID:    downloadID,
		Status:        "downloading",
		Progress:      0.65,
		DownloadSpeed: 1024 * 1024, // 1MB/s
		UploadSpeed:   1024,        // 1KB/s
		Seeders:       10,
		Leechers:      5,
		ETA:           "00:15:30",
	}

	c.JSON(http.StatusOK, DownloadStatusResponse{
		Status: status,
	})
}

// CancelDownload 取消下载
// @Summary 取消下载
// @Description 根据下载ID取消下载任务
// @Tags downloads
// @Accept json
// @Produce json
// @Param id path string true "下载ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/downloads/{id} [delete]
func (h *DownloadHandler) CancelDownload(c *gin.Context) {
	downloadID := c.Param("id")
	if downloadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "下载ID不能为空",
			Message: "请提供有效的下载ID",
		})
		return
	}

	// 这里应该调用下载管理器的取消方法
	// 暂时返回成功
	h.logger.Info("下载任务已取消", "download_id", downloadID)

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "下载任务已取消",
	})
}

// validateCreateRequest 验证创建请求参数
func (h *DownloadHandler) validateCreateRequest(req *CreateDownloadRequest) error {
	if req.Downloader == "" {
		return &ValidationError{Field: "downloader", Message: "下载器类型不能为空"}
	}

	if len(req.Torrents) == 0 {
		return &ValidationError{Field: "torrents", Message: "种子列表不能为空"}
	}

	for i, torrent := range req.Torrents {
		if torrent.URL == "" {
			return &ValidationError{Field: fmt.Sprintf("torrents[%d].url", i), Message: "种子URL不能为空"}
		}
	}

	return nil
}

// getSuccessCount 获取成功的数量
func (h *DownloadHandler) getSuccessCount(results []*AddDownloadResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

// getFailureCount 获取失败的数量
func (h *DownloadHandler) getFailureCount(results []*AddDownloadResult) int {
	count := 0
	for _, result := range results {
		if !result.Success {
			count++
		}
	}
	return count
}

// 请求和响应结构定义

// CreateDownloadRequest 创建下载请求
type CreateDownloadRequest struct {
	WorkflowID int64           `json:"workflow_id" binding:"required"`
	Downloader string          `json:"downloader" binding:"required"`
	SavePath   string          `json:"save_path"`
	Labels     []string        `json:"labels"`
	OnlyLack   bool            `json:"only_lack"`
	Sites      []int           `json:"sites"`
	Quality    string          `json:"quality"`
	Resolution string          `json:"resolution"`
	Torrents   []TorrentRequest `json:"torrents" binding:"required"`
}

// TorrentRequest 种子请求结构
type TorrentRequest struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	URL    string `json:"url" binding:"required"`
	SiteID int    `json:"site_id"`
	Size   int64  `json:"size"`
}

// CreateDownloadResponse 创建下载响应
type CreateDownloadResponse struct {
	Total    int              `json:"total"`
	Success  int              `json:"success"`
	Failures int              `json:"failures"`
	Results  []DownloadResult `json:"results"`
}

// DownloadResult 下载结果
type DownloadResult struct {
	Success      bool   `json:"success"`
	DownloadID   string `json:"download_id,omitempty"`
	Message      string `json:"message"`
	TorrentID    string `json:"torrent_id"`
	TorrentTitle string `json:"torrent_title"`
}

// DownloadStatusResponse 下载状态响应
type DownloadStatusResponse struct {
	Status *types.DownloadStatus `json:"status"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("字段 '%s': %s", e.Field, e.Message)
}
