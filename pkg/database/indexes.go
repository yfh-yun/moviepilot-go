package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	appLogger "moviepilot-go/pkg/logger"
)

// IndexDefinition 索引定义
type IndexDefinition struct {
	Name  string
	Table string
	SQL   string
}

// GetIndexDefinitions 获取所有索引定义
func GetIndexDefinitions() []IndexDefinition {
	return []IndexDefinition{
		// Subscribe 表索引
		{
			Name:  "idx_subscribes_user_state_updated",
			Table: "subscribes",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_subscribes_user_state_updated ON subscribes(username, state, updated_at DESC)",
		},
		{
			Name:  "idx_subscribes_tmdb_type",
			Table: "subscribes",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_subscribes_tmdb_type ON subscribes(tmdb_id, type) WHERE tmdb_id IS NOT NULL",
		},
		{
			Name:  "idx_subscribes_state_last_update",
			Table: "subscribes",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_subscribes_state_last_update ON subscribes(state, last_update DESC) WHERE state IN ('R', 'N')",
		},

		// DownloadHistory 表索引
		{
			Name:  "idx_download_history_hash_state",
			Table: "downloadhistories",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_download_history_hash_state ON downloadhistories(download_hash, state)",
		},
		{
			Name:  "idx_download_history_user_date",
			Table: "downloadhistories",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_download_history_user_date ON downloadhistories(username, date DESC)",
		},
		{
			Name:  "idx_download_history_tmdb_type",
			Table: "downloadhistories",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_download_history_tmdb_type ON downloadhistories(tmdb_id, type) WHERE tmdb_id IS NOT NULL",
		},

		// TransferHistory 表索引
		{
			Name:  "idx_transfer_history_src_status",
			Table: "transfer_histories",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_transfer_history_src_status ON transfer_histories(src, status)",
		},
		{
			Name:  "idx_transfer_history_hash_date",
			Table: "transfer_histories",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_transfer_history_hash_date ON transfer_histories(download_hash, date DESC)",
		},
		{
			Name:  "idx_transfer_history_tmdb_type",
			Table: "transfer_histories",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_transfer_history_tmdb_type ON transfer_histories(tmdb_id, type) WHERE tmdb_id IS NOT NULL",
		},

		// Site 表索引
		{
			Name:  "idx_sites_active_pri",
			Table: "sites",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_sites_active_pri ON sites(is_active, pri DESC) WHERE is_active = true",
		},
		{
			Name:  "idx_sites_domain_active",
			Table: "sites",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_sites_domain_active ON sites(domain, is_active)",
		},

		// SiteUserData 表索引
		{
			Name:  "idx_site_user_data_name_updated",
			Table: "siteuserdatas",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_site_user_data_name_updated ON siteuserdatas(site_name, updated_at DESC)",
		},
		{
			Name:  "idx_site_user_data_userid",
			Table: "siteuserdatas",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_site_user_data_userid ON siteuserdatas(userid)",
		},

		// Download 表索引
		{
			Name:  "idx_downloads_status_created",
			Table: "downloads",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_downloads_status_created ON downloads(status, created_at DESC)",
		},
		{
			Name:  "idx_downloads_subscribe_id",
			Table: "downloads",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_downloads_subscribe_id ON downloads(subscribe_id) WHERE subscribe_id IS NOT NULL",
		},

		// Media 表索引
		{
			Name:  "idx_medias_tmdb_type",
			Table: "medias",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_medias_tmdb_type ON medias(tmdb_id, type) WHERE tmdb_id IS NOT NULL",
		},
		{
			Name:  "idx_medias_title_year",
			Table: "medias",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_medias_title_year ON medias(title, year)",
		},

		// PluginData 表索引
		{
			Name:  "idx_plugin_data_composite",
			Table: "plugindatas",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_plugin_data_composite ON plugindatas(plugin_key, data_key, userid)",
		},

		// SystemConfig 表索引
		{
			Name:  "idx_system_config_key",
			Table: "systemconfigs",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_system_config_key ON systemconfigs(key)",
		},

		// UserConfig 表索引
		{
			Name:  "idx_user_config_userid_key",
			Table: "userconfigs",
			SQL:   "CREATE INDEX IF NOT EXISTS idx_user_config_userid_key ON userconfigs(userid, key)",
		},
	}
}

