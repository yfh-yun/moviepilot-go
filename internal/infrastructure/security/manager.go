package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"moviepilot-go/pkg/logger"
)

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		SecretKey:           generateRandomSecret(),
		JWTExpirationHours:  24,
		SessionTimeout:      30,
		MaxFailedAttempts:   5,
		AccountLockDuration: 30,
		MinPasswordLength:   8,
		MaxPasswordLength:   128,
		DefaultPasswordLength: 12,
		RequireStrongPassword: true,
		MaxSessionsPerUser:  5,
		BCryptCost:          12,
	}
}

// generateRandomSecret 生成随机密钥
func generateRandomSecret() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机生成失败，使用时间戳作为回退
		return fmt.Sprintf("fallback-secret-%d", time.Now().UnixNano())
	}
	return base64.StdEncoding.EncodeToString(b)
}

// SecurityManagerImpl SecurityManager接口的实现
type SecurityManagerImpl struct {
	config             *SecurityConfig
	userManager        *MemoryUserManager
	roleManager        *MemoryRoleManager
	permissionManager  *MemoryPermissionManager
	sessionManager     *MemoryUserSessionManager
	jwtManager         *JWTManager
	passwordManager    *PasswordManager
	logger             logger.Logger
	initialized        bool
	mutex              sync.RWMutex
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager(config *SecurityConfig, logger logger.Logger) *SecurityManagerImpl {
	// 如果配置为空，使用默认配置
	if config == nil {
		config = DefaultSecurityConfig()
	}

	manager := &SecurityManagerImpl{
		config:    config,
		logger:    logger,
		initialized: false,
	}

	// 初始化各个管理器
	manager.jwtManager = NewJWTManager(config, logger)
	manager.passwordManager = NewPasswordManager(config)
	manager.roleManager = NewMemoryRoleManager(logger)
	manager.permissionManager = NewMemoryPermissionManager(manager.roleManager, logger)
	manager.userManager = NewMemoryUserManager(config, logger)
	manager.sessionManager = NewMemoryUserSessionManager(config, manager.jwtManager, logger)

	return manager
}

// Initialize 初始化安全管理器
func (sm *SecurityManagerImpl) Initialize() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.initialized {
		return nil
	}

	// 创建默认权限
	sm.createDefaultPermissions()

	// 创建默认角色
	sm.createDefaultRoles()

	// 创建默认管理员用户（如果不存在）
	sm.createDefaultAdminUser()

	// 启动会话清理任务
	sm.sessionManager.StartCleanupTask()

	sm.initialized = true
	sm.logger.Info("Security manager initialized successfully")
	return nil
}

// GetUserManager 获取用户管理器
func (sm *SecurityManagerImpl) GetUserManager() UserManager {
	return sm.userManager
}

// GetRoleManager 获取角色管理器
func (sm *SecurityManagerImpl) GetRoleManager() RoleManager {
	return sm.roleManager
}

// GetPermissionManager 获取权限管理器
func (sm *SecurityManagerImpl) GetPermissionManager() PermissionManager {
	return sm.permissionManager
}

// GetSessionManager 获取会话管理器
func (sm *SecurityManagerImpl) GetSessionManager() UserSessionManager {
	return sm.sessionManager
}

// GetJWTManager 获取JWT管理器
func (sm *SecurityManagerImpl) GetJWTManager() *JWTManager {
	return sm.jwtManager
}

// GetPasswordManager 获取密码管理器
func (sm *SecurityManagerImpl) GetPasswordManager() *PasswordManager {
	return sm.passwordManager
}

// Authenticate 用户认证
func (sm *SecurityManagerImpl) Authenticate(username, password string) (*User, string, error) {
	// 验证用户
	user, err := sm.userManager.Authenticate(username, password)
	if err != nil {
		return nil, "", err
	}

	// 创建会话
	session, err := sm.sessionManager.CreateSession(user.ID, "", "")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, session.Token, nil
}

// VerifyToken 验证令牌
func (sm *SecurityManagerImpl) VerifyToken(token string) (*User, error) {
	// 验证令牌
	claims, err := sm.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// 获取用户
	user, err := sm.userManager.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, fmt.Errorf("user account is %s", user.Status)
	}

	return user, nil
}

// RefreshToken 刷新令牌
func (sm *SecurityManagerImpl) RefreshToken(token string) (string, error) {
	// 验证当前令牌
	user, err := sm.VerifyToken(token)
	if err != nil {
		return "", err
	}

	// 生成新令牌
	newToken, err := sm.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}

	return newToken, nil
}

// Logout 用户登出
func (sm *SecurityManagerImpl) Logout(token string) error {
	// 查找并使会话失效
	session, err := sm.sessionManager.VerifySessionToken(token)
	if err != nil {
		return err
	}

	return sm.sessionManager.InvalidateSession(session.SessionID)
}

// CheckPermission 检查用户权限
func (sm *SecurityManagerImpl) CheckPermission(userID, permissionCode string) (bool, error) {
	return sm.permissionManager.CheckPermission(userID, permissionCode)
}

