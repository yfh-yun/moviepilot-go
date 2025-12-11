package user

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"moviepilot-go/internal/business/services/base"
	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/security"
	"moviepilot-go/pkg/utils"
)

// UserService 用户服务
// 原UserChain，负责用户认证和管理
type UserService struct {
	*base.ServiceBase
	repo           interfaces.UserRepository
	userConfigRepo interfaces.UserConfigRepository
}

// NewUserService 创建UserService实例
func NewUserService(repo interfaces.UserRepository, userConfigRepo interfaces.UserConfigRepository) *UserService {
	return &UserService{
		ServiceBase:    base.NewServiceBase(),
		repo:           repo,
		userConfigRepo: userConfigRepo,
	}
}

// Initialize 初始化服务
func (s *UserService) Initialize() error {
	return nil
}

// Name 获取服务名称
func (s *UserService) Name() string {
	return "UserService"
}

// Close 关闭服务
func (s *UserService) Close() error {
	return nil
}

// GetCurrentUser 根据访问 Token 获取当前用户信息
// 等价于 Python get_current_user (同步版本)
func (s *UserService) GetCurrentUser(ctx context.Context, token string) (*dto.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}
	// 使用默认密钥和超时时间创建 JWT 管理器
	jwtManager := security.NewJWTManager("default_secret_key", 24*time.Hour, 7*24*time.Hour)
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	idStr := strconv.FormatUint(uint64(claims.UserID), 10)
	user, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return dbUserToDTO(user)
}

// GetCurrentActiveUser 获取当前激活用户
// 等价于 Python get_current_active_user*
func (s *UserService) GetCurrentActiveUser(ctx context.Context, token string) (*dto.User, error) {
	user, err := s.GetCurrentUser(ctx, token)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, fmt.Errorf("user not active")
	}
	return user, nil
}

// GetCurrentActiveSuperuser 获取当前激活的超级管理员
// 等价于 Python get_current_active_superuser*
func (s *UserService) GetCurrentActiveSuperuser(ctx context.Context, token string) (*dto.User, error) {
	user, err := s.GetCurrentActiveUser(ctx, token)
	if err != nil {
		return nil, err
	}
	if !user.IsSuperuser {
		return nil, fmt.Errorf("permission denied: not superuser")
	}
	return user, nil
}

// Authenticate 用户认证
func (s *UserService) Authenticate(ctx context.Context, username, password string) (*dto.Token, error) {
	// TODO: 实现认证逻辑
	// 1. 验证用户名密码
	// 2. 检查OTP（如果启用）
	// 3. 生成Token
	return nil, nil
}

// AuthenticateWithOTP 使用OTP认证
func (s *UserService) AuthenticateWithOTP(ctx context.Context, username, password, otpCode string) (*dto.Token, error) {
	// TODO: 实现OTP认证逻辑
	return nil, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, user *dto.UserCreate) (*dto.User, error) {
	// 检查用户名是否已存在
	existingUser, err := s.repo.GetByUsername(ctx, user.Name)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 密码哈希
	passwordManager := security.NewPasswordManager(security.DefaultPasswordConfig)
	passwordHash, err := passwordManager.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户模型
	dbUser := &database.User{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: passwordHash,
		IsActive:     user.IsActive,
		IsSuperuser:  user.IsSuperuser,
		Avatar:       user.Avatar,
		IsOTP:        user.IsOTP,
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, dbUser); err != nil {
		return nil, err
	}

	return dbUserToDTO(dbUser)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, user *dto.UserUpdate) error {
	// 获取现有用户
	idStr := strconv.Itoa(user.ID)
	dbUser, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return err
	}
	if dbUser == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查用户名是否已被其他用户使用
	existingUser, err := s.repo.GetByUsername(ctx, user.Name)
	if err != nil {
		return err
	}
	if existingUser != nil && int(existingUser.ID) != user.ID {
		return fmt.Errorf("用户名已被使用")
	}

	// 更新用户字段
	dbUser.Name = user.Name
	dbUser.Email = user.Email
	dbUser.IsActive = user.IsActive
	dbUser.IsSuperuser = user.IsSuperuser
	dbUser.Avatar = user.Avatar
	dbUser.IsOTP = user.IsOTP

	// 如果提供了密码，则更新密码
	if user.Password != "" {
		passwordManager := security.NewPasswordManager(security.DefaultPasswordConfig)
		passwordHash, err := passwordManager.HashPassword(user.Password)
		if err != nil {
			return err
		}
		dbUser.PasswordHash = passwordHash
	}

	// 保存更新
	return s.repo.Update(ctx, dbUser)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, userID int) error {
	idStr := strconv.Itoa(userID)
	return s.repo.Delete(ctx, idStr)
}

