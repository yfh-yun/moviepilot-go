package logger

import (
	"fmt"
	"strings"
	"time"
)

// LogLevel 日志级别类型
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	CRITICAL
)

// LogLevelName 获取日志级别名称
func (l LogLevel) Name() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ColorCode 获取日志级别对应的颜色代�?func (l LogLevel) ColorCode() string {
	switch l {
	case DEBUG:
		return "\033[36m" // 青色
	case INFO:
		return "\033[32m" // 绿色
	case WARNING:
		return "\033[33m" // 黄色
	case ERROR:
		return "\033[31m" // 红色
	case CRITICAL:
		return "\033[91m" // 亮红�?	default:
		return "\033[0m" // 默认颜色
	}
}

// CustomFormatter 自定义日志输出格�?type CustomFormatter struct {
	FormatPattern string
}

// NewCustomFormatter 创建新的自定义格式化�?func NewCustomFormatter(format string) *CustomFormatter {
	return &CustomFormatter{
		FormatPattern: format,
	}
}

// FormatMessage 格式化日志条�?func (f *CustomFormatter) FormatMessage(level LogLevel, name, message string, timestamp time.Time) string {
	// 应用颜色
	coloredLevel := fmt.Sprintf("%s%s\033[0m", level.ColorCode(), level.Name())
	
	// 计算分隔符空格数
	separator := ""
	if len(level.Name()) < 8 {
		separator = fmt.Sprintf("%*s", 8-len(level.Name()), "")
	}
	
	levelText := fmt.Sprintf("%s:%s", coloredLevel, separator)
	
	// 根据格式字符串进行替�?	formatted := f.FormatPattern
	formatted = strings.Replace(formatted, "%(leveltext)s", levelText, -1)
	formatted = strings.Replace(formatted, "%(name)s", name, -1)
	formatted = strings.Replace(formatted, "%(asctime)s", timestamp.Format("2006/01/02 15:04:05"), -1)
	formatted = strings.Replace(formatted, "%(message)s", message, -1)
	
	return formatted
}
