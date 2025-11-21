# API 文档

## 🌐 API 概览

MoviePilot Go 提供 RESTful API 和 gRPC 接口，支持媒体资源管理的完整功能集。

### 基础信息
- **Base URL**: `http://localhost:3001/api/v1`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

### 通用响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid-string"
}
```

### 错误响应格式
```json
{
  "code": 400,
  "message": "error message",
  "error": "detailed error info",
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid-string"
}
```

## 🔐 认证授权

### 获取访问令牌
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "login success",
  "data": {
    "access_token": "jwt-token-string",
    "refresh_token": "refresh-token-string",
    "expires_in": 3600,
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin"
    }
  }
}
```

### 使用访问令牌
```http
Authorization: Bearer <access_token>
```

## 👤 用户管理 API

### 获取用户信息
```http
GET /api/v1/users/profile
Authorization: Bearer <token>
```

### 更新用户信息
```http
PUT /api/v1/users/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "nickname": "New Nickname",
  "email": "user@example.com"
}
```

### 用户列表（管理员）
```http
GET /api/v1/users?page=1&size=20&search=keyword
Authorization: Bearer <admin_token>
```

## 🎬 媒体管理 API

### 获取媒体列表
```http
GET /api/v1/media?type=movie&page=1&size=20&sort=created_at&order=desc
Authorization: Bearer <token>
```

**查询参数**:
- `type`: 媒体类型 (movie, tv, documentary)
- `page`: 页码 (默认: 1)
- `size`: 每页数量 (默认: 20, 最大: 100)
- `sort`: 排序字段 (created_at, title, rating)
- `order`: 排序方向 (asc, desc)
- `search`: 搜索关键词
- `genre`: 类型筛选
- `year`: 年份筛选

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "title": "Movie Title",
        "type": "movie",
        "year": 2024,
        "genre": ["Action", "Adventure"],
        "rating": 8.5,
        "poster_url": "https://example.com/poster.jpg",
        "overview": "Movie description...",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "size": 20,
      "total": 100,
      "pages": 5
    }
  }
}
```

### 添加媒体
```http
POST /api/v1/media
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "New Movie",
  "type": "movie",
  "year": 2024,
  "genre": ["Action"],
  "overview": "Movie description",
  "poster_url": "https://example.com/poster.jpg"
}
```

### 更新媒体信息
```http
PUT /api/v1/media/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Title",
  "overview": "Updated description"
}
```

### 删除媒体
```http
DELETE /api/v1/media/{id}
Authorization: Bearer <token>
```

## 📥 传输管理 API

### 获取传输历史
```http
GET /api/v1/transfers?page=1&size=20&status=completed
Authorization: Bearer <token>
```

**查询参数**:
- `status`: 传输状态 (pending, running, completed, failed)
- `source_type`: 源类型 (local, remote, torrent)
- `start_date`: 开始日期 (YYYY-MM-DD)
- `end_date`: 结束日期 (YYYY-MM-DD)

### 创建传输任务
```http
POST /api/v1/transfers
Authorization: Bearer <token>
Content-Type: application/json

{
  "source_path": "/path/to/source",
  "destination_path": "/path/to/destination",
  "source_type": "torrent",
  "priority": "normal",
  "auto_start": true
}
```

### 获取传输状态
```http
GET /api/v1/transfers/{id}
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "status": "running",
    "progress": 75.5,
    "speed": "10.5 MB/s",
    "source_path": "/path/to/source",
    "destination_path": "/path/to/destination",
    "created_at": "2024-01-01T00:00:00Z",
    "started_at": "2024-01-01T00:01:00Z",
    "estimated_completion": "2024-01-01T00:05:00Z"
  }
}
```

### 取消传输
```http
DELETE /api/v1/transfers/{id}
Authorization: Bearer <token>
```

## 🔍 搜索 API

### 全局搜索
```http
GET /api/v1/search?q=keyword&type=all&page=1&size=20
Authorization: Bearer <token>
```

**查询参数**:
- `q`: 搜索关键词
- `type`: 搜索类型 (all, media, transfers, users)
- `page`: 页码
- `size`: 每页数量

### 媒体搜索
```http
GET /api/v1/search/media?q=movie title&genre=action&year=2024
Authorization: Bearer <token>
```

## 🔌 插件 API

### 获取插件列表
```http
GET /api/v1/plugins
Authorization: Bearer <token>
```

### 启用插件
```http
POST /api/v1/plugins/{id}/enable
Authorization: Bearer <token>
```

### 禁用插件
```http
POST /api/v1/plugins/{id}/disable
Authorization: Bearer <token>
```

### 配置插件
```http
PUT /api/v1/plugins/{id}/config
Authorization: Bearer <token>
Content-Type: application/json

