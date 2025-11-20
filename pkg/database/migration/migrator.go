// Package migration 数据库迁移包
package migration

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/database"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// MigrationConfig 迁移配置
type MigrationConfig struct {
	// 离线模式参数
	URL            string                 `json:"url"`
	TargetMetadata *MetadataRegistry      `json:"target_metadata"`
	LiteralBinds   bool                   `json:"literal_binds"`
	DialectOptions map[string]interface{} `json:"dialect_options"`
	RenderAsBatch  bool                   `json:"render_as_batch"`

	// 在线模式参数
	Connection *database.Connection `json:"connection"`

	// 通用参数
	DatabaseType  string `json:"database_type"`
	IsOfflineMode bool   `json:"is_offline_mode"`
}

// Migrator 数据库迁移器
type Migrator struct {
	config            *MigrationConfig
	connection        *database.Connection
	migrations        []*Migration
	appliedMigrations map[string]bool
}

// Migration 迁移定义
type Migration struct {
	Version       string                                            `json:"version"`
	DownRevision  string                                            `json:"down_revision"`
	BranchLabels  []string                                          `json:"branch_labels"`
	DependsOn     []string                                          `json:"depends_on"`
	UpgradeFunc   func(context.Context, *database.Connection) error `json:"-"`
	DowngradeFunc func(context.Context, *database.Connection) error `json:"-"`
	CreateDate    time.Time                                         `json:"create_date"`
	Description   string                                            `json:"description"`
}

// MigrationFile 迁移文件
type MigrationFile struct {
	Filename     string    `json:"filename"`
	Version      string    `json:"version"`
	DownRevision string    `json:"down_revision"`
	BranchLabels []string  `json:"branch_labels"`
	DependsOn    []string  `json:"depends_on"`
	Content      string    `json:"content"`
	UpgradeSQL   string    `json:"upgrade_sql"`
	DowngradeSQL string    `json:"downgrade_sql"`
	CreateDate   time.Time `json:"create_date"`
}

// NewMigrator 创建迁移器
func NewMigrator(config *MigrationConfig) *Migrator {
	return &Migrator{
		config:            config,
		migrations:        []*Migration{},
		appliedMigrations: make(map[string]bool),
	}
}

// Run 运行迁移
func (m *Migrator) Run(ctx context.Context) error {
	logger.Info("开始运行数据库迁移")

	// 初始化连接
	if err := m.initializeConnection(ctx); err != nil {
		return fmt.Errorf("初始化数据库连接失败: %w", err)
	}
	defer m.cleanup()

	// 创建迁移表
	if err := m.createMigrationTable(ctx); err != nil {
		return fmt.Errorf("创建迁移表失败: %w", err)
	}

	// 加载已应用的迁移
	if err := m.loadAppliedMigrations(ctx); err != nil {
		return fmt.Errorf("加载已应用迁移失败: %w", err)
	}

	// 加载所有迁移文件
	if err := m.loadMigrationFiles(ctx); err != nil {
		return fmt.Errorf("加载迁移文件失败: %w", err)
	}

	// 排序迁移
	m.sortMigrations()

	// 执行待应用的迁移
	if err := m.applyMigrations(ctx); err != nil {
		return fmt.Errorf("应用迁移失败: %w", err)
	}

	logger.Info("数据库迁移完成")
	return nil
}

// initializeConnection 初始化数据库连接
func (m *Migrator) initializeConnection(ctx context.Context) error {
	if m.config.IsOfflineMode {
		// 离线模式，不需要实际连接
		return nil
	}

	if m.config.Connection != nil {
		m.connection = m.config.Connection
		return nil
	}

	// 创建连接（这里需要根据实际的数据库管理器创建）
	return fmt.Errorf("在线模式下必须提供数据库连接")
}

// cleanup 清理资源
func (m *Migrator) cleanup() {
	// 清理连接等资源
}

