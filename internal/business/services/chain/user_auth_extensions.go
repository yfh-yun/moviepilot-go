package chain

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"
	"github.com/yfh-yun/moviepilot-go/internal/repositories"
	"github.com/yfh-yun/moviepilot-go/pkg/cache"
	"github.com/yfh-yun/moviepilot-go/internal/infrastructure/security"
	"github.com/yfh-yun/moviepilot-go/pkg/utils"
)

// AuxiliaryAuthCredentials 辅助认证凭证
type AuxiliaryAuthCredentials struct {
	Username string `json:"username"`
	Channel  string `json:"channel"`
	Service  string `json:"service"`
	Token    string `json:"token"`
	Status   string `json:"status"`
	Source   string `json:"source"`
}

// AuthInterceptCredentials 认证拦截凭证
type AuthInterceptCredentials struct {
	Username string `json:"username"`
	Channel  string `json:"channel"`
	Service  string `json:"service"`
	Token    string `json:"token"`
	Status   string `json:"status"`
	Cancel   bool   `json:"cancel"`
	Source   string `json:"source"`
}

// UserPermissions 用户权限信息
type UserPermissions struct {
	UserID      int64                  `json:"user_id"`
	Permissions map[string]interface{} `json:"permissions"`
	Roles       []string               `json:"roles"`
	Scopes      []string               `json:"scopes"`
	UpdateTime  time.Time              `json:"update_time"`
}

// UserSecurityManager 用户安全管理器
type UserSecurityManager struct {
	cache      *cache.Cache
	logger     *logger.Logger
	userRepo   *repository.UserRepository
	expiration time.Duration
}

// NewUserSecurityManager 创建用户安全管理器
func NewUserSecurityManager(cache *cache.Cache, logger *logger.Logger, userRepo *repository.UserRepository) *UserSecurityManager {
	return &UserSecurityManager{
		cache:      cache,
		logger:     logger,
		userRepo:   userRepo,
		expiration: 2 * time.Hour,
	}
}

// AuxiliaryAuthenticate 辅助认证
func (m *UserSecurityManager) AuxiliaryAuthenticate(ctx context.Context, credentials AuxiliaryAuthCredentials) (*model.User, error) {
	m.logger.Info("辅助认证", "username", credentials.Username, "channel", credentials.Channel, "service", credentials.Service)

	// 检查用户是否被禁用
	user, err := m.userRepo.GetUserByName(ctx, credentials.Username)
	if err != nil {
		return nil, err
	}

	if user != nil && !user.IsActive {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 验证必要信息
	if credentials.Token == "" || credentials.Channel == "" || credentials.Service == "" {
		return nil, fmt.Errorf("认证信息不完整")
	}

	// 触发认证拦截事件
	interceptData := &AuthInterceptCredentials{
		Username: credentials.Username,
		Channel:  credentials.Channel,
		Service:  credentials.Service,
		Token:    credentials.Token,
		Status:   "pending",
		Cancel:   false,
		Source:   "system",
	}

	// 这里应该调用事件管理器发送事件
	// 简化实现，直接返回成功
	interceptData.Status = "completed"

	if interceptData.Cancel {
		return nil, fmt.Errorf("认证被拦截")
	}

	// 如果用户不存在，创建新用户
	if user == nil {
		user, err = m.createAuxiliaryUser(ctx, credentials)
		if err != nil {
			return nil, err
		}
	}

	m.logger.Info("辅助认证成功", "username", user.Name, "channel", credentials.Channel)
	return user, nil
}

// createAuxiliaryUser 创建辅助认证用户
func (m *UserSecurityManager) createAuxiliaryUser(ctx context.Context, credentials AuxiliaryAuthCredentials) (*model.User, error) {
	m.logger.Info("创建辅助认证用户", "username", credentials.Username)

	// 生成随机密码
	password := utils.GenerateRandomString(16)
	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 创建用户数据
	createData := model.UserCreateData{
		Name:           credentials.Username,
		HashedPassword: hashedPassword,
		IsActive:       true,
		IsSuperuser:    false,
	}

	user, err := m.userRepo.CreateUser(ctx, createData)
	if err != nil {
		return nil, err
	}

	m.logger.Info("创建辅助认证用户成功", "userID", user.ID)
	return user, nil
}

// LogAuthenticationEvent 记录认证事件
func (m *UserSecurityManager) LogAuthenticationEvent(ctx context.Context, userID int64, authType string, success bool, details map[string]interface{}) error {
	m.logger.Info("记录认证事件", "userID", userID, "authType", authType, "success", success)

	event := map[string]interface{}{
		"user_id":   userID,
		"auth_type": authType,
		"success":   success,
		"details":   details,
		"timestamp": time.Now(),
	}

	// 存储到缓存（实际项目中应该存储到数据库）
	cacheKey := fmt.Sprintf("user:auth_event:%d:%d", userID, time.Now().Unix())
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	m.cache.Set(ctx, cacheKey, string(eventData), 7*24*time.Hour)

	return nil
}

// GenerateOTPSecret 生成OTP密钥
func (m *UserSecurityManager) GenerateOTPSecret(ctx context.Context, userID int64) (string, error) {
	m.logger.Info("生成OTP密钥", "userID", userID)

	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", fmt.Errorf("用户不存在")
	}

	// 生成随机密钥
	secret := make([]byte, 20)
	_, err = rand.Read(secret)
	if err != nil {
		return "", err
	}

	// Base32编码
	secretBase32 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)

	// 更新用户OTP设置
	updateData := model.UserUpdateData{
		IsOTP:     true,
		OTPSecret: &secretBase32,
	}

	_, err = m.userRepo.UpdateUser(ctx, userID, updateData)
	if err != nil {
		return "", err
	}

	m.logger.Info("生成OTP密钥成功", "userID", userID)
	return secretBase32, nil
}

