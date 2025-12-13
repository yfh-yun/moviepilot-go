package media

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MetaBase 统一的元信息基础结构与公共逻辑
type MetaBase struct {
	// 是否处理的文件
	Isfile bool `json:"isfile"` // 是否处理的文件

	// 名称相关
	Title     string `json:"title"`      // 标题
	OrgString string `json:"org_string"` // 原始字符串
	Subtitle  string `json:"subtitle"`   // 副标题
	CnName    string `json:"cn_name"`    // 中文名
	EnName    string `json:"en_name"`    // 英文名
	Year      string `json:"year"`       // 年份

	// 媒体类型
	Type MediaType `json:"type"` // 媒体类型

	// 季相关
	BeginSeason *int `json:"begin_season"` // 开始季
	EndSeason   *int `json:"end_season"`   // 结束季
	TotalSeason int  `json:"total_season"` // 总季数

	// 集相关
	BeginEpisode *int `json:"begin_episode"` // 开始集
	EndEpisode   *int `json:"end_episode"`   // 结束集
	TotalEpisode int  `json:"total_episode"` // 总集数

	// 资源属性
	ResourceType   ResourceType   `json:"resource_type"`   // 资源类型
	ResourceEffect ResourceEffect `json:"resource_effect"` // 资源效果
	ResourcePix    ResourcePix    `json:"resource_pix"`    // 资源分辨率
	ResourceTeam   string         `json:"resource_team"`   // 资源团队
	Customization  string         `json:"customization"`   // 自定义标识
	WebSource      string         `json:"web_source"`      // 网络来源
	VideoEncode    string         `json:"video_encode"`    // 视频编码
	AudioEncode    string         `json:"audio_encode"`    // 音频编码

	// 其他
	Part         string   `json:"part"`          // Partx Cd Dvd Disk Disc
	TMDBID       *int64   `json:"tmdbid"`        // TMDB ID
	DoubanID     string   `json:"doubanid"`      // 豆瓣 ID
	AppliedWords []string `json:"applied_words"` // 应用的识别词
}

// NewMetaBase 创建新的MetaBase实例
func NewMetaBase(title, orgString, subtitle string, isfile bool) *MetaBase {
	return &MetaBase{
		Isfile:    isfile,
		Title:     title,
		OrgString: orgString,
		Subtitle:  subtitle,
		Type:      MediaTypeUnknown,
	}
}

// Name 获取名称，优先返回中文名，其次英文名
func (m *MetaBase) Name() string {
	if m.CnName != "" {
		return m.CnName
	}
	if m.EnName != "" {
		return m.EnName
	}
	return m.CnName
}

// SetName 设置名称，根据字符串是否全为中文来决定是中文名还是英文名
func (m *MetaBase) SetName(name string) {
	// 判断是否全为中文
	if isAllChinese(name) {
		m.CnName = name
		m.EnName = ""
	} else {
		m.EnName = name
		m.CnName = ""
	}
}

// isAllChinese 判断字符串是否全为中文
func isAllChinese(s string) bool {
	for _, r := range s {
		if r < 0x4e00 || r > 0x9fff {
			return false
		}
	}
	return true
}

