package template

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	textTemplate "text/template"

	"go.uber.org/zap"

	"moviepilot-go/internal/models"
)

// MessageTemplate 消息模板接口
type MessageTemplate interface {
	Render(templateName string, data interface{}) (string, error)
	RenderHTML(templateName string, data interface{}) (string, error)
	Parse(templateString string) (Template, error)
	RegisterTemplate(name string, templateString string) error
	GetTemplate(name string) (Template, error)
}

// Template 模板接口
type Template interface {
	Execute(data interface{}) (string, error)
	ExecuteHTML(data interface{}) (string, error)
}

// GoTemplate Go模板实现
type GoTemplate struct {
	textTemplates map[string]*textTemplate.Template
	htmlTemplates map[string]*template.Template
	logger        *zap.Logger
}

// NewGoTemplate 创建Go模板引擎
func NewGoTemplate(logger *zap.Logger) *GoTemplate {
	return &GoTemplate{
		textTemplates: make(map[string]*textTemplate.Template),
		htmlTemplates: make(map[string]*template.Template),
		logger:        logger.Named("go_template"),
	}
}

// Render 渲染文本模板
func (g *GoTemplate) Render(templateName string, data interface{}) (string, error) {
	g.logger.Debug("渲染文本模板", "name", templateName)

	tmpl, err := g.getTextTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("获取文本模板失败: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		g.logger.Error("执行文本模板失败", "name", templateName, "error", err)
		return "", fmt.Errorf("执行文本模板失败: %w", err)
	}

	g.logger.Debug("文本模板渲染成功", "name", templateName, "length", buf.Len())
	return buf.String(), nil
}

// RenderHTML 渲染HTML模板
func (g *GoTemplate) RenderHTML(templateName string, data interface{}) (string, error) {
	g.logger.Debug("渲染HTML模板", "name", templateName)

	tmpl, err := g.getHTMLTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("获取HTML模板失败: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		g.logger.Error("执行HTML模板失败", "name", templateName, "error", err)
		return "", fmt.Errorf("执行HTML模板失败: %w", err)
	}

	g.logger.Debug("HTML模板渲染成功", "name", templateName, "length", buf.Len())
	return buf.String(), nil
}

// Parse 解析模板字符串
func (g *GoTemplate) Parse(templateString string) (Template, error) {
	g.logger.Debug("解析模板字符串")

	// 解析文本模板
	textTmpl, err := textTemplate.New("").Parse(templateString)
	if err != nil {
		g.logger.Error("解析文本模板失败", "error", err)
		return nil, fmt.Errorf("解析文本模板失败: %w", err)
	}

	// 解析HTML模板
	htmlTmpl, err := template.New("").Parse(templateString)
	if err != nil {
		g.logger.Error("解析HTML模板失败", "error", err)
		return nil, fmt.Errorf("解析HTML模板失败: %w", err)
	}

	return &GoTemplateInstance{
		textTemplate: textTmpl,
		htmlTemplate: htmlTmpl,
		logger:       g.logger,
	}, nil
}

// RegisterTemplate 注册模板
func (g *GoTemplate) RegisterTemplate(name string, templateString string) error {
	g.logger.Debug("注册模板", "name", name)

	// 解析文本模板
	textTmpl, err := textTemplate.New(name).Parse(templateString)
	if err != nil {
		g.logger.Error("注册文本模板失败", "name", name, "error", err)
		return fmt.Errorf("注册文本模板失败: %w", err)
	}

	// 解析HTML模板
	htmlTmpl, err := template.New(name).Parse(templateString)
	if err != nil {
		g.logger.Error("注册HTML模板失败", "name", name, "error", err)
		return fmt.Errorf("注册HTML模板失败: %w", err)
	}

	g.textTemplates[name] = textTmpl
	g.htmlTemplates[name] = htmlTmpl

	g.logger.Debug("模板注册成功", "name", name)
	return nil
}

