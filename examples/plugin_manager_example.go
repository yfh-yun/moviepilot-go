package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/plugin"
)

func main() {
	// 创建插件配置
	config := &plugin.Config{
		Plugins: plugin.PluginConfig{
			Native: plugin.NativePluginConfig{
				Path:    "./plugins",
				Enabled: true,
			},
			Python: plugin.PythonPluginConfig{
				Host:    "localhost",
				Port:    5000,
				Timeout: 30,
				Enabled: false, // 禁用Python插件以简化示例
			},
			Web: plugin.WebPluginConfig{
				Enabled: false,
				Path:    "./web-plugins",
			},
		},
	}

	// 创建混合插件管理器
	manager, err := plugin.NewHybridPluginManager(config)
	if err != nil {
		log.Fatalf("Failed to create plugin manager: %v", err)
	}

	// 启动健康监控（在后台运行）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go manager.MonitorPluginHealth(ctx)

	// 示例1：加载原生插件
	fmt.Println("=== Loading Native Plugin ===")
	err = manager.LoadPlugin("./plugins/hello_plugin.so", plugin.PluginTypeNative)
	if err != nil {
		log.Printf("Failed to load native plugin: %v", err)
	} else {
		fmt.Println("Native plugin loaded successfully")
	}

	// 示例2：获取插件列表
	fmt.Println("\n=== Listing Plugins ===")
	plugins := manager.ListPlugins()
	for _, p := range plugins {
		fmt.Printf("Plugin: %s (%s) - %s\n", p.Name, p.Version, p.State)
	}

	// 示例3：初始化和启动插件
	if len(plugins) > 0 {
		pluginID := plugins[0].ID
		
		fmt.Printf("\n=== Initializing Plugin: %s ===\n", pluginID)
		config := map[string]interface{}{
			"greeting": "Hello from MoviePilot!",
			"enabled":   true,
		}
		
		err = manager.InitializePlugin(pluginID)
		if err != nil {
			log.Printf("Failed to initialize plugin: %v", err)
		} else {
			fmt.Println("Plugin initialized successfully")
		}

		fmt.Printf("\n=== Starting Plugin: %s ===\n", pluginID)
		err = manager.StartPlugin(pluginID)
		if err != nil {
			log.Printf("Failed to start plugin: %v", err)
		} else {
			fmt.Println("Plugin started successfully")
		}

		// 示例4：调用插件方法
		fmt.Printf("\n=== Calling Plugin Method: %s ===\n", pluginID)
		result, err := manager.CallPluginMethod(pluginID, "hello", "MoviePilot")
		if err != nil {
			log.Printf("Failed to call plugin method: %v", err)
		} else {
			fmt.Printf("Method result: %v\n", result)
		}

		// 示例5：发布事件
		fmt.Printf("\n=== Publishing Event ===\n")
		event := plugin.CreateEvent("test.event", "example", map[string]interface{}{
			"message": "Hello from event system!",
			"time":    time.Now(),
		})
		manager.PublishEvent(event)
		fmt.Println("Event published successfully")

		// 示例6：获取插件详细信息
		fmt.Printf("\n=== Getting Plugin Info: %s ===\n", pluginID)
		info, err := manager.GetPluginInfo(pluginID)
		if err != nil {
			log.Printf("Failed to get plugin info: %v", err)
		} else {
			infoJSON, _ := json.MarshalIndent(info, "", "  ")
			fmt.Printf("Plugin Info:\n%s\n", string(infoJSON))
		}

		// 示例7：停止插件
		fmt.Printf("\n=== Stopping Plugin: %s ===\n", pluginID)
		err = manager.StopPlugin(pluginID)
		if err != nil {
			log.Printf("Failed to stop plugin: %v", err)
		} else {
			fmt.Println("Plugin stopped successfully")
		}
	}

	// 等待一段时间以观察健康监控
	fmt.Println("\n=== Monitoring Plugin Health ===")
	fmt.Println("Waiting for 5 seconds to observe health monitoring...")
	time.Sleep(5 * time.Second)

	// 示例8：卸载插件
	if len(plugins) > 0 {
		pluginID := plugins[0].ID
		fmt.Printf("\n=== Unloading Plugin: %s ===\n", pluginID)
		err = manager.UnloadPlugin(pluginID)
		if err != nil {
			log.Printf("Failed to unload plugin: %v", err)
		} else {
			fmt.Println("Plugin unloaded successfully")
		}
	}

	fmt.Println("\n=== Plugin Manager Example Completed ===")
}