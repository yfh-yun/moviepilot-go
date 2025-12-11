package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/apis/middlewares"
	authservice "moviepilot-go/internal/business/services/auth"
	"moviepilot-go/pkg/logger"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService authservice.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService authservice.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger.GetLogger(),
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body authservice.RegisterRequest true "注册信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Auth v1 Register started", zap.String("request_id", reqID))

	var req authservice.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Auth v1 Register invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Auth v1 Register failed",
			zap.String("request_id", reqID),
			zap.String("username", req.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("Auth v1 Register succeeded",
		zap.String("request_id", reqID),
		zap.String("username", req.Username),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"user":    user,
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body authservice.LoginRequest true "登录信息"
// @Success 200 {object} authservice.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Auth v1 Login started",
		zap.String("request_id", reqID),
		zap.String("client_ip", c.ClientIP()),
	)

	var req authservice.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Auth v1 Login invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取客户端IP
	ip := c.ClientIP()

	resp, err := h.authService.Login(c.Request.Context(), &req, ip)
	if err != nil {
		h.logger.Error("Auth v1 Login failed",
			zap.String("request_id", reqID),
			zap.String("username", req.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("Auth v1 Login succeeded",
		zap.String("request_id", reqID),
		zap.String("username", req.Username),
	)

	c.JSON(http.StatusOK, resp)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出
// @Tags 认证
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	h.logger.Debug("Auth v1 Logout started",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		h.logger.Error("Auth v1 Logout failed",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("Auth v1 Logout succeeded",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body map[string]string true "刷新令牌"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Auth v1 RefreshToken started", zap.String("request_id", reqID))

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	accessToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("Auth v1 RefreshToken failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_in":   900,
	})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags 认证
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body authservice.ChangePasswordRequest true "密码信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	h.logger.Debug("Auth v1 ChangePassword started",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	var req authservice.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		h.logger.Error("Auth v1 ChangePassword failed",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密码修改成功",
	})
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户
// @Description 获取当前登录用户信息
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, _ := middlewares.GetUserID(c)
	h.logger.Debug("Auth v1 GetCurrentUser started",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Auth v1 GetCurrentUser failed",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("Auth v1 GetCurrentUser succeeded",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
