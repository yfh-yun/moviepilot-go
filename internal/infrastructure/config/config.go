package config

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"moviepilot-go/internal/infrastructure/config/models"
	"moviepilot-go/pkg/cache"
)

// PathConfig 路径配置（动态计算）
type PathConfig struct {
	Root       string
	Config     string
	Temp       string
	Cache      string
	PluginData string
	Log        string
	Cookie     string
}

// Config 全局配置结构体
type Config struct {
	mu sync.RWMutex // 读写锁，支持运行时更新

	// 各分类配置（组合模式）
	App         *models.AppConfig
	Database    *models.DatabaseConfig
	Cache       *models.CacheConfig
	Security    *models.SecurityConfig
	Media       *models.MediaConfig
	TMDB        *models.TMDBConfig
	Site        *models.SiteConfig
	Download    *models.DownloadConfig
	CookieCloud *models.CookieCloudConfig
	Transfer    *models.TransferConfig
	Plugin      *models.PluginConfig
	Performance *models.PerformanceConfig
	Scheduler   *models.SchedulerConfig
	Subscribe   *models.SubscribeConfig
	Network     *models.NetworkConfig

	// 动态计算属性（通过方法实现）
	paths *PathConfig // 内部字段
}

// GetCacheBackendConfig 将缓存相关配置映射为 pkg/cache.Config
// 供上层初始化统一的缓存 Backend 使用
func (c *Config) GetCacheBackendConfig() cache.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()

	backendType := cache.BackendMemory
	switch c.Cache.BackendType {
	case "redis":
		backendType = cache.BackendRedis
	case "file":
		backendType = cache.BackendFile
	default:
		// "cachetools" 或其他值都统一视为内存缓存
		backendType = cache.BackendMemory
	}

	// 统一的默认TTL：优先使用 MetaCacheExpire（单位小时），0 表示永不过期
	var defaultTTL time.Duration
	if c.Cache.MetaCacheExpire > 0 {
		defaultTTL = time.Duration(c.Cache.MetaCacheExpire) * time.Hour
	}

	return cache.Config{
		Type:       backendType,
		DefaultTTL: defaultTTL,
		// File 缓存使用全局 Cache 目录
		FileBaseDir: c.paths.Cache,
		// Redis 缓存使用统一的 BackendURL
		RedisURL: c.Cache.BackendURL,
	}
}

// initPaths 初始化动态路径配置
func (c *Config) initPaths() {
	c.paths = &PathConfig{}

	// Root Path
	c.paths.Root = c.getRootPath()

	// Config Path
	if c.App.ConfigDir != "" {
		c.paths.Config = c.App.ConfigDir
	} else if isDocker() {
		c.paths.Config = "/config"
	} else if isFrozen() {
		execPath, _ := os.Executable()
		c.paths.Config = filepath.Join(filepath.Dir(execPath), "config")
	} else {
		c.paths.Config = filepath.Join(c.paths.Root, "config")
	}

	// 其他路径
	c.paths.Temp = filepath.Join(c.paths.Config, "temp")
	c.paths.Cache = filepath.Join(c.paths.Config, "cache")
	c.paths.PluginData = filepath.Join(c.paths.Config, "plugins")
	c.paths.Log = filepath.Join(c.paths.Config, "logs")
	c.paths.Cookie = filepath.Join(c.paths.Config, "cookies")

	// 确保目录存在
	c.ensureDirsExist()
}

// GetConfigPath 获取配置目录路径
func (c *Config) GetConfigPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.paths.Config
}

// GetTempPath 获取临时文件路径
func (c *Config) GetTempPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.paths.Temp
}

// GetCachePath 获取缓存目录路径
func (c *Config) GetCachePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.paths.Cache
}

// GetPluginDataPath 获取插件数据目录路径
func (c *Config) GetPluginDataPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.paths.PluginData
}

// GetLogPath 获取日志目录路径
func (c *Config) GetLogPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.paths.Log
}

// GetCookiePath 获取Cookie目录路径
func (c *Config) GetCookiePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.paths.Cookie
}

// GetSystemConf 根据性能模式返回系统配置
func (c *Config) GetSystemConf() *models.SystemConfModel {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.Performance.BigMemoryMode {
		metaExpire := c.Cache.MetaCacheExpire
		if metaExpire == 0 {
			metaExpire = 72
		}
		return &models.SystemConfModel{
			Torrents:   200,
			Refresh:    100,
			TMDB:       1024,
			Douban:     512,
			Bangumi:    512,
			Fanart:     512,
			Meta:       metaExpire * 3600,
			Scheduler:  100,
			ThreadPool: 100,
		}
	}

	metaExpire := c.Cache.MetaCacheExpire
	if metaExpire == 0 {
		metaExpire = 24
	}
	return &models.SystemConfModel{
		Torrents:   100,
		Refresh:    50,
		TMDB:       256,
		Douban:     256,
		Bangumi:    256,
		Fanart:     128,
		Meta:       metaExpire * 3600,
		Scheduler:  50,
		ThreadPool: 50,
	}
}

// GetProxy 获取代理配置
func (c *Config) GetProxy() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.Network.ProxyHost != "" {
		return map[string]string{
			"http":  c.Network.ProxyHost,
			"https": c.Network.ProxyHost,
		}
	}
	return nil
}

// getRootPath 获取应用根路径
func (c *Config) getRootPath() string {
	if execPath, err := os.Executable(); err == nil {
		return filepath.Dir(execPath)
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

// ensureDirsExist 确保必要的目录存在
func (c *Config) ensureDirsExist() {
	dirs := []string{
		c.paths.Config,
		c.paths.Temp,
		c.paths.Cache,
		c.paths.PluginData,
		c.paths.Log,
		c.paths.Cookie,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			// 这里可以添加日志记录
		}
	}
}

// isDocker 检查是否在Docker容器中运行
func isDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/proc/self/cgroup"); err == nil {
		content, _ := os.ReadFile("/proc/self/cgroup")
		return len(content) > 0 && (contains(string(content), "docker") || contains(string(content), "containerd"))
	}
	return false
}

// isFrozen 检查是否为编译后的二进制文件
func isFrozen() bool {
	// 简单判断：如果可执行文件路径包含"/bin/"或"\bin\"，则认为是编译后的二进制
	execPath, err := os.Executable()
	if err != nil {
		return false
	}
	return contains(execPath, "/bin/") || contains(execPath, "\\bin\\")
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
