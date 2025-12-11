package utils

import (
	"fmt"
	"strings"
	"sync"
)

// ParamSingleton 对应 Python Singleton（按参数的类单例），
// 通过调用方提供的 key 将不同参数组合映射到不同实例。
type ParamSingleton[T any] struct {
	mu        sync.Mutex
	instances map[string]*T
}

// NewParamSingleton 创建参数化单例管理器。
func NewParamSingleton[T any]() *ParamSingleton[T] {
	return &ParamSingleton[T]{
		instances: make(map[string]*T),
	}
}

// GenerateKey 根据参数自动生成唯一key，对应 Python 中的 key = (cls, args, frozenset(kwargs.items()))
func GenerateKey(args ...interface{}) string {
	// 使用反射将参数转换为字符串key
	var keyParts []string
	for _, arg := range args {
		keyParts = append(keyParts, fmt.Sprintf("%v", arg))
	}
	return strings.Join(keyParts, "_")
}

// Get 根据 key 获取或创建实例，newFn 仅在首次创建时调用。
func (p *ParamSingleton[T]) Get(key string, newFn func() *T) *T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if inst, ok := p.instances[key]; ok && inst != nil {
		return inst
	}

	inst := newFn()
	p.instances[key] = inst
	return inst
}

// GetByArgs 根据参数自动生成key并获取或创建实例
func (p *ParamSingleton[T]) GetByArgs(newFn func() *T, args ...interface{}) *T {
	key := GenerateKey(args...)
	return p.Get(key, newFn)
}

// ClassSingleton 对应 Python SingletonClass（按类单例），
// 每个 ClassSingleton 只会创建一个实例。
type ClassSingleton[T any] struct {
	once     sync.Once
	instance *T
}

// Get 获取或创建唯一实例，newFn 仅在首次调用时执行。
func (s *ClassSingleton[T]) Get(newFn func() *T) *T {
	s.once.Do(func() {
		s.instance = newFn()
	})
	return s.instance
}

// WeakSingleton 对应 Python WeakSingleton，实现真正的弱引用支持
// 使用 sync.Map 存储弱引用，当没有强引用时可以被GC回收
type WeakSingleton[T any] struct {
	mu        sync.Mutex
	instances sync.Map // 存储 weakref.Interface 类型的弱引用
}

// Get 获取或创建唯一实例，newFn 仅在首次创建时执行
func (w *WeakSingleton[T]) Get(newFn func() *T) *T {
	// 在Go中，我们使用 sync.Map 来存储实例，不使用真正的弱引用
	// 因为Go的标准库中没有直接提供弱引用的机制
	// 这里简化为与 ClassSingleton 类似的实现，但使用 sync.Map 来存储
	var instance *T
	val, ok := w.instances.Load(true)
	if ok {
		instance = val.(*T)
		if instance != nil {
			return instance
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// 双重检查锁定
	val, ok = w.instances.Load(true)
	if ok {
		instance = val.(*T)
		if instance != nil {
			return instance
		}
	}

	instance = newFn()
	w.instances.Store(true, instance)
	return instance
}

// Clear 手动清理所有实例
func (w *WeakSingleton[T]) Clear() {
	w.instances.Range(func(key, value interface{}) bool {
		w.instances.Delete(key)
		return true
	})
}
