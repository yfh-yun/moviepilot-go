package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"moviepilot-go/internal/config"
	"moviepilot-go/internal/core"
	"moviepilot-go/internal/db"
	"moviepilot-go/internal/logger"
	"moviepilot-go/internal/utils"
	"moviepilot-go/pkg/models"
)

// TemplateContextBuilder 模板上下文构建器
type TemplateContextBuilder struct {
	context map[string]interface{}
}

// NewTemplateContextBuilder 创建一个新的模板上下文构建器实�?func NewTemplateContextBuilder() *TemplateContextBuilder {
	return &TemplateContextBuilder{
		context: make(map[string]interface{}),
	}
}

// Build 构建模板上下�?func (tcb *TemplateContextBuilder) Build(
	meta *models.MetaBase,
	mediainfo *models.MediaInfo,
	torrentinfo *models.TorrentInfo,
	transferinfo *models.TransferInfo,
	fileExtension *string,
	episodesInfo []models.TmdbEpisode,
	includeRawObjects bool,
	additional map[string]interface{},
) map[string]interface{} {
	/*
	 * :param meta: 媒体信息
	 * :param mediainfo: 媒体信息
	 * :param torrentinfo: 种子信息
	 * :param transferinfo: 传输信息
	 * :param fileExtension: 文件扩展�?	 * :param episodesInfo: 剧集信息
	 * :param includeRawObjects: 是否包含原始对象
	 * :param additional: 额外参数
	 * :return: 渲染上下文字�?	 */
	tcb.context = make(map[string]interface{})
	tcb.addEpisodeDetails(meta, episodesInfo)
	tcb.addMediaInfo(mediainfo)
	tcb.addTransferInfo(transferinfo)
	tcb.addTorrentInfo(torrentinfo)
	tcb.addFileInfo(fileExtension)
	
	if additional != nil {
		for k, v := range additional {
			tcb.context[k] = v
		}
	}
	
	if includeRawObjects {
		tcb.addRawObjects(meta, mediainfo, torrentinfo, transferinfo, episodesInfo)
	}
	
	// 移除空�?	result := make(map[string]interface{})
	for k, v := range tcb.context {
		if v != nil {
			result[k] = v
		}
	}
	return result
}

// addMediaInfo 增加媒体信息
func (tcb *TemplateContextBuilder) addMediaInfo(mediainfo *models.MediaInfo) {
	/*
	 * 增加媒体信息
	 */
	if mediainfo == nil {
		return
	}
	
	var seasonFmt *string
	if mediainfo.Season != nil {
		sf := fmt.Sprintf("S%02d", *mediainfo.Season)
		seasonFmt = &sf
	}
	
	baseInfo := map[string]interface{}{
		// 标题
		"title": tcb.convertInvalidCharacters(mediainfo.Title),
		// 英文标题
		"en_title": tcb.convertInvalidCharacters(mediainfo.EnTitle),
		// 原语种标�?		"original_title": tcb.convertInvalidCharacters(mediainfo.OriginalTitle),
		// 季号
		"season": utils.ValueOrDefault(tcb.context["season"], mediainfo.Season),
		// Sxx
		"season_fmt": utils.ValueOrDefault(tcb.context["season_fmt"], seasonFmt),
		// 年份
		"year": utils.ValueOrDefault(mediainfo.Year, tcb.context["year"]),
		// 媒体标题 + 年份
		"title_year": utils.ValueOrDefault(mediainfo.TitleYear, tcb.context["title_year"]),
	}
	
	metaSeason := tcb.context["season"]
	mediaInfo := map[string]interface{}{
		// 类型
		"type": mediainfo.Type,
		// 类别
		"category": mediainfo.Category,
		// 评分
		"vote_average": mediainfo.VoteAverage,
		// 海报
		"poster": mediainfo.PosterPath,
		// 背景�?		"backdrop": mediainfo.BackdropPath,
		// 季年份根据season值获�?		"season_year": tcb.getSeasonYear(mediainfo.SeasonYears, metaSeason),
		// 演员
		"actors": tcb.formatActors(mediainfo.Actors),
		// 简�?		"overview": mediainfo.Overview,
		// TMDBID
		"tmdbid": mediainfo.TmdbID,
		// IMDBID
		"imdbid": mediainfo.ImdbID,
		// 豆瓣ID
		"doubanid": mediainfo.DoubanID,
	}
	
	for k, v := range baseInfo {
		tcb.context[k] = v
	}
	for k, v := range mediaInfo {
		tcb.context[k] = v
	}
}

