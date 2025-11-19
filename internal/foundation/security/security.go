package security

import (
	"context"
	"errors"
	"time"
)

// 定义常见错误
var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidToken          = errors.New("invalid token")
	ErrExpiredToken          = errors.New("token has expired")
	ErrInsufficientPermission = errors.New("insufficient permission")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserLocked            = errors.New("user account is locked")
	ErrUserDisabled          = errors.New("user account is disabled")
	ErrAuthenticationFailed  = errors.New("authentication failed")
	ErrAuthorizationFailed   = errors.New("authorization failed")
	ErrPasswordPolicy        = errors.New("password does not meet policy requirements")
	ErrSessionExpired        = errors.New("session has expired")
)

// SecurityManager 安全管理器接口
type SecurityManager interface {
	// 初始化
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error

	// 认证相关
	Authenticate(ctx context.Context, username, password string) (*AuthResult, error)
	AuthenticateWithToken(ctx context.Context, token string) (*AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	RevokeToken(ctx context.Context, token string) error

	// 用户管理
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUsers(ctx context.Context, page, pageSize int) ([]*User, int, error)

	// 角色管理
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, roleID string) error
	GetRoleByID(ctx context.Context, roleID string) (*Role, error)
	GetRoleByName(ctx context.Context, roleName string) (*Role, error)
	GetRoles(ctx context.Context) ([]*Role, error)

	// 用户-角色关联
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*Role, error)
	GetRoleUsers(ctx context.Context, roleID string) ([]*User, error)

	// 权限管理
	CheckPermission(ctx context.Context, userID, resource, action string) (bool, error)
	GrantPermission(ctx context.Context, roleID, resource, action string) error
	RevokePermission(ctx context.Context, roleID, resource, action string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*Permission, error)
	GetUserPermissions(ctx context.Context, userID string) ([]*Permission, error)

	// 密码管理
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, userID, newPassword string) error
	GeneratePasswordHash(password string) (string, error)
	VerifyPassword(password, hash string) (bool, error)

	// JWT相关
	GenerateToken(user *User) (*TokenPair, error)
	ValidateToken(token string) (*Claims, error)

	// 配置相关
	SetConfig(config *SecurityConfig)
	GetConfig() *SecurityConfig

	// 统计和监控
	GetStatistics(ctx context.Context) map[string]interface{}
	GetRecentLoginAttempts(ctx context.Context, limit int) ([]*LoginAttempt, error)
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// JWT配置
	JWTSecretKey      string        `json:"jwt_secret_key"`
	JWTExpiration     time.Duration `json:"jwt_expiration"`
	JWTRefreshExpiration time.Duration `json:"jwt_refresh_expiration"`
	JWTAlgorithm      string        `json:"jwt_algorithm"`

	// 密码策略
	PasswordMinLength     int      `json:"password_min_length"`
	PasswordRequireUppercase bool   `json:"password_require_uppercase"`
	PasswordRequireLowercase bool   `json:"password_require_lowercase"`
	PasswordRequireDigit    bool   `json:"password_require_digit"`
	PasswordRequireSpecial  bool   `json:"password_require_special"`
	PasswordHistorySize     int    `json:"password_history_size"`
	PasswordMaxAge          time.Duration `json:"password_max_age"`

	// 认证策略
	MaxLoginAttempts     int           `json:"max_login_attempts"`
	AccountLockDuration  time.Duration `json:"account_lock_duration"`
	SessionTimeout       time.Duration `json:"session_timeout"`
	TokenBlacklistTTL    time.Duration `json:"token_blacklist_ttl"`

	// 安全选项
	EnableRBAC           bool   `json:"enable_rbac"`
	EnableAuditLogging   bool   `json:"enable_audit_logging"`
	EnableTwoFactorAuth  bool   `json:"enable_two_factor_auth"`
	AllowedOrigins       []string `json:"allowed_origins"`

	// 加密配置
	BCryptCost          int    `json:"bcrypt_cost"`
}

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		// JWT配置
		JWTSecretKey:         "your-secret-key-here", // 生产环境中应通过环境变量设置
		JWTExpiration:        24 * time.Hour,
		JWTRefreshExpiration: 7 * 24 * time.Hour,
		JWTAlgorithm:         "HS256",

		// 密码策略
		PasswordMinLength:        8,
		PasswordRequireUppercase: true,
		PasswordRequireLowercase: true,
		PasswordRequireDigit:     true,
		PasswordRequireSpecial:   true,
		PasswordHistorySize:      5,
		PasswordMaxAge:           90 * 24 * time.Hour,

		// 认证策略
		MaxLoginAttempts:    5,
		AccountLockDuration: 30 * time.Minute,
		SessionTimeout:      24 * time.Hour,
		TokenBlacklistTTL:   7 * 24 * time.Hour,

		// 安全选项
		EnableRBAC:          true,
		EnableAuditLogging:  true,
		EnableTwoFactorAuth: false,
		AllowedOrigins:      []string{"*"},

		// 加密配置
		BCryptCost: 12,
	}
}

