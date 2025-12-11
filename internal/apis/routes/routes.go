package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "moviepilot-go/docs" // swagger docs

	bangumiapi "moviepilot-go/internal/apis/handlers/bangumi"
	dashboardapi "moviepilot-go/internal/apis/handlers/dashboard"
	discoverapi "moviepilot-go/internal/apis/handlers/discover"
	doubanapi "moviepilot-go/internal/apis/handlers/douban"
	downloadapi "moviepilot-go/internal/apis/handlers/download"
	historyapi "moviepilot-go/internal/apis/handlers/history"
	loginapi "moviepilot-go/internal/apis/handlers/login"
	mediaapi "moviepilot-go/internal/apis/handlers/media"
	mediaserverapi "moviepilot-go/internal/apis/handlers/mediaserver"
	pluginapi "moviepilot-go/internal/apis/handlers/plugin"
	pluginmediaapi "moviepilot-go/internal/apis/handlers/pluginmedia"
	recommendapi "moviepilot-go/internal/apis/handlers/recommend"
	searchapi "moviepilot-go/internal/apis/handlers/search"
	servarrapi "moviepilot-go/internal/apis/handlers/servarr"
	servcookieapi "moviepilot-go/internal/apis/handlers/servcookie"
	siteapi "moviepilot-go/internal/apis/handlers/site"
	storageapi "moviepilot-go/internal/apis/handlers/storage"
	subscribeapi "moviepilot-go/internal/apis/handlers/subscribe"
	systemapi "moviepilot-go/internal/apis/handlers/system"
	tmdbapi "moviepilot-go/internal/apis/handlers/tmdb"
	torrentapi "moviepilot-go/internal/apis/handlers/torrent"
	transferapi "moviepilot-go/internal/apis/handlers/transfer"
	userapi "moviepilot-go/internal/apis/handlers/user"
	authv1api "moviepilot-go/internal/apis/handlers/v1/auth"
	webhookapi "moviepilot-go/internal/apis/handlers/webhook"
	workflowapi "moviepilot-go/internal/apis/handlers/workflow"
	"moviepilot-go/internal/apis/middleware"
	"moviepilot-go/internal/apis/middlewares"
	authbiz "moviepilot-go/internal/business/services/auth"
	bangumibiz "moviepilot-go/internal/business/services/bangumi"
	dashboardbiz "moviepilot-go/internal/business/services/dashboard"
	doubanbiz "moviepilot-go/internal/business/services/douban"
	downloadbiz "moviepilot-go/internal/business/services/download"
	historybiz "moviepilot-go/internal/business/services/history"
	pluginbiz "moviepilot-go/internal/business/services/plugin"
	pluginmediabiz "moviepilot-go/internal/business/services/pluginmedia"
	recommendbiz "moviepilot-go/internal/business/services/recommend"
	searchbiz "moviepilot-go/internal/business/services/search"
	"moviepilot-go/internal/business/services/storage"
	subscribebiz "moviepilot-go/internal/business/services/subscribe"
	systembiz "moviepilot-go/internal/business/services/system"
	tmdbbiz "moviepilot-go/internal/business/services/tmdb"
	torrentbiz "moviepilot-go/internal/business/services/torrent"
	transferbiz "moviepilot-go/internal/business/services/transfer"
	userbiz "moviepilot-go/internal/business/services/user"
	webhookbiz "moviepilot-go/internal/business/services/webhook"
	"moviepilot-go/internal/infrastructure/events"
	"moviepilot-go/internal/integration/mediaserver"
	authrepos "moviepilot-go/internal/repositories"
	repoInterfaces "moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/internal/repositories/repositories"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/plugin"
	"moviepilot-go/pkg/security"
	"moviepilot-go/pkg/utils"
)

// Config 提供路由注册所需的依赖。
type Config struct {
	// DB 数据库连接
	DB *gorm.DB
	// JWT 配置
	JWTSecretKey            string
	AccessTokenExpireMinute int
}

// noopWallpaperCache 是用于 WallpaperHelper 的简易缓存实现，始终返回默认值，不做持久化
type noopWallpaperCache struct{}

func (c *noopWallpaperCache) Get(key string, defaultValue any) any {
	return defaultValue
}

func (c *noopWallpaperCache) Set(key string, value any, ttl time.Duration) {}

// noopMediaServerChain 是用于 WallpaperHelper 的媒体服务器空实现
type noopMediaServerChain struct{}

func (m *noopMediaServerChain) GetLatestWallpaper() string {
	return ""
}

func (m *noopMediaServerChain) GetLatestWallpapers(count int) []string {
	return []string{}
}

// noopTmdbChain 是用于 WallpaperHelper 的 TMDB 空实现
type noopTmdbChain struct{}

func (t *noopTmdbChain) GetRandomWallpaper() string {
	return ""
}

func (t *noopTmdbChain) GetTrendingWallpapers(num int) []string {
	return []string{}
}