// getSeasonYear 获取季年�?func (tcb *TemplateContextBuilder) getSeasonYear(seasonYears map[int]int, metaSeason interface{}) interface{} {
	if seasonYears == nil || metaSeason == nil {
		return nil
	}
	
	if season, ok := metaSeason.(int); ok {
		if year, exists := seasonYears[season]; exists {
			return year
		}
	}
	return nil
}

// formatActors 格式化演员列�?func (tcb *TemplateContextBuilder) formatActors(actors []map[string]interface{}) string {
	if len(actors) == 0 {
		return ""
	}
	
	var names []string
	limit := len(actors)
	if limit > 5 {
		limit = 5
	}
	
	for i := 0; i < limit; i++ {
		if name, ok := actors[i]["name"].(string); ok {
			names = append(names, name)
		}
	}
	
	return strings.Join(names, "�?")
}

// addEpisodeDetails 添加剧集详细信息
func (tcb *TemplateContextBuilder) addEpisodeDetails(meta *models.MetaBase, episodes []models.TmdbEpisode) {
	/*
	 * 添加剧集详细信息
	 */
	if meta == nil {
		return
	}
	
	episodeData := map[string]interface{}{
		"episode_title": nil,
		"episode_date":  nil,
	}
	
	if meta.BeginEpisode != nil && episodes != nil {
		for _, episode := range episodes {
			if episode.EpisodeNumber == *meta.BeginEpisode {
				episodeData["episode_title"] = tcb.convertInvalidCharacters(episode.Name)
				episodeData["episode_date"] = episode.AirDate
				break
			}
		}
	}
	
	var titleYear string
	if meta.Year != "" {
		titleYear = fmt.Sprintf("%s (%s)", meta.Name, meta.Year)
	} else {
		titleYear = meta.Name
	}
	
	metaInfo := map[string]interface{}{
		// 原文件名
		"original_name": meta.Title,
		// 识别名称（优先使用中文）
		"name": meta.Name,
		// 识别的英文名称（可能为空�?		"en_name": meta.EnName,
		// 年份
		"year": meta.Year,
		// 名字 + 年份
		"title_year": utils.ValueOrDefault(tcb.context["title_year"], titleYear),
		// 季号
		"season": meta.SeasonSeq,
		// Sxx
		"season_fmt": meta.Season,
		// 集号
		"episode": meta.EpisodeSeqs,
		// 季集 SxxExx
		"season_episode": fmt.Sprintf("%s%s", meta.Season, meta.Episode),
		// �?�?		"part": meta.Part,
		// 自定义占位符
		"customization": meta.Customization,
	}
	
	techMetadata := map[string]interface{}{
		// 资源类型
		"resourceType": meta.ResourceType,
		// 特效
		"effect": meta.ResourceEffect,
		// 版本
		"edition": meta.Edition,
		// 分辨�?		"videoFormat": meta.ResourcePix,
		// 质量
		"resource_term": meta.ResourceTerm,
		// 制作�?字幕�?		"releaseGroup": meta.ResourceTeam,
		// 视频编码
		"videoCodec": meta.VideoEncode,
		// 音频编码
		"audioCodec": meta.AudioEncode,
		// 流媒体平�?		"webSource": meta.WebSource,
	}
	
	for k, v := range episodeData {
		tcb.context[k] = v
	}
	for k, v := range metaInfo {
		tcb.context[k] = v
	}
	for k, v := range techMetadata {
		tcb.context[k] = v
	}
}

// addTorrentInfo 添加种子信息
func (tcb *TemplateContextBuilder) addTorrentInfo(torrentinfo *models.TorrentInfo) {
	/*
	 * 添加种子信息
	 */
	if torrentinfo == nil {
		return
	}
	
	var size interface{}
	if torrentinfo.Size != nil {
		if utils.IsNumeric(fmt.Sprintf("%v", torrentinfo.Size)) {
			size = utils.StringUtils.StrFilesize(*torrentinfo.Size)
		} else {
			size = torrentinfo.Size
		}
	} else {
		size = 0
	}
	
	var description *string
	if torrentinfo.Description != nil {
		// 移除HTML标签
		re := regexp.MustCompile(`<[^>]+>`)
		desc := re.ReplaceAllString(*torrentinfo.Description, "")
		desc = html.UnescapeString(desc)
		description = &desc
	}
	
	torrentInfo := map[string]interface{}{
		// 种子标题
		"torrent_title": torrentinfo.Title,
		// 发布时间
		"pubdate": torrentinfo.Pubdate,
		// 免费剩余时间
		"freedate": torrentinfo.FreedateDiff,
		// 做种�?		"seeders": torrentinfo.Seeders,
		// 促销信息
		"volume_factor": torrentinfo.VolumeFactor,
		// Hit&Run
		"hit_and_run": func() string {
			if torrentinfo.HitAndRun {
				return "�?
			}
			return "�?
		}(),
		// 种子标签
		"labels": strings.Join(torrentinfo.Labels, " "),
		// 描述
		"description": description,
		// 站点名称
		"site_name": torrentinfo.SiteName,
		// 种子大小
		"size": size,
	}
	
	for k, v := range torrentInfo {
		tcb.context[k] = v
	}
}

