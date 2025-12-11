package utils

import (
	"strconv"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// BaseRateLimiter 定义限流器通用接口，对应 Python BaseRateLimiter
// Go 中不做装饰器，而是通过 ExecuteWithLimiter 等高阶函数调用
type BaseRateLimiter interface {
	CanCall() (bool, string)
	Reset()
	TriggerLimit()
	RecordCall()
	ResetOnSuccess() bool
}

type baseLimiter struct {
	source        string
	enableLogging bool
	mu            sync.Mutex
}

func newBaseLimiter(source string, enableLogging bool) baseLimiter {
	return baseLimiter{source: source, enableLogging: enableLogging}
}

func (b *baseLimiter) lock()   { b.mu.Lock() }
func (b *baseLimiter) unlock() { b.mu.Unlock() }

func (b *baseLimiter) formatLog(msg string) string {
	if b.source == "" {
		return msg
	}
	return "[" + b.source + "] " + msg
}

func (b *baseLimiter) logInfo(msg string) {
	if !b.enableLogging {
		return
	}
	logger.GetLogger().Info(b.formatLog(msg))
}

func (b *baseLimiter) logWarn(msg string) {
	if !b.enableLogging {
		return
	}
	logger.GetLogger().Warn(b.formatLog(msg))
}

// ExponentialBackoffRateLimiter 指数退避限流器
// 对应 Python ExponentialBackoffRateLimiter
type ExponentialBackoffRateLimiter struct {
	baseLimiter
	NextAllowedTime float64
	CurrentWait     float64
	BaseWait        float64
	MaxWait         float64
	BackoffFactor   float64
}

func NewExponentialBackoffRateLimiter(baseWait, maxWait, backoffFactor float64, source string, enableLogging bool) *ExponentialBackoffRateLimiter {
	if baseWait <= 0 {
		baseWait = 60
	}
	if maxWait <= 0 {
		maxWait = 600
	}
	if backoffFactor <= 1 {
		backoffFactor = 2
	}

	return &ExponentialBackoffRateLimiter{
		baseLimiter:     newBaseLimiter(source, enableLogging),
		NextAllowedTime: 0,
		CurrentWait:     baseWait,
		BaseWait:        baseWait,
		MaxWait:         maxWait,
		BackoffFactor:   backoffFactor,
	}
}

func (e *ExponentialBackoffRateLimiter) ResetOnSuccess() bool { return true }

func (e *ExponentialBackoffRateLimiter) CanCall() (bool, string) {
	now := float64(time.Now().UnixNano()) / 1e9 // 使用纳秒精度转换为浮点数秒
	e.lock()
	defer e.unlock()
	if now >= e.NextAllowedTime {
		return true, ""
	}
	wait := e.NextAllowedTime - now
	msg := "限流期间，跳过调用，将在 " + formatSeconds(wait) + " 秒后允许继续调用"
	e.logInfo(msg)
	return false, e.formatLog(msg)
}

func (e *ExponentialBackoffRateLimiter) Reset() {
	e.lock()
	defer e.unlock()
	if e.NextAllowedTime != 0 || e.CurrentWait > e.BaseWait {
		e.logInfo("调用成功，重置限流等待时间为 " + formatSeconds(e.BaseWait) + " 秒")
	}
	e.NextAllowedTime = 0
	e.CurrentWait = e.BaseWait
}

func (e *ExponentialBackoffRateLimiter) TriggerLimit() {
	now := float64(time.Now().UnixNano()) / 1e9 // 使用纳秒精度转换为浮点数秒
	e.lock()
	defer e.unlock()
	e.NextAllowedTime = now + e.CurrentWait
	if e.CurrentWait*e.BackoffFactor > e.MaxWait {
		e.CurrentWait = e.MaxWait
	} else {
		e.CurrentWait = e.CurrentWait * e.BackoffFactor
	}
	wait := e.NextAllowedTime - now
	e.logWarn("触发限流，将在 " + formatSeconds(wait) + " 秒后允许继续调用")
}

func (e *ExponentialBackoffRateLimiter) RecordCall() {}

// WindowRateLimiter 时间窗口限流器，对应 Python WindowRateLimiter
type WindowRateLimiter struct {
	baseLimiter
	MaxCalls      int
	WindowSeconds float64
	callTimes     []float64
}

func NewWindowRateLimiter(maxCalls int, windowSeconds float64, source string, enableLogging bool) *WindowRateLimiter {
	if maxCalls <= 0 {
		maxCalls = 1
	}
	if windowSeconds <= 0 {
		windowSeconds = 60
	}
	return &WindowRateLimiter{
		baseLimiter:   newBaseLimiter(source, enableLogging),
		MaxCalls:      maxCalls,
		WindowSeconds: windowSeconds,
		callTimes:     make([]float64, 0),
	}
}

func (w *WindowRateLimiter) ResetOnSuccess() bool { return false }

func (w *WindowRateLimiter) CanCall() (bool, string) {
	now := float64(time.Now().UnixNano()) / 1e9 // 使用纳秒精度转换为浮点数秒
	w.lock()
	defer w.unlock()

	// 清理超出窗口的调用记录
	filtered := w.callTimes[:0]
	for _, t := range w.callTimes {
		if now-t <= w.WindowSeconds {
			filtered = append(filtered, t)
		}
	}
	w.callTimes = filtered

	if len(w.callTimes) < w.MaxCalls {
		return true, ""
	}

	if len(w.callTimes) == 0 {
		return false, w.formatLog("限流，当前无有效窗口记录")
	}
	wait := w.WindowSeconds - (now - w.callTimes[0])
	msg := "限流期间，跳过调用，将在 " + formatSeconds(wait) + " 秒后允许继续调用"
	w.logInfo(msg)
	return false, w.formatLog(msg)
}

func (w *WindowRateLimiter) Reset() {
	w.lock()
	defer w.unlock()
	w.callTimes = w.callTimes[:0]
}

func (w *WindowRateLimiter) TriggerLimit() {}

func (w *WindowRateLimiter) RecordCall() {
	now := float64(time.Now().UnixNano()) / 1e9 // 使用纳秒精度转换为浮点数秒
	w.lock()
	defer w.unlock()
	w.callTimes = append(w.callTimes, now)
}

// CompositeRateLimiter 组合限流器，对应 Python CompositeRateLimiter
type CompositeRateLimiter struct {
	baseLimiter
	Limiters []BaseRateLimiter
}

func NewCompositeRateLimiter(limiters []BaseRateLimiter, source string, enableLogging bool) *CompositeRateLimiter {
	return &CompositeRateLimiter{
		baseLimiter: newBaseLimiter(source, enableLogging),
		Limiters:    limiters,
	}
}

func (c *CompositeRateLimiter) ResetOnSuccess() bool { return false }

func (c *CompositeRateLimiter) CanCall() (bool, string) {
	for _, l := range c.Limiters {
		ok, msg := l.CanCall()
		if !ok {
			return false, msg
		}
	}
	return true, ""
}

func (c *CompositeRateLimiter) Reset() {
	for _, l := range c.Limiters {
		l.Reset()
	}
}

func (c *CompositeRateLimiter) TriggerLimit() {}

func (c *CompositeRateLimiter) RecordCall() {
	for _, l := range c.Limiters {
		l.RecordCall()
	}
}

// ExecuteWithLimiter 使用限流器执行函数，语义对应 Python rate_limit_handler 装饰器
//   - 当限流时返回 (nil, errorMessage)
//   - 当函数执行报错时由调用方自行处理
func ExecuteWithLimiter[T any](limiter BaseRateLimiter, fn func() (T, error), raiseOnLimit bool) (T, error) {
	var zero T

	canCall, msg := limiter.CanCall()
	if !canCall {
		if raiseOnLimit {
			return zero, &RateLimitError{Message: msg}
		}
		return zero, nil
	}

	result, err := fn()
	if err == nil {
		limiter.RecordCall()
		if limiter.ResetOnSuccess() {
			limiter.Reset()
		}
		return result, nil
	}

	// 对应 Python 中捕获 LimitException 后触发 TriggerLimit，这里由调用方自行决定何时调用
	return result, err
}

// ExecuteWithLimiterAsync 使用限流器异步执行函数，支持异步函数的限流处理
// 对应 Python 中的 async_wrapper
func ExecuteWithLimiterAsync[T any](limiter BaseRateLimiter, fn func() (T, error), raiseOnLimit bool) (T, error) {
	// 在 Go 中，异步通常通过 goroutines 实现，这里使用同步方式实现基础功能
	// 完整的异步支持可以通过 context 和 channels 扩展
	return ExecuteWithLimiter(limiter, fn, raiseOnLimit)
}

// RateLimitExceededException 创建限流异常，对应 Python RateLimitExceededException
func RateLimitExceededException(message string) error {
	return &RateLimitError{Message: message}
}

// LimitException 基础限流异常类型，对应 Python LimitException
func LimitException(message string) error {
	return &RateLimitError{Message: message}
}

// RateLimitError 表示限流错误
type RateLimitError struct {
	Message string
}

func (e *RateLimitError) Error() string { return e.Message }

// formatSeconds 简单格式化秒数为字符串
func formatSeconds(sec float64) string {
	if sec < 0 {
		sec = 0
	}
	return strconv.FormatFloat(sec, 'f', 2, 64)
}
