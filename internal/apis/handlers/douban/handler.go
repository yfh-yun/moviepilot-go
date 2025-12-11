package douban

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	doubanservice "moviepilot-go/internal/business/services/douban"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/pkg/logger"
)

// Handler 豆瓣 API 处理器
type Handler struct {
	doubanService *doubanservice.DoubanService
	logger        *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(doubanService *doubanservice.DoubanService) *Handler {
	return &Handler{
		doubanService: doubanService,
		logger:        logger.GetLogger(),
	}
}

// GetMovieTop250 获取豆瓣电影TOP250
// @Summary 获取豆瓣电影TOP250
// @Description 获取豆瓣电影TOP250榜单
// @Tags douban
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.MediaInfo
// @Router /api/douban/movie/top250 [get]
func (h *Handler) GetMovieTop250(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	results, err := h.doubanService.MovieTop250(c.Request.Context(), page, count)
	if err != nil {
		h.logger.Error("获取TOP250失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("page", page),
			zap.Int("count", count),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取TOP250成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("page", page),
		zap.Int("count", count),
	)

	c.JSON(http.StatusOK, results)
}

// GetMovieShowing 获取正在上映
// @Summary 获取正在上映的电影
// @Description 获取正在上映的电影列表
// @Tags douban
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.MediaInfo
// @Router /api/douban/movie/showing [get]
func (h *Handler) GetMovieShowing(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	results, err := h.doubanService.MovieShowing(c.Request.Context(), page, count)
	if err != nil {
		h.logger.Error("获取正在上映失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("page", page),
			zap.Int("count", count),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取正在上映成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("page", page),
		zap.Int("count", count),
	)

	c.JSON(http.StatusOK, results)
}

// GetTVWeeklyChinese 获取本周中国剧集榜
// @Summary 获取本周中国剧集榜
// @Description 获取本周中国剧集排行榜
// @Tags douban
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.MediaInfo
// @Router /api/douban/tv/weekly_chinese [get]
func (h *Handler) GetTVWeeklyChinese(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	results, err := h.doubanService.TVWeeklyChinese(c.Request.Context(), page, count)
	if err != nil {
		h.logger.Error("获取中国剧集榜失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("page", page),
			zap.Int("count", count),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取中国剧集榜成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("page", page),
		zap.Int("count", count),
	)

	c.JSON(http.StatusOK, results)
}

// GetTVWeeklyGlobal 获取本周全球剧集榜
// @Summary 获取本周全球剧集榜
// @Description 获取本周全球剧集排行榜
// @Tags douban
// @Produce json
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.MediaInfo
// @Router /api/douban/tv/weekly_global [get]
func (h *Handler) GetTVWeeklyGlobal(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	results, err := h.doubanService.TVWeeklyGlobal(c.Request.Context(), page, count)
	if err != nil {
		h.logger.Error("获取全球剧集榜失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Int("page", page),
			zap.Int("count", count),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取全球剧集榜成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.Int("page", page),
		zap.Int("count", count),
	)

	c.JSON(http.StatusOK, results)
}

// GetPersonDetail 获取人物详情
// @Summary 获取人物详情
// @Description 获取影人详细信息
// @Tags douban
// @Produce json
// @Param person_id path int true "人物ID"
// @Success 200 {object} dto.MediaPerson
// @Failure 400 {object} map[string]interface{}
// @Router /api/douban/person/{person_id} [get]
func (h *Handler) GetPersonDetail(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	personIDStr := c.Param("person_id")
	personID, err := strconv.Atoi(personIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的人物ID"})
		return
	}

	result, err := h.doubanService.PersonDetail(c.Request.Context(), personID)
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
// @Description 获取影人参演的影视作品列表
// @Tags douban
// @Produce json
// @Param person_id path int true "人物ID"
// @Param page query int false "页码" default(1)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/douban/person/credits/{person_id} [get]
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

	result, err := h.doubanService.PersonCredits(c.Request.Context(), personID, page)
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

	c.JSON(http.StatusOK, result)
}

// GetCredits 获取豆瓣演员阵容
// @Summary 获取豆瓣演员阵容
// @Description 获取电影或电视剧的演员阵容
// @Tags douban
// @Produce json
// @Param doubanid path string true "豆瓣ID"
// @Param type_name path string true "媒体类型: 电影/电视剧"
// @Success 200 {array} dto.MediaPerson
// @Failure 400 {object} map[string]interface{}
// @Router /api/douban/credits/{doubanid}/{type_name} [get]
func (h *Handler) GetCredits(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	doubanID := c.Param("doubanid")
	typeName := c.Param("type_name")

	var result []*dto.MediaPerson
	var err error

	switch typeName {
	case "电影":
		result, err = h.doubanService.MovieCredits(c.Request.Context(), doubanID)
	case "电视剧":
		result, err = h.doubanService.TVCredits(c.Request.Context(), doubanID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型"})
		return
	}

	if err != nil {
		h.logger.Error("获取演员阵容失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("douban_id", doubanID),
			zap.String("type_name", typeName),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取演员阵容成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("douban_id", doubanID),
		zap.String("type_name", typeName),
	)

	c.JSON(http.StatusOK, result)
}

// GetRecommend 获取豆瓣推荐内容
// @Summary 获取豆瓣推荐内容
// @Description 获取电影或电视剧的推荐内容
// @Tags douban
// @Produce json
// @Param doubanid path string true "豆瓣ID"
// @Param type_name path string true "媒体类型: 电影/电视剧"
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/douban/recommend/{doubanid}/{type_name} [get]
func (h *Handler) GetRecommend(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	doubanID := c.Param("doubanid")
	typeName := c.Param("type_name")

	var result []*dto.MediaInfo
	var err error

	switch typeName {
	case "电影":
		result, err = h.doubanService.MovieRecommend(c.Request.Context(), doubanID)
	case "电视剧":
		result, err = h.doubanService.TVRecommend(c.Request.Context(), doubanID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的媒体类型"})
		return
	}

	if err != nil {
		h.logger.Error("获取推荐内容失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("douban_id", doubanID),
			zap.String("type_name", typeName),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("获取推荐内容成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("douban_id", doubanID),
		zap.String("type_name", typeName),
	)

	c.JSON(http.StatusOK, result)
}

// GetDoubanInfo 获取豆瓣媒体信息
// @Summary 获取豆瓣媒体信息
// @Description 根据豆瓣ID获取媒体详情
// @Tags douban
// @Produce json
// @Param doubanid path string true "豆瓣ID"
// @Success 200 {object} dto.MediaInfo
// @Router /api/douban/{doubanid} [get]
func (h *Handler) GetDoubanInfo(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	doubanID := c.Param("doubanid")

	result, err := h.doubanService.DoubanInfo(c.Request.Context(), doubanID)
	if err != nil {
		h.logger.Error("获取豆瓣媒体信息失败",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("douban_id", doubanID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result == nil {
		// 返回空的MediaInfo对象
		result = &dto.MediaInfo{}
	}

	h.logger.Info("获取豆瓣媒体信息成功",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("douban_id", doubanID),
	)

	c.JSON(http.StatusOK, result)
}
