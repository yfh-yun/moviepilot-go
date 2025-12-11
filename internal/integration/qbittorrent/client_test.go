package qbittorrent

import (
	"context"
	"testing"
	"time"

	"moviepilot-go/internal/integration/downloader"
)

// TestNewClient 测试创建客户端
func TestNewClient(t *testing.T) {
	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
		Timeout:  30 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	if client == nil {
		t.Fatal("客户端为空")
	}

	if client.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL 错误: got %s, want http://localhost:8080", client.baseURL)
	}
}

// TestLogin 测试登录（需要真实的qBittorrent服务）
func TestLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()
	err = client.Login(ctx)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
}

// TestAddTorrent 测试添加种子
func TestAddTorrent(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	// 测试添加磁力链接
	req := &downloader.AddTorrentRequest{
		URL:      "magnet:?xt=urn:btih:test",
		SavePath: "/downloads",
		Category: "test",
		Tags:     []string{"test", "auto"},
		Paused:   true,
	}

	torrent, err := client.AddTorrent(ctx, req)
	if err != nil {
		t.Fatalf("添加种子失败: %v", err)
	}

	if torrent == nil {
		t.Fatal("返回的种子信息为空")
	}

	t.Logf("种子添加成功: %s", torrent.Name)
}

// TestListTorrents 测试列出种子
func TestListTorrents(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	torrents, err := client.ListTorrents(ctx, nil)
	if err != nil {
		t.Fatalf("获取种子列表失败: %v", err)
	}

	t.Logf("找到 %d 个种子", len(torrents))

	for _, torrent := range torrents {
		t.Logf("种子: %s, 状态: %s, 进度: %.2f%%",
			torrent.Name,
			torrent.State,
			torrent.Progress*100)
	}
}

// TestGetTorrentInfo 测试获取种子详情
func TestGetTorrentInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	// 先获取种子列表
	torrents, err := client.ListTorrents(ctx, nil)
	if err != nil {
		t.Fatalf("获取种子列表失败: %v", err)
	}

	if len(torrents) == 0 {
		t.Skip("没有种子可测试")
	}

	// 获取第一个种子的详情
	hash := torrents[0].Hash
	info, err := client.GetTorrentInfo(ctx, hash)
	if err != nil {
		t.Fatalf("获取种子详情失败: %v", err)
	}

	t.Logf("种子详情: %s", info.Name)
	t.Logf("文件数量: %d", len(info.Files))
	t.Logf("做种者: %d, 下载者: %d", info.Seeders, info.Leechers)
}

// TestControlTorrent 测试控制种子
func TestControlTorrent(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	// 先获取种子列表
	torrents, err := client.ListTorrents(ctx, nil)
	if err != nil {
		t.Fatalf("获取种子列表失败: %v", err)
	}

	if len(torrents) == 0 {
		t.Skip("没有种子可测试")
	}

	hash := torrents[0].Hash

	// 测试暂停
	err = client.PauseTorrent(ctx, hash)
	if err != nil {
		t.Errorf("暂停种子失败: %v", err)
	}

	// 等待状态更新
	time.Sleep(time.Second)

	// 测试恢复
	err = client.ResumeTorrent(ctx, hash)
	if err != nil {
		t.Errorf("恢复种子失败: %v", err)
	}
}

// TestSetTorrentCategory 测试设置分类
func TestSetTorrentCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	torrents, err := client.ListTorrents(ctx, nil)
	if err != nil {
		t.Fatalf("获取种子列表失败: %v", err)
	}

	if len(torrents) == 0 {
		t.Skip("没有种子可测试")
	}

	hash := torrents[0].Hash

	err = client.SetTorrentCategory(ctx, hash, "movies")
	if err != nil {
		t.Errorf("设置分类失败: %v", err)
	}
}

// TestGetVersion 测试获取版本
func TestGetVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	version, err := client.GetVersion(ctx)
	if err != nil {
		t.Fatalf("获取版本失败: %v", err)
	}

	t.Logf("qBittorrent 版本: %s", version)

	if version == "" {
		t.Error("版本号为空")
	}
}

// TestTestConnection 测试连接测试
func TestTestConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	config := Config{
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Password: "adminpass",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	err = client.TestConnection(ctx)
	if err != nil {
		t.Fatalf("连接测试失败: %v", err)
	}

	t.Log("连接测试成功")
}
