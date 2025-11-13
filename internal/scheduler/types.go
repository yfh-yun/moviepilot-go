package scheduler

// ScheduleInfo 定时任务信息
type ScheduleInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
	NextRun  string `json:"next_run,omitempty"`
}

// ScheduleJob 定时任务接口
type ScheduleJob interface {
	GetID() string
	GetName() string
	GetStatus() string
	GetProvider() string
	GetNextRun() string
}
