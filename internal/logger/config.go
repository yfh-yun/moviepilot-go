package logger

import (
	"path/filepath"
	"sync"
	
	"github.com/spf13/viper"
)

// LogConfig 日志配置结构�?type LogConfig struct {
	// 配置文件目录
	ConfigDir string `mapstructure:"CONFIG_DIR"`
	
	// 是否为调试模�?	Debug bool `mapstructure:"DEBUG"`
	
	// 日志级别（DEBUG、INFO、WARNING、ERROR等）
	LogLevel string `mapstructure:"LOG_LEVEL"`
	
	// 日志文件最大大小（单位：MB�?	LogMaxFileSize int `mapstructure:"LOG_MAX_FILE_SIZE"`
	
	// 备份的日志文件数�?	LogBackupCount int `mapstructure:"LOG_BACKUP_COUNT"`
	
	// 控制台日志格�?	LogConsoleFormat string `mapstructure:"LOG_CONSOLE_FORMAT"`
	
	// 文件日志格式
	LogFileFormat string `mapstructure:"LOG_FILE_FORMAT"`
	
	// 异步文件写入队列大小
	AsyncFileQueueSize int `mapstructure:"ASYNC_FILE_QUEUE_SIZE"`
	
	// 异步文件写入线程�?	AsyncFileWorkers int `mapstructure:"ASYNC_FILE_WORKERS"`
	
	// 批量写入大小
	BatchWriteSize int `mapstructure:"BATCH_WRITE_SIZE"`
	
	// 写入超时时间（秒�?	WriteTimeout float64 `mapstructure:"WRITE_TIMEOUT"`
}

// logSettings 日志设置实例
var logSettings *LogConfig
var once sync.Once

// GetLogSettings 获取日志设置单例
func GetLogSettings() *LogConfig {
	once.Do(func() {
		logSettings = &LogConfig{
			Debug:              false,
			LogLevel:           "INFO",
			LogMaxFileSize:     5,
			LogBackupCount:     10,
			LogConsoleFormat:   "%(leveltext)s[%(name)s] %(asctime)s %(message)s",
			LogFileFormat:      "�?(levelname)s�?(asctime)s - %(message)s",
			AsyncFileQueueSize: 1000,
			AsyncFileWorkers:   2,
			BatchWriteSize:     50,
			WriteTimeout:       3.0,
		}
	})
	return logSettings
}

// LoadConfig 从配置文件加载日志配�?func (c *LogConfig) LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	
	// 设置默认�?	viper.SetDefault("DEBUG", false)
	viper.SetDefault("LOG_LEVEL", "INFO")
	viper.SetDefault("LOG_MAX_FILE_SIZE", 5)
	viper.SetDefault("LOG_BACKUP_COUNT", 10)
	viper.SetDefault("LOG_CONSOLE_FORMAT", "%(leveltext)s[%(name)s] %(asctime)s %(message)s")
	viper.SetDefault("LOG_FILE_FORMAT", "�?(levelname)s�?(asctime)s - %(message)s")
	viper.SetDefault("ASYNC_FILE_QUEUE_SIZE", 1000)
	viper.SetDefault("ASYNC_FILE_WORKERS", 2)
	viper.SetDefault("BATCH_WRITE_SIZE", 50)
	viper.SetDefault("WRITE_TIMEOUT", 3.0)
	
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	
	if err := viper.Unmarshal(c); err != nil {
		return err
	}
	
	return nil
}

// GetConfigPath 获取配置路径
func (c *LogConfig) GetConfigPath() string {
	if c.ConfigDir != "" {
		return c.ConfigDir
	}
	// 默认使用当前目录下的config目录
	return filepath.Join(".", "config")
}

// GetLogPath 获取日志存储路径
func (c *LogConfig) GetLogPath() string {
	return filepath.Join(c.GetConfigPath(), "logs")
}

// GetLogMaxFileSizeBytes 将日志文件大小转换为字节（MB -> Bytes�?func (c *LogConfig) GetLogMaxFileSizeBytes() int {
	return c.LogMaxFileSize * 1024 * 1024
}