// createMigrationTable 创建迁移表
func (m *Migrator) createMigrationTable(ctx context.Context) error {
	if m.config.IsOfflineMode {
		// 离线模式下，跳过表创建
		return nil
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version_num VARCHAR(32) NOT NULL PRIMARY KEY
	)`

	if _, err := m.connection.Exec(ctx, createTableSQL); err != nil {
		return fmt.Errorf("创建迁移表失败: %w", err)
	}

	logger.Info("迁移表已创建或已存在")
	return nil
}

// loadAppliedMigrations 加载已应用的迁移
func (m *Migrator) loadAppliedMigrations(ctx context.Context) error {
	if m.config.IsOfflineMode {
		// 离线模式下，假设没有已应用的迁移
		return nil
	}

	query := "SELECT version_num FROM schema_migrations"
	rows, err := m.connection.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("查询已应用迁移失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("扫描迁移版本失败: %w", err)
		}
		m.appliedMigrations[version] = true
	}

	logger.Info("加载已应用迁移", zap.Int("数量", len(m.appliedMigrations)))
	return nil
}

// loadMigrationFiles 加载迁移文件
func (m *Migrator) loadMigrationFiles(ctx context.Context) error {
	// 这里应该从文件系统加载迁移文件
	// 为简化，直接创建一些示例迁移

	// 2.0.0版本迁移
	migrations := m.getDefinedMigrations()

	for _, migration := range migrations {
		m.migrations = append(m.migrations, migration)
	}

	logger.Info("加载迁移文件", zap.Int("数量", len(m.migrations)))
	return nil
}

// getDefinedMigrations 获取定义的迁移
func (m *Migrator) getDefinedMigrations() []*Migration {
	return []*Migration{
		{
			Version:       "294b007932ef",
			DownRevision:  "",
			Description:   "2.0.0版本迁移",
			CreateDate:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			UpgradeFunc:   m.upgradeTo200,
			DowngradeFunc: m.downgradeFrom200,
		},
		{
			Version:       "5b3355c964bb",
			DownRevision:  "294b007932ef",
			Description:   "2.2.0版本迁移 - PostgreSQL序列修复",
			CreateDate:    time.Date(2025, 8, 19, 12, 27, 8, 0, time.UTC),
			UpgradeFunc:   m.upgradeTo220,
			DowngradeFunc: m.downgradeFrom220,
		},
		{
			Version:       "a946dae52526",
			DownRevision:  "5b3355c964bb",
			Description:   "2.2.1版本迁移 - PostgreSQL用户ID迁移",
			CreateDate:    time.Date(2025, 8, 20, 17, 50, 0, 0, time.UTC),
			UpgradeFunc:   m.upgradeTo221,
			DowngradeFunc: m.downgradeFrom221,
		},
	}
}

// sortMigrations 排序迁移
func (m *Migrator) sortMigrations() {
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})
}

// applyMigrations 应用迁移
func (m *Migrator) applyMigrations(ctx context.Context) error {
	appliedCount := 0

	for _, migration := range m.migrations {
		if m.appliedMigrations[migration.Version] {
			logger.Debug("迁移已应用", zap.String("版本", migration.Version))
			continue
		}

		logger.Info("应用迁移",
			zap.String("版本", migration.Version),
			zap.String("描述", migration.Description))

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("应用迁移 %s 失败: %w", migration.Version, err)
		}

		appliedCount++
	}

	if appliedCount == 0 {
		logger.Info("所有迁移都已应用")
	} else {
		logger.Info("成功应用迁移", zap.Int("数量", appliedCount))
	}

	return nil
}

// applyMigration 应用单个迁移
func (m *Migrator) applyMigration(ctx context.Context, migration *Migration) error {
	if migration.UpgradeFunc == nil {
		logger.Warn("迁移没有升级函数", zap.String("版本", migration.Version))
		return nil
	}

	// 开始事务
	if m.config.IsOfflineMode {
		// 离线模式，直接记录迁移
		return m.recordMigrationOffline(ctx, migration)
	}

	tx, err := m.connection.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 执行迁移
	if err := migration.UpgradeFunc(ctx, &database.Connection{DB: tx}); err != nil {
		return fmt.Errorf("执行迁移升级失败: %w", err)
	}

	// 记录迁移
	if err := m.recordMigration(ctx, tx, migration.Version); err != nil {
		return fmt.Errorf("记录迁移失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	logger.Info("迁移应用成功", zap.String("版本", migration.Version))
	return nil
}

// recordMigration 记录迁移
func (m *Migrator) recordMigration(ctx context.Context, tx *sql.Tx, version string) error {
	insertSQL := "INSERT INTO schema_migrations (version_num) VALUES (?)"
	_, err := tx.ExecContext(ctx, insertSQL, version)
	return err
}

// recordMigrationOffline 离线模式记录迁移
func (m *Migrator) recordMigrationOffline(ctx context.Context, migration *Migration) error {
	logger.Info("离线模式记录迁移", zap.String("版本", migration.Version))
	return nil
}

// upgradeTo200 2.0.0版本升级
func (m *Migrator) upgradeTo200(ctx context.Context, conn *database.Connection) error {
	// 这里实现2.0.0版本的升级逻辑
	// 例如：创建新表、修改现有表结构等

	// 示例：创建用户配置表
	createUserConfigSQL := `
	CREATE TABLE IF NOT EXISTS user_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id VARCHAR(100) NOT NULL,
		config_key VARCHAR(100) NOT NULL,
		config_value TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := conn.Exec(ctx, createUserConfigSQL); err != nil {
		return fmt.Errorf("创建用户配置表失败: %w", err)
	}

	logger.Info("2.0.0版本升级完成")
	return nil
}

