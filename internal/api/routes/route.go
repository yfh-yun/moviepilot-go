// Package routes MoviePilot API路由配置模块
// 提供RESTful API路由配置，遵循Go项目标准布局
package routes

import (
	"github.com/yfh-yun/moviepilot-go/internal/api/handlers"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouterConfig 路由配置结构体
// 集中管理所有处理器，简化依赖注入
type RouterConfig struct {
	BaseHandler *handlers.BaseHandler
}

// SetupRouter 设置主路由
// 初始化Gin引擎并配置所有中间件和路由
func SetupRouter(config *RouterConfig) *gin.Engine {
	// 创建Gin引擎
	router := gin.New()

	// 设置全局中间件
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())

	// API版本1路由组
	v1 := router.Group("/api/v1")
	setupV1Routes(v1, config)

	// 健康检查路由（不需要版本控制）
	router.GET("/health", healthCheck)
	router.GET("/version", getVersion)

	logger.Info("Router setup completed")
	return router
}

// setupV1Routes 设置v1版本API路由
func setupV1Routes(rg *gin.RouterGroup, config *RouterConfig) {
	if config == nil {
		logger.Error("Router config is nil")
		return
	}

	// 公开路由（不需要认证）
	public := rg.Group("")
	{
		setupAuthRoutes(public, config.BaseHandler)
		setupSystemRoutes(public, config.BaseHandler)
	}

	// 需要认证的路由
	protected := rg.Group("")
	// TODO: 添加JWT认证中间件
	// protected.Use(middleware.JWTAuthMiddleware())
	{
		setupUserRoutes(protected, config.BaseHandler)
		setupMediaRoutes(protected, config.BaseHandler)
		setupDownloadRoutes(protected, config.BaseHandler)
		setupSiteRoutes(protected, config.BaseHandler)
		setupPluginRoutes(protected, config.BaseHandler)
		setupSystemProtectedRoutes(protected, config.BaseHandler)
	}
}

// setupAuthRoutes 认证相关路由
func setupAuthRoutes(rg *gin.RouterGroup, handler *handlers.BaseHandler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/logout", handler.Logout)
		auth.POST("/refresh", handler.RefreshToken)
		auth.GET("/wallpaper", handler.GetWallpaper)
		auth.GET("/wallpapers", handler.GetWallpapers)
	}
}

// setupSystemRoutes 系统公开路由
func setupSystemRoutes(rg *gin.RouterGroup, handler *handlers.BaseHandler) {
	system := rg.Group("/system")
	{
		system.GET("/info", handler.GetSystemInfo)
		system.GET("/health", handler.HealthCheck)
		system.GET("/version", handler.GetVersion)
	}
}

// setupUserRoutes 用户管理路由
func setupUserRoutes(rg *gin.RouterGroup, handler *handlers.BaseHandler) {
	user := rg.Group("/user")
	{
		user.GET("/profile", handler.GetUserProfile)
		user.PUT("/profile", handler.UpdateUserProfile)
		user.GET("/preferences", handler.GetUserPreferences)
		user.PUT("/preferences", handler.UpdateUserPreferences)
	}
}

// setupMediaRoutes 媒体管理路由
func setupMediaRoutes(rg *gin.RouterGroup, handler *handlers.BaseHandler) {
	media := rg.Group("/media")
	{
		media.GET("/", handler.GetMediaList)
		media.GET("/:id", handler.GetMedia)
		media.POST("/", handler.CreateMedia)
		media.PUT("/:id", handler.UpdateMedia)
		media.DELETE("/:id", handler.DeleteMedia)
		media.GET("/search", handler.SearchMedia)
		media.POST("/recognize", handler.RecognizeMedia)
	}

	// TMDB相关
	tmdb := rg.Group("/tmdb")
	{
		tmdb.GET("/movie/:id", handler.GetTMDBMovie)
		tmdb.GET("/tv/:id", handler.GetTMDBTV)
		tmdb.GET("/person/:id", handler.GetTMDBPerson)
		tmdb.GET("/search/movie", handler.SearchTMDBMovies)
		tmdb.GET("/search/tv", handler.SearchTMDBTVs)
		tmdb.GET("/search/person", handler.SearchTMDBPersons)
	}

	// 豆瓣相关
	douban := rg.Group("/douban")
	{
		douban.GET("/movie/:id", handler.GetDoubanMovie)
		douban.GET("/tv/:id", handler.GetDoubanTV)
		douban.GET("/search", handler.SearchDouban)
	}
}

