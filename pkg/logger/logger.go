// Package logger MoviePilot日志管理模块
package logger

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
	// Sugar 全局sugar日志实例（提供更灵活的接口）
	Sugar *zap.SugaredLogger

	// 上下文键定义
	ContextKeyRequestID = contextKey("request_id")
	ContextKeyUserID    = contextKey("user_id")
	ContextKeyTraceID   = contextKey("trace_id")

	// 环境变量前缀
	envPrefix = "LOGGER_"
)

// contextKey 上下文键类型
type contextKey string

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnvOrDefault 获取整数环境变量或返回默认值
func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getBoolEnvOrDefault 获取布尔环境变量或返回默认值
func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// ContextLogger 带上下文的日志结构体
type ContextLogger struct {
	logger *zap.Logger
}

// Init 初始化日志系统
func Init() error {
	// 创建日志配置
	logConfig := buildLogConfig()

	var logger *zap.Logger
	var err error

	// 检查是否需要多路输出
	if len(logConfig.OutputPaths) > 0 && logConfig.OutputPaths[0] == "multiwriter:" {
		// 处理多路输出的特殊配置
		logger, err = buildMultiWriterLogger(&logConfig)
	} else {
		// 标准配置构建
		logger, err = logConfig.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	}

	if err != nil {
		return err
	}

	Logger = logger
	Sugar = logger.Sugar()

	return nil
}