// addTransferInfo 添加文件转移上下�?func (tcb *TemplateContextBuilder) addTransferInfo(transferinfo *models.TransferInfo) {
	/*
	 * 添加文件转移上下�?	 */
	if transferinfo == nil {
		return
	}
	
	ctx := map[string]interface{}{
		"transfer_type": transferinfo.TransferType,
		"file_count":    transferinfo.FileCount,
		"total_size":    utils.StringUtils.StrFilesize(float64(transferinfo.TotalSize)),
		"err_msg":       transferinfo.Message,
	}
	
	for k, v := range ctx {
		tcb.context[k] = v
	}
}

// addFileInfo 添加文件信息
func (tcb *TemplateContextBuilder) addFileInfo(fileExtension *string) {
	/*
	 * 添加文件信息
	 */
	if fileExtension == nil {
		return
	}
	
	fileInfo := map[string]interface{}{
		// 文件后缀
		"fileExt": *fileExtension,
	}
	
	for k, v := range fileInfo {
		tcb.context[k] = v
	}
}

// addRawObjects 添加原始对象引用
func (tcb *TemplateContextBuilder) addRawObjects(
	meta *models.MetaBase,
	mediainfo *models.MediaInfo,
	torrentinfo *models.TorrentInfo,
	transferinfo *models.TransferInfo,
	episodesInfo []models.TmdbEpisode,
) {
	/*
	 * 添加原始对象引用
	 */
	rawObjects := map[string]interface{}{
		// 文件元数�?		"__meta__": meta,
		// 识别的媒体信�?		"__mediainfo__": mediainfo,
		// 种子信息
		"__torrentinfo__": torrentinfo,
		// 文件转移信息
		"__transferinfo__": transferinfo,
		// 当前季的全部集信�?		"__episodes_info__": episodesInfo,
	}
	
	for k, v := range rawObjects {
		tcb.context[k] = v
	}
}

// convertInvalidCharacters 将不支持的字符转换为全角字符
func (tcb *TemplateContextBuilder) convertInvalidCharacters(filename string) string {
	/*
	 * 将不支持的字符转换为全角字符
	 */
	if filename == "" {
		return filename
	}
	
	invalidCharacters := `\/:*?"<>|`
	// 创建半角到全角字符的转换�?	var halfwidthChars, fullwidthChars strings.Builder
	for i := 33; i < 127; i++ {
		halfwidthChars.WriteByte(byte(i))
		fullwidthChars.WriteByte(byte(i + 0xFEE0))
	}
	
	translationTable := make(map[rune]rune)
	halfRunes := []rune(halfwidthChars.String())
	fullRunes := []rune(fullwidthChars.String())
	for i, char := range halfRunes {
		translationTable[char] = fullRunes[i]
	}
	
	// 将不支持的字符替换为对应的全角字�?	var result strings.Builder
	for _, char := range filename {
		if strings.ContainsRune(invalidCharacters, char) {
			if translated, exists := translationTable[char]; exists {
				result.WriteRune(translated)
			} else {
				result.WriteRune(char)
			}
		} else {
			result.WriteRune(char)
		}
	}
	
	return result.String()
}

// TemplateHelper 模板格式渲染帮助�?type TemplateHelper struct {
	builder *TemplateContextBuilder
	cache   *utils.TTLCache
}

var (
	templateHelperInstance *TemplateHelper
	templateHelperOnce     sync.Once
)

// NewTemplateHelper 创建模板帮助类单例实�?func NewTemplateHelper() *TemplateHelper {
	templateHelperOnce.Do(func() {
		templateHelperInstance = &TemplateHelper{
			builder: NewTemplateContextBuilder(),
			cache:   utils.NewTTLCache("notification", 100, 600*time.Second),
		}
	})
	return templateHelperInstance
}

// generateCacheKey 生成缓存�?func (th *TemplateHelper) generateCacheKey(content interface{}) string {
	/*
	 * 生成缓存�?	 */
	switch c := content.(type) {
	case map[string]interface{}:
		var baseStr string
		if title, ok := c["title"].(string); ok {
			baseStr += title
		}
		if text, ok := c["text"].(string); ok {
			baseStr += text
		}
		return utils.StringUtils.MD5Hash([]byte(baseStr))
	case string:
		return utils.StringUtils.MD5Hash([]byte(c))
	default:
		return utils.StringUtils.MD5Hash([]byte(fmt.Sprintf("%v", c)))
	}
}

