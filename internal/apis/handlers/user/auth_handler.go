package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/apis/middleware"
	userbiz "moviepilot-go/internal/business/services/user"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/pkg/logger"
)

// AuthHandler 认证 API 处理器
type AuthHandler struct {
	authService *userbiz.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *userbiz.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger.GetLogger(),
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Avatar string `json:"avatar"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} dto.User
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid register request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 创建用户
	userCreate := &dto.UserCreate{
		UserBase: dto.UserBase{
			Name:  req.Username,
			Email: req.Email,
		},
		Password: req.Password,
	}

	user, err := h.authService.Register(c.Request.Context(), userCreate)
	if err != nil {
		h.logger.Error("Register failed",
			zap.String("username", req.Username),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("User registered successfully",
		zap.String("username", req.Username))

	c.JSON(http.StatusCreated, user)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} dto.Token
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Username, req.Password, "")
	if err != nil {
		h.logger.Warn("Login failed",
			zap.String("username", req.Username),
			zap.Error(err))

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户名或密码错误",
		})
		return
	}

	h.logger.Info("User logged in successfully",
		zap.String("username", req.Username))

	c.JSON(http.StatusOK, token)
}

// GetProfile 获取个人资料
// @Summary 获取个人资料
// @Description 获取当前登录用户的个人资料
// @Tags 用户管理
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.User
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/users/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证",
		})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile",
			zap.String("user_id", userID),
			zap.Error(err))

		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile 更新个人资料
// @Summary 更新个人资料
// @Description 更新当前登录用户的个人资料
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "更新信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update profile request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	userUpdate := &dto.UserUpdate{
		UserBase: dto.UserBase{
			Email:  req.Email,
			Avatar: req.Avatar,
		},
	}

	if err := h.authService.UpdateUser(c.Request.Context(), userID, userUpdate); err != nil {
		h.logger.Error("Failed to update user profile",
			zap.String("user_id", userID),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("User profile updated successfully",
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户管理
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "密码信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDStr := middleware.GetUserID(c)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证",
		})
		return
	}

	// 将 userID 从 string 转换为 uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid user ID", zap.String("user_id", userIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户ID",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid change password request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), uint(userID), req.OldPassword, req.NewPassword); err != nil {
		h.logger.Warn("Failed to change password",
			zap.String("user_id", userIDStr),
			zap.Error(err))

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("Password changed successfully",
		zap.String("user_id", userIDStr))

	c.JSON(http.StatusOK, gin.H{
		"message": "密码修改成功",
	})
}
