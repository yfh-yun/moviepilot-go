// Package main 配置使用示例
package main

import (
    "context"
    "log"
    
    "github.com/yfh-yun/moviepilot-go/config"
)

func main() {
    // 初始化配置管理器
    manager, err := config.Init()
    if err != nil {
        log.Fatal(err)
    }
    
    // 加载配置
    ctx := context.Background()
    appConfig, err := manager.Load(ctx, "config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用配置
    if app, ok := appConfig["app"].(map[string]interface{}); ok {
        log.Printf("应用名称: %v", app["name"])
        log.Printf("应用版本: %v", app["version"])
        log.Printf("运行环境: %v", app["env"])
    }
    
    if server, ok := appConfig["server"].(map[string]interface{}); ok {
        log.Printf("服务端口: %v", server["port"])
    }
}
