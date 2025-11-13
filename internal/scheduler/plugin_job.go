package scheduler

import "time"

// PluginJob 插件定时任务
type PluginJob struct {
	// 插件ID
	PluginID string
	
	// 任务ID
	JobID string
	
	// 任务名称
	Name string
	
	// 任务函数
	Func func()
	
	// 是否正在运行
	Running bool
	
	// 插件名称
	PluginName string
	
	// 参数
	Kwargs map[string]interface{}
}

// Run 执行任务
func (pj *PluginJob) Run() {
	// 实际的任务执行逻辑应该在这里实�?}

// GetID 获取任务ID
func (pj *PluginJob) GetID() string {
	return pj.JobID
}

// GetName 获取任务名称
func (pj *PluginJob) GetName() string {
	return pj.Name
}

// GetNextRunTime 获取下次运行时间
func (pj *PluginJob) GetNextRunTime() time.Time {
	// 简化实现，实际应该返回下次运行时间
	return time.Now()
}

// IsRunning 是否正在运行
func (pj *PluginJob) IsRunning() bool {
	return pj.Running
}
