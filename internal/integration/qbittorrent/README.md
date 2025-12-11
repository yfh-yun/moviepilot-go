# qBittorrent 客户端

> qBittorrent Web API 客户端实现

---

## 功能特性

- ✅ 完整的 qBittorrent Web API 支持
- ✅ 自动登录和会话管理
- ✅ 种子添加（URL/磁力链接/文件）
- ✅ 种子列表和详情查询
- ✅ 种子控制（暂停/恢复/删除）
- ✅ 分类和标签管理
- ✅ 文件列表查询
- ✅ Tracker 信息查询
- ✅ 连接测试

---

## 快速开始

### 1. 创建客户端

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "moviepilot-go/internal/integration/qbittorrent"
    "moviepilot-go/internal/integration/downloader"
)

func main() {
    // 创建客户端配置
    config := qbittorrent.Config{
        BaseURL:  "http://localhost:8080",
        Username: "admin",
        Password: "adminpass",
        Timeout:  30 * time.Second,
    }
    
    // 创建客户端
    client, err := qbittorrent.NewClient(config)
    if err != nil {
        panic(err)
    }
    
    ctx := context.Background()
    
    // 测试连接
    if err := client.TestConnection(ctx); err != nil {
        panic(err)
    }
    
    fmt.Println("连接成功！")
}
```

### 2. 添加种子

```go
// 添加磁力链接
req := &downloader.AddTorrentRequest{
    URL:      "magnet:?xt=urn:btih:...",
    SavePath: "/downloads/movies",
    Category: "movies",
    Tags:     []string{"auto", "1080p"},
    Paused:   false,
}

torrent, err := client.AddTorrent(ctx, req)
if err != nil {
    panic(err)
}

fmt.Printf("种子添加成功: %s\n", torrent.Name)
```

### 3. 列出种子

```go
// 列出所有种子
torrents, err := client.ListTorrents(ctx, nil)
if err != nil {
    panic(err)
}

for _, torrent := range torrents {
    fmt.Printf("名称: %s\n", torrent.Name)
    fmt.Printf("状态: %s\n", torrent.State)
    fmt.Printf("进度: %.2f%%\n", torrent.Progress*100)
    fmt.Printf("下载速度: %d KB/s\n", torrent.DownloadSpeed/1024)
    fmt.Println("---")
}

// 按分类过滤
filter := &downloader.TorrentFilter{
    Category: "movies",
}
movieTorrents, err := client.ListTorrents(ctx, filter)
```

### 4. 获取种子详情

```go
hash := "abc123..."
info, err := client.GetTorrentInfo(ctx, hash)
if err != nil {
    panic(err)
}

fmt.Printf("种子: %s\n", info.Name)
fmt.Printf("总大小: %d MB\n", info.Size/1024/1024)
fmt.Printf("做种者: %d\n", info.Seeders)
fmt.Printf("下载者: %d\n", info.Leechers)
fmt.Printf("文件数量: %d\n", len(info.Files))

// 列出文件
for _, file := range info.Files {
    fmt.Printf("  - %s (%.2f%%)\n", file.Name, file.Progress*100)
}
```

### 5. 控制种子

```go
hash := "abc123..."

// 暂停种子
err := client.PauseTorrent(ctx, hash)

// 恢复种子
err = client.ResumeTorrent(ctx, hash)

// 删除种子（保留文件）
err = client.RemoveTorrent(ctx, hash, false)

// 删除种子（同时删除文件）
err = client.RemoveTorrent(ctx, hash, true)
```

### 6. 设置分类和标签

```go
hash := "abc123..."

// 设置分类
err := client.SetTorrentCategory(ctx, hash, "movies")

