package media

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"moviepilot-go/internal/models"
)

// NFOData NFO文件数据结构
type NFOData struct {
	XMLName xml.Name `xml:"movie"`
	Title   string   `xml:"title,omitempty"`
	Plot    string   `xml:"plot,omitempty"`
	Year    int      `xml:"year,omitempty"`
	Runtime int      `xml:"runtime,omitempty"`
	Rating  float64  `xml:"rating,omitempty"`
	Votes   int      `xml:"votes,omitempty"`
	IMDBID  string   `xml:"imdbid,omitempty"`
	TMDBID  int      `xml:"tmdbid,omitempty"`
	Poster  string   `xml:"thumb,omitempty"`
	// 电影特有字段
	Tagline string   `xml:"tagline,omitempty"`
	Genres  []string `xml:"genre,omitempty"`
	Actors  []Actor  `xml:"actor"`
	Director []string `xml:"director"`
	Writer  []string `xml:"writer"`
	// 电视剧特有字段
	ShowTitle string `xml:"showtitle,omitempty"`
	Season    int    `xml:"season,omitempty"`
	Episode   int    `xml:"episode,omitempty"`
	EpisodeTitle string `xml:"episodetitle,omitempty"`
	AiredDate string `xml:"aired,omitempty"`
}

// TVShowNFOData 电视剧NFO文件数据结构
type TVShowNFOData struct {
	XMLName xml.Name `xml:"tvshow"`
	Title   string   `xml:"title,omitempty"`
	Plot    string   `xml:"plot,omitempty"`
	Rating  float64  `xml:"rating,omitempty"`
	Votes   int      `xml:"votes,omitempty"`
	IMDBID  string   `xml:"imdbid,omitempty"`
	TMDBID  int      `xml:"tmdbid,omitempty"`
	Poster  string   `xml:"thumb,omitempty"`
	Tagline string   `xml:"tagline,omitempty"`
	Genres  []string `xml:"genre,omitempty"`
	Actors  []Actor  `xml:"actor"`
	Director []string `xml:"director"`
	Writer  []string `xml:"writer"`
	Year    int      `xml:"year,omitempty"`
	Studio  string   `xml:"studio,omitempty"`
	Seasons []Season `xml:"season"`
}

// EpisodeNFOData 集NFO文件数据结构
type EpisodeNFOData struct {
	XMLName xml.Name `xml:"episodedetails"`
	Title   string   `xml:"title,omitempty"`
	ShowTitle string `xml:"showtitle,omitempty"`
	Season  int    `xml:"season,omitempty"`
	Episode int    `xml:"episode,omitempty"`
	Plot    string   `xml:"plot,omitempty"`
	Rating  float64  `xml:"rating,omitempty"`
	Votes   int      `xml:"votes,omitempty"`
	IMDBID  string   `xml:"imdbid,omitempty"`
	TMDBID  int      `xml:"tmdbid,omitempty"`
	Thumb   string   `xml:"thumb,omitempty"`
	AiredDate string `xml:"aired,omitempty"`
	Director []string `xml:"director"`
	Writer  []string `xml:"writer"`
	Actors  []Actor  `xml:"actor"`
	Runtime int      `xml:"runtime,omitempty"`
	Year    int      `xml:"year,omitempty"`
}

// Actor 演员
type Actor struct {
	Name  string `xml:"name"`
	Role  string `xml:"role"`
	Order int    `xml:"order"`
}

// Season 季信息
type Season struct {
	Number int    `xml:"number"`
	Name   string `xml:"name"`
}

// ReadNFO 读取NFO文件
func ReadNFO(path string, logger *zap.Logger) (*NFOData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open NFO file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read NFO file: %w", err)
	}

	var nfo NFOData
	err = xml.Unmarshal(data, &nfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NFO file: %w", err)
	}

	if logger != nil {
		logger.Debug("NFO file read successfully", zap.String("path", path), zap.String("title", nfo.Title))
	}

	return &nfo, nil
}

// ReadTVShowNFO 读取电视剧NFO文件
func ReadTVShowNFO(path string, logger *zap.Logger) (*TVShowNFOData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open TV show NFO file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read TV show NFO file: %w", err)
	}

	var nfo TVShowNFOData
	err = xml.Unmarshal(data, &nfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TV show NFO file: %w", err)
	}

	if logger != nil {
		logger.Debug("TV show NFO file read successfully", zap.String("path", path), zap.String("title", nfo.Title))
	}

	return &nfo, nil
}

// ReadEpisodeNFO 读取集NFO文件
func ReadEpisodeNFO(path string, logger *zap.Logger) (*EpisodeNFOData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open episode NFO file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read episode NFO file: %w", err)
	}

	var nfo EpisodeNFOData
	err = xml.Unmarshal(data, &nfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse episode NFO file: %w", err)
	}

	if logger != nil {
		logger.Debug("Episode NFO file read successfully", zap.String("path", path), zap.String("title", nfo.Title))
	}

	return &nfo, nil
}

