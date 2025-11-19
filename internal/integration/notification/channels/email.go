package channels

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/integration/notification"
)

// EmailConfig 邮件通知配置
type EmailConfig struct {
	SMTPHost      string `json:"smtp_host"`      // SMTP服务器地址
	SMTPPort      int    `json:"smtp_port"`      // SMTP服务器端口
	Username      string `json:"username"`       // 邮箱用户名
	Password      string `json:"password"`       // 邮箱密码或授权码
	From          string `json:"from"`           // 发件人邮箱
	To            string `json:"to"`             // 收件人邮箱，多个用逗号分隔
	SubjectPrefix string `json:"subject_prefix"` // 邮件主题前缀
	UseTLS        bool   `json:"use_tls"`        // 是否使用TLS
}

// EmailSender 邮件通知发送器
type EmailSender struct {
	config *EmailConfig
}

// NewEmailSender 创建新的邮件通知发送器
func NewEmailSender(config *EmailConfig) *EmailSender {
	return &EmailSender{
		config: config,
	}
}

// Name 返回发送器名称
func (e *EmailSender) Name() string {
	return "email"
}

// SupportedLevels 返回支持的通知级别
func (e *EmailSender) SupportedLevels() []notification.NotificationLevel {
	return []notification.NotificationLevel{
		notification.LevelInfo,
		notification.LevelWarning,
		notification.LevelError,
		notification.LevelSuccess,
	}
}

// Validate 验证消息
func (e *EmailSender) Validate(message *notification.Message) error {
	if message.Title == "" {
		return fmt.Errorf("email subject cannot be empty")
	}

	if message.Content == "" {
		return fmt.Errorf("email content cannot be empty")
	}

	if e.config.To == "" {
		return fmt.Errorf("email recipient is not configured")
	}

	return nil
}

// Send 发送邮件通知
func (e *EmailSender) Send(ctx context.Context, message *notification.Message) error {
	if err := e.Validate(message); err != nil {
		return err
	}

	// 构建邮件主题
	subject := e.buildSubject(message)

	// 构建邮件内容
	body := e.buildBody(message)

	// 构建邮件头
	header := make(map[string]string)
	header["From"] = e.config.From
	header["To"] = e.config.To
	header["Subject"] = subject
	header["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件消息
	messageBody := ""
	for k, v := range header {
		messageBody += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	messageBody += "\r\n" + body

	// 连接到SMTP服务器
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)

	server := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	err := smtp.SendMail(server, auth, e.config.From, strings.Split(e.config.To, ","), []byte(messageBody))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildSubject 构建邮件主题
func (e *EmailSender) buildSubject(message *notification.Message) string {
	subject := message.Title

	// 添加前缀
	if e.config.SubjectPrefix != "" {
		subject = fmt.Sprintf("%s %s", e.config.SubjectPrefix, subject)
	}

	// 添加级别标识
	switch message.Level {
	case notification.LevelError:
		subject = "🚨 " + subject
	case notification.LevelWarning:
		subject = "⚠️ " + subject
	case notification.LevelSuccess:
		subject = "✅ " + subject
	case notification.LevelInfo:
		subject = "ℹ️ " + subject
	}

	return subject
}

// buildBody 构建邮件正文
func (e *EmailSender) buildBody(message *notification.Message) string {
	// 构建HTML邮件内容
	body := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { border-bottom: 2px solid #2c3e50; margin-bottom: 20px; padding-bottom: 10px; }
        .header h1 { color: #2c3e50; margin: 0; }
        .content { line-height: 1.6; color: #333; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; color: #666; font-size: 12px; }
        .level-info { color: #3498db; }
        .level-warning { color: #f39c12; }
        .level-error { color: #e74c3c; }
        .level-success { color: #27ae60; }
        .link { color: #3498db; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="level-%s">%s</h1>
        </div>
        <div class="content">
            %s
        </div>
        %s
        <div class="footer">
            <p>此邮件由 MoviePilot 系统自动发送</p>
            <p>发送时间: %s</p>
            <p>如需停止接收此类通知，请联系系统管理员</p>
        </div>
    </div>
</body>
</html>`

	// 获取级别名称
	levelName := "info"
	switch message.Level {
	case notification.LevelInfo:
		levelName = "info"
	case notification.LevelWarning:
		levelName = "warning"
	case notification.LevelError:
		levelName = "error"
	case notification.LevelSuccess:
		levelName = "success"
	}

	// 构建链接部分（如果有）
	linkSection := ""
	if message.LinkURL != "" {
		linkSection = fmt.Sprintf(`<p><a href="%s" class="link">查看详情</a></p>`, message.LinkURL)
	}

	// 将内容中的换行符转换为HTML段落
	content := strings.ReplaceAll(message.Content, "\n", "</p><p>")
	content = fmt.Sprintf("<p>%s</p>", content)

	return fmt.Sprintf(body, message.Title, levelName, message.Title, content, linkSection, time.Now().Format("2006-01-02 15:04:05"))
}

// HealthCheck 健康检查
func (e *EmailSender) HealthCheck(ctx context.Context) error {
	// 测试连接SMTP服务器
	if e.config.SMTPHost == "" || e.config.SMTPPort == 0 {
		return fmt.Errorf("email configuration is incomplete")
	}

	// 尝试简单的连接测试
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort))
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// 尝试认证
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	return nil
}

// Close 关闭发送器
func (e *EmailSender) Close() error {
	return nil
}

// CreateEmailChannel 创建邮件通知渠道
func CreateEmailChannel(config *EmailConfig) *notification.Channel {
	enabled := config.SMTPHost != "" && config.SMTPPort > 0 &&
		config.Username != "" && config.Password != "" &&
		config.From != "" && config.To != ""

	return &notification.Channel{
		Name:        "email",
		Description: "邮件通知渠道",
		Enabled:     enabled,
		Sender:      NewEmailSender(config),
		Config: map[string]string{
			"smtp_host":      config.SMTPHost,
			"smtp_port":      fmt.Sprintf("%d", config.SMTPPort),
			"from":           config.From,
			"to":             config.To,
			"subject_prefix": config.SubjectPrefix,
		},
	}
}
