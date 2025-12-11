package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"moviepilot-go/internal/infrastructure/config/models"
	"moviepilot-go/pkg/logger"
)

// Loader 配置加载器
type Loader struct {
	logger *zap.Logger
}

// NewLoader 创建配置加载器实例
func NewLoader() *Loader {
	return &Loader{
		logger: logger.GetLogger(),
	}
}

// Load 加载配置（优先级：环境变量 > .env 文件 > 默认值）
func (l *Loader) Load() (*Config, error) {
	// 1. 加载 .env 文件
	envPath := l.getEnvPath()
	if err := godotenv.Load(envPath); err != nil {
		l.logger.Warn("failed to load .env file", zap.String("path", envPath), zap.Error(err))
	}

	// 2. 初始化 viper
	v := viper.New()
	v.AutomaticEnv() // 自动读取环境变量

	// 3. 设置默认值
	l.setDefaults(v)

	// 4. 解析配置
	cfg := &Config{}
	if err := l.unmarshal(v, cfg); err != nil {
		return nil, err
	}

	// 5. 校验配置
	if err := l.validate(cfg); err != nil {
		return nil, err
	}

	// 6. 初始化动态属性
	cfg.initPaths()

	l.logger.Info("configuration loaded successfully")
	return cfg, nil
}

// getEnvPath 获取 .env 文件路径
func (l *Loader) getEnvPath() string {
	// 优先级：环境变量 > Docker > 二进制 > 开发环境
	if envPath := os.Getenv("ENV_FILE_PATH"); envPath != "" {
		return envPath
	}

	if isDocker() {
		return "/config/app.env"
	}

	if isFrozen() {
		execPath, _ := os.Executable()
		return fmt.Sprintf("%s/config/app.env", execPath)
	}

	return "./config/app.env"
}

// setDefaults 设置默认配置值
func (l *Loader) setDefaults(v *viper.Viper) {
	// 应用配置
	v.SetDefault("PROJECT_NAME", "MoviePilot")
	v.SetDefault("API_V1_STR", "/api/v1")
	v.SetDefault("FRONTEND_PATH", "/public")
	v.SetDefault("TZ", "Asia/Shanghai")
	v.SetDefault("HOST", "0.0.0.0")
	v.SetDefault("PORT", 3001)
	v.SetDefault("NGINX_PORT", 3000)
	v.SetDefault("DEBUG", false)
	v.SetDefault("DEV", false)
	v.SetDefault("ADVANCED_MODE", true)

	// 数据库配置
	v.SetDefault("DB_TYPE", "sqlite")
	v.SetDefault("DB_ECHO", false)
	v.SetDefault("DB_TIMEOUT", 60)
	v.SetDefault("DB_WAL_ENABLE", true)
	v.SetDefault("DB_POOL_TYPE", "QueuePool")
	v.SetDefault("DB_POOL_PRE_PING", true)
	v.SetDefault("DB_POOL_RECYCLE", 300)
	v.SetDefault("DB_POOL_TIMEOUT", 30)
	v.SetDefault("DB_SQLITE_POOL_SIZE", 10)
	v.SetDefault("DB_SQLITE_MAX_OVERFLOW", 50)
	v.SetDefault("DB_POSTGRESQL_HOST", "localhost")
	v.SetDefault("DB_POSTGRESQL_PORT", 5432)
	v.SetDefault("DB_POSTGRESQL_DATABASE", "moviepilot")
	v.SetDefault("DB_POSTGRESQL_USERNAME", "moviepilot")
	v.SetDefault("DB_POSTGRESQL_PASSWORD", "moviepilot")
	v.SetDefault("DB_POSTGRESQL_POOL_SIZE", 10)
	v.SetDefault("DB_POSTGRESQL_MAX_OVERFLOW", 50)

	// 缓存配置
	v.SetDefault("CACHE_BACKEND_TYPE", "cachetools")
	v.SetDefault("CACHE_BACKEND_URL", "redis://localhost:6379")
	v.SetDefault("GLOBAL_IMAGE_CACHE", false)
	v.SetDefault("GLOBAL_IMAGE_CACHE_DAYS", 7)
	v.SetDefault("TEMP_FILE_DAYS", 3)
	v.SetDefault("META_CACHE_EXPIRE", 0)

	// 安全配置
	v.SetDefault("ALLOWED_HOSTS", []string{"*"})
	v.SetDefault("ACCESS_TOKEN_EXPIRE_MINUTES", 11520)
	v.SetDefault("RESOURCE_ACCESS_TOKEN_EXPIRE_SECONDS", 1800)
	v.SetDefault("SUPERUSER", "admin")
	v.SetDefault("AUXILIARY_AUTH_ENABLE", false)

	// 媒体配置
	v.SetDefault("SEARCH_SOURCE", "tmdb")
	v.SetDefault("RECOGNIZE_SOURCE", "tmdb")

	// TMDB配置
	v.SetDefault("TMDB_IMAGE_DOMAIN", "image.tmdb.org")
	v.SetDefault("TMDB_LANGUAGE", "zh-CN")
	v.SetDefault("TMDB_REGION", "CN")
	v.SetDefault("TMDB_PROXY_ENABLE", false)

	// 站点配置
	v.SetDefault("SITEDATA_REFRESH_INTERVAL", 3600)
	v.SetDefault("SITE_MAX_CONCURRENT_TASKS", 5)
	v.SetDefault("SITE_REQUEST_TIMEOUT", 30)
	v.SetDefault("SITE_RETRY_TIMES", 3)
	v.SetDefault("SITE_RETRY_INTERVAL", 5)

	// 下载配置
	v.SetDefault("TORRENT_TAG", "MoviePilot")
	v.SetDefault("DOWNLOAD_SUBTITLE", true)
	v.SetDefault("MAX_CONCURRENT_DOWNLOADS", 5)

	// 整理配置
	v.SetDefault("TRANSFER_ENABLE", true)
	v.SetDefault("AUTO_TRANSFER", true)
	v.SetDefault("DELETE_SOURCE", false)

	// 插件配置
	v.SetDefault("PLUGIN_AUTO_RELOAD", false)
	v.SetDefault("PLUGIN_MAX_WORKERS", 10)
	v.SetDefault("PLUGIN_TIMEOUT", 30)

	// CookieCloud配置
	v.SetDefault("COOKIECLOUD_INTERVAL", 3600)
	v.SetDefault("COOKIECLOUD_ENABLE", false)

	// 性能配置
	v.SetDefault("BIG_MEMORY_MODE", false)
	v.SetDefault("ENCODING_DETECTION_PERFORMANCE_MODE", true)
	v.SetDefault("ENCODING_DETECTION_MIN_CONFIDENCE", 0.8)
	v.SetDefault("MEMORY_GC_INTERVAL", 30)

	// 网络配置
	v.SetDefault("DOH_ENABLE", false)

	// 调度器配置
	v.SetDefault("SCHEDULER_MAX_CONCURRENT_TASKS", 10)
	v.SetDefault("SCHEDULER_TASK_TIMEOUT", 3600)

	// 订阅配置
	v.SetDefault("SUBSCRIBE_CHECK_INTERVAL", 300)
	v.SetDefault("SUBSCRIBE_MAX_CONCURRENT_CHECKS", 5)
	v.SetDefault("SUBSCRIBE_NOTIFICATION_ENABLED", true)
}