// 设置标签
err = client.SetTorrentTags(ctx, hash, []string{"1080p", "bluray"})
```

---

## API 参考

### Client 方法

#### AddTorrent
添加种子到下载器

```go
func (c *Client) AddTorrent(ctx context.Context, req *downloader.AddTorrentRequest) (*downloader.Torrent, error)
```

**参数**：
- `URL`: 种子URL或磁力链接
- `TorrentData`: 种子文件数据（二进制）
- `SavePath`: 保存路径
- `Category`: 分类
- `Tags`: 标签列表
- `Paused`: 是否暂停
- `SkipChecking`: 跳过校验
- `SequentialDownload`: 顺序下载
- `FirstLastPiecePrio`: 优先下载首尾块

#### ListTorrents
列出所有种子

```go
func (c *Client) ListTorrents(ctx context.Context, filter *downloader.TorrentFilter) ([]*downloader.Torrent, error)
```

**过滤器**：
- `Category`: 按分类过滤
- `Tag`: 按标签过滤
- `State`: 按状态过滤
- `Hashes`: 指定hash列表

#### GetTorrentInfo
获取种子详细信息

```go
func (c *Client) GetTorrentInfo(ctx context.Context, hash string) (*downloader.TorrentInfo, error)
```

#### PauseTorrent
暂停种子

```go
func (c *Client) PauseTorrent(ctx context.Context, hash string) error
```

#### ResumeTorrent
恢复种子

```go
func (c *Client) ResumeTorrent(ctx context.Context, hash string) error
```

#### RemoveTorrent
删除种子

```go
func (c *Client) RemoveTorrent(ctx context.Context, hash string, deleteFiles bool) error
```

#### SetTorrentCategory
设置种子分类

```go
func (c *Client) SetTorrentCategory(ctx context.Context, hash string, category string) error
```

#### SetTorrentTags
设置种子标签

```go
func (c *Client) SetTorrentTags(ctx context.Context, hash string, tags []string) error
```

#### GetVersion
获取qBittorrent版本

```go
func (c *Client) GetVersion(ctx context.Context) (string, error)
```

#### TestConnection
测试连接

```go
func (c *Client) TestConnection(ctx context.Context) error
```

---

## 种子状态

| 状态 | 说明 |
|------|------|
| `StateError` | 错误 |
| `StateMissingFiles` | 文件丢失 |
| `StateUploading` | 上传中（做种） |
| `StatePausedUP` | 暂停上传 |
| `StateDownloading` | 下载中 |
| `StatePausedDL` | 暂停下载 |
| `StateQueuedDL` | 排队下载 |
| `StateStalledDL` | 停滞下载 |
| `StateCheckingDL` | 检查下载 |
| `StateAllocating` | 分配空间 |
| `StateMoving` | 移动中 |

### 状态判断方法

```go
state := torrent.State

// 是否正在下载
if state.IsDownloading() {
    fmt.Println("正在下载")
}

// 是否已完成
if state.IsCompleted() {
    fmt.Println("下载完成")
}

// 是否已暂停
if state.IsPaused() {
    fmt.Println("已暂停")
}

// 是否错误
if state.IsError() {
    fmt.Println("发生错误")
}
```

---

## 测试

### 运行单元测试

```bash
go test ./internal/integration/qbittorrent -v
```

### 运行集成测试

需要先启动 qBittorrent 服务：

```bash
# 使用 Docker 启动 qBittorrent
docker run -d \
  --name qbittorrent \
  -p 8080:8080 \
  -e WEBUI_PORT=8080 \
  -e PUID=1000 \
  -e PGID=1000 \
  linuxserver/qbittorrent

# 运行集成测试
go test ./internal/integration/qbittorrent -v -run TestLogin
go test ./internal/integration/qbittorrent -v -run TestAddTorrent
go test ./internal/integration/qbittorrent -v -run TestListTorrents
```

---

## 错误处理

```go
torrent, err := client.AddTorrent(ctx, req)
if err != nil {
    // 检查错误类型
    if strings.Contains(err.Error(), "登录失败") {
        // 处理登录错误
    } else if strings.Contains(err.Error(), "添加种子失败") {
        // 处理添加失败
    }
    return err
}
```

---

## 最佳实践

### 1. 连接复用

```go
// 创建一个全局客户端实例
var qbClient *qbittorrent.Client

func init() {
    config := qbittorrent.Config{
        BaseURL:  os.Getenv("QB_URL"),
        Username: os.Getenv("QB_USER"),
        Password: os.Getenv("QB_PASS"),
    }
    
    var err error
    qbClient, err = qbittorrent.NewClient(config)
    if err != nil {
        panic(err)
    }
}
```

### 2. 错误重试

```go
func addTorrentWithRetry(ctx context.Context, req *downloader.AddTorrentRequest, maxRetries int) (*downloader.Torrent, error) {
    var torrent *downloader.Torrent
    var err error
    
    for i := 0; i < maxRetries; i++ {
        torrent, err = qbClient.AddTorrent(ctx, req)
        if err == nil {
            return torrent, nil
        }
        
        // 等待后重试
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    return nil, fmt.Errorf("添加种子失败，已重试%d次: %w", maxRetries, err)
}
```

### 3. 批量操作

```go
// 批量暂停种子
func pauseTorrents(ctx context.Context, hashes []string) error {
    for _, hash := range hashes {
        if err := qbClient.PauseTorrent(ctx, hash); err != nil {
            logger.Error("暂停种子失败", zap.String("hash", hash), zap.Error(err))
            continue
        }
    }
    return nil
}
```

---

## 相关文档

- [qBittorrent Web API 文档](https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1))
- [下载器统一接口](../downloader/interface.go)
- [Transmission 客户端](../transmission/README.md)

---

**最后更新**: 2025-12-02  
**维护者**: MoviePilot Go 开发团队
