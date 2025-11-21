package fanart

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"moviepilot-go/internal/infrastructure/config"
	"moviepilot-go/pkg/logger"
	"moviepilot-go/pkg/httpclient"
)

// Client Fanart API客户端
type Client struct {
	httpClient *httpclient.Client
	baseURL    string
	apiKey     string
	cache      *Cache
	logger     *logger.Logger
}

// ImageType 图片类型
type ImageType string

const (
	MoviePoster     ImageType = "movieposter"
	MovieBackground ImageType = "moviebackground"
	MovieLogo       ImageType = "movielogo"
	MovieDisc       ImageType = "moviedisc"
	TVPoster        ImageType = "tvposter"
	TVBackground    ImageType = "tvbackground"
	TVLogo          ImageType = "tvlogo"
	TVSeasonPoster  ImageType = "seasonposter"
	TVSeasonBanner  ImageType = "seasonbanner"
	TVSeasonThumb   ImageType = "seasonthumb"
)

// Image 图片信息
type Image struct {
	ID    string    `json:"id"`
	URL   string    `json:"url"`
	Lang  string    `json:"lang"`
	Likes string    `json:"likes"`
	Type  ImageType `json:"type"`
}

// MovieImages 电影图片集
type MovieImages struct {
	TMDBID      string  `json:"tmdb_id"`
	IMDBID      string  `json:"imdb_id"`
	Title       string  `json:"title"`
	Posters     []Image `json:"posters"`
	Backgrounds []Image `json:"backgrounds"`
	Logos       []Image `json:"logos"`
	Discs       []Image `json:"discs"`
}

// TVImages 电视剧图片集
type TVImages struct {
	TVDBID      string                   `json:"tvdb_id"`
	IMDBID      string                   `json:"imdb_id"`
	Title       string                   `json:"title"`
	Posters     []Image                  `json:"posters"`
	Backgrounds []Image                  `json:"backgrounds"`
	Logos       []Image                  `json:"logos"`
	Seasons     map[string]*SeasonImages `json:"seasons"`
}

// SeasonImages 季图片集
type SeasonImages struct {
	Season  int     `json:"season"`
	Posters []Image `json:"posters"`
	Banners []Image `json:"banners"`
	Thumbs  []Image `json:"thumbs"`
}

// NewClient 创建Fanart客户端
func NewClient(cfg *config.Config) *Client {
	httpClient := httpclient.NewClient(&httpclient.Config{
		BaseURL:   "https://webservice.fanart.tv/v3/",
		Timeout:   30 * time.Second,
		UserAgent: "MoviePilot/1.0",
	})

	return &Client{
		httpClient: httpClient,
		baseURL:    "https://webservice.fanart.tv/v3/",
		apiKey:     cfg.Fanart.APIKey,
		cache:      NewCache(24*time.Hour, 1000),
		logger:     logger.NewLogger("fanart"),
	}
}