// Merge 合并元信息
func (m *MetaBase) Merge(other *MetaBase) {
	if other == nil {
		return
	}

	// 合并类型
	if m.Type == MediaTypeUnknown && other.Type != MediaTypeUnknown {
		m.Type = other.Type
	}

	// 合并名称相关
	if m.CnName == "" {
		m.CnName = other.CnName
	}
	if m.EnName == "" {
		m.EnName = other.EnName
	}
	if m.Year == "" {
		m.Year = other.Year
	}

	// 合并季相关
	if m.Type == MediaTypeTV {
		if m.BeginSeason == nil && other.BeginSeason != nil {
			m.BeginSeason = other.BeginSeason
			m.EndSeason = other.EndSeason
			m.TotalSeason = other.TotalSeason
		}
	}

	// 合并集相关
	if m.Type == MediaTypeTV {
		if m.BeginEpisode == nil && other.BeginEpisode != nil {
			m.BeginEpisode = other.BeginEpisode
			m.EndEpisode = other.EndEpisode
			m.TotalEpisode = other.TotalEpisode
		}
	}

	// 合并资源属性
	if m.ResourceType == "" {
		m.ResourceType = other.ResourceType
	}
	if m.ResourceEffect == "" {
		m.ResourceEffect = other.ResourceEffect
	}
	if m.ResourcePix == "" {
		m.ResourcePix = other.ResourcePix
	}
	if m.ResourceTeam == "" {
		m.ResourceTeam = other.ResourceTeam
	}
	if m.Customization == "" {
		m.Customization = other.Customization
	}
	if m.WebSource == "" {
		m.WebSource = other.WebSource
	}
	if m.VideoEncode == "" {
		m.VideoEncode = other.VideoEncode
	}
	if m.AudioEncode == "" {
		m.AudioEncode = other.AudioEncode
	}

	// 合并其他
	if m.Part == "" {
		m.Part = other.Part
	}
	if m.TMDBID == nil && other.TMDBID != nil {
		m.TMDBID = other.TMDBID
	}
	if m.DoubanID == "" {
		m.DoubanID = other.DoubanID
	}
	if len(m.AppliedWords) == 0 {
		m.AppliedWords = other.AppliedWords
	}
}

// Season 获取季数（单季时返回具体季数，多季时返回范围，带S前缀）
func (m *MetaBase) Season() string {
	if m.BeginSeason == nil {
		if m.Type == MediaTypeTV {
			return "S01"
		}
		return ""
	}
	if m.EndSeason == nil {
		return fmt.Sprintf("S%02d", *m.BeginSeason)
	}
	return fmt.Sprintf("S%02d-S%02d", *m.BeginSeason, *m.EndSeason)
}

// Sea 获取开始季字符串，确定是剧集没有季的返回空
func (m *MetaBase) Sea() string {
	if m.BeginSeason == nil {
		return ""
	}
	return m.Season()
}

// SeasonSeq 获取begin_season的数字，电视剧没有季的返回1
func (m *MetaBase) SeasonSeq() string {
	if m.BeginSeason == nil {
		if m.Type == MediaTypeTV {
			return "1"
		}
		return ""
	}
	return fmt.Sprintf("%d", *m.BeginSeason)
}

// SeasonList 获取季列表
func (m *MetaBase) SeasonList() []int {
	var seasons []int
	if m.BeginSeason == nil {
		if m.Type == MediaTypeTV {
			return []int{1}
		}
		return seasons
	}
	if m.EndSeason == nil {
		return []int{*m.BeginSeason}
	}
	for i := *m.BeginSeason; i <= *m.EndSeason; i++ {
		seasons = append(seasons, i)
	}
	return seasons
}

// Episode 获取集数（单集时返回具体集数，多集时返回范围，带E前缀）
func (m *MetaBase) Episode() string {
	if m.BeginEpisode == nil {
		return ""
	}
	if m.EndEpisode == nil {
		return fmt.Sprintf("E%02d", *m.BeginEpisode)
	}
	return fmt.Sprintf("E%02d-E%02d", *m.BeginEpisode, *m.EndEpisode)
}

// EpisodeList 获取集列表
func (m *MetaBase) EpisodeList() []int {
	if m.BeginEpisode == nil {
		return []int{}
	}
	if m.EndEpisode == nil {
		return []int{*m.BeginEpisode}
	}
	episodes := make([]int, 0, *m.EndEpisode-*m.BeginEpisode+1)
	for i := *m.BeginEpisode; i <= *m.EndEpisode; i++ {
		episodes = append(episodes, i)
	}
	return episodes
}

// Episodes 获取单文件多集的集数表达方式，用于支持单文件多集
func (m *MetaBase) Episodes() string {
	episodes := m.EpisodeList()
	if len(episodes) == 0 {
		return ""
	}
	var episodeStrs []string
	for _, ep := range episodes {
		episodeStrs = append(episodeStrs, fmt.Sprintf("E%02d", ep))
	}
	return strings.Join(episodeStrs, "")
}

