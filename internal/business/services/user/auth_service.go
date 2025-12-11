package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"moviepilot-go/internal/models/database"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/internal/repositories/interfaces"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/security"

	"go.uber.org/zap"
)

var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrInvalidPassword    = errors.New("密码错误")
	ErrUserExists         = errors.New("用户已存在")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
)

// AuthService 认证服务
type AuthService struct {
	userRepo   interfaces.UserRepository
	logger     *zap.Logger
	jwtManager security.JWTManager
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo interfaces.UserRepository, jwtManager security.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		logger:     logger.GetLogger(),
		jwtManager: jwtManager,
	}
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *dto.UserCreate) (*dto.User, error) {
	s.logger.Info("User registration attempt",
		zap.String("username", req.Name))

	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Name)
	if err != nil {
		s.logger.Error("Failed to check existing user", zap.Error(err))
		return nil, fmt.Errorf("检查用户失败: %w", err)
	}
	if existingUser != nil {
		s.logger.Warn("User already exists", zap.String("username", req.Name))
		return nil, ErrUserExists
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		existingEmail, emailErr := s.userRepo.GetByEmail(ctx, req.Email)
		if emailErr != nil {
			s.logger.Error("Failed to check existing email", zap.Error(emailErr))
			return nil, fmt.Errorf("检查邮箱失败: %w", emailErr)
		}
		if existingEmail != nil {
			s.logger.Warn("Email already exists", zap.String("email", req.Email))
			return nil, errors.New("邮箱已被使用")
		}
	}

	// 加密密码
	passwordManager := security.NewPasswordManager(security.DefaultPasswordConfig)
	hashedPassword, err := passwordManager.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &database.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		IsActive:     true,
		IsSuperuser:  false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user",
			zap.String("username", req.Name),
			zap.Error(err))
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	s.logger.Info("User registered successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Name))

	return &dto.User{
		UserInDBBase: dto.UserInDBBase{
			UserBase: dto.UserBase{
				Name:        user.Name,
				Email:       user.Email,
				IsActive:    user.IsActive,
				IsSuperuser: user.IsSuperuser,
				Avatar:      user.Avatar,
			},
			ID: int(user.ID),
		},
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, username, password, otpPassword string) (*dto.Token, error) {
	s.logger.Info("User login attempt", zap.String("username", username), zap.Bool("has_otp", otpPassword != ""))

	// 首次登录初始化逻辑：当系统中还没有任何用户时，自动创建超级管理员
	hasAny, err := s.userRepo.HasAny(ctx)
	if err != nil {
		s.logger.Error("Failed to check existing users", zap.Error(err))
		return nil, fmt.Errorf("检查用户失败: %w", err)
	}
	if !hasAny {
		if username == "" {
			return nil, errors.New("首次登录必须提供用户名")
		}

		// 生成随机强密码
		passwordManager := security.NewPasswordManager(security.DefaultPasswordConfig)
		randomPassword, generateErr := generateRandomPassword(16)
		if generateErr != nil {
			s.logger.Error("Failed to generate random password", zap.Error(generateErr))
			return nil, fmt.Errorf("生成随机密码失败: %w", generateErr)
		}

		hashedPassword, hashErr := passwordManager.HashPassword(randomPassword)
		if hashErr != nil {
			s.logger.Error("Failed to hash initial admin password", zap.Error(hashErr))
			return nil, fmt.Errorf("初始化管理员密码加密失败: %w", hashErr)
		}

		// 创建超级管理员用户
		user := &database.User{
			Name:         username,
			Email:        "",
			PasswordHash: hashedPassword,
			IsActive:     true,
			IsSuperuser:  true,
		}
		if createErr := s.userRepo.Create(ctx, user); createErr != nil {
			s.logger.Error("Failed to create initial superuser",
				zap.String("username", username),
				zap.Error(createErr))
			return nil, fmt.Errorf("创建初始管理员失败: %w", createErr)
		}

		s.logger.Info("Initial superuser created",
			zap.Uint("user_id", user.ID),
			zap.String("username", user.Name))

		// 为初始管理员生成访问 Token
		roles := []string{"user", "admin"}
		accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Name, roles)
		if err != nil {
			s.logger.Error("Failed to generate access token for initial superuser", zap.Error(err))
			return nil, fmt.Errorf("生成 Token 失败: %w", err)
		}

		// 返回 Token，并在初次登录时附带随机密码，便于前端提示用户修改
		return &dto.Token{
			AccessToken:     accessToken,
			TokenType:       "Bearer",
			SuperUser:       true,
			UserID:          int(user.ID),
			UserName:        user.Name,
			Avatar:          user.Avatar,
			InitialPassword: randomPassword,
		}, nil
	}

	// 正常登录流程
	// 查询用户
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		s.logger.Warn("User not found", zap.String("username", username))
		return nil, ErrInvalidCredentials
	}

	// 检查用户是否激活
	if !user.IsActive {
		s.logger.Warn("User is not active", zap.String("username", username))
		return nil, errors.New("用户已被禁用")
	}

	// 验证密码
	passwordManager := security.NewPasswordManager(security.DefaultPasswordConfig)
	if verifyErr := passwordManager.VerifyPassword(user.PasswordHash, password); verifyErr != nil {
		s.logger.Warn("Invalid password", zap.String("username", username))
		return nil, ErrInvalidCredentials
	}

	// 生成 Token
	roles := []string{"user"}
	if user.IsSuperuser {
		roles = append(roles, "admin")
	}
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Name, roles)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	// 更新最后登录时间
	if updateErr := s.userRepo.UpdateLastLogin(ctx, fmt.Sprintf("%d", user.ID)); updateErr != nil {
		s.logger.Warn("Failed to update last login time", zap.Error(updateErr))
		// 不影响登录流程
	}

	s.logger.Info("User logged in successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Name))

	return &dto.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		SuperUser:   user.IsSuperuser,
		UserID:      int(user.ID),
		UserName:    user.Name,
		Avatar:      user.Avatar,
	}, nil
}

