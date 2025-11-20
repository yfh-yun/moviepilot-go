package mediaserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/business/services/mediaserver"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"
)

// Handler 表示媒体服务器API处理器
type Handler struct {
	logger  *zap.Logger
	service *mediaserver.Service
}

// NewHandler 创建新的媒体服务器处理器
func NewHandler(logger *zap.Logger, service *mediaserver.Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// HealthCheck 健康检查接口
// @Summary 媒体服务器健康检查
// @Description 检查所有媒体服务器的连接状态
// @Tags 媒体服务器
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 500 {object} response.Response
// @Router /api/v1/mediaserver/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	results := h.service.HealthCheckAll()

	status := make(map[string]string)
	for serverType, err := range results {
		if err != nil {
			status[string(serverType)] = "unhealthy"
		} else {
			status[string(serverType)] = "healthy"
		}
	}

	response.Success(c, status)
}

// ListServers 获取服务器列表
// @Summary 获取媒体服务器列表
// @Description 获取所有已配置的媒体服务器
// @Tags 媒体服务器
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Failure 500 {object} response.Response
// @Router /api/v1/mediaserver/servers [get]
func (h *Handler) ListServers(c *gin.Context) {
	serverTypes := h.service.ListServers()

	servers := make([]string, len(serverTypes))
	for i, serverType := range serverTypes {
		servers[i] = string(serverType)
	}

	response.Success(c, servers)
}

// GetLibraries 获取媒体库列表
// @Summary 获取媒体库列表
// @Description 获取指定媒体服务器的媒体库列表
// @Tags 媒体服务器
// @Accept json
// @Produce json
// @Param server_type path string true "服务器类型(emby|jellyfin|plex)"
// @Success 200 {object} response.Response{data=[]mediaserver.Library}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/mediaserver/{server_type}/libraries [get]
func (h *Handler) GetLibraries(c *gin.Context) {
	serverType := mediaserver.MediaServerType(c.Param("server_type"))

	server, err := h.service.GetServer(serverType)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.Wrap(err, "获取媒体服务器失败"))
		return
	}

	libraries, err := server.GetLibraries()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errors.Wrap(err, "获取媒体库列表失败"))
		return
	}

	response.Success(c, libraries)
}

// RefreshLibrary 刷新媒体库
// @Summary 刷新媒体库
// @Description 刷新指定媒体服务器的媒体库
// @Tags 媒体服务器
// @Accept json
// @Produce json
// @Param server_type path string true "服务器类型(emby|jellyfin|plex)"
// @Param library_id path string true "媒体库ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/mediaserver/{server_type}/libraries/{library_id}/refresh [post]
func (h *Handler) RefreshLibrary(c *gin.Context) {
	serverType := mediaserver.MediaServerType(c.Param("server_type"))
	libraryID := c.Param("library_id")

	if err := h.service.RefreshLibrary(c.Request.Context(), serverType, libraryID); err != nil {
		response.Error(c, http.StatusInternalServerError, errors.Wrap(err, "刷新媒体库失败"))
		return
	}

	response.Success(c, nil)
}

// GetPlaybackSessions 获取播放会话
// @Summary 获取播放会话
// @Description 获取指定媒体服务器的当前播放会话
// @Tags 媒体服务器
// @Accept json
// @Produce json
// @Param server_type path string true "服务器类型(emby|jellyfin|plex)"
// @Success 200 {object} response.Response{data=[]mediaserver.PlaybackSession}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/mediaserver/{server_type}/sessions [get]
func (h *Handler) GetPlaybackSessions(c *gin.Context) {
	serverType := mediaserver.MediaServerType(c.Param("server_type"))

	sessions, err := h.service.GetPlaybackSessions(c.Request.Context(), serverType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, errors.Wrap(err, "获取播放会话失败"))
		return
	}

	response.Success(c, sessions)
}

// SyncLibraries 同步媒体库
// @Summary 同步媒体库
// @Description 同步指定媒体服务器的媒体库到本地数据库
// @Tags 媒体服务器
// @Accept json
// @Produce json
// @Param server_type path string true "服务器类型(emby|jellyfin|plex)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/mediaserver/{server_type}/sync [post]
func (h *Handler) SyncLibraries(c *gin.Context) {
	serverType := mediaserver.MediaServerType(c.Param("server_type"))

	if err := h.service.SyncMediaLibraries(c.Request.Context(), serverType); err != nil {
		response.Error(c, http.StatusInternalServerError, errors.Wrap(err, "同步媒体库失败"))
		return
	}

	response.Success(c, nil)
}

// 参数验证结构体

type RefreshLibraryRequest struct {
	LibraryID string `json:"library_id" binding:"required"`
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	mediaserverGroup := router.Group("/mediaserver")
	{
		mediaserverGroup.GET("/health", h.HealthCheck)
		mediaserverGroup.GET("/servers", h.ListServers)

		serverGroup := mediaserverGroup.Group("/:server_type")
		{
			serverGroup.GET("/libraries", h.GetLibraries)
			serverGroup.POST("/libraries/:library_id/refresh", h.RefreshLibrary)
			serverGroup.GET("/sessions", h.GetPlaybackSessions)
			serverGroup.POST("/sync", h.SyncLibraries)
		}
	}
}

// 输入验证函数
func (h *Handler) validateRefreshLibraryRequest(c *gin.Context) (*RefreshLibraryRequest, error) {
	var req RefreshLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, validator.HandleValidationError(err)
	}

	if req.LibraryID == "" {
		return nil, errors.New("library_id不能为空")
	}

	return &req, nil
}
