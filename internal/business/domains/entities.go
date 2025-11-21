package domains

import "time"

// User 用户领域实体
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // 不在JSON中返回密码
	Role      string    `json:"role" gorm:"default:'user'"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsAdmin 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// CanAccess 检查用户是否有访问权限
func (u *User) CanAccess(resource string) bool {
	// 简单的权限检查逻辑
	if u.IsAdmin() {
		return true
	}
	// 这里可以扩展更复杂的权限逻辑
	return true
}

// Media 媒体领域实体
type Media struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TMDBID      int       `json:"tmdb_id" gorm:"index"`
	Title       string    `json:"title" gorm:"not null"`
	OriginalTitle string  `json:"original_title"`
	Overview    string    `json:"overview"`
	ReleaseDate time.Time `json:"release_date"`
	PosterPath  string    `json:"poster_path"`
	BackdropPath string   `json:"backdrop_path"`
	MediaType   string    `json:"media_type" gorm:"not null"` // movie, tv
	Genres      []Genre   `json:"genres" gorm:"many2many:media_genres;"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IsMovie 检查是否为电影
func (m *Media) IsMovie() bool {
	return m.MediaType == "movie"
}

// IsTV 检查是否为电视剧
func (m *Media) IsTV() bool {
	return m.MediaType == "tv"
}

// Genre 类型领域实体
type Genre struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
	TMDBID int  `json:"tmdb_id" gorm:"uniqueIndex"`
}

// Subscribe 订阅领域实体
type Subscribe struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	MediaID     uint      `json:"media_id" gorm:"index"`
	Media       Media     `json:"media" gorm:"foreignKey:MediaID"`
	Season      int       `json:"season"` // 电视剧季数，电影为0
	Episode     int       `json:"episode"` // 电视剧集数，电影为0
	Quality     string    `json:"quality"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	UserID      uint      `json:"user_id" gorm:"index"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IsTVSeries 检查是否为电视剧订阅
func (s *Subscribe) IsTVSeries() bool {
	return s.Season > 0 || s.Episode > 0
}

// Transfer 转移领域实体
type Transfer struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SourcePath  string    `json:"source_path" gorm:"not null"`
	TargetPath  string    `json:"target_path" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, processing, completed, failed
	Progress    int       `json:"progress" gorm:"default:0"`
	FileSize    int64     `json:"file_size"`
	Speed       int64     `json:"speed"` // bytes per second
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	UserID      uint      `json:"user_id" gorm:"index"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	MediaID     *uint     `json:"media_id" gorm:"index"`
	Media       *Media    `json:"media" gorm:"foreignKey:MediaID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IsCompleted 检查转移是否完成
func (t *Transfer) IsCompleted() bool {
	return t.Status == "completed"
}

// IsFailed 检查转移是否失败
func (t *Transfer) IsFailed() bool {
	return t.Status == "failed"
}

// Duration 计算转移耗时
func (t *Transfer) Duration() time.Duration {
	if t.EndTime == nil {
		return time.Since(t.StartTime)
	}
	return t.EndTime.Sub(t.StartTime)
}