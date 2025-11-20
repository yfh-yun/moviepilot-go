package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// APIKeyMiddleware API密钥中间件
type APIKeyMiddleware struct {
	apiKeys map[string]bool
	logger  *zap.Logger
}

// NewAPIKeyMiddleware 创建API密钥中间件
func NewAPIKeyMiddleware(apiKeys []string, logger *zap.Logger) *APIKeyMiddleware {
	keyMap := make(map[string]bool)
	for _, key := range apiKeys {
		keyMap[key] = true
	}

	return &APIKeyMiddleware{
		apiKeys: keyMap,
		logger:  logger,
	}
}

// RequireAPIKey 需要API密钥的中间件
func (am *APIKeyMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从多个来源获取API密钥
		apiKey := am.getAPIKey(c)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    1002,
				"error":   "缺少API密钥",
			})
			c.Abort()
			return
		}

		// 验证API密钥
		if !am.apiKeys[apiKey] {
			am.logger.Warn("无效的API密钥尝试", 
				zap.String("ip", c.ClientIP()),
				zap.String("user_agent", c.GetHeader("User-Agent")))
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    1002,
				"error":   "无效的API密钥",
			})
			c.Abort()
			return
		}

		// 将API密钥信息存储到上下文
		c.Set("api_key", apiKey)
		
		c.Next()
	}
}

// getAPIKey 从请求中获取API密钥
func (am *APIKeyMiddleware) getAPIKey(c *gin.Context) string {
	// 1. 从Authorization头获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return authHeader[7:]
		}
		if strings.HasPrefix(authHeader, "API-Key ") {
			return authHeader[8:]
		}
	}

	// 2. 从X-API-Key头获取
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// 3. 从查询参数获取
	apiKey = c.Query("api_key")
	if apiKey != "" {
		return apiKey
	}

	return ""
}

// AddAPIKey 添加新的API密钥
func (am *APIKeyMiddleware) AddAPIKey(apiKey string) {
	am.apiKeys[apiKey] = true
}

// RemoveAPIKey 移除API密钥
func (am *APIKeyMiddleware) RemoveAPIKey(apiKey string) {
	delete(am.apiKeys, apiKey)
}

// HasAPIKey 检查API密钥是否存在
func (am *APIKeyMiddleware) HasAPIKey(apiKey string) bool {
	return am.apiKeys[apiKey]
}