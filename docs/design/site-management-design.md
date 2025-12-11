# 站点管理系统设计文档

> **版本**: v1.0.0  
> **创建时间**: 2025-12-02  
> **设计阶段**: Phase 2 准备

---

## 📋 目录

1. [概述](#概述)
2. [系统架构](#系统架构)
3. [站点配置模型](#站点配置模型)
4. [Cookie 同步机制](#cookie-同步机制)
5. [签到任务调度](#签到任务调度)
6. [数据库设计](#数据库设计)
7. [API 设计](#api-设计)
8. [实施计划](#实施计划)

---

## 概述

### 设计目标

MoviePilot Go 站点管理系统旨在提供：
- ✅ 统一的站点配置管理
- ✅ 自动 Cookie 同步和刷新
- ✅ 定时签到任务调度
- ✅ 站点状态监控
- ✅ 流量统计和分析

### 核心特性

1. **站点配置**：支持多站点配置
2. **Cookie 管理**：自动同步和刷新
3. **签到调度**：定时自动签到
4. **状态监控**：实时监控站点状态
5. **流量统计**：记录上传下载流量

---

## 系统架构

### 架构图

```
┌─────────────────────────────────────────────┐
│              Web UI / API                   │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│         Site Management Service             │
│  ┌──────────────┐  ┌──────────────┐        │
│  │SiteService   │  │CookieService │        │
│  └──────────────┘  └──────────────┘        │
│  ┌──────────────┐  ┌──────────────┐        │
│  │CheckinService│  │MonitorService│        │
│  └──────────────┘  └──────────────┘        │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│           Task Scheduler (Cron)             │
│  ┌──────────────┐  ┌──────────────┐        │
│  │Cookie Sync   │  │Auto Checkin  │        │
│  │  (1h)        │  │  (Daily)     │        │
│  └──────────────┘  └──────────────┘        │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│              Repository Layer               │
│  ┌──────────────┐  ┌──────────────┐        │
│  │  SiteRepo    │  │CheckinLogRepo│        │
│  └──────────────┘  └──────────────┘        │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│         Database (PostgreSQL)               │
│  ┌──────┐ ┌────────┐ ┌──────────┐          │
│  │ sites│ │ cookies│ │checkin_logs│        │
│  └──────┘ └────────┘ └──────────┘          │
└─────────────────────────────────────────────┘
```

### 核心组件

1. **SiteService**：站点配置管理
2. **CookieService**：Cookie 同步和刷新
3. **CheckinService**：签到任务执行
4. **MonitorService**：站点状态监控
5. **TaskScheduler**：定时任务调度

---

## 站点配置模型

### 站点类型

MoviePilot 支持以下站点类型：

| 类型 | 说明 | 示例 |
|------|------|------|
| `pt` | PT 站点 | M-Team, HDChina |
| `public` | 公开 Tracker | RARBG, 1337x |
| `rss` | RSS 订阅源 | 自定义 RSS |

### 站点配置结构

```go
type Site struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`           // 站点名称
    URL         string    `json:"url"`            // 站点 URL
    Type        string    `json:"type"`           // pt, public, rss
    Priority    int       `json:"priority"`       // 优先级 (1-10)
    Enabled     bool      `json:"enabled"`        // 是否启用
    
    // 认证信息
    Cookie      string    `json:"cookie"`         // Cookie
    UserAgent   string    `json:"user_agent"`     // User-Agent
    Proxy       string    `json:"proxy"`          // 代理地址
    
    // 签到配置
    CheckinEnabled bool   `json:"checkin_enabled"` // 是否启用签到
    CheckinCron    string `json:"checkin_cron"`    // 签到 Cron 表达式
    CheckinURL     string `json:"checkin_url"`     // 签到 URL
    
    // 流量统计
    Upload      int64     `json:"upload"`         // 上传量（字节）
    Download    int64     `json:"download"`       // 下载量（字节）
    Ratio       float64   `json:"ratio"`          // 分享率
    
    // 状态信息
    Status      string    `json:"status"`         // active, error, disabled
    LastCheckin time.Time `json:"last_checkin"`   // 最后签到时间
    LastSync    time.Time `json:"last_sync"`      // 最后同步时间
    
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### 站点配置示例

```json
{
  "name": "M-Team",
  "url": "https://kp.m-team.cc",
  "type": "pt",
  "priority": 10,
  "enabled": true,
  "cookie": "c_secure_uid=xxx; c_secure_pass=xxx",
  "user_agent": "Mozilla/5.0...",
  "checkin_enabled": true,
  "checkin_cron": "0 8 * * *",
  "checkin_url": "https://kp.m-team.cc/attendance.php"
}
```

---

## Cookie 同步机制

### 同步策略

1. **定时同步**：每小时自动同步一次
2. **手动同步**：用户可手动触发同步
3. **失败重试**：同步失败后自动重试（最多 3 次）
4. **过期检测**：检测 Cookie 是否过期

### 同步流程

```
定时触发（每小时）
  ↓
遍历所有启用的站点
  ↓
发送 HTTP 请求到站点首页
  ↓
检查响应状态码
  ↓
提取用户信息（用户名、等级、流量等）
  ↓
更新站点状态和流量信息
  ↓
记录同步日志
  ↓
如果失败，标记站点状态为 error
```

### Cookie 验证

```go
func ValidateCookie(site *Site) error {
    // 1. 构建请求
    req, _ := http.NewRequest("GET", site.URL, nil)
    req.Header.Set("Cookie", site.Cookie)
    req.Header.Set("User-Agent", site.UserAgent)
    
    // 2. 发送请求
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    // 3. 检查状态码
    if resp.StatusCode != 200 {
        return fmt.Errorf("invalid status code: %d", resp.StatusCode)
    }
    
    // 4. 检查是否需要登录
    body, _ := io.ReadAll(resp.Body)
    if strings.Contains(string(body), "login") {
        return fmt.Errorf("cookie expired")
    }
    
    return nil
}
```

### Cookie 刷新

对于支持的站点，可以自动刷新 Cookie：

```go
func RefreshCookie(site *Site) (string, error) {
    // 1. 使用旧 Cookie 访问站点
    // 2. 提取新的 Cookie
    // 3. 更新数据库
    // 4. 返回新 Cookie
}
```

---

## 签到任务调度

### 签到策略

1. **定时签到**：每天固定时间签到（默认早上 8 点）
2. **随机延迟**：避免同时签到，随机延迟 0-30 分钟
3. **失败重试**：签到失败后 1 小时后重试
4. **签到验证**：验证签到是否成功

### 签到流程

```
Cron 触发（每天 8:00）
  ↓
遍历所有启用签到的站点
  ↓
随机延迟 0-30 分钟
  ↓
发送签到请求
  ↓
解析响应，提取签到结果
  ↓
更新签到状态和奖励
  ↓
记录签到日志
  ↓
如果失败，1 小时后重试
```

### 签到实现

```go
func Checkin(site *Site) (*CheckinResult, error) {
    // 1. 构建签到请求
    req, _ := http.NewRequest("GET", site.CheckinURL, nil)
    req.Header.Set("Cookie", site.Cookie)
    req.Header.Set("User-Agent", site.UserAgent)
    
    // 2. 发送请求
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 3. 解析响应
    body, _ := io.ReadAll(resp.Body)
    result := ParseCheckinResponse(body)
    
    // 4. 记录日志
    LogCheckin(site.ID, result)
    
    return result, nil
}
```

### Cron 表达式

| 表达式 | 说明 |
|--------|------|
| `0 8 * * *` | 每天 8:00 |
| `0 */6 * * *` | 每 6 小时 |
| `0 0 * * 0` | 每周日 0:00 |

### 签到结果

```go
type CheckinResult struct {
    Success     bool      `json:"success"`
    Message     string    `json:"message"`
    Bonus       int       `json:"bonus"`        // 获得的魔力值/积分
    ContinueDays int      `json:"continue_days"` // 连续签到天数
    CheckinTime time.Time `json:"checkin_time"`
}
```

---

## 数据库设计

### 表结构定义

#### 1. sites 表

```sql
CREATE TABLE sites (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,              -- 所属用户
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    type VARCHAR(20) NOT NULL,         -- pt, public, rss
    priority INT DEFAULT 5,
    enabled BOOLEAN DEFAULT TRUE,
    
    -- 认证信息
    cookie TEXT,
    user_agent VARCHAR(500),
    proxy VARCHAR(200),
    
    -- 签到配置
    checkin_enabled BOOLEAN DEFAULT FALSE,
    checkin_cron VARCHAR(50) DEFAULT '0 8 * * *',
    checkin_url VARCHAR(500),
    
    -- 流量统计
    upload BIGINT DEFAULT 0,
    download BIGINT DEFAULT 0,
    ratio DECIMAL(10, 2) DEFAULT 0,
    
    -- 状态信息
    status VARCHAR(20) DEFAULT 'active', -- active, error, disabled
    last_checkin TIMESTAMP,
    last_sync TIMESTAMP,
    error_message TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_enabled (enabled)
);
```

#### 2. site_cookies 表（Cookie 历史）

```sql
CREATE TABLE site_cookies (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL,
    cookie TEXT NOT NULL,
    is_valid BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    INDEX idx_site_id (site_id),
    INDEX idx_is_valid (is_valid)
);
```

#### 3. checkin_logs 表

```sql
CREATE TABLE checkin_logs (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL,
    success BOOLEAN NOT NULL,
    message TEXT,
    bonus INT DEFAULT 0,
    continue_days INT DEFAULT 0,
    error_message TEXT,
    checkin_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    INDEX idx_site_id (site_id),
    INDEX idx_checkin_time (checkin_time),
    INDEX idx_success (success)
);
```

#### 4. site_stats 表（流量统计）

```sql
CREATE TABLE site_stats (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL,
    date DATE NOT NULL,
    upload_delta BIGINT DEFAULT 0,      -- 当天上传增量
    download_delta BIGINT DEFAULT 0,    -- 当天下载增量
    upload_total BIGINT DEFAULT 0,      -- 总上传
    download_total BIGINT DEFAULT 0,    -- 总下载
    ratio DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    UNIQUE (site_id, date),
    INDEX idx_site_id (site_id),
    INDEX idx_date (date)
);
```

#### 5. sync_logs 表（同步日志）

```sql
CREATE TABLE sync_logs (
    id SERIAL PRIMARY KEY,
    site_id INT NOT NULL,
    success BOOLEAN NOT NULL,
    duration_ms INT,                    -- 同步耗时（毫秒）
    error_message TEXT,
    synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    INDEX idx_site_id (site_id),
    INDEX idx_synced_at (synced_at)
);
```

---

## API 设计

### 站点管理 API

#### 1. 创建站点

```
POST /api/v1/sites
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "name": "M-Team",
  "url": "https://kp.m-team.cc",
  "type": "pt",
  "cookie": "c_secure_uid=xxx; c_secure_pass=xxx",
  "checkin_enabled": true
}

Response: 201 Created
{
  "code": 201,
  "message": "站点创建成功",
  "data": {
    "id": 1,
    "name": "M-Team",
    "status": "active"
  }
}
```

#### 2. 获取站点列表

```
GET /api/v1/sites?type=pt&enabled=true
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "M-Team",
        "url": "https://kp.m-team.cc",
        "type": "pt",
        "status": "active",
        "upload": 1099511627776,
        "download": 549755813888,
        "ratio": 2.0
      }
    ],
    "total": 1
  }
}
```

#### 3. 更新站点

```
PUT /api/v1/sites/:id
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "cookie": "new_cookie_value",
  "checkin_enabled": false
}

Response: 200 OK
{
  "code": 200,
  "message": "更新成功"
}
```

#### 4. 删除站点

```
DELETE /api/v1/sites/:id
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "message": "删除成功"
}
```

### Cookie 管理 API

#### 1. 验证 Cookie

```
POST /api/v1/sites/:id/validate
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "data": {
    "valid": true,
    "username": "john_doe",
    "level": "VIP",
    "upload": 1099511627776,
    "download": 549755813888
  }
}
```

#### 2. 同步站点信息

```
POST /api/v1/sites/:id/sync
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "message": "同步成功",
  "data": {
    "upload": 1099511627776,
    "download": 549755813888,
    "ratio": 2.0
  }
}
```

#### 3. 批量同步

```
POST /api/v1/sites/sync-all
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "data": {
    "success": 5,
    "failed": 1,
    "total": 6
  }
}
```

### 签到管理 API

#### 1. 手动签到

```
POST /api/v1/sites/:id/checkin
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "data": {
    "success": true,
    "message": "签到成功",
    "bonus": 100,
    "continue_days": 30
  }
}
```

#### 2. 获取签到历史

```
GET /api/v1/sites/:id/checkin-logs?page=1&limit=20
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "data": {
    "items": [
      {
        "id": 1,
        "success": true,
        "message": "签到成功",
        "bonus": 100,
        "checkin_time": "2025-12-02T08:00:00Z"
      }
    ],
    "total": 30
  }
}
```

#### 3. 批量签到

```
POST /api/v1/sites/checkin-all
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "data": {
    "success": 5,
    "failed": 1,
    "total": 6
  }
}
```

### 统计分析 API

#### 1. 获取流量统计

```
GET /api/v1/sites/:id/stats?start_date=2025-11-01&end_date=2025-12-01
Authorization: Bearer <token>

Response: 200 OK
{
  "code": 200,
  "data": {
    "daily_stats": [
      {
        "date": "2025-11-01",
        "upload": 10737418240,
        "download": 5368709120
      }
    ],
    "total_upload": 1099511627776,
    "total_download": 549755813888,
    "avg_ratio": 2.0
  }
}
```

---

## 实施计划

### Week 7 实施任务

#### Day 1-2: 数据库和模型

- [ ] 创建站点相关数据库表
- [ ] 定义 GORM 模型（Site、CheckinLog、SiteStats）
- [ ] 实现 Repository 层

#### Day 3-4: 业务逻辑

- [ ] 实现 SiteService（CRUD）
- [ ] 实现 CookieService（验证、同步）
- [ ] 实现 CheckinService（签到逻辑）
- [ ] 实现 MonitorService（状态监控）

#### Day 5: 任务调度

- [ ] 集成 Cron 调度器
- [ ] 实现 Cookie 同步任务
- [ ] 实现自动签到任务
- [ ] 编写单元测试

### 验收标准

- ✅ 站点可以正常添加、编辑、删除
- ✅ Cookie 可以自动同步和验证
- ✅ 签到任务可以定时执行
- ✅ 流量统计正常记录
- ✅ API 文档完整

---

## 附录

### A. 站点模板

常见 PT 站点的配置模板：

```json
{
  "M-Team": {
    "url": "https://kp.m-team.cc",
    "checkin_url": "https://kp.m-team.cc/attendance.php",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
  },
  "HDChina": {
    "url": "https://hdchina.org",
    "checkin_url": "https://hdchina.org/attendance.php",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
  }
}
```

### B. Cookie 提取工具

提供浏览器插件或脚本，帮助用户提取 Cookie：

```javascript
// Chrome Console
document.cookie
```

### C. 签到结果解析

不同站点的签到响应格式不同，需要针对性解析：

```go
func ParseCheckinResponse(siteType string, body []byte) *CheckinResult {
    switch siteType {
    case "mteam":
        return parseMTeamCheckin(body)
    case "hdchina":
        return parseHDChinaCheckin(body)
    default:
        return parseGenericCheckin(body)
    }
}
```

---

**文档状态**: ✅ 设计完成，待实施  
**下一步**: Week 7 开始实施
