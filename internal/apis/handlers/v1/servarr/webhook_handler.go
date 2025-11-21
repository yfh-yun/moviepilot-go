package servarr

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	
	"moviepilot-go/internal/business/services/chain"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"
	"moviepilot-go/internal/repositories"

	"github.com/gin-gonic/gin"
)

// WebhookEventType Webhook事件类型
type WebhookEventType string

const (
	// 电影事件
	WebhookEventMovieDownload   WebhookEventType = "Download"
	WebhookEventMovieGrab      WebhookEventType = "Grab"
	WebhookEventMovieFileDelete WebhookEventType = "FileDelete"
	
	// 剧集事件
	WebhookEventEpisodeDownload   WebhookEventType = "EpisodeFileDelete"
	WebhookEventEpisodeGrab      WebhookEventType = "Grab"
	WebhookEventEpisodeFileDelete WebhookEventType = "Download"
	
	// 测试事件
	WebhookEventTest WebhookEventType = "Test"
)

// WebhookEventTypeDesc 事件类型描述
var WebhookEventTypeDesc = map[WebhookEventType]string{
	WebhookEventMovieDownload:   "电影下载完成",
	WebhookEventMovieGrab:      "电影抓取",
	WebhookEventMovieFileDelete: "电影文件删除",
	WebhookEventEpisodeDownload:   "剧集下载完成",
	WebhookEventEpisodeGrab:      "剧集抓取",
	WebhookEventEpisodeFileDelete: "剧集文件删除",
	WebhookEventTest: "测试事件",
}

// ServArrWebhookHandler ServArr Webhook处理器
type ServArrWebhookHandler struct {
	logger          logger.Logger
	mediaChain      *chain.MediaChain
	subscribeChain  *chain.SubscribeChain
	transferChain   *chain.TransferChain
	subscribeRepo   repository.SubscribeRepository
	downloadRepo    repository.DownloadRepository
	messageRepo     repository.MessageRepository
}

// NewServArrWebhookHandler 创建ServArr Webhook处理器
func NewServArrWebhookHandler(
	logger logger.Logger,
	mediaChain *chain.MediaChain,
	subscribeChain *chain.SubscribeChain,
	transferChain *chain.TransferChain,
	subscribeRepo repository.SubscribeRepository,
	downloadRepo repository.DownloadRepository,
	messageRepo repository.MessageRepository,
) *ServArrWebhookHandler {
	return &ServArrWebhookHandler{
		logger:         logger,
		mediaChain:     mediaChain,
		subscribeChain: subscribeChain,
		transferChain:  transferChain,
		subscribeRepo:  subscribeRepo,
		downloadRepo:   downloadRepo,
		messageRepo:    messageRepo,
	}
}

// HandleMovieWebhook 处理电影Webhook
func (h *ServArrWebhookHandler) HandleMovieWebhook(c *gin.Context) {
	ctx := c.Request.Context()
	eventType := c.GetHeader("X-Servarr-Event")
	
	h.logger.Info("收到电影Webhook事件", 
		logger.String("event_type", eventType),
		logger.String("source", c.ClientIP()))

	switch WebhookEventType(eventType) {
	case WebhookEventMovieDownload:
		h.handleMovieDownload(ctx, c)
	case WebhookEventMovieGrab:
		h.handleMovieGrab(ctx, c)
	case WebhookEventMovieFileDelete:
		h.handleMovieFileDelete(ctx, c)
	case WebhookEventTest:
		h.handleTestWebhook(ctx, c)
	default:
		h.logger.Warn("未知的Webhook事件类型", logger.String("event_type", eventType))
		c.JSON(200, gin.H{"message": "事件已接收"})
	}
}

// HandleSeriesWebhook 处理剧集Webhook
func (h *ServArrWebhookHandler) HandleSeriesWebhook(c *gin.Context) {
	ctx := c.Request.Context()
	eventType := c.GetHeader("X-Servarr-Event")
	
	h.logger.Info("收到剧集Webhook事件", 
		logger.String("event_type", eventType),
		logger.String("source", c.ClientIP()))

	switch WebhookEventType(eventType) {
	case WebhookEventEpisodeDownload:
		h.handleEpisodeDownload(ctx, c)
	case WebhookEventEpisodeGrab:
		h.handleEpisodeGrab(ctx, c)
	case WebhookEventEpisodeFileDelete:
		h.handleEpisodeFileDelete(ctx, c)
	case WebhookEventTest:
		h.handleTestWebhook(ctx, c)
	default:
		h.logger.Warn("未知的Webhook事件类型", logger.String("event_type", eventType))
		c.JSON(200, gin.H{"message": "事件已接收"})
	}
}

