package cache

import (
	"time"
)

// KeyBuilder 键生成器函数类型
// 用于根据函数参数生成缓存键
// 原Python: 对应@cached装饰器中自动生成key的逻辑
type KeyBuilder func() string

// CachedFunc 缓存函数结果
// 原Python: @cached装饰器
// backend: 缓存后端
// region: 缓存区域
// ttl: 过期时间
// keyBuilder: 键生成器
// fn: 要缓存的函数
func CachedFunc[T any](backend Backend, region string, ttl time.Duration, keyBuilder KeyBuilder, fn func() (T, error)) (T, error) {
	var zero T

	// 生成缓存键
	key := keyBuilder()

	// 尝试从缓存获取
	var cached T
	hit, err := backend.Get(region, key, &cached)
	if err != nil {
		// 缓存读取失败，执行函数并返回结果
		return fn()
	}

	if hit {
		// 缓存命中，返回结果
		return cached, nil
	}

	// 缓存未命中，执行函数
	res, err := fn()
	if err != nil {
		return zero, err
	}

	// 将结果存入缓存
	_ = backend.Set(region, key, res, int64(ttl.Seconds()))

	return res, nil
}

// CachedFunc2 缓存带两个参数的函数结果
// 原Python: @cached装饰器的泛型版本
func CachedFunc2[T1, T2, R any](backend Backend, region string, ttl time.Duration, keyBuilder func(T1, T2) string, fn func(T1, T2) (R, error)) func(T1, T2) (R, error) {
	return func(arg1 T1, arg2 T2) (R, error) {
		var zero R

		// 生成缓存键
		key := keyBuilder(arg1, arg2)

		// 尝试从缓存获取
		var cached R
		hit, err := backend.Get(region, key, &cached)
		if err != nil {
			// 缓存读取失败，执行函数并返回结果
			return fn(arg1, arg2)
		}

		if hit {
			// 缓存命中，返回结果
			return cached, nil
		}

		// 缓存未命中，执行函数
		res, err := fn(arg1, arg2)
		if err != nil {
			return zero, err
		}

		// 将结果存入缓存
		_ = backend.Set(region, key, res, int64(ttl.Seconds()))

		return res, nil
	}
}

// CachedFunc3 缓存带三个参数的函数结果
func CachedFunc3[T1, T2, T3, R any](backend Backend, region string, ttl time.Duration, keyBuilder func(T1, T2, T3) string, fn func(T1, T2, T3) (R, error)) func(T1, T2, T3) (R, error) {
	return func(arg1 T1, arg2 T2, arg3 T3) (R, error) {
		var zero R

		// 生成缓存键
		key := keyBuilder(arg1, arg2, arg3)

		// 尝试从缓存获取
		var cached R
		hit, err := backend.Get(region, key, &cached)
		if err != nil {
			// 缓存读取失败，执行函数并返回结果
			return fn(arg1, arg2, arg3)
		}

		if hit {
			// 缓存命中，返回结果
			return cached, nil
		}

		// 缓存未命中，执行函数
		res, err := fn(arg1, arg2, arg3)
		if err != nil {
			return zero, err
		}

		// 将结果存入缓存
		_ = backend.Set(region, key, res, int64(ttl.Seconds()))

		return res, nil
	}
}
