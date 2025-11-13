package scheduler

// InitScheduler 初始化定时器
func InitScheduler() *Scheduler {
	// 获取并返回定时任务管理器单例实例
	// 在Go版本中，GetScheduler()已经创建了单例实�?	scheduler := GetScheduler()
	// 调用Init方法初始化定时器
	scheduler.Init()
	return scheduler
}

// StopScheduler 停止定时�?func StopScheduler() {
	// 获取定时任务管理器单例实例并停止
	scheduler := GetScheduler()
	scheduler.Stop()
}

// RestartScheduler 重启定时�?func RestartScheduler() {
	// 获取定时任务管理器单例实例并重新初始�?	scheduler := GetScheduler()
	scheduler.Init()
}

// InitPluginScheduler 初始化插件定时器
func InitPluginScheduler() {
	// 获取定时任务管理器单例实例并初始化插件任�?	scheduler := GetScheduler()
	// 注意：在当前代码中没有看到InitPluginJobs方法的实�?	// 需要根据实际的插件系统实现来完善此方法
	scheduler.InitPluginJobs()
}
