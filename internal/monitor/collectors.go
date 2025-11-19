package monitor

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// CPUCollector CPU指标收集器
type CPUCollector struct {
	lastCPUTime []cpu.TimesStat
}

// NewCPUCollector 创建CPU收集器
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

// Collect 收集CPU指标
func (c *CPUCollector) Collect() (*MonitorMetrics, error) {
	// 获取CPU使用率
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("获取CPU使用率失败: %w", err)
	}
	
	// 获取CPU时间
	times, err := cpu.Times(false)
	if err != nil {
		return nil, fmt.Errorf("获取CPU时间失败: %w", err)
	}
	
	// 获取负载
	loadAvg, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("获取负载失败: %w", err)
	}
	
	// 获取CPU数量
	counts, err := cpu.Counts(true)
	if err != nil {
		counts = runtime.NumCPU()
	}
	
	// 获取主机信息
	hostInfo, _ := host.Info()
	hostname := "localhost"
	if hostInfo != nil {
		hostname = hostInfo.Hostname
	}
	
	var cpuUsage, user, system, idle, wait float64
	if len(times) > 0 {
		total := times[0].Total()
		if total > 0 {
			user = (times[0].User / total) * 100
			system = (times[0].System / total) * 100
			idle = (times[0].Idle / total) * 100
			wait = (times[0].Iowait / total) * 100
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
				Name:      "cpu_usage_percent",
				Type:      MetricTypeCPU,
				Value:     cpuUsage,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system"},
				Labels:    map[string]interface{}{"cores": counts},
			},
			{
				Name:      "cpu_user_percent",
				Type:      MetricTypeCPU,
				Value:     user,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system", "mode": "user"},
			},
			{
				Name:      "cpu_system_percent",
				Type:      MetricTypeCPU,
				Value:     system,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system", "mode": "system"},
			},
			{
				Name:      "cpu_idle_percent",
				Type:      MetricTypeCPU,
				Value:     idle,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system", "mode": "idle"},
			},
			{
				Name:      "cpu_wait_percent",
				Type:      MetricTypeCPU,
				Value:     wait,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system", "mode": "wait"},
			},
			{
				Name:      "load_1min",
				Type:      MetricTypeCPU,
				Value:     loadAvg.Load1,
				Unit:      "load",
				Tags:      map[string]string{"source": "system", "period": "1min"},
			},
			{
				Name:      "load_5min",
				Type:      MetricTypeCPU,
				Value:     loadAvg.Load5,
				Unit:      "load",
				Tags:      map[string]string{"source": "system", "period": "5min"},
			},
			{
				Name:      "load_15min",
				Type:      MetricTypeCPU,
				Value:     loadAvg.Load15,
				Unit:      "load",
				Tags:      map[string]string{"source": "system", "period": "15min"},
			},
		},
	}
	
	return metrics, nil
}

// GetMetricType 获取指标类型
func (c *CPUCollector) GetMetricType() MetricType {
	return MetricTypeCPU
}

// MemoryCollector 内存指标收集器
type MemoryCollector struct{}

// NewMemoryCollector 创建内存收集器
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

// Collect 收集内存指标
func (c *MemoryCollector) Collect() (*MonitorMetrics, error) {
	// 获取虚拟内存
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("获取虚拟内存失败: %w", err)
	}
	
	// 获取交换内存
	swapMem, err := mem.SwapMemory()
	if err != nil {
		return nil, fmt.Errorf("获取交换内存失败: %w", err)
	}
	
	// 获取主机信息
	hostInfo, _ := host.Info()
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
				Name:      "memory_total_bytes",
				Type:      MetricTypeMemory,
				Value:     float64(vMem.Total),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:      "memory_used_bytes",
				Type:      MetricTypeMemory,
				Value:     float64(vMem.Used),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:      "memory_free_bytes",
				Type:      MetricTypeMemory,
				Value:     float64(vMem.Free),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:      "memory_available_bytes",
				Type:      MetricTypeMemory,
				Value:     float64(vMem.Available),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:      "memory_used_percent",
				Type:      MetricTypeMemory,
				Value:     vMem.UsedPercent,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system", "type": "virtual"},
			},
			{
				Name:      "memory_buffers_bytes",
				Type:      MetricTypeMemory,
				Value:     float64(vMem.Buffers),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "type": "buffers"},
			},
			{
				Name:      "memory_cached_bytes",
				Type:      MetricTypeMemory,
				Value:     float64(vMem.Cached),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "type": "cached"},
			},
			{
				Name:      "swap_total_bytes",
				Type:      MetricTypeMemory,
				Value:     float64(swapMem.Total),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "type": "swap"},
			},
			{
				Name:      "swap_used_bytes",
				Type:      MetricTypeMemory,
				Value:     float64(swapMem.Used),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "type": "swap"},
			},
			{
				Name:      "swap_used_percent",
				Type:      MetricTypeMemory,
				Value:     swapMem.UsedPercent,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system", "type": "swap"},
			},
		},
	}
	
	return metrics, nil
}

