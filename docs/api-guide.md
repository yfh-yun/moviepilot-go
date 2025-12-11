# MoviePilot API 使用指南

> **版本**: 2.8.1  
> **更新时间**: 2025-12-02

---

## 📋 目录

1. [快速开始](#快速开始)
2. [认证方式](#认证方式)
3. [API 概览](#api-概览)
4. [核心 API 详解](#核心-api-详解)
5. [错误处理](#错误处理)
6. [最佳实践](#最佳实践)

---

## 快速开始

### 访问 Swagger UI

启动 MoviePilot Go 服务后，访问：

```
http://localhost:3001/swagger/index.html
```

Swagger UI 提供：
- 📖 完整的 API 文档
- 🧪 在线测试工具
- 📝 请求/响应示例
- 🔐 认证配置

### 基础配置

**Base URL**: `http://localhost:3001`  
**API 前缀**: `/api/v1`

---

## 认证方式

MoviePilot API 使用 **Bearer Token** 认证。

### 1. 获取令牌

**请求**:
```http
POST /api/user/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your_password"
}
```

**响应**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

### 2. 使用令牌

在所有需要认证的请求中添加 Header：

```http
Authorization: Bearer <access_token>
```

### 3. 刷新令牌

当 `access_token` 过期时，使用 `refresh_token` 获取新令牌：

```http
POST /api/user/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## API 概览

### 用户管理 (User)

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/user/login` | 用户登录 | ❌ |
| POST | `/api/user/logout` | 用户登出 | ✅ |
| POST | `/api/user/refresh` | 刷新令牌 | ❌ |
| PUT | `/api/user/password` | 修改密码 | ✅ |
| GET | `/api/user/validate` | 验证令牌 | ✅ |
| GET | `/api/user/permissions` | 获取权限 | ✅ |
| GET | `/api/user/permission/check` | 检查权限 | ✅ |

### 订阅管理 (Subscribe)

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/subscribes` | 创建订阅 | ✅ |
| GET | `/api/subscribes/{id}` | 获取订阅详情 | ✅ |
| PUT | `/api/subscribes/{id}` | 更新订阅 | ✅ |
| DELETE | `/api/subscribes/{id}` | 删除订阅 | ✅ |
| GET | `/api/subscribes` | 列出订阅 | ✅ |
| POST | `/api/subscribes/{id}/refresh` | 刷新订阅 | ✅ |

### 下载管理 (Download)

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/downloads` | 添加下载任务 | ✅ |
| GET | `/api/downloads/{id}` | 获取下载详情 | ✅ |
| GET | `/api/downloads` | 列出下载任务 | ✅ |
| DELETE | `/api/downloads/{id}` | 删除下载任务 | ✅ |
| POST | `/api/downloads/{id}/pause` | 暂停下载 | ✅ |
| POST | `/api/downloads/{id}/resume` | 恢复下载 | ✅ |

### 媒体服务器 (Media Server)

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/mediaserver/info` | 获取服务器信息 | ✅ |
| GET | `/api/mediaserver/libraries` | 列出媒体库 | ✅ |
| GET | `/api/mediaserver/items/{id}` | 获取媒体条目 | ✅ |
| GET | `/api/mediaserver/search` | 搜索媒体 | ✅ |
| POST | `/api/mediaserver/sync` | 同步媒体库 | ✅ |

### 元数据 (Metadata)

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/metadata/movie/search` | 搜索电影 | ✅ |
| GET | `/api/metadata/movie/{id}` | 获取电影详情 | ✅ |
| GET | `/api/metadata/tv/search` | 搜索剧集 | ✅ |
| GET | `/api/metadata/tv/{id}` | 获取剧集详情 | ✅ |
| POST | `/api/metadata/aggregate/movie` | 聚合电影元数据 | ✅ |
| POST | `/api/metadata/aggregate/tv` | 聚合剧集元数据 | ✅ |

### 系统管理 (System)

| 方法 | 路径 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/system/info` | 获取系统信息 | ✅ |
| GET | `/api/system/health` | 健康检查 | ❌ |
| GET | `/api/system/stats` | 系统统计 | ✅ |
| POST | `/api/system/restart` | 重启系统 | ✅ |

---

## 核心 API 详解

### 1. 创建订阅

订阅功能是 MoviePilot 的核心功能之一，用于自动追踪和下载媒体内容。

**请求**:
```http
POST /api/subscribes
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "The Last of Us",
  "year": 2023,
  "type": "tv",
  "tmdb_id": 100088,
  "quality": "1080p",
  "season": 1,
  "total_episode": 9,
  "start_episode": 1,
  "best": true,
  "save_path": "/media/tv/The Last of Us"
}
```

**字段说明**:
- `title` (必填): 媒体标题
- `year` (可选): 年份
- `type` (必填): 类型 (`movie` 或 `tv`)
- `tmdb_id` (可选): TMDB ID
- `quality` (可选): 质量要求 (`720p`, `1080p`, `4k` 等)
- `season` (剧集必填): 季数
- `total_episode` (可选): 总集数
- `start_episode` (可选): 起始集数
- `best` (可选): 是否只下载最佳资源
- `save_path` (可选): 保存路径

**响应**:
```json
{
  "subscribe_id": 123,
  "title": "The Last of Us",
  "type": "tv",
  "state": "active",
  "created_at": "2025-12-02T07:00:00Z"
}
```

### 2. 搜索电影元数据

使用聚合服务搜索电影，自动从 TMDB/豆瓣等多个数据源获取信息。

**请求**:
```http
GET /api/metadata/movie/search?title=Inception&year=2010
Authorization: Bearer <token>
```

**响应**:
```json
{
  "results": [
    {
      "id": "27205",
      "provider": "tmdb",
      "title": "Inception",
      "original_title": "Inception",
      "year": 2010,
      "overview": "Cobb, a skilled thief...",
      "poster_url": "https://image.tmdb.org/t/p/w500/...",
      "backdrop_url": "https://image.tmdb.org/t/p/original/...",
      "tmdb_id": 27205,
      "imdb_id": "tt1375666"
    }
  ],
  "total": 1
}
```

### 3. 添加下载任务

**请求**:
```http
POST /api/downloads
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "magnet:?xt=urn:btih:...",
  "title": "The Last of Us S01E01",
  "save_path": "/media/tv/The Last of Us/Season 1",
  "category": "tv"
}
```

**响应**:
```json
{
  "download_id": "abc123",
  "title": "The Last of Us S01E01",
  "status": "downloading",
  "progress": 0,
  "speed": 0,
  "eta": 0
}
```

### 4. 媒体库同步

同步媒体服务器（Emby/Plex/Jellyfin）的媒体库。

**请求**:
```http
POST /api/mediaserver/sync
Authorization: Bearer <token>
Content-Type: application/json

{
  "library_id": "1",
  "full_sync": false
}
```

**响应**:
```json
{
  "task_id": "sync-123",
  "status": "running",
  "message": "同步已开始"
}
```

---

## 错误处理

### 标准错误响应

所有 API 错误都遵循统一格式：

```json
{
  "error": "error_code",
  "message": "详细错误信息",
  "details": {
    "field": "具体字段错误"
  }
}
```

### 常见错误码

| HTTP 状态码 | 错误码 | 描述 |
|------------|--------|------|
| 400 | `invalid_request` | 请求参数错误 |
| 401 | `unauthorized` | 未认证或令牌无效 |
| 403 | `forbidden` | 无权限访问 |
| 404 | `not_found` | 资源不存在 |
| 409 | `conflict` | 资源冲突 |
| 429 | `rate_limit_exceeded` | 请求频率超限 |
| 500 | `internal_error` | 服务器内部错误 |

### 错误处理示例

```javascript
fetch('http://localhost:3001/api/subscribes', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(subscribeData)
})
.then(response => {
  if (!response.ok) {
    return response.json().then(err => {
      throw new Error(err.message || '请求失败');
    });
  }
  return response.json();
})
.then(data => {
  console.log('订阅创建成功:', data);
})
.catch(error => {
  console.error('错误:', error.message);
});
```

---

## 最佳实践

### 1. 令牌管理

- ✅ 将 `access_token` 存储在内存中（如 JavaScript 变量）
- ✅ 将 `refresh_token` 存储在 HttpOnly Cookie 或安全存储中
- ✅ 在令牌过期前主动刷新
- ❌ 不要将令牌存储在 localStorage（XSS 风险）

### 2. 请求优化

- ✅ 使用分页参数减少数据量：`?page=1&limit=20`
- ✅ 使用字段过滤减少响应大小：`?fields=id,title,year`
- ✅ 合理使用缓存，避免重复请求
- ✅ 批量操作时使用批量 API

### 3. 错误重试

对于可重试的错误（如网络超时、5xx 错误），建议使用指数退避策略：

```javascript
async function fetchWithRetry(url, options, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(url, options);
      if (response.ok) {
        return response.json();
      }
      if (response.status >= 400 && response.status < 500) {
        // 客户端错误，不重试
        throw new Error(`Client error: ${response.status}`);
      }
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      // 指数退避: 2^i * 1000ms
      await new Promise(resolve => setTimeout(resolve, Math.pow(2, i) * 1000));
    }
  }
}
```

### 4. 限流处理

API 有速率限制（默认 100 req/min）。建议：

- ✅ 监听 `429` 状态码
- ✅ 读取 `Retry-After` 响应头
- ✅ 实现客户端限流队列

### 5. WebSocket 订阅

对于实时更新（如下载进度），建议使用 WebSocket：

```javascript
const ws = new WebSocket('ws://localhost:3001/api/ws');

ws.onopen = () => {
  // 订阅下载进度
  ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'downloads',
    id: 'abc123'
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('下载进度:', data.progress);
};
```

---

## 附录

### A. 完整请求示例（cURL）

```bash
# 登录
curl -X POST http://localhost:3001/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# 创建订阅
curl -X POST http://localhost:3001/api/subscribes \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Last of Us",
    "year": 2023,
    "type": "tv",
    "season": 1
  }'

