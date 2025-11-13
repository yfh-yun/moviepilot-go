package objs

import (
	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Genre 类型对象
type Genre struct {
	tmdb *tmdbv3api.TMDb
}

// NewGenre 创建Genre实例
func NewGenre(tmdb *tmdbv3api.TMDb) *Genre {
	return &Genre{
		tmdb: tmdb,
	}
}

// MovieList 获取电影的官方类型列�?/*
Get the list of official genres for movies.
*/
func (g *Genre) MovieList() ([]interface{}, error) {
	action := "/genre/movie/list"
	key := "genres"
	
	result, err := g.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// TVList 获取电视节目的官方类型列�?/*
Get the list of official genres for TV shows.
*/
func (g *Genre) TVList() ([]interface{}, error) {
	action := "/genre/tv/list"
	key := "genres"
	
	result, err := g.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
