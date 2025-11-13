package helper

import (
	"testing"
)

func TestTorrentHelper_DownloadTorrent(t *testing.T) {
	// 创建TorrentHelper实例
	th := NewTorrentHelper()
	
	// 测试磁力链接
	result := th.DownloadTorrent("magnet:?xt=urn:btih:1234567890", "", "", "", false)
	if !result.IsMagnet {
		t.Error("Expected magnet link to be detected")
	}
	
	if result.ErrorMsg != "磁力链接" {
		t.Error("Expected error message to be '磁力链接'")
	}
}

func TestTorrentHelper_GetTorrentEpisodes(t *testing.T) {
	th := NewTorrentHelper()
	
	// 测试文件列表
	files := []string{
		"test.s01e01.mp4",
		"test.s01e02.mp4",
		"test.s01e03.mp4",
	}
	
	episodes := th.GetTorrentEpisodes(files)
	if len(episodes) != 3 {
		t.Errorf("Expected 3 episodes, got %d", len(episodes))
	}
}

func TestTorrentHelper_IsInvalid(t *testing.T) {
	th := NewTorrentHelper()
	
	// 初始状态下URL应该是有效的
	if th.IsInvalid("http://example.com/test.torrent") {
		t.Error("Expected URL to be valid initially")
	}
	
	// 添加无效URL后应该返回true
	th.AddInvalid("http://example.com/test.torrent")
	if !th.IsInvalid("http://example.com/test.torrent") {
		t.Error("Expected URL to be invalid after adding to invalid list")
	}
}
