package media

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	"moviepilot-go/internal/business/services/media"
	"moviepilot-go/internal/models/dto"
)

// Handler 媒体处理器
type Handler struct {
	logger       *zap.Logger
	mediaService *media.MediaService
}

// NewHandler 创建媒体处理器
func NewHandler(logger *zap.Logger, mediaService *media.MediaService) *Handler {
	return &Handler{
		logger:       logger,
		mediaService: mediaService,
	}
}

// BatchScrape 批量媒体识别
// @Summary 批量媒体识别
// @Description 批量对多个媒体文件进行识别
// @Tags media
// @Accept json
// @Produce json
// @Param request body dto.MediaBatchScrapeRequest true "批量识别请求"
// @Success 200 {object} dto.MediaBatchScrapeResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/batch-scrape [post]
func (h *Handler) BatchScrape(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	var req dto.MediaBatchScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid batch scrape request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Batch media scrape requested",
		zap.Int("count", len(req.FilePaths)),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现批量媒体识别逻辑
	// 1. 验证文件路径
	// 2. 创建批量任务
	// 3. 并行处理识别任务

	response := dto.MediaBatchScrapeResponse{
		TaskID:    "batch_" + generateTaskID(), // TODO: 生成唯一任务ID
		Status:    "created",
		Total:     len(req.FilePaths),
		Processed: 0,
		Failed:    0,
		Message:   "Batch scrape task created successfully",
	}

	h.logger.Info("Batch media scrape task created",
		zap.String("task_id", response.TaskID),
		zap.Int("total_files", response.Total),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)
	c.JSON(http.StatusOK, response)
}

