package meta

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MetaAnime 动漫元数据类
type MetaAnime struct {
	MetaVideo
	
	// 动漫特有信息
	AnimeID      string `json:"anime_id"`       // 动漫平台ID
	BangumiID    int    `json:"bangumi_id"`     // Bangumi ID
	MyAnimeListID int   `json:"myanimelist_id"` // MyAnimeList ID
	
	// 动漫特殊属性
	OriginalName string   `json:"original_name"` // 原始名称（通常是日语）
	Aliases      []string `json:"aliases"`       // 别名列表
	Studio       []string `json:"studio"`        // 制作公司
	Tags         []string `json:"tags"`          // 标签
	
	// 番剧相关
	SeasonInfo   string   `json:"season_info"`   // 季度信息（如 "2024年1月"）
	IsSpecial    bool     `json:"is_special"`    // 是否特别篇
	IsOVA        bool     `json:"is_ova"`        // 是否OVA
	IsMovie      bool     `json:"is_movie"`      // 是否剧场版
	IsWeb        bool     `json:"is_web"`        // 是否网络动画
	IsTV         bool     `json:"is_tv"`         // 是否TV动画
	
	// 集数信息
	TotalEpisodes int    `json:"total_episodes"` // 总集数
	CurrentEpisode int    `json:"current_episode"` // 当前集数
	StartDate     string `json:"start_date"`     // 开播日期
	EndDate       string `json:"end_date"`       // 完结日期
	IsOngoing     bool   `json:"is_ongoing"`     // 是否连载中
	
	// 评分相关
	Rating        float64 `json:"rating"`        // 评分
	ScoreCount    int     `json:"score_count"`   // 评分人数
	Rank          int     `json:"rank"`          // 排名
	
	// 观看状态
	WatchStatus   string `json:"watch_status"`   // 观看状态
	
	// 搜索和匹配
	SearchTerms   []string `json:"search_terms"` // 搜索词列表
	MatchScore    float64  `json:"match_score"`  // 匹配得分
}

// NewMetaAnime 创建动漫元数据实例
func NewMetaAnime(name string) *MetaAnime {
	return &MetaAnime{
		MetaVideo: *NewMetaVideo(name).(*MetaVideo),
		MediaType:     MediaTypeAnime,
		IsAnime:       true,
		IsAnimation:   true,
		Aliases:       make([]string, 0),
		Studio:        make([]string, 0),
		Tags:          make([]string, 0),
		SearchTerms:   make([]string, 0),
		SearchKeywords: make([]string, 0),
	}
}

// Clone 克隆动漫元数据
func (a *MetaAnime) Clone() MetaInfo {
	clone := *a
	// 深拷贝所有切片
	clone.Tags = make([]string, len(a.Tags))
	copy(clone.Tags, a.Tags)
	clone.Aliases = make([]string, len(a.Aliases))
	copy(clone.Aliases, a.Aliases)
	clone.Studio = make([]string, len(a.Studio))
	copy(clone.Studio, a.Studio)
	clone.SearchTerms = make([]string, len(a.SearchTerms))
	copy(clone.SearchTerms, a.SearchTerms)
	clone.SearchKeywords = make([]string, len(a.SearchKeywords))
	copy(clone.SearchKeywords, a.SearchKeywords)
	return &clone
}

// IsValid 判断动漫元数据是否有效
func (a *MetaAnime) IsValid() bool {
	return (a.ParseStatus == ParseStatusSuccess || a.ParseStatus == ParseStatusPartially) && 
	       (a.AnimeID != "" || a.BangumiID > 0 || a.MyAnimeListID > 0 || a.Title != "")
}

// ToString 转换为字符串表示
func (a *MetaAnime) ToString() string {
	if a.IsTV && a.CurrentEpisode > 0 {
		return fmt.Sprintf("%s - 第%d话 (%s)", a.Title, a.CurrentEpisode, a.Year)
	} else if a.IsMovie {
		return fmt.Sprintf("%s - 剧场版 (%s)", a.Title, a.Year)
	} else if a.IsOVA {
		return fmt.Sprintf("%s - OVA (%s)", a.Title, a.Year)
	}
	return fmt.Sprintf("%s (%s) [%s]", a.Title, a.Year, a.MediaType)
}

