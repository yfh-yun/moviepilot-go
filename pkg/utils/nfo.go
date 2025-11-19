package utils

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NFOReader NFO文件读取器
type NFOReader struct {
	XMLFilePath string
	Root        *NFORoot
}

// NFORoot NFO文件根元素
type NFORoot struct {
	XMLName xml.Name `xml:"movie"`
	Title   string   `xml:"title"`
	Year    string   `xml:"year"`
	Plot    string   `xml:"plot"`
	Runtime string   `xml:"runtime"`
	Rating  string   `xml:"rating"`
	Votes   string   `xml:"votes"`
	Director string  `xml:"director"`
	Genre   []string `xml:"genre"`
	Actor   []Actor  `xml:"actor"`
}

// Actor 演员信息
type Actor struct {
	Name string `xml:"name"`
	Role string `xml:"role"`
}

// TVShowRoot 电视剧NFO根元素
type TVShowRoot struct {
	XMLName xml.Name `xml:"tvshow"`
	Title   string   `xml:"title"`
	Plot    string   `xml:"plot"`
	Runtime string   `xml:"runtime"`
	Rating  string   `xml:"rating"`
	Votes   string   `xml:"votes"`
	Genre   []string `xml:"genre"`
	Actor   []Actor  `xml:"actor"`
	Season  []Season `xml:"season"`
}

// Season 季信息
type Season struct {
	Number string `xml:"num,attr"`
	Name   string `xml:"name"`
}

// EpisodeRoot 剧集NFO根元素
type EpisodeRoot struct {
	XMLName   xml.Name `xml:"episodedetails"`
	Title     string   `xml:"title"`
	Season    string   `xml:"season"`
	Episode   string   `xml:"episode"`
	Aired     string   `xml:"aired"`
	Plot      string   `xml:"plot"`
	Runtime   string   `xml:"runtime"`
	Rating    string   `xml:"rating"`
	Director  string   `xml:"director"`
	Genre     []string `xml:"genre"`
	Actor     []Actor  `xml:"actor"`
}

// NewNFOReader 创建NFO读取器实例
func NewNFOReader(xmlFilePath string) (*NFOReader, error) {
	if !filepath.IsAbs(xmlFilePath) {
		return nil, fmt.Errorf("file path must be absolute: %s", xmlFilePath)
	}

	if _, err := os.Stat(xmlFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", xmlFilePath)
	}

	return &NFOReader{
		XMLFilePath: xmlFilePath,
	}, nil
}

