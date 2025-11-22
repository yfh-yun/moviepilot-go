package monitor

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
)

// SystemMonitor 系统监控器
type SystemMonitor struct {
	*BaseMonitor
	config MonitorConfig
	wg     sync.WaitGroup

	// 缓存
	lastMetrics  map[string]float64
	lastUpdate   time.Time
	metricsMutex sync.RWMutex // 保护 lastMetrics 和 lastUpdate 的互斥锁

	// 告警管理
	alertManager *AlertManager

	// 指标收集器
	collectors     map[MetricType]MetricCollector
	collectorMutex sync.RWMutex // 保护 collectors 的互斥锁
}

// MetricCollector 指标收集器接口
type MetricCollector interface {
	Collect() (*MonitorMetrics, error)
	GetMetricType() MetricType
}

// NewSystemMonitor 创建系统监控器
func NewSystemMonitor(config MonitorConfig, logger *zap.Logger) *SystemMonitor {
	baseMonitor := NewBaseMonitor(logger)

	sm := &SystemMonitor{
		BaseMonitor:  baseMonitor,
		config:       config,
		lastMetrics:  make(map[string]float64),
		collectors:   make(map[MetricType]MetricCollector),
		alertManager: NewAlertManager(config.AlertRules, logger),
	}

	// 注册收集器
	sm.registerCollectors()

	return sm
}

// registerCollectors 注册指标收集器
func (sm *SystemMonitor) registerCollectors() {
	for _, metric := range sm.config.Metrics {
		switch MetricType(metric) {
		case MetricTypeCPU:
			sm.collectors[MetricTypeCPU] = NewCPUCollector()
		case MetricTypeMemory:
			sm.collectors[MetricTypeMemory] = NewMemoryCollector()
		case MetricTypeDisk:
			sm.collectors[MetricTypeDisk] = NewDiskCollector()
		case MetricTypeNetwork:
			sm.collectors[MetricTypeNetwork] = NewNetworkCollector()
		case MetricTypeProcess:
			sm.collectors[MetricTypeProcess] = NewProcessCollector()
		}
	}
}

// Start 启动系统监控
func (sm *SystemMonitor) Start() error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	if sm.Running {
		return fmt.Errorf("系统监控已经在运行")
	}

	sm.Logger.Info("启动系统监控器",
		zap.Strings("metrics", sm.config.Metrics),
		zap.Duration("interval", sm.config.Interval))

	// 启动指标收集循环
	sm.wg.Add(1)
	go sm.collectLoop()

	// 启动告警检查循环
	sm.wg.Add(1)
	go sm.alertLoop()

	sm.Running = true
	return nil
}

// Stop 停止系统监控
func (sm *SystemMonitor) Stop() error {
	sm.Mu.Lock()
	if !sm.Running {
		sm.Mu.Unlock()
		return nil
	}

	sm.Logger.Info("停止系统监控器")
	sm.Running = false
	sm.Cancel()    // 取消上下文，通知所有监听器
	sm.Mu.Unlock() // 在 Wait 之前释放锁，避免死锁

	sm.wg.Wait() // 等待所有 goroutine 退出

	return nil
}

// runWithTicker 使用定时器运行指定函数，直到上下文被取消
func (sm *SystemMonitor) runWithTicker(interval time.Duration, f func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.Ctx.Done():
			return
		case <-ticker.C:
			f()
		}
	}
}

// collectLoop 指标收集循环
func (sm *SystemMonitor) collectLoop() {
	defer sm.wg.Done()
	sm.runWithTicker(sm.config.Interval, sm.collectMetrics)
}

// alertLoop 告警检查循环
func (sm *SystemMonitor) alertLoop() {
	defer sm.wg.Done()
	sm.runWithTicker(30*time.Second, sm.checkAlerts)
}

