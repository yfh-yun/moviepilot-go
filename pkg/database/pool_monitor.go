package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	appLogger "moviepilot-go/pkg/logger"
)

// PoolMonitor 连接池监控器
type PoolMonitor struct {
	db       *gorm.DB
	logger   *zap.Logger
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup

	// 统计数据
	mu              sync.RWMutex
	lastStats       map[string]any
	alertThresholds AlertThresholds
}

// AlertThresholds 告警阈值
type AlertThresholds struct {
	UtilizationWarning  float64 // 连接池利用率告警阈值（默认80%）
	UtilizationCritical float64 // 连接池利用率严重告警阈值（默认95%）
	WaitCountWarning    int64   // 等待次数告警阈值（默认1000）
	WaitDurationWarning int64   // 等待时间告警阈值（毫秒，默认1000ms）
}

// DefaultAlertThresholds 默认告警阈值
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		UtilizationWarning:  80.0,
		UtilizationCritical: 95.0,
		WaitCountWarning:    1000,
		WaitDurationWarning: 1000,
	}
}

// NewPoolMonitor 创建连接池监控器
func NewPoolMonitor(db *gorm.DB, interval time.Duration) *PoolMonitor {
	return &PoolMonitor{
		db:              db,
		logger:          appLogger.GetLogger(),
		interval:        interval,
		stopCh:          make(chan struct{}),
		alertThresholds: DefaultAlertThresholds(),
	}
}

// Start 启动监控
func (m *PoolMonitor) Start() {
	m.wg.Add(1)
	go m.monitor()
	m.logger.Info("连接池监控已启动", zap.Duration("interval", m.interval))
}

// Stop 停止监控
func (m *PoolMonitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
	m.logger.Info("连接池监控已停止")
}

// monitor 监控循环
func (m *PoolMonitor) monitor() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectAndAnalyze()
		case <-m.stopCh:
			return
		}
	}
}

// collectAndAnalyze 收集并分析统计数据
func (m *PoolMonitor) collectAndAnalyze() {
	stats, err := GetConnectionPoolStats(m.db)
	if err != nil {
		m.logger.Error("获取连接池统计失败", zap.Error(err))
		return
	}

	m.mu.Lock()
	m.lastStats = stats
	m.mu.Unlock()

	// 记录统计信息
	m.logStats(stats)

	// 检查告警
	m.checkAlerts(stats)
}

// logStats 记录统计信息
func (m *PoolMonitor) logStats(stats map[string]any) {
	m.logger.Debug("连接池统计",
		zap.Int("max_open", stats["max_open_connections"].(int)),
		zap.Int("open", stats["open_connections"].(int)),
		zap.Int("in_use", stats["in_use"].(int)),
		zap.Int("idle", stats["idle"].(int)),
		zap.Int64("wait_count", stats["wait_count"].(int64)),
		zap.Int64("wait_duration_ms", stats["wait_duration_ms"].(int64)),
		zap.Float64("utilization", stats["utilization_percentage"].(float64)))
}

// checkAlerts 检查告警
func (m *PoolMonitor) checkAlerts(stats map[string]any) {
	utilization := stats["utilization_percentage"].(float64)
	waitCount := stats["wait_count"].(int64)
	waitDuration := stats["wait_duration_ms"].(int64)

	// 检查连接池利用率
	if utilization >= m.alertThresholds.UtilizationCritical {
		m.logger.Error("连接池利用率严重告警",
			zap.Float64("utilization", utilization),
			zap.Float64("threshold", m.alertThresholds.UtilizationCritical),
			zap.Int("in_use", stats["in_use"].(int)),
			zap.Int("max_open", stats["max_open_connections"].(int)))
	} else if utilization >= m.alertThresholds.UtilizationWarning {
		m.logger.Warn("连接池利用率告警",
			zap.Float64("utilization", utilization),
			zap.Float64("threshold", m.alertThresholds.UtilizationWarning),
			zap.Int("in_use", stats["in_use"].(int)),
			zap.Int("max_open", stats["max_open_connections"].(int)))
	}

	// 检查等待次数
	if waitCount >= m.alertThresholds.WaitCountWarning {
		m.logger.Warn("连接池等待次数告警",
			zap.Int64("wait_count", waitCount),
			zap.Int64("threshold", m.alertThresholds.WaitCountWarning))
	}

	// 检查等待时间
	if waitDuration >= m.alertThresholds.WaitDurationWarning {
		m.logger.Warn("连接池等待时间告警",
			zap.Int64("wait_duration_ms", waitDuration),
			zap.Int64("threshold_ms", m.alertThresholds.WaitDurationWarning))
	}
}

// GetLastStats 获取最后一次统计数据
func (m *PoolMonitor) GetLastStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastStats
}

