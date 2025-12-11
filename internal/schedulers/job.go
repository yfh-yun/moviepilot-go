package schedulers

import (
	"context"
	"time"
)

// Job 任务接口
type Job interface {
	ID() string
	Name() string
	Run(ctx context.Context) error
}

// Trigger 触发器接口
type Trigger interface {
	Next(after time.Time) time.Time
}

// JobInfo 任务信息
type JobInfo struct {
	ID          string
	Name        string
	Status      string // running/paused/stopped
	NextRunTime time.Time
	LastRunTime time.Time
	LastError   string
}

// IntervalTrigger 间隔触发器
type IntervalTrigger struct {
	Interval time.Duration
}

// NewIntervalTrigger 创建间隔触发器
func NewIntervalTrigger(interval time.Duration) *IntervalTrigger {
	return &IntervalTrigger{
		Interval: interval,
	}
}

// Next 返回下一次执行时间
func (t *IntervalTrigger) Next(after time.Time) time.Time {
	return after.Add(t.Interval)
}

// CronTrigger Cron表达式触发器
type CronTrigger struct {
	Spec string
}

// NewCronTrigger 创建Cron触发器
func NewCronTrigger(spec string) *CronTrigger {
	return &CronTrigger{
		Spec: spec,
	}
}

// Next 返回下一次执行时间
// 注意：实际实现需要解析Cron表达式，这里简化处理
func (t *CronTrigger) Next(after time.Time) time.Time {
	// 实际实现中需要使用cron库解析spec
	// 这里暂时返回1小时后，实际使用时需要替换
	return after.Add(1 * time.Hour)
}
