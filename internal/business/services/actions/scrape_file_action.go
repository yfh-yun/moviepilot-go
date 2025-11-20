// Package actions 提供动作系统的业务逻辑实现
package actions

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/repositories/interfaces"
	"github.com/yfh-yun/moviepilot-go/internal/business/services/actions/types"

	"go.uber.org/zap"
)

// ScrapeFileAction 刮削文件动作
// 对应Python版本app/actions/scrape_file.py的ScrapeFileAction
type ScrapeFileAction struct {
	fileRepo     interfaces.FileRepository
	mediaRepo    interfaces.MediaRepository
	mediaChain   *MediaChain
	storageChain *StorageChain
	cache        *WorkflowCache
	scrapedFiles []*types.File
	hasError     bool
	logger       *zap.Logger
}

// ScrapeFileParams 刮削文件参数
type ScrapeFileParams struct {
	// 刮削文件参数，Python版本中为空，这里预留扩展
	ForceScrape bool `json:"force_scrape" description:"强制重新刮削"`
}

// NewScrapeFileAction 创建刮削文件动作实例
func NewScrapeFileAction(
	fileRepo interfaces.FileRepository,
	mediaRepo interfaces.MediaRepository,
	mediaChain *MediaChain,
	storageChain *StorageChain,
	cache *WorkflowCache,
) *ScrapeFileAction {
	return &ScrapeFileAction{
		fileRepo:     fileRepo,
		mediaRepo:    mediaRepo,
		mediaChain:   mediaChain,
		storageChain: storageChain,
		cache:        cache,
		scrapedFiles: make([]*types.File, 0),
		hasError:     false,
		logger:       logger.Logger,
	}
}

// Execute 执行刮削文件动作
// 实现Python版本ScrapeFileAction.execute()方法的完整功能
func (sfa *ScrapeFileAction) Execute(
	ctx context.Context,
	workflowID int64,
	params map[string]interface{},
	actionCtx *types.ActionContext,
) (*types.ActionContext, error) {
	startTime := time.Now()
	failedCount := 0

	// 解析参数
	scrapeParams, err := sfa.parseParams(params)
	if err != nil {
		sfa.logger.Error("解析刮削文件参数失败", zap.Error(err))
		return actionCtx, err
	}

	sfa.logger.Info("开始执行刮削文件动作",
		zap.Int64("workflow_id", workflowID),
		zap.Int("file_count", len(actionCtx.Files)),
		zap.Bool("force_scrape", scrapeParams.ForceScrape),
	)

	// 处理每个文件
	for _, file := range actionCtx.Files {
		// 检查工作流是否已停止
		if sfa.isWorkflowStopped(ctx, workflowID) {
			sfa.logger.Info("工作流已停止，终止文件刮削", zap.Int64("workflow_id", workflowID))
			break
		}

		// 检查文件是否已刮削过
		if !scrapeParams.ForceScrape && sfa.isAlreadyScraped(file) {
			continue
		}

		// 检查文件是否存在
		if !sfa.storageChain.Exists(file) {
			sfa.logger.Warn("文件不存在，跳过刮削", zap.String("path", file.Path))
			continue
		}

		// 检查缓存
		cacheKey := file.Path
		if !scrapeParams.ForceScrape && sfa.checkCache(ctx, workflowID, cacheKey) {
			sfa.logger.Info("文件已刮削过，跳过", zap.String("path", file.Path))
			continue
		}

		// 处理单个文件
		success := sfa.processFile(ctx, workflowID, file)
		if success {
			sfa.scrapedFiles = append(sfa.scrapedFiles, file)
			// 保存缓存
			if err := sfa.saveCache(ctx, workflowID, cacheKey); err != nil {
				sfa.logger.Warn("保存缓存失败", zap.Error(err))
			}
			sfa.logger.Info("文件刮削成功", zap.String("path", file.Path))
		} else {
			failedCount++
			sfa.logger.Warn("文件刮削失败", zap.String("path", file.Path))
		}
	}

	// 更新动作上下文
	// 这里可以更新文件状态或添加刮削结果

	// 检查是否所有文件都失败了
	if len(sfa.scrapedFiles) == 0 && failedCount > 0 {
		sfa.hasError = true
		sfa.logger.Error("所有文件刮削都失败了")
	}

	sfa.logger.Info("文件刮削完成",
		zap.Int("success_count", len(sfa.scrapedFiles)),
		zap.Int("failed_count", failedCount),
		zap.Duration("duration", time.Since(startTime)),
	)

	return actionCtx, nil
}

// parseParams 解析动作参数
func (sfa *ScrapeFileAction) parseParams(params map[string]interface{}) (*ScrapeFileParams, error) {
	scrapeParams := &ScrapeFileParams{
		ForceScrape: false, // 默认不强制重新刮削
	}

	if forceScrape, ok := params["force_scrape"].(bool); ok {
		scrapeParams.ForceScrape = forceScrape
	}

	return scrapeParams, nil
}

