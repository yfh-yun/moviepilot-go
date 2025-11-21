package chain

import (
	"context"
	"errors"
	"fmt"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/cache"
	"moviepilot-go/internal/infrastructure/security"
)

// UserChain 用户认证处理链
type UserChain struct {
	cache                *cache.Cache
	logger               *logger.Logger
	userRepo             *repository.UserRepository
	userService          *service.UserService
	preferenceManager    *UserPreferenceManager
	activityTracker      *UserActivityTracker
	securityManager      *UserSecurityManager
	permissionManager    *UserPermissionManager
	recommendationEngine *PersonalizedRecommendationEngine
}

// NewUserChain 创建用户认证处理链实例
func NewUserChain(cache *cache.Cache, logger *logger.Logger, userRepo *repository.UserRepository) *UserChain {
	// 创建各个管理器
	preferenceManager := NewUserPreferenceManager(cache, logger, userRepo)
	activityTracker := NewUserActivityTracker(cache, logger, userRepo)
	securityManager := NewUserSecurityManager(cache, logger, userRepo)
	permissionManager := NewUserPermissionManager(cache, logger, userRepo)
	recommendationEngine := NewPersonalizedRecommendationEngine(preferenceManager, activityTracker, logger)

	return &UserChain{
		cache:                cache,
		logger:               logger,
		userRepo:             userRepo,
		userService:          service.NewUserService(userRepo, logger),
		preferenceManager:    preferenceManager,
		activityTracker:      activityTracker,
		securityManager:      securityManager,
		permissionManager:    permissionManager,
		recommendationEngine: recommendationEngine,
	}
}

// UserAuthenticate 用户认证
func (c *UserChain) UserAuthenticate(ctx context.Context, credentials model.AuthCredentials) (*model.User, error) {
	c.logger.Info("用户认证", "username", credentials.Username, "grantType", credentials.GrantType)

	// 根据认证类型处理
	switch credentials.GrantType {
	case "password":
		return c.passwordAuthenticate(ctx, credentials)
	case "authorization_code":
		return c.authorizationCodeAuthenticate(ctx, credentials)
	default:
		return nil, errors.New("不支持的认证类型")
	}
}

