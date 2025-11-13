package utils

import (
	"reflect"

	"moviepilot-go/internal/logger"
)

// RateLimitHandler 通用装饰器，允许用户传递自定义的限流器实例，用于处理限流逻辑
// 该装饰器可灵活支持任意实�?BaseRateLimiter 接口的限流器
type RateLimitHandler struct {
	limiter      BaseRateLimiter
	raiseOnLimit bool
}

// NewRateLimitHandler 创建新的限流处理�?func NewRateLimitHandler(limiter BaseRateLimiter, raiseOnLimit bool) *RateLimitHandler {
	return &RateLimitHandler{
		limiter:      limiter,
		raiseOnLimit: raiseOnLimit,
	}
}

// Handle 同步函数处理
func (r *RateLimitHandler) Handle(fn interface{}, args ...interface{}) (interface{}, error) {
	// 检查是否可以进行调�?	canCall, message := r.limiter.CanCall()
	if !canCall {
		// 如果调用受限，并�?raiseOnLimit �?true，则抛出限流异常
		if r.raiseOnLimit {
			return nil, &RateLimitExceededException{Message: message}
		}
		// 如果不抛出异常，则返�?nil 表示跳过调用
		return nil, nil
	}

	// 如果调用允许，执行目标函数，并记录一次调�?	result, err := r.executeFunction(fn, args...)
	if err != nil {
		// 检查是否为限流相关的异�?		if r.isLimitException(err) {
			// 如果目标函数触发了限流相关的异常，执行限流器的触发逻辑
			r.limiter.TriggerLimit()
			log := logger.GetLoggerManager()
			log.Error(r.limiter.FormatLog("触发限流�? + err.Error()))
			
			// 如果 raiseOnLimit �?true，则抛出异常，否则返�?nil
			if r.raiseOnLimit {
				return nil, err
			}
			return nil, nil
		}
		// 其他异常直接返回
		return nil, err
	}

	// 记录调用并根据需要重�?	r.limiter.RecordCall()
	if r.limiter.ResetOnSuccess() {
		r.limiter.Reset()
	}

	return result, nil
}

// executeFunction 执行函数
func (r *RateLimitHandler) executeFunction(fn interface{}, args ...interface{}) (interface{}, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// 检查函数签�?	if fnType.Kind() != reflect.Func {
		return nil, &LimitException{Message: "提供的参数不是函�?}
	}

	// 准备参数
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// 调用函数
	results := fnValue.Call(in)

	// 处理返回�?	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		return results[0].Interface(), nil
	case 2:
		// 假设第二个返回值是错误
		var err error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}
		return results[0].Interface(), err
	default:
		// 多于两个返回值的情况，只处理前两�?		var err error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}
		return results[0].Interface(), err
	}
}

// isLimitException 检查是否为限流相关异常
func (r *RateLimitHandler) isLimitException(err error) bool {
	_, isLimit := err.(*LimitException)
	_, isRateLimit := err.(*RateLimitExceededException)
	return isLimit || isRateLimit
}

// RateLimitExponential 装饰器工厂函数，用于应用指数退避限流策�?// 通过逐渐增加调用等待时间控制调用频率。每次触发限流时，等待时间会成倍增加，直到达到最大等待时�?func RateLimitExponential(baseWait, maxWait, backoffFactor float64, raiseOnLimit bool, source string, enableLogging bool) *RateLimitHandler {
	// 实例�?ExponentialBackoffRateLimiter，并传入相关参数
	limiter := NewExponentialBackoffRateLimiter(baseWait, maxWait, backoffFactor, source, enableLogging)
	// 使用通用装饰器逻辑包装该限流器
	return NewRateLimitHandler(limiter, raiseOnLimit)
}

// RateLimitWindow 装饰器工厂函数，用于应用时间窗口限流策略
// 在固定的时间窗口内限制调用次数，当调用次数超过最大值时，触发限流，直到时间窗口结束
func RateLimitWindow(maxCalls int, windowSeconds float64, raiseOnLimit bool, source string, enableLogging bool) *RateLimitHandler {
	// 实例�?WindowRateLimiter，并传入相关参数
	limiter := NewWindowRateLimiter(maxCalls, windowSeconds, source, enableLogging)
	// 使用通用装饰器逻辑包装该限流器
	return NewRateLimitHandler(limiter, raiseOnLimit)
}

// RateLimitComposite 装饰器工厂函数，用于应用组合限流策略
func RateLimitComposite(limiters []BaseRateLimiter, raiseOnLimit bool, source string, enableLogging bool) *RateLimitHandler {
	// 实例�?CompositeRateLimiter，并传入相关参数
	limiter := NewCompositeRateLimiter(limiters, source, enableLogging)
	// 使用通用装饰器逻辑包装该限流器
	return NewRateLimitHandler(limiter, raiseOnLimit)
}
