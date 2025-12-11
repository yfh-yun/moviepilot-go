package models

// CacheConfig 缓存配置
type CacheConfig struct {
	BackendType      string `mapstructure:"CACHE_BACKEND_TYPE" default:"cachetools"`
	BackendURL       string `mapstructure:"CACHE_BACKEND_URL" default:"redis://localhost:6379"`
	RedisMaxMemory   string `mapstructure:"CACHE_REDIS_MAXMEMORY"`
	GlobalImageCache bool   `mapstructure:"GLOBAL_IMAGE_CACHE" default:"false"`
	ImageCacheDays   int    `mapstructure:"GLOBAL_IMAGE_CACHE_DAYS" default:"7"`
	TempFileDays     int    `mapstructure:"TEMP_FILE_DAYS" default:"3"`
	MetaCacheExpire  int    `mapstructure:"META_CACHE_EXPIRE" default:"0"`
}
