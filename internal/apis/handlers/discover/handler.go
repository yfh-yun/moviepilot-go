package discover

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	"moviepilot-go/internal/business/services/bangumi"
	"moviepilot-go/internal/business/services/douban"
	"moviepilot-go/internal/business/services/tmdb"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/models/types"
	"moviepilot-go/pkg/logger"
)

// Handler Discover API 处理器
type Handler struct {
	bangumiService *bangumi.BangumiService
	doubanService  *douban.DoubanService
	tmdbService    *tmdb.TmdbService
	logger         *zap.Logger
}

// NewHandler 创建Discover处理器
func NewHandler(
	bangumiService *bangumi.BangumiService,
	doubanService *douban.DoubanService,
	tmdbService *tmdb.TmdbService,
) *Handler {
	return &Handler{
		bangumiService: bangumiService,
		doubanService:  doubanService,
		tmdbService:    tmdbService,
		logger:         logger.GetLogger(),
	}
}

// GetSource 获取探索数据源
// @Summary 获取探索数据源
// @Description 获取探索数据源
// @Tags discover
// @Produce json
// @Success 200 {array} dto.DiscoverMediaSource
// @Router /api/discover/source [get]
func (h *Handler) GetSource(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	h.logger.Debug("GetSource called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现事件系统，获取额外的探索数据源
	// 当前返回空列表
	c.JSON(http.StatusOK, []dto.DiscoverMediaSource{})
}

// DiscoverBangumi 探索Bangumi
// @Summary 探索Bangumi
// @Description 根据条件探索Bangumi内容
// @Tags discover
// @Produce json
// @Param type query int false "类型" default(2)
// @Param cat query int false "分类"
// @Param sort query string false "排序" default("rank")
// @Param year query string false "年份"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {array} dto.MediaInfo
// @Router /api/discover/bangumi [get]
func (h *Handler) DiscoverBangumi(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	// 获取查询参数
	typeStr := c.DefaultQuery("type", "2")
	catStr := c.DefaultQuery("cat", "0")
	sort := c.DefaultQuery("sort", "rank")
	year := c.DefaultQuery("year", "")
	pageStr := c.DefaultQuery("page", "1")
	countStr := c.DefaultQuery("count", "30")

	// 转换参数类型
	typeVal, _ := strconv.Atoi(typeStr)
	cat, _ := strconv.Atoi(catStr)
	page, _ := strconv.Atoi(pageStr)
	count, _ := strconv.Atoi(countStr)

	h.logger.Debug("DiscoverBangumi called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("type", typeVal),
		zap.Int("cat", cat),
		zap.String("sort", sort),
		zap.String("year", year),
		zap.Int("page", page),
		zap.Int("count", count),
	)

	// 计算偏移量
	offset := (page - 1) * count

	// 调用Bangumi服务探索内容
	medias, err := h.bangumiService.Discover(c.Request.Context(), typeVal, cat, sort, year, count, offset)
	if err != nil {
		h.logger.Error("DiscoverBangumi failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to discover Bangumi"})
		return
	}

	c.JSON(http.StatusOK, medias)
}