// CreateOptimizedIndexes 创建所有优化索引
func CreateOptimizedIndexes(db *gorm.DB) error {
	logger := appLogger.GetLogger()
	ctx := context.Background()

	logger.Info("开始创建数据库索引...")
	startTime := time.Now()

	indexes := GetIndexDefinitions()
	successCount := 0
	failCount := 0

	for _, idx := range indexes {
		logger.Debug("创建索引",
			zap.String("name", idx.Name),
			zap.String("table", idx.Table))

		if err := db.WithContext(ctx).Exec(idx.SQL).Error; err != nil {
			logger.Error("创建索引失败",
				zap.String("name", idx.Name),
				zap.String("table", idx.Table),
				zap.Error(err))
			failCount++
			continue
		}

		successCount++
		logger.Debug("索引创建成功", zap.String("name", idx.Name))
	}

	duration := time.Since(startTime)
	logger.Info("索引创建完成",
		zap.Int("total", len(indexes)),
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
		zap.Duration("duration", duration))

	if failCount > 0 {
		return fmt.Errorf("部分索引创建失败: %d/%d", failCount, len(indexes))
	}

	return nil
}

// DropOptimizedIndexes 删除所有优化索引（用于测试或重建）
func DropOptimizedIndexes(db *gorm.DB) error {
	logger := appLogger.GetLogger()
	ctx := context.Background()

	logger.Info("开始删除数据库索引...")

	indexes := GetIndexDefinitions()
	for _, idx := range indexes {
		sql := fmt.Sprintf("DROP INDEX IF EXISTS %s", idx.Name)
		if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
			logger.Warn("删除索引失败",
				zap.String("name", idx.Name),
				zap.Error(err))
		}
	}

	logger.Info("索引删除完成")
	return nil
}

// AnalyzeIndexUsage 分析索引使用情况
func AnalyzeIndexUsage(db *gorm.DB) ([]map[string]any, error) {
	logger := appLogger.GetLogger()
	ctx := context.Background()

	// PostgreSQL 索引使用统计查询
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_scan as scans,
			idx_tup_read as tuples_read,
			idx_tup_fetch as tuples_fetched,
			pg_size_pretty(pg_relation_size(indexrelid)) as size
		FROM pg_stat_user_indexes
		WHERE schemaname = 'public'
		ORDER BY idx_scan ASC, tablename, indexname
	`

	var results []map[string]any
	rows, err := db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		logger.Error("查询索引使用情况失败", zap.Error(err))
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
			logger.Error("扫描索引数据失败", zap.Error(err))
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

// GetSlowQueries 获取慢查询
func GetSlowQueries(db *gorm.DB, minDuration time.Duration) ([]map[string]any, error) {
	logger := appLogger.GetLogger()
	ctx := context.Background()

	// 需要先启用 pg_stat_statements 扩展
	// CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

	query := `
		SELECT 
			query,
			calls,
			total_exec_time,
			mean_exec_time,
			max_exec_time,
			stddev_exec_time,
			rows
		FROM pg_stat_statements
		WHERE mean_exec_time > $1
		ORDER BY mean_exec_time DESC
		LIMIT 20
	`

	var results []map[string]any
	rows, err := db.WithContext(ctx).Raw(query, minDuration.Milliseconds()).Rows()
	if err != nil {
		logger.Error("查询慢查询失败", zap.Error(err))
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
			logger.Error("扫描慢查询数据失败", zap.Error(err))
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

// EnableSlowQueryLog 启用慢查询日志
func EnableSlowQueryLog(db *gorm.DB) error {
	logger := appLogger.GetLogger()
	ctx := context.Background()

	// 创建 pg_stat_statements 扩展
	if err := db.WithContext(ctx).Exec("CREATE EXTENSION IF NOT EXISTS pg_stat_statements").Error; err != nil {
		logger.Error("创建 pg_stat_statements 扩展失败", zap.Error(err))
		return err
	}

	logger.Info("慢查询日志已启用")
	return nil
}
