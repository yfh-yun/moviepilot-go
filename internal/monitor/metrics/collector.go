package metrics

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// SystemMetrics 封装系统监控指标结构
// 该结构与 internal/apis/handlers/system/handler.go 中 MetricsResponse 及相关子结构字段对应，
// 方便 handler 直接复用或做轻量转换。
type SystemMetrics struct {
	CPUUsagePercent []float64
	CPUCores        int

	MemoryTotal        uint64
	MemoryUsed         uint64
	MemoryFree         uint64
	MemoryUsagePercent float64

	DiskTotal        uint64
	DiskUsed         uint64
	DiskFree         uint64
	DiskUsagePercent float64

	Hostname        string
	OS              string
	Platform        string
	PlatformVersion string
	Uptime          uint64

	GoGoroutines  int
	GoMemoryAlloc uint64
	GoMemorySys   uint64
	GoNumGC       uint32
	GoLastGCTime  string
	GoHeapObjects uint64
	GoVersion     string
	GoNumCPU      int
	GoOS          string
	GoArch        string
	GoCompiler    string

	Timestamp time.Time
}

// Collector 抽象系统指标采集器接口，便于未来扩展/替换实现。
type Collector interface {
	Collect(ctx context.Context) (*SystemMetrics, error)
}

// GopsutilCollector 基于 gopsutil 的默认实现
// 负责从操作系统采集 CPU/内存/磁盘/主机信息，同时结合 runtime 指标。
type GopsutilCollector struct {
	logger *zap.Logger
}

// NewGopsutilCollector 创建基于 gopsutil 的采集器
func NewGopsutilCollector() *GopsutilCollector {
	return &GopsutilCollector{logger: logger.GetLogger()}
}

// Collect 采集系统指标
func (c *GopsutilCollector) Collect(ctx context.Context) (*SystemMetrics, error) {
	m := &SystemMetrics{Timestamp: time.Now()}

	// CPU
	cpuPercent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		c.logger.Error("采集CPU使用率失败", zap.Error(err))
	} else {
		m.CPUUsagePercent = cpuPercent
	}
	m.CPUCores = runtime.NumCPU()

	// 内存
	if memInfo, err := mem.VirtualMemoryWithContext(ctx); err != nil {
		c.logger.Error("采集内存信息失败", zap.Error(err))
	} else if memInfo != nil {
		m.MemoryTotal = memInfo.Total
		m.MemoryUsed = memInfo.Used
		m.MemoryFree = memInfo.Free
		m.MemoryUsagePercent = memInfo.UsedPercent
	}

	// 磁盘（默认根分区）
	if diskInfo, err := disk.UsageWithContext(ctx, "/"); err != nil {
		c.logger.Error("采集磁盘信息失败", zap.Error(err))
	} else if diskInfo != nil {
		m.DiskTotal = diskInfo.Total
		m.DiskUsed = diskInfo.Used
		m.DiskFree = diskInfo.Free
		m.DiskUsagePercent = diskInfo.UsedPercent
	}

	// 主机信息
	if hostInfo, err := host.InfoWithContext(ctx); err != nil {
		c.logger.Error("采集主机信息失败", zap.Error(err))
	} else if hostInfo != nil {
		m.Hostname = hostInfo.Hostname
		m.OS = hostInfo.OS
		m.Platform = hostInfo.Platform
		m.PlatformVersion = hostInfo.PlatformVersion
		m.Uptime = hostInfo.Uptime
	}

	// Go runtime 指标
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.GoGoroutines = runtime.NumGoroutine()
	m.GoMemoryAlloc = memStats.Alloc
	m.GoMemorySys = memStats.Sys
	m.GoNumGC = memStats.NumGC
	m.GoLastGCTime = time.Unix(0, int64(memStats.LastGC)).Format(time.RFC3339)
	m.GoHeapObjects = memStats.HeapObjects
	m.GoVersion = runtime.Version()
	m.GoNumCPU = runtime.NumCPU()
	m.GoOS = runtime.GOOS
	m.GoArch = runtime.GOARCH
	m.GoCompiler = runtime.Compiler

	return m, nil
}
