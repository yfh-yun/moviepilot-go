package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// TransferDirectoryConf 文件整理目录配置
type TransferDirectoryConf struct {
	// 名称
	Name string `mapstructure:"name"`
	// 优先级
	Priority int `mapstructure:"priority"`
	// 存储
	Storage string `mapstructure:"storage"`
	// 下载目录
	DownloadPath string `mapstructure:"download_path"`
	// 适用媒体类型
	MediaType string `mapstructure:"media_type"`
	// 适用媒体类别
	MediaCategory string `mapstructure:"media_category"`
	// 下载类型子目录
	DownloadTypeFolder bool `mapstructure:"download_type_folder"`
	// 下载类别子目录
	DownloadCategoryFolder bool `mapstructure:"download_category_folder"`
	// 监控方式 downloader/monitor，None为不监控
	MonitorType string `mapstructure:"monitor_type"`
	// 监控模式 fast / compatibility
	MonitorMode string `mapstructure:"monitor_mode"`
	// 整理方式 move/copy/link/softlink
	TransferType string `mapstructure:"transfer_type"`
	// 文件覆盖模式 always/size/never/latest
	OverwriteMode string `mapstructure:"overwrite_mode"`
	// 整理到媒体库目录
	LibraryPath string `mapstructure:"library_path"`
	// 媒体库目录存储
	LibraryStorage string `mapstructure:"library_storage"`
	// 智能重命名
	Renaming bool `mapstructure:"renaming"`
	// 刮削
	Scraping bool `mapstructure:"scraping"`
	// 是否发送通知
	Notify bool `mapstructure:"notify"`
	// 媒体库类型子目录
	LibraryTypeFolder bool `mapstructure:"library_type_folder"`
	// 媒体库类别子目录
	LibraryCategoryFolder bool `mapstructure:"library_category_folder"`
}

// AppConfig 应用配置结构体
type AppConfig struct {
	// 项目配置
	ProjectName string `mapstructure:"project_name"`
	Version     string `mapstructure:"version"`
	APIVersion  string `mapstructure:"api_version"`
	
	// 服务器配置
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	NginxPort int    `mapstructure:"nginx_port"`
	
	// 数据库配置
	DBType     string `mapstructure:"db_type"`
	DBPath     string `mapstructure:"db_path"`
	
	// 数据库详细配置
	DBTimeout   int  `mapstructure:"db_timeout"`
	DBWalEnable bool `mapstructure:"db_wal_enable"`
	DBPoolType  string `mapstructure:"db_pool_type"`
	DBPoolPrePing bool `mapstructure:"db_pool_pre_ping"`
	DBPoolRecycle int `mapstructure:"db_pool_recycle"`
	DBSQLitePoolSize int `mapstructure:"db_sqlite_pool_size"`
	DBPoolTimeout int `mapstructure:"db_pool_timeout"`
	DBSQLiteMaxOverflow int `mapstructure:"db_sqlite_max_overflow"`
	DBPostgreSQLHost string `mapstructure:"db_postgresql_host"`
	DBPostgreSQLPort string `mapstructure:"db_postgresql_port"`
	DBPostgreSQLDatabase string `mapstructure:"db_postgresql_database"`
	DBPostgreSQLUsername string `mapstructure:"db_postgresql_username"`
	DBPostgreSQLPassword string `mapstructure:"db_postgresql_password"`
	DBPostgreSQLPoolSize int `mapstructure:"db_postgresql_pool_size"`
	DBPostgreSQLMaxOverflow int `mapstructure:"db_postgresql_max_overflow"`
	DBEcho bool `mapstructure:"db_echo"`
	
	// 日志配置
	LogLevel string `mapstructure:"log_level"`
	LogPath  string `mapstructure:"log_path"`
	
	// 媒体配置
	MediaPaths []string `mapstructure:"media_paths"`
	
	// 插件配置
	PluginPath string `mapstructure:"plugin_path"`
	
	// CookieCloud配置
	CookieCloudHost       string `mapstructure:"cookiecloud_host"`
	CookieCloudKey        string `mapstructure:"cookiecloud_key"`
	CookieCloudPassword   string `mapstructure:"cookiecloud_password"`
	CookieCloudEnableLocal bool  `mapstructure:"cookiecloud_enable_local"`
	CookiePath            string `mapstructure:"cookie_path"`
	
	// DoH配置
	DohEnable   bool   `mapstructure:"doh_enable"`
	DohDomains  string `mapstructure:"doh_domains"`
	DohResolvers string `mapstructure:"doh_resolvers"`
	
	// 目录配置
	Directories []TransferDirectoryConf `mapstructure:"directories"`
	
	// 临时文件和缓存配置
	TempPath             string `mapstructure:"temp_path"`
	CachePath            string `mapstructure:"cache_path"`
	TempFileDays         int    `mapstructure:"temp_file_days"`
	GlobalImageCacheDays int    `mapstructure:"global_image_cache_days"`
	
	// 浏览器仿真配置
	BrowserEmulation string `mapstructure:"browser_emulation"`
	FlareSolverrURL  string `mapstructure:"flaresolverr_url"`
	
	// 开发模式
	Debug bool `mapstructure:"debug"`
	
	// 允许的主机
	AllowedHosts []string `mapstructure:"allowed_hosts"`
}

