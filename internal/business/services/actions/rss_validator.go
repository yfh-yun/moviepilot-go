package actions

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// RSSValidator RSS验证器
type RSSValidator struct{}

// NewRSSValidator 创建RSS验证器实例
func NewRSSValidator() *RSSValidator {
	return &RSSValidator{}
}

// ValidateFetchParams 验证RSS获取参数
func (v *RSSValidator) ValidateFetchParams(params *FetchRSSParams) error {
	if params == nil {
		return errors.New("参数不能为空")
	}

	// 验证URL
	if err := v.validateURL(params.FeedURL); err != nil {
		return err
	}

	// 验证格式
	if err := v.validateFormat(params.Format); err != nil {
		return err
	}

	// 验证超时时间
	if err := v.validateTimeout(params.Timeout); err != nil {
		return err
	}

	// 验证重试次数
	if err := v.validateRetries(params.Retries); err != nil {
		return err
	}

	// 验证延迟时间
	if err := v.validateDelay(params.Delay); err != nil {
		return err
	}

	// 验证限制数量
	if err := v.validateLimit(params.Limit); err != nil {
		return err
	}

	// 验证缓存TTL
	if err := v.validateCacheTTL(params.CacheTTL); err != nil {
		return err
	}

	// 验证认证信息
	if err := v.validateAuth(params.Username, params.Password); err != nil {
		return err
	}

	// 验证过滤器
	if params.Filters != nil {
		if err := v.validateFilters(params.Filters); err != nil {
			return err
		}
	}

	// 验证自定义头
	if err := v.validateHeaders(params.Headers); err != nil {
		return err
	}

	return nil
}

// ValidateFeedURL 验证RSS源URL
func (v *RSSValidator) ValidateFeedURL(feedURL string) error {
	return v.validateURL(feedURL)
}

// ValidateRSSFilter 验证RSS过滤器
func (v *RSSValidator) ValidateRSSFilter(filter *RSSFilters) error {
	if filter == nil {
		return nil // 空过滤器是允许的
	}

	return v.validateFilters(filter)
}

// ValidateBatchParams 验证批量获取参数
func (v *RSSValidator) ValidateBatchParams(paramsList []*FetchRSSParams) error {
	if paramsList == nil || len(paramsList) == 0 {
		return errors.New("批量参数列表不能为空")
	}

	if len(paramsList) > 10 {
		return errors.New("批量参数数量不能超过10个")
	}

	// 验证每个参数
	for i, params := range paramsList {
		if err := v.ValidateFetchParams(params); err != nil {
			return fmt.Errorf("第%d个参数验证失败: %w", i+1, err)
		}
	}

	return nil
}

// 私有验证方法

// validateURL 验证URL格式
func (v *RSSValidator) validateURL(feedURL string) error {
	if feedURL == "" {
		return errors.New("RSS源URL不能为空")
	}

	// 检查URL长度
	if len(feedURL) > 2000 {
		return errors.New("RSS源URL长度不能超过2000个字符")
	}

	// 解析URL
	parsedURL, err := url.ParseRequestURI(feedURL)
	if err != nil {
		return fmt.Errorf("无效的RSS源URL: %w", err)
	}

	// 验证协议
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("RSS源URL必须使用HTTP或HTTPS协议")
	}

	// 验证主机名
	if parsedURL.Host == "" {
		return errors.New("RSS源URL必须包含主机名")
	}

	return nil
}

