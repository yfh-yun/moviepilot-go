# Transmission 客户端

> Transmission RPC API 客户端实现

---

## 功能特性

- ✅ 完整的 Transmission RPC API 支持
- ✅ 自动 Session ID 管理
- ✅ 种子添加（URL/磁力链接/文件）
- ✅ 种子列表和详情查询
- ✅ 种子控制（启动/停止/删除）
- ✅ 标签管理（模拟分类）
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
    
    "moviepilot-go/internal/integration/transmission"
    "moviepilot-go/internal/integration/downloader"
)

func main() {
    // 创建客户端配置
    config := transmission.Config{
        BaseURL:  "http://localhost:9091",
        Username: "admin",
        Password: "password",
        Timeout:  30 * time.Second,
    }
    
    // 创建客户端
    client, err := transmission.NewClient(config)
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
    Category: "movies",  // 通过标签实现
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
    fmt.Printf("ID: %s\n", torrent.Hash)  // Transmission 使用 ID
    fmt.Printf("名称: %s\n", torrent.Name)
    fmt.Printf("状态: %s\n", torrent.State)
    fmt.Printf("进度: %.2f%%\n", torrent.Progress*100)
    fmt.Println("---")
}

// 按标签过滤
filter := &downloader.TorrentFilter{
    Tag: "movies",
}
movieTorrents, err := client.ListTorrents(ctx, filter)
```

### 4. 获取种子详情

```go
id := "123"  // Transmission 使用数字 ID
info, err := client.GetTorrentInfo(ctx, id)
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
id := "123"

// 暂停种子
err := client.PauseTorrent(ctx, id)

// 恢复种子
err = client.ResumeTorrent(ctx, id)

// 删除种子（保留文件）
err = client.RemoveTorrent(ctx, id, false)

// 删除种子（同时删除文件）
err = client.RemoveTorrent(ctx, id, true)
```

### 6. 设置标签

```go
id := "123"

// 设置分类（通过标签实现）
err := client.SetTorrentCategory(ctx, id, "movies")

// 设置多个标签
err = client.SetTorrentTags(ctx, id, []string{"1080p", "bluray", "movies"})
```

---

## API 参考

### Client 方法

#### RPC
执行原始 RPC 调用

```go
func (c *Client) RPC(ctx context.Context, method string, arguments interface{}) (json.RawMessage, error)
```

#### AddTorrent
添加种子到下载器

```go
func (c *Client) AddTorrent(ctx context.Context, req *downloader.AddTorrentRequest) (*downloader.Torrent, error)
```

#### ListTorrents
列出所有种子

```go
func (c *Client) ListTorrents(ctx context.Context, filter *downloader.TorrentFilter) ([]*downloader.Torrent, error)
```

#### GetTorrentInfo
获取种子详细信息

```go
func (c *Client) GetTorrentInfo(ctx context.Context, hash string) (*downloader.TorrentInfo, error)
```

**注意**: `hash` 参数实际上是 Transmission 的种子 ID（数字）

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
设置种子分类（通过标签实现）

```go
func (c *Client) SetTorrentCategory(ctx context.Context, hash string, category string) error
```

#### SetTorrentTags
设置种子标签

```go
func (c *Client) SetTorrentTags(ctx context.Context, hash string, tags []string) error
```

#### GetVersion
获取 Transmission 版本

```go
func (c *Client) GetVersion(ctx context.Context) (string, error)
```

#### TestConnection
测试连接

```go
func (c *Client) TestConnection(ctx context.Context) error
```

---

## Transmission 状态映射

| Transmission 状态 | 状态码 | 映射到统一状态 |
|------------------|--------|---------------|
| 已停止 | 0 | `StatePausedDL` / `StatePausedUP` |
| 检查等待 | 1 | `StateCheckingResumeData` |
| 检查中 | 2 | `StateCheckingDL` / `StateCheckingUP` |
| 下载等待 | 3 | `StateQueuedDL` |
| 下载中 | 4 | `StateDownloading` / `StateStalledDL` |
| 做种等待 | 5 | `StateQueuedUP` |
| 做种中 | 6 | `StateUploading` / `StateStalledUP` |

---

## 与 qBittorrent 的差异

### 1. 种子标识
- **qBittorrent**: 使用 hash（字符串）
- **Transmission**: 使用 ID（数字）

### 2. 分类管理
- **qBittorrent**: 原生支持分类
- **Transmission**: 通过标签（labels）模拟分类

### 3. Session 管理
- **qBittorrent**: Cookie 认证
- **Transmission**: Session ID 认证（自动处理 409 错误）

### 4. RPC 协议
- **qBittorrent**: RESTful API
- **Transmission**: JSON-RPC 2.0

---

## 测试

### 运行单元测试

```bash
go test ./internal/integration/transmission -v
```

### 运行集成测试

需要先启动 Transmission 服务：

```bash
# 使用 Docker 启动 Transmission
docker run -d \
  --name transmission \
  -p 9091:9091 \
  -e USER=admin \
  -e PASS=password \
  -e PUID=1000 \
  -e PGID=1000 \
  linuxserver/transmission