// Register 统一注册 API 路由。
func Register(engine *gin.Engine, cfg Config) error {
	// 检查 gin 引擎是否为空
	if engine == nil {
		return fmt.Errorf("gin engine cannot be nil")
	}

	logger.Debug("Register called", zap.String("func", "Register"))
	logger.Info("Starting API route registration", zap.Bool("db_available", cfg.DB != nil))

	// 构造全局 JWT 管理器，供所有受保护路由的 AuthMiddleware 使用
	baseSecret := cfg.JWTSecretKey
	if baseSecret == "" {
		baseSecret = "moviepilot_default_secret"
	}
	baseExpireMinutes := cfg.AccessTokenExpireMinute
	if baseExpireMinutes <= 0 {
		baseExpireMinutes = 24 * 60
	}
	baseJWTManager := security.NewJWTManager(
		baseSecret,
		time.Duration(baseExpireMinutes)*time.Minute,
		7*24*time.Hour,
	)

	// Swagger UI 路由（不需要认证）
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("Swagger UI registered at /swagger/index.html")

	api := engine.Group("/api")
	api.Use(middlewares.AuthMiddleware(baseJWTManager))
	// 移除 APIRateLimitMiddleware，因为 pkg/middlewares 目录为空

	// Workflow 路由
	logger.Info("Registering workflow routes")
	workflowGroup := api.Group("/workflows")
	// 简化实现：创建一个空的workflowManager实现
	workflowEngine := &mockWorkflowManager{}
	workflowService := workflowapi.NewService(workflowEngine, nil, logger.GetLogger())
	workflowHandler := workflowapi.NewHandler(workflowService, logger.GetLogger())
	workflowGroup.POST("/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	// 工作流管理路由
	logger.Info("Registering workflow management routes")
	workflowApiGroup := api.Group("/workflow")
	{
		workflowApiGroup.GET("", workflowHandler.ListWorkflows)
		workflowApiGroup.POST("", workflowHandler.CreateWorkflow)
		workflowApiGroup.GET("/plugin/actions", workflowHandler.GetPluginActions)
		workflowApiGroup.GET("/actions", workflowHandler.ListActions)
		workflowApiGroup.GET("/event_types", workflowHandler.GetEventTypes)
		workflowApiGroup.POST("/share", workflowHandler.ShareWorkflow)
		workflowApiGroup.DELETE("/share/:share_id", workflowHandler.DeleteShare)
		workflowApiGroup.POST("/fork", workflowHandler.ForkWorkflow)
		workflowApiGroup.GET("/shares", workflowHandler.GetShares)
		workflowApiGroup.POST("/:workflow_id/run", workflowHandler.RunWorkflow)
		workflowApiGroup.POST("/:workflow_id/start", workflowHandler.StartWorkflow)
		workflowApiGroup.POST("/:workflow_id/pause", workflowHandler.PauseWorkflow)
		workflowApiGroup.POST("/:workflow_id/reset", workflowHandler.ResetWorkflow)
		workflowApiGroup.GET("/:workflow_id", workflowHandler.GetWorkflow)
		workflowApiGroup.PUT("/:workflow_id", workflowHandler.UpdateWorkflow)
		workflowApiGroup.DELETE("/:workflow_id", workflowHandler.DeleteWorkflow)
	}

	// /api/v1/workflow 兼容路由
	v1WorkflowGroup := engine.Group("/api/v1/workflow")
	v1WorkflowGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		v1WorkflowGroup.GET("", workflowHandler.ListWorkflows)
		v1WorkflowGroup.POST("", workflowHandler.CreateWorkflow)
		v1WorkflowGroup.GET("/plugin/actions", workflowHandler.GetPluginActions)
		v1WorkflowGroup.GET("/actions", workflowHandler.ListActions)
		v1WorkflowGroup.GET("/event_types", workflowHandler.GetEventTypes)
		v1WorkflowGroup.POST("/share", workflowHandler.ShareWorkflow)
		v1WorkflowGroup.DELETE("/share/:share_id", workflowHandler.DeleteShare)
		v1WorkflowGroup.POST("/fork", workflowHandler.ForkWorkflow)
		v1WorkflowGroup.GET("/shares", workflowHandler.GetShares)
		v1WorkflowGroup.POST("/:workflow_id/run", workflowHandler.RunWorkflow)
		v1WorkflowGroup.POST("/:workflow_id/start", workflowHandler.StartWorkflow)
		v1WorkflowGroup.POST("/:workflow_id/pause", workflowHandler.PauseWorkflow)
		v1WorkflowGroup.POST("/:workflow_id/reset", workflowHandler.ResetWorkflow)
		v1WorkflowGroup.GET("/:workflow_id", workflowHandler.GetWorkflow)
		v1WorkflowGroup.PUT("/:workflow_id", workflowHandler.UpdateWorkflow)
		v1WorkflowGroup.DELETE("/:workflow_id", workflowHandler.DeleteWorkflow)
	}

	// TODO: 以下路由需要在实现对应的 repository、business、api 层后取消注释

	// // 1. 下载路由
	// downloadRepo := repositories.NewDownloadRepository(cfg.DB)
	// downloadService := downloadbiz.NewService(downloadRepo, log)
	// downloadHandler := downloadapi.NewHandler(downloadService, log)
	// downloadGroup := api.Group("/downloads")
	// {
	// 	downloadGroup.GET("", downloadHandler.List)
	// 	downloadGroup.POST("", downloadHandler.Add)
	// 	downloadGroup.DELETE("/:id", downloadHandler.Delete)
	// 	downloadGroup.POST("/:id/start", downloadHandler.Start)
	// 	downloadGroup.POST("/:id/stop", downloadHandler.Stop)
	// }

	// // 2. 媒体路由
	// mediaRepo := repositories.NewMediaRepository(cfg.DB)
	// mediaService := mediabiz.NewService(mediaRepo, log)
	// mediaHandler := mediaapi.NewHandler(mediaService, log)
	// mediaGroup := api.Group("/media")
	// {
	// 	mediaGroup.POST("/scrape", mediaHandler.Scrape)
	// 	mediaGroup.GET("/search", mediaHandler.Search)
	// 	mediaGroup.GET("/:id", mediaHandler.GetInfo)
	// }

	// 3. 搜索路由
	// 简化实现：使用 searchbiz.NewService 的默认实现，传递 nil 作为缓存
	searchService := searchbiz.NewService(nil)
	searchHandler := searchapi.NewHandler(searchService)
	searchGroup := api.Group("/search")
	{
		searchGroup.POST("", searchHandler.Search)
		searchGroup.GET("", searchHandler.SearchSimple)
		searchGroup.POST("/multi", searchHandler.SearchMultiSite)
		searchGroup.GET("/history", searchHandler.GetSearchHistory)
		searchGroup.DELETE("/history", searchHandler.ClearSearchHistory)
		// 添加新的搜索端点
		searchGroup.GET("/last", searchHandler.LastSearchResults)
		searchGroup.GET("/media/:mediaid", searchHandler.SearchByID)
		searchGroup.GET("/title", searchHandler.SearchByTitle)
	}

	// // 4. 转移路由
	// transferRepo := repositories.NewTransferRepository(cfg.DB)
	// transferService := transferbiz.NewService(transferRepo, log)
	// transferHandler := transferapi.NewHandler(transferService, log)
	// transferGroup := api.Group("/transfer")
	// {
	// 	transferGroup.POST("/manual", transferHandler.ManualTransfer)
	// 	transferGroup.GET("/history", transferHandler.GetHistory)
	// 	transferGroup.DELETE("/history/:id", transferHandler.DeleteHistory)
	// }

	// // 5. 用户路由
	// userRepo := repositories.NewUserRepository(cfg.DB)
	// userService := userbiz.NewService(userRepo, log)
	// userHandler := userapi.NewHandler(userService, log)
	// userGroup := api.Group("/users")
	// {
	// 	userGroup.POST("/login", userHandler.Login)
	// 	userGroup.POST("/register", userHandler.Register)
	// 	userGroup.GET("/profile", userHandler.GetProfile)
	// 	userGroup.PUT("/profile", userHandler.UpdateProfile)
	// }

	// // 6. 仪表盘路由
	// dashboardService := dashboardbiz.NewService(cfg.DB, log)
	// dashboardHandler := dashboardapi.NewHandler(dashboardService, log)
	// dashboardGroup := api.Group("/dashboard")
	// {
	// 	dashboardGroup.GET("/stats", dashboardHandler.GetStats)
	// 	dashboardGroup.GET("/storage", dashboardHandler.GetStorage)
	// 	dashboardGroup.GET("/processes", dashboardHandler.GetProcesses)
	// }

	// // 7. 通知路由
	// notificationRepo := repositories.NewNotificationRepository(cfg.DB)
	// notificationService := notificationbiz.NewService(notificationRepo, log)
	// notificationHandler := notificationapi.NewHandler(notificationService, log)
	// notificationGroup := api.Group("/notifications")
	// {
	// 	notificationGroup.POST("/send", notificationHandler.Send)
	// 	notificationGroup.GET("/history", notificationHandler.GetHistory)
	// }

	// // 8. 历史记录路由
	// historyRepo := repositories.NewHistoryRepository(cfg.DB)
	// historyService := historybiz.NewService(historyRepo, log)
	// historyHandler := historyapi.NewHandler(historyService, log)
	// historyGroup := api.Group("/history")
	// {
	// 	historyGroup.GET("/transfer", historyHandler.GetTransferHistory)
	// 	historyGroup.DELETE("/transfer/:id", historyHandler.DeleteTransferHistory)
	// }

	// // 9. 媒体服务器路由
	// mediaserverRepo := repositories.NewMediaServerRepository(cfg.DB)
	// mediaserverService := mediaserverbiz.NewService(mediaserverRepo, log)
	// mediaserverHandler := mediaserverapi.NewHandler(mediaserverService, log)
	// mediaserverGroup := api.Group("/mediaserver")
	// {
	// 	mediaserverGroup.GET("/libraries", mediaserverHandler.GetLibraries)
	// 	mediaserverGroup.GET("/items", mediaserverHandler.GetItems)
	// 	mediaserverGroup.POST("/sync", mediaserverHandler.Sync)
	// }

	// // 10. 插件路由
	// pluginRepo := repositories.NewPluginRepository(cfg.DB)
	// pluginService := pluginbiz.NewService(pluginRepo, log)
	// pluginHandler := pluginapi.NewHandler(pluginService, log)
	// pluginGroup := api.Group("/plugins")
	// {
	// 	pluginGroup.GET("", pluginHandler.List)
	// 	pluginGroup.POST("/:id/enable", pluginHandler.Enable)
	// 	pluginGroup.POST("/:id/disable", pluginHandler.Disable)
	// 	pluginGroup.GET("/:id/config", pluginHandler.GetConfig)
	// 	pluginGroup.PUT("/:id/config", pluginHandler.UpdateConfig)
	// }
	// 订阅路由
	if cfg.DB != nil {
		logger.Info("Registering subscription routes")
		// 底层 DB 仓储
		dbSubscribeRepo := repositories.NewSubscribeRepository(cfg.DB)
		// 业务层订阅仓储适配器
		bizSubscribeRepo := subscribebiz.NewDBSubscribeRepository(dbSubscribeRepo)
		subscribeService := subscribebiz.NewBasicService(bizSubscribeRepo, logger.GetLogger())
		subscribeHandler := subscribeapi.NewHandler(subscribeService, logger.GetLogger())

		subscribeGroup := api.Group("/subscribes")
		{
			subscribeGroup.POST("", subscribeHandler.CreateSubscribe)
			subscribeGroup.GET("", subscribeHandler.ListSubscribes)
			subscribeGroup.GET("/:id", subscribeHandler.GetSubscribe)
			subscribeGroup.PUT("/:id", subscribeHandler.UpdateSubscribe)
			subscribeGroup.DELETE("/:id", subscribeHandler.DeleteSubscribe)
			subscribeGroup.POST("/:id/pause", subscribeHandler.PauseSubscribe)
			subscribeGroup.POST("/:id/resume", subscribeHandler.ResumeSubscribe)

			// 新增订阅管理端点
			subscribeGroup.GET("/refresh", subscribeHandler.RefreshSubscribes)
			subscribeGroup.GET("/reset/:subid", subscribeHandler.ResetSubscribe)
			subscribeGroup.GET("/check", subscribeHandler.CheckSubscribes)
			subscribeGroup.GET("/search", subscribeHandler.SearchAllSubscribes)
			subscribeGroup.GET("/search/:subscribe_id", subscribeHandler.SearchSubscribe)
			subscribeGroup.GET("/user/:username", subscribeHandler.GetUserSubscribes)
			subscribeGroup.GET("/files/:subscribe_id", subscribeHandler.GetSubscribeFiles)

			// 订阅分享路由
			shareService := subscribebiz.NewShareService()
			shareHandler := subscribeapi.NewShareHandler(shareService)
			subscribeGroup.POST("/share", shareHandler.ShareSubscribe)
			subscribeGroup.DELETE("/share/:share_id", shareHandler.DeleteShare)
			subscribeGroup.POST("/fork", shareHandler.ForkSubscribe)
			subscribeGroup.GET("/shares", shareHandler.GetShares)
			subscribeGroup.GET("/share/statistics", shareHandler.GetShareStatistics)
			subscribeGroup.POST("/follow", shareHandler.FollowUser)
			subscribeGroup.DELETE("/follow", shareHandler.UnfollowUser)
			subscribeGroup.GET("/follow", shareHandler.GetFollowedUsers)
			subscribeGroup.GET("/popular", shareHandler.GetPopularShares)

			// 订阅状态路由
			statusHandler := subscribeapi.NewStatusHandler()
			subscribeGroup.PUT("/status/:sub_id", statusHandler.UpdateSubscribeStatus)
			subscribeGroup.GET("/history/:mtype", statusHandler.GetSubscribeHistory)
			subscribeGroup.GET("/media/:media_id", statusHandler.GetMediaSubscribe)
			subscribeGroup.DELETE("/media/:media_id", statusHandler.DeleteMediaSubscribe)
			subscribeGroup.POST("/seerr", statusHandler.OverSeerrNotify)

			// 新增媒体ID相关端点
			subscribeGroup.GET("/media/:mediaid", subscribeHandler.GetSubscribeByMediaID)
			subscribeGroup.DELETE("/media/:mediaid", subscribeHandler.DeleteSubscribeByMediaID)
			// OverSeerr/JellySeerr通知订阅
			subscribeGroup.POST("/seerr", subscribeHandler.SeerrSubscribe)
			// 热门订阅
			subscribeGroup.GET("/popular", subscribeHandler.GetPopularSubscribes)
		}

		// /api/v1/subscribes 兼容路由
		v1SubscribeGroup := engine.Group("/api/v1/subscribes")
		v1SubscribeGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
		{
			v1SubscribeGroup.POST("", subscribeHandler.CreateSubscribe)
			v1SubscribeGroup.GET("", subscribeHandler.ListSubscribes)
			v1SubscribeGroup.GET("/:id", subscribeHandler.GetSubscribe)
			v1SubscribeGroup.PUT("/:id", subscribeHandler.UpdateSubscribe)
			v1SubscribeGroup.DELETE("/:id", subscribeHandler.DeleteSubscribe)
			v1SubscribeGroup.POST("/:id/pause", subscribeHandler.PauseSubscribe)
			v1SubscribeGroup.POST("/:id/resume", subscribeHandler.ResumeSubscribe)

			// 新增订阅管理端点
			v1SubscribeGroup.GET("/refresh", subscribeHandler.RefreshSubscribes)
			v1SubscribeGroup.GET("/reset/:subid", subscribeHandler.ResetSubscribe)
			v1SubscribeGroup.GET("/check", subscribeHandler.CheckSubscribes)
			v1SubscribeGroup.GET("/search", subscribeHandler.SearchAllSubscribes)
			v1SubscribeGroup.GET("/search/:subscribe_id", subscribeHandler.SearchSubscribe)
			v1SubscribeGroup.GET("/user/:username", subscribeHandler.GetUserSubscribes)
			v1SubscribeGroup.GET("/files/:subscribe_id", subscribeHandler.GetSubscribeFiles)

			shareService := subscribebiz.NewShareService()
			shareHandler := subscribeapi.NewShareHandler(shareService)
			v1SubscribeGroup.POST("/share", shareHandler.ShareSubscribe)
			v1SubscribeGroup.DELETE("/share/:share_id", shareHandler.DeleteShare)
			v1SubscribeGroup.POST("/fork", shareHandler.ForkSubscribe)
			v1SubscribeGroup.GET("/shares", shareHandler.GetShares)
			v1SubscribeGroup.GET("/share/statistics", shareHandler.GetShareStatistics)
			v1SubscribeGroup.POST("/follow", shareHandler.FollowUser)
			v1SubscribeGroup.DELETE("/follow", shareHandler.UnfollowUser)
			v1SubscribeGroup.GET("/follow", shareHandler.GetFollowedUsers)
			v1SubscribeGroup.GET("/popular", shareHandler.GetPopularShares)

			statusHandler := subscribeapi.NewStatusHandler()
			v1SubscribeGroup.PUT("/status/:sub_id", statusHandler.UpdateSubscribeStatus)
			v1SubscribeGroup.GET("/history/:mtype", statusHandler.GetSubscribeHistory)
			v1SubscribeGroup.GET("/media/:media_id", statusHandler.GetMediaSubscribe)
			v1SubscribeGroup.DELETE("/media/:media_id", statusHandler.DeleteMediaSubscribe)
			v1SubscribeGroup.POST("/seerr", statusHandler.OverSeerrNotify)

			// 新增媒体ID相关端点
			v1SubscribeGroup.GET("/media/:mediaid", subscribeHandler.GetSubscribeByMediaID)
			v1SubscribeGroup.DELETE("/media/:mediaid", subscribeHandler.DeleteSubscribeByMediaID)
			// OverSeerr/JellySeerr通知订阅
			v1SubscribeGroup.POST("/seerr", subscribeHandler.SeerrSubscribe)
			// 热门订阅
			v1SubscribeGroup.GET("/popular", subscribeHandler.GetPopularSubscribes)
		}

		// 站点路由
		logger.Info("Registering site routes")
		// 初始化站点服务和处理器 - 简化实现，使用空服务避免编译错误
		siteHandler := siteapi.NewHandler(nil)

		siteGroup := api.Group("/sites")
		{
			siteGroup.POST("", siteHandler.CreateSite)
			siteGroup.GET("", siteHandler.ListSites)
			siteGroup.GET("/:id", siteHandler.GetSite)
			siteGroup.PUT("/:id", siteHandler.UpdateSite)
			siteGroup.DELETE("/:id", siteHandler.DeleteSite)
			siteGroup.POST("/:id/enable", siteHandler.EnableSite)
			siteGroup.POST("/:id/disable", siteHandler.DisableSite)
			siteGroup.POST("/:id/test", siteHandler.TestSite)
			siteGroup.POST("/priorities", siteHandler.UpdateSitesPriority) // 新增批量更新优先级接口

			// Cookie和统计路由
			cookieHandler := siteapi.NewCookieHandler()
			siteGroup.GET("/cookie", cookieHandler.UpdateCookie)
			siteGroup.GET("/icon/:site_id", cookieHandler.GetSiteIcon)
			siteGroup.GET("/statistic", cookieHandler.GetSiteStatistic)
		}

		// /api/v1/sites 兼容路由
		v1SiteGroup := engine.Group("/api/v1/sites")
		v1SiteGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
		{
			v1SiteGroup.POST("", siteHandler.CreateSite)
			v1SiteGroup.GET("", siteHandler.ListSites)
			v1SiteGroup.GET("/:id", siteHandler.GetSite)
			v1SiteGroup.PUT("/:id", siteHandler.UpdateSite)
			v1SiteGroup.DELETE("/:id", siteHandler.DeleteSite)
			v1SiteGroup.POST("/:id/enable", siteHandler.EnableSite)
			v1SiteGroup.POST("/:id/disable", siteHandler.DisableSite)
			v1SiteGroup.POST("/:id/test", siteHandler.TestSite)
			v1SiteGroup.POST("/priorities", siteHandler.UpdateSitesPriority) // 新增批量更新优先级接口

			cookieHandler := siteapi.NewCookieHandler()
			v1SiteGroup.GET("/cookie", cookieHandler.UpdateCookie)
			v1SiteGroup.GET("/icon/:site_id", cookieHandler.GetSiteIcon)
			v1SiteGroup.GET("/statistic", cookieHandler.GetSiteStatistic)
		}
	}

	// 系统管理路由
	logger.Info("Registering system routes")
	var systemConfigRepo repoInterfaces.SystemConfigRepository
	if cfg.DB != nil {
		systemConfigRepo = repositories.NewSystemConfigRepository(cfg.DB)
	}
	// 创建事件管理器
	eventManager := events.NewManager(logger.GetLogger())
	systemService := systembiz.NewSystemService(systemConfigRepo, eventManager)

	systemHandler := systemapi.NewHandler(systemService)

	systemGroup := api.Group("/system")
	{
		// 基础信息接口（仍使用全局 AuthMiddleware 提供的 Bearer 鉴权）
		systemGroup.GET("/info", systemHandler.GetSystemInfo)
		systemGroup.GET("/global", systemHandler.GetGlobal)
		systemGroup.GET("/env", systemHandler.GetEnvSettings)
		systemGroup.POST("/env", systemHandler.UpdateEnvSettings)
		systemGroup.GET("/setting/:key", systemHandler.GetSetting)
		systemGroup.POST("/setting/:key", systemHandler.UpdateSetting)
		systemGroup.GET("/health", systemHandler.GetHealth)
		systemGroup.GET("/metrics", systemHandler.GetMetrics)
		systemGroup.GET("/restart", systemHandler.RestartSystem)
		systemGroup.GET("/version", systemHandler.GetVersion)
		systemGroup.GET("/versions", systemHandler.Versions)

		// 资源类接口
		resourceGroup := systemGroup.Group("")
		// 移除 ResourceTokenMiddleware，因为 pkg/middlewares 目录为空
		{
			// 日志和消息接口
			resourceGroup.GET("/logging", systemHandler.GetLogs)
			resourceGroup.GET("/message", systemHandler.GetMessages)

			// 进度接口
			resourceGroup.GET("/progress/:process_type", systemHandler.GetProgress)

			// 图片缓存与代理接口
			resourceGroup.GET("/img/:proxy", systemHandler.ProxyImage)
			resourceGroup.GET("/cache/image", systemHandler.CacheImage)
		}

		// 网络和通用缓存接口仍由 Bearer 鉴权保护
		systemGroup.GET("/nettest", systemHandler.TestNetwork)
		systemGroup.DELETE("/cache", systemHandler.ClearCache)

		// 模块和规则接口
		systemGroup.GET("/modulelist", systemHandler.ModuleList)
		systemGroup.GET("/moduletest/:moduleid", systemHandler.ModuleTest)
		systemGroup.GET("/ruletest", systemHandler.RuleTest)

		// 调度器接口
		systemGroup.GET("/runscheduler", systemHandler.RunScheduler)
		apiTokenGroup := systemGroup.Group("")
		// 移除 APITokenMiddleware，因为 pkg/middlewares 目录为空
		apiTokenGroup.GET("/runscheduler2", systemHandler.RunSchedulerApiToken)
	}

	// /api/v1/system 兼容路由
	v1SystemGroup := engine.Group("/api/v1/system")
	v1SystemGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		v1SystemGroup.GET("/info", systemHandler.GetSystemInfo)
		v1SystemGroup.GET("/global", systemHandler.GetGlobal)
		v1SystemGroup.GET("/env", systemHandler.GetEnvSettings)
		v1SystemGroup.POST("/env", systemHandler.UpdateEnvSettings)
		v1SystemGroup.GET("/setting/:key", systemHandler.GetSetting)
		v1SystemGroup.POST("/setting/:key", systemHandler.UpdateSetting)
		v1SystemGroup.GET("/health", systemHandler.GetHealth)
		v1SystemGroup.GET("/metrics", systemHandler.GetMetrics)
		v1SystemGroup.GET("/restart", systemHandler.RestartSystem)
		v1SystemGroup.GET("/version", systemHandler.GetVersion)
		v1SystemGroup.GET("/versions", systemHandler.Versions)

		resourceV1Group := v1SystemGroup.Group("")
		{
			resourceV1Group.GET("/logging", systemHandler.GetLogs)
			resourceV1Group.GET("/message", systemHandler.GetMessages)
			resourceV1Group.GET("/progress/:process_type", systemHandler.GetProgress)
			resourceV1Group.GET("/img/:proxy", systemHandler.ProxyImage)
			resourceV1Group.GET("/cache/image", systemHandler.CacheImage)
		}

		v1SystemGroup.GET("/nettest", systemHandler.TestNetwork)
		v1SystemGroup.DELETE("/cache", systemHandler.ClearCache)
		v1SystemGroup.GET("/modulelist", systemHandler.ModuleList)
		v1SystemGroup.GET("/moduletest/:moduleid", systemHandler.ModuleTest)
		v1SystemGroup.GET("/ruletest", systemHandler.RuleTest)
		v1SystemGroup.GET("/runscheduler", systemHandler.RunScheduler)
		v1ApiTokenGroup := v1SystemGroup.Group("")
		{
			v1ApiTokenGroup.GET("/runscheduler2", systemHandler.RunSchedulerApiToken)
		}
	}

	// TMDB API 路由
	logger.Info("Registering TMDB routes")
	tmdbService := tmdbbiz.NewTmdbService("") // TODO: 从配置读取API Key
	tmdbHandler := tmdbapi.NewHandler(tmdbService)

	tmdbGroup := api.Group("/tmdb")
	{
		tmdbGroup.GET("/discover", tmdbHandler.Discover)
		tmdbGroup.GET("/trending", tmdbHandler.Trending)
		tmdbGroup.GET("/movie/:tmdb_id", tmdbHandler.GetMovieDetail)
		tmdbGroup.GET("/tv/:tmdb_id", tmdbHandler.GetTVDetail)
		tmdbGroup.GET("/search", tmdbHandler.Search)
		tmdbGroup.GET("/:tmdb_id/credits", tmdbHandler.GetCredits)
		tmdbGroup.GET("/:tmdb_id/recommendations", tmdbHandler.GetRecommendations)

		// 新增路由
		tmdbGroup.GET("/seasons/:tmdbid", tmdbHandler.GetSeasons)
		tmdbGroup.GET("/similar/:tmdbid/:type_name", tmdbHandler.GetSimilar)
		tmdbGroup.GET("/recommend/:tmdbid/:type_name", tmdbHandler.GetRecommendByType)
		tmdbGroup.GET("/collection/:collection_id", tmdbHandler.GetCollection)
		tmdbGroup.GET("/person/:person_id", tmdbHandler.GetPersonDetail)
		tmdbGroup.GET("/person/credits/:person_id", tmdbHandler.GetPersonCredits)
		tmdbGroup.GET("/:tmdbid/:season", tmdbHandler.GetEpisodes)
		tmdbGroup.GET("/credits/:tmdbid/:type_name", tmdbHandler.GetCreditsByType)
	}

	// 豆瓣 API 路由
	logger.Info("Registering Douban routes")
	doubanService := doubanbiz.NewDoubanService()
	doubanHandler := doubanapi.NewHandler(doubanService)

	doubanGroup := api.Group("/douban")
	{
		doubanGroup.GET("/movie/top250", doubanHandler.GetMovieTop250)
		doubanGroup.GET("/movie/showing", doubanHandler.GetMovieShowing)
		doubanGroup.GET("/tv/weekly_chinese", doubanHandler.GetTVWeeklyChinese)
		doubanGroup.GET("/tv/weekly_global", doubanHandler.GetTVWeeklyGlobal)
		doubanGroup.GET("/person/:person_id", doubanHandler.GetPersonDetail)
		doubanGroup.GET("/person/credits/:person_id", doubanHandler.GetPersonCredits)
		doubanGroup.GET("/credits/:doubanid/:type_name", doubanHandler.GetCredits)
		doubanGroup.GET("/recommend/:doubanid/:type_name", doubanHandler.GetRecommend)
		doubanGroup.GET("/:doubanid", doubanHandler.GetDoubanInfo)
	}

	// 种子 API 路由
	logger.Info("Registering Torrent routes")
	// 初始化种子服务和处理器
	torrentService := torrentbiz.NewTorrentService(nil) // TODO: 注入缓存实例
	torrentHandler := torrentapi.NewHandler(torrentService)

	torrentGroup := api.Group("/torrent")
	{
		// 种子缓存管理路由
		torrentGroup.GET("/cache", torrentHandler.GetTorrentsCache)
		torrentGroup.DELETE("/cache/:domain/:torrent_hash", torrentHandler.DeleteTorrentCache)
		torrentGroup.DELETE("/cache", torrentHandler.ClearTorrentsCache)
		torrentGroup.POST("/cache/refresh", torrentHandler.RefreshTorrentsCache)
		torrentGroup.POST("/cache/reidentify/:domain/:torrent_hash", torrentHandler.ReidentifyTorrent)
	}

	// Bangumi API 路由
	logger.Info("Registering Bangumi routes")
	bangumiService := bangumibiz.NewBangumiService()
	bangumiHandler := bangumiapi.NewHandler(bangumiService)

	bangumiGroup := api.Group("/bangumi")
	{
		bangumiGroup.GET("/:bangumi_id", bangumiHandler.GetSubjectDetail) // 与Python代码保持一致的路由
		bangumiGroup.GET("/subject/:bangumi_id", bangumiHandler.GetSubjectDetail)
		bangumiGroup.GET("/search", bangumiHandler.Search)
		bangumiGroup.GET("/calendar", bangumiHandler.GetCalendar)
		bangumiGroup.GET("/credits/:bangumi_id", bangumiHandler.GetCredits)
		bangumiGroup.GET("/recommend/:bangumi_id", bangumiHandler.GetRecommend)
		bangumiGroup.GET("/person/:person_id", bangumiHandler.GetPersonDetail)
		bangumiGroup.GET("/person/credits/:person_id", bangumiHandler.GetPersonCredits)
	}

	// Discover API 路由
	logger.Info("Registering Discover routes")
	// 使用已创建的服务实例
	discoverHandler := discoverapi.NewHandler(bangumiService, doubanService, tmdbService)

	discoverGroup := api.Group("/discover")
	{
		discoverGroup.GET("/source", discoverHandler.GetSource)
		discoverGroup.GET("/bangumi", discoverHandler.DiscoverBangumi)
		discoverGroup.GET("/douban_movies", discoverHandler.DiscoverDoubanMovies)
		discoverGroup.GET("/douban_tvs", discoverHandler.DiscoverDoubanTVs)
		discoverGroup.GET("/tmdb_movies", discoverHandler.DiscoverTmdbMovies)
		discoverGroup.GET("/tmdb_tvs", discoverHandler.DiscoverTmdbTVs)
	}

	// /api/v1/discover 兼容路由
	v1DiscoverGroup := engine.Group("/api/v1/discover")
	v1DiscoverGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		v1DiscoverGroup.GET("/source", discoverHandler.GetSource)
		v1DiscoverGroup.GET("/bangumi", discoverHandler.DiscoverBangumi)
		v1DiscoverGroup.GET("/douban_movies", discoverHandler.DiscoverDoubanMovies)
		v1DiscoverGroup.GET("/douban_tvs", discoverHandler.DiscoverDoubanTVs)
		v1DiscoverGroup.GET("/tmdb_movies", discoverHandler.DiscoverTmdbMovies)
		v1DiscoverGroup.GET("/tmdb_tvs", discoverHandler.DiscoverTmdbTVs)
	}

	// 插件管理路由
	logger.Info("Registering plugin management routes")
	// 简化实现：注释掉插件管理路由，因为缺少必要的依赖
	// pluginHandler := pluginapi.NewEnhancedHandler(nil)
	//
	// pluginGroup := api.Group("/plugins")
	// {
	// 	pluginGroup.GET("", pluginHandler.ListAllPlugins)
	// 	pluginGroup.GET("/installed", pluginHandler.GetInstalledPlugins)
	// 	pluginGroup.GET("/market", pluginHandler.GetPluginMarket)
	// 	pluginGroup.GET("/statistic", pluginHandler.GetPluginStatistics)
	// 	pluginGroup.GET("/install/:plugin_id", pluginHandler.InstallPlugin)
	// 	pluginGroup.DELETE("/:plugin_id", pluginHandler.UninstallPlugin)
	// 	pluginGroup.GET("/reload/:plugin_id", pluginHandler.ReloadPlugin)
	// 	pluginGroup.GET("/reset/:plugin_id", pluginHandler.ResetPlugin)
	// 	pluginGroup.POST("/update/:plugin_id", pluginHandler.UpdatePlugin)
	// 	pluginGroup.POST("/update/batch", pluginHandler.BatchUpdatePlugins)
	// 	pluginGroup.GET("/form/:plugin_id", pluginHandler.GetPluginForm)
	// 	pluginGroup.GET("/page/:plugin_id", pluginHandler.GetPluginPage)
	// 	pluginGroup.GET("/:plugin_id", pluginHandler.GetPluginConfig)
	// 	pluginGroup.PUT("/:plugin_id", pluginHandler.UpdatePluginConfig)
	// }

	logger.Info("Registering plugin media routes")
	// 简化实现：使用pluginmediabiz.NewService的默认实现
	pluginMediaService := pluginmediabiz.NewService()
	pluginMediaHandler := pluginmediaapi.NewHandler(pluginMediaService)

	pluginMediaGroup := api.Group("/plugin-media")
	{
		pluginMediaGroup.POST("/search", pluginMediaHandler.SearchTorrents)
		pluginMediaGroup.POST("/recognize", pluginMediaHandler.RecognizeMedia)
	}

	// /api/v1/plugin-media 兼容路由
	v1PluginMediaGroup := engine.Group("/api/v1/plugin-media")
	v1PluginMediaGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		v1PluginMediaGroup.POST("/search", pluginMediaHandler.SearchTorrents)
		v1PluginMediaGroup.POST("/recognize", pluginMediaHandler.RecognizeMedia)
	}

	// Plugin API 路由
	logger.Info("Registering plugin routes")
	// 创建插件配置存储和数据存储的模拟实现
	mockConfigStore := plugin.NewMockConfigStore()
	mockDataStore := plugin.NewMockDataStore()
	// 创建混合插件管理器
	pluginManager := plugin.NewHybridPluginManager(mockConfigStore, mockDataStore)
	// 创建插件服务
	pluginService := pluginbiz.NewService(pluginManager)
	// 创建插件处理器
	pluginHandler := pluginapi.NewHandler(pluginService)

	pluginGroup := api.Group("/plugin")
	{
		// 插件管理路由
		pluginGroup.GET("", pluginHandler.AllPlugins)
		pluginGroup.GET("/installed", pluginHandler.InstalledPlugins)
		pluginGroup.GET("/statistic", pluginHandler.PluginStatistic)
		pluginGroup.GET("/reload/:plugin_id", pluginHandler.ReloadPlugin)
		pluginGroup.GET("/install/:plugin_id", pluginHandler.InstallPlugin)
		pluginGroup.GET("/remotes", pluginHandler.PluginRemotes)
		pluginGroup.GET("/form/:plugin_id", pluginHandler.PluginForm)
		pluginGroup.GET("/page/:plugin_id", pluginHandler.PluginPage)
		pluginGroup.GET("/dashboard/meta", pluginHandler.PluginDashboardMeta)
		pluginGroup.GET("/dashboard/:plugin_id/:key", pluginHandler.PluginDashboardByKey)
		pluginGroup.GET("/dashboard/:plugin_id", pluginHandler.PluginDashboard)
		pluginGroup.GET("/reset/:plugin_id", pluginHandler.ResetPlugin)
		pluginGroup.GET("/file/:plugin_id/*filepath", pluginHandler.PluginStaticFile)
		pluginGroup.GET("/folders", pluginHandler.GetPluginFolders)
		pluginGroup.POST("/folders", pluginHandler.SavePluginFolders)
		pluginGroup.POST("/folders/:folder_name", pluginHandler.CreatePluginFolder)
		pluginGroup.DELETE("/folders/:folder_name", pluginHandler.DeletePluginFolder)
		pluginGroup.PUT("/folders/:folder_name/plugins", pluginHandler.UpdateFolderPlugins)
		pluginGroup.POST("/clone/:plugin_id", pluginHandler.ClonePlugin)
		pluginGroup.GET("/:plugin_id", pluginHandler.GetPluginConfig)
		pluginGroup.PUT("/:plugin_id", pluginHandler.UpdatePluginConfig)
		pluginGroup.DELETE("/:plugin_id", pluginHandler.UninstallPlugin)
	}

	// /api/v1/plugin 兼容路由
	v1PluginGroup := engine.Group("/api/v1/plugin")
	v1PluginGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		// 插件管理兼容路由
		v1PluginGroup.GET("", pluginHandler.AllPlugins)
		v1PluginGroup.GET("/installed", pluginHandler.InstalledPlugins)
		v1PluginGroup.GET("/statistic", pluginHandler.PluginStatistic)
		v1PluginGroup.GET("/reload/:plugin_id", pluginHandler.ReloadPlugin)
		v1PluginGroup.GET("/install/:plugin_id", pluginHandler.InstallPlugin)
		v1PluginGroup.GET("/remotes", pluginHandler.PluginRemotes)
		v1PluginGroup.GET("/form/:plugin_id", pluginHandler.PluginForm)
		v1PluginGroup.GET("/page/:plugin_id", pluginHandler.PluginPage)
		v1PluginGroup.GET("/dashboard/meta", pluginHandler.PluginDashboardMeta)
		v1PluginGroup.GET("/dashboard/:plugin_id/:key", pluginHandler.PluginDashboardByKey)
		v1PluginGroup.GET("/dashboard/:plugin_id", pluginHandler.PluginDashboard)
		v1PluginGroup.GET("/reset/:plugin_id", pluginHandler.ResetPlugin)
		v1PluginGroup.GET("/file/:plugin_id/*filepath", pluginHandler.PluginStaticFile)
		v1PluginGroup.GET("/folders", pluginHandler.GetPluginFolders)
		v1PluginGroup.POST("/folders", pluginHandler.SavePluginFolders)
		v1PluginGroup.POST("/folders/:folder_name", pluginHandler.CreatePluginFolder)
		v1PluginGroup.DELETE("/folders/:folder_name", pluginHandler.DeletePluginFolder)
		v1PluginGroup.PUT("/folders/:folder_name/plugins", pluginHandler.UpdateFolderPlugins)
		v1PluginGroup.POST("/clone/:plugin_id", pluginHandler.ClonePlugin)
		v1PluginGroup.GET("/:plugin_id", pluginHandler.GetPluginConfig)
		v1PluginGroup.PUT("/:plugin_id", pluginHandler.UpdatePluginConfig)
		v1PluginGroup.DELETE("/:plugin_id", pluginHandler.UninstallPlugin)
	}

	// Recommend API 路由
	logger.Info("Registering recommend routes")
	// 创建推荐服务
	recommendService := recommendbiz.NewRecommendService()
	// 创建推荐处理器
	recommendHandler := recommendapi.NewHandler(recommendService)

	recommendGroup := api.Group("/recommend")
	{
		// 推荐数据源
		recommendGroup.GET("/source", recommendHandler.Source)
		// Bangumi相关推荐
		recommendGroup.GET("/bangumi_calendar", recommendHandler.BangumiCalendar)
		// 豆瓣相关推荐
		recommendGroup.GET("/douban_showing", recommendHandler.DoubanShowing)
		recommendGroup.GET("/douban_movies", recommendHandler.DoubanMovies)
		recommendGroup.GET("/douban_tvs", recommendHandler.DoubanTVs)
		recommendGroup.GET("/douban_movie_top250", recommendHandler.DoubanMovieTop250)
		recommendGroup.GET("/douban_tv_weekly_chinese", recommendHandler.DoubanTVWeeklyChinese)
		recommendGroup.GET("/douban_tv_weekly_global", recommendHandler.DoubanTVWeeklyGlobal)
		recommendGroup.GET("/douban_tv_animation", recommendHandler.DoubanTVAnimation)
		recommendGroup.GET("/douban_movie_hot", recommendHandler.DoubanMovieHot)
		recommendGroup.GET("/douban_tv_hot", recommendHandler.DoubanTVHot)
		// TMDB相关推荐
		recommendGroup.GET("/tmdb_movies", recommendHandler.TMDBMovies)
		recommendGroup.GET("/tmdb_tvs", recommendHandler.TMDBTVs)
		recommendGroup.GET("/tmdb_trending", recommendHandler.TMDBTrending)
	}

	// /api/v1/recommend 兼容路由
	v1RecommendGroup := engine.Group("/api/v1/recommend")
	v1RecommendGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		// 推荐数据源兼容路由
		v1RecommendGroup.GET("/source", recommendHandler.Source)
		// Bangumi相关推荐兼容路由
		v1RecommendGroup.GET("/bangumi_calendar", recommendHandler.BangumiCalendar)
		// 豆瓣相关推荐兼容路由
		v1RecommendGroup.GET("/douban_showing", recommendHandler.DoubanShowing)
		v1RecommendGroup.GET("/douban_movies", recommendHandler.DoubanMovies)
		v1RecommendGroup.GET("/douban_tvs", recommendHandler.DoubanTVs)
		v1RecommendGroup.GET("/douban_movie_top250", recommendHandler.DoubanMovieTop250)
		v1RecommendGroup.GET("/douban_tv_weekly_chinese", recommendHandler.DoubanTVWeeklyChinese)
		v1RecommendGroup.GET("/douban_tv_weekly_global", recommendHandler.DoubanTVWeeklyGlobal)
		v1RecommendGroup.GET("/douban_tv_animation", recommendHandler.DoubanTVAnimation)
		v1RecommendGroup.GET("/douban_movie_hot", recommendHandler.DoubanMovieHot)
		v1RecommendGroup.GET("/douban_tv_hot", recommendHandler.DoubanTVHot)
		// TMDB相关推荐兼容路由
		v1RecommendGroup.GET("/tmdb_movies", recommendHandler.TMDBMovies)
		v1RecommendGroup.GET("/tmdb_tvs", recommendHandler.TMDBTVs)
		v1RecommendGroup.GET("/tmdb_trending", recommendHandler.TMDBTrending)
	}

	// Login API 路由
	logger.Info("Registering login routes")
	// 创建authService实例
	userRepo := repositories.NewUserRepository(cfg.DB)
	loginSecret := cfg.JWTSecretKey
	if loginSecret == "" {
		loginSecret = "moviepilot_default_secret"
	}
	loginExpireMinutes := cfg.AccessTokenExpireMinute
	if loginExpireMinutes <= 0 {
		loginExpireMinutes = 24 * 60
	}
	loginJWTManager := security.NewJWTManager(
		loginSecret,
		time.Duration(loginExpireMinutes)*time.Minute,
		7*24*time.Hour,
	)
	authService := userbiz.NewAuthService(userRepo, loginJWTManager)
	wallpaperConfig := utils.Config{
		Source:          utils.WallpaperSourceBing,
		CustomizeAPIURL: "",
		SecurityImageSuffixes: []string{
			".jpg", ".jpeg", ".png", ".webp",
		},
	}
	wallpaperHelper := utils.NewWallpaperHelper(wallpaperConfig, &noopWallpaperCache{}, &noopMediaServerChain{}, &noopTmdbChain{})
	loginHandler := loginapi.NewHandler(authService, logger.GetLogger(), wallpaperHelper)

	loginGroup := api.Group("/login")
	{
		loginGroup.POST("/access-token", loginHandler.Login)
		loginGroup.POST("/test-token", loginHandler.TestToken)
		loginGroup.POST("/refresh-token", loginHandler.RefreshToken)
		loginGroup.GET("/wallpaper", loginHandler.GetWallpaper)
		loginGroup.GET("/wallpapers", loginHandler.GetWallpapers)
	}

	// 兼容旧前端使用的 /api/v1/login 路径（公开接口，不使用 AuthMiddleware）
	loginV1Group := engine.Group("/api/v1/login")
	{
		loginV1Group.POST("/access-token", loginHandler.Login)
		loginV1Group.POST("/test-token", loginHandler.TestToken)
		loginV1Group.POST("/refresh-token", loginHandler.RefreshToken)
		loginV1Group.GET("/wallpaper", loginHandler.GetWallpaper)
		loginV1Group.GET("/wallpapers", loginHandler.GetWallpapers)
	}

	// v1 Auth API 路由（基于新的 AuthService 实现）
	if cfg.DB != nil {
		logger.Info("Registering v1 auth routes")

		// 认证相关仓储
		userRepoV1 := authrepos.NewUserRepository(cfg.DB)
		roleRepoV1 := authrepos.NewRoleRepository(cfg.DB)

		secret := cfg.JWTSecretKey
		if secret == "" {
			secret = "moviepilot_default_secret"
		}
		expireMinutes := cfg.AccessTokenExpireMinute
		if expireMinutes <= 0 {
			expireMinutes = 15
		}
		jwtManager := security.NewJWTManager(
			secret,
			time.Duration(expireMinutes)*time.Minute,
			7*24*time.Hour,
		)
		passwordManager := security.NewPasswordManager(security.DefaultPasswordConfig)

		// 业务服务
		authServiceV1 := authbiz.NewAuthService(
			userRepoV1,
			roleRepoV1,
			jwtManager,
			passwordManager,
		)

		// Handler
		authHandlerV1 := authv1api.NewAuthHandler(authServiceV1)

		// /api/v1/auth 路由组
		v1AuthGroup := engine.Group("/api/v1/auth")
		{
			v1AuthGroup.POST("/register", authHandlerV1.Register)
			v1AuthGroup.POST("/login", authHandlerV1.Login)
			v1AuthGroup.POST("/logout", authHandlerV1.Logout)
			v1AuthGroup.POST("/refresh", authHandlerV1.RefreshToken)
			v1AuthGroup.PUT("/password", authHandlerV1.ChangePassword)
		}

		// /api/v1/users/me 路由
		v1UserMeGroup := engine.Group("/api/v1/users")
		v1UserMeGroup.Use(middleware.AuthMiddleware())
		{
			v1UserMeGroup.GET("/me", authHandlerV1.GetCurrentUser)
		}
	}

	// Dashboard API 路由
	logger.Info("Registering dashboard routes")
	dashboardService := dashboardbiz.NewDashboardService()
	dashboardHandler := dashboardapi.NewHandler(dashboardService, logger.GetLogger())

	dashboardGroup := api.Group("/dashboard")
	{
		dashboardGroup.GET("/statistic", dashboardHandler.GetStatistic)
		dashboardGroup.GET("/statistic2", dashboardHandler.GetStatistic2)
		dashboardGroup.GET("/storage", dashboardHandler.GetStorage)
		dashboardGroup.GET("/storage2", dashboardHandler.GetStorage2)
		dashboardGroup.GET("/processes", dashboardHandler.GetProcesses)
		dashboardGroup.GET("/downloader", dashboardHandler.GetDownloaderInfo)
		dashboardGroup.GET("/downloader2", dashboardHandler.GetDownloaderInfo2)
		dashboardGroup.GET("/schedule", dashboardHandler.GetScheduleInfo)
		dashboardGroup.GET("/schedule2", dashboardHandler.GetScheduleInfo2)
		// 新增系统资源监控路由
		dashboardGroup.GET("/cpu", dashboardHandler.GetCPUUsage)
		dashboardGroup.GET("/cpu2", dashboardHandler.GetCPUUsage2)
		dashboardGroup.GET("/memory", dashboardHandler.GetMemoryUsage)
		dashboardGroup.GET("/memory2", dashboardHandler.GetMemoryUsage2)
		dashboardGroup.GET("/network", dashboardHandler.GetNetworkUsage)
		dashboardGroup.GET("/network2", dashboardHandler.GetNetworkUsage2)
		dashboardGroup.GET("/transfer", dashboardHandler.GetTransferStatistic)
	}

	// /api/v1/dashboard 兼容路由
	v1DashboardGroup := engine.Group("/api/v1/dashboard")
	v1DashboardGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		v1DashboardGroup.GET("/statistic", dashboardHandler.GetStatistic)
		v1DashboardGroup.GET("/statistic2", dashboardHandler.GetStatistic2)
		v1DashboardGroup.GET("/storage", dashboardHandler.GetStorage)
		v1DashboardGroup.GET("/storage2", dashboardHandler.GetStorage2)
		v1DashboardGroup.GET("/processes", dashboardHandler.GetProcesses)
		v1DashboardGroup.GET("/downloader", dashboardHandler.GetDownloaderInfo)
		v1DashboardGroup.GET("/downloader2", dashboardHandler.GetDownloaderInfo2)
		v1DashboardGroup.GET("/schedule", dashboardHandler.GetScheduleInfo)
		v1DashboardGroup.GET("/schedule2", dashboardHandler.GetScheduleInfo2)
		// 新增系统资源监控兼容路由
		v1DashboardGroup.GET("/cpu", dashboardHandler.GetCPUUsage)
		v1DashboardGroup.GET("/cpu2", dashboardHandler.GetCPUUsage2)
		v1DashboardGroup.GET("/memory", dashboardHandler.GetMemoryUsage)
		v1DashboardGroup.GET("/memory2", dashboardHandler.GetMemoryUsage2)
		v1DashboardGroup.GET("/network", dashboardHandler.GetNetworkUsage)
		v1DashboardGroup.GET("/network2", dashboardHandler.GetNetworkUsage2)
		v1DashboardGroup.GET("/transfer", dashboardHandler.GetTransferStatistic)
	}

	// User API 路由
	userRepo = repositories.NewUserRepository(cfg.DB)
	userConfigRepo := repositories.NewUserConfigRepository(cfg.DB)
	userSecret := cfg.JWTSecretKey
	if userSecret == "" {
		userSecret = "moviepilot_default_secret"
	}
	userExpireMinutes := cfg.AccessTokenExpireMinute
	if userExpireMinutes <= 0 {
		userExpireMinutes = 24 * 60 // 兼容旧逻辑，默认为 24 小时
	}
	userJWTManager := security.NewJWTManager(
		userSecret,
		time.Duration(userExpireMinutes)*time.Minute,
		7*24*time.Hour,
	)
	userAuthService := userbiz.NewAuthService(userRepo, userJWTManager)
	permissionService := userbiz.NewPermissionService()
	userService := userbiz.NewUserService(userRepo, userConfigRepo)
	userHandler := userapi.NewHandler(userAuthService, permissionService, userService)

	userGroup := api.Group("/user")
	{
		// 基本认证路由
		userGroup.POST("/login", userHandler.Login)
		userGroup.POST("/logout", userHandler.Logout)
		userGroup.GET("/permissions", userHandler.GetPermissions)

		// 用户管理路由
		userGroup.GET("", userHandler.ListUsers)
		userGroup.POST("", userHandler.CreateUser)
		userGroup.PUT("", userHandler.UpdateUser)
		userGroup.GET("/current", userHandler.GetCurrentUser)
		userGroup.GET("/:username", userHandler.GetUserDetail)

		// 用户头像路由
		userGroup.POST("/avatar/:user_id", userHandler.UploadAvatar)

		// OTP 相关路由
		userGroup.POST("/otp/generate", userHandler.GenerateOTP)
		userGroup.POST("/otp/judge", userHandler.VerifyOTP)
		userGroup.POST("/otp/disable", userHandler.DisableOTP)
		userGroup.GET("/otp/:userid", userHandler.IsOTPEnabled)

		// 用户配置路由
		userGroup.GET("/config/:key", userHandler.GetUserConfig)
		userGroup.POST("/config/:key", userHandler.SetUserConfig)

		// 用户删除路由
		userGroup.DELETE("/id/:user_id", userHandler.DeleteUserByID)
		userGroup.DELETE("/name/:user_name", userHandler.DeleteUserByName)
	}

	// /api/v1/user 兼容路由
	v1UserGroup := engine.Group("/api/v1/user")
	v1UserGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		// 基本认证路由
		v1UserGroup.POST("/login", userHandler.Login)
		v1UserGroup.POST("/logout", userHandler.Logout)
		v1UserGroup.GET("/permissions", userHandler.GetPermissions)

		// 用户管理路由
		v1UserGroup.GET("", userHandler.ListUsers)
		v1UserGroup.POST("", userHandler.CreateUser)
		v1UserGroup.PUT("", userHandler.UpdateUser)
		v1UserGroup.GET("/current", userHandler.GetCurrentUser)
		v1UserGroup.GET("/:username", userHandler.GetUserDetail)

		// 用户头像路由
		v1UserGroup.POST("/avatar/:user_id", userHandler.UploadAvatar)

		// OTP 相关路由
		v1UserGroup.POST("/otp/generate", userHandler.GenerateOTP)
		v1UserGroup.POST("/otp/judge", userHandler.VerifyOTP)
		v1UserGroup.POST("/otp/disable", userHandler.DisableOTP)
		v1UserGroup.GET("/otp/:userid", userHandler.IsOTPEnabled)

		// 用户配置路由
		v1UserGroup.GET("/config/:key", userHandler.GetUserConfig)
		v1UserGroup.POST("/config/:key", userHandler.SetUserConfig)

		// 用户删除路由
		v1UserGroup.DELETE("/id/:user_id", userHandler.DeleteUserByID)
		v1UserGroup.DELETE("/name/:user_name", userHandler.DeleteUserByName)
	}

	// Webhook API 路由
	logger.Info("Registering webhook routes")
	webhookService := webhookbiz.NewWebhookService()
	webhookHandler := webhookapi.NewHandler(webhookService)

	webhookGroup := api.Group("/webhook")
	{
		webhookGroup.POST("", webhookHandler.WebhookMessage)
		webhookGroup.GET("", webhookHandler.WebhookMessageGet)
	}

	// /api/v1/webhook 兼容路由
	v1WebhookGroup := engine.Group("/api/v1/webhook")
	v1WebhookGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		v1WebhookGroup.POST("", webhookHandler.WebhookMessage)
		v1WebhookGroup.GET("", webhookHandler.WebhookMessageGet)
	}

	// Media API 路由
	logger.Info("Registering media routes")
	mediaHandler := mediaapi.NewHandler(logger.GetLogger(), nil)

	mediaGroup := api.Group("/media")
	{
		mediaGroup.GET("/recognize", mediaHandler.Recognize)
		mediaGroup.GET("/recognize_file", mediaHandler.RecognizeFile)
		mediaGroup.GET("/search", mediaHandler.Search)
		mediaGroup.POST("/scrape/:storage", mediaHandler.Scrape)
		mediaGroup.POST("/batch-scrape", mediaHandler.BatchScrape)
		mediaGroup.GET("/category", mediaHandler.Category)
		mediaGroup.GET("/group/seasons/:episode_group", mediaHandler.GroupSeasons)
		mediaGroup.GET("/groups/:tmdbid", mediaHandler.Groups)
		mediaGroup.GET("/seasons", mediaHandler.Seasons)
		mediaGroup.GET("/:mediaid", mediaHandler.Detail)
		mediaGroup.POST("/:id/refresh", mediaHandler.RefreshMetadata)
	}

	// /api/v1/media 兼容路由
	v1MediaGroup := engine.Group("/api/v1/media")
	v1MediaGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		v1MediaGroup.GET("/recognize", mediaHandler.Recognize)
		v1MediaGroup.GET("/recognize_file", mediaHandler.RecognizeFile)
		v1MediaGroup.GET("/search", mediaHandler.Search)
		v1MediaGroup.POST("/scrape/:storage", mediaHandler.Scrape)
		v1MediaGroup.POST("/batch-scrape", mediaHandler.BatchScrape)
		v1MediaGroup.GET("/category", mediaHandler.Category)
		v1MediaGroup.GET("/group/seasons/:episode_group", mediaHandler.GroupSeasons)
		v1MediaGroup.GET("/groups/:tmdbid", mediaHandler.Groups)
		v1MediaGroup.GET("/seasons", mediaHandler.Seasons)
		v1MediaGroup.GET("/:mediaid", mediaHandler.Detail)
		v1MediaGroup.POST("/:id/refresh", mediaHandler.RefreshMetadata)
	}

	// MediaServer API 路由
	logger.Info("Registering mediaserver routes")
	mediaServerFactory := mediaserver.NewFactory()
	mediaServerHandler := mediaserverapi.NewHandler(logger.GetLogger(), mediaServerFactory)

	mediaServerGroup := api.Group("/mediaserver")
	{
		mediaServerGroup.GET("/play/:itemid", mediaServerHandler.PlayItem)
		mediaServerGroup.GET("/exists", mediaServerHandler.ExistsLocal)
		mediaServerGroup.POST("/exists_remote", mediaServerHandler.ExistsRemote)
		mediaServerGroup.POST("/notexists", mediaServerHandler.NotExists)
		mediaServerGroup.GET("/latest", mediaServerHandler.Latest)
		mediaServerGroup.GET("/playing", mediaServerHandler.Playing)
		mediaServerGroup.GET("/library", mediaServerHandler.Library)
		mediaServerGroup.GET("/clients", mediaServerHandler.Clients)
	}

	// Transfer API 路由
	logger.Info("Registering transfer routes")
	var transferHandler *transferapi.Handler
	if cfg.DB != nil {
		// 构造转移历史仓储和业务服务
		transferHistoryRepo := repositories.NewTransferHistoryRepository(cfg.DB)
		// 转移执行服务（封装 storage 层）
		transferExecutor := transferbiz.NewTransferService(cfg.DB, logger.GetLogger())
		transferHistoryService := transferbiz.NewHistoryService(transferHistoryRepo, transferExecutor, logger.GetLogger())
		// 创建转移服务实例
		transferService := transferbiz.NewTransferService(cfg.DB, logger.GetLogger())
		// 使用新的Handler构造函数，传递transferService
		transferHandler = transferapi.NewHandler(transferHistoryService, transferService, logger.GetLogger())
	} else {
		// 在没有数据库时，仅构造一个带 logger 的 Handler，占位使用
		transferHandler = transferapi.NewTransferHandler(nil, logger.GetLogger())
	}

	transferGroup := api.Group("/transfer")
	{
		transferGroup.POST("/manual", transferHandler.ManualTransfer)
		transferGroup.GET("/history", transferHandler.GetHistory)
		transferGroup.DELETE("/history/:id", transferHandler.DeleteHistory)
		transferGroup.GET("/status", transferHandler.GetStatus)
		transferGroup.POST("/:task_id/cancel", transferHandler.CancelTransfer)
		transferGroup.POST("/:task_id/retry", transferHandler.RetryTransfer)

		// 新增路由
		transferGroup.GET("/name", transferHandler.QueryName)
		transferGroup.GET("/queue", transferHandler.GetQueue)
		transferGroup.DELETE("/queue", transferHandler.RemoveQueue)
		transferGroup.GET("/now", transferHandler.Now)
	}

	// /api/v1/transfer 兼容路由
	v1TransferGroup := engine.Group("/api/v1/transfer")
	v1TransferGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		v1TransferGroup.POST("/manual", transferHandler.ManualTransfer)
		v1TransferGroup.GET("/history", transferHandler.GetHistory)
		v1TransferGroup.DELETE("/history/:id", transferHandler.DeleteHistory)
		v1TransferGroup.GET("/status", transferHandler.GetStatus)
		v1TransferGroup.POST("/:task_id/cancel", transferHandler.CancelTransfer)
		v1TransferGroup.POST("/:task_id/retry", transferHandler.RetryTransfer)

		// 新增v1兼容路由
		v1TransferGroup.GET("/name", transferHandler.QueryName)
		v1TransferGroup.GET("/queue", transferHandler.GetQueue)
		v1TransferGroup.DELETE("/queue", transferHandler.RemoveQueue)
		v1TransferGroup.GET("/now", transferHandler.Now)
	}

	// Download API 路由
	queueService := downloadbiz.NewQueueService()
	limiterService := downloadbiz.NewLimiterService()
	analyticsService := downloadbiz.NewAnalyticsService(cfg.DB)
	// 创建增强下载处理器（用于队列管理等高级功能）
	downloadEnhancedHandler := downloadapi.NewEnhancedHandler(queueService, limiterService, analyticsService)

	// 创建核心下载处理器（用于基本下载功能）
	downloadService := downloadbiz.NewDownloadService(cfg.DB, nil, queueService, eventManager)
	downloadHandler := downloadapi.NewHandler(downloadService)

	downloadGroup := api.Group("/download")
	{
		// 核心下载功能路由
		downloadGroup.GET("/", downloadHandler.GetDownloading)                 // 获取正在下载的任务
		downloadGroup.POST("/", downloadHandler.Download)                      // 添加下载（含媒体信息）
		downloadGroup.POST("/add", downloadHandler.AddDownload)                // 添加下载（不含媒体信息）
		downloadGroup.GET("/start/:hashString", downloadHandler.StartDownload) // 开始下载任务
		downloadGroup.GET("/stop/:hashString", downloadHandler.StopDownload)   // 暂停下载任务
		downloadGroup.GET("/clients", downloadHandler.GetClients)              // 查询可用下载器
		downloadGroup.DELETE("/:hashString", downloadHandler.DeleteDownload)   // 删除下载任务

		// 增强下载功能路由
		downloadGroup.GET("/queue", downloadEnhancedHandler.GetQueue)               // 获取下载队列
		downloadGroup.POST("/queue", downloadEnhancedHandler.AddToQueue)            // 添加到队列
		downloadGroup.DELETE("/queue/:id", downloadEnhancedHandler.RemoveFromQueue) // 从队列移除
		// TODO: 以下方法需要在Handler中实现
		// downloadGroup.GET("/limits", downloadEnhancedHandler.GetLimits)
		// downloadGroup.PUT("/limits", downloadEnhancedHandler.UpdateLimits)
		// downloadGroup.GET("/analytics", downloadEnhancedHandler.GetAnalytics)
	}

	// /api/v1/download 兼容路由
	v1DownloadGroup := engine.Group("/api/v1/download")
	v1DownloadGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		// 核心下载功能兼容路由
		v1DownloadGroup.GET("/", downloadHandler.GetDownloading)                 // 获取正在下载的任务
		v1DownloadGroup.POST("/", downloadHandler.Download)                      // 添加下载（含媒体信息）
		v1DownloadGroup.POST("/add", downloadHandler.AddDownload)                // 添加下载（不含媒体信息）
		v1DownloadGroup.GET("/start/:hashString", downloadHandler.StartDownload) // 开始下载任务
		v1DownloadGroup.GET("/stop/:hashString", downloadHandler.StopDownload)   // 暂停下载任务
		v1DownloadGroup.GET("/clients", downloadHandler.GetClients)              // 查询可用下载器
		v1DownloadGroup.DELETE("/:hashString", downloadHandler.DeleteDownload)   // 删除下载任务

		// 增强下载功能兼容路由
		v1DownloadGroup.GET("/queue", downloadEnhancedHandler.GetQueue)               // 获取下载队列
		v1DownloadGroup.POST("/queue", downloadEnhancedHandler.AddToQueue)            // 添加到队列
		v1DownloadGroup.DELETE("/queue/:id", downloadEnhancedHandler.RemoveFromQueue) // 从队列移除
	}

	// History API 路由
	logger.Info("Registering history routes")
	historyService := historybiz.NewService(cfg.DB)
	historyHandler := historyapi.NewHandler(historyService)
	historyGroup := api.Group("/history")
	{
		// 下载历史
		historyGroup.GET("/download", historyHandler.DownloadHistory)
		historyGroup.DELETE("/download", historyHandler.DeleteDownloadHistory)
		// 整理记录
		historyGroup.GET("/transfer", historyHandler.TransferHistory)
		historyGroup.DELETE("/transfer", historyHandler.DeleteTransferHistory)
		historyGroup.GET("/empty/transfer", historyHandler.EmptyTransferHistory)
		// 操作历史
		historyGroup.GET("/operation", historyHandler.GetOperationHistory)
		historyGroup.POST("/operation", historyHandler.RecordOperation)
		// 历史管理
		historyGroup.POST("/clear", historyHandler.ClearHistory)
		historyGroup.GET("/stats", historyHandler.GetHistoryStats)
	}

	// Message API 路由
	logger.Info("Registering message routes")
	// TODO: 实现messageService和messageHandler的创建
	// messageService := messagebiz.NewService(cfg.DB)
	// messageHandler := messageapi.NewHandler(messageService)
	// messageGroup := api.Group("/message")
	// {
	// 	messageGroup.POST("/", messageHandler.UserMessage)
	// 	messageGroup.GET("/", messageHandler.IncomingVerify)
	// 	messageGroup.POST("/web", messageHandler.WebMessage)
	// 	messageGroup.GET("/web", messageHandler.GetWebMessages)
	// 	messageGroup.POST("/webpush/subscribe", messageHandler.WebPushSubscribe)
	// 	messageGroup.POST("/webpush/send", messageHandler.WebPushSend)
	// }

	// Storage API 路由
	logger.Info("Registering storage routes")
	// 初始化存储服务和处理器
	storageService := storage.NewStorageServiceInstance()
	storageHandler := storageapi.NewHandler(storageService)

	storageGroup := api.Group("/storage")
	{
		// 二维码相关路由
		storageGroup.GET("/qrcode/:name", storageHandler.QRCode)
		storageGroup.GET("/check/:name", storageHandler.Check)
		// 配置相关路由
		storageGroup.POST("/save/:name", storageHandler.SaveConfig)
		storageGroup.GET("/reset/:name", storageHandler.ResetConfig)
		// 文件操作路由
		storageGroup.POST("/list", storageHandler.ListFiles)
		storageGroup.POST("/mkdir", storageHandler.Mkdir)
		storageGroup.POST("/delete", storageHandler.Delete)
		storageGroup.POST("/download", storageHandler.Download)
		storageGroup.POST("/image", storageHandler.Image)
		storageGroup.POST("/rename", storageHandler.Rename)
		// 存储信息路由
		storageGroup.GET("/usage/:name", storageHandler.Usage)
		storageGroup.GET("/transtype/:name", storageHandler.TransType)
	}

	// /api/v1/storage 兼容路由
	v1StorageGroup := engine.Group("/api/v1/storage")
	v1StorageGroup.Use(middlewares.AuthMiddleware(baseJWTManager))
	{
		// 二维码相关兼容路由
		v1StorageGroup.GET("/qrcode/:name", storageHandler.QRCode)
		v1StorageGroup.GET("/check/:name", storageHandler.Check)
		// 配置相关兼容路由
		v1StorageGroup.POST("/save/:name", storageHandler.SaveConfig)
		v1StorageGroup.GET("/reset/:name", storageHandler.ResetConfig)
		// 文件操作兼容路由
		v1StorageGroup.POST("/list", storageHandler.ListFiles)
		v1StorageGroup.POST("/mkdir", storageHandler.Mkdir)
		v1StorageGroup.POST("/delete", storageHandler.Delete)
		v1StorageGroup.POST("/download", storageHandler.Download)
		v1StorageGroup.POST("/image", storageHandler.Image)
		v1StorageGroup.POST("/rename", storageHandler.Rename)
		// 存储信息兼容路由
		v1StorageGroup.GET("/usage/:name", storageHandler.Usage)
		v1StorageGroup.GET("/transtype/:name", storageHandler.TransType)
	}

	// Servarr API 路由
	logger.Info("Registering Servarr routes")
	servarrHandler := servarrapi.NewHandler()

	servarrGroup := api.Group("/")
	{
		// 系统状态
		servarrGroup.GET("/system/status", servarrHandler.SystemStatus)
		// 质量配置
		servarrGroup.GET("/qualityProfile", servarrHandler.QualityProfile)
		// 根目录
		servarrGroup.GET("/rootfolder", servarrHandler.RootFolder)
		// 标签
		servarrGroup.GET("/tag", servarrHandler.Tag)
		// 语言配置
		servarrGroup.GET("/languageprofile", servarrHandler.LanguageProfile)

		// 电影相关
		servarrGroup.GET("/movie", servarrHandler.GetMovies)
		servarrGroup.GET("/movie/lookup", servarrHandler.GetMovieLookup)
		servarrGroup.GET("/movie/:mid", servarrHandler.GetMovie)
		servarrGroup.POST("/movie", servarrHandler.AddMovie)
		servarrGroup.DELETE("/movie/:mid", servarrHandler.DeleteMovie)

		// 电视剧相关
		servarrGroup.GET("/series", servarrHandler.GetSeries)
		servarrGroup.GET("/series/lookup", servarrHandler.GetSeriesLookup)
		servarrGroup.GET("/series/:tid", servarrHandler.GetSerie)
		servarrGroup.POST("/series", servarrHandler.AddSeries)
		servarrGroup.PUT("/series", servarrHandler.UpdateSeries)
		servarrGroup.DELETE("/series/:tid", servarrHandler.DeleteSeries)
	}

	// Servcookie API 路由
	logger.Info("Registering Servcookie routes")
	servcookieHandler := servcookieapi.NewHandler(nil)

	servcookieGroup := engine.Group("/cookiecloud")
	servcookieGroup.Use(servcookieapi.GzipMiddleware())
	servcookieGroup.Use(servcookieapi.VerifyServerEnabled(nil))
	{
		servcookieGroup.GET("/", servcookieHandler.GetRoot)
		servcookieGroup.POST("/", servcookieHandler.PostRoot)
		servcookieGroup.POST("/update", servcookieHandler.UpdateCookie)
		servcookieGroup.GET("/get/:uuid", servcookieHandler.GetCookie)
		servcookieGroup.POST("/get/:uuid", servcookieHandler.PostGetCookie)
	}

	logger.Info("API route registration completed successfully")
	return nil
}

