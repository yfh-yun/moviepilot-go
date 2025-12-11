package bootstrap

import (
	"moviepilot-go/internal/infrastructure/config"
)

// initConfig 初始化配置系统
func initConfig(app *App) error {
	// 使用配置加载器，从默认值 + .env + 环境变量加载完整配置
	loader := config.NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		return err
	}

	app.Config = cfg
	return nil
}