// GetTemplate 获取模板
func (g *GoTemplate) GetTemplate(name string) (Template, error) {
	textTmpl, err := g.getTextTemplate(name)
	if err != nil {
		return nil, err
	}

	htmlTmpl, err := g.getHTMLTemplate(name)
	if err != nil {
		return nil, err
	}

	return &GoTemplateInstance{
		textTemplate: textTmpl,
		htmlTemplate: htmlTmpl,
		logger:       g.logger,
	}, nil
}

// getTextTemplate 获取文本模板
func (g *GoTemplate) getTextTemplate(name string) (*textTemplate.Template, error) {
	tmpl, exists := g.textTemplates[name]
	if !exists {
		return nil, fmt.Errorf("文本模板不存在: %s", name)
	}
	return tmpl, nil
}

// getHTMLTemplate 获取HTML模板
func (g *GoTemplate) getHTMLTemplate(name string) (*template.Template, error) {
	tmpl, exists := g.htmlTemplates[name]
	if !exists {
		return nil, fmt.Errorf("HTML模板不存在: %s", name)
	}
	return tmpl, nil
}

// GoTemplateInstance Go模板实例
type GoTemplateInstance struct {
	textTemplate *textTemplate.Template
	htmlTemplate *template.Template
	logger       *zap.Logger
}

// Execute 执行文本模板
func (g *GoTemplateInstance) Execute(data interface{}) (string, error) {
	var buf bytes.Buffer
	err := g.textTemplate.Execute(&buf, data)
	if err != nil {
		g.logger.Error("执行模板失败", "error", err)
		return "", fmt.Errorf("执行模板失败: %w", err)
	}
	return buf.String(), nil
}

// ExecuteHTML 执行HTML模板
func (g *GoTemplateInstance) ExecuteHTML(data interface{}) (string, error) {
	var buf bytes.Buffer
	err := g.htmlTemplate.Execute(&buf, data)
	if err != nil {
		g.logger.Error("执行HTML模板失败", "error", err)
		return "", fmt.Errorf("执行HTML模板失败: %w", err)
	}
	return buf.String(), nil
}

// MessageTemplateHelper 消息模板助手
type MessageTemplateHelper struct {
	templateEngine MessageTemplate
	logger        *zap.Logger
}

// NewMessageTemplateHelper 创建消息模板助手
func NewMessageTemplateHelper(templateEngine MessageTemplate, logger *zap.Logger) *MessageTemplateHelper {
	return &MessageTemplateHelper{
		templateEngine: templateEngine,
		logger:        logger.Named("message_template_helper"),
	}
}

// RenderMediaNotification 渲染媒体通知消息
func (m *MessageTemplateHelper) RenderMediaNotification(notification *models.Notification, mediaInfo *models.MediaInfo) (string, error) {
	m.logger.Debug("渲染媒体通知消息", "title", notification.Title, "media", mediaInfo.Title)

	data := struct {
		*models.Notification
		MediaInfo *models.MediaInfo
	}{
		Notification: notification,
		MediaInfo:   mediaInfo,
	}

	return m.templateEngine.Render("media_notification", data)
}

// RenderTorrentNotification 渲染种子通知消息
func (m *MessageTemplateHelper) RenderTorrentNotification(notification *models.Notification, torrentInfo *models.TorrentInfo) (string, error) {
	m.logger.Debug("渲染种子通知消息", "title", notification.Title, "torrent", torrentInfo.Title)

	data := struct {
		*models.Notification
		TorrentInfo *models.TorrentInfo
	}{
		Notification: notification,
		TorrentInfo:  torrentInfo,
	}

	return m.templateEngine.Render("torrent_notification", data)
}

