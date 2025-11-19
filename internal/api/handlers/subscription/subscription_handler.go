// Package subscription 订阅和下载API处理器
package subscription

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yfh-yun/moviepilot-go/internal/service"
	"github.com/yfh-yun/moviepilot-go/pkg/response"
	"github.com/yfh-yun/moviepilot-go/pkg/validator"
)

// SubscriptionHandler 订阅和下载处理器
// 提供媒体订阅、下载队列管理、下载状态监控等功能
type SubscriptionHandler struct {
	subscriptionService service.SubscriptionService
	downloadService     service.DownloadService
	logger              *zap.Logger
}

// NewSubscriptionHandler 创建订阅和下载处理器
func NewSubscriptionHandler(
	subscriptionService service.SubscriptionService,
	downloadService service.DownloadService,
	logger *zap.Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		downloadService:     downloadService,
		logger:              logger,
	}
}

// SubscribeRequest 订阅请求结构体
type SubscribeRequest struct {
	MediaID   string `json:"media_id" binding:"required"`
	NotifyURL string `json:"notify_url" validate:"omitempty,url"`
}

// SubscribeResponse 订阅响应结构体
type SubscribeResponse struct {
	SubscriptionID string `json:"subscription_id"`
	MediaID        string `json:"media_id"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

// Subscribe 订阅媒体
// @Summary 订阅媒体
// @Description 订阅指定的媒体，系统会自动下载并通知用户
// @Tags subscription
// @Accept json
// @Produce json
// @Param request body SubscribeRequest true "订阅参数"
// @Success 201 {object} response.SuccessResponse{data=SubscribeResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscription/subscribe [post]
func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	
	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("订阅请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}
	
	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("订阅请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}
	
	ctx := c.Request.Context()
	userID := c.GetString("user_id") // 从认证中间件获取用户ID
	
	// 创建订阅
	subscription, err := h.subscriptionService.CreateSubscription(ctx, service.CreateSubscriptionParams{
		UserID:    userID,
		MediaID:   req.MediaID,
		NotifyURL: req.NotifyURL,
	})
	
	if err != nil {
		if err == service.ErrMediaNotFound {
			response.NotFound(c, "媒体不存在")
			return
		}
		h.logger.Error("创建订阅失败", zap.Error(err), zap.String("media_id", req.MediaID))
		response.InternalServerError(c, "创建订阅失败")
		return
	}
	
	response.Created(c, SubscribeResponse{
		SubscriptionID: subscription.ID,
		MediaID:        subscription.MediaID,
		Status:         subscription.Status,
		CreatedAt:      subscription.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// UnsubscribeRequest 取消订阅请求结构体
type UnsubscribeRequest struct {
	SubscriptionID string `json:"subscription_id" binding:"required"`
}

// Unsubscribe 取消订阅
// @Summary 取消订阅
// @Description 取消指定的媒体订阅
// @Tags subscription
// @Accept json
// @Produce json
// @Param request body UnsubscribeRequest true "取消订阅参数"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscription/unsubscribe [post]
func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	var req UnsubscribeRequest
	
	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("取消订阅请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}
	
	ctx := c.Request.Context()
	userID := c.GetString("user_id") // 从认证中间件获取用户ID
	
	err := h.subscriptionService.CancelSubscription(ctx, req.SubscriptionID, userID)
	if err != nil {
		if err == service.ErrSubscriptionNotFound {
			response.NotFound(c, "订阅不存在")
			return
		}
		h.logger.Error("取消订阅失败", zap.Error(err), zap.String("subscription_id", req.SubscriptionID))
		response.InternalServerError(c, "取消订阅失败")
		return
	}
	
	response.Success(c, gin.H{
		"message":         "取消订阅成功",
		"subscription_id": req.SubscriptionID,
	})
}

// GetUserSubscriptions 获取用户订阅列表
// @Summary 获取用户订阅列表
// @Description 获取当前用户的所有订阅记录
// @Tags subscription
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param status query string false "订阅状态"
// @Success 200 {object} response.SuccessResponse{data=UserSubscriptionsResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/subscription/list [get]
func (h *SubscriptionHandler) GetUserSubscriptions(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")
	status := c.Query("status")
	
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		response.BadRequest(c, "页码参数错误")
		return
	}
	
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt < 1 || pageSizeInt > 100 {
		response.BadRequest(c, "每页数量参数错误")
		return
	}
	
	ctx := c.Request.Context()
	userID := c.GetString("user_id")
	
	subscriptions, total, err := h.subscriptionService.GetUserSubscriptions(ctx, service.GetUserSubscriptionsParams{
		UserID:   userID,
		Status:   status,
		Page:     pageInt,
		PageSize: pageSizeInt,
	})
	
	if err != nil {
		h.logger.Error("获取用户订阅列表失败", zap.Error(err))
		response.InternalServerError(c, "获取订阅列表失败")
		return
	}
	
	var responseList []*SubscriptionItem
	for _, sub := range subscriptions {
		responseList = append(responseList, &SubscriptionItem{
			SubscriptionID: sub.ID,
			MediaID:        sub.MediaID,
			Status:         sub.Status,
			CreatedAt:      sub.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      sub.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	
	response.Success(c, UserSubscriptionsResponse{
		Subscriptions: responseList,
		Total:         total,
		Page:          pageInt,
		PageSize:      pageSizeInt,
	})
}

// DownloadRequest 下载请求结构体
type DownloadRequest struct {
	MediaID string `json:"media_id" binding:"required"`
	Format  string `json:"format" validate:"omitempty,oneof=mp4 avi mkv"`
	Quality string `json:"quality" validate:"omitempty,oneof=low medium high"`
}

// DownloadResponse 下载响应结构体
type DownloadResponse struct {
	DownloadID string `json:"download_id"`
	MediaID    string `json:"media_id"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	CreatedAt  string `json:"created_at"`
}