// DiscoverDoubanMovies 探索豆瓣电影
// @Summary 探索豆瓣电影
// @Description 浏览豆瓣电影信息
// @Tags discover
// @Produce json
// @Param sort query string false "排序" default("R")
// @Param tags query string false "标签"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {array} dto.MediaInfo
// @Router /api/discover/douban_movies [get]
func (h *Handler) DiscoverDoubanMovies(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	// 获取查询参数
	sort := c.DefaultQuery("sort", "R")
	tags := c.DefaultQuery("tags", "")
	pageStr := c.DefaultQuery("page", "1")
	countStr := c.DefaultQuery("count", "30")

	// 转换参数类型
	page, _ := strconv.Atoi(pageStr)
	count, _ := strconv.Atoi(countStr)

	h.logger.Debug("DiscoverDoubanMovies called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("sort", sort),
		zap.String("tags", tags),
		zap.Int("page", page),
		zap.Int("count", count),
	)

	// 调用Douban服务探索电影
	movies, err := h.doubanService.Discover(c.Request.Context(), "movie", sort, tags, page, count)
	if err != nil {
		h.logger.Error("DiscoverDoubanMovies failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to discover Douban movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
}

// DiscoverDoubanTVs 探索豆瓣剧集
// @Summary 探索豆瓣剧集
// @Description 浏览豆瓣剧集信息
// @Tags discover
// @Produce json
// @Param sort query string false "排序" default("R")
// @Param tags query string false "标签"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {array} dto.MediaInfo
// @Router /api/discover/douban_tvs [get]
func (h *Handler) DiscoverDoubanTVs(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	// 获取查询参数
	sort := c.DefaultQuery("sort", "R")
	tags := c.DefaultQuery("tags", "")
	pageStr := c.DefaultQuery("page", "1")
	countStr := c.DefaultQuery("count", "30")

	// 转换参数类型
	page, _ := strconv.Atoi(pageStr)
	count, _ := strconv.Atoi(countStr)

	h.logger.Debug("DiscoverDoubanTVs called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("sort", sort),
		zap.String("tags", tags),
		zap.Int("page", page),
		zap.Int("count", count),
	)

	// 调用Douban服务探索剧集
	tvs, err := h.doubanService.Discover(c.Request.Context(), "tv", sort, tags, page, count)
	if err != nil {
		h.logger.Error("DiscoverDoubanTVs failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to discover Douban TVs"})
		return
	}

	c.JSON(http.StatusOK, tvs)
}

// DiscoverTmdbMovies 探索TMDB电影
// @Summary 探索TMDB电影
// @Description 浏览TMDB电影信息
// @Tags discover
// @Produce json
// @Param sort_by query string false "排序" default("popularity.desc")
// @Param with_genres query string false "类型ID"
// @Param with_original_language query string false "原始语言"
// @Param with_keywords query string false "关键词ID"
// @Param with_watch_providers query string false "播放平台ID"
// @Param vote_average query float false "最低评分" default(0.0)
// @Param vote_count query int false "最低投票数" default(0)
// @Param release_date query string false "发布日期"
// @Param page query int false "页码" default(1)
// @Success 200 {array} dto.MediaInfo
// @Router /api/discover/tmdb_movies [get]
func (h *Handler) DiscoverTmdbMovies(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	// 获取查询参数
	sortBy := c.DefaultQuery("sort_by", "popularity.desc")
	withGenres := c.DefaultQuery("with_genres", "")
	withOriginalLanguage := c.DefaultQuery("with_original_language", "")
	withKeywords := c.DefaultQuery("with_keywords", "")
	withWatchProviders := c.DefaultQuery("with_watch_providers", "")
	voteAverageStr := c.DefaultQuery("vote_average", "0.0")
	voteCountStr := c.DefaultQuery("vote_count", "0")
	releaseDate := c.DefaultQuery("release_date", "")
	pageStr := c.DefaultQuery("page", "1")

	// 转换参数类型
	voteAverage, _ := strconv.ParseFloat(voteAverageStr, 64)
	voteCount, _ := strconv.Atoi(voteCountStr)
	page, _ := strconv.Atoi(pageStr)

	// 构建查询参数
	params := map[string]any{
		"sort_by":                sortBy,
		"with_genres":            withGenres,
		"with_original_language": withOriginalLanguage,
		"with_keywords":          withKeywords,
		"with_watch_providers":   withWatchProviders,
		"vote_average.gte":       voteAverage,
		"vote_count.gte":         voteCount,
		"release_date.gte":       releaseDate,
		"page":                   page,
	}

	h.logger.Debug("DiscoverTmdbMovies called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Any("params", params),
	)

	// 调用TMDB服务探索电影
	movies, err := h.tmdbService.Discover(c.Request.Context(), types.MediaTypeMovie, params)
	if err != nil {
		h.logger.Error("DiscoverTmdbMovies failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to discover TMDB movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
}

// DiscoverTmdbTVs 探索TMDB剧集
// @Summary 探索TMDB剧集
// @Description 浏览TMDB剧集信息
// @Tags discover
// @Produce json
// @Param sort_by query string false "排序" default("popularity.desc")
// @Param with_genres query string false "类型ID"
// @Param with_original_language query string false "原始语言"
// @Param with_keywords query string false "关键词ID"
// @Param with_watch_providers query string false "播放平台ID"
// @Param vote_average query float false "最低评分" default(0.0)
// @Param vote_count query int false "最低投票数" default(0)
// @Param release_date query string false "发布日期"
// @Param page query int false "页码" default(1)
// @Success 200 {array} dto.MediaInfo
// @Router /api/discover/tmdb_tvs [get]
func (h *Handler) DiscoverTmdbTVs(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	// 获取查询参数
	sortBy := c.DefaultQuery("sort_by", "popularity.desc")
	withGenres := c.DefaultQuery("with_genres", "")
	withOriginalLanguage := c.DefaultQuery("with_original_language", "")
	withKeywords := c.DefaultQuery("with_keywords", "")
	withWatchProviders := c.DefaultQuery("with_watch_providers", "")
	voteAverageStr := c.DefaultQuery("vote_average", "0.0")
	voteCountStr := c.DefaultQuery("vote_count", "0")
	releaseDate := c.DefaultQuery("release_date", "")
	pageStr := c.DefaultQuery("page", "1")

	// 转换参数类型
	voteAverage, _ := strconv.ParseFloat(voteAverageStr, 64)
	voteCount, _ := strconv.Atoi(voteCountStr)
	page, _ := strconv.Atoi(pageStr)

	// 构建查询参数
	params := map[string]any{
		"sort_by":                sortBy,
		"with_genres":            withGenres,
		"with_original_language": withOriginalLanguage,
		"with_keywords":          withKeywords,
		"with_watch_providers":   withWatchProviders,
		"vote_average.gte":       voteAverage,
		"vote_count.gte":         voteCount,
		"first_air_date.gte":     releaseDate,
		"page":                   page,
	}

	h.logger.Debug("DiscoverTmdbTVs called",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Any("params", params),
	)

	// 调用TMDB服务探索剧集
	tvs, err := h.tmdbService.Discover(c.Request.Context(), types.MediaTypeTV, params)
	if err != nil {
		h.logger.Error("DiscoverTmdbTVs failed",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to discover TMDB TVs"})
		return
	}

	c.JSON(http.StatusOK, tvs)
}