// GetMetricType 获取指标类型
func (c *MemoryCollector) GetMetricType() MetricType {
	return MetricTypeMemory
}

// DiskCollector 磁盘指标收集器
type DiskCollector struct{}

// NewDiskCollector 创建磁盘收集器
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

// Collect 收集磁盘指标
func (c *DiskCollector) Collect() (*MonitorMetrics, error) {
	// 获取磁盘分区
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘分区失败: %w", err)
	}
	
	// 获取主机信息
	hostInfo, _ := host.Info()
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
			continue // 忽略无法获取使用率的分区
		}
		
		// 添加分区指标
		partitionMetrics := []MetricData{
			{
				Name:      "disk_total_bytes",
				Type:      MetricTypeDisk,
				Value:     float64(usage.Total),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "mountpoint": usage.Path, "device": usage.Device, "fstype": usage.Fstype},
			},
			{
				Name:      "disk_used_bytes",
				Type:      MetricTypeDisk,
				Value:     float64(usage.Used),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "mountpoint": usage.Path, "device": usage.Device, "fstype": usage.Fstype},
			},
			{
				Name:      "disk_free_bytes",
				Type:      MetricTypeDisk,
				Value:     float64(usage.Free),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "mountpoint": usage.Path, "device": usage.Device, "fstype": usage.Fstype},
			},
			{
				Name:      "disk_used_percent",
				Type:      MetricTypeDisk,
				Value:     usage.UsedPercent,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system", "mountpoint": usage.Path, "device": usage.Device, "fstype": usage.Fstype},
			},
			{
				Name:      "disk_inodes_total",
				Type:      MetricTypeDisk,
				Value:     float64(usage.InodesTotal),
				Unit:      "count",
				Tags:      map[string]string{"source": "system", "mountpoint": usage.Path, "device": usage.Device, "fstype": usage.Fstype},
			},
			{
				Name:      "disk_inodes_used",
				Type:      MetricTypeDisk,
				Value:     float64(usage.InodesUsed),
				Unit:      "count",
				Tags:      map[string]string{"source": "system", "mountpoint": usage.Path, "device": usage.Device, "fstype": usage.Fstype},
			},
			{
				Name:      "disk_inodes_free",
				Type:      MetricTypeDisk,
				Value:     float64(usage.InodesFree),
				Unit:      "count",
				Tags:      map[string]string{"source": "system", "mountpoint": usage.Path, "device": usage.Device, "fstype": usage.Fstype},
			},
			{
				Name:      "disk_inodes_used_percent",
				Type:      MetricTypeDisk,
				Value:     usage.InodesUsedPercent,
				Unit:      "percent",
				Tags:      map[string]string{"source": "system", "mountpoint": usage.Path, "device": usage.Device, "fstype": usage.Fstype},
			},
		}
		
		metrics = append(metrics, partitionMetrics...)
	}
	
	return &MonitorMetrics{
		Source:    "system_monitor",
		Host:      hostname,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}, nil
}

// GetMetricType 获取指标类型
func (c *DiskCollector) GetMetricType() MetricType {
	return MetricTypeDisk
}

// NetworkCollector 网络指标收集器
type NetworkCollector struct {
	lastIOStats []net.IOCountersStat
}

// NewNetworkCollector 创建网络收集器
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

