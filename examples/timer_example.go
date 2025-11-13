package main

import (
	"fmt"
	"time"

	utils "moviepilot-go/internal/utils"
)

func main() {
	timerUtils := utils.NewTimerUtils()

	// 测试 RandomScheduler
	fmt.Println("=== 测试 RandomScheduler ===")
	triggers := timerUtils.RandomScheduler(5, 7, 23, 20, 40)
	fmt.Printf("生成�?%d 个随机时�?\n", len(triggers))
	for i, trigger := range triggers {
		fmt.Printf("  %d. %s\n", i+1, trigger.Format("2006-01-02 15:04:05"))
	}

	// 测试 RandomEvenScheduler
	fmt.Println("\n=== 测试 RandomEvenScheduler ===")
	evenTriggers := timerUtils.RandomEvenScheduler(3, 9, 21)
	fmt.Printf("生成�?%d 个平均分布的随机时间:\n", len(evenTriggers))
	for i, trigger := range evenTriggers {
		fmt.Printf("  %d. %s\n", i+1, trigger.Format("2006-01-02 15:04:05"))
	}

	// 测试 TimeDifference
	fmt.Println("\n=== 测试 TimeDifference ===")
	futureTime := time.Now().Add(2*time.Hour + 30*time.Minute + 15*time.Second)
	diffStr := timerUtils.TimeDifference(futureTime)
	fmt.Printf("距离未来时间 %s 还有: %s\n", futureTime.Format("15:04:05"), diffStr)

	pastTime := time.Now().Add(-1 * time.Hour)
	diffStr = timerUtils.TimeDifference(pastTime)
	fmt.Printf("距离过去时间 %s 的差�? '%s'\n", pastTime.Format("15:04:05"), diffStr)

	// 测试 DiffMinutes
	fmt.Println("\n=== 测试 DiffMinutes ===")
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	minutesDiff := timerUtils.DiffMinutes(fiveMinutesAgo)
	fmt.Printf("距离 %s 已过�?%d 分钟\n", fiveMinutesAgo.Format("15:04:05"), minutesDiff)

	tenMinutesFuture := time.Now().Add(10 * time.Minute)
	minutesDiff = timerUtils.DiffMinutes(tenMinutesFuture)
	fmt.Printf("距离 %s 还有 %d 分钟\n", tenMinutesFuture.Format("15:04:05"), minutesDiff)

	fmt.Println("\n=== 测试完成 ===")
}
