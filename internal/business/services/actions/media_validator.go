package actions

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// MediaValidator 媒体参数验证器
type MediaValidator struct{}

// NewMediaValidator 创建媒体验证器实例
func NewMediaValidator() *MediaValidator {
	return &MediaValidator{}
}

// ValidateFetchMediasParams 验证获取媒体参数
func (v *MediaValidator) ValidateFetchMediasParams(params *FetchMediasParams) error {
	if params == nil {
		return errors.New("参数不能为空")
	}

	// 验证数据源类型
	if params.SourceType != "" {
		validSourceTypes := []string{"recommend", "search", "detail", "custom"}
		valid := false
		for _, t := range validSourceTypes {
			if params.SourceType == t {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("无效的数据源类型，支持: recommend, search, detail, custom")
		}
	}

	// 验证数量限制
	if params.Limit < 0 {
		return errors.New("数量限制不能为负数")
	}
	if params.Limit > 100 {
		return errors.New("数量限制不能超过100")
	}

	// 验证评分
	if params.Rating < 0 || params.Rating > 10 {
		return errors.New("评分必须在0-10之间")
	}

	// 验证排序方式
	if params.SortBy != "" {
		validSortBy := []string{"title", "rating", "year", "release_date", "popularity"}
		valid := false
		for _, s := range validSortBy {
			if params.SortBy == s {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("无效的排序方式")
		}
	}

	// 验证排序顺序
	if params.OrderBy != "" {
		validOrderBy := []string{"asc", "desc"}
		valid := false
		for _, o := range validOrderBy {
			if params.OrderBy == o {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("无效的排序顺序，支持: asc, desc")
		}
	}

	// 自定义API路径验证
	if params.APIPath != "" && !strings.HasPrefix(params.APIPath, "/") {
		return errors.New("自定义API路径必须以/开头")
	}

	return nil
}

// ValidateSearchParams 验证搜索参数
func (v *MediaValidator) ValidateSearchParams(query string, mediaType string, year int) error {
	// 搜索关键词验证
	if query != "" {
		if len(strings.TrimSpace(query)) < 2 {
			return errors.New("搜索关键词至少需要2个字符")
		}
		if len(query) > 100 {
			return errors.New("搜索关键词不能超过100个字符")
		}
	}

	// 媒体类型验证
	if mediaType != "" {
		validTypes := []string{"movie", "tv", "series", "anime", "documentary"}
		valid := false
		for _, t := range validTypes {
			if mediaType == t {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("无效的媒体类型")
		}
	}

	// 年份验证
	if year != 0 {
		currentYear := 2024 // 实际实现中应该使用当前年份
		if year < 1895 || year > currentYear+1 {
			return errors.New("无效的年份")
		}
	}

	return nil
}

// ValidateMediaID 验证媒体ID
func (v *MediaValidator) ValidateMediaID(mediaID string) error {
	if mediaID == "" {
		return errors.New("媒体ID不能为空")
	}

	if len(mediaID) > 100 {
		return errors.New("媒体ID长度不能超过100个字符")
	}

	// 检查ID格式（可以根据实际需求调整）
	// 允许字母、数字、下划线、连字符和冒号
	for _, char := range mediaID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-' || char == ':') {
			return errors.New("媒体ID包含无效字符")
		}
	}

	return nil
}

// ValidateSeasonAndEpisode 验证季节和剧集编号
func (v *MediaValidator) ValidateSeasonAndEpisode(seasonNumber, episodeNumber int) error {
	if seasonNumber < 0 {
		return errors.New("季节编号不能为负数")
	}

	if seasonNumber > 100 {
		return errors.New("季节编号不能超过100")
	}

	if episodeNumber < 0 {
		return errors.New("剧集编号不能为负数")
	}

	if episodeNumber > 999 {
		return errors.New("剧集编号不能超过999")
	}

	return nil
}

// RegisterValidators 注册验证器到Gin
func RegisterValidators() error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证函数
		// 验证媒体类型
		if err := v.RegisterValidation("media_type", validateMediaType); err != nil {
			return err
		}

		// 验证排序方式
		if err := v.RegisterValidation("sort_by", validateSortBy); err != nil {
			return err
		}

		// 验证排序顺序
		if err := v.RegisterValidation("order_by", validateOrderBy); err != nil {
			return err
		}

		return nil
	}

	return errors.New("无法获取验证器引擎")
}

// 自定义验证函数

// validateMediaType 验证媒体类型
func validateMediaType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // 允许为空
	}

	validTypes := []string{"movie", "tv", "series", "anime", "documentary"}
	for _, t := range validTypes {
		if value == t {
			return true
		}
	}

	return false
}

// validateSortBy 验证排序方式
func validateSortBy(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // 允许为空
	}

	validSortBy := []string{"title", "rating", "year", "release_date", "popularity"}
	for _, s := range validSortBy {
		if value == s {
			return true
		}
	}

	return false
}

// validateOrderBy 验证排序顺序
func validateOrderBy(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // 允许为空
	}

	return value == "asc" || value == "desc"
}