// Collect 收集网络指标
func (c *NetworkCollector) Collect() (*MonitorMetrics, error) {
	// 获取网络IO统计
	ioStats, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("获取网络IO失败: %w", err)
	}
	
	// 获取主机信息
	hostInfo, _ := host.Info()
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
				Name:      "network_bytes_sent_total",
				Type:      MetricTypeNetwork,
				Value:     float64(ioStat.BytesSent),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:      "network_bytes_recv_total",
				Type:      MetricTypeNetwork,
				Value:     float64(ioStat.BytesRecv),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:      "network_packets_sent_total",
				Type:      MetricTypeNetwork,
				Value:     float64(ioStat.PacketsSent),
				Unit:      "packets",
				Tags:      map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:      "network_packets_recv_total",
				Type:      MetricTypeNetwork,
				Value:     float64(ioStat.PacketsRecv),
				Unit:      "packets",
				Tags:      map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:      "network_errin_total",
				Type:      MetricTypeNetwork,
				Value:     float64(ioStat.Errin),
				Unit:      "errors",
				Tags:      map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:      "network_errout_total",
				Type:      MetricTypeNetwork,
				Value:     float64(ioStat.Errout),
				Unit:      "errors",
				Tags:      map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:      "network_dropin_total",
				Type:      MetricTypeNetwork,
				Value:     float64(ioStat.Dropin),
				Unit:      "packets",
				Tags:      map[string]string{"source": "system", "interface": ioStat.Name},
			},
			{
				Name:      "network_dropout_total",
				Type:      MetricTypeNetwork,
				Value:     float64(ioStat.Dropout),
				Unit:      "packets",
				Tags:      map[string]string{"source": "system", "interface": ioStat.Name},
			},
		}
		
		metrics = append(metrics, interfaceMetrics...)
	}
	
	return &MonitorMetrics{
		Source:    "system_monitor",
		Host:      hostname,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}, nil
}

// GetMetricType 获取指标类型
func (c *NetworkCollector) GetMetricType() MetricType {
	return MetricTypeNetwork
}

// ProcessCollector 进程指标收集器
type ProcessCollector struct{}

// NewProcessCollector 创建进程收集器
func NewProcessCollector() *ProcessCollector {
	return &ProcessCollector{}
}

// Collect 收集进程指标
func (c *ProcessCollector) Collect() (*MonitorMetrics, error) {
	// 获取所有进程
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("获取进程列表失败: %w", err)
	}
	
	// 获取主机信息
	hostInfo, _ := host.Info()
	hostname := "localhost"
	if hostInfo != nil {
		hostname = hostInfo.Hostname
	}
	
	var metrics []MetricData
	
	// 统计进程总数
	totalProcesses := float64(len(processes))
	metrics = append(metrics, MetricData{
		Name:      "process_total",
		Type:      MetricTypeProcess,
		Value:     totalProcesses,
		Unit:      "count",
		Tags:      map[string]string{"source": "system"},
	})
	
	// 获取当前进程信息
	currentPID := int32(runtime.Getpid())
	currentProcess, err := process.NewProcess(currentPID)
	if err == nil {
		// 获取当前进程的CPU和内存使用
		cpuPercent, _ := currentProcess.CPUPercent()
		memInfo, _ := currentProcess.MemoryInfo()
		
		if memInfo != nil {
			metrics = append(metrics, MetricData{
				Name:      "process_memory_bytes",
				Type:      MetricTypeProcess,
				Value:     float64(memInfo.RSS),
				Unit:      "bytes",
				Tags:      map[string]string{"source": "system", "name": "moviepilot", "pid": fmt.Sprintf("%d", currentPID)},
			})
		}
		
		metrics = append(metrics, MetricData{
			Name:      "process_cpu_percent",
			Type:      MetricTypeProcess,
			Value:     cpuPercent,
			Unit:      "percent",
			Tags:      map[string]string{"source": "system", "name": "moviepilot", "pid": fmt.Sprintf("%d", currentPID)},
		})
	}
	
	// 统计不同状态的进程数量
	statusCount := make(map[string]int)
	for _, proc := range processes {
		status, err := proc.Status()
		if err == nil {
			statusCount[status]++
		}
	}
	
	for status, count := range statusCount {
		metrics = append(metrics, MetricData{
			Name:      "process_status_count",
			Type:      MetricTypeProcess,
			Value:     float64(count),
			Unit:      "count",
			Tags:      map[string]string{"source": "system", "status": status},
		})
	}
	
	return &MonitorMetrics{
		Source:    "system_monitor",
		Host:      hostname,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}, nil
}

// GetMetricType 获取指标类型
func (c *ProcessCollector) GetMetricType() MetricType {
	return MetricTypeProcess
}