// EpisodeSeqs 获取单文件多集的集数表达方式，用于支持单文件多集
func (m *MetaBase) EpisodeSeqs() string {
	episodes := m.EpisodeList()
	if len(episodes) == 0 {
		return ""
	}
	if len(episodes) == 1 {
		return fmt.Sprintf("%d", episodes[0])
	}
	return fmt.Sprintf("%d-%d", episodes[0], episodes[len(episodes)-1])
}

// EpisodeSeq 获取begin_episode的数字
func (m *MetaBase) EpisodeSeq() string {
	episodes := m.EpisodeList()
	if len(episodes) == 0 {
		return ""
	}
	return fmt.Sprintf("%d", episodes[0])
}

// SeasonEpisode 获取季集字符串
func (m *MetaBase) SeasonEpisode() string {
	if m.Type != MediaTypeTV {
		return ""
	}
	sea := m.Sea()
	episode := m.Episode()
	if sea != "" && episode != "" {
		return fmt.Sprintf("%s %s", sea, episode)
	}
	if sea != "" {
		return sea
	}
	if episode != "" {
		return episode
	}
	return ""
}

// ResourceTerm 获取资源类型字符串，含分辨率
func (m *MetaBase) ResourceTerm() string {
	ret := ""
	if m.ResourceType != "" {
		ret += fmt.Sprintf(" %s", m.ResourceType)
	}
	if m.ResourceEffect != "" {
		ret += fmt.Sprintf(" %s", m.ResourceEffect)
	}
	if m.ResourcePix != "" {
		ret += fmt.Sprintf(" %s", m.ResourcePix)
	}
	return strings.TrimSpace(ret)
}

// Edition 获取资源类型字符串，不含分辨率
func (m *MetaBase) Edition() string {
	ret := ""
	if m.ResourceType != "" {
		ret += fmt.Sprintf(" %s", m.ResourceType)
	}
	if m.ResourceEffect != "" {
		ret += fmt.Sprintf(" %s", m.ResourceEffect)
	}
	return strings.TrimSpace(ret)
}

// ReleaseGroup 获取发布组/字幕组字符串
func (m *MetaBase) ReleaseGroup() string {
	return m.ResourceTeam
}

// VideoTerm 获取视频编码
func (m *MetaBase) VideoTerm() string {
	return m.VideoEncode
}

// AudioTerm 获取音频编码
func (m *MetaBase) AudioTerm() string {
	return m.AudioEncode
}

// IsInSeason 判断是否在指定季范围内，支持列表类型
func (m *MetaBase) IsInSeason(season interface{}) bool {
	if m.BeginSeason == nil {
		// 没有季信息，电视剧默认返回1
		if m.Type == MediaTypeTV {
			return isInValue(1, season)
		}
		return false
	}

	beginSeason := *m.BeginSeason
	endSeason := beginSeason
	if m.EndSeason != nil {
		endSeason = *m.EndSeason
	}

	return isInRange(beginSeason, endSeason, season)
}

// IsInEpisode 判断是否在指定集范围内，支持列表类型
func (m *MetaBase) IsInEpisode(episode interface{}) bool {
	if m.BeginEpisode == nil {
		return false
	}

	beginEpisode := *m.BeginEpisode
	endEpisode := beginEpisode
	if m.EndEpisode != nil {
		endEpisode = *m.EndEpisode
	}

	return isInRange(beginEpisode, endEpisode, episode)
}

// isInRange 判断值是否在范围内，支持列表类型
func isInRange(begin, end int, value interface{}) bool {
	switch v := value.(type) {
	case []int:
		// 列表类型，检查所有值是否都在范围内
		for _, item := range v {
			if item < begin || item > end {
				return false
			}
		}
		return true
	case int:
		// 单个值，检查是否在范围内
		return v >= begin && v <= end
	case string:
		// 字符串类型，转换为整数后检查
		intValue, err := parseInt(v)
		if err != nil {
			return false
		}
		return intValue >= begin && intValue <= end
	default:
		return false
	}
}

// isInValue 判断值是否等于指定值，支持列表类型
func isInValue(value int, check interface{}) bool {
	switch v := check.(type) {
	case []int:
		// 列表类型，检查是否包含指定值
		for _, item := range v {
			if item == value {
				return true
			}
		}
		return false
	case int:
		// 单个值，直接比较
		return v == value
	case string:
		// 字符串类型，转换为整数后比较
		intValue, err := parseInt(v)
		if err != nil {
			return false
		}
		return intValue == value
	default:
		return false
	}
}

