package command

import (
	"fmt"
	"moviepilot-go/internal/scheduler"
	"moviepilot-go/internal/core/event"
)

// ExampleCommandManager 使用示例
func ExampleCommandManager() {
	// 创建调度器和事件管理器实�?	s := scheduler.NewScheduler()
	em := event.NewManager()
	
	// 创建命令管理�?	cm := NewCommandManager(s, em)
	
	// 注册一个自定义命令
	cm.Register("/hello", func(args ...interface{}) error {
		fmt.Println("Hello, World!")
		if len(args) > 0 {
			fmt.Printf("参数: %v\n", args)
		}
		return nil
	}, nil, "示例命令", "测试", true)
	
	// 获取命令列表
	commands := cm.GetCommands()
	fmt.Printf("总共�?%d 个命令\n", len(commands))
	
	// 执行命令
	cm.Execute("/hello", "test argument", "channel1", "source1", "user1")
}