// getCacheContext 获取缓存上下�?func (th *TemplateHelper) getCacheContext(content interface{}) interface{} {
	/*
	 * 获取缓存上下�?	 */
	cacheKey := th.generateCacheKey(content)
	return th.cache.Get(cacheKey)
}

// setCacheContext 设置缓存上下�?func (th *TemplateHelper) setCacheContext(content interface{}, context interface{}) {
	/*
	 * 设置缓存上下�?	 */
	cacheKey := th.generateCacheKey(content)
	th.cache.Set(cacheKey, context)
}

// Render 根据模板格式渲染内容
func (th *TemplateHelper) Render(
	templateContent string,
	templateType string,
	meta *models.MetaBase,
	mediainfo *models.MediaInfo,
	torrentinfo *models.TorrentInfo,
	transferinfo *models.TransferInfo,
	fileExtension *string,
	episodesInfo []models.TmdbEpisode,
	includeRawObjects bool,
	additional map[string]interface{},
) (interface{}, error) {
	/*
	 * 根据模板格式渲染内容
	 * :param templateContent: 模板字符�?	 * :param templateType: 模板字符串类�?消息通知`literal`, 路径`string`)
	 * :param meta: 媒体信息
	 * :param mediainfo: 媒体信息
	 * :param torrentinfo: 种子信息
	 * :param transferinfo: 传输信息
	 * :param fileExtension: 文件扩展�?	 * :param episodesInfo: 剧集信息
	 * :param includeRawObjects: 是否包含原始对象
	 * :param additional: 额外参数
	 * :return: 渲染后的结果
	 */
	
	// 解析模板字符
	parsed, err := th.parseTemplateContent(templateContent, templateType)
	if err != nil {
		return nil, fmt.Errorf("模板解析失败: %v", err)
	}
	
	if parsed == "" {
		return nil, fmt.Errorf("模板解析失败")
	}
	
	context := th.builder.Build(meta, mediainfo, torrentinfo, transferinfo, fileExtension, episodesInfo, includeRawObjects, additional)
	if len(context) == 0 {
		return nil, fmt.Errorf("上下文构建失�?)
	}
	
	rendered, err := th.renderWithContext(parsed, context)
	if err != nil {
		return nil, fmt.Errorf("模板渲染失败: %v", err)
	}
	
	if rendered == "" {
		return nil, fmt.Errorf("模板渲染失败")
	}
	
	var processed interface{}
	if templateType == "string" {
		processed = rendered
	} else {
		processed, err = th.processFormattedString(rendered)
		if err != nil {
			return nil, fmt.Errorf("处理格式化字符串失败: %v", err)
		}
	}
	
	if processed != nil {
		// 缓存上下�?		th.setCacheContext(processed, context)
		// 返回渲染结果
		return processed, nil
	}
	
	return nil, nil
}

// renderWithContext 使用指定上下文渲染模板字符串
func (th *TemplateHelper) renderWithContext(templateContent string, context map[string]interface{}) (string, error) {
	/*
	 * 使用指定上下文渲染模板字符串
	 * templateContent: 模板字符�?	 * context: 渲染用的上下文数�?	 */
	// 简单的模板渲染实现（Go标准库没有Jinja2�?	// 这里实现一个基础的变量替换功能，支持条件判断和循�?	result := templateContent
	
	// 处理简单的变量替换 {{variable}}
	for key, value := range context {
		placeholder := fmt.Sprintf("{{%s}}", key)
		strValue := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}
	
	// 处理简单的条件判断 {% if variable %}...{% endif %}
	result = th.renderConditions(result, context)
	
	return result, nil
}

// renderConditions 处理模板中的条件判断
func (th *TemplateHelper) renderConditions(templateContent string, context map[string]interface{}) string {
	// 简单的条件判断处理
	re := regexp.MustCompile(`{%\s*if\s+(\w+)\s*%}([^]*?){%\s*endif\s*%}`)
	result := re.ReplaceAllStringFunc(templateContent, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			conditionVar := submatches[1]
			content := submatches[2]
			
			// 检查条件变量是否存在且为真
			if value, exists := context[conditionVar]; exists {
				// 简单判断非空和非false
				if value != nil && value != false && value != "" {
					return content
				}
			}
		}
		return ""
	})
	return result
}