// collectMetrics 收集指标
func (sm *SystemMonitor) collectMetrics() {
	start := time.Now()

	sm.collectorMutex.RLock()
	collectors := make([]MetricCollector, 0, len(sm.collectors))
	for _, collector := range sm.collectors {
		collectors = append(collectors, collector)
	}
	sm.collectorMutex.RUnlock()

	var metricsCollected int
	var metricsProcessed int

	for _, collector := range collectors {
		metrics, err := collector.Collect()
		if err != nil {
			sm.Logger.Error("收集指标失败",
				zap.String("metric", string(collector.GetMetricType())),
				zap.Error(err))
			continue
		}
		metricsCollected++

		// 更新指标缓存
		sm.metricsMutex.Lock()
		for _, metric := range metrics.Metrics {
			sm.lastMetrics[metric.Name] = metric.Value
			metricsProcessed++
		}
		sm.lastUpdate = time.Now()
		sm.metricsMutex.Unlock()

		// 记录指标
		sm.logMetrics(metrics)
	}

	duration := time.Since(start)
	sm.Logger.Debug("指标收集完成",
		zap.Duration("duration", duration),
		zap.Int("collectors", len(collectors)),
		zap.Int("metrics_collected", metricsCollected),
		zap.Int("metrics_processed", metricsProcessed))
}

// logMetrics 记录指标
func (sm *SystemMonitor) logMetrics(metrics *MonitorMetrics) {
	for _, metric := range metrics.Metrics {
		sm.Logger.Debug("系统指标",
			zap.String("name", metric.Name),
			zap.String("type", string(metric.Type)),
			zap.Float64("value", metric.Value),
			zap.String("unit", metric.Unit),
			zap.Any("tags", metric.Tags))
	}
}

// checkAlerts 检查告警
func (sm *SystemMonitor) checkAlerts() {
	alerts := sm.alertManager.CheckAlerts(sm.lastMetrics)

	for _, alert := range alerts {
		sm.Logger.Warn("触发告警",
			zap.String("rule", alert.RuleName),
			zap.String("level", string(alert.Level)),
			zap.String("message", alert.Message),
			zap.Float64("value", alert.Value),
			zap.Float64("threshold", alert.Threshold))

		// 发送通知
		if sm.config.Notification {
			sm.sendNotification(alert)
		}
	}
}

// sendNotification 发送通知
func (sm *SystemMonitor) sendNotification(alert Alert) {
	// 这里可以集成通知系统
	sm.Logger.Info("发送告警通知",
		zap.String("alert_id", alert.ID),
		zap.String("message", alert.Message))
}

// GetSystemInfo 获取系统信息
func (sm *SystemMonitor) GetSystemInfo() (*SystemInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("获取主机信息失败: %w", err)
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("获取CPU信息失败: %w", err)
	}

	loadAvg, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("获取负载信息失败: %w", err)
	}

	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("获取进程信息失败: %w", err)
	}

	numCPU := runtime.NumCPU()
	if len(cpuInfo) > 0 {
		numCPU = int(cpuInfo[0].Cores)
	}

	return &SystemInfo{
		Hostname:   hostInfo.Hostname,
		OS:         fmt.Sprintf("%s %s", hostInfo.OS, hostInfo.PlatformVersion),
		Arch:       runtime.GOARCH,
		Uptime:     hostInfo.Uptime,
		LoadAvg:    []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15},
		NumCPU:     numCPU,
		NumProcess: len(processes),
	}, nil
}

// GetMetrics 获取当前指标的副本
func (sm *SystemMonitor) GetMetrics() map[string]float64 {
	sm.metricsMutex.RLock()
	defer sm.metricsMutex.RUnlock()

	// 创建并返回一个副本，避免外部修改影响内部状态
	result := make(map[string]float64, len(sm.lastMetrics))
	for k, v := range sm.lastMetrics {
		result[k] = v
	}
	return result
}

// UpdateConfig 更新配置
func (sm *SystemMonitor) UpdateConfig(config MonitorConfig) error {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	// 验证配置
	if config.Interval <= 0 {
		return fmt.Errorf("无效的监控间隔: %v", config.Interval)
	}

	oldConfig := sm.config
	sm.config = config

	// 重新注册收集器
	sm.registerCollectors()

	// 更新告警规则
	sm.alertManager.UpdateRules(config.AlertRules)

	sm.Logger.Info("更新系统监控配置",
		zap.Strings("old_metrics", oldConfig.Metrics),
		zap.Strings("new_metrics", config.Metrics),
		zap.Duration("interval", config.Interval))

	return nil
}
