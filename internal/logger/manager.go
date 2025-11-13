package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerManager 日志管理�?type LoggerManager struct {
	// 管理所有的 Logger
	loggers map[string]*zap.Logger
	mutex   sync.RWMutex
	
	// 默认日志文件名称
	defaultLogFile string
	
	// 非阻塞文件处理器
	fileHandler *NonBlockingFileHandler
}

// loggerManagerInstance 日志管理器单�?var loggerManagerInstance *LoggerManager
var managerOnce sync.Once

// GetLoggerManager 获取日志管理器单�?func GetLoggerManager() *LoggerManager {
	managerOnce.Do(func() {
		loggerManagerInstance = &LoggerManager{
			loggers:        make(map[string]*zap.Logger),
			defaultLogFile: "moviepilot.log",
			fileHandler:    GetNonBlockingFileHandler(),
		}
	})
	return loggerManagerInstance
}

// GetLogger 获取一个指定名称的、独立的日志记录�?// 创建一个独立的日志文件，例�?'diag_memory.log'
// name: 日志记录器的名称，也将用作文件名
func (lm *LoggerManager) GetLogger(name string) *zap.Logger {
	logFile := fmt.Sprintf("%s.log", name)
	
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	// 检查是否已经创建过这个 logger
	if logger, exists := lm.loggers[logFile]; exists {
		return logger
	}
	
	// 如果没有，就创建一个新�?	logger := lm.setupConsoleLogger(logFile)
	lm.loggers[logFile] = logger
	return logger
}

// getCaller 获取调用者的文件名称与插件名�?// 如果是插件调用内置的模块, 也能写入到插件日志文件中
func (lm *LoggerManager) getCaller() (callerName, pluginName string) {
	// 获取调用栈，需要跳过更多的层级来获取正确的调用�?	for i := 3; i <= 6; i++ {
		pc, file, _, ok := runtime.Caller(i)
		if !ok {
			continue
		}
		
		// 获取调用者文件名�?		filePath := filepath.Base(file)
		callerName = strings.TrimSuffix(filePath, filepath.Ext(filePath))
		
		// 获取调用者函数名�?		funcName := runtime.FuncForPC(pc).Name()
		
		// 检查是否是插件调用
		if strings.Contains(funcName, "plugin") || strings.Contains(file, "plugins") {
			// 提取插件名称
			parts := strings.Split(file, string(filepath.Separator))
			for j, part := range parts {
				if part == "plugins" && j+1 < len(parts) {
					pluginName = parts[j+1]
					break
				}
			}
			break
		}
		
		// 如果找到了非日志模块的调用者，就使用它
		if !strings.Contains(file, "logger") && callerName != "" {
			break
		}
	}
	
	// 默认�?	if callerName == "" {
		callerName = "logger.go"
	}
	
	return callerName, pluginName
}

