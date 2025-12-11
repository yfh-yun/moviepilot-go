package bootstrap

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/internal/infrastructure/config"
	"moviepilot-go/pkg/logger"
)

// App 应用上下文，持有所有已初始化的组件
type App struct {
	Config  *config.Config
	DB      *gorm.DB
	Modules *ModuleRegistry
	Router  *gin.Engine
	Logger  *zap.Logger
}

// Bootstrap 统一初始化入口
func Bootstrap() (*App, error) {
	app := &App{}

	// 1. 初始化日志
	if err := initLogger(); err != nil {
		return nil, err
	}
	app.Logger = logger.GetLogger()
	app.Logger.Info("Starting MoviePilot Go server...")

	// 2. 加载配置
	if err := initConfig(app); err != nil {
		return nil, err
	}

	// 3. 初始化数据库
	if err := initDatabase(app); err != nil {
		return nil, err
	}

	// 4. 初始化模块
	if err := initModules(app); err != nil {
		return nil, err
	}

	// 5. 初始化路由
	if err := initRouter(app); err != nil {
		return nil, err
	}

	return app, nil
}

// Shutdown 优雅关闭
func (app *App) Shutdown(ctx context.Context) error {
	app.Logger.Info("Shutting down server...")

	// 按相反顺序关闭各组件
	if app.Modules != nil {
		app.Modules.Stop()
	}

	if app.DB != nil {
		sqlDB, err := app.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	app.Logger.Info("Server exited gracefully")
	return nil
}
