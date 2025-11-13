package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/driver/postgres"
	
	"moviepilot-go/internal/config"
	"moviepilot-go/internal/logger"
)

// 数据库引擎实例
var (
	Engine       *gorm.DB
	AsyncEngine  *gorm.DB
	dbMutex      sync.Mutex
)

// 数据库配置
type dbConfig struct {
	Type     string
	Host     string
	Port     string
	Database string
	Username string
	Password string
	Path     string
	WalEnable bool
	Timeout   int
	PoolType  string
	PoolPrePing bool
	PoolRecycle int
	SqlitePoolSize int
	PoolTimeout int
	SqliteMaxOverflow int
	PostgreSQLPoolSize int
	PostgreSQLMaxOverflow int
	Echo bool
}

// Init 初始化数据库连接
func Init() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if Engine != nil {
		return nil // 已经初始化
	}

	appConfig := config.GetConfig()
	cfg := getDBConfig(appConfig)

	var err error
	Engine, err = createEngine(cfg, false)
	if err != nil {
		return fmt.Errorf("failed to create database engine: %v", err)
	}

	AsyncEngine, err = createEngine(cfg, true)
	if err != nil {
		return fmt.Errorf("failed to create async database engine: %v", err)
	}

	logger.GetLoggerManager().Info("数据库初始化成功")
	return nil
}

// getDBConfig 根据应用配置获取数据库配置
func getDBConfig(appConfig *config.Config) *dbConfig {
	cfg := &dbConfig{
		Type:     appConfig.DBType,
		WalEnable: appConfig.DBWalEnable,
		Timeout:  appConfig.DBTimeout,
		PoolType: appConfig.DBPoolType,
		PoolPrePing: appConfig.DBPoolPrePing,
		PoolRecycle: appConfig.DBPoolRecycle,
		SqlitePoolSize: appConfig.DBSQLitePoolSize,
		PoolTimeout: appConfig.DBPoolTimeout,
		SqliteMaxOverflow: appConfig.DBSQLiteMaxOverflow,
		PostgreSQLPoolSize: appConfig.DBPostgreSQLPoolSize,
		PostgreSQLMaxOverflow: appConfig.DBPostgreSQLMaxOverflow,
		Echo: appConfig.DBEcho,
	}

	if cfg.Type == "postgresql" {
		cfg.Host = appConfig.DBPostgreSQLHost
		cfg.Port = appConfig.DBPostgreSQLPort
		cfg.Database = appConfig.DBPostgreSQLDatabase
		cfg.Username = appConfig.DBPostgreSQLUsername
		cfg.Password = appConfig.DBPostgreSQLPassword
	} else {
		cfg.Path = appConfig.DBPath
	}

	return cfg
}

// createEngine 创建数据库引擎
func createEngine(cfg *dbConfig, isAsync bool) (*gorm.DB, error) {
	var dialector gorm.Dialector
	
	if cfg.Type == "postgresql" {
		var dsn string
		if cfg.Password != "" {
			dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
				cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port)
		} else {
			dsn = fmt.Sprintf("host=%s user=%s dbname=%s port=%s",
				cfg.Host, cfg.Username, cfg.Database, cfg.Port)
		}
		
		dialector = postgres.New(postgres.Config{
			DSN: dsn,
		})
	} else {
		// SQLite
		connStr := cfg.Path
		if cfg.WalEnable {
			connStr += "?_pragma=journal_mode(WAL)"
		}
		dialector = sqlite.Open(connStr)
	}
	
	logLevel := logger.Silent
	if cfg.Echo {
		logLevel = logger.Info
	}
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}
	
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}
	
	// 配置连接池
	if sqlDB, err := db.DB(); err == nil {
		// 设置连接池参数
		if cfg.PoolType != "NullPool" {
			if cfg.Type == "postgresql" {
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetMaxOpenConns(cfg.PostgreSQLPoolSize)
			} else {
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetMaxOpenConns(cfg.SqlitePoolSize)
			}
		}
		
		// 设置连接超时
		// 注意: GORM/数据库驱动的超时配置方式不同，这里仅设置连接池相关参数
	}
	
	return db, nil
}

// GetDB 获取数据库会话（用于WEB请求）
func GetDB() *gorm.DB {
	return Engine
}

// GetAsyncDB 获取异步数据库会话
func GetAsyncDB() *gorm.DB {
	return AsyncEngine
}

// CloseDatabase 关闭所有数据库连接并清理资源
func CloseDatabase() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	
	if Engine != nil {
		sqlDB, err := Engine.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %v", err)
		}
		Engine = nil
	}
	
	if AsyncEngine != nil {
		sqlDB, err := AsyncEngine.DB()
		if err != nil {
			return fmt.Errorf("failed to get async database instance: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close async database: %v", err)
		}
		AsyncEngine = nil
	}
	
	return nil
}

// DBUpdate 数据库更新类操作装饰器
func DBUpdate(fn func(db *gorm.DB) error) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// 开启事务
		return db.Transaction(fn)
	}
}

// AsyncDBUpdate 异步数据库更新类操作装饰器
func AsyncDBUpdate(fn func(db *gorm.DB) error) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// 开启事务
		return db.Transaction(fn)
	}
}

// DBQuery 数据库查询操作装饰器
func DBQuery(fn func(db *gorm.DB) error) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// 直接执行查询
		return fn(db)
	}
}

// AsyncDBQuery 异步数据库查询操作装饰器
func AsyncDBQuery(fn func(db *gorm.DB) error) func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		// 直接执行查询
		return fn(db)
	}
}

// GetIDColumn 根据数据库类型返回合适的ID列定义
func GetIDColumn() string {
	// GORM默认使用ID作为主键名，类型为uint时自动设置为自增
	return "id"
}

// WithTransaction 为函数添加事务支持
func WithTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}

// WithContext 为数据库操作添加上下文
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx)
}