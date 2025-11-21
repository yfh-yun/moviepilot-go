package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"moviepilot-go/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQLClient PostgreSQL客户端
type PostgreSQLClient struct {
	pool   *pgxpool.Pool
	config *PostgreSQLConfig
}

// PostgreSQLConfig PostgreSQL配置
type PostgreSQLConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Database        string        `json:"database"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	SSLMode         string        `json:"ssl_mode"`
	MaxConnections  int           `json:"max_connections"`
	MinConnections  int           `json:"min_connections"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time"`
	ConnectTimeout  time.Duration `json:"connect_timeout"`
}

// NewPostgreSQLClient 创建PostgreSQL客户端
func NewPostgreSQLClient(config *PostgreSQLConfig) (*PostgreSQLClient, error) {
	if config == nil {
		return nil, fmt.Errorf("PostgreSQL config is required")
	}

	// 设置默认值
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.MaxConnections == 0 {
		config.MaxConnections = 20
	}
	if config.MinConnections == 0 {
		config.MinConnections = 5
	}
	if config.MaxConnLifetime == 0 {
		config.MaxConnLifetime = time.Hour
	}
	if config.MaxConnIdleTime == 0 {
		config.MaxConnIdleTime = time.Minute * 30
	}
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = time.Second * 10
	}

	// 构建连接字符串
	connString := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		config.Host, config.Port, config.Database, config.Username, config.Password, config.SSLMode,
	)

	// 配置连接池
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL config: %w", err)
	}

	poolConfig.MaxConns = int32(config.MaxConnections)
	poolConfig.MinConns = int32(config.MinConnections)
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = time.Minute
	poolConfig.ConnectTimeout = config.ConnectTimeout

	// 创建连接池
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
	}

	client := &PostgreSQLClient{
		pool:   pool,
		config: config,
	}

	// 测试连接
	if err := client.Test(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("PostgreSQL connection test failed: %w", err)
	}

	logger.Info("PostgreSQL client connected successfully to %s:%d/%s", config.Host, config.Port, config.Database)
	return client, nil
}

// Test 测试连接
func (c *PostgreSQLClient) Test(ctx context.Context) error {
	if c.pool == nil {
		return fmt.Errorf("PostgreSQL connection pool is not initialized")
	}

	// 执行简单查询测试连接
	var result int
	err := c.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("PostgreSQL connection test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("PostgreSQL connection test returned unexpected result: %d", result)
	}

	logger.Info("PostgreSQL connection test successful")
	return nil
}

// Close 关闭连接
func (c *PostgreSQLClient) Close() error {
	if c.pool != nil {
		c.pool.Close()
		logger.Info("PostgreSQL connection pool closed")
	}
	return nil
}

// GetPool 获取连接池
func (c *PostgreSQLClient) GetPool() *pgxpool.Pool {
	return c.pool
}

// GetStats 获取连接池统计信息
func (c *PostgreSQLClient) GetStats() *pgxpool.Stat {
	if c.pool == nil {
		return nil
	}
	stats := c.pool.Stat()
	return &stats
}

// Execute 执行SQL语句
func (c *PostgreSQLClient) Execute(ctx context.Context, query string, args ...interface{}) (int64, error) {
	if c.pool == nil {
		return 0, fmt.Errorf("PostgreSQL connection pool is not initialized")
	}

	commandTag, err := c.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute SQL: %w", err)
	}

	return commandTag.RowsAffected(), nil
}

// QueryRow 查询单行
func (c *PostgreSQLClient) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	if c.pool == nil {
		return nil
	}
	return c.pool.QueryRow(ctx, query, args...)
}

// Query 查询多行
func (c *PostgreSQLClient) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	if c.pool == nil {
		return nil, fmt.Errorf("PostgreSQL connection pool is not initialized")
	}

	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	return rows, nil
}

// QueryJSON 查询并返回JSON结果
func (c *PostgreSQLClient) QueryJSON(ctx context.Context, query string, args ...interface{}) ([]byte, error) {
	rows, err := c.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 获取列信息
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// 读取所有行
	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to get row values: %w", err)
		}

		row := make(map[string]interface{})
		for i, value := range values {
			row[columns[i]] = value
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// 转换为JSON
	return json.Marshal(results)
}

// BeginTransaction 开始事务
func (c *PostgreSQLClient) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	if c.pool == nil {
		return nil, fmt.Errorf("PostgreSQL connection pool is not initialized")
	}

	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

