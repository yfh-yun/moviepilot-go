package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// TimerUtils 定时器工具类
type TimerUtils struct{}

// NewTimerUtils 创建一个新�?TimerUtils 实例
func NewTimerUtils() *TimerUtils {
	return &TimerUtils{}
}

// RandomScheduler 按执行次数生成随机定时器
// numExecutions: 执行次数
// beginHour: 开始时�?// endHour: 结束时间
// minInterval: 最小间隔分�?// maxInterval: 最大间隔分�?func (t *TimerUtils) RandomScheduler(numExecutions, beginHour, endHour, minInterval, maxInterval int) []time.Time {
	trigger := make([]time.Time, 0)
	
	// 当前时间
	now := time.Now()
	
	// 创建随机的时间触发器
	randomTrigger := time.Date(now.Year(), now.Month(), now.Day(), beginHour, 0, 0, 0, now.Location())
	
	for i := 0; i < numExecutions; i++ {
		// 随机生成下一个任务的时间间隔
		intervalMinutes := rand.Intn(maxInterval-minInterval+1) + minInterval
		randomInterval := time.Duration(intervalMinutes) * time.Minute
		
		// 记录上一个任务的时间触发�?		lastRandomTrigger := randomTrigger
		
		// 更新当前时间为下一个任务的时间触发�?		randomTrigger = randomTrigger.Add(randomInterval)
		
		// 达到结束时间或者时间出现倒退时退�?		if randomTrigger.Hour() > endHour || randomTrigger.Hour() < lastRandomTrigger.Hour() {
			break
		}
		
		// 添加到队�?		trigger = append(trigger, randomTrigger)
	}
	
	return trigger
}

// RandomEvenScheduler 按执行次数尽可能平均生成随机定时�?// numExecutions: 执行次数
// beginHour: 计划范围开始的小时�?// endHour: 计划范围结束的小时数
func (t *TimerUtils) RandomEvenScheduler(numExecutions, beginHour, endHour int) []time.Time {
	triggerTimes := make([]time.Time, 0)
	
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), beginHour, 0, 0, 0, now.Location())
	endTime := time.Date(now.Year(), now.Month(), now.Day(), endHour, 0, 0, 0, now.Location())
	
	// 计算范围内的总分钟数
	totalMinutes := int(endTime.Sub(startTime).Minutes())
	
	// 计算每个执行时间段的平均长度
	segmentLength := totalMinutes / numExecutions
	
	for i := 0; i < numExecutions; i++ {
		// 在每个段内随机选择一个点
		startSegment := segmentLength * i
		endSegment := startSegment + segmentLength
		// 修复：确保不会出现负数或超出范围的情�?		if endSegment <= startSegment {
			endSegment = startSegment + 1
		}
		minute := rand.Intn(endSegment-startSegment) + startSegment
		triggerTime := startTime.Add(time.Duration(minute) * time.Minute)
		triggerTimes = append(triggerTimes, triggerTime)
	}
	
	return triggerTimes
}

// TimeDifference 判断输入时间与当前的时间差，如果输入时间大于当前时间则返回时间差，否则返回空字符�?func (t *TimerUtils) TimeDifference(inputTime time.Time) string {
	// 如果输入时间为空，则返回空字符串
	if inputTime.IsZero() {
		return ""
	}
	
	currentTime := time.Now()
	timeDifference := inputTime.Sub(currentTime)
	
	// 如果时间差小�?，返回空字符�?	if timeDifference < 0 {
		return ""
	}
	
	days := int(timeDifference.Hours()) / 24
	hours := int(timeDifference.Hours()) % 24
	minutes := int(timeDifference.Minutes()) % 60
	seconds := int(timeDifference.Seconds()) % 60
	
	timeDifferenceString := ""
	if days > 0 {
		timeDifferenceString += fmt.Sprintf("%d�?, days)
	}
	if hours > 0 {
		timeDifferenceString += fmt.Sprintf("%d小时", hours)
	}
	if minutes > 0 {
		timeDifferenceString += fmt.Sprintf("%d分钟", minutes)
	}
	if timeDifferenceString == "" && seconds > 0 {
		timeDifferenceString = fmt.Sprintf("%d�?, seconds)
	}
	
	return timeDifferenceString
}

// DiffMinutes 计算当前时间与输入时间的分钟�?func (t *TimerUtils) DiffMinutes(inputTime time.Time) int {
	// 如果输入时间为空，返�?
	if inputTime.IsZero() {
		return 0
	}
	
	timeDifference := time.Now().Sub(inputTime)
	return int(timeDifference.Minutes())
}