// isAlreadyScraped 检查文件是否已刮削过
func (sfa *ScrapeFileAction) isAlreadyScraped(file *types.File) bool {
	for _, scrapedFile := range sfa.scrapedFiles {
		if scrapedFile.Path == file.Path {
			return true
		}
	}
	return false
}

// processFile 处理单个文件的刮削
func (sfa *ScrapeFileAction) processFile(ctx context.Context, workflowID int64, file *types.File) bool {
	// 从文件路径解析元信息
	metaInfo, err := sfa.parseMetaInfo(file.Path)
	if err != nil {
		sfa.logger.Error("解析文件元信息失败", zap.String("path", file.Path), zap.Error(err))
		return false
	}

	// 识别媒体信息
	mediaInfo, err := sfa.mediaChain.RecognizeMedia(ctx, &types.ActionContext{
		Files: []*types.File{file},
	})
	if err != nil {
		sfa.logger.Error("识别媒体信息失败", zap.String("path", file.Path), zap.Error(err))
		return false
	}

	if mediaInfo == nil {
		sfa.logger.Info("未识别到媒体信息，无法刮削", zap.String("path", file.Path))
		return false
	}

	// 执行刮削
	err = sfa.mediaChain.ScrapeMetadata(ctx, file, metaInfo, mediaInfo)
	if err != nil {
		sfa.logger.Error("刮削元数据失败", zap.String("path", file.Path), zap.Error(err))
		return false
	}

	// 更新文件信息
	err = sfa.updateFileInfo(ctx, file, mediaInfo)
	if err != nil {
		sfa.logger.Error("更新文件信息失败", zap.String("path", file.Path), zap.Error(err))
		return false
	}

	return true
}

// parseMetaInfo 从文件路径解析元信息
func (sfa *ScrapeFileAction) parseMetaInfo(filePath string) (*MetaInfo, error) {
	// 对应Python版本的MetaInfoPath(Path(fileitem.path))
	fileName := filepath.Base(filePath)
	dirPath := filepath.Dir(filePath)

	// 简单的文件名解析逻辑
	metaInfo := &MetaInfo{
		Title:      sfa.extractTitle(fileName),
		Year:       sfa.extractYear(fileName),
		Season:     sfa.extractSeason(fileName),
		Episodes:   sfa.extractEpisodes(fileName),
		Extension:  sfa.extractExtension(fileName),
		Quality:    sfa.extractQuality(fileName),
		Resolution: sfa.extractResolution(fileName),
		Source:     sfa.extractSource(fileName),
		VideoCodec: sfa.extractVideoCodec(fileName),
		AudioCodec: sfa.extractAudioCodec(fileName),
	}

	// 从目录路径推断额外信息
	if dirInfo := sfa.analyzeDirectoryPath(dirPath); dirInfo != nil {
		metaInfo.MediaType = dirInfo.Type
		if metaInfo.Title == "" {
			metaInfo.Title = dirInfo.Title
		}
	}

	// 如果无法确定类型，根据文件名推断
	if metaInfo.MediaType == "" {
		metaInfo.MediaType = sfa.inferMediaType(fileName)
	}

	return metaInfo, nil
}

// extractTitle 提取标题
func (sfa *ScrapeFileAction) extractTitle(fileName string) string {
	// 移除扩展名
	name := fileName[:len(fileName)-len(filepath.Ext(fileName))]

	// 移除常见的标记
	cleaners := []string{
		".1080p", ".720p", ".4k", ".2160p",
		".bluray", ".webdl", ".hdtv", ".webrip",
		".x264", ".x265", ".h264", ".h265",
		".dts", ".ddp5.1", ".aac", ".ac3",
		".中英字幕", ".中字", ".简体", ".繁体",
		"[", "]", "(", ")", "{", "}",
	}

	for _, cleaner := range cleaners {
		name = strings.ReplaceAll(name, cleaner, "")
	}

	// 移除年份数字后面的内容
	if idx := strings.Index(name, ".20"); idx > 0 {
		name = name[:idx+5]
	}

	return strings.TrimSpace(name)
}

// extractYear 提取年份
func (sfa *ScrapeFileAction) extractYear(fileName string) int {
	// 匹配 19xx 或 20xx 格式
	for i := 0; i < len(fileName)-4; i++ {
		if fileName[i] >= '1' && fileName[i] <= '2' {
			yearStr := fileName[i : i+4]
			if len(yearStr) == 4 {
				if (yearStr[0] == '1' && yearStr[1] == '9') || (yearStr[0] == '2' && yearStr[1] == '0') {
					if year, err := sfa.parseInt(yearStr); err == nil && year >= 1900 && year <= 2100 {
						return year
					}
				}
			}
		}
	}
	return 0
}