{
  "config": {
    "api_key": "your-api-key",
    "timeout": 30,
    "retry_count": 3
  }
}
```

## 📊 统计 API

### 获取系统统计
```http
GET /api/v1/stats/system
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "media_count": 1500,
    "transfer_count": 250,
    "active_transfers": 3,
    "storage_used": "2.5 TB",
    "storage_total": "5 TB",
    "cpu_usage": 25.5,
    "memory_usage": 60.2,
    "uptime": "5 days, 12 hours"
  }
}
```

### 获取媒体统计
```http
GET /api/v1/stats/media?period=30d
Authorization: Bearer <token>
```

## 🔔 通知 API

### 获取通知列表
```http
GET /api/v1/notifications?page=1&size=20&read=false
Authorization: Bearer <token>
```

### 标记通知已读
```http
PUT /api/v1/notifications/{id}/read
Authorization: Bearer <token>
```

### 删除通知
```http
DELETE /api/v1/notifications/{id}
Authorization: Bearer <token>
```

## ⚙️ 配置 API

### 获取系统配置
```http
GET /api/v1/config
Authorization: Bearer <admin_token>
```

### 更新系统配置
```http
PUT /api/v1/config
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "site_name": "MoviePilot",
  "max_concurrent_transfers": 5,
  "default_quality": "1080p",
  "auto_subtitle": true
}
```

## 🏥 健康检查

### 系统健康状态
```http
GET /api/v1/health
```

**响应**:
```json
{
  "code": 200,
  "message": "healthy",
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "5 days, 12 hours",
    "checks": {
      "database": "healthy",
      "redis": "healthy",
      "storage": "healthy",
      "plugins": "healthy"
    }
  }
}
```

## 📝 请求限制

### 速率限制
- **默认限制**: 100 请求/分钟
- **认证用户**: 1000 请求/分钟
- **管理员**: 5000 请求/分钟

### 请求大小限制
- **最大请求体**: 10MB
- **文件上传**: 100MB

## 🔄 WebSocket API

### 实时事件推送
```javascript
const ws = new WebSocket('ws://localhost:3001/api/v1/ws/events');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Event:', data);
};
```

**事件类型**:
- `transfer.progress`: 传输进度更新
- `transfer.completed`: 传输完成
- `media.added`: 新媒体添加
- `system.alert`: 系统告警

## 🧪 API 测试

### Postman Collection
提供完整的 Postman 测试集合，包含所有 API 端点的示例请求。

### OpenAPI/Swagger
访问 `http://localhost:3001/swagger/index.html` 查看交互式 API 文档。

## 📚 SDK 和客户端库

### Go SDK
```go
import "github.com/yfh-yun/moviepilot-go/sdk/go"

client := moviepilot.NewClient("http://localhost:3001", "your-token")
media, err := client.Media.Get(1)
```

### Python SDK
```python
from moviepilot_sdk import MoviePilotClient

client = MoviePilotClient("http://localhost:3001", "your-token")
media = client.media.get(1)
```

---

**注意**: 所有 API 都需要有效的认证令牌，除了健康检查和公开的 API 端点。