// Package utils 提供内存回收相关的工具函�?package utils

import (
	"fmt"
	"runtime"
	"time"

	"moviepilot-go/internal/logger"
)

// GCUtils 内存回收工具�?type GCUtils struct{}

// MemoryGCResult 包装函数执行结果和相关信�?type MemoryGCResult struct {
	Result       interface{}
	MemoryBefore float64
	MemoryAfter  float64
	MemoryFreed  float64
	TimeTaken    time.Duration
	ObjectsFreed int
}

// MemoryGC 内存回收装饰�?// forceCollect: 是否强制执行垃圾回收
// logMemoryUsage: 是否记录内存使用日志
// 返回装饰器函�?func (g *GCUtils) MemoryGC(forceCollect, logMemoryUsage bool) func(func(...interface{}) (interface{}, error)) func(...interface{}) (interface{}, error) {
	return func(fn func(...interface{}) (interface{}, error)) func(...interface{}) (interface{}, error) {
		return func(args ...interface{}) (interface{}, error) {
			var mBefore, mAfter, mAfterGC runtime.MemStats
			var memoryBefore, memoryAfter, memoryAfterGC float64
			
			// 记录函数执行前的内存使用情况
			if logMemoryUsage {
				runtime.ReadMemStats(&mBefore)
				memoryBefore = float64(mBefore.Alloc) / 1024 / 1024 // 转换为MB
				logger.GetLoggerManager().Info(fmt.Sprintf("函数执行前内存使�? %.2f MB", memoryBefore))
			}
			
			// 执行原函数并计时
			startTime := time.Now()
			result, err := fn(args...)
			timeTaken := time.Since(startTime)
			
			// 记录函数执行后的内存使用情况
			if logMemoryUsage {
				runtime.ReadMemStats(&mAfter)
				memoryAfter = float64(mAfter.Alloc) / 1024 / 1024 // 转换为MB
				logger.GetLoggerManager().Info(fmt.Sprintf("函数执行后内存使�? %.2f MB", memoryAfter))
				
				if memoryBefore > 0 {
					memDiff := memoryAfter - memoryBefore
					logger.GetLoggerManager().Info(fmt.Sprintf("函数内存变化: %.2f MB", memDiff))
				}
				logger.GetLoggerManager().Info(fmt.Sprintf("函数执行耗时: %v", timeTaken))
			}
			
			var objectsFreed int
			// 强制垃圾回收
			if forceCollect {
				gcStart := time.Now()
				objectsFreed = g.ForceGC()
				gcDuration := time.Since(gcStart)
				
				if logMemoryUsage {
					logger.GetLoggerManager().Info(fmt.Sprintf("垃圾回收完成，回收对象数: %d，耗时: %v", objectsFreed, gcDuration))
				}
			}
			
			var memoryFreed float64
			// 记录垃圾回收后的内存使用情况
			if logMemoryUsage {
				runtime.ReadMemStats(&mAfterGC)
				memoryAfterGC = float64(mAfterGC.Alloc) / 1024 / 1024 // 转换为MB
				logger.GetLoggerManager().Info(fmt.Sprintf("垃圾回收后内存使�? %.2f MB", memoryAfterGC))
				
				if memoryAfter > 0 {
					memoryFreed = memoryAfter - memoryAfterGC
					logger.GetLoggerManager().Info(fmt.Sprintf("释放内存: %.2f MB", memoryFreed))
				}
			}
			
			// 如果需要，可以返回详细的内存信�?			if logMemoryUsage {
				logger.GetLoggerManager().Info(fmt.Sprintf("函数执行完成，总耗时: %v", timeTaken))
			}
			
			return result, err
		}
	}
}

// ForceGC 强制执行垃圾回收
// 返回回收的对象数量（Go中无法准确获取，返回估计值）
func (g *GCUtils) ForceGC() int {
	// Go的垃圾回收机制与Python不同，无法获取准确的回收对象�?	// 我们通过统计GC前后的数值来估算
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	
	runtime.GC()
	runtime.GC() // 多次调用确保回收
	
	runtime.ReadMemStats(&after)
	
	// 估算回收的对象数（不准确，仅作参考）
	freedObjects := int((before.Frees - after.Frees) - (before.Mallocs - after.Mallocs))
	if freedObjects < 0 {
		freedObjects = 0
	}
	
	return freedObjects
}

// GetMemoryUsage 获取当前进程的内存使用情况（MB�?// 返回内存使用量（MB�?func (g *GCUtils) GetMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024 // 转换为MB
}

// MemoryMonitor 内存监控装饰器，当内存使用超过阈值时自动触发垃圾回收
// thresholdMB: 内存阈值（MB），超过此值将触发垃圾回收
// 返回装饰器函�?func (g *GCUtils) MemoryMonitor(thresholdMB float64) func(func(...interface{}) (interface{}, error)) func(...interface{}) (interface{}, error) {
	return func(fn func(...interface{}) (interface{}, error)) func(...interface{}) (interface{}, error) {
		return func(args ...interface{}) (interface{}, error) {
			// 检查内存使用情�?			currentMemory := g.GetMemoryUsage()
			
			if thresholdMB > 0 && currentMemory > thresholdMB {
				logger.GetLoggerManager().Warn(fmt.Sprintf("内存使用超过阈�?%.2fMB，当前使�? %.2fMB", thresholdMB, currentMemory))
				collected := g.ForceGC()
				logger.GetLoggerManager().Info(fmt.Sprintf("自动垃圾回收完成，回收对象数: %d", collected))
			}
			
			// 执行原函�?			result, err := fn(args...)
			
			// 执行后再次检查并回收
			if thresholdMB > 0 {
				memoryAfter := g.GetMemoryUsage()
				if memoryAfter > thresholdMB {
					collected := g.ForceGC()
					logger.GetLoggerManager().Info(fmt.Sprintf("函数执行后垃圾回收完成，回收对象�? %d", collected))
				}
			}
			
			return result, err
		}
	}
}

// MemoryCleanup 便捷的装饰器别名
func (g *GCUtils) MemoryCleanup(forceCollect, logMemoryUsage bool) func(func(...interface{}) (interface{}, error)) func(...interface{}) (interface{}, error) {
	return g.MemoryGC(forceCollect, logMemoryUsage)
}

// AutoGC 便捷的装饰器别名
func (g *GCUtils) AutoGC() func(func(...interface{}) (interface{}, error)) func(...interface{}) (interface{}, error) {
	return g.MemoryGC(true, true)
}

// MemoryWatch 便捷的装饰器别名
func (g *GCUtils) MemoryWatch(thresholdMB float64) func(func(...interface{}) (interface{}, error)) func(...interface{}) (interface{}, error) {
	return g.MemoryMonitor(thresholdMB)
}
