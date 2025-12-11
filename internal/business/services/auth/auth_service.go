package auth

import (
	"context"
	"fmt"
	"time"

	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"
	"moviepilot-go/pkg/security"
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *models.User `json:"user"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// AuthService 认证服务接口
type AuthService interface {
	// Register 用户注册
	Register(ctx context.Context, req *RegisterRequest) (*models.User, error)
	// Login 用户登录
	Login(ctx context.Context, req *LoginRequest, ip string) (*LoginResponse, error)
	// Logout 用户登出
	Logout(ctx context.Context, userID uint) error
	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error
	// GetCurrentUser 获取当前用户
	GetCurrentUser(ctx context.Context, userID uint) (*models.User, error)
}

// authService 认证服务实现
type authService struct {
	userRepo        repositories.UserRepository
	roleRepo        repositories.RoleRepository
	jwtManager      security.JWTManager
	passwordManager security.PasswordManager
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	jwtManager security.JWTManager,
	passwordManager security.PasswordManager,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
	}
}

// Register 用户注册
func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
	// 检查用户名是否已存在
	existingUser, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, _ = s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Nickname:     req.Nickname,
		Status:       "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 分配默认角色（user）
	defaultRole, err := s.roleRepo.GetByName(ctx, "user")
	if err == nil && defaultRole != nil {
		user.Roles = []models.Role{*defaultRole}
	}

	return user, nil
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req *LoginRequest, ip string) (*LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 验证密码
	if err := s.passwordManager.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 获取用户角色
	userWithRoles, err := s.userRepo.GetWithRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %w", err)
	}

	// 提取角色名称
	roleNames := make([]string, len(userWithRoles.Roles))
	for i, role := range userWithRoles.Roles {
		roleNames[i] = role.Name
	}

	// 生成访问令牌
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Username, roleNames)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ip); err != nil {
		// 记录日志但不影响登录
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15分钟
		User:         user,
	}, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, userID uint) error {
	// 这里可以实现 token 黑名单机制
	// 目前简单返回成功
	return nil
}

// RefreshToken 刷新令牌
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// 使用 JWT Manager 刷新令牌
	newAccessToken, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("刷新令牌失败: %w", err)
	}

	return newAccessToken, nil
}

// ChangePassword 修改密码
func (s *authService) ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	if err := s.passwordManager.VerifyPassword(user.PasswordHash, req.OldPassword); err != nil {
		return fmt.Errorf("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := s.passwordManager.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("新密码加密失败: %w", err)
	}

	// 更新密码
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// GetCurrentUser 获取当前用户
func (s *authService) GetCurrentUser(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.userRepo.GetWithRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	return user, nil
}
