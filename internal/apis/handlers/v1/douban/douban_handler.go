// Package douban Douban API处理器模块
package douban

import (
	"net/http"
	"strconv"

	"moviepilot-go/internal/business/services/douban"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler Douban API处理器
type Handler struct {
	service douban.Service
	logger  *zap.Logger
}

// NewHandler 创建新的Douban处理器
func NewHandler(service douban.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	doubanGroup := router.Group("/douban")
	{
		// 豆瓣基础信息
		doubanGroup.GET("/movie/:id", h.GetDoubanMovie)
		doubanGroup.GET("/tv/:id", h.GetDoubanTV)
		doubanGroup.GET("/person/:id", h.GetDoubanPerson)
		doubanGroup.GET("/search", h.SearchDouban)

		// 用户相关
		doubanGroup.GET("/user/:id/wish", h.GetDoubanWish)
		doubanGroup.GET("/user/:id/do", h.GetDoubanDo)
		doubanGroup.GET("/user/:id/collect", h.GetDoubanCollect)

		// 人物详情和作品
		doubanGroup.GET("/person/:person_id", h.GetPersonDetail)
		doubanGroup.GET("/person/credits/:person_id", h.GetPersonCredits)

		// 演员阵容
		doubanGroup.GET("/credits/:doubanid/:type_name", h.GetDoubanCredits)

		// 推荐内容
		doubanGroup.GET("/recommend/:doubanid/:type_name", h.GetDoubanRecommend)
	}
}

// GetDoubanMovie 获取豆瓣电影信息
// @Summary 获取豆瓣电影信息
// @Description 根据豆瓣ID获取电影详细信息
// @Tags 豆瓣
// @Produce json
// @Param id path string true "豆瓣电影ID"
// @Success 200 {object} Response{data=MovieInfo}
// @Router /douban/movie/{id} [get]
func (h *Handler) GetDoubanMovie(c *gin.Context) {
	doubanID := c.Param("id")
	if doubanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "豆瓣ID不能为空",
		})
		return
	}

	movieInfo, err := h.service.GetMovieInfo(doubanID)
	if err != nil {
		h.logger.Error("Failed to get douban movie", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣电影信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movieInfo,
	})
}

// GetDoubanTV 获取豆瓣电视剧信息
// @Summary 获取豆瓣电视剧信息
// @Description 根据豆瓣ID获取电视剧详细信息
// @Tags 豆瓣
// @Produce json
// @Param id path string true "豆瓣电视剧ID"
// @Success 200 {object} Response{data=TVInfo}
// @Router /douban/tv/{id} [get]
func (h *Handler) GetDoubanTV(c *gin.Context) {
	doubanID := c.Param("id")
	if doubanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "豆瓣ID不能为空",
		})
		return
	}

	tvInfo, err := h.service.GetTVInfo(doubanID)
	if err != nil {
		h.logger.Error("Failed to get douban TV", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣电视剧信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tvInfo,
	})
}

// GetDoubanPerson 获取豆瓣人物信息
// @Summary 获取豆瓣人物信息
// @Description 根据豆瓣ID获取人物详细信息
// @Tags 豆瓣
// @Produce json
// @Param id path string true "豆瓣人物ID"
// @Success 200 {object} Response{data=PersonInfo}
// @Router /douban/person/{id} [get]
func (h *Handler) GetDoubanPerson(c *gin.Context) {
	doubanID := c.Param("id")
	if doubanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "豆瓣ID不能为空",
		})
		return
	}

	personInfo, err := h.service.GetPersonInfo(doubanID)
	if err != nil {
		h.logger.Error("Failed to get douban person", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣人物信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    personInfo,
	})
}

// SearchDouban 搜索豆瓣内容
// @Summary 搜索豆瓣内容
// @Description 搜索豆瓣电影、电视剧、人物
// @Tags 豆瓣
// @Produce json
// @Param q query string true "搜索关键词"
// @Param type query string false "搜索类型 (movie,tv,person)" default(movie)
// @Param page query int false "页码" default(1)
// @Success 200 {object} Response{data=SearchResult}
// @Router /douban/search [get]
func (h *Handler) SearchDouban(c *gin.Context) {
	query := c.Query("q")
	searchType := c.DefaultQuery("type", "movie")
	pageParam := c.DefaultQuery("page", "1")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "搜索关键词不能为空",
		})
		return
	}

	page, _ := strconv.Atoi(pageParam)

	searchResult, err := h.service.SearchDouban(query, searchType, page)
	if err != nil {
		h.logger.Error("Failed to search douban", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "搜索豆瓣失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    searchResult,
	})
}