var (
	config *AppConfig
)

// GetConfig 获取应用配置实例（单例模式）
func GetConfig() *AppConfig {
	if config == nil {
		config = NewConfig()
	}
	return config
}

// NewConfig 创建默认配置
func NewConfig() *AppConfig {
	// 获取当前工作目录
	wd, _ := os.Getwd()
	
	// 根据CPU核心数计算worker数量，类似Python版本中的 multiprocessing.cpu_count() * 2 + 1
	numWorkers := runtime.NumCPU()*2 + 1
	
	return &AppConfig{
		ProjectName:            "MoviePilot-Go",
		Version:                "1.0.0",
		APIVersion:             "1.0.0",
		Host:                   "0.0.0.0",
		Port:                   3001,
		NginxPort:              3000,
		DBType:                 "sqlite",
		DBPath:                 filepath.Join(wd, "data", "moviepilot.db"),
		DBTimeout:              30,
		DBWalEnable:            true,
		DBPoolType:             "QueuePool",
		DBPoolPrePing:          true,
		DBPoolRecycle:          3600,
		DBSQLitePoolSize:       30,
		DBPoolTimeout:          30,
		DBSQLiteMaxOverflow:    10,
		DBPostgreSQLHost:       "localhost",
		DBPostgreSQLPort:       "5432",
		DBPostgreSQLDatabase:   "moviepilot",
		DBPostgreSQLUsername:   "moviepilot",
		DBPostgreSQLPassword:   "moviepilot",
		DBPostgreSQLPoolSize:   20,
		DBPostgreSQLMaxOverflow: 30,
		DBEcho:                 false,
		LogLevel:               "info",
		LogPath:                filepath.Join(wd, "logs"),
		MediaPaths:             []string{filepath.Join(wd, "media")},
		PluginPath:             filepath.Join(wd, "plugins"),
		CookieCloudHost:        "",
		CookieCloudKey:         "",
		CookieCloudPassword:    "",
		CookieCloudEnableLocal: false,
		CookiePath:             filepath.Join(wd, "data"),
		DohEnable:              false,
		DohDomains:             "",
		DohResolvers:           "cloudflare-dns.com,dns.google,dns.quad9.net",
		Directories:            []TransferDirectoryConf{},
		TempPath:               filepath.Join(wd, "temp"),
		CachePath:              filepath.Join(wd, "cache"),
		TempFileDays:           3,
		GlobalImageCacheDays:   7,
		BrowserEmulation:       "",
		FlareSolverrURL:        "",
		Debug:                  false,
		AllowedHosts:           []string{"*"}, // 允许所有主机访问
	}
}

// LoadConfig 从配置文件加载配置
func (c *AppConfig) LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}
	
	return nil
}

// Validate 验证配置
func (c *AppConfig) Validate() error {
	// TODO: 实现配置验证逻辑
	fmt.Println("验证配置...")
	return nil
}

// GetRootPath 获取项目根路径
func (c *AppConfig) GetRootPath() string {
	// 获取当前工作目录作为根路径
	wd, _ := os.Getwd()
	return wd
}

// GetConfigPath 获取配置路径
func (c *AppConfig) GetConfigPath() string {
	return filepath.Join(c.GetRootPath(), "config")
}

// GetTempPath 获取临时文件路径
func (c *AppConfig) GetTempPath() string {
	return c.TempPath
}

// GetCachePath 获取缓存路径
func (c *AppConfig) GetCachePath() string {
	return c.CachePath
}