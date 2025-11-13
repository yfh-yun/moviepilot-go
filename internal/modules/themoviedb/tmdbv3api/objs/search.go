package objs

import (
	"net/url"
	"strconv"
	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Search 搜索对象
type Search struct {
	tmdb *tmdbv3api.TMDb
}

// NewSearch 创建Search实例
func NewSearch(tmdb *tmdbv3api.TMDb) *Search {
	return &Search{
		tmdb: tmdb,
	}
}

// Companies 搜索公司
/*
Search for companies.
:param term: 搜索关键�?:param page: 页码
:return: 搜索结果
*/
func (s *Search) Companies(term string, page int) ([]interface{}, error) {
	action := "/search/company"
	params := "query=" + url.QueryEscape(term) + "&page=" + strconv.Itoa(page)
	key := "results"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Collections 搜索集合
/*
Search for collections.
:param term: 搜索关键�?:param page: 页码
:return: 搜索结果
*/
func (s *Search) Collections(term string, page int) ([]interface{}, error) {
	action := "/search/collection"
	params := "query=" + url.QueryEscape(term) + "&page=" + strconv.Itoa(page)
	key := "results"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Keywords 搜索关键�?/*
Search for keywords.
:param term: 搜索关键�?:param page: 页码
:return: 搜索结果
*/
func (s *Search) Keywords(term string, page int) ([]interface{}, error) {
	action := "/search/keyword"
	params := "query=" + url.QueryEscape(term) + "&page=" + strconv.Itoa(page)
	key := "results"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Movies 搜索电影
/*
Search for movies.
:param term: 搜索关键�?:param adult: 是否包含成人内容
:param region: 地区
:param year: 年份
:param release_year: 发行年份
:param page: 页码
:return: 搜索结果
*/
func (s *Search) Movies(term string, adult *bool, region *string, year *int, releaseYear *int, page int) ([]interface{}, error) {
	action := "/search/movie"
	params := "query=" + url.QueryEscape(term) + "&page=" + strconv.Itoa(page)
	if adult != nil {
		if *adult {
			params += "&include_adult=true"
		} else {
			params += "&include_adult=false"
		}
	}
	if region != nil {
		params += "&region=" + url.QueryEscape(*region)
	}
	if year != nil {
		params += "&year=" + strconv.Itoa(*year)
	}
	if releaseYear != nil {
		params += "&primary_release_year=" + strconv.Itoa(*releaseYear)
	}
	key := "results"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// Multi 多类型搜�?/*
Search multiple models in a single request.
Multi search currently supports searching for movies, tv shows and people in a single request.
:param term: 搜索关键�?:param adult: 是否包含成人内容
:param region: 地区
:param page: 页码
:return: 搜索结果
*/
func (s *Search) Multi(term string, adult *bool, region *string, page int) ([]interface{}, error) {
	action := "/search/multi"
	params := "query=" + url.QueryEscape(term) + "&page=" + strconv.Itoa(page)
	if adult != nil {
		if *adult {
			params += "&include_adult=true"
		} else {
			params += "&include_adult=false"
		}
	}
	if region != nil {
		params += "&region=" + url.QueryEscape(*region)
	}
	key := "results"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// People 搜索人物
/*
Search for people.
:param term: 搜索关键�?:param adult: 是否包含成人内容
:param region: 地区
:param page: 页码
:return: 搜索结果
*/
func (s *Search) People(term string, adult *bool, region *string, page int) ([]interface{}, error) {
	action := "/search/person"
	params := "query=" + url.QueryEscape(term) + "&page=" + strconv.Itoa(page)
	if adult != nil {
		if *adult {
			params += "&include_adult=true"
		} else {
			params += "&include_adult=false"
		}
	}
	if region != nil {
		params += "&region=" + url.QueryEscape(*region)
	}
	key := "results"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// TVShows 搜索电视�?/*
Search for a TV show.
:param term: 搜索关键�?:param adult: 是否包含成人内容
:param release_year: 发行年份
:param page: 页码
:return: 搜索结果
*/
func (s *Search) TVShows(term string, adult *bool, releaseYear *int, page int) ([]interface{}, error) {
	action := "/search/tv"
	params := "query=" + url.QueryEscape(term) + "&page=" + strconv.Itoa(page)
	if adult != nil {
		if *adult {
			params += "&include_adult=true"
		} else {
			params += "&include_adult=false"
		}
	}
	if releaseYear != nil {
		params += "&first_air_date_year=" + strconv.Itoa(*releaseYear)
	}
	key := "results"

	result, err := s.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
