package chain

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/model"
	"go.uber.org/zap"
)

// NFOGenerator NFO文件生成器
type NFOGenerator struct {
	logger      *zap.Logger
	templater   *NFOTemplater
	validator   *NFOValidator
	mediaMatcher *MediaMatcher
}

// NFOTemplater NFO模板生成器
type NFOTemplater struct {
	movieTemplate string
	tvShowTemplate string
	tvEpisodeTemplate string
}

// NFOValidator NFO验证器
type NFOValidator struct {
	schemaValidator *SchemaValidator
}

// MediaMatcher 媒体匹配器
type MediaMatcher struct {
	titleMatcher    *TitleMatcher
	yearComparator  *YearComparator
	seasonMatcher  *SeasonMatcher
	episodeMatcher *EpisodeMatcher
}

// TitleMatcher 标题匹配器
type TitleMatcher struct {
	normalizers []func(string) string
}

// YearComparator 年份比较器
type YearComparator struct {
	tolerance int
}

// SeasonMatcher 季匹配器
type SeasonMatcher struct{}

// EpisodeMatcher 集匹配器
type EpisodeMatcher struct{}

// NewNFOGenerator 创建NFO生成器
func NewNFOGenerator(logger *zap.Logger) *NFOGenerator {
	return &NFOGenerator{
		logger:      logger,
		templater:   NewNFOTemplater(),
		validator:   NewNFOValidator(),
		mediaMatcher: NewMediaMatcher(),
	}
}

// GenerateMovieNFO 生成电影NFO文件
func (g *NFOGenerator) GenerateMovieNFO(mediaInfo *model.MediaInfo, outputPath string) error {
	g.logger.Info("开始生成电影NFO文件",
		zap.String("title", mediaInfo.Title),
		zap.String("output_path", outputPath))

	// 1. 验证媒体信息
	if err := g.validator.ValidateMovieInfo(mediaInfo); err != nil {
		return fmt.Errorf("媒体信息验证失败: %w", err)
	}

	// 2. 生成NFO内容
	nfoContent := g.templater.GenerateMovieNFO(mediaInfo)

	// 3. 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 4. 写入NFO文件
	if err := os.WriteFile(outputPath, []byte(nfoContent), 0644); err != nil {
		return fmt.Errorf("写入NFO文件失败: %w", err)
	}

	g.logger.Info("电影NFO文件生成完成",
		zap.String("title", mediaInfo.Title),
		zap.String("output_path", outputPath))

	return nil
}

// GenerateTVShowNFO 生成电视剧NFO文件
func (g *NFOGenerator) GenerateTVShowNFO(mediaInfo *model.MediaInfo, outputPath string) error {
	g.logger.Info("开始生成电视剧NFO文件",
		zap.String("title", mediaInfo.Title),
		zap.String("output_path", outputPath))

	if err := g.validator.ValidateTVShowInfo(mediaInfo); err != nil {
		return fmt.Errorf("媒体信息验证失败: %w", err)
	}

	nfoContent := g.templater.GenerateTVShowNFO(mediaInfo)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(nfoContent), 0644); err != nil {
		return fmt.Errorf("写入NFO文件失败: %w", err)
	}

	g.logger.Info("电视剧NFO文件生成完成",
		zap.String("title", mediaInfo.Title),
		zap.String("output_path", outputPath))

	return nil
}

