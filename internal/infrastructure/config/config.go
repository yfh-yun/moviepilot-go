// Package config MoviePilot配置管理模块
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	// Config 全局配置实例
	Config *viper.Viper
)

// Init 初始化配置系统
func Init() error {
	Config = viper.New()

	// 设置配置文件名和路径
	Config.SetConfigName("config")
	Config.SetConfigType("yaml")

	// 添加配置搜索路径
	configPaths := []string{
		".",
		"./configs",
		"/etc/moviepilot",
		"$HOME/.moviepilot",
	}

	for _, path := range configPaths {
		if expandedPath := os.ExpandEnv(path); expandedPath != "" {
			Config.AddConfigPath(expandedPath)
		}
	}

	// 设置环境变量前缀
	Config.SetEnvPrefix("MOVIEPILOT")
	Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	Config.AutomaticEnv()

	// 设置默认配置值
	setDefaults()

	// 读取配置文件
	if err := Config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置
			fmt.Println("Config file not found, using default configuration")
		} else {
			return errors.Wrap(err, "failed to read config file")
		}
	}

	// 验证配置
	if err := validateConfig(); err != nil {
		return errors.Wrap(err, "config validation failed")
	}

	return nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 应用配置
	Config.SetDefault("app.name", "MoviePilot")
	Config.SetDefault("app.version", "2.8.1")
	Config.SetDefault("app.env", "development")
	Config.SetDefault("app.debug", true)
	Config.SetDefault("app.timezone", "Asia/Shanghai")

	// 服务器配置
	Config.SetDefault("server.host", "0.0.0.0")
	Config.SetDefault("server.port", 3000)
	Config.SetDefault("server.read_timeout", 30)
	Config.SetDefault("server.write_timeout", 30)
	Config.SetDefault("server.idle_timeout", 60)
	Config.SetDefault("server.max_header_bytes", 1048576)

	// 数据库配置
	Config.SetDefault("database.driver", "sqlite")
	Config.SetDefault("database.host", "localhost")
	Config.SetDefault("database.port", 5432)
	Config.SetDefault("database.name", "moviepilot.db")
	Config.SetDefault("database.username", "moviepilot")
	Config.SetDefault("database.password", "")
	Config.SetDefault("database.ssl_mode", "disable")
	Config.SetDefault("database.max_open_conns", 100)
	Config.SetDefault("database.max_idle_conns", 10)
	Config.SetDefault("database.conn_max_lifetime", 3600)

	// Redis配置
	Config.SetDefault("redis.host", "localhost")
	Config.SetDefault("redis.port", 6379)
	Config.SetDefault("redis.password", "")
	Config.SetDefault("redis.db", 0)
	Config.SetDefault("redis.pool_size", 10)
	Config.SetDefault("redis.min_idle_conns", 3)

	// JWT配置
	Config.SetDefault("jwt.secret", "moviepilot-secret-key")
	Config.SetDefault("jwt.expire_minutes", 1440)
	Config.SetDefault("jwt.refresh_expire_days", 7)

	// 日志配置
	Config.SetDefault("log.level", "info")
	Config.SetDefault("log.format", "json")
	Config.SetDefault("log.output", "stdout")
	Config.SetDefault("log.file_path", "/var/log/moviepilot/app.log")
	Config.SetDefault("log.max_size", 100)
	Config.SetDefault("log.max_backups", 3)
	Config.SetDefault("log.max_age", 28)
	Config.SetDefault("log.compress", true)

	// 安全配置
	Config.SetDefault("security.cors.enabled", true)
	Config.SetDefault("security.cors.allowed_origins", []string{"*"})
	Config.SetDefault("security.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	Config.SetDefault("security.cors.allowed_headers", []string{"*"})
	Config.SetDefault("security.rate_limit.enabled", true)
	Config.SetDefault("security.rate_limit.requests_per_minute", 100)

	// 缓存配置
	Config.SetDefault("cache.type", "memory")
	Config.SetDefault("cache.ttl", 3600)
	Config.SetDefault("cache.prefix", "moviepilot:")
	Config.SetDefault("cache.redis.addr", "localhost:6379")
	Config.SetDefault("cache.redis.password", "")
	Config.SetDefault("cache.redis.db", 0)
	Config.SetDefault("cache.memory.life_window", 3600)
	Config.SetDefault("cache.memory.max_entries", 10000)

	// 下载器配置
	Config.SetDefault("downloader.qbittorrent.enabled", true)
	Config.SetDefault("downloader.qbittorrent.url", "http://localhost:8080")
	Config.SetDefault("downloader.qbittorrent.username", "admin")
	Config.SetDefault("downloader.qbittorrent.password", "adminadmin")
	Config.SetDefault("downloader.transmission.enabled", false)
	Config.SetDefault("downloader.transmission.url", "http://localhost:9091/transmission/rpc")
	Config.SetDefault("downloader.transmission.username", "")
	Config.SetDefault("downloader.transmission.password", "")

	// 媒体服务器配置
	Config.SetDefault("mediaserver.type", "emby")
	Config.SetDefault("mediaserver.emby.url", "http://localhost:8096")
	Config.SetDefault("mediaserver.emby.api_key", "")
	Config.SetDefault("mediaserver.jellyfin.url", "http://localhost:8097")
	Config.SetDefault("mediaserver.jellyfin.api_key", "")
	Config.SetDefault("mediaserver.plex.url", "http://localhost:32400")
	Config.SetDefault("mediaserver.plex.token", "")

	// API配置
	Config.SetDefault("api.tmdb.api_key", "")
	Config.SetDefault("api.tmdb.language", "zh-CN")
	Config.SetDefault("api.tmdb.region", "CN")
	Config.SetDefault("api.douban.api_key", "")
	Config.SetDefault("api.bangumi.enabled", false)
}

