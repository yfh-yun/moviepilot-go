package models

import (
	"time"
)

// RequestMetrics 请求指标模型
type RequestMetrics struct {
	// 路径
	Path string `json:"path"`
	// 方法
	Method string `json:"method"`
	// 状态码
	StatusCode int `json:"status_code"`
	// 响应时间
	ResponseTime float64 `json:"response_time"`
	// 时间�?	Timestamp time.Time `json:"timestamp"`
	// 客户端IP
	ClientIP string `json:"client_ip"`
	// 用户代理
	UserAgent string `json:"user_agent"`
}

// PerformanceSnapshot 性能快照模型
type PerformanceSnapshot struct {
	// 时间�?	Timestamp time.Time `json:"timestamp"`
	// CPU使用�?	CPUUsage float64 `json:"cpu_usage"`
	// 内存使用�?	MemoryUsage float64 `json:"memory_usage"`
	// 活跃请求�?	ActiveRequests int `json:"active_requests"`
	// 请求速率
	RequestRate float64 `json:"request_rate"`
	// 平均响应时间
	AvgResponseTime float64 `json:"avg_response_time"`
	// 错误�?	ErrorRate float64 `json:"error_rate"`
	// 慢请�?	SlowRequests int `json:"slow_requests"`
}

// EndpointStats 端点统计模型
type EndpointStats struct {
	// 端点
	Endpoint string `json:"endpoint"`
	// 计数
	Count int `json:"count"`
	// 总时�?	TotalTime float64 `json:"total_time"`
	// 错误�?	Errors int `json:"errors"`
	// 平均时间
	AvgTime float64 `json:"avg_time"`
}

// ErrorRequest 错误请求模型
type ErrorRequest struct {
	// 时间�?	Timestamp string `json:"timestamp"`
	// 方法
	Method string `json:"method"`
	// 路径
	Path string `json:"path"`
	// 状态码
	StatusCode int `json:"status_code"`
	// 响应时间
	ResponseTime float64 `json:"response_time"`
	// 客户端IP
	ClientIP string `json:"client_ip"`
}

// MonitoringOverview 监控概览模型
type MonitoringOverview struct {
	// 性能
	Performance PerformanceSnapshot `json:"performance"`
	// 顶级端点
	TopEndpoints []EndpointStats `json:"top_endpoints"`
	// 最近错�?	RecentErrors []ErrorRequest `json:"recent_errors"`
	// 警报
	Alerts []string `json:"alerts"`
}

// MonitoringConfig 监控配置模型
type MonitoringConfig struct {
	// 慢请求阈�?	SlowRequestThreshold float64 `json:"slow_request_threshold"`
	// 错误阈�?	ErrorThreshold float64 `json:"error_threshold"`
	// CPU阈�?	CPUThreshold float64 `json:"cpu_threshold"`
	// 内存阈�?	MemoryThreshold float64 `json:"memory_threshold"`
	// 最大历史记�?	MaxHistory int `json:"max_history"`
	// 窗口大小
	WindowSize int `json:"window_size"`
}

// NewRequestMetrics 创建一个新�?RequestMetrics 实例
func NewRequestMetrics() *RequestMetrics {
	return &RequestMetrics{}
}

// NewPerformanceSnapshot 创建一个新�?PerformanceSnapshot 实例
func NewPerformanceSnapshot() *PerformanceSnapshot {
	return &PerformanceSnapshot{}
}

// NewEndpointStats 创建一个新�?EndpointStats 实例
func NewEndpointStats() *EndpointStats {
	return &EndpointStats{}
}

// NewErrorRequest 创建一个新�?ErrorRequest 实例
func NewErrorRequest() *ErrorRequest {
	return &ErrorRequest{}
}

// NewMonitoringOverview 创建一个新�?MonitoringOverview 实例
func NewMonitoringOverview() *MonitoringOverview {
	return &MonitoringOverview{
		TopEndpoints: make([]EndpointStats, 0),
		RecentErrors: make([]ErrorRequest, 0),
		Alerts:       make([]string, 0),
	}
}

// NewMonitoringConfig 创建一个新�?MonitoringConfig 实例
func NewMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		SlowRequestThreshold: 1.0,
		ErrorThreshold:       0.05,
		CPUThreshold:         80.0,
		MemoryThreshold:      80.0,
		MaxHistory:           1000,
		WindowSize:           60,
	}
}
