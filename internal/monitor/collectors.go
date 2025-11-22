package monitor

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"moviepilot-go/pkg/errors"
	"moviepilot-go/pkg/logger"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
)

// CPUCollector CPU指标收集器
type CPUCollector struct {
	*BaseCollector
	lastCPUTime []cpu.TimesStat
}

// NewCPUCollector 创建CPU收集器
func NewCPUCollector() *CPUCollector {
	baseCollector := NewBaseCollector(MetricTypeCPU)
	return &CPUCollector{
		BaseCollector: baseCollector,
	}
}

// Collect 收集CPU指标
func (c *CPUCollector) Collect() (*MonitorMetrics, error) {
	c.LogCollectionStart("CPUCollector")

	// 获取CPU使用率
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		c.LogCollectionError("CPUCollector", err, "failed to get CPU usage percent")
		return nil, errors.WrapError(err, "failed to get CPU usage percent")
	}

	// 获取CPU时间
	times, err := cpu.Times(false)
	if err != nil {
		c.LogCollectionError("CPUCollector", err, "failed to get CPU times")
		return nil, errors.WrapError(err, "failed to get CPU times")
	}

	// 获取负载
	loadAvg, err := load.Avg()
	if err != nil {
		logger.Error("Failed to get load average",
			zap.String("func", "CPUCollector.Collect"),
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to get load average")
	}

	// 获取CPU数量
	counts, err := cpu.Counts(true)
	if err != nil {
		logger.Warn("Failed to get CPU counts, using runtime.NumCPU",
			zap.String("func", "CPUCollector.Collect"),
			zap.Error(err))
		counts = runtime.NumCPU()
	}

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		logger.Warn("Failed to get host info, using localhost",
			zap.String("func", "CPUCollector.Collect"),
			zap.Error(err))
		hostInfo = nil
	}
	hostname := "localhost"
	if hostInfo != nil {
		hostname = hostInfo.Hostname
	}

	// 计算CPU使用率
	var cpuUsage, user, system, idle, wait float64
	if len(times) > 0 {
		// 手动计算总时间，避免使用已弃用的Total()方法
		totalTime := times[0].User + times[0].System + times[0].Idle + times[0].Iowait +
			times[0].Irq + times[0].Softirq + times[0].Steal + times[0].Guest + times[0].GuestNice
		if totalTime > 0 {
			user = (times[0].User / totalTime) * 100
			system = (times[0].System / totalTime) * 100
			idle = (times[0].Idle / totalTime) * 100
			wait = (times[0].Iowait / totalTime) * 100
			cpuUsage = user + system + wait
		}
	}

	// 如果没有获取到使用率，使用默认值
	if len(percent) > 0 {
		cpuUsage = percent[0]
	}

	metrics := &MonitorMetrics{
		Source:    "system_monitor",
		Host:      hostname,
		Timestamp: time.Now(),
		Metrics: []MetricData{
			{
				Name:   "cpu_usage_percent",
				Type:   MetricTypeCPU,
				Value:  cpuUsage,
				Unit:   "percent",
				Tags:   map[string]string{"source": "system"},
				Labels: map[string]interface{}{"cores": counts},
			},
			{
				Name:  "cpu_user_percent",
				Type:  MetricTypeCPU,
				Value: user,
				Unit:  "percent",
				Tags:  map[string]string{"source": "system", "mode": "user"},
			},
			{
				Name:  "cpu_system_percent",
				Type:  MetricTypeCPU,
				Value: system,
				Unit:  "percent",
				Tags:  map[string]string{"source": "system", "mode": "system"},
			},
			{
				Name:  "cpu_idle_percent",
				Type:  MetricTypeCPU,
				Value: idle,
				Unit:  "percent",
				Tags:  map[string]string{"source": "system", "mode": "idle"},
			},
			{
				Name:  "cpu_wait_percent",
				Type:  MetricTypeCPU,
				Value: wait,
				Unit:  "percent",
				Tags:  map[string]string{"source": "system", "mode": "wait"},
			},
			{
				Name:  "load_1min",
				Type:  MetricTypeCPU,
				Value: loadAvg.Load1,
				Unit:  "load",
				Tags:  map[string]string{"source": "system", "period": "1min"},
			},
			{
				Name:  "load_5min",
				Type:  MetricTypeCPU,
				Value: loadAvg.Load5,
				Unit:  "load",
				Tags:  map[string]string{"source": "system", "period": "5min"},
			},
			{
				Name:  "load_15min",
				Type:  MetricTypeCPU,
				Value: loadAvg.Load15,
				Unit:  "load",
				Tags:  map[string]string{"source": "system", "period": "15min"},
			},
		},
	}

	logger.Debug("CPU metrics collected successfully",
		zap.String("func", "CPUCollector.Collect"),
		zap.String("hostname", hostname),
		zap.Float64("cpu_usage", cpuUsage),
		zap.Float64("load_1min", loadAvg.Load1),
		zap.Int("metrics_count", len(metrics.Metrics)))

	return metrics, nil
}