// User 用户模型
type User struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	PasswordHash    string    `json:"-"` // 不返回密码哈希
	Email           string    `json:"email"`
	FullName        string    `json:"full_name"`
	Status          string    `json:"status"` // active, locked, disabled
	LastLogin       time.Time `json:"last_login"`
	FailedAttempts  int       `json:"failed_attempts"`
	LockedUntil     time.Time `json:"locked_until"`
	PasswordChanged time.Time `json:"password_changed"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	RoleIDs         []string  `json:"role_ids"`
	Attributes      map[string]interface{} `json:"attributes"`
}

// Role 角色模型
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Permissions []*Permission `json:"permissions"`
}

// Permission 权限模型
type Permission struct {
	ID          string    `json:"id"`
	RoleID      string    `json:"role_id"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuthResult 认证结果
type AuthResult struct {
	User        *User      `json:"user"`
	AccessToken string     `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn   int64      `json:"expires_in"`
	TokenType   string     `json:"token_type"`
	Permissions []*Permission `json:"permissions"`
	LoginTime   time.Time  `json:"login_time"`
	IPAddress   string     `json:"ip_address"`
	UserAgent   string     `json:"user_agent"`
}

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType    string    `json:"token_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// Claims JWT声明
type Claims struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	RoleIDs   []string  `json:"role_ids"`
	Exp       int64     `json:"exp"`
	Iat       int64     `json:"iat"`
	Iss       string    `json:"iss"`
	Sub       string    `json:"sub"`
	Jti       string    `json:"jti"` // JWT ID
	Type      string    `json:"type"` // access or refresh
}

// LoginAttempt 登录尝试记录
type LoginAttempt struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Success     bool      `json:"success"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	AttemptTime time.Time `json:"attempt_time"`
	Error       string    `json:"error,omitempty"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Timestamp   time.Time              `json:"timestamp"`
	Success     bool                   `json:"success"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// SecurityContext 安全上下文
type SecurityContext struct {
	Authenticated bool
	User          *User
	Token         string
	Claims        *Claims
	IPAddress     string
	UserAgent     string
	Permissions   []string // 缓存的权限列表，格式为 "resource:action"
}

// HasPermission 检查是否有指定权限
func (sc *SecurityContext) HasPermission(resource, action string) bool {
	if !sc.Authenticated {
		return false
	}

	permissionKey := resource + ":" + action
	for _, perm := range sc.Permissions {
		if perm == permissionKey {
			return true
		}
	}

	return false
}

// IsAdmin 检查用户是否有管理员权限
func (sc *SecurityContext) IsAdmin() bool {
	return sc.HasPermission("admin", "access")
}

// IsAuthenticated 检查是否已认证
func (sc *SecurityContext) IsAuthenticated() bool {
	return sc.Authenticated
}

// NewSecurityContext 创建新的安全上下文
func NewSecurityContext() *SecurityContext {
	return &SecurityContext{
		Authenticated: false,
		Permissions:   make([]string, 0),
		Attributes:    make(map[string]interface{}),
	}
}

// 安全工具函数

// ValidatePasswordPolicy 验证密码是否符合策略
func ValidatePasswordPolicy(password string, config *SecurityConfig) error {
	if len(password) < config.PasswordMinLength {
		return ErrPasswordPolicy
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case char >= 33 && char <= 47 || char >= 58 && char <= 64 || char >= 91 && char <= 96 || char >= 123 && char <= 126:
			hasSpecial = true
		}
	}

	if config.PasswordRequireUppercase && !hasUpper {
		return ErrPasswordPolicy
	}

	if config.PasswordRequireLowercase && !hasLower {
		return ErrPasswordPolicy
	}

	if config.PasswordRequireDigit && !hasDigit {
		return ErrPasswordPolicy
	}

	if config.PasswordRequireSpecial && !hasSpecial {
		return ErrPasswordPolicy
	}

	return nil
}

// IsPasswordExpired 检查密码是否过期
func IsPasswordExpired(lastChanged time.Time, maxAge time.Duration) bool {
	if maxAge <= 0 {
		return false
	}

	return time.Since(lastChanged) > maxAge
}

// IsAccountLocked 检查账户是否被锁定
func IsAccountLocked(lockedUntil time.Time) bool {
	return !lockedUntil.IsZero() && time.Now().Before(lockedUntil)
}

// IsAccountActive 检查账户是否激活
func IsAccountActive(user *User) bool {
	return user.Status == "active" && !IsAccountLocked(user.LockedUntil)
}

// 常量定义
const (
	// 用户状态
	UserStatusActive   = "active"
	UserStatusLocked   = "locked"
	UserStatusDisabled = "disabled"

	// 令牌类型
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	// 资源类型
	ResourceUser    = "user"
	ResourceRole    = "role"
	ResourceConfig  = "config"
	ResourceMedia   = "media"
	ResourceTask    = "task"
	ResourcePlugin  = "plugin"
	ResourceAdmin   = "admin"

	// 操作类型
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"
	ActionAccess = "access"
	ActionExecute = "execute"
	ActionManage = "manage"

	// 系统角色
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleGuest     = "guest"
	RoleSystem    = "system"
)

// GenerateRandomString 生成随机字符串（用于密码重置、临时令牌等）
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	s := make([]byte, length)
	for i := range s {
		s[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond) // 确保每次生成不同的随机数
	}
	return string(s)
}