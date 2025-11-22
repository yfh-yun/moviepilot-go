package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// Options 控制配置加载行为
type Options struct {
	ConfigFile  string
	ConfigPaths []string
}

// Manager 负责加载和管理配置
type Manager struct {
	mu  sync.RWMutex
	cfg Config
	v   *viper.Viper
	log *zap.Logger
}

// NewManager 创建配置管理器并立即加载配置
func NewManager(opts Options) (*Manager, error) {
	log := logger.GetLogger()
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("MOVIEPILOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if opts.ConfigFile != "" {
		v.SetConfigFile(opts.ConfigFile)
	} else {
		v.SetConfigName("core")
		paths := opts.ConfigPaths
		if len(paths) == 0 {
			paths = []string{"./configs", "./config", "."}
		}
		for _, p := range paths {
			v.AddConfigPath(p)
		}
	}

	if err := readConfig(v, log); err != nil {
		return nil, err
	}

	mgr := &Manager{v: v, log: log}
	if err := mgr.reload(); err != nil {
		return nil, err
	}

	return mgr, nil
}

// Get 返回当前配置的副本
func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

// reload 从 viper 反序列化配置
func (m *Manager) reload() error {
	var cfg Config
	if err := m.v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	return nil
}

func readConfig(v *viper.Viper, log *zap.Logger) error {
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn("Config file not found, using defaults and env overrides")
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	log.Info("Configuration loaded",
		zap.String("config_file", v.ConfigFileUsed()),
	)
	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "moviepilot-go")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 3001)
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("server.shutdown_timeout", "30s")

	v.SetDefault("database.type", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "moviepilot")
	v.SetDefault("database.user", "moviepilot")
	v.SetDefault("database.password", "moviepilot")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "300s")

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 3)

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.output", "stdout")
	v.SetDefault("logger.file_path", filepath.Join(os.TempDir(), "moviepilot", "app.log"))
	v.SetDefault("logger.max_size", 100)
	v.SetDefault("logger.max_backups", 3)
	v.SetDefault("logger.max_age", 28)
	v.SetDefault("logger.compress", true)

	v.SetDefault("auth.jwt_secret", "your-secret-key-change-in-production")
	v.SetDefault("auth.jwt_expiry", "24h")
	v.SetDefault("auth.api_key_header", "X-API-Key")
	v.SetDefault("auth.api_keys", []string{"development-api-key"})

	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.requests_per_minute", 60)
	v.SetDefault("rate_limit.burst", 10)
	v.SetDefault("rate_limit.api_requests_per_minute", 120)
	v.SetDefault("rate_limit.api_burst", 20)

	v.SetDefault("storage.download_path", "/downloads")
	v.SetDefault("storage.library_path", "/library")
	v.SetDefault("storage.temp_path", "/tmp")
	v.SetDefault("storage.max_file_size", "50GB")

	v.SetDefault("tmdb.api_key", "")
	v.SetDefault("tmdb.language", "zh-CN")
	v.SetDefault("tmdb.include_adult", false)
	v.SetDefault("tmdb.timeout", "30s")

	v.SetDefault("workflow.max_concurrent_workflows", 5)
	v.SetDefault("workflow.task_timeout", "300s")
	v.SetDefault("workflow.retry_attempts", 3)
	v.SetDefault("workflow.retry_delay", "5s")

	v.SetDefault("monitoring.enabled", true)
	v.SetDefault("monitoring.prometheus.enabled", true)
	v.SetDefault("monitoring.prometheus.port", 9090)
	v.SetDefault("monitoring.prometheus.path", "/metrics")
	v.SetDefault("monitoring.health_check.enabled", true)
	v.SetDefault("monitoring.health_check.port", 3002)
	v.SetDefault("monitoring.health_check.path", "/health")

	v.SetDefault("plugins.enabled", true)
	v.SetDefault("plugins.python_service.host", "localhost")
	v.SetDefault("plugins.python_service.port", 5000)
	v.SetDefault("plugins.python_service.timeout", "30s")
	v.SetDefault("plugins.directories", []string{"./plugins/go", "./plugins/python"})
}
