package helper

import (
	"moviepilot-go/pkg/models"
	"testing"
)

func TestProgressHelper(t *testing.T) {
	// 测试创建ProgressHelper实例
	t.Run("创建ProgressHelper实例", func(t *testing.T) {
		progressHelper := NewProgressHelper(models.ProgressKeySearch)
		if progressHelper == nil {
			t.Error("无法创建ProgressHelper实例")
		}
		
		// 测试单例模式
		progressHelper2 := NewProgressHelper(models.ProgressKeySearch)
		if progressHelper != progressHelper2 {
			t.Error("ProgressHelper应该使用单例模式")
		}
	})

	// 测试Start方法
	t.Run("测试Start方法", func(t *testing.T) {
		progressHelper := NewProgressHelper("test-key")
		progressHelper.Start()
		
		data := progressHelper.Get()
		if data == nil {
			t.Error("Start后应该能获取到进度数�?)
			return
		}
		
		if !data.Enable {
			t.Error("Start后进度应该启�?)
		}
		
		if data.Value != 0 {
			t.Error("Start后进度值应该为0")
		}
		
		if data.Text != "请稍�?.." {
			t.Error("Start后进度文本应该为'请稍�?..'")
		}
	})

	// 测试End方法
	t.Run("测试End方法", func(t *testing.T) {
		progressHelper := NewProgressHelper("test-key-2")
		progressHelper.Start()
		progressHelper.End()
		
		data := progressHelper.Get()
		if data == nil {
			t.Error("End后应该能获取到进度数�?)
			return
		}
		
		if data.Enable {
			t.Error("End后进度应该禁�?)
		}
		
		if data.Value != 100 {
			t.Error("End后进度值应该为100")
		}
		
		if data.Text != "" {
			t.Error("End后进度文本应该为�?)
		}
	})

	// 测试Update方法
	t.Run("测试Update方法", func(t *testing.T) {
		progressHelper := NewProgressHelper("test-key-3")
		progressHelper.Start()
		
		// 更新进度�?		value := 50.0
		progressHelper.Update(&value, nil, nil)
		
		data := progressHelper.Get()
		if data == nil {
			t.Error("Update后应该能获取到进度数�?)
			return
		}
		
		if data.Value != 50.0 {
			t.Errorf("Update后进度值应该为50.0，实际为%f", data.Value)
		}
		
		// 更新进度文本
		text := "处理�?.."
		progressHelper.Update(nil, &text, nil)
		
		data = progressHelper.Get()
		if data.Text != "处理�?.." {
			t.Errorf("Update后进度文本应该为'处理�?..'，实际为'%s'", data.Text)
		}
		
		// 更新进度数据
		updateData := map[string]interface{}{
			"file": "test.txt",
			"size": 1024,
		}
		progressHelper.Update(nil, nil, updateData)
		
		data = progressHelper.Get()
		if data.Data == nil {
			t.Error("Update后进度数据不应该为nil")
			return
		}
		
		if file, exists := data.Data["file"]; !exists || file != "test.txt" {
			t.Error("Update后进度数据应该包含file字段")
		}
		
		if size, exists := data.Data["size"]; !exists || size != 1024 {
			t.Error("Update后进度数据应该包含size字段")
		}
	})

	// 测试Update方法在未启用进度时的行为
	t.Run("测试Update方法在未启用进度时的行为", func(t *testing.T) {
		progressHelper := NewProgressHelper("test-key-4")
		// 不调用Start，直接Update
		
		value := 50.0
		progressHelper.Update(&value, nil, nil)
		
		data := progressHelper.Get()
		if data != nil && data.Enable {
			t.Error("未启用进度时，Update不应该生�?)
		}
	})
}