// unmarshal 解析配置到结构体
func (l *Loader) unmarshal(v *viper.Viper, cfg *Config) error {
	// 初始化各配置结构体
	cfg.App = &models.AppConfig{}
	cfg.Database = &models.DatabaseConfig{}
	cfg.Cache = &models.CacheConfig{}
	cfg.Security = &models.SecurityConfig{}
	cfg.Media = &models.MediaConfig{}
	cfg.TMDB = &models.TMDBConfig{}
	cfg.Site = &models.SiteConfig{}
	cfg.Download = &models.DownloadConfig{}
	cfg.CookieCloud = &models.CookieCloudConfig{}
	cfg.Transfer = &models.TransferConfig{}
	cfg.Plugin = &models.PluginConfig{}
	cfg.Performance = &models.PerformanceConfig{}
	cfg.Scheduler = &models.SchedulerConfig{}
	cfg.Subscribe = &models.SubscribeConfig{}
	cfg.Network = &models.NetworkConfig{}

	// 解析各配置结构体
	if err := v.Unmarshal(&cfg.App); err != nil {
		return fmt.Errorf("failed to unmarshal app config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Database); err != nil {
		return fmt.Errorf("failed to unmarshal database config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Cache); err != nil {
		return fmt.Errorf("failed to unmarshal cache config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Security); err != nil {
		return fmt.Errorf("failed to unmarshal security config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Media); err != nil {
		return fmt.Errorf("failed to unmarshal media config: %w", err)
	}

	if err := v.Unmarshal(&cfg.TMDB); err != nil {
		return fmt.Errorf("failed to unmarshal tmdb config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Site); err != nil {
		return fmt.Errorf("failed to unmarshal site config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Download); err != nil {
		return fmt.Errorf("failed to unmarshal download config: %w", err)
	}

	if err := v.Unmarshal(&cfg.CookieCloud); err != nil {
		return fmt.Errorf("failed to unmarshal cookiecloud config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Transfer); err != nil {
		return fmt.Errorf("failed to unmarshal transfer config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Plugin); err != nil {
		return fmt.Errorf("failed to unmarshal plugin config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Performance); err != nil {
		return fmt.Errorf("failed to unmarshal performance config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Scheduler); err != nil {
		return fmt.Errorf("failed to unmarshal scheduler config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Subscribe); err != nil {
		return fmt.Errorf("failed to unmarshal subscribe config: %w", err)
	}

	if err := v.Unmarshal(&cfg.Network); err != nil {
		return fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	return nil
}

// validate 校验配置
func (l *Loader) validate(cfg *Config) error {
	// 这里可以添加配置校验逻辑
	// 例如：检查必填项、校验格式等
	return nil
}
