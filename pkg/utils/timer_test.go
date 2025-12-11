package utils

import (
	"testing"
	"time"
)

// TestRandomScheduler 测试RandomScheduler函数
func TestRandomScheduler(t *testing.T) {
	// 测试正常情况
	triggers := RandomScheduler(3, 7, 23, 20, 40)
	if len(triggers) < 1 {
		t.Errorf("Expected at least 1 trigger, got %d", len(triggers))
	}

	// 测试边界情况：执行次数为0
	triggers = RandomScheduler(0, 7, 23, 20, 40)
	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(triggers))
	}

	// 测试边界情况：无效的小时范围
	triggers = RandomScheduler(3, 25, 30, 20, 40)
	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(triggers))
	}

	// 测试边界情况：最小间隔大于最大间隔
	triggers = RandomScheduler(3, 7, 23, 40, 20)
	if len(triggers) < 1 {
		t.Errorf("Expected at least 1 trigger, got %d", len(triggers))
	}
}

// TestRandomEvenScheduler 测试RandomEvenScheduler函数
func TestRandomEvenScheduler(t *testing.T) {
	// 测试正常情况
	triggers := RandomEvenScheduler(3, 7, 23)
	if len(triggers) != 3 {
		t.Errorf("Expected 3 triggers, got %d", len(triggers))
	}

	// 测试边界情况：执行次数为0
	triggers = RandomEvenScheduler(0, 7, 23)
	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(triggers))
	}

	// 测试边界情况：无效的小时范围
	triggers = RandomEvenScheduler(3, 23, 7)
	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(triggers))
	}

	// 测试边界情况：相同的开始和结束小时
	triggers = RandomEvenScheduler(3, 12, 12)
	if len(triggers) != 0 {
		t.Errorf("Expected 0 triggers, got %d", len(triggers))
	}
}

// TestTimeDifference 测试TimeDifference函数
func TestTimeDifference(t *testing.T) {
	// 测试正常情况：未来时间
	futureTime := time.Now().Add(2 * time.Hour)
	result := TimeDifference(futureTime)
	if result == "" {
		t.Errorf("Expected non-empty result for future time, got empty")
	}

	// 测试正常情况：过去时间
	pastTime := time.Now().Add(-2 * time.Hour)
	result = TimeDifference(pastTime)
	if result != "" {
		t.Errorf("Expected empty result for past time, got %s", result)
	}

	// 测试边界情况：零值时间
	result = TimeDifference(time.Time{})
	if result != "" {
		t.Errorf("Expected empty result for zero time, got %s", result)
	}

	// 测试边界情况：精确到秒
	futureTime = time.Now().Add(30 * time.Second)
	result = TimeDifference(futureTime)
	if result == "" {
		t.Errorf("Expected non-empty result for future time (30s), got empty")
	}
}

// TestDiffMinutes 测试DiffMinutes函数
func TestDiffMinutes(t *testing.T) {
	// 测试正常情况：过去时间
	pastTime := time.Now().Add(-2 * time.Hour)
	result := DiffMinutes(pastTime)
	if result < 115 || result > 125 {
		t.Errorf("Expected approximately 120 minutes, got %d", result)
	}

	// 测试正常情况：未来时间
	futureTime := time.Now().Add(2 * time.Hour)
	result = DiffMinutes(futureTime)
	if result > -115 || result < -125 {
		t.Errorf("Expected approximately -120 minutes, got %d", result)
	}

	// 测试边界情况：零值时间
	result = DiffMinutes(time.Time{})
	if result != 0 {
		t.Errorf("Expected 0 for zero time, got %d", result)
	}

	// 测试边界情况：当前时间
	result = DiffMinutes(time.Now())
	if result < -1 || result > 1 {
		t.Errorf("Expected approximately 0 minutes for current time, got %d", result)
	}
}