// passwordAuthenticate 密码认证
func (c *UserChain) passwordAuthenticate(ctx context.Context, credentials model.AuthCredentials) (*model.User, error) {
	// 参数验证
	if credentials.Username == "" || credentials.Password == "" {
		return nil, errors.New("用户名或密码不能为空")
	}

	// 查询用户
	user, err := c.userRepo.GetUserByName(ctx, credentials.Username)
	if err != nil {
		c.logger.Error("查询用户失败", "error", err)
		return nil, err
	}

	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if !user.IsActive {
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	if !security.VerifyPassword(credentials.Password, user.HashedPassword) {
		return nil, errors.New("用户名或密码不正确")
	}

	// 验证MFA（如果需要）
	if user.IsOTP {
		if credentials.MfaCode == "" {
			return nil, errors.New("需要二次验证码")
		}
		if !c.verifyMFA(user, credentials.MfaCode) {
			return nil, errors.New("二次验证码不正确")
		}
	}

	c.logger.Info("密码认证成功", "username", user.Name)
	return user, nil
}

// authorizationCodeAuthenticate 授权码认证
func (c *UserChain) authorizationCodeAuthenticate(ctx context.Context, credentials model.AuthCredentials) (*model.User, error) {
	if credentials.Code == "" {
		return nil, errors.New("授权码不能为空")
	}

	// 这里可以实现第三方认证逻辑
	// 例如：通过授权码获取用户信息，然后创建或更新本地用户

	c.logger.Warn("授权码认证暂未实现")
	return nil, errors.New("授权码认证暂未实现")
}

// verifyMFA 验证MFA
func (c *UserChain) verifyMFA(user *model.User, mfaCode string) bool {
	// 这里可以实现MFA验证逻辑
	// 例如：使用TOTP验证

	// 暂简单实现，实际应该使用TOTP算法
	if user.IsOTP && mfaCode != "" {
		// 简化验证，实际应该使用安全的MFA验证
		return true
	}

	return false
}

// CreateUser 创建用户
func (c *UserChain) CreateUser(ctx context.Context, userData model.UserCreateData) (*model.User, error) {
	c.logger.Info("创建用户", "username", userData.Name)

	// 验证用户名是否已存在
	existingUser, err := c.userRepo.GetUserByName(ctx, userData.Name)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 创建用户
	user, err := c.userService.CreateUser(ctx, userData)
	if err != nil {
		c.logger.Error("创建用户失败", "error", err)
		return nil, err
	}

	c.logger.Info("创建用户成功", "userID", user.ID)
	return user, nil
}

// UpdateUser 更新用户信息
func (c *UserChain) UpdateUser(ctx context.Context, userID int64, updateData model.UserUpdateData) (*model.User, error) {
	c.logger.Info("更新用户信息", "userID", userID)

	user, err := c.userService.UpdateUser(ctx, userID, updateData)
	if err != nil {
		c.logger.Error("更新用户失败", "error", err)
		return nil, err
	}

	c.logger.Info("更新用户成功", "userID", userID)
	return user, nil
}

// DeleteUser 删除用户
func (c *UserChain) DeleteUser(ctx context.Context, userID int64) error {
	c.logger.Info("删除用户", "userID", userID)

	err := c.userService.DeleteUser(ctx, userID)
	if err != nil {
		c.logger.Error("删除用户失败", "error", err)
		return err
	}

	c.logger.Info("删除用户成功", "userID", userID)
	return nil
}

// GetUserList 获取用户列表
func (c *UserChain) GetUserList(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	c.logger.Info("获取用户列表", "page", page, "pageSize", pageSize)

	users, total, err := c.userRepo.GetUserList(ctx, page, pageSize)
	if err != nil {
		c.logger.Error("获取用户列表失败", "error", err)
		return nil, 0, err
	}

	c.logger.Info("获取用户列表成功", "count", len(users))
	return users, total, nil
}

// ResetPassword 重置密码
func (c *UserChain) ResetPassword(ctx context.Context, userID int64, newPassword string) error {
	c.logger.Info("重置用户密码", "userID", userID)

	// 验证新密码
	if newPassword == "" {
		return errors.New("新密码不能为空")
	}

	// 更新密码
	err := c.userService.UpdatePassword(ctx, userID, newPassword)
	if err != nil {
		c.logger.Error("重置密码失败", "error", err)
		return err
	}

	c.logger.Info("重置密码成功", "userID", userID)
	return nil
}

// EnableUser 启用用户
func (c *UserChain) EnableUser(ctx context.Context, userID int64) error {
	c.logger.Info("启用用户", "userID", userID)

	err := c.userService.EnableUser(ctx, userID)
	if err != nil {
		c.logger.Error("启用用户失败", "error", err)
		return err
	}

	c.logger.Info("启用用户成功", "userID", userID)
	return nil
}

// DisableUser 禁用用户
func (c *UserChain) DisableUser(ctx context.Context, userID int64) error {
	c.logger.Info("禁用用户", "userID", userID)

	err := c.userService.DisableUser(ctx, userID)
	if err != nil {
		c.logger.Error("禁用用户失败", "error", err)
		return err
	}

	c.logger.Info("禁用用户成功", "userID", userID)
	return nil
}

// ===== 用户偏好管理方法 =====

// GetUserPreferences 获取用户偏好设置
func (c *UserChain) GetUserPreferences(ctx context.Context, userID int64) (*UserPreference, error) {
	return c.preferenceManager.GetUserPreferences(ctx, userID)
}

// UpdateUserPreferences 更新用户偏好设置
func (c *UserChain) UpdateUserPreferences(ctx context.Context, userID int64, preferences map[string]interface{}) error {
	return c.preferenceManager.UpdateUserPreferences(ctx, userID, preferences)
}

// GetFavoriteGenres 获取用户喜欢的类型
func (c *UserChain) GetFavoriteGenres(ctx context.Context, userID int64) ([]string, error) {
	return c.preferenceManager.GetFavoriteGenres(ctx, userID)
}

// GetPreferredQuality 获取用户偏好的质量
func (c *UserChain) GetPreferredQuality(ctx context.Context, userID int64) ([]string, error) {
	return c.preferenceManager.GetPreferredQuality(ctx, userID)
}

// GetPreferredLanguage 获取用户偏好的语言
func (c *UserChain) GetPreferredLanguage(ctx context.Context, userID int64) ([]string, error) {
	return c.preferenceManager.GetPreferredLanguage(ctx, userID)
}

// ===== 用户活动跟踪方法 =====

// TrackActivity 跟踪用户活动
func (c *UserChain) TrackActivity(ctx context.Context, activity *UserActivity) error {
	return c.activityTracker.TrackActivity(ctx, activity)
}

// TrackLogin 跟踪登录活动
func (c *UserChain) TrackLogin(ctx context.Context, userID int64, ipAddress, userAgent string) error {
	activity := &UserActivity{
		UserID:    userID,
		Action:    "login",
		Resource:  "system",
		Details:   map[string]interface{}{"action": "user_login"},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	return c.activityTracker.TrackActivity(ctx, activity)
}

// TrackDownload 跟踪下载活动
func (c *UserChain) TrackDownload(ctx context.Context, userID int64, mediaID int64, title string) error {
	activity := &UserActivity{
		UserID:   userID,
		Action:   "download",
		Resource: "media",
		Details: map[string]interface{}{
			"media_id": mediaID,
			"title":    title,
		},
	}

	return c.activityTracker.TrackActivity(ctx, activity)
}

// TrackSearch 跟踪搜索活动
func (c *UserChain) TrackSearch(ctx context.Context, userID int64, keyword string, resultCount int) error {
	activity := &UserActivity{
		UserID:   userID,
		Action:   "search",
		Resource: "search",
		Details: map[string]interface{}{
			"keyword":      keyword,
			"result_count": resultCount,
		},
	}

	return c.activityTracker.TrackActivity(ctx, activity)
}

// GetUserActivities 获取用户活动列表
func (c *UserChain) GetUserActivities(ctx context.Context, userID int64, limit int) ([]*UserActivity, error) {
	return c.activityTracker.GetUserActivities(ctx, userID, limit)
}

// GetUserStats 获取用户统计信息
func (c *UserChain) GetUserStats(ctx context.Context, userID int64) (*UserStats, error) {
	return c.activityTracker.GetUserStats(ctx, userID)
}

// ===== 个性化推荐方法 =====

// GetPersonalizedRecommendations 获取个性化推荐
func (c *UserChain) GetPersonalizedRecommendations(ctx context.Context, userID int64, limit int) (*model.PersonalizedRecommendations, error) {
	return c.recommendationEngine.GetPersonalizedRecommendations(ctx, userID, limit)
}

// UpdateRecommendationWeights 更新推荐权重
func (c *UserChain) UpdateRecommendationWeights(ctx context.Context, userID int64, action string, mediaInfo *model.MediaInfo) error {
	return c.recommendationEngine.UpdateRecommendationWeights(ctx, userID, action, mediaInfo)
}

// ===== 用户安全方法 =====

// AuxiliaryAuthenticate 辅助认证
func (c *UserChain) AuxiliaryAuthenticate(ctx context.Context, credentials AuxiliaryAuthCredentials) (*model.User, error) {
	return c.securityManager.AuxiliaryAuthenticate(ctx, credentials)
}

// GenerateOTPSecret 生成OTP密钥
func (c *UserChain) GenerateOTPSecret(ctx context.Context, userID int64) (string, error) {
	return c.securityManager.GenerateOTPSecret(ctx, userID)
}

// EnableOTP 启用OTP
func (c *UserChain) EnableOTP(ctx context.Context, userID int64, secret string) error {
	return c.securityManager.EnableOTP(ctx, userID, secret)
}

// DisableOTP 禁用OTP
func (c *UserChain) DisableOTP(ctx context.Context, userID int64) error {
	return c.securityManager.DisableOTP(ctx, userID)
}

// VerifyOTP 验证OTP
func (c *UserChain) VerifyOTP(ctx context.Context, userID int64, code string) error {
	return c.securityManager.VerifyOTP(ctx, userID, code)
}

// LogAuthenticationEvent 记录认证事件
func (c *UserChain) LogAuthenticationEvent(ctx context.Context, userID int64, authType string, success bool, details map[string]interface{}) error {
	return c.securityManager.LogAuthenticationEvent(ctx, userID, authType, success, details)
}

// ===== 用户权限管理方法 =====

// GetUserPermissions 获取用户权限
func (c *UserChain) GetUserPermissions(ctx context.Context, userID int64) (*UserPermissions, error) {
	return c.permissionManager.GetUserPermissions(ctx, userID)
}

// UpdateUserPermissions 更新用户权限
func (c *UserChain) UpdateUserPermissions(ctx context.Context, userID int64, permissions map[string]interface{}) error {
	return c.permissionManager.UpdateUserPermissions(ctx, userID, permissions)
}

// CheckPermission 检查用户权限
func (c *UserChain) CheckPermission(ctx context.Context, userID int64, resource, action string) bool {
	return c.permissionManager.CheckPermission(ctx, userID, resource, action)
}

// GrantRole 授予角色
func (c *UserChain) GrantRole(ctx context.Context, userID int64, role string) error {
	return c.permissionManager.GrantRole(ctx, userID, role)
}

// RevokeRole 撤销角色
func (c *UserChain) RevokeRole(ctx context.Context, userID int64, role string) error {
	return c.permissionManager.RevokeRole(ctx, userID, role)
}

// ===== 用户管理扩展方法 =====

// GetUserByName 根据用户名获取用户
func (c *UserChain) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	return c.userRepo.GetUserByName(ctx, name)
}

// GetUserByID 根据ID获取用户
func (c *UserChain) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return c.userRepo.GetUserByID(ctx, userID)
}

