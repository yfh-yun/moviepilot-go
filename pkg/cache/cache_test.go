package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMemoryBackend 测试内存缓存后端
func TestMemoryBackend(t *testing.T) {
	// 创建内存缓存配置
	config := Config{
		Type:        BackendMemory,
		DefaultTTL:  time.Second * 5, // 5秒过期
		MaxSize:     DefaultCacheSize,
		FileBaseDir: "",
		RedisURL:    "",
	}

	// 创建内存缓存后端
	backend, err := NewBackend(config)
	if err != nil {
		t.Fatalf("创建内存缓存后端失败: %v", err)
	}

	// 测试设置缓存
	key := "test_key"
	value := "test_value"
	err = backend.Set(key, value, time.Second*10, DefaultCacheRegion)
	if err != nil {
		t.Fatalf("设置缓存失败: %v", err)
	}

	// 测试获取缓存
	cachedValue, hit, err := backend.Get(key, DefaultCacheRegion)
	if err != nil {
		t.Fatalf("获取缓存失败: %v", err)
	}
	if !hit {
		t.Fatalf("缓存未命中")
	}
	if cachedValue != value {
		t.Fatalf("缓存值不匹配，期望: %v, 实际: %v", value, cachedValue)
	}

	// 测试缓存存在
	exists, err := backend.Exists(key, DefaultCacheRegion)
	if err != nil {
		t.Fatalf("检查缓存存在失败: %v", err)
	}
	if !exists {
		t.Fatalf("缓存应该存在")
	}

	// 测试获取所有缓存项
	items, err := backend.Items(DefaultCacheRegion)
	if err != nil {
		t.Fatalf("获取所有缓存项失败: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("缓存项数量不匹配，期望: 1, 实际: %v", len(items))
	}

	// 测试删除缓存
	err = backend.Delete(key, DefaultCacheRegion)
	if err != nil {
		t.Fatalf("删除缓存失败: %v", err)
	}

	// 测试缓存不存在
	exists, err = backend.Exists(key, DefaultCacheRegion)
	if err != nil {
		t.Fatalf("检查缓存存在失败: %v", err)
	}
	if exists {
		t.Fatalf("缓存应该不存在")
	}

	// 测试清空缓存
	err = backend.Clear(DefaultCacheRegion)
	if err != nil {
		t.Fatalf("清空缓存失败: %v", err)
	}

	items, err = backend.Items(DefaultCacheRegion)
	if err != nil {
		t.Fatalf("获取所有缓存项失败: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("缓存项数量不匹配，期望: 0, 实际: %v", len(items))
	}
}

// TestTTLCache 测试TTL缓存
func TestTTLCache(t *testing.T) {
	// 创建TTL缓存
	ttlCache := NewTTLCache(DefaultCacheRegion, 100, int64(time.Second*2))

	// 测试设置缓存
	key := "ttl_test_key"
	value := "ttl_test_value"
	err := ttlCache.Set(key, value, time.Second*1, WithMaxSize(100))
	if err != nil {
		t.Fatalf("设置TTL缓存失败: %v", err)
	}

	// 测试获取缓存
	cachedValue, hit, err := ttlCache.Get(key)
	if err != nil {
		t.Fatalf("获取TTL缓存失败: %v", err)
	}
	if !hit {
		t.Fatalf("TTL缓存未命中")
	}
	if cachedValue != value {
		t.Fatalf("TTL缓存值不匹配，期望: %v, 实际: %v", value, cachedValue)
	}

	// 等待缓存过期
	time.Sleep(time.Second * 3)

	// 测试缓存过期
	cachedValue, hit, err = ttlCache.Get(key)
	if err != nil {
		t.Fatalf("获取TTL缓存失败: %v", err)
	}
	if hit {
		t.Fatalf("TTL缓存应该过期")
	}
}

// TestLRUCache 测试LRU缓存
func TestLRUCache(t *testing.T) {
	// 创建LRU缓存，最大容量为2
	lruCache := NewLRUCache(DefaultCacheRegion, 2)

	// 测试设置缓存
	keys := []string{"lru_key1", "lru_key2", "lru_key3"}
	values := []string{"lru_value1", "lru_value2", "lru_value3"}

	for i := 0; i < len(keys); i++ {
		err := lruCache.Set(keys[i], values[i], 0, WithMaxSize(2))
		if err != nil {
			t.Fatalf("设置LRU缓存失败: %v", err)
		}
	}

	// 测试LRU淘汰机制
	// 由于容量为2，第一个键应该被淘汰
	_, hit, err := lruCache.Get(keys[0])
	if err != nil {
		t.Fatalf("获取LRU缓存失败: %v", err)
	}
	if hit {
		t.Fatalf("LRU缓存应该淘汰第一个键")
	}

	// 测试其他键应该存在
	for i := 1; i < len(keys); i++ {
		_, hit, err := lruCache.Get(keys[i])
		if err != nil {
			t.Fatalf("获取LRU缓存失败: %v", err)
		}
		if !hit {
			t.Fatalf("LRU缓存应该命中键: %v", keys[i])
		}
	}
}

// TestFileBackend 测试文件缓存后端
func TestFileBackend(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "cache_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建文件缓存配置
	config := Config{
		Type:        BackendFile,
		DefaultTTL:  time.Second * 5,
		MaxSize:     DefaultCacheSize,
		FileBaseDir: tempDir,
		RedisURL:    "",
	}

	// 创建文件缓存后端
	backend, err := NewBackend(config)
	if err != nil {
		t.Fatalf("创建文件缓存后端失败: %v", err)
	}

	// 测试设置缓存
	key := "file_test_key"
	value := "file_test_value"
	err = backend.Set(key, value, time.Second*10, DefaultCacheRegion)
	if err != nil {
		t.Fatalf("设置文件缓存失败: %v", err)
	}

	// 测试文件是否创建
	// 文件缓存使用MD5哈希生成文件名，所以我们无法直接预测文件名
	// 检查目录是否存在即可
	dirPath := filepath.Join(tempDir, DefaultCacheRegion)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Fatalf("文件缓存目录未创建: %v", dirPath)
	}

	// 测试获取缓存
	cachedValue, hit, err := backend.Get(key, DefaultCacheRegion)
	if err != nil {
		t.Fatalf("获取文件缓存失败: %v", err)
	}
	if !hit {
		t.Fatalf("文件缓存未命中")
	}
	if cachedValue != value {
		t.Fatalf("文件缓存值不匹配，期望: %v, 实际: %v", value, cachedValue)
	}

	// 测试删除缓存
	err = backend.Delete(key, DefaultCacheRegion)
	if err != nil {
		t.Fatalf("删除文件缓存失败: %v", err)
	}

	// 测试清空缓存
	err = backend.Clear(DefaultCacheRegion)
	if err != nil {
		t.Fatalf("清空文件缓存失败: %v", err)
	}
}

// TestCacheFactory 测试缓存工厂函数
func TestCacheFactory(t *testing.T) {
	// 测试Cache函数
	cacheBackend := Cache("ttl", DefaultCacheSize, 10) // 10秒过期
	if cacheBackend == nil {
		t.Fatalf("Cache函数返回nil")
	}

	// 测试AsyncCache函数
	asyncCacheBackend := AsyncCache("lru", DefaultCacheSize, 0) // LRU缓存
	if asyncCacheBackend == nil {
		t.Fatalf("AsyncCache函数返回nil")
	}

	// 测试FileCache函数
	fileCacheBackend := FileCache("/tmp", 5)
	if fileCacheBackend == nil {
		t.Fatalf("FileCache函数返回nil")
	}

	// 测试AsyncFileCache函数
	asyncFileCacheBackend := AsyncFileCache("/tmp", 5)
	if asyncFileCacheBackend == nil {
		t.Fatalf("AsyncFileCache函数返回nil")
	}
}

// TestCacheProxy 测试缓存代理
func TestCacheProxy(t *testing.T) {
	// 创建内存缓存配置
	config := Config{
		Type:        BackendMemory,
		DefaultTTL:  time.Second * 5,
		MaxSize:     DefaultCacheSize,
		FileBaseDir: "",
		RedisURL:    "",
	}

	// 创建内存缓存后端
	backend, err := NewBackend(config)
	if err != nil {
		t.Fatalf("创建内存缓存后端失败: %v", err)
	}

	// 创建缓存代理
	proxy := NewCacheProxy(backend, DefaultCacheRegion)

	// 测试设置缓存
	key := "proxy_test_key"
	value := "proxy_test_value"
	err = proxy.Set(key, value, time.Second*10)
	if err != nil {
		t.Fatalf("设置缓存代理失败: %v", err)
	}

	// 测试获取缓存
	cachedValue, hit, err := proxy.Get(key)
	if err != nil {
		t.Fatalf("获取缓存代理失败: %v", err)
	}
	if !hit {
		t.Fatalf("缓存代理未命中")
	}
	if cachedValue != value {
		t.Fatalf("缓存代理值不匹配，期望: %v, 实际: %v", value, cachedValue)
	}

	// 测试删除缓存
	err = proxy.Delete(key)
	if err != nil {
		t.Fatalf("删除缓存代理失败: %v", err)
	}

	// 测试清空缓存
	err = proxy.Clear()
	if err != nil {
		t.Fatalf("清空缓存代理失败: %v", err)
	}
}

// TestPopAndPopitem 测试Pop和Popitem方法
func TestPopAndPopitem(t *testing.T) {
	// 创建内存缓存配置
	config := Config{
		Type:        BackendMemory,
		DefaultTTL:  0, // 永不过期
		MaxSize:     DefaultCacheSize,
		FileBaseDir: "",
		RedisURL:    "",
	}

	// 创建内存缓存后端
	backend, err := NewBackend(config)
	if err != nil {
		t.Fatalf("创建内存缓存后端失败: %v", err)
	}

	// 测试设置多个缓存
	keys := []string{"pop_key1", "pop_key2", "pop_key3"}
	values := []string{"pop_value1", "pop_value2", "pop_value3"}

	for i := 0; i < len(keys); i++ {
		err := backend.Set(keys[i], values[i], 0, DefaultCacheRegion)
		if err != nil {
			t.Fatalf("设置缓存失败: %v", err)
		}
	}

	// 测试Pop方法
	poppedValue, err := backend.Pop(keys[0], DefaultCacheRegion)
	if err != nil {
		t.Fatalf("Pop方法失败: %v", err)
	}
	if poppedValue != values[0] {
		t.Fatalf("Pop值不匹配，期望: %v, 实际: %v", values[0], poppedValue)
	}

	// 测试Pop不存在的键
	defaultValue := "default_value"
	poppedValue, err = backend.Pop("non_existent_key", DefaultCacheRegion, defaultValue)
	if err != nil {
		t.Fatalf("Pop不存在的键失败: %v", err)
	}
	if poppedValue != defaultValue {
		t.Fatalf("Pop默认值不匹配，期望: %v, 实际: %v", defaultValue, poppedValue)
	}

	// 测试Popitem方法
	key, value, err := backend.Popitem(DefaultCacheRegion)
	if err != nil {
		t.Fatalf("Popitem方法失败: %v", err)
	}
	if key == "" || value == nil {
		t.Fatalf("Popitem返回空值")
	}
}

// TestSetdefault 测试Setdefault方法
func TestSetdefault(t *testing.T) {
	// 创建内存缓存配置
	config := Config{
		Type:        BackendMemory,
		DefaultTTL:  0, // 永不过期
		MaxSize:     DefaultCacheSize,
		FileBaseDir: "",
		RedisURL:    "",
	}

	// 创建内存缓存后端
	backend, err := NewBackend(config)
	if err != nil {
		t.Fatalf("创建内存缓存后端失败: %v", err)
	}

	// 测试Setdefault不存在的键
	key := "setdefault_key"
	defaultValue := "setdefault_value"
	value, err := backend.Setdefault(key, defaultValue, DefaultCacheRegion, 0)
	if err != nil {
		t.Fatalf("Setdefault方法失败: %v", err)
	}
	if value != defaultValue {
		t.Fatalf("Setdefault值不匹配，期望: %v, 实际: %v", defaultValue, value)
	}

	// 测试Setdefault已存在的键
	value, err = backend.Setdefault(key, "new_value", DefaultCacheRegion, 0)
	if err != nil {
		t.Fatalf("Setdefault方法失败: %v", err)
	}
	if value != defaultValue {
		t.Fatalf("Setdefault应该返回原有的值，期望: %v, 实际: %v", defaultValue, value)
	}
}
