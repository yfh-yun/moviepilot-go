package meta

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MetaVideo 视频元数据类
type MetaVideo struct {
	MetaBase
	
	// 视频特有信息
	IMDBID        string `json:"imdb_id"`         // IMDB ID
	TMDBID        int    `json:"tmdb_id"`         // TMDB ID
	TVDBID        int    `json:"tvdb_id"`         // TVDB ID
	DoubanID      string `json:"douban_id"`       // 豆瓣 ID
	
	// 电影特有信息
	Directors     []string `json:"directors"`     // 导演
	Writers       []string `json:"writers"`       // 编剧
	Actors        []string `json:"actors"`        // 演员
	Genres        []string `json:"genres"`        // 类型
	Countries     []string `json:"countries"`     // 国家/地区
	Runtime       int      `json:"runtime"`       // 时长（分钟）
	Rating        float64  `json:"rating"`        // 评分
	ReleaseDate   string   `json:"release_date"`  // 上映日期
	
	// 剧集特有信息
	SeasonCount   int      `json:"season_count"`  // 总季数
	EpisodeCount  int      `json:"episode_count"` // 总集数
	Season        int      `json:"season"`        // 当前季数
	Episode       int      `json:"episode"`       // 当前集数
	TotalEpisodes []int    `json:"total_episodes"` // 每季度集数
	
	// 高级信息
	Subtitles     []string `json:"subtitles"`     // 字幕列表
	Is4K          bool     `json:"is_4k"`         // 是否4K
	IsDolbyVision bool     `json:"is_dolby_vision"` // 是否杜比视界
	IsHDR         bool     `json:"is_hdr"`        // 是否HDR
	IsDolbyAtmos  bool     `json:"is_dolby_atmos"` // 是否杜比全景声
	
	// 额外标识
	IsMovie       bool     `json:"is_movie"`      // 是否电影
	IsSeries      bool     `json:"is_series"`     // 是否电视剧
	IsAnime       bool     `json:"is_anime"`      // 是否动漫
	IsAnimation   bool     `json:"is_animation"`  // 是否动画
	IsDocumentary bool     `json:"is_documentary"` // 是否纪录片
	IsShort       bool     `json:"is_short"`      // 是否短片
	
	// 搜索相关
	SearchKeywords []string `json:"search_keywords"` // 搜索关键词
	SearchScore    float64  `json:"search_score"`    // 搜索得分
}

// NewMetaVideo 创建视频元数据实例
func NewMetaVideo(name string) *MetaVideo {
	return &MetaVideo{
		MetaBase:      *NewMetaBase(name).(*MetaBase),
		MediaType:     MediaTypeVideo,
		Directors:     make([]string, 0),
		Writers:       make([]string, 0),
		Actors:        make([]string, 0),
		Genres:        make([]string, 0),
		Countries:     make([]string, 0),
		Subtitles:     make([]string, 0),
		TotalEpisodes: make([]int, 0),
		SearchKeywords: make([]string, 0),
	}
}

// NewMovieMeta 创建电影元数据实例
func NewMovieMeta(name string) *MetaVideo {
	video := NewMetaVideo(name)
	video.IsMovie = true
	video.MediaType = MediaTypeMovie
	return video
}

// NewSeriesMeta 创建剧集元数据实例
func NewSeriesMeta(name string) *MetaVideo {
	video := NewMetaVideo(name)
	video.IsSeries = true
	video.MediaType = MediaTypeSeries
	return video
}

// 基础方法

// Clone 克隆视频元数据
func (v *MetaVideo) Clone() MetaInfo {
	clone := *v
	// 深拷贝所有切片
	clone.Tags = make([]string, len(v.Tags))
	copy(clone.Tags, v.Tags)
	clone.Directors = make([]string, len(v.Directors))
	copy(clone.Directors, v.Directors)
	clone.Writers = make([]string, len(v.Writers))
	copy(clone.Writers, v.Writers)
	clone.Actors = make([]string, len(v.Actors))
	copy(clone.Actors, v.Actors)
	clone.Genres = make([]string, len(v.Genres))
	copy(clone.Genres, v.Genres)
	clone.Countries = make([]string, len(v.Countries))
	copy(clone.Countries, v.Countries)
	clone.Subtitles = make([]string, len(v.Subtitles))
	copy(clone.Subtitles, v.Subtitles)
	clone.TotalEpisodes = make([]int, len(v.TotalEpisodes))
	copy(clone.TotalEpisodes, v.TotalEpisodes)
	clone.SearchKeywords = make([]string, len(v.SearchKeywords))
	copy(clone.SearchKeywords, v.SearchKeywords)
	return &clone
}