// validateConfig 验证配置
func validateConfig() error {
	// 验证必需的配置项
	requiredFields := []string{
		"database.host",
		"database.name",
		"database.username",
		"jwt.secret",
	}

	for _, field := range requiredFields {
		if Config.GetString(field) == "" {
			return fmt.Errorf("required config field '%s' is empty", field)
		}
	}

	// 验证端口号
	port := Config.GetInt("server.port")
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid server port: %d", port)
	}

	// 验证日志级别
	logLevel := Config.GetString("log.level")
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if logLevel == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		return fmt.Errorf("invalid log level: %s", logLevel)
	}

	return nil
}

// GetDatabaseDSN 获取数据库连接字符串
func GetDatabaseDSN() string {
	driver := Config.GetString("database.driver")
	host := Config.GetString("database.host")
	port := Config.GetInt("database.port")
	name := Config.GetString("database.name")
	username := Config.GetString("database.username")
	password := Config.GetString("database.password")
	sslMode := Config.GetString("database.ssl_mode")

	switch driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
			host, port, username, password, name, sslMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			username, password, host, port, name)
	case "sqlite":
		return name
	default:
		return name
	}
}

// GetRedisAddr 获取Redis连接地址
func GetRedisAddr() string {
	host := Config.GetString("redis.host")
	port := Config.GetInt("redis.port")
	return fmt.Sprintf("%s:%d", host, port)
}

// GetJWTSecret 获取JWT密钥
func GetJWTSecret() string {
	return Config.GetString("jwt.secret")
}

// GetJWTExpireMinutes 获取JWT过期时间（分钟）
func GetJWTExpireMinutes() int {
	return Config.GetInt("jwt.expire_minutes")
}

// IsDebug 是否为调试模式
func IsDebug() bool {
	return Config.GetBool("app.debug")
}

// GetLogLevel 获取日志级别
func GetLogLevel() string {
	return Config.GetString("log.level")
}

// GetLogFormat 获取日志格式
func GetLogFormat() string {
	return Config.GetString("log.format")
}

// GetLogOutput 获取日志输出
func GetLogOutput() string {
	return Config.GetString("log.output")
}

// GetGoVersion 获取Go版本信息
func GetGoVersion() string {
	return "go1.21+"
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	if Config.ConfigFileUsed() != "" {
		absPath, _ := filepath.Abs(Config.ConfigFileUsed())
		return absPath
	}
	return ""
}

// WatchConfig 监听配置文件变化
func WatchConfig(callback func()) {
	Config.WatchConfig()
	Config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
		if callback != nil {
			callback()
		}
	})
}
