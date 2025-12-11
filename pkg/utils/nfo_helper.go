package utils

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// NFOType 定义NFO文件类型
type NFOType string

const (
	NFOTypeMovie   NFOType = "movie"
	NFOTypeTVShow  NFOType = "tvshow"
	NFOTypeEpisode NFOType = "episode"
)

// Actor 演员信息
type Actor struct {
	XMLName xml.Name `xml:"actor"`
	Name    string   `xml:"name"`
	Role    string   `xml:"role"`
	Order   int      `xml:"order,attr,omitempty"`
	Thumb   string   `xml:"thumb,omitempty"`
}

// Season 季信息
type Season struct {
	XMLName  xml.Name `xml:"season"`
	Number   int      `xml:"number"`
	Name     string   `xml:"name,omitempty"`
	Thumb    string   `xml:"thumb,omitempty"`
	Overview string   `xml:"overview,omitempty"`
}

// NFOData 电影NFO数据
type NFOData struct {
	XMLName        xml.Name       `xml:"movie"`
	Title          string         `xml:"title"`
	OriginalTitle  string         `xml:"originaltitle,omitempty"`
	SortTitle      string         `xml:"sorttitle,omitempty"`
	Year           int            `xml:"year"`
	Rating         float64        `xml:"rating,omitempty"`
	Votes          int            `xml:"votes,omitempty"`
	Top250         int            `xml:"top250,omitempty"`
	Outline        string         `xml:"outline,omitempty"`
	Plot           string         `xml:"plot,omitempty"`
	Tagline        string         `xml:"tagline,omitempty"`
	Runtime        int            `xml:"runtime,omitempty"`
	MPAA           string         `xml:"mpaa,omitempty"`
	PlayCount      int            `xml:"playcount,omitempty"`
	LastPlayed     string         `xml:"lastplayed,omitempty"`
	ID             string         `xml:"id,omitempty"`
	IMDBID         string         `xml:"imdbid,omitempty"`
	TMDBID         int            `xml:"tmdbid,omitempty"`
	UniqueID       []UniqueID     `xml:"uniqueid,omitempty"`
	Genre          []string       `xml:"genre,omitempty"`
	Country        []string       `xml:"country,omitempty"`
	Credits        string         `xml:"credits,omitempty"`
	Director       []string       `xml:"director,omitempty"`
	Writer         []string       `xml:"writer,omitempty"`
	Actor          []Actor        `xml:"actor,omitempty"`
	Producer       []string       `xml:"producer,omitempty"`
	Music          []string       `xml:"music,omitempty"`
	Studio         []string       `xml:"studio,omitempty"`
	Trailer        string         `xml:"trailer,omitempty"`
	Premiered      string         `xml:"premiered,omitempty"`
	Released       string         `xml:"released,omitempty"`
	DVDReleaseDate string         `xml:"dvdrelease_date,omitempty"`
	Status         string         `xml:"status,omitempty"`
	Code           string         `xml:"code,omitempty"`
	ShowLink       []string       `xml:"showlink,omitempty"`
	Album          string         `xml:"album,omitempty"`
	Artist         []string       `xml:"artist,omitempty"`
	Track          int            `xml:"track,omitempty"`
	Disc           int            `xml:"disc,omitempty"`
	Tag            []string       `xml:"tag,omitempty"`
	Set            *Set           `xml:"set,omitempty"`
	Thumb          []Thumb        `xml:"thumb,omitempty"`
	Fanart         *Fanart        `xml:"fanart,omitempty"`
	FileInfo       *FileInfo      `xml:"fileinfo,omitempty"`
	StreamDetails  *StreamDetails `xml:"streamdetails,omitempty"`
}

