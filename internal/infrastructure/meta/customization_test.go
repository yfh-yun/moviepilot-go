package meta

import (
	"testing"

	"moviepilot-go/pkg/cache"
)

// mockSystemConfigOper 模拟系统配置操作
type mockSystemConfigOper struct {
	customization interface{}
}

func (m *mockSystemConfigOper) Get(key string) (interface{}, error) {
	if key == "Customization" {
		return m.customization, nil
	}
	return nil, nil
}

func TestCustomizationMatcher_Match(t *testing.T) {
	// 创建缓存实例
	cache := cache.Cache("ttl", 100, 300)

	// 测试用例1: 空标题
	t.Run("EmptyTitle", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1|test2"}
		cm := NewCustomizationMatcher(systemConfig, cache)
		result := cm.Match("")
		if result != "" {
			t.Errorf("Expected empty string for empty title, got %s", result)
		}
	})

	// 测试用例2: 无匹配项
	t.Run("NoMatch", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1|test2"}
		cm := NewCustomizationMatcher(systemConfig, cache)
		result := cm.Match("no match here")
		if result != "" {
			t.Errorf("Expected empty string for no match, got %s", result)
		}
	})

	// 测试用例3: 单个匹配项
	t.Run("SingleMatch", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1|test2"}
		cm := NewCustomizationMatcher(systemConfig, cache)
		result := cm.Match("this is a test1 example")
		if result != "test1" {
			t.Errorf("Expected 'test1', got %s", result)
		}
	})

	// 测试用例4: 多个匹配项
	t.Run("MultipleMatches", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1|test2|test3"}
		cm := NewCustomizationMatcher(systemConfig, cache)
		result := cm.Match("this has test2 and test1 and test3")
		expected := "test1@test2"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例5: 重复匹配项
	t.Run("DuplicateMatches", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1|test2"}
		cm := NewCustomizationMatcher(systemConfig, cache)
		result := cm.Match("test1 test2 test1 test2")
		expected := "test1@test2"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例6: 字符串格式配置
	t.Run("StringConfig", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1\ntest2|test3"}
		cm := NewCustomizationMatcher(systemConfig, cache)
		result := cm.Match("test3 test1")
		expected := "test1"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例7: 数组格式配置
	t.Run("ArrayConfig", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: []string{"test1", "test2", "test3"}}
		cm := NewCustomizationMatcher(systemConfig, cache)
		result := cm.Match("test2")
		expected := "test2"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例8: 自定义分隔符
	t.Run("CustomSeparator", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1|test2"}
		cm := NewCustomizationMatcher(systemConfig, cache)
		cm.SetCustomSeparator("$")
		result := cm.Match("test1 test2")
		expected := "test1$test2"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例9: 缓存功能
	t.Run("CacheFunctionality", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1|test2"}
		cm := NewCustomizationMatcher(systemConfig, cache)

		// 第一次调用，应该生成缓存
		result1 := cm.Match("test1")
		if result1 != "test1" {
			t.Errorf("First call: expected 'test1', got %s", result1)
		}

		// 第二次调用，应该命中缓存
		result2 := cm.Match("test1")
		if result2 != "test1" {
			t.Errorf("Second call (cache hit): expected 'test1', got %s", result2)
		}
	})

	// 测试用例10: 清空缓存
	t.Run("ClearCache", func(t *testing.T) {
		systemConfig := &mockSystemConfigOper{customization: "test1|test2"}
		cm := NewCustomizationMatcher(systemConfig, cache)

		// 第一次调用，生成缓存
		result1 := cm.Match("test1")
		if result1 != "test1" {
			t.Errorf("First call: expected 'test1', got %s", result1)
		}

		// 清空缓存
		cm.ClearCache()

		// 再次调用，应该重新生成缓存
		result2 := cm.Match("test1")
		if result2 != "test1" {
			t.Errorf("Call after clear cache: expected 'test1', got %s", result2)
		}
	})
}

func TestCustomizationMatcher_Singleton(t *testing.T) {
	// 测试单例模式
	cache := cache.Cache("ttl", 100, 300)
	systemConfig := &mockSystemConfigOper{customization: "test1|test2"}

	instance1 := NewCustomizationMatcher(systemConfig, cache)
	instance2 := NewCustomizationMatcher(systemConfig, cache)

	if instance1 != instance2 {
		t.Error("Expected singleton instances to be the same")
	}
}
