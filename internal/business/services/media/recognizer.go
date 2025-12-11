package media

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// Recognizer 媒体识别器
type Recognizer struct {
	logger *zap.Logger
}

// NewRecognizer 创建媒体识别器
func NewRecognizer(logger *zap.Logger) *Recognizer {
	return &Recognizer{
		logger: logger,
	}
}

// MediaInfo 媒体信息
type MediaInfo struct {
	Title        string // 标题
	Year         int    // 年份
	Season       int    // 季数
	Episode      int    // 集数
	Type         string // 类型 (movie, tv)
	Quality      string // 质量
	Source       string // 来源
	Codec        string // 编码
	Audio        string // 音频
	Group        string // 发布组
	OriginalName string // 原始文件名
}

// RecognizeFromPath 从路径识别媒体信息
func (r *Recognizer) RecognizeFromPath(ctx context.Context, path string) (*MediaInfo, error) {
	filename := filepath.Base(path)
	return r.RecognizeFromFilename(ctx, filename)
}

// RecognizeFromFilename 从文件名识别媒体信息
func (r *Recognizer) RecognizeFromFilename(ctx context.Context, filename string) (*MediaInfo, error) {
	info := &MediaInfo{
		OriginalName: filename,
	}

	// 移除扩展名
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// 识别季集信息
	if season, episode, found := r.extractSeasonEpisode(name); found {
		info.Season = season
		info.Episode = episode
		info.Type = "tv"
	} else {
		info.Type = "movie"
	}

	// 识别年份
	if year, found := r.extractYear(name); found {
		info.Year = year
	}

	// 识别质量
	info.Quality = r.extractQuality(name)

	// 识别来源
	info.Source = r.extractSource(name)

	// 识别编码
	info.Codec = r.extractCodec(name)

	// 识别音频
	info.Audio = r.extractAudio(name)

	// 识别发布组
	info.Group = r.extractGroup(name)

	// 识别标题
	info.Title = r.extractTitle(name, info)

	if r.logger != nil {
		r.logger.Debug("media recognized",
			zap.String("filename", filename),
			zap.String("title", info.Title),
			zap.String("type", info.Type))
	}

	return info, nil
}

// extractSeasonEpisode 提取季集信息
func (r *Recognizer) extractSeasonEpisode(name string) (season int, episode int, found bool) {
	// S01E05, S1E5
	re := regexp.MustCompile(`[Ss](\d{1,2})[Ee](\d{1,3})`)
	if matches := re.FindStringSubmatch(name); len(matches) == 3 {
		season, _ = strconv.Atoi(matches[1])
		episode, _ = strconv.Atoi(matches[2])
		return season, episode, true
	}

	// 1x05
	re = regexp.MustCompile(`(\d{1,2})x(\d{1,3})`)
	if matches := re.FindStringSubmatch(name); len(matches) == 3 {
		season, _ = strconv.Atoi(matches[1])
		episode, _ = strconv.Atoi(matches[2])
		return season, episode, true
	}

	return 0, 0, false
}

// extractYear 提取年份
func (r *Recognizer) extractYear(name string) (int, bool) {
	re := regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	if matches := re.FindStringSubmatch(name); len(matches) > 0 {
		year, _ := strconv.Atoi(matches[1])
		return year, true
	}
	return 0, false
}

// extractQuality 提取质量
func (r *Recognizer) extractQuality(name string) string {
	qualities := []string{"2160p", "1080p", "720p", "480p", "4K", "UHD"}
	nameLower := strings.ToLower(name)
	for _, quality := range qualities {
		if strings.Contains(nameLower, strings.ToLower(quality)) {
			return quality
		}
	}
	return ""
}

// extractSource 提取来源
func (r *Recognizer) extractSource(name string) string {
	sources := []string{"BluRay", "WEB-DL", "WEBRip", "HDTV", "DVDRip", "BDRip"}
	nameLower := strings.ToLower(name)
	for _, source := range sources {
		if strings.Contains(nameLower, strings.ToLower(source)) {
			return source
		}
	}
	return ""
}

// extractCodec 提取编码
func (r *Recognizer) extractCodec(name string) string {
	codecs := []string{"x265", "x264", "H.265", "H.264", "HEVC", "AVC"}
	nameLower := strings.ToLower(name)
	for _, codec := range codecs {
		if strings.Contains(nameLower, strings.ToLower(codec)) {
			return codec
		}
	}
	return ""
}

// extractAudio 提取音频
func (r *Recognizer) extractAudio(name string) string {
	audios := []string{"Atmos", "DTS-HD", "DTS", "AC3", "AAC", "TrueHD"}
	nameLower := strings.ToLower(name)
	for _, audio := range audios {
		if strings.Contains(nameLower, strings.ToLower(audio)) {
			return audio
		}
	}
	return ""
}

// extractGroup 提取发布组
func (r *Recognizer) extractGroup(name string) string {
	// 通常在文件名末尾，用 - 分隔
	re := regexp.MustCompile(`-([A-Za-z0-9]+)$`)
	if matches := re.FindStringSubmatch(name); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractTitle 提取标题
func (r *Recognizer) extractTitle(name string, info *MediaInfo) string {
	// 移除年份
	if info.Year > 0 {
		name = strings.ReplaceAll(name, fmt.Sprintf("%d", info.Year), "")
	}

	// 移除季集信息
	if info.Season > 0 {
		re := regexp.MustCompile(`[Ss]\d{1,2}[Ee]\d{1,3}`)
		name = re.ReplaceAllString(name, "")
		re = regexp.MustCompile(`\d{1,2}x\d{1,3}`)
		name = re.ReplaceAllString(name, "")
	}

	// 移除质量、来源等信息
	removePatterns := []string{
		info.Quality, info.Source, info.Codec, info.Audio, info.Group,
		"BluRay", "WEB-DL", "WEBRip", "HDTV", "x264", "x265", "HEVC",
	}

	for _, pattern := range removePatterns {
		if pattern != "" {
			name = strings.ReplaceAll(name, pattern, "")
			name = strings.ReplaceAll(name, strings.ToLower(pattern), "")
			name = strings.ReplaceAll(name, strings.ToUpper(pattern), "")
		}
	}

	// 清理特殊字符
	name = strings.ReplaceAll(name, ".", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// 移除多余空格
	name = strings.TrimSpace(name)
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	return name
}
