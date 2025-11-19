// Package webhook Webhook API处理器模块
package webhook

import (
	"net/http"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/service/webhook"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler Webhook API处理器
type Handler struct {
	service webhook.Service
	logger  *logger.Logger
}

// NewHandler 创建新的Webhook处理器
func NewHandler(service webhook.Service, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	webhookGroup := router.Group("/webhook")
	{
		webhookGroup.GET("/", h.GetWebhookList)
		webhookGroup.POST("/", h.CreateWebhook)
		webhookGroup.PUT("/:id", h.UpdateWebhook)
		webhookGroup.DELETE("/:id", h.DeleteWebhook)
		webhookGroup.POST("/test", h.TestWebhook)
	}
}

// GetWebhookList 获取Webhook列表
// @Summary 获取Webhook列表
// @Description 获取Webhook配置列表
// @Tags Webhook
// @Produce json
// @Success 200 {object} Response{data=[]WebhookInfo}
// @Router /webhook [get]
func (h *Handler) GetWebhookList(c *gin.Context) {
	webhooks, err := h.service.GetWebhookList()
	if err != nil {
		h.logger.Error("Failed to get webhook list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取Webhook列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    webhooks,
	})
}

// CreateWebhook 创建Webhook
// @Summary 创建Webhook
// @Description 创建新的Webhook配置
// @Tags Webhook
// @Produce json
// @Param webhook body WebhookConfig true "Webhook配置"
// @Success 200 {object} Response{data=WebhookInfo}
// @Router /webhook [post]
func (h *Handler) CreateWebhook(c *gin.Context) {
	var config WebhookConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	webhookInfo, err := h.service.CreateWebhook(config)
	if err != nil {
		h.logger.Error("Failed to create webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建Webhook失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    webhookInfo,
		"message": "Webhook创建成功",
	})
}

// UpdateWebhook 更新Webhook
// @Summary 更新Webhook
// @Description 更新Webhook配置
// @Tags Webhook
// @Produce json
// @Param id path string true "Webhook ID"
// @Param webhook body WebhookConfig true "Webhook配置"
// @Success 200 {object} Response{data=WebhookInfo}
// @Router /webhook/{id} [put]
func (h *Handler) UpdateWebhook(c *gin.Context) {
	webhookID := c.Param("id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Webhook ID不能为空",
		})
		return
	}

	var config WebhookConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	webhookInfo, err := h.service.UpdateWebhook(webhookID, config)
	if err != nil {
		h.logger.Error("Failed to update webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新Webhook失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    webhookInfo,
		"message": "Webhook更新成功",
	})
}

// DeleteWebhook 删除Webhook
// @Summary 删除Webhook
// @Description 删除Webhook配置
// @Tags Webhook
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 200 {object} Response
// @Router /webhook/{id} [delete]
func (h *Handler) DeleteWebhook(c *gin.Context) {
	webhookID := c.Param("id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Webhook ID不能为空",
		})
		return
	}

	err := h.service.DeleteWebhook(webhookID)
	if err != nil {
		h.logger.Error("Failed to delete webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除Webhook失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook删除成功",
	})
}

// TestWebhook 测试Webhook
// @Summary 测试Webhook
// @Description 测试Webhook配置是否正常工作
// @Tags Webhook
// @Produce json
// @Param id query string true "Webhook ID"
// @Success 200 {object} Response{data=TestResult}
// @Router /webhook/test [post]
func (h *Handler) TestWebhook(c *gin.Context) {
	webhookID := c.Query("id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Webhook ID不能为空",
		})
		return
	}

	result, err := h.service.TestWebhook(webhookID)
	if err != nil {
		h.logger.Error("Failed to test webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "测试Webhook失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Webhook测试完成",
	})
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// WebhookInfo Webhook信息结构
type WebhookInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Payload    string            `json:"payload"`
	Events     []string          `json:"events"`
	Enabled    bool              `json:"enabled"`
	CreateTime string            `json:"create_time"`
	UpdateTime string            `json:"update_time"`
}

// WebhookConfig Webhook配置结构
type WebhookConfig struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Payload string            `json:"payload"`
	Events  []string          `json:"events"`
	Enabled bool              `json:"enabled"`
}

// TestResult Webhook测试结果结构
type TestResult struct {
	Success      bool   `json:"success"`
	StatusCode   int    `json:"status_code"`
	Response     string `json:"response"`
	Error        string `json:"error"`
	ResponseTime string `json:"response_time"`
}
