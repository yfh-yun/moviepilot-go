# 数据库设计文档

## 🗄️ 数据库概览

MoviePilot Go 使用 PostgreSQL 作为主数据库，Redis 作为缓存层，采用 GORM 作为 ORM 框架。

### 技术栈
- **主数据库**: PostgreSQL 14+
- **缓存**: Redis 6+
- **ORM**: GORM v2
- **连接池**: 内置连接池管理
- **迁移**: GORM AutoMigrate + 自定义迁移

## 📊 数据库架构

### 核心表结构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      users      │    │      media      │    │   transfers     │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ id (PK)         │    │ id (PK)         │    │ id (PK)         │
│ username        │    │ title           │    │ source_path      │
│ email           │    │ type            │    │ destination_path │
│ password_hash   │    │ year            │    │ status           │
│ role            │    │ genre           │    │ progress         │
│ created_at      │    │ rating          │    │ created_at       │
│ updated_at      │    │ overview        │    │ completed_at     │
└─────────────────┘    │ created_at      │    └─────────────────┘
         │              │ updated_at      │             │
         │              └─────────────────┘             │
         │                      │                      │
         └──────────────────────┼──────────────────────┘
                                │
                    ┌─────────────────┐
                    │   subtitles     │
                    ├─────────────────┤
                    │ id (PK)         │
                    │ media_id (FK)   │
                    │ language        │
                    │ format          │
                    │ path            │
                    │ created_at      │
                    └─────────────────┘
```

## 📋 数据模型定义

### 1. 用户模型 (User)

```go
// internal/models/user.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID           string         `json:"id" gorm:"primaryKey;type:varchar(36);default:uuid_generate_v4()"`
    Username     string         `json:"username" gorm:"uniqueIndex;not null;size:50"`
    Email        string         `json:"email" gorm:"uniqueIndex;not null;size:255"`
    PasswordHash string         `json:"-" gorm:"not null;size:255"`
    Role         string         `json:"role" gorm:"not null;size:20;default:'user'"`
    Avatar       string         `json:"avatar" gorm:"size:500"`
    IsActive     bool           `json:"is_active" gorm:"not null;default:true"`
    LastLoginAt  *time.Time     `json:"last_login_at"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联关系
    Transfers    []Transfer     `json:"transfers,omitempty" gorm:"foreignKey:UserID"`
    Notifications []Notification `json:"notifications,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (User) TableName() string {
    return "users"
}

// BeforeCreate GORM 钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = generateUUID()
    }
    return nil
}
```

### 2. 媒体模型 (Media)

```go
// internal/models/media.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Media struct {
    ID           string         `json:"id" gorm:"primaryKey;type:varchar(36);default:uuid_generate_v4()"`
    Title        string         `json:"title" gorm:"not null;size:500;index"`
    OriginalTitle string        `json:"original_title" gorm:"size:500"`
    Type         string         `json:"type" gorm:"not null;size:20;index;check:type IN ('movie', 'tv', 'documentary')"`
    Year         int            `json:"year" gorm:"index"`
    Genre        []string       `json:"genre" gorm:"serializer:json"`
    Rating       float64        `json:"rating" gorm:"index"`
    Overview     string         `json:"overview" gorm:"type:text"`
    PosterURL    string         `json:"poster_url" gorm:"size:500"`
    BackdropURL  string         `json:"backdrop_url" gorm:"size:500"`
    IMDBID       string         `json:"imdb_id" gorm:"size:20;index"`
    TMDBID       int            `json:"tmdb_id" gorm:"index"`
    Status       string         `json:"status" gorm:"not null;size:20;default:'active';index"`
    Source       string         `json:"source" gorm:"size:100;index"`
    FilePath     string         `json:"file_path" gorm:"size:1000"`
    FileSize     int64          `json:"file_size"`
    Duration     int            `json:"duration"` // 秒
    Quality      string         `json:"quality" gorm:"size:20"`
    Resolution   string         `json:"resolution" gorm:"size:20"`
    Codec        string         `json:"codec" gorm:"size:50"`
    Container    string         `json:"container" gorm:"size:20"`
    CreatedAt    time.Time      `json:"created_at" gorm:"index"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联关系
    Transfers    []Transfer     `json:"transfers,omitempty" gorm:"foreignKey:MediaID"`
    Subtitles    []Subtitle     `json:"subtitles,omitempty" gorm:"foreignKey:MediaID"`
}

func (Media) TableName() string {
    return "media"
}

func (m *Media) BeforeCreate(tx *gorm.DB) error {
    if m.ID == "" {
        m.ID = generateUUID()
    }
    return nil
}
```