// IsValid 判断视频元数据是否有效
func (v *MetaVideo) IsValid() bool {
	return (v.ParseStatus == ParseStatusSuccess || v.ParseStatus == ParseStatusPartially) && 
	       (v.TMDBID > 0 || v.IMDBID != "" || v.DoubanID != "" || v.Title != "")
}

// ToString 转换为字符串表示
func (v *MetaVideo) ToString() string {
	if v.IsSeries {
		if v.Season > 0 && v.Episode > 0 {
			return fmt.Sprintf("%s S%02dE%02d (%d) [%s]", v.Title, v.Season, v.Episode, v.Year, v.MediaType)
		} else if v.Season > 0 {
			return fmt.Sprintf("%s S%02d (%d) [%s]", v.Title, v.Season, v.Year, v.MediaType)
		}
	}
	return fmt.Sprintf("%s (%d) [%s]", v.Title, v.Year, v.MediaType)
}

// 标识相关方法

// IsMovie 判断是否为电影
func (v *MetaVideo) IsMovieType() bool {
	return v.IsMovie
}

// IsSeries 判断是否为剧集
func (v *MetaVideo) IsSeriesType() bool {
	return v.IsSeries
}

// IsAnime 判断是否为动漫
func (v *MetaVideo) IsAnimeType() bool {
	return v.IsAnime
}

// IsAnimation 判断是否为动画
func (v *MetaVideo) IsAnimationType() bool {
	return v.IsAnimation
}

// IsDocumentary 判断是否为纪录片
func (v *MetaVideo) IsDocumentaryType() bool {
	return v.IsDocumentary
}

// IsShort 判断是否为短片
func (v *MetaVideo) IsShortType() bool {
	return v.IsShort
}

// 设置标识方法
func (v *MetaVideo) SetMovie() {
	v.IsMovie = true
	v.IsSeries = false
	v.IsAnime = false
	v.MediaType = MediaTypeMovie
	v.UpdatedAt = time.Now()
}

func (v *MetaVideo) SetSeries() {
	v.IsMovie = false
	v.IsSeries = true
	v.IsAnime = false
	v.MediaType = MediaTypeSeries
	v.UpdatedAt = time.Now()
}

func (v *MetaVideo) SetAnime() {
	v.IsMovie = false
	v.IsSeries = false
	v.IsAnime = true
	v.IsAnimation = true
	v.MediaType = MediaTypeAnime
	v.UpdatedAt = time.Now()
}

func (v *MetaVideo) SetAnimation() {
	v.IsAnimation = true
	v.UpdatedAt = time.Now()
}

func (v *MetaVideo) SetDocumentary() {
	v.IsDocumentary = true
	v.UpdatedAt = time.Now()
}

func (v *MetaVideo) SetShort() {
	v.IsShort = true
	v.UpdatedAt = time.Now()
}

// ID相关方法

// GetIMDBID 获取IMDB ID
func (v *MetaVideo) GetIMDBID() string {
	return v.IMDBID
}

// SetIMDBID 设置IMDB ID
func (v *MetaVideo) SetIMDBID(id string) {
	v.IMDBID = id
	v.UpdatedAt = time.Now()
}

// GetTMDBID 获取TMDB ID
func (v *MetaVideo) GetTMDBID() int {
	return v.TMDBID
}

// SetTMDBID 设置TMDB ID
func (v *MetaVideo) SetTMDBID(id int) {
	v.TMDBID = id
	v.UpdatedAt = time.Now()
}

// GetTVDBID 获取TVDB ID
func (v *MetaVideo) GetTVDBID() int {
	return v.TVDBID
}

// SetTVDBID 设置TVDB ID
func (v *MetaVideo) SetTVDBID(id int) {
	v.TVDBID = id
	v.UpdatedAt = time.Now()
}

// GetDoubanID 获取豆瓣 ID
func (v *MetaVideo) GetDoubanID() string {
	return v.DoubanID
}

// SetDoubanID 设置豆瓣 ID
func (v *MetaVideo) SetDoubanID(id string) {
	v.DoubanID = id
	v.UpdatedAt = time.Now()
}

