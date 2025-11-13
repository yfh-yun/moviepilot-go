package scheduler

import "time"

// WorkflowJob 工作流定时任�?type WorkflowJob struct {
	// 工作流ID
	WorkflowID string
	
	// 任务ID
	JobID string
	
	// 任务名称
	Name string
	
	// 任务函数
	Func func()
	
	// 是否正在运行
	Running bool
	
	// 提供者名�?	ProviderName string
	
	// 参数
	Kwargs map[string]interface{}
}

// Run 执行任务
func (wj *WorkflowJob) Run() {
	// 实际的任务执行逻辑应该在这里实�?}

// GetID 获取任务ID
func (wj *WorkflowJob) GetID() string {
	return wj.JobID
}

// GetName 获取任务名称
func (wj *WorkflowJob) GetName() string {
	return wj.Name
}

// GetNextRunTime 获取下次运行时间
func (wj *WorkflowJob) GetNextRunTime() time.Time {
	// 简化实现，实际应该返回下次运行时间
	return time.Now()
}

// IsRunning 是否正在运行
func (wj *WorkflowJob) IsRunning() bool {
	return wj.Running
}
