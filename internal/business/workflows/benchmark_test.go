package workflows

import (
	"context"
	"testing"

	"moviepilot-go/internal/business/workflows/registry"
	"moviepilot-go/internal/business/workflows/types"
)

// BenchmarkActionCreation 基准测试动作创建
func BenchmarkActionCreation(b *testing.B) {
	reg := registry.GetDefaultRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		action, err := reg.CreateAction("file_scanner")
		if err != nil {
			b.Fatal(err)
		}
		_ = action
	}
}

// BenchmarkActionExecution 基准测试动作执行
func BenchmarkActionExecution(b *testing.B) {
	reg := registry.GetDefaultRegistry()
	action, err := reg.CreateAction("file_scanner")
	if err != nil {
		b.Fatal(err)
	}

	err = action.Initialize()
	if err != nil {
		b.Fatal(err)
	}
	defer action.Cleanup()

	params := map[string]interface{}{
		"scan_path": []string{"/tmp"},
		"limit":     10,
	}

	actionContext := &types.ActionContext{
		WorkflowID: 12345,
		Variables:  make(map[string]interface{}),
		Metadata:   make(map[string]string),
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := action.Execute(ctx, int64(i), params, actionContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRegistryOperations 基准测试注册表操作
func BenchmarkRegistryOperations(b *testing.B) {
	reg := registry.NewActionRegistry()

	// 注册测试动作
	err := reg.Register("benchmark_action", func() interfaces.Action {
		return &TestAction{}
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// 基准测试获取动作工厂
	b.Run("GetFactory", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := reg.GetFactory("benchmark_action")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 基准测试创建动作
	b.Run("CreateAction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := reg.CreateAction("benchmark_action")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 基准测试列出动作
	b.Run("ListActions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = reg.ListActions()
		}
	})
}

// BenchmarkCaching 基准测试缓存操作
func BenchmarkCaching(b *testing.B) {
	reg := registry.GetDefaultRegistry()
	action, err := reg.CreateAction("file_scanner")
	if err != nil {
		b.Fatal(err)
	}

	err = action.Initialize()
	if err != nil {
		b.Fatal(err)
	}
	defer action.Cleanup()

	ctx := context.Background()
	workflowID := int64(12345)

	// 预先设置缓存
	err = action.SaveCache(ctx, workflowID, "test_key", "test_value", 0)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	// 基准测试检查缓存
	b.Run("CheckCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = action.CheckCache(ctx, workflowID, "test_key")
		}
	})

	// 基准测试获取缓存
	b.Run("GetCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = action.GetCache(ctx, workflowID, "test_key")
		}
	})
}