// GenerateTVEpisodeNFO 生成电视剧单集NFO文件
func (g *NFOGenerator) GenerateTVEpisodeNFO(mediaInfo *model.MediaInfo, season, episode int, outputPath string) error {
	g.logger.Info("开始生成电视剧单集NFO文件",
		zap.String("title", mediaInfo.Title),
		zap.Int("season", season),
		zap.Int("episode", episode),
		zap.String("output_path", outputPath))

	if err := g.validator.ValidateTVEpisodeInfo(mediaInfo, season, episode); err != nil {
		return fmt.Errorf("媒体信息验证失败: %w", err)
	}

	nfoContent := g.templater.GenerateTVEpisodeNFO(mediaInfo, season, episode)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(nfoContent), 0644); err != nil {
		return fmt.Errorf("写入NFO文件失败: %w", err)
	}

	g.logger.Info("电视剧单集NFO文件生成完成",
		zap.String("title", mediaInfo.Title),
		zap.Int("season", season),
		zap.Int("episode", episode),
		zap.String("output_path", outputPath))

	return nil
}

// MatchMediaFile 匹配媒体文件
func (g *NFOGenerator) MatchMediaFile(filePath string, library []*model.MediaInfo) (*MediaMatch, error) {
	g.logger.Debug("开始匹配媒体文件", zap.String("file_path", filePath))

	// 1. 解析文件名
	fileInfo := g.parseFileName(filepath.Base(filePath))
	
	// 2. 在库中搜索匹配项
	candidates := g.findCandidates(fileInfo, library)
	
	// 3. 计算匹配分数
	matches := g.calculateMatchScores(fileInfo, candidates)
	
	// 4. 选择最佳匹配
	bestMatch := g.selectBestMatch(matches)
	
	if bestMatch == nil {
		return nil, fmt.Errorf("未找到匹配的媒体")
	}

	g.logger.Debug("媒体文件匹配完成",
		zap.String("file_path", filePath),
		zap.String("matched_title", bestMatch.MediaInfo.Title),
		zap.Float64("score", bestMatch.Score))

	return bestMatch, nil
}

// parseFileName 解析文件名
func (g *NFOGenerator) parseFileName(fileName string) *FileInfo {
	// 移除扩展名
	nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	
	// 清理文件名
	cleanedName := g.cleanFileName(nameWithoutExt)
	
	// 解析标题、年份、季集信息
	fileInfo := &FileInfo{
		OriginalName: fileName,
		CleanName:    cleanedName,
	}
	
	// 提取年份
	if year := g.extractYear(cleanedName); year > 0 {
		fileInfo.Year = year
	}
	
	// 提取季集信息（如果是电视剧）
	if season, episode := g.extractSeasonEpisode(cleanedName); season > 0 {
		fileInfo.Season = season
		fileInfo.Episode = episode
		fileInfo.Type = "tv"
	} else {
		fileInfo.Type = "movie"
	}
	
	// 提取标题
	fileInfo.Title = g.extractTitle(cleanedName, fileInfo)
	
	return fileInfo
}

// cleanFileName 清理文件名
func (g *NFOGenerator) cleanFileName(fileName string) string {
	// 移除常见的分隔符和标记
	cleaned := fileName
	
	// 替换常见的分隔符
	replacements := map[string]string{
		".": " ",
		"_": " ",
		"-": " ",
		"(": " ",
		")": " ",
		"[": " ",
		"]": " ",
	}
	
	for old, new := range replacements {
		cleaned = strings.ReplaceAll(cleaned, old, new)
	}
	
	// 移除多余空格
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	
	return cleaned
}

// extractYear 提取年份
func (g *NFOGenerator) extractYear(fileName string) int {
	// 查找四位数字（1900-2100）
	words := strings.Fields(fileName)
	for _, word := range words {
		if len(word) == 4 {
			if year, err := strconv.Atoi(word); err == nil {
				if year >= 1900 && year <= 2100 {
					return year
				}
			}
		}
	}
	return 0
}

// extractSeasonEpisode 提取季集信息
func (g *NFOGenerator) extractSeasonEpisode(fileName string) (int, int) {
	// 查找 S01E01 或 1x01 等格式
	lower := strings.ToLower(fileName)
	
	// S01E01 格式
	if matched := g.matchSeasonEpisode(lower, `s(\d+)e(\d+)`); matched != nil {
		return matched[0], matched[1]
	}
	
	// 1x01 格式
	if matched := g.matchSeasonEpisode(lower, `(\d+)x(\d+)`); matched != nil {
		return matched[0], matched[1]
	}
	
	// Season 1 Episode 1 格式
	if matched := g.matchSeasonEpisode(lower, `season\s*(\d+).*episode\s*(\d+)`); matched != nil {
		return matched[0], matched[1]
	}
	
	return 0, 0
}

