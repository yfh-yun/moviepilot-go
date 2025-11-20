package security

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// User 用户模型
type User struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"` // 不输出密码哈希
	Nickname      string    `json:"nickname"`
	Avatar        string    `json:"avatar"`
	RoleIDs       []string  `json:"role_ids"`
	LastLoginAt   time.Time `json:"last_login_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Status        string    `json:"status"` // active, locked, disabled
	FailedAttempt int       `json:"failed_attempt"`
	LastFailedAt  time.Time `json:"last_failed_at"`
	MFAEnabled    bool      `json:"mfa_enabled"`
	MFAKey        string    `json:"-"` // 不输出MFA密钥
}

// Role 角色模型
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission 权限模型
type Permission struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserManager 用户管理器接口
type UserManager interface {
	CreateUser(user *User) error
	GetUserByID(id string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error
	ListUsers(page, pageSize int) ([]*User, int64, error)
	Authenticate(username, password string) (*User, error)
	UpdatePassword(userID, newPassword string) error
	ResetPassword(userID string) (string, error)
	LockUser(userID string) error
	UnlockUser(userID string) error
	UpdateLoginStatus(userID string, success bool) error
	SearchUsers(keyword string, page, pageSize int) ([]*User, int64, error)
}

// RoleManager 角色管理器接口
type RoleManager interface {
	CreateRole(role *Role) error
	GetRoleByID(id string) (*Role, error)
	GetRoleByName(name string) (*Role, error)
	UpdateRole(role *Role) error
	DeleteRole(id string) error
	ListRoles() ([]*Role, error)
	AddPermissionToRole(roleID, permissionCode string) error
	RemovePermissionFromRole(roleID, permissionCode string) error
	GetUserRoles(userID string) ([]*Role, error)
	AssignRoleToUser(userID, roleID string) error
	RemoveRoleFromUser(userID, roleID string) error
}

// PermissionManager 权限管理器接口
type PermissionManager interface {
	CreatePermission(perm *Permission) error
	GetPermissionByID(id string) (*Permission, error)
	GetPermissionByCode(code string) (*Permission, error)
	UpdatePermission(perm *Permission) error
	DeletePermission(id string) error
	ListPermissions() ([]*Permission, error)
	ListPermissionsByCategory(category string) ([]*Permission, error)
	CheckPermission(userID, permissionCode string) (bool, error)
	CheckUserHasPermission(user *User, permissionCode string) bool
}

// UserSession 用户会话
type UserSession struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	Token        string    `json:"token"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsActive     bool      `json:"is_active"`
}

// UserSessionManager 用户会话管理器接口
type UserSessionManager interface {
	CreateSession(userID, ip, userAgent string) (*UserSession, error)
	GetSession(sessionID string) (*UserSession, error)
	UpdateSessionActivity(sessionID string) error
	InvalidateSession(sessionID string) error
	InvalidateUserSessions(userID string) error
	ListUserSessions(userID string) ([]*UserSession, error)
	GetActiveSessionsCount() int
}

// MemoryUserManager 内存用户管理器实现
type MemoryUserManager struct {
	users       map[string]*User
	usernameMap map[string]*User
	emailMap    map[string]*User
	mutex       sync.RWMutex
	passwordManager *PasswordManager
	config      *SecurityConfig
	logger      logger.Logger
}

// NewMemoryUserManager 创建内存用户管理器
func NewMemoryUserManager(config *SecurityConfig, logger logger.Logger) *MemoryUserManager {
	return &MemoryUserManager{
		users:           make(map[string]*User),
		usernameMap:     make(map[string]*User),
		emailMap:        make(map[string]*User),
		passwordManager: NewPasswordManager(config),
		config:          config,
		logger:          logger,
	}
}

// CreateUser 创建用户
func (um *MemoryUserManager) CreateUser(user *User) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	// 检查用户名是否已存在
	if _, exists := um.usernameMap[user.Username]; exists {
		return errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if user.Email != "" {
		if _, exists := um.emailMap[user.Email]; exists {
			return errors.New("email already exists")
		}
	}

	// 设置默认值
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()
	if user.Status == "" {
		user.Status = "active"
	}

	// 存储用户
	um.users[user.ID] = user
	um.usernameMap[user.Username] = user
	if user.Email != "" {
		um.emailMap[user.Email] = user
	}

	um.logger.Info("User created successfully", "user_id", user.ID, "username", user.Username)
	return nil
}