// validateFormat 验证RSS格式
func (v *RSSValidator) validateFormat(format RSSFormat) error {
	validFormats := []RSSFormat{RSSFormatXML, RSSFormatJSON, RSSFormatCustom}
	valid := false

	for _, validFormat := range validFormats {
		if format == validFormat {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("无效的RSS格式: %s，支持的格式: xml, json, custom", format)
	}

	return nil
}

// validateTimeout 验证超时时间
func (v *RSSValidator) validateTimeout(timeout int) error {
	if timeout < 0 {
		return errors.New("超时时间不能为负数")
	}

	if timeout > 300 { // 最大5分钟
		return errors.New("超时时间不能超过300秒")
	}

	return nil
}

// validateRetries 验证重试次数
func (v *RSSValidator) validateRetries(retries int) error {
	if retries < 0 {
		return errors.New("重试次数不能为负数")
	}

	if retries > 10 { // 最大10次重试
		return errors.New("重试次数不能超过10次")
	}

	return nil
}

// validateDelay 验证重试延迟
func (v *RSSValidator) validateDelay(delay int) error {
	if delay < 0 {
		return errors.New("重试延迟不能为负数")
	}

	if delay > 60 { // 最大1分钟延迟
		return errors.New("重试延迟不能超过60秒")
	}

	return nil
}

// validateLimit 验证返回条目数量限制
func (v *RSSValidator) validateLimit(limit int) error {
	if limit < 0 {
		return errors.New("返回条目数量限制不能为负数")
	}

	if limit > 500 { // 最大500条
		return errors.New("返回条目数量限制不能超过500条")
	}

	return nil
}

// validateCacheTTL 验证缓存时间
func (v *RSSValidator) validateCacheTTL(cacheTTL int) error {
	if cacheTTL < 0 {
		return errors.New("缓存时间不能为负数")
	}

	if cacheTTL > 1440 { // 最大24小时（分钟）
		return errors.New("缓存时间不能超过1440分钟")
	}

	return nil
}

// validateAuth 验证认证信息
func (v *RSSValidator) validateAuth(username, password string) error {
	// 检查用户名和密码是否同时提供
	if (username != "" && password == "") || (username == "" && password != "") {
		return errors.New("用户名和密码必须同时提供")
	}

	// 检查用户名长度
	if len(username) > 100 {
		return errors.New("用户名长度不能超过100个字符")
	}

	// 检查密码长度
	if len(password) > 200 {
		return errors.New("密码长度不能超过200个字符")
	}

	return nil
}

// validateFilters 验证过滤器
func (v *RSSValidator) validateFilters(filters *RSSFilters) error {
	// 验证标题关键词
	if err := v.validateKeywords(filters.IncludeTitle, "包含标题关键词"); err != nil {
		return err
	}

	if err := v.validateKeywords(filters.ExcludeTitle, "排除标题关键词"); err != nil {
		return err
	}

	// 验证内容关键词
	if err := v.validateKeywords(filters.IncludeKeywords, "包含内容关键词"); err != nil {
		return err
	}

	if err := v.validateKeywords(filters.ExcludeKeywords, "排除内容关键词"); err != nil {
		return err
	}

	// 验证文件大小
	if err := v.validateFileSize(filters.MinSize, filters.MaxSize); err != nil {
		return err
	}

	// 验证做种数
	if err := v.validateSeeders(filters.MinSeeders); err != nil {
		return err
	}

	// 验证媒体类型
	if err := v.validateMediaTypes(filters.MediaTypes); err != nil {
		return err
	}

	return nil
}

// validateKeywords 验证关键词列表
func (v *RSSValidator) validateKeywords(keywords []string, keywordType string) error {
	if keywords == nil {
		return nil
	}

	// 检查关键词数量
	if len(keywords) > 20 {
		return fmt.Errorf("%s数量不能超过20个", keywordType)
	}

	// 检查每个关键词
	for _, keyword := range keywords {
		// 去除空白字符
		keyword = strings.TrimSpace(keyword)

		if keyword == "" {
			return fmt.Errorf("%s不能为空", keywordType)
		}

		if len(keyword) > 50 {
			return fmt.Errorf("%s长度不能超过50个字符", keywordType)
		}

		// 检查是否包含特殊字符（可选，根据需求调整）
		if strings.ContainsAny(keyword, "\x00-\x1f\x7f") {
			return fmt.Errorf("%s不能包含控制字符", keywordType)
		}
	}

	return nil
}

// validateFileSize 验证文件大小限制
func (v *RSSValidator) validateFileSize(minSize, maxSize int64) error {
	// 检查最小大小
	if minSize < 0 {
		return errors.New("最小文件大小不能为负数")
	}

	// 检查最大大小
	if maxSize < 0 {
		return errors.New("最大文件大小不能为负数")
	}

	// 检查大小范围
	if minSize > 0 && maxSize > 0 && minSize > maxSize {
		return errors.New("最小文件大小不能大于最大文件大小")
	}

	// 检查最大值限制
	if maxSize > 10995116277760 { // 10TB
		return errors.New("最大文件大小不能超过10TB")
	}

	return nil
}

// validateSeeders 验证做种数限制
func (v *RSSValidator) validateSeeders(minSeeders int) error {
	if minSeeders < 0 {
		return errors.New("最小做种数不能为负数")
	}

	if minSeeders > 10000 {
		return errors.New("最小做种数不能超过10000")
	}

	return nil
}

// validateMediaTypes 验证媒体类型
func (v *RSSValidator) validateMediaTypes(mediaTypes []string) error {
	if mediaTypes == nil {
		return nil
	}

	// 检查媒体类型数量
	if len(mediaTypes) > 5 {
		return errors.New("媒体类型数量不能超过5个")
	}

	// 检查有效的媒体类型
	validMediaTypes := []RSSType{
		RSSTypeMovie,
		RSSTypeSeries,
		RSSTypeAnimation,
		RSSTypeDocumentary,
		RSSTypeOther,
	}

	for _, mediaType := range mediaTypes {
		valid := false
		for _, validType := range validMediaTypes {
			if mediaType == string(validType) {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("无效的媒体类型: %s", mediaType)
		}
	}

	return nil
}

// validateHeaders 验证自定义头
func (v *RSSValidator) validateHeaders(headers map[string]string) error {
	if headers == nil {
		return nil
	}

	// 检查头数量
	if len(headers) > 20 {
		return errors.New("自定义头数量不能超过20个")
	}

	// 检查每个头
	for key, value := range headers {
		// 检查头名称
		if key == "" {
			return errors.New("自定义头名称不能为空")
		}

		if len(key) > 50 {
			return errors.New("自定义头名称长度不能超过50个字符")
		}

		// 检查是否为保留头
		reservedHeaders := []string{
			"Host", "Content-Length", "Transfer-Encoding", "Connection",
			"Upgrade", "TE", "Trailer", "Content-Encoding",
			"Content-Type", "Authorization", // 授权头由系统处理
		}

		keyLower := strings.ToLower(key)
		for _, reserved := range reservedHeaders {
			if keyLower == strings.ToLower(reserved) {
				return fmt.Errorf("自定义头 '%s' 是保留头，不能自定义", key)
			}
		}

		// 检查头值
		if len(value) > 1000 {
			return fmt.Errorf("自定义头 '%s' 的值长度不能超过1000个字符", key)
		}

		// 检查是否包含控制字符
		if strings.ContainsAny(value, "\x00-\x1f\x7f") {
			return fmt.Errorf("自定义头 '%s' 的值不能包含控制字符", key)
		}
	}

	return nil
}

// ValidateUserAgent 验证User-Agent
func (v *RSSValidator) ValidateUserAgent(userAgent string) error {
	if userAgent == "" {
		return nil // 空User-Agent是允许的
	}

	if len(userAgent) > 500 {
		return errors.New("User-Agent长度不能超过500个字符")
	}

	// 检查是否包含控制字符
	if strings.ContainsAny(userAgent, "\x00-\x1f\x7f") {
		return errors.New("User-Agent不能包含控制字符")
	}

	return nil
}

// ValidateCookies 验证Cookies
func (v *RSSValidator) ValidateCookies(cookies string) error {
	if cookies == "" {
		return nil // 空Cookies是允许的
	}

	if len(cookies) > 2000 {
		return errors.New("Cookies长度不能超过2000个字符")
	}

	// 基本的Cookies格式验证
	// 实际项目中可能需要更严格的验证
	if strings.ContainsAny(cookies, "\x00-\x1f\x7f") {
		return errors.New("Cookies不能包含控制字符")
	}

	return nil
}

// ValidateRSSResponse 验证RSS响应
func (v *RSSValidator) ValidateRSSResponse(response *RSSResponse) error {
	if response == nil {
		return errors.New("RSS响应不能为空")
	}

	// 检查条目数量
	if len(response.Entries) > 1000 {
		return errors.New("RSS响应条目数量不能超过1000条")
	}

	// 检查错误信息
	if !response.Success && response.Error == nil {
		return errors.New("失败的响应必须包含错误信息")
	}

	// 验证条目
	for i, entry := range response.Entries {
		if err := v.validateEntry(&entry); err != nil {
			return fmt.Errorf("第%d个条目验证失败: %w", i+1, err)
		}
	}

	return nil
}

// validateEntry 验证RSS条目
func (v *RSSValidator) validateEntry(entry *RSSEntry) error {
	if entry == nil {
		return errors.New("RSS条目不能为空")
	}

	// 检查ID
	if entry.ID == "" {
		return errors.New("RSS条目ID不能为空")
	}

	// 检查标题
	if entry.Title == "" {
		return errors.New("RSS条目标题不能为空")
	}

	if len(entry.Title) > 500 {
		return errors.New("RSS条目标题长度不能超过500个字符")
	}

	// 检查链接
	if entry.Link != "" {
		if err := v.validateURL(entry.Link); err != nil {
			return fmt.Errorf("无效的RSS条目链接: %w", err)
		}
	}

	// 检查磁力链接
	if entry.MagnetURL != "" && !strings.HasPrefix(entry.MagnetURL, "magnet:") {
		return errors.New("无效的磁力链接")
	}

	// 检查种子链接
	if entry.TorrentURL != "" {
		if err := v.validateURL(entry.TorrentURL); err != nil {
			return fmt.Errorf("无效的种子链接: %w", err)
		}
		if !strings.HasSuffix(entry.TorrentURL, ".torrent") {
			return errors.New("种子链接必须以.torrent结尾")
		}
	}

	// 检查发布时间
	if !entry.PublishedAt.IsZero() && entry.PublishedAt.After(time.Now().Add(24 * time.Hour)) {
		return errors.New("RSS条目发布时间不能是未来时间超过24小时")
	}

	return nil
}
