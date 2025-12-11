package bangumi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	bangumiservice "moviepilot-go/internal/business/services/bangumi"
	"moviepilot-go/pkg/logger"
)

// Handler Bangumi API 处理器
type Handler struct {
	bangumiService *bangumiservice.BangumiService
	logger         *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(bangumiService *bangumiservice.BangumiService) *Handler {
	return &Handler{
		bangumiService: bangumiService,
		logger:         logger.GetLogger(),
	}
}

// GetSubjectDetail 获取条目详情
// @Summary 获取Bangumi条目详情
// @Description 根据Bangumi ID获取条目详细信息
// @Tags bangumi
// @Produce json
// @Param bangumi_id path int true "Bangumi ID"
// @Success 200 {object} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/bangumi/subject/{bangumi_id} [get]
func (h *Handler) GetSubjectDetail(c *gin.Context) {
	bangumiIDStr := c.Param("bangumi_id")
	bangumiID, err := strconv.Atoi(bangumiIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的Bangumi ID"})
		return
	}

	result, err := h.bangumiService.GetSubjectDetail(c.Request.Context(), bangumiID)
	if err != nil {
		h.logger.Error("获取条目详情失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Search 搜索
// @Summary 搜索Bangumi条目
// @Description 搜索动画、漫画等条目
// @Tags bangumi
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/bangumi/search [get]
func (h *Handler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	results, err := h.bangumiService.Search(c.Request.Context(), keyword)
	if err != nil {
		h.logger.Error("搜索失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetCalendar 获取每日放送
// @Summary 获取每日放送
// @Description 获取Bangumi每日放送时间表
// @Tags bangumi
// @Produce json
// @Success 200 {array} dto.MediaInfo
// @Router /api/bangumi/calendar [get]
func (h *Handler) GetCalendar(c *gin.Context) {
	results, err := h.bangumiService.GetCalendar(c.Request.Context())
	if err != nil {
		h.logger.Error("获取每日放送失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetCredits 获取演职员表
// @Summary 查询Bangumi演职员表
// @Description 根据Bangumi ID获取演职员表信息
// @Tags bangumi
// @Produce json
// @Param bangumi_id path int true "Bangumi ID"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.MediaPerson
// @Failure 400 {object} map[string]interface{}
// @Router /api/bangumi/credits/{bangumi_id} [get]
func (h *Handler) GetCredits(c *gin.Context) {
	bangumiIDStr := c.Param("bangumi_id")
	bangumiID, err := strconv.Atoi(bangumiIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的Bangumi ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	if page < 1 {
		page = 1
	}
	if count < 1 || count > 100 {
		count = 20
	}

	persons, err := h.bangumiService.GetCredits(c.Request.Context(), bangumiID)
	if err != nil {
		h.logger.Error("获取演职员表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 分页处理
	start := (page - 1) * count
	end := start + count
	if start >= len(persons) {
		c.JSON(http.StatusOK, []any{})
		return
	}
	if end > len(persons) {
		end = len(persons)
	}

	c.JSON(http.StatusOK, persons[start:end])
}

// GetRecommend 获取推荐
// @Summary 查询Bangumi推荐
// @Description 根据Bangumi ID获取相关推荐
// @Tags bangumi
// @Produce json
// @Param bangumi_id path int true "Bangumi ID"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/bangumi/recommend/{bangumi_id} [get]
func (h *Handler) GetRecommend(c *gin.Context) {
	bangumiIDStr := c.Param("bangumi_id")
	bangumiID, err := strconv.Atoi(bangumiIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的Bangumi ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	if page < 1 {
		page = 1
	}
	if count < 1 || count > 100 {
		count = 20
	}

	medias, err := h.bangumiService.GetRecommend(c.Request.Context(), bangumiID)
	if err != nil {
		h.logger.Error("获取推荐失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 分页处理
	start := (page - 1) * count
	end := start + count
	if start >= len(medias) {
		c.JSON(http.StatusOK, []any{})
		return
	}
	if end > len(medias) {
		end = len(medias)
	}

	c.JSON(http.StatusOK, medias[start:end])
}

// GetPersonDetail 获取人物详情
// @Summary 人物详情
// @Description 根据人物ID查询人物详情
// @Tags bangumi
// @Produce json
// @Param person_id path int true "人物ID"
// @Success 200 {object} dto.MediaPerson
// @Failure 400 {object} map[string]interface{}
// @Router /api/bangumi/person/{person_id} [get]
func (h *Handler) GetPersonDetail(c *gin.Context) {
	personIDStr := c.Param("person_id")
	personID, err := strconv.Atoi(personIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的人物ID"})
		return
	}

	person, err := h.bangumiService.GetPersonDetail(c.Request.Context(), personID)
	if err != nil {
		h.logger.Error("获取人物详情失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, person)
}

// GetPersonCredits 获取人物参演作品
// @Summary 人物参演作品
// @Description 根据人物ID查询人物参演作品
// @Tags bangumi
// @Produce json
// @Param person_id path int true "人物ID"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(20)
// @Success 200 {array} dto.MediaInfo
// @Failure 400 {object} map[string]interface{}
// @Router /api/bangumi/person/credits/{person_id} [get]
func (h *Handler) GetPersonCredits(c *gin.Context) {
	personIDStr := c.Param("person_id")
	personID, err := strconv.Atoi(personIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的人物ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))

	if page < 1 {
		page = 1
	}
	if count < 1 || count > 100 {
		count = 20
	}

	medias, err := h.bangumiService.GetPersonCredits(c.Request.Context(), personID)
	if err != nil {
		h.logger.Error("获取人物参演作品失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 分页处理
	start := (page - 1) * count
	end := start + count
	if start >= len(medias) {
		c.JSON(http.StatusOK, []any{})
		return
	}
	if end > len(medias) {
		end = len(medias)
	}

	c.JSON(http.StatusOK, medias[start:end])
}
