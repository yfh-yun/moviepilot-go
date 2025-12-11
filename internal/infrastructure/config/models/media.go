package models

// MediaConfig 媒体配置
type MediaConfig struct {
	SearchSource      string `mapstructure:"SEARCH_SOURCE" default:"tmdb"`
	RecognizeSource   string `mapstructure:"RECOGNIZE_SOURCE" default:"tmdb"`
	MovieRenameFormat string `mapstructure:"MOVIE_RENAME_FORMAT"`
	TVRenameFormat    string `mapstructure:"TV_RENAME_FORMAT"`
}

// TMDBConfig TMDB配置
type TMDBConfig struct {
	APIKey       string `mapstructure:"TMDB_API_KEY"`
	ImageDomain  string `mapstructure:"TMDB_IMAGE_DOMAIN" default:"image.tmdb.org"`
	Language     string `mapstructure:"TMDB_LANGUAGE" default:"zh-CN"`
	Region       string `mapstructure:"TMDB_REGION" default:"CN"`
	ProxyEnabled bool   `mapstructure:"TMDB_PROXY_ENABLE" default:"false"`
	ProxyURL     string `mapstructure:"TMDB_PROXY_URL"`
}

// SiteConfig 站点配置
type SiteConfig struct {
	DataRefreshInterval int    `mapstructure:"SITEDATA_REFRESH_INTERVAL" default:"3600"`
	OCRHost             string `mapstructure:"OCR_HOST"`
	OCRAPIKey           string `mapstructure:"OCR_API_KEY"`
	MaxConcurrentTasks  int    `mapstructure:"SITE_MAX_CONCURRENT_TASKS" default:"5"`
	RequestTimeout      int    `mapstructure:"SITE_REQUEST_TIMEOUT" default:"30"`
	RetryTimes          int    `mapstructure:"SITE_RETRY_TIMES" default:"3"`
	RetryInterval       int    `mapstructure:"SITE_RETRY_INTERVAL" default:"5"`
}
