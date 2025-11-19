// Package logger 日志系统使用示例
package logger

import (
	"context"
	"errors"
)

// ExampleBasicUsage 基础日志使用示例
func ExampleBasicUsage() {
	// 初始化日志系统（通常在应用启动时调用一次）
	_ = Init()
	defer Sync()

	// 使用不同级别的日志
	Debug("这是调试日志", "key1", "value1")
	Info("这是信息日志", "user_id", "123", "action", "login")
	Warn("这是警告日志", "deprecated", true, "reason", "API即将废弃")
	Error("这是错误日志", "error", errors.New("操作失败"), "attempts", 3)

	// 使用格式化日志
	Debugf("调试信息: %s, 值: %d", "test", 42)
	Infof("用户 %s 成功执行了操作", "admin")
	Warnf("资源使用率达到 %d%%", 85)
	Errorf("处理请求失败: %v", errors.New("网络错误"))
}

// ExampleWithContext 带上下文的日志使用示例
func ExampleWithContext() {
	// 创建包含上下文信息的context
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyRequestID, "req-123456")
	ctx = context.WithValue(ctx, ContextKeyUserID, "user-7890")
	ctx = context.WithValue(ctx, ContextKeyTraceID, "trace-abc123")

	// 获取带上下文的日志实例
	logger := WithContext(ctx)

	// 使用带上下文的日志
	logger.Debug("处理请求开始", "path", "/api/v1/users")
	logger.Info("数据库查询完成", "rows", 10, "duration", "123ms")

	// 使用格式化日志
	logger.Infof("处理用户 %s 的请求完成", "admin")
	logger.Errorf("请求处理失败: %v", errors.New("业务错误"))
}

// ExampleAPIUsage API层日志使用示例
func ExampleAPIUsage(ctx context.Context) {
	// 获取请求ID等信息（通常从请求头或中间件获取）
	requestID := "api-req-001"
	userID := "api-user-001"

	// 设置上下文
	ctx = context.WithValue(ctx, ContextKeyRequestID, requestID)
	ctx = context.WithValue(ctx, ContextKeyUserID, userID)

	// 获取带上下文的日志
	logger := WithContext(ctx)

	// API请求开始
	logger.Info("API请求开始", 
		"method", "GET", 
		"path", "/api/v1/resources", 
		"ip", "192.168.1.1")

	// 业务逻辑处理...

	// API请求结束
	logger.Info("API请求结束", 
		"status_code", 200, 
		"duration", "45ms", 
		"response_size", 1024)
}

// ExampleServiceUsage 服务层日志使用示例
func ExampleServiceUsage() {
	// 记录服务方法开始
	Debug("CreateUser服务方法开始", "func", "CreateUser")

	// 业务逻辑处理...
	userID := "new-user-001"

	// 记录关键业务节点
	Info("用户创建成功", 
		"user_id", userID,
		"username", "testuser",
		"email", "test@example.com")

	// 错误处理示例
	if err := processUserData(userID); err != nil {
		Error("用户数据处理失败", 
			"error", err.Error(),
			"user_id", userID,
			"retry_count", 2)
	}
}

// processUserData 模拟处理用户数据的辅助函数
func processUserData(userID string) error {
	// 模拟业务逻辑
	return nil // 模拟成功场景
}

// ExampleRepositoryUsage 数据访问层日志使用示例
func ExampleRepositoryUsage() {
	// 记录数据库操作
	Debug("数据库查询开始", 
		"func", "GetUserByID",
		"query", "SELECT * FROM users WHERE id = ?")

	// 执行数据库操作...
	userID := "query-user-001"

	// 记录查询结果
	Info("数据库查询完成", 
		"user_id", userID,
		"duration", "15ms",
		"result_count", 1)

	// 数据库错误处理
	if err := saveUserData(userID, map[string]interface{}{"name": "updated"}); err != nil {
		Error("数据库保存失败", 
			"error", err.Error(),
			"user_id", userID,
			"operation", "update")
	}
}

// saveUserData 模拟保存用户数据的辅助函数
func saveUserData(userID string, data map[string]interface{}) error {
	// 模拟业务逻辑
	return nil // 模拟成功场景
}