// setupDownloadRoutes 下载管理路由
func setupDownloadRoutes(rg *gin.RouterGroup, handler *handlers.BaseHandler) {
	download := rg.Group("/download")
	{
		download.GET("/", handler.GetDownloadList)
		download.GET("/:id", handler.GetDownload)
		download.POST("/", handler.CreateDownload)
		download.DELETE("/:id", handler.DeleteDownload)
		download.POST("/:id/pause", handler.PauseDownload)
		download.POST("/:id/resume", handler.ResumeDownload)
	}

	// 种子管理
	torrent := rg.Group("/torrent")
	{
		torrent.GET("/", handler.GetTorrentList)
		torrent.GET("/:id", handler.GetTorrent)
		torrent.POST("/add", handler.AddTorrent)
		torrent.DELETE("/:id", handler.DeleteTorrent)
		torrent.POST("/:id/start", handler.StartTorrent)
		torrent.POST("/:id/stop", handler.StopTorrent)
	}
}

// setupSiteRoutes 站点管理路由
func setupSiteRoutes(rg *gin.RouterGroup, handler *handlers.BaseHandler) {
	site := rg.Group("/site")
	{
		site.GET("/", handler.GetSiteList)
		site.GET("/:id", handler.GetSite)
		site.POST("/", handler.CreateSite)
		site.PUT("/:id", handler.UpdateSite)
		site.DELETE("/:id", handler.DeleteSite)
		site.POST("/:id/test", handler.TestSite)
		site.GET("/:id/statistics", handler.GetSiteStatistics)
	}

	// 订阅管理
	subscribe := rg.Group("/subscribe")
	{
		subscribe.GET("/", handler.GetSubscribeList)
		subscribe.GET("/:id", handler.GetSubscribe)
		subscribe.POST("/", handler.CreateSubscribe)
		subscribe.PUT("/:id", handler.UpdateSubscribe)
		subscribe.DELETE("/:id", handler.DeleteSubscribe)
		subscribe.POST("/:id/refresh", handler.RefreshSubscribe)
		subscribe.GET("/search", handler.SearchSubscribe)
	}
}

// setupPluginRoutes 插件管理路由
func setupPluginRoutes(rg *gin.RouterGroup, handler *handlers.BaseHandler) {
	plugin := rg.Group("/plugin")
	{
		plugin.GET("/", handler.GetPluginList)
		plugin.GET("/:id", handler.GetPlugin)
		plugin.POST("/:id/install", handler.InstallPlugin)
		plugin.DELETE("/:id", handler.UninstallPlugin)
		plugin.POST("/:id/enable", handler.EnablePlugin)
		plugin.POST("/:id/disable", handler.DisablePlugin)
		plugin.POST("/:id/config", handler.ConfigurePlugin)
	}
}

// setupSystemProtectedRoutes 系统管理路由（需要认证）
func setupSystemProtectedRoutes(rg *gin.RouterGroup, handler *handlers.BaseHandler) {
	// 系统配置
	config := rg.Group("/config")
	{
		config.GET("/", handler.GetSystemConfig)
		config.PUT("/", handler.UpdateSystemConfig)
		config.GET("/backup", handler.BackupConfig)
		config.POST("/restore", handler.RestoreConfig)
	}

	// 消息管理
	message := rg.Group("/message")
	{
		message.GET("/", handler.GetMessageList)
		message.GET("/:id", handler.GetMessage)
		message.PUT("/:id/read", handler.MarkMessageRead)
		message.DELETE("/:id", handler.DeleteMessage)
		message.POST("/send", handler.SendMessage)
	}

	// 文件管理
	file := rg.Group("/file")
	{
		file.GET("/exists", handler.FileExists)
		file.GET("/list", handler.ListDirectory)
		file.GET("/info", handler.GetFileInfo)
		file.GET("/hash", handler.GetFileHash)
		file.GET("/read", handler.ReadFile)
		file.POST("/write", handler.WriteFile)
		file.POST("/create-dir", handler.CreateDirectory)
		file.PUT("/move", handler.MoveFile)
		file.POST("/copy", handler.CopyFile)
		file.DELETE("/delete", handler.DeleteFile)
	}

	// 历史记录
	history := rg.Group("/history")
	{
		history.GET("/download", handler.GetDownloadHistory)
		history.GET("/transfer", handler.GetTransferHistory)
		history.GET("/subscribe", handler.GetSubscribeHistory)
	}

	// 搜索
	search := rg.Group("/search")
	{
		search.GET("/media", handler.SearchMediaItems)
		search.GET("/torrent", handler.SearchTorrents)
		search.GET("/suggest", handler.GetSearchSuggestions)
	}

	// 仪表板
	dashboard := rg.Group("/dashboard")
	{
		dashboard.GET("/overview", handler.GetDashboardOverview)
		dashboard.GET("/statistics", handler.GetDashboardStatistics)
		dashboard.GET("/recent", handler.GetRecentActivities)
	}
}

// healthCheck 健康检查端点
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "MoviePilot API is running",
	})
}

// getVersion 获取版本信息端点
func getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": "2.8.1",
		"name":    "MoviePilot",
		"mode":    gin.Mode(),
	})
}