### 3. 传输模型 (Transfer)

```go
// internal/models/transfer.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Transfer struct {
    ID              string         `json:"id" gorm:"primaryKey;type:varchar(36);default:uuid_generate_v4()"`
    UserID          string         `json:"user_id" gorm:"not null;index"`
    MediaID         *string        `json:"media_id" gorm:"index"`
    SourcePath      string         `json:"source_path" gorm:"not null;size:1000"`
    DestinationPath string         `json:"destination_path" gorm:"not null;size:1000"`
    SourceType      string         `json:"source_type" gorm:"not null;size:20;index;check:source_type IN ('local', 'remote', 'torrent', 'magnet')"`
    Status          string         `json:"status" gorm:"not null;size:20;index;default:'pending';check:status IN ('pending', 'running', 'completed', 'failed', 'paused', 'cancelled')"`
    Progress        float64        `json:"progress" gorm:"default:0;check:progress >= 0 AND progress <= 100"`
    Speed           int64         `json:"speed"` // bytes per second
    TotalSize       int64         `json:"total_size"`
    TransferredSize int64         `json:"transferred_size"`
    Priority        string         `json:"priority" gorm:"not null;size:20;default:'normal';check:priority IN ('low', 'normal', 'high')"`
    AutoStart       bool           `json:"auto_start" gorm:"not null;default:true"`
    RetryCount      int            `json:"retry_count" gorm:"default:0"`
    MaxRetries      int            `json:"max_retries" gorm:"default:3"`
    ErrorMessage    string         `json:"error_message" gorm:"type:text"`
    Metadata        string         `json:"metadata" gorm:"type:jsonb"`
    StartedAt       *time.Time     `json:"started_at"`
    CompletedAt     *time.Time     `json:"completed_at"`
    CreatedAt       time.Time      `json:"created_at" gorm:"index"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联关系
    User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Media   *Media   `json:"media,omitempty" gorm:"foreignKey:MediaID"`
}

func (Transfer) TableName() string {
    return "transfers"
}

func (t *Transfer) BeforeCreate(tx *gorm.DB) error {
    if t.ID == "" {
        t.ID = generateUUID()
    }
    return nil
}
```

### 4. 字幕模型 (Subtitle)

```go
// internal/models/subtitle.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Subtitle struct {
    ID          string         `json:"id" gorm:"primaryKey;type:varchar(36);default:uuid_generate_v4()"`
    MediaID     string         `json:"media_id" gorm:"not null;index"`
    Language    string         `json:"language" gorm:"not null;size:10;index"`
    Format      string         `json:"format" gorm:"not null;size:10;check:format IN ('srt', 'ass', 'vtt')"`
    Path        string         `json:"path" gorm:"not null;size:1000"`
    Size        int64          `json:"size"`
    Encoding    string         `json:"encoding" gorm:"size:20;default:'utf-8'"`
    Source      string         `json:"source" gorm:"size:100"` // 字幕来源
    IsDefault   bool           `json:"is_default" gorm:"not null;default:false"`
    IsExternal  bool           `json:"is_external" gorm:"not null;default:true"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联关系
    Media Media `json:"media,omitempty" gorm:"foreignKey:MediaID"`
}

func (Subtitle) TableName() string {
    return "subtitles"
}

func (s *Subtitle) BeforeCreate(tx *gorm.DB) error {
    if s.ID == "" {
        s.ID = generateUUID()
    }
    return nil
}
```

