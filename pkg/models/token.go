package models

// Token 令牌信息模型
type Token struct {
	// 令牌
	AccessToken string `json:"access_token"`
	// 令牌类型
	TokenType string `json:"token_type"`
	// 超级用户
	SuperUser bool `json:"super_user"`
	// 用户ID
	UserID int `json:"user_id"`
	// 用户�?	UserName string `json:"user_name"`
	// 头像
	Avatar string `json:"avatar,omitempty"`
	// 权限级别
	Level int `json:"level"`
	// 详细权限
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	// 是否显示配置向导
	Widzard bool `json:"widzard,omitempty"`
}

// TokenPayload 令牌载荷模型
type TokenPayload struct {
	// 用户ID
	Sub int `json:"sub,omitempty"`
	// 用户�?	Username string `json:"username,omitempty"`
	// 超级用户
	SuperUser bool `json:"super_user,omitempty"`
	// 权限级别
	Level int `json:"level,omitempty"`
	// 令牌用�?authentication\resource
	Purpose string `json:"purpose,omitempty"`
}

// NewToken 创建一个新�?Token 实例
func NewToken() *Token {
	return &Token{
		Level:       1,
		Permissions: make(map[string]interface{}),
	}
}

// NewTokenPayload 创建一个新�?TokenPayload 实例
func NewTokenPayload() *TokenPayload {
	return &TokenPayload{}
}
