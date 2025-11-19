// Package user 用户管理API处理器
package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/service/user"
	"github.com/yfh-yun/moviepilot-go/pkg/jwt"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService user.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UpdateProfileRequest 更新用户信息请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50,alphanum"`
	Email    string `json:"email" binding:"omitempty,email"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
}

// UpdateSettingsRequest 更新用户设置请求
type UpdateSettingsRequest struct {
	Settings map[string]interface{} `json:"settings" binding:"required"`
}

// UpdatePermissionsRequest 更新用户权限请求
type UpdatePermissionsRequest struct {
	Permissions map[string]bool `json:"permissions" binding:"required"`
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// GetUserActivityRequest 用户活动记录请求
type GetUserActivityRequest struct {
	Days int `form:"days" binding:"omitempty,min=1,max=30"`
}

// GetUserProfile 获取当前用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=models.User}
// @Failure 401 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	// 调用服务层获取用户信息
	userProfile, err := h.userService.GetUserProfile(c.Request.Context(), userID.(uint))
	if err != nil {
		logger.Error("get user profile failed", zap.Uint("user_id", userID.(uint)), zap.Error(err))
		response.InternalError(c, "get user profile failed")
		return
	}

	response.Success(c, userProfile)
}

// UpdateUserProfile 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前登录用户的基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "用户信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/user/profile [put]
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req UpdateProfileRequest

	// 绑定并验证参数
	if err := validator.BindJSONAndValidate(c, &req); err != nil {
		logger.Error("bind update profile params failed", zap.Error(err))
		validator.HandleValidationError(c, err)
		return
	}

	// 调用服务层更新用户信息
	if err := h.userService.UpdateUserProfile(c.Request.Context(), userID.(uint), req.Username, req.Email, req.Avatar); err != nil {
		logger.Error("update user profile failed", zap.Uint("user_id", userID.(uint)), zap.Error(err))
		response.Error(c, response.CodeOperationFailed, err.Error())
		return
	}

	logger.Info("update user profile success", zap.Uint("user_id", userID.(uint)))
	response.Success(c, nil)
}

// GetUserSettings 获取用户设置
// @Summary 获取用户设置
// @Description 获取当前登录用户的个性化设置
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 401 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/user/settings [get]
func (h *UserHandler) GetUserSettings(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	// 调用服务层获取用户设置
	settings, err := h.userService.GetUserSettings(c.Request.Context(), userID.(uint))
	if err != nil {
		logger.Error("get user settings failed", zap.Uint("user_id", userID.(uint)), zap.Error(err))
		response.InternalError(c, "get user settings failed")
		return
	}

	response.Success(c, settings)
}

// UpdateUserSettings 更新用户设置
// @Summary 更新用户设置
// @Description 更新当前登录用户的个性化设置
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body UpdateSettingsRequest true "用户设置"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/user/settings [put]
func (h *UserHandler) UpdateUserSettings(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req UpdateSettingsRequest

	// 绑定并验证参数
	if err := validator.BindJSONAndValidate(c, &req); err != nil {
		logger.Error("bind update settings params failed", zap.Error(err))
		validator.HandleValidationError(c, err)
		return
	}

	// 调用服务层更新用户设置
	if err := h.userService.UpdateUserSettings(c.Request.Context(), userID.(uint), req.Settings); err != nil {
		logger.Error("update user settings failed", zap.Uint("user_id", userID.(uint)), zap.Error(err))
		response.Error(c, response.CodeOperationFailed, err.Error())
		return
	}

	logger.Info("update user settings success", zap.Uint("user_id", userID.(uint)))
	response.Success(c, nil)
}

// GetUserStats 获取用户统计信息
// @Summary 获取用户统计
// @Description 获取当前登录用户的统计信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=user.UserStats}
// @Failure 401 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/user/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	// 调用服务层获取用户统计
	stats, err := h.userService.GetUserStats(c.Request.Context(), userID.(uint))
	if err != nil {
		logger.Error("get user stats failed", zap.Uint("user_id", userID.(uint)), zap.Error(err))
		response.InternalError(c, "get user stats failed")
		return
	}

	response.Success(c, stats)
}

// GetUserActivity 获取用户活动记录
// @Summary 获取用户活动
// @Description 获取当前登录用户的活动记录
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request query GetUserActivityRequest true "活动记录查询"
// @Success 200 {object} response.Response{data=[]user.UserActivity}
// @Failure 401 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/user/activity [get]
func (h *UserHandler) GetUserActivity(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req GetUserActivityRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("bind get user activity params failed", zap.Error(err))
		validator.HandleValidationError(c, err)
		return
	}

	// 设置默认值
	if req.Days == 0 {
		req.Days = 7
	}

	// 调用服务层获取用户活动
	activities, err := h.userService.GetUserActivity(c.Request.Context(), userID.(uint), req.Days)
	if err != nil {
		logger.Error("get user activity failed", zap.Uint("user_id", userID.(uint)), zap.Error(err))
		response.InternalError(c, "get user activity failed")
		return
	}

	response.Success(c, activities)
}

// ListUsers 获取用户列表（管理员）
// @Summary 获取用户列表
// @Description 管理员获取用户列表，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request query ListUsersRequest true "分页参数"
// @Success 200 {object} response.Response{data=[]models.User}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 检查管理员权限
	if !h.checkAdminPermission(c) {
		return
	}

	var req ListUsersRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("bind list users params failed", zap.Error(err))
		validator.HandleValidationError(c, err)
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	// 计算offset
	offset := (req.Page - 1) * req.Limit

	// 调用服务层获取用户列表
	users, total, err := h.userService.ListUsers(c.Request.Context(), offset, req.Limit)
	if err != nil {
		logger.Error("list users failed", zap.Error(err))
		response.InternalError(c, "list users failed")
		return
	}

	response.SuccessWithPagination(c, users, req.Page, req.Limit, total)
}

// EnableUser 启用用户（管理员）
// @Summary 启用用户
// @Description 管理员启用被禁用的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/admin/users/{user_id}/enable [post]
func (h *UserHandler) EnableUser(c *gin.Context) {
	// 检查管理员权限
	adminID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if !h.checkAdminPermission(c) {
		return
	}

	// 获取目标用户ID
	userID, err := validator.GetUintParam(c, "user_id")
	if err != nil {
		logger.Error("get user_id param failed", zap.Error(err))
		response.BadRequest(c, "invalid user_id")
		return
	}

	// 调用服务层启用用户
	if err := h.userService.EnableUser(c.Request.Context(), adminID.(uint), userID); err != nil {
		logger.Error("enable user failed", zap.Uint("user_id", userID), zap.Error(err))
		response.Error(c, response.CodeOperationFailed, err.Error())
		return
	}

	logger.Info("enable user success", zap.Uint("user_id", userID))
	response.Success(c, nil)
}

// DisableUser 禁用用户（管理员）
// @Summary 禁用用户
// @Description 管理员禁用用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/admin/users/{user_id}/disable [post]
func (h *UserHandler) DisableUser(c *gin.Context) {
	// 检查管理员权限
	adminID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if !h.checkAdminPermission(c) {
		return
	}

	// 获取目标用户ID
	userID, err := validator.GetUintParam(c, "user_id")
	if err != nil {
		logger.Error("get user_id param failed", zap.Error(err))
		response.BadRequest(c, "invalid user_id")
		return
	}

	// 调用服务层禁用用户
	if err := h.userService.DisableUser(c.Request.Context(), adminID.(uint), userID); err != nil {
		logger.Error("disable user failed", zap.Uint("user_id", userID), zap.Error(err))
		response.Error(c, response.CodeOperationFailed, err.Error())
		return
	}

	logger.Info("disable user success", zap.Uint("user_id", userID))
	response.Success(c, nil)
}

// checkAdminPermission 检查管理员权限
func (h *UserHandler) checkAdminPermission(c *gin.Context) bool {
	// 从上下文获取Token
	token, exists := c.Get("token")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return false
	}

	// 解析Token检查管理员权限
	claims, err := jwt.ParseToken(token.(string))
	if err != nil {
		response.Unauthorized(c, "invalid token")
		return false
	}

	// 检查角色权限
	if claims.Role != "admin" && claims.Role != "superadmin" {
		response.Forbidden(c, "insufficient permissions")
		return false
	}

	return true
}
