package routes

import (
	"github.com/yfh-yun/moviepilot-go/internal/api/handlers/site"
	"github.com/yfh-yun/moviepilot-go/internal/middleware"
	"github.com/yfh-yun/moviepilot-go/internal/repository/repositories"
	"github.com/yfh-yun/moviepilot-go/internal/service/site"

	"github.com/gin-gonic/gin"
)

// RegisterSiteRoutes 注册站点路由
func RegisterSiteRoutes(router *gin.RouterGroup) {
	// 创建仓储实例
	siteRepo := repositories.NewSiteRepository()
	siteUserDataRepo := repositories.NewSiteUserDataRepository()
	siteStatisticRepo := repositories.NewSiteStatisticRepository()
	siteIconRepo := repositories.NewSiteIconRepository()

	// 创建服务实例
	siteService := site.NewSiteService(siteRepo, siteUserDataRepo, siteStatisticRepo, siteIconRepo)

	// 创建处理器实例
	siteHandler := site.NewSiteHandler(siteService)

	// 站点管理路由组
	siteGroup := router.Group("/sites")
	siteGroup.Use(middleware.RequireAuth())

	{
		// 站点操作
		siteGroup.POST("", siteHandler.CreateSite)
		siteGroup.GET("", siteHandler.ListSites)
		siteGroup.GET("/search", siteHandler.SearchSites)
		siteGroup.POST("/import", siteHandler.ImportSites)

		// 特定站点操作
		siteGroup.GET("/:id", siteHandler.GetSiteByID)
		siteGroup.PUT("/:id", siteHandler.UpdateSite)
		siteGroup.DELETE("/:id", siteHandler.DeleteSite)
		siteGroup.PUT("/:id/toggle-active", siteHandler.ToggleSiteActive)
		siteGroup.PUT("/:id/cookie", siteHandler.UpdateSiteCookie)
		siteGroup.PUT("/:id/settings", siteHandler.UpdateSiteSettings)

		// 按名称操作
		siteGroup.GET("/name/:name", siteHandler.GetSiteByName)
		siteGroup.GET("/:name/statistics", siteHandler.GetSiteStatistics)
		siteGroup.GET("/:name/user-data", siteHandler.GetSiteUserData)

		// 特殊站点类型
		siteGroup.GET("/active", siteHandler.GetActiveSites)
		siteGroup.GET("/rss", siteHandler.GetRSSSites)
		siteGroup.GET("/search-enabled", siteHandler.GetSearchSites)
	}
}