// generateRandomPassword 生成指定长度的随机密码（字母数字组合）
func generateRandomPassword(length int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length < 8 {
		length = 8
	}
	buf := make([]byte, length)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		buf[i] = letters[n.Int64()]
	}
	return string(buf), nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	userIDStr := fmt.Sprintf("%d", userID)
	s.logger.Info("Change password attempt", zap.String("user_id", userIDStr))

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userIDStr)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 验证旧密码
	passwordManager := security.NewPasswordManager(security.DefaultPasswordConfig)
	if verifyErr := passwordManager.VerifyPassword(user.PasswordHash, oldPassword); verifyErr != nil {
		s.logger.Warn("Invalid old password", zap.String("user_id", userIDStr))
		return ErrInvalidPassword
	}

	// 加密新密码
	hashedPassword, hashErr := passwordManager.HashPassword(newPassword)
	if hashErr != nil {
		s.logger.Error("Failed to hash new password", zap.Error(hashErr))
		return fmt.Errorf("密码加密失败: %w", hashErr)
	}

	// 更新密码
	if updateErr := s.userRepo.UpdatePassword(ctx, userIDStr, hashedPassword); updateErr != nil {
		s.logger.Error("Failed to update password", zap.Error(updateErr))
		return fmt.Errorf("更新密码失败: %w", updateErr)
	}

	s.logger.Info("Password changed successfully", zap.String("user_id", userIDStr))
	return nil
}

// GetUserByID 根据 ID 获取用户
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*dto.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &dto.User{
		UserInDBBase: dto.UserInDBBase{
			UserBase: dto.UserBase{
				Name:        user.Name,
				Email:       user.Email,
				IsActive:    user.IsActive,
				IsSuperuser: user.IsSuperuser,
				Avatar:      user.Avatar,
			},
			ID: int(user.ID),
		},
	}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.Token, error) {
	s.logger.Info("Refresh token attempt")

	// 验证刷新令牌
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		s.logger.Warn("Invalid refresh token", zap.Error(err))
		return nil, fmt.Errorf("无效的刷新令牌: %w", err)
	}

	// 获取用户信息
	userIDStr := fmt.Sprintf("%d", claims.UserID)
	user, err := s.userRepo.GetByID(ctx, userIDStr)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 检查用户是否激活
	if !user.IsActive {
		s.logger.Warn("User is not active", zap.String("user_id", userIDStr))
		return nil, errors.New("用户已被禁用")
	}

	// 生成新的访问令牌
	roles := []string{"user"}
	if user.IsSuperuser {
		roles = append(roles, "admin")
	}
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Name, roles)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	s.logger.Info("Token refreshed successfully", zap.String("user_id", fmt.Sprintf("%d", claims.UserID)))

	return &dto.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		SuperUser:   user.IsSuperuser,
		UserID:      int(user.ID),
		UserName:    user.Name,
		Avatar:      user.Avatar,
	}, nil
}

// ValidateToken 验证令牌
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*security.Claims, error) {
	s.logger.Debug("Validate token attempt")

	// 解析令牌
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		s.logger.Warn("Invalid token", zap.Error(err))
		return nil, fmt.Errorf("无效的令牌: %w", err)
	}

	// TODO: 检查令牌是否在黑名单中

	s.logger.Debug("Token validated successfully", zap.String("user_id", fmt.Sprintf("%d", claims.UserID)))
	return claims, nil
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, token string) error {
	s.logger.Info("User logout attempt")

	// TODO: 实现令牌黑名单机制
	// 可以将令牌加入 Redis 黑名单,设置过期时间与令牌过期时间一致
	// 或者从令牌缓存中移除该令牌

	// 验证令牌是否有效
	if _, err := s.jwtManager.ValidateToken(token); err != nil {
		s.logger.Warn("Invalid token for logout", zap.Error(err))
		return fmt.Errorf("无效的令牌: %w", err)
	}

	s.logger.Info("User logged out successfully")
	return nil
}

// UpdateUser 更新用户信息
func (s *AuthService) UpdateUser(ctx context.Context, userID string, req *dto.UserUpdate) error {
	s.logger.Info("Update user attempt", zap.String("user_id", userID))

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 更新字段
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("更新用户失败: %w", err)
	}

	s.logger.Info("User updated successfully", zap.String("user_id", userID))
	return nil
}
