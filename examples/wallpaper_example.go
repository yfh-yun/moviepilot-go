package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
)

func main() {
	// 获取壁纸帮助类实�?	wallpaperHelper := helper.GetWallpaperHelper()

	// 示例1: 获取单张壁纸
	fmt.Println("=== 获取单张壁纸 ===")
	wallpaper := wallpaperHelper.GetWallpaper()
	if wallpaper != "" {
		fmt.Printf("获取到壁�? %s\n", wallpaper)
	} else {
		fmt.Println("未配置壁纸源或获取失�?)
	}

	// 示例2: 获取多张壁纸
	fmt.Println("\n=== 获取多张壁纸 ===")
	wallpapers := wallpaperHelper.GetWallpapers(5)
	if len(wallpapers) > 0 {
		fmt.Printf("获取�?%d 张壁�?\n", len(wallpapers))
		for i, wp := range wallpapers {
			fmt.Printf("  %d. %s\n", i+1, wp)
		}
	} else {
		fmt.Println("未配置壁纸源或获取失�?)
	}

	// 示例3: 分别测试各种壁纸�?	fmt.Println("\n=== 测试各种壁纸�?===")
	
	// Bing壁纸
	bingWallpaper := wallpaperHelper.GetBingWallpaper()
	fmt.Printf("Bing壁纸: %s\n", bingWallpaper)
	
	bingWallpapers := wallpaperHelper.GetBingWallpapers(3)
	fmt.Printf("获取�?%d 张Bing壁纸\n", len(bingWallpapers))
	
	// 自定义壁�?	customWallpaper := wallpaperHelper.GetCustomizeWallpaper()
	fmt.Printf("自定义壁�? %s\n", customWallpaper)
	
	customWallpapers := wallpaperHelper.GetCustomizeWallpapers()
	fmt.Printf("获取�?%d 张自定义壁纸\n", len(customWallpapers))
}
