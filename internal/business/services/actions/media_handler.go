package actions

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"moviepilot-go/internal/business/services/actions/types"
	"moviepilot-go/pkg/logger"
)

// MediaHandler 媒体处理器
type MediaHandler struct {
	mediaFetcher *MediaFetcher
	logger       *zap.Logger
}

// NewMediaHandler 创建媒体处理器实例
func NewMediaHandler(mediaFetcher *MediaFetcher) *MediaHandler {
	return &MediaHandler{
		mediaFetcher: mediaFetcher,
		logger:       logger.Logger,
	}
}

// FetchMedias 获取媒体数据
// @Summary 获取媒体数据
// @Description 从各种数据源获取媒体信息
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param request body FetchMediasParams true "获取媒体参数"
// @Success 200 {array} FetchMediasResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/medias/fetch [post]
func (h *MediaHandler) FetchMedias(c *gin.Context) {
	var params FetchMediasParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Error("参数绑定失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数格式错误: " + err.Error(),
		})
		return
	}

	// 获取工作流ID，默认为0
	workflowID := int64(0)
	if workflowIDStr := c.Query("workflow_id"); workflowIDStr != "" {
		if id, err := strconv.ParseInt(workflowIDStr, 10, 64); err == nil {
			workflowID = id
		}
	}

	results, err := h.mediaFetcher.FetchMedias(c.Request.Context(), workflowID, &params)
	if err != nil {
		h.logger.Error("获取媒体数据失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取媒体数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetAvailableSources 获取可用数据源
// @Summary 获取可用数据源
// @Description 获取系统中配置的所有可用媒体数据源
// @Tags 媒体管理
// @Produce json
// @Success 200 {array} MediaSource
// @Failure 500 {object} map[string]string
// @Router /api/v1/medias/sources [get]
func (h *MediaHandler) GetAvailableSources(c *gin.Context) {
	sources := h.mediaFetcher.GetAvailableSources()
	c.JSON(http.StatusOK, sources)
}

// EnableSource 启用数据源
// @Summary 启用数据源
// @Description 启用指定的媒体数据源
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param sourceName path string true "数据源名称"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/medias/sources/{sourceName}/enable [post]
func (h *MediaHandler) EnableSource(c *gin.Context) {
	sourceName := c.Param("sourceName")
	if sourceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "数据源名称不能为空",
		})
		return
	}

	err := h.mediaFetcher.EnableSource(sourceName)
	if err != nil {
		h.logger.Error("启用数据源失败", zap.String("source", sourceName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "启用数据源失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据源启用成功",
		"source":  sourceName,
	})
}

// DisableSource 禁用数据源
// @Summary 禁用数据源
// @Description 禁用指定的媒体数据源
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param sourceName path string true "数据源名称"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/medias/sources/{sourceName}/disable [post]
func (h *MediaHandler) DisableSource(c *gin.Context) {
	sourceName := c.Param("sourceName")
	if sourceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "数据源名称不能为空",
		})
		return
	}

	err := h.mediaFetcher.DisableSource(sourceName)
	if err != nil {
		h.logger.Error("禁用数据源失败", zap.String("source", sourceName), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "禁用数据源失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据源禁用成功",
		"source":  sourceName,
	})
}

// SearchMedias 搜索媒体（兼容接口）
// @Summary 搜索媒体
// @Description 根据条件搜索媒体信息
// @Tags 媒体管理
// @Accept json
// @Produce json
// @Param query query string false "搜索关键词"
// @Param type query string false "媒体类型"
// @Param year query int false "年份"
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/medias/search [get]
func (h *MediaHandler) SearchMedias(c *gin.Context) {
	// 构建搜索参数
	params := &FetchMediasParams{
		SourceType: "search",
		APIPath:    "search/media",
	}

	// 解析查询参数
	if query := c.Query("query"); query != "" {
		// 在实际实现中，这里应该将查询参数传递给相应的搜索函数
		h.logger.Debug("搜索媒体", zap.String("query", query))
	}

	if mediaType := c.Query("type"); mediaType != "" {
		// 处理媒体类型过滤
	}

	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			params.Year = year
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	} else {
		params.Limit = 20
	}

	// 执行搜索
	results, err := h.mediaFetcher.FetchMedias(c.Request.Context(), 0, params)
	if err != nil {
		h.logger.Error("搜索媒体失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索媒体失败: " + err.Error(),
		})
		return
	}

	// 聚合结果
	aggregated := []*types.MediaInfo{}
	for _, result := range results {
		if result.Success {
			aggregated = append(aggregated, result.Medias...)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":    aggregated,
		"total":    len(aggregated),
		"page":     1,
		"page_size": params.Limit,
	})
}

// GetMediaByID 根据ID获取媒体详情（兼容接口）
// @Summary 获取媒体详情
// @Description 根据媒体ID获取详细信息
// @Tags 媒体管理
// @Produce json
// @Param mediaId path string true "媒体ID"
// @Param includeEpisodes query bool false "是否包含剧集信息"
// @Success 200 {object} types.MediaInfo
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/medias/{mediaId} [get]
func (h *MediaHandler) GetMediaByID(c *gin.Context) {
	mediaID := c.Param("mediaId")
	if mediaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "媒体ID不能为空",
		})
		return
	}

	includeEpisodes := false
	if includeStr := c.Query("includeEpisodes"); includeStr == "true" {
		includeEpisodes = true
	}

	// 构建获取详情的参数
	params := &FetchMediasParams{
		SourceType: "detail",
		APIPath:    "detail/media",
		// 这里需要设置媒体ID参数
	}

	// 在实际实现中，这里应该调用专门的获取详情函数
	results, err := h.mediaFetcher.FetchMedias(c.Request.Context(), 0, params)
	if err != nil {
		h.logger.Error("获取媒体详情失败", zap.String("media_id", mediaID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取媒体详情失败: " + err.Error(),
		})
		return
	}

	// 从结果中提取第一个媒体信息
	if len(results) > 0 && len(results[0].Medias) > 0 {
		c.JSON(http.StatusOK, results[0].Medias[0])
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "未找到媒体信息",
		})
	}
}