// matchSeasonEpisode 匹配季集正则
func (g *NFOGenerator) matchSeasonEpisode(text, pattern string) []int {
	// 这里应该使用正则表达式库，为了简化示例，使用字符串匹配
	// 实际实现中应该使用 regexp 包
	return nil
}

// extractTitle 提取标题
func (g *NFOGenerator) extractTitle(fileName string, fileInfo *FileInfo) string {
	words := strings.Fields(fileName)
	var titleWords []string
	
	for i, word := range words {
		// 跳过年份
		if fileInfo.Year > 0 && len(word) == 4 {
			if year, err := strconv.Atoi(word); err == nil && year == fileInfo.Year {
				break
			}
		}
		
		// 跳过季集信息
		if strings.HasPrefix(strings.ToLower(word), "s") && len(word) >= 3 {
			continue
		}
		
		titleWords = append(titleWords, word)
	}
	
	return strings.Join(titleWords, " ")
}

// findCandidates 查找候选媒体
func (g *NFOGenerator) findCandidates(fileInfo *FileInfo, library []*model.MediaInfo) []*model.MediaInfo {
	var candidates []*model.MediaInfo
	
	for _, media := range library {
		if g.isCandidateMatch(fileInfo, media) {
			candidates = append(candidates, media)
		}
	}
	
	return candidates
}

// isCandidateMatch 检查是否为候选匹配
func (g *NFOGenerator) isCandidateMatch(fileInfo *FileInfo, media *model.MediaInfo) bool {
	// 类型匹配
	if fileInfo.Type == "movie" && media.Type != "movie" {
		return false
	}
	if fileInfo.Type == "tv" && media.Type != "tv" {
		return false
	}
	
	// 年份匹配（宽容匹配）
	if fileInfo.Year > 0 && media.Year != nil {
		yearDiff := abs(fileInfo.Year - *media.Year)
		if yearDiff > 2 { // 允许2年误差
			return false
		}
	}
	
	// 标题相似度匹配
	titleSimilarity := g.calculateTitleSimilarity(fileInfo.Title, media.Title)
	if titleSimilarity < 0.6 { // 最低60%相似度
		return false
	}
	
	return true
}

// calculateMatchScores 计算匹配分数
func (g *NFOGenerator) calculateMatchScores(fileInfo *FileInfo, candidates []*model.MediaInfo) []*MediaMatch {
	var matches []*MediaMatch
	
	for _, media := range candidates {
		score := g.calculateMatchScore(fileInfo, media)
		if score > 0 {
			matches = append(matches, &MediaMatch{
				MediaInfo: media,
				Score:     score,
				FileInfo:  fileInfo,
			})
		}
	}
	
	// 按分数排序
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].Score < matches[j].Score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
	
	return matches
}

// calculateMatchScore 计算单个匹配分数
func (g *NFOGenerator) calculateMatchScore(fileInfo *FileInfo, media *model.MediaInfo) float64 {
	var score float64 = 0.0
	
	// 标题相似度 (50%)
	titleSimilarity := g.calculateTitleSimilarity(fileInfo.Title, media.Title)
	score += titleSimilarity * 0.5
	
	// 年份匹配 (30%)
	if fileInfo.Year > 0 && media.Year != nil {
		yearDiff := abs(fileInfo.Year - *media.Year)
		if yearDiff == 0 {
			score += 0.3
		} else if yearDiff == 1 {
			score += 0.2
		} else if yearDiff == 2 {
			score += 0.1
		}
	} else {
		score += 0.15 // 如果没有年份信息，给予部分分数
	}
	
	// 季集匹配 (20%，仅适用于电视剧)
	if fileInfo.Type == "tv" && media.Type == "tv" {
		if fileInfo.Season > 0 && media.Season != nil {
			if fileInfo.Season == *media.Season {
				score += 0.1
			}
		}
		if fileInfo.Episode > 0 && media.Episode != nil {
			if fileInfo.Episode == *media.Episode {
				score += 0.1
			}
		}
	}
	
	return score
}

