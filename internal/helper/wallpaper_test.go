package helper

import (
	"testing"
)

func TestWallpaperHelper_GetWallpaper(t *testing.T) {
	// 获取壁纸帮助类实�?	wallpaperHelper := GetWallpaperHelper()
	
	// 测试获取壁纸（由于依赖外部配置，这里只是基本测试结构�?	wallpaper := wallpaperHelper.GetWallpaper()
	// 由于没有配置，应该返回空字符�?	if wallpaper != "" {
		t.Logf("Got wallpaper: %s", wallpaper)
	} else {
		t.Log("No wallpaper configured, returned empty string as expected")
	}
}

func TestWallpaperHelper_GetWallpapers(t *testing.T) {
	// 获取壁纸帮助类实�?	wallpaperHelper := GetWallpaperHelper()
	
	// 测试获取壁纸列表
	wallpapers := wallpaperHelper.GetWallpapers(5)
	// 由于没有配置，应该返回空列表
	if len(wallpapers) == 0 {
		t.Log("No wallpapers configured, returned empty list as expected")
	} else {
		t.Logf("Got %d wallpapers", len(wallpapers))
	}
}

func TestWallpaperHelper_FindFilesWithSuffixes(t *testing.T) {
	// 获取壁纸帮助类实�?	wallpaperHelper := GetWallpaperHelper()
	
	// 测试字符�?	result := wallpaperHelper.findFilesWithSuffixes("test.jpg", []string{".jpg", ".png"})
	if len(result) != 1 || result[0] != "test.jpg" {
		t.Errorf("Expected [\"test.jpg\"], got %v", result)
	}
	
	// 测试列表
	listData := []interface{}{"image1.jpg", "document.pdf", "image2.png"}
	result = wallpaperHelper.findFilesWithSuffixes(listData, []string{".jpg", ".png"})
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}
	
	// 测试字典
	mapData := map[string]interface{}{
		"file1": "image1.jpg",
		"file2": "document.pdf",
		"file3": "image2.png",
	}
	result = wallpaperHelper.findFilesWithSuffixes(mapData, []string{".jpg", ".png"})
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}
}