// setupConsoleLogger 初始化控制台日志实例
func (lm *LoggerManager) setupConsoleLogger(logFile string) *zap.Logger {
	settings := GetLogSettings()
	
	// 设置日志级别
	level := zapcore.InfoLevel
	switch strings.ToUpper(settings.LogLevel) {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARNING", "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "CRITICAL":
		level = zapcore.FatalLevel
	}
	
	if settings.Debug {
		level = zapcore.DebugLevel
	}
	
	// 创建编码器配�?	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	
	// 处理控制台格�?	consoleFormat := settings.LogConsoleFormat
	if consoleFormat != "" {
		// 解析格式字符串并设置相应的编码器配置
		if strings.Contains(consoleFormat, "%(leveltext)s") {
			// 已经在EncodeLevel中处理颜�?		}
		if strings.Contains(consoleFormat, "%(name)s") {
			encoderConfig.NameKey = "name"
		}
		if strings.Contains(consoleFormat, "%(asctime)s") {
			encoderConfig.TimeKey = "time"
		}
		if strings.Contains(consoleFormat, "%(message)s") {
			encoderConfig.MessageKey = "message"
		}
	}
	
	// 创建控制台核�?	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)
	
	// 创建logger
	logger := zap.New(consoleCore, zap.AddCaller(), zap.AddCallerSkip(2))
	return logger
}

// getLogLevel 获取当前日志级别
func (lm *LoggerManager) getLogLevel() LogLevel {
	settings := GetLogSettings()
	
	level := INFO
	switch strings.ToUpper(settings.LogLevel) {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARNING", "WARN":
		level = WARNING
	case "ERROR":
		level = ERROR
	case "CRITICAL":
		level = CRITICAL
	}
	
	if settings.Debug {
		level = DEBUG
	}
	
	return level
}

// log 记录日志的通用方法
func (lm *LoggerManager) log(level LogLevel, msg string, args ...interface{}) {
	// 获取当前日志级别设置
	currentLevel := lm.getLogLevel()
	
	// 如果当前方法的级别低于设定的日志级别，则不处�?	if level < currentLevel {
		return
	}
	
	// 获取调用者文件名和插件名
	callerName, pluginName := lm.getCaller()
	
	// 格式化消�?	formattedMsg := fmt.Sprintf("%s - %s", callerName, msg)
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(formattedMsg, args...)
	}
	
	// 构建日志文件路径
	settings := GetLogSettings()
	var logFilePath string
	if pluginName != "" {
		// 使用插件日志文件
		logFilePath = filepath.Join(settings.GetLogPath(), "plugins", fmt.Sprintf("%s.log", pluginName))
	} else {
		// 使用默认日志文件
		logFilePath = filepath.Join(settings.GetLogPath(), lm.defaultLogFile)
	}
	
	// 使用非阻塞文件处理器写入文件日志
	lm.fileHandler.WriteLog(level, formattedMsg, logFilePath)
	
	// 同时保持控制台输�?	lm.mutex.RLock()
	logger, exists := lm.loggers[filepath.Base(logFilePath)]
	lm.mutex.RUnlock()
	
	if !exists {
		lm.mutex.Lock()
		// 双重检�?		logger, exists = lm.loggers[filepath.Base(logFilePath)]
		if !exists {
			logger = lm.setupConsoleLogger(filepath.Base(logFilePath))
			lm.loggers[filepath.Base(logFilePath)] = logger
		}
		lm.mutex.Unlock()
	}
	
	// 控制台输�?	switch level {
	case DEBUG:
		logger.Debug(formattedMsg)
	case INFO:
		logger.Info(formattedMsg)
	case WARNING:
		logger.Warn(formattedMsg)
	case ERROR:
		logger.Error(formattedMsg)
	case CRITICAL:
		logger.Fatal(formattedMsg)
	}
}

// Info 输出信息级别日志
func (lm *LoggerManager) Info(msg string, args ...interface{}) {
	lm.log(INFO, msg, args...)
}

// Debug 输出调试级别日志
func (lm *LoggerManager) Debug(msg string, args ...interface{}) {
	lm.log(DEBUG, msg, args...)
}

// Warning 输出警告级别日志
func (lm *LoggerManager) Warning(msg string, args ...interface{}) {
	lm.log(WARNING, msg, args...)
}

// Warn 输出警告级别日志（兼容）
func (lm *LoggerManager) Warn(msg string, args ...interface{}) {
	lm.log(WARNING, msg, args...)
}

// Error 输出错误级别日志
func (lm *LoggerManager) Error(msg string, args ...interface{}) {
	lm.log(ERROR, msg, args...)
}

// Critical 输出严重错误级别日志
func (lm *LoggerManager) Critical(msg string, args ...interface{}) {
	lm.log(CRITICAL, msg, args...)
}

// UpdateLoggers 更新日志实例
func (lm *LoggerManager) UpdateLoggers() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	// 重新创建所有logger以应用新的配�?	newLoggers := make(map[string]*zap.Logger)
	for logFile, oldLogger := range lm.loggers {
		newLogger := lm.setupConsoleLogger(logFile)
		newLoggers[logFile] = newLogger
		oldLogger.Sync()
	}
	
	lm.loggers = newLoggers
	
	// 同时更新文件处理器的配置
	if lm.fileHandler != nil {
		// 重新初始化文件处理器以应用新配置
		lm.fileHandler.Shutdown()
		lm.fileHandler = GetNonBlockingFileHandler()
	}
}

// Shutdown 关闭日志管理器，清理资源
func (lm *LoggerManager) Shutdown() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()
	
	// 关闭所有logger
	for _, logger := range lm.loggers {
		logger.Sync()
	}
	
	// 关闭文件处理�?	if lm.fileHandler != nil {
		lm.fileHandler.Shutdown()
	}
	
	// 清空logger映射
	lm.loggers = make(map[string]*zap.Logger)
}
