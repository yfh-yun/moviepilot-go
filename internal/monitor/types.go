package monitor

import (
	"time"
)

// MonitorType 监控类型
type MonitorType string

const (
	MonitorTypeSystem     MonitorType = "system"     // 系统监控
	MonitorTypeService    MonitorType = "service"    // 服务监控
	MonitorTypeCustom     MonitorType = "custom"     // 自定义监控
	MonitorTypeBrowser    MonitorType = "browser"    // 浏览器监控
)

// MetricType 指标类型
type MetricType string

const (
	MetricTypeCPU     MetricType = "cpu"
	MetricTypeMemory  MetricType = "memory"
	MetricTypeDisk    MetricType = "disk"
	MetricTypeNetwork MetricType = "network"
	MetricTypeProcess MetricType = "process"
	MetricTypeCustom  MetricType = "custom"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// MonitorConfig 监控配置
type MonitorConfig struct {
	Enabled         bool          `json:"enabled"`
	Interval        time.Duration `json:"interval"`
	Retention       time.Duration `json:"retention"`
	Metrics         []string      `json:"metrics"`
	AlertRules      []AlertRule   `json:"alert_rules"`
	Notification    bool          `json:"notification"`
	WebhookURL      string        `json:"webhook_url,omitempty"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Metric      string      `json:"metric"`
	Operator    string      `json:"operator"`    // >, <, >=, <=, ==, !=
	Threshold   float64     `json:"threshold"`
	Level       AlertLevel  `json:"level"`
	Duration    time.Duration `json:"duration"`
	Enabled     bool        `json:"enabled"`
	Description string      `json:"description"`
}

// MetricData 指标数据
type MetricData struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit"`
	Tags      map[string]string     `json:"tags"`
	Timestamp time.Time              `json:"timestamp"`
	Labels    map[string]interface{} `json:"labels"`
}

// MonitorMetrics 监控指标集合
type MonitorMetrics struct {
	Source    string       `json:"source"`
	Host      string       `json:"host"`
	Timestamp time.Time    `json:"timestamp"`
	Metrics   []MetricData `json:"metrics"`
}

// Alert 告警信息
type Alert struct {
	ID          string            `json:"id"`
	RuleID      string            `json:"rule_id"`
	RuleName    string            `json:"rule_name"`
	Level       AlertLevel        `json:"level"`
	Message     string            `json:"message"`
	Value       float64           `json:"value"`
	Threshold   float64           `json:"threshold"`
	Metric      string            `json:"metric"`
	Tags        map[string]string `json:"tags"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     *time.Time        `json:"end_time,omitempty"`
	Status      string            `json:"status"` // firing, resolved
	Annotations map[string]string `json:"annotations"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Hostname   string    `json:"hostname"`
	OS         string    `json:"os"`
	Arch       string    `json:"arch"`
	Uptime     uint64    `json:"uptime"`
	LoadAvg    []float64 `json:"load_avg"`     // 1min, 5min, 15min
	NumCPU     int       `json:"num_cpu"`
	NumProcess int       `json:"num_process"`
}

// CPUInfo CPU信息
type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
	User        float64 `json:"user"`
	System      float64 `json:"system"`
	Idle        float64 `json:"idle"`
	Wait        float64 `json:"wait"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	Buffers     uint64  `json:"buffers"`
	Cached      uint64  `json:"cached"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Mountpoint  string  `json:"mountpoint"`
	Device      string  `json:"device"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
	FSType      string  `json:"fs_type"`
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	Interface string `json:"interface"`
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrorsIn  uint64 `json:"errors_in"`
	ErrorsOut uint64 `json:"errors_out"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID        int     `json:"pid"`
	Name       string  `json:"name"`
	CmdLine    string  `json:"cmd_line"`
	Status     string  `json:"status"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float64 `json:"mem_percent"`
	MemoryRSS  uint64  `json:"memory_rss"`
	NumThreads int     `json:"num_threads"`
	CreateTime int64   `json:"create_time"`
}

// BrowserMetrics 浏览器指标
type BrowserMetrics struct {
	BrowserID     string            `json:"browser_id"`
	IsRunning     bool              `json:"is_running"`
	WindowTitle   string            `json:"window_title"`
	URL           string            `json:"url"`
	MemoryUsage   uint64            `json:"memory_usage"`
	CPUUsage      float64           `json:"cpu_usage"`
	NetworkStats  BrowserNetwork    `json:"network_stats"`
	Javascript    BrowserJS         `json:"javascript"`
	Performance   BrowserPerformance `json:"performance"`
}

// BrowserNetwork 浏览器网络信息
type BrowserNetwork struct {
	RequestsCount    int   `json:"requests_count"`
	TotalBytes       uint64 `json:"total_bytes"`
	FailedRequests   int   `json:"failed_requests"`
	AverageLatency   float64 `json:"average_latency"`
}

// BrowserJS 浏览器JS信息
type BrowserJS struct {
	JSHeapUsedSize     uint64 `json:"js_heap_used_size"`
	JSHeapTotalSize    uint64 `json:"js_heap_total_size"`
	JSHeapSizeLimit    uint64 `json:"js_heap_size_limit"`
	DOMNodes          int    `json:"dom_nodes"`
	JSEventListeners  int    `json:"js_event_listeners"`
}

// BrowserPerformance 浏览器性能信息
type BrowserPerformance struct {
	FCP     float64 `json:"fcp"`     // First Contentful Paint
	LCP     float64 `json:"lcp"`     // Largest Contentful Paint
	FID     float64 `json:"fid"`     // First Input Delay
	CLS     float64 `json:"cls"`     // Cumulative Layout Shift
	TTFB    float64 `json:"ttfb"`    // Time to First Byte
}

// CloudflareInfo Cloudflare信息
type CloudflareInfo struct {
	UserAgent    string            `json:"user_agent"`
	Cookies      map[string]string `json:"cookies"`
	Headers      map[string]string `json:"headers"`
	ChallengeDetected bool          `json:"challenge_detected"`
	ChallengeType    string        `json:"challenge_type"`
	Solved         bool             `json:"solved"`
	SolveTime      time.Duration    `json:"solve_time"`
	RequiredAction string           `json:"required_action"`
}

// BrowserAction 浏览器动作
type BrowserAction struct {
	Type        string                 `json:"type"`
	URL         string                 `json:"url,omitempty"`
	Selector    string                 `json:"selector,omitempty"`
	Value       string                 `json:"value,omitempty"`
	WaitTime    time.Duration          `json:"wait_time,omitempty"`
	Screenshot  bool                   `json:"screenshot,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
}

// BrowserActionResult 浏览器动作结果
type BrowserActionResult struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Screenshot  string                 `json:"screenshot,omitempty"`
	HTML        string                 `json:"html,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
}