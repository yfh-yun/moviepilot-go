package meta

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yfh-yun/moviepilot-go/internal/core/config"
	"github.com/yfh-yun/moviepilot-go/internal/integration/tmdb"
	"github.com/yfh-yun/moviepilot-go/internal/integration/tvdb"
	"github.com/yfh-yun/moviepilot-go/pkg/logger"
)

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeMovie       MediaType = "movie"
	MediaTypeTV          MediaType = "tv"
	MediaTypeEpisode     MediaType = "episode"
	MediaTypeDocumentary MediaType = "documentary"
	MediaTypeAnime       MediaType = "anime"
)

// MediaInfo 媒体信息
type MediaInfo struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	OriginalTitle string    `json:"originalTitle"`
	Type        MediaType   `json:"type"`
	Year        int         `json:"year"`
	ReleaseDate string      `json:"releaseDate"`
	Overview    string      `json:"overview"`
	Rating      float64     `json:"rating"`
	VoteCount   int         `json:"voteCount"`
	Genres      []string    `json:"genres"`
	Languages   []string    `json:"languages"`
	Runtime     int         `json:"runtime"`
	PosterPath  string      `json:"posterPath"`
	BackdropPath string     `json:"backdropPath"`
	VideoPath   string      `json:"videoPath"`
	Subtitles   []string    `json:"subtitles"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// TVSeriesInfo 电视剧信息
type TVSeriesInfo struct {
	*MediaInfo
	Seasons        []SeasonInfo `json:"seasons"`
	EpisodeCount   int          `json:"episodeCount"`
	CurrentSeason  int          `json:"currentSeason"`
	Network        string       `json:"network"`
	FirstAired     string       `json:"firstAired"`
	LastAired      string       `json:"lastAired"`
	Status         string       `json:"status"`
}

// SeasonInfo 季节信息
type SeasonInfo struct {
	SeasonNumber int    `json:"seasonNumber"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	EpisodeCount int    `json:"episodeCount"`
	PosterPath   string `json:"posterPath"`
	AirDate      string `json:"airDate"`
}

// EpisodeInfo 剧集信息
type EpisodeInfo struct {
	*MediaInfo
	SeasonNumber  int    `json:"seasonNumber"`
	EpisodeNumber int    `json:"episodeNumber"`
	AbsoluteNumber int   `json:"absoluteNumber"`
	TVSeriesID    string `json:"tvSeriesId"`
}

// ParseResult 解析结果
type ParseResult struct {
	MediaInfo   *MediaInfo  `json:"mediaInfo"`
	FileInfo    FileInfo    `json:"fileInfo"`
	Confidence  float64     `json:"confidence"`
	ParseErrors []string    `json:"parseErrors"`
}