// parseInt 将字符串转换为整数
func parseInt(s string) (int, error) {
	// 这里可以添加中文数字转换逻辑，暂时只处理阿拉伯数字
	return strconv.Atoi(s)
}

// SetSeason 设置季信息，支持列表/字符串类型
func (m *MetaBase) SetSeason(sea interface{}) {
	var beginSeason, endSeason *int

	switch v := sea.(type) {
	case []int:
		if len(v) == 1 {
			begin := v[0]
			beginSeason = &begin
			endSeason = nil
		} else if len(v) > 1 {
			begin := v[0]
			beginSeason = &begin
			end := v[len(v)-1]
			endSeason = &end
		}
	case int:
		begin := v
		beginSeason = &begin
		endSeason = nil
	case string:
		// 处理字符串类型，例如 "1", "1-2"
		if strings.Contains(v, "-") {
			parts := strings.Split(v, "-")
			if len(parts) == 2 {
				begin, err1 := parseInt(strings.TrimSpace(parts[0]))
				end, err2 := parseInt(strings.TrimSpace(parts[1]))
				if err1 == nil && err2 == nil {
					beginSeason = &begin
					endSeason = &end
				}
			}
		} else {
			season, err := parseInt(strings.TrimSpace(v))
			if err == nil {
				beginSeason = &season
				endSeason = nil
			}
		}
	}

	// 更新季信息
	m.BeginSeason = beginSeason
	m.EndSeason = endSeason

	// 更新总季数
	if m.BeginSeason != nil {
		if m.EndSeason != nil {
			m.TotalSeason = *m.EndSeason - *m.BeginSeason + 1
		} else {
			m.TotalSeason = 1
		}
	}
}

// SetEpisode 设置集信息，支持列表/字符串类型
func (m *MetaBase) SetEpisode(ep interface{}) {
	var beginEpisode, endEpisode *int

	switch v := ep.(type) {
	case []int:
		if len(v) == 1 {
			begin := v[0]
			beginEpisode = &begin
			endEpisode = nil
		} else if len(v) > 1 {
			begin := v[0]
			beginEpisode = &begin
			end := v[len(v)-1]
			endEpisode = &end
		}
	case int:
		begin := v
		beginEpisode = &begin
		endEpisode = nil
	case string:
		// 处理字符串类型，例如 "1", "1-2"
		if strings.Contains(v, "-") {
			parts := strings.Split(v, "-")
			if len(parts) == 2 {
				begin, err1 := parseInt(strings.TrimSpace(parts[0]))
				end, err2 := parseInt(strings.TrimSpace(parts[1]))
				if err1 == nil && err2 == nil {
					beginEpisode = &begin
					endEpisode = &end
				}
			}
		} else {
			episode, err := parseInt(strings.TrimSpace(v))
			if err == nil {
				beginEpisode = &episode
				endEpisode = nil
			}
		}
	}

	// 更新集信息
	m.BeginEpisode = beginEpisode
	m.EndEpisode = endEpisode

	// 更新总集数
	if m.BeginEpisode != nil {
		if m.EndEpisode != nil {
			m.TotalEpisode = *m.EndEpisode - *m.BeginEpisode + 1
		} else {
			m.TotalEpisode = 1
		}
	}
}

// SetEpisodes 设置开始集结束集
func (m *MetaBase) SetEpisodes(begin, end *int) {
	m.BeginEpisode = begin
	m.EndEpisode = end
	if begin != nil && end != nil {
		m.TotalEpisode = *end - *begin + 1
	}
}