// downgradeFrom200 2.0.0版本降级
func (m *Migrator) downgradeFrom200(ctx context.Context, conn *database.Connection) error {
	// 降级逻辑
	dropUserConfigSQL := "DROP TABLE IF EXISTS user_config"
	if _, err := conn.Exec(ctx, dropUserConfigSQL); err != nil {
		return fmt.Errorf("删除用户配置表失败: %w", err)
	}

	logger.Info("2.0.0版本降级完成")
	return nil
}

// upgradeTo220 2.2.0版本升级
func (m *Migrator) upgradeTo220(ctx context.Context, conn *database.Connection) error {
	if m.config.DatabaseType == "postgresql" {
		return m.fixPostgreSQLSequences(ctx, conn)
	}

	logger.Info("2.2.0版本升级完成（非PostgreSQL）")
	return nil
}

// downgradeFrom220 2.2.0版本降级
func (m *Migrator) downgradeFrom220(ctx context.Context, conn *database.Connection) error {
	logger.Info("2.2.0版本降级完成")
	return nil
}

// upgradeTo221 2.2.1版本升级
func (m *Migrator) upgradeTo221(ctx context.Context, conn *database.Connection) error {
	if m.config.DatabaseType == "postgresql" {
		return m.migratePostgreSQLUserID(ctx, conn)
	}

	logger.Info("2.2.1版本升级完成（非PostgreSQL）")
	return nil
}

// downgradeFrom221 2.2.1版本降级
func (m *Migrator) downgradeFrom221(ctx context.Context, conn *database.Connection) error {
	logger.Info("2.2.1版本降级完成")
	return nil
}

// fixPostgreSQLSequences 修复PostgreSQL序列问题
func (m *Migrator) fixPostgreSQLSequences(ctx context.Context, conn *database.Connection) error {
	logger.Info("开始修复PostgreSQL序列问题")

	// 获取所有表名
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("查询表名失败: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("扫描表名失败: %w", err)
		}
		tables = append(tables, tableName)
	}

	logger.Info("发现表", zap.Int("数量", len(tables)))

	// 为每个表修复序列
	for _, tableName := range tables {
		if err := m.fixTableSequence(ctx, conn, tableName); err != nil {
			logger.Warn("修复表序列失败",
				zap.String("表名", tableName),
				zap.Error(err))
			continue
		}
	}

	logger.Info("PostgreSQL序列修复完成")
	return nil
}

// fixTableSequence 修复单个表的序列
func (m *Migrator) fixTableSequence(ctx context.Context, conn *database.Connection, tableName string) error {
	// 获取表的主键列
	columnQuery := `
		SELECT column_name 
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_name = $1 AND tc.constraint_type = 'PRIMARY KEY'
	`

	rows, err := conn.Query(ctx, columnQuery, tableName)
	if err != nil {
		return fmt.Errorf("查询主键列失败: %w", err)
	}
	defer rows.Close()

	var primaryKey string
	for rows.Next() {
		if err := rows.Scan(&primaryKey); err != nil {
			return fmt.Errorf("扫描主键列失败: %w", err)
		}
		break // 取第一个主键列
	}

	if primaryKey == "" {
		logger.Debug("表没有主键，跳过", zap.String("表名", tableName))
		return nil
	}

	// 创建序列并设置默认值
	sequenceName := fmt.Sprintf("%s_%s_seq", tableName, primaryKey)

	// 删除现有序列（如果存在）
	dropSeqSQL := fmt.Sprintf("DROP SEQUENCE IF EXISTS %s", sequenceName)
	if _, err := conn.Exec(ctx, dropSeqSQL); err != nil {
		return fmt.Errorf("删除序列失败: %w", err)
	}

	// 创建序列
	createSeqSQL := fmt.Sprintf("CREATE SEQUENCE %s", sequenceName)
	if _, err := conn.Exec(ctx, createSeqSQL); err != nil {
		return fmt.Errorf("创建序列失败: %w", err)
	}

	// 设置列为IDENTITY
	identitySQL := fmt.Sprintf(`
		ALTER TABLE %s ALTER COLUMN %s 
		SET GENERATED BY DEFAULT AS IDENTITY
	`, tableName, primaryKey)

	if _, err := conn.Exec(ctx, identitySQL); err != nil {
		// 如果IDENTITY不支持，使用序列
		defaultSQL := fmt.Sprintf(`
			ALTER TABLE %s ALTER COLUMN %s 
			SET DEFAULT nextval('%s')
		`, tableName, primaryKey, sequenceName)

		if _, err := conn.Exec(ctx, defaultSQL); err != nil {
			return fmt.Errorf("设置默认值失败: %w", err)
		}
	}

	logger.Debug("修复表序列完成",
		zap.String("表名", tableName),
		zap.String("主键", primaryKey))

	return nil
}

