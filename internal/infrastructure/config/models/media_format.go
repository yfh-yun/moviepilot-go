package models

// MediaFormatConfig 媒体文件格式配置
type MediaFormatConfig struct {
	// 支持的媒体文件后缀格式
	MediaExt []string `mapstructure:"RMT_MEDIAEXT" default:":[".mp4", ".mkv", ".ts", ".iso", ".rmvb", ".avi", ".mov", ".mpeg", ".mpg", ".wmv", ".3gp", ".asf", ".m4v", ".flv", ".m2ts", ".strm", ".tp", ".f4v"]"`
	
	// 支持的字幕文件后缀格式
	SubExt []string `mapstructure:"RMT_SUBEXT" default:":[".srt", ".ass", ".ssa", ".sup"]"`
	
	// 支持的音轨文件后缀格式
	AudioTrackExt []string `mapstructure:"RMT_AUDIO_TRACK_EXT" default:":[".mka"]"`
	
	// 支持的音频文件后缀格式
	AudioExt []string `mapstructure:"RMT_AUDIOEXT" default:":[".aac", ".ac3", ".amr", ".caf", ".cda", ".dsf", ".dff", ".kar", ".m4a", ".mp1", ".mp2", ".mp3", ".mid", ".mod", ".mka", ".mpc", ".nsf", ".ogg", ".pcm", ".rmi", ".s3m", ".snd", ".spx", ".tak", ".tta", ".vqf", ".wav", ".wma", ".aifc", ".aiff", ".alac", ".adif", ".adts", ".flac", ".midi", ".opus", ".sfalc"]"`
	
	// 下载器临时文件后缀
	DownloadTmpExt []string `mapstructure:"DOWNLOAD_TMPEXT" default:":[".!qb", ".part"]"`
}