# 运行集成测试
go test ./internal/integration/transmission -v -run TestRPC
go test ./internal/integration/transmission -v -run TestAddTorrent
go test ./internal/integration/transmission -v -run TestListTorrents
```

---

## 错误处理

### Session ID 自动更新

Transmission 需要 Session ID 进行认证。客户端会自动处理 409 错误并更新 Session ID：

```go
// 自动处理 409 错误
if resp.StatusCode == http.StatusConflict {
    c.sessionID = resp.Header.Get("X-Transmission-Session-Id")
    // 自动重试请求
    return c.RPC(ctx, method, arguments)
}
```

### RPC 错误处理

```go
torrent, err := client.AddTorrent(ctx, req)
if err != nil {
    if strings.Contains(err.Error(), "RPC 调用失败") {
        // 处理 RPC 错误
    } else if strings.Contains(err.Error(), "种子已存在") {
        // 处理重复添加
    }
    return err
}
```

---

## 最佳实践

### 1. ID 转换

Transmission 使用数字 ID，需要注意类型转换：

```go
// 从种子列表获取 ID
torrents, _ := client.ListTorrents(ctx, nil)
for _, t := range torrents {
    id := t.Hash  // 实际上是字符串形式的 ID
    info, _ := client.GetTorrentInfo(ctx, id)
}
```

### 2. 标签管理

使用标签模拟分类功能：

```go
// 设置分类（实际上是添加标签）
client.SetTorrentCategory(ctx, id, "movies")

// 添加更多标签
client.SetTorrentTags(ctx, id, []string{"movies", "1080p", "bluray"})

// 查询时按标签过滤
filter := &downloader.TorrentFilter{
    Tag: "movies",
}
torrents, _ := client.ListTorrents(ctx, filter)
```

### 3. 批量操作

```go
// 批量暂停种子
func pauseTorrents(ctx context.Context, ids []string) error {
    for _, id := range ids {
        if err := client.PauseTorrent(ctx, id); err != nil {
            logger.Error("暂停种子失败", zap.String("id", id), zap.Error(err))
            continue
        }
    }
    return nil
}
```

---

## 高级用法

### 直接调用 RPC

```go
// 获取会话统计
result, err := client.RPC(ctx, "session-stats", nil)
if err != nil {
    panic(err)
}

var stats struct {
    ActiveTorrentCount int64 `json:"activeTorrentCount"`
    DownloadSpeed      int64 `json:"downloadSpeed"`
    UploadSpeed        int64 `json:"uploadSpeed"`
}
json.Unmarshal(result, &stats)

fmt.Printf("活跃种子: %d\n", stats.ActiveTorrentCount)
fmt.Printf("下载速度: %d KB/s\n", stats.DownloadSpeed/1024)
fmt.Printf("上传速度: %d KB/s\n", stats.UploadSpeed/1024)
```

### 设置全局限速

```go
args := map[string]interface{}{
    "speed-limit-down": 1024,  // KB/s
    "speed-limit-up":   512,   // KB/s
    "speed-limit-down-enabled": true,
    "speed-limit-up-enabled":   true,
}

_, err := client.RPC(ctx, "session-set", args)
```

---

## 相关文档

- [Transmission RPC 规范](https://github.com/transmission/transmission/blob/main/docs/rpc-spec.md)
- [下载器统一接口](../downloader/interface.go)
- [qBittorrent 客户端](../qbittorrent/README.md)

---

**最后更新**: 2025-12-02  
**维护者**: MoviePilot Go 开发团队
