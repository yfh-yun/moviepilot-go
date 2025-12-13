package media

import (
	"fmt"
	"strings"
)

// MediaSource 媒体来源枚举
type MediaSource string

const (
	MediaSourceTMDB    MediaSource = "themoviedb"
	MediaSourceDouban  MediaSource = "douban"
	MediaSourceBangumi MediaSource = "bangumi"
)

// MediaInfo 媒体上下文（TMDB/豆瓣/Bangumi 聚合元数据）
type MediaInfo struct {
	Source MediaSource `json:"source"`
	Type   MediaType   `json:"type"` // 对应 shared/schemas 或 internal/models 枚举

	Title   string `json:"title"`
	EnTitle string `json:"en_title"`
	HKTitle string `json:"hk_title"`
	TWTitle string `json:"tw_title"`
	SGTitle string `json:"sg_title"`

	Year   string `json:"year"`
	Season int    `json:"season"`

	TMDBID    int64  `json:"tmdb_id"`
	IMDBID    string `json:"imdb_id"`
	TVDBID    int64  `json:"tvdb_id"`
	DoubanID  string `json:"douban_id"`
	BangumiID int64  `json:"bangumi_id"`

	CollectionID int64 `json:"collection_id"`

	OriginalLanguage string `json:"original_language"`
	OriginalTitle    string `json:"original_title"`
	ReleaseDate      string `json:"release_date"` // 原始字符串保留，方便展示

	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`
	LogoPath     string `json:"logo_path"`

	VoteAverage float64 `json:"vote_average"`
	Overview    string  `json:"overview"`

	GenreIDs []int64  `json:"genre_ids"`
	Names    []string `json:"names"`

	Seasons     map[int][]int    `json:"seasons"`
	SeasonInfo  []map[string]any `json:"season_info"`
	SeasonYears map[int]string   `json:"season_years"`

	TMDBRaw    map[string]any `json:"tmdb_raw"`
	DoubanRaw  map[string]any `json:"douban_raw"`
	BangumiRaw map[string]any `json:"bangumi_raw"`

	Directors []map[string]any `json:"directors"`
	Actors    []map[string]any `json:"actors"`

	// 其他字段按需添加
	Adult          bool    `json:"adult"`
	EpisodeRunTime []int   `json:"episode_run_time"`
	Popularity     float64 `json:"popularity"`
	VoteCount      int     `json:"vote_count"`
	Runtime        int     `json:"runtime"`
}

// TitleYear 返回 "标题 (年份)"
func (m *MediaInfo) TitleYear() string {
	if m.Year != "" {
		return fmt.Sprintf("%s (%s)", m.Title, m.Year)
	}
	return m.Title
}

// DetailLink 返回 TMDB/豆瓣/Bangumi 的详情链接
func (m *MediaInfo) DetailLink() string {
	switch m.Source {
	case MediaSourceTMDB:
		if m.Type == MediaTypeMovie {
			return fmt.Sprintf("https://www.themoviedb.org/movie/%d", m.TMDBID)
		} else {
			return fmt.Sprintf("https://www.themoviedb.org/tv/%d", m.TMDBID)
		}
	case MediaSourceDouban:
		return fmt.Sprintf("https://movie.douban.com/subject/%s/", m.DoubanID)
	case MediaSourceBangumi:
		return fmt.Sprintf("https://bgm.tv/subject/%d", m.BangumiID)
	default:
		return ""
	}
}

// Stars 评分星星渲染
func (m *MediaInfo) Stars() string {
	stars := int(m.VoteAverage / 2)
	return strings.Repeat("★", stars) + strings.Repeat("☆", 5-stars)
}

// VoteStar 评分星星渲染
func (m *MediaInfo) VoteStar() string {
	return m.Stars()
}

// BackdropImage 返回背景图片 URL
func (m *MediaInfo) BackdropImage(defaultURL string) string {
	if m.BackdropPath != "" {
		return fmt.Sprintf("https://image.tmdb.org/t/p/original%s", m.BackdropPath)
	}
	return defaultURL
}

// PosterImage 返回海报图片 URL
func (m *MediaInfo) PosterImage(defaultURL string) string {
	if m.PosterPath != "" {
		return fmt.Sprintf("https://image.tmdb.org/t/p/original%s", m.PosterPath)
	}
	return defaultURL
}

// MessageImage 返回消息通知使用的图片 URL
func (m *MediaInfo) MessageImage(defaultURL string) string {
	// 优先使用海报，没有则使用背景图
	if m.PosterPath != "" {
		return m.PosterImage(defaultURL)
	}
	return m.BackdropImage(defaultURL)
}

// OverviewString 裁剪简介文本
func (m *MediaInfo) OverviewString(maxLen int) string {
	if len(m.Overview) <= maxLen {
		return m.Overview
	}
	return m.Overview[:maxLen-3] + "..."
}

// ClearHeavyFields 清理大字段，减小体积（适合缓存/消息传输）
func (m *MediaInfo) ClearHeavyFields() {
	m.TMDBRaw = nil
	m.DoubanRaw = nil
	m.BangumiRaw = nil
	m.SeasonInfo = nil
	m.Directors = nil
	m.Actors = nil
}

// ToDict 将MediaInfo转换为字典，与Python版本to_dict功能一致
func (m *MediaInfo) ToDict() map[string]interface{} {
	result := make(map[string]interface{})
	
	// 基础字段
	result["source"] = string(m.Source)
	result["type"] = string(m.Type)
	result["title"] = m.Title
	result["en_title"] = m.EnTitle
	result["hk_title"] = m.HKTitle
	result["tw_title"] = m.TWTitle
	result["sg_title"] = m.SGTitle
	result["year"] = m.Year
	result["season"] = m.Season
	result["tmdb_id"] = m.TMDBID
	result["imdb_id"] = m.IMDBID
	result["tvdb_id"] = m.TVDBID
	result["douban_id"] = m.DoubanID
	result["bangumi_id"] = m.BangumiID
	result["collection_id"] = m.CollectionID
	result["original_language"] = m.OriginalLanguage
	result["original_title"] = m.OriginalTitle
	result["release_date"] = m.ReleaseDate
	result["backdrop_path"] = m.BackdropPath
	result["poster_path"] = m.PosterPath
	result["logo_path"] = m.LogoPath
	result["vote_average"] = m.VoteAverage
	result["overview"] = m.Overview
	result["genre_ids"] = m.GenreIDs
	result["names"] = m.Names
	result["seasons"] = m.Seasons
	result["season_years"] = m.SeasonYears
	result["adult"] = m.Adult
	result["episode_run_time"] = m.EpisodeRunTime
	result["popularity"] = m.Popularity
	result["vote_count"] = m.VoteCount
	result["runtime"] = m.Runtime
	
	// 计算属性
	result["detail_link"] = m.DetailLink()
	result["title_year"] = m.TitleYear()
	result["stars"] = m.Stars()
	result["vote_star"] = m.VoteStar()
	
	// 移除原始数据字段，减小体积
	result["tmdb_info"] = nil
	result["douban_info"] = nil
	result["bangumi_info"] = nil
	
	return result
}
