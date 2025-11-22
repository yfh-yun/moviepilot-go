package workflowapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/response"
	"moviepilot-go/pkg/validator"
)

// WorkflowService 定义 handler 所需的最小接口，方便测试替换。
type WorkflowService interface {
	StartLocalFileWorkflow(ctx context.Context, req StartLocalFileWorkflowRequest) (*StartLocalFileWorkflowResponse, error)
}

// Handler 负责处理 workflow 相关 API。
type Handler struct {
	service WorkflowService
	logger  *zap.Logger
}

// NewHandler 创建 Handler。
func NewHandler(service WorkflowService, log *zap.Logger) *Handler {
	if log == nil {
		log = logger.GetLogger()
	}
	return &Handler{service: service, logger: log}
}

// StartLocalFileWorkflow 触发“本地文件 → 刮削 → 转移”链路。
func (h *Handler) StartLocalFileWorkflow(c *gin.Context) {
	var req StartLocalFileWorkflowRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		h.logger.Warn("workflow request validation failed",
			zap.String("path", c.FullPath()),
			zap.Error(err),
		)
		response.InvalidParams(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("workflow request accepted",
		zap.String("path", c.FullPath()),
		zap.String("root_path", req.RootPath),
		zap.Bool("include_fetch", req.IncludeFetch),
		zap.Bool("wait_for_completion", req.WaitForCompletion),
	)

	result, err := h.service.StartLocalFileWorkflow(ctx, req)
	if err != nil {
		h.logger.Error("workflow start failed",
			zap.String("path", c.FullPath()),
			zap.Error(err),
		)
		response.ErrorWithLog(c, response.CodeServerError, "workflow start failed", err)
		return
	}

	statusCode := http.StatusAccepted
	message := "workflow started"
	if req.WaitForCompletion {
		statusCode = http.StatusOK
		if result != nil && result.Message != "" {
			message = result.Message
		} else {
			message = "workflow completed"
		}
	}

	c.JSON(statusCode, response.Response{
		Success:   true,
		Code:      response.CodeSuccess,
		Message:   message,
		Data:      result,
		Timestamp: time.Now().Unix(),
	})
}