// ID相关方法

// GetAnimeID 获取动漫平台ID
func (a *MetaAnime) GetAnimeID() string {
	return a.AnimeID
}

// SetAnimeID 设置动漫平台ID
func (a *MetaAnime) SetAnimeID(id string) {
	a.AnimeID = id
	a.UpdatedAt = time.Now()
}

// GetBangumiID 获取Bangumi ID
func (a *MetaAnime) GetBangumiID() int {
	return a.BangumiID
}

// SetBangumiID 设置Bangumi ID
func (a *MetaAnime) SetBangumiID(id int) {
	a.BangumiID = id
	a.UpdatedAt = time.Now()
}

// GetMyAnimeListID 获取MyAnimeList ID
func (a *MetaAnime) GetMyAnimeListID() int {
	return a.MyAnimeListID
}

// SetMyAnimeListID 设置MyAnimeList ID
func (a *MetaAnime) SetMyAnimeListID(id int) {
	a.MyAnimeListID = id
	a.UpdatedAt = time.Now()
}

// 名称相关方法

// GetOriginalName 获取原始名称
func (a *MetaAnime) GetOriginalName() string {
	return a.OriginalName
}

// SetOriginalName 设置原始名称
func (a *MetaAnime) SetOriginalName(name string) {
	a.OriginalName = name
	a.UpdatedAt = time.Now()
}

// GetAliases 获取别名列表
func (a *MetaAnime) GetAliases() []string {
	return a.Aliases
}

// AddAlias 添加别名
func (a *MetaAnime) AddAlias(alias string) {
	for _, a := range a.Aliases {
		if a == alias {
			return // 避免重复
		}
	}
	a.Aliases = append(a.Aliases, alias)
	a.UpdatedAt = time.Now()
}

// 制作相关方法

// GetStudio 获取制作公司列表
func (a *MetaAnime) GetStudio() []string {
	return a.Studio
}

// AddStudio 添加制作公司
func (a *MetaAnime) AddStudio(studio string) {
	for _, s := range a.Studio {
		if s == studio {
			return // 避免重复
		}
	}
	a.Studio = append(a.Studio, studio)
	a.UpdatedAt = time.Now()
}

// 类型相关方法

// IsTVSeries 判断是否为TV动画
func (a *MetaAnime) IsTVSeries() bool {
	return a.IsTV
}

// SetTV 设置为TV动画
func (a *MetaAnime) SetTV() {
	a.IsTV = true
	a.IsMovie = false
	a.IsOVA = false
	a.IsSpecial = false
	a.IsWeb = false
	a.MediaType = MediaTypeAnimeTV
	a.UpdatedAt = time.Now()
}

// IsMovie 判断是否为剧场版
func (a *MetaAnime) IsMovieType() bool {
	return a.IsMovie
}

// SetMovie 设置为剧场版
func (a *MetaAnime) SetMovie() {
	a.IsMovie = true
	a.IsTV = false
	a.IsOVA = false
	a.IsSpecial = false
	a.IsWeb = false
	a.MediaType = MediaTypeAnimeMovie
	a.UpdatedAt = time.Now()
}

// IsOVA 判断是否为OVA
func (a *MetaAnime) IsOVA() bool {
	return a.IsOVA
}

// SetOVA 设置为OVA
func (a *MetaAnime) SetOVA() {
	a.IsOVA = true
	a.IsTV = false
	a.IsMovie = false
	a.IsSpecial = false
	a.IsWeb = false
	a.MediaType = MediaTypeAnimeOVA
	a.UpdatedAt = time.Now()
}

// IsSpecial 判断是否为特别篇
func (a *MetaAnime) IsSpecial() bool {
	return a.IsSpecial
}

