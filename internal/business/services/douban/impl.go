// Package douban 豆瓣服务实现
package douban

import (
	"context"
	"fmt"
	"strconv"

	"moviepilot-go/internal/integration/douban"
	"go.uber.org/zap"
)

// impl 豆瓣服务实现
type impl struct {
	client *douban.Client
	logger *zap.Logger
}

// NewService 创建豆瓣服务实例
func NewService(logger *zap.Logger) Service {
	return &impl{
		client: douban.NewClient(),
		logger: logger,
	}
}

// GetMovieInfo 获取电影信息
func (s *impl) GetMovieInfo(doubanID string) (*douban.SubjectDetails, error) {
	if !douban.IsValidDoubanID(doubanID) {
		return nil, fmt.Errorf("invalid douban ID: %s", doubanID)
	}

	ctx := context.Background()
	return s.client.GetMovieDetails(ctx, doubanID)
}

// GetTVInfo 获取电视剧信息
func (s *impl) GetTVInfo(doubanID string) (*douban.SubjectDetails, error) {
	if !douban.IsValidDoubanID(doubanID) {
		return nil, fmt.Errorf("invalid douban ID: %s", doubanID)
	}

	// 豆瓣API中电影和电视剧使用相同的接口
	ctx := context.Background()
	details, err := s.client.GetMovieDetails(ctx, doubanID)
	if err != nil {
		return nil, err
	}

	// 确保返回的是电视剧类型
	if details.Subtype != "tv" {
		return nil, fmt.Errorf("not a TV series: %s", doubanID)
	}

	return details, nil
}

// GetPersonInfo 获取人物信息
func (s *impl) GetPersonInfo(doubanID string) (*douban.Person, error) {
	if !douban.IsValidDoubanID(doubanID) {
		return nil, fmt.Errorf("invalid douban ID: %s", doubanID)
	}

	// TODO: 实现获取人物信息的逻辑
	// 目前返回模拟数据
	return &douban.Person{
		ID:        doubanID,
		Name:      "Person " + doubanID,
		Alt:       "",
		AvatarURL: "",
	}, nil
}

// SearchDouban 搜索豆瓣内容
func (s *impl) SearchDouban(query, searchType string, page int) (*douban.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if page < 1 {
		page = 1
	}

	start := (page - 1) * 20 // 每页20条
	count := 20

	ctx := context.Background()
	
	// 根据搜索类型调用不同的搜索方法
	switch searchType {
	case "movie":
		return s.client.SearchMovie(ctx, query, start, count)
	case "tv":
		// 豆瓣API中电视剧搜索也使用movie接口
		result, err := s.client.SearchMovie(ctx, query, start, count)
		if err != nil {
			return nil, err
		}
		
		// 过滤出电视剧类型的结果
		tvResults := make([]douban.SubjectDetails, 0)
		for _, subject := range result.Subjects {
			if subject.Subtype == "tv" {
				tvResults = append(tvResults, subject)
			}
		}
		
		result.Subjects = tvResults
		result.Count = len(tvResults)
		return result, nil
	case "person":
		// TODO: 实现人物搜索
		return &douban.SearchResult{
			Count:    0,
			Start:    start,
			Total:    0,
			Subjects: []douban.SubjectDetails{},
		}, nil
	default:
		// 默认搜索电影
		return s.client.SearchMovie(ctx, query, start, count)
	}
}

// GetUserWishList 获取用户想看列表
func (s *impl) GetUserWishList(userID, contentType string) (*douban.SearchResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	if contentType == "" {
		contentType = "movie"
	}

	// TODO: 实现获取用户想看列表的逻辑
	// 目前返回空结果
	return &douban.SearchResult{
		Count:    0,
		Start:    0,
		Total:    0,
		Subjects: []douban.SubjectDetails{},
	}, nil
}

// GetUserDoList 获取用户在看列表
func (s *impl) GetUserDoList(userID, contentType string) (*douban.SearchResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	if contentType == "" {
		contentType = "tv"
	}

	// TODO: 实现获取用户在看列表的逻辑
	// 目前返回空结果
	return &douban.SearchResult{
		Count:    0,
		Start:    0,
		Total:    0,
		Subjects: []douban.SubjectDetails{},
	}, nil
}

// GetUserCollectList 获取用户已看列表
func (s *impl) GetUserCollectList(userID, contentType string) (*douban.SearchResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	if contentType == "" {
		contentType = "movie"
	}

	// TODO: 实现获取用户已看列表的逻辑
	// 目前返回空结果
	return &douban.SearchResult{
		Count:    0,
		Start:    0,
		Total:    0,
		Subjects: []douban.SubjectDetails{},
	}, nil
}

// GetPersonDetail 获取人物详情
func (s *impl) GetPersonDetail(personID int) (*douban.Person, error) {
	personIDStr := strconv.Itoa(personID)
	if !douban.IsValidDoubanID(personIDStr) {
		return nil, fmt.Errorf("invalid person ID: %d", personID)
	}

	// TODO: 实现获取人物详情的逻辑
	// 目前返回模拟数据
	return &douban.Person{
		ID:        personIDStr,
		Name:      "Person " + personIDStr,
		Alt:       "",
		AvatarURL: "",
	}, nil
}

// GetPersonCredits 获取人物作品
func (s *impl) GetPersonCredits(personID, page int) (*douban.SearchResult, error) {
	personIDStr := strconv.Itoa(personID)
	if !douban.IsValidDoubanID(personIDStr) {
		return nil, fmt.Errorf("invalid person ID: %d", personID)
	}

	if page < 1 {
		page = 1
	}

	// TODO: 实现获取人物作品的逻辑
	// 目前返回空结果
	return &douban.SearchResult{
		Count:    0,
		Start:    (page - 1) * 20,
		Total:    0,
		Subjects: []douban.SubjectDetails{},
	}, nil
}

// GetDoubanCredits 获取演职员表
func (s *impl) GetDoubanCredits(doubanID, typeName string) ([]douban.Person, error) {
	if !douban.IsValidDoubanID(doubanID) {
		return nil, fmt.Errorf("invalid douban ID: %s", doubanID)
	}

	if typeName == "" {
		typeName = "movie" // 默认为电影类型
	}

	// 获取详细信息以提取演职员表
	ctx := context.Background()
	details, err := s.client.GetMovieDetails(ctx, doubanID)
	if err != nil {
		return nil, err
	}

	// 根据类型返回相应的演职员表
	switch typeName {
	case "movie", "tv", "电影", "电视剧":
		// 合并导演和演员
		credits := make([]douban.Person, 0)
		credits = append(credits, details.Directors...)
		credits = append(credits, details.Casts...)
		return credits, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", typeName)
	}
}

// GetDoubanRecommendations 获取推荐内容
func (s *impl) GetDoubanRecommendations(doubanID, typeName string) (*douban.SearchResult, error) {
	if !douban.IsValidDoubanID(doubanID) {
		return nil, fmt.Errorf("invalid douban ID: %s", doubanID)
	}

	if typeName == "" {
		typeName = "movie" // 默认为电影类型
	}

	// TODO: 实现获取推荐内容的逻辑
	// 目前返回空结果
	return &douban.SearchResult{
		Count:    0,
		Start:    0,
		Total:    0,
		Subjects: []douban.SubjectDetails{},
	}, nil
}