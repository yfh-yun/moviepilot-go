package bootstrap

import (
	"fmt"
	"moviepilot-go/pkg/database"
)

// initDatabase 初始化数据库系统
func initDatabase(app *App) error {
	// 使用PostgreSQL数据库
	dbConfig := database.Config{
		Host:     app.Config.Database.PostgreSQLHost,
		Port:     app.Config.Database.PostgreSQLPort,
		User:     app.Config.Database.PostgreSQLUsername,
		Password: app.Config.Database.PostgreSQLPassword,
		DBName:   app.Config.Database.PostgreSQLDatabase,
		SSLMode:  "disable",
		MaxOpenConns:    app.Config.Database.PostgreSQLPoolSize,
		MaxIdleConns:    app.Config.Database.PostgreSQLPoolSize / 2,
		ConnMaxLifetime: 300,
		ConnMaxIdleTime: 10,
	}
	
	// 连接到PostgreSQL数据库
	db, err := database.Connect(dbConfig, app.Logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	app.DB = db

	// 返回nil，迁移和初始化数据将在其他地方处理
	return nil
}