// ExecuteInTransaction 在事务中执行多个SQL
func (c *PostgreSQLClient) ExecuteInTransaction(ctx context.Context, queries []QueryWithArgs) error {
	tx, err := c.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, query := range queries {
		_, err := tx.Exec(ctx, query.Query, query.Args...)
		if err != nil {
			return fmt.Errorf("failed to execute query in transaction: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryWithArgs 查询和参数
type QueryWithArgs struct {
	Query string
	Args  []interface{}
}

// CreateTable 创建表
func (c *PostgreSQLClient) CreateTable(ctx context.Context, tableName string, columns []ColumnDefinition) error {
	if len(columns) == 0 {
		return fmt.Errorf("columns are required")
	}

	var columnDefs []string
	for _, col := range columns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)

		if col.PrimaryKey {
			def += " PRIMARY KEY"
		}

		if col.NotNull {
			def += " NOT NULL"
		}

		if col.Unique {
			def += " UNIQUE"
		}

		if col.Default != "" {
			def += fmt.Sprintf(" DEFAULT %s", col.Default)
		}

		columnDefs = append(columnDefs, def)
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columnDefs, ", "))

	_, err := c.Execute(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	logger.Info("Table %s created successfully", tableName)
	return nil
}

// ColumnDefinition 列定义
type ColumnDefinition struct {
	Name       string
	Type       string
	PrimaryKey bool
	NotNull    bool
	Unique     bool
	Default    string
}

// DropTable 删除表
func (c *PostgreSQLClient) DropTable(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName)

	_, err := c.Execute(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %w", tableName, err)
	}

	logger.Info("Table %s dropped successfully", tableName)
	return nil
}

// TableExists 检查表是否存在
func (c *PostgreSQLClient) TableExists(ctx context.Context, tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`

	var exists bool
	err := c.QueryRow(ctx, query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if table %s exists: %w", tableName, err)
	}

	return exists, nil
}

// GetTableInfo 获取表信息
func (c *PostgreSQLClient) GetTableInfo(ctx context.Context, tableName string) (*TableInfo, error) {
	if !c.TableExists(ctx, tableName) {
		return nil, fmt.Errorf("table %s does not exist", tableName)
	}

	// 获取列信息
	columnsQuery := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := c.Query(ctx, columnsQuery, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get column info for table %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var maxLength sql.NullInt64

		err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &col.Default, &maxLength)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}

		if maxLength.Valid {
			col.MaxLength = int(maxLength.Int64)
		}

		columns = append(columns, col)
	}

	// 获取索引信息
	indexesQuery := `
		SELECT 
			indexname,
			indexdef
		FROM pg_indexes
		WHERE tablename = $1
	`

	indexRows, err := c.Query(ctx, indexesQuery, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get index info for table %s: %w", tableName, err)
	}
	defer indexRows.Close()

	var indexes []IndexInfo
	for indexRows.Next() {
		var idx IndexInfo
		err := indexRows.Scan(&idx.Name, &idx.Definition)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index info: %w", err)
		}
		indexes = append(indexes, idx)
	}

	return &TableInfo{
		Name:    tableName,
		Columns: columns,
		Indexes: indexes,
	}, nil
}

// TableInfo 表信息
type TableInfo struct {
	Name    string       `json:"name"`
	Columns []ColumnInfo `json:"columns"`
	Indexes []IndexInfo  `json:"indexes"`
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Nullable  string `json:"nullable"`
	Default   string `json:"default"`
	MaxLength int    `json:"max_length"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

// BackupTable 备份表
func (c *PostgreSQLClient) BackupTable(ctx context.Context, tableName, backupName string) error {
	if !c.TableExists(ctx, tableName) {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	query := fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s", backupName, tableName)

	_, err := c.Execute(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to backup table %s to %s: %w", tableName, backupName, err)
	}

	logger.Info("Table %s backed up to %s successfully", tableName, backupName)
	return nil
}

// RestoreTable 恢复表
func (c *PostgreSQLClient) RestoreTable(ctx context.Context, tableName, backupName string) error {
	if !c.TableExists(ctx, backupName) {
		return fmt.Errorf("backup table %s does not exist", backupName)
	}

	// 删除原表
	if err := c.DropTable(ctx, tableName); err != nil {
		return fmt.Errorf("failed to drop original table %s: %w", tableName, err)
	}

	// 重命名备份表
	query := fmt.Sprintf("ALTER TABLE %s RENAME TO %s", backupName, tableName)

	_, err := c.Execute(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to restore table %s from %s: %w", tableName, backupName, err)
	}

	logger.Info("Table %s restored from %s successfully", tableName, backupName)
	return nil
}