### 5. 通知模型 (Notification)

```go
// internal/models/notification.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Notification struct {
    ID        string         `json:"id" gorm:"primaryKey;type:varchar(36);default:uuid_generate_v4()"`
    UserID    string         `json:"user_id" gorm:"not null;index"`
    Type      string         `json:"type" gorm:"not null;size:50;index"`
    Title     string         `json:"title" gorm:"not null;size:500"`
    Content   string         `json:"content" gorm:"type:text"`
    IsRead    bool           `json:"is_read" gorm:"not null;default:false;index"`
    Level     string         `json:"level" gorm:"not null;size:20;default:'info';check:level IN ('info', 'warning', 'error', 'success')"`
    Data      string         `json:"data" gorm:"type:jsonb"` // 额外数据
    ExpiresAt *time.Time     `json:"expires_at"`
    CreatedAt time.Time      `json:"created_at" gorm:"index"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联关系
    User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Notification) TableName() string {
    return "notifications"
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
    if n.ID == "" {
        n.ID = generateUUID()
    }
    return nil
}
```

## 🔧 数据库配置

### 1. GORM 配置

```go
// pkg/database/database.go
package database

import (
    "fmt"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    
    "moviepilot-go/internal/models"
    "moviepilot-go/pkg/logger"
)

type Config struct {
    Host            string
    Port            int
    User            string
    Password        string
    DBName          string
    SSLMode         string
    MaxIdleConns    int
    MaxOpenConns    int
    ConnMaxLifetime time.Duration
    LogLevel        logger.LogLevel
}

type DB struct {
    *gorm.DB
}

func NewDB(config Config, appLogger logger.Logger) (*DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
        config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
    )
    
    gormConfig := &gorm.Config{
        Logger: logger.New(
            appLogger,
            logger.Config{
                SlowThreshold:             time.Second,
                LogLevel:                  config.LogLevel,
                IgnoreRecordNotFoundError: true,
                Colorful:                  false,
            },
        ),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
    }
    
    db, err := gorm.Open(postgres.Open(dsn), gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
    }
    
    // 配置连接池
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
    
    return &DB{DB: db}, nil
}

func (db *DB) AutoMigrate() error {
    return db.DB.AutoMigrate(
        &models.User{},
        &models.Media{},
        &models.Transfer{},
        &models.Subtitle{},
        &models.Notification{},
    )
}

func (db *DB) GetDB() *gorm.DB {
    return db.DB
}
```

### 2. 数据库迁移

```go
// pkg/database/migration.go
package database

import (
    "gorm.io/gorm"
)

type Migration struct {
    Version     string `gorm:"primaryKey"`
    Description string
    ExecutedAt  time.Time
}

func (m *Migration) TableName() string {
    return "schema_migrations"
}

type Migrator struct {
    db *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
    return &Migrator{db: db}
}

func (m *Migrator) RunMigrations() error {
    // 创建迁移表
    if err := m.db.AutoMigrate(&Migration{}); err != nil {
        return fmt.Errorf("failed to create migration table: %w", err)
    }
    
    migrations := []MigrationFunc{
        m.addMediaIndexes,
        m.addTransferIndexes,
        m.addNotificationIndexes,
        m.createMediaFullTextIndex,
    }
    
    for _, migration := range migrations {
        if err := migration(); err != nil {
            return fmt.Errorf("migration failed: %w", err)
        }
    }
    
    return nil
}

type MigrationFunc func() error

