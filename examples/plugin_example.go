package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
)

func main() {
	// 创建PluginHelper实例
	pluginHelper := helper.NewPluginHelper()
	
	// 示例：获取插件列�?	repoURL := "https://github.com/jxxghp/MoviePilot-Plugins"
	plugins, err := pluginHelper.GetPlugins(repoURL, nil, true)
	if err != nil {
		fmt.Printf("获取插件列表失败: %v\n", err)
	} else {
		fmt.Printf("获取�?%d 个插件\n", len(plugins))
	}
	
	// 示例：安装插�?	pid := "example-plugin"
	success, message := pluginHelper.Install(pid, repoURL, nil, false)
	
	if success {
		fmt.Printf("插件 %s 安装成功\n", pid)
	} else {
		fmt.Printf("插件 %s 安装失败: %s\n", pid, message)
	}
}
