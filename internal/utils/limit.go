package utils

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/internal/logger"
)

// RateLimitExceededException 限流异常
type RateLimitExceededException struct {
	Message string
}

func (e *RateLimitExceededException) Error() string {
	return e.Message
}

// LimitException 限流基础异常
type LimitException struct {
	Message string
}

func (e *LimitException) Error() string {
	return e.Message
}

// BaseRateLimiter 限流器基类接口，定义了限流器的通用接口
type BaseRateLimiter interface {
	// CanCall 检查是否可以进行调�?	// 返回�? 如果允许调用，返�?true 和空消息，否则返�?false 和限流消�?	CanCall() (bool, string)
	
	// Reset 重置限流状�?	Reset()
	
	// TriggerLimit 触发限流
	TriggerLimit()
	
	// RecordCall 记录一次调�?	RecordCall()
	
	// ResetOnSuccess 是否在成功调用后自动重置限流器状�?	ResetOnSuccess() bool
	
	// FormatLog 格式化日志消�?	FormatLog(message string) string
	
	// Log 根据日志级别记录日志
	Log(level string, message string)
	
	// LogInfo 记录信息日志
	LogInfo(message string)
	
	// LogWarning 记录警告日志
	LogWarning(message string)
}

// baseRateLimiter 限流器基类实�?type baseRateLimiter struct {
	source        string
	enableLogging bool
	lock          sync.Mutex
}

// NewBaseRateLimiter 创建新的基础限流�?func NewBaseRateLimiter(source string, enableLogging bool) *baseRateLimiter {
	return &baseRateLimiter{
		source:        source,
		enableLogging: enableLogging,
	}
}

// ResetOnSuccess 是否在成功调用后自动重置限流器状态，默认�?false
func (b *baseRateLimiter) ResetOnSuccess() bool {
	return false
}

// FormatLog 格式化日志消�?func (b *baseRateLimiter) FormatLog(message string) string {
	if b.source != "" {
		return fmt.Sprintf("[%s] %s", b.source, message)
	}
	return message
}

// Log 根据日志级别记录日志
func (b *baseRateLimiter) Log(level string, message string) {
	if b.enableLogging {
		log := logger.GetLoggerManager()
		switch level {
		case "debug":
			log.Debug(message)
		case "info":
			log.Info(message)
		case "warning":
			log.Warning(message)
		case "error":
			log.Error(message)
		default:
			log.Info(message)
		}
	}
}

// LogInfo 记录信息日志
func (b *baseRateLimiter) LogInfo(message string) {
	b.Log("info", message)
}

// LogWarning 记录警告日志
func (b *baseRateLimiter) LogWarning(message string) {
	b.Log("warning", message)
}

// ExponentialBackoffRateLimiter 基于指数退避的限流器，用于处理单次调用频率的控�?// 每次触发限流时，等待时间会成倍增加，直到达到最大等待时�?type ExponentialBackoffRateLimiter struct {
	*baseRateLimiter
	nextAllowedTime float64
	currentWait     float64
	baseWait        float64
	maxWait         float64
	backoffFactor   float64
}

// NewExponentialBackoffRateLimiter 创建新的指数退避限流器
func NewExponentialBackoffRateLimiter(baseWait, maxWait, backoffFactor float64, source string, enableLogging bool) *ExponentialBackoffRateLimiter {
	return &ExponentialBackoffRateLimiter{
		baseRateLimiter: NewBaseRateLimiter(source, enableLogging),
		nextAllowedTime: 0.0,
		currentWait:     baseWait,
		baseWait:        baseWait,
		maxWait:         maxWait,
		backoffFactor:   backoffFactor,
	}
}

// ResetOnSuccess 指数退避限流器在调用成功后应重置等待时�?func (e *ExponentialBackoffRateLimiter) ResetOnSuccess() bool {
	return true
}

// CanCall 检查是否可以进行调用，如果当前时间超过下一次允许调用的时间，则允许调用
func (e *ExponentialBackoffRateLimiter) CanCall() (bool, string) {
	currentTime := float64(time.Now().UnixNano()) / 1e9
	e.lock.Lock()
	defer e.lock.Unlock()
	
	if currentTime >= e.nextAllowedTime {
		return true, ""
	}
	
	waitTime := e.nextAllowedTime - currentTime
	message := fmt.Sprintf("限流期间，跳过调用，将在 %.2f 秒后允许继续调用", waitTime)
	e.LogInfo(message)
	return false, e.FormatLog(message)
}