// GetMovieImages 获取电影图片
func (c *Client) GetMovieImages(ctx context.Context, tmdbID int) (*MovieImages, error) {
	// 检查缓存
	if cached, found := c.cache.GetMovieImages(tmdbID); found {
		return cached, nil
	}

	url := fmt.Sprintf("%smovies/%d?api_key=%s", c.baseURL, tmdbID, c.apiKey)

	resp, err := c.httpClient.Get(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("获取电影图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	movieImages := c.parseMovieImages(result, tmdbID)

	// 缓存结果
	c.cache.SetMovieImages(tmdbID, movieImages)

	return movieImages, nil
}

// GetTVImages 获取电视剧图片
func (c *Client) GetTVImages(ctx context.Context, tvdbID int) (*TVImages, error) {
	// 检查缓存
	if cached, found := c.cache.GetTVImages(tvdbID); found {
		return cached, nil
	}

	url := fmt.Sprintf("%stv/%d?api_key=%s", c.baseURL, tvdbID, c.apiKey)

	resp, err := c.httpClient.Get(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("获取电视剧图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	tvImages := c.parseTVImages(result, tvdbID)

	// 缓存结果
	c.cache.SetTVImages(tvdbID, tvImages)

	return tvImages, nil
}

// parseMovieImages 解析电影图片数据
func (c *Client) parseMovieImages(data map[string]interface{}, tmdbID int) *MovieImages {
	movie := &MovieImages{
		TMDBID: fmt.Sprintf("%d", tmdbID),
		Title:  c.getString(data, "name"),
		IMDBID: c.getString(data, "imdb_id"),
	}

	// 解析海报
	if posters, ok := data["movieposter"].([]interface{}); ok {
		movie.Posters = c.parseImages(posters, MoviePoster)
	}

	// 解析背景图
	if backgrounds, ok := data["moviebackground"].([]interface{}); ok {
		movie.Backgrounds = c.parseImages(backgrounds, MovieBackground)
	}

	// 解析Logo
	if logos, ok := data["movielogo"].([]interface{}); ok {
		movie.Logos = c.parseImages(logos, MovieLogo)
	}

	// 解析光盘封面
	if discs, ok := data["moviedisc"].([]interface{}); ok {
		movie.Discs = c.parseImages(discs, MovieDisc)
	}

	return movie
}

// parseTVImages 解析电视剧图片数据
func (c *Client) parseTVImages(data map[string]interface{}, tvdbID int) *TVImages {
	tv := &TVImages{
		TVDBID:  fmt.Sprintf("%d", tvdbID),
		Title:   c.getString(data, "name"),
		IMDBID:  c.getString(data, "imdb_id"),
		Seasons: make(map[string]*SeasonImages),
	}

	// 解析海报
	if posters, ok := data["tvposter"].([]interface{}); ok {
		tv.Posters = c.parseImages(posters, TVPoster)
	}

	// 解析背景图
	if backgrounds, ok := data["showbackground"].([]interface{}); ok {
		tv.Backgrounds = c.parseImages(backgrounds, TVBackground)
	}

	// 解析Logo
	if logos, ok := data["hdtvlogo"].([]interface{}); ok {
		tv.Logos = c.parseImages(logos, TVLogo)
	}

	// 解析季图片
	c.parseSeasonImages(data, tv)

	return tv
}

// parseSeasonImages 解析季图片数据
func (c *Client) parseSeasonImages(data map[string]interface{}, tv *TVImages) {
	// 季海报
	if seasonPosters, ok := data["seasonposter"].(map[string]interface{}); ok {
		for seasonNum, images := range seasonPosters {
			if seasonImages, ok := images.([]interface{}); ok {
				if tv.Seasons[seasonNum] == nil {
					tv.Seasons[seasonNum] = &SeasonImages{}
				}
				tv.Seasons[seasonNum].Posters = c.parseImages(seasonImages, TVSeasonPoster)
			}
		}
	}

	// 季横幅
	if seasonBanners, ok := data["seasonbanner"].(map[string]interface{}); ok {
		for seasonNum, images := range seasonBanners {
			if seasonImages, ok := images.([]interface{}); ok {
				if tv.Seasons[seasonNum] == nil {
					tv.Seasons[seasonNum] = &SeasonImages{}
				}
				tv.Seasons[seasonNum].Banners = c.parseImages(seasonImages, TVSeasonBanner)
			}
		}
	}

	// 季缩略图
	if seasonThumbs, ok := data["seasonthumb"].(map[string]interface{}); ok {
		for seasonNum, images := range seasonThumbs {
			if seasonImages, ok := images.([]interface{}); ok {
				if tv.Seasons[seasonNum] == nil {
					tv.Seasons[seasonNum] = &SeasonImages{}
				}
				tv.Seasons[seasonNum].Thumbs = c.parseImages(seasonImages, TVSeasonThumb)
			}
		}
	}
}

// parseImages 解析图片数组
func (c *Client) parseImages(images []interface{}, imageType ImageType) []Image {
	var result []Image

	for _, img := range images {
		if imgMap, ok := img.(map[string]interface{}); ok {
			image := Image{
				ID:    c.getString(imgMap, "id"),
				URL:   c.getString(imgMap, "url"),
				Lang:  c.getString(imgMap, "lang"),
				Likes: c.getString(imgMap, "likes"),
				Type:  imageType,
			}
			result = append(result, image)
		}
	}

	return result
}

// getString 安全获取字符串值
func (c *Client) getString(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// FilterImages 根据质量和语言筛选图片
func (c *Client) FilterImages(images []Image, qualityThreshold int, preferredLang string) []Image {
	var filtered []Image

	for _, img := range images {
		// 转换likes为数字
		likes := 0
		fmt.Sscanf(img.Likes, "%d", &likes)

		// 筛选质量
		if likes < qualityThreshold {
			continue
		}

		// 优选指定语言
		if img.Lang == preferredLang {
			filtered = append(filtered, img)
		}
	}

	// 如果没有指定语言的图片，返回所有符合条件的图片
	if len(filtered) == 0 {
		for _, img := range images {
			likes := 0
			fmt.Sscanf(img.Likes, "%d", &likes)
			if likes >= qualityThreshold {
				filtered = append(filtered, img)
			}
		}
	}

	return filtered
}
