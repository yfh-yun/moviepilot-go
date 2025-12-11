package tmdb

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	tmdbservice "moviepilot-go/internal/business/services/tmdb"
	"moviepilot-go/internal/models/types"
	"moviepilot-go/pkg/logger"
)

// Handler TMDB API 处理器
type Handler struct {
	tmdbService *tmdbservice.TmdbService
	logger      *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(tmdbService *tmdbservice.TmdbService) *Handler {
	return &Handler{
		tmdbService: tmdbService,
		logger:      logger.GetLogger(),
	}
}

// Discover 发现媒体
// @Summary 发现媒体
// @Description 根据条件发现媒体内容
// @Tags tmdb
// @Security BearerAuth
// @Produce json
// @Param media_type query string true "媒体类型" Enums(movie, tv)
// @Param page query int false "页码" default(1)
// @Success 200 {array} dto.MediaInfo
// @Router /api/tmdb/discover [get]
func (h *Handler) Discover(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	mediaTypeStr := c.Query("media_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	var mediaType types.MediaType
	switch mediaTypeStr {
	case "movie":
		mediaType = types.MediaTypeMovie
	case "tv":
		mediaType = types.MediaTypeTV
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型"})
		return
	}

	params := make(map[string]any)
	params["page"] = page

	results, err := h.tmdbService.Discover(c.Request.Context(), mediaType, params)
	if err != nil {
		h.logger.Error("发现媒体失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("media_type", mediaTypeStr),
			zap.Int("page", page),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("发现媒体成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("media_type", mediaTypeStr),
		zap.Int("page", page),
	)

	c.JSON(http.StatusOK, results)
}

// Trending 获取热门
// @Summary 获取热门内容
// @Description 获取本周热门电影和电视剧
// @Tags tmdb
// @Produce json
// @Param page query int false "页码" default(1)
// @Success 200 {array} dto.MediaInfo
// @Router /api/tmdb/trending [get]
func (h *Handler) Trending(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	results, err := h.tmdbService.Trending(c.Request.Context(), page)
	if err != nil {
		h.logger.Error("获取热门失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("page", page),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取热门成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("page", page),
	)

	c.JSON(http.StatusOK, results)
}

// GetMovieDetail 获取电影详情
// @Summary 获取电影详情
// @Description 根据TMDB ID获取电影详细信息
// @Tags tmdb
// @Produce json
// @Param tmdb_id path int true "TMDB ID"
// @Success 200 {object} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/movie/{tmdb_id} [get]
func (h *Handler) GetMovieDetail(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdb_id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	result, err := h.tmdbService.GetMovieDetail(c.Request.Context(), tmdbID)
	if err != nil {
		h.logger.Error("获取电影详情失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取电影详情成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
	)

	c.JSON(http.StatusOK, result)
}

// GetTVDetail 获取电视剧详情
// @Summary 获取电视剧详情
// @Description 根据TMDB ID获取电视剧详细信息
// @Tags tmdb
// @Produce json
// @Param tmdb_id path int true "TMDB ID"
// @Success 200 {object} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/tv/{tmdb_id} [get]
func (h *Handler) GetTVDetail(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdb_id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	result, err := h.tmdbService.GetTVDetail(c.Request.Context(), tmdbID)
	if err != nil {
		h.logger.Error("获取电视剧详情失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取电视剧详情成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
	)

	c.JSON(http.StatusOK, result)
}

// Search 搜索
// @Summary 搜索媒体
// @Description 搜索电影、电视剧等内容
// @Tags tmdb
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param media_type query string false "媒体类型" Enums(movie, tv, multi) default(multi)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/search [get]
func (h *Handler) Search(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	mediaTypeStr := c.DefaultQuery("media_type", "multi")
	var mediaType types.MediaType
	switch mediaTypeStr {
	case "movie":
		mediaType = types.MediaTypeMovie
	case "tv":
		mediaType = types.MediaTypeTV
	default:
		mediaType = types.MediaTypeUnknown
	}

	results, err := h.tmdbService.Search(c.Request.Context(), keyword, mediaType)
	if err != nil {
		h.logger.Error("搜索失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("keyword", keyword),
			zap.String("media_type", mediaTypeStr),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("TMDB 搜索成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("keyword", keyword),
		zap.String("media_type", mediaTypeStr),
	)

	c.JSON(http.StatusOK, results)
}

// GetCredits 获取演职员表
// @Summary 获取演职员表
// @Description 获取电影或电视剧的演职员信息
// @Tags tmdb
// @Produce json
// @Param tmdb_id path int true "TMDB ID"
// @Param media_type query string true "媒体类型" Enums(movie, tv)
// @Success 200 {array} dto.MediaPerson
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/{tmdb_id}/credits [get]
func (h *Handler) GetCredits(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdb_id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	mediaTypeStr := c.Query("media_type")
	var mediaType types.MediaType
	switch mediaTypeStr {
	case "movie":
		mediaType = types.MediaTypeMovie
	case "tv":
		mediaType = types.MediaTypeTV
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型"})
		return
	}

	results, err := h.tmdbService.GetCredits(c.Request.Context(), tmdbID, mediaType)
	if err != nil {
		h.logger.Error("获取演职员表失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
			zap.String("media_type", mediaTypeStr),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取演职员表成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
		zap.String("media_type", mediaTypeStr),
	)

	c.JSON(http.StatusOK, results)
}

// GetRecommendations 获取推荐
// @Summary 获取推荐内容
// @Description 根据媒体ID获取相关推荐
// @Tags tmdb
// @Produce json
// @Param tmdb_id path int true "TMDB ID"
// @Param media_type query string true "媒体类型" Enums(movie, tv)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/{tmdb_id}/recommendations [get]
func (h *Handler) GetRecommendations(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdb_id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	mediaTypeStr := c.Query("media_type")
	var mediaType types.MediaType
	switch mediaTypeStr {
	case "movie":
		mediaType = types.MediaTypeMovie
	case "tv":
		mediaType = types.MediaTypeTV
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型"})
		return
	}

	results, err := h.tmdbService.GetRecommendations(c.Request.Context(), tmdbID, mediaType)
	if err != nil {
		h.logger.Error("获取推荐失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
			zap.String("media_type", mediaTypeStr),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取推荐成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
		zap.String("media_type", mediaTypeStr),
	)

	c.JSON(http.StatusOK, results)
}

// GetSeasons 获取所有季信息
// @Summary 获取所有季信息
// @Description 根据TMDB ID获取所有季信息
// @Tags tmdb
// @Produce json
// @Param tmdbid path int true "TMDB ID"
// @Success 200 {array} dto.TmdbSeason
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/seasons/{tmdbid} [get]
func (h *Handler) GetSeasons(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdbid")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	results, err := h.tmdbService.GetSeasons(c.Request.Context(), tmdbID)
	if err != nil {
		h.logger.Error("获取所有季信息失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取所有季信息成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
	)

	c.JSON(http.StatusOK, results)
}

// GetSimilar 获取类似媒体
// @Summary 获取类似媒体
// @Description 根据TMDB ID获取类似电影/电视剧
// @Tags tmdb
// @Produce json
// @Param tmdbid path int true "TMDB ID"
// @Param type_name path string true "媒体类型" Enums(电影, 电视剧)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/similar/{tmdbid}/{type_name} [get]
func (h *Handler) GetSimilar(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdbid")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	typeName := c.Param("type_name")
	var mediaType types.MediaType
	switch typeName {
	case "电影":
		mediaType = types.MediaTypeMovie
	case "电视剧":
		mediaType = types.MediaTypeTV
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型"})
		return
	}

	results, err := h.tmdbService.GetSimilar(c.Request.Context(), tmdbID, mediaType)
	if err != nil {
		h.logger.Error("获取类似媒体失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
			zap.String("type_name", typeName),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取类似媒体成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
		zap.String("type_name", typeName),
	)

	c.JSON(http.StatusOK, results)
}

// GetRecommendByType 获取推荐媒体
// @Summary 获取推荐媒体
// @Description 根据TMDB ID获取推荐电影/电视剧
// @Tags tmdb
// @Produce json
// @Param tmdbid path int true "TMDB ID"
// @Param type_name path string true "媒体类型" Enums(电影, 电视剧)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/recommend/{tmdbid}/{type_name} [get]
func (h *Handler) GetRecommendByType(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdbid")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	typeName := c.Param("type_name")
	var mediaType types.MediaType
	switch typeName {
	case "电影":
		mediaType = types.MediaTypeMovie
	case "电视剧":
		mediaType = types.MediaTypeTV
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型"})
		return
	}

	results, err := h.tmdbService.GetRecommendations(c.Request.Context(), tmdbID, mediaType)
	if err != nil {
		h.logger.Error("获取推荐媒体失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
			zap.String("type_name", typeName),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取推荐媒体成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
		zap.String("type_name", typeName),
	)

	c.JSON(http.StatusOK, results)
}

// GetCollection 获取系列合集
// @Summary 获取系列合集
// @Description 根据合集ID获取合集详情
// @Tags tmdb
// @Produce json
// @Param collection_id path int true "合集ID"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/collection/{collection_id} [get]
func (h *Handler) GetCollection(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	collectionIDStr := c.Param("collection_id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的合集ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	results, err := h.tmdbService.GetCollection(c.Request.Context(), collectionID)
	if err != nil {
		h.logger.Error("获取系列合集失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("collection_id", collectionID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 分页处理
	start := (page - 1) * count
	end := start + count
	if start >= len(results) {
		c.JSON(http.StatusOK, []any{})
		return
	}
	if end > len(results) {
		end = len(results)
	}

	h.logger.Info("获取系列合集成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("collection_id", collectionID),
		zap.Int("page", page),
		zap.Int("count", count),
	)

	c.JSON(http.StatusOK, results[start:end])
}

// GetPersonDetail 获取人物详情
// @Summary 获取人物详情
// @Description 根据人物ID获取人物详细信息
// @Tags tmdb
// @Produce json
// @Param person_id path int true "人物ID"
// @Success 200 {object} dto.MediaPerson
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/person/{person_id} [get]
func (h *Handler) GetPersonDetail(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	personIDStr := c.Param("person_id")
	personID, err := strconv.Atoi(personIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的人物ID"})
		return
	}

	result, err := h.tmdbService.GetPersonDetail(c.Request.Context(), personID)
	if err != nil {
		h.logger.Error("获取人物详情失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("person_id", personID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取人物详情成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("person_id", personID),
	)

	c.JSON(http.StatusOK, result)
}

// GetPersonCredits 获取人物参演作品
// @Summary 获取人物参演作品
// @Description 根据人物ID获取人物参演作品
// @Tags tmdb
// @Produce json
// @Param person_id path int true "人物ID"
// @Param page query int false "页码" default(1)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/person/credits/{person_id} [get]
func (h *Handler) GetPersonCredits(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	personIDStr := c.Param("person_id")
	personID, err := strconv.Atoi(personIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的人物ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	results, err := h.tmdbService.GetPersonCredits(c.Request.Context(), personID, page)
	if err != nil {
		h.logger.Error("获取人物参演作品失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("person_id", personID),
			zap.Int("page", page),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取人物参演作品成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("person_id", personID),
		zap.Int("page", page),
	)

	c.JSON(http.StatusOK, results)
}

// GetEpisodes 获取季的所有集
// @Summary 获取季的所有集
// @Description 根据TMDB ID和季号获取季的所有集信息
// @Tags tmdb
// @Produce json
// @Param tmdbid path int true "TMDB ID"
// @Param season path int true "季号"
// @Param episode_group query string false "剧集组"
// @Success 200 {array} dto.TmdbEpisode
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/{tmdbid}/{season} [get]
func (h *Handler) GetEpisodes(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdbid")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	seasonStr := c.Param("season")
	season, err := strconv.Atoi(seasonStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的季号"})
		return
	}

	episodeGroup := c.Query("episode_group")

	results, err := h.tmdbService.GetEpisodes(c.Request.Context(), tmdbID, season, episodeGroup)
	if err != nil {
		h.logger.Error("获取季的所有集失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
			zap.Int("season", season),
			zap.String("episode_group", episodeGroup),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取季的所有集成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
		zap.Int("season", season),
		zap.String("episode_group", episodeGroup),
	)

	c.JSON(http.StatusOK, results)
}

// GetCreditsByType 获取演员阵容
// @Summary 获取演员阵容
// @Description 根据TMDB ID和媒体类型获取演员阵容
// @Tags tmdb
// @Produce json
// @Param tmdbid path int true "TMDB ID"
// @Param type_name path string true "媒体类型" Enums(电影, 电视剧)
// @Param page query int false "页码" default(1)
// @Success 200 {array} dto.MediaPerson
// @Failure 400 {object} map[string]interface{}
// @Router /api/tmdb/credits/{tmdbid}/{type_name} [get]
func (h *Handler) GetCreditsByType(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	tmdbIDStr := c.Param("tmdbid")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的TMDB ID"})
		return
	}

	typeName := c.Param("type_name")
	var mediaType types.MediaType
	switch typeName {
	case "电影":
		mediaType = types.MediaTypeMovie
	case "电视剧":
		mediaType = types.MediaTypeTV
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型"})
		return
	}

	results, err := h.tmdbService.GetCredits(c.Request.Context(), tmdbID, mediaType)
	if err != nil {
		h.logger.Error("获取演员阵容失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("tmdb_id", tmdbID),
			zap.String("type_name", typeName),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取演员阵容成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("tmdb_id", tmdbID),
		zap.String("type_name", typeName),
	)

	c.JSON(http.StatusOK, results)
}
