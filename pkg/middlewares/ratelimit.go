package middlewares

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	cacheRedis "moviepilot-go/pkg/cache/redis"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/response"
)

// RateLimitMiddleware 基于令牌桶的限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 构造 Redis key
		key := fmt.Sprintf("ratelimit:%s", clientIP)

		// 使用 Redis 实现令牌桶算法
		// 默认配置：每秒 10 个请求，突发 20 个
		limit := 10.0
		burst := 20

		// 检查限流
		allowed, err := cacheRedis.AllowRequest(key, limit, burst, time.Second)
		if err != nil {
			logger.Error("rate limit check failed",
				zap.String("key", key),
				zap.Error(err),
			)
			// 如果 Redis 失败，允许请求通过（降级）
			c.Next()
			return
		}

		if !allowed {
			logger.Warn("rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.String("path", c.FullPath()),
				zap.Float64("limit", limit),
				zap.Int("burst", burst),
			)

			response.RateLimit(c)
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", burst-1))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

		c.Next()
	}
}

// APIRateLimitMiddleware 针对 API 的更严格限流
func APIRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对 API 路径进行限流
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("api_ratelimit:%s", clientIP)

		// API 更严格的限制：每秒 5 个请求，突发 10 个
		limit := 5.0
		burst := 10

		allowed, err := cacheRedis.AllowRequest(key, limit, burst, time.Second)
		if err != nil {
			logger.Error("API rate limit check failed",
				zap.String("key", key),
				zap.Error(err),
			)
			c.Next()
			return
		}

		if !allowed {
			logger.Warn("API rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.String("path", c.FullPath()),
				zap.Float64("limit", limit),
				zap.Int("burst", burst),
			)

			response.Error(c, response.CodeRateLimit, "API 请求过于频繁，请稍后再试")
			return
		}

		c.Next()
	}
}
