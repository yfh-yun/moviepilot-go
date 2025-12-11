package utils

import (
	"runtime"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// GetMemoryUsageMB 获取当前进程的内存使用情况（MB）
// 对应 Python get_memory_usage 函数
func GetMemoryUsageMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024.0 / 1024.0
}

// MemoryGCOptions 内存回收选项
// 对应 Python memory_gc 装饰器参数
type MemoryGCOptions struct {
	ForceCollect   bool    // 是否强制执行垃圾回收
	LogMemoryUsage bool    // 是否记录内存使用日志
	FuncName       string  // 函数名称
}

// WithMemoryGC 执行函数并进行内存回收
// 对应 Python memory_gc 装饰器
func WithMemoryGC(opts MemoryGCOptions, fn func() error) error {
	l := logger.GetLogger()
	var before, after, afterGC float64
	
	// 记录函数执行前的内存使用情况
	if opts.LogMemoryUsage {
		before = GetMemoryUsageMB()
		l.Info("函数执行前内存使用", zap.String("func", opts.FuncName), zap.Float64("memory_mb", before))
	}

	err := fn()

	// 记录函数执行后的内存使用情况
	if opts.LogMemoryUsage {
		after = GetMemoryUsageMB()
		l.Info("函数执行后内存使用", zap.String("func", opts.FuncName), zap.Float64("memory_mb", after))
		if before > 0 {
			diff := after - before
			l.Info("函数内存变化", zap.String("func", opts.FuncName), zap.Float64("diff_mb", diff))
		}
	}

	// 强制垃圾回收
	if opts.ForceCollect {
		runtime.GC()
		if opts.LogMemoryUsage {
			l.Info("垃圾回收完成", zap.String("func", opts.FuncName))
		}
	}

	// 记录垃圾回收后的内存使用情况
	if opts.LogMemoryUsage {
		afterGC = GetMemoryUsageMB()
		l.Info("垃圾回收后内存使用", zap.String("func", opts.FuncName), zap.Float64("memory_mb", afterGC))
		if after > 0 {
			freed := after - afterGC
			l.Info("释放内存", zap.String("func", opts.FuncName), zap.Float64("freed_mb", freed))
		}
	}

	return err
}

// MemoryMonitorOptions 内存监控选项
// 对应 Python memory_monitor 装饰器参数
type MemoryMonitorOptions struct {
	ThresholdMB float64 // 内存阈值（MB）
	FuncName    string  // 函数名称
}

// WithMemoryMonitor 执行函数并监控内存使用
// 对应 Python memory_monitor 装饰器
func WithMemoryMonitor(opts MemoryMonitorOptions, fn func() error) error {
	l := logger.GetLogger()
	
	// 执行前检查内存使用情况
	if opts.ThresholdMB > 0 {
		current := GetMemoryUsageMB()
		if current > opts.ThresholdMB {
			l.Warn("内存使用超过阈值", zap.String("func", opts.FuncName), zap.Float64("threshold_mb", opts.ThresholdMB), zap.Float64("current_mb", current))
			runtime.GC()
			l.Info("自动垃圾回收完成", zap.String("func", opts.FuncName))
		}
	}

	err := fn()

	// 执行后检查内存使用情况
	if opts.ThresholdMB > 0 {
		after := GetMemoryUsageMB()
		if after > opts.ThresholdMB {
			runtime.GC()
			l.Info("函数执行后垃圾回收完成", zap.String("func", opts.FuncName))
		}
	}

	return err
}

// MemoryCleanup 便捷的内存清理函数
// 对应 Python memory_cleanup 装饰器别名
func MemoryCleanup(fn func() error, name string) error {
	return WithMemoryGC(MemoryGCOptions{ForceCollect: true, LogMemoryUsage: false, FuncName: name}, fn)
}

// AutoGC 自动垃圾回收函数
// 对应 Python auto_gc 装饰器别名
func AutoGC(fn func() error, name string) error {
	return WithMemoryGC(MemoryGCOptions{ForceCollect: true, LogMemoryUsage: true, FuncName: name}, fn)
}

// MemoryWatch 内存监控函数
// 对应 Python memory_watch 装饰器别名
func MemoryWatch(thresholdMB float64, fn func() error, name string) error {
	return WithMemoryMonitor(MemoryMonitorOptions{ThresholdMB: thresholdMB, FuncName: name}, fn)
}

// MemoryGC 内存回收函数（便捷版本）
// 对应 Python memory_gc 装饰器
func MemoryGC(forceCollect bool, logMemoryUsage bool, fn func() error, name string) error {
	return WithMemoryGC(MemoryGCOptions{
		ForceCollect:   forceCollect,
		LogMemoryUsage: logMemoryUsage,
		FuncName:       name,
	}, fn)
}

// MemoryMonitor 内存监控函数（便捷版本）
// 对应 Python memory_monitor 装饰器
func MemoryMonitor(thresholdMB float64, fn func() error, name string) error {
	return WithMemoryMonitor(MemoryMonitorOptions{
		ThresholdMB: thresholdMB,
		FuncName:    name,
	}, fn)
}

// 便捷的函数别名，与 Python 装饰器别名保持一致
// 在 Go 中，函数名本身就是标识符，不需要额外的变量别名
// 以下是可用的函数别名说明：
// - MemoryCleanup: 内存清理函数
// - AutoGC: 自动垃圾回收函数
// - MemoryWatch: 内存监控函数
// - MemoryGC: 内存回收函数
// - MemoryMonitor: 内存监控函数
