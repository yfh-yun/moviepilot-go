package bootstrap

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"moviepilot-go/internal/apis/middleware"
	"moviepilot-go/internal/apis/routes"
)

// initRouter 初始化路由系统
func initRouter(app *App) error {
	// 设置Gin模式
	if !app.Config.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 初始化Gin引擎
	engine := gin.New()

	// 添加必要的中间件
	engine.Use(gin.Recovery())   // 替换 middlewares.RecoveryMiddleware
	engine.Use(corsMiddleware()) // 自定义 CORS 中间件，替换 middlewares.CORSMiddleware

	// 添加请求日志中间件
	engine.Use(middleware.RequestLoggingMiddleware())

	// 健康检查端点
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   "2.8.1", // 可以从配置或常量中获取
			"port":      app.Config.App.Port,
		})
	})

	// 注册API路由
	if err := routes.Register(engine, routes.Config{
		DB:                      app.DB,
		JWTSecretKey:            app.Config.Security.SecretKey,
		AccessTokenExpireMinute: app.Config.Security.AccessTokenExpireMinutes,
	}); err != nil {
		return err
	}

	app.Router = engine
	return nil
}

// corsMiddleware 自定义CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
