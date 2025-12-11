package mediaserver

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	middlewares "moviepilot-go/internal/apis/middlewares"
	"moviepilot-go/internal/integration/mediaserver"
	"moviepilot-go/internal/models/dto"
)

// Handler 媒体服务器处理器
type Handler struct {
	logger        *zap.Logger
	clientFactory *mediaserver.Factory
}

// NewHandler 创建媒体服务器处理器
func NewHandler(logger *zap.Logger, clientFactory *mediaserver.Factory) *Handler {
	return &Handler{
		logger:        logger,
		clientFactory: clientFactory,
	}
}

// PlayItem 在线播放
// @Summary 在线播放
// @Description 获取媒体服务器播放页面地址
// @Tags mediaserver
// @Produce json
// @Param itemid path string true "媒体条目ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/play/{itemid} [get]
func (h *Handler) PlayItem(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	itemID := c.Param("itemid")

	if itemID == "" {
		h.logger.Error("Missing itemid for play",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "参数错误",
		})
		return
	}

	h.logger.Info("Media server play requested",
		zap.String("itemid", itemID),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取媒体服务器播放页面地址逻辑
	// 1. 遍历所有媒体服务器
	// 2. 尝试获取播放地址
	// 3. 返回第一个可用的播放地址

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]any{
			"url": "",
		},
	})
}

// ExistsLocal 查询本地是否存在（数据库）
// @Summary 查询本地是否存在（数据库）
// @Description 判断本地是否存在媒体
// @Tags mediaserver
// @Produce json
// @Param title query string false "标题"
// @Param year query string false "年份"
// @Param mtype query string false "媒体类型"
// @Param tmdbid query int false "TMDB ID"
// @Param season query int false "季号"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/exists [get]
func (h *Handler) ExistsLocal(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	title := c.Query("title")
	year := c.Query("year")
	mtype := c.Query("mtype")
	tmdbIDStr := c.Query("tmdbid")
	seasonStr := c.Query("season")

	var tmdbID int
	var season int
	var err error

	if tmdbIDStr != "" {
		tmdbID, err = strconv.Atoi(tmdbIDStr)
		if err != nil {
			tmdbID = 0
		}
	}

	if seasonStr != "" {
		season, err = strconv.Atoi(seasonStr)
		if err != nil {
			season = 0
		}
	}

	h.logger.Info("Media server exists local requested",
		zap.String("title", title),
		zap.String("year", year),
		zap.String("mtype", mtype),
		zap.Int("tmdbid", tmdbID),
		zap.Int("season", season),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现查询本地是否存在逻辑
	// 1. 查询本地数据库
	// 2. 返回存在状态和媒体ID

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"data": map[string]any{
			"item": map[string]any{},
		},
	})
}

// ExistsRemote 查询已存在的剧集信息（媒体服务器）
// @Summary 查询已存在的剧集信息（媒体服务器）
// @Description 根据媒体信息查询媒体库已存在的剧集信息
// @Tags mediaserver
// @Accept json
// @Produce json
// @Param media body dto.MediaInfo true "媒体信息"
// @Success 200 {object} map[int][]int
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/exists_remote [post]
func (h *Handler) ExistsRemote(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	var mediaInfo dto.MediaInfo
	if err := c.ShouldBindJSON(&mediaInfo); err != nil {
		h.logger.Error("Invalid media info for exists_remote",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的媒体信息",
		})
		return
	}

	// 处理可空字段
	var tmdbID int
	if mediaInfo.TmdbID != nil {
		tmdbID = *mediaInfo.TmdbID
	}

	var season int
	if mediaInfo.Season != nil {
		season = *mediaInfo.Season
	}

	h.logger.Info("Media server exists remote requested",
		zap.String("title", mediaInfo.Title),
		zap.Int("tmdb_id", tmdbID),
		zap.Int("season", season),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现查询已存在的剧集信息逻辑
	// 1. 遍历所有媒体服务器
	// 2. 查询媒体库中已存在的剧集
	// 3. 返回存在的剧集信息

	c.JSON(http.StatusOK, map[int][]int{})
}

// NotExists 查询媒体库缺失信息（媒体服务器）
// @Summary 查询媒体库缺失信息（媒体服务器）
// @Description 根据媒体信息查询缺失电影/剧集
// @Tags mediaserver
// @Accept json
// @Produce json
// @Param media body dto.MediaInfo true "媒体信息"
// @Success 200 {object} []map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/notexists [post]
func (h *Handler) NotExists(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	var mediaInfo dto.MediaInfo
	if err := c.ShouldBindJSON(&mediaInfo); err != nil {
		h.logger.Error("Invalid media info for notexists",
			zap.Error(err),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的媒体信息",
		})
		return
	}

	// 处理可空字段
	var tmdbID int
	if mediaInfo.TmdbID != nil {
		tmdbID = *mediaInfo.TmdbID
	}

	var season int
	if mediaInfo.Season != nil {
		season = *mediaInfo.Season
	}

	h.logger.Info("Media server notexists requested",
		zap.String("title", mediaInfo.Title),
		zap.Int("tmdb_id", tmdbID),
		zap.Int("season", season),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现查询媒体库缺失信息逻辑
	// 1. 遍历所有媒体服务器
	// 2. 查询媒体库中缺失的剧集
	// 3. 返回缺失的剧集信息

	c.JSON(http.StatusOK, []map[string]any{})
}

