// Package middleware 全局中间件
package middlewares

import (
	"go.uber.org/zap"
)

var (
	// 全局中间件实例（用于向后兼容）
	GlobalLogger *zap.Logger
)

// JWTAuthMiddleware 返回JWT认证中间件（向后兼容）
// 注意：这是一个临时的实现，实际使用时需要注入相应的服务
func JWTAuthMiddleware() func(interface{}) interface{} {
	return func(next interface{}) interface{} {
		return func(c interface{}) interface{} {
			// 这里需要实际的JWT认证逻辑
			// 暂时返回一个空的中间件
			return func() {}
		}
	}
}

// RequireAPIKey 返回API密钥中间件（向后兼容）
// 注意：这是一个临时的实现，实际使用时需要注入相应的服务
func RequireAPIKey() func(interface{}) interface{} {
	return func(next interface{}) interface{} {
		return func(c interface{}) interface{} {
			// 这里需要实际的API密钥验证逻辑
			// 暂时返回一个空的中间件
			return func() {}
		}
	}
}