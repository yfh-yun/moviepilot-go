# Week 8 Day 1 实施指南

> **任务**: 订阅数据模型和 Repository  
> **日期**: 2025-12-03  
> **状态**: 🚀 开始

---

## 📋 任务清单

### 1. 数据库迁移（3 个表）

#### subscriptions 表 ✅
- [x] 创建 `000012_create_subscriptions_table.up.sql`
- [x] 创建 `000012_create_subscriptions_table.down.sql`

#### subscription_items 表 ⏳
- [ ] 创建 `000013_create_subscription_items_table.up.sql`
- [ ] 创建 `000013_create_subscription_items_table.down.sql`

```sql
CREATE TABLE IF NOT EXISTS subscription_items (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    torrent_url VARCHAR(500),
    magnet_link TEXT,
    size BIGINT,
    seeders INT,
    leechers INT,
    publish_date TIMESTAMP,
    matched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    downloaded BOOLEAN DEFAULT false,
    download_task_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
);
```

#### subscription_history 表 ⏳
- [ ] 创建 `000014_create_subscription_history_table.up.sql`
- [ ] 创建 `000014_create_subscription_history_table.down.sql`

```sql
CREATE TABLE IF NOT EXISTS subscription_history (
    id SERIAL PRIMARY KEY,
    subscription_id INT NOT NULL,
    action VARCHAR(50) NOT NULL,  -- refresh, match, download
    items_found INT DEFAULT 0,
    items_matched INT DEFAULT 0,
    items_downloaded INT DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
);
```

### 2. GORM 模型（3 个）

#### Subscription 模型 ⏳
文件: `internal/models/subscription.go`

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Subscription struct {
    ID             uint           `gorm:"primaryKey" json:"id"`
    UserID         uint           `gorm:"not null;index" json:"user_id"`
    Name           string         `gorm:"size:100;not null" json:"name"`
    Type           string         `gorm:"size:20;not null" json:"type"` // movie, tv
    TMDBID         *int           `gorm:"index" json:"tmdb_id,omitempty"`
    IMDBID         *string        `gorm:"size:20" json:"imdb_id,omitempty"`
    Season         *int           `json:"season,omitempty"`
    Quality        string         `gorm:"size:50" json:"quality,omitempty"`
    Resolution     string         `gorm:"size:20" json:"resolution,omitempty"`
    Source         string         `gorm:"size:50" json:"source,omitempty"`
    Codec          string         `gorm:"size:20" json:"codec,omitempty"`
    FilterRules    string         `gorm:"type:jsonb" json:"filter_rules,omitempty"`
    Enabled        bool           `gorm:"default:true;not null" json:"enabled"`
    AutoDownload   bool           `gorm:"default:true;not null" json:"auto_download"`
    NotifyOnMatch  bool           `gorm:"default:false;not null" json:"notify_on_match"`
    LastRefreshAt  *time.Time     `json:"last_refresh_at,omitempty"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
    DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
    
    // 关联
    User  User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Items []SubscriptionItem  `gorm:"foreignKey:SubscriptionID" json:"items,omitempty"`
}
```

#### SubscriptionItem 模型 ⏳
文件: `internal/models/subscription_item.go`

#### SubscriptionHistory 模型 ⏳
文件: `internal/models/subscription_history.go`

### 3. Repository 层（2 个）

#### SubscriptionRepository ⏳
文件: `internal/repositories/subscription_repository.go`

**接口定义**:
```go
type SubscriptionRepository interface {
    Create(ctx context.Context, subscription *models.Subscription) error
    GetByID(ctx context.Context, id uint) (*models.Subscription, error)
    Update(ctx context.Context, subscription *models.Subscription) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, userID uint, page, limit int) ([]*models.Subscription, int64, error)
    GetEnabledSubscriptions(ctx context.Context, userID uint) ([]*models.Subscription, error)
    GetByTMDBID(ctx context.Context, userID uint, tmdbID int, season *int) (*models.Subscription, error)
}
```

#### SubscriptionItemRepository ⏳
文件: `internal/repositories/subscription_item_repository.go`

**接口定义**:
```go
type SubscriptionItemRepository interface {
    Create(ctx context.Context, item *models.SubscriptionItem) error
    GetByID(ctx context.Context, id uint) (*models.SubscriptionItem, error)
    GetBySubscriptionID(ctx context.Context, subscriptionID uint, page, limit int) ([]*models.SubscriptionItem, int64, error)
    MarkAsDownloaded(ctx context.Context, id uint, taskID uint) error
    Delete(ctx context.Context, id uint) error
}
```

---

## 🎯 完成标准

- [ ] 3 个数据库迁移脚本创建完成
- [ ] 3 个 GORM 模型定义完成
- [ ] 2 个 Repository 接口和实现完成
- [ ] 所有代码编译通过
- [ ] 运行迁移成功

---

## 🚀 执行步骤

### Step 1: 创建迁移脚本
```bash
# 创建 subscription_items 迁移
touch database/migrations/000013_create_subscription_items_table.up.sql
touch database/migrations/000013_create_subscription_items_table.down.sql

# 创建 subscription_history 迁移
touch database/migrations/000014_create_subscription_history_table.up.sql
touch database/migrations/000014_create_subscription_history_table.down.sql
```

### Step 2: 创建 GORM 模型
```bash
# 创建模型文件
touch internal/models/subscription.go
touch internal/models/subscription_item.go
touch internal/models/subscription_history.go
```

### Step 3: 创建 Repository
```bash
# 创建 Repository 文件
touch internal/repositories/subscription_repository.go
touch internal/repositories/subscription_item_repository.go
```

### Step 4: 运行迁移
```bash
# 执行迁移
make migrate-up

# 验证迁移
make migrate-status
```

### Step 5: 编译验证
```bash
# 编译模型
go build ./internal/models/...

# 编译 Repository
go build ./internal/repositories/...
```

---

## 📊 预期成果

| 类别 | 文件数 | 代码行数 |
|------|--------|---------|
| 数据库迁移 | 6 | - |
| GORM 模型 | 3 | 200 |
| Repository | 2 | 250 |
| **总计** | **11** | **450** |

---

## 💡 技术要点

### 订阅匹配规则
```json
{
  "include_keywords": ["1080p", "BluRay"],
  "exclude_keywords": ["CAM", "TS"],
  "min_size": 1073741824,  // 1GB
  "max_size": 10737418240,  // 10GB
  "min_seeders": 5
}
```

### 订阅类型
- `movie`: 电影订阅
- `tv`: 剧集订阅（可指定季数）

### 质量选项
- 分辨率: 720p, 1080p, 2160p (4K)
- 来源: BluRay, WEB-DL, HDTV
- 编码: H264, H265 (HEVC), AV1

---

## 🔍 注意事项

1. **外键约束**: 确保 `user_id` 和 `subscription_id` 的外键正确设置
2. **软删除**: Subscription 使用软删除，关联数据需要考虑级联
3. **索引优化**: 为常用查询字段添加索引
4. **JSONB 字段**: `filter_rules` 使用 JSONB 存储复杂规则

---

**状态**: 🚀 Day 1 开始  
**下一步**: 完成所有迁移和模型后，进入 Day 2 的服务层开发
