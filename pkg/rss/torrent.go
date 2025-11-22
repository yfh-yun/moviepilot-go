package rss

import (
	"regexp"
	"strconv"
	"strings"
)

// TorrentInfo Torrent 信息
type TorrentInfo struct {
	// 原始信息
	Title       string
	Size        int64
	Seeders     int
	Leechers    int
	DownloadURL string
	MagnetLink  string
	InfoHash    string

	// 解析出的媒体信息
	MediaTitle string
	Season     int
	Episode    int
	Quality    string
	Resolution string
	Source     string
	Codec      string
	Audio      string
	Group      string
	Year       int
}

var (
	// 季集正则
	seasonEpisodePattern = regexp.MustCompile(`[Ss](\d{1,2})[Ee](\d{1,3})`)
	// 年份正则
	yearPattern = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	// 分辨率正则
	resolutionPattern = regexp.MustCompile(`(?i)(2160p|1080p|720p|480p|4K|UHD)`)
	// 来源正则
	sourcePattern = regexp.MustCompile(`(?i)(BluRay|Blu-ray|WEB-DL|WEBDL|WEBRip|HDTV|DVDRip|BDRip)`)
	// 编码正则
	codecPattern = regexp.MustCompile(`(?i)(x264|x265|H\.264|H\.265|HEVC|AVC)`)
	// 音频正则
	audioPattern = regexp.MustCompile(`(?i)(DTS|AC3|AAC|FLAC|TrueHD|Atmos)`)
	// 发布组正则
	groupPattern = regexp.MustCompile(`-([A-Za-z0-9]+)$`)
)

// ExtractTorrentInfo 从 RSS Item 提取 Torrent 信息
func ExtractTorrentInfo(item RSSItem) (*TorrentInfo, error) {
	info := &TorrentInfo{
		Title:       item.Title,
		DownloadURL: item.Link,
		Size:        item.Enclosure.Length,
	}

	// 如果有 enclosure，使用其 URL
	if item.Enclosure.URL != "" {
		info.DownloadURL = item.Enclosure.URL
	}

	// 解析标题
	parseTitle(info)

	return info, nil
}

// parseTitle 解析标题提取媒体信息
func parseTitle(info *TorrentInfo) {
	title := info.Title

	// 提取季集信息
	if matches := seasonEpisodePattern.FindStringSubmatch(title); len(matches) > 2 {
		info.Season, _ = strconv.Atoi(matches[1])
		info.Episode, _ = strconv.Atoi(matches[2])
	}

	// 提取年份
	if matches := yearPattern.FindStringSubmatch(title); len(matches) > 1 {
		info.Year, _ = strconv.Atoi(matches[1])
	}

	// 提取分辨率
	if matches := resolutionPattern.FindStringSubmatch(title); len(matches) > 1 {
		info.Resolution = strings.ToLower(matches[1])
		if info.Resolution == "4k" || info.Resolution == "uhd" {
			info.Resolution = "2160p"
		}
		info.Quality = info.Resolution
	}

	// 提取来源
	if matches := sourcePattern.FindStringSubmatch(title); len(matches) > 1 {
		info.Source = matches[1]
		// 标准化来源名称
		info.Source = strings.ReplaceAll(info.Source, "Blu-ray", "BluRay")
		info.Source = strings.ReplaceAll(info.Source, "WEBDL", "WEB-DL")
	}

	// 提取编码
	if matches := codecPattern.FindStringSubmatch(title); len(matches) > 1 {
		info.Codec = matches[1]
		// 标准化编码名称
		if strings.Contains(strings.ToLower(info.Codec), "265") ||
			strings.Contains(strings.ToUpper(info.Codec), "HEVC") {
			info.Codec = "x265"
		} else if strings.Contains(strings.ToLower(info.Codec), "264") ||
			strings.Contains(strings.ToUpper(info.Codec), "AVC") {
			info.Codec = "x264"
		}
	}

	// 提取音频
	if matches := audioPattern.FindStringSubmatch(title); len(matches) > 1 {
		info.Audio = matches[1]
	}

	// 提取发布组
	if matches := groupPattern.FindStringSubmatch(title); len(matches) > 1 {
		info.Group = matches[1]
	}

	// 提取媒体标题（移除所有识别出的信息）
	mediaTitle := title

	// 移除季集信息
	mediaTitle = seasonEpisodePattern.ReplaceAllString(mediaTitle, "")
	// 移除年份
	mediaTitle = yearPattern.ReplaceAllString(mediaTitle, "")
	// 移除分辨率
	mediaTitle = resolutionPattern.ReplaceAllString(mediaTitle, "")
	// 移除来源
	mediaTitle = sourcePattern.ReplaceAllString(mediaTitle, "")
	// 移除编码
	mediaTitle = codecPattern.ReplaceAllString(mediaTitle, "")
	// 移除音频
	mediaTitle = audioPattern.ReplaceAllString(mediaTitle, "")
	// 移除发布组
	mediaTitle = groupPattern.ReplaceAllString(mediaTitle, "")

	// 清理标题
	mediaTitle = strings.ReplaceAll(mediaTitle, ".", " ")
	mediaTitle = strings.ReplaceAll(mediaTitle, "_", " ")
	mediaTitle = regexp.MustCompile(`\s+`).ReplaceAllString(mediaTitle, " ")
	mediaTitle = strings.TrimSpace(mediaTitle)

	info.MediaTitle = mediaTitle
}

// MatchesQuality 检查是否匹配质量要求
func (t *TorrentInfo) MatchesQuality(required string) bool {
	if required == "" {
		return true
	}
	return strings.EqualFold(t.Quality, required)
}

// MatchesSource 检查是否匹配来源要求
func (t *TorrentInfo) MatchesSource(required string) bool {
	if required == "" {
		return true
	}
	return strings.Contains(strings.ToLower(t.Source), strings.ToLower(required))
}

// ContainsKeyword 检查是否包含关键词
func (t *TorrentInfo) ContainsKeyword(keyword string) bool {
	if keyword == "" {
		return true
	}
	titleLower := strings.ToLower(t.Title)
	keywordLower := strings.ToLower(keyword)
	return strings.Contains(titleLower, keywordLower)
}

// ExcludesKeyword 检查是否排除关键词
func (t *TorrentInfo) ExcludesKeyword(keyword string) bool {
	if keyword == "" {
		return true
	}
	titleLower := strings.ToLower(t.Title)
	keywordLower := strings.ToLower(keyword)
	return !strings.Contains(titleLower, keywordLower)
}
