package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config *SystemConfig
	logger *logger.Logger
}

// SystemConfig 系统配置结构
type SystemConfig struct {
	// 基础配置
	AppName      string `json:"app_name"`
	AppVersion   string `json:"app_version"`
	Debug        bool   `json:"debug"`
	LogLevel     string `json:"log_level"`
	LogPath      string `json:"log_path"`
	LogRetention int    `json:"log_retention"`
	WorkDir      string `json:"work_dir"`

	// API配置
	APIHost                          string `json:"api_host"`
	APIPort                          int    `json:"api_port"`
	APIKey                           string `json:"api_key"`
	APIToken                         string `json:"api_token"`
	ResourceSecretKey                string `json:"resource_secret_key"`
	AccessTokenExpireMinutes         int    `json:"access_token_expire_minutes"`
	ResourceAccessTokenExpireSeconds int    `json:"resource_access_token_expire_seconds"`

	// 数据库配置
	DBType            string        `json:"db_type"`
	DBHost            string        `json:"db_host"`
	DBPort            int           `json:"db_port"`
	DBUser            string        `json:"db_user"`
	DBPassword        string        `json:"db_password"`
	DBName            string        `json:"db_name"`
	DBMaxIdleConns    int           `json:"db_max_idle_conns"`
	DBMaxOpenConns    int           `json:"db_max_open_conns"`
	DBConnMaxLifetime time.Duration `json:"db_conn_max_lifetime"`

	// 缓存配置
	RedisHost     string        `json:"redis_host"`
	RedisPort     int           `json:"redis_port"`
	RedisPassword string        `json:"redis_password"`
	RedisDB       int           `json:"redis_db"`
	CacheExpire   time.Duration `json:"cache_expire"`

	// 网络配置
	ProxyURL      string        `json:"proxy_url"`
	ProxyUsername string        `json:"proxy_username"`
	ProxyPassword string        `json:"proxy_password"`
	Timeout       time.Duration `json:"timeout"`
	RetryCount    int           `json:"retry_count"`

	// 媒体配置
	MediaExtensions    []string `json:"media_extensions"`
	SubtitleExtensions []string `json:"subtitle_extensions"`
	RenameRule         string   `json:"rename_rule"`
	DeleteSource       bool     `json:"delete_source"`

	// 安全配置
	SecretKey   string   `json:"secret_key"`
	Salt        string   `json:"salt"`
	CorsOrigins []string `json:"cors_origins"`

	// 插件配置
	PluginDir     string `json:"plugin_dir"`
	PluginAPIHost string `json:"plugin_api_host"`
	PluginAPIPort int    `json:"plugin_api_port"`

	// 工作流配置
	WorkflowMaxConcurrent int           `json:"workflow_max_concurrent"`
	WorkflowTimeout       time.Duration `json:"workflow_timeout"`
}