// RefreshMetadata 刷新媒体元数据
// @Summary 刷新媒体元数据
// @Description 重新获取和更新媒体元数据
// @Tags media
// @Accept json
// @Produce json
// @Param id path string true "媒体ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/{id}/refresh [post]
func (h *Handler) RefreshMetadata(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	mediaID := c.Param("id")
	if mediaID == "" {
		h.logger.Error("Missing media ID for metadata refresh",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Media ID is required",
		})
		return
	}

	h.logger.Info("Media metadata refresh requested",
		zap.String("media_id", mediaID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现元数据刷新逻辑
	// 1. 重新调用外部API获取最新信息
	// 2. 更新数据库记录
	// 3. 清除相关缓存

	c.JSON(http.StatusOK, gin.H{
		"media_id": mediaID,
		"status":   "refreshing",
		"message":  "Metadata refresh started",
	})
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	// TODO: 实现生成唯一任务ID的逻辑
	// 可以使用UUID或时间戳+随机数
	return "temp_task_id"
}

// Recognize 识别媒体信息（种子）
// @Summary 识别媒体信息（种子）
// @Description 根据标题、副标题识别媒体信息
// @Tags media
// @Produce json
// @Param title query string true "标题"
// @Param subtitle query string false "副标题"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/recognize [get]
func (h *Handler) Recognize(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	title := c.Query("title")
	subtitle := c.Query("subtitle")

	if title == "" {
		h.logger.Error("Missing title for media recognize",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "标题不能为空",
		})
		return
	}

	h.logger.Info("Media recognize requested",
		zap.String("title", title),
		zap.String("subtitle", subtitle),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用媒体服务识别媒体信息
	metaInfo, err := h.mediaService.RecognizeMedia(c.Request.Context(), title)
	if err != nil {
		h.logger.Error("Media recognize failed",
			zap.Error(err),
			zap.String("title", title),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "识别媒体信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metaInfo)
}

// RecognizeFile 识别媒体信息（文件）
// @Summary 识别媒体信息（文件）
// @Description 根据文件路径识别媒体信息
// @Tags media
// @Produce json
// @Param path query string true "文件路径"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/recognize_file [get]
func (h *Handler) RecognizeFile(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	path := c.Query("path")

	if path == "" {
		h.logger.Error("Missing path for media recognize_file",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件路径不能为空",
		})
		return
	}

	h.logger.Info("Media recognize_file requested",
		zap.String("path", path),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用媒体服务识别媒体信息
	metaInfo, err := h.mediaService.RecognizeMedia(c.Request.Context(), path)
	if err != nil {
		h.logger.Error("Media recognize_file failed",
			zap.Error(err),
			zap.String("path", path),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "识别媒体信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metaInfo)
}

// Search 搜索媒体/人物信息
// @Summary 搜索媒体/人物信息
// @Description 模糊搜索媒体/人物信息列表 media：媒体信息，person：人物信息
// @Tags media
// @Produce json
// @Param title query string true "搜索关键词"
// @Param type query string false "搜索类型：media, person, collection"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(8)
// @Success 200 {object} []map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/search [get]
func (h *Handler) Search(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	title := c.Query("title")
	searchType := c.DefaultQuery("type", "media")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "8"))

	if title == "" {
		h.logger.Error("Missing title for media search",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "搜索关键词不能为空",
		})
		return
	}

	h.logger.Info("Media search requested",
		zap.String("title", title),
		zap.String("type", searchType),
		zap.Int("page", page),
		zap.Int("count", count),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用媒体服务搜索媒体信息
	results, err := h.mediaService.SearchMedia(c.Request.Context(), title, searchType)
	if err != nil {
		h.logger.Error("Media search failed",
			zap.Error(err),
			zap.String("title", title),
			zap.String("type", searchType),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索媒体信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// Scrape 刮削媒体信息
// @Summary 刮削媒体信息
// @Description 刮削媒体信息
// @Tags media
// @Accept json
// @Produce json
// @Param storage path string false "存储类型" default(local)
// @Param fileitem body dto.MediaScrapeRequest true "文件信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/scrape/{storage} [post]
func (h *Handler) Scrape(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	storage := c.DefaultQuery("storage", "local")
	var req dto.MediaScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid scrape request",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	h.logger.Info("Media scrape requested",
		zap.String("file_path", req.FilePath),
		zap.String("storage", storage),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用媒体服务刮削媒体信息
	mediaInfo, err := h.mediaService.RecognizeMedia(c.Request.Context(), req.FilePath)
	if err != nil {
		h.logger.Error("Media scrape failed",
			zap.Error(err),
			zap.String("file_path", req.FilePath),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "刮削媒体信息失败: " + err.Error(),
		})
		return
	}

	// 刮削元数据
	err = h.mediaService.ScrapeMetadata(c.Request.Context(), mediaInfo, req.FilePath)
	if err != nil {
		h.logger.Error("Media metadata scrape failed",
			zap.Error(err),
			zap.String("file_path", req.FilePath),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "刮削元数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": req.FilePath + " 刮削完成",
	})
}

// Category 查询自动分类配置
// @Summary 查询自动分类配置
// @Description 查询自动分类配置
// @Tags media
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/category [get]
func (h *Handler) Category(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("Media category requested",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现查询自动分类配置逻辑
	// 暂时返回空对象
	c.JSON(http.StatusOK, map[string]any{})
}

// GroupSeasons 查询剧集组季信息
// @Summary 查询剧集组季信息
// @Description 查询剧集组季信息（themoviedb）
// @Tags media
// @Produce json
// @Param episode_group path string true "剧集组ID"
// @Success 200 {object} []map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/group/seasons/{episode_group} [get]
func (h *Handler) GroupSeasons(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	episodeGroup := c.Param("episode_group")
	if episodeGroup == "" {
		h.logger.Error("Missing episode_group for group seasons",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "剧集组ID不能为空",
		})
		return
	}

	h.logger.Info("Media group seasons requested",
		zap.String("episode_group", episodeGroup),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现查询剧集组季信息逻辑
	c.JSON(http.StatusOK, []map[string]any{})
}

// Groups 查询媒体剧集组
// @Summary 查询媒体剧集组
// @Description 查询媒体剧集组列表（themoviedb）
// @Tags media
// @Produce json
// @Param tmdbid path int true "TMDB ID"
// @Success 200 {object} []map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/groups/{tmdbid} [get]
func (h *Handler) Groups(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	tmdbIDStr := c.Param("tmdbid")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		h.logger.Error("Invalid tmdbid for media groups",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的TMDB ID",
		})
		return
	}

	h.logger.Info("Media groups requested",
		zap.Int("tmdbid", tmdbID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 调用媒体服务获取媒体信息
	_, err = h.mediaService.GetMediaInfo(c.Request.Context(), tmdbID, "tv")
	if err != nil {
		h.logger.Error("Get media info failed",
			zap.Error(err),
			zap.Int("tmdbid", tmdbID),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取媒体信息失败: " + err.Error(),
		})
		return
	}

	// TODO: 实现查询媒体剧集组逻辑
	c.JSON(http.StatusOK, []map[string]any{})
}

// Seasons 查询媒体季信息
// @Summary 查询媒体季信息
// @Description 查询媒体季信息
// @Tags media
// @Produce json
// @Param mediaid query string false "媒体ID"
// @Param title query string false "标题"
// @Param year query string false "年份"
// @Param season query int false "季号"
// @Success 200 {object} []map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/seasons [get]
func (h *Handler) Seasons(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	mediaID := c.Query("mediaid")
	title := c.Query("title")
	year := c.Query("year")
	seasonStr := c.Query("season")
	var season int
	if seasonStr != "" {
		season, _ = strconv.Atoi(seasonStr)
	}

	h.logger.Info("Media seasons requested",
		zap.String("mediaid", mediaID),
		zap.String("title", title),
		zap.String("year", year),
		zap.Int("season", season),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现查询媒体季信息逻辑
	c.JSON(http.StatusOK, []map[string]any{})
}

// Detail 查询媒体详情
// @Summary 查询媒体详情
// @Description 根据媒体ID查询themoviedb或豆瓣媒体信息
// @Tags media
// @Produce json
// @Param mediaid path string true "媒体ID"
// @Param type_name query string true "媒体类型: 电影/电视剧"
// @Param title query string false "标题"
// @Param year query string false "年份"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/media/{mediaid} [get]
func (h *Handler) Detail(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	mediaID := c.Param("mediaid")
	typeName := c.Query("type_name")
	title := c.Query("title")
	year := c.Query("year")

	if mediaID == "" || typeName == "" {
		h.logger.Error("Missing mediaid or type_name for media detail",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "媒体ID和类型不能为空",
		})
		return
	}

	// 转换媒体类型
	mediaType := "movie"
	if typeName == "电视剧" || typeName == "tv" {
		mediaType = "tv"
	}

	h.logger.Info("Media detail requested",
		zap.String("mediaid", mediaID),
		zap.String("type_name", typeName),
		zap.String("media_type", mediaType),
		zap.String("title", title),
		zap.String("year", year),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 解析TMDB ID
	var tmdbID int
	if len(mediaID) > 5 && mediaID[:5] == "tmdb:" {
		tmdbIDStr := mediaID[5:]
		var err error
		tmdbID, err = strconv.Atoi(tmdbIDStr)
		if err != nil {
			h.logger.Error("Invalid tmdbid format",
				zap.Error(err),
				zap.String("mediaid", mediaID),
				zap.String("request_id", reqID),
				zap.Uint("user_id", userID),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的TMDB ID格式",
			})
			return
		}
	}

	// 调用媒体服务获取媒体详情
	mediaInfo, err := h.mediaService.GetMediaInfo(c.Request.Context(), tmdbID, mediaType)
	if err != nil {
		h.logger.Error("Get media detail failed",
			zap.Error(err),
			zap.String("mediaid", mediaID),
			zap.String("media_type", mediaType),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取媒体详情失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, mediaInfo)
}