// parseTemplateContent 解析模板字符
func (th *TemplateHelper) parseTemplateContent(templateContent string, templateType string) (string, error) {
	/*
	 * 解析模板字符
	 * :param templateContent 模板格式字符
	 * :param templateType 模板字符类型
	 */
	
	parseLiteral := func(content string) (string, error) {
		/*
		 * 解析Python字面�?		 */
		var templateDict map[string]interface{}
		if err := json.Unmarshal([]byte(content), &templateDict); err != nil {
			return "", fmt.Errorf("无效的Python字面量格�? %v", err)
		}
		jsonBytes, err := json.Marshal(templateDict)
		if err != nil {
			return "", fmt.Errorf("序列化失�? %v", err)
		}
		return string(jsonBytes), nil
	}
	
	if templateType != "" {
		switch templateType {
		case "string":
			return templateContent, nil
		case "dict":
			jsonBytes, err := json.Marshal(templateContent)
			if err != nil {
				return "", fmt.Errorf("序列化失�? %v", err)
			}
			return string(jsonBytes), nil
		case "literal":
			return parseLiteral(templateContent)
		default:
			return "", fmt.Errorf("不支持的模板类型: %s", templateType)
		}
	}
	
	// 自动判断模板类型
	if strings.HasPrefix(templateContent, "{") && strings.HasSuffix(templateContent, "}") {
		// 尝试解析为JSON
		var temp interface{}
		if err := json.Unmarshal([]byte(templateContent), &temp); err == nil {
			return templateContent, nil
		}
	}
	
	// 尝试解析为字面量
	if parsed, err := parseLiteral(templateContent); err == nil {
		return parsed, nil
	}
	
	return templateContent, nil
}

// processFormattedString 处理格式化字符串
func (th *TemplateHelper) processFormattedString(rendered string) (interface{}, error) {
	/*
	 * 处理格式化字符串
	 * 保留转义字符
	 */
	
	restoreChars := func(obj interface{}) interface{} {
		/* 恢复特殊字符 */
		switch v := obj.(type) {
		case string:
			v = strings.ReplaceAll(v, "\\n", "\n")
			v = strings.ReplaceAll(v, "\\r", "\r")
			v = strings.ReplaceAll(v, "\\t", "\t")
			v = strings.ReplaceAll(v, "\\b", "\b")
			v = strings.ReplaceAll(v, "\\f", "\f")
			return v
		case map[string]interface{}:
			result := make(map[string]interface{})
			for k, val := range v {
				result[k] = restoreChars(val)
			}
			return result
		case []interface{}:
			result := make([]interface{}, len(v))
			for i, item := range v {
				result[i] = restoreChars(item)
			}
			return result
		default:
			return obj
		}
	}
	
	// 定义特殊字符映射
	specialChars := map[string]string{
		"\n": "\\n", // 换行�?		"\r": "\\r", // 回车�?		"\t": "\\t", // 制表�?		"\b": "\\b", // 退格符
		"\f": "\\f", // 换页�?	}
	
	// 处理特殊字符
	processed := rendered
	for char, escape := range specialChars {
		processed = strings.ReplaceAll(processed, char, escape)
	}
	
	// 尝试解析为JSON
	var renderedDict map[string]interface{}
	if err := json.Unmarshal([]byte(processed), &renderedDict); err == nil {
		return restoreChars(renderedDict), nil
	}
	
	return rendered, nil
}

// Close 清理资源
func (th *TemplateHelper) Close() {
	/*
	 * 清理资源
	 */
	if th.cache != nil {
		th.cache.Close()
	}
}

// MessageTemplateHelper 消息模板渲染�?type MessageTemplateHelper struct{}

// NewMessageTemplateHelper 创建消息模板渲染器实�?func NewMessageTemplateHelper() *MessageTemplateHelper {
	return &MessageTemplateHelper{}
}

// Render 渲染消息模板
func (mth *MessageTemplateHelper) Render(
	message *models.Notification,
	meta *models.MetaBase,
	mediainfo *models.MediaInfo,
	torrentinfo *models.TorrentInfo,
	transferinfo *models.TransferInfo,
	fileExtension *string,
	episodesInfo []models.TmdbEpisode,
	includeRawObjects bool,
	additional map[string]interface{},
) (*models.Notification, error) {
	/*
	 * 渲染消息模板
	 */
	if !mth.isInstanceValid(message) {
		if mth.meetsUpdateConditions(message, additional) {
			logger.GetLoggerManager().Info("将使用模板渲染消息内�?)
			return mth.applyTemplateData(message, meta, mediainfo, torrentinfo, transferinfo, fileExtension, episodesInfo, includeRawObjects, additional)
		}
	}
	return message, nil
}

// isInstanceValid 检查消息是否有�?func (mth *MessageTemplateHelper) isInstanceValid(message *models.Notification) bool {
	/*
	 * 检查消息是否有�?	 */
	if message != nil {
		return (message.Title != nil && *message.Title != "") || (message.Text != nil && *message.Text != "")
	}
	return false
}