// MemoryCollector 内存指标收集器
type MemoryCollector struct {
	*BaseCollector
}

// NewMemoryCollector 创建内存收集器
func NewMemoryCollector() *MemoryCollector {
	baseCollector := NewBaseCollector(MetricTypeMemory)
	return &MemoryCollector{
		BaseCollector: baseCollector,
	}
}

// Collect 收集内存指标
func (c *MemoryCollector) Collect() (*MonitorMetrics, error) {
	c.LogCollectionStart("MemoryCollector")

	// 获取虚拟内存
	vMem, err := mem.VirtualMemory()
	if err != nil {
		logger.Error("Failed to get virtual memory info",
			zap.String("func", "MemoryCollector.Collect"),
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to get virtual memory info")
	}

	// 获取交换内存
	swapMem, err := mem.SwapMemory()
	if err != nil {
		logger.Error("Failed to get swap memory info",
			zap.String("func", "MemoryCollector.Collect"),
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to get swap memory info")
	}

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		logger.Warn("Failed to get host info, using localhost",
			zap.String("func", "MemoryCollector.Collect"),
			zap.Error(err))
		hostInfo = nil
	}
	hostname := "localhost"
	if hostInfo != nil {
		hostname = hostInfo.Hostname
	}

	metrics := &MonitorMetrics{
		Source:    "system_monitor",
		Host:      hostname,
		Timestamp: time.Now(),
		Metrics: []MetricData{
			{
				Name:  "memory_total_bytes",
				Type:  MetricTypeMemory,
				Value: float64(vMem.Total),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:  "memory_used_bytes",
				Type:  MetricTypeMemory,
				Value: float64(vMem.Used),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:  "memory_free_bytes",
				Type:  MetricTypeMemory,
				Value: float64(vMem.Free),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:  "memory_available_bytes",
				Type:  MetricTypeMemory,
				Value: float64(vMem.Available),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:  "memory_used_percent",
				Type:  MetricTypeMemory,
				Value: vMem.UsedPercent,
				Unit:  "percent",
				Tags:  map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:  "memory_buffers_bytes",
				Type:  MetricTypeMemory,
				Value: float64(vMem.Buffers),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "type": "buffers"},
			},
			{
				Name:  "memory_cached_bytes",
				Type:  MetricTypeMemory,
				Value: float64(vMem.Cached),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "type": "cached"},
			},
			{
				Name:  "swap_total_bytes",
				Type:  MetricTypeMemory,
				Value: float64(swapMem.Total),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "type": "swap"},
			},
			{
				Name:  "swap_used_bytes",
				Type:  MetricTypeMemory,
				Value: float64(swapMem.Used),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "type": "swap"},
			},
			{
				Name:  "swap_used_percent",
				Type:  MetricTypeMemory,
				Value: swapMem.UsedPercent,
				Unit:  "percent",
				Tags:  map[string]string{"source": "system", "type": "swap"},
			},
		},
	}

	logger.Debug("Memory metrics collected successfully",
		zap.String("func", "MemoryCollector.Collect"),
		zap.String("hostname", hostname),
		zap.Float64("memory_used_percent", vMem.UsedPercent),
		zap.Float64("swap_used_percent", swapMem.UsedPercent),
		zap.Int("metrics_count", len(metrics.Metrics)))

	return metrics, nil
}

