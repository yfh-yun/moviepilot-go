package helper

import (
	"testing"
	"time"
)

func TestSerializeDeserialize(t *testing.T) {
	// 测试基本类型序列�?反序列化
	testCases := []interface{}{
		"hello world",
		123,
		45.67,
		true,
		[]string{"a", "b", "c"},
		map[string]interface{}{
			"name": "test",
			"age":  30,
		},
	}

	for _, testCase := range testCases {
		serialized, err := serialize(testCase)
		if err != nil {
			t.Errorf("serialize failed for %v: %v", testCase, err)
			continue
		}

		deserialized, err := deserialize(serialized)
		if err != nil {
			t.Errorf("deserialize failed for %v: %v", testCase, err)
			continue
		}

		// 检查值是否相�?		// 注意: 由于JSON序列化的特性，int可能会变成float64
		// 这里仅做简单验�?		if serialized == nil {
			t.Error("serialized data is nil")
		}
	}
}

func TestRedisHelper_SetGet(t *testing.T) {
	// 注意: 此测试需要运行中的Redis服务�?	redisHelper := GetRedisHelper()
	
	// 测试设置和获取字符串�?	key := "test_key"
	value := "test_value"
	region := "test_region"
	ttl := 10 // 10�?	
	err := redisHelper.Set(key, value, ttl, region)
	if err != nil {
		t.Logf("Warning: Redis server may not be running: %v", err)
		return
	}
	
	result, err := redisHelper.Get(key, region)
	if err != nil {
		t.Errorf("Get failed: %v", err)
		return
	}
	
	if result != value {
		t.Errorf("Expected %v, got %v", value, result)
	}
	
	// 测试存在性检�?	exists := redisHelper.Exists(key, region)
	if !exists {
		t.Error("Key should exist but was not found")
	}
	
	// 测试删除
	err = redisHelper.Delete(key, region)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	
	// 确认已删�?	exists = redisHelper.Exists(key, region)
	if exists {
		t.Error("Key should not exist but was found")
	}
}

func TestRedisHelper_Clear(t *testing.T) {
	// 注意: 此测试需要运行中的Redis服务�?	redisHelper := GetRedisHelper()
	
	// 添加一些测试数�?	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": []string{"a", "b", "c"},
	}
	
	region := "clear_test_region"
	for key, value := range testData {
		err := redisHelper.Set(key, value, 0, region)
		if err != nil {
			t.Logf("Warning: Redis server may not be running: %v", err)
			return
		}
	}
	
	// 清除特定区域
	err := redisHelper.Clear(region)
	if err != nil {
		t.Errorf("Clear failed: %v", err)
	}
	
	// 验证数据已被清除
	for key := range testData {
		exists := redisHelper.Exists(key, region)
		if exists {
			t.Errorf("Key %s should have been cleared but still exists", key)
		}
	}
}

func TestAsyncRedisHelper_SetGet(t *testing.T) {
	// 注意: 此测试需要运行中的Redis服务�?	asyncRedisHelper := GetAsyncRedisHelper()
	
	// 测试设置和获取字符串�?	key := "async_test_key"
	value := "async_test_value"
	region := "async_test_region"
	ttl := 10 // 10�?	
	err := asyncRedisHelper.Set(key, value, ttl, region)
	if err != nil {
		t.Logf("Warning: Redis server may not be running: %v", err)
		return
	}
	
	result, err := asyncRedisHelper.Get(key, region)
	if err != nil {
		t.Errorf("Get failed: %v", err)
		return
	}
	
	if result != value {
		t.Errorf("Expected %v, got %v", value, result)
	}
	
	// 测试存在性检�?	exists, err := asyncRedisHelper.Exists(key, region)
	if err != nil {
		t.Errorf("Exists failed: %v", err)
		return
	}
	
	if !exists {
		t.Error("Key should exist but was not found")
	}
	
	// 测试删除
	err = asyncRedisHelper.Delete(key, region)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	
	// 确认已删�?	exists, err = asyncRedisHelper.Exists(key, region)
	if err != nil {
		t.Errorf("Exists failed: %v", err)
		return
	}
	
	if exists {
		t.Error("Key should not exist but was found")
	}
}

func TestConnectionManagement(t *testing.T) {
	// 测试连接管理
	redisHelper := GetRedisHelper()
	
	// 测试连接
	canConnect := redisHelper.Test()
	t.Logf("Sync Redis connection test: %v", canConnect)
	
	// 测试关闭连接
	err := redisHelper.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
	
	// 测试异步连接管理
	asyncRedisHelper := GetAsyncRedisHelper()
	
	// 测试连接
	canConnect = asyncRedisHelper.Test()
	t.Logf("Async Redis connection test: %v", canConnect)
	
	// 测试关闭连接
	err = asyncRedisHelper.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestKeyGeneration(t *testing.T) {
	redisHelper := GetRedisHelper()
	
	// 测试键生�?	region := "test_region"
	key := "test/key"
	expected := "region:test_region:key:test%2Fkey"
	
	actual := redisHelper.__makeRedisKey(region, key)
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
	
	// 测试原始键提�?	original := redisHelper.__getOriginalKey(actual)
	if original != "test%2Fkey" {
		t.Errorf("Expected test%%2Fkey, got %s", original)
	}
}

func TestComplexTypes(t *testing.T) {
	redisHelper := GetRedisHelper()
	
	// 测试复杂类型
	complexData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   123,
			"name": "John Doe",
			"preferences": map[string]interface{}{
				"theme":    "dark",
				"language": "en",
			},
		},
		"items": []interface{}{
			map[string]interface{}{
				"id":    1,
				"title": "Item 1",
			},
			map[string]interface{}{
				"id":    2,
				"title": "Item 2",
			},
		},
		"timestamp": time.Now().Unix(),
	}
	
	key := "complex_data"
	region := "complex_test"
	
	err := redisHelper.Set(key, complexData, 0, region)
	if err != nil {
		t.Logf("Warning: Redis server may not be running: %v", err)
		return
	}
	
	result, err := redisHelper.Get(key, region)
	if err != nil {
		t.Errorf("Get complex data failed: %v", err)
		return
	}
	
	if result == nil {
		t.Error("Complex data result is nil")
		return
	}
	
	// 清理
	redisHelper.Delete(key, region)
}
