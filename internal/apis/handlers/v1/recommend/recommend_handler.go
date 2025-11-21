// Package recommend Recommend API处理器模块
package recommend

import (
	"net/http"
	"strconv"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/business/services/recommend"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler Recommend API处理器
type Handler struct {
	service recommend.Service
	logger  *logger.Logger
}

// NewHandler 创建新的Recommend处理器
func NewHandler(service recommend.Service, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	recommendGroup := router.Group("/recommend")
	{
		recommendGroup.GET("/source", h.GetSource)
		recommendGroup.GET("/bangumi/calendar", h.BangumiCalendar)
		recommendGroup.GET("/douban/movies/showing", h.DoubanShowing)
		recommendGroup.GET("/douban/movies", h.DoubanMovies)
		recommendGroup.GET("/douban/tvs", h.DoubanTVs)
		recommendGroup.GET("/douban/movies/top250", h.DoubanMovieTop250)
		recommendGroup.GET("/douban/tvs/chinese", h.DoubanTVChinese)
		recommendGroup.GET("/douban/tvs/global", h.DoubanTVGlobal)
		recommendGroup.GET("/douban/tvs/animation", h.DoubanTVAnimation)
		recommendGroup.GET("/douban/movies/hot", h.DoubanMovieHot)
		recommendGroup.GET("/douban/tvs/hot", h.DoubanTVHot)
		recommendGroup.GET("/tmdb/movies", h.TMDbMovies)
		recommendGroup.GET("/tmdb/tvs", h.TMDbTVs)
		recommendGroup.GET("/tmdb/trending", h.TMDbTrending)
		recommendGroup.GET("/similar/:id", h.GetSimilarMedia)
		recommendGroup.GET("/preferences", h.GetPreferenceBasedRecommendations)
	}
}

// GetSource 获取推荐数据源
// @Summary 获取推荐数据源
// @Description 获取可用的推荐数据源列表
// @Tags 推荐
// @Produce json
// @Success 200 {object} Response{data=[]MediaSource}
// @Router /recommend/source [get]
func (h *Handler) GetSource(c *gin.Context) {
	sources, err := h.service.GetSources()
	if err != nil {
		h.logger.Error("Failed to get recommend sources", zap.Error(err))
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

// BangumiCalendar Bangumi每日放送
// @Summary Bangumi每日放送
// @Description 获取Bangumi每日放送内容
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/bangumi/calendar [get]
func (h *Handler) BangumiCalendar(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Page:  page,
		Count: count,
		Type:  "calendar",
	}

	medias, err := h.service.BangumiCalendar(params)
	if err != nil {
		h.logger.Error("Failed to get Bangumi calendar", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取Bangumi每日放送失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
	})
}

// DoubanShowing 豆瓣正在热映
// @Summary 豆瓣正在热映
// @Description 获取豆瓣正在热映的电影
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/movies/showing [get]
func (h *Handler) DoubanShowing(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Page:  page,
		Count: count,
		Type:  "showing",
	}

	movies, err := h.service.DoubanShowing(params)
	if err != nil {
		h.logger.Error("Failed to get Douban showing movies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣正在热映失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// DoubanMovies 豆瓣电影推荐
// @Summary 豆瓣电影推荐
// @Description 获取豆瓣电影推荐
// @Tags 推荐
// @Produce json
// @Param sort query string false "排序方式 (R=热度,T=时间,S=评分)" default(R)
// @Param tags query string false "标签"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/movies [get]
func (h *Handler) DoubanMovies(c *gin.Context) {
	sortParam := c.DefaultQuery("sort", "R")
	tagsParam := c.Query("tags")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Sort:  sortParam,
		Tags:  tagsParam,
		Page:  page,
		Count: count,
		Type:  "movie",
	}

	movies, err := h.service.DoubanMovies(params)
	if err != nil {
		h.logger.Error("Failed to get Douban movies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣电影推荐失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// DoubanTVs 豆瓣剧集推荐
// @Summary 豆瓣剧集推荐
// @Description 获取豆瓣剧集推荐
// @Tags 推荐
// @Produce json
// @Param sort query string false "排序方式 (R=热度,T=时间,S=评分)" default(R)
// @Param tags query string false "标签"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/tvs [get]
func (h *Handler) DoubanTVs(c *gin.Context) {
	sortParam := c.DefaultQuery("sort", "R")
	tagsParam := c.Query("tags")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Sort:  sortParam,
		Tags:  tagsParam,
		Page:  page,
		Count: count,
		Type:  "tv",
	}

	tvs, err := h.service.DoubanTVs(params)
	if err != nil {
		h.logger.Error("Failed to get Douban TVs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣剧集推荐失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvs,
	})
}

// DoubanMovieTop250 豆瓣电影TOP250
// @Summary 豆瓣电影TOP250
// @Description 获取豆瓣电影TOP250
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/movies/top250 [get]
func (h *Handler) DoubanMovieTop250(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Page:  page,
		Count: count,
		Type:  "top250",
	}

	movies, err := h.service.DoubanMovieTop250(params)
	if err != nil {
		h.logger.Error("Failed to get Douban movie top250", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣电影TOP250失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// DoubanTVChinese 豆瓣国产剧集周榜
// @Summary 豆瓣国产剧集周榜
// @Description 获取豆瓣国产剧集周榜
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/tvs/chinese [get]
func (h *Handler) DoubanTVChinese(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Page:  page,
		Count: count,
		Type:  "chinese",
	}

	tvs, err := h.service.DoubanTVChinese(params)
	if err != nil {
		h.logger.Error("Failed to get Douban Chinese TVs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣国产剧集周榜失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvs,
	})
}

// DoubanTVGlobal 豆瓣全球剧集周榜
// @Summary 豆瓣全球剧集周榜
// @Description 获取豆瓣全球剧集周榜
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/tvs/global [get]
func (h *Handler) DoubanTVGlobal(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Page:  page,
		Count: count,
		Type:  "global",
	}

	tvs, err := h.service.DoubanTVGlobal(params)
	if err != nil {
		h.logger.Error("Failed to get Douban global TVs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣全球剧集周榜失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvs,
	})
}

// DoubanTVAnimation 豆瓣动画剧集推荐
// @Summary 豆瓣动画剧集推荐
// @Description 获取豆瓣动画剧集推荐
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/tvs/animation [get]
func (h *Handler) DoubanTVAnimation(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Page:  page,
		Count: count,
		Type:  "animation",
	}

	tvs, err := h.service.DoubanTVAnimation(params)
	if err != nil {
		h.logger.Error("Failed to get Douban animation TVs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣动画剧集推荐失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvs,
	})
}

// DoubanMovieHot 豆瓣热门电影
// @Summary 豆瓣热门电影
// @Description 获取豆瓣热门电影推荐
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/movies/hot [get]
func (h *Handler) DoubanMovieHot(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Page:  page,
		Count: count,
		Type:  "hot",
	}

	movies, err := h.service.DoubanMovieHot(params)
	if err != nil {
		h.logger.Error("Failed to get Douban hot movies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣热门电影失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// DoubanTVHot 豆瓣热门电视剧
// @Summary 豆瓣热门电视剧
// @Description 获取豆瓣热门电视剧推荐
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/douban/tvs/hot [get]
func (h *Handler) DoubanTVHot(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := recommend.RecommendParams{
		Page:  page,
		Count: count,
		Type:  "hot",
	}

	tvs, err := h.service.DoubanTVHot(params)
	if err != nil {
		h.logger.Error("Failed to get Douban hot TVs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣热门电视剧失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvs,
	})
}

// TMDbMovies TMDB电影推荐
// @Summary TMDB电影推荐
// @Description 获取TMDB电影推荐
// @Tags 推荐
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
// @Router /recommend/tmdb/movies [get]
func (h *Handler) TMDbMovies(c *gin.Context) {
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

	params := recommend.TMDbRecommendParams{
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

	movies, err := h.service.TMDbMovies(params)
	if err != nil {
		h.logger.Error("Failed to get TMDb movies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取TMDB电影推荐失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
	})
}

// TMDbTVs TMDB剧集推荐
// @Summary TMDB剧集推荐
// @Description 获取TMDB剧集推荐
// @Tags 推荐
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
// @Router /recommend/tmdb/tvs [get]
func (h *Handler) TMDbTVs(c *gin.Context) {
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

	params := recommend.TMDbRecommendParams{
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

	tvs, err := h.service.TMDbTVs(params)
	if err != nil {
		h.logger.Error("Failed to get TMDb TVs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取TMDB剧集推荐失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvs,
	})
}

// TMDbTrending TMDB流行趋势
// @Summary TMDB流行趋势
// @Description 获取TMDB流行趋势内容
// @Tags 推荐
// @Produce json
// @Param page query int false "页码" default(1)
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/tmdb/trending [get]
func (h *Handler) TMDbTrending(c *gin.Context) {
	pageParam := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageParam)

	params := recommend.RecommendParams{
		Page: page,
		Type: "trending",
	}

	medias, err := h.service.TMDbTrending(params)
	if err != nil {
		h.logger.Error("Failed to get TMDb trending", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取TMDB流行趋势失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
	})
}

// GetSimilarMedia 获取相似媒体
// @Summary 获取相似媒体
// @Description 根据媒体ID获取相似的内容推荐
// @Tags 推荐
// @Produce json
// @Param id path string true "媒体ID"
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/similar/{id} [get]
func (h *Handler) GetSimilarMedia(c *gin.Context) {
	mediaID := c.Param("id")
	if mediaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "媒体ID不能为空",
		})
		return
	}

	medias, err := h.service.GetSimilarMedia(mediaID)
	if err != nil {
		h.logger.Error("Failed to get similar media", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取相似媒体失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
	})
}

// GetPreferenceBasedRecommendations 基于偏好的推荐
// @Summary 基于偏好的推荐
// @Description 根据用户偏好获取个性化推荐
// @Tags 推荐
// @Produce json
// @Success 200 {object} Response{data=[]MediaInfo}
// @Router /recommend/preferences [get]
func (h *Handler) GetPreferenceBasedRecommendations(c *gin.Context) {
	// TODO: 从上下文中获取用户信息
	userID := "" // 从上下文中获取

	medias, err := h.service.GetPreferenceBasedRecommendations(userID)
	if err != nil {
		h.logger.Error("Failed to get preference-based recommendations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取基于偏好的推荐失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    medias,
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
