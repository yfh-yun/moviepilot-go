package main

import (
	"fmt"
	"sync"

	"moviepilot-go/internal/utils"
)

// 示例结构�?type Database struct {
	Host string
	Port int
	Name string
}

// 示例结构体构造函�?func NewDatabase(host string, port int, name string) *Database {
	return &Database{
		Host: host,
		Port: port,
		Name: name,
	}
}

// 示例结构体（按类单例�?type Logger struct {
	Level string
}

// 示例结构体构造函数（按类单例�?func NewLogger() *Logger {
	return &Logger{
		Level: "INFO",
	}
}

// 示例结构体（弱引用单例）
type Cache struct {
	Data map[string]interface{}
}

// 示例结构体构造函数（弱引用单例）
func NewCache() *Cache {
	return &Cache{
		Data: make(map[string]interface{}),
	}
}

func main() {
	fmt.Println("=== 单例模式示例 ===")

	// 测试按参数的单例模式
	fmt.Println("\n--- 按参数的单例模式 ---")
	testSingletonByKey()

	// 测试按类的单例模�?	fmt.Println("\n--- 按类的单例模�?---")
	testSingletonByClass()

	// 测试弱引用单例模�?	fmt.Println("\n--- 弱引用单例模�?---")
	testWeakSingleton()

	// 测试并发安全
	fmt.Println("\n--- 并发安全测试 ---")
	testConcurrency()
}

func testSingletonByKey() {
	// 创建相同参数的实�?	db1 := utils.SingletonByKey("db1", func() interface{} {
		return NewDatabase("localhost", 3306, "test")
	}).(*Database)

	db2 := utils.SingletonByKey("db1", func() interface{} {
		return NewDatabase("localhost", 3306, "test")
	}).(*Database)

	// 验证是否为同一实例
	fmt.Printf("db1 == db2: %v\n", db1 == db2)
	fmt.Printf("db1 地址: %p\n", db1)
	fmt.Printf("db2 地址: %p\n", db2)

	// 创建不同参数的实�?	db3 := utils.SingletonByKey("db2", func() interface{} {
		return NewDatabase("localhost", 3306, "production")
	}).(*Database)

	fmt.Printf("db1 == db3: %v\n", db1 == db3)
	fmt.Printf("db3 地址: %p\n", db3)

	// 创建带不同参数的实例
	db4 := utils.SingletonByKey("db1_with_port", func() interface{} {
		return NewDatabase("localhost", 5432, "test")
	}).(*Database)

	fmt.Printf("db1 == db4: %v\n", db1 == db4)
	fmt.Printf("db4 地址: %p\n", db4)
}

func testSingletonByClass() {
	// 创建Logger实例
	logger1 := utils.SingletonByClass("Logger", func() interface{} {
		return NewLogger()
	}).(*Logger)

	logger2 := utils.SingletonByClass("Logger", func() interface{} {
		return NewLogger()
	}).(*Logger)

	// 验证是否为同一实例（按类单例）
	fmt.Printf("logger1 == logger2: %v\n", logger1 == logger2)
	fmt.Printf("logger1 地址: %p\n", logger1)
	fmt.Printf("logger2 地址: %p\n", logger2)

	// 创建不同类的实例
	db1 := utils.SingletonByClass("Database", func() interface{} {
		return NewDatabase("localhost", 3306, "test")
	}).(*Database)

	fmt.Printf("logger1 == db1: %v (不同类型)\n", logger1 == db1)
	fmt.Printf("db1 地址: %p\n", db1)
}

func testWeakSingleton() {
	// 创建Cache实例
	cache1 := utils.WeakSingleton("Cache", func() interface{} {
		return NewCache()
	}).(*Cache)

	cache2 := utils.WeakSingleton("Cache", func() interface{} {
		return NewCache()
	}).(*Cache)

	// 验证是否为同一实例
	fmt.Printf("cache1 == cache2: %v\n", cache1 == cache2)
	fmt.Printf("cache1 地址: %p\n", cache1)
	fmt.Printf("cache2 地址: %p\n", cache2)

	// 添加一些数�?	cache1.Data["key1"] = "value1"
	fmt.Printf("cache2.Data['key1']: %v\n", cache2.Data["key1"])

	// 清理弱引用实�?	utils.CleanupWeakInstances()
	
	// 创建新的实例（应该是一个新实例�?	cache3 := utils.WeakSingleton("Cache", func() interface{} {
		return NewCache()
	}).(*Cache)
	
	fmt.Printf("cache2 == cache3: %v\n", cache2 == cache3)
	fmt.Printf("cache3 地址: %p\n", cache3)
}

func testConcurrency() {
	var wg sync.WaitGroup
	const numGoroutines = 10
	
	// 测试按参数的单例在并发环境下的安全�?	instances := make([]*Database, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			instance := utils.SingletonByKey("concurrent_db", func() interface{} {
				return NewDatabase("localhost", 3306, "concurrent_test")
			}).(*Database)
			instances[index] = instance
		}(i)
	}
	
	wg.Wait()
	
	// 验证所有goroutine获取的是同一实例
	firstInstance := instances[0]
	allSame := true
	for i := 1; i < numGoroutines; i++ {
		if instances[i] != firstInstance {
			allSame = false
			break
		}
	}
	
	fmt.Printf("并发环境下按参数单例是否一�? %v\n", allSame)
	fmt.Printf("第一个实例地址: %p\n", firstInstance)
	
	// 测试按类的单例在并发环境下的安全�?	classInstances := make([]*Logger, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			instance := utils.SingletonByClass("Logger", func() interface{} {
				return NewLogger()
			}).(*Logger)
			classInstances[index] = instance
		}(i)
	}
	
	wg.Wait()
	
	// 验证所有goroutine获取的是同一实例
	firstClassInstance := classInstances[0]
	allClassSame := true
	for i := 1; i < numGoroutines; i++ {
		if classInstances[i] != firstClassInstance {
			allClassSame = false
			break
		}
	}
	
	fmt.Printf("并发环境下按类单例是否一�? %v\n", allClassSame)
	fmt.Printf("第一个类实例地址: %p\n", firstClassInstance)
}
