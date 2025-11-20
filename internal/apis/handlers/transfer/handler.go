// Package transfer Transfer API处理器模块
package transfer

import (
	"net/http"
	"strconv"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/transfer"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler Transfer API处理器
type Handler struct {
	service transfer.Service
	logger  *logger.Logger
}

// NewHandler 创建新的Transfer处理器
func NewHandler(service transfer.Service, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	transferGroup := router.Group("/transfer")
	{
		transferGroup.GET("/", h.GetTransferList)
		transferGroup.GET("/:id", h.GetTransfer)
		transferGroup.POST("/", h.CreateTransfer)
		transferGroup.DELETE("/:id", h.DeleteTransfer)
		transferGroup.POST("/:id/retry", h.RetryTransfer)
		transferGroup.POST("/:id/pause", h.PauseTransfer)
		transferGroup.POST("/:id/resume", h.ResumeTransfer)
	}
}

// GetTransferList 获取转移列表
// @Summary 获取转移列表
// @Description 获取文件转移任务列表
// @Tags 转移
// @Produce json
// @Param status query string false "状态过滤 (pending,running,completed,failed)"
// @Param page query int false "页码" default(1)
// @Param count query int false "每页数量" default(30)
// @Success 200 {object} Response{data=[]TransferInfo}
// @Router /transfer [get]
func (h *Handler) GetTransferList(c *gin.Context) {
	status := c.Query("status")
	pageParam := c.DefaultQuery("page", "1")
	countParam := c.DefaultQuery("count", "30")

	page, _ := strconv.Atoi(pageParam)
	count, _ := strconv.Atoi(countParam)

	params := transfer.ListParams{
		Status: status,
		Page:   page,
		Count:  count,
	}

	transfers, err := h.service.GetTransferList(params)
	if err != nil {
		h.logger.Error("Failed to get transfer list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取转移列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transfers,
	})
}

// GetTransfer 获取转移详情
// @Summary 获取转移详情
// @Description 获取文件转移任务详细信息
// @Tags 转移
// @Produce json
// @Param id path string true "转移任务ID"
// @Success 200 {object} Response{data=TransferDetail}
// @Router /transfer/{id} [get]
func (h *Handler) GetTransfer(c *gin.Context) {
	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "转移任务ID不能为空",
		})
		return
	}

	transferDetail, err := h.service.GetTransfer(transferID)
	if err != nil {
		h.logger.Error("Failed to get transfer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取转移详情失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transferDetail,
	})
}

// CreateTransfer 创建转移任务
// @Summary 创建转移任务
// @Description 创建新的文件转移任务
// @Tags 转移
// @Produce json
// @Param transfer body CreateTransferRequest true "转移任务信息"
// @Success 200 {object} Response{data=TransferInfo}
// @Router /transfer [post]
func (h *Handler) CreateTransfer(c *gin.Context) {
	var request CreateTransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	transferInfo, err := h.service.CreateTransfer(request)
	if err != nil {
		h.logger.Error("Failed to create transfer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建转移任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    transferInfo,
		"message": "转移任务创建成功",
	})
}

// DeleteTransfer 删除转移任务
// @Summary 删除转移任务
// @Description 删除文件转移任务
// @Tags 转移
// @Produce json
// @Param id path string true "转移任务ID"
// @Param delete_files query bool false "是否删除文件" default(false)
// @Success 200 {object} Response
// @Router /transfer/{id} [delete]
func (h *Handler) DeleteTransfer(c *gin.Context) {
	transferID := c.Param("id")
	deleteFilesParam := c.DefaultQuery("delete_files", "false")

	if transferID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "转移任务ID不能为空",
		})
		return
	}

	deleteFiles := deleteFilesParam == "true"

	err := h.service.DeleteTransfer(transferID, deleteFiles)
	if err != nil {
		h.logger.Error("Failed to delete transfer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除转移任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "转移任务删除成功",
	})
}

// RetryTransfer 重试转移任务
// @Summary 重试转移任务
// @Description 重试失败的转移任务
// @Tags 转移
// @Produce json
// @Param id path string true "转移任务ID"
// @Success 200 {object} Response
// @Router /transfer/{id}/retry [post]
func (h *Handler) RetryTransfer(c *gin.Context) {
	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "转移任务ID不能为空",
		})
		return
	}

	err := h.service.RetryTransfer(transferID)
	if err != nil {
		h.logger.Error("Failed to retry transfer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "重试转移任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "转移任务重试成功",
	})
}

// PauseTransfer 暂停转移任务
// @Summary 暂停转移任务
// @Description 暂停文件转移任务
// @Tags 转移
// @Produce json
// @Param id path string true "转移任务ID"
// @Success 200 {object} Response
// @Router /transfer/{id}/pause [post]
func (h *Handler) PauseTransfer(c *gin.Context) {
	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "转移任务ID不能为空",
		})
		return
	}

	err := h.service.PauseTransfer(transferID)
	if err != nil {
		h.logger.Error("Failed to pause transfer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "暂停转移任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "转移任务已暂停",
	})
}

// ResumeTransfer 恢复转移任务
// @Summary 恢复转移任务
// @Description 恢复暂停的文件转移任务
// @Tags 转移
// @Produce json
// @Param id path string true "转移任务ID"
// @Success 200 {object} Response
// @Router /transfer/{id}/resume [post]
func (h *Handler) ResumeTransfer(c *gin.Context) {
	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "转移任务ID不能为空",
		})
		return
	}

	err := h.service.ResumeTransfer(transferID)
	if err != nil {
		h.logger.Error("Failed to resume transfer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "恢复转移任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "转移任务已恢复",
	})
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// TransferInfo 转移信息结构
type TransferInfo struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	SourcePath    string  `json:"source_path"`
	DestPath      string  `json:"dest_path"`
	Status        string  `json:"status"`
	Size          int64   `json:"size"`
	Transferred   int64   `json:"transferred"`
	Progress      float64 `json:"progress"`
	Speed         int64   `json:"speed"`
	TimeRemaining string  `json:"time_remaining"`
	CreateTime    string  `json:"create_time"`
	UpdateTime    string  `json:"update_time"`
}

// TransferDetail 转移详情结构
type TransferDetail struct {
	TransferInfo
	Files        []TransferFile `json:"files"`
	Error        string         `json:"error"`
	RetryCount   int            `json:"retry_count"`
	TransferMode string         `json:"transfer_mode"`
	Checksum     string         `json:"checksum"`
}

// CreateTransferRequest 创建转移任务请求结构
type CreateTransferRequest struct {
	SourcePath string `json:"source_path"`
	DestPath   string `json:"dest_path"`
	Mode       string `json:"mode"`
	Overwrite  bool   `json:"overwrite"`
	Validate   bool   `json:"validate"`
}

// TransferFile 转移文件结构
type TransferFile struct {
	Name        string  `json:"name"`
	Size        int64   `json:"size"`
	Transferred int64   `json:"transferred"`
	Status      string  `json:"status"`
	Progress    float64 `json:"progress"`
}
