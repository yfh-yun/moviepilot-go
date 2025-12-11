package database

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	appLogger "moviepilot-go/pkg/logger"
)

// OptimizationConfig 优化配置
type OptimizationConfig struct {
	// 连接池配置
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 连接最大空闲时间

	// 性能配置
	EnablePreparedStmt bool // 启用预编译语句
	EnableQueryCache   bool // 启用查询缓存
}

// ProductionConfig 生产环境优化配置
func ProductionConfig() OptimizationConfig {
	return OptimizationConfig{
		MaxOpenConns:       100,              // 增加到100个连接
		MaxIdleConns:       10,               // 保持10个空闲连接
		ConnMaxLifetime:    time.Hour,        // 连接最多存活1小时
		ConnMaxIdleTime:    10 * time.Minute, // 空闲连接10分钟后关闭
		EnablePreparedStmt: true,             // 启用预编译语句
		EnableQueryCache:   true,             // 启用查询缓存
	}
}

// DevelopmentConfig 开发环境配置
func DevelopmentConfig() OptimizationConfig {
	return OptimizationConfig{
		MaxOpenConns:       25,
		MaxIdleConns:       5,
		ConnMaxLifetime:    30 * time.Minute,
		ConnMaxIdleTime:    5 * time.Minute,
		EnablePreparedStmt: true,
		EnableQueryCache:   false, // 开发环境关闭缓存
	}
}

// ApplyOptimization 应用优化配置
func ApplyOptimization(db *gorm.DB, config OptimizationConfig) error {
	logger := appLogger.GetLogger()

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("获取 sql.DB 失败", zap.Error(err))
		return err
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	logger.Info("数据库连接池优化配置已应用",
		zap.Int("max_open_conns", config.MaxOpenConns),
		zap.Int("max_idle_conns", config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", config.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", config.ConnMaxIdleTime))

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		logger.Error("数据库连接测试失败", zap.Error(err))
		return err
	}

	logger.Info("数据库连接测试成功")
	return nil
}

// GetConnectionPoolStats 获取连接池统计信息
func GetConnectionPoolStats(db *gorm.DB) (map[string]any, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()

	return map[string]any{
		"max_open_connections":   stats.MaxOpenConnections,
		"open_connections":       stats.OpenConnections,
		"in_use":                 stats.InUse,
		"idle":                   stats.Idle,
		"wait_count":             stats.WaitCount,
		"wait_duration_ms":       stats.WaitDuration.Milliseconds(),
		"max_idle_closed":        stats.MaxIdleClosed,
		"max_idle_time_closed":   stats.MaxIdleTimeClosed,
		"max_lifetime_closed":    stats.MaxLifetimeClosed,
		"utilization_percentage": float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100,
	}, nil
}

// MonitorConnectionPool 监控连接池（定期输出统计信息）
func MonitorConnectionPool(db *gorm.DB, interval time.Duration, stop <-chan struct{}) {
	logger := appLogger.GetLogger()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats, err := GetConnectionPoolStats(db)
			if err != nil {
				logger.Error("获取连接池统计失败", zap.Error(err))
				continue
			}

			logger.Info("连接池统计",
				zap.Int("max_open", stats["max_open_connections"].(int)),
				zap.Int("open", stats["open_connections"].(int)),
				zap.Int("in_use", stats["in_use"].(int)),
				zap.Int("idle", stats["idle"].(int)),
				zap.Float64("utilization", stats["utilization_percentage"].(float64)))

		case <-stop:
			logger.Info("连接池监控已停止")
			return
		}
	}
}

// VacuumAnalyze 执行 VACUUM ANALYZE（优化表和更新统计信息）
func VacuumAnalyze(db *gorm.DB, tables []string) error {
	logger := appLogger.GetLogger()
	ctx := context.Background()

	for _, table := range tables {
		logger.Info("执行 VACUUM ANALYZE", zap.String("table", table))

		sql := "VACUUM ANALYZE " + table
		if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
			logger.Error("VACUUM ANALYZE 失败",
				zap.String("table", table),
				zap.Error(err))
			return err
		}
	}

	logger.Info("VACUUM ANALYZE 完成", zap.Int("tables", len(tables)))
	return nil
}

// GetTableSizes 获取表大小统计
func GetTableSizes(db *gorm.DB) ([]map[string]any, error) {
	logger := appLogger.GetLogger()
	ctx := context.Background()

	query := `
		SELECT 
			schemaname,
			tablename,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
			pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS indexes_size,
			pg_total_relation_size(schemaname||'.'||tablename) AS total_bytes
		FROM pg_tables
		WHERE schemaname = 'public'
		ORDER BY total_bytes DESC
	`

	var results []map[string]any
	rows, err := db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		logger.Error("查询表大小失败", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			logger.Error("扫描表大小数据失败", zap.Error(err))
			continue
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}
