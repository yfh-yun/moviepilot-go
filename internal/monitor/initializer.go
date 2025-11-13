package monitor

var monitorInstance *Monitor

// InitMonitor 初始化监控器
func InitMonitor() *Monitor {
	// 使用单例模式确保只创建一个监控器实例
	if monitorInstance == nil {
		monitorInstance = NewMonitor()
	}
	return monitorInstance
}

// StopMonitor 停止监控�?func StopMonitor() {
	// 获取监控器实例并停止
	if monitorInstance != nil {
		monitorInstance.Stop()
	}
}
