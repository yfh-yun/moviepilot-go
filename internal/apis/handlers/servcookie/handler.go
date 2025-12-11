package servcookie

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/internal/infrastructure/config"
	"moviepilot-go/pkg/logger"
)

// Handler Servcookie API 处理器
type Handler struct {
	logger *zap.Logger
	config *config.Config
}

// NewHandler 创建 Servcookie 处理器
func NewHandler(config *config.Config) *Handler {
	return &Handler{
		logger: logger.GetLogger(),
		config: config,
	}
}

// CookieData Cookie数据结构
type CookieData struct {
	UUID      string `json:"uuid"`
	Encrypted string `json:"encrypted"`
}

// CookiePassword Cookie密码结构
type CookiePassword struct {
	Password string `json:"password"`
}

// GzipMiddleware Gzip压缩中间件
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是gzip压缩的请求，解压
		if c.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gzip request"})
				c.Abort()
				return
			}
			defer reader.Close()
			c.Request.Body = io.NopCloser(reader)
		}
		c.Next()
	}
}

// VerifyServerEnabled 验证CookieCloud服务是否启用
func VerifyServerEnabled(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现配置检查逻辑
		// if !config.CookieCloud.EnableLocal {
		//     c.JSON(http.StatusBadRequest, gin.H{"error": "本地CookieCloud服务器未启用"})
		//     c.Abort()
		//     return
		// }
		c.Next()
	}
}

// GetRoot 获取根路径
// @Summary 根路径
// @Description CookieCloud API根路径
// @Tags servcookie
// @Produce plain
// @Success 200 {string} string
// @Router /cookiecloud [get]
func (h *Handler) GetRoot(c *gin.Context) {
	h.logger.Debug("GetRoot called")
	c.String(http.StatusOK, "Hello MoviePilot! COOKIECLOUD API ROOT = /cookiecloud")
}

// PostRoot POST根路径
// @Summary POST根路径
// @Description CookieCloud API根路径
// @Tags servcookie
// @Produce plain
// @Success 200 {string} string
// @Router /cookiecloud [post]
func (h *Handler) PostRoot(c *gin.Context) {
	h.logger.Debug("PostRoot called")
	c.String(http.StatusOK, "Hello MoviePilot! COOKIECLOUD API ROOT = /cookiecloud")
}

// UpdateCookie 更新Cookie数据
// @Summary 上传Cookie数据
// @Description 上传Cookie数据
// @Tags servcookie
// @Accept json
// @Produce json
// @Param cookie body CookieData true "Cookie数据"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cookiecloud/update [post]
func (h *Handler) UpdateCookie(c *gin.Context) {
	h.logger.Debug("UpdateCookie called")

	var req CookieData
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: 获取配置中的Cookie路径
	cookiePath := "/config/cookies"
	if _, err := os.Stat(cookiePath); os.IsNotExist(err) {
		if err := os.MkdirAll(cookiePath, 0755); err != nil {
			h.logger.Error("Failed to create cookie directory", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cookie directory"})
			return
		}
	}

	filePath := filepath.Join(cookiePath, req.UUID+".json")
	content := map[string]string{
		"encrypted": req.Encrypted,
	}

	fileContent, err := json.Marshal(content)
	if err != nil {
		h.logger.Error("Failed to marshal cookie data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal cookie data"})
		return
	}

	if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
		h.logger.Error("Failed to write cookie data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write cookie data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"action": "done"})
}

// GetCookie 获取Cookie数据
// @Summary 下载Cookie数据
// @Description 下载Cookie数据
// @Tags servcookie
// @Produce json
// @Param uuid path string true "UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /cookiecloud/get/{uuid} [get]
func (h *Handler) GetCookie(c *gin.Context) {
	h.logger.Debug("GetCookie called")

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID is required"})
		return
	}

	// TODO: 获取配置中的Cookie路径
	cookiePath := "/config/cookies"
	filePath := filepath.Join(cookiePath, uuid+".json")

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		h.logger.Error("Failed to read cookie file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read cookie file"})
		return
	}

	var data map[string]string
	if err := json.Unmarshal(content, &data); err != nil {
		h.logger.Error("Failed to unmarshal cookie data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal cookie data"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// PostGetCookie POST获取Cookie数据
// @Summary POST下载Cookie数据
// @Description POST下载Cookie数据并解密
// @Tags servcookie
// @Accept json
// @Produce json
// @Param uuid path string true "UUID"
// @Param password body CookiePassword true "Cookie密码"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cookiecloud/get/{uuid} [post]
func (h *Handler) PostGetCookie(c *gin.Context) {
	h.logger.Debug("PostGetCookie called")

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID is required"})
		return
	}

	var req CookiePassword
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: 获取配置中的Cookie路径
	cookiePath := "/config/cookies"
	filePath := filepath.Join(cookiePath, uuid+".json")

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		h.logger.Error("Failed to read cookie file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read cookie file"})
		return
	}

	var data map[string]string
	if err := json.Unmarshal(content, &data); err != nil {
		h.logger.Error("Failed to unmarshal cookie data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal cookie data"})
		return
	}

	encrypted := data["encrypted"]
	if encrypted == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No encrypted data found"})
		return
	}

	// 解密Cookie数据
	decryptedData, err := decryptCookieData(uuid, req.Password, encrypted)
	if err != nil {
		h.logger.Error("Failed to decrypt cookie data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt cookie data"})
		return
	}

	c.JSON(http.StatusOK, decryptedData)
}

// decryptCookieData 解密Cookie数据
func decryptCookieData(uuid, password, encrypted string) (map[string]any, error) {
	// 简化实现，暂时返回空数据
	return map[string]any{"cookie_data": map[string]any{}}, nil
}