// 电影相关方法

// GetDirectors 获取导演列表
func (v *MetaVideo) GetDirectors() []string {
	return v.Directors
}

// AddDirector 添加导演
func (v *MetaVideo) AddDirector(director string) {
	for _, d := range v.Directors {
		if d == director {
			return // 避免重复
		}
	}
	v.Directors = append(v.Directors, director)
	v.UpdatedAt = time.Now()
}

// GetWriters 获取编剧列表
func (v *MetaVideo) GetWriters() []string {
	return v.Writers
}

// AddWriter 添加编剧
func (v *MetaVideo) AddWriter(writer string) {
	for _, w := range v.Writers {
		if w == writer {
			return // 避免重复
		}
	}
	v.Writers = append(v.Writers, writer)
	v.UpdatedAt = time.Now()
}

// GetActors 获取演员列表
func (v *MetaVideo) GetActors() []string {
	return v.Actors
}

// AddActor 添加演员
func (v *MetaVideo) AddActor(actor string) {
	for _, a := range v.Actors {
		if a == actor {
			return // 避免重复
		}
	}
	v.Actors = append(v.Actors, actor)
	v.UpdatedAt = time.Now()
}

// GetGenres 获取类型列表
func (v *MetaVideo) GetGenres() []string {
	return v.Genres
}

// AddGenre 添加类型
func (v *MetaVideo) AddGenre(genre string) {
	for _, g := range v.Genres {
		if g == genre {
			return // 避免重复
		}
	}
	v.Genres = append(v.Genres, genre)
	v.UpdatedAt = time.Now()
}

// GetCountries 获取国家/地区列表
func (v *MetaVideo) GetCountries() []string {
	return v.Countries
}

// AddCountry 添加国家/地区
func (v *MetaVideo) AddCountry(country string) {
	for _, c := range v.Countries {
		if c == country {
			return // 避免重复
		}
	}
	v.Countries = append(v.Countries, country)
	v.UpdatedAt = time.Now()
}

// GetRuntime 获取时长
func (v *MetaVideo) GetRuntime() int {
	return v.Runtime
}

// SetRuntime 设置时长
func (v *MetaVideo) SetRuntime(runtime int) {
	v.Runtime = runtime
	v.UpdatedAt = time.Now()
}

// GetRating 获取评分
func (v *MetaVideo) GetRating() float64 {
	return v.Rating
}

// SetRating 设置评分
func (v *MetaVideo) SetRating(rating float64) {
	v.Rating = rating
	v.UpdatedAt = time.Now()
}

// GetReleaseDate 获取上映日期
func (v *MetaVideo) GetReleaseDate() string {
	return v.ReleaseDate
}

// SetReleaseDate 设置上映日期
func (v *MetaVideo) SetReleaseDate(date string) {
	v.ReleaseDate = date
	v.UpdatedAt = time.Now()
}

// 剧集相关方法

// GetSeasonCount 获取总季数
func (v *MetaVideo) GetSeasonCount() int {
	return v.SeasonCount
}

// SetSeasonCount 设置总季数
func (v *MetaVideo) SetSeasonCount(count int) {
	v.SeasonCount = count
	v.UpdatedAt = time.Now()
}

// GetEpisodeCount 获取总集数
func (v *MetaVideo) GetEpisodeCount() int {
	return v.EpisodeCount
}

// SetEpisodeCount 设置总集数
func (v *MetaVideo) SetEpisodeCount(count int) {
	v.EpisodeCount = count
	v.UpdatedAt = time.Now()
}

// GetSeason 获取当前季数
func (v *MetaVideo) GetSeason() int {
	return v.Season
}

// SetSeason 设置当前季数
func (v *MetaVideo) SetSeason(season int) {
	v.Season = season
	v.UpdatedAt = time.Now()
}

// GetEpisode 获取当前集数
func (v *MetaVideo) GetEpisode() int {
	return v.Episode
}

// SetEpisode 设置当前集数
func (v *MetaVideo) SetEpisode(episode int) {
	v.Episode = episode
	v.UpdatedAt = time.Now()
}

// GetTotalEpisodes 获取每季度集数
func (v *MetaVideo) GetTotalEpisodes() []int {
	return v.TotalEpisodes
}