// InitSubtitle 初始化副标题，解析季集信息
func (m *MetaBase) InitSubtitle(subtitle string) {
	if subtitle == "" {
		return
	}

	// 格式化副标题，添加空格以便于正则匹配
	subtitle = " " + strings.TrimSpace(subtitle) + " "

	// 定义正则表达式模式
	// 注意：Go 不支持负向断言，所以简化了正则表达式
	titleEpisodelRe := regexp.MustCompile(`(?i)Episode\s+(\d{1,4})`)
	subtitleSeasonAllRe := regexp.MustCompile(`(?i)[全共]\s*(\d+)\s*季`)
	subtitleSeasonRe := regexp.MustCompile(`(?i)[第\s]+([0-9S\-]+)\s*季`)
	subtitleEpisodeBetweenRe := regexp.MustCompile(`(?i)[第]*\s*(\d+)\s*[集话話期幕]?\s*-\s*第*\s*(\d+)\s*[集话話期幕]`)
	subtitleEpisodeRe := regexp.MustCompile(`(?i)[第\s]+([0-9EP]+)\s*[集话話期幕]`)
	subtitleEpisodeAllRe := regexp.MustCompile(`(?i)(\d+)\s*集\s*全|[全共]\s*(\d+)\s*[集话話期幕]`)

	// 1. 匹配 Episode X 格式
	if match := titleEpisodelRe.FindStringSubmatch(subtitle); match != nil {
		episode, err := strconv.Atoi(match[1])
		if err == nil && episode < 10000 {
			m.BeginEpisode = &episode
			m.TotalEpisode = 1
			m.Type = MediaTypeTV
		}
		return
	}

	// 2. 匹配 全X季 / X季全 格式
	if match := subtitleSeasonAllRe.FindStringSubmatch(subtitle); match != nil {
		seasonAll := match[1]
		if seasonAll != "" && m.BeginSeason == nil && m.BeginEpisode == nil {
			season, err := strconv.Atoi(seasonAll)
			if err == nil {
				m.TotalSeason = season
				beginSeason := 1
				m.BeginSeason = &beginSeason
				m.EndSeason = &season
				m.Type = MediaTypeTV
			}
		}
		return
	}

	// 3. 匹配 第X季 / 第X-Y季 格式
	if match := subtitleSeasonRe.FindStringSubmatch(subtitle); match != nil {
		seasons := strings.ToUpper(match[1])
		seasons = strings.ReplaceAll(seasons, "S", "")
		seasons = strings.TrimSpace(seasons)

		var beginSeason, endSeason int
		var err error

		if strings.Contains(seasons, "-") {
			// 多季格式：X-Y
			seasonParts := strings.Split(seasons, "-")
			if len(seasonParts) >= 2 {
				beginSeason, err = strconv.Atoi(strings.TrimSpace(seasonParts[0]))
				if err != nil {
					return
				}
				endSeason, err = strconv.Atoi(strings.TrimSpace(seasonParts[1]))
				if err != nil {
					return
				}
			}
		} else {
			// 单季格式：X
			beginSeason, err = strconv.Atoi(seasons)
			if err != nil {
				return
			}
			endSeason = beginSeason
		}

		// 验证季数范围
		if beginSeason > 100 || endSeason > 100 {
			return
		}

		// 更新季信息
		if m.BeginSeason == nil {
			m.BeginSeason = &beginSeason
			m.TotalSeason = 1
		}

		if m.BeginSeason != nil && m.EndSeason == nil && endSeason != beginSeason {
			m.EndSeason = &endSeason
			m.TotalSeason = endSeason - beginSeason + 1
		}

		m.Type = MediaTypeTV
	}

	// 4. 匹配 第X-Y集 / X-Y集 格式
	if match := subtitleEpisodeBetweenRe.FindStringSubmatch(subtitle); match != nil {
		beginEpisodeStr := match[1]
		endEpisodeStr := match[2]

		beginEpisode, err1 := strconv.Atoi(beginEpisodeStr)
		endEpisode, err2 := strconv.Atoi(endEpisodeStr)

		if err1 == nil && err2 == nil && beginEpisode < 10000 && endEpisode < 10000 {
			m.BeginEpisode = &beginEpisode
			m.EndEpisode = &endEpisode
			m.TotalEpisode = endEpisode - beginEpisode + 1
			m.Type = MediaTypeTV
		}
		return
	}

	// 6. 匹配 X集全 / 全X集 格式
	if match := subtitleEpisodeAllRe.FindStringSubmatch(subtitle); match != nil {
		episodeAll := match[1]
		if episodeAll == "" {
			episodeAll = match[2]
		}

		if episodeAll != "" && m.BeginEpisode == nil {
			totalEpisode, err := strconv.Atoi(episodeAll)
			if err == nil {
				m.TotalEpisode = totalEpisode
				m.Type = MediaTypeTV
			}
		}
		return
	}

	// 5. 匹配 第X集 / X集 格式
	if match := subtitleEpisodeRe.FindStringSubmatch(subtitle); match != nil {
		episodes := strings.ToUpper(match[1])
		episodes = strings.ReplaceAll(episodes, "E", "")
		episodes = strings.ReplaceAll(episodes, "P", "")
		episodes = strings.TrimSpace(episodes)

		var beginEpisode, endEpisode int
		var err error

		if strings.Contains(episodes, "-") {
			// 多集格式：X-Y
			episodeParts := strings.Split(episodes, "-")
			if len(episodeParts) >= 2 {
				beginEpisode, err = strconv.Atoi(strings.TrimSpace(episodeParts[0]))
				if err != nil {
					return
				}
				endEpisode, err = strconv.Atoi(strings.TrimSpace(episodeParts[1]))
				if err != nil {
					return
				}
			}
		} else {
			// 单集格式：X
			beginEpisode, err = strconv.Atoi(episodes)
			if err != nil {
				return
			}
			endEpisode = beginEpisode
		}

		// 验证集数范围
		if beginEpisode >= 10000 || endEpisode >= 10000 {
			return
		}

		// 更新集信息
		if m.BeginEpisode == nil {
			m.BeginEpisode = &beginEpisode
			m.TotalEpisode = 1
		}

		if m.BeginEpisode != nil && m.EndEpisode == nil && endEpisode != beginEpisode {
			m.EndEpisode = &endEpisode
			m.TotalEpisode = endEpisode - beginEpisode + 1
		}

		m.Type = MediaTypeTV
		return
	}
}