// DiskCollector 磁盘指标收集器
type DiskCollector struct {
	*BaseCollector
}

// NewDiskCollector 创建磁盘收集器
func NewDiskCollector() *DiskCollector {
	logger.Debug("Creating new DiskCollector instance", zap.String("func", "NewDiskCollector"))
	return &DiskCollector{
		BaseCollector: NewBaseCollector(MetricTypeDisk),
	}
}

// 确保 DiskCollector 实现了 MetricCollector 接口
var _ MetricCollector = (*DiskCollector)(nil)

// Collect 收集磁盘指标
func (c *DiskCollector) Collect() (*MonitorMetrics, error) {
	logger.Debug("Collecting disk metrics", zap.String("func", "DiskCollector.Collect"))

	// 获取磁盘分区
	partitions, err := disk.Partitions(false)
	if err != nil {
		logger.Error("Failed to get disk partitions",
			zap.String("func", "DiskCollector.Collect"),
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to get disk partitions")
	}

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		logger.Warn("Failed to get host info, using localhost",
			zap.String("func", "DiskCollector.Collect"),
			zap.Error(err))
		hostInfo = nil
	}
	hostname := "localhost"
	if hostInfo != nil {
		hostname = hostInfo.Hostname
	}

	var metrics []MetricData

	// 遍历所有分区
	for _, partition := range partitions {
		// 获取分区使用率
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			logger.Warn("Failed to get disk usage for partition, skipping",
				zap.String("func", "DiskCollector.Collect"),
				zap.String("mountpoint", partition.Mountpoint),
				zap.String("device", partition.Device),
				zap.Error(err))
			continue // 忽略无法获取使用率的分区
		}

		// 添加分区指标
		partitionMetrics := []MetricData{
			{
				Name:  "disk_total_bytes",
				Type:  MetricTypeDisk,
				Value: float64(usage.Total),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "mountpoint": usage.Path, "device": partition.Device, "fstype": partition.Fstype},
			},
			{
				Name:  "disk_used_bytes",
				Type:  MetricTypeDisk,
				Value: float64(usage.Used),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "mountpoint": usage.Path, "device": partition.Device, "fstype": partition.Fstype},
			},
			{
				Name:  "disk_free_bytes",
				Type:  MetricTypeDisk,
				Value: float64(usage.Free),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "mountpoint": usage.Path, "device": partition.Device, "fstype": partition.Fstype},
			},
			{
				Name:  "disk_used_percent",
				Type:  MetricTypeDisk,
				Value: usage.UsedPercent,
				Unit:  "percent",
				Tags:  map[string]string{"source": "system", "mountpoint": usage.Path, "device": partition.Device, "fstype": partition.Fstype},
			},
			{
				Name:  "disk_inodes_total",
				Type:  MetricTypeDisk,
				Value: float64(usage.InodesTotal),
				Unit:  "count",
				Tags:  map[string]string{"source": "system", "mountpoint": usage.Path, "device": partition.Device, "fstype": partition.Fstype},
			},
			{
				Name:  "disk_inodes_used",
				Type:  MetricTypeDisk,
				Value: float64(usage.InodesUsed),
				Unit:  "count",
				Tags:  map[string]string{"source": "system", "mountpoint": usage.Path, "device": partition.Device, "fstype": partition.Fstype},
			},
			{
				Name:  "disk_inodes_free",
				Type:  MetricTypeDisk,
				Value: float64(usage.InodesFree),
				Unit:  "count",
				Tags:  map[string]string{"source": "system", "mountpoint": usage.Path, "device": partition.Device, "fstype": partition.Fstype},
			},
			{
				Name:  "disk_inodes_used_percent",
				Type:  MetricTypeDisk,
				Value: usage.InodesUsedPercent,
				Unit:  "percent",
				Tags:  map[string]string{"source": "system", "mountpoint": usage.Path, "device": partition.Device, "fstype": partition.Fstype},
			},
		}

		metrics = append(metrics, partitionMetrics...)
	}

	result := &MonitorMetrics{
		Source:    "system_monitor",
		Host:      hostname,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}

	logger.Debug("Disk metrics collected successfully",
		zap.String("func", "DiskCollector.Collect"),
		zap.String("hostname", hostname),
		zap.Int("partition_count", len(partitions)),
		zap.Int("metrics_count", len(metrics)))

	return result, nil
}