// migratePostgreSQLUserID PostgreSQL用户ID迁移
func (m *Migrator) migratePostgreSQLUserID(ctx context.Context, conn *database.Connection) error {
	logger.Info("开始PostgreSQL用户ID迁移")

	// 1. 创建临时列
	addColumnSQL := `
		ALTER TABLE siteuserdata 
		ADD COLUMN IF NOT EXISTS userid_new VARCHAR
	`

	if _, err := conn.Exec(ctx, addColumnSQL); err != nil {
		return fmt.Errorf("添加临时列失败: %w", err)
	}

	// 2. 迁移数据
	migrateSQL := `
		UPDATE siteuserdata 
		SET userid_new = CAST(userid AS VARCHAR)
		WHERE userid_new IS NULL
	`

	if _, err := conn.Exec(ctx, migrateSQL); err != nil {
		return fmt.Errorf("迁移数据失败: %w", err)
	}

	// 3. 删除旧列
	dropColumnSQL := `
		ALTER TABLE siteuserdata 
		DROP COLUMN IF EXISTS userid
	`

	if _, err := conn.Exec(ctx, dropColumnSQL); err != nil {
		return fmt.Errorf("删除旧列失败: %w", err)
	}

	// 4. 重命名新列
	renameColumnSQL := `
		ALTER TABLE siteuserdata 
		RENAME COLUMN userid_new TO userid
	`

	if _, err := conn.Exec(ctx, renameColumnSQL); err != nil {
		return fmt.Errorf("重命名列失败: %w", err)
	}

	logger.Info("PostgreSQL用户ID迁移完成")
	return nil
}

// GetMigrationStatus 获取迁移状态
func (m *Migrator) GetMigrationStatus(ctx context.Context) (*MigrationStatus, error) {
	status := &MigrationStatus{
		TotalMigrations:   len(m.migrations),
		AppliedMigrations: len(m.appliedMigrations),
		PendingMigrations: 0,
		DatabaseType:      m.config.DatabaseType,
		IsOfflineMode:     m.config.IsOfflineMode,
		MigrationDetails:  make([]*MigrationDetail, 0),
	}

	for _, migration := range m.migrations {
		detail := &MigrationDetail{
			Version:      migration.Version,
			DownRevision: migration.DownRevision,
			Description:  migration.Description,
			CreateDate:   migration.CreateDate,
			Applied:      m.appliedMigrations[migration.Version],
		}

		status.MigrationDetails = append(status.MigrationDetails, detail)
		if !detail.Applied {
			status.PendingMigrations++
		}
	}

	return status, nil
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	TotalMigrations   int                `json:"total_migrations"`
	AppliedMigrations int                `json:"applied_migrations"`
	PendingMigrations int                `json:"pending_migrations"`
	DatabaseType      string             `json:"database_type"`
	IsOfflineMode     bool               `json:"is_offline_mode"`
	CurrentVersion    string             `json:"current_version"`
	LatestVersion     string             `json:"latest_version"`
	MigrationDetails  []*MigrationDetail `json:"migration_details"`
}

// MigrationDetail 迁移详情
type MigrationDetail struct {
	Version      string    `json:"version"`
	DownRevision string    `json:"down_revision"`
	Description  string    `json:"description"`
	CreateDate   time.Time `json:"create_date"`
	Applied      bool      `json:"applied"`
	AppliedAt    time.Time `json:"applied_at"`
}
