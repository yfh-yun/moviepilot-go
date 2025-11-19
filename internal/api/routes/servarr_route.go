package routes

import (
	"github.com/yfh-yun/moviepilot-go/internal/api/handlers/servarr"
	"github.com/yfh-yun/moviepilot-go/internal/middleware"
	"github.com/yfh-yun/moviepilot-go/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupServArrRoutes 设置ServArr路由
func SetupServArrRoutes(
	router *gin.RouterGroup,
	servarrHandler *servarr.ServArrHandler,
	authService service.AuthService,
) {
	// ServArr路由组，需要API Key认证
	servarrGroup := router.Group("/servarr")
	servarrGroup.Use(middleware.RequireAPIKey(authService))
	{
		// 系统相关
		servarrGroup.GET("/system/status", servarrHandler.GetSystemStatus)
		
		// 配置相关
		servarrGroup.GET("/qualityProfile", servarrHandler.GetQualityProfiles)
		servarrGroup.GET("/rootfolder", servarrHandler.GetRootFolders)
		servarrGroup.GET("/tag", servarrHandler.GetTags)
		servarrGroup.GET("/languageprofile", servarrHandler.GetLanguageProfiles)
		
		// 电影相关
		servarrGroup.GET("/movie", servarrHandler.GetMovies)
		servarrGroup.GET("/movie/lookup", servarrHandler.LookupMovie)
		servarrGroup.GET("/movie/:mid", servarrHandler.GetMovie)
		servarrGroup.POST("/movie", servarrHandler.AddMovie)
		servarrGroup.DELETE("/movie/:mid", servarrHandler.DeleteMovie)
		
		// 剧集相关
		servarrGroup.GET("/series", servarrHandler.GetSeries)
		// TODO: 实现更多剧集相关API
		// servarrGroup.GET("/series/lookup", servarrHandler.LookupSeries)
		// servarrGroup.GET("/series/:sid", servarrHandler.GetSeries)
		// servarrGroup.POST("/series", servarrHandler.AddSeries)
		// servarrGroup.DELETE("/series/:sid", servarrHandler.DeleteSeries)
	}
}