// CreateUser 创建用户
func (sm *SecurityManagerImpl) CreateUser(user *User, password string) error {
	// 生成密码哈希
	hash, err := sm.passwordManager.GeneratePasswordHash(password)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	user.PasswordHash = hash
	return sm.userManager.CreateUser(user)
}

// UpdatePassword 更新用户密码
func (sm *SecurityManagerImpl) UpdatePassword(userID, oldPassword, newPassword string) error {
	// 验证旧密码
	user, err := sm.userManager.GetUserByID(userID)
	if err != nil {
		return err
	}

	matches, err := sm.passwordManager.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil || !matches {
		return errors.New("invalid old password")
	}

	// 更新密码
	return sm.userManager.UpdatePassword(userID, newPassword)
}

// createDefaultPermissions 创建默认权限
func (sm *SecurityManagerImpl) createDefaultPermissions() {
	defaultPermissions := []*Permission{
		// 用户管理权限
		{ID: "perm_user_read", Code: "user:read", Name: "读取用户", Description: "读取用户信息的权限", Category: "用户管理"},
		{ID: "perm_user_create", Code: "user:create", Name: "创建用户", Description: "创建新用户的权限", Category: "用户管理"},
		{ID: "perm_user_update", Code: "user:update", Name: "更新用户", Description: "更新用户信息的权限", Category: "用户管理"},
		{ID: "perm_user_delete", Code: "user:delete", Name: "删除用户", Description: "删除用户的权限", Category: "用户管理"},
		{ID: "perm_user_manage", Code: "user:manage", Name: "管理用户", Description: "完全管理用户的权限", Category: "用户管理"},

		// 角色管理权限
		{ID: "perm_role_read", Code: "role:read", Name: "读取角色", Description: "读取角色信息的权限", Category: "角色管理"},
		{ID: "perm_role_create", Code: "role:create", Name: "创建角色", Description: "创建新角色的权限", Category: "角色管理"},
		{ID: "perm_role_update", Code: "role:update", Name: "更新角色", Description: "更新角色信息的权限", Category: "角色管理"},
		{ID: "perm_role_delete", Code: "role:delete", Name: "删除角色", Description: "删除角色的权限", Category: "角色管理"},
		{ID: "perm_role_manage", Code: "role:manage", Name: "管理角色", Description: "完全管理角色的权限", Category: "角色管理"},

		// 权限管理权限
		{ID: "perm_perm_read", Code: "permission:read", Name: "读取权限", Description: "读取权限信息的权限", Category: "权限管理"},
		{ID: "perm_perm_create", Code: "permission:create", Name: "创建权限", Description: "创建新权限的权限", Category: "权限管理"},
		{ID: "perm_perm_update", Code: "permission:update", Name: "更新权限", Description: "更新权限信息的权限", Category: "权限管理"},
		{ID: "perm_perm_delete", Code: "permission:delete", Name: "删除权限", Description: "删除权限的权限", Category: "权限管理"},
		{ID: "perm_perm_manage", Code: "permission:manage", Name: "管理权限", Description: "完全管理权限的权限", Category: "权限管理"},

		// 系统管理权限
		{ID: "perm_system_config", Code: "system:config", Name: "系统配置", Description: "修改系统配置的权限", Category: "系统管理"},
		{ID: "perm_system_log", Code: "system:log", Name: "查看日志", Description: "查看系统日志的权限", Category: "系统管理"},
		{ID: "perm_system_monitor", Code: "system:monitor", Name: "系统监控", Description: "监控系统状态的权限", Category: "系统管理"},
		{ID: "perm_system_manage", Code: "system:manage", Name: "系统管理", Description: "完全管理系统的权限", Category: "系统管理"},

		// 媒体管理权限
		{ID: "perm_media_read", Code: "media:read", Name: "读取媒体", Description: "读取媒体信息的权限", Category: "媒体管理"},
		{ID: "perm_media_manage", Code: "media:manage", Name: "管理媒体", Description: "完全管理媒体的权限", Category: "媒体管理"},

		// 下载管理权限
		{ID: "perm_download_read", Code: "download:read", Name: "读取下载", Description: "读取下载任务的权限", Category: "下载管理"},
		{ID: "perm_download_create", Code: "download:create", Name: "创建下载", Description: "创建下载任务的权限", Category: "下载管理"},
		{ID: "perm_download_manage", Code: "download:manage", Name: "管理下载", Description: "完全管理下载的权限", Category: "下载管理"},
	}

	for _, perm := range defaultPermissions {
		if err := sm.permissionManager.CreatePermission(perm); err != nil {
			sm.logger.Warn("Failed to create default permission", "code", perm.Code, "error", err.Error())
		} else {
			sm.logger.Debug("Default permission created", "code", perm.Code)
		}
	}
}