// GetDoubanWish 获取用户想看列表
// @Summary 获取用户想看列表
// @Description 获取豆瓣用户想看的电影/电视剧列表
// @Tags 豆瓣
// @Produce json
// @Param id path string true "豆瓣用户ID"
// @Param type query string false "内容类型 (movie,tv)" default(movie)
// @Success 200 {object} Response{data=[]WishItem}
// @Router /douban/user/{id}/wish [get]
func (h *Handler) GetDoubanWish(c *gin.Context) {
	userID := c.Param("id")
	contentType := c.DefaultQuery("type", "movie")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	wishList, err := h.service.GetUserWishList(userID, contentType)
	if err != nil {
		h.logger.Error("Failed to get douban wish list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户想看列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    wishList,
	})
}

// GetDoubanDo 获取用户在看列表
// @Summary 获取用户在看列表
// @Description 获取豆瓣用户在看的电影/电视剧列表
// @Tags 豆瓣
// @Produce json
// @Param id path string true "豆瓣用户ID"
// @Param type query string false "内容类型 (movie,tv)" default(tv)
// @Success 200 {object} Response{data=[]DoItem}
// @Router /douban/user/{id}/do [get]
func (h *Handler) GetDoubanDo(c *gin.Context) {
	userID := c.Param("id")
	contentType := c.DefaultQuery("type", "tv")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	doList, err := h.service.GetUserDoList(userID, contentType)
	if err != nil {
		h.logger.Error("Failed to get douban do list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户在看列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    doList,
	})
}

// GetDoubanCollect 获取用户看过列表
// @Summary 获取用户看过列表
// @Description 获取豆瓣用户看过的电影/电视剧列表
// @Tags 豆瓣
// @Produce json
// @Param id path string true "豆瓣用户ID"
// @Param type query string false "内容类型 (movie,tv)" default(movie)
// @Success 200 {object} Response{data=[]CollectItem}
// @Router /douban/user/{id}/collect [get]
func (h *Handler) GetDoubanCollect(c *gin.Context) {
	userID := c.Param("id")
	contentType := c.DefaultQuery("type", "movie")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	collectList, err := h.service.GetUserCollectList(userID, contentType)
	if err != nil {
		h.logger.Error("Failed to get douban collect list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户看过列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    collectList,
	})
}

// GetPersonDetail 获取人物详情
// @Summary 获取人物详情
// @Description 根据人物ID查询人物详情
// @Tags 豆瓣
// @Produce json
// @Param person_id path int true "豆瓣人物ID"
// @Success 200 {object} Response{data=PersonDetail}
// @Router /douban/person/{person_id} [get]
func (h *Handler) GetPersonDetail(c *gin.Context) {
	personIDParam := c.Param("person_id")
	personID, err := strconv.Atoi(personIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的人物ID",
		})
		return
	}

	personDetail, err := h.service.GetPersonDetail(personID)
	if err != nil {
		h.logger.Error("Failed to get person detail", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取人物详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    personDetail,
	})
}

// GetPersonCredits 获取人物参演作品
// @Summary 获取人物参演作品
// @Description 根据人物ID查询人物参演作品
// @Tags 豆瓣
// @Produce json
// @Param person_id path int true "豆瓣人物ID"
// @Param page query int false "页码" default(1)
// @Success 200 {object} Response{data=[]PersonCredit}
// @Router /douban/person/credits/{person_id} [get]
func (h *Handler) GetPersonCredits(c *gin.Context) {
	personIDParam := c.Param("person_id")
	pageParam := c.DefaultQuery("page", "1")

	personID, err := strconv.Atoi(personIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的人物ID",
		})
		return
	}

	page, _ := strconv.Atoi(pageParam)

	credits, err := h.service.GetPersonCredits(personID, page)
	if err != nil {
		h.logger.Error("Failed to get person credits", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取人物参演作品失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    credits,
	})
}

// GetDoubanCredits 获取豆瓣演员阵容
// @Summary 获取豆瓣演员阵容
// @Description 根据豆瓣ID查询演员阵容
// @Tags 豆瓣
// @Produce json
// @Param doubanid path string true "豆瓣ID"
// @Param type_name path string true "类型名称 (电影/电视剧)"
// @Success 200 {object} Response{data=[]CreditInfo}
// @Router /douban/credits/{doubanid}/{type_name} [get]
func (h *Handler) GetDoubanCredits(c *gin.Context) {
	doubanID := c.Param("doubanid")
	typeName := c.Param("type_name")

	if doubanID == "" || typeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "豆瓣ID和类型名称不能为空",
		})
		return
	}

	credits, err := h.service.GetDoubanCredits(doubanID, typeName)
	if err != nil {
		h.logger.Error("Failed to get douban credits", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣演员阵容失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    credits,
	})
}

