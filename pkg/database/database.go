// Package database MoviePilot数据库管理模块
package database

import (
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/config"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repository/migrations"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var (
	// DB 全局数据库实例
	DB *gorm.DB
)

// Init 初始化数据库连接
func Init() error {
	dsn := config.GetDatabaseDSN()
	driver := config.Config.GetString("database.driver")

	logger.Info("Initializing database connection",
		zap.String("driver", driver),
		zap.String("host", config.Config.GetString("database.host")),
		zap.Int("port", config.Config.GetInt("database.port")),
		zap.String("database", config.Config.GetString("database.name")),
	)

	// 配置GORM日志
	gormLogLevel := gormLogger.Silent
	if config.IsDebug() {
		gormLogLevel = gormLogger.Info
	}

	// 创建数据库连接
	var err error
	switch driver {
	case "postgres":
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogLevel),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
		})
	case "mysql":
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogLevel),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
		})
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogLevel),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
		})
	default:
		// 默认使用PostgreSQL
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogLevel),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
		})
	}

	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return err
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Error("Failed to get underlying sql.DB", zap.Error(err))
		return err
	}

	// 配置连接池
	maxOpenConns := config.Config.GetInt("database.max_open_conns")
	if maxOpenConns == 0 {
		maxOpenConns = 100
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := config.Config.GetInt("database.max_idle_conns")
	if maxIdleConns == 0 {
		maxIdleConns = 10
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	connMaxLifetime := config.Config.GetInt("database.conn_max_lifetime")
	if connMaxLifetime == 0 {
		connMaxLifetime = 3600
	}
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return err
	}

	logger.Info("Database connection established successfully")

	// 自动迁移数据库表
	if err := autoMigrate(); err != nil {
		logger.Error("Failed to auto migrate database", zap.Error(err))
		return err
	}

	return nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	logger.Info("Starting database auto migration")

	// 使用迁移模块进行完整的数据库迁移
	migration := migrations.NewMigration(DB)
	if err := migration.Run(); err != nil {
		return err
	}

	logger.Info("Database auto migration completed")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Health 检查数据库健康状态
func Health() error {
	if DB == nil {
		return gorm.ErrInvalidTransaction
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// GetStats 获取数据库统计信息
func GetStats() map[string]interface{} {
	if DB == nil {
		return map[string]interface{}{
			"connected": false,
		}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return map[string]interface{}{
			"connected": false,
			"error":     err.Error(),
		}
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"connected":            true,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