func (m *Migrator) addMediaIndexes() error {
    version := "20240101_001_add_media_indexes"
    if m.isMigrationExecuted(version) {
        return nil
    }
    
    indexes := []string{
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_media_title_gin ON media USING gin(title gin_trgm_ops)",
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_media_type_year ON media(type, year DESC)",
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_media_rating ON media(rating DESC) WHERE rating > 0",
        "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_media_status_source ON media(status, source)",
    }
    
    for _, index := range indexes {
        if err := m.db.Exec(index).Error; err != nil {
            return fmt.Errorf("failed to create index: %w", err)
        }
    }
    
    return m.recordMigration(version, "Add media indexes")
}

func (m *Migrator) createMediaFullTextIndex() error {
    version := "20240101_002_media_fulltext"
    if m.isMigrationExecuted(version) {
        return nil
    }
    
    // 启用 pg_trgm 扩展
    if err := m.db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
        return fmt.Errorf("failed to enable pg_trgm extension: %w", err)
    }
    
    return m.recordMigration(version, "Create media full-text search index")
}

func (m *Migrator) isMigrationExecuted(version string) bool {
    var count int64
    m.db.Model(&Migration{}).Where("version = ?", version).Count(&count)
    return count > 0
}

func (m *Migrator) recordMigration(version, description string) error {
    migration := Migration{
        Version:     version,
        Description: description,
        ExecutedAt:  time.Now(),
    }
    return m.db.Create(&migration).Error
}
```

## 📈 性能优化

### 1. 索引策略

#### 主要索引
```sql
-- 用户表索引
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY idx_users_username ON users(username);
CREATE INDEX CONCURRENTLY idx_users_role_active ON users(role, is_active);

-- 媒体表索引
CREATE INDEX CONCURRENTLY idx_media_type ON media(type);
CREATE INDEX CONCURRENTLY idx_media_year ON media(year DESC);
CREATE INDEX CONCURRENTLY idx_media_rating ON media(rating DESC) WHERE rating > 0;
CREATE INDEX CONCURRENTLY idx_media_status ON media(status);
CREATE INDEX CONCURRENTLY idx_media_created_at ON media(created_at DESC);

-- 传输表索引
CREATE INDEX CONCURRENTLY idx_transfers_user_id ON transfers(user_id);
CREATE INDEX CONCURRENTLY idx_transfers_media_id ON transfers(media_id);
CREATE INDEX CONCURRENTLY idx_transfers_status ON transfers(status);
CREATE INDEX CONCURRENTLY idx_transfers_created_at ON transfers(created_at DESC);

-- 复合索引
CREATE INDEX CONCURRENTLY idx_media_type_year_rating ON media(type, year DESC, rating DESC);
CREATE INDEX CONCURRENTLY idx_transfers_user_status ON transfers(user_id, status);
```

#### 全文搜索索引
```sql
-- 启用 pg_trgm 扩展
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 创建 GIN 索引用于模糊搜索
CREATE INDEX CONCURRENTLY idx_media_title_gin ON media USING gin(title gin_trgm_ops);
CREATE INDEX CONCURRENTLY idx_media_overview_gin ON media USING gin(overview gin_trgm_ops);

-- 创建全文搜索索引
ALTER TABLE media ADD COLUMN search_vector tsvector;
CREATE INDEX CONCURRENTLY idx_media_search ON media USING gin(search_vector);