// SetTotalEpisodes 设置每季度集数
func (v *MetaVideo) SetTotalEpisodes(episodes []int) {
	v.TotalEpisodes = episodes
	v.UpdatedAt = time.Now()
}

// GetSeasonEpisodes 获取指定季的集数
func (v *MetaVideo) GetSeasonEpisodes(season int) int {
	if season > 0 && season <= len(v.TotalEpisodes) {
		return v.TotalEpisodes[season-1]
	}
	return 0
}

// SetSeasonEpisodes 设置指定季的集数
func (v *MetaVideo) SetSeasonEpisodes(season, count int) {
	if season > 0 {
		// 扩展切片以容纳新的季数
		if season > len(v.TotalEpisodes) {
			newEpisodes := make([]int, season)
			copy(newEpisodes, v.TotalEpisodes)
			v.TotalEpisodes = newEpisodes
		}
		v.TotalEpisodes[season-1] = count
		v.UpdatedAt = time.Now()
	}
}

// 高级信息方法

// GetSubtitles 获取字幕列表
func (v *MetaVideo) GetSubtitles() []string {
	return v.Subtitles
}

// AddSubtitle 添加字幕
func (v *MetaVideo) AddSubtitle(subtitle string) {
	for _, s := range v.Subtitles {
		if s == subtitle {
			return // 避免重复
		}
	}
	v.Subtitles = append(v.Subtitles, subtitle)
	v.UpdatedAt = time.Now()
}

// Is4KVideo 判断是否为4K视频
func (v *MetaVideo) Is4KVideo() bool {
	return v.Is4K || v.Resolution == Resolution4K || v.Resolution == Resolution8K
}

// Set4K 设置4K标识
func (v *MetaVideo) Set4K(is4K bool) {
	v.Is4K = is4K
	v.UpdatedAt = time.Now()
}

// IsDolbyVisionVideo 判断是否为杜比视界视频
func (v *MetaVideo) IsDolbyVisionVideo() bool {
	return v.IsDolbyVision
}

// SetDolbyVision 设置杜比视界标识
func (v *MetaVideo) SetDolbyVision(isDV bool) {
	v.IsDolbyVision = isDV
	v.UpdatedAt = time.Now()
}

// IsHDRVideo 判断是否为HDR视频
func (v *MetaVideo) IsHDRVideo() bool {
	return v.IsHDR
}

// SetHDR 设置HDR标识
func (v *MetaVideo) SetHDR(isHDR bool) {
	v.IsHDR = isHDR
	v.UpdatedAt = time.Now()
}

// IsDolbyAtmosAudio 判断是否为杜比全景声音频
func (v *MetaVideo) IsDolbyAtmosAudio() bool {
	return v.IsDolbyAtmos || strings.Contains(strings.ToLower(v.AudioCodec), "atmos")
}

// SetDolbyAtmos 设置杜比全景声标识
func (v *MetaVideo) SetDolbyAtmos(isAtmos bool) {
	v.IsDolbyAtmos = isAtmos
	v.UpdatedAt = time.Now()
}

// 搜索相关方法

// GetSearchKeywords 获取搜索关键词
func (v *MetaVideo) GetSearchKeywords() []string {
	return v.SearchKeywords
}

// AddSearchKeyword 添加搜索关键词
func (v *MetaVideo) AddSearchKeyword(keyword string) {
	for _, k := range v.SearchKeywords {
		if k == keyword {
			return // 避免重复
		}
	}
	v.SearchKeywords = append(v.SearchKeywords, keyword)
	v.UpdatedAt = time.Now()
}

// GetSearchScore 获取搜索得分
func (v *MetaVideo) GetSearchScore() float64 {
	return v.SearchScore
}

// SetSearchScore 设置搜索得分
func (v *MetaVideo) SetSearchScore(score float64) {
	v.SearchScore = score
	v.UpdatedAt = time.Now()
}

// 解析相关方法