// GetDoubanRecommend 获取豆瓣推荐内容
// @Summary 获取豆瓣推荐内容
// @Description 根据豆瓣ID查询推荐电影/电视剧
// @Tags 豆瓣
// @Produce json
// @Param doubanid path string true "豆瓣ID"
// @Param type_name path string true "类型名称 (电影/电视剧)"
// @Success 200 {object} Response{data=[]RecommendInfo}
// @Router /douban/recommend/{doubanid}/{type_name} [get]
func (h *Handler) GetDoubanRecommend(c *gin.Context) {
	doubanID := c.Param("doubanid")
	typeName := c.Param("type_name")

	if doubanID == "" || typeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "豆瓣ID和类型名称不能为空",
		})
		return
	}

	recommendations, err := h.service.GetDoubanRecommendations(doubanID, typeName)
	if err != nil {
		h.logger.Error("Failed to get douban recommendations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取豆瓣推荐内容失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recommendations,
	})
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// MovieInfo 电影信息结构
type MovieInfo struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	OriginalTitle string   `json:"original_title"`
	Year          int      `json:"year"`
	Rating        float64  `json:"rating"`
	Votes         int      `json:"votes"`
	Poster        string   `json:"poster"`
	Summary       string   `json:"summary"`
	Genres        []string `json:"genres"`
	Countries     []string `json:"countries"`
	Languages     []string `json:"languages"`
	Duration      int      `json:"duration"`
	ReleaseDate   string   `json:"release_date"`
}

// TVInfo 电视剧信息结构
type TVInfo struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	OriginalTitle string   `json:"original_title"`
	Year          int      `json:"year"`
	Rating        float64  `json:"rating"`
	Votes         int      `json:"votes"`
	Poster        string   `json:"poster"`
	Summary       string   `json:"summary"`
	Genres        []string `json:"genres"`
	Countries     []string `json:"countries"`
	Languages     []string `json:"languages"`
	Episodes      int      `json:"episodes"`
	Seasons       int      `json:"seasons"`
	ReleaseDate   string   `json:"release_date"`
}

// PersonInfo 人物信息结构
type PersonInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Birthday   string   `json:"birthday"`
	Birthplace string   `json:"birthplace"`
	Occupation []string `json:"occupation"`
	Avatar     string   `json:"avatar"`
	Gender     string   `json:"gender"`
	Summary    string   `json:"summary"`
	Works      []string `json:"works"`
}

// SearchResult 搜索结果结构
type SearchResult struct {
	Total   int          `json:"total"`
	Results []SearchItem `json:"results"`
}

// SearchItem 搜索项结构
type SearchItem struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Year          int     `json:"year"`
	Rating        float64 `json:"rating"`
	Poster        string  `json:"poster"`
	Type          string  `json:"type"`
}

// WishItem 想看项结构
type WishItem struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Rating float64 `json:"rating"`
	Poster string  `json:"poster"`
	Date   string  `json:"date"`
}

// DoItem 在看项结构
type DoItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Episode int    `json:"episode"`
	Date    string `json:"date"`
}

// CollectItem 看过项结构
type CollectItem struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Rating float64 `json:"rating"`
	Date   string  `json:"date"`
}

// PersonDetail 人物详情结构
type PersonDetail struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Birthday      string   `json:"birthday"`
	Birthplace    string   `json:"birthplace"`
	Gender        string   `json:"gender"`
	Profession    []string `json:"profession"`
	Constellation string   `json:"constellation"`
	Avatar        string   `json:"avatar"`
	Alias         string   `json:"alias"`
	Family        string   `json:"family"`
	Works         int      `json:"works"`
	Summary       string   `json:"summary"`
}

// PersonCredit 人物作品结构
type PersonCredit struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Year   int    `json:"year"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	Poster string `json:"poster"`
}

// CreditInfo 演员阵容信息结构
type CreditInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Avatar    string `json:"avatar"`
	Character string `json:"character"`
}

// RecommendInfo 推荐信息结构
type RecommendInfo struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Year       int     `json:"year"`
	Rating     float64 `json:"rating"`
	Poster     string  `json:"poster"`
	Similarity float64 `json:"similarity"`
}
