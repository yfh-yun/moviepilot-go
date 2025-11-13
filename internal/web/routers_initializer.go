package web

import (
	"github.com/gin-gonic/gin"
)

// InitRouters 初始化路�?func InitRouters(engine *gin.Engine) {
	// 注册API路由
	registerAPIRoutes(engine)
	
	// 注册Radarr、Sonarr路由
	registerArrRoutes(engine)
	
	// 注册CookieCloud路由
	registerCookieRoutes(engine)
	
	// 添加健康检查端�?	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "MoviePilot-Go is running",
		})
	})
	
	// 添加OpenAPI文档端点（模拟FastAPI的openapi.json�?	engine.GET("/api/v1/openapi.json", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"openapi": "3.0.0",
			"info": gin.H{
				"title":   "MoviePilot-Go",
				"version": "1.0.0",
			},
			"paths": gin.H{},
		})
	})
}

// registerAPIRoutes 注册API路由
func registerAPIRoutes(engine *gin.Engine) {
	// 对应Python中的 app.include_router(api_router, prefix=settings.API_V1_STR)
	// 这里假设settings.API_V1_STR�?/api/v1"
	
	api := engine.Group("/api/v1")
	{
		// 在这里添加具体的API路由
		api.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "API路由已注�?,
			})
		})
		
		// TODO: 添加更多API路由
	}
}

// registerArrRoutes 注册Radarr、Sonarr路由
func registerArrRoutes(engine *gin.Engine) {
	// 对应Python中的 app.include_router(arr_router, prefix="/api/v3")
	
	arr := engine.Group("/api/v3")
	{
		// 在这里添加Radarr、Sonarr相关路由
		arr.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Radarr/Sonarr路由已注�?,
			})
		})
		
		// TODO: 添加更多Radarr/Sonarr路由
	}
}

// registerCookieRoutes 注册CookieCloud路由
func registerCookieRoutes(engine *gin.Engine) {
	// 对应Python中的 app.include_router(cookie_router, prefix="/cookiecloud")
	
	cookie := engine.Group("/cookiecloud")
	{
		// 在这里添加CookieCloud相关路由
		cookie.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "CookieCloud路由已注�?,
			})
		})
		
		// TODO: 添加更多CookieCloud路由
	}
}
