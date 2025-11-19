// Package logger 日志系统测试文件
package logger

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"
)

// TestLoggerInitialization 测试日志初始化
func TestLoggerInitialization(t *testing.T) {
	// 保存原始的环境变量
	oldLevel := os.Getenv("LOGGER_LEVEL")
	defer os.Setenv("LOGGER_LEVEL", oldLevel)

	// 设置测试环境变量
	os.Setenv("LOGGER_LEVEL", "debug")
	os.Setenv("LOGGER_FORMAT", "json")

	// 初始化日志系统
	err := Init()
	if err != nil {
		t.Fatalf("日志初始化失败: %v", err)
	}
	defer Sync()

	// 验证全局Logger不为nil
	if Logger == nil {
		t.Error("Logger全局实例为nil")
	}

	// 验证全局Sugar不为nil
	if Sugar == nil {
		t.Error("Sugar全局实例为nil")
	}
}

// TestBasicLogging 测试基本日志功能
func TestBasicLogging(t *testing.T) {
	// 初始化测试用日志
	mockWriter := &bytes.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			MessageKey:     "message",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
		}),
		zapcore.AddSync(mockWriter),
		zapcore.InfoLevel,
	)
	Logger = zapcore.New(core)
	Sugar = Logger.Sugar()

	// 测试Info日志
	Info("测试信息日志", "key", "value")
	output := mockWriter.String()
	if !strings.Contains(output, "info") || !strings.Contains(output, "测试信息日志") {
		t.Errorf("Info日志输出不符合预期: %s", output)
	}

	// 清空缓冲区
	mockWriter.Reset()

	// 测试Error日志
	Error("测试错误日志", "error", "test error")
	output = mockWriter.String()
	if !strings.Contains(output, "error") || !strings.Contains(output, "测试错误日志") {
		t.Errorf("Error日志输出不符合预期: %s", output)
	}
}

// TestContextLogging 测试带上下文的日志功能
func TestContextLogging(t *testing.T) {
	// 初始化测试用日志
	mockWriter := &bytes.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			MessageKey:     "message",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
		}),
		zapcore.AddSync(mockWriter),
		zapcore.InfoLevel,
	)
	Logger = zapcore.New(core)

	// 创建上下文
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyRequestID, "test-req-001")
	ctx = context.WithValue(ctx, ContextKeyUserID, "test-user-001")

	// 使用带上下文的日志
	logger := WithContext(ctx)
	logger.Info("带上下文的测试日志")

	// 验证输出包含上下文信息
	output := mockWriter.String()
	if !strings.Contains(output, "test-req-001") || !strings.Contains(output, "test-user-001") {
		t.Errorf("上下文日志输出中未包含预期的上下文信息: %s", output)
	}
}

// TestEnvironmentConfiguration 测试环境变量配置
func TestEnvironmentConfiguration(t *testing.T) {
	// 保存原始的环境变量
	oldLevel := os.Getenv("LOGGER_LEVEL")
	oldFile := os.Getenv("LOGGER_FILE")
	defer func() {
		os.Setenv("LOGGER_LEVEL", oldLevel)
		os.Setenv("LOGGER_FILE", oldFile)
	}()

	// 设置测试环境变量
	testLevel := "debug"
	testFile := "/tmp/test-log.log"
	os.Setenv("LOGGER_LEVEL", testLevel)
	os.Setenv("LOGGER_FILE", testFile)

	// 这里我们不能直接测试配置是否被正确读取，因为我们无法访问内部的config包
	// 但我们可以测试初始化不会失败
	err := Init()
	if err != nil {
		t.Fatalf("使用环境变量配置初始化日志失败: %v", err)
	}
	defer Sync()

	// 清理测试文件
	if _, err := os.Stat(testFile); err == nil {
		_ = os.Remove(testFile)
	}
}

// TestGetLogger 测试GetLogger函数
func TestGetLogger(t *testing.T) {
	// 保存原始的Logger
	oldLogger := Logger
	defer func() {
		Logger = oldLogger
	}()

	// 测试当Logger为nil时的情况
	Logger = nil
	logger := GetLogger()
	if logger == nil {
		t.Error("当Logger为nil时，GetLogger应返回默认日志实例")
	}

	// 测试当Logger不为nil时的情况
	mockWriter := &bytes.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{}),
		zapcore.AddSync(mockWriter),
		zapcore.InfoLevel,
	)
	Logger = zapcore.New(core)
	logger = GetLogger()
	if logger != Logger {
		t.Error("GetLogger应返回现有的Logger实例")
	}
}

// TestGetSugar 测试GetSugar函数
func TestGetSugar(t *testing.T) {
	// 保存原始的Sugar
	oldSugar := Sugar
	defer func() {
		Sugar = oldSugar
	}()

	// 测试当Sugar为nil时的情况
	Sugar = nil
	sugar := GetSugar()
	if sugar == nil {
		t.Error("当Sugar为nil时，GetSugar应返回默认sugar实例")
	}

	// 测试当Sugar不为nil时的情况
	mockWriter := &bytes.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{}),
		zapcore.AddSync(mockWriter),
		zapcore.InfoLevel,
	)
	Logger = zapcore.New(core)
	Sugar = Logger.Sugar()
	sugar = GetSugar()
	if sugar != Sugar {
		t.Error("GetSugar应返回现有的Sugar实例")
	}
}