// Reset 重置等待时间
// 当调用成功时调用此方法，重置当前等待时间为基础等待时间
func (e *ExponentialBackoffRateLimiter) Reset() {
	e.lock.Lock()
	defer e.lock.Unlock()
	
	if e.nextAllowedTime != 0 || e.currentWait > e.baseWait {
		e.LogInfo(fmt.Sprintf("调用成功，重置限流等待时间为 %.2f �?, e.baseWait))
	}
	
	e.nextAllowedTime = 0.0
	e.currentWait = e.baseWait
}

// TriggerLimit 触发限流
// 当触发限流异常时调用此方法，增加下一次允许调用的时间并更新当前等待时�?func (e *ExponentialBackoffRateLimiter) TriggerLimit() {
	currentTime := float64(time.Now().UnixNano()) / 1e9
	e.lock.Lock()
	defer e.lock.Unlock()
	
	e.nextAllowedTime = currentTime + e.currentWait
	e.currentWait = e.currentWait * e.backoffFactor
	if e.currentWait > e.maxWait {
		e.currentWait = e.maxWait
	}
	
	waitTime := e.nextAllowedTime - currentTime
	e.LogWarning(fmt.Sprintf("触发限流，将�?%.2f 秒后允许继续调用", waitTime))
}

// RecordCall 记录一次调�?func (e *ExponentialBackoffRateLimiter) RecordCall() {
	// 指数退避限流器不需要记录调�?}

// WindowRateLimiter 基于时间窗口的限流器，用于限制在特定时间窗口内的调用次数
// 如果超过允许的最大调用次数，则限流直到窗口期结束
type WindowRateLimiter struct {
	*baseRateLimiter
	maxCalls      int
	windowSeconds float64
	callTimes     *list.List // 使用链表存储调用时间�?}

// NewWindowRateLimiter 创建新的时间窗口限流�?func NewWindowRateLimiter(maxCalls int, windowSeconds float64, source string, enableLogging bool) *WindowRateLimiter {
	return &WindowRateLimiter{
		baseRateLimiter: NewBaseRateLimiter(source, enableLogging),
		maxCalls:        maxCalls,
		windowSeconds:   windowSeconds,
		callTimes:       list.New(),
	}
}

// CanCall 检查是否可以进行调用，如果在时间窗口内的调用次数少于最大允许次数，则允许调�?func (w *WindowRateLimiter) CanCall() (bool, string) {
	currentTime := float64(time.Now().UnixNano()) / 1e9
	w.lock.Lock()
	defer w.lock.Unlock()
	
	// 清理超出时间窗口的调用记�?	for e := w.callTimes.Front(); e != nil; {
		callTime := e.Value.(float64)
		next := e.Next()
		if currentTime-callTime > w.windowSeconds {
			w.callTimes.Remove(e)
		} else {
			// 由于是按时间顺序存储的，后面的元素肯定也未过�?			break
		}
		e = next
	}
	
	if w.callTimes.Len() < w.maxCalls {
		return true, ""
	}
	
	// 计算等待时间
	firstCallTime := w.callTimes.Front().Value.(float64)
	waitTime := w.windowSeconds - (currentTime - firstCallTime)
	message := fmt.Sprintf("限流期间，跳过调用，将在 %.2f 秒后允许继续调用", waitTime)
	w.LogInfo(message)
	return false, w.FormatLog(message)
}

// Reset 重置时间窗口内的调用记录
// 当调用成功时调用此方法，清空时间窗口内的调用记录
func (w *WindowRateLimiter) Reset() {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.callTimes.Init()
}

// RecordCall 记录当前时间戳，用于限流检�?func (w *WindowRateLimiter) RecordCall() {
	currentTime := float64(time.Now().UnixNano()) / 1e9
	w.lock.Lock()
	defer w.lock.Unlock()
	w.callTimes.PushBack(currentTime)
}

// TriggerLimit 触发限流
func (w *WindowRateLimiter) TriggerLimit() {
	// 时间窗口限流器不需要特殊处�?}

// ResetOnSuccess 是否在成功调用后自动重置限流器状�?func (w *WindowRateLimiter) ResetOnSuccess() bool {
	return false
}

// CompositeRateLimiter 组合限流器，可以组合多个限流策略
// 当任意一个限流策略触发限流时，都会阻止调�?type CompositeRateLimiter struct {
	*baseRateLimiter
	limiters []BaseRateLimiter
}

// NewCompositeRateLimiter 创建新的组合限流�?func NewCompositeRateLimiter(limiters []BaseRateLimiter, source string, enableLogging bool) *CompositeRateLimiter {
	return &CompositeRateLimiter{
		baseRateLimiter: NewBaseRateLimiter(source, enableLogging),
		limiters:        limiters,
	}
}

// CanCall 检查是否可以进行调用，当组合的任意限流器触发限流时，阻止调�?func (c *CompositeRateLimiter) CanCall() (bool, string) {
	for _, limiter := range c.limiters {
		canCall, message := limiter.CanCall()
		if !canCall {
			return false, message
		}
	}
	return true, ""
}

// Reset 重置所有组合的限流器状�?func (c *CompositeRateLimiter) Reset() {
	for _, limiter := range c.limiters {
		limiter.Reset()
	}
}

// RecordCall 记录所有组合的限流器的调用时间
func (c *CompositeRateLimiter) RecordCall() {
	for _, limiter := range c.limiters {
		limiter.RecordCall()
	}
}

// TriggerLimit 触发限流
func (c *CompositeRateLimiter) TriggerLimit() {
	for _, limiter := range c.limiters {
		limiter.TriggerLimit()
	}
}

// ResetOnSuccess 是否在成功调用后自动重置限流器状�?func (c *CompositeRateLimiter) ResetOnSuccess() bool {
	return false
}
