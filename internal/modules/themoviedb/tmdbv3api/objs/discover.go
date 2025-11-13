package objs

import (
	"net/url"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Discover 发现对象
type Discover struct {
	tmdb *tmdbv3api.TMDb
}

// NewDiscover 创建Discover实例
func NewDiscover(tmdb *tmdbv3api.TMDb) *Discover {
	return &Discover{
		tmdb: tmdb,
	}
}

// DiscoverMovies 通过不同类型的数据发现电影，如平均评分、投票数、类型和分级
/*
Discover movies by different types of data like average rating, number of votes, genres and certifications.
*/
func (d *Discover) DiscoverMovies(params map[string]string) ([]interface{}, error) {
	action := "/discover/movie"
	
	// 构建参数字符�?	paramStr := buildParamString(params)
	key := "results"
	
	result, err := d.tmdb.RequestObj(action, paramStr, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// DiscoverTVShows 通过不同类型的数据发现电视节目，如平均评分、投票数、类型、播出网络和播出日期
/*
Discover TV shows by different types of data like average rating, number of votes, genres,
the network they aired on and air dates.
*/
func (d *Discover) DiscoverTVShows(params map[string]string) ([]interface{}, error) {
	action := "/discover/tv"
	
	// 构建参数字符�?	paramStr := buildParamString(params)
	key := "results"
	
	result, err := d.tmdb.RequestObj(action, paramStr, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// buildParamString 构建参数字符�?func buildParamString(params map[string]string) string {
	// 使用url.Values来正确编码参�?	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return values.Encode()
}
