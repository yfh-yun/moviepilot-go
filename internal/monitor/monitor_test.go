package monitor

import (
	"testing"
	"time"
)

// TestMonitor 测试监控功能
func TestMonitor(t *testing.T) {
	// 创建监控实例
	monitor := NewMonitor()
	
	// 等待一段时间观察监控效�?	time.Sleep(5 * time.Second)
	
	// 测试快照功能
	snapshot := make(map[string]interface{})
	snapshot["test.txt"] = map[string]interface{}{
		"size":       1024,
		"modify_time": float64(time.Now().Unix()),
	}
	
	// 保存快照
	monitor.SaveSnapshot("local", snapshot, 1, nil)
	
	// 加载快照
	loadedSnapshot := monitor.LoadSnapshot("local")
	if loadedSnapshot == nil {
		t.Error("加载快照失败")
	}
	
	// 重置快照
	if !monitor.ResetSnapshot("local") {
		t.Error("重置快照失败")
	}
	
	// 停止监控
	monitor.Stop()
	
	t.Log("监控测试完成")
}