// Load 加载NFO文件
func (nfo *NFOReader) Load() error {
	data, err := os.ReadFile(nfo.XMLFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// 尝试解析为电影NFO
	var movieRoot NFORoot
	if err := xml.Unmarshal(data, &movieRoot); err == nil && movieRoot.Title != "" {
		nfo.Root = &NFORoot{
			Title:    movieRoot.Title,
			Year:     movieRoot.Year,
			Plot:     movieRoot.Plot,
			Runtime:  movieRoot.Runtime,
			Rating:   movieRoot.Rating,
			Votes:    movieRoot.Votes,
			Director: movieRoot.Director,
			Genre:    movieRoot.Genre,
			Actor:    movieRoot.Actor,
		}
		return nil
	}

	// 尝试解析为电视剧NFO
	var tvShowRoot TVShowRoot
	if err := xml.Unmarshal(data, &tvShowRoot); err == nil && tvShowRoot.Title != "" {
		nfo.Root = &NFORoot{
			Title:   tvShowRoot.Title,
			Plot:    tvShowRoot.Plot,
			Runtime: tvShowRoot.Runtime,
			Rating:  tvShowRoot.Rating,
			Votes:   tvShowRoot.Votes,
			Genre:   tvShowRoot.Genre,
			Actor:   tvShowRoot.Actor,
		}
		return nil
	}

	// 尝试解析为剧集NFO
	var episodeRoot EpisodeRoot
	if err := xml.Unmarshal(data, &episodeRoot); err == nil && episodeRoot.Title != "" {
		nfo.Root = &NFORoot{
			Title:    episodeRoot.Title,
			Plot:     episodeRoot.Plot,
			Runtime:  episodeRoot.Runtime,
			Rating:   episodeRoot.Rating,
			Director: episodeRoot.Director,
			Genre:    episodeRoot.Genre,
			Actor:    episodeRoot.Actor,
		}
		return nil
	}

	return fmt.Errorf("failed to parse NFO file: unsupported format")
}

// GetElementValue 获取元素值
func (nfo *NFOReader) GetElementValue(elementPath string) string {
	if nfo.Root == nil {
		return ""
	}

	parts := strings.Split(elementPath, ".")
	if len(parts) == 0 {
		return ""
	}

	switch parts[0] {
	case "title":
		return nfo.Root.Title
	case "year":
		return nfo.Root.Year
	case "plot":
		return nfo.Root.Plot
	case "runtime":
		return nfo.Root.Runtime
	case "rating":
		return nfo.Root.Rating
	case "votes":
		return nfo.Root.Votes
	case "director":
		return nfo.Root.Director
	case "genre":
		if len(parts) > 1 {
			index := 0
			if len(parts) > 2 {
				if idx, err := parseIndex(parts[2]); err == nil {
					index = idx
				}
			}
			if index < len(nfo.Root.Genre) {
				return nfo.Root.Genre[index]
			}
		} else if len(nfo.Root.Genre) > 0 {
			return strings.Join(nfo.Root.Genre, ", ")
		}
	case "actor":
		if len(parts) > 1 {
			index := 0
			if len(parts) > 2 {
				if idx, err := parseIndex(parts[2]); err == nil {
					index = idx
				}
			}
			if index < len(nfo.Root.Actor) {
				actor := nfo.Root.Actor[index]
				if len(parts) > 1 {
					switch parts[1] {
					case "name":
						return actor.Name
					case "role":
						return actor.Role
					}
				}
				return actor.Name
			}
		}
	}

	return ""
}

// GetElements 获取元素列表
func (nfo *NFOReader) GetElements(elementPath string) []string {
	if nfo.Root == nil {
		return nil
	}

	parts := strings.Split(elementPath, ".")
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "genre":
		return nfo.Root.Genre
	case "actor":
		var actors []string
		for _, actor := range nfo.Root.Actor {
			if len(parts) > 1 {
				switch parts[1] {
				case "name":
					actors = append(actors, actor.Name)
				case "role":
					actors = append(actors, actor.Role)
				}
			} else {
				actors = append(actors, actor.Name)
			}
		}
		return actors
	}

	return nil
}

// GetTitle 获取标题
func (nfo *NFOReader) GetTitle() string {
	if nfo.Root == nil {
		return ""
	}
	return nfo.Root.Title
}

// GetYear 获取年份
func (nfo *NFOReader) GetYear() string {
	if nfo.Root == nil {
		return ""
	}
	return nfo.Root.Year
}

// GetPlot 获取剧情简介
func (nfo *NFOReader) GetPlot() string {
	if nfo.Root == nil {
		return ""
	}
	return nfo.Root.Plot
}

// GetGenres 获取类型列表
func (nfo *NFOReader) GetGenres() []string {
	if nfo.Root == nil {
		return nil
	}
	return nfo.Root.Genre
}

// GetActors 获取演员列表
func (nfo *NFOReader) GetActors() []string {
	if nfo.Root == nil {
		return nil
	}
	var actors []string
	for _, actor := range nfo.Root.Actor {
		actors = append(actors, actor.Name)
	}
	return actors
}

// GetDirector 获取导演
func (nfo *NFOReader) GetDirector() string {
	if nfo.Root == nil {
		return ""
	}
	return nfo.Root.Director
}

// GetRating 获取评分
func (nfo *NFOReader) GetRating() string {
	if nfo.Root == nil {
		return ""
	}
	return nfo.Root.Rating
}

// parseIndex 解析索引
func parseIndex(indexStr string) (int, error) {
	// 移除方括号
	indexStr = strings.Trim(indexStr, "[]")
	var index int
	_, err := fmt.Sscanf(indexStr, "%d", &index)
	return index, err
}

// IsNFOFile 检查文件是否为NFO文件
func IsNFOFile(filePath string) bool {
	return strings.ToLower(filepath.Ext(filePath)) == ".nfo"
}

// FindNFOFile 在目录中查找NFO文件
func FindNFOFile(dirPath string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && IsNFOFile(entry.Name()) {
			return filepath.Join(dirPath, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no NFO file found in directory: %s", dirPath)
}