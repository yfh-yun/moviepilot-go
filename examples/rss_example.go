package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
	"moviepilot-go/internal/logger"
)

func main() {
	fmt.Println("RSS Helper Example")
	
	// 创建RSS帮助类实�?	rssHelper := helper.NewRssHelper()
	
	if rssHelper == nil {
		logger.Error("Failed to create RssHelper")
		return
	}
	
	fmt.Println("RssHelper created successfully")
	
	// 显示配置信息
	fmt.Printf("Max RSS Size: %d bytes\n", rssHelper.MaxRssSize)
	fmt.Printf("Max RSS Items: %d\n", rssHelper.MaxRssItems)
	fmt.Printf("Supported sites count: %d\n", len(rssHelper.rssLinkConf))
	
	// 测试解析一个公开的RSS源（示例�?	fmt.Println("\n=== Testing RSS Parse ===")
	// 注意：这需要网络连接，且URL需要是有效的RSS�?	/*
	items, resultFlag, err := rssHelper.Parse("https://feeds.bbci.co.uk/news/rss.xml", false, 30, nil)
	if err != nil {
		fmt.Printf("Error parsing RSS: %v\n", err)
	} else if resultFlag != nil && !*resultFlag {
		fmt.Println("Failed to parse RSS")
	} else if items == nil {
		fmt.Println("RSS feed expired")
	} else {
		fmt.Printf("Successfully parsed %d RSS items\n", len(items))
		for i, item := range items {
			if i >= 5 { // 只显示前5�?				break
			}
			fmt.Printf("  %d. %s\n", i+1, item.Title)
		}
	}
	*/
	
	// 测试获取RSS链接
	fmt.Println("\n=== Testing Get RSS Link ===")
	// 注意：这需要有效的站点URL和相关认证信�?	/*
	link, errMsg := rssHelper.GetRssLink("https://example.com", "", "", false, 30)
	if errMsg != "" {
		fmt.Printf("Error getting RSS link: %s\n", errMsg)
	} else {
		fmt.Printf("RSS link: %s\n", link)
	}
	*/
	
	// 显示一些预定义的站点配�?	fmt.Println("\n=== Site Configurations ===")
	for site, config := range rssHelper.rssLinkConf {
		if site == "default" || site == "m-team.io" || site == "hdchina.org" {
			fmt.Printf("Site: %s\n", site)
			fmt.Printf("  URL: %s\n", config.URL)
			fmt.Printf("  XPath: %s\n", config.XPath)
			fmt.Printf("  Render: %v\n", config.Render)
			fmt.Printf("  Params count: %d\n", len(config.Params))
			fmt.Println()
		}
	}
	
	fmt.Println("Example completed")
}
