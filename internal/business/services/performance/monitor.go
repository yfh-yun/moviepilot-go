package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// MonitorService 性能监控服务
type MonitorService interface {
	// GetMetrics 获取性能指标
	GetMetrics(ctx context.Context) (*Metrics, error)

	// GetHistory 获取历史数据
	GetHistory(ctx context.Context, duration time.Duration) ([]*MetricsSnapshot, error)

	// StartMonitoring 开始监控
	StartMonitoring(ctx context.Context, interval time.Duration)

	// StopMonitoring 停止监控
	StopMonitoring()
}

// monitorService 监控服务实现
type monitorService struct {
	history    []*MetricsSnapshot
	maxHistory int
	mutex      sync.RWMutex
	stopChan   chan struct{}
	logger     *zap.Logger
}

// NewMonitorService 创建监控服务
func NewMonitorService() MonitorService {
	return &monitorService{
		history:    make([]*MetricsSnapshot, 0),
		maxHistory: 1000, // 保留最近1000个快照
		logger:     logger.GetLogger(),
	}
}

// Metrics 性能指标
type Metrics struct {
	// CPU
	CPUUsage     float64 `json:"cpu_usage"`     // CPU使用率（百分比）
	NumCPU       int     `json:"num_cpu"`       // CPU核心数
	NumGoroutine int     `json:"num_goroutine"` // Goroutine数量

	// 内存
	MemoryAlloc      uint64  `json:"memory_alloc"`       // 已分配内存（字节）
	MemoryTotalAlloc uint64  `json:"memory_total_alloc"` // 总分配内存（字节）
	MemorySys        uint64  `json:"memory_sys"`         // 系统内存（字节）
	MemoryUsage      float64 `json:"memory_usage"`       // 内存使用率（百分比）

	// GC
	NumGC        uint32 `json:"num_gc"`         // GC次数
	GCPauseTotal uint64 `json:"gc_pause_total"` // GC暂停总时间（纳秒）
	LastGCTime   uint64 `json:"last_gc_time"`   // 最后GC时间

	// 系统
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	Metrics
	ID string `json:"id"`
}

// GetMetrics 获取性能指标
func (m *monitorService) GetMetrics(ctx context.Context) (*Metrics, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := &Metrics{
		NumCPU:           runtime.NumCPU(),
		NumGoroutine:     runtime.NumGoroutine(),
		MemoryAlloc:      memStats.Alloc,
		MemoryTotalAlloc: memStats.TotalAlloc,
		MemorySys:        memStats.Sys,
		MemoryUsage:      float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		NumGC:            memStats.NumGC,
		GCPauseTotal:     memStats.PauseTotalNs,
		LastGCTime:       memStats.LastGC,
		Timestamp:        time.Now(),
	}

	return metrics, nil
}

// GetHistory 获取历史数据
func (m *monitorService) GetHistory(ctx context.Context, duration time.Duration) ([]*MetricsSnapshot, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cutoff := time.Now().Add(-duration)
	result := make([]*MetricsSnapshot, 0)

	for _, snapshot := range m.history {
		if snapshot.Timestamp.After(cutoff) {
			result = append(result, snapshot)
		}
	}

	return result, nil
}

// StartMonitoring 开始监控
func (m *monitorService) StartMonitoring(ctx context.Context, interval time.Duration) {
	m.logger.Info("开始性能监控", zap.Duration("interval", interval))

	m.stopChan = make(chan struct{})
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				metrics, err := m.GetMetrics(ctx)
				if err != nil {
					m.logger.Error("获取性能指标失败", zap.Error(err))
					continue
				}

				snapshot := &MetricsSnapshot{
					Metrics: *metrics,
					ID:      time.Now().Format("20060102150405"),
				}

				m.addSnapshot(snapshot)

			case <-m.stopChan:
				ticker.Stop()
				m.logger.Info("性能监控已停止")
				return
			}
		}
	}()
}

// StopMonitoring 停止监控
func (m *monitorService) StopMonitoring() {
	if m.stopChan != nil {
		close(m.stopChan)
	}
}

// addSnapshot 添加快照
func (m *monitorService) addSnapshot(snapshot *MetricsSnapshot) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.history = append(m.history, snapshot)

	// 保持历史记录在限制内
	if len(m.history) > m.maxHistory {
		m.history = m.history[len(m.history)-m.maxHistory:]
	}
}

// FormatBytes 格式化字节数
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes >= GB {
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	} else if bytes >= MB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	} else if bytes >= KB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	}
	return fmt.Sprintf("%d B", bytes)
}