-- 更新搜索向量的触发器
CREATE OR REPLACE FUNCTION update_media_search_vector()
RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', 
        COALESCE(NEW.title, '') || ' ' || 
        COALESCE(NEW.original_title, '') || ' ' || 
        COALESCE(NEW.overview, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_media_search_vector
    BEFORE INSERT OR UPDATE ON media
    FOR EACH ROW EXECUTE FUNCTION update_media_search_vector();
```

### 2. 查询优化

#### 分页查询
```go
// repositories/media_repository.go
func (r *mediaRepository) ListMedia(ctx context.Context, filter MediaFilter, pagination Pagination) ([]*models.Media, int64, error) {
    var media []*models.Media
    var total int64
    
    query := r.db.WithContext(ctx).Model(&models.Media{})
    
    // 应用过滤条件
    if filter.Type != "" {
        query = query.Where("type = ?", filter.Type)
    }
    if filter.Year > 0 {
        query = query.Where("year = ?", filter.Year)
    }
    if filter.MinRating > 0 {
        query = query.Where("rating >= ?", filter.MinRating)
    }
    if filter.Search != "" {
        query = query.Where("search_vector @@ plainto_tsquery('english', ?)", filter.Search)
    }
    
    // 获取总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // 应用排序和分页
    offset := (pagination.Page - 1) * pagination.Size
    err := query.
        Preload("Subtitles").
        Order(fmt.Sprintf("%s %s", pagination.Sort, pagination.Order)).
        Offset(offset).
        Limit(pagination.Size).
        Find(&media).Error
    
    return media, total, err
}
```

#### 批量操作
```go
// repositories/user_repository.go
func (r *userRepository) CreateUsersBatch(ctx context.Context, users []*models.User) error {
    if len(users) == 0 {
        return nil
    }
    
    // 使用事务确保数据一致性
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        batchSize := 100
        for i := 0; i < len(users); i += batchSize {
            end := i + batchSize
            if end > len(users) {
                end = len(users)
            }
            
            batch := users[i:end]
            if err := tx.CreateInBatches(batch, batchSize).Error; err != nil {
                return err
            }
        }
        return nil
    })
}
```

## 🔍 查询示例

### 1. 复杂查询

```go
// 获取用户最近的传输记录
func (r *transferRepository) GetUserRecentTransfers(ctx context.Context, userID string, limit int) ([]*models.Transfer, error) {
    var transfers []*models.Transfer
    
    err := r.db.WithContext(ctx).
        Preload("Media").
        Preload("User").
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Limit(limit).
        Find(&transfers).Error
    
    return transfers, err
}

// 搜索媒体（支持多种条件）
func (r *mediaRepository) SearchMedia(ctx context.Context, params SearchParams) ([]*models.Media, int64, error) {
    var media []*models.Media
    var total int64
    
    query := r.db.WithContext(ctx).Model(&models.Media{})
    
    // 文本搜索
    if params.Query != "" {
        query = query.Where("search_vector @@ plainto_tsquery('english', ?)", params.Query)
    }
    
    // 类型过滤
    if len(params.Types) > 0 {
        query = query.Where("type IN ?", params.Types)
    }
    
    // 年份范围
    if params.YearFrom > 0 {
        query = query.Where("year >= ?", params.YearFrom)
    }
    if params.YearTo > 0 {
        query = query.Where("year <= ?", params.YearTo)
    }
    
    // 评分范围
    if params.MinRating > 0 {
        query = query.Where("rating >= ?", params.MinRating)
    }
    if params.MaxRating > 0 {
        query = query.Where("rating <= ?", params.MaxRating)
    }
    
    // 获取总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // 应用排序和分页
    err := query.
        Preload("Subtitles").
        Order(params.SortBy + " " + params.SortOrder).
        Offset(params.Offset).
        Limit(params.Limit).
        Find(&media).Error
    
    return media, total, err
}
```

### 2. 聚合查询

```go
// 获取传输统计
func (r *transferRepository) GetTransferStats(ctx context.Context, userID string) (*TransferStats, error) {
    var stats TransferStats
    
    // 总传输数
    err := r.db.WithContext(ctx).
        Model(&models.Transfer{}).
        Where("user_id = ?", userID).
        Count(&stats.TotalTransfers).Error
    if err != nil {
        return nil, err
    }
    
    // 按状态分组统计
    var statusStats []struct {
        Status string
        Count  int64
    }
    err = r.db.WithContext(ctx).
        Model(&models.Transfer{}).
        Select("status, COUNT(*) as count").
        Where("user_id = ?", userID).
        Group("status").
        Scan(&statusStats).Error
    if err != nil {
        return nil, err
    }
    
    for _, stat := range statusStats {
        switch stat.Status {
        case "completed":
            stats.CompletedTransfers = stat.Count
        case "failed":
            stats.FailedTransfers = stat.Count
        case "running":
            stats.RunningTransfers = stat.Count
        }
    }
    
    // 总传输大小
    err = r.db.WithContext(ctx).
        Model(&models.Transfer{}).
        Select("COALESCE(SUM(total_size), 0) as total_size").
        Where("user_id = ? AND status = 'completed'", userID).
        Scan(&stats.TotalTransferredSize).Error
    
    return &stats, err
}
```

## 🛡️ 数据安全

### 1. 数据加密

```go
// pkg/security/encryption.go
package security

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

