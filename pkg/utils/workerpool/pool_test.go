package workerpool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// 测试默认大小
	pool := New(0)
	assert.NotNil(t, pool)
	assert.Equal(t, 10, pool.Size())
	pool.Stop()

	// 测试自定义大小
	pool = New(20)
	assert.NotNil(t, pool)
	assert.Equal(t, 20, pool.Size())
	pool.Stop()
}

func TestPool_Submit(t *testing.T) {
	pool := New(5)
	defer pool.Stop()

	var counter int32
	var wg sync.WaitGroup

	// 提交10个任务
	for i := 0; i < 10; i++ {
		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()
			atomic.AddInt32(&counter, 1)
			time.Sleep(10 * time.Millisecond)
		})
	}

	wg.Wait()
	assert.Equal(t, int32(10), atomic.LoadInt32(&counter))
}

func TestPool_Wait(t *testing.T) {
	pool := New(5)

	var counter int32

	// 提交任务
	for i := 0; i < 100; i++ {
		pool.Submit(func() {
			atomic.AddInt32(&counter, 1)
		})
	}

	// 等待所有任务完成
	pool.Wait()

	assert.Equal(t, int32(100), atomic.LoadInt32(&counter))
}

func TestPool_Stop(t *testing.T) {
	pool := New(5)

	var counter int32

	// 提交一些任务
	for i := 0; i < 10; i++ {
		pool.Submit(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(10 * time.Millisecond)
		})
	}

	// 立即停止
	pool.Stop()

	// 停止后提交任务应该被忽略
	pool.Submit(func() {
		atomic.AddInt32(&counter, 1)
	})

	// counter 应该小于等于 10
	assert.LessOrEqual(t, atomic.LoadInt32(&counter), int32(10))
}

func TestPool_Concurrency(t *testing.T) {
	pool := New(3)
	defer pool.Stop()

	var (
		running    int32
		maxRunning int32
		mu         sync.Mutex
	)

	var wg sync.WaitGroup

	// 提交10个任务
	for i := 0; i < 10; i++ {
		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()

			// 增加运行计数
			current := atomic.AddInt32(&running, 1)

			// 更新最大运行数
			mu.Lock()
			if current > maxRunning {
				maxRunning = current
			}
			mu.Unlock()

			// 模拟工作
			time.Sleep(50 * time.Millisecond)

			// 减少运行计数
			atomic.AddInt32(&running, -1)
		})
	}

	wg.Wait()

	// 最大并发数应该不超过池大小
	assert.LessOrEqual(t, maxRunning, int32(3))
}

// 基准测试
func BenchmarkPool_Submit(b *testing.B) {
	pool := New(10)
	defer pool.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(func() {
			// 空任务
		})
	}
}

func BenchmarkPool_SubmitAndWait(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := New(10)

		for j := 0; j < 100; j++ {
			pool.Submit(func() {
				// 简单计算
				_ = 1 + 1
			})
		}

		pool.Wait()
	}
}

func BenchmarkPool_WithWork(b *testing.B) {
	pool := New(10)
	defer pool.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		pool.Submit(func() {
			defer wg.Done()
			// 模拟一些工作
			sum := 0
			for j := 0; j < 1000; j++ {
				sum += j
			}
		})
		wg.Wait()
	}
}
