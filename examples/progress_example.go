package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
	"moviepilot-go/pkg/models"
	"time"
)

func main() {
	// 创建ProgressHelper实例
	progressHelper := helper.NewProgressHelper(models.ProgressKeySearch)
	
	// 开始进�?	progressHelper.Start()
	fmt.Println("开始进�?..")
	
	// 获取初始进度
	data := progressHelper.Get()
	if data != nil {
		fmt.Printf("初始进度: 启用=%t, �?%.1f, 文本=%s\n", data.Enable, data.Value, data.Text)
	}
	
	// 模拟进度更新
	for i := 0; i <= 10; i++ {
		value := float64(i * 10)
		text := fmt.Sprintf("处理�?.. %d%%", int(value))
		
		// 更新进度
		progressHelper.Update(&value, &text, map[string]interface{}{
			"step": i,
			"info": "正在处理文件",
		})
		
		// 获取当前进度
		data = progressHelper.Get()
		if data != nil {
			fmt.Printf("当前进度: 启用=%t, �?%.1f, 文本=%s\n", data.Enable, data.Value, data.Text)
			if data.Data != nil {
				fmt.Printf("  额外数据: %+v\n", data.Data)
			}
		}
		
		time.Sleep(500 * time.Millisecond)
	}
	
	// 结束进度
	progressHelper.End()
	fmt.Println("进度结束")
	
	// 获取最终进�?	data = progressHelper.Get()
	if data != nil {
		fmt.Printf("最终进�? 启用=%t, �?%.1f, 文本=%s\n", data.Enable, data.Value, data.Text)
	}
}