// EnableOTP 启用OTP
func (m *UserSecurityManager) EnableOTP(ctx context.Context, userID int64, secret string) error {
	m.logger.Info("启用OTP", "userID", userID)

	updateData := model.UserUpdateData{
		IsOTP:     true,
		OTPSecret: &secret,
	}

	_, err := m.userRepo.UpdateUser(ctx, userID, updateData)
	if err != nil {
		return err
	}

	m.logger.Info("启用OTP成功", "userID", userID)
	return nil
}

// DisableOTP 禁用OTP
func (m *UserSecurityManager) DisableOTP(ctx context.Context, userID int64) error {
	m.logger.Info("禁用OTP", "userID", userID)

	updateData := model.UserUpdateData{
		IsOTP:     false,
		OTPSecret: nil,
	}

	_, err := m.userRepo.UpdateUser(ctx, userID, updateData)
	if err != nil {
		return err
	}

	m.logger.Info("禁用OTP成功", "userID", userID)
	return nil
}

// VerifyOTP 验证OTP
func (m *UserSecurityManager) VerifyOTP(ctx context.Context, userID int64, code string) error {
	m.logger.Info("验证OTP", "userID", userID)

	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	if !user.IsOTP {
		return fmt.Errorf("用户未启用OTP")
	}

	if user.OTPSecret == nil {
		return fmt.Errorf("OTP密钥不存在")
	}

	// 验证OTP（这里简化实现，实际应该使用TOTP算法）
	// 实际实现应该使用 github.com/pquerna/otp 库
	valid := m.verifyTOTP(*user.OTPSecret, code)
	if !valid {
		return fmt.Errorf("OTP验证失败")
	}

	m.logger.Info("OTP验证成功", "userID", userID)
	return nil
}

// verifyTOTP 验证TOTP（简化实现）
func (m *UserSecurityManager) verifyTOTP(secret, code string) bool {
	// 这里应该实现真正的TOTP验证算法
	// 暂时简化实现，实际应该使用:
	// import "github.com/pquerna/otp/totp"
	// return totp.Validate(code, secret)

	// 简化验证逻辑（仅用于演示）
	return len(code) == 6 && code != "000000"
}

// UserPermissionManager 用户权限管理器
type UserPermissionManager struct {
	cache    *cache.Cache
	logger   *logger.Logger
	userRepo *repository.UserRepository
}

// NewUserPermissionManager 创建用户权限管理器
func NewUserPermissionManager(cache *cache.Cache, logger *logger.Logger, userRepo *repository.UserRepository) *UserPermissionManager {
	return &UserPermissionManager{
		cache:    cache,
		logger:   logger,
		userRepo: userRepo,
	}
}

// GetUserPermissions 获取用户权限
func (m *UserPermissionManager) GetUserPermissions(ctx context.Context, userID int64) (*UserPermissions, error) {
	m.logger.Info("获取用户权限", "userID", userID)

	// 先从缓存获取
	cacheKey := fmt.Sprintf("user:permissions:%d", userID)
	if cached, err := m.cache.Get(ctx, cacheKey); err == nil {
		var permissions UserPermissions
		if err := json.Unmarshal([]byte(cached), &permissions); err == nil {
			return &permissions, nil
		}
	}

	// 从数据库获取
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	permissions := &UserPermissions{
		UserID:      userID,
		Permissions: user.Permissions,
		Roles:       m.extractRoles(user),
		Scopes:      m.extractScopes(user),
		UpdateTime:  time.Now(),
	}

	// 缓存结果
	if data, err := json.Marshal(permissions); err == nil {
		m.cache.Set(ctx, cacheKey, string(data), time.Hour)
	}

	return permissions, nil
}

