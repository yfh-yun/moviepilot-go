package cache

import (
	"testing"
	"time"
)

// TestMemoryBackend 测试内存缓存后端
func TestMemoryBackend(t *testing.T) {
	// 创建配置
	config := Config{
		Type:       BackendMemory,
		DefaultTTL: 10 * time.Second,
		MaxSize:    100,
	}

	// 创建内存缓存
	backend := NewMemoryBackend(config)

	// 测试数据
	type TestData struct {
		Name  string
		Value int
	}

	testData := TestData{
		Name:  "test",
		Value: 123,
	}

	// 测试Set和Get
	err := backend.Set("test_region", "test_key", testData, 10)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	var retrieved TestData
	hit, err := backend.Get("test_region", "test_key", &retrieved)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !hit {
		t.Fatal("Get missed, expected hit")
	}

	if retrieved.Name != testData.Name || retrieved.Value != testData.Value {
		t.Fatalf("Retrieved data mismatch: got %+v, expected %+v", retrieved, testData)
	}

	// 测试Exists
	exists, err := backend.Exists("test_region", "test_key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if !exists {
		t.Fatal("Exists returned false, expected true")
	}

	// 测试Delete
	err = backend.Delete("test_region", "test_key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	exists, err = backend.Exists("test_region", "test_key")
	if err != nil {
		t.Fatalf("Exists after delete failed: %v", err)
	}

	if exists {
		t.Fatal("Exists returned true after delete, expected false")
	}

	// 测试Clear
	err = backend.Set("test_region", "key1", "value1", 10)
	if err != nil {
		t.Fatalf("Set key1 failed: %v", err)
	}

	err = backend.Set("test_region", "key2", "value2", 10)
	if err != nil {
		t.Fatalf("Set key2 failed: %v", err)
	}

	err = backend.Clear("test_region")
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	exists, err = backend.Exists("test_region", "key1")
	if err != nil {
		t.Fatalf("Exists after clear failed: %v", err)
	}

	if exists {
		t.Fatal("Exists returned true after clear, expected false")
	}

	// 测试Close
	err = backend.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestCachedFunc 测试缓存函数
func TestCachedFunc(t *testing.T) {
	// 创建配置
	config := Config{
		Type:       BackendMemory,
		DefaultTTL: 10 * time.Second,
		MaxSize:    100,
	}

	// 创建内存缓存
	backend := NewMemoryBackend(config)

	// 测试函数调用计数
	callCount := 0

	// 测试函数
	testFn := func() (string, error) {
		callCount++
		return "test_result", nil
	}

	// 缓存函数
	cachedFn := func() (string, error) {
		return CachedFunc(backend, "test_region", 5*time.Second, func() string { return "test_key" }, testFn)
	}

	// 第一次调用，应该执行函数
	result1, err := cachedFn()
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("First call: expected callCount=1, got %d", callCount)
	}

	if result1 != "test_result" {
		t.Fatalf("First call: expected 'test_result', got '%s'", result1)
	}

	// 第二次调用，应该命中缓存
	result2, err := cachedFn()
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("Second call: expected callCount=1, got %d", callCount)
	}

	if result2 != "test_result" {
		t.Fatalf("Second call: expected 'test_result', got '%s'", result2)
	}
}