// ToDict 将MetaBase转换为字典，与Python版本to_dict功能一致
func (m *MetaBase) ToDict() map[string]interface{} {
	result := make(map[string]interface{})

	// 基本属性
	result["isfile"] = m.Isfile
	result["title"] = m.Title
	result["org_string"] = m.OrgString
	result["subtitle"] = m.Subtitle
	result["cn_name"] = m.CnName
	result["en_name"] = m.EnName
	result["year"] = m.Year

	// 媒体类型
	result["type"] = string(m.Type)

	// 季相关
	result["begin_season"] = m.BeginSeason
	result["end_season"] = m.EndSeason
	result["total_season"] = m.TotalSeason
	result["season"] = m.Season()
	result["sea"] = m.Sea()
	result["season_seq"] = m.SeasonSeq()
	result["season_list"] = m.SeasonList()

	// 集相关
	result["begin_episode"] = m.BeginEpisode
	result["end_episode"] = m.EndEpisode
	result["total_episode"] = m.TotalEpisode
	result["episode"] = m.Episode()
	result["episode_list"] = m.EpisodeList()
	result["episodes"] = m.Episodes()
	result["episode_seqs"] = m.EpisodeSeqs()
	result["episode_seq"] = m.EpisodeSeq()

	// 组合属性
	result["season_episode"] = m.SeasonEpisode()

	// 资源属性
	result["resource_type"] = string(m.ResourceType)
	result["resource_effect"] = string(m.ResourceEffect)
	result["resource_pix"] = string(m.ResourcePix)
	result["resource_team"] = m.ResourceTeam
	result["customization"] = m.Customization
	result["web_source"] = m.WebSource
	result["video_encode"] = m.VideoEncode
	result["audio_encode"] = m.AudioEncode
	result["resource_term"] = m.ResourceTerm()
	result["edition"] = m.Edition()
	result["release_group"] = m.ReleaseGroup()
	result["video_term"] = m.VideoTerm()
	result["audio_term"] = m.AudioTerm()

	// 其他
	result["part"] = m.Part
	result["tmdbid"] = m.TMDBID
	result["doubanid"] = m.DoubanID
	result["applied_words"] = m.AppliedWords
	result["name"] = m.Name()

	return result
}
