package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"moviepilot-go/internal/business/services/auth"
)

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(permissionService auth.PermissionService, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未认证",
			})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := permissionService.CheckPermission(c.Request.Context(), userID, requiredPermission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "权限检查失败",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 要求特定角色
func RequireRole(permissionService auth.PermissionService, roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未认证",
			})
			c.Abort()
			return
		}

		// 检查角色
		hasRole, err := permissionService.HasRole(c.Request.Context(), userID, roleName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "角色检查失败",
			})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "需要" + roleName + "角色",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 要求管理员角色
func RequireAdmin(permissionService auth.PermissionService) gin.HandlerFunc {
	return RequireRole(permissionService, "admin")
}
