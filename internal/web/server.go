package web

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
)

// Server Web服务器结�?type Server struct {
	// Gin引擎
	engine *gin.Engine
	
	// HTTP服务�?	httpServer *http.Server
	
	// 配置
	config *config.AppConfig
	
	// 日志记录�?	logger *logger.LoggerManager
}

// NewServer 创建新的Web服务器实�?func NewServer() *Server {
	// 获取配置
	appConfig := config.GetConfig()
	
	// 获取日志记录�?	logManager := logger.GetLoggerManager()
	
	// 设置Gin模式
	if appConfig.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// 创建Gin引擎
	engine := gin.New()
	
	// 添加日志中间�?	engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	
	// 添加恢复中间�?	engine.Use(gin.Recovery())
	
	// 配置CORS中间�?	engine.Use(cors.New(cors.Config{
		AllowOrigins:     appConfig.AllowedHosts,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	// 根据CPU核心数设置工作协程数，类似Python版本中的 workers=multiprocessing.cpu_count() * 2 + 1
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	server := &Server{
		engine: engine,
		config: appConfig,
		logger: logManager,
	}
	
	// 初始化路�?	InitRouters(engine)
	
	return server
}

// Start 启动Web服务�?func (s *Server) Start() error {
	// 创建HTTP服务�?	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.engine,
	}
	
	s.logger.Info(fmt.Sprintf("正在启动Web服务器，监听端口 %d", s.config.Port))
	
	// 启动HTTP服务�?	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	
	return nil
}

// Stop 停止Web服务�?func (s *Server) Stop() error {
	s.logger.Info("正在停止Web服务�?..")
	
	if s.httpServer != nil {
		// 创建一�?0秒的超时上下文，与Python版本中的timeout_graceful_shutdown=60保持一�?		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		
		// 优雅地关闭服务器
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error(fmt.Sprintf("Web服务器关闭错�? %v", err))
			return err
		}
		
		s.logger.Info("Web服务器已停止")
	}
	
	return nil
}

// GetEngine 获取Gin引擎实例
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}