// GetUserByID 根据ID获取用户
func (um *MemoryUserManager) GetUserByID(id string) (*User, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	user, exists := um.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	// 返回用户副本
	return user.clone(), nil
}

// GetUserByUsername 根据用户名获取用户
func (um *MemoryUserManager) GetUserByUsername(username string) (*User, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	user, exists := um.usernameMap[strings.ToLower(username)]
	if !exists {
		return nil, errors.New("user not found")
	}

	// 返回用户副本
	return user.clone(), nil
}

// GetUserByEmail 根据邮箱获取用户
func (um *MemoryUserManager) GetUserByEmail(email string) (*User, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	user, exists := um.emailMap[strings.ToLower(email)]
	if !exists {
		return nil, errors.New("user not found")
	}

	// 返回用户副本
	return user.clone(), nil
}

// UpdateUser 更新用户
func (um *MemoryUserManager) UpdateUser(user *User) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	originalUser, exists := um.users[user.ID]
	if !exists {
		return errors.New("user not found")
	}

	// 检查用户名是否被其他用户使用
	if user.Username != originalUser.Username {
		if _, exists := um.usernameMap[strings.ToLower(user.Username)]; exists {
			return errors.New("username already exists")
		}
		delete(um.usernameMap, strings.ToLower(originalUser.Username))
		um.usernameMap[strings.ToLower(user.Username)] = user
	}

	// 检查邮箱是否被其他用户使用
	if user.Email != originalUser.Email {
		if originalUser.Email != "" {
			delete(um.emailMap, strings.ToLower(originalUser.Email))
		}
		if user.Email != "" {
			if _, exists := um.emailMap[strings.ToLower(user.Email)]; exists {
				return errors.New("email already exists")
			}
			um.emailMap[strings.ToLower(user.Email)] = user
		}
	}

	// 更新时间
	user.UpdatedAt = time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = originalUser.CreatedAt
	}

	// 更新用户
	um.users[user.ID] = user

	um.logger.Info("User updated successfully", "user_id", user.ID, "username", user.Username)
	return nil
}

// DeleteUser 删除用户
func (um *MemoryUserManager) DeleteUser(id string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[id]
	if !exists {
		return errors.New("user not found")
	}

	// 删除用户引用
	delete(um.users, id)
	delete(um.usernameMap, strings.ToLower(user.Username))
	if user.Email != "" {
		delete(um.emailMap, strings.ToLower(user.Email))
	}

	um.logger.Info("User deleted successfully", "user_id", id, "username", user.Username)
	return nil
}

// ListUsers 列出用户
func (um *MemoryUserManager) ListUsers(page, pageSize int) ([]*User, int64, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	// 转换为切片
	userList := make([]*User, 0, len(um.users))
	for _, user := range um.users {
		userList = append(userList, user.clone())
	}

	total := int64(len(userList))

	// 分页处理
	start := (page - 1) * pageSize
	end := start + pageSize

	if start < 0 {
		start = 0
	}
	if end > len(userList) {
		end = len(userList)
	}

	if start >= len(userList) {
		return []*User{}, total, nil
	}

	return userList[start:end], total, nil
}

// Authenticate 用户认证
func (um *MemoryUserManager) Authenticate(username, password string) (*User, error) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.usernameMap[strings.ToLower(username)]
	if !exists {
		return nil, errors.New("invalid username or password")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, fmt.Errorf("user account is %s", user.Status)
	}

	// 检查是否被锁定
	if um.config.MaxFailedAttempts > 0 && user.FailedAttempt >= um.config.MaxFailedAttempts {
		// 检查锁定时间是否已过
		lockDuration := time.Duration(um.config.AccountLockDuration) * time.Minute
		if time.Since(user.LastFailedAt) < lockDuration {
			remainingTime := lockDuration - time.Since(user.LastFailedAt)
			return nil, fmt.Errorf("account locked, try again in %.0f minutes", remainingTime.Minutes())
		}
		// 重置失败尝试次数
		user.FailedAttempt = 0
	}

	// 验证密码
	matches, err := um.passwordManager.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		um.logger.Error("Failed to verify password", "user_id", user.ID, "error", err.Error())
		return nil, errors.New("authentication failed")
	}

	if !matches {
		// 增加失败尝试次数
		user.FailedAttempt++
		user.LastFailedAt = time.Now()
		um.logger.Warn("Failed login attempt", "username", username, "attempt", user.FailedAttempt)
		return nil, errors.New("invalid username or password")
	}

	// 认证成功，重置失败尝试次数
	user.FailedAttempt = 0
	user.LastLoginAt = time.Now()

	um.logger.Info("User authenticated successfully", "user_id", user.ID, "username", user.Username)
	return user.clone(), nil
}

