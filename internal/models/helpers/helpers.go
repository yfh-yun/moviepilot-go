package helpers

import (
	"encoding/json"
	"fmt"

	. "moviepilot-go/internal/models/dto"
	. "moviepilot-go/internal/models/types"
)

// IntPtr 创建int指针
func IntPtr(i int) *int {
	return &i
}

// StringPtr 创建string指针
func StringPtr(s string) *string {
	return &s
}

// BoolPtr 创建bool指针
func BoolPtr(b bool) *bool {
	return &b
}

// Float64Ptr 创建float64指针
func Float64Ptr(f float64) *float64 {
	return &f
}

// IntValue 安全获取int指针的值，如果为nil返回默认值
func IntValue(p *int, defaultValue int) int {
	if p == nil {
		return defaultValue
	}
	return *p
}

// StringValue 安全获取string指针的值，如果为nil返回默认值
func StringValue(p *string, defaultValue string) string {
	if p == nil {
		return defaultValue
	}
	return *p
}

// BoolValue 安全获取bool指针的值，如果为nil返回默认值
func BoolValue(p *bool, defaultValue bool) bool {
	if p == nil {
		return defaultValue
	}
	return *p
}

// Float64Value 安全获取float64指针的值，如果为nil返回默认值
func Float64Value(p *float64, defaultValue float64) float64 {
	if p == nil {
		return defaultValue
	}
	return *p
}

// ToJSON 将结构体转换为JSON字符串
func ToJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONIndent 将结构体转换为格式化的JSON字符串
func ToJSONIndent(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析到结构体
func FromJSON(jsonStr string, v any) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// ToMap 将结构体转换为map[string]interface{}
func ToMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FromMap 从map[string]interface{}转换到结构体
func FromMap(m map[string]any, v any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// IsValidMediaType 验证媒体类型是否有效
func IsValidMediaType(mediaType string) bool {
	validTypes := map[string]bool{
		string(MediaTypeMovie):      true,
		string(MediaTypeTV):         true,
		string(MediaTypeCollection): true,
		string(MediaTypeUnknown):    true,
	}
	return validTypes[mediaType]
}

// IsValidEventType 验证事件类型是否有效
func IsValidEventType(eventType string) bool {
	_, exists := EventTypeNames[EventType(eventType)]
	return exists
}

// GetEventTypeName 获取事件类型的中文名称
func GetEventTypeName(eventType EventType) string {
	if name, exists := EventTypeNames[eventType]; exists {
		return name
	}
	return "未知事件"
}

// IsValidMessageChannel 验证消息渠道是否有效
func IsValidMessageChannel(channel string) bool {
	validChannels := map[string]bool{
		string(MessageChannelWechat):       true,
		string(MessageChannelTelegram):     true,
		string(MessageChannelSlack):        true,
		string(MessageChannelSynologyChat): true,
		string(MessageChannelVoceChat):     true,
		string(MessageChannelWeb):          true,
		string(MessageChannelWebPush):      true,
	}
	return validChannels[channel]
}

// IsValidDownloaderType 验证下载器类型是否有效
func IsValidDownloaderType(downloaderType string) bool {
	validTypes := map[string]bool{
		string(DownloaderTypeQbittorrent):  true,
		string(DownloaderTypeTransmission): true,
	}
	return validTypes[downloaderType]
}

// IsValidMediaServerType 验证媒体服务器类型是否有效
func IsValidMediaServerType(serverType string) bool {
	validTypes := map[string]bool{
		string(MediaServerTypeEmby):       true,
		string(MediaServerTypeJellyfin):   true,
		string(MediaServerTypePlex):       true,
		string(MediaServerTypeTrimeMedia): true,
	}
	return validTypes[serverType]
}

// IsValidStorageSchema 验证存储类型是否有效
func IsValidStorageSchema(schema string) bool {
	validSchemas := map[string]bool{
		string(StorageSchemaLocal):  true,
		string(StorageSchemaAlipan): true,
		string(StorageSchemaU115):   true,
		string(StorageSchemaRclone): true,
		string(StorageSchemaAlist):  true,
		string(StorageSchemaSMB):    true,
	}
	return validSchemas[schema]
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// FormatRatio 格式化分享率
func FormatRatio(upload, download int64) float64 {
	if download == 0 {
		if upload > 0 {
			return 999.99 // 无限大
		}
		return 0.0
	}
	ratio := float64(upload) / float64(download)
	return ratio
}

// MergeEpisodeList 合并集数列表，去重并排序
func MergeEpisodeList(lists ...[]int) []int {
	episodeMap := make(map[int]bool)
	for _, list := range lists {
		for _, ep := range list {
			episodeMap[ep] = true
		}
	}

	var result []int
	for ep := range episodeMap {
		result = append(result, ep)
	}

	// 简单排序
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// BuildSeasonEpisode 构建SxxExx格式字符串
func BuildSeasonEpisode(season, episode int) string {
	return fmt.Sprintf("S%02dE%02d", season, episode)
}

// BuildSeasonRange 构建季范围字符串
func BuildSeasonRange(start, end int) string {
	if start == end {
		return fmt.Sprintf("S%02d", start)
	}
	return fmt.Sprintf("S%02d-S%02d", start, end)
}

// BuildEpisodeRange 构建集范围字符串
func BuildEpisodeRange(start, end int) string {
	if start == end {
		return fmt.Sprintf("E%02d", start)
	}
	return fmt.Sprintf("E%02d-E%02d", start, end)
}

// CloneMediaInfo 深拷贝MediaInfo
func CloneMediaInfo(src *MediaInfo) (*MediaInfo, error) {
	if src == nil {
		return nil, nil
	}

	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	var dst MediaInfo
	err = json.Unmarshal(data, &dst)
	if err != nil {
		return nil, err
	}

	return &dst, nil
}

// CloneMetaInfo 深拷贝MetaInfo
func CloneMetaInfo(src *MetaInfo) (*MetaInfo, error) {
	if src == nil {
		return nil, nil
	}

	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	var dst MetaInfo
	err = json.Unmarshal(data, &dst)
	if err != nil {
		return nil, err
	}

	return &dst, nil
}

// ValidateSubscribe 验证订阅信息的必填字段
func ValidateSubscribe(sub *Subscribe) error {
	if sub == nil {
		return fmt.Errorf("subscribe is nil")
	}

	if sub.Name == "" {
		return fmt.Errorf("subscribe name is required")
	}

	if sub.Type == "" {
		return fmt.Errorf("subscribe type is required")
	}

	if !IsValidMediaType(sub.Type) {
		return fmt.Errorf("invalid subscribe type: %s", sub.Type)
	}

	return nil
}

// ValidateMediaInfo 验证媒体信息的必填字段
func ValidateMediaInfo(media *MediaInfo) error {
	if media == nil {
		return fmt.Errorf("media info is nil")
	}

	if media.Title == "" {
		return fmt.Errorf("media title is required")
	}

	if media.Type == "" {
		return fmt.Errorf("media type is required")
	}

	if !IsValidMediaType(media.Type) {
		return fmt.Errorf("invalid media type: %s", media.Type)
	}

	return nil
}

// ValidateSite 验证站点信息的必填字段
func ValidateSite(site *Site) error {
	if site == nil {
		return fmt.Errorf("site is nil")
	}

	if site.Name == "" {
		return fmt.Errorf("site name is required")
	}

	if site.Domain == "" {
		return fmt.Errorf("site domain is required")
	}

	if site.URL == "" {
		return fmt.Errorf("site url is required")
	}

	return nil
}
