package actions

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"moviepilot-go/pkg/logger"
)

// RSSHandler RSS处理器接口
type RSSHandler interface {
	// FetchRSS 获取RSS数据
	FetchRSS(c *gin.Context)
	// FetchMultipleRSS 批量获取RSS数据
	FetchMultipleRSS(c *gin.Context)
	// ValidateRSSFeed 验证RSS源
	ValidateRSSFeed(c *gin.Context)
	// GetRSSStats 获取RSS统计信息
	GetRSSStats(c *gin.Context)
	// ClearCache 清除RSS缓存
	ClearCache(c *gin.Context)
	// GetAvailableRSSSources 获取可用的RSS源列表
	GetAvailableRSSSources(c *gin.Context)
}

// RSSHandlerImpl RSS处理器实现
type RSSHandlerImpl struct {
	rssFetcher RSSFetcher
	validator  *RSSValidator
	logger     *zap.Logger
}

// NewRSSHandler 创建RSS处理器实例
func NewRSSHandler(rssFetcher RSSFetcher, validator *RSSValidator) *RSSHandlerImpl {
	return &RSSHandlerImpl{
		rssFetcher: rssFetcher,
		validator:  validator,
		logger:     logger.Logger,
	}
}

// @Summary 获取RSS数据
// @Description 从指定的RSS源获取数据，支持过滤、缓存等功能
// @Tags RSS
// @Accept json
// @Produce json
// @Param url query string true "RSS源URL"
// @Param format query string false "RSS格式(xml/json/custom)" default(xml)
// @Param timeout query int false "超时时间(秒)" default(30)
// @Param retries query int false "重试次数" default(3)
// @Param delay query int false "重试延迟(秒)" default(2)
// @Param limit query int false "返回条目数量限制"
// @Param cache query bool false "是否使用缓存" default(true)
// @Param cache_ttl query int false "缓存时间(分钟)" default(30)
// @Param include_title query string false "包含标题关键词(逗号分隔)"
// @Param exclude_title query string false "排除标题关键词(逗号分隔)"
// @Param include_keywords query string false "包含内容关键词(逗号分隔)"
// @Param exclude_keywords query string false "排除内容关键词(逗号分隔)"
// @Param min_size query int false "最小文件大小(字节)"
// @Param max_size query int false "最大文件大小(字节)"
// @Param min_seeders query int false "最小做种数"
// @Success 200 {object} RSSResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/rss/fetch [get]
func (h *RSSHandlerImpl) FetchRSS(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// 解析查询参数
	params, err := h.parseFetchParams(c)
	if err != nil {
		h.logger.Error("Failed to parse fetch params", zap.String("request_id", requestID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "参数解析失败: " + err.Error(),
		})
		return
	}

	// 验证参数
	if err := h.validator.ValidateFetchParams(params); err != nil {
		h.logger.Warn("Invalid fetch params", zap.String("request_id", requestID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "参数验证失败: " + err.Error(),
		})
		return
	}

	// 记录请求日志
	h.logger.Info("Fetching RSS data", 
		zap.String("request_id", requestID),
		zap.String("url", params.FeedURL),
		zap.String("format", params.Format),
	)

	// 执行获取
	response, err := h.rssFetcher.FetchRSS(ctx, params)
	if err != nil {
		h.logger.Error("Failed to fetch RSS", 
			zap.String("request_id", requestID),
			zap.String("url", params.FeedURL),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取RSS数据失败: " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, response)
}

// @Summary 批量获取RSS数据
// @Description 从多个RSS源批量获取数据
// @Tags RSS
// @Accept json
// @Produce json
// @Param request body []FetchRSSParams true "RSS获取请求列表"
// @Success 200 {array} RSSResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/rss/fetch-multiple [post]
func (h *RSSHandlerImpl) FetchMultipleRSS(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// 解析请求体
	var paramsList []*FetchRSSParams
	if err := c.ShouldBindJSON(&paramsList); err != nil {
		h.logger.Error("Failed to parse request body", zap.String("request_id", requestID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求体解析失败: " + err.Error(),
		})
		return
	}

	// 验证请求数量
	if len(paramsList) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求列表不能为空",
		})
		return
	}

	if len(paramsList) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求数量不能超过10个",
		})
		return
	}

	// 验证每个参数
	for _, params := range paramsList {
		if err := h.validator.ValidateFetchParams(params); err != nil {
			h.logger.Warn("Invalid fetch params in batch", zap.String("request_id", requestID), zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "参数验证失败: " + err.Error(),
			})
			return
		}
	}

	// 记录请求日志
	h.logger.Info("Batch fetching RSS data", 
		zap.String("request_id", requestID),
		zap.Int("count", len(paramsList)),
	)

	// 执行批量获取
	responses, err := h.rssFetcher.FetchMultipleRSS(ctx, paramsList)
	if err != nil {
		h.logger.Error("Failed to batch fetch RSS", 
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "批量获取RSS数据失败: " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, responses)
}

