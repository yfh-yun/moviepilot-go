package actions

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"moviepilot-go/pkg/logger"
)

// TorrentFilterHandler 种子过滤器HTTP处理器接口
type TorrentFilterHandler interface {
	// FilterTorrents 过滤种子列表
	FilterTorrents(c *gin.Context)
	// ValidateFilter 验证过滤参数
	ValidateFilter(c *gin.Context)
	// GetFilterStats 获取过滤统计信息
	GetFilterStats(c *gin.Context)
	// GetFilterSuggestions 获取过滤建议
	GetFilterSuggestions(c *gin.Context)
	// PreviewFilter 预览过滤结果
	PreviewFilter(c *gin.Context)
	// ExportFilterResults 导出过滤结果
	ExportFilterResults(c *gin.Context)
}

// torrentFilterHandler 种子过滤器HTTP处理器实现
type torrentFilterHandler struct {
	logger         logger.Logger
	torrentFilter  TorrentFilter
	torrentManager TorrentManager
}

// NewTorrentFilterHandler 创建种子过滤器HTTP处理器实例
func NewTorrentFilterHandler(logger logger.Logger, torrentFilter TorrentFilter, torrentManager TorrentManager) TorrentFilterHandler {
	return &torrentFilterHandler{
		logger:         logger,
		torrentFilter:  torrentFilter,
		torrentManager: torrentManager,
	}
}

// FilterTorrents 过滤种子列表
// @Summary 过滤种子列表
// @Description 根据指定条件过滤种子列表，支持分页、排序和复杂过滤
// @Tags 种子过滤
// @Accept json
// @Produce json
// @Param filter body TorrentFilterParams false "过滤参数"
// @Success 200 {object} TorrentFilterResponse "过滤结果"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/torrents/filter [post]
func (h *torrentFilterHandler) FilterTorrents(c *gin.Context) {
	log := h.logger.WithContext(c)
	log.Debug("Starting torrent filter request")

	startTime := time.Now()
	ctx := c.Request.Context()

	// 解析请求参数
	var params TorrentFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		// 尝试从查询参数获取简单过滤条件
		params = h.parseQueryParams(c)
		log.Warn("Failed to bind JSON, using query params", "error", err.Error())
	}

	// 验证参数
	if err := h.torrentFilter.ValidateFilter(ctx, &params); err != nil {
		log.Error("Invalid filter parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "参数验证失败",
			Details: err.Error(),
		})
		return
	}

	// 执行过滤
	response, err := h.torrentFilter.FilterTorrents(ctx, &params)
	if err != nil {
		log.Error("Failed to filter torrents", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "过滤种子失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Torrent filtering completed successfully", 
		"total", response.Total, 
		"returned", len(response.Items),
		"elapsed", time.Since(startTime).Seconds())

	c.JSON(http.StatusOK, response)
}

// ValidateFilter 验证过滤参数
// @Summary 验证过滤参数
// @Description 验证种子过滤参数的有效性
// @Tags 种子过滤
// @Accept json
// @Produce json
// @Param filter body TorrentFilterParams true "过滤参数"
// @Success 200 {object} ValidationResponse "验证成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Router /api/v1/torrents/filter/validate [post]
func (h *torrentFilterHandler) ValidateFilter(c *gin.Context) {
	log := h.logger.WithContext(c)
	log.Debug("Validating torrent filter parameters")

	// 解析请求参数
	var params TorrentFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Error("Failed to bind filter parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "参数格式错误",
			Details: err.Error(),
		})
		return
	}

	// 执行验证
	if err := h.torrentFilter.ValidateFilter(c.Request.Context(), &params); err != nil {
		log.Warn("Filter parameters validation failed", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "参数验证失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Filter parameters validation successful")
	c.JSON(http.StatusOK, ValidationResponse{
		Valid:   true,
		Message: "参数验证通过",
	})
}

