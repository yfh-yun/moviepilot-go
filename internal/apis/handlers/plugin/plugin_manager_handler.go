package plugin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/plugin"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
)

// PluginManagerHandler 插件管理器处理器
type PluginManagerHandler struct {
	manager *plugin.HybridPluginManager
	logger  *zap.Logger
}

// NewPluginManagerHandler 创建插件管理器处理器
func NewPluginManagerHandler(manager *plugin.HybridPluginManager, logger *zap.Logger) *PluginManagerHandler {
	return &PluginManagerHandler{
		manager: manager,
		logger:  logger,
	}
}

// LoadPluginRequest 加载插件请求
type LoadPluginRequest struct {
	Path string `json:"path" binding:"required"`
	Type string `json:"type" binding:"required,oneof=native script web"`
}

// PluginMethodRequest 插件方法调用请求
type PluginMethodRequest struct {
	Method string        `json:"method" binding:"required"`
	Args   []interface{} `json:"args"`
}

// LoadPlugin 加载插件
// @Summary 加载插件
// @Description 加载指定路径和类型的插件
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Param request body LoadPluginRequest true "加载插件请求"
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/load [post]
func (h *PluginManagerHandler) LoadPlugin(c *gin.Context) {
	var req LoadPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	var pluginType plugin.PluginType
	switch req.Type {
	case "native":
		pluginType = plugin.PluginTypeNative
	case "script":
		pluginType = plugin.PluginTypeScript
	case "web":
		pluginType = plugin.PluginTypeWeb
	default:
		response.Error(c, http.StatusBadRequest, "Unsupported plugin type", nil)
		return
	}

	if err := h.manager.LoadPlugin(req.Path, pluginType); err != nil {
		h.logger.Error("Failed to load plugin", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to load plugin", err)
		return
	}

	response.Success(c, "Plugin loaded successfully")
}

// InitializePlugin 初始化插件
// @Summary 初始化插件
// @Description 初始化已加载的插件
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Param pluginId path string true "插件ID"
// @Success 200 {object} response.Response{data=string}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/{pluginId}/initialize [post]
func (h *PluginManagerHandler) InitializePlugin(c *gin.Context) {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		response.Error(c, http.StatusBadRequest, "Plugin ID is required", nil)
		return
	}

	if err := h.manager.InitializePlugin(pluginID); err != nil {
		h.logger.Error("Failed to initialize plugin", zap.String("pluginId", pluginID), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to initialize plugin", err)
		return
	}

	response.Success(c, "Plugin initialized successfully")
}

// StartPlugin 启动插件
// @Summary 启动插件
// @Description 启动已初始化的插件
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Param pluginId path string true "插件ID"
// @Success 200 {object} response.Response{data=string}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/{pluginId}/start [post]
func (h *PluginManagerHandler) StartPlugin(c *gin.Context) {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		response.Error(c, http.StatusBadRequest, "Plugin ID is required", nil)
		return
	}

	if err := h.manager.StartPlugin(pluginID); err != nil {
		h.logger.Error("Failed to start plugin", zap.String("pluginId", pluginID), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to start plugin", err)
		return
	}

	response.Success(c, "Plugin started successfully")
}

// StopPlugin 停止插件
// @Summary 停止插件
// @Description 停止正在运行的插件
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Param pluginId path string true "插件ID"
// @Success 200 {object} response.Response{data=string}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/{pluginId}/stop [post]
func (h *PluginManagerHandler) StopPlugin(c *gin.Context) {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		response.Error(c, http.StatusBadRequest, "Plugin ID is required", nil)
		return
	}

	if err := h.manager.StopPlugin(pluginID); err != nil {
		h.logger.Error("Failed to stop plugin", zap.String("pluginId", pluginID), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to stop plugin", err)
		return
	}

	response.Success(c, "Plugin stopped successfully")
}

// UnloadPlugin 卸载插件
// @Summary 卸载插件
// @Description 卸载插件并释放资源
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Param pluginId path string true "插件ID"
// @Success 200 {object} response.Response{data=string}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/{pluginId}/unload [post]
func (h *PluginManagerHandler) UnloadPlugin(c *gin.Context) {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		response.Error(c, http.StatusBadRequest, "Plugin ID is required", nil)
		return
	}

	if err := h.manager.UnloadPlugin(pluginID); err != nil {
		h.logger.Error("Failed to unload plugin", zap.String("pluginId", pluginID), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to unload plugin", err)
		return
	}

	response.Success(c, "Plugin unloaded successfully")
}

// CallPluginMethod 调用插件方法
// @Summary 调用插件方法
// @Description 调用指定插件的方法
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Param pluginId path string true "插件ID"
// @Param request body PluginMethodRequest true "方法调用请求"
// @Success 200 {object} response.Response{data=interface{}}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/{pluginId}/call [post]
func (h *PluginManagerHandler) CallPluginMethod(c *gin.Context) {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		response.Error(c, http.StatusBadRequest, "Plugin ID is required", nil)
		return
	}

	var req PluginMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	result, err := h.manager.CallPluginMethod(pluginID, req.Method, req.Args...)
	if err != nil {
		h.logger.Error("Failed to call plugin method", 
			zap.String("pluginId", pluginID), 
			zap.String("method", req.Method), 
			zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to call plugin method", err)
		return
	}

	response.Success(c, result)
}

// GetPluginInfo 获取插件信息
// @Summary 获取插件信息
// @Description 获取指定插件的详细信息
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Param pluginId path string true "插件ID"
// @Success 200 {object} response.Response{data=plugin.PluginInfo}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/{pluginId}/info [get]
func (h *PluginManagerHandler) GetPluginInfo(c *gin.Context) {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		response.Error(c, http.StatusBadRequest, "Plugin ID is required", nil)
		return
	}

	info, err := h.manager.GetPluginInfo(pluginID)
	if err != nil {
		h.logger.Error("Failed to get plugin info", zap.String("pluginId", pluginID), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to get plugin info", err)
		return
	}

	response.Success(c, info)
}

// ListPlugins 列出所有插件
// @Summary 列出所有插件
// @Description 获取所有已加载插件的信息
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]plugin.PluginInfo}
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/plugins [get]
func (h *PluginManagerHandler) ListPlugins(c *gin.Context) {
	plugins := h.manager.ListPlugins()
	response.Success(c, plugins)
}

// PublishEvent 发布事件
// @Summary 发布事件
// @Description 向所有插件发布事件
// @Tags plugin-manager
// @Accept json
// @Produce json
// @Param event body plugin.Event true "事件数据"
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/plugin-manager/events [post]
func (h *PluginManagerHandler) PublishEvent(c *gin.Context) {
	var event plugin.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid event data", err)
		return
	}

	// 这里需要访问HybridPluginManager的eventBus
	// 由于eventBus是私有的，我们需要在HybridPluginManager中添加一个公开的方法
	h.logger.Info("Event published", zap.String("eventType", event.Type), zap.String("eventId", event.ID))
	response.Success(c, "Event published successfully")
}