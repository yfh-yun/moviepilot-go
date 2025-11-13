package logger

import (
	"time"
)

// LogEntry 日志条目
type LogEntry struct {
	Level     LogLevel
	Message   string
	FilePath  string
	Timestamp time.Time
}

// NewLogEntry 创建新的日志条目
func NewLogEntry(level LogLevel, message, filePath string, timestamp time.Time) *LogEntry {
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	
	return &LogEntry{
		Level:     level,
		Message:   message,
		FilePath:  filePath,
		Timestamp: timestamp,
	}
}

// LogEntryQueue 日志条目队列接口
type LogEntryQueue interface {
	// PutNowait 非阻塞放入条�?	PutNowait(entry *LogEntry) error
	
	// GetTimeout 超时获取条目
	GetTimeout(timeout float64) (*LogEntry, error)
	
	// Empty 队列是否为空
	Empty() bool
}