// GetFilterStats 获取过滤统计信息
// @Summary 获取过滤统计信息
// @Description 获取符合过滤条件的种子统计信息
// @Tags 种子过滤
// @Accept json
// @Produce json
// @Param filter body TorrentFilterParams false "过滤参数"
// @Success 200 {object} TorrentFilterStats "统计信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/torrents/filter/stats [post]
func (h *torrentFilterHandler) GetFilterStats(c *gin.Context) {
	log := h.logger.WithContext(c)
	log.Debug("Getting filter statistics")

	// 解析请求参数
	var params TorrentFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		// 尝试从查询参数获取简单过滤条件
		params = h.parseQueryParams(c)
		log.Warn("Failed to bind JSON, using query params", "error", err.Error())
	}

	// 获取统计信息
	stats, err := h.torrentFilter.GetFilterStats(c.Request.Context(), &params)
	if err != nil {
		log.Error("Failed to get filter statistics", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取统计信息失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Filter statistics retrieved successfully", "total", stats.TotalCount)
	c.JSON(http.StatusOK, stats)
}

// GetFilterSuggestions 获取过滤建议
// @Summary 获取过滤建议
// @Description 获取种子过滤的建议值
// @Tags 种子过滤
// @Accept json
// @Produce json
// @Param filter body TorrentFilterParams false "过滤参数"
// @Success 200 {array} TorrentFilterSuggestion "过滤建议列表"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/torrents/filter/suggestions [post]
func (h *torrentFilterHandler) GetFilterSuggestions(c *gin.Context) {
	log := h.logger.WithContext(c)
	log.Debug("Getting filter suggestions")

	// 解析请求参数
	var params TorrentFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		// 尝试从查询参数获取简单过滤条件
		params = h.parseQueryParams(c)
		log.Warn("Failed to bind JSON, using query params", "error", err.Error())
	}

	// 获取建议
	suggestions, err := h.torrentFilter.GetFilterSuggestions(c.Request.Context(), &params)
	if err != nil {
		log.Error("Failed to get filter suggestions", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取过滤建议失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Filter suggestions retrieved successfully", "count", len(suggestions))
	c.JSON(http.StatusOK, suggestions)
}

// PreviewFilter 预览过滤结果
// @Summary 预览过滤结果
// @Description 预览过滤条件的效果，返回样本数据
// @Tags 种子过滤
// @Accept json
// @Produce json
// @Param filter body TorrentFilterParams true "过滤参数"
// @Success 200 {object} TorrentFilterPreview "预览结果"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/torrents/filter/preview [post]
func (h *torrentFilterHandler) PreviewFilter(c *gin.Context) {
	log := h.logger.WithContext(c)
	log.Debug("Previewing filter results")

	// 解析请求参数
	var params TorrentFilterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Error("Failed to bind filter parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "参数格式错误",
			Details: err.Error(),
		})
		return
	}

	// 执行预览
	preview, err := h.torrentFilter.PreviewFilter(c.Request.Context(), &params)
	if err != nil {
		log.Error("Failed to preview filter results", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "预览过滤结果失败",
			Details: err.Error(),
		})
		return
	}

	log.Info("Filter preview generated successfully", "preview_count", preview.PreviewCount)
	c.JSON(http.StatusOK, preview)
}