// SetSpecial 设置为特别篇
func (a *MetaAnime) SetSpecial() {
	a.IsSpecial = true
	a.IsTV = false
	a.IsMovie = false
	a.IsOVA = false
	a.IsWeb = false
	a.MediaType = MediaTypeAnimeSpecial
	a.UpdatedAt = time.Now()
}

// IsWeb 判断是否为网络动画
func (a *MetaAnime) IsWeb() bool {
	return a.IsWeb
}

// SetWeb 设置为网络动画
func (a *MetaAnime) SetWeb() {
	a.IsWeb = true
	a.IsTV = false
	a.IsMovie = false
	a.IsOVA = false
	a.IsSpecial = false
	a.MediaType = MediaTypeAnimeWeb
	a.UpdatedAt = time.Now()
}

// 季度信息

// GetSeasonInfo 获取季度信息
func (a *MetaAnime) GetSeasonInfo() string {
	return a.SeasonInfo
}

// SetSeasonInfo 设置季度信息
func (a *MetaAnime) SetSeasonInfo(season string) {
	a.SeasonInfo = season
	a.UpdatedAt = time.Now()
}

// 集数信息

// GetTotalEpisodes 获取总集数
func (a *MetaAnime) GetTotalEpisodes() int {
	return a.TotalEpisodes
}

// SetTotalEpisodes 设置总集数
func (a *MetaAnime) SetTotalEpisodes(episodes int) {
	a.TotalEpisodes = episodes
	a.UpdatedAt = time.Now()
}

// GetCurrentEpisode 获取当前集数
func (a *MetaAnime) GetCurrentEpisode() int {
	return a.CurrentEpisode
}

// SetCurrentEpisode 设置当前集数
func (a *MetaAnime) SetCurrentEpisode(episode int) {
	a.CurrentEpisode = episode
	a.UpdatedAt = time.Now()
}

// 日期信息

// GetStartDate 获取开播日期
func (a *MetaAnime) GetStartDate() string {
	return a.StartDate
}

// SetStartDate 设置开播日期
func (a *MetaAnime) SetStartDate(date string) {
	a.StartDate = date
	a.UpdatedAt = time.Now()
}

// GetEndDate 获取完结日期
func (a *MetaAnime) GetEndDate() string {
	return a.EndDate
}

// SetEndDate 设置完结日期
func (a *MetaAnime) SetEndDate(date string) {
	a.EndDate = date
	// 如果有完结日期，则不是连载中
	if date != "" {
		a.IsOngoing = false
	}
	a.UpdatedAt = time.Now()
}

// IsOngoing 判断是否连载中
func (a *MetaAnime) IsOngoing() bool {
	return a.IsOngoing
}

// SetOngoing 设置连载状态
func (a *MetaAnime) SetOngoing(ongoing bool) {
	a.IsOngoing = ongoing
	if ongoing {
		a.EndDate = ""
	}
	a.UpdatedAt = time.Now()
}

// 评分信息

// GetRating 获取评分
func (a *MetaAnime) GetRating() float64 {
	return a.Rating
}

// SetRating 设置评分
func (a *MetaAnime) SetRating(rating float64) {
	a.Rating = rating
	a.UpdatedAt = time.Now()
}

// GetScoreCount 获取评分人数
func (a *MetaAnime) GetScoreCount() int {
	return a.ScoreCount
}

// SetScoreCount 设置评分人数
func (a *MetaAnime) SetScoreCount(count int) {
	a.ScoreCount = count
	a.UpdatedAt = time.Now()
}

// GetRank 获取排名
func (a *MetaAnime) GetRank() int {
	return a.Rank
}

// SetRank 设置排名
func (a *MetaAnime) SetRank(rank int) {
	a.Rank = rank
	a.UpdatedAt = time.Now()
}

// 观看状态

// GetWatchStatus 获取观看状态
func (a *MetaAnime) GetWatchStatus() string {
	return a.WatchStatus
}

