package models

// User 用户模型
type User struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 用户名
	Name string `json:"name" gorm:"not null;uniqueIndex"`
	
	// 邮箱
	Email string `json:"email,omitempty"`
	
	// 状态
	IsActive bool `json:"is_active,omitempty" gorm:"default:true"`
	
	// 超级管理员
	IsSuperuser bool `json:"is_superuser" gorm:"default:false"`
	
	// 头像
	Avatar string `json:"avatar,omitempty"`
	
	// 是否开启二次验证
	IsOtp bool `json:"is_otp,omitempty" gorm:"default:false;column:is_otp"`
	
	// 权限
	Permissions map[string]interface{} `json:"permissions,omitempty" gorm:"serializer:json"`
	
	// 个性化设置
	Settings map[string]interface{} `json:"settings,omitempty" gorm:"serializer:json"`
	
	// 哈希密码
	HashedPassword string `json:"hashed_password" gorm:"column:hashed_password"`
}

// TableName 设置表名
func (User) TableName() string {
	return "user"
}

// UserBase Shared properties
type UserBase struct {
	// 用户��?	Name string `json:"name"`
	// 邮箱，未启用
	Email string `json:"email,omitempty"`
	// 状��?	IsActive bool `json:"is_active,omitempty"`
	// 超级管理��?	IsSuperuser bool `json:"is_superuser"`
	// 头像
	Avatar string `json:"avatar,omitempty"`
	// 是否开启二次验��?	IsOtp bool `json:"is_otp,omitempty"`
	// 权限
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	// 个性化设置
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// UserCreate Properties to receive via API on creation
type UserCreate struct {
	// 用户��?	Name string `json:"name"`
	// 邮箱，未启用
	Email string `json:"email,omitempty"`
	// 密码
	Password string `json:"password,omitempty"`
	// 权限
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	// 个性化设置
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// UserUpdate Properties to receive via API on update
type UserUpdate struct {
	// ID
	ID int `json:"id"`
	// 用户��?	Name string `json:"name"`
	// 邮箱，未启用
	Email string `json:"email,omitempty"`
	// 密码
	Password string `json:"password,omitempty"`
	// 权限
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	// 个性化设置
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// UserInDBBase Base properties for DB models
type UserInDBBase struct {
	// ID
	ID int `json:"id,omitempty"`
	// 用户��?	Name string `json:"name"`
	// 邮箱，未启用
	Email string `json:"email,omitempty"`
	// 状��?	IsActive bool `json:"is_active,omitempty"`
	// 超级管理��?	IsSuperuser bool `json:"is_superuser"`
	// 头像
	Avatar string `json:"avatar,omitempty"`
	// 是否开启二次验��?	IsOtp bool `json:"is_otp,omitempty"`
	// 权限
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	// 个性化设置
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// User Additional properties to return via API
type User struct {
	// ID
	ID int `json:"id,omitempty"`
	// 用户��?	Name string `json:"name"`
	// 邮箱，未启用
	Email string `json:"email,omitempty"`
	// 状��?	IsActive bool `json:"is_active,omitempty"`
	// 超级管理��?	IsSuperuser bool `json:"is_superuser"`
	// 头像
	Avatar string `json:"avatar,omitempty"`
	// 是否开启二次验��?	IsOtp bool `json:"is_otp,omitempty"`
	// 权限
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	// 个性化设置
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// UserInDB Additional properties stored in DB
type UserInDB struct {
	// ID
	ID int `json:"id,omitempty"`
	// 用户��?	Name string `json:"name"`
	// 邮箱，未启用
	Email string `json:"email,omitempty"`
	// 状��?	IsActive bool `json:"is_active,omitempty"`
	// 超级管理��?	IsSuperuser bool `json:"is_superuser"`
	// 头像
	Avatar string `json:"avatar,omitempty"`
	// 是否开启二次验��?	IsOtp bool `json:"is_otp,omitempty"`
	// 权限
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	// 个性化设置
	Settings map[string]interface{} `json:"settings,omitempty"`
	// 哈希密码
	HashedPassword string `json:"hashed_password"`
}

// NewUserBase 创建一个新��?UserBase 实例
func NewUserBase() *UserBase {
	return &UserBase{
		IsActive:    true,
		IsSuperuser: false,
		IsOtp:       false,
		Permissions: make(map[string]interface{}),
		Settings:    make(map[string]interface{}),
	}
}

// NewUserCreate 创建一个新��?UserCreate 实例
func NewUserCreate() *UserCreate {
	return &UserCreate{
		Permissions: make(map[string]interface{}),
		Settings:    make(map[string]interface{}),
	}
}

// NewUserUpdate 创建一个新��?UserUpdate 实例
func NewUserUpdate() *UserUpdate {
	return &UserUpdate{
		Permissions: make(map[string]interface{}),
		Settings:    make(map[string]interface{}),
	}
}

// NewUserInDBBase 创建一个新��?UserInDBBase 实例
func NewUserInDBBase() *UserInDBBase {
	return &UserInDBBase{
		IsActive:    true,
		IsSuperuser: false,
		IsOtp:       false,
		Permissions: make(map[string]interface{}),
		Settings:    make(map[string]interface{}),
	}
}

// NewUser 创建一个新��?User 实例
func NewUser() *User {
	return &User{
		IsActive:    true,
		IsSuperuser: false,
		IsOtp:       false,
		Permissions: make(map[string]interface{}),
		Settings:    make(map[string]interface{}),
	}
}

// NewUserInDB 创建一个新��?UserInDB 实例
func NewUserInDB() *UserInDB {
	return &UserInDB{
		IsActive:       true,
		IsSuperuser:    false,
		IsOtp:          false,
		Permissions:    make(map[string]interface{}),
		Settings:       make(map[string]interface{}),
		HashedPassword: "",
	}
}