// DownloadMedia 下载媒体文件
// @Summary 下载媒体文件
// @Description 下载指定的媒体文件到本地
// @Tags download
// @Accept json
// @Produce json
// @Param request body DownloadRequest true "下载参数"
// @Success 201 {object} response.SuccessResponse{data=DownloadResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/download/start [post]
func (h *SubscriptionHandler) DownloadMedia(c *gin.Context) {
	var req DownloadRequest
	
	// 绑定JSON数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("下载请求参数绑定失败", zap.Error(err))
		response.BadRequest(c, "请求参数格式错误")
		return
	}
	
	// 验证请求参数
	if err := validator.Validate().Struct(req); err != nil {
		h.logger.Warn("下载请求参数验证失败", zap.Error(err))
		response.BadRequest(c, validator.TranslateError(err))
		return
	}
	
	ctx := c.Request.Context()
	userID := c.GetString("user_id")
	
	download, err := h.downloadService.StartDownload(ctx, service.StartDownloadParams{
		UserID:  userID,
		MediaID: req.MediaID,
		Format:  req.Format,
		Quality: req.Quality,
	})
	
	if err != nil {
		if err == service.ErrMediaNotFound {
			response.NotFound(c, "媒体不存在")
			return
		}
		h.logger.Error("开始下载失败", zap.Error(err), zap.String("media_id", req.MediaID))
		response.InternalServerError(c, "开始下载失败")
		return
	}
	
	response.Created(c, DownloadResponse{
		DownloadID: download.ID,
		MediaID:    download.MediaID,
		Status:     download.Status,
		Progress:   download.Progress,
		CreatedAt:  download.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// GetDownloadStatus 获取下载状态
// @Summary 获取下载状态
// @Description 根据下载ID获取下载状态和进度
// @Tags download
// @Accept json
// @Produce json
// @Param download_id path string true "下载ID"
// @Success 200 {object} response.SuccessResponse{data=DownloadResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/download/{download_id}/status [get]
func (h *SubscriptionHandler) GetDownloadStatus(c *gin.Context) {
	downloadID := c.Param("download_id")
	
	if downloadID == "" {
		response.BadRequest(c, "下载ID不能为空")
		return
	}
	
	ctx := c.Request.Context()
	userID := c.GetString("user_id")
	
	download, err := h.downloadService.GetDownloadStatus(ctx, downloadID, userID)
	if err != nil {
		if err == service.ErrDownloadNotFound {
			response.NotFound(c, "下载任务不存在")
			return
		}
		h.logger.Error("获取下载状态失败", zap.Error(err), zap.String("download_id", downloadID))
		response.InternalServerError(c, "获取下载状态失败")
		return
	}
	
	response.Success(c, DownloadResponse{
		DownloadID: download.ID,
		MediaID:    download.MediaID,
		Status:     download.Status,
		Progress:   download.Progress,
		CreatedAt:  download.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// CancelDownload 取消下载
// @Summary 取消下载
// @Description 取消指定的下载任务
// @Tags download
// @Accept json
// @Produce json
// @Param download_id path string true "下载ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/download/{download_id}/cancel [post]
func (h *SubscriptionHandler) CancelDownload(c *gin.Context) {
	downloadID := c.Param("download_id")
	
	if downloadID == "" {
		response.BadRequest(c, "下载ID不能为空")
		return
	}
	
	ctx := c.Request.Context()
	userID := c.GetString("user_id")
	
	err := h.downloadService.CancelDownload(ctx, downloadID, userID)
	if err != nil {
		if err == service.ErrDownloadNotFound {
			response.NotFound(c, "下载任务不存在")
			return
		}
		h.logger.Error("取消下载失败", zap.Error(err), zap.String("download_id", downloadID))
		response.InternalServerError(c, "取消下载失败")
		return
	}
	
	response.Success(c, gin.H{
		"message":     "下载取消成功",
		"download_id": downloadID,
	})
}

// UserSubscriptionsResponse 用户订阅列表响应
type UserSubscriptionsResponse struct {
	Subscriptions []*SubscriptionItem `json:"subscriptions"`
	Total         int64                `json:"total"`
	Page          int                  `json:"page"`
	PageSize      int                  `json:"page_size"`
}

// SubscriptionItem 订阅项目结构体
type SubscriptionItem struct {
	SubscriptionID string `json:"subscription_id"`
	MediaID        string `json:"media_id"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// Add missing import
import "strconv"