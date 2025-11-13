package postgresql

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"moviepilot-go/internal/config"
)

// PostgreSQL PostgreSQL数据库连接和操作�?type PostgreSQL struct {
	dbType     string
	dbHost     string
	dbPort     string
	dbName     string
	dbUsername string
	dbPassword string
}

// NewPostgreSQL 创建PostgreSQL实例
func NewPostgreSQL() *PostgreSQL {
	// 从环境变量获取PostgreSQL配置信息
	host := getEnvOrDefault("DB_POSTGRESQL_HOST", "localhost")
	port := getEnvOrDefault("DB_POSTGRESQL_PORT", "5432")
	database := getEnvOrDefault("DB_POSTGRESQL_DATABASE", "moviepilot")
	username := getEnvOrDefault("DB_POSTGRESQL_USERNAME", "moviepilot")
	password := getEnvOrDefault("DB_POSTGRESQL_PASSWORD", "moviepilot")
	
	appConfig := config.GetConfig()
	return &PostgreSQL{
		dbType:     appConfig.DBType,
		dbHost:     host,
		dbPort:     port,
		dbName:     database,
		dbUsername: username,
		dbPassword: password,
	}
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认�?func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// connect 建立PostgreSQL连接
func (p *PostgreSQL) connect() (*sql.DB, error) {
	// 构建PostgreSQL连接字符�?	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		p.dbHost, p.dbPort, p.dbUsername, p.dbPassword, p.dbName)
	
	// 建立数据库连�?	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("创建PostgreSQL连接失败: %v", err)
	}
	
	// 设置连接池参�?	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	
	return db, nil
}

// Test 测试PostgreSQL连接�?func (p *PostgreSQL) Test() (bool, string) {
	// 检查数据库类型是否为PostgreSQL
	if p.dbType != "postgresql" {
		return true, "" // 如果不是PostgreSQL后端，返回成功但不测�?	}
	
	// 测试数据库连�?	db, err := p.connect()
	if err != nil {
		return false, fmt.Sprintf("PostgreSQL连接失败�?v", err)
	}
	defer db.Close()
	
	// 执行简单查询测试连�?	if err := db.Ping(); err != nil {
		return false, fmt.Sprintf("PostgreSQL连接失败�?v", err)
	}
	
	return true, ""
}
