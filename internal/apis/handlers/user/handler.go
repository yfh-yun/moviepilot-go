package user

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/apis/middlewares"
	userbiz "moviepilot-go/internal/business/services/user"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/pkg/logger"
)

// Handler 用户 API 处理器
type Handler struct {
	authService       *userbiz.AuthService
	permissionService userbiz.PermissionService
	userService       *userbiz.UserService
	logger            *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(
	authService *userbiz.AuthService,
	permissionService userbiz.PermissionService,
	userService *userbiz.UserService,
) *Handler {
	return &Handler{
		authService:       authService,
		permissionService: permissionService,
		userService:       userService,
		logger:            logger.GetLogger(),
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags user
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "登录凭证"
// @Success 200 {object} userbiz.AuthToken
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/user/login [post]
func (h *Handler) Login(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API Login started", zap.String("request_id", reqID))

	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("User API Login invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Username, req.Password, "")
	if err != nil {
		h.logger.Error("User API Login failed",
			zap.String("request_id", reqID),
			zap.String("username", req.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	h.logger.Info("User API Login succeeded",
		zap.String("request_id", reqID),
		zap.String("username", req.Username),
	)

	c.JSON(http.StatusOK, token)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，使令牌失效
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/user/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	reqID := c.GetString("request_id")
	token := c.GetHeader("Authorization")
	if token == "" {
		h.logger.Warn("User API Logout missing token",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供令牌"})
		return
	}

	// 移除 "Bearer " 前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		h.logger.Error("User API Logout failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("User API Logout succeeded", zap.String("request_id", reqID))

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags user
// @Accept json
// @Produce json
// @Param token body RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} userbiz.AuthToken
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/user/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	reqID := c.GetString("request_id")
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("User API RefreshToken failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的刷新令牌"})
		return
	}

	c.JSON(http.StatusOK, token)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param password body ChangePasswordRequest true "密码信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/user/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	reqID := c.GetString("request_id")
	var req ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := middlewares.GetUserID(c)
	if !ok || userID == 0 {
		h.logger.Warn("User API ChangePassword unauthorized",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		h.logger.Error("User API ChangePassword failed",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("User API ChangePassword succeeded",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// CheckPermission 检查权限
// @Summary 检查权限
// @Description 检查用户是否有指定资源的操作权限
// @Tags user
// @Security BearerAuth
// @Produce json
// @Param resource query string true "资源名称"
// @Param action query string true "操作名称"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/user/permission/check [get]
func (h *Handler) CheckPermission(c *gin.Context) {
	reqID := c.GetString("request_id")
	resource := c.Query("resource")
	action := c.Query("action")

	if resource == "" || action == "" {
		h.logger.Warn("User API CheckPermission missing params",
			zap.String("request_id", reqID),
			zap.String("resource", resource),
			zap.String("action", action),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "资源和操作不能为空"})
		return
	}

	userID, ok := middlewares.GetUserID(c)
	if !ok || userID == 0 {
		h.logger.Warn("User API CheckPermission unauthorized",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	hasPermission, err := h.permissionService.CheckPermission(c.Request.Context(), userID, resource, action)
	if err != nil {
		h.logger.Error("User API CheckPermission failed",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.String("resource", resource),
			zap.String("action", action),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{
			"has_permission": false,
			"message":        "没有权限",
		})
		return
	}

	h.logger.Info("User API CheckPermission succeeded",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action),
	)

	c.JSON(http.StatusOK, gin.H{
		"has_permission": true,
		"message":        "有权限",
	})
}

// GetPermissions 获取用户权限
// @Summary 获取用户权限
// @Description 获取当前用户的所有权限
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {array} userbiz.Permission
// @Failure 401 {object} map[string]interface{}
// @Router /api/user/permissions [get]
func (h *Handler) GetPermissions(c *gin.Context) {
	reqID := c.GetString("request_id")
	userID, ok := middlewares.GetUserID(c)
	if !ok || userID == 0 {
		h.logger.Warn("User API GetPermissions unauthorized",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	permissions, err := h.permissionService.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("User API GetPermissions failed",
			zap.String("request_id", reqID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("User API GetPermissions succeeded",
		zap.String("request_id", reqID),
		zap.Uint("user_id", userID),
	)

	c.JSON(http.StatusOK, permissions)
}

// ValidateToken 验证令牌
// @Summary 验证令牌
// @Description 验证访问令牌是否有效
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} userbiz.Claims
// @Failure 401 {object} map[string]interface{}
// @Router /api/user/validate [get]
func (h *Handler) ValidateToken(c *gin.Context) {
	reqID := c.GetString("request_id")
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供令牌"})
		return
	}

	// 移除 "Bearer " 前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims, err := h.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		h.logger.Error("User API ValidateToken failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
		return
	}

	h.logger.Info("User API ValidateToken succeeded", zap.String("request_id", reqID))

	c.JSON(http.StatusOK, claims)
}

// ListUsers 获取所有用户列表
// @Summary 所有用户
// @Description 查询用户列表
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {array} userbiz.User
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user [get]
func (h *Handler) ListUsers(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API ListUsers started", zap.String("request_id", reqID))

	users, err := h.userService.ListUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("User API ListUsers failed",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("User API ListUsers succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_count", len(users)),
	)

	c.JSON(http.StatusOK, users)
}

// CreateUser 新增用户
// @Summary 新增用户
// @Description 新增用户
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body dto.UserCreate true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user [post]
func (h *Handler) CreateUser(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API CreateUser started", zap.String("request_id", reqID))

	var req dto.UserCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("User API CreateUser invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("User API CreateUser failed",
			zap.String("request_id", reqID),
			zap.String("username", req.Name),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API CreateUser succeeded",
		zap.String("request_id", reqID),
		zap.String("username", user.Name),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body dto.UserUpdate true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API UpdateUser started", zap.String("request_id", reqID))

	var req dto.UserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("User API UpdateUser invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateUser(c.Request.Context(), &req); err != nil {
		h.logger.Error("User API UpdateUser failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", req.ID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API UpdateUser succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_id", req.ID),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetCurrentUser 获取当前登录用户信息
// @Summary 当前登录用户信息
// @Description 获取当前登录用户信息
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} userbiz.User
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/current [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API GetCurrentUser started", zap.String("request_id", reqID))

	// 从上下文中获取用户ID
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User API GetCurrentUser missing user_id in context",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID.(int))
	if err != nil {
		h.logger.Error("User API GetCurrentUser failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("User API GetCurrentUser succeeded",
		zap.String("request_id", reqID),
		zap.String("username", user.Name),
	)

	c.JSON(http.StatusOK, user)
}

// UploadAvatar 上传用户头像
// @Summary 上传用户头像
// @Description 上传用户头像
// @Tags user
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param user_id path int true "用户ID"
// @Param file formData file true "头像文件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/avatar/{user_id} [post]
func (h *Handler) UploadAvatar(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API UploadAvatar started", zap.String("request_id", reqID))

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.logger.Warn("User API UploadAvatar invalid user_id",
			zap.String("request_id", reqID),
			zap.String("user_id", userIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Warn("User API UploadAvatar failed to get file",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败"})
		return
	}
	defer file.Close()

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.logger.Warn("User API UploadAvatar failed to read file",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取文件失败"})
		return
	}

	// 转换为Base64
	fileBase64 := base64.StdEncoding.EncodeToString(fileBytes)
	avatar := fmt.Sprintf("data:image/ico;base64,%s", fileBase64)

	// 更新用户头像
	if err := h.userService.UpdateAvatar(c.Request.Context(), userID, avatar); err != nil {
		h.logger.Error("User API UploadAvatar failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API UploadAvatar succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_id", userID),
		zap.String("filename", header.Filename),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": header.Filename})
}

// GenerateOTP 生成OTP验证URI
// @Summary 生成OTP验证URI
// @Description 生成OTP验证URI
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/otp/generate [post]
func (h *Handler) GenerateOTP(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API GenerateOTP started", zap.String("request_id", reqID))

	// 从上下文中获取用户ID
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User API GenerateOTP missing user_id in context",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	secret, uri, err := h.userService.GenerateOTPURI(c.Request.Context(), userID.(int))
	if err != nil {
		h.logger.Error("User API GenerateOTP failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API GenerateOTP succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_id", userID.(int)),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"secret": secret, "uri": uri}})
}

// VerifyOTP 判断OTP验证是否通过
// @Summary 判断OTP验证是否通过
// @Description 判断OTP验证是否通过
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body map[string]string true "OTP验证数据"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/otp/judge [post]
func (h *Handler) VerifyOTP(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API VerifyOTP started", zap.String("request_id", reqID))

	// 从上下文中获取用户ID
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User API VerifyOTP missing user_id in context",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("User API VerifyOTP invalid request",
			zap.String("request_id", reqID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	otpCode := req["otpPassword"]
	if otpCode == "" {
		h.logger.Warn("User API VerifyOTP missing otpPassword",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少OTP验证码"})
		return
	}

	// 验证OTP
	isValid, err := h.userService.VerifyOTP(c.Request.Context(), userID.(int), otpCode)
	if err != nil {
		h.logger.Error("User API VerifyOTP failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	if !isValid {
		h.logger.Warn("User API VerifyOTP failed: invalid code",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
		)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "验证码错误"})
		return
	}

	// 启用OTP
	if _, err := h.userService.EnableOTP(c.Request.Context(), userID.(int)); err != nil {
		h.logger.Error("User API VerifyOTP failed to enable OTP",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API VerifyOTP succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_id", userID.(int)),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DisableOTP 关闭当前用户的OTP验证
// @Summary 关闭当前用户的OTP验证
// @Description 关闭当前用户的OTP验证
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/otp/disable [post]
func (h *Handler) DisableOTP(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API DisableOTP started", zap.String("request_id", reqID))

	// 从上下文中获取用户ID
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User API DisableOTP missing user_id in context",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	if err := h.userService.DisableOTP(c.Request.Context(), userID.(int)); err != nil {
		h.logger.Error("User API DisableOTP failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API DisableOTP succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_id", userID.(int)),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// IsOTPEnabled 判断当前用户是否开启OTP验证
// @Summary 判断当前用户是否开启OTP验证
// @Description 判断当前用户是否开启OTP验证
// @Tags user
// @Security BearerAuth
// @Produce json
// @Param userid path string true "用户名"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/otp/{userid} [get]
func (h *Handler) IsOTPEnabled(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API IsOTPEnabled started", zap.String("request_id", reqID))

	userID := c.Param("userid")

	user, err := h.userService.GetUserByUsername(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("User API IsOTPEnabled failed",
			zap.String("request_id", reqID),
			zap.String("username", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	if user == nil {
		h.logger.Warn("User API IsOTPEnabled user not found",
			zap.String("request_id", reqID),
			zap.String("username", userID),
		)
		c.JSON(http.StatusOK, gin.H{"success": false})
		return
	}

	h.logger.Info("User API IsOTPEnabled succeeded",
		zap.String("request_id", reqID),
		zap.String("username", userID),
		zap.Bool("is_otp_enabled", user.IsOTP),
	)

	c.JSON(http.StatusOK, gin.H{"success": user.IsOTP})
}

// GetUserConfig 查询用户配置
// @Summary 查询用户配置
// @Description 查询用户配置
// @Tags user
// @Security BearerAuth
// @Produce json
// @Param key path string true "配置键"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/config/{key} [get]
func (h *Handler) GetUserConfig(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API GetUserConfig started", zap.String("request_id", reqID))

	// 从上下文中获取用户ID
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User API GetUserConfig missing user_id in context",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	key := c.Param("key")

	value, err := h.userService.GetUserConfig(c.Request.Context(), userID.(int), key)
	if err != nil {
		h.logger.Error("User API GetUserConfig failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.String("key", key),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API GetUserConfig succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_id", userID.(int)),
		zap.String("key", key),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"value": value}})
}

// SetUserConfig 更新用户配置
// @Summary 更新用户配置
// @Description 更新用户配置
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path string true "配置键"
// @Param value body interface{} true "配置值"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/config/{key} [post]
func (h *Handler) SetUserConfig(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API SetUserConfig started", zap.String("request_id", reqID))

	// 从上下文中获取用户ID
	userID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User API SetUserConfig missing user_id in context",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	key := c.Param("key")

	var value any
	if err := c.ShouldBindJSON(&value); err != nil {
		h.logger.Warn("User API SetUserConfig invalid request",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.String("key", key),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 将value转换为字符串
	valueStr, err := json.Marshal(value)
	if err != nil {
		h.logger.Warn("User API SetUserConfig failed to marshal value",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.String("key", key),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置值"})
		return
	}

	if err := h.userService.SetUserConfig(c.Request.Context(), userID.(int), key, string(valueStr)); err != nil {
		h.logger.Error("User API SetUserConfig failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID.(int)),
			zap.String("key", key),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API SetUserConfig succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_id", userID.(int)),
		zap.String("key", key),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteUserByID 通过ID删除用户
// @Summary 通过ID删除用户
// @Description 通过ID删除用户
// @Tags user
// @Security BearerAuth
// @Produce json
// @Param user_id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/id/{user_id} [delete]
func (h *Handler) DeleteUserByID(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API DeleteUserByID started", zap.String("request_id", reqID))

	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.logger.Warn("User API DeleteUserByID invalid user_id",
			zap.String("request_id", reqID),
			zap.String("user_id", userIDStr),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		h.logger.Error("User API DeleteUserByID failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API DeleteUserByID succeeded",
		zap.String("request_id", reqID),
		zap.Int("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteUserByName 通过用户名删除用户
// @Summary 通过用户名删除用户
// @Description 通过用户名删除用户
// @Tags user
// @Security BearerAuth
// @Produce json
// @Param user_name path string true "用户名"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/name/{user_name} [delete]
func (h *Handler) DeleteUserByName(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API DeleteUserByName started", zap.String("request_id", reqID))

	userName := c.Param("user_name")

	user, err := h.userService.GetUserByUsername(c.Request.Context(), userName)
	if err != nil {
		h.logger.Error("User API DeleteUserByName failed to get user",
			zap.String("request_id", reqID),
			zap.String("username", userName),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	if user == nil {
		h.logger.Warn("User API DeleteUserByName user not found",
			zap.String("request_id", reqID),
			zap.String("username", userName),
		)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), user.ID); err != nil {
		h.logger.Error("User API DeleteUserByName failed",
			zap.String("request_id", reqID),
			zap.Int("user_id", user.ID),
			zap.String("username", userName),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	h.logger.Info("User API DeleteUserByName succeeded",
		zap.String("request_id", reqID),
		zap.String("username", userName),
	)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetUserDetail 获取用户详情
// @Summary 用户详情
// @Description 查询用户详情
// @Tags user
// @Security BearerAuth
// @Produce json
// @Param username path string true "用户名"
// @Success 200 {object} userbiz.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/user/{username} [get]
func (h *Handler) GetUserDetail(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("User API GetUserDetail started", zap.String("request_id", reqID))

	username := c.Param("username")

	user, err := h.userService.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("User API GetUserDetail failed",
			zap.String("request_id", reqID),
			zap.String("username", username),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		h.logger.Warn("User API GetUserDetail user not found",
			zap.String("request_id", reqID),
			zap.String("username", username),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 从上下文中获取当前用户ID
	currentUserID, ok := c.Get("user_id")
	if !ok {
		h.logger.Warn("User API GetUserDetail missing current user_id in context",
			zap.String("request_id", reqID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 检查权限：只有超级管理员或用户本人可以查看用户详情
	if currentUserID.(int) != user.ID {
		// TODO: 检查当前用户是否为超级管理员
		// 暂时简单实现，只允许用户本人查看
		h.logger.Warn("User API GetUserDetail permission denied",
			zap.String("request_id", reqID),
			zap.Int("current_user_id", currentUserID.(int)),
			zap.Int("target_user_id", user.ID),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "用户权限不足"})
		return
	}

	h.logger.Info("User API GetUserDetail succeeded",
		zap.String("request_id", reqID),
		zap.String("username", username),
	)

	c.JSON(http.StatusOK, user)
}

// LoginRequest 登录请求

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新令牌请求

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
