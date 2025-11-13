package db

import (
	"fmt"
	"os"
	"path/filepath"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
	"moviepilot-go/pkg/models"
)

// InitTables 初始化数据库表结构
// 对应Python版本的init_db函数
func InitTables() error {
	appConfig := config.GetConfig()
	
	// 确保数据库目录存在
	dbDir := filepath.Dir(appConfig.DBPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err = os.MkdirAll(dbDir, 0755)
		if err != nil {
			return fmt.Errorf("创建数据库目录失败: %v", err)
		}
	}
	
	// 初始化数据库连接
	err := Init()
	if err != nil {
		return fmt.Errorf("初始化数据库连接失败: %v", err)
	}
	
	// 自动迁移创建表结构
	err = GetDB().AutoMigrate(
		&models.User{},
		&models.SystemConfig{},
		&models.DownloadHistory{},
		&models.DownloadFiles{},
	)
	
	if err != nil {
		return fmt.Errorf("初始化数据库表结构失败: %v", err)
	}
	
	logger.GetLoggerManager().Info("数据库表结构初始化成功")
	return nil
}

// UpdateTables 更新数据库表结构
// 对应Python版本的update_db函数
func UpdateTables() error {
	appConfig := config.GetConfig()
	
	// 初始化数据库连接
	err := Init()
	if err != nil {
		return fmt.Errorf("初始化数据库连接失败: %v", err)
	}
	
	// 自动迁移更新表结构
	err = GetDB().AutoMigrate(
		&models.User{},
		&models.SystemConfig{},
		&models.DownloadHistory{},
		&models.DownloadFiles{},
	)
	
	if err != nil {
		logger.GetLoggerManager().Error(fmt.Sprintf("更新数据库表结构失败: %v", err))
		return err
	}
	
	logger.GetLoggerManager().Info("数据库表结构更新完成")
	return nil
}