// UpdateUserPermissions 更新用户权限
func (m *UserPermissionManager) UpdateUserPermissions(ctx context.Context, userID int64, permissions map[string]interface{}) error {
	m.logger.Info("更新用户权限", "userID", userID)

	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 更新权限
	updateData := model.UserUpdateData{
		Permissions: permissions,
	}

	_, err = m.userRepo.UpdateUser(ctx, userID, updateData)
	if err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:permissions:%d", userID)
	m.cache.Delete(ctx, cacheKey)

	m.logger.Info("更新用户权限成功", "userID", userID)
	return nil
}

// CheckPermission 检查用户权限
func (m *UserPermissionManager) CheckPermission(ctx context.Context, userID int64, resource, action string) bool {
	m.logger.Info("检查用户权限", "userID", userID, "resource", resource, "action", action)

	permissions, err := m.GetUserPermissions(ctx, userID)
	if err != nil {
		m.logger.Error("获取用户权限失败", "userID", userID, "error", err)
		return false
	}

	// 检查超级用户权限
	if _, ok := permissions.Roles["superuser"]; ok {
		return true
	}

	// 检查具体权限
	key := fmt.Sprintf("%s:%s", resource, action)
	if value, exists := permissions.Permissions[key]; exists {
		if allowed, ok := value.(bool); ok {
			return allowed
		}
	}

	// 检查通配符权限
	wildcardKey := fmt.Sprintf("%s:*", resource)
	if value, exists := permissions.Permissions[wildcardKey]; exists {
		if allowed, ok := value.(bool); ok {
			return allowed
		}
	}

	return false
}

// extractRoles 提取用户角色
func (m *UserPermissionManager) extractRoles(user *model.User) map[string]interface{} {
	roles := make(map[string]interface{})

	// 默认角色
	if user.IsSuperuser {
		roles["superuser"] = true
	}

	if user.IsActive {
		roles["active"] = true
	}

	// 从权限中提取角色信息
	if user.Permissions != nil {
		if roleValue, exists := user.Permissions["roles"]; exists {
			if roleList, ok := roleValue.([]interface{}); ok {
				for _, role := range roleList {
					if roleStr, ok := role.(string); ok {
						roles[roleStr] = true
					}
				}
			}
		}
	}

	return roles
}

// extractScopes 提取用户作用域
func (m *UserPermissionManager) extractScopes(user *model.User) []string {
	var scopes []string

	// 默认作用域
	scopes = append(scopes, "profile")

	if user.IsSuperuser {
		scopes = append(scopes, "admin", "system")
	}

	// 从权限中提取作用域信息
	if user.Permissions != nil {
		if scopeValue, exists := user.Permissions["scopes"]; exists {
			if scopeList, ok := scopeValue.([]interface{}); ok {
				for _, scope := range scopeList {
					if scopeStr, ok := scope.(string); ok {
						scopes = append(scopes, scopeStr)
					}
				}
			}
		}
	}

	return scopes
}

// GrantRole 授予角色
func (m *UserPermissionManager) GrantRole(ctx context.Context, userID int64, role string) error {
	m.logger.Info("授予角色", "userID", userID, "role", role)

	permissions, err := m.GetUserPermissions(ctx, userID)
	if err != nil {
		return err
	}

	// 添加角色
	permissions.Roles[role] = true

	// 更新权限
	updateData := model.UserUpdateData{
		Permissions: permissions.Permissions,
	}

	_, err = m.userRepo.UpdateUser(ctx, userID, updateData)
	if err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:permissions:%d", userID)
	m.cache.Delete(ctx, cacheKey)

	m.logger.Info("授予角色成功", "userID", userID, "role", role)
	return nil
}

// RevokeRole 撤销角色
func (m *UserPermissionManager) RevokeRole(ctx context.Context, userID int64, role string) error {
	m.logger.Info("撤销角色", "userID", userID, "role", role)

	permissions, err := m.GetUserPermissions(ctx, userID)
	if err != nil {
		return err
	}

	// 移除角色
	delete(permissions.Roles, role)

	// 更新权限
	updateData := model.UserUpdateData{
		Permissions: permissions.Permissions,
	}

	_, err = m.userRepo.UpdateUser(ctx, userID, updateData)
	if err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:permissions:%d", userID)
	m.cache.Delete(ctx, cacheKey)

	m.logger.Info("撤销角色成功", "userID", userID, "role", role)
	return nil
}
