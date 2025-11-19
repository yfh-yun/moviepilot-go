# API 文档

## 概述

MoviePilot Go 提供RESTful API，支持媒体库管理的各种功能。

## 基础信息

- **Base URL**: `http://localhost:3001/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: JWT Token / API Key

## 认证

### JWT认证

```http
Authorization: Bearer <jwt_token>
```

### API Key认证

```http
X-API-Key: <api_key>
```

## API端点

### 用户管理

#### 用户登录
```http
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

#### 获取用户信息
```http
GET /users/profile
Authorization: Bearer <token>
```

### 媒体管理

#### 获取媒体列表
```http
GET /media?type=movie&page=1&size=20
Authorization: Bearer <token>
```

#### 搜索媒体
```http
GET /media/search?q=keyword&type=movie
Authorization: Bearer <token>
```

### 订阅管理

#### 创建订阅
```http
POST /subscribe
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "订阅名称",
  "type": "movie",
  "tmdb_id": 12345,
  "quality": "1080p"
}
```

#### 获取订阅列表
```http
GET /subscribe?page=1&size=20
Authorization: Bearer <token>
```

### 下载管理

#### 获取下载列表
```http
GET /downloads?page=1&size=20
Authorization: Bearer <token>
```

#### 创建下载任务
```http
POST /downloads
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "magnet:?xt=urn:btih:...",
  "save_path": "/downloads/movies"
}
```

### 站点管理

#### 获取站点列表
```http
GET /sites
Authorization: Bearer <token>
```

#### 添加站点
```http
POST /sites
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "站点名称",
  "url": "https://example.com",
  "type": "nexusphp"
}
```

## 错误响应

所有API错误响应都遵循统一格式：

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述",
    "details": {}
  },
  "timestamp": "2023-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

### 常见错误码

| 错误码 | HTTP状态码 | 描述 |
|--------|------------|------|
| INVALID_REQUEST | 400 | 请求参数无效 |
| UNAUTHORIZED | 401 | 未授权访问 |
| FORBIDDEN | 403 | 禁止访问 |
| NOT_FOUND | 404 | 资源不存在 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |

## 限流

API实施限流策略：
- 每个IP每分钟最多100个请求
- 超出限制返回429状态码

## 版本控制

API版本通过URL路径控制：
- v1: `/api/v1/`
- v2: `/api/v2/` (未来版本)

## Swagger文档

启动服务后访问: http://localhost:3001/swagger/index.html

## 示例代码

### JavaScript/Node.js

```javascript
const axios = require('axios');

const client = axios.create({
  baseURL: 'http://localhost:3001/api/v1',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  }
});

// 获取媒体列表
async function getMovies() {
  try {
    const response = await client.get('/media?type=movie');
    return response.data;
  } catch (error) {
    console.error('获取媒体列表失败:', error.response.data);
  }
}
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type APIClient struct {
    BaseURL    string
    Token      string
    HTTPClient *http.Client
}

func NewAPIClient(baseURL, token string) *APIClient {
    return &APIClient{
        BaseURL: baseURL,
        Token:   token,
        HTTPClient: &http.Client{},
    }
}

func (c *APIClient) GetMovies() ([]Movie, error) {
    req, _ := http.NewRequest("GET", c.BaseURL+"/media?type=movie", nil)
    req.Header.Set("Authorization", "Bearer "+c.Token)
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var movies []Movie
    err = json.NewDecoder(resp.Body).Decode(&movies)
    return movies, err
}
```

### Python

```python
import requests

class MoviePilotAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {token}'
        }
    
    def get_movies(self):
        response = requests.get(
            f'{self.base_url}/media?type=movie',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()
```