package utils

import (
	"sync"
)

// Singleton 按参数的单例模式接口
type Singleton interface {
	// Instance 获取实例的通用方法
	Instance() interface{}
}

// singletonManager 单例管理�?type singletonManager struct {
	// 按参数区分的实例映射
	instances map[interface{}]interface{}
	// 按类区分的实例映�?	classInstances map[interface{}]interface{}
	// 弱引用实例映射（Go中通过手动管理实现类似效果�?	weakInstances map[interface{}]interface{}
	
	// 用于保护并发访问的互斥锁
	mutex sync.RWMutex
	// 弱引用单例的专用�?	weakMutex sync.RWMutex
}

// singletonMgr 全局单例管理器实�?var singletonMgr *singletonManager
var once sync.Once

// getSingletonManager 获取单例管理器实�?func getSingletonManager() *singletonManager {
	once.Do(func() {
		singletonMgr = &singletonManager{
			instances:      make(map[interface{}]interface{}),
			classInstances: make(map[interface{}]interface{}),
			weakInstances:  make(map[interface{}]interface{}),
		}
	})
	return singletonMgr
}

// SingletonByKey 按参数的单例模式
// 类似于Python的Singleton类，根据类、参数和关键字参数创建唯一实例
func SingletonByKey(key interface{}, constructor func() interface{}) interface{} {
	manager := getSingletonManager()
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	if instance, exists := manager.instances[key]; exists {
		return instance
	}
	
	instance := constructor()
	manager.instances[key] = instance
	return instance
}

// SingletonByClass 按类的单例模�?// 类似于Python的SingletonClass类，每个类只有一个实例，不考虑参数
func SingletonByClass(class interface{}, constructor func() interface{}) interface{} {
	manager := getSingletonManager()
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	if instance, exists := manager.classInstances[class]; exists {
		return instance
	}
	
	instance := constructor()
	manager.classInstances[class] = instance
	return instance
}

// WeakSingleton 弱引用单例模�?// 类似于Python的WeakSingleton类，使用读写锁保护并发访�?func WeakSingleton(class interface{}, constructor func() interface{}) interface{} {
	manager := getSingletonManager()
	manager.weakMutex.Lock()
	defer manager.weakMutex.Unlock()
	
	if instance, exists := manager.weakInstances[class]; exists {
		return instance
	}
	
	instance := constructor()
	manager.weakInstances[class] = instance
	return instance
}

// CleanupWeakInstances 清理弱引用实�?// 在适当的时候手动调用此方法来清理不需要的实例
func CleanupWeakInstances() {
	manager := getSingletonManager()
	manager.weakMutex.Lock()
	defer manager.weakMutex.Unlock()
	
	// 在Go中需要手动清理，这里简单清空弱引用实例映射
	// 实际使用中可以根据具体需求实现更复杂的清理逻辑
	manager.weakInstances = make(map[interface{}]interface{})
}

// GetInstanceCount 获取当前管理的实例数�?func (s *singletonManager) GetInstanceCount() (int, int, int) {
	s.mutex.RLock()
	s.weakMutex.RLock()
	defer s.mutex.RUnlock()
	defer s.weakMutex.RUnlock()
	
	return len(s.instances), len(s.classInstances), len(s.weakInstances)
}
