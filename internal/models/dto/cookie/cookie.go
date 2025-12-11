package cookie

// CookieData Cookie数据
type CookieData struct {
	// 加密数据
	Encrypted string `json:"encrypted" validate:"required,min=1,max=52428800"` // 50MB
	// UUID
	UUID string `json:"uuid" validate:"required,min=5,alphanum"`
}

// CookiePassword Cookie密码
type CookiePassword struct {
	// 密码
	Password string `json:"password" validate:"required"`
}
