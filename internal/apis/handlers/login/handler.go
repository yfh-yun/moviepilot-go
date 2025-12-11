package login

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/business/services/user"
	"moviepilot-go/internal/models/dto"
	"moviepilot-go/pkg/utils"
)

// Handler Login处理器
type Handler struct {
	authService     *user.AuthService
	logger          *zap.Logger
	wallpaperHelper *utils.WallpaperHelper
}

// NewHandler 创建Login处理器
func NewHandler(authService *user.AuthService, logger *zap.Logger, wallpaperHelper *utils.WallpaperHelper) *Handler {
	return &Handler{
		authService:     authService,
		logger:          logger,
		wallpaperHelper: wallpaperHelper,
	}
}

// Login 用户登录
// @Summary 用户登录获取Token
// @Description 用户登录认证，返回访问Token
// @Tags login
// @Accept json,form-data
// @Produce json
// @Param username formData string true "用户名"
// @Param password formData string true "密码"
// @Param otp_password formData string false "MFA验证码"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/login/access-token [post]
func (h *Handler) Login(c *gin.Context) {
	// 支持JSON和Form-Data两种格式
	var req dto.LoginRequest

	// 首先尝试绑定JSON格式
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果JSON绑定失败，尝试绑定Form-Data格式
		req.Username = c.PostForm("username")
		req.Password = c.PostForm("password")
		req.OtpPassword = c.PostForm("otp_password")
	}

	// 验证必填字段
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	// 用户认证
	token, err := h.authService.Login(c.Request.Context(), req.Username, req.Password, req.OtpPassword)
	if err != nil {
		h.logger.Warn("登录失败",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("用户登录成功",
		zap.String("username", req.Username),
		zap.Int("user_id", token.UserID),
	)

	c.JSON(http.StatusOK, token)
}

// TestToken 测试Token有效性
// @Summary 测试Token有效性
// @Description 验证Token是否有效
// @Tags login
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.TestTokenResponse
// @Failure 401 {object} map[string]interface{}
// @Router /api/login/test-token [post]
func (h *Handler) TestToken(c *gin.Context) {
	// 从请求头获取Token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供Token"})
		return
	}

	// 验证Token
	claims, err := h.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		h.logger.Warn("Token验证失败", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "user_id": claims.UserID})
}

// RefreshToken 刷新Token
// @Summary 刷新访问Token
// @Description 使用刷新Token获取新的访问Token
// @Tags login
// @Accept json
// @Produce json
// @Param refresh body dto.RefreshTokenRequest true "刷新Token"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/login/refresh-token [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 刷新Token
	token, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Warn("Token刷新失败", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Token刷新成功", zap.Int("user_id", token.UserID))

	c.JSON(http.StatusOK, token)
}

// GetWallpaper 获取登录页面壁纸
// @Summary 获取登录页面壁纸
// @Description 获取随机电影海报作为登录页面壁纸
// @Tags login
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/login/wallpaper [get]
func (h *Handler) GetWallpaper(c *gin.Context) {
	url := ""
	if h.wallpaperHelper != nil {
		url = h.wallpaperHelper.GetWallpaper()
	}
	if url == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": url,
	})
}

// GetWallpapers 获取登录页面壁纸列表
// @Summary 获取登录页面壁纸列表
// @Description 获取所有可用的登录页面壁纸
// @Tags login
// @Produce json
// @Success 200 {array} string
// @Router /api/login/wallpapers [get]
func (h *Handler) GetWallpapers(c *gin.Context) {
	wallpapers := []string{}
	if h.wallpaperHelper != nil {
		wallpapers = h.wallpaperHelper.GetWallpapers(0)
	}
	c.JSON(http.StatusOK, wallpapers)
}