// @Summary 验证RSS源
// @Description 验证指定的RSS源是否有效
// @Tags RSS
// @Accept json
// @Produce json
// @Param url query string true "RSS源URL"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/rss/validate [get]
func (h *RSSHandlerImpl) ValidateRSSFeed(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// 获取URL参数
	feedURL := c.Query("url")
	if feedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "URL参数不能为空",
		})
		return
	}

	// 记录请求日志
	h.logger.Info("Validating RSS feed", 
		zap.String("request_id", requestID),
		zap.String("url", feedURL),
	)

	// 执行验证
	valid, err := h.rssFetcher.ValidateRSSFeed(ctx, feedURL)
	if err != nil {
		h.logger.Warn("RSS feed validation failed", 
			zap.String("request_id", requestID),
			zap.String("url", feedURL),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"valid":     false,
			"error":     "验证失败: " + err.Error(),
			"feed_url":  feedURL,
			"timestamp": time.Now(),
		})
		return
	}

	// 返回结果
	status := "有效"
	if !valid {
		status = "无效"
	}

	h.logger.Info("RSS feed validation completed", 
		zap.String("request_id", requestID),
		zap.String("url", feedURL),
		zap.Bool("valid", valid),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"valid":     valid,
		"status":    status,
		"feed_url":  feedURL,
		"timestamp": time.Now(),
	})
}

// @Summary 获取RSS统计信息
// @Description 获取RSS获取器的统计信息
// @Tags RSS
// @Accept json
// @Produce json
// @Success 200 {object} RSSStats
// @Failure 500 {object} gin.H
// @Router /api/v1/rss/stats [get]
func (h *RSSHandlerImpl) GetRSSStats(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")

	// 记录请求日志
	h.logger.Info("Getting RSS statistics", zap.String("request_id", requestID))

	// 获取统计信息
	stats, err := h.rssFetcher.GetRSSStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get RSS statistics", 
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取统计信息失败: " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// @Summary 清除RSS缓存
// @Description 清除指定RSS源的缓存
// @Tags RSS
// @Accept json
// @Produce json
// @Param url query string true "RSS源URL"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/rss/clear-cache [post]
func (h *RSSHandlerImpl) ClearCache(c *gin.Context) {
	requestID := c.GetString("request_id")

	// 获取URL参数
	feedURL := c.Query("url")
	if feedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "URL参数不能为空",
		})
		return
	}

	// 记录请求日志
	h.logger.Info("Clearing RSS cache", 
		zap.String("request_id", requestID),
		zap.String("url", feedURL),
	)

	// 执行清除缓存
	err := h.rssFetcher.ClearCache(feedURL)
	if err != nil {
		h.logger.Error("Failed to clear RSS cache", 
			zap.String("request_id", requestID),
			zap.String("url", feedURL),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "清除缓存失败: " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "缓存已清除",
		"feed_url":  feedURL,
		"timestamp": time.Now(),
	})
}

