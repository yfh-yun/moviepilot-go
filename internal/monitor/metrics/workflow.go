package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WorkflowExecutionTotal Workflow 执行总数
	WorkflowExecutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_workflow_execution_total",
			Help: "Total number of workflow executions",
		},
		[]string{"workflow_type", "status"},
	)

	// WorkflowExecutionDuration Workflow 执行时长
	WorkflowExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moviepilot_workflow_execution_duration_seconds",
			Help:    "Workflow execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"workflow_type"},
	)

	// WorkflowFilesProcessed Workflow 处理的文件数
	WorkflowFilesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_workflow_files_processed_total",
			Help: "Total number of files processed by workflows",
		},
		[]string{"workflow_type"},
	)

	// WorkflowErrors Workflow 错误数
	WorkflowErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_workflow_errors_total",
			Help: "Total number of workflow errors",
		},
		[]string{"workflow_type", "error_type"},
	)
)

// RecordWorkflowExecution 记录 Workflow 执行
func RecordWorkflowExecution(workflowType, status string, duration time.Duration) {
	WorkflowExecutionTotal.WithLabelValues(workflowType, status).Inc()
	WorkflowExecutionDuration.WithLabelValues(workflowType).Observe(duration.Seconds())
}

// RecordFilesProcessed 记录处理的文件数
func RecordFilesProcessed(workflowType string, count int) {
	WorkflowFilesProcessed.WithLabelValues(workflowType).Add(float64(count))
}

// RecordWorkflowError 记录 Workflow 错误
func RecordWorkflowError(workflowType, errorType string) {
	WorkflowErrors.WithLabelValues(workflowType, errorType).Inc()
}