// Latest 最新入库条目
// @Summary 最新入库条目
// @Description 获取媒体服务器最新入库条目
// @Tags mediaserver
// @Produce json
// @Param server query string true "服务器名称"
// @Param count query int false "数量" default(20)
// @Success 200 {object} []map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/latest [get]
func (h *Handler) Latest(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	server := c.Query("server")
	countStr := c.DefaultQuery("count", "20")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		count = 20
	}

	if server == "" {
		h.logger.Error("Missing server for latest",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "服务器名称不能为空",
		})
		return
	}

	h.logger.Info("Media server latest requested",
		zap.String("server", server),
		zap.Int("count", count),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取最新入库条目逻辑
	// 1. 获取指定服务器的客户端
	// 2. 查询最新入库的媒体条目
	// 3. 返回结果

	c.JSON(http.StatusOK, []map[string]any{})
}

// Playing 正在播放条目
// @Summary 正在播放条目
// @Description 获取媒体服务器正在播放条目
// @Tags mediaserver
// @Produce json
// @Param server query string true "服务器名称"
// @Param count query int false "数量" default(12)
// @Success 200 {object} []map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/playing [get]
func (h *Handler) Playing(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	server := c.Query("server")
	countStr := c.DefaultQuery("count", "12")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		count = 12
	}

	if server == "" {
		h.logger.Error("Missing server for playing",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "服务器名称不能为空",
		})
		return
	}

	h.logger.Info("Media server playing requested",
		zap.String("server", server),
		zap.Int("count", count),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// TODO: 实现获取正在播放条目逻辑
	// 1. 获取指定服务器的客户端
	// 2. 查询正在播放的媒体条目
	// 3. 返回结果

	c.JSON(http.StatusOK, []map[string]any{})
}

// Library 媒体库列表
// @Summary 媒体库列表
// @Description 获取媒体服务器媒体库列表
// @Tags mediaserver
// @Produce json
// @Param server query string true "服务器名称"
// @Param hidden query bool false "是否显示隐藏媒体库" default(false)
// @Success 200 {object} []map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/library [get]
func (h *Handler) Library(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	server := c.Query("server")
	hiddenStr := c.DefaultQuery("hidden", "false")
	hidden, _ := strconv.ParseBool(hiddenStr)

	if server == "" {
		h.logger.Error("Missing server for library",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "服务器名称不能为空",
		})
		return
	}

	h.logger.Info("Media server library requested",
		zap.String("server", server),
		zap.Bool("hidden", hidden),
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 获取指定服务器的客户端
	client, ok := h.clientFactory.GetClient(server)
	if !ok {
		h.logger.Error("Media server client not found",
			zap.String("server", server),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "未找到媒体服务器客户端",
		})
		return
	}

	// 列出媒体库
	libraries, err := client.ListLibraries(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list libraries",
			zap.Error(err),
			zap.String("server", server),
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取媒体库列表失败: " + err.Error(),
		})
		return
	}

	// 转换为响应格式
	response := make([]map[string]any, 0, len(libraries))
	for _, lib := range libraries {
		response = append(response, map[string]any{
			"id":         lib.ID,
			"name":       lib.Name,
			"type":       string(lib.Type),
			"item_count": lib.ItemCount,
		})
	}

	c.JSON(http.StatusOK, response)
}

// Clients 查询可用媒体服务器
// @Summary 查询可用媒体服务器
// @Description 查询可用媒体服务器
// @Tags mediaserver
// @Produce json
// @Success 200 {object} []map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/mediaserver/clients [get]
func (h *Handler) Clients(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)

	h.logger.Info("Media server clients requested",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	// 获取所有已注册的客户端名称
	clients := h.clientFactory.ListClients()

	// 转换为响应格式
	response := make([]map[string]any, 0, len(clients))
	for _, clientName := range clients {
		// 获取客户端信息
		client, ok := h.clientFactory.GetClient(clientName)
		if !ok {
			continue
		}

		// 获取服务器信息
		serverInfo, err := client.GetServerInfo(c.Request.Context())
		if err != nil {
			h.logger.Error("Failed to get server info",
				zap.Error(err),
				zap.String("client", clientName),
				zap.String("request_id", reqID),
				zap.Uint("user_id", userID),
			)
			continue
		}

		response = append(response, map[string]any{
			"name": clientName,
			"type": serverInfo.Type,
		})
	}

	c.JSON(http.StatusOK, response)
}
