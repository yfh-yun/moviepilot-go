package config

import "time"

// Config 聚合所有配置段
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	Auth       AuthConfig       `mapstructure:"auth"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Storage    StorageConfig    `mapstructure:"storage"`
	TMDB       TMDBConfig       `mapstructure:"tmdb"`
	Workflow   WorkflowConfig   `mapstructure:"workflow"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Plugins    PluginsConfig    `mapstructure:"plugins"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Type            string        `mapstructure:"type"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type AuthConfig struct {
	JWTSecret    string        `mapstructure:"jwt_secret"`
	JWTExpiry    time.Duration `mapstructure:"jwt_expiry"`
	APIKeyHeader string        `mapstructure:"api_key_header"`
	APIKeys      []string      `mapstructure:"api_keys"`
}

type RateLimitConfig struct {
	Enabled              bool `mapstructure:"enabled"`
	RequestsPerMinute    int  `mapstructure:"requests_per_minute"`
	Burst                int  `mapstructure:"burst"`
	APIRequestsPerMinute int  `mapstructure:"api_requests_per_minute"`
	APIBurst             int  `mapstructure:"api_burst"`
}

type StorageConfig struct {
	DownloadPath string `mapstructure:"download_path"`
	LibraryPath  string `mapstructure:"library_path"`
	TempPath     string `mapstructure:"temp_path"`
	MaxFileSize  string `mapstructure:"max_file_size"`
}

type TMDBConfig struct {
	APIKey       string        `mapstructure:"api_key"`
	Language     string        `mapstructure:"language"`
	IncludeAdult bool          `mapstructure:"include_adult"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

type WorkflowConfig struct {
	MaxConcurrent int           `mapstructure:"max_concurrent_workflows"`
	TaskTimeout   time.Duration `mapstructure:"task_timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`
}

type MonitoringConfig struct {
	Enabled     bool              `mapstructure:"enabled"`
	Prometheus  PrometheusConfig  `mapstructure:"prometheus"`
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
}

type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

type HealthCheckConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

type PluginsConfig struct {
	Enabled       bool                `mapstructure:"enabled"`
	PythonService PythonServiceConfig `mapstructure:"python_service"`
	Directories   []string            `mapstructure:"directories"`
}

type PythonServiceConfig struct {
	Host    string        `mapstructure:"host"`
	Port    int           `mapstructure:"port"`
	Timeout time.Duration `mapstructure:"timeout"`
}
