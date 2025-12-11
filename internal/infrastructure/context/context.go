package context

import (
	"context"
	"github.com/google/uuid"
)

// 上下文键类型，避免与其他包冲突
type contextKey string

const (
	// RequestIDKey 请求ID上下文键
	RequestIDKey contextKey = "request_id"
	// UserIDKey 用户ID上下文键
	UserIDKey contextKey = "user_id"
	// WorkflowIDKey 工作流ID上下文键
	WorkflowIDKey contextKey = "workflow_id"
	// TraceIDKey 追踪ID上下文键
	TraceIDKey contextKey = "trace_id"
)

// WithRequestID 添加请求ID到上下文
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// RequestIDFrom 从上下文获取请求ID
func RequestIDFrom(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithUserID 添加用户ID到上下文
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// UserIDFrom 从上下文获取用户ID
func UserIDFrom(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// WithWorkflowID 添加工作流ID到上下文
func WithWorkflowID(ctx context.Context, workflowID string) context.Context {
	return context.WithValue(ctx, WorkflowIDKey, workflowID)
}

// WorkflowIDFrom 从上下文获取工作流ID
func WorkflowIDFrom(ctx context.Context) string {
	if workflowID, ok := ctx.Value(WorkflowIDKey).(string); ok {
		return workflowID
	}
	return ""
}

// WithTraceID 添加追踪ID到上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = uuid.New().String()
	}
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// TraceIDFrom 从上下文获取追踪ID
func TraceIDFrom(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// NewContext 创建包含基本上下文信息的上下文
func NewContext() context.Context {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "")
	ctx = WithRequestID(ctx, "")
	return ctx
}

// GetAllFrom 从上下文获取所有上下文信息
func GetAllFrom(ctx context.Context) map[string]string {
	return map[string]string{
		"request_id":  RequestIDFrom(ctx),
		"user_id":     UserIDFrom(ctx),
		"workflow_id": WorkflowIDFrom(ctx),
		"trace_id":    TraceIDFrom(ctx),
	}
}