# 列出订阅
curl -X GET "http://localhost:3001/api/subscribes?page=1&limit=20" \
  -H "Authorization: Bearer <token>"
```

### B. SDK 示例（JavaScript）

```javascript
class MoviePilotClient {
  constructor(baseURL, token) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async request(method, path, data = null) {
    const options = {
      method,
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      }
    };

    if (data) {
      options.body = JSON.stringify(data);
    }

    const response = await fetch(`${this.baseURL}${path}`, options);
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message);
    }

    return response.json();
  }

  // 订阅相关
  async createSubscribe(data) {
    return this.request('POST', '/api/subscribes', data);
  }

  async getSubscribe(id) {
    return this.request('GET', `/api/subscribes/${id}`);
  }

  async listSubscribes(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request('GET', `/api/subscribes?${query}`);
  }

  // 下载相关
  async addDownload(data) {
    return this.request('POST', '/api/downloads', data);
  }

  async getDownload(id) {
    return this.request('GET', `/api/downloads/${id}`);
  }
}

// 使用示例
const client = new MoviePilotClient('http://localhost:3001', 'your_token');

// 创建订阅
const subscribe = await client.createSubscribe({
  title: 'The Last of Us',
  year: 2023,
  type: 'tv',
  season: 1
});

console.log('订阅创建成功:', subscribe);
```

---

## 更新日志

### v2.8.1 (2025-12-02)

- ✅ 新增媒体服务器集成 API（Emby/Plex/Jellyfin）
- ✅ 新增元数据聚合 API（TMDB/TVDB/豆瓣）
- ✅ 优化订阅管理 API
- ✅ 完善 Swagger 文档
- ✅ 新增 API 使用指南

---

**需要帮助？**

- 📖 查看 [Swagger UI](http://localhost:3001/swagger/index.html)
- 🐛 提交 [Issue](https://github.com/moviepilot/moviepilot-go/issues)
- 💬 加入 [讨论](https://github.com/moviepilot/moviepilot-go/discussions)