// meetsUpdateConditions 判断是否满足消息实例更新条件
func (mth *MessageTemplateHelper) meetsUpdateConditions(message *models.Notification, additional map[string]interface{}) bool {
	/*
	 * 判断是否满足消息实例更新条件
	 *
	 * 满足条件需同时具备�?	 * 1. 消息为有效Notification实例
	 * 2. 消息指定了模板类�?ctype)
	 * 3. 存在待渲染的模板变量数据
	 */
	if message != nil {
		return message.CType != nil && (additional != nil && len(additional) > 0)
	}
	return false
}

// applyTemplateData 更新消息实例
func (mth *MessageTemplateHelper) applyTemplateData(
	message *models.Notification,
	meta *models.MetaBase,
	mediainfo *models.MediaInfo,
	torrentinfo *models.TorrentInfo,
	transferinfo *models.TransferInfo,
	fileExtension *string,
	episodesInfo []models.TmdbEpisode,
	includeRawObjects bool,
	additional map[string]interface{},
) (*models.Notification, error) {
	/*
	 * 更新消息实例
	 */
	defer func() {
		if r := recover(); r != nil {
			logger.GetLoggerManager().Errorf("更新Notification时出现错误：%v", r)
		}
	}()
	
	if template := mth.getTemplate(message); template != "" {
		rendered, err := NewTemplateHelper().Render(template, "literal", meta, mediainfo, torrentinfo, transferinfo, fileExtension, episodesInfo, includeRawObjects, additional)
		if err != nil {
			return message, err
		}
		
		if renderedMap, ok := rendered.(map[string]interface{}); ok {
			// 使用反射更新message字段
			message.UpdateFromMap(renderedMap)
		}
	}
	return message, nil
}

// getTemplate 获取消息模板
func (mth *MessageTemplateHelper) getTemplate(message *models.Notification) string {
	/*
	 * 获取消息模板
	 */
	configOper := db.NewSystemConfigOper()
	templateDict := configOper.Get(models.SystemConfigKeyNotificationTemplates)
	
	if templateMap, ok := templateDict.(map[string]interface{}); ok {
		if ctype := message.CType; ctype != nil {
			if template, exists := templateMap[*ctype]; exists {
				if templateStr, ok := template.(string); ok {
					return templateStr
				}
			}
		}
	}
	
	return ""
}

// MessageQueueManager 消息发送队列管理器
type MessageQueueManager struct {
	schedulePeriods []SchedulePeriod
	queue           chan MessageQueueItem
	sendCallback    func(args []interface{}, kwargs map[string]interface{})
	checkInterval   time.Duration
	running         bool
	mutex           sync.Mutex
}

// SchedulePeriod 时间段结�?type SchedulePeriod struct {
	StartH int
	StartM int
	EndH   int
	EndM   int
}

// MessageQueueItem 消息队列�?type MessageQueueItem struct {
	Args   []interface{}
	Kwargs map[string]interface{}
}

var (
	messageQueueManagerInstance *MessageQueueManager
	messageQueueManagerOnce     sync.Once
)

// NewMessageQueueManager 创建消息队列管理器单例实�?func NewMessageQueueManager(sendCallback func(args []interface{}, kwargs map[string]interface{})) *MessageQueueManager {
	messageQueueManagerOnce.Do(func() {
		messageQueueManagerInstance = &MessageQueueManager{
			schedulePeriods: make([]SchedulePeriod, 0),
			queue:           make(chan MessageQueueItem, 100), // 缓冲通道
			sendCallback:    sendCallback,
			checkInterval:   10 * time.Second,
			running:         true,
		}
		messageQueueManagerInstance.initConfig()
		go messageQueueManagerInstance.monitorLoop()
	})
	return messageQueueManagerInstance
}

// initConfig 初始化配�?func (mqm *MessageQueueManager) initConfig() {
	/*
	 * 初始化配�?	 */
	configOper := db.NewSystemConfigOper()
	periods := configOper.Get(models.SystemConfigKeyNotificationSendTime)
	mqm.schedulePeriods = mqm.parseSchedule(periods)
}

