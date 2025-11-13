package main

import (
	"fmt"
	"moviepilot-go/internal/helper"
)

func main() {
	// 创建TorrentHelper实例
	th := helper.NewTorrentHelper()

	// 测试下载种子功能（使用磁力链接示例）
	fmt.Println("Testing DownloadTorrent with magnet link...")
	result := th.DownloadTorrent("magnet:?xt=urn:btih:1234567890", "", "", "", false)
	fmt.Printf("IsMagnet: %v, ErrorMsg: %s\n", result.IsMagnet, result.ErrorMsg)

	// 测试无效种子功能
	fmt.Println("\nTesting invalid torrent functionality...")
	url := "http://example.com/test.torrent"
	fmt.Printf("IsInvalid(%s): %v\n", url, th.IsInvalid(url))
	
	th.AddInvalid(url)
	fmt.Printf("IsInvalid(%s) after AddInvalid: %v\n", url, th.IsInvalid(url))

	// 测试获取种子集数功能
	fmt.Println("\nTesting GetTorrentEpisodes...")
	files := []string{
		"test.s01e01.mp4",
		"test.s01e02.mp4",
		"test.s01e03.mp4",
	}
	episodes := th.GetTorrentEpisodes(files)
	fmt.Printf("Episodes found: %v\n", episodes)
}