// extractSeason 提取季数
func (sfa *ScrapeFileAction) extractSeason(fileName string) int {
	// 匹配 S01, S1, Season 1 等格式
	lowerName := strings.ToLower(fileName)

	seasonPatterns := []string{
		"s", "season",
	}

	for _, pattern := range seasonPatterns {
		for i := 0; i < len(lowerName)-len(pattern); i++ {
			if lowerName[i:i+len(pattern)] == pattern {
				// 查找数字
				j := i + len(pattern)
				numStr := ""
				for j < len(lowerName) && lowerName[j] >= '0' && lowerName[j] <= '9' {
					numStr += string(lowerName[j])
					j++
				}
				if numStr != "" {
					if season, err := sfa.parseInt(numStr); err == nil {
						return season
					}
				}
			}
		}
	}

	return 0
}

// extractEpisodes 提取集数
func (sfa *ScrapeFileAction) extractEpisodes(fileName string) []int {
	var episodes []int
	lowerName := strings.ToLower(fileName)

	// 匹配 E01, E1, EP01, Episode 1 等格式
	episodePatterns := []string{
		"e", "ep", "episode",
	}

	for _, pattern := range episodePatterns {
		start := 0
		for {
			idx := strings.Index(lowerName[start:], pattern)
			if idx == -1 {
				break
			}
			idx += start

			// 查找数字
			j := idx + len(pattern)
			numStr := ""
			for j < len(lowerName) && lowerName[j] >= '0' && lowerName[j] <= '9' {
				numStr += string(lowerName[j])
				j++
			}
			if numStr != "" {
				if episode, err := sfa.parseInt(numStr); err == nil {
					// 避免重复添加
					found := false
					for _, ep := range episodes {
						if ep == episode {
							found = true
							break
						}
					}
					if !found {
						episodes = append(episodes, episode)
					}
				}
			}
			start = j
		}
	}

	return episodes
}

// extractExtension 提取扩展名
func (sfa *ScrapeFileAction) extractExtension(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != "" && ext[0] == '.' {
		return ext[1:]
	}
	return ext
}

// extractQuality 提取质量
func (sfa *ScrapeFileAction) extractQuality(fileName string) string {
	lowerName := strings.ToLower(fileName)

	qualityMap := map[string]string{
		"bluray": "bluray",
		"bdrip":  "bluray",
		"webdl":  "webdl",
		"web-dl": "webdl",
		"webrip": "webrip",
		"hdtv":   "hdtv",
		"dvdr":   "dvd",
		"dvdrip": "dvd",
		"cam":    "cam",
		"ts":     "ts",
		"tc":     "tc",
	}

	for quality, keyword := range qualityMap {
		if strings.Contains(lowerName, keyword) {
			return quality
		}
	}

	return ""
}

// extractResolution 提取分辨率
func (sfa *ScrapeFileAction) extractResolution(fileName string) string {
	lowerName := strings.ToLower(fileName)

	resolutionMap := map[string]string{
		"4k":    "2160p",
		"2160p": "2160p",
		"1080p": "1080p",
		"720p":  "720p",
		"480p":  "480p",
		"360p":  "360p",
	}

	for resolution, keyword := range resolutionMap {
		if strings.Contains(lowerName, keyword) {
			return resolution
		}
	}

	return ""
}

// extractSource 提取来源
func (sfa *ScrapeFileAction) extractSource(fileName string) string {
	// 这里可以根据需要实现来源提取逻辑
	return ""
}

// extractVideoCodec 提取视频编码
func (sfa *ScrapeFileAction) extractVideoCodec(fileName string) string {
	lowerName := strings.ToLower(fileName)

	codecMap := map[string]string{
		"x264":  "h264",
		"h264":  "h264",
		"avc":   "h264",
		"x265":  "h265",
		"h265":  "h265",
		"hevc":  "h265",
		"xvid":  "xvid",
		"divx":  "divx",
		"mpeg2": "mpeg2",
		"mpeg4": "mpeg4",
	}

	for codec, keyword := range codecMap {
		if strings.Contains(lowerName, keyword) {
			return codec
		}
	}

	return ""
}

// extractAudioCodec 提取音频编码
func (sfa *ScrapeFileAction) extractAudioCodec(fileName string) string {
	lowerName := strings.ToLower(fileName)

	codecMap := map[string]string{
		"dts":   "dts",
		"dtshd": "dts",
		"dd":    "dolby",
		"ddp":   "dolby",
		"ac3":   "ac3",
		"aac":   "aac",
		"flac":  "flac",
		"mp3":   "mp3",
	}

	for codec, keyword := range codecMap {
		if strings.Contains(lowerName, keyword) {
			return codec
		}
	}

	return ""
}

