package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	workflowapi "moviepilot-go/internal/api/workflow"
	wf "moviepilot-go/internal/platform/workflow"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/middlewares"
)

// Config 提供路由注册所需的依赖。
type Config struct {
	// Logger 日志记录器
	Logger *zap.Logger
}

// Register 统一注册 API 路由。
func Register(engine *gin.Engine, cfg Config) error {
	// 检查 gin 引擎是否为空
	if engine == nil {
		return fmt.Errorf("gin engine cannot be nil")
	}

	// 获取日志记录器
	log := cfg.Logger
	if log == nil {
		log = logger.GetLogger()
	}

	api := engine.Group("/api")
	api.Use(middlewares.AuthMiddleware())
	api.Use(middlewares.APIRateLimitMiddleware())

	workflowGroup := api.Group("/workflows")
	workflowManager := wf.NewWorkflowManager(log)
	workflowService := workflowapi.NewService(workflowManager, wf.LocalFileWorkflowConfig{Logger: log}, log)
	workflowHandler := workflowapi.NewHandler(workflowService, log)
	workflowGroup.POST("/local-file-scrape-transfer", workflowHandler.StartLocalFileWorkflow)

	return nil
}