// createDefaultRoles 创建默认角色
func (sm *SecurityManagerImpl) createDefaultRoles() {
	// 创建管理员角色
	adminRole := &Role{
		ID:          "role_admin",
		Name:        "管理员",
		Description: "系统管理员，拥有所有权限",
		Permissions: []string{
			"user:manage", "role:manage", "permission:manage", "system:manage",
			"media:manage", "download:manage",
		},
	}

	if err := sm.roleManager.CreateRole(adminRole); err != nil {
		sm.logger.Warn("Failed to create admin role", "error", err.Error())
	} else {
		sm.logger.Debug("Admin role created")
	}

	// 创建普通用户角色
	userRole := &Role{
		ID:          "role_user",
		Name:        "普通用户",
		Description: "普通用户，拥有基本权限",
		Permissions: []string{
			"user:read", "media:read", "download:read", "download:create",
		},
	}

	if err := sm.roleManager.CreateRole(userRole); err != nil {
		sm.logger.Warn("Failed to create user role", "error", err.Error())
	} else {
		sm.logger.Debug("User role created")
	}
}

// createDefaultAdminUser 创建默认管理员用户
func (sm *SecurityManagerImpl) createDefaultAdminUser() {
	// 检查是否已存在管理员用户
	_, err := sm.userManager.GetUserByUsername("admin")
	if err == nil {
		sm.logger.Debug("Admin user already exists")
		return
	}

	// 创建默认管理员密码
	defaultPassword := "admin123" // 生产环境应该修改此默认密码
	if sm.config.RequireStrongPassword {
		// 如果需要强密码，生成随机密码
		defaultPassword = GenerateRandomPassword(sm.config.DefaultPasswordLength)
		sm.logger.Warn("Generated random password for default admin user", "password", defaultPassword)
	}

	// 创建管理员用户
	adminUser := &User{
		ID:           "admin",
		Username:     "admin",
		Email:        "admin@example.com",
		Nickname:     "管理员",
		RoleIDs:      []string{"role_admin"},
		Status:       "active",
		FailedAttempt: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 设置密码哈希
	hash, err := sm.passwordManager.GeneratePasswordHash(defaultPassword)
	if err != nil {
		sm.logger.Error("Failed to generate password hash for admin user", "error", err.Error())
		return
	}

	adminUser.PasswordHash = hash

	// 创建用户
	if err := sm.userManager.CreateUser(adminUser); err != nil {
		sm.logger.Error("Failed to create default admin user", "error", err.Error())
	} else {
		// 分配角色
		sm.roleManager.AssignRoleToUser(adminUser.ID, "role_admin")
		sm.logger.Info("Default admin user created", "username", adminUser.Username, "password", defaultPassword)
	}
}

// GetConfig 获取安全配置
func (sm *SecurityManagerImpl) GetConfig() *SecurityConfig {
	return sm.config
}

// UpdateConfig 更新安全配置
func (sm *SecurityManagerImpl) UpdateConfig(config *SecurityConfig) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 验证配置
	if config.SecretKey == "" {
		return errors.New("secret key cannot be empty")
	}

	if config.JWTExpirationHours <= 0 {
		return errors.New("JWT expiration hours must be positive")
	}

	if config.MinPasswordLength < 6 {
		return errors.New("minimum password length must be at least 6")
	}

	// 更新配置
	sm.config = config

	// 更新相关管理器的配置
	sm.jwtManager.UpdateConfig(config)
	sm.passwordManager = NewPasswordManager(config)
	sm.sessionManager = NewMemoryUserSessionManager(config, sm.jwtManager, sm.logger)

	sm.logger.Info("Security configuration updated")
	return nil
}

// IsInitialized 检查是否已初始化
func (sm *SecurityManagerImpl) IsInitialized() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.initialized
}

// GetSystemStatus 获取系统安全状态
func (sm *SecurityManagerImpl) GetSystemStatus() map[string]interface{} {
	stats := sm.sessionManager.GetSessionStats()

	// 获取用户和角色统计
	users, _, _ := sm.userManager.ListUsers(1, 1)
	userCount := 0
	if len(users) > 0 {
		_, total, _ := sm.userManager.ListUsers(1, 1000)
		userCount = int(total)
	}

	roles, _ := sm.roleManager.ListRoles()
	permissions, _ := sm.permissionManager.ListPermissions()

	return map[string]interface{}{
		"initialized":      sm.initialized,
		"user_count":       userCount,
		"role_count":       len(roles),
		"permission_count": len(permissions),
		"session_stats":    stats,
		"config": map[string]interface{}{
			"jwt_expiration_hours":   sm.config.JWTExpirationHours,
			"session_timeout":        sm.config.SessionTimeout,
			"max_failed_attempts":    sm.config.MaxFailedAttempts,
			"require_strong_password": sm.config.RequireStrongPassword,
		},
	}
}

// Shutdown 关闭安全管理器
func (sm *SecurityManagerImpl) Shutdown() error {
	sm.logger.Info("Security manager shutting down")
	// 可以在这里添加清理逻辑
	return nil
}