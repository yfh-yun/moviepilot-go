package utils

import (
	"math/rand"
	"strconv"
	"time"
)

// init 初始化随机数生成器种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomScheduler 按执行次数生成随机定时器，对应 Python TimerUtils.random_scheduler
// 返回从今天 beginHour 开始，在 [minInterval,maxInterval] 分钟随机间隔内的触发时间列表，超过 endHour 或时间倒退则停止
func RandomScheduler(numExecutions, beginHour, endHour, minInterval, maxInterval int) []time.Time {
	if numExecutions <= 0 {
		return nil
	}
	if beginHour < 0 || beginHour > 23 || endHour < 0 || endHour > 23 {
		return nil
	}
	if minInterval <= 0 {
		minInterval = 20
	}
	if maxInterval < minInterval {
		maxInterval = minInterval
	}

	now := time.Now()
	randomTrigger := time.Date(now.Year(), now.Month(), now.Day(), beginHour, 0, 0, 0, now.Location())
	var triggers []time.Time

	for i := 0; i < numExecutions; i++ {
		intervalMinutes := rand.Intn(maxInterval-minInterval+1) + minInterval
		lastTrigger := randomTrigger
		randomTrigger = randomTrigger.Add(time.Duration(intervalMinutes) * time.Minute)

		if randomTrigger.Hour() > endHour || randomTrigger.Hour() < lastTrigger.Hour() {
			break
		}
		triggers = append(triggers, randomTrigger)
	}

	return triggers
}

// RandomEvenScheduler 按执行次数尽可能平均生成随机定时器，对应 Python TimerUtils.random_even_scheduler
// 在 [beginHour,endHour] 范围内平均分段，每段内随机选择一个时间点
func RandomEvenScheduler(numExecutions, beginHour, endHour int) []time.Time {
	if numExecutions <= 0 {
		return nil
	}
	if beginHour < 0 || beginHour > 23 || endHour <= beginHour || endHour > 23 {
		return nil
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), beginHour, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), endHour, 0, 0, 0, now.Location())

	totalMinutes := int(end.Sub(start).Minutes())
	if totalMinutes <= 0 {
		return nil
	}

	segmentLen := totalMinutes / numExecutions
	if segmentLen <= 0 {
		segmentLen = 1
	}

	triggers := make([]time.Time, 0, numExecutions)
	for i := 0; i < numExecutions; i++ {
		startSegment := segmentLen * i
		endSegment := startSegment + segmentLen
		if endSegment > totalMinutes {
			endSegment = totalMinutes
		}
		if endSegment <= startSegment {
			minute := startSegment
			triggers = append(triggers, start.Add(time.Duration(minute)*time.Minute))
			continue
		}
		minute := rand.Intn(endSegment-startSegment) + startSegment
		triggers = append(triggers, start.Add(time.Duration(minute)*time.Minute))
	}

	return triggers
}

// TimeDifference 计算输入时间与当前时间的差值，返回形如 "x天x小时x分钟" 的字符串，若已过期则返回空字符串
// 对应 Python TimerUtils.time_difference
func TimeDifference(input time.Time) string {
	if input.IsZero() {
		return ""
	}

	// 使用与Python版本一致的时区处理：UTC时间转换为本地时区
	now := time.Now().UTC().Local()
	diff := input.Sub(now)
	if diff <= 0 {
		return ""
	}

	days := int(diff.Hours()) / 24
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60

	result := ""
	if days > 0 {
		result += strconv.Itoa(days) + "天"
	}
	if hours > 0 {
		result += strconv.Itoa(hours) + "小时"
	}
	if minutes > 0 {
		result += strconv.Itoa(minutes) + "分钟"
	}
	if result == "" && seconds > 0 {
		result = strconv.Itoa(seconds) + "秒"
	}

	return result
}

// DiffMinutes 计算当前时间与输入时间的分钟差，对应 Python TimerUtils.diff_minutes
func DiffMinutes(input time.Time) int {
	if input.IsZero() {
		return 0
	}
	timeDifference := time.Now().Sub(input)
	return int(timeDifference.Seconds() / 60)
}