// UpdatePassword 更新密码
func (um *MemoryUserManager) UpdatePassword(userID, newPassword string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// 生成新的密码哈希
	hash, err := um.passwordManager.GeneratePasswordHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now()

	um.logger.Info("User password updated successfully", "user_id", userID)
	return nil
}

// ResetPassword 重置密码
func (um *MemoryUserManager) ResetPassword(userID string) (string, error) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return "", errors.New("user not found")
	}

	// 生成随机密码
	newPassword := GenerateRandomPassword(um.config.DefaultPasswordLength)

	// 生成密码哈希
	hash, err := um.passwordManager.GeneratePasswordHash(newPassword)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %w", err)
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now()

	um.logger.Info("User password reset successfully", "user_id", userID)
	return newPassword, nil
}

// LockUser 锁定用户
func (um *MemoryUserManager) LockUser(userID string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	user.Status = "locked"
	user.UpdatedAt = time.Now()

	um.logger.Info("User locked", "user_id", userID)
	return nil
}

// UnlockUser 解锁用户
func (um *MemoryUserManager) UnlockUser(userID string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	user.Status = "active"
	user.FailedAttempt = 0
	user.UpdatedAt = time.Now()

	um.logger.Info("User unlocked", "user_id", userID)
	return nil
}

// UpdateLoginStatus 更新登录状态
func (um *MemoryUserManager) UpdateLoginStatus(userID string, success bool) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	if success {
		user.LastLoginAt = time.Now()
		user.FailedAttempt = 0
		um.logger.Info("User login successful", "user_id", userID)
	} else {
		user.FailedAttempt++
		user.LastFailedAt = time.Now()
		um.logger.Warn("User login failed", "user_id", userID, "attempt", user.FailedAttempt)

		// 检查是否需要锁定账户
		if um.config.MaxFailedAttempts > 0 && user.FailedAttempt >= um.config.MaxFailedAttempts {
			user.Status = "locked"
			um.logger.Warn("User account locked due to too many failed attempts", "user_id", userID)
		}
	}

	user.UpdatedAt = time.Now()
	return nil
}

// SearchUsers 搜索用户
func (um *MemoryUserManager) SearchUsers(keyword string, page, pageSize int) ([]*User, int64, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	keyword = strings.ToLower(keyword)
	var filteredUsers []*User

	for _, user := range um.users {
		if strings.Contains(strings.ToLower(user.Username), keyword) ||
			strings.Contains(strings.ToLower(user.Email), keyword) ||
			strings.Contains(strings.ToLower(user.Nickname), keyword) {
			filteredUsers = append(filteredUsers, user.clone())
		}
	}

	total := int64(len(filteredUsers))

	// 分页处理
	start := (page - 1) * pageSize
	end := start + pageSize

	if start < 0 {
		start = 0
	}
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}

	if start >= len(filteredUsers) {
		return []*User{}, total, nil
	}

	return filteredUsers[start:end], total, nil
}

// clone 克隆用户
func (u *User) clone() *User {
	clone := *u
	// 深拷贝切片
	clone.RoleIDs = make([]string, len(u.RoleIDs))
	copy(clone.RoleIDs, u.RoleIDs)
	return &clone
}

// ValidatePasswordPolicy 验证密码策略
func ValidatePasswordPolicy(password string, config *SecurityConfig) error {
	if len(password) < config.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", config.MinPasswordLength)
	}

	if len(password) > config.MaxPasswordLength {
		return fmt.Errorf("password must not exceed %d characters", config.MaxPasswordLength)
	}

	// 检查密码强度
	isStrong, details := IsPasswordStrong(password)
	if config.RequireStrongPassword && !isStrong {
		var issues []string
		if !details["min_length"] {
			issues = append(issues, "minimum 8 characters")
		}
		if !details["uppercase"] {
			issues = append(issues, "at least one uppercase letter")
		}
		if !details["lowercase"] {
			issues = append(issues, "at least one lowercase letter")
		}
		if !details["digit"] {
			issues = append(issues, "at least one digit")
		}
		if !details["special"] {
			issues = append(issues, "at least one special character")
		}
		return fmt.Errorf("password must meet the following requirements: %s", strings.Join(issues, ", "))
	}

	return nil
}