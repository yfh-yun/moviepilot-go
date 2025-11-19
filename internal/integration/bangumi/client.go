// Package bangumi Bangumi API集成客户端
// 提供番组计划(Bangumi)的API访问功能，主要用于番剧信息获取
package bangumi

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
	// DefaultAPIURL Bangumi API默认基础URL
	DefaultAPIURL = "https://api.bgm.tv"
	// DefaultWebURL Bangumi网页默认基础URL
	DefaultWebURL = "https://bgm.tv"
)

// Client Bangumi API客户端
type Client struct {
	httpClient *httpclient.Client
	logger     *zap.Logger
}

// NewClient 创建Bangumi客户端
// 返回: Bangumi客户端实例
func NewClient() *Client {
	client := &Client{
		httpClient: httpclient.NewClient(httpclient.Options{
			BaseURL: DefaultAPIURL,
			Headers: map[string]string{
				"Accept":     "application/json",
				"User-Agent": "MoviePilot/1.0",
			},
			Logger: logger.Logger,
		}),
		logger: logger.Logger,
	}

	return client
}

// SubjectDetails 条目详情
type SubjectDetails struct {
	ID         int64      `json:"id"`
	Type       int        `json:"type"` // 1=书籍 2=动画 3=音乐 4=游戏 6=三次元
	Name       string     `json:"name"`
	NameCN     string     `json:"name_cn"`
	Summary    string     `json:"summary"`
	Air        string     `json:"air_date"`
	AirWeekday int        `json:"air_weekday"`
	Rating     Rating     `json:"rating"`
	Images     Images     `json:"images"`
	Eps        int        `json:"eps"`
	EpsCount   int        `json:"eps_count"`
	Collection Collection `json:"collection"`
}

// Rating 评分
type Rating struct {
	Total int            `json:"total"` // 评分人数
	Count map[string]int `json:"count"` // 评分分布
	Score float64        `json:"score"` // 平均分
}

// Images 图片
type Images struct {
	Small  string `json:"small"`
	Grid   string `json:"grid"`
	Large  string `json:"large"`
	Medium string `json:"medium"`
	Common string `json:"common"`
}

// Collection 收藏统计
type Collection struct {
	Wish    int `json:"wish"`    // 想看
	Collect int `json:"collect"` // 看过
	Doing   int `json:"doing"`   // 在看
	OnHold  int `json:"on_hold"` // 搁置
	Dropped int `json:"dropped"` // 抛弃
}

// SearchResult 搜索结果
type SearchResult struct {
	Results int          `json:"results"`
	List    []SearchItem `json:"list"`
}

// SearchItem 搜索条目
type SearchItem struct {
	ID      int64   `json:"id"`
	Type    int     `json:"type"`
	Name    string  `json:"name"`
	NameCN  string  `json:"name_cn"`
	Summary string  `json:"summary"`
	Images  Images  `json:"images"`
	Score   float64 `json:"score"`
	Rank    int     `json:"rank"`
}

// GetSubjectDetails 获取条目详情
// ctx: 上下文
// subjectID: 条目ID
// 返回: 条目详情和错误信息
func (c *Client) GetSubjectDetails(ctx context.Context, subjectID int64) (*SubjectDetails, error) {
	path := fmt.Sprintf("/v0/subjects/%d", subjectID)

	var result SubjectDetails
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Get bangumi subject details failed",
			zap.Int64("subject_id", subjectID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get bangumi subject details: %w", err)
	}

	c.logger.Info("Get bangumi subject details success",
		zap.Int64("subject_id", subjectID),
		zap.String("name", result.Name),
		zap.Float64("score", result.Rating.Score),
	)

	return &result, nil
}

// SearchSubject 搜索条目
// ctx: 上下文
// query: 搜索关键词
// subjectType: 条目类型（0=全部, 1=书籍, 2=动画, 3=音乐, 4=游戏, 6=三次元）
// 返回: 搜索结果和错误信息
func (c *Client) SearchSubject(ctx context.Context, query string, subjectType int) (*SearchResult, error) {
	path := fmt.Sprintf("/search/subject/%s?type=%d",
		url.QueryEscape(query), subjectType)

	var result SearchResult
	if err := c.httpClient.Get(ctx, path, &result); err != nil {
		c.logger.Error("Search bangumi subject failed",
			zap.String("query", query),
			zap.Int("type", subjectType),
			zap.Error(err),
		)
		return nil, fmt.Errorf("search bangumi subject: %w", err)
	}

	c.logger.Info("Search bangumi subject success",
		zap.String("query", query),
		zap.Int("type", subjectType),
		zap.Int("results", result.Results),
	)

	return &result, nil
}

// ParseBangumiID 从Bangumi URL解析ID
// bangumiURL: Bangumi网址
// 返回: ID和错误信息
// 示例URL: https://bgm.tv/subject/12345
func ParseBangumiID(bangumiURL string) (int64, error) {
	// 正则匹配Bangumi ID
	re := regexp.MustCompile(`subject/(\d+)`)
	matches := re.FindStringSubmatch(bangumiURL)

	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid bangumi URL format")
	}

	id, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse bangumi ID: %w", err)
	}

	return id, nil
}

// GetRatingScore 获取评分分数
// subjectID: 条目ID
// 返回: 评分（0-10）和错误信息
func (c *Client) GetRatingScore(ctx context.Context, subjectID int64) (float64, error) {
	details, err := c.GetSubjectDetails(ctx, subjectID)
	if err != nil {
		return 0, err
	}

	return details.Rating.Score, nil
}

// ConvertToStandardRating 转换Bangumi评分到标准评分（0-10）
// bangumiRating: Bangumi评分（0-10）
// 返回: 标准评分（与TMDB等保持一致）
func ConvertToStandardRating(bangumiRating float64) float64 {
	// Bangumi评分本身就是0-10，无需转换
	return bangumiRating
}

// IsAnime 判断条目是否为动画
// subjectType: 条目类型
// 返回: 是否为动画
func IsAnime(subjectType int) bool {
	return subjectType == 2
}

// IsValidBangumiID 验证Bangumi ID是否有效
// id: Bangumi ID
// 返回: 是否有效
func IsValidBangumiID(id int64) bool {
	return id > 0
}