// buildLogConfig 构建日志配置
func buildLogConfig() zap.Config {
	// 从环境变量获取配置，如果没有则使用默认值
	level := getEnvOrDefault(envPrefix+"LEVEL", "info")
	format := getEnvOrDefault(envPrefix+"FORMAT", "json")
	output := getEnvOrDefault(envPrefix+"OUTPUT", "stdout")

	// 创建基础配置
	logConfig := zap.NewProductionConfig()

	// 设置日志级别
	switch level {
	case "debug":
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "fatal":
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case "panic":
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	default:
		logConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// 设置编码器
	if format == "console" {
		logConfig.Encoding = "console"
		logConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		logConfig.Encoding = "json"
		logConfig.EncoderConfig = zap.NewProductionEncoderConfig()
		logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	// 设置输出
	switch output {
	case "file":
		configureFileOutput(&logConfig, false)
	case "both":
		configureFileOutput(&logConfig, true)
	default:
		// 标准输出
		logConfig.OutputPaths = []string{"stdout"}
		logConfig.ErrorOutputPaths = []string{"stderr"}
	}

	return logConfig
}

// buildMultiWriterLogger 构建多路输出的日志器
func buildMultiWriterLogger(logConfig *zap.Config) (*zap.Logger, error) {
	// 提取文件路径
	filePath := ""
	if len(logConfig.OutputPaths) > 0 {
		path := logConfig.OutputPaths[0]
		if len(path) > 12 && path[:12] == "multiwriter:" {
			filePath = path[12:]
		}
	}

	if filePath == "" {
		filePath = getEnvOrDefault(envPrefix+"FILE", "/var/log/moviepilot/app.log")
	}

	// 创建文件写入器（使用真实的 lumberjack.Logger 进行日志轮转）
	fileWriter := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    getIntEnvOrDefault(envPrefix+"MAX_SIZE", 100),
		MaxBackups: getIntEnvOrDefault(envPrefix+"MAX_BACKUPS", 3),
		MaxAge:     getIntEnvOrDefault(envPrefix+"MAX_AGE", 28),
		Compress:   getBoolEnvOrDefault(envPrefix+"COMPRESS", true),
		LocalTime:  true,
	}

	// 创建多路写入器
	multiWriter := zapcore.AddSync(io.MultiWriter(os.Stdout, fileWriter))

	// 创建编码器
	var encoder zapcore.Encoder
	if logConfig.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(logConfig.EncoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(logConfig.EncoderConfig)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, multiWriter, logConfig.Level)

	// 构建日志器
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// configureFileOutput 配置文件输出
func configureFileOutput(logConfig *zap.Config, includeStdout bool) {
	filePath := getEnvOrDefault(envPrefix+"FILE", "/var/log/moviepilot/app.log")

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		// 如果创建目录失败，回退到标准输出
		// 注意：这里不能使用logger，因为logger可能还未初始化
		logConfig.OutputPaths = []string{"stdout"}
		logConfig.ErrorOutputPaths = []string{"stderr"}
		return
	}

	// 配置输出路径
	if includeStdout {
		// 注意：多路输出将在 Init() 函数中处理
		// 这里只配置标记

		// 创建新的配置以支持多路输出
		*logConfig = zap.Config{
			Level:             logConfig.Level,
			Development:       logConfig.Development,
			DisableCaller:     logConfig.DisableCaller,
			DisableStacktrace: logConfig.DisableStacktrace,
			Sampling:          logConfig.Sampling,
			Encoding:          logConfig.Encoding,
			EncoderConfig:     logConfig.EncoderConfig,
			OutputPaths:       []string{"stdout"}, // 用于 Build() 方法
			ErrorOutputPaths:  []string{"stderr"}, // 用于 Build() 方法
		}

		// 注意：实际的多路输出需要在 Init() 中特殊处理
		// 这里标记需要使用自定义核心
		logConfig.OutputPaths = []string{"multiwriter:" + filePath}
	} else {
		// 仅文件输出
		logConfig.OutputPaths = []string{filePath}
		logConfig.ErrorOutputPaths = []string{filePath}
	}

	// 创建自定义编码器配置
	encoderConfig := logConfig.EncoderConfig
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.LevelKey = "level"
	encoderConfig.NameKey = "logger"
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "message"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	logConfig.EncoderConfig = encoderConfig
}

// createDefaultLogger 创建默认日志实例
func createDefaultLogger() (*zap.Logger, error) {
	// 创建默认配置，确保符合规范
	config := zap.NewProductionConfig()
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// WithContext 从上下文中提取信息并创建带上下文的日志实例
func WithContext(ctx context.Context) *ContextLogger {
	fields := []zap.Field{}

	// 从上下文提取request_id
	if requestID, ok := ctx.Value(ContextKeyRequestID).(string); ok && requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	// 从上下文提取user_id
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	// 从上下文提取trace_id
	if traceID, ok := ctx.Value(ContextKeyTraceID).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	return &ContextLogger{
		logger: GetLogger().With(fields...),
	}
}

// GetLogger 获取日志实例
func GetLogger() *zap.Logger {
	if Logger == nil {
		// 如果没有初始化，创建一个默认的
		logger, err := createDefaultLogger()
		if err != nil {
			// 如果创建失败，返回一个开发模式的日志器
			logger, _ = zap.NewDevelopment()
		}
		return logger
	}
	return Logger
}

// Debug 带上下文的调试日志
func (l *ContextLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info 带上下文的信息日志
func (l *ContextLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn 带上下文的警告日志
func (l *ContextLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error 带上下文的错误日志
func (l *ContextLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal 带上下文的致命错误日志
func (l *ContextLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// Panic 带上下文的Panic日志
func (l *ContextLogger) Panic(msg string, fields ...zap.Field) {
	l.logger.Panic(msg, fields...)
}

// Debugf 带上下文的格式化调试日志
func (l *ContextLogger) Debugf(template string, args ...any) {
	l.logger.Sugar().Debugf(template, args...)
}

// Infof 带上下文的格式化信息日志
func (l *ContextLogger) Infof(template string, args ...any) {
	l.logger.Sugar().Infof(template, args...)
}

// Warnf 带上下文的格式化警告日志
func (l *ContextLogger) Warnf(template string, args ...any) {
	l.logger.Sugar().Warnf(template, args...)
}

// Errorf 带上下文的格式化错误日志
func (l *ContextLogger) Errorf(template string, args ...any) {
	l.logger.Sugar().Errorf(template, args...)
}

// Fatalf 带上下文的格式化致命错误日志
func (l *ContextLogger) Fatalf(template string, args ...any) {
	l.logger.Sugar().Fatalf(template, args...)
}

// Panicf 带上下文的格式化Panic日志
func (l *ContextLogger) Panicf(template string, args ...any) {
	l.logger.Sugar().Panicf(template, args...)
}

// GetSugar 获取Sugar日志实例
func GetSugar() *zap.SugaredLogger {
	if Sugar == nil {
		// 如果没有初始化，创建一个默认的
		logger, err := createDefaultLogger()
		if err != nil {
			// 如果创建失败，返回一个开发模式的日志器
			logger, _ = zap.NewDevelopment()
		}
		return logger.Sugar()
	}
	return Sugar
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Panic Panic日志
func Panic(msg string, fields ...zap.Field) {
	GetLogger().Panic(msg, fields...)
}

// Debugf 格式化调试日志
func Debugf(template string, args ...any) {
	GetSugar().Debugf(template, args...)
}

// Infof 格式化信息日志
func Infof(template string, args ...any) {
	GetSugar().Infof(template, args...)
}

// Warnf 格式化警告日志
func Warnf(template string, args ...any) {
	GetSugar().Warnf(template, args...)
}

// Errorf 格式化错误日志
func Errorf(template string, args ...any) {
	GetSugar().Errorf(template, args...)
}

// Fatalf 格式化致命错误日志
func Fatalf(template string, args ...any) {
	GetSugar().Fatalf(template, args...)
}

// Panicf 格式化Panic日志
func Panicf(template string, args ...any) {
	GetSugar().Panicf(template, args...)
}