// GetHealthStatus 获取健康状态
func (m *PoolMonitor) GetHealthStatus() HealthStatus {
	m.mu.RLock()
	stats := m.lastStats
	m.mu.RUnlock()

	if stats == nil {
		return HealthStatus{
			Status:  "unknown",
			Message: "暂无统计数据",
		}
	}

	utilization := stats["utilization_percentage"].(float64)
	waitCount := stats["wait_count"].(int64)

	if utilization >= m.alertThresholds.UtilizationCritical {
		return HealthStatus{
			Status:  "critical",
			Message: fmt.Sprintf("连接池利用率过高: %.2f%%", utilization),
			Details: stats,
		}
	}

	if utilization >= m.alertThresholds.UtilizationWarning || waitCount >= m.alertThresholds.WaitCountWarning {
		return HealthStatus{
			Status:  "warning",
			Message: fmt.Sprintf("连接池利用率: %.2f%%, 等待次数: %d", utilization, waitCount),
			Details: stats,
		}
	}

	return HealthStatus{
		Status:  "healthy",
		Message: fmt.Sprintf("连接池运行正常, 利用率: %.2f%%", utilization),
		Details: stats,
	}
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status  string         `json:"status"`  // healthy, warning, critical, unknown
	Message string         `json:"message"` // 状态描述
	Details map[string]any `json:"details"` // 详细统计
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	StartTime       time.Time        `json:"start_time"`
	EndTime         time.Time        `json:"end_time"`
	Duration        time.Duration    `json:"duration"`
	AvgUtilization  float64          `json:"avg_utilization"`
	MaxUtilization  float64          `json:"max_utilization"`
	TotalWaitCount  int64            `json:"total_wait_count"`
	TotalWaitTime   time.Duration    `json:"total_wait_time"`
	ConnectionLeaks int              `json:"connection_leaks"`
	Recommendations []string         `json:"recommendations"`
	DetailedStats   []map[string]any `json:"detailed_stats"`
}

// GenerateReport 生成性能报告
func (m *PoolMonitor) GenerateReport(ctx context.Context, duration time.Duration) (*PerformanceReport, error) {
	startTime := time.Now()
	endTime := startTime.Add(duration)

	report := &PerformanceReport{
		StartTime:     startTime,
		EndTime:       endTime,
		Duration:      duration,
		DetailedStats: make([]map[string]any, 0),
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var totalUtilization float64
	var samples int
	var maxUtil float64

	for {
		select {
		case <-ticker.C:
			stats, err := GetConnectionPoolStats(m.db)
			if err != nil {
				m.logger.Error("获取统计失败", zap.Error(err))
				continue
			}

			utilization := stats["utilization_percentage"].(float64)
			totalUtilization += utilization
			samples++

			if utilization > maxUtil {
				maxUtil = utilization
			}

			report.DetailedStats = append(report.DetailedStats, stats)
			report.TotalWaitCount = stats["wait_count"].(int64)
			report.TotalWaitTime = time.Duration(stats["wait_duration_ms"].(int64)) * time.Millisecond

		case <-ctx.Done():
			goto done
		case <-time.After(duration):
			goto done
		}
	}

done:
	if samples > 0 {
		report.AvgUtilization = totalUtilization / float64(samples)
	}
	report.MaxUtilization = maxUtil

	// 生成建议
	report.Recommendations = m.generateRecommendations(report)

	return report, nil
}

// generateRecommendations 生成优化建议
func (m *PoolMonitor) generateRecommendations(report *PerformanceReport) []string {
	recommendations := make([]string, 0)

	// 检查平均利用率
	if report.AvgUtilization > 80 {
		recommendations = append(recommendations,
			fmt.Sprintf("平均连接池利用率过高(%.2f%%)，建议增加 MaxOpenConns", report.AvgUtilization))
	} else if report.AvgUtilization < 20 {
		recommendations = append(recommendations,
			fmt.Sprintf("平均连接池利用率过低(%.2f%%)，建议减少 MaxOpenConns 以节省资源", report.AvgUtilization))
	}

	// 检查等待情况
	if report.TotalWaitCount > 1000 {
		recommendations = append(recommendations,
			fmt.Sprintf("连接等待次数过多(%d次)，建议增加连接池大小或优化查询", report.TotalWaitCount))
	}

	if report.TotalWaitTime > 10*time.Second {
		recommendations = append(recommendations,
			fmt.Sprintf("连接等待时间过长(%s)，建议优化数据库查询或增加连接数", report.TotalWaitTime))
	}

	// 检查连接泄漏
	if len(report.DetailedStats) > 0 {
		lastStats := report.DetailedStats[len(report.DetailedStats)-1]
		idleClosed := lastStats["max_idle_closed"].(int64)
		lifetimeClosed := lastStats["max_lifetime_closed"].(int64)

		if idleClosed > 100 || lifetimeClosed > 100 {
			recommendations = append(recommendations,
				"检测到大量连接关闭，建议检查是否存在连接泄漏")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "连接池配置良好，无需调整")
	}

	return recommendations
}
