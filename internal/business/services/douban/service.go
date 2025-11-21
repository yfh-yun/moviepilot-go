// Package douban 豆瓣服务接口定义
package douban

import (
	"moviepilot-go/internal/integration/douban"
)

// Service 豆瓣服务接口
type Service interface {
	// GetMovieInfo 获取电影信息
	GetMovieInfo(doubanID string) (*douban.SubjectDetails, error)
	
	// GetTVInfo 获取电视剧信息  
	GetTVInfo(doubanID string) (*douban.SubjectDetails, error)
	
	// GetPersonInfo 获取人物信息
	GetPersonInfo(doubanID string) (*douban.Person, error)
	
	// SearchDouban 搜索豆瓣内容
	SearchDouban(query, searchType string, page int) (*douban.SearchResult, error)
	
	// GetUserWishList 获取用户想看列表
	GetUserWishList(userID, contentType string) (*douban.SearchResult, error)
	
	// GetUserDoList 获取用户在看列表
	GetUserDoList(userID, contentType string) (*douban.SearchResult, error)
	
	// GetUserCollectList 获取用户已看列表
	GetUserCollectList(userID, contentType string) (*douban.SearchResult, error)
	
	// GetPersonDetail 获取人物详情
	GetPersonDetail(personID int) (*douban.Person, error)
	
	// GetPersonCredits 获取人物作品
	GetPersonCredits(personID, page int) (*douban.SearchResult, error)
	
	// GetDoubanCredits 获取演职员表
	GetDoubanCredits(doubanID, typeName string) ([]douban.Person, error)
	
	// GetDoubanRecommendations 获取推荐内容
	GetDoubanRecommendations(doubanID, typeName string) (*douban.SearchResult, error)
}