// parseSchedule 将字符串时间格式转换为分钟数元组
func (mqm *MessageQueueManager) parseSchedule(periods interface{}) []SchedulePeriod {
	/*
	 * 将字符串时间格式转换为分钟数元组
	 * 支持格式�?'HH:MM' �?'HH:MM:SS' 的时间字符串
	 */
	parsed := make([]SchedulePeriod, 0)
	
	if periods == nil {
		return parsed
	}
	
	// 尝试将periods转换为切�?	var periodsSlice []interface{}
	switch p := periods.(type) {
	case []interface{}:
		periodsSlice = p
	case []map[string]interface{}:
		// 转换为[]interface{}
		periodsSlice = make([]interface{}, len(p))
		for i, v := range p {
			periodsSlice[i] = v
		}
	default:
		return parsed
	}
	
	for _, period := range periodsSlice {
		if periodMap, ok := period.(map[string]interface{}); ok {
			if start, startExists := periodMap["start"].(string); startExists {
				if end, endExists := periodMap["end"].(string); endExists {
					startParts := strings.Split(start, ":")
					endParts := strings.Split(end, ":")
					
					if len(startParts) >= 2 && len(endParts) >= 2 {
						startH, err1 := strconv.Atoi(startParts[0])
						startM, err2 := strconv.Atoi(startParts[1])
						endH, err3 := strconv.Atoi(endParts[0])
						endM, err4 := strconv.Atoi(endParts[1])
						
						if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
							parsed = append(parsed, SchedulePeriod{
								StartH: startH,
								StartM: startM,
								EndH:   endH,
								EndM:   endM,
							})
						}
					}
				}
			}
		}
	}
	
	return parsed
}

// isInScheduledTime 检查当前时间是否在允许发送的时间段内
func (mqm *MessageQueueManager) isInScheduledTime(currentTime time.Time) bool {
	/*
	 * 检查当前时间是否在允许发送的时间段内
	 */
	if len(mqm.schedulePeriods) == 0 {
		return true
	}
	
	currentMinutes := currentTime.Hour()*60 + currentTime.Minute()
	for _, period := range mqm.schedulePeriods {
		start := period.StartH*60 + period.StartM
		end := period.EndH*60 + period.EndM
		
		if start <= end {
			if start <= currentMinutes && currentMinutes <= end {
				return true
			}
		} else {
			if currentMinutes >= start || currentMinutes <= end {
				return true
			}
		}
	}
	return false
}

// SendMessage 发送消息（立即发送或加入队列�?func (mqm *MessageQueueManager) SendMessage(args []interface{}, kwargs map[string]interface{}) {
	/*
	 * 发送消息（立即发送或加入队列�?	 */
	immediately := false
	if imm, exists := kwargs["immediately"]; exists {
		if immBool, ok := imm.(bool); ok {
			immediately = immBool
		}
		delete(kwargs, "immediately")
	}
	
	if immediately || mqm.isInScheduledTime(time.Now()) {
		mqm.send(args, kwargs)
	} else {
		select {
		case mqm.queue <- MessageQueueItem{Args: args, Kwargs: kwargs}:
			logger.GetLoggerManager().Infof("消息已加入队列，当前队列长度�?d", len(mqm.queue))
		default:
			logger.GetLoggerManager().Warn("消息队列已满，丢弃消�?)
		}
	}
}

// AsyncSendMessage 异步发送消息（直接加入队列�?func (mqm *MessageQueueManager) AsyncSendMessage(args []interface{}, kwargs map[string]interface{}) {
	/*
	 * 异步发送消息（直接加入队列�?	 */
	delete(kwargs, "immediately")
	
	select {
	case mqm.queue <- MessageQueueItem{Args: args, Kwargs: kwargs}:
		logger.GetLoggerManager().Infof("消息已加入队列，当前队列长度�?d", len(mqm.queue))
	default:
		logger.GetLoggerManager().Warn("消息队列已满，丢弃消�?)
	}
}

// send 实际发送消息（可通过回调函数自定义）
func (mqm *MessageQueueManager) send(args []interface{}, kwargs map[string]interface{}) {
	/*
	 * 实际发送消息（可通过回调函数自定义）
	 */
	if mqm.sendCallback != nil {
		defer func() {
			if r := recover(); r != nil {
				logger.GetLoggerManager().Errorf("发送消息错误：%v", r)
			}
		}()
		
		logger.GetLoggerManager().Infof("发送消息：%v", kwargs)
		mqm.sendCallback(args, kwargs)
	}
}

// monitorLoop 后台线程循环检查时间并处理队列
func (mqm *MessageQueueManager) monitorLoop() {
	/*
	 * 后台线程循环检查时间并处理队列
	 */
	for mqm.running {
		currentTime := time.Now()
		if mqm.isInScheduledTime(currentTime) {
			for {
				select {
				case item := <-mqm.queue:
					// 检查系统是否停�?					if core.GlobalVars.IsSystemStopped() {
						return
					}
					
					// 再次检查是否在允许发送的时间段内
					if !mqm.isInScheduledTime(time.Now()) {
						// 如果不在时间段内，将消息重新放回队列
						select {
						case mqm.queue <- item:
						default:
							logger.GetLoggerManager().Warn("消息队列已满，丢弃消�?)
						}
						break
					}
					
					mqm.send(item.Args, item.Kwargs)
					logger.GetLoggerManager().Infof("队列剩余消息�?d", len(mqm.queue))
				default:
					// 队列为空，跳出循�?					break
				}
				break
			}
		}
		time.Sleep(mqm.checkInterval)
	}
}