// TVShowNFOData 电视剧NFO数据
type TVShowNFOData struct {
	XMLName       xml.Name       `xml:"tvshow"`
	Title         string         `xml:"title"`
	OriginalTitle string         `xml:"originaltitle,omitempty"`
	SortTitle     string         `xml:"sorttitle,omitempty"`
	Year          int            `xml:"year"`
	Rating        float64        `xml:"rating,omitempty"`
	Votes         int            `xml:"votes,omitempty"`
	Top250        int            `xml:"top250,omitempty"`
	Outline       string         `xml:"outline,omitempty"`
	Plot          string         `xml:"plot,omitempty"`
	Tagline       string         `xml:"tagline,omitempty"`
	Runtime       int            `xml:"runtime,omitempty"`
	MPAA          string         `xml:"mpaa,omitempty"`
	PlayCount     int            `xml:"playcount,omitempty"`
	LastPlayed    string         `xml:"lastplayed,omitempty"`
	ID            string         `xml:"id,omitempty"`
	IMDBID        string         `xml:"imdbid,omitempty"`
	TMDBID        int            `xml:"tmdbid,omitempty"`
	TVDBID        int            `xml:"tvdbid,omitempty"`
	UniqueID      []UniqueID     `xml:"uniqueid,omitempty"`
	Genre         []string       `xml:"genre,omitempty"`
	Country       []string       `xml:"country,omitempty"`
	Credits       string         `xml:"credits,omitempty"`
	Director      []string       `xml:"director,omitempty"`
	Writer        []string       `xml:"writer,omitempty"`
	Actor         []Actor        `xml:"actor,omitempty"`
	Producer      []string       `xml:"producer,omitempty"`
	Music         []string       `xml:"music,omitempty"`
	Studio        []string       `xml:"studio,omitempty"`
	Trailer       string         `xml:"trailer,omitempty"`
	Premiered     string         `xml:"premiered,omitempty"`
	FirstAired    string         `xml:"firstaired,omitempty"`
	Status        string         `xml:"status,omitempty"`
	EpisodeGuide  string         `xml:"episodeguide,omitempty"`
	Season        []Season       `xml:"season,omitempty"`
	Tag           []string       `xml:"tag,omitempty"`
	Set           *Set           `xml:"set,omitempty"`
	Thumb         []Thumb        `xml:"thumb,omitempty"`
	Fanart        *Fanart        `xml:"fanart,omitempty"`
	FileInfo      *FileInfo      `xml:"fileinfo,omitempty"`
	StreamDetails *StreamDetails `xml:"streamdetails,omitempty"`
}

// EpisodeNFOData 剧集NFO数据
type EpisodeNFOData struct {
	XMLName        xml.Name       `xml:"episodedetails"`
	Title          string         `xml:"title"`
	OriginalTitle  string         `xml:"originaltitle,omitempty"`
	ShowTitle      string         `xml:"showtitle,omitempty"`
	SortTitle      string         `xml:"sorttitle,omitempty"`
	Season         int            `xml:"season"`
	Episode        int            `xml:"episode"`
	DisplaySeason  int            `xml:"displayseason,omitempty"`
	DisplayEpisode int            `xml:"displayepisode,omitempty"`
	AbsoluteNumber int            `xml:"absolute_number,omitempty"`
	Rating         float64        `xml:"rating,omitempty"`
	Votes          int            `xml:"votes,omitempty"`
	Outline        string         `xml:"outline,omitempty"`
	Plot           string         `xml:"plot,omitempty"`
	Tagline        string         `xml:"tagline,omitempty"`
	Runtime        int            `xml:"runtime,omitempty"`
	MPAA           string         `xml:"mpaa,omitempty"`
	PlayCount      int            `xml:"playcount,omitempty"`
	LastPlayed     string         `xml:"lastplayed,omitempty"`
	ID             string         `xml:"id,omitempty"`
	IMDBID         string         `xml:"imdbid,omitempty"`
	TMDBID         int            `xml:"tmdbid,omitempty"`
	TVDBID         int            `xml:"tvdbid,omitempty"`
	UniqueID       []UniqueID     `xml:"uniqueid,omitempty"`
	Genre          []string       `xml:"genre,omitempty"`
	Country        []string       `xml:"country,omitempty"`
	Credits        string         `xml:"credits,omitempty"`
	Director       []string       `xml:"director,omitempty"`
	Writer         []string       `xml:"writer,omitempty"`
	Actor          []Actor        `xml:"actor,omitempty"`
	Producer       []string       `xml:"producer,omitempty"`
	Music          []string       `xml:"music,omitempty"`
	Studio         []string       `xml:"studio,omitempty"`
	Trailer        string         `xml:"trailer,omitempty"`
	Premiered      string         `xml:"premiered,omitempty"`
	FirstAired     string         `xml:"firstaired,omitempty"`
	DVDReleaseDate string         `xml:"dvdrelease_date,omitempty"`
	Status         string         `xml:"status,omitempty"`
	Tag            []string       `xml:"tag,omitempty"`
	Thumb          []Thumb        `xml:"thumb,omitempty"`
	Fanart         *Fanart        `xml:"fanart,omitempty"`
	FileInfo       *FileInfo      `xml:"fileinfo,omitempty"`
	StreamDetails  *StreamDetails `xml:"streamdetails,omitempty"`
}