// @Summary 获取可用的RSS源列表
// @Description 获取系统中可用的RSS源列表
// @Tags RSS
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/rss/sources [get]
func (h *RSSHandlerImpl) GetAvailableRSSSources(c *gin.Context) {
	requestID := c.GetString("request_id")

	// 记录请求日志
	h.logger.Info("Getting available RSS sources", zap.String("request_id", requestID))

	// 返回预定义的RSS源列表
	// 实际项目中应该从配置或数据库获取
	sources := []map[string]interface{}{
		{
			"id":          "movie_source_1",
			"name":        "电影资源源1",
			"url":         "https://example.com/movies.rss",
			"type":        RSSTypeMovie,
			"description": "提供最新电影资源的RSS源",
			"enabled":     true,
		},
		{
			"id":          "series_source_1",
			"name":        "剧集资源源1",
			"url":         "https://example.com/series.rss",
			"type":        RSSTypeSeries,
			"description": "提供最新剧集资源的RSS源",
			"enabled":     true,
		},
		{
			"id":          "anime_source_1",
			"name":        "动漫资源源1",
			"url":         "https://example.com/anime.rss",
			"type":        RSSTypeAnimation,
			"description": "提供最新动漫资源的RSS源",
			"enabled":     true,
		},
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"sources":  sources,
		"count":    len(sources),
		"timestamp": time.Now(),
	})
}

// 私有辅助方法

// parseFetchParams 解析获取参数
func (h *RSSHandlerImpl) parseFetchParams(c *gin.Context) (*FetchRSSParams, error) {
	params := &FetchRSSParams{
		FeedURL:  c.Query("url"),
		Format:   c.DefaultQuery("format", RSSFormatXML),
		UserAgent: c.Query("user_agent"),
		Username:  c.Query("username"),
		Password:  c.Query("password"),
		Cookies:   c.Query("cookies"),
		Headers:   make(map[string]string),
	}

	// 解析超时时间
	if timeoutStr := c.Query("timeout"); timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, err
		}
		params.Timeout = timeout
	}

	// 解析重试次数
	if retriesStr := c.Query("retries"); retriesStr != "" {
		retries, err := strconv.Atoi(retriesStr)
		if err != nil {
			return nil, err
		}
		params.Retries = retries
	}

	// 解析延迟时间
	if delayStr := c.Query("delay"); delayStr != "" {
		delay, err := strconv.Atoi(delayStr)
		if err != nil {
			return nil, err
		}
		params.Delay = delay
	}

	// 解析限制数量
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, err
		}
		params.Limit = limit
	}

	// 解析缓存设置
	if cacheStr := c.Query("cache"); cacheStr != "" {
		cache, err := strconv.ParseBool(cacheStr)
		if err != nil {
			return nil, err
		}
		params.CacheEnabled = cache
	}

	// 解析缓存TTL
	if cacheTTLStr := c.Query("cache_ttl"); cacheTTLStr != "" {
		cacheTTL, err := strconv.Atoi(cacheTTLStr)
		if err != nil {
			return nil, err
		}
		params.CacheTTL = cacheTTL
	}

	// 解析过滤器
	params.Filters = &RSSFilters{}

	// 解析标题关键词
	if includeTitle := c.Query("include_title"); includeTitle != "" {
		params.Filters.IncludeTitle = parseCommaList(includeTitle)
	}

	if excludeTitle := c.Query("exclude_title"); excludeTitle != "" {
		params.Filters.ExcludeTitle = parseCommaList(excludeTitle)
	}

	// 解析内容关键词
	if includeKeywords := c.Query("include_keywords"); includeKeywords != "" {
		params.Filters.IncludeKeywords = parseCommaList(includeKeywords)
	}

	if excludeKeywords := c.Query("exclude_keywords"); excludeKeywords != "" {
		params.Filters.ExcludeKeywords = parseCommaList(excludeKeywords)
	}

	// 解析大小限制
	if minSizeStr := c.Query("min_size"); minSizeStr != "" {
		minSize, err := strconv.ParseInt(minSizeStr, 10, 64)
		if err != nil {
			return nil, err
		}
		params.Filters.MinSize = minSize
	}

	if maxSizeStr := c.Query("max_size"); maxSizeStr != "" {
		maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
		if err != nil {
			return nil, err
		}
		params.Filters.MaxSize = maxSize
	}

	// 解析做种数限制
	if minSeedersStr := c.Query("min_seeders"); minSeedersStr != "" {
		minSeeders, err := strconv.Atoi(minSeedersStr)
		if err != nil {
			return nil, err
		}
		params.Filters.MinSeeders = minSeeders
	}

	return params, nil
}

// parseCommaList 解析逗号分隔的列表
func parseCommaList(s string) []string {
	var result []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
