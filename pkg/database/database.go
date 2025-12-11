package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	appLogger "moviepilot-go/pkg/logger"
)

// Config 数据库配置
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	LogLevel        logger.LogLevel
}

// getLogger 返回可用的 *zap.Logger，如果调用方未提供，则回退到全局应用日志。
func getLogger(log *zap.Logger) *zap.Logger {
	if log != nil {
		return log
	}
	return appLogger.GetLogger()
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		DBName:          "moviepilot",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        logger.Info,
	}
}

// Connect 连接到 PostgreSQL 数据库
func Connect(config Config, log *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	// 配置 GORM，使用 zap 驱动的 GORM Logger，将 SQL 日志纳入统一日志体系
	gormLogger := NewZapGormLogger(getLogger(log)).LogMode(config.LogLevel)
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		l := getLogger(log)
		l.Error("failed to connect to database",
			zap.String("host", config.Host),
			zap.Int("port", config.Port),
			zap.String("database", config.DBName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		l := getLogger(log)
		l.Error("failed to get sql.DB", zap.Error(err))
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		l := getLogger(log)
		l.Error("failed to ping database",
			zap.String("host", config.Host),
			zap.Int("port", config.Port),
			zap.String("database", config.DBName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	getLogger(log).Info("database connected successfully",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.DBName))

	return db, nil
}

// Close 关闭数据库连接
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to get sql.DB when closing database", zap.Error(err))
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		appLogger.Error("failed to close database", zap.Error(err))
		return err
	}

	return nil
}

// HealthCheck 健康检查
func HealthCheck(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to get sql.DB in health check", zap.Error(err))
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		appLogger.Error("database ping failed", zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Stats 获取数据库统计信息
func Stats(db *gorm.DB) (map[string]any, error) {
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to get sql.DB when collecting stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	stats := sqlDB.Stats()

	return map[string]any{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}
