package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TransferTotal 转移总数
	TransferTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_transfer_total",
			Help: "Total number of file transfers",
		},
		[]string{"mode", "status"},
	)

	// TransferBytesTotal 转移字节总数
	TransferBytesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_transfer_bytes_total",
			Help: "Total bytes transferred",
		},
		[]string{"mode"},
	)

	// TransferDuration 转移时长
	TransferDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moviepilot_transfer_duration_seconds",
			Help:    "File transfer duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
		},
		[]string{"mode"},
	)

	// TransferErrors 转移错误数
	TransferErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moviepilot_transfer_errors_total",
			Help: "Total number of transfer errors",
		},
		[]string{"mode", "error_type"},
	)
)

// RecordTransfer 记录转移
func RecordTransfer(mode, status string, bytes int64, durationSeconds float64) {
	TransferTotal.WithLabelValues(mode, status).Inc()
	if status == "success" {
		TransferBytesTotal.WithLabelValues(mode).Add(float64(bytes))
	}
	TransferDuration.WithLabelValues(mode).Observe(durationSeconds)
}

// RecordTransferError 记录转移错误
func RecordTransferError(mode, errorType string) {
	TransferErrors.WithLabelValues(mode, errorType).Inc()
}
