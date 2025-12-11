package user

// UserBase 用户基础信息
type UserBase struct {
	// 用户名
	Name string `json:"name"`
	// 邮箱，未启用
	Email string `json:"email,omitempty"`
	// 状态
	IsActive bool `json:"is_active,omitempty"`
	// 超级管理员
	IsSuperuser bool `json:"is_superuser,omitempty"`
	// 头像
	Avatar string `json:"avatar,omitempty"`
	// 是否开启二次验证
	IsOTP bool `json:"is_otp,omitempty"`
	// 权限
	Permissions map[string]any `json:"permissions,omitempty"`
	// 个性化设置
	Settings map[string]any `json:"settings,omitempty"`
}

// UserCreate 创建用户请求
type UserCreate struct {
	UserBase
	// 密码
	Password string `json:"password,omitempty"`
}

// UserUpdate 更新用户请求
type UserUpdate struct {
	UserBase
	// ID
	ID int `json:"id"`
	// 密码
	Password string `json:"password,omitempty"`
}

// UserInDBBase 数据库用户基础信息
type UserInDBBase struct {
	UserBase
	// ID
	ID int `json:"id,omitempty"`
}

// User 用户信息（API返回）
type User struct {
	UserInDBBase
}

// UserInDB 数据库用户信息（含密码哈希）
type UserInDB struct {
	UserInDBBase
	// 密码哈希
	HashedPassword string `json:"-"`
}
