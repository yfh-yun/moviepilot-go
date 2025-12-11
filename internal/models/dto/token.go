package dto

// Token 令牌信息
type Token struct {
	// 令牌
	AccessToken string `json:"access_token"`
	// 令牌类型
	TokenType string `json:"token_type"`
	// 超级用户
	SuperUser bool `json:"super_user"`
	// 用户ID
	UserID int `json:"user_id"`
	// 用户名
	UserName string `json:"user_name"`
	// 头像
	Avatar string `json:"avatar,omitempty"`
	// 权限级别
	Level int `json:"level,omitempty"`
	// 详细权限
	Permissions map[string]any `json:"permissions,omitempty"`
	// 是否显示配置向导
	Widzard *bool `json:"widzard,omitempty"`
	// 首次登录时返回的随机初始密码（仅用于第一次创建管理员时）
	InitialPassword string `json:"initial_password,omitempty"`
}

// TokenPayload 令牌载荷
type TokenPayload struct {
	// 用户ID
	Sub *int `json:"sub,omitempty"`
	// 用户名
	Username string `json:"username,omitempty"`
	// 超级用户
	SuperUser *bool `json:"super_user,omitempty"`
	// 权限级别
	Level *int `json:"level,omitempty"`
	// 令牌用途 authentication\resource
	Purpose string `json:"purpose,omitempty"`
}
