// Package douban 豆瓣API集成客户端
// 提供豆瓣电影/电视剧评分和信息获取功能
package douban

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/pkg/httpclient"

	"go.uber.org/zap"
)

const (
	// DefaultAPIURL 豆瓣API默认基础URL
	DefaultAPIURL = "https://api.douban.com/v2"
	// DefaultWebURL 豆瓣网页默认基础URL
	DefaultWebURL = "https://movie.douban.com"
)

// Client 豆瓣API客户端
type Client struct {
	httpClient *httpclient.Client
	logger     *zap.Logger
}

// NewClient 创建豆瓣客户端
// 返回: 豆瓣客户端实例
func NewClient() *Client {
	client := &Client{
		httpClient: httpclient.NewClient(httpclient.Options{
			BaseURL: DefaultAPIURL,
			Headers: map[string]string{
				"Accept":     "application/json",
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			Logger: logger.Logger,
		}),
		logger: logger.Logger,
	}

	return client
}

// SubjectDetails 豆瓣条目详情
type SubjectDetails struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	OriginalTitle string   `json:"original_title"`
	Year          string   `json:"year"`
	Subtype       string   `json:"subtype"` // movie或tv
	Rating        Rating   `json:"rating"`
	Summary       string   `json:"summary"`
	Directors     []Person `json:"directors"`
	Casts         []Person `json:"casts"`
	Genres        []string `json:"genres"`
	Countries     []string `json:"countries"`
	Images        Images   `json:"images"`
	AltTitle      string   `json:"alt_title"`
	PubDate       []string `json:"pubdate"`
	CollectCount  int      `json:"collect_count"`
	CommentsCount int      `json:"comments_count"`
	ReviewsCount  int      `json:"reviews_count"`
}

// Rating 评分
type Rating struct {
	Average float64        `json:"average"` // 平均分
	Max     int            `json:"max"`     // 最高分
	Min     int            `json:"min"`     // 最低分
	Stars   string         `json:"stars"`   // 星级
	Details map[string]int `json:"details"` // 评分分布
}

// Person 人物
type Person struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Alt       string `json:"alt"`
	AvatarURL string `json:"avatars"`
}

// Images 图片
type Images struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Count    int              `json:"count"`
	Start    int              `json:"start"`
	Total    int              `json:"total"`
	Subjects []SubjectDetails `json:"subjects"`
}

// GetMovieDetails 获取电影详情
// ctx: 上下文
// movieID: 电影ID
// 返回: 电影详情和错误信息
func (c *Client) GetMovieDetails(ctx context.Context, movieID string) (*SubjectDetails, error) {
	path := fmt.Sprintf("/movie/subject/%s", movieID)

	var result SubjectDetails
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get douban movie details failed",
			zap.String("movie_id", movieID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get douban movie details: %w", err)
	}

	c.logger.Info("Get douban movie details success",
		zap.String("movie_id", movieID),
		zap.String("title", result.Title),
		zap.Float64("rating", result.Rating.Average),
	)

	return &result, nil
}

// SearchMovie 搜索电影
// ctx: 上下文
// query: 搜索关键词
// start: 起始位置（分页用）
// count: 返回数量
// 返回: 搜索结果和错误信息
func (c *Client) SearchMovie(ctx context.Context, query string, start, count int) (*SearchResult, error) {
	if start < 0 {
		start = 0
	}
	if count <= 0 {
		count = 10
	}

	path := fmt.Sprintf("/movie/search?q=%s&start=%d&count=%d",
		url.QueryEscape(query), start, count)

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Search douban movie failed",
			zap.String("query", query),
			zap.Int("start", start),
			zap.Int("count", count),
			zap.Error(err),
		)
		return nil, fmt.Errorf("search douban movie: %w", err)
	}

	c.logger.Info("Search douban movie success",
		zap.String("query", query),
		zap.Int("start", start),
		zap.Int("count", count),
		zap.Int("total", result.Total),
	)

	return &result, nil
}

// ParseDoubanID 从豆瓣URL解析ID
// doubanURL: 豆瓣网址
// 返回: ID和错误信息
// 示例URL: https://movie.douban.com/subject/1292052/
func ParseDoubanID(doubanURL string) (string, error) {
	// 正则匹配豆瓣ID
	re := regexp.MustCompile(`subject/(\d+)`)
	matches := re.FindStringSubmatch(doubanURL)

	if len(matches) < 2 {
		return "", fmt.Errorf("invalid douban URL format")
	}

	return matches[1], nil
}

// GetRatingScore 获取评分分数
// subjectID: 条目ID
// 返回: 评分（0-10）和错误信息
func (c *Client) GetRatingScore(ctx context.Context, subjectID string) (float64, error) {
	details, err := c.GetMovieDetails(ctx, subjectID)
	if err != nil {
		return 0, err
	}

	return details.Rating.Average, nil
}

// ConvertToStandardRating 转换豆瓣评分到标准评分（0-10）
// doubanRating: 豆瓣评分（0-10）
// 返回: 标准评分（与TMDB等保持一致）
func ConvertToStandardRating(doubanRating float64) float64 {
	// 豆瓣评分本身就是0-10，无需转换
	return doubanRating
}

// IsValidDoubanID 验证豆瓣ID是否有效
// id: 豆瓣ID
// 返回: 是否有效
func IsValidDoubanID(id string) bool {
	if id == "" {
		return false
	}

	// 豆瓣ID应该是纯数字
	_, err := strconv.ParseInt(id, 10, 64)
	return err == nil
}
