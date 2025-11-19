package monitor

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
)

// SystemMonitor 系统监控器
type SystemMonitor struct {
	config     MonitorConfig
	logger     *zap.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	mu         sync.RWMutex
	
	// 缓存
	lastMetrics map[string]float64
	lastUpdate  time.Time
	
	// 告警管理
	alertManager *AlertManager
	
	// 指标收集器
	collectors map[MetricType]MetricCollector
}

// MetricCollector 指标收集器接口
type MetricCollector interface {
	Collect() (*MonitorMetrics, error)
	GetMetricType() MetricType
}

// NewSystemMonitor 创建系统监控器
func NewSystemMonitor(config MonitorConfig, logger *zap.Logger) *SystemMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	sm := &SystemMonitor{
		config:       config,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
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
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.running {
		return fmt.Errorf("系统监控已经在运行")
	}
	
	sm.logger.Info("启动系统监控器", 
		zap.Strings("metrics", sm.config.Metrics),
		zap.Duration("interval", sm.config.Interval))
	
	// 启动指标收集循环
	sm.wg.Add(1)
	go sm.collectLoop()
	
	// 启动告警检查循环
	sm.wg.Add(1)
	go sm.alertLoop()
	
	sm.running = true
	return nil
}

// Stop 停止系统监控
func (sm *SystemMonitor) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.running {
		return nil
	}
	
	sm.logger.Info("停止系统监控器")
	
	sm.cancel()
	sm.wg.Wait()
	
	sm.running = false
	return nil
}

// collectLoop 指标收集循环
func (sm *SystemMonitor) collectLoop() {
	defer sm.wg.Done()
	
	ticker := time.NewTicker(sm.config.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.collectMetrics()
		}
	}
}

// alertLoop 告警检查循环
func (sm *SystemMonitor) alertLoop() {
	defer sm.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second) // 30秒检查一次告警
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.checkAlerts()
		}
	}
}

// collectMetrics 收集指标
func (sm *SystemMonitor) collectMetrics() {
	start := time.Now()
	
	for metricType, collector := range sm.collectors {
		metrics, err := collector.Collect()
		if err != nil {
			sm.logger.Error("收集指标失败", 
				zap.String("metric", string(metricType)),
				zap.Error(err))
			continue
		}
		
		// 缓存指标
		for _, metric := range metrics.Metrics {
			sm.lastMetrics[metric.Name] = metric.Value
		}
		
		// 记录指标
		sm.logMetrics(metrics)
	}
	
	sm.lastUpdate = time.Now()
	sm.logger.Debug("指标收集完成", 
		zap.Duration("duration", time.Since(start)),
		zap.Int("metrics", len(sm.collectors)))
}

// logMetrics 记录指标
func (sm *SystemMonitor) logMetrics(metrics *MonitorMetrics) {
	for _, metric := range metrics.Metrics {
		sm.logger.Debug("系统指标",
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
		sm.logger.Warn("触发告警",
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
	sm.logger.Info("发送告警通知",
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

// GetMetrics 获取当前指标
func (sm *SystemMonitor) GetMetrics() map[string]float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	result := make(map[string]float64)
	for k, v := range sm.lastMetrics {
		result[k] = v
	}
	return result
}

// IsRunning 检查监控器是否运行
func (sm *SystemMonitor) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.running
}

// UpdateConfig 更新配置
func (sm *SystemMonitor) UpdateConfig(config MonitorConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	oldConfig := sm.config
	sm.config = config
	
	// 重新注册收集器
	sm.registerCollectors()
	
	// 更新告警规则
	sm.alertManager.UpdateRules(config.AlertRules)
	
	sm.logger.Info("更新系统监控配置",
		zap.Strings("old_metrics", oldConfig.Metrics),
		zap.Strings("new_metrics", config.Metrics))
	
	return nil
}