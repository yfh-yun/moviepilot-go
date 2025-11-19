package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/api/handlers/plugin"
	"github.com/yfh-yun/moviepilot-go/internal/middleware"
	"github.com/yfh-yun/moviepilot-go/pkg/plugin"
)

// RegisterPluginManagerRoutes 注册插件管理器路由
func RegisterPluginManagerRoutes(
	router *gin.RouterGroup,
	manager *plugin.HybridPluginManager,
	logger *zap.Logger,
) {
	// 创建处理器
	handler := plugin.NewPluginManagerHandler(manager, logger)

	// 插件管理器路由组
	managerGroup := router.Group("/plugin-manager")
	managerGroup.Use(middleware.Auth())
	managerGroup.Use(middleware.RequestLogger())
	{
		// 插件生命周期管理
		managerGroup.POST("/load", handler.LoadPlugin)
		managerGroup.POST("/:pluginId/initialize", handler.InitializePlugin)
		managerGroup.POST("/:pluginId/start", handler.StartPlugin)
		managerGroup.POST("/:pluginId/stop", handler.StopPlugin)
		managerGroup.POST("/:pluginId/unload", handler.UnloadPlugin)

		// 插件信息和方法调用
		managerGroup.GET("/:pluginId/info", handler.GetPluginInfo)
		managerGroup.POST("/:pluginId/call", handler.CallPluginMethod)
		managerGroup.GET("/plugins", handler.ListPlugins)

		// 事件管理
		managerGroup.POST("/events", handler.PublishEvent)
	}
}