// GetUser 获取用户
func (s *UserService) GetUser(ctx context.Context, userID int) (*dto.User, error) {
	idStr := strconv.Itoa(userID)
	user, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return dbUserToDTO(user)
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context) ([]*dto.User, error) {
	params := interfaces.ListUserParams{
		Page:     1,
		PageSize: 1000,
	}
	users, _, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.User, 0, len(users))
	for _, u := range users {
		du, err := dbUserToDTO(u)
		if err != nil {
			return nil, err
		}
		result = append(result, du)
	}
	return result, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	idStr := strconv.Itoa(userID)
	user, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	passwordManager := security.NewPasswordManager(security.DefaultPasswordConfig)
	if verifyErr := passwordManager.VerifyPassword(user.PasswordHash, oldPassword); verifyErr != nil {
		return fmt.Errorf("旧密码错误")
	}

	// 哈希新密码
	passwordHash, err := passwordManager.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	user.PasswordHash = passwordHash
	return s.repo.Update(ctx, user)
}

// EnableOTP 启用OTP
func (s *UserService) EnableOTP(ctx context.Context, userID int) (string, error) {
	idStr := strconv.Itoa(userID)
	user, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", fmt.Errorf("用户不存在")
	}

	// 生成OTP密钥
	secret, _ := utils.GenerateSecretKey(user.Name)
	if secret == "" {
		return "", fmt.Errorf("生成OTP密钥失败")
	}
	user.OTPSecret = secret
	user.IsOTP = true

	// 保存更新
	if err := s.repo.Update(ctx, user); err != nil {
		return "", err
	}

	return secret, nil
}

// DisableOTP 禁用OTP
func (s *UserService) DisableOTP(ctx context.Context, userID int) error {
	idStr := strconv.Itoa(userID)
	user, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 禁用OTP
	user.IsOTP = false
	user.OTPSecret = ""

	// 保存更新
	return s.repo.Update(ctx, user)
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*dto.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return dbUserToDTO(user)
}

// UpdateAvatar 更新用户头像
func (s *UserService) UpdateAvatar(ctx context.Context, userID int, avatar string) error {
	idStr := strconv.Itoa(userID)
	user, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 更新头像
	user.Avatar = avatar

	// 保存更新
	return s.repo.Update(ctx, user)
}

// GenerateOTPURI 生成OTP验证URI
func (s *UserService) GenerateOTPURI(ctx context.Context, userID int) (string, string, error) {
	idStr := strconv.Itoa(userID)
	user, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", fmt.Errorf("用户不存在")
	}

	// 生成OTP密钥和URI
	secret, uri := utils.GenerateSecretKey(user.Name)
	return secret, uri, nil
}

// VerifyOTP 验证OTP
func (s *UserService) VerifyOTP(ctx context.Context, userID int, otpCode string) (bool, error) {
	idStr := strconv.Itoa(userID)
	user, err := s.repo.GetByID(ctx, idStr)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, fmt.Errorf("用户不存在")
	}

	// 验证OTP
	return utils.Check(user.OTPSecret, otpCode), nil
}

// GetUserConfig 获取用户配置
func (s *UserService) GetUserConfig(ctx context.Context, userID int, key string) (string, error) {
	idStr := strconv.Itoa(userID)
	params := interfaces.ListUserConfigParams{
		Page:     1,
		PageSize: 1,
		UserID:   idStr,
		Key:      key,
	}

	configs, _, err := s.userConfigRepo.List(ctx, params)
	if err != nil {
		return "", err
	}

	if len(configs) > 0 {
		return configs[0].Value, nil
	}

	return "", nil
}

// SetUserConfig 设置用户配置
func (s *UserService) SetUserConfig(ctx context.Context, userID int, key, value string) error {
	idStr := strconv.Itoa(userID)
	params := interfaces.ListUserConfigParams{
		Page:     1,
		PageSize: 1,
		UserID:   idStr,
		Key:      key,
	}

	configs, _, err := s.userConfigRepo.List(ctx, params)
	if err != nil {
		return err
	}

	if len(configs) > 0 {
		// 更新现有配置
		config := configs[0]
		config.Value = value
		return s.userConfigRepo.Update(ctx, config)
	}

	// 创建新配置
	newConfig := &database.UserConfig{
		UserID: idStr,
		Key:    key,
		Value:  value,
		Type:   "string",
	}

	return s.userConfigRepo.Create(ctx, newConfig)
}

// dbUserToDTO 将数据库用户模型转换为 DTO 用户
func dbUserToDTO(u *database.User) (*dto.User, error) {
	if u == nil {
		return nil, fmt.Errorf("nil user")
	}
	res := &dto.User{
		UserInDBBase: dto.UserInDBBase{
			UserBase: dto.UserBase{
				Name:        u.Name,
				Email:       u.Email,
				IsActive:    u.IsActive,
				IsSuperuser: u.IsSuperuser,
				Avatar:      u.Avatar,
				IsOTP:       u.IsOTP,
			},
			ID: int(u.ID),
		},
	}

	// 解析权限和设置（JSON 字符串），忽略解析错误
	if u.Permissions != "" {
		var perms map[string]any
		if err := json.Unmarshal([]byte(u.Permissions), &perms); err == nil {
			res.Permissions = perms
		}
	}
	if u.Settings != "" {
		var settings map[string]any
		if err := json.Unmarshal([]byte(u.Settings), &settings); err == nil {
			res.Settings = settings
		}
	}
	return res, nil
}
