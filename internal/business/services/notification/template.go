package notification

import (
	"bytes"
	"fmt"
	"text/template"
)

// TemplateService 模板服务
type TemplateService interface {
	// Render 渲染模板
	Render(templateName string, data any) (string, error)

	// RegisterTemplate 注册模板
	RegisterTemplate(name, content string) error

	// GetTemplate 获取模板
	GetTemplate(name string) (*template.Template, error)
}

// templateService 模板服务实现
type templateService struct {
	templates map[string]*template.Template
}

// NewTemplateService 创建模板服务
func NewTemplateService() TemplateService {
	service := &templateService{
		templates: make(map[string]*template.Template),
	}

	// 注册默认模板
	service.registerDefaultTemplates()

	return service
}

// Render 渲染模板
func (s *templateService) Render(templateName string, data any) (string, error) {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return "", fmt.Errorf("模板不存在: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RegisterTemplate 注册模板
func (s *templateService) RegisterTemplate(name, content string) error {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return err
	}

	s.templates[name] = tmpl
	return nil
}

// GetTemplate 获取模板
func (s *templateService) GetTemplate(name string) (*template.Template, error) {
	tmpl, ok := s.templates[name]
	if !ok {
		return nil, fmt.Errorf("模板不存在: %s", name)
	}
	return tmpl, nil
}

// registerDefaultTemplates 注册默认模板
func (s *templateService) registerDefaultTemplates() {
	// 下载完成模板
	s.RegisterTemplate("download_complete", `
下载完成通知

标题: {{.Title}}
大小: {{.Size}}
保存路径: {{.SavePath}}
用时: {{.Duration}}
`)

	// 订阅更新模板
	s.RegisterTemplate("subscribe_update", `
订阅更新通知

订阅名称: {{.Name}}
新增集数: {{.Episodes}}
质量: {{.Quality}}
`)

	// 站点签到模板
	s.RegisterTemplate("site_checkin", `
站点签到通知

站点: {{.SiteName}}
状态: {{.Status}}
{{if .Bonus}}获得积分: {{.Bonus}}{{end}}
{{if .Upload}}获得上传: {{.Upload}}{{end}}
`)

	// 系统错误模板
	s.RegisterTemplate("system_error", `
系统错误通知

错误类型: {{.Type}}
错误信息: {{.Message}}
发生时间: {{.Time}}
`)
}

// DownloadCompleteData 下载完成数据
type DownloadCompleteData struct {
	Title    string
	Size     string
	SavePath string
	Duration string
}

// SubscribeUpdateData 订阅更新数据
type SubscribeUpdateData struct {
	Name     string
	Episodes string
	Quality  string
}

// SiteCheckinData 站点签到数据
type SiteCheckinData struct {
	SiteName string
	Status   string
	Bonus    string
	Upload   string
}

// SystemErrorData 系统错误数据
type SystemErrorData struct {
	Type    string
	Message string
	Time    string
}

// BuildNotificationFromTemplate 从模板构建通知
func BuildNotificationFromTemplate(templateService TemplateService, templateName string, data any, notifType NotificationType) (*Notification, error) {
	content, err := templateService.Render(templateName, data)
	if err != nil {
		return nil, err
	}

	// 根据模板名称生成标题
	title := getTemplateTitle(templateName)

	return &Notification{
		Title:    title,
		Content:  content,
		Type:     notifType,
		Priority: PriorityNormal,
	}, nil
}

// getTemplateTitle 获取模板标题
func getTemplateTitle(templateName string) string {
	titles := map[string]string{
		"download_complete": "下载完成",
		"subscribe_update":  "订阅更新",
		"site_checkin":      "站点签到",
		"system_error":      "系统错误",
	}

	if title, ok := titles[templateName]; ok {
		return title
	}
	return "通知"
}
