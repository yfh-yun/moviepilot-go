package naming

import (
	"reflect"
	"strings"
	"text/template"

	"moviepilot-go/internal/models/database"
)

// Template 命名模板
type Template struct {
	pattern string
	tmpl    *template.Template
}

// ParseTemplate 解析模板
func ParseTemplate(pattern string) (*Template, error) {
	// 转换模板格式：{title} -> {{ .Title }}
	// 首先处理带格式的变量，如 {season:02d}
	formattedPattern := pattern
	
	// 转换 {title} -> {{ .Title }}
	formattedPattern = strings.ReplaceAll(formattedPattern, "{title}", "{{ .Title }}")
	formattedPattern = strings.ReplaceAll(formattedPattern, "{season}", "{{ .Season }}")
	formattedPattern = strings.ReplaceAll(formattedPattern, "{episode}", "{{ .Episode }}")
	formattedPattern = strings.ReplaceAll(formattedPattern, "{year}", "{{ .Year }}")
	
	// 转换 {season:02d} -> {{ printf "%02d" .Season }}
	formattedPattern = strings.ReplaceAll(formattedPattern, "{season:02d}", "{{ printf \"%02d\" .Season }}")
	formattedPattern = strings.ReplaceAll(formattedPattern, "{episode:02d}", "{{ printf \"%02d\" .Episode }}")
	
	// 解析模板
	tmpl, err := template.New("naming").Parse(formattedPattern)
	if err != nil {
		return nil, err
	}
	
	return &Template{
		pattern: pattern,
		tmpl:    tmpl,
	}, nil
}

// MediaToVars 将媒体转换为模板变量
func MediaToVars(media any, sourcePath string) map[string]any {
	vars := make(map[string]any)
	
	// 处理 database.Media 类型
	if dbMedia, ok := media.(*database.Media); ok {
		vars["Title"] = dbMedia.Title
		vars["OriginalTitle"] = dbMedia.OriginalTitle
		if dbMedia.Year != nil {
			vars["Year"] = *dbMedia.Year
		} else {
			vars["Year"] = ""
		}
		vars["Type"] = dbMedia.Type
		if dbMedia.Season != nil {
			vars["Season"] = *dbMedia.Season
		} else {
			vars["Season"] = 0
		}
		if dbMedia.Episode != nil {
			vars["Episode"] = *dbMedia.Episode
		} else {
			vars["Episode"] = 0
		}
		return vars
	}
	
	// 处理其他类型，使用反射
	v := reflect.ValueOf(media)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return vars
	}
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i)
		
		// 获取 JSON 标签作为变量名
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		
		// 提取变量名（忽略标签中的其他选项）
		varName := strings.Split(jsonTag, ",")[0]
		varName = strings.Title(varName) // 转换为驼峰命名
		
		// 根据字段类型处理值
		switch value.Kind() {
		case reflect.String:
			vars[varName] = value.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vars[varName] = value.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			vars[varName] = value.Uint()
		case reflect.Float32, reflect.Float64:
			vars[varName] = value.Float()
		case reflect.Bool:
			vars[varName] = value.Bool()
		case reflect.Ptr:
			if !value.IsNil() {
				// 递归处理指针指向的值
				switch value.Elem().Kind() {
				case reflect.String:
					vars[varName] = value.Elem().String()
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					vars[varName] = value.Elem().Int()
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					vars[varName] = value.Elem().Uint()
				case reflect.Float32, reflect.Float64:
					vars[varName] = value.Elem().Float()
				case reflect.Bool:
					vars[varName] = value.Elem().Bool()
				}
			}
		}
	}
	
	return vars
}

// GetDefaultTemplate 获取默认模板
func GetDefaultTemplate(mediaType string) string {
	switch mediaType {
	case "movie":
		return "{{ .Title }} ({{ .Year }})"
	case "tv":
		return "{{ .Title }}/Season {{ .Season }}/{{ .Title }} - S{{ .Season }}E{{ .Episode }} - {{ .EpisodeTitle }}"
	case "anime":
		return "{{ .Title }}/{{ .Title }} - {{ .Episode }}.mkv"
	default:
		return "{{ .Title }}"
	}
}

// Render 渲染模板
func (t *Template) Render(vars map[string]any) (string, error) {
	var result strings.Builder
	err := t.tmpl.Execute(&result, vars)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

// Raw 获取模板原始字符串
func (t *Template) Raw() string {
	return t.pattern
}
