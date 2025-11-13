package db

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/pkg/models"
)

var DB *gorm.DB

// InitDB 初始化数据库 - 保持向后兼容
func InitDB() error {
	appConfig := config.GetConfig()
	
	// 确保数据库目录存在
	dbDir := filepath.Dir(appConfig.DBPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err = os.MkdirAll(dbDir, 0755)
		if err != nil {
			return fmt.Errorf("创建数据库目录失败: %v", err)
		}
	}
	
	// 使用新的初始化方法
	err := Init()
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	
	DB = GetDB()
	
	// 获取通用数据库对象 sql.DB 以进行后续操作
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库对象失败: %v", err)
	}
	
	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	
	// 初始化表结构
	err = initTables()
	if err != nil {
		return fmt.Errorf("初始化数据库表结构失败: %v", err)
	}
	
	logger.GetLoggerManager().Info("数据库初始化成功")
	return nil
}

// UpdateDB 更新数据库（在Go中通常通过GORM的自动迁移实现）
func UpdateDB() error {
	// 在GORM中，数据库更新通常通过AutoMigrate实现
	// 这里执行自动迁移来更新数据库结构
	err := updateTables()
	if err != nil {
		logger.GetLoggerManager().Error(fmt.Sprintf("数据库更新失败: %v", err))
		return err
	}
	
	logger.GetLoggerManager().Info("数据库更新检查完成")
	return nil
}

// initTables 初始化数据库表结构
func initTables() error {
	// 自动迁移创建表结构
	return DB.AutoMigrate(
		&models.User{},
		&models.SystemConfig{},
		&models.DownloadHistory{},
		&models.DownloadFiles{},
	)
}

// updateTables 更新数据库表结构
func updateTables() error {
	// 使用GORM的自动迁移功能来更新表结构
	return DB.AutoMigrate(
		&models.User{},
		&models.SystemConfig{},
		&models.DownloadHistory{},
		&models.DownloadFiles{},
	)
}