// SetWatchStatus 设置观看状态
func (a *MetaAnime) SetWatchStatus(status string) {
	a.WatchStatus = status
	a.UpdatedAt = time.Now()
}

// 搜索相关方法

// GetSearchTerms 获取搜索词列表
func (a *MetaAnime) GetSearchTerms() []string {
	return a.SearchTerms
}

// AddSearchTerm 添加搜索词
func (a *MetaAnime) AddSearchTerm(term string) {
	for _, t := range a.SearchTerms {
		if t == term {
			return // 避免重复
		}
	}
	a.SearchTerms = append(a.SearchTerms, term)
	a.UpdatedAt = time.Now()
}

// GetMatchScore 获取匹配得分
func (a *MetaAnime) GetMatchScore() float64 {
	return a.MatchScore
}

// SetMatchScore 设置匹配得分
func (a *MetaAnime) SetMatchScore(score float64) {
	a.MatchScore = score
	a.UpdatedAt = time.Now()
}

// 动漫特有的解析方法

// ParseAnimeName 解析动漫名称
func (a *MetaAnime) ParseAnimeName() {
	a.ParseStatus = ParseStatusFailed
	a.Confidence = 0.0
	
	// 调用基础解析方法
	a.ParseName()
	
	// 重新设置为动漫类型
	a.IsAnime = true
	a.IsAnimation = true
	a.MediaType = MediaTypeAnime
	
	// 进一步解析动漫特有的信息
	a.ParseEpisodeNumber()
	a.ParseAnimeType()
	a.ParseAnimeSeason()
	
	// 生成搜索词
	a.GenerateAnimeSearchTerms()
	
	// 更新解析状态
	if a.Title != "" || a.OriginalName != "" {
		a.ParseStatus = ParseStatusSuccess
		if a.Confidence < 0.5 {
			a.Confidence = 0.5
		}
	}
}

// ParseEpisodeNumber 解析动漫集数
func (a *MetaAnime) ParseEpisodeNumber() {
	name := a.Name
	
	// 常见的动漫集数格式
	patterns := []struct {
		Regex *regexp.Regexp
		EpisodeIndex int
	}{{
		// 第1话 格式
		Regex: regexp.MustCompile(`第(\d{1,3})话`),
		EpisodeIndex: 1,
	}, {
		// 第1集 格式
		Regex: regexp.MustCompile(`第(\d{1,3})集`),
		EpisodeIndex: 1,
	}, {
		// [01] 格式
		Regex: regexp.MustCompile(`\[(\d{1,3})\]`),
		EpisodeIndex: 1,
	}, {
		// - 01 格式
		Regex: regexp.MustCompile(`\-\s*(\d{1,3})`),
		EpisodeIndex: 1,
	}, {
		// _01 格式
		Regex: regexp.MustCompile(`_(\d{1,3})`),
		EpisodeIndex: 1,
	}, {
		// episode 01 格式
		Regex: regexp.MustCompile(`episode\s+(\d{1,3})`),
		EpisodeIndex: 1,
	}, {
		// ep01 格式
		Regex: regexp.MustCompile(`ep(\d{1,3})`),
		EpisodeIndex: 1,
	}, {
		// 单独数字格式，前后非数字
		Regex: regexp.MustCompile(`[^\d](\d{1,3})[^\d]`),
		EpisodeIndex: 1,
	}}
	
	// 转换为小写以进行匹配
	nameLower := strings.ToLower(name)
	
	for _, pattern := range patterns {
		matches := pattern.Regex.FindStringSubmatch(nameLower)
		if len(matches) >= 2 {
			episode, err := strconv.Atoi(matches[pattern.EpisodeIndex])
			if err == nil && episode > 0 {
				a.CurrentEpisode = episode
				a.Episode = episode // 与通用视频保持一致
				
				// 检查是否为OVA或特别篇
				a.checkSpecialTypes(nameLower)
				
				return
			}
		}
	}
}

