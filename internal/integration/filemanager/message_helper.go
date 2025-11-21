// Package filemanager 消息助手
package filemanager

import (
	"context"
	"fmt"
	"text/template"

	"moviepilot-go/pkg/logger"
	"moviepilot-go/internal/models"

	"go.uber.org/zap"
)

// MessageHelper 消息助手
type MessageHelper struct {
	logger *zap.Logger
}

// NewMessageHelper 创建消息助手
func NewMessageHelper() *MessageHelper {
	return &MessageHelper{
		logger: logger.Logger,
	}
}

// SendMessage 发送消息
func (mh *MessageHelper) SendMessage(ctx context.Context, title, content string, level string) error {
	mh.logger.Info("Sending message",
		zap.String("title", title),
		zap.String("content", content),
		zap.String("level", level))

	// 这里应该调用实际的消息发送服务
	// 为了演示，只记录日志
	return nil
}

// SendTransferNotification 发送转移通知
func (mh *MessageHelper) SendTransferNotification(ctx context.Context, transferInfo *TransferInfo, mediaInfo *models.MediaInfo) error {
	title := "文件转移完成"
	
	content, err := mh.generateTransferContent(transferInfo, mediaInfo)
	if err != nil {
		return fmt.Errorf("failed to generate transfer content: %w", err)
	}

	level := "info"
	if !transferInfo.Success {
		level = "error"
	}

	return mh.SendMessage(ctx, title, content, level)
}

// SendScrapeNotification 发送刮削通知
func (mh *MessageHelper) SendScrapeNotification(ctx context.Context, filePath string, mediaInfo *models.MediaInfo, success bool) error {
	title := "元数据刮削"
	
	var content string
	if success {
		content = fmt.Sprintf("文件 %s 的元数据刮削完成\n标题: %s\n年份: %d",
			filePath, mediaInfo.Title, mediaInfo.Year)
	} else {
		content = fmt.Sprintf("文件 %s 的元数据刮削失败", filePath)
	}

	level := "info"
	if !success {
		level = "error"
	}

	return mh.SendMessage(ctx, title, content, level)
}

// SendErrorNotification 发送错误通知
func (mh *MessageHelper) SendErrorNotification(ctx context.Context, operation string, err error) error {
	title := "操作失败"
	content := fmt.Sprintf("操作: %s\n错误: %s", operation, err.Error())
	
	return mh.SendMessage(ctx, title, content, "error")
}

// generateTransferContent 生成转移内容
func (mh *MessageHelper) generateTransferContent(transferInfo *TransferInfo, mediaInfo *models.MediaInfo) (string, error) {
	tmpl := `文件转移通知

状态: {{if .Success}}成功{{else}}失败{{end}}
媒体标题: {{.MediaInfo.Title}}
媒体类型: {{.MediaInfo.Type}}
年份: {{.MediaInfo.Year}}

{{if .Success}}
源文件: {{.SourceFile}}
目标文件: {{.TargetFile}}
转移文件数: {{.TransferCount}}
总文件数: {{.TotalCount}}
{{if .NeedScrape}}
元数据刮削: {{if .ScrapeStatus}}成功{{else}}失败{{end}}
{{end}}
{{if .NeedNotify}}
通知发送: {{if .NotifyStatus}}成功{{else}}失败{{end}}
{{end}}
{{else}}
错误信息: {{.Message}}
失败文件数: {{len .FailedFiles}}
{{end}}
`

	t, err := template.New("transfer").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		*TransferInfo
		MediaInfo *models.MediaInfo
	}{
		TransferInfo: transferInfo,
		MediaInfo:    mediaInfo,
	}

	var result strings.Builder
	if err := t.Execute(&result, data); err != nil {
		return "", err
	}

	return result.String(), nil
}

// SendProgressNotification 发送进度通知
func (mh *MessageHelper) SendProgressNotification(ctx context.Context, operation string, current, total int) error {
	title := "操作进度"
	content := fmt.Sprintf("操作: %s\n进度: %d/%d (%.1f%%)",
		operation, current, total, float64(current)/float64(total)*100)

	return mh.SendMessage(ctx, title, content, "info")
}

// SendCompletionNotification 发送完成通知
func (mh *MessageHelper) SendCompletionNotification(ctx context.Context, operation string, success bool, summary string) error {
	title := fmt.Sprintf("%s完成", operation)
	level := "info"
	if !success {
		title = fmt.Sprintf("%s失败", operation)
		level = "error"
	}

	content := summary
	if content == "" {
		if success {
			content = fmt.Sprintf("%s操作已成功完成", operation)
		} else {
			content = fmt.Sprintf("%s操作失败", operation)
		}
	}

	return mh.SendMessage(ctx, title, content, level)
}