// ParseName 解析视频名称
func (v *MetaVideo) ParseName() {
	name := v.Name
	v.ParseStatus = ParseStatusFailed
	v.Confidence = 0.0
	
	// 清理名称
	cleanName := v.CleanupName()
	if cleanName == "" {
		return
	}
	
	// 解析季数和集数（剧集）
	season, episode := v.ParseSeasonEpisode(name)
	if season > 0 && episode > 0 {
		v.IsSeries = true
		v.Season = season
		v.Episode = episode
		v.MediaType = MediaTypeSeries
		v.ParseStatus = ParseStatusSuccess
		v.Confidence = 0.8
	} else if season > 0 {
		v.IsSeries = true
		v.Season = season
		v.MediaType = MediaTypeSeries
		v.ParseStatus = ParseStatusPartially
		v.Confidence = 0.7
	}
	
	// 解析年份
	year := v.ParseYearFromName()
	if year > 0 {
		v.Year = year
		// 增加置信度
		if v.Confidence > 0 {
			v.Confidence += 0.1
		} else {
			v.Confidence = 0.5
		}
	}
	
	// 尝试识别动漫
	if v.IsAnimeType(name) {
		v.IsAnime = true
		v.IsAnimation = true
		v.MediaType = MediaTypeAnime
		v.ParseStatus = ParseStatusSuccess
		v.Confidence = 0.9
	}
	
	// 设置标题
	if v.Title == "" {
		// 清理后的名称作为标题
		title := cleanName
		
		// 如果是剧集，移除季数和集数部分
		if v.IsSeries {
			title = v.RemoveSeasonEpisodeFromTitle(title)
		}
		
		// 如果有年份，移除年份部分
		if v.Year > 0 {
			title = v.RemoveYearFromTitle(title)
		}
		
		v.Title = strings.TrimSpace(title)
	}
	
	// 如果标题已设置，设置解析成功
	if v.Title != "" && v.ParseStatus == ParseStatusFailed {
		v.ParseStatus = ParseStatusSuccess
		v.Confidence = 0.6
	}
	
	// 设置高级标识
	v.ParseVideoFeatures(name)
	
	// 设置默认值
	v.SetDefaultValues()
}

// ParseSeasonEpisode 从名称中解析季数和集数
func (v *MetaVideo) ParseSeasonEpisode(name string) (season, episode int) {
	// 转换为小写以进行匹配
	name = strings.ToLower(name)
	
	// 常见的季集格式正则表达式
	patterns := []struct {
		Regex *regexp.Regexp
		SeasonIndex, EpisodeIndex int
	}{{
		// S01E01 格式
		Regex: regexp.MustCompile(`s(\d{1,3})e(\d{1,3})`),
		SeasonIndex: 1,
		EpisodeIndex: 2,
	}, {
		// 第1季第1集 格式
		Regex: regexp.MustCompile(`第(\d{1,3})季第(\d{1,3})集`),
		SeasonIndex: 1,
		EpisodeIndex: 2,
	}, {
		// 第1季01集 格式
		Regex: regexp.MustCompile(`第(\d{1,3})季(\d{1,3})集`),
		SeasonIndex: 1,
		EpisodeIndex: 2,
	}, {
		// S01E01E02 多集格式
		Regex: regexp.MustCompile(`s(\d{1,3})e(\d{1,3})`),
		SeasonIndex: 1,
		EpisodeIndex: 2,
	}, {
		// Season 1 Episode 1 格式
		Regex: regexp.MustCompile(`season\s+(\d{1,3})\s+episode\s+(\d{1,3})`),
		SeasonIndex: 1,
		EpisodeIndex: 2,
	}, {
		// 1x01 格式
		Regex: regexp.MustCompile(`(\d{1,3})x(\d{1,3})`),
		SeasonIndex: 1,
		EpisodeIndex: 2,
	}}
	
	for _, pattern := range patterns {
		matches := pattern.Regex.FindStringSubmatch(name)
		if len(matches) >= 3 {
			s, err1 := strconv.Atoi(matches[pattern.SeasonIndex])
			e, err2 := strconv.Atoi(matches[pattern.EpisodeIndex])
			if err1 == nil && err2 == nil && s > 0 && e > 0 {
				return s, e
			}
		}
	}
	
	// 特殊处理单季数
	seasonPattern := regexp.MustCompile(`s(\d{1,3})`)
	matches := seasonPattern.FindStringSubmatch(name)
	if len(matches) >= 2 {
		s, err := strconv.Atoi(matches[1])
		if err == nil && s > 0 {
			return s, 0
		}
	}
	
	// 特殊处理单集数（常见于动漫）
	episodePattern := regexp.MustCompile(`第(\d{1,3})集`)
	matches = episodePattern.FindStringSubmatch(name)
	if len(matches) >= 2 {
		e, err := strconv.Atoi(matches[1])
		if err == nil && e > 0 {
			return 1, e // 默认为第一季
		}
	}
	
	return 0, 0
}