// GetUserPermissionsByName 根据用户名获取权限
func (c *UserChain) GetUserPermissionsByName(ctx context.Context, name string) (map[string]interface{}, error) {
	user, err := c.userRepo.GetUserByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	return user.Permissions, nil
}

// GetUserSettingsByName 根据用户名获取设置
func (c *UserChain) GetUserSettingsByName(ctx context.Context, name string) (map[string]interface{}, error) {
	user, err := c.userRepo.GetUserByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	return user.Settings, nil
}

// GetUserSettingByName 根据用户名获取特定设置
func (c *UserChain) GetUserSettingByName(ctx context.Context, name, key string) (string, error) {
	settings, err := c.GetUserSettingsByName(ctx, name)
	if err != nil {
		return "", err
	}

	if value, exists := settings[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue, nil
		}
	}

	return "", nil
}

// GetUserNameByBinding 根据绑定账号获取用户名
func (c *UserChain) GetUserNameByBinding(ctx context.Context, bindings map[string]interface{}) (string, error) {
	users, _, err := c.userRepo.GetUserList(ctx, 1, 1000)
	if err != nil {
		return "", err
	}

	for _, user := range users {
		if user.Settings != nil {
			match := true
			for key, value := range bindings {
				if userSettingsValue, exists := user.Settings[key]; exists {
					if fmt.Sprintf("%v", userSettingsValue) != fmt.Sprintf("%v", value) {
						match = false
						break
					}
				} else {
					match = false
					break
				}
			}
			if match {
				return user.Name, nil
			}
		}
	}

	return "", nil
}
