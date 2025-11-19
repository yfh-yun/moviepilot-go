package routes

import (
	"github.com/yfh-yun/moviepilot-go/internal/api/handlers/search"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/service/search"

	"github.com/gin-gonic/gin"
)

// RegisterSearchRoutes 注册搜索相关路由
func RegisterSearchRoutes(router *gin.RouterGroup, searchService *search.SearchService, logger *logger.Logger) {
	searchHandler := search.NewSearchHandler(searchService, logger)

	searchGroup := router.Group("/search")
	{
		// 搜索接口
		searchGroup.POST("/media", searchHandler.MediaSearch)
		searchGroup.POST("/torrent", searchHandler.TorrentSearch)
		searchGroup.POST("/site", searchHandler.SiteSearch)

		// 搜索历史
		searchGroup.GET("/history", searchHandler.GetSearchHistory)
		searchGroup.DELETE("/history", searchHandler.ClearSearchHistory)

		// 搜索建议和热门
		searchGroup.GET("/suggestions", searchHandler.GetSearchSuggestions)
		searchGroup.GET("/trending", searchHandler.GetTrendingSearches)
	}
}
