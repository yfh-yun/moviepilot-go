package database

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"

	appLogger "moviepilot-go/pkg/logger"
)

// zapGormLogger 是一个使用 zap 的 GORM 日志实现
// 仅用于数据库层内部，使 SQL 日志进入统一的日志体系。
type zapGormLogger struct {
	logger                    *zap.Logger
	logLevel                  logger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// NewZapGormLogger 创建一个基于 zap 的 GORM 日志器。
func NewZapGormLogger(baseLogger *zap.Logger) logger.Interface {
	l := baseLogger
	if l == nil {
		l = appLogger.GetLogger()
	}

	return &zapGormLogger{
		logger:                    l,
		logLevel:                  logger.Info,
		slowThreshold:             200 * time.Millisecond,
		ignoreRecordNotFoundError: true,
	}
}

// LogMode 设置日志级别
func (l *zapGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 记录信息级别日志
func (l *zapGormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= logger.Info {
		l.logger.Sugar().Infof(msg, data...)
	}
}

// Warn 记录警告级别日志
func (l *zapGormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= logger.Warn {
		l.logger.Sugar().Warnf(msg, data...)
	}
}

// Error 记录错误级别日志
func (l *zapGormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= logger.Error {
		l.logger.Sugar().Errorf(msg, data...)
	}
}

// Trace 记录 SQL 执行相关日志
func (l *zapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Duration("duration", elapsed),
		zap.Int64("rows", rows),
	}

	if err != nil {
		if l.ignoreRecordNotFoundError && err == logger.ErrRecordNotFound {
			return
		}
		if l.logLevel >= logger.Error {
			l.logger.Error("gorm query failed", append(fields, zap.Error(err))...)
		}
		return
	}

	if l.slowThreshold > 0 && elapsed > l.slowThreshold && l.logLevel >= logger.Warn {
		l.logger.Warn("gorm slow query", fields...)
		return
	}

	if l.logLevel >= logger.Info {
		l.logger.Debug("gorm query", fields...)
	}
}
