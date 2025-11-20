// Package routes MoviePilot API路由配置模块
// 提供RESTful API路由配置，遵循Go项目标准布局
package routes

import (
	"net/http"
	
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/auth"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/dashboard"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/discover"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/douban"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/download"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/file"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/history"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/media"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/mediaserver"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/message"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/notification"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/plugin"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/recommend"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/scheduler"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/search"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/servarr"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/settings"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/site"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/storage"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/subscribe"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/subscription"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/system"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/torrent"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/transfer"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/user"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/webhook"
	"github.com/yfh-yun/moviepilot-go/internal/apis/handlers/workflow"
	"github.com/yfh-yun/moviepilot-go/internal/apis/middlewares"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	
	"github.com/gin-gonic/gin"
)

// RouterConfig 路由配置结构体
// 集中管理所有处理器，简化依赖注入
type RouterConfig struct {
	BaseHandler        *handlers.BaseHandler
	AuthHandler        *auth.AuthHandler
	DashboardHandler   *dashboard.DashboardHandler
	DiscoverHandler    *discover.DiscoverHandler
	DoubanHandler      *douban.DoubanHandler
	DownloadHandler    *download.DownloadHandler
	FileHandler        *file.FileHandler
	HistoryHandler     *history.HistoryHandler
	MediaHandler       *media.MediaHandler
	MediaServerHandler *mediaserver.MediaServerHandler
	MessageHandler     *message.MessageHandler
	NotificationHandler *notification.NotificationHandler
	PluginHandler      *plugin.PluginHandler
	RecommendHandler   *recommend.RecommendHandler
	SchedulerHandler   *scheduler.SchedulerHandler
	SearchHandler      *search.SearchHandler
	ServarrHandler     *servarr.ServArrHandler
	SettingsHandler    *settings.SettingsHandler
	SiteHandler        *site.SiteHandler
	StorageHandler     *storage.StorageHandler
	SubscribeHandler   *subscribe.SubscribeHandler
	SubscriptionHandler *subscription.SubscriptionHandler
	SystemHandler      *system.SystemHandler
	TorrentHandler     *torrent.TorrentHandler
	TransferHandler    *transfer.TransferHandler
	UserHandler        *user.UserHandler
	WebhookHandler     *webhook.WebhookHandler
	WorkflowHandler    *workflow.WorkflowHandler
}

// SetupRouter 设置主路由
// 初始化Gin引擎并配置所有中间件和路由
func SetupRouter(config *RouterConfig) *gin.Engine {
	// 创建Gin引擎
	router := gin.New()

	// 设置全局中间件
	router.Use(middlewares.RequestIDMiddleware())
	router.Use(middlewares.LoggerMiddleware())
	router.Use(middlewares.RecoveryMiddleware())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.SecurityHeadersMiddleware())
	router.Use(middlewares.ErrorHandlerMiddleware())

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
		setupAuthRoutes(public, config.AuthHandler)
		setupSystemRoutes(public, config.BaseHandler)
		setupDiscoverRoutes(public, config.DiscoverHandler)
		setupDoubanRoutes(public, config.DoubanHandler)
	}

	// 需要认证的路由
	protected := rg.Group("")
	protected.Use(middlewares.AuthMiddleware())
	{
		// 用户相关
		setupUserRoutes(protected, config.UserHandler)
		
		// 媒体相关
		setupMediaRoutes(protected, config.MediaHandler)
		setupMediaServerRoutes(protected, config.MediaServerHandler)
		
		// 下载和传输
		setupDownloadRoutes(protected, config.DownloadHandler)
		setupTorrentRoutes(protected, config.TorrentHandler)
		setupTransferRoutes(protected, config.TransferHandler)
		
		// 文件和存储
		setupFileRoutes(protected, config.FileHandler)
		setupStorageRoutes(protected, config.StorageHandler)
		
		// 订阅相关
		setupSubscribeRoutes(protected, config.SubscribeHandler)
		setupSubscriptionRoutes(protected, config.SubscriptionHandler)
		
		// 站点相关
		setupSiteRoutes(protected, config.SiteHandler)
		
		// 搜索和推荐
		setupSearchRoutes(protected, config.SearchHandler)
		setupRecommendRoutes(protected, config.RecommendHandler)
		
		// 系统功能
		setupHistoryRoutes(protected, config.HistoryHandler)
		setupMessageRoutes(protected, config.MessageHandler)
		setupNotificationRoutes(protected, config.NotificationHandler)
		setupPluginRoutes(protected, config.PluginHandler)
		setupServarrRoutes(protected, config.ServarrHandler)
		setupSchedulerRoutes(protected, config.SchedulerHandler)
		setupWebhookRoutes(protected, config.WebhookHandler)
		setupWorkflowRoutes(protected, config.WorkflowHandler)
		
		// 界面和配置
		setupDashboardRoutes(protected, config.DashboardHandler)
		setupSettingsRoutes(protected, config.SettingsHandler)
		
		// 系统管理
		setupSystemProtectedRoutes(protected, config.SystemHandler)
	}
}