type Encryptor struct {
    gcm cipher.AEAD
}

func NewEncryptor(key string) (*Encryptor, error) {
    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    return &Encryptor{gcm: gcm}, nil
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
    nonce := make([]byte, e.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }
    
    nonceSize := e.gcm.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}
```

### 2. 敏感数据处理

```go
// models/user.go (续)
type User struct {
    // ... 其他字段
    
    // 加密字段
    EncryptedAPIKey string `json:"-" gorm:"column:encrypted_api_key;size:500"`
}

// APIKey 访问器
func (u *User) SetAPIKey(apiKey string, encryptor *security.Encryptor) error {
    if apiKey == "" {
        u.EncryptedAPIKey = ""
        return nil
    }
    
    encrypted, err := encryptor.Encrypt(apiKey)
    if err != nil {
        return fmt.Errorf("failed to encrypt API key: %w", err)
    }
    
    u.EncryptedAPIKey = encrypted
    return nil
}

func (u *User) GetAPIKey(encryptor *security.Encryptor) (string, error) {
    if u.EncryptedAPIKey == "" {
        return "", nil
    }
    
    return encryptor.Decrypt(u.EncryptedAPIKey)
}
```

## 📊 监控和维护

### 1. 数据库监控

```go
// pkg/database/monitor.go
package database

import (
    "context"
    "time"
    
    "gorm.io/gorm"
)

type Monitor struct {
    db *gorm.DB
}

type DBStats struct {
    Connections     int
    IdleConnections int
    OpenConnections int
    InUse          int
    MaxIdleClosed  int64
    MaxLifetimeClosed int64
    WaitCount      int64
    WaitDuration   time.Duration
    MaxIdleClosedTotal int64
}

func NewMonitor(db *gorm.DB) *Monitor {
    return &Monitor{db: db}
}

func (m *Monitor) GetStats(ctx context.Context) (*DBStats, error) {
    sqlDB, err := m.db.DB()
    if err != nil {
        return nil, err
    }
    
    stats := sqlDB.Stats()
    
    return &DBStats{
        Connections:         stats.MaxOpenConnections,
        IdleConnections:     stats.Idle,
        OpenConnections:     stats.OpenConnections,
        InUse:              stats.InUse,
        MaxIdleClosed:       stats.MaxIdleClosed,
        MaxLifetimeClosed:   stats.MaxLifetimeClosed,
        WaitCount:          stats.WaitCount,
        WaitDuration:       stats.WaitDuration,
        MaxIdleClosedTotal: stats.MaxIdleClosedTotal,
    }, nil
}

func (m *Monitor) CheckHealth(ctx context.Context) error {
    sqlDB, err := m.db.DB()
    if err != nil {
        return err
    }
    
    return sqlDB.PingContext(ctx)
}
```

### 2. 性能分析

```sql
-- 查询慢查询
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- 查看表大小
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 查看索引使用情况
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes 
ORDER BY idx_scan DESC;
```

---

**注意**: 数据库设计和优化是一个持续的过程，需要根据实际使用情况和性能指标进行调整。建议定期监控数据库性能，并根据需要进行优化。