// RemoveSeasonEpisodeFromTitle 从标题中移除季数和集数部分
func (v *MetaVideo) RemoveSeasonEpisodeFromTitle(title string) string {
	// 移除常见的季集标识
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`s\d{1,3}e\d{1,3}`),              // S01E01
		regexp.MustCompile(`s\d{1,3}\s*`),                   // S01
		regexp.MustCompile(`第\d{1,3}季第\d{1,3}集`),         // 第1季第1集
		regexp.MustCompile(`第\d{1,3}季\d{1,3}集`),           // 第1季01集
		regexp.MustCompile(`第\d{1,3}季`),                     // 第1季
		regexp.MustCompile(`第\d{1,3}集`),                     // 第1集
		regexp.MustCompile(`\d{1,3}x\d{1,3}`),                // 1x01
		regexp.MustCompile(`season\s+\d{1,3}\s+episode\s+\d{1,3}`), // Season 1 Episode 1
	}
	
	for _, pattern := range patterns {
		title = strings.ToLower(title)
		title = pattern.ReplaceAllString(title, "")
	}
	
	return strings.TrimSpace(title)
}

// RemoveYearFromTitle 从标题中移除年份部分
func (v *MetaVideo) RemoveYearFromTitle(title string) string {
	if v.Year > 0 {
		yearRegex := regexp.MustCompile(fmt.Sprintf(`%d`, v.Year))
		title = yearRegex.ReplaceAllString(title, "")
	}
	return strings.TrimSpace(title)
}

