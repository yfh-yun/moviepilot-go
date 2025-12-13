package plugin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/plugin"
	"moviepilot-go/pkg/logger"
)

// Handler 插件管理API处理器
type Handler struct {
	pluginService plugin.Service
	logger        *zap.Logger
}

// NewHandler 创建插件管理API处理器
func NewHandler(pluginService plugin.Service) *Handler {
	return &Handler{
		pluginService: pluginService,
		logger:        logger.GetLogger(),
	}
}

// RegisterRoutes 注册插件管理API路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	pluginRoutes := router.Group("/plugins")
	{
		pluginRoutes.GET("", h.ListPlugins)                    // 获取插件列表
		pluginRoutes.GET("/:id", h.GetPlugin)                  // 获取插件详情
		pluginRoutes.POST("/:id/start", h.StartPlugin)         // 启动插件
		pluginRoutes.POST("/:id/stop", h.StopPlugin)           // 停止插件
		pluginRoutes.GET("/:id/config", h.GetPluginConfig)     // 获取插件配置
		pluginRoutes.PUT("/:id/config", h.UpdatePluginConfig)  // 更新插件配置
		pluginRoutes.GET("/:id/commands", h.GetPluginCommands) // 获取插件命令
		pluginRoutes.GET("/:id/apis", h.GetPluginAPIs)         // 获取插件API
		pluginRoutes.POST("/:id/execute", h.ExecutePlugin)     // 执行插件方法
	}
}

// ListPlugins 获取插件列表
// @Summary 获取插件列表
// @Description 获取所有插件的列表信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} pluginpkg.PluginInfo
// @Router /api/v1/plugins [get]
func (h *Handler) ListPlugins(c *gin.Context) {
	plugins, err := h.pluginService.ListPlugins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, plugins)
}

// GetPlugin 获取插件详情
// @Summary 获取插件详情
// @Description 获取指定插件的详细信息
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} pluginpkg.PluginInfo
// @Router /api/v1/plugins/{id} [get]
func (h *Handler) GetPlugin(c *gin.Context) {
	id := c.Param("id")
	pluginInfo, err := h.pluginService.GetPlugin(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pluginInfo)
}

// StartPlugin 启动插件
// @Summary 启动插件
// @Description 启动指定的插件
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} gin.H
// @Router /api/v1/plugins/{id}/start [post]
func (h *Handler) StartPlugin(c *gin.Context) {
	id := c.Param("id")
	if err := h.pluginService.EnablePlugin(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "插件启动成功",
	})
}

// StopPlugin 停止插件
// @Summary 停止插件
// @Description 停止指定的插件
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} gin.H
// @Router /api/v1/plugins/{id}/stop [post]
func (h *Handler) StopPlugin(c *gin.Context) {
	id := c.Param("id")
	if err := h.pluginService.DisablePlugin(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "插件停止成功",
	})
}

// GetPluginConfig 获取插件配置
// @Summary 获取插件配置
// @Description 获取指定插件的配置
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/plugins/{id}/config [get]
func (h *Handler) GetPluginConfig(c *gin.Context) {
	id := c.Param("id")
	config, err := h.pluginService.GetPluginConfig(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdatePluginConfig 更新插件配置
// @Summary 更新插件配置
// @Description 更新指定插件的配置
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Param config body map[string]any true "插件配置"
// @Success 200 {object} gin.H
// @Router /api/v1/plugins/{id}/config [put]
func (h *Handler) UpdatePluginConfig(c *gin.Context) {
	id := c.Param("id")
	var config map[string]any
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.pluginService.ConfigurePlugin(c.Request.Context(), id, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "插件配置更新成功",
	})
}

// GetPluginCommands 获取插件命令
// @Summary 获取插件命令
// @Description 获取指定插件的命令列表
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {array} pluginpkg.Command
// @Router /api/v1/plugins/{id}/commands [get]
func (h *Handler) GetPluginCommands(c *gin.Context) {
	id := c.Param("id")
	pluginInfo, err := h.pluginService.GetPlugin(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pluginInfo.Commands)
}

// GetPluginAPIs 获取插件API
// @Summary 获取插件API
// @Description 获取指定插件的API列表
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Success 200 {array} pluginpkg.API
// @Router /api/v1/plugins/{id}/apis [get]
func (h *Handler) GetPluginAPIs(c *gin.Context) {
	id := c.Param("id")
	pluginInfo, err := h.pluginService.GetPlugin(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pluginInfo.APIs)
}

// ExecutePlugin 执行插件方法
// @Summary 执行插件方法
// @Description 执行指定插件的方法
// @Tags 插件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "插件ID"
// @Param request body map[string]any true "执行请求"
// @Success 200 {object} map[string]any
// @Router /api/v1/plugins/{id}/execute [post]
func (h *Handler) ExecutePlugin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "执行插件方法功能暂未实现",
	})
}