// RenderTransferNotification 渲染转移通知消息
func (m *MessageTemplateHelper) RenderTransferNotification(notification *models.Notification, transferInfo *models.TransferInfo) (string, error) {
	m.logger.Debug("渲染转移通知消息", "title", notification.Title, "path", transferInfo.TargetPath)

	data := struct {
		*models.Notification
		TransferInfo *models.TransferInfo
	}{
		Notification: notification,
		TransferInfo: transferInfo,
	}

	return m.templateEngine.Render("transfer_notification", data)
}

// RenderMediaList 渲染媒体列表
func (m *MessageTemplateHelper) RenderMediaList(notification *models.Notification, mediaList []*models.MediaInfo) (string, error) {
	m.logger.Debug("渲染媒体列表", "count", len(mediaList))

	data := struct {
		*models.Notification
		MediaList []*models.MediaInfo
	}{
		Notification: notification,
		MediaList:   mediaList,
	}

	return m.templateEngine.Render("media_list", data)
}

// RenderTorrentList 渲染种子列表
func (m *MessageTemplateHelper) RenderTorrentList(notification *models.Notification, torrentList []*models.TorrentInfo) (string, error) {
	m.logger.Debug("渲染种子列表", "count", len(torrentList))

	data := struct {
		*models.Notification
		TorrentList []*models.TorrentInfo
	}{
		Notification: notification,
		TorrentList:  torrentList,
	}

	return m.templateEngine.Render("torrent_list", data)
}

// InitDefaultTemplates 初始化默认模板
func (m *MessageTemplateHelper) InitDefaultTemplates() error {
	m.logger.Info("初始化默认消息模板")

	// 媒体通知模板
	mediaTemplate := `🎬 {{.Title}}

📺 {{.MediaInfo.Title}}
📅 上映年份: {{.MediaInfo.Year}}
⭐ 评分: {{.MediaInfo.Rating}}
🎭 类型: {{range .MediaInfo.Genres}}{{.}} {{end}}

{{if .Content}}📝 {{.Content}}{{end}}`

	// 种子通知模板
	torrentTemplate := `🌱 {{.Title}}

📦 {{.TorrentInfo.Title}}
📊 大小: {{.TorrentInfo.Size}}
👤 制作组: {{.TorrentInfo.Uploader}}
⏰ 时间: {{.TorrentInfo.PublishDate}}

{{if .Content}}📝 {{.Content}}{{end}}`

	// 转移通知模板
	transferTemplate := `📁 {{.Title}}

🔄 转移完成
📂 源路径: {{.TransferInfo.SourcePath}}
🎯 目标路径: {{.TransferInfo.TargetPath}}

{{if .Content}}📝 {{.Content}}{{end}}`

	// 媒体列表模板
	mediaListTemplate := `📋 {{.Title}}

{{range .MediaList}}🎬 {{.Title}} ({{.Year}})
⭐ {{.Rating}} | 🎭 {{range .Genres}}{{.}} {{end}}

{{end}}`

	// 种子列表模板
	torrentListTemplate := `🔍 {{.Title}}

{{range .TorrentList}}🌱 {{.Title}}
📊 {{.Size}} | 👤 {{.Uploader}}
⏰ {{.PublishDate}}

{{end}}`

	// 注册模板
	templates := map[string]string{
		"media_notification":    mediaTemplate,
		"torrent_notification": torrentTemplate,
		"transfer_notification": transferTemplate,
		"media_list":          mediaListTemplate,
		"torrent_list":        torrentListTemplate,
	}

	for name, templateString := range templates {
		err := m.templateEngine.RegisterTemplate(name, templateString)
		if err != nil {
			m.logger.Error("注册默认模板失败", "name", name, "error", err)
			return fmt.Errorf("注册默认模板失败 %s: %w", name, err)
		}
	}

	m.logger.Info("默认消息模板初始化完成")
	return nil
}

// FormatMessage 格式化消息
func (m *MessageTemplateHelper) FormatMessage(content string, maxLength int) string {
	if maxLength > 0 && len(content) > maxLength {
		content = content[:maxLength] + "..."
	}
	
	// 替换多余的换行符
	content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	
	return content
}