// IsAnimeType 判断是否为动漫类型
func (v *MetaVideo) IsAnimeType(name string) bool {
	name = strings.ToLower(name)
	
	// 常见的动漫关键词
	animeKeywords := []string{
		"anime", "动漫", "アニメ", "动画", "cartoon",
		"番剧", "番組", "ova", "oad", "剧场版", "movie",
	}
	
	// 检查关键词
	for _, keyword := range animeKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}
	
	// 检查是否有特殊的编号格式（常见于动漫）
	animePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\[(\d{2,3})\]`),   // [01] 格式
		regexp.MustCompile(`第(\d{2,3})话`),    // 第01话 格式
		regexp.MustCompile(`第(\d{2,3})集`),    // 第01集 格式（如果前面没有季数）
	}
	
	// 如果没有检测到季数但有集数，可能是动漫
	season, episode := v.ParseSeasonEpisode(name)
	if season == 0 && episode > 0 {
		return true
	}
	
	// 检查动漫特有格式
	for _, pattern := range animePatterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	
	return false
}

// ParseVideoFeatures 解析视频高级特性
func (v *MetaVideo) ParseVideoFeatures(name string) {
	name = strings.ToLower(name)
	
	// 检查4K
	if strings.Contains(name, "4k") || strings.Contains(name, "2160p") || strings.Contains(name, "uhd") {
		v.Is4K = true
	}
	
	// 检查8K
	if strings.Contains(name, "8k") || strings.Contains(name, "4320p") {
		v.Is4K = true // 8K也视为4K的一种
	}
	
	// 检查杜比视界
	if strings.Contains(name, "dolby vision") || strings.Contains(name, "dv") {
		v.IsDolbyVision = true
		v.IsHDR = true // 杜比视界包含HDR
	}
	
	// 检查HDR
	if strings.Contains(name, "hdr") || strings.Contains(name, "hdr10") || strings.Contains(name, "hdr10+") {
		v.IsHDR = true
	}
	
	// 检查杜比全景声
	if strings.Contains(name, "dolby atmos") || strings.Contains(name, "atmos") {
		v.IsDolbyAtmos = true
	}
	
	// 检查是否为动画
	animationKeywords := []string{"animation", "animated", "cartoon", "动画", "卡通"}
	for _, keyword := range animationKeywords {
		if strings.Contains(name, keyword) {
			v.IsAnimation = true
			break
		}
	}
	
	// 检查是否为纪录片
	documentaryKeywords := []string{"documentary", "纪录片", "纪实", "docu"}
	for _, keyword := range documentaryKeywords {
		if strings.Contains(name, keyword) {
			v.IsDocumentary = true
			break
		}
	}
	
	// 检查是否为短片
	shortKeywords := []string{"short", "短片", "short film"}
	for _, keyword := range shortKeywords {
		if strings.Contains(name, keyword) {
			v.IsShort = true
			break
		}
	}
}

// FormatVideoQuality 格式化视频质量信息
func (v *MetaVideo) FormatVideoQuality() string {
	quality := []string{}
	
	// 分辨率
	if v.Is4KVideo() {
		quality = append(quality, "4K")
		if strings.Contains(v.Resolution, "8K") {
			quality = append(quality, "8K")
		}
	} else if v.Resolution != ResolutionUnknown {
		quality = append(quality, v.Resolution)
	}
	
	// HDR相关
	if v.IsDolbyVisionVideo() {
		quality = append(quality, "Dolby Vision")
	} else if v.IsHDRVideo() {
		quality = append(quality, "HDR")
	}
	
	// 音频相关
	if v.IsDolbyAtmosAudio() {
		quality = append(quality, "Dolby Atmos")
	}
	
	// 视频编码
	if v.VideoCodec != VideoCodecUnknown {
		quality = append(quality, v.VideoCodec)
	}
	
	return strings.Join(quality, " ")
}

// GenerateSearchKeywords 生成搜索关键词
func (v *MetaVideo) GenerateSearchKeywords() {
	keywords := make([]string, 0)
	
	// 添加标题关键词
	if v.Title != "" {
		// 清理标题，移除特殊字符
		cleanTitle := regexp.MustCompile(`[^\w\s]`).ReplaceAllString(v.Title, "")
		// 分词并添加
		words := strings.Fields(strings.ToLower(cleanTitle))
		for _, word := range words {
			if len(word) > 1 { // 忽略单字符
				keywords = append(keywords, word)
			}
		}
	}
	
	// 添加原始标题关键词
	if v.OriginalTitle != "" && v.OriginalTitle != v.Title {
		cleanTitle := regexp.MustCompile(`[^\w\s]`).ReplaceAllString(v.OriginalTitle, "")
		words := strings.Fields(strings.ToLower(cleanTitle))
		for _, word := range words {
			if len(word) > 1 {
				keywords = append(keywords, word)
			}
		}
	}
	
	// 添加导演、演员等作为关键词
	for _, person := range append(append(v.Directors, v.Writers...), v.Actors...) {
		if person != "" {
			cleanName := regexp.MustCompile(`[^\w\s]`).ReplaceAllString(person, "")
			words := strings.Fields(strings.ToLower(cleanName))
			for _, word := range words {
				if len(word) > 1 {
					keywords = append(keywords, word)
				}
			}
		}
	}
	
	// 添加类型作为关键词
	for _, genre := range v.Genres {
		if genre != "" {
			keywords = append(keywords, strings.ToLower(genre))
		}
	}
	
	// 去重
	uniqueKeywords := make([]string, 0)
	keywordMap := make(map[string]bool)
	for _, keyword := range keywords {
		if !keywordMap[keyword] {
			keywordMap[keyword] = true
			uniqueKeywords = append(uniqueKeywords, keyword)
		}
	}
	
	v.SearchKeywords = uniqueKeywords
}

// CalculateMatchScore 计算匹配分数
func (v *MetaVideo) CalculateMatchScore(query string) float64 {
	if query == "" {
		return 0.0
	}
	
	query = strings.ToLower(query)
	score := 0.0
	
	// 精确匹配标题
	if strings.ToLower(v.Title) == query {
		score += 1.0
	}
	
	// 部分匹配标题
	if strings.Contains(strings.ToLower(v.Title), query) {
		score += 0.8
	}
	
	// 匹配原始标题
	if strings.ToLower(v.OriginalTitle) == query {
		score += 0.9
	}
	if strings.Contains(strings.ToLower(v.OriginalTitle), query) {
		score += 0.7
	}
	
	// 匹配搜索关键词
	for _, keyword := range v.SearchKeywords {
		if keyword == query {
			score += 0.6
			break
		}
		if strings.Contains(keyword, query) {
			score += 0.4
		}
	}
	
	// 匹配年份（如果查询中包含年份）
	yearMatch := regexp.MustCompile(`\b(19|20)\d{2}\b`).FindString(query)
	if yearMatch != "" && strconv.Itoa(v.Year) == yearMatch {
		score += 0.3
	}
	
	// 限制分数范围在0-1之间
	if score > 1.0 {
		score = 1.0
	}
	
	v.SearchScore = score
	return score
}