// ExportFilterResults 导出过滤结果
// @Summary 导出过滤结果
// @Description 将过滤结果导出为指定格式的文件
// @Tags 种子过滤
// @Accept json
// @Produce application/octet-stream
// @Param export body TorrentExportParams true "导出参数"
// @Success 200 {file} file "导出文件"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/torrents/filter/export [post]
func (h *torrentFilterHandler) ExportFilterResults(c *gin.Context) {
	log := h.logger.WithContext(c)
	log.Debug("Exporting filter results")

	// 解析请求参数
	var params TorrentExportParams
	if err := c.ShouldBindJSON(&params); err != nil {
		log.Error("Failed to bind export parameters", "error", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "参数格式错误",
			Details: err.Error(),
		})
		return
	}

	// 执行导出
	data, fileName, err := h.torrentFilter.ExportFilterResults(c.Request.Context(), &params)
	if err != nil {
		log.Error("Failed to export filter results", "error", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "导出过滤结果失败",
			Details: err.Error(),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.Itoa(len(data)))

	log.Info("Filter results exported successfully", "file", fileName, "size", len(data))
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// parseQueryParams 从查询参数解析简单的过滤条件
func (h *torrentFilterHandler) parseQueryParams(c *gin.Context) TorrentFilterParams {
	params := TorrentFilterParams{}

	// 分页参数
	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		params.Page = page
	}

	if limit, err := strconv.Atoi(c.Query("limit")); err == nil && limit > 0 {
		params.Limit = limit
	}

	if offset, err := strconv.Atoi(c.Query("offset")); err == nil && offset >= 0 {
		params.Offset = offset
	}

	// 排序参数
	if sortBy := c.Query("sort_by"); sortBy != "" {
		params.SortBy = TorrentSortField(sortBy)
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		params.SortOrder = SortOrder(sortOrder)
	}

	// 状态过滤
	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		params.Statuses = make([]TorrentStatus, len(statuses))
		for i, status := range statuses {
			params.Statuses[i] = TorrentStatus(status)
		}
	}

	// 类型过滤
	if types := c.QueryArray("type"); len(types) > 0 {
		params.Types = make([]TorrentType, len(types))
		for i, t := range types {
			params.Types[i] = TorrentType(t)
		}
	}

	// 分类过滤
	if categories := c.QueryArray("category"); len(categories) > 0 {
		params.Categories = categories
	}

	// 下载器过滤
	if downloaders := c.QueryArray("downloader"); len(downloaders) > 0 {
		params.Downloaders = downloaders
	}

	// 媒体类型过滤
	if mediaTypes := c.QueryArray("media_type"); len(mediaTypes) > 0 {
		params.MediaTypes = make([]MediaType, len(mediaTypes))
		for i, mt := range mediaTypes {
			params.MediaTypes[i] = MediaType(mt)
		}
	}

	// 质量过滤
	if qualities := c.QueryArray("quality"); len(qualities) > 0 {
		params.Qualities = qualities
	}

	// 名称过滤
	if name := c.Query("name"); name != "" {
		params.Names = []string{name}
	}

	// 精确匹配
	if exactMatch := c.Query("exact_match"); exactMatch == "true" {
		params.ExactMatch = true
	}

	// 特殊状态过滤
	if onlyActive := c.Query("only_active"); onlyActive == "true" {
		params.OnlyActive = true
	}

	if onlyCompleted := c.Query("only_completed"); onlyCompleted == "true" {
		params.OnlyCompleted = true
	}

	if onlyDownloading := c.Query("only_downloading"); onlyDownloading == "true" {
		params.OnlyDownloading = true
	}

	if onlySeeding := c.Query("only_seeding"); onlySeeding == "true" {
		params.OnlySeeding = true
	}

	if onlyPaused := c.Query("only_paused"); onlyPaused == "true" {
		params.OnlyPaused = true
	}

	if onlyStalled := c.Query("only_stalled"); onlyStalled == "true" {
		params.OnlyStalled = true
	}

	// 数值范围过滤
	if sizeMin, err := strconv.ParseInt(c.Query("size_min"), 10, 64); err == nil {
		params.SizeMin = &sizeMin
	}

	if sizeMax, err := strconv.ParseInt(c.Query("size_max"), 10, 64); err == nil {
		params.SizeMax = &sizeMax
	}

	if ratioMin, err := strconv.ParseFloat(c.Query("ratio_min"), 64); err == nil {
		params.RatioMin = &ratioMin
	}

	if ratioMax, err := strconv.ParseFloat(c.Query("ratio_max"), 64); err == nil {
		params.RatioMax = &ratioMax
	}

	if progressMin, err := strconv.ParseFloat(c.Query("progress_min"), 64); err == nil {
		params.ProgressMin = &progressMin
	}

	if progressMax, err := strconv.ParseFloat(c.Query("progress_max"), 64); err == nil {
		params.ProgressMax = &progressMax
	}

	return params
}
