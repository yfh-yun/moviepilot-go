package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus 指标定义
var (
	// HTTP 请求指标
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moviepilot_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 数据库查询指标
	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moviepilot_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// 缓存指标
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	// 订阅指标
	SubscriptionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "moviepilot_subscriptions_active",
			Help: "Number of active subscriptions",
		},
	)

	SubscriptionRefreshTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "moviepilot_subscription_refresh_total",
			Help: "Total number of subscription refreshes",
		},
	)

	// 下载指标
	DownloadsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "moviepilot_downloads_active",
			Help: "Number of active downloads",
		},
	)

	DownloadBytesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "moviepilot_download_bytes_total",
			Help: "Total bytes downloaded",
		},
	)

	DownloadSpeed = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "moviepilot_download_speed_bytes_per_second",
			Help: "Current download speed in bytes per second",
		},
	)

	// 插件指标
	PluginsLoaded = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "moviepilot_plugins_loaded",
			Help: "Number of loaded plugins",
		},
	)

	PluginExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_plugin_executions_total",
			Help: "Total number of plugin executions",
		},
		[]string{"plugin_id", "status"},
	)

	// 工作流指标
	WorkflowExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_workflow_executions_total",
			Help: "Total number of workflow executions",
		},
		[]string{"workflow_id", "status"},
	)

	WorkflowExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moviepilot_workflow_execution_duration_seconds",
			Help:    "Workflow execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"workflow_id"},
	)

	// 动作指标
	ActionExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_action_executions_total",
			Help: "Total number of action executions",
		},
		[]string{"action_name", "status"},
	)

	ActionExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moviepilot_action_execution_duration_seconds",
			Help:    "Action execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"action_name"},
	)

	// 下载动作指标
	DownloadActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_download_actions_total",
			Help: "Total number of download actions",
		},
		[]string{"status"},
	)

	DownloadTasksAdded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "moviepilot_download_tasks_added_total",
			Help: "Total number of download tasks added",
		},
	)

	DownloadTasksFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "moviepilot_download_tasks_failed_total",
			Help: "Total number of failed download tasks",
		},
	)

	// 订阅动作指标
	SubscribeActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_subscribe_actions_total",
			Help: "Total number of subscribe actions",
		},
		[]string{"status"},
	)

	SubscribeTasksAdded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "moviepilot_subscribe_tasks_added_total",
			Help: "Total number of subscribe tasks added",
		},
	)

	SubscribeTasksFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "moviepilot_subscribe_tasks_failed_total",
			Help: "Total number of failed subscribe tasks",
		},
	)

	// 系统资源指标
	MemoryUsageBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "moviepilot_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
	)

	GoroutinesActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "moviepilot_goroutines_active",
			Help: "Number of active goroutines",
		},
	)

	// gRPC 指标
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "status"},
	)

	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moviepilot_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)
)

// RecordHTTPRequest 记录 HTTP 请求
func RecordHTTPRequest(method, endpoint, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordDBQuery 记录数据库查询
func RecordDBQuery(operation, table string, duration float64) {
	DBQueriesTotal.WithLabelValues(operation, table).Inc()
	DBQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordCacheHit 记录缓存命中
func RecordCacheHit(cacheType string) {
	CacheHitsTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss 记录缓存未命中
func RecordCacheMiss(cacheType string) {
	CacheMissesTotal.WithLabelValues(cacheType).Inc()
}

// RecordWorkflowExecution 记录工作流执行
func RecordWorkflowExecution(workflowID, status string, duration float64) {
	WorkflowExecutionsTotal.WithLabelValues(workflowID, status).Inc()
	WorkflowExecutionDuration.WithLabelValues(workflowID).Observe(duration)
}

// RecordGRPCRequest 记录 gRPC 请求
func RecordGRPCRequest(service, method, status string, duration float64) {
	GRPCRequestsTotal.WithLabelValues(service, method, status).Inc()
	GRPCRequestDuration.WithLabelValues(service, method).Observe(duration)
}

// RecordActionExecution 记录动作执行
func RecordActionExecution(actionName, status string, duration float64) {
	ActionExecutionsTotal.WithLabelValues(actionName, status).Inc()
	ActionExecutionDuration.WithLabelValues(actionName).Observe(duration)
}

// RecordDownloadAction 记录下载动作执行
func RecordDownloadAction(status string) {
	DownloadActionsTotal.WithLabelValues(status).Inc()
}

// RecordDownloadTaskAdded 记录成功添加的下载任务
func RecordDownloadTaskAdded() {
	DownloadTasksAdded.Inc()
}

// RecordDownloadTaskFailed 记录失败的下载任务
func RecordDownloadTaskFailed() {
	DownloadTasksFailed.Inc()
}

// RecordSubscribeAction 记录订阅动作执行
func RecordSubscribeAction(status string) {
	SubscribeActionsTotal.WithLabelValues(status).Inc()
}

// RecordSubscribeTaskAdded 记录成功添加的订阅任务
func RecordSubscribeTaskAdded() {
	SubscribeTasksAdded.Inc()
}

// RecordSubscribeTaskFailed 记录失败的订阅任务
func RecordSubscribeTaskFailed() {
	SubscribeTasksFailed.Inc()
}