// checkSpecialTypes 检查是否为特殊类型
func (a *MetaAnime) checkSpecialTypes(name string) {
	// 检查是否为OVA
	if strings.Contains(name, "ova") {
		a.SetOVA()
		a.Confidence += 0.1
	}
	// 检查是否为剧场版
	if strings.Contains(name, "movie") || strings.Contains(name, "剧场") || strings.Contains(name, "劇場") {
		a.SetMovie()
		a.Confidence += 0.1
	}
	// 检查是否为特别篇
	if strings.Contains(name, "special") || strings.Contains(name, "特别") || strings.Contains(name, "SP") {
		a.SetSpecial()
		a.Confidence += 0.1
	}
	// 检查是否为网络动画
	if strings.Contains(name, "web") || strings.Contains(name, "网络") {
		a.SetWeb()
		a.Confidence += 0.1
	}
	// 如果没有特殊类型，默认为TV动画
	if !a.IsOVA && !a.IsMovie && !a.IsSpecial && !a.IsWeb {
		a.SetTV()
	}
}

// ParseAnimeType 解析动漫类型
func (a *MetaAnime) ParseAnimeType() {
	name := strings.ToLower(a.Name)
	
	// 检查是否为OVA
	if strings.Contains(name, "ova") || strings.Contains(name, "ova") || strings.Contains(name, "oad") {
		a.SetOVA()
	}
	// 检查是否为剧场版
	if strings.Contains(name, "movie") || strings.Contains(name, "剧场") || strings.Contains(name, "劇場") || strings.Contains(name, "cinema") {
		a.SetMovie()
	}
	// 检查是否为特别篇
	if strings.Contains(name, "special") || strings.Contains(name, "特别") || strings.Contains(name, "SP") {
		a.SetSpecial()
	}
	// 检查是否为网络动画
	if strings.Contains(name, "web") || strings.Contains(name, "网络") || strings.Contains(name, "online") {
		a.SetWeb()
	}
}

// ParseAnimeSeason 解析动漫季度
func (a *MetaAnime) ParseAnimeSeason() {
	name := strings.ToLower(a.Name)
	
	// 常见的季度格式
	seasonPatterns := []struct {
		Regex *regexp.Regexp
		YearIndex, SeasonIndex int
	}{{
		// 2024年1月 格式
		Regex: regexp.MustCompile(`(\d{4})年(\d{1,2})月`),
		YearIndex: 1,
		SeasonIndex: 2,
	}, {
		// 2024 Winter 格式
		Regex: regexp.MustCompile(`(\d{4})\s+(winter|spring|summer|fall|autumn)`),
		YearIndex: 1,
		SeasonIndex: 2,
	}, {
		// 24冬 格式
		Regex: regexp.MustCompile(`(\d{2})(冬|春|夏|秋)`),
		YearIndex: 1,
		SeasonIndex: 2,
	}}
	
	for _, pattern := range seasonPatterns {
		matches := pattern.Regex.FindStringSubmatch(name)
		if len(matches) >= 3 {
			// 处理年份
			yearStr := matches[pattern.YearIndex]
			var year int
			var err error
			
			if len(yearStr) == 2 {
				// 处理两位数年份
				year, err = strconv.Atoi("20" + yearStr)
			} else {
				year, err = strconv.Atoi(yearStr)
			}
			
			if err == nil {
				a.Year = year
				
				// 处理季度
				season := matches[pattern.SeasonIndex]
				seasonInfo := a.formatSeason(year, season)
				a.SeasonInfo = seasonInfo
				
				return
			}
		}
	}
}

