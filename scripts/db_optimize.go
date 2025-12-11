package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"moviepilot-go/pkg/database"
	"moviepilot-go/pkg/logger"
)

func main() {
	// 命令行参数
	action := flag.String("action", "analyze", "操作类型: analyze(分析), create-indexes(创建索引), drop-indexes(删除索引), optimize(优化连接池), vacuum(清理表)")
	host := flag.String("host", "localhost", "数据库主机")
	port := flag.Int("port", 5432, "数据库端口")
	user := flag.String("user", "postgres", "数据库用户")
	password := flag.String("password", "", "数据库密码")
	dbname := flag.String("dbname", "moviepilot", "数据库名称")
	flag.Parse()

	// 初始化日志
	if err := logger.Init(); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	log := logger.GetLogger()

	// 连接数据库
	config := database.Config{
		Host:            *host,
		Port:            *port,
		User:            *user,
		Password:        *password,
		DBName:          *dbname,
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db, err := database.Connect(config, log)
	if err != nil {
		log.Fatal("连接数据库失败", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 执行操作
	switch *action {
	case "analyze":
		analyzeDatabase(db, log)
	case "create-indexes":
		createIndexes(db, log)
	case "drop-indexes":
		dropIndexes(db, log)
	case "optimize":
		optimizeConnectionPool(db, log)
	case "vacuum":
		vacuumTables(db, log)
	case "slow-queries":
		analyzeSlowQueries(db, log)
	default:
		log.Error("未知操作", zap.String("action", *action))
		os.Exit(1)
	}
}

// analyzeDatabase 分析数据库
func analyzeDatabase(db *gorm.DB, log *zap.Logger) {
	log.Info("=== 开始数据库分析 ===")

	// 1. 获取连接池统计
	log.Info("\n--- 连接池统计 ---")
	poolStats, err := database.GetConnectionPoolStats(db)
	if err != nil {
		log.Error("获取连接池统计失败", zap.Error(err))
	} else {
		for key, value := range poolStats {
			log.Info(fmt.Sprintf("%s: %v", key, value))
		}
	}

	// 2. 获取表大小
	log.Info("\n--- 表大小统计 ---")
	tableSizes, err := database.GetTableSizes(db)
	if err != nil {
		log.Error("获取表大小失败", zap.Error(err))
	} else {
		for _, table := range tableSizes {
			log.Info(fmt.Sprintf("表: %s, 总大小: %s, 表大小: %s, 索引大小: %s",
				table["tablename"],
				table["total_size"],
				table["table_size"],
				table["indexes_size"]))
		}
	}

	// 3. 分析索引使用情况
	log.Info("\n--- 索引使用情况 ---")
	indexUsage, err := database.AnalyzeIndexUsage(db)
	if err != nil {
		log.Error("分析索引使用失败", zap.Error(err))
	} else {
		for _, idx := range indexUsage {
			scans := idx["scans"]
			if scans == 0 {
				log.Warn(fmt.Sprintf("未使用的索引: %s.%s (大小: %s)",
					idx["tablename"],
					idx["indexname"],
					idx["size"]))
			}
		}
	}

	log.Info("=== 数据库分析完成 ===")
}

// createIndexes 创建优化索引
func createIndexes(db *gorm.DB, log *zap.Logger) {
	log.Info("=== 开始创建优化索引 ===")

	// 启用慢查询日志
	if err := database.EnableSlowQueryLog(db); err != nil {
		log.Warn("启用慢查询日志失败", zap.Error(err))
	}

	// 创建索引
	if err := database.CreateOptimizedIndexes(db); err != nil {
		log.Error("创建索引失败", zap.Error(err))
		os.Exit(1)
	}

	log.Info("=== 索引创建完成 ===")
}

// dropIndexes 删除优化索引
func dropIndexes(db *gorm.DB, log *zap.Logger) {
	log.Info("=== 开始删除优化索引 ===")

	if err := database.DropOptimizedIndexes(db); err != nil {
		log.Error("删除索引失败", zap.Error(err))
		os.Exit(1)
	}

	log.Info("=== 索引删除完成 ===")
}

// optimizeConnectionPool 优化连接池
func optimizeConnectionPool(db *gorm.DB, log *zap.Logger) {
	log.Info("=== 开始优化连接池 ===")

	// 应用生产环境配置
	config := database.ProductionConfig()
	if err := database.ApplyOptimization(db, config); err != nil {
		log.Error("优化连接池失败", zap.Error(err))
		os.Exit(1)
	}

	// 显示当前统计
	stats, _ := database.GetConnectionPoolStats(db)
	for key, value := range stats {
		log.Info(fmt.Sprintf("%s: %v", key, value))
	}

	log.Info("=== 连接池优化完成 ===")
}

// vacuumTables 清理表
func vacuumTables(db *gorm.DB, log *zap.Logger) {
	log.Info("=== 开始清理表 ===")

	tables := []string{
		"subscribes",
		"downloadhistories",
		"transfer_histories",
		"sites",
		"siteuserdatas",
		"downloads",
		"medias",
	}

	if err := database.VacuumAnalyze(db, tables); err != nil {
		log.Error("清理表失败", zap.Error(err))
		os.Exit(1)
	}

	log.Info("=== 表清理完成 ===")
}

// analyzeSlowQueries 分析慢查询
func analyzeSlowQueries(db *gorm.DB, log *zap.Logger) {
	log.Info("=== 开始分析慢查询 ===")

	// 启用慢查询日志
	if err := database.EnableSlowQueryLog(db); err != nil {
		log.Error("启用慢查询日志失败", zap.Error(err))
		os.Exit(1)
	}

	// 获取慢查询（平均执行时间 > 100ms）
	slowQueries, err := database.GetSlowQueries(db, 100*time.Millisecond)
	if err != nil {
		log.Error("获取慢查询失败", zap.Error(err))
		os.Exit(1)
	}

	if len(slowQueries) == 0 {
		log.Info("未发现慢查询")
		return
	}

	log.Info(fmt.Sprintf("发现 %d 个慢查询:", len(slowQueries)))
	for i, query := range slowQueries {
		log.Info(fmt.Sprintf("\n慢查询 #%d:", i+1))
		log.Info(fmt.Sprintf("  SQL: %s", query["query"]))
		log.Info(fmt.Sprintf("  调用次数: %v", query["calls"]))
		log.Info(fmt.Sprintf("  平均执行时间: %.2f ms", query["mean_exec_time"]))
		log.Info(fmt.Sprintf("  最大执行时间: %.2f ms", query["max_exec_time"]))
	}

	log.Info("=== 慢查询分析完成 ===")
}
