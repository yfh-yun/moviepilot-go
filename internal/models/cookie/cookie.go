package schemas

// CookieData Cookie数据结构
type CookieData struct {
	UUID      string `json:"uuid" binding:"required"`
	Encrypted string `json:"encrypted" binding:"required"`
}

// CookiePassword Cookie密码结构
type CookiePassword struct {
	Password string `json:"password" binding:"required"`
}

// CookieResponse Cookie响应结构
type CookieResponse struct {
	Action string `json:"action"`
	Data   any    `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
}

// CookieInfo Cookie信息结构
type CookieInfo struct {
	UUID          string `json:"uuid"`
	Exists        bool   `json:"exists"`
	FileSize      int64  `json:"file_size"`
	LastModified  string `json:"last_modified"`
	HasEncrypted  bool   `json:"has_encrypted"`
	EncryptedSize int    `json:"encrypted_size"`
}

// CookieListResponse Cookie列表响应结构
type CookieListResponse struct {
	Cookies []CookieInfo `json:"cookies"`
	Total   int          `json:"total"`
}

// CookieCloudData CookieCloud数据结构
type CookieCloudData struct {
	CookieData map[string]any `json:"cookie_data"`
	DeviceInfo map[string]any `json:"device_info,omitempty"`
}
