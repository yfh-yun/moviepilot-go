package monitoring

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	registry   *prometheus.Registry
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	summaries  map[string]*prometheus.SummaryVec
	mu         sync.RWMutex
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		registry:   prometheus.NewRegistry(),
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
	}
}

// RegisterCounter 注册计数器
func (m *MetricsCollector) RegisterCounter(name, help string, labels []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.counters[name]; exists {
		return fmt.Errorf("counter %s already registered", name)
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	if err := m.registry.Register(counter); err != nil {
		return fmt.Errorf("failed to register counter %s: %w", name, err)
	}

	m.counters[name] = counter
	return nil
}

// RegisterGauge 注册仪表
func (m *MetricsCollector) RegisterGauge(name, help string, labels []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.gauges[name]; exists {
		return fmt.Errorf("gauge %s already registered", name)
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)

	if err := m.registry.Register(gauge); err != nil {
		return fmt.Errorf("failed to register gauge %s: %w", name, err)
	}

	m.gauges[name] = gauge
	return nil
}

// RegisterHistogram 注册直方图
func (m *MetricsCollector) RegisterHistogram(name, help string, labels []string, buckets []float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.histograms[name]; exists {
		return fmt.Errorf("histogram %s already registered", name)
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)

	if err := m.registry.Register(histogram); err != nil {
		return fmt.Errorf("failed to register histogram %s: %w", name, err)
	}

	m.histograms[name] = histogram
	return nil
}

// RegisterSummary 注册摘要
func (m *MetricsCollector) RegisterSummary(name, help string, labels []string, objectives map[float64]float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.summaries[name]; exists {
		return fmt.Errorf("summary %s already registered", name)
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)

	if err := m.registry.Register(summary); err != nil {
		return fmt.Errorf("failed to register summary %s: %w", name, err)
	}

	m.summaries[name] = summary
	return nil
}

// IncrementCounter 增加计数器值
func (m *MetricsCollector) IncrementCounter(name string, value float64, labels map[string]string) error {
	m.mu.RLock()
	counter, exists := m.counters[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("counter %s not registered", name)
	}

	counter.With(labels).Add(value)
	return nil
}

// SetGauge 设置仪表值
func (m *MetricsCollector) SetGauge(name string, value float64, labels map[string]string) error {
	m.mu.RLock()
	gauge, exists := m.gauges[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("gauge %s not registered", name)
	}

	gauge.With(labels).Set(value)
	return nil
}

// ObserveHistogram 观察直方图值
func (m *MetricsCollector) ObserveHistogram(name string, value float64, labels map[string]string) error {
	m.mu.RLock()
	histogram, exists := m.histograms[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("histogram %s not registered", name)
	}

	histogram.With(labels).Observe(value)
	return nil
}

// ObserveSummary 观察摘要值
func (m *MetricsCollector) ObserveSummary(name string, value float64, labels map[string]string) error {
	m.mu.RLock()
	summary, exists := m.summaries[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("summary %s not registered", name)
	}

	summary.With(labels).Observe(value)
	return nil
}

// GetHandler 获取HTTP处理程序
func (m *MetricsCollector) GetHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		Registry: m.registry,
	})
}

// SetupDefaultMetrics 设置默认指标
func (m *MetricsCollector) SetupDefaultMetrics() error {
	// HTTP请求相关指标
	err := m.RegisterCounter(
		"http_requests_total",
		"Total number of HTTP requests by method, status and path.",
		[]string{"method", "status", "path"},
	)
	if err != nil {
		return err
	}

	err = m.RegisterHistogram(
		"http_request_duration_seconds",
		"HTTP request duration in seconds.",
		[]string{"method", "path"},
		prometheus.DefBuckets,
	)
	if err != nil {
		return err
	}

	err = m.RegisterGauge(
		"http_requests_in_progress",
		"Current number of HTTP requests in progress.",
		[]string{"method", "path"},
	)
	if err != nil {
		return err
	}

	// 业务相关指标
	err = m.RegisterCounter(
		"moviepilot_tasks_total",
		"Total number of tasks processed by type and status.",
		[]string{"task_type", "status"},
	)
	if err != nil {
		return err
	}

	err = m.RegisterGauge(
		"moviepilot_tasks_in_progress",
		"Current number of tasks in progress.",
		[]string{"task_type"},
	)
	if err != nil {
		return err
	}

	err = m.RegisterHistogram(
		"moviepilot_task_duration_seconds",
		"Task processing duration in seconds.",
		[]string{"task_type"},
		[]float64{0.1, 0.5, 1, 5, 10, 30, 60},
	)
	if err != nil {
		return err
	}

	// 通知相关指标
	err = m.RegisterCounter(
		"moviepilot_notifications_total",
		"Total number of notifications sent by channel and status.",
		[]string{"channel", "status"},
	)
	if err != nil {
		return err
	}

	err = m.RegisterGauge(
		"moviepilot_notifications_queue_size",
		"Current size of notification queue.",
		[]string{"priority"},
	)
	if err != nil {
		return err
	}

	// 数据库相关指标
	err = m.RegisterGauge(
		"moviepilot_database_connections",
		"Current number of database connections.",
		[]string{"state"},
	)
	if err != nil {
		return err
	}

	err = m.RegisterHistogram(
		"moviepilot_database_query_duration_seconds",
		"Database query duration in seconds.",
		[]string{"table", "operation"},
		[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
	)
	if err != nil {
		return err
	}

	// 缓存相关指标
	err = m.RegisterGauge(
		"moviepilot_cache_hit_rate",
		"Cache hit rate percentage.",
		[]string{"cache_type"},
	)
	if err != nil {
		return err
	}

	err = m.RegisterCounter(
		"moviepilot_cache_operations_total",
		"Total number of cache operations by type and result.",
		[]string{"cache_type", "operation", "result"},
	)
	if err != nil {
		return err
	}

	return nil
}

// HTTPRequestMiddleware HTTP请求指标中间件
func (m *MetricsCollector) HTTPRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		// 记录正在进行的请求
		m.SetGauge("http_requests_in_progress", 1, map[string]string{
			"method": method,
			"path":   path,
		})

		// 包装ResponseWriter以捕获状态码
		wrapper := &responseWrapper{w: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapper, r)

		// 记录请求完成
		duration := time.Since(start).Seconds()

		m.IncrementCounter("http_requests_total", 1, map[string]string{
			"method": method,
			"status": fmt.Sprintf("%d", wrapper.statusCode),
			"path":   path,
		})

		m.ObserveHistogram("http_request_duration_seconds", duration, map[string]string{
			"method": method,
			"path":   path,
		})

		m.SetGauge("http_requests_in_progress", 0, map[string]string{
			"method": method,
			"path":   path,
		})
	})
}

// responseWrapper 包装ResponseWriter以捕获状态码
type responseWrapper struct {
	w          http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) Header() http.Header {
	return rw.w.Header()
}

func (rw *responseWrapper) Write(data []byte) (int, error) {
	return rw.w.Write(data)
}

func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.w.WriteHeader(statusCode)
}
