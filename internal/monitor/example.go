package monitor

import (
	"time"
)

// ExampleMonitor 使用示例
func ExampleMonitor() {
	// 创建监控实例
	monitor := NewMonitor()
	
	// 等待一段时间观察监控效�?	time.Sleep(30 * time.Second)
	
	// 停止监控
	monitor.Stop()
}