// FileInfo 文件信息
type FileInfo struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Extension   string `json:"extension"`
	MimeType    string `json:"mimeType"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	Checksum    string `json:"checksum"`
}

// MetadataService 元数据服务
type MetadataService struct {
	tmdbClient *tmdb.Client
	tvdbClient *tvdb.Client
	logger     *logger.Logger
	cache      *MetadataCache
}

// NewMetadataService 创建元数据服务
func NewMetadataService(cfg *config.Config) (*MetadataService, error) {
	tmdbClient, err := tmdb.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建TMDB客户端失败: %w", err)
	}

	tvdbClient := tvdb.NewClient(cfg)

	return &MetadataService{
		tmdbClient: tmdbClient,
		tvdbClient: tvdbClient,
		logger:     logger.NewLogger("metadata"),
		cache:      NewMetadataCache(24*time.Hour, 1000),
	}, nil
}

// ParseMediaFile 解析媒体文件
func (ms *MetadataService) ParseMediaFile(ctx context.Context, filePath string) (*ParseResult, error) {
	// 检查缓存
	if cached, found := ms.cache.Get(filePath); found {
		return cached, nil
	}

	// 解析文件名
	parseResult := &ParseResult{
		FileInfo:   ms.getFileInfo(filePath),
		Confidence: 0.0,
	}

	// 提取文件名中的媒体信息
	mediaInfo, err := ms.extractFromFilename(filePath)
	if err != nil {
		parseResult.ParseErrors = append(parseResult.ParseErrors, err.Error())
	} else {
		parseResult.MediaInfo = mediaInfo
		parseResult.Confidence += 0.3
	}

	// 尝试从文件内容解析
	if contentInfo, err := ms.extractFromContent(filePath); err == nil {
		// 合并内容解析结果
		ms.mergeMediaInfo(parseResult.MediaInfo, contentInfo)
		parseResult.Confidence += 0.4
	}

	// 如果置信度足够高，尝试在线查询
	if parseResult.Confidence >= 0.5 && parseResult.MediaInfo != nil {
		if onlineInfo, err := ms.queryOnline(ctx, parseResult.MediaInfo); err == nil {
			ms.mergeMediaInfo(parseResult.MediaInfo, onlineInfo)
			parseResult.Confidence += 0.3
		}
	}

	// 缓存结果
	ms.cache.Set(filePath, parseResult)

	return parseResult, nil
}

// extractFromFilename 从文件名提取媒体信息
func (ms *MetadataService) extractFromFilename(filePath string) (*MediaInfo, error) {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	fileNameWithoutExt := strings.TrimSuffix(fileName, ext)

	// 常见的命名模式匹配
	patterns := []struct {
		pattern string
		type    MediaType
	}{
		{".+[Ss]\d{2}[Ee]\d{2}.*", MediaTypeEpisode},
		{".+\d{4}\.\d{2}\.\d{2}.*", MediaTypeMovie},
		{".+\.S\d{2}E\d{2}\..*", MediaTypeEpisode},
		{".+\.\d{4}\..*", MediaTypeMovie},
	}

	var mediaType MediaType
	for _, p := range patterns {
		if matched, _ := filepath.Match(p.pattern, fileName); matched {
			mediaType = p.type
			break
		}
	}

	// 提取年份
	year := ms.extractYear(fileNameWithoutExt)

	// 提取标题（去除无用的字符和序列）
	title := ms.cleanTitle(fileNameWithoutExt)

	return &MediaInfo{
		Title:    title,
		Type:     mediaType,
		Year:     year,
		VideoPath: filePath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// extractFromContent 从文件内容提取信息
func (ms *MetadataService) extractFromContent(filePath string) (*MediaInfo, error) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// 尝试读取文件头信息
	headerInfo, err := ms.readFileHeader(filePath)
	if err != nil {
		return nil, err
	}

	return &MediaInfo{
		VideoPath: filePath,
		Runtime:   headerInfo.Duration,
		Metadata: map[string]interface{}{
			"fileSize": fileInfo.Size(),
			"format":   headerInfo.Format,
			"video":    headerInfo.Video,
			"audio":    headerInfo.Audio,
		},
	}, nil
}

// queryOnline 在线查询媒体信息
func (ms *MetadataService) queryOnline(ctx context.Context, mediaInfo *MediaInfo) (*MediaInfo, error) {
	// 根据媒体类型选择不同的API
	switch mediaInfo.Type {
	case MediaTypeMovie:
		return ms.queryMovie(ctx, mediaInfo.Title, mediaInfo.Year)
	case MediaTypeTV, MediaTypeEpisode:
		return ms.queryTV(ctx, mediaInfo.Title, mediaInfo.Year)
	case MediaTypeAnime:
		return ms.queryAnime(ctx, mediaInfo)
	default:
		return nil, fmt.Errorf("不支持的媒体类型: %s", mediaInfo.Type)
	}
}

// queryMovie 查询电影信息
func (ms *MetadataService) queryMovie(ctx context.Context, title string, year int) (*MediaInfo, error) {
	// 使用TMDB查询
	movies, err := ms.tmdbClient.SearchMovie(ctx, title, tmdb.SearchOptions{
		Year: year,
	})
	if err != nil || len(movies) == 0 {
		return nil, fmt.Errorf("TMDB查询失败")
	}

	// 获取详细信息
	movie, err := ms.tmdbClient.GetMovieDetails(ctx, movies[0].ID)
	if err != nil {
		return nil, err
	}

	return &MediaInfo{
		ID:           fmt.Sprintf("%d", movie.ID),
		Title:        movie.Title,
		OriginalTitle: movie.OriginalTitle,
		Type:         MediaTypeMovie,
		Year:         movie.ReleaseDate.Year(),
		ReleaseDate:  movie.ReleaseDate.Format("2006-01-02"),
		Overview:     movie.Overview,
		Rating:       movie.VoteAverage,
		VoteCount:    movie.VoteCount,
		Genres:       ms.extractGenres(movie.Genres),
		Runtime:      movie.Runtime,
		PosterPath:   ms.tmdbClient.GetImageURL(movie.PosterPath, "w500"),
		BackdropPath: ms.tmdbClient.GetImageURL(movie.BackdropPath, "w1280"),
		Metadata: map[string]interface{}{
			"tmdbId": movie.ID,
			"imdbId": movie.IMDBID,
		},
	}, nil
}

// queryTV 查询电视剧信息
func (ms *MetadataService) queryTV(ctx context.Context, title string, year int) (*MediaInfo, error) {
	// 使用TVDB查询
	seriesList, err := ms.tvdbClient.SearchSeries(ctx, title)
	if err != nil || len(seriesList) == 0 {
		return nil, fmt.Errorf("TVDB查询失败")
	}

	// 获取详细信息
	series, err := ms.tvdbClient.GetSeries(ctx, seriesList[0].ID)
	if err != nil {
		return nil, err
	}
}

// queryAnime 查询动漫信息
func (ms *MetadataService) queryAnime(ctx context.Context, mediaInfo *MediaInfo) (*MediaInfo, error) {
	// 优先使用TMDB查询动漫（TMDB有动漫分类）
	movieInfo, err := ms.queryMovie(ctx, mediaInfo.Title, mediaInfo.Year)
	if err == nil {
		// 检查是否为动漫类型
		if ms.isAnimeMediaType(movieInfo) {
			movieInfo.Type = MediaTypeAnime
			return movieInfo, nil
		}
	}

	// 如果TMDB没有找到，尝试TV API（动漫通常以TV形式存储）
	tvInfo, err := ms.queryTV(ctx, mediaInfo.Title, mediaInfo.Year)
	if err == nil {
		tvInfo.Type = MediaTypeAnime
		return tvInfo, nil
	}

	return nil, fmt.Errorf("未找到动漫信息: %s", mediaInfo.Title)
}

// isAnimeMediaType 检查TMDB返回的媒体是否为动漫类型
func (ms *MetadataService) isAnimeMediaType(mediaInfo *MediaInfo) bool {
	// 检查类型关键词
	animeKeywords := []string{
		"animation", "anime", "cartoon", "animated",
		"动画", "动漫", "卡通",
	}

	// 检查类型
	for _, genre := range mediaInfo.Genres {
		genreLower := strings.ToLower(genre)
		for _, keyword := range animeKeywords {
			if strings.Contains(genreLower, keyword) {
				return true
			}
		}
	}

	// 检查标题中是否包含动漫关键词
	titleLower := strings.ToLower(mediaInfo.Title)
	for _, keyword := range animeKeywords {
		if strings.Contains(titleLower, keyword) {
			return true
		}
	}

	return false
}

	return &MediaInfo{
		ID:          fmt.Sprintf("%d", series.ID),
		Title:       series.Name,
		Type:        MediaTypeTV,
		Year:        ms.extractYearFromDate(series.FirstAired),
		ReleaseDate: series.FirstAired,
		Overview:    series.Overview,
		Runtime:     series.Runtime,
		PosterPath:  series.Poster,
		BackdropPath: series.Fanart,
		Metadata: map[string]interface{}{
			"tvdbId": series.ID,
			"imdbId": series.IMDBID,
			"status": series.Status,
		},
	}, nil
}

// 辅助方法
func (ms *MetadataService) getFileInfo(filePath string) FileInfo {
	info, _ := os.Stat(filePath)
	return FileInfo{
		Path:      filePath,
		Name:      filepath.Base(filePath),
		Size:      info.Size(),
		Extension: filepath.Ext(filePath),
		ModifiedAt: info.ModTime(),
	}
}

func (ms *MetadataService) extractYear(text string) int {
	// 从文本中提取年份
	return 0
}

func (ms *MetadataService) cleanTitle(title string) string {
	// 清理标题
	return title
}

func (ms *MetadataService) readFileHeader(filePath string) (*FileHeaderInfo, error) {
	// 读取文件头信息
	return &FileHeaderInfo{}, nil
}

func (ms *MetadataService) extractGenres(genres []tmdb.Genre) []string {
	var result []string
	for _, genre := range genres {
		result = append(result, genre.Name)
	}
	return result
}

func (ms *MetadataService) extractYearFromDate(date string) int {
	if len(date) >= 4 {
		year := 0
		fmt.Sscanf(date[:4], "%d", &year)
		return year
	}
	return 0
}

func (ms *MetadataService) mergeMediaInfo(dest *MediaInfo, source *MediaInfo) {
	// 合并媒体信息
}

// FileHeaderInfo 文件头信息
type FileHeaderInfo struct {
	Duration int
	Format   string
	Video    map[string]interface{}
	Audio    map[string]interface{}
}

// 其他方法实现...

// HealthCheck 健康检查
func (ms *MetadataService) HealthCheck(ctx context.Context) error {
	return nil
}