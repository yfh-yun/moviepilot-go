package models

// PluginConfig 插件配置
type PluginConfig struct {
	Market     string `mapstructure:"PLUGIN_MARKET"`
	AutoReload bool   `mapstructure:"PLUGIN_AUTO_RELOAD" default:"false"`
	MaxWorkers int    `mapstructure:"PLUGIN_MAX_WORKERS" default:"10"`
	Timeout    int    `mapstructure:"PLUGIN_TIMEOUT" default:"30"`
}

// CookieCloudConfig CookieCloud配置
type CookieCloudConfig struct {
	Host     string `mapstructure:"COOKIECLOUD_HOST"`
	Key      string `mapstructure:"COOKIECLOUD_KEY"`
	Password string `mapstructure:"COOKIECLOUD_PASSWORD"`
	Interval int    `mapstructure:"COOKIECLOUD_INTERVAL" default:"3600"`
	Enabled  bool   `mapstructure:"COOKIECLOUD_ENABLE" default:"false"`
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	TorrentTag             string `mapstructure:"TORRENT_TAG" default:"MoviePilot"`
	DownloadSubtitle       bool   `mapstructure:"DOWNLOAD_SUBTITLE" default:"true"`
	MaxConcurrentDownloads int    `mapstructure:"MAX_CONCURRENT_DOWNLOADS" default:"5"`
	DownloadPath           string `mapstructure:"DOWNLOAD_PATH"`
	TempPath               string `mapstructure:"TEMP_PATH"`
}

// TransferConfig 整理配置
type TransferConfig struct {
	Enabled      bool   `mapstructure:"TRANSFER_ENABLE" default:"true"`
	MoviePath    string `mapstructure:"MOVIE_PATH"`
	TVPath       string `mapstructure:"TV_PATH"`
	AutoTransfer bool   `mapstructure:"AUTO_TRANSFER" default:"true"`
	DeleteSource bool   `mapstructure:"DELETE_SOURCE" default:"false"`
}
