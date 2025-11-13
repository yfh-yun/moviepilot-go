package scheduler

import (
	"time"
)

// Job 定时任务接口
type Job interface {
	Run()
	GetID() string
	GetName() string
	GetNextRunTime() time.Time
	IsRunning() bool
}