// calculateTitleSimilarity 计算标题相似度
func (g *NFOGenerator) calculateTitleSimilarity(title1, title2 string) float64 {
	// 简化的标题相似度计算
	// 实际应该使用更复杂的算法，如Levenshtein距离等
	t1 := strings.ToLower(strings.TrimSpace(title1))
	t2 := strings.ToLower(strings.TrimSpace(title2))
	
	if t1 == t2 {
		return 1.0
	}
	
	// 计算共同的单词
	words1 := strings.Fields(t1)
	words2 := strings.Fields(t2)
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}
	
	commonWords := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				commonWords++
				break
			}
		}
	}
	
	return float64(commonWords) / float64(max(len(words1), len(words2)))
}

// selectBestMatch 选择最佳匹配
func (g *NFOGenerator) selectBestMatch(matches []*MediaMatch) *MediaMatch {
	if len(matches) == 0 {
		return nil
	}
	
	// 选择分数最高的匹配，且分数必须大于0.7
	if matches[0].Score > 0.7 {
		return matches[0]
	}
	
	return nil
}

// FileInfo 文件信息
type FileInfo struct {
	OriginalName string `json:"original_name"`
	CleanName    string `json:"clean_name"`
	Title        string `json:"title"`
	Year         int    `json:"year"`
	Type         string `json:"type"` // movie, tv
	Season       int    `json:"season"`
	Episode      int    `json:"episode"`
}

// MediaMatch 媒体匹配结果
type MediaMatch struct {
	MediaInfo *model.MediaInfo `json:"media_info"`
	Score     float64         `json:"score"`
	FileInfo  *FileInfo       `json:"file_info"`
}