// handleMovieDownload 处理电影下载完成事件
func (h *ServArrWebhookHandler) handleMovieDownload(ctx context.Context, c *gin.Context) {
	var webhookData models.ServArrMovie
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		h.logger.Error("解析电影下载数据失败", logger.Error(err))
		c.JSON(400, gin.H{"error": "数据格式错误"})
		return
	}

	h.logger.Info("处理电影下载完成事件",
		logger.String("title", webhookData.Title),
		logger.Int("year", webhookData.Year),
		logger.Int("tmdb_id", webhookData.TMDBID))

	// 1. 更新订阅状态
	if webhookData.ID > 0 {
		err := h.subscribeRepo.UpdateDownloaded(ctx, webhookData.ID, true)
		if err != nil {
			h.logger.Error("更新订阅状态失败", logger.Error(err))
		}
	}

	// 2. 记录下载历史
	downloadHistory := &model.DownloadHistory{
		TMDBID:      webhookData.TMDBID,
		IMDBID:      webhookData.IMDBID,
		Title:       webhookData.Title,
		Year:        webhookData.Year,
		Season:      0,
		Episode:     0,
		FilePath:    webhookData.Path,
		FileSize:    webhookData.SizeOnDisk,
		DownloadTime: time.Now(),
		Status:      model.DownloadStatusCompleted,
		Source:      "ServArr",
	}

	err := h.downloadRepo.Create(ctx, downloadHistory)
	if err != nil {
		h.logger.Error("记录下载历史失败", logger.Error(err))
	}

	// 3. 触发文件转移
	if webhookData.Path != "" {
		transferRequest := &chain.TransferRequest{
			SourcePath: webhookData.Path,
			MediaType:  model.MediaTypeMovie,
			Title:      webhookData.Title,
			Year:       webhookData.Year,
			TMDBID:     webhookData.TMDBID,
			IMDBID:     webhookData.IMDBID,
		}

		_, err = h.transferChain.TransferFile(ctx, transferRequest)
		if err != nil {
			h.logger.Error("触发文件转移失败", logger.Error(err))
		} else {
			h.logger.Info("文件转移任务已触发", logger.String("path", webhookData.Path))
		}
	}

	// 4. 发送通知消息
	h.sendMessage(ctx, &model.Message{
		Type:    model.MessageTypeDownload,
		Title:   "电影下载完成",
		Content: fmt.Sprintf("电影《%s (%d)》已下载完成", webhookData.Title, webhookData.Year),
		Level:   model.MessageLevelInfo,
		Source:  "ServArr",
	})

	c.JSON(200, gin.H{"message": "事件处理完成"})
}

// handleMovieGrab 处理电影抓取事件
func (h *ServArrWebhookHandler) handleMovieGrab(ctx context.Context, c *gin.Context) {
	var webhookData models.ServArrMovie
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		h.logger.Error("解析电影抓取数据失败", logger.Error(err))
		c.JSON(400, gin.H{"error": "数据格式错误"})
		return
	}

	h.logger.Info("处理电影抓取事件",
		logger.String("title", webhookData.Title),
		logger.Int("year", webhookData.Year),
		logger.Int("tmdb_id", webhookData.TMDBID))

	// 记录抓取历史
	downloadHistory := &model.DownloadHistory{
		TMDBID:      webhookData.TMDBID,
		IMDBID:      webhookData.IMDBID,
		Title:       webhookData.Title,
		Year:        webhookData.Year,
		Season:      0,
		Episode:     0,
		DownloadTime: time.Now(),
		Status:      model.DownloadStatusGrabbed,
		Source:      "ServArr",
	}

	err := h.downloadRepo.Create(ctx, downloadHistory)
	if err != nil {
		h.logger.Error("记录抓取历史失败", logger.Error(err))
	}

	c.JSON(200, gin.H{"message": "事件处理完成"})
}

