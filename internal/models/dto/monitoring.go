package dto

import "time"

// RequestMetrics 请求指标模型
type RequestMetrics struct {
	Path         string    `json:"path"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	ResponseTime float64   `json:"response_time"`
	Timestamp    time.Time `json:"timestamp"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
}

// PerformanceSnapshot 性能快照模型
type PerformanceSnapshot struct {
	Timestamp       time.Time `json:"timestamp"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	ActiveRequests  int       `json:"active_requests"`
	RequestRate     float64   `json:"request_rate"`
	AvgResponseTime float64   `json:"avg_response_time"`
	ErrorRate       float64   `json:"error_rate"`
	SlowRequests    int       `json:"slow_requests"`
}

// EndpointStats 端点统计模型
type EndpointStats struct {
	Endpoint  string  `json:"endpoint"`
	Count     int     `json:"count"`
	TotalTime float64 `json:"total_time"`
	Errors    int     `json:"errors"`
	AvgTime   float64 `json:"avg_time"`
}

// ErrorRequest 错误请求模型
type ErrorRequest struct {
	Timestamp    string  `json:"timestamp"`
	Method       string  `json:"method"`
	Path         string  `json:"path"`
	StatusCode   int     `json:"status_code"`
	ResponseTime float64 `json:"response_time"`
	ClientIP     string  `json:"client_ip"`
}

// MonitoringOverview 监控概览模型
type MonitoringOverview struct {
	Performance  PerformanceSnapshot `json:"performance"`
	TopEndpoints []EndpointStats     `json:"top_endpoints"`
	RecentErrors []ErrorRequest      `json:"recent_errors"`
	Alerts       []string            `json:"alerts"`
}

// MonitoringConfig 监控配置模型
type MonitoringConfig struct {
	// 慢请求阈值（秒）
	SlowRequestThreshold float64 `json:"slow_request_threshold,omitempty"`
	// 错误率阈值
	ErrorThreshold float64 `json:"error_threshold,omitempty"`
	// CPU使用率阈值
	CPUThreshold float64 `json:"cpu_threshold,omitempty"`
	// 内存使用率阈值
	MemoryThreshold float64 `json:"memory_threshold,omitempty"`
	// 最大历史记录数
	MaxHistory int `json:"max_history,omitempty"`
	// 时间窗口大小（秒）
	WindowSize int `json:"window_size,omitempty"`
}

// DefaultMonitoringConfig 返回默认监控配置
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		SlowRequestThreshold: 1.0,
		ErrorThreshold:       0.05,
		CPUThreshold:         80.0,
		MemoryThreshold:      80.0,
		MaxHistory:           1000,
		WindowSize:           60,
	}
}