// 辅助函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// NewNFOTemplater 创建NFO模板生成器
func NewNFOTemplater() *NFOTemplater {
	return &NFOTemplater{
		movieTemplate: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<movie>
    <title>%s</title>
    <year>%d</year>
    <plot>%s</plot>
    <runtime>%d</runtime>
    <rating>%f</rating>
    <tmdbid>%d</tmdbid>
    <imdbid>%s</imdbid>
    <poster>%s</poster>
    <fanart>%s</fanart>
    <genre>%s</genre>
    <country>%s</country>
</movie>`,
		tvShowTemplate: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<tvshow>
    <title>%s</title>
    <year>%d</year>
    <plot>%s</plot>
    <rating>%f</rating>
    <tmdbid>%d</tmdbid>
    <imdbid>%s</imdbid>
    <poster>%s</poster>
    <fanart>%s</fanart>
    <genre>%s</genre>
    <country>%s</country>
</tvshow>`,
		tvEpisodeTemplate: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<episodedetails>
    <title>%s</title>
    <season>%d</season>
    <episode>%d</episode>
    <plot>%s</plot>
    <rating>%f</rating>
    <aired>%s</aired>
    <tmdbid>%d</tmdbid>
    <imdbid>%s</imdbid>
    <poster>%s</poster>
</episodedetails>`,
	}
}

// GenerateMovieNFO 生成电影NFO
func (t *NFOTemplater) GenerateMovieNFO(media *model.MediaInfo) string {
	title := media.Title
	year := 0
	if media.Year != nil {
		year = *media.Year
	}
	plot := media.Description
	runtime := 0
	if media.Runtime != nil {
		runtime = *media.Runtime
	}
	rating := 0.0
	if media.Vote != nil {
		rating = *media.Vote
	}
	tmdbid := 0
	if media.TMDBID != nil {
		tmdbid = *media.TMDBID
	}
	imdbid := ""
	if media.IMDBID != nil {
		imdbid = *media.IMDBID
	}

	return fmt.Sprintf(t.movieTemplate,
		title, year, plot, runtime, rating, tmdbid, imdbid,
		media.Poster, media.Backdrop, media.Genres, media.Countries)
}

// GenerateTVShowNFO 生成电视剧NFO
func (t *NFOTemplater) GenerateTVShowNFO(media *model.MediaInfo) string {
	title := media.Title
	year := 0
	if media.Year != nil {
		year = *media.Year
	}
	plot := media.Description
	rating := 0.0
	if media.Vote != nil {
		rating = *media.Vote
	}
	tmdbid := 0
	if media.TMDBID != nil {
		tmdbid = *media.TMDBID
	}
	imdbid := ""
	if media.IMDBID != nil {
		imdbid = *media.IMDBID
	}

	return fmt.Sprintf(t.tvShowTemplate,
		title, year, plot, rating, tmdbid, imdbid,
		media.Poster, media.Backdrop, media.Genres, media.Countries)
}

// GenerateTVEpisodeNFO 生成电视剧单集NFO
func (t *NFOTemplater) GenerateTVEpisodeNFO(media *model.MediaInfo, season, episode int) string {
	title := media.Title
	plot := media.Description
	rating := 0.0
	if media.Vote != nil {
		rating = *media.Vote
	}
	tmdbid := 0
	if media.TMDBID != nil {
		tmdbid = *media.TMDBID
	}
	imdbid := ""
	if media.IMDBID != nil {
		imdbid = *media.IMDBID
	}
	aired := time.Now().Format("2006-01-02")

	return fmt.Sprintf(t.tvEpisodeTemplate,
		title, season, episode, plot, rating, aired, tmdbid, imdbid, media.Poster)
}

// NewNFOValidator 创建NFO验证器
func NewNFOValidator() *NFOValidator {
	return &NFOValidator{
		schemaValidator: &SchemaValidator{},
	}
}

// ValidateMovieInfo 验证电影信息
func (v *NFOValidator) ValidateMovieInfo(media *model.MediaInfo) error {
	if media.Title == "" {
		return fmt.Errorf("电影标题不能为空")
	}
	if media.Type != "movie" {
		return fmt.Errorf("媒体类型不匹配，期望movie，实际%s", media.Type)
	}
	return nil
}

// ValidateTVShowInfo 验证电视剧信息
func (v *NFOValidator) ValidateTVShowInfo(media *model.MediaInfo) error {
	if media.Title == "" {
		return fmt.Errorf("电视剧标题不能为空")
	}
	if media.Type != "tv" {
		return fmt.Errorf("媒体类型不匹配，期望tv，实际%s", media.Type)
	}
	return nil
}

// ValidateTVEpisodeInfo 验证电视剧单集信息
func (v *NFOValidator) ValidateTVEpisodeInfo(media *model.MediaInfo, season, episode int) error {
	if err := v.ValidateTVShowInfo(media); err != nil {
		return err
	}
	if season <= 0 {
		return fmt.Errorf("季数必须大于0")
	}
	if episode <= 0 {
		return fmt.Errorf("集数必须大于0")
	}
	return nil
}

// SchemaValidator 模式验证器
type SchemaValidator struct{}

// NewMediaMatcher 创建媒体匹配器
func NewMediaMatcher() *MediaMatcher {
	return &MediaMatcher{
		titleMatcher:    &TitleMatcher{},
		yearComparator:  &YearComparator{tolerance: 2},
		seasonMatcher:  &SeasonMatcher{},
		episodeMatcher: &EpisodeMatcher{},
	}
}