// formatSeason 格式化季度信息
func (a *MetaAnime) formatSeason(year int, season string) string {
	// 转换季节名称为中文
	seasonCN := season
	switch strings.ToLower(season) {
	case "winter", "冬":
		seasonCN = "冬季"
	case "spring", "春":
		seasonCN = "春季"
	case "summer", "夏":
		seasonCN = "夏季"
	case "fall", "autumn", "秋":
		seasonCN = "秋季"
	default:
		// 处理数字月份
		if month, err := strconv.Atoi(season); err == nil {
			if month >= 1 && month <= 3 {
				seasonCN = "冬季"
			} else if month >= 4 && month <= 6 {
				seasonCN = "春季"
			} else if month >= 7 && month <= 9 {
				seasonCN = "夏季"
			} else if month >= 10 && month <= 12 {
				seasonCN = "秋季"
			}
		}
	}
	
	return fmt.Sprintf("%d年%s", year, seasonCN)
}

// GenerateAnimeSearchTerms 生成动漫搜索词
func (a *MetaAnime) GenerateAnimeSearchTerms() {
	terms := make([]string, 0)
	
	// 添加标题
	if a.Title != "" {
		terms = append(terms, strings.ToLower(a.Title))
	}
	
	// 添加原始名称
	if a.OriginalName != "" {
		terms = append(terms, strings.ToLower(a.OriginalName))
	}
	
	// 添加别名
	for _, alias := range a.Aliases {
		if alias != "" {
			terms = append(terms, strings.ToLower(alias))
		}
	}
	
	// 添加制作公司
	for _, studio := range a.Studio {
		if studio != "" {
			terms = append(terms, strings.ToLower(studio))
		}
	}
	
	// 特殊处理动漫名称的常见缩写和变体
	for _, term := range terms {
		// 添加无空格版本
		noSpaceTerm := strings.ReplaceAll(term, " ", "")
		if noSpaceTerm != term && !contains(terms, noSpaceTerm) {
			terms = append(terms, noSpaceTerm)
		}
		
		// 添加无连字符版本
		noHyphenTerm := strings.ReplaceAll(term, "-", "")
		if noHyphenTerm != term && !contains(terms, noHyphenTerm) {
			terms = append(terms, noHyphenTerm)
		}
	}
	
	// 去重
	uniqueTerms := make([]string, 0)
	termMap := make(map[string]bool)
	for _, term := range terms {
		if term != "" && !termMap[term] {
			termMap[term] = true
			uniqueTerms = append(uniqueTerms, term)
		}
	}
	
	a.SearchTerms = uniqueTerms
	
	// 同时更新通用搜索关键词
	a.SearchKeywords = uniqueTerms
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// CalculateAnimeMatchScore 计算动漫匹配分数
func (a *MetaAnime) CalculateAnimeMatchScore(query string) float64 {
	if query == "" {
		return 0.0
	}
	
	query = strings.ToLower(query)
	score := 0.0
	
	// 精确匹配标题
	if strings.ToLower(a.Title) == query {
		score += 1.0
	}
	
	// 部分匹配标题
	if strings.Contains(strings.ToLower(a.Title), query) {
		score += 0.8
	}
	
	// 匹配原始名称
	if strings.ToLower(a.OriginalName) == query {
		score += 0.9
	}
	if strings.Contains(strings.ToLower(a.OriginalName), query) {
		score += 0.7
	}
	
	// 匹配别名
	for _, alias := range a.Aliases {
		if strings.ToLower(alias) == query {
			score += 0.85
			break
		}
		if strings.Contains(strings.ToLower(alias), query) {
			score += 0.65
		}
	}
	
	// 匹配搜索词
	for _, term := range a.SearchTerms {
		if term == query {
			score += 0.7
			break
		}
		if strings.Contains(term, query) {
			score += 0.5
		}
	}
	
	// 匹配集数
	if a.CurrentEpisode > 0 {
		episodeStr := strconv.Itoa(a.CurrentEpisode)
		if strings.Contains(query, episodeStr) {
			score += 0.3
		}
	}
	
	// 匹配年份
	if a.Year > 0 {
		yearStr := strconv.Itoa(a.Year)
		if strings.Contains(query, yearStr) {
			score += 0.3
		}
	}
	
	// 特殊处理动漫标题中的数字（如第二季、第三季等）
	numberPattern := regexp.MustCompile(`第(\d+)季`)
	matches := numberPattern.FindStringSubmatch(query)
	if len(matches) >= 2 {
		if strings.Contains(strings.ToLower(a.Title), matches[0]) {
			score += 0.4
		}
	}
	
	// 限制分数范围在0-1之间
	if score > 1.0 {
		score = 1.0
	}
	
	a.MatchScore = score
	a.SearchScore = score // 与通用搜索分数保持一致
	return score
}

// FormatAnimeInfo 格式化动漫信息
func (a *MetaAnime) FormatAnimeInfo() string {
	info := []string{}
	
	// 添加标题
	if a.Title != "" {
		info = append(info, a.Title)
	}
	
	// 添加季度信息
	if a.SeasonInfo != "" {
		info = append(info, a.SeasonInfo)
	}
	
	// 添加类型
	if a.IsMovie {
		info = append(info, "剧场版")
	} else if a.IsOVA {
		info = append(info, "OVA")
	} else if a.IsSpecial {
		info = append(info, "特别篇")
	} else if a.IsWeb {
		info = append(info, "网络动画")
	} else {
		info = append(info, "TV动画")
	}
	
	// 添加集数信息
	if a.IsTV || a.IsWeb {
		if a.CurrentEpisode > 0 {
			if a.TotalEpisodes > 0 {
				info = append(info, fmt.Sprintf("第%d话/%d话", a.CurrentEpisode, a.TotalEpisodes))
			} else {
				info = append(info, fmt.Sprintf("第%d话", a.CurrentEpisode))
			}
		}
	}
	
	// 添加评分信息
	if a.Rating > 0 {
		info = append(info, fmt.Sprintf("%.1f分", a.Rating))
	}
	
	// 添加连载状态
	if a.IsOngoing {
		info = append(info, "连载中")
	} else if a.EndDate != "" {
		info = append(info, "已完结")
	}
	
	return strings.Join(info, " | ")
}

// GetAnimeStatusText 获取动漫状态文本
func (a *MetaAnime) GetAnimeStatusText() string {
	if a.IsOngoing {
		return "连载中"
	} else if a.EndDate != "" {
		return "已完结"
	} else if a.StartDate != "" {
		return "未开播"
	}
	return "未知"
}

// IsEpisodeMatch 检查是否匹配指定集数
func (a *MetaAnime) IsEpisodeMatch(episode int) bool {
	return a.CurrentEpisode == episode
}

// IsNewerEpisode 检查是否为更新的集数
func (a *MetaAnime) IsNewerEpisode(episode int) bool {
	return a.CurrentEpisode > episode
}

// GetNextEpisode 获取下一话
func (a *MetaAnime) GetNextEpisode() int {
	return a.CurrentEpisode + 1
}

// IsComplete 判断是否完结
func (a *MetaAnime) IsComplete() bool {
	return !a.IsOngoing && a.EndDate != ""
}

// GenerateAnimeFileName 生成标准的动漫文件名
func (a *MetaAnime) GenerateAnimeFileName() string {
	parts := []string{}
	
	// 标题
	if a.Title != "" {
		parts = append(parts, a.Title)
	} else if a.OriginalName != "" {
		parts = append(parts, a.OriginalName)
	}
	
	// 年份
	if a.Year > 0 {
		parts = append(parts, fmt.Sprintf("%d", a.Year))
	}
	
	// 集数
	if a.CurrentEpisode > 0 {
		parts = append(parts, fmt.Sprintf("第%02d话", a.CurrentEpisode))
	}
	
	// 类型标识
	if a.IsMovie {
		parts = append(parts, "剧场版")
	} else if a.IsOVA {
		parts = append(parts, "OVA")
	} else if a.IsSpecial {
		parts = append(parts, "特别篇")
	}
	
	// 视频质量
	quality := a.FormatVideoQuality()
	if quality != "" {
		parts = append(parts, quality)
	}
	
	return strings.Join(parts, " ")
}