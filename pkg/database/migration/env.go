// Package migration 数据库迁移包
package migration

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/config"
	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// Env 迁移环境配置
type Env struct {
	config         *config.DatabaseConfig
	dbManager      *database.Manager
	migrationPath  string
	isOfflineMode  bool
	offlineURL     string
	targetMetadata *MetadataRegistry
}

// NewEnv 创建迁移环境
func NewEnv(cfg *config.DatabaseConfig, dbManager *database.Manager) *Env {
	return &Env{
		config:         cfg,
		dbManager:      dbManager,
		migrationPath:  "internal/database/migrations",
		targetMetadata: NewMetadataRegistry(),
	}
}

// SetOfflineMode 设置离线模式
func (e *Env) SetOfflineMode(isOffline bool, url string) {
	e.isOfflineMode = isOffline
	e.offlineURL = url
}

// GetOfflineURL 获取离线模式URL
func (e *Env) GetOfflineURL() string {
	if e.offlineURL != "" {
		return e.offlineURL
	}
	return e.config.GetDSN()
}

// RunMigrationsOffline 在离线模式下运行迁移
func (e *Env) RunMigrationsOffline(ctx context.Context) error {
	logger.Info("开始离线模式数据库迁移")

	if e.offlineURL == "" {
		e.offlineURL = e.GetOfflineURL()
	}

	// 根据数据库类型配置不同的参数
	if e.isPostgreSQL() {
		return e.runPostgreSQLOffline(ctx, e.offlineURL)
	} else {
		return e.runSQLiteOffline(ctx, e.offlineURL)
	}
}

// RunMigrationsOnline 在在线模式下运行迁移
func (e *Env) RunMigrationsOnline(ctx context.Context) error {
	logger.Info("开始在线模式数据库迁移")

	// 创建连接
	db, err := e.dbManager.GetConnection()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 根据数据库类型配置不同的参数
	if e.isPostgreSQL() {
		return e.runPostgreSQLOnline(ctx, db)
	} else {
		return e.runSQLiteOnline(ctx, db)
	}
}

// isPostgreSQL 检查是否为PostgreSQL数据库
func (e *Env) isPostgreSQL() bool {
	return strings.Contains(strings.ToLower(e.config.Type), "postgres")
}

// runPostgreSQLOffline PostgreSQL离线迁移
func (e *Env) runPostgreSQLOffline(ctx context.Context, url string) error {
	logger.Info("PostgreSQL离线模式迁移")

	config := &MigrationConfig{
		URL:            url,
		TargetMetadata: e.targetMetadata,
		LiteralBinds:   true,
		DialectOptions: map[string]interface{}{
			"paramstyle": "named",
		},
		DatabaseType:  "postgresql",
		IsOfflineMode: true,
	}

	migrator := NewMigrator(config)
	return migrator.Run(ctx)
}

// runSQLiteOffline SQLite离线迁移
func (e *Env) runSQLiteOffline(ctx context.Context, url string) error {
	logger.Info("SQLite离线模式迁移")

	config := &MigrationConfig{
		URL:            url,
		TargetMetadata: e.targetMetadata,
		LiteralBinds:   true,
		DialectOptions: map[string]interface{}{
			"paramstyle": "named",
		},
		RenderAsBatch: true,
		DatabaseType:  "sqlite",
		IsOfflineMode: true,
	}

	migrator := NewMigrator(config)
	return migrator.Run(ctx)
}

// runPostgreSQLOnline PostgreSQL在线迁移
func (e *Env) runPostgreSQLOnline(ctx context.Context, db *database.Connection) error {
	logger.Info("PostgreSQL在线模式迁移")

	config := &MigrationConfig{
		Connection:     db,
		TargetMetadata: e.targetMetadata,
		DatabaseType:   "postgresql",
		IsOfflineMode:  false,
	}

	migrator := NewMigrator(config)
	return migrator.Run(ctx)
}

// runSQLiteOnline SQLite在线迁移
func (e *Env) runSQLiteOnline(ctx context.Context, db *database.Connection) error {
	logger.Info("SQLite在线模式迁移")

	config := &MigrationConfig{
		Connection:     db,
		TargetMetadata: e.targetMetadata,
		RenderAsBatch:  true,
		DatabaseType:   "sqlite",
		IsOfflineMode:  false,
	}

	migrator := NewMigrator(config)
	return migrator.Run(ctx)
}

