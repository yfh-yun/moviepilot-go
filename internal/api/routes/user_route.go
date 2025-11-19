// Package routes 用户管理路由配置
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/yfh-yun/moviepilot-go/internal/api/handlers/user"
	"github.com/yfh-yun/moviepilot-go/internal/middleware"
)

// SetupUserRoutes 设置用户管理路由
func SetupUserRoutes(rg *gin.RouterGroup, userHandler *user.UserHandler) {
	// 用户管理路由组
	userGroup := rg.Group("/user")
	userGroup.Use(middleware.JWTAuthMiddleware())
	{
		// 用户基本信息
		userGroup.GET("/profile", userHandler.GetUserProfile)
		userGroup.PUT("/profile", userHandler.UpdateUserProfile)

		// 用户设置
		userGroup.GET("/settings", userHandler.GetUserSettings)
		userGroup.PUT("/settings", userHandler.UpdateUserSettings)

		// 用户统计和活动
		userGroup.GET("/stats", userHandler.GetUserStats)
		userGroup.GET("/activity", userHandler.GetUserActivity)
	}

	// 管理员用户管理路由组
	adminGroup := rg.Group("/admin")
	adminGroup.Use(middleware.JWTAuthMiddleware())
	{
		// 用户列表
		adminGroup.GET("/users", userHandler.ListUsers)

		// 启用/禁用用户
		adminGroup.POST("/users/:user_id/enable", userHandler.EnableUser)
		adminGroup.POST("/users/:user_id/disable", userHandler.DisableUser)
	}
}
