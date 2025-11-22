package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ActionExecutionTotal Action 执行总数
	ActionExecutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_action_execution_total",
			Help: "Total number of action executions",
		},
		[]string{"action_name", "status"},
	)

	// ActionExecutionDuration Action 执行时长
	ActionExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moviepilot_action_execution_duration_seconds",
			Help:    "Action execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"action_name"},
	)

	// ActionErrors Action 错误数
	ActionErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_action_errors_total",
			Help: "Total number of action errors",
		},
		[]string{"action_name", "error_type"},
	)
)

// RecordActionExecution 记录 Action 执行
func RecordActionExecution(actionName, status string, duration time.Duration) {
	ActionExecutionTotal.WithLabelValues(actionName, status).Inc()
	ActionExecutionDuration.WithLabelValues(actionName).Observe(duration.Seconds())
}

// RecordActionError 记录 Action 错误
func RecordActionError(actionName, errorType string) {
	ActionErrors.WithLabelValues(actionName, errorType).Inc()
}