// UniqueID 唯一标识符
type UniqueID struct {
	Type    string `xml:"type,attr"`
	Default bool   `xml:"default,attr,omitempty"`
	Value   string `xml:"-,chardata"`
}

// Set 合集信息
type Set struct {
	Name  string `xml:"name"`
	Thumb string `xml:"thumb,omitempty"`
}

// Thumb 缩略图信息
type Thumb struct {
	Aspect  string `xml:"aspect,attr,omitempty"`
	Preview string `xml:"preview,attr,omitempty"`
	Value   string `xml:"-,chardata"`
}

// Fanart 粉丝艺术图
type Fanart struct {
	XMLName xml.Name `xml:"fanart"`
	Thumb   []Thumb  `xml:"thumb"`
}

// FileInfo 文件信息
type FileInfo struct {
	XMLName       xml.Name       `xml:"fileinfo"`
	StreamDetails *StreamDetails `xml:"streamdetails"`
}

// StreamDetails 流详情
type StreamDetails struct {
	XMLName  xml.Name         `xml:"streamdetails"`
	Video    []VideoStream    `xml:"video"`
	Audio    []AudioStream    `xml:"audio"`
	Subtitle []SubtitleStream `xml:"subtitle"`
}

// VideoStream 视频流信息
type VideoStream struct {
	Codec      string  `xml:"codec"`
	Aspect     float64 `xml:"aspect"`
	Width      int     `xml:"width"`
	Height     int     `xml:"height"`
	Duration   float64 `xml:"duration"`
	Bitrate    int     `xml:"bitrate,omitempty"`
	BitDepth   int     `xml:"bitdepth,omitempty"`
	ColorSpace string  `xml:"colorspace,omitempty"`
	FrameRate  float64 `xml:"framerate"`
	StereoMode string  `xml:"stereomode,omitempty"`
	ScanType   string  `xml:"scantype,omitempty"`
	Language   string  `xml:"language,omitempty"`
	Default    bool    `xml:"default,omitempty"`
	Forced     bool    `xml:"forced,omitempty"`
}

// AudioStream 音频流信息
type AudioStream struct {
	Codec      string `xml:"codec"`
	Language   string `xml:"language"`
	Channels   int    `xml:"channels"`
	Bitrate    int    `xml:"bitrate,omitempty"`
	SampleRate int    `xml:"samplerate,omitempty"`
	BitDepth   int    `xml:"bitdepth,omitempty"`
	Default    bool   `xml:"default,omitempty"`
	Forced     bool   `xml:"forced,omitempty"`
}

// SubtitleStream 字幕流信息
type SubtitleStream struct {
	Language string `xml:"language"`
	Format   string `xml:"format,omitempty"`
	Default  bool   `xml:"default,omitempty"`
	Forced   bool   `xml:"forced,omitempty"`
}

// NFOHelper NFO文件处理助手
type NFOHelper struct {
	logger *zap.Logger
}

// NewNFOHelper 创建NFO助手实例
func NewNFOHelper(logger *zap.Logger) *NFOHelper {
	return &NFOHelper{
		logger: logger,
	}
}

// WriteNFO 写入NFO文件
func (h *NFOHelper) WriteNFO(filePath string, data any) error {
	// 创建目录
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.logger.Error("创建NFO目录失败", zap.Error(err))
		return err
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("创建NFO文件失败", zap.Error(err))
		return err
	}
	defer file.Close()

	// 写入XML头
	if _, err := file.WriteString(xml.Header); err != nil {
		h.logger.Error("写入XML头失败", zap.Error(err))
		return err
	}

	// 编码XML
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(data); err != nil {
		h.logger.Error("编码XML失败", zap.Error(err))
		return err
	}

	h.logger.Info("NFO文件写入成功", zap.String("file_path", filePath))
	return nil
}

