package bootstrap

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// initDatabase 初始化数据库系统
func initDatabase(app *App) error {
	// 简化实现：使用SQLite数据库，固定连接到内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return err
	}
	app.DB = db

	// 返回nil，跳过迁移和初始化数据，因为这是简化实现
	return nil
}