// NetworkCollector 网络指标收集器
type NetworkCollector struct {
	*BaseCollector
	lastIOStats []net.IOCountersStat
}

// NewNetworkCollector 创建网络收集器
func NewNetworkCollector() *NetworkCollector {
	logger.Debug("Creating new NetworkCollector instance", zap.String("func", "NewNetworkCollector"))
	return &NetworkCollector{
		BaseCollector: NewBaseCollector(MetricTypeNetwork),
	}
}

// Collect 收集网络指标
func (c *NetworkCollector) Collect() (*MonitorMetrics, error) {
	logger.Debug("Collecting network metrics", zap.String("func", "NetworkCollector.Collect"))

	// 获取网络IO统计
	ioStats, err := net.IOCounters(true)
	if err != nil {
		logger.Error("Failed to get network IO counters",
			zap.String("func", "NetworkCollector.Collect"),
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to get network IO counters")
	}

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		logger.Warn("Failed to get host info, using localhost",
			zap.String("func", "NetworkCollector.Collect"),
			zap.Error(err))
		hostInfo = nil
	}
	hostname := "localhost"
	if hostInfo != nil {
		hostname = hostInfo.Hostname
	}

	var metrics []MetricData

	// 遍历所有网络接口
	for _, ioStat := range ioStats {
		// 跳过回环接口
		if ioStat.Name == "lo" {
			continue
		}

		interfaceMetrics := []MetricData{
			{
				Name:  "network_bytes_sent_total",
				Type:  MetricTypeNetwork,
				Value: float64(ioStat.BytesSent),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:  "network_bytes_recv_total",
				Type:  MetricTypeNetwork,
				Value: float64(ioStat.BytesRecv),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:  "network_packets_sent_total",
				Type:  MetricTypeNetwork,
				Value: float64(ioStat.PacketsSent),
				Unit:  "packets",
				Tags:  map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:  "network_packets_recv_total",
				Type:  MetricTypeNetwork,
				Value: float64(ioStat.PacketsRecv),
				Unit:  "packets",
				Tags:  map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:  "network_errin_total",
				Type:  MetricTypeNetwork,
				Value: float64(ioStat.Errin),
				Unit:  "errors",
				Tags:  map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:  "network_errout_total",
				Type:  MetricTypeNetwork,
				Value: float64(ioStat.Errout),
				Unit:  "errors",
				Tags:  map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:  "network_dropin_total",
				Type:  MetricTypeNetwork,
				Value: float64(ioStat.Dropin),
				Unit:  "packets",
				Tags:  map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:  "network_dropout_total",
				Type:  MetricTypeNetwork,
				Value: float64(ioStat.Dropout),
				Unit:  "packets",
				Tags:  map[string]string{"source": "system", "interface": ioStat.Name},
			},
		}

		metrics = append(metrics, interfaceMetrics...)
	}

	result := &MonitorMetrics{
		Source:    "system_monitor",
		Host:      hostname,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}

	logger.Debug("Network metrics collected successfully",
		zap.String("func", "NetworkCollector.Collect"),
		zap.String("hostname", hostname),
		zap.Int("interface_count", len(ioStats)),
		zap.Int("metrics_count", len(metrics)))

	return result, nil

}

// ProcessCollector 进程指标收集器
type ProcessCollector struct {
	*BaseCollector
}

