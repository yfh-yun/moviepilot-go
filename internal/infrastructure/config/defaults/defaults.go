package defaults

import (
	"moviepilot-go/internal/infrastructure/config/models"
)

// GetDefaultAppConfig 获取默认应用配置
func GetDefaultAppConfig() *models.AppConfig {
	return &models.AppConfig{
		ProjectName:  "MoviePilot",
		APIV1Str:     "/api/v1",
		FrontendPath: "/public",
		TZ:           "Asia/Shanghai",
		Host:         "0.0.0.0",
		Port:         3001,
		NginxPort:    3000,
		Debug:        false,
		Dev:          false,
		AdvancedMode: true,
	}
}

// GetDefaultDatabaseConfig 获取默认数据库配置
func GetDefaultDatabaseConfig() *models.DatabaseConfig {
	return &models.DatabaseConfig{
		Type:                  "sqlite",
		Echo:                  false,
		Timeout:               60,
		WALEnable:             true,
		PoolType:              "QueuePool",
		PoolPrePing:           true,
		PoolRecycle:           300,
		PoolTimeout:           30,
		SQLitePoolSize:        10,
		SQLiteMaxOverflow:     50,
		PostgreSQLHost:        "localhost",
		PostgreSQLPort:        5432,
		PostgreSQLDatabase:    "moviepilot",
		PostgreSQLUsername:    "moviepilot",
		PostgreSQLPassword:    "moviepilot",
		PostgreSQLPoolSize:    10,
		PostgreSQLMaxOverflow: 50,
	}
}

// GetDefaultCacheConfig 获取默认缓存配置
func GetDefaultCacheConfig() *models.CacheConfig {
	return &models.CacheConfig{
		BackendType:      "cachetools",
		BackendURL:       "redis://localhost:6379",
		GlobalImageCache: false,
		ImageCacheDays:   7,
		TempFileDays:     3,
		MetaCacheExpire:  0,
	}
}

// GetDefaultSecurityConfig 获取默认安全配置
func GetDefaultSecurityConfig() *models.SecurityConfig {
	return &models.SecurityConfig{
		AllowedHosts:                     []string{"*"},
		AccessTokenExpireMinutes:         11520,
		ResourceAccessTokenExpireSeconds: 1800,
		SuperUser:                        "admin",
		AuxiliaryAuthEnable:              false,
	}
}

// GetDefaultMediaConfig 获取默认媒体配置
func GetDefaultMediaConfig() *models.MediaConfig {
	return &models.MediaConfig{
		SearchSource:    "tmdb",
		RecognizeSource: "tmdb",
	}
}

// GetDefaultTMDBConfig 获取默认TMDB配置
func GetDefaultTMDBConfig() *models.TMDBConfig {
	return &models.TMDBConfig{
		ImageDomain:  "image.tmdb.org",
		Language:     "zh-CN",
		Region:       "CN",
		ProxyEnabled: false,
	}
}

// GetDefaultSiteConfig 获取默认站点配置
func GetDefaultSiteConfig() *models.SiteConfig {
	return &models.SiteConfig{
		DataRefreshInterval: 3600,
		MaxConcurrentTasks:  5,
		RequestTimeout:      30,
		RetryTimes:          3,
		RetryInterval:       5,
	}
}

// GetDefaultDownloadConfig 获取默认下载配置
func GetDefaultDownloadConfig() *models.DownloadConfig {
	return &models.DownloadConfig{
		TorrentTag:             "MoviePilot",
		DownloadSubtitle:       true,
		MaxConcurrentDownloads: 5,
	}
}

// GetDefaultCookieCloudConfig 获取默认CookieCloud配置
func GetDefaultCookieCloudConfig() *models.CookieCloudConfig {
	return &models.CookieCloudConfig{
		Interval: 3600,
		Enabled:  false,
	}
}

// GetDefaultTransferConfig 获取默认整理配置
func GetDefaultTransferConfig() *models.TransferConfig {
	return &models.TransferConfig{
		Enabled:      true,
		AutoTransfer: true,
		DeleteSource: false,
	}
}

// GetDefaultPluginConfig 获取默认插件配置
func GetDefaultPluginConfig() *models.PluginConfig {
	return &models.PluginConfig{
		AutoReload: false,
		MaxWorkers: 10,
		Timeout:    30,
	}
}

// GetDefaultPerformanceConfig 获取默认性能配置
func GetDefaultPerformanceConfig() *models.PerformanceConfig {
	return &models.PerformanceConfig{
		BigMemoryMode:                    false,
		EncodingDetectionPerformanceMode: true,
		EncodingDetectionMinConfidence:   0.8,
		MemoryGCInterval:                 30,
	}
}

// GetDefaultNetworkConfig 获取默认网络配置
func GetDefaultNetworkConfig() *models.NetworkConfig {
	return &models.NetworkConfig{
		DOHEnable: false,
	}
}

// GetDefaultSchedulerConfig 获取默认调度器配置
func GetDefaultSchedulerConfig() *models.SchedulerConfig {
	return &models.SchedulerConfig{
		MaxConcurrentTasks: 10,
		TaskTimeout:        3600,
	}
}

// GetDefaultSubscribeConfig 获取默认订阅配置
func GetDefaultSubscribeConfig() *models.SubscribeConfig {
	return &models.SubscribeConfig{
		CheckInterval:       300,
		MaxConcurrentChecks: 5,
		NotificationEnabled: true,
	}
}
