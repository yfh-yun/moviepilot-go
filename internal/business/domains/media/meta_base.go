package media

import (
	"fmt"
	"strings"
)

// MetaBase 统一的元信息基础结构与公共逻辑
type MetaBase struct {
	// 名称相关
	Title     string `json:"title"`      // 标题
	OrgString string `json:"org_string"` // 原始字符串
	Subtitle  string `json:"subtitle"`   // 副标题
	CnName    string `json:"cn_name"`    // 中文名
	EnName    string `json:"en_name"`    // 英文名
	Year      int    `json:"year"`       // 年份

	// 媒体类型
	Type MediaType `json:"type"` // 媒体类型

	// 季相关
	BeginSeason int `json:"begin_season"` // 开始季
	EndSeason   int `json:"end_season"`   // 结束季
	TotalSeason int `json:"total_season"` // 总季数

	// 集相关
	BeginEpisode int `json:"begin_episode"` // 开始集
	EndEpisode   int `json:"end_episode"`   // 结束集
	TotalEpisode int `json:"total_episode"` // 总集数

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
	Part         string   `json:"part"`          // 部分
	TMDBID       int64    `json:"tmdbid"`        // TMDB ID
	DoubanID     string   `json:"doubanid"`      // 豆瓣 ID
	AppliedWords []string `json:"applied_words"` // 应用的识别词
}

// NewMetaBase 创建新的MetaBase实例
func NewMetaBase(title, orgString, subtitle string) *MetaBase {
	return &MetaBase{
		Title:     title,
		OrgString: orgString,
		Subtitle:  subtitle,
		Type:      MediaTypeUnknown,
	}
}

// Merge 合并元信息
func (m *MetaBase) Merge(other *MetaBase) {
	if other == nil {
		return
	}

	// 合并名称相关
	if m.CnName == "" {
		m.CnName = other.CnName
	}
	if m.EnName == "" {
		m.EnName = other.EnName
	}
	if m.Year == 0 {
		m.Year = other.Year
	}

	// 合并季相关
	if m.BeginSeason == 0 {
		m.BeginSeason = other.BeginSeason
	}
	if m.EndSeason == 0 {
		m.EndSeason = other.EndSeason
	}
	if m.TotalSeason == 0 {
		m.TotalSeason = other.TotalSeason
	}

	// 合并集相关
	if m.BeginEpisode == 0 {
		m.BeginEpisode = other.BeginEpisode
	}
	if m.EndEpisode == 0 {
		m.EndEpisode = other.EndEpisode
	}
	if m.TotalEpisode == 0 {
		m.TotalEpisode = other.TotalEpisode
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
	if m.TMDBID == 0 {
		m.TMDBID = other.TMDBID
	}
	if m.DoubanID == "" {
		m.DoubanID = other.DoubanID
	}
	if len(m.AppliedWords) == 0 {
		m.AppliedWords = other.AppliedWords
	}
}

// Season 获取季数（单季时返回具体季数，多季时返回范围）
func (m *MetaBase) Season() string {
	if m.BeginSeason == m.EndSeason && m.BeginSeason > 0 {
		return fmt.Sprintf("%d", m.BeginSeason)
	} else if m.BeginSeason > 0 && m.EndSeason > 0 {
		return fmt.Sprintf("%d-%d", m.BeginSeason, m.EndSeason)
	}
	return ""
}

// SeasonList 获取季列表
func (m *MetaBase) SeasonList() []int {
	var seasons []int
	if m.BeginSeason > 0 && m.EndSeason > 0 {
		for i := m.BeginSeason; i <= m.EndSeason; i++ {
			seasons = append(seasons, i)
		}
	} else if m.BeginSeason > 0 {
		seasons = append(seasons, m.BeginSeason)
	}
	return seasons
}

// Episode 获取集数（单集时返回具体集数，多集时返回范围）
func (m *MetaBase) Episode() string {
	if m.BeginEpisode == m.EndEpisode && m.BeginEpisode > 0 {
		return fmt.Sprintf("%d", m.BeginEpisode)
	} else if m.BeginEpisode > 0 && m.EndEpisode > 0 {
		return fmt.Sprintf("%d-%d", m.BeginEpisode, m.EndEpisode)
	}
	return ""
}

// EpisodeList 获取集列表
func (m *MetaBase) EpisodeList() []int {
	var episodes []int
	if m.BeginEpisode > 0 && m.EndEpisode > 0 {
		for i := m.BeginEpisode; i <= m.EndEpisode; i++ {
			episodes = append(episodes, i)
		}
	} else if m.BeginEpisode > 0 {
		episodes = append(episodes, m.BeginEpisode)
	}
	return episodes
}

// IsInSeason 判断是否在指定季范围内
func (m *MetaBase) IsInSeason(season int) bool {
	return m.BeginSeason <= season && season <= m.EndSeason
}

// IsInEpisode 判断是否在指定集范围内
func (m *MetaBase) IsInEpisode(episode int) bool {
	return m.BeginEpisode <= episode && episode <= m.EndEpisode
}

// SetSeason 设置季信息
func (m *MetaBase) SetSeason(season int) {
	m.BeginSeason = season
	m.EndSeason = season
	m.TotalSeason = 1
}

// SetEpisodes 设置集范围
func (m *MetaBase) SetEpisodes(begin, end int) {
	m.BeginEpisode = begin
	m.EndEpisode = end
	if begin > 0 && end > 0 {
		m.TotalEpisode = end - begin + 1
	}
}

// SetEpisode 设置单集
func (m *MetaBase) SetEpisode(episode int) {
	m.SetEpisodes(episode, episode)
}

// InitSubtitle 初始化副标题，解析季集信息
func (m *MetaBase) InitSubtitle(subtitle string) {
	if subtitle == "" {
		return
	}

	// 这里实现副标题解析逻辑，解析“第X季/全X季/第X集-第Y集/全X集”等
	// 简化实现，后续可以扩展
	subtitle = strings.TrimSpace(subtitle)

	// 示例：解析“第1季”
	if strings.Contains(subtitle, "第") && strings.Contains(subtitle, "季") {
		// 提取季数
		var season int
		fmt.Sscanf(subtitle, "第%d季", &season)
		if season > 0 {
			m.SetSeason(season)
		}
	}

	// 示例：解析“第1集-第10集”
	if strings.Contains(subtitle, "第") && strings.Contains(subtitle, "集") {
		var begin, end int
		if strings.Contains(subtitle, "-") {
			fmt.Sscanf(subtitle, "第%d集-第%d集", &begin, &end)
			if begin > 0 && end > 0 {
				m.SetEpisodes(begin, end)
			}
		} else {
			fmt.Sscanf(subtitle, "第%d集", &begin)
			if begin > 0 {
				m.SetEpisode(begin)
			}
		}
	}
}
