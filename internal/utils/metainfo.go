package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// MetaInfo 识别元数�?type MetaInfo struct {
	// 是否处理的文�?	IsFile bool
	
	// 原字符串
	OrgString string
	
	// 原标�?	Title string
	
	// 副标�?	Subtitle string
	
	// 类型 电影、电视剧
	Type string
	
	// 名称
	Name string
	
	// 识别的中文名
	CnName string
	
	// 识别的英文名
	EnName string
	
	// 年份
	Year string
	
	// 总季�?	TotalSeason int
	
	// 识别的开始季 数字
	BeginSeason int
	
	// 识别的结束季 数字
	EndSeason int
	
	// 总集�?	TotalEpisode int
	
	// 识别的开始集
	BeginEpisode int
	
	// 识别的结束集
	EndEpisode int
	
	// SxxExx
	SeasonEpisode string
	
	// 集列�?	EpisodeList []int
	
	// Partx Cd Dvd Disk Disc
	Part string
	
	// 识别的资源类�?	ResourceType string
	
	// 识别的效�?	ResourceEffect string
	
	// 识别的分辨率
	ResourcePix string
	
	// 识别的制作组/字幕�?	ResourceTeam string
	
	// 视频编码
	VideoEncode string
	
	// 音频编码
	AudioEncode string
	
	// 资源类型
	Edition string
	
	// 流媒体平�?	WebSource string
	
	// 应用的识别词信息
	ApplyWords []string
}

// NewMetaInfo 创建新的MetaInfo实例
func NewMetaInfo(filename string) *MetaInfo {
	meta := &MetaInfo{
		OrgString:   filename,
		EpisodeList: make([]int, 0),
		ApplyWords:  make([]string, 0),
	}
	
	// 解析文件名获取集数等信息
	meta.parseFilename(filename)
	
	return meta
}

// parseFilename 解析文件�?func (m *MetaInfo) parseFilename(filename string) {
	// 提取集数信息
	episodeRegex := regexp.MustCompile(`[Ee](\d+)`)
	matches := episodeRegex.FindAllStringSubmatch(filename, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			if episode, err := strconv.Atoi(match[1]); err == nil {
				m.EpisodeList = append(m.EpisodeList, episode)
			}
		}
	}
	
	// 提取季信�?	seasonRegex := regexp.MustCompile(`[Ss](\d+)`)
	seasonMatches := seasonRegex.FindAllStringSubmatch(filename, -1)
	
	for _, match := range seasonMatches {
		if len(match) > 1 {
			if season, err := strconv.Atoi(match[1]); err == nil {
				if m.BeginSeason == 0 {
					m.BeginSeason = season
				}
				m.EndSeason = season
			}
		}
	}
	
	// 构造SeasonEpisode
	if m.BeginSeason > 0 {
		m.SeasonEpisode = "S" + strconv.Itoa(m.BeginSeason)
		if len(m.EpisodeList) > 0 {
			m.SeasonEpisode += "E"
			for i, episode := range m.EpisodeList {
				if i > 0 {
					m.SeasonEpisode += "_"
				}
				m.SeasonEpisode += strconv.Itoa(episode)
			}
		}
	} else if len(m.EpisodeList) > 0 {
		m.SeasonEpisode = "E"
		for i, episode := range m.EpisodeList {
			if i > 0 {
				m.SeasonEpisode += "_"
			}
			m.SeasonEpisode += strconv.Itoa(episode)
		}
	}
	
	// 提取年份信息
	yearRegex := regexp.MustCompile(`(19|20)\d{2}`)
	yearMatches := yearRegex.FindAllString(filename, -1)
	if len(yearMatches) > 0 {
		m.Year = yearMatches[0]
	}
	
	// 简化处理标�?	parts := strings.Split(filename, ".")
	if len(parts) > 0 {
		m.Name = parts[0]
	}
}