// SendStorageNotification 发送存储通知
func (mh *MessageHelper) SendStorageNotification(ctx context.Context, storageType string, usage *StorageUsage) error {
	title := "存储使用情况"
	
	percentage := float64(usage.Used) / float64(usage.Total) * 100
	content := fmt.Sprintf(`存储类型: %s
总空间: %s
已使用: %s
可用空间: %s
使用率: %.1f%%`,
		storageType,
		mh.formatBytes(usage.Total),
		mh.formatBytes(usage.Used),
		mh.formatBytes(usage.Free),
		percentage)

	level := "info"
	if percentage > 90 {
		level = "warning"
	}
	if percentage > 95 {
		level = "error"
	}

	return mh.SendMessage(ctx, title, content, level)
}

// formatBytes 格式化字节数
func (mh *MessageHelper) formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// SendCustomNotification 发送自定义通知
func (mh *MessageHelper) SendCustomNotification(ctx context.Context, title, content string, templateData interface{}, templateStr string) error {
	if templateStr != "" {
		tmpl, err := template.New("custom").Parse(templateStr)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		var result strings.Builder
		if err := tmpl.Execute(&result, templateData); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		content = result.String()
	}

	return mh.SendMessage(ctx, title, content, "info")
}

// SendBatchNotification 发送批量通知
func (mh *MessageHelper) SendBatchNotification(ctx context.Context, notifications []*NotificationMessage) error {
	for _, notification := range notifications {
		if err := mh.SendMessage(ctx, notification.Title, notification.Content, notification.Level); err != nil {
			mh.logger.Error("Failed to send notification",
				zap.String("title", notification.Title),
				zap.Error(err))
			// 继续发送其他通知，不返回错误
		}
	}
	return nil
}

// NotificationMessage 通知消息
type NotificationMessage struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Level   string `json:"level"`
}

// SendMediaAddedNotification 发送媒体添加通知
func (mh *MessageHelper) SendMediaAddedNotification(ctx context.Context, mediaInfo *models.MediaInfo, filePath string) error {
	title := "新媒体添加"
	
	content := fmt.Sprintf(`标题: %s
类型: %s
年份: %d
文件路径: %s`,
		mediaInfo.Title,
		mediaInfo.Type,
		mediaInfo.Year,
		filePath)

	return mh.SendMessage(ctx, title, content, "info")
}

// SendMediaUpdatedNotification 发送媒体更新通知
func (mh *MessageHelper) SendMediaUpdatedNotification(ctx context.Context, mediaInfo *models.MediaInfo, changes []string) error {
	title := "媒体信息更新"
	
	content := fmt.Sprintf(`标题: %s
类型: %s
年份: %d

更新内容:
%s`,
		mediaInfo.Title,
		mediaInfo.Type,
		mediaInfo.Year,
		strings.Join(changes, "\n"))

	return mh.SendMessage(ctx, title, content, "info")
}

// SendSystemNotification 发送系统通知
func (mh *MessageHelper) SendSystemNotification(ctx context.Context, message string, level string) error {
	title := "系统通知"
	return mh.SendMessage(ctx, title, message, level)
}

// ValidateMessage 验证消息内容
func (mh *MessageHelper) ValidateMessage(title, content string) error {
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	if len(title) > 200 {
		return fmt.Errorf("title too long (max 200 characters)")
	}
	if len(content) > 2000 {
		return fmt.Errorf("content too long (max 2000 characters)")
	}
	return nil
}

// GetNotificationStats 获取通知统计
func (mh *MessageHelper) GetNotificationStats(ctx context.Context) (*NotificationStats, error) {
	// 这里应该从数据库获取实际统计信息
	// 为了演示，返回模拟数据
	stats := &NotificationStats{
		TotalSent:     1000,
		SuccessCount:  950,
		FailureCount:  50,
		LastSentTime:  time.Now().Add(-time.Hour),
		AverageDelay:  time.Second * 5,
	}

	return stats, nil
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalSent    int64     `json:"total_sent"`
	SuccessCount int64     `json:"success_count"`
	FailureCount int64     `json:"failure_count"`
	LastSentTime time.Time `json:"last_sent_time"`
	AverageDelay time.Duration `json:"average_delay"`
}