// NewProcessCollector 创建进程收集器
func NewProcessCollector() *ProcessCollector {
	logger.Debug("Creating new ProcessCollector instance", zap.String("func", "NewProcessCollector"))
	return &ProcessCollector{
		BaseCollector: NewBaseCollector(MetricTypeProcess),
	}
}

// Collect 收集进程指标
func (c *ProcessCollector) Collect() (*MonitorMetrics, error) {
	logger.Debug("Collecting process metrics", zap.String("func", "ProcessCollector.Collect"))

	// 获取所有进程
	processes, err := process.Processes()
	if err != nil {
		logger.Error("Failed to get process list",
			zap.String("func", "ProcessCollector.Collect"),
			zap.Error(err))
		return nil, errors.WrapError(err, "failed to get process list")
	}

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		logger.Warn("Failed to get host info, using localhost",
			zap.String("func", "ProcessCollector.Collect"),
			zap.Error(err))
		hostInfo = nil
	}
	hostname := "localhost"
	if hostInfo != nil {
		hostname = hostInfo.Hostname
	}

	var metrics []MetricData

	// 统计进程总数
	totalProcesses := float64(len(processes))
	metrics = append(metrics, MetricData{
		Name:  "process_total",
		Type:  MetricTypeProcess,
		Value: totalProcesses,
		Unit:  "count",
		Tags:  map[string]string{"source": "system"},
	})

	// 获取当前进程信息
	currentPID := int32(os.Getpid())
	currentProcess, err := process.NewProcess(currentPID)
	if err != nil {
		logger.Warn("Failed to get current process info",
			zap.String("func", "ProcessCollector.Collect"),
			zap.Int32("pid", currentPID),
			zap.Error(err))
	} else {
		// 获取当前进程的CPU和内存使用
		cpuPercent, err := currentProcess.CPUPercent()
		if err != nil {
			logger.Warn("Failed to get current process CPU percent",
				zap.String("func", "ProcessCollector.Collect"),
				zap.Int32("pid", currentPID),
				zap.Error(err))
		}

		memInfo, err := currentProcess.MemoryInfo()
		if err != nil {
			logger.Warn("Failed to get current process memory info",
				zap.String("func", "ProcessCollector.Collect"),
				zap.Int32("pid", currentPID),
				zap.Error(err))
		}

		if memInfo != nil {
			metrics = append(metrics, MetricData{
				Name:  "process_memory_bytes",
				Type:  MetricTypeProcess,
				Value: float64(memInfo.RSS),
				Unit:  "bytes",
				Tags:  map[string]string{"source": "system", "name": "moviepilot", "pid": fmt.Sprintf("%d", currentPID)},
			})
		}

		metrics = append(metrics, MetricData{
			Name:  "process_cpu_percent",
			Type:  MetricTypeProcess,
			Value: cpuPercent,
			Unit:  "percent",
			Tags:  map[string]string{"source": "system", "name": "moviepilot", "pid": fmt.Sprintf("%d", currentPID)},
		})
	}

	// 统计不同状态的进程数量
	statusCount := make(map[string]int)
	for _, proc := range processes {
		statuses, err := proc.Status()
		if err == nil {
			// Status 返回的是字符串数组，取第一个状态
			if len(statuses) > 0 {
				statusCount[statuses[0]]++
			}
		} else {
			logger.Debug("Failed to get process status",
				zap.String("func", "ProcessCollector.Collect"),
				zap.Int32("pid", proc.Pid),
				zap.Error(err))
		}
	}

	for status, count := range statusCount {
		metrics = append(metrics, MetricData{
			Name:  "process_status_count",
			Type:  MetricTypeProcess,
			Value: float64(count),
			Unit:  "count",
			Tags:  map[string]string{"source": "system", "status": status},
		})
	}

	result := &MonitorMetrics{
		Source:    "system_monitor",
		Host:      hostname,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}

	logger.Debug("Process metrics collected successfully",
		zap.String("func", "ProcessCollector.Collect"),
		zap.String("hostname", hostname),
		zap.Int("process_count", len(processes)),
		zap.Int("metrics_count", len(metrics)))

	return result, nil
}
