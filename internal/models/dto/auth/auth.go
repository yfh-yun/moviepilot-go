package auth

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" form:"username" binding:"required"`
	Password    string `json:"password" form:"password" binding:"required"`
	OtpPassword string `json:"otp_password" form:"otp_password"` // MFA验证码
}

// TokenResponse Token响应
type TokenResponse struct {
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	SuperUser   bool           `json:"super_user"`
	UserID      uint           `json:"user_id"`
	UserName    string         `json:"user_name"`
	Avatar      string         `json:"avatar"`
	Level       int            `json:"level"`
	Permissions map[string]any `json:"permissions"`
	Wizard      bool           `json:"wizard"` // 是否显示配置向导
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TestTokenResponse 测试Token响应
type TestTokenResponse struct {
	Valid   bool   `json:"valid"`
	UserID  uint   `json:"user_id"`
	Message string `json:"message"`
}
