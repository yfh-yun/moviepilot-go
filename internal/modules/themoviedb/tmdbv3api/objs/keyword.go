package objs

import (
	"fmt"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Keyword 关键词对�?type Keyword struct {
	tmdb *tmdbv3api.TMDb
}

// NewKeyword 创建Keyword实例
func NewKeyword(tmdb *tmdbv3api.TMDb) *Keyword {
	return &Keyword{
		tmdb: tmdb,
	}
}

// Details 通过ID获取关键词详�?/*
Get a keywords details by id.
*/
func (k *Keyword) Details(keywordID int) (map[string]interface{}, error) {
	action := fmt.Sprintf("/keyword/%d", keywordID)
	
	result, err := k.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Movies 通过ID获取关键词的电影
/*
Get the movies of a keyword by id.
*/
func (k *Keyword) Movies(keywordID int) ([]interface{}, error) {
	action := fmt.Sprintf("/keyword/%d/movies", keywordID)
	key := "results"
	
	result, err := k.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