// Stop 停止队列管理�?func (mqm *MessageQueueManager) Stop() {
	/*
	 * 停止队列管理�?	 */
	mqm.mutex.Lock()
	defer mqm.mutex.Unlock()
	
	mqm.running = false
	logger.GetLoggerManager().Info("正在停止消息队列...")
	close(mqm.queue)
	logger.GetLoggerManager().Info("消息队列已停�?)
}

// MessageHelper 消息队列管理器，包括系统消息和用户消�?type MessageHelper struct {
	sysQueue  chan string
	userQueue chan string
}

var (
	messageHelperInstance *MessageHelper
	messageHelperOnce     sync.Once
)

// NewMessageHelper 创建消息帮助类单例实�?func NewMessageHelper() *MessageHelper {
	messageHelperOnce.Do(func() {
		messageHelperInstance = &MessageHelper{
			sysQueue:  make(chan string, 100),
			userQueue: make(chan string, 100),
		}
	})
	return messageHelperInstance
}

// Put 存消�?func (mh *MessageHelper) Put(message interface{}, role string, title string, note interface{}) {
	/*
	 * 存消�?	 * :param message: 消息
	 * :param role: 消息通道 system：系统消息，plugin：插件消息，user：用户消�?	 * :param title: 标题
	 * :param note: 附件json
	 */
	msgData := map[string]interface{}{
		"date": time.Now().Format("2006-01-02 15:04:05"),
		"note": note,
	}
	
	if role == "system" || role == "plugin" {
		// 没有标题时获取插件名�?		if role == "plugin" && title == "" {
			title = "插件通知"
		}
		// 系统通知，默�?		msgData["type"] = role
		msgData["title"] = title
		
		switch m := message.(type) {
		case string:
			msgData["text"] = m
		default:
			msgData["text"] = fmt.Sprintf("%v", m)
		}
		
		jsonData, _ := json.Marshal(msgData)
		select {
		case mh.sysQueue <- string(jsonData):
		default:
			logger.GetLoggerManager().Warn("系统消息队列已满，丢弃消�?)
		}
	} else {
		if msgStr, ok := message.(string); ok {
			// 非系统的文本通知
			msgData["title"] = title
			msgData["text"] = msgStr
			
			jsonData, _ := json.Marshal(msgData)
			select {
			case mh.userQueue <- string(jsonData):
			default:
				logger.GetLoggerManager().Warn("用户消息队列已满，丢弃消�?)
			}
		} else {
			// 非系统的复杂结构通知，如媒体信息/种子列表等�?			if toDict, ok := message.(ToDict); ok {
				content := toDict.ToDict()
				content["title"] = title
				content["date"] = time.Now().Format("2006-01-02 15:04:05")
				content["note"] = note
				
				jsonData, _ := json.Marshal(content)
				select {
				case mh.userQueue <- string(jsonData):
				default:
					logger.GetLoggerManager().Warn("用户消息队列已满，丢弃消�?)
				}
			} else {
				// 尝试转换为map
				if msgMap, ok := message.(map[string]interface{}); ok {
					msgMap["title"] = title
					msgMap["date"] = time.Now().Format("2006-01-02 15:04:05")
					msgMap["note"] = note
					
					jsonData, _ := json.Marshal(msgMap)
					select {
					case mh.userQueue <- string(jsonData):
					default:
						logger.GetLoggerManager().Warn("用户消息队列已满，丢弃消�?)
					}
				}
			}
		}
	}
}

// ToDict 可转换为字典的接�?type ToDict interface {
	ToDict() map[string]interface{}
}

// Get 取消�?func (mh *MessageHelper) Get(role string) *string {
	/*
	 * 取消�?	 * :param role: 消息通道 system：系统消息，plugin：插件消息，user：用户消�?	 */
	select {
	case msg := <-mh.sysQueue:
		if role == "system" {
			return &msg
		}
	default:
	}
	
	select {
	case msg := <-mh.userQueue:
		if role != "system" {
			return &msg
		}
	default:
	}
	
	return nil
}

// StopMessage 停止消息服务
func StopMessage() {
	/*
	 * 停止消息服务
	 */
	// 停止消息队列
	if messageQueueManagerInstance != nil {
		messageQueueManagerInstance.Stop()
	}
	
	// 关闭消息渲染�?	NewTemplateHelper().Close()
}
