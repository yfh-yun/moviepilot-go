// Package auth 认证API处理器
// 重构合并后的统一认证处理器，消除代码冗余
package auth

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/business/services"
	"moviepilot-go/pkg/response"
	"moviepilot-go/pkg/validator"
)

// AuthHandler 统一认证处理器
type AuthHandler struct {
	authService service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger.Logger,
	}
}

// === 请求结构体定义 ===

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" binding:"required,username"` // 3-20位字母数字下划线
	Password    string `json:"password" binding:"required,password"` // 至少8位，包含字母数字特殊字符
	OTPPassword string `json:"otp_password,omitempty"`               // 双因素认证码（可选）
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"required,email"`
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// GetWallpaperRequest 获取壁纸请求
type GetWallpaperRequest struct {
	Type   string `form:"type" binding:"omitempty,oneof=random daily weekly"` // 壁纸类型
	Width  int    `form:"width" binding:"omitempty,min=100,max=3840"`         // 壁纸宽度
	Height int    `form:"height" binding:"omitempty,min=100,max=2160"`        // 壁纸高度
}

// === API方法实现 ===

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌，支持双因素认证
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求参数"
// @Success 200 {object} response.Response{data=jwt.TokenPair}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "认证失败"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := validator.BindAndValidate(c, &req); err != nil {
		return
	}

	tokenPair, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.handleAuthError(c, err, req.Username)
		return
	}

	h.logger.Info("用户登录成功",
		zap.String("username", req.Username),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	response.Success(c, tokenPair)
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=models.User}
// @Failure 400 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := validator.BindAndValidate(c, &req); err != nil {
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Username, req.Password, req.Email)
	if err != nil {
		h.logger.Error("注册失败",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		h.handleAuthError(c, err, req.Username)
		return
	}

	h.logger.Info("用户注册成功", zap.String("username", req.Username))
	response.Success(c, user)
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} response.Response{data=jwt.TokenPair}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "令牌无效"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	if err := validator.BindAndValidate(c, &req); err != nil {
		return
	}

	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("令牌刷新失败",
			zap.Error(err),
			zap.String("ip", c.ClientIP()),
		)
		response.Unauthorized(c, "令牌刷新失败")
		return
	}

	response.Success(c, tokenPair)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，使访问令牌和刷新令牌失效
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "登出请求参数"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest

	if err := validator.BindAndValidate(c, &req); err != nil {
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		// 如果没有用户ID，只记录刷新令牌的登出
		h.logger.Info("令牌登出（无用户ID）",
			zap.String("ip", c.ClientIP()),
		)
	} else {
		h.authService.Logout(c.Request.Context(), userID.(uint))
	}

	h.logger.Info("用户登出成功",
		zap.Any("user_id", userID),
		zap.String("ip", c.ClientIP()),
	)

	response.Success(c, nil)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 用户修改自己的密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	var req ChangePasswordRequest

	if err := validator.BindAndValidate(c, &req); err != nil {
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		h.logger.Error("修改密码失败",
			zap.Error(err),
			zap.Any("user_id", userID),
		)
		if err.Error() == "invalid old password" {
			response.Error(c, response.CodeInvalidPassword, "旧密码错误")
		} else {
			response.ServerError(c, "密码修改失败")
		}
		return
	}

	h.logger.Info("密码修改成功", zap.Any("user_id", userID))
	response.Success(c, nil)
}

// GetUserInfo 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.User}
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/auth/userinfo [get]
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}

	// 从Authorization header获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "缺少认证令牌")
		return
	}

	// 移除 "Bearer " 前缀
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	userInfo, err := h.authService.GetUserByToken(c.Request.Context(), token)
	if err != nil {
		h.logger.Error("获取用户信息失败",
			zap.Error(err),
			zap.Any("user_id", userID),
		)
		response.ServerError(c, "获取用户信息失败")
		return
	}

	response.Success(c, userInfo)
}

// GetWallpaper 获取壁纸
// @Summary 获取壁纸
// @Description 获取随机壁纸或指定类型的壁纸
// @Tags 认证
// @Accept json
// @Produce json
// @Param type query string false "壁纸类型: random(随机), daily(每日), weekly(每周)"
// @Param width query int false "壁纸宽度" minimum(100) maximum(3840)
// @Param height query int false "壁纸高度" minimum(100) maximum(2160)
// @Success 200 {object} response.Response{data=[]models.Wallpaper}
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/auth/wallpaper [get]
func (h *AuthHandler) GetWallpaper(c *gin.Context) {
	var req GetWallpaperRequest

	if err := validator.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	// 暂时返回空数组，因为壁纸服务尚未实现
	h.logger.Info("获取壁纸请求",
		zap.String("type", req.Type),
		zap.Int("width", req.Width),
		zap.Int("height", req.Height),
	)
	response.Success(c, []interface{}{})
}

// RegisterRoutes 注册认证相关路由
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
		authGroup.POST("/register", h.Register)
		authGroup.POST("/logout", h.Logout)
		authGroup.POST("/refresh", h.RefreshToken)
		authGroup.GET("/wallpaper", h.GetWallpaper)

		protected := authGroup.Group("")
		{
			protected.GET("/userinfo", h.GetUserInfo)
			protected.POST("/change-password", h.ChangePassword)
		}
	}
}

// === 辅助函数 ===

// handleAuthError 统一处理认证错误
func (h *AuthHandler) handleAuthError(c *gin.Context, err error, username string) {
	switch {
	case err.Error() == "invalid password" || err.Error() == "user not found: ":
		h.logger.Warn("用户认证失败",
			zap.String("username", username),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
		)
		response.Unauthorized(c, "用户名或密码错误")

	case err.Error() == "username already exists":
		h.logger.Warn("用户已存在",
			zap.String("username", username),
			zap.String("ip", c.ClientIP()),
		)
		response.Error(c, response.CodeUserExists, "用户已存在")

	case err.Error() == "user is disabled":
		h.logger.Warn("用户已被禁用",
			zap.String("username", username),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
		)
		response.Error(c, response.CodeUserDisabled, "用户已被禁用")

	default:
		h.logger.Error("认证服务错误",
			zap.String("username", username),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
		)
		response.ServerError(c, "认证失败")
	}
}