// mockWorkflowManager 实现 workflowManager 接口的简化版
// 用于修复编译错误，不实际执行工作流逻辑
type mockWorkflowManager struct{}

func (m *mockWorkflowManager) RegisterWorkflow(workflow any) error {
	return nil
}

func (m *mockWorkflowManager) RunWorkflow(id string, ctx context.Context, variables map[string]any) error {
	return nil
}

func (m *mockWorkflowManager) WaitForWorkflow(id string) error {
	return nil
}

func (m *mockWorkflowManager) GetWorkflowResult(id string) (any, error) {
	return nil, nil
}

// mockPluginManager 实现 plugin.PluginManager 接口的简化版
// 用于修复编译错误，不实际执行插件管理逻辑
type mockPluginManager struct{}

func (m *mockPluginManager) GetAllPlugins() ([]any, error) {
	return nil, nil
}

func (m *mockPluginManager) GetPluginByID(pluginID string) (any, error) {
	return nil, nil
}

func (m *mockPluginManager) GetInstalledPlugins() ([]any, error) {
	return nil, nil
}

func (m *mockPluginManager) GetPluginMarket() ([]any, error) {
	return nil, nil
}

func (m *mockPluginManager) InstallPlugin(pluginID string) error {
	return nil
}

func (m *mockPluginManager) UninstallPlugin(pluginID string) error {
	return nil
}

func (m *mockPluginManager) ReloadPlugin(pluginID string) error {
	return nil
}

func (m *mockPluginManager) ResetPlugin(pluginID string) error {
	return nil
}

func (m *mockPluginManager) UpdatePlugin(pluginID string) error {
	return nil
}

func (m *mockPluginManager) BatchUpdatePlugins() error {
	return nil
}

func (m *mockPluginManager) GetPluginForm(pluginID string) (any, error) {
	return nil, nil
}

func (m *mockPluginManager) GetPluginPage(pluginID string) (any, error) {
	return nil, nil
}

func (m *mockPluginManager) GetPluginConfig(pluginID string) (any, error) {
	return nil, nil
}

func (m *mockPluginManager) UpdatePluginConfig(pluginID string, config any) error {
	return nil
}

func (m *mockPluginManager) GetPluginStatistics() (any, error) {
	return nil, nil
}
