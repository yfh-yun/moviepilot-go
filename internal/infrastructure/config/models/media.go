package models

// MediaConfig 媒体配置
type MediaConfig struct {
	SearchSource      string   `mapstructure:"SEARCH_SOURCE" default:"themoviedb"`
	RecognizeSource   string   `mapstructure:"RECOGNIZE_SOURCE" default:"themoviedb"`
	ScrapSource       string   `mapstructure:"SCRAP_SOURCE" default:"themoviedb"`
	MovieRenameFormat string   `mapstructure:"MOVIE_RENAME_FORMAT" default:"{{title}}{% if year %} ({{year}}){% endif %}/{{title}}{% if year %} ({{year}}){% endif %}{% if part %}-{{part}}{% endif %}{% if videoFormat %} - {{videoFormat}}{% endif %}{{fileExt}}"`
	TVRenameFormat    string   `mapstructure:"TV_RENAME_FORMAT" default:"{{title}}{% if year %} ({{year}}){% endif %}/Season {{season}}/{{title}} - {{season_episode}}{% if part %}-{{part}}{% endif %}{% if episode %} - 第 {{episode}} 集{% endif %}{{fileExt}}"`
	RenameFormatS0Names []string `mapstructure:"RENAME_FORMAT_S0_NAMES" default:":["Specials", "SPs"]"`
	DefaultSub        string   `mapstructure:"DEFAULT_SUB" default:"zh-cn"`
	ScrapFollowTMDB   bool     `mapstructure:"SCRAP_FOLLOW_TMDB" default:"true"`
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