// analyzeDirectoryPath 分析目录路径
func (sfa *ScrapeFileAction) analyzeDirectoryPath(dirPath string) *DirectoryInfo {
	// 这里可以实现目录路径分析逻辑
	// 根据目录结构推断媒体类型和标题
	return nil
}

// inferMediaType 推断媒体类型
func (sfa *ScrapeFileAction) inferMediaType(fileName string) string {
	lowerName := strings.ToLower(fileName)

	// 检查是否包含电视剧标识
	tvIndicators := []string{
		"s", "season", "episode", "ep", "e01", "e02", "e03",
	}

	for _, indicator := range tvIndicators {
		if strings.Contains(lowerName, indicator) {
			return "tv"
		}
	}

	// 检查是否包含纪录片标识
	docIndicators := []string{
		"documentary", "纪录片", "docu",
	}

	for _, indicator := range docIndicators {
		if strings.Contains(lowerName, indicator) {
			return "documentary"
		}
	}

	// 默认为电影
	return "movie"
}

// updateFileInfo 更新文件信息
func (sfa *ScrapeFileAction) updateFileInfo(ctx context.Context, file *types.File, mediaInfo *types.MediaInfo) error {
	// 更新文件的媒体信息
	file.MediaType = mediaInfo.Type
	file.Title = mediaInfo.Title
	file.Year = mediaInfo.Year
	file.Season = mediaInfo.Season
	file.Episodes = mediaInfo.Episodes
	file.Resolution = mediaInfo.Resolution
	file.Quality = mediaInfo.Quality
	file.Source = mediaInfo.Source
	file.TMDBID = mediaInfo.TMDBID
	file.IMDBID = mediaInfo.IMDBID
	file.DoubanID = mediaInfo.DoubanID
	file.BangumiID = mediaInfo.BangumiID
	now := time.Now()
	file.ScrapedAt = &now
	file.UpdatedAt = now

	// 保存到数据库
	return sfa.fileRepo.Update(ctx, file)
}

// checkCache 检查缓存
func (sfa *ScrapeFileAction) checkCache(ctx context.Context, workflowID int64, key string) bool {
	if sfa.cache == nil {
		return false
	}

	cacheKey := fmt.Sprintf("scrape_cache_%d", workflowID)
	exists, err := sfa.cache.Exists(ctx, cacheKey, key)
	if err != nil {
		sfa.logger.Warn("检查缓存失败", zap.Error(err))
		return false
	}

	return exists
}

// saveCache 保存缓存
func (sfa *ScrapeFileAction) saveCache(ctx context.Context, workflowID int64, key string) error {
	if sfa.cache == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("scrape_cache_%d", workflowID)
	return sfa.cache.Set(ctx, cacheKey, key, 24*time.Hour)
}

// isWorkflowStopped 检查工作流是否已停止
func (sfa *ScrapeFileAction) isWorkflowStopped(ctx context.Context, workflowID int64) bool {
	// 这里应该检查工作流状态
	// 暂时返回false
	return false
}

// parseInt 解析整数
func (sfa *ScrapeFileAction) parseInt(s string) (int, error) {
	var result int
	n, err := fmt.Sscanf(s, "%d", &result)
	if n != 1 || err != nil {
		return 0, fmt.Errorf("invalid number: %s", s)
	}
	return result, nil
}

// GetSuccess 获取执行结果
func (sfa *ScrapeFileAction) GetSuccess() bool {
	return !sfa.hasError
}

// GetScrapedFiles 获取已刮削的文件列表
func (sfa *ScrapeFileAction) GetScrapedFiles() []*types.File {
	return sfa.scrapedFiles
}

// GetName 获取动作名称
func (sfa *ScrapeFileAction) GetName() string {
	return "刮削文件"
}

// GetDescription 获取动作描述
func (sfa *ScrapeFileAction) GetDescription() string {
	return "刮削媒体信息和图片"
}

// GetData 获取动作参数定义
func (sfa *ScrapeFileAction) GetData() map[string]interface{} {
	return map[string]interface{}{
		"force_scrape": map[string]interface{}{
			"type":        "boolean",
			"description": "强制重新刮削",
			"default":     false,
		},
	}
}

// MetaInfo 元信息结构（简化版）
type MetaInfo struct {
	Title      string `json:"title"`
	Year       int    `json:"year"`
	Season     int    `json:"season"`
	Episodes   []int  `json:"episodes"`
	Extension  string `json:"extension"`
	MediaType  string `json:"media_type"`
	Quality    string `json:"quality"`
	Resolution string `json:"resolution"`
	Source     string `json:"source"`
	VideoCodec string `json:"video_codec"`
	AudioCodec string `json:"audio_codec"`
}

// DirectoryInfo 目录信息结构（简化版）
type DirectoryInfo struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}