// NewConfigManager 创建配置管理器
func NewConfigManager(log *logger.Logger) *ConfigManager {
	return &ConfigManager{
		logger: log,
		config: DefaultConfig(),
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *SystemConfig {
	return &SystemConfig{
		// 基础配置
		AppName:      "MoviePilot",
		AppVersion:   "2.8.1",
		Debug:        false,
		LogLevel:     "info",
		LogPath:      "./logs",
		LogRetention: 7,
		WorkDir:      ".",

		// API配置
		APIHost:                          "0.0.0.0",
		APIPort:                          3001,
		APIKey:                           "",
		APIToken:                         "",
		ResourceSecretKey:                "",
		AccessTokenExpireMinutes:         1440,
		ResourceAccessTokenExpireSeconds: 86400,

		// 数据库配置
		DBType:            "postgres",
		DBHost:            "localhost",
		DBPort:            5432,
		DBUser:            "postgres",
		DBPassword:        "postgres",
		DBName:            "moviepilot",
		DBMaxIdleConns:    10,
		DBMaxOpenConns:    100,
		DBConnMaxLifetime: 1 * time.Hour,

		// 缓存配置
		RedisHost:     "localhost",
		RedisPort:     6379,
		RedisPassword: "",
		RedisDB:       0,
		CacheExpire:   30 * time.Minute,

		// 网络配置
		ProxyURL:      "",
		ProxyUsername: "",
		ProxyPassword: "",
		Timeout:       30 * time.Second,
		RetryCount:    3,

		// 媒体配置
		MediaExtensions:    []string{"mp4", "mkv", "avi", "mov", "wmv", "flv", "ts", "webm", "iso"},
		SubtitleExtensions: []string{"srt", "ass", "ssa", "idx", "sub", "vtt"},
		RenameRule:         "{title} - S{season:02d}E{episode:02d} - {resolution}",
		DeleteSource:       false,

		// 安全配置
		SecretKey:   "moviepilot-secret-key",
		Salt:        "moviepilot-salt",
		CorsOrigins: []string{"*"},

		// 插件配置
		PluginDir:     "./plugins",
		PluginAPIHost: "localhost",
		PluginAPIPort: 5000,

		// 工作流配置
		WorkflowMaxConcurrent: 5,
		WorkflowTimeout:       24 * time.Hour,
	}
}

// Load 从文件加载配置
func (cm *ConfigManager) Load(configPath string) error {
	cm.logger.Info("Loading configuration", "path", configPath)

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cm.logger.Warn("Config file not found, creating default config", "path", configPath)
		return cm.Save(configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	var config SystemConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	cm.config = &config
	cm.logger.Info("Configuration loaded successfully")
	return nil
}

// Save 保存配置到文件
func (cm *ConfigManager) Save(configPath string) error {
	cm.logger.Info("Saving configuration", "path", configPath)

	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cm.logger.Info("Configuration saved successfully")
	return nil
}

// Get 获取配置
func (cm *ConfigManager) Get() *SystemConfig {
	return cm.config
}

// Set 设置配置
func (cm *ConfigManager) Set(config *SystemConfig) {
	cm.config = config
}

// Update 更新配置项
func (cm *ConfigManager) Update(key string, value interface{}) error {
	// 这里可以实现动态更新配置项的逻辑
	// 例如使用反射来设置字段值
	cm.logger.Info("Updating config", "key", key)
	return nil
}

// Validate 验证配置
func (cm *ConfigManager) Validate() error {
	// 验证必要的配置项
	if cm.config.SecretKey == "" {
		return fmt.Errorf("secret_key cannot be empty")
	}

	if cm.config.WorkDir == "" {
		cm.config.WorkDir = "."
	}

	// 验证日志路径
	if err := os.MkdirAll(cm.config.LogPath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	cm.logger.Info("Configuration validation passed")
	return nil
}

// Reload 重新加载配置
func (cm *ConfigManager) Reload(configPath string) error {
	return cm.Load(configPath)
}

// GetAPIURL 获取API URL
func (cm *ConfigManager) GetAPIURL() string {
	return fmt.Sprintf("http://%s:%d", cm.config.APIHost, cm.config.APIPort)
}

// GetPluginAPIURL 获取插件API URL
func (cm *ConfigManager) GetPluginAPIURL() string {
	return fmt.Sprintf("http://%s:%d", cm.config.PluginAPIHost, cm.config.PluginAPIPort)
}

// GetDatabaseDSN 获取数据库连接字符串
func (cm *ConfigManager) GetDatabaseDSN() string {
	switch cm.config.DBType {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cm.config.DBHost, cm.config.DBPort, cm.config.DBUser, cm.config.DBPassword, cm.config.DBName)
	case "sqlite":
		return filepath.Join(cm.config.WorkDir, "data", "moviepilot.db")
	default:
		return ""
	}
}

// GetRedisAddr 获取Redis地址
func (cm *ConfigManager) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", cm.config.RedisHost, cm.config.RedisPort)
}
