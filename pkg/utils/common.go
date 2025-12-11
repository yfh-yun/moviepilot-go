package utils

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
)

// ImmediateError 用于表示不应重试的错误，对应 Python 的 ImmediateException
// 调用方可以使用 errors.As(err, &ImmediateError{}) 判断是否为立即失败
type ImmediateError struct {
	err error
}

func (e ImmediateError) Error() string {
	if e.err == nil {
		return "immediate error"
	}
	return e.err.Error()
}

func (e ImmediateError) Unwrap() error { return e.err }

// NewImmediateError 创建一个立即失败的错误
func NewImmediateError(err error) error {
	if err == nil {
		return ImmediateError{err: errors.New("immediate error")}
	}
	return ImmediateError{err: err}
}

// RetryOptions 重试配置，对应 Python retry 装饰器参数
//   - Tries: 最大重试次数（含最后一次）
//   - Delay: 初始重试间隔
//   - Backoff: 每次重试间隔乘以的倍数
//   - Logger: 可选日志对象，需实现 Debug(msg string) / Warn(msg string) 类接口
//
// Logger 设计为 interface，调用方可传入 zap.Logger 的轻量包装，避免 pkg 依赖业务日志包。
type RetryOptions struct {
	Tries   int
	Delay   time.Duration
	Backoff float64
	Logger  interface{ Warn(msg string) }
}

// RetryFunc 被重试的函数签名
// 返回 nil 表示成功，非 nil 表示失败
// 如需中断重试，可返回 NewImmediateError 包装过的错误
type RetryFunc func() error

// Retry 执行带指数退避的重试逻辑
// 行为对应 Python retry 装饰器的同步版本：
//   - 最大重试次数为 opts.Tries（小于等于 1 时只执行一次）
//   - 首次失败后等待 opts.Delay，再按 Backoff 倍数递增
//   - 遇到 ImmediateError 时立即返回，不再重试
func Retry(fn RetryFunc, opts RetryOptions) error {
	if fn == nil {
		return errors.New("retry: fn is nil")
	}
	if opts.Tries <= 0 {
		opts.Tries = 1
	}
	if opts.Delay <= 0 {
		opts.Delay = 3 * time.Second
	}
	if opts.Backoff <= 0 {
		opts.Backoff = 2
	}

	tries := opts.Tries
	delay := opts.Delay

	for tries > 1 {
		if err := fn(); err != nil {
			var immErr ImmediateError
			if errors.As(err, &immErr) {
				// 立即失败，不再重试
				return err
			}

			msg := fmt.Sprintf("%v, %s 后重试 ...", err, delay)
			if opts.Logger != nil {
				opts.Logger.Warn(msg)
			} else {
				fmt.Println(msg)
			}

			time.Sleep(delay)
			tries--
			delay = time.Duration(float64(delay) * opts.Backoff)
			continue
		}
		// 成功
		return nil
	}

	// 最后一次尝试，直接返回结果
	return fn()
}

// RetryAsync 执行带指数退避的异步重试逻辑
// 行为对应 Python retry 装饰器的异步版本
func RetryAsync(ctx context.Context, fn func() error, opts RetryOptions) error {
	if fn == nil {
		return errors.New("retry: fn is nil")
	}
	if opts.Tries <= 0 {
		opts.Tries = 1
	}
	if opts.Delay <= 0 {
		opts.Delay = 3 * time.Second
	}
	if opts.Backoff <= 0 {
		opts.Backoff = 2
	}

	tries := opts.Tries
	delay := opts.Delay

	for tries > 1 {
		if err := fn(); err != nil {
			var immErr ImmediateError
			if errors.As(err, &immErr) {
				// 立即失败，不再重试
				return err
			}

			msg := fmt.Sprintf("%v, %s 后重试 ...", err, delay)
			if opts.Logger != nil {
				opts.Logger.Warn(msg)
			} else {
				fmt.Println(msg)
			}

			// 异步等待，支持上下文取消
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// 继续重试
			}

			tries--
			delay = time.Duration(float64(delay) * opts.Backoff)
			continue
		}
		// 成功
		return nil
	}

	// 最后一次尝试，直接返回结果
	return fn()
}

// LogExecutionTime 记录函数执行时间的辅助工具
// 用法通常是：
//
//	start := time.Now()
//	defer LogExecutionTime(logger, "FuncName", start)
//
// 若 logger 为 nil，则打印到标准输出
type debugLogger interface{ Debug(msg string) }

type debugWarnLogger interface {
	Debug(msg string)
	Warn(msg string)
}

func LogExecutionTime(logger any, name string, start time.Time) {
	duration := time.Since(start)
	msg := fmt.Sprintf("%s execution time: %.2f seconds", name, duration.Seconds())

	switch l := logger.(type) {
	case debugWarnLogger:
		l.Debug(msg)
	case debugLogger:
		l.Debug(msg)
	case nil:
		fmt.Println(msg)
	default:
		// 兜底：未知类型但不 panic
		fmt.Println(msg)
	}
}

// WithExecutionTime 包装一个无参函数，返回其执行结果并记录耗时
// 更接近 Python log_execution_time 装饰器的写法
func WithExecutionTime(logger any, name string, fn func() error) error {
	start := time.Now()
	err := fn()
	LogExecutionTime(logger, name, start)
	return err
}

// WithExecutionTimeAsync 包装一个异步无参函数，返回其执行结果并记录耗时
func WithExecutionTimeAsync(ctx context.Context, logger any, name string, fn func() error) error {
	start := time.Now()
	err := fn()
	LogExecutionTime(logger, name, start)
	return err
}

// LogExecutionTimeFunc 通用函数执行时间记录器
// 支持任意签名的函数，使用反射实现
// 示例：
//
//	result, err := LogExecutionTimeFunc(logger, "MyFunc", func() (string, error) {
//	    // 函数逻辑
//	    return "result", nil
//	})
func LogExecutionTimeFunc(logger any, name string, fn any) []any {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("LogExecutionTimeFunc: fn must be a function")
	}

	start := time.Now()

	// 调用函数，传入空参数列表
	results := fnValue.Call(nil)

	LogExecutionTime(logger, name, start)

	// 将 reflect.Value 转换为 interface{} 并返回
	interfaceResults := make([]any, len(results))
	for i, v := range results {
		interfaceResults[i] = v.Interface()
	}

	return interfaceResults
}

// LogExecutionTimeFuncWithContext 带上下文的通用函数执行时间记录器
// 支持任意签名的函数，第一个参数必须是 context.Context
// 示例：
//
//	result, err := LogExecutionTimeFuncWithContext(ctx, logger, "MyFunc", func(ctx context.Context) (string, error) {
//	    // 函数逻辑
//	    return "result", nil
//	})
func LogExecutionTimeFuncWithContext(ctx context.Context, logger any, name string, fn any) []any {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("LogExecutionTimeFuncWithContext: fn must be a function")
	}

	start := time.Now()

	// 检查函数第一个参数是否为 context.Context
	if fnValue.Type().NumIn() > 0 && fnValue.Type().In(0) == reflect.TypeOf(ctx) {
		// 调用函数，传入上下文
		results := fnValue.Call([]reflect.Value{reflect.ValueOf(ctx)})

		LogExecutionTime(logger, name, start)

		// 将 reflect.Value 转换为 interface{} 并返回
		interfaceResults := make([]any, len(results))
		for i, v := range results {
			interfaceResults[i] = v.Interface()
		}

		return interfaceResults
	} else {
		// 如果函数不接受上下文，直接调用
		return LogExecutionTimeFunc(logger, name, fn)
	}
}
