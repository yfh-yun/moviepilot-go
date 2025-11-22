package naming

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"moviepilot-go/internal/models"
	"moviepilot-go/pkg/logger"

	"go.uber.org/zap"
)

// Template 命名模板
type Template struct {
	raw       string
	variables []string
}

// TemplateVars 模板变量集合
type TemplateVars struct {
	// 通用变量
	Title         string
	OriginalTitle string
	Year          string
	Type          string
	Extension     string

	// 电影变量
	Resolution string
	Source     string
	Codec      string
	Audio      string
	Subtitle   string
	Group      string

	// 电视剧变量
	Season       string // S01
	SeasonNum    string // 1
	Episode      string // E01
	EpisodeNum   string // 1
	EpisodeTitle string

	// 额外信息
	TMDBID string
	IMDBID string
}

var (
	// 变量匹配正则: ${variable_name}
	varPattern = regexp.MustCompile(`\$\{([a-z_]+)\}`)
)

// ParseTemplate 解析模板字符串
func ParseTemplate(template string) (*Template, error) {
	if template == "" {
		return nil, fmt.Errorf("template cannot be empty")
	}

	// 提取所有变量
	matches := varPattern.FindAllStringSubmatch(template, -1)
	variables := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			variables = append(variables, match[1])
		}
	}

	return &Template{
		raw:       template,
		variables: variables,
	}, nil
}

// Render 渲染模板
func (t *Template) Render(vars TemplateVars) (string, error) {
	result := t.raw

	// 替换所有变量
	varMap := t.buildVarMap(vars)
	for varName, varValue := range varMap {
		placeholder := fmt.Sprintf("${%s}", varName)
		result = strings.ReplaceAll(result, placeholder, varValue)
	}

	// 检查是否还有未替换的变量
	if varPattern.MatchString(result) {
		if logger.Logger != nil {
			logger.Logger.Warn("template has unresolved variables",
				zap.String("template", t.raw),
				zap.String("result", result))
		}
	}

	// 清理路径
	result = cleanPath(result)

	return result, nil
}

// buildVarMap 构建变量映射
func (t *Template) buildVarMap(vars TemplateVars) map[string]string {
	return map[string]string{
		// 通用变量
		"title":          sanitizeFileName(vars.Title),
		"original_title": sanitizeFileName(vars.OriginalTitle),
		"year":           vars.Year,
		"type":           vars.Type,
		"ext":            vars.Extension,

		// 电影变量
		"resolution": vars.Resolution,
		"source":     vars.Source,
		"codec":      vars.Codec,
		"audio":      vars.Audio,
		"subtitle":   vars.Subtitle,
		"group":      vars.Group,

		// 电视剧变量
		"season":        vars.Season,
		"season_num":    vars.SeasonNum,
		"episode":       vars.Episode,
		"episode_num":   vars.EpisodeNum,
		"episode_title": sanitizeFileName(vars.EpisodeTitle),

		// 额外信息
		"tmdbid": vars.TMDBID,
		"imdbid": vars.IMDBID,
	}
}

// Variables 返回模板中使用的变量列表
func (t *Template) Variables() []string {
	return t.variables
}

// Raw 返回原始模板字符串
func (t *Template) Raw() string {
	return t.raw
}

// MediaToVars 将 Media 模型转换为模板变量
func MediaToVars(media *models.Media, sourcePath string) TemplateVars {
	vars := TemplateVars{
		Title:         media.Title,
		OriginalTitle: media.OriginalTitle,
		Type:          media.Type,
		Extension:     filepath.Ext(sourcePath),
	}

	// 年份
	if media.Year != nil {
		vars.Year = *media.Year
	}

	// 季集信息
	if media.Season != nil {
		vars.SeasonNum = strconv.Itoa(*media.Season)
		vars.Season = fmt.Sprintf("S%02d", *media.Season)
	}
	if media.Episode != nil {
		vars.EpisodeNum = strconv.Itoa(*media.Episode)
		vars.Episode = fmt.Sprintf("E%02d", *media.Episode)
	}
	// EpisodeTitle 需要从其他地方获取，Media 模型中没有这个字段
	// 可以在调用时手动设置

	// ID信息
	if media.TMDBID != nil {
		vars.TMDBID = strconv.Itoa(*media.TMDBID)
	}
	if media.IMDBID != nil {
		vars.IMDBID = *media.IMDBID
	}

	return vars
}

// sanitizeFileName 清理文件名中的非法字符
func sanitizeFileName(name string) string {
	if name == "" {
		return ""
	}

	// 移除或替换非法字符: <>:"/\|?*
	replacements := map[string]string{
		"<":  "",
		">":  "",
		":":  " -",
		"\"": "'",
		"/":  "-",
		"\\": "-",
		"|":  "-",
		"?":  "",
		"*":  "",
	}

	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// 移除前后空格
	result = strings.TrimSpace(result)

	// 限制长度 (保留扩展名空间)
	maxLen := 200
	if len(result) > maxLen {
		result = result[:maxLen]
	}

	return result
}

// cleanPath 清理路径
func cleanPath(path string) string {
	// 移除多余的斜杠
	path = filepath.Clean(path)

	// 移除连续的空格
	path = regexp.MustCompile(`\s+`).ReplaceAllString(path, " ")

	// 移除路径中的 ".."
	path = strings.ReplaceAll(path, "..", ".")

	return path
}

// DefaultTemplates 默认模板
var DefaultTemplates = map[string]string{
	"movie": "${title} (${year})/${title}.${year}.${resolution}.${source}.${codec}${ext}",
	"tv":    "${title}/Season ${season_num}/${title}.${season}${episode}.${episode_title}${ext}",
	"anime": "${title}/Season ${season_num}/[${group}] ${title} - ${episode_num}${ext}",
}

// GetDefaultTemplate 获取默认模板
func GetDefaultTemplate(mediaType string) string {
	if template, ok := DefaultTemplates[mediaType]; ok {
		return template
	}
	return DefaultTemplates["movie"]
}
