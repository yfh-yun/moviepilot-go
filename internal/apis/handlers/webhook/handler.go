package webhook

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	webhookbiz "moviepilot-go/internal/business/services/webhook"
	"moviepilot-go/pkg/logger"
)

// Handler Webhook API处理器
type Handler struct {
	webhookService *webhookbiz.WebhookService
	logger         *zap.Logger
}

// NewHandler 创建Webhook处理器
func NewHandler(webhookService *webhookbiz.WebhookService) *Handler {
	return &Handler{
		webhookService: webhookService,
		logger:         logger.GetLogger(),
	}
}

// WebhookMessage Webhook消息响应
// @Summary Webhook消息响应
// @Description Webhook响应，配置请求中需要添加参数：token=API_TOKEN&source=媒体服务器名
// @Tags webhook
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/webhook [post]
func (h *Handler) WebhookMessage(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Webhook API WebhookMessage started", zap.String("request_id", reqID))

	// TODO: 实现API令牌验证
	// token := c.Query("token")
	// if err := security.VerifyAPIToken(token, expectedToken); err != nil {
	//     h.logger.Warn("Webhook API WebhookMessage invalid token",
	//         zap.String("request_id", reqID),
	//         zap.Error(err),
	//     )
	//     c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
	//     return
	// }

	// TODO: 实现后台任务处理webhook消息
	// body, err := c.GetRawData()
	// if err != nil {
	//     h.logger.Warn("Webhook API WebhookMessage failed to get body",
	//         zap.String("request_id", reqID),
	//         zap.Error(err),
	//     )
	//     body = nil
	// }

	// var form map[string]string
	// if err := c.ShouldBind(&form); err != nil {
	//     form = nil
	// }

	// args := c.Request.URL.Query()
	// background_tasks.add_task(start_webhook_chain, body, form, args)

	h.logger.Info("Webhook API WebhookMessage succeeded", zap.String("request_id", reqID))

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// WebhookMessageGet Webhook消息响应（GET方法）
// @Summary Webhook消息响应
// @Description Webhook响应，配置请求中需要添加参数：token=API_TOKEN&source=媒体服务器名
// @Tags webhook
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/webhook [get]
func (h *Handler) WebhookMessageGet(c *gin.Context) {
	reqID := c.GetString("request_id")
	h.logger.Debug("Webhook API WebhookMessageGet started", zap.String("request_id", reqID))

	// TODO: 实现API令牌验证
	// token := c.Query("token")
	// if err := security.VerifyAPIToken(token, expectedToken); err != nil {
	//     h.logger.Warn("Webhook API WebhookMessageGet invalid token",
	//         zap.String("request_id", reqID),
	//         zap.Error(err),
	//     )
	//     c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
	//     return
	// }

	// TODO: 实现后台任务处理webhook消息
	// args := c.Request.URL.Query()
	// background_tasks.add_task(start_webhook_chain, nil, nil, args)

	h.logger.Info("Webhook API WebhookMessageGet succeeded", zap.String("request_id", reqID))

	c.JSON(http.StatusOK, gin.H{"success": true})
}