// ShouldRunMigrations 判断是否需要运行迁移
func (e *Env) ShouldRunMigrations(ctx context.Context) (bool, error) {
	if e.isOfflineMode {
		return true, nil
	}

	db, err := e.dbManager.GetConnection()
	if err != nil {
		return false, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 检查迁移表是否存在
	exists, err := db.TableExists(ctx, "schema_migrations")
	if err != nil {
		return false, fmt.Errorf("检查迁移表失败: %w", err)
	}

	if !exists {
		logger.Info("迁移表不存在，需要创建和运行迁移")
		return true, nil
	}

	// 获取当前数据库版本
	currentVersion, err := e.getCurrentVersion(ctx, db)
	if err != nil {
		return false, fmt.Errorf("获取当前版本失败: %w", err)
	}

	// 获取最新版本
	latestVersion, err := e.getLatestVersion()
	if err != nil {
		return false, fmt.Errorf("获取最新版本失败: %w", err)
	}

	shouldMigrate := currentVersion < latestVersion
	if shouldMigrate {
		logger.Info("检测到新版本，需要运行迁移",
			zap.String("当前版本", currentVersion),
			zap.String("最新版本", latestVersion))
	}

	return shouldMigrate, nil
}

// getCurrentVersion 获取当前数据库版本
func (e *Env) getCurrentVersion(ctx context.Context, db *database.Connection) (string, error) {
	if !e.isOfflineMode {
		query := "SELECT version_num FROM schema_migrations ORDER BY version_num DESC LIMIT 1"
		row := db.QueryRow(ctx, query)
		var version string
		if err := row.Scan(&version); err != nil {
			return "", nil // 没有记录返回空字符串
		}
		return version, nil
	}
	return "", nil
}

// getLatestVersion 获取最新迁移版本
func (e *Env) getLatestVersion() (string, error) {
	migrations, err := e.loadMigrationFiles()
	if err != nil {
		return "", err
	}

	if len(migrations) == 0 {
		return "", nil
	}

	return migrations[len(migrations)-1].Version, nil
}

// loadMigrationFiles 加载迁移文件
func (e *Env) loadMigrationFiles() ([]*MigrationFile, error) {
	// 这里应该加载迁移文件
	// 为简化，返回空列表
	return []*MigrationFile{}, nil
}

// GetMigrationPath 获取迁移文件路径
func (e *Env) GetMigrationPath() string {
	return e.migrationPath
}

// SetMigrationPath 设置迁移文件路径
func (e *Env) SetMigrationPath(path string) {
	e.migrationPath = path
}

// ValidateConfig 验证配置
func (e *Env) ValidateConfig() error {
	if e.config == nil {
		return fmt.Errorf("数据库配置不能为空")
	}

	if !e.isOfflineMode && e.dbManager == nil {
		return fmt.Errorf("在线模式下需要数据库管理器")
	}

	if e.targetMetadata == nil {
		return fmt.Errorf("目标元数据不能为空")
	}

	return nil
}

// GetDatabaseType 获取数据库类型
func (e *Env) GetDatabaseType() string {
	if e.isPostgreSQL() {
		return "postgresql"
	}
	return "sqlite"
}

// GetConnectionInfo 获取连接信息
type ConnectionInfo struct {
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Database     string `json:"database"`
	Username     string `json:"username"`
	SSLMode      string `json:"ssl_mode"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
}

// GetConnectionInfo 获取连接信息
func (e *Env) GetConnectionInfo() *ConnectionInfo {
	info := &ConnectionInfo{
		Type:         e.config.Type,
		MaxOpenConns: e.config.MaxOpenConns,
		MaxIdleConns: e.config.MaxIdleConns,
	}

	if e.config.Type == "postgres" {
		info.Host = e.config.Host
		info.Port = e.config.Port
		info.Database = e.config.Name
		info.Username = e.config.Username
		info.SSLMode = e.config.SSLMode
	} else {
		info.Database = e.config.DatabasePath
	}

	return info
}

// CreateMigrationEnvironment 创建迁移环境
func CreateMigrationEnvironment(cfg *config.DatabaseConfig) (*Env, error) {
	env := &Env{
		config:         cfg,
		targetMetadata: NewMetadataRegistry(),
	}

	// 验证配置
	if err := env.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("创建迁移环境失败: %w", err)
	}

	return env, nil
}