// setupAuthRoutes 认证相关路由
func setupAuthRoutes(rg *gin.RouterGroup, handler *auth.AuthHandler) {
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
func setupUserRoutes(rg *gin.RouterGroup, handler *auth.AuthHandler) {
	user := rg.Group("/user")
	{
		user.GET("/profile", handler.GetUserProfile)
		user.PUT("/profile", handler.UpdateUserProfile)
		user.GET("/preferences", handler.GetUserPreferences)
		user.PUT("/preferences", handler.UpdateUserPreferences)
	}
}

// setupMediaRoutes 媒体管理路由
func setupMediaRoutes(rg *gin.RouterGroup, handler *media.MediaHandler) {
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
func setupDownloadRoutes(rg *gin.RouterGroup, handler *download.DownloadHandler) {
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

// setupDiscoverRoutes 发现内容路由
func setupDiscoverRoutes(rg *gin.RouterGroup, handler *discover.DiscoverHandler) {
	discover := rg.Group("/discover")
	{
		discover.GET("/movies", handler.GetDiscoverMovies)
		discover.GET("/tvs", handler.GetDiscoverTVs)
		discover.GET("/recommendations", handler.GetRecommendations)
	}
}

// setupFileRoutes 文件管理路由
func setupFileRoutes(rg *gin.RouterGroup, handler *file.FileHandler) {
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
}

// setupHistoryRoutes 历史记录路由
func setupHistoryRoutes(rg *gin.RouterGroup, handler *history.HistoryHandler) {
	history := rg.Group("/history")
	{
		history.GET("/download", handler.GetDownloadHistory)
		history.GET("/transfer", handler.GetTransferHistory)
		history.GET("/subscribe", handler.GetSubscribeHistory)
	}
}

// setupMessageRoutes 消息管理路由
func setupMessageRoutes(rg *gin.RouterGroup, handler *message.MessageHandler) {
	message := rg.Group("/message")
	{
		message.GET("/", handler.GetMessageList)
		message.GET("/:id", handler.GetMessage)
		message.PUT("/:id/read", handler.MarkMessageRead)
		message.DELETE("/:id", handler.DeleteMessage)
		message.POST("/send", handler.SendMessage)
	}
}

// setupPluginRoutes 插件管理路由
func setupPluginRoutes(rg *gin.RouterGroup, handler *plugin.PluginHandler) {
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

// setupSearchRoutes 搜索路由
func setupSearchRoutes(rg *gin.RouterGroup, handler *search.SearchHandler) {
	search := rg.Group("/search")
	{
		search.POST("/media", handler.MediaSearch)
		search.POST("/torrent", handler.TorrentSearch)
		search.POST("/site", handler.SiteSearch)
		search.GET("/history", handler.GetSearchHistory)
		search.DELETE("/history", handler.ClearSearchHistory)
		search.GET("/suggestions", handler.GetSearchSuggestions)
		search.GET("/trending", handler.GetTrendingSearches)
	}
}

// setupServarrRoutes ServArr集成路由
func setupServarrRoutes(rg *gin.RouterGroup, handler *servarr.ServArrHandler) {
	servarr := rg.Group("/servarr")
	{
		// 系统相关
		servarr.GET("/system/status", handler.GetSystemStatus)
		
		// 配置相关
		servarr.GET("/qualityProfile", handler.GetQualityProfiles)
		servarr.GET("/rootfolder", handler.GetRootFolders)
		servarr.GET("/tag", handler.GetTags)
		servarr.GET("/languageprofile", handler.GetLanguageProfiles)
		
		// 电影相关
		servarr.GET("/movie", handler.GetMovies)
		servarr.GET("/movie/lookup", handler.LookupMovie)
		servarr.GET("/movie/:mid", handler.GetMovie)
		servarr.POST("/movie", handler.AddMovie)
		servarr.DELETE("/movie/:mid", handler.DeleteMovie)
		
		// 剧集相关
		servarr.GET("/series", handler.GetSeries)
		// TODO: 实现更多剧集相关API
	}
}

// setupDashboardRoutes 仪表板路由
func setupDashboardRoutes(rg *gin.RouterGroup, handler *dashboard.DashboardHandler) {
	dashboard := rg.Group("/dashboard")
	{
		dashboard.GET("/overview", handler.GetDashboardOverview)
		dashboard.GET("/statistics", handler.GetDashboardStatistics)
		dashboard.GET("/recent", handler.GetRecentActivities)
	}
}

// setupDoubanRoutes 豆瓣相关路由
func setupDoubanRoutes(rg *gin.RouterGroup, handler *douban.DoubanHandler) {
	douban := rg.Group("/douban")
	{
		douban.GET("/movie/:id", handler.GetDoubanMovie)
		douban.GET("/tv/:id", handler.GetDoubanTV)
		douban.GET("/search", handler.SearchDouban)
	}
}

// setupMediaServerRoutes 媒体服务器路由
func setupMediaServerRoutes(rg *gin.RouterGroup, handler *mediaserver.MediaServerHandler) {
	mediaserver := rg.Group("/mediaserver")
	{
		mediaserver.GET("/", handler.GetMediaServerList)
		mediaserver.GET("/:id", handler.GetMediaServer)
		mediaserver.POST("/", handler.CreateMediaServer)
		mediaserver.PUT("/:id", handler.UpdateMediaServer)
		mediaserver.DELETE("/:id", handler.DeleteMediaServer)
		mediaserver.POST("/:id/test", handler.TestMediaServer)
		mediaserver.POST("/:id/sync", handler.SyncMediaServer)
	}
}

// setupNotificationRoutes 通知管理路由
func setupNotificationRoutes(rg *gin.RouterGroup, handler *notification.NotificationHandler) {
	notification := rg.Group("/notification")
	{
		notification.GET("/", handler.GetNotificationList)
		notification.GET("/:id", handler.GetNotification)
		notification.POST("/", handler.CreateNotification)
		notification.PUT("/:id", handler.UpdateNotification)
		notification.DELETE("/:id", handler.DeleteNotification)
		notification.POST("/send", handler.SendNotification)
		notification.POST("/test", handler.TestNotification)
	}
}

// setupRecommendRoutes 推荐路由
func setupRecommendRoutes(rg *gin.RouterGroup, handler *recommend.RecommendHandler) {
	recommend := rg.Group("/recommend")
	{
		recommend.GET("/movies", handler.GetMovieRecommendations)
		recommend.GET("/tvs", handler.GetTVRecommendations)
		recommend.GET("/personal", handler.GetPersonalRecommendations)
		recommend.POST("/feedback", handler.SubmitRecommendationFeedback)
	}
}

// setupSchedulerRoutes 调度器路由
func setupSchedulerRoutes(rg *gin.RouterGroup, handler *scheduler.SchedulerHandler) {
	scheduler := rg.Group("/scheduler")
	{
		scheduler.GET("/jobs", handler.GetJobs)
		scheduler.GET("/jobs/:id", handler.GetJob)
		scheduler.POST("/jobs", handler.CreateJob)
		scheduler.PUT("/jobs/:id", handler.UpdateJob)
		scheduler.DELETE("/jobs/:id", handler.DeleteJob)
		scheduler.POST("/jobs/:id/run", handler.RunJob)
		scheduler.POST("/jobs/:id/stop", handler.StopJob)
	}
}

// setupSettingsRoutes 设置管理路由
func setupSettingsRoutes(rg *gin.RouterGroup, handler *settings.SettingsHandler) {
	settings := rg.Group("/settings")
	{
		settings.GET("/", handler.GetSettings)
		settings.PUT("/", handler.UpdateSettings)
		settings.GET("/backup", handler.BackupSettings)
		settings.POST("/restore", handler.RestoreSettings)
		settings.POST("/reset", handler.ResetSettings)
	}
}

// setupSiteRoutes 站点管理路由
func setupSiteRoutes(rg *gin.RouterGroup, handler *site.SiteHandler) {
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
}

// setupStorageRoutes 存储管理路由
func setupStorageRoutes(rg *gin.RouterGroup, handler *storage.StorageHandler) {
	storage := rg.Group("/storage")
	{
		storage.GET("/info", handler.GetStorageInfo)
		storage.GET("/usage", handler.GetStorageUsage)
		storage.POST("/cleanup", handler.CleanupStorage)
		storage.POST("/analyze", handler.AnalyzeStorage)
	}
}

// setupSubscribeRoutes 订阅路由
func setupSubscribeRoutes(rg *gin.RouterGroup, handler *subscribe.SubscribeHandler) {
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

// setupSubscriptionRoutes 订阅管理路由
func setupSubscriptionRoutes(rg *gin.RouterGroup, handler *subscription.SubscriptionHandler) {
	subscription := rg.Group("/subscription")
	{
		subscription.GET("/", handler.GetSubscriptionList)
		subscription.GET("/:id", handler.GetSubscription)
		subscription.POST("/", handler.CreateSubscription)
		subscription.PUT("/:id", handler.UpdateSubscription)
		subscription.DELETE("/:id", handler.DeleteSubscription)
		subscription.POST("/:id/enable", handler.EnableSubscription)
		subscription.POST("/:id/disable", handler.DisableSubscription)
	}
}

// setupSystemProtectedRoutes 系统管理路由（需要认证）
func setupSystemProtectedRoutes(rg *gin.RouterGroup, handler *system.SystemHandler) {
	system := rg.Group("/system")
	{
		system.GET("/info", handler.GetSystemInfo)
		system.GET("/status", handler.GetSystemStatus)
		system.GET("/logs", handler.GetSystemLogs)
		system.POST("/restart", handler.RestartSystem)
		system.POST("/shutdown", handler.ShutdownSystem)
	}
}

// setupTorrentRoutes 种子管理路由
func setupTorrentRoutes(rg *gin.RouterGroup, handler *torrent.TorrentHandler) {
	torrent := rg.Group("/torrent")
	{
		torrent.GET("/", handler.GetTorrentList)
		torrent.GET("/:id", handler.GetTorrent)
		torrent.POST("/add", handler.AddTorrent)
		torrent.DELETE("/:id", handler.DeleteTorrent)
		torrent.POST("/:id/start", handler.StartTorrent)
		torrent.POST("/:id/stop", handler.StopTorrent)
		torrent.POST("/:id/pause", handler.PauseTorrent)
		torrent.POST("/:id/resume", handler.ResumeTorrent)
	}
}

// setupTransferRoutes 传输管理路由
func setupTransferRoutes(rg *gin.RouterGroup, handler *transfer.TransferHandler) {
	transfer := rg.Group("/transfer")
	{
		transfer.GET("/", handler.GetTransferList)
		transfer.GET("/:id", handler.GetTransfer)
		transfer.POST("/", handler.CreateTransfer)
		transfer.DELETE("/:id", handler.DeleteTransfer)
		transfer.POST("/:id/retry", handler.RetryTransfer)
	}
}

// setupUserRoutes 用户管理路由
func setupUserRoutes(rg *gin.RouterGroup, handler *user.UserHandler) {
	user := rg.Group("/user")
	{
		user.GET("/profile", handler.GetUserProfile)
		user.PUT("/profile", handler.UpdateUserProfile)
		user.GET("/preferences", handler.GetUserPreferences)
		user.PUT("/preferences", handler.UpdateUserPreferences)
		user.GET("/avatar", handler.GetUserAvatar)
		user.POST("/avatar", handler.UpdateUserAvatar)
	}
}

// setupWebhookRoutes Webhook路由
func setupWebhookRoutes(rg *gin.RouterGroup, handler *webhook.WebhookHandler) {
	webhook := rg.Group("/webhook")
	{
		webhook.GET("/", handler.GetWebhookList)
		webhook.GET("/:id", handler.GetWebhook)
		webhook.POST("/", handler.CreateWebhook)
		webhook.PUT("/:id", handler.UpdateWebhook)
		webhook.DELETE("/:id", handler.DeleteWebhook)
		webhook.POST("/:id/test", handler.TestWebhook)
	}
}

// setupWorkflowRoutes 工作流路由
func setupWorkflowRoutes(rg *gin.RouterGroup, handler *workflow.WorkflowHandler) {
	workflow := rg.Group("/workflow")
	{
		workflow.GET("/", handler.GetWorkflowList)
		workflow.GET("/:id", handler.GetWorkflow)
		workflow.POST("/", handler.CreateWorkflow)
		workflow.PUT("/:id", handler.UpdateWorkflow)
		workflow.DELETE("/:id", handler.DeleteWorkflow)
		workflow.POST("/:id/run", handler.RunWorkflow)
		workflow.POST("/:id/stop", handler.StopWorkflow)
	}
}
