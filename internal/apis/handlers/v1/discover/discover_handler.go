// Package discover Discover API处理器模块
package discover

import (
	"net/http"
	"strconv"

	"moviepilot-go/internal/business/services/discover"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler Discover API处理器
// handler.go
// @version 1.0.0
// @title MoviePilot Discover API
// @description 探索和发现媒体内容的API接口
// @contact.name MoviePilot Team
// @contact.url https://github.com/moviepilot
// @contact.email support@moviepilot.org
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:3001
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 输入"Bearer <your-token>"
type Handler struct {
	service discover.Service
	logger  *zap.Logger
}

// NewHandler 创建新的Discover处理器
func NewHandler(service discover.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	discoverGroup := router.Group("/discover")
	{
		discoverGroup.GET("/source", h.GetSource)
		discoverGroup.GET("/bangumi", h.DiscoverBangumi)
		discoverGroup.GET("/douban/movies", h.DiscoverDoubanMovies)
		discoverGroup.GET("/douban/tvs", h.DiscoverDoubanTVs)
		discoverGroup.GET("/tmdb/movies", h.DiscoverTMDbMovies)
		discoverGroup.GET("/tmdb/tvs", h.DiscoverTMDbTVs)
		discoverGroup.GET("/trending", h.GetTrendingMedia)
		discoverGroup.GET("/popular", h.GetPopularMedia)
		discoverGroup.GET("/new", h.GetNewMedia)
		discoverGroup.GET("/random", h.GetRandomMedia)
	}
}

// GetSource 获取探索数据源
// @Summary 获取探索数据源
// @Description 获取可用的探索数据源列表
// @Tags 发现
// @Produce json
// @Success 200 {object} Response{data=[]MediaSource}
// @Router /discover/source [get]
func (h *Handler) GetSource(c *gin.Context) {
	sources, err := h.service.GetSources()
	if err != nil {
		h.logger.Error("Failed to get discover sources", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取数据源失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sources,
	})
}

// DiscoverBangumi 探索Bangumi
// @Summary 探索Bangumi
// @Description 从Bangumi平台探索动漫内容
// @Tags 发现
// @Produce json
// @Param type query int false "类型 (1=电影, 2=剧集, 3=动画)" default(2)
// @Param cat query int false "分类ID"
// @Param sort query string false "排序方式 (rank,date,score)" default(rank)
// @Param year query string false "年份"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/bangumi [get]
func (h *Handler) DiscoverBangumi(c *gin.Context) {
	typeParam := c.DefaultQuery("type", "2")
	catParam := c.Query("cat")
	sortParam := c.DefaultQuery("sort", "rank")
	yearParam := c.Query("year")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	mediaType, _ := strconv.Atoi(typeParam)
	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := discover.DiscoverParams{
		Type:     mediaType,
		Category: catParam,
		Sort:     sortParam,
		Year:     yearParam,
		Page:     page,
		Count:    count,
	}

	medias, err := h.service.DiscoverBangumi(params)
	if err != nil {
		h.logger.Error("Failed to discover Bangumi", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "探索Bangumi失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
	})
}

// DiscoverDoubanMovies 探索豆瓣电影
// @Summary 探索豆瓣电影
// @Description 从豆瓣平台探索电影内容
// @Tags 发现
// @Produce json
// @Param sort query string false "排序方式 (R=热度,T=时间,S=评分)" default(R)
// @Param tags query string false "标签"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/douban/movies [get]
func (h *Handler) DiscoverDoubanMovies(c *gin.Context) {
	sortParam := c.DefaultQuery("sort", "R")
	tagsParam := c.Query("tags")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := discover.DiscoverParams{
		Sort:  sortParam,
		Tags:  tagsParam,
		Page:  page,
		Count: count,
		Type:  1, // 电影类型
	}

	movies, err := h.service.DiscoverDoubanMovies(params)
	if err != nil {
		h.logger.Error("Failed to discover Douban movies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "探索豆瓣电影失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// DiscoverDoubanTVs 探索豆瓣剧集
// @Summary 探索豆瓣剧集
// @Description 从豆瓣平台探索剧集内容
// @Tags 发现
// @Produce json
// @Param sort query string false "排序方式 (R=热度,T=时间,S=评分)" default(R)
// @Param tags query string false "标签"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/douban/tvs [get]
func (h *Handler) DiscoverDoubanTVs(c *gin.Context) {
	sortParam := c.DefaultQuery("sort", "R")
	tagsParam := c.Query("tags")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := discover.DiscoverParams{
		Sort:  sortParam,
		Tags:  tagsParam,
		Page:  page,
		Count: count,
		Type:  2, // 剧集类型
	}

	tvs, err := h.service.DiscoverDoubanTVs(params)
	if err != nil {
		h.logger.Error("Failed to discover Douban TVs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "探索豆瓣剧集失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvs,
	})
}

// DiscoverTMDbMovies 探索TMDB电影
// @Summary 探索TMDB电影
// @Description 从TMDB平台探索电影内容
// @Tags 发现
// @Produce json
// @Param sort_by query string false "排序方式" default(popularity.desc)
// @Param with_genres query string false "类型ID"
// @Param with_original_language query string false "原始语言"
// @Param with_keywords query string false "关键词"
// @Param with_watch_providers query string false "观看提供商"
// @Param vote_average query float false "评分最低值" default(0.0)
// @Param vote_count query int false "评分人数最低值" default(0)
// @Param release_date query string false "发布日期"
// @Param page query int false "页码" default(1)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/tmdb/movies [get]
func (h *Handler) DiscoverTMDbMovies(c *gin.Context) {
	sortBy := c.DefaultQuery("sort_by", "popularity.desc")
	withGenres := c.Query("with_genres")
	withOriginalLanguage := c.Query("with_original_language")
	withKeywords := c.Query("with_keywords")
	withWatchProviders := c.Query("with_watch_providers")
	voteAverageParam := c.DefaultQuery("vote_average", "0.0")
	voteCountParam := c.DefaultQuery("vote_count", "0")
	releaseDate := c.Query("release_date")
	pageParam := c.DefaultQuery("page", "1")

	voteAverage, _ := strconv.ParseFloat(voteAverageParam, 64)
	voteCount, _ := strconv.Atoi(voteCountParam)
	page, _ := strconv.Atoi(pageParam)

	params := discover.TMDbDiscoverParams{
		SortBy:               sortBy,
		WithGenres:           withGenres,
		WithOriginalLanguage: withOriginalLanguage,
		WithKeywords:         withKeywords,
		WithWatchProviders:   withWatchProviders,
		VoteAverage:          voteAverage,
		VoteCount:            voteCount,
		ReleaseDate:          releaseDate,
		Page:                 page,
	}

	movies, err := h.service.DiscoverTMDbMovies(params)
	if err != nil {
		h.logger.Error("Failed to discover TMDb movies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "探索TMDB电影失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// DiscoverTMDbTVs 探索TMDB剧集
// @Summary 探索TMDB剧集
// @Description 从TMDB平台探索剧集内容
// @Tags 发现
// @Produce json
// @Param sort_by query string false "排序方式" default(popularity.desc)
// @Param with_genres query string false "类型ID"
// @Param with_original_language query string false "原始语言"
// @Param with_keywords query string false "关键词"
// @Param with_watch_providers query string false "观看提供商"
// @Param vote_average query float false "评分最低值" default(0.0)
// @Param vote_count query int false "评分人数最低值" default(0)
// @Param release_date query string false "发布日期"
// @Param page query int false "页码" default(1)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/tmdb/tvs [get]
func (h *Handler) DiscoverTMDbTVs(c *gin.Context) {
	sortBy := c.DefaultQuery("sort_by", "popularity.desc")
	withGenres := c.Query("with_genres")
	withOriginalLanguage := c.Query("with_original_language")
	withKeywords := c.Query("with_keywords")
	withWatchProviders := c.Query("with_watch_providers")
	voteAverageParam := c.DefaultQuery("vote_average", "0.0")
	voteCountParam := c.DefaultQuery("vote_count", "0")
	releaseDate := c.Query("release_date")
	pageParam := c.DefaultQuery("page", "1")

	voteAverage, _ := strconv.ParseFloat(voteAverageParam, 64)
	voteCount, _ := strconv.Atoi(voteCountParam)
	page, _ := strconv.Atoi(pageParam)

	params := discover.TMDbDiscoverParams{
		SortBy:               sortBy,
		WithGenres:           withGenres,
		WithOriginalLanguage: withOriginalLanguage,
		WithKeywords:         withKeywords,
		WithWatchProviders:   withWatchProviders,
		VoteAverage:          voteAverage,
		VoteCount:            voteCount,
		ReleaseDate:          releaseDate,
		Page:                 page,
	}

	tvs, err := h.service.DiscoverTMDbTVs(params)
	if err != nil {
		h.logger.Error("Failed to discover TMDb TVs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "探索TMDB剧集失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvs,
	})
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// MediaSource 媒体数据源结构
type MediaSource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// MediaInfo 媒体信息结构
type MediaInfo struct {
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	OriginalTitle string   `json:"original_title,omitempty"`
	Type          string   `json:"type"`
	Year          int      `json:"year,omitempty"`
	Rating        float64  `json:"rating,omitempty"`
	VoteCount     int      `json:"vote_count,omitempty"`
	Poster        string   `json:"poster,omitempty"`
	Backdrop      string   `json:"backdrop,omitempty"`
	Overview      string   `json:"overview,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Countries     []string `json:"countries,omitempty"`
	Languages     []string `json:"languages,omitempty"`
	ReleaseDate   string   `json:"release_date,omitempty"`
	Runtime       int      `json:"runtime,omitempty"`
	Popularity    float64  `json:"popularity,omitempty"`
	Source        string   `json:"source"`
	SourceURL     string   `json:"source_url,omitempty"`
}

// GetTrendingMedia 获取热门媒体内容
// @Summary 获取热门媒体内容
// @Description 获取当前热门的媒体内容，基于多种数据源的热度指标
// @Tags 发现
// @Produce json
// @Param type query string false "媒体类型 (movie,tv,all)" default(all)
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Param time_window query string false "时间窗口 (day,week)" default(day)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/trending [get]
func (h *Handler) GetTrendingMedia(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "all")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")
	timeWindow := c.DefaultQuery("time_window", "day")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := discover.TrendingParams{
		Type:       mediaType,
		Page:       page,
		Count:      count,
		TimeWindow: timeWindow,
	}

	medias, err := h.service.GetTrendingMedia(params)
	if err != nil {
		h.logger.Error("Failed to get trending media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取热门内容失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
	})
}

// GetPopularMedia 获取流行媒体内容
// @Summary 获取流行媒体内容
// @Description 获取当前流行的媒体内容，基于综合评分和流行度
// @Tags 发现
// @Produce json
// @Param type query string false "媒体类型 (movie,tv,all)" default(all)
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Param region query string false "地区代码" default(US)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/popular [get]
func (h *Handler) GetPopularMedia(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "all")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")
	region := c.DefaultQuery("region", "US")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := discover.PopularParams{
		Type:   mediaType,
		Page:   page,
		Count:  count,
		Region: region,
	}

	medias, err := h.service.GetPopularMedia(params)
	if err != nil {
		h.logger.Error("Failed to get popular media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取流行内容失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
	})
}

// GetNewMedia 获取最新媒体内容
// @Summary 获取最新媒体内容
// @Description 获取最新发布的媒体内容
// @Tags 发现
// @Produce json
// @Param type query string false "媒体类型 (movie,tv,all)" default(all)
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Param region query string false "地区代码" default(US)
// @Param primary_release_year query string false "主要发行年份"
// @Param first_air_date_year query string false "首次播出年份"
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/new [get]
func (h *Handler) GetNewMedia(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "all")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")
	region := c.DefaultQuery("region", "US")
	primaryReleaseYear := c.Query("primary_release_year")
	firstAirDateYear := c.Query("first_air_date_year")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := discover.NewMediaParams{
		Type:               mediaType,
		Page:               page,
		Count:              count,
		Region:             region,
		PrimaryReleaseYear: primaryReleaseYear,
		FirstAirDateYear:   firstAirDateYear,
	}

	medias, err := h.service.GetNewMedia(params)
	if err != nil {
		h.logger.Error("Failed to get new media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取最新内容失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
	})
}

// GetRandomMedia 获取随机媒体内容
// @Summary 获取随机媒体内容
// @Description 获取随机推荐的媒体内容
// @Tags 发现
// @Produce json
// @Param type query string false "媒体类型 (movie,tv,all)" default(all)
// @Param count query int false "数量" default(10)
// @Param genre query string false "类型ID"
// @Param min_rating query float false "最低评分" default(6.0)
// @Param min_votes query int false "最低评分人数" default(100)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /discover/random [get]
func (h *Handler) GetRandomMedia(c *gin.Context) {
	mediaType := c.DefaultQuery("type", "all")
	countParam := c.DefaultQuery("count", "10")
	genre := c.Query("genre")
	minRatingParam := c.DefaultQuery("min_rating", "6.0")
	minVotesParam := c.DefaultQuery("min_votes", "100")

	count, _ := strconv.Atoi(countParam)
	minRating, _ := strconv.ParseFloat(minRatingParam, 64)
	minVotes, _ := strconv.Atoi(minVotesParam)

	params := discover.RandomMediaParams{
		Type:      mediaType,
		Count:     count,
		Genre:     genre,
		MinRating: minRating,
		MinVotes:  minVotes,
	}

	medias, err := h.service.GetRandomMedia(params)
	if err != nil {
		h.logger.Error("Failed to get random media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取随机推荐失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
	})
}