// ReadNFO 读取NFO文件
func (h *NFOHelper) ReadNFO(filePath string) (any, error) {
	// 检测NFO类型
	nfoType, err := h.DetectNFOType(filePath)
	if err != nil {
		return nil, err
	}

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		h.logger.Error("读取NFO文件失败", zap.Error(err))
		return nil, err
	}

	// 解码XML
	var result any
	switch nfoType {
	case NFOTypeMovie:
		var movie NFOData
		if err := xml.Unmarshal(data, &movie); err != nil {
			return nil, err
		}
		result = movie
	case NFOTypeTVShow:
		var tvshow TVShowNFOData
		if err := xml.Unmarshal(data, &tvshow); err != nil {
			return nil, err
		}
		result = tvshow
	case NFOTypeEpisode:
		var episode EpisodeNFOData
		if err := xml.Unmarshal(data, &episode); err != nil {
			return nil, err
		}
		result = episode
	default:
		return nil, fmt.Errorf("未知的NFO类型: %s", nfoType)
	}

	h.logger.Info("NFO文件读取成功", zap.String("file_path", filePath))
	return result, nil
}

// DetectNFOType 检测NFO文件类型
func (h *NFOHelper) DetectNFOType(filePath string) (NFOType, error) {
	// 从文件名检测
	fileName := filepath.Base(filePath)
	if strings.Contains(fileName, "tvshow") {
		return NFOTypeTVShow, nil
	} else if strings.Contains(fileName, "season") || strings.Contains(fileName, "episode") {
		return NFOTypeEpisode, nil
	} else {
		// 读取文件内容检测
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		content := string(data)
		if strings.Contains(content, "<tvshow>") {
			return NFOTypeTVShow, nil
		} else if strings.Contains(content, "<episodedetails>") {
			return NFOTypeEpisode, nil
		} else {
			return NFOTypeMovie, nil
		}
	}
}

// WriteMovieNFO 写入电影NFO文件
func (h *NFOHelper) WriteMovieNFO(filePath string, data NFOData) error {
	return h.WriteNFO(filePath, data)
}

// WriteTVShowNFO 写入电视剧NFO文件
func (h *NFOHelper) WriteTVShowNFO(filePath string, data TVShowNFOData) error {
	return h.WriteNFO(filePath, data)
}

// WriteEpisodeNFO 写入剧集NFO文件
func (h *NFOHelper) WriteEpisodeNFO(filePath string, data EpisodeNFOData) error {
	return h.WriteNFO(filePath, data)
}

// ReadMovieNFO 读取电影NFO文件
func (h *NFOHelper) ReadMovieNFO(filePath string) (NFOData, error) {
	result, err := h.ReadNFO(filePath)
	if err != nil {
		return NFOData{}, err
	}
	if movieData, ok := result.(NFOData); ok {
		return movieData, nil
	}
	return NFOData{}, fmt.Errorf("NFO文件不是电影类型")
}

// ReadTVShowNFO 读取电视剧NFO文件
func (h *NFOHelper) ReadTVShowNFO(filePath string) (TVShowNFOData, error) {
	result, err := h.ReadNFO(filePath)
	if err != nil {
		return TVShowNFOData{}, err
	}
	if tvshowData, ok := result.(TVShowNFOData); ok {
		return tvshowData, nil
	}
	return TVShowNFOData{}, fmt.Errorf("NFO文件不是电视剧类型")
}

// ReadEpisodeNFO 读取剧集NFO文件
func (h *NFOHelper) ReadEpisodeNFO(filePath string) (EpisodeNFOData, error) {
	result, err := h.ReadNFO(filePath)
	if err != nil {
		return EpisodeNFOData{}, err
	}
	if episodeData, ok := result.(EpisodeNFOData); ok {
		return episodeData, nil
	}
	return EpisodeNFOData{}, fmt.Errorf("NFO文件不是剧集类型")
}