// WriteMovieNFO 写入电影NFO文件
func WriteMovieNFO(media *models.Media, path string, logger *zap.Logger) error {
	nfo := &NFOData{
		XMLName: xml.Name{Local: "movie"},
		Title:    media.Title,
		Plot:     media.Description,
		Rating:   getFloatValue(media.Vote),
		IMDBID:   getStringValue(media.IMDBID),
		Poster:   media.Poster,
	}

	// 设置年份
	if media.Year != nil {
		if year, err := strconv.Atoi(*media.Year); err == nil {
			nfo.Year = year
		}
	}

	// 设置TMDB ID
	if media.TMDBID != nil {
		nfo.TMDBID = *media.TMDBID
	}

	// 设置运行时间
	if media.Runtime != nil {
		nfo.Runtime = *media.Runtime
	}

	// 解析类型
	if media.Genres != "" {
		var genres []string
		if err := parseJSONField(media.Genres, &genres); err == nil {
			nfo.Genres = genres
		}
	}

	return writeNFOFile(nfo, path, logger)
}

// WriteTVShowNFO 写入电视剧NFO文件
func WriteTVShowNFO(media *models.Media, path string, logger *zap.Logger) error {
	nfo := &TVShowNFOData{
		XMLName: xml.Name{Local: "tvshow"},
		Title:    media.Title,
		Plot:     media.Description,
		Rating:   getFloatValue(media.Vote),
		IMDBID:   getStringValue(media.IMDBID),
		Poster:   media.Poster,
	}

	// 设置年份
	if media.Year != nil {
		if year, err := strconv.Atoi(*media.Year); err == nil {
			nfo.Year = year
		}
	}

	// 设置TMDB ID
	if media.TMDBID != nil {
		nfo.TMDBID = *media.TMDBID
	}

	// 解析类型
	if media.Genres != "" {
		var genres []string
		if err := parseJSONField(media.Genres, &genres); err == nil {
			nfo.Genres = genres
		}
	}

	return writeNFOFile(nfo, path, logger)
}

// WriteEpisodeNFO 写入集NFO文件
func WriteEpisodeNFO(media *models.Media, season, episode int, episodeTitle string, path string, logger *zap.Logger) error {
	nfo := &EpisodeNFOData{
		XMLName: xml.Name{Local: "episodedetails"},
		Title:    episodeTitle,
		ShowTitle: media.Title,
		Season:   season,
		Episode:  episode,
		Plot:     media.Description,
		Rating:   getFloatValue(media.Vote),
		IMDBID:   getStringValue(media.IMDBID),
		Thumb:    media.Poster,
	}

	// 设置年份
	if media.Year != nil {
		if year, err := strconv.Atoi(*media.Year); err == nil {
			nfo.Year = year
		}
	}

	// 设置TMDB ID
	if media.TMDBID != nil {
		nfo.TMDBID = *media.TMDBID
	}

	// 设置运行时间
	if media.Runtime != nil {
		nfo.Runtime = *media.Runtime
	}

	// 设置播出日期
	if media.Year != nil {
		nfo.AiredDate = *media.Year + "-01-01" // 简化处理
	}

	return writeNFOFile(nfo, path, logger)
}

// writeNFOFile 写入NFO文件到磁盘
func writeNFOFile(data interface{}, path string, logger *zap.Logger) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create NFO file: %w", err)
	}
	defer file.Close()

	// XML头部
	_, err = file.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	if err != nil {
		return fmt.Errorf("failed to write XML header: %w", err)
	}

	// 序列化XML数据
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode NFO data: %w", err)
	}

	if logger != nil {
		logger.Debug("NFO file written successfully", zap.String("path", path))
	}

	return nil
}

// DetectNFOType 检测NFO文件类型
func DetectNFOType(path string) string {
	basename := strings.ToLower(filepath.Base(path))
	
	switch {
	case strings.Contains(basename, "tvshow"):
		return "tvshow"
	case strings.Contains(basename, "season"):
		return "season"
	case strings.Contains(basename, "episode"):
		return "episode"
	default:
		// 根据文件位置判断
		parent := strings.ToLower(filepath.Base(filepath.Dir(path)))
		if strings.Contains(parent, "season") || strings.Contains(parent, "s0") {
			return "episode"
		}
		return "movie"
	}
}

// NFOToMedia 将NFO数据转换为Media模型
func NFOToMedia(nfo *NFOData, mediaType string) models.Media {
	media := models.Media{
		Title:       nfo.Title,
		Description: nfo.Plot,
		Type:        mediaType,
		Poster:      nfo.Poster,
	}

	// 设置年份
	if nfo.Year > 0 {
		year := strconv.Itoa(nfo.Year)
		media.Year = &year
	}

	// 设置TMDB ID
	if nfo.TMDBID > 0 {
		media.TMDBID = &nfo.TMDBID
	}

	// 设置IMDB ID
	if nfo.IMDBID != "" {
		media.IMDBID = &nfo.IMDBID
	}

	// 设置评分
	if nfo.Rating > 0 {
		media.Vote = &nfo.Rating
	}

	// 设置运行时间
	if nfo.Runtime > 0 {
		media.Runtime = &nfo.Runtime
	}

	// 设置类型
	if len(nfo.Genres) > 0 {
		if data, err := json.Marshal(nfo.Genres); err == nil {
			media.Genres = string(data)
		}
	}

	return media
}

// 辅助函数
func getFloatValue(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}

func getStringValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

func parseJSONField(jsonStr string, dest interface{}) error {
	return json.Unmarshal([]byte(jsonStr), dest)
}