// handleMovieFileDelete 处理电影文件删除事件
func (h *ServArrWebhookHandler) handleMovieFileDelete(ctx context.Context, c *gin.Context) {
	var webhookData models.ServArrMovie
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		h.logger.Error("解析电影删除数据失败", logger.Error(err))
		c.JSON(400, gin.H{"error": "数据格式错误"})
		return
	}

	h.logger.Info("处理电影文件删除事件",
		logger.String("title", webhookData.Title),
		logger.Int("tmdb_id", webhookData.TMDBID))

	// 1. 更新订阅状态
	if webhookData.ID > 0 {
		err := h.subscribeRepo.UpdateDownloaded(ctx, webhookData.ID, false)
		if err != nil {
			h.logger.Error("更新订阅状态失败", logger.Error(err))
		}
	}

	// 2. 发送通知消息
	h.sendMessage(ctx, &model.Message{
		Type:    model.MessageTypeDownload,
		Title:   "电影文件已删除",
		Content: fmt.Sprintf("电影《%s》的文件已被删除", webhookData.Title),
		Level:   model.MessageLevelWarning,
		Source:  "ServArr",
	})

	c.JSON(200, gin.H{"message": "事件处理完成"})
}

// handleEpisodeDownload 处理剧集下载完成事件
func (h *ServArrWebhookHandler) handleEpisodeDownload(ctx context.Context, c *gin.Context) {
	var webhookData models.ServArrSeries
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		h.logger.Error("解析剧集下载数据失败", logger.Error(err))
		c.JSON(400, gin.H{"error": "数据格式错误"})
		return
	}

	h.logger.Info("处理剧集下载完成事件",
		logger.String("title", webhookData.Title),
		logger.Int("tvdb_id", webhookData.TVDBID))

	// TODO: 处理具体的剧集下载完成逻辑
	// 需要从webhook中解析具体的剧集信息

	// 发送通知消息
	h.sendMessage(ctx, &model.Message{
		Type:    model.MessageTypeDownload,
		Title:   "剧集下载完成",
		Content: fmt.Sprintf("剧集《%s》已下载完成", webhookData.Title),
		Level:   model.MessageLevelInfo,
		Source:  "ServArr",
	})

	c.JSON(200, gin.H{"message": "事件处理完成"})
}

// handleEpisodeGrab 处理剧集抓取事件
func (h *ServArrWebhookHandler) handleEpisodeGrab(ctx context.Context, c *gin.Context) {
	var webhookData models.ServArrSeries
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		h.logger.Error("解析剧集抓取数据失败", logger.Error(err))
		c.JSON(400, gin.H{"error": "数据格式错误"})
		return
	}

	h.logger.Info("处理剧集抓取事件",
		logger.String("title", webhookData.Title),
		logger.Int("tvdb_id", webhookData.TVDBID))

	c.JSON(200, gin.H{"message": "事件处理完成"})
}

// handleEpisodeFileDelete 处理剧集文件删除事件
func (h *ServArrWebhookHandler) handleEpisodeFileDelete(ctx context.Context, c *gin.Context) {
	var webhookData models.ServArrSeries
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		h.logger.Error("解析剧集删除数据失败", logger.Error(err))
		c.JSON(400, gin.H{"error": "数据格式错误"})
		return
	}

	h.logger.Info("处理剧集文件删除事件",
		logger.String("title", webhookData.Title),
		logger.Int("tvdb_id", webhookData.TVDBID))

	// 发送通知消息
	h.sendMessage(ctx, &model.Message{
		Type:    model.MessageTypeDownload,
		Title:   "剧集文件已删除",
		Content: fmt.Sprintf("剧集《%s》的文件已被删除", webhookData.Title),
		Level:   model.MessageLevelWarning,
		Source:  "ServArr",
	})

	c.JSON(200, gin.H{"message": "事件处理完成"})
}

// handleTestWebhook 处理测试Webhook
func (h *ServArrWebhookHandler) handleTestWebhook(ctx context.Context, c *gin.Context) {
	h.logger.Info("收到ServArr测试Webhook")

	// 发送测试消息
	h.sendMessage(ctx, &model.Message{
		Type:    model.MessageTypeSystem,
		Title:   "ServArr连接测试",
		Content: "ServArr Webhook连接测试成功",
		Level:   model.MessageLevelInfo,
		Source:  "ServArr",
	})

	c.JSON(200, gin.H{"message": "测试成功"})
}

// sendMessage 发送消息
func (h *ServArrWebhookHandler) sendMessage(ctx context.Context, message *model.Message) {
	err := h.messageRepo.Create(ctx, message)
	if err != nil {
		h.logger.Error("发送消息失败", logger.Error(err))
	}
}

// ParseWebhookData 解析Webhook数据通用函数
func (h *ServArrWebhookHandler) ParseWebhookData(c *gin.Context, data interface{}) error {
	body, err := c.GetRawData()
	if err != nil {
		return fmt.Errorf("获取请求体失败: %w", err)
	}

	err = json.Unmarshal(body, data)
	if err != nil {
		return fmt.Errorf("解析JSON失败: %w", err)
	}

	return nil
}