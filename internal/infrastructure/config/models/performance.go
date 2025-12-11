package models

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	BigMemoryMode                    bool    `mapstructure:"BIG_MEMORY_MODE" default:"false"`
	EncodingDetectionPerformanceMode bool    `mapstructure:"ENCODING_DETECTION_PERFORMANCE_MODE" default:"true"`
	EncodingDetectionMinConfidence   float64 `mapstructure:"ENCODING_DETECTION_MIN_CONFIDENCE" default:"0.8"`
	MemoryGCInterval                 int     `mapstructure:"MEMORY_GC_INTERVAL" default:"30"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ProxyHost     string   `mapstructure:"PROXY_HOST"`
	ProxyPort     int      `mapstructure:"PROXY_PORT"`
	ProxyUsername string   `mapstructure:"PROXY_USERNAME"`
	ProxyPassword string   `mapstructure:"PROXY_PASSWORD"`
	DOHEnable     bool     `mapstructure:"DOH_ENABLE" default:"false"`
	DOHDomains    []string `mapstructure:"DOH_DOMAINS"`
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	MaxConcurrentTasks int `mapstructure:"SCHEDULER_MAX_CONCURRENT_TASKS" default:"10"`
	TaskTimeout        int `mapstructure:"SCHEDULER_TASK_TIMEOUT" default:"3600"`
}

// SubscribeConfig 订阅配置
type SubscribeConfig struct {
	CheckInterval       int  `mapstructure:"SUBSCRIBE_CHECK_INTERVAL" default:"300"`
	MaxConcurrentChecks int  `mapstructure:"SUBSCRIBE_MAX_CONCURRENT_CHECKS" default:"5"`
	NotificationEnabled bool `mapstructure:"SUBSCRIBE_NOTIFICATION_ENABLED" default:"true"`
}

// SystemConfModel 系统资源配置（根据性能模式动态计算）
type SystemConfModel struct {
	Torrents   int
	Refresh    int
	TMDB       int
	Douban     int
	Bangumi    int
	Fanart     int
	Meta       int
	Scheduler  int
	ThreadPool int
}
