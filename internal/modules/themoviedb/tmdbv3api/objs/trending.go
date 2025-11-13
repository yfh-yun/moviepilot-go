package objs

import (
	"fmt"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Trending 趋势对象
type Trending struct {
	tmdb *tmdbv3api.TMDb
}

// NewTrending 创建Trending实例
func NewTrending(tmdb *tmdbv3api.TMDb) *Trending {
	return &Trending{
		tmdb: tmdb,
	}
}

// trending 获取趋势数据的内部方�?/*
Get trending, TTLCache 12 hours
:param media_type: 媒体类型 (all, movie, tv, person)
:param time_window: 时间窗口 (day, week)
:param page: 页码
:return: 趋势数据结果
*/
func (t *Trending) trending(mediaType string, timeWindow string, page int) ([]interface{}, error) {
	action := fmt.Sprintf("/trending/%s/%s", mediaType, timeWindow)
	params := "page=" + strconv.Itoa(page)
	key := "results"

	result, err := t.tmdb.RequestObj(action, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}

	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}

	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AllDay 获取所有类型日趋势
/*
Get all daily trending
:param page: 页码
:return: 日趋势数�?*/
func (t *Trending) AllDay(page int) ([]interface{}, error) {
	return t.trending("all", "day", page)
}

// AllWeek 获取所有类型周趋势
/*
Get all weekly trending
:param page: 页码
:return: 周趋势数�?*/
func (t *Trending) AllWeek(page int) ([]interface{}, error) {
	return t.trending("all", "week", page)
}

// MovieDay 获取电影日趋�?/*
Get movie daily trending
:param page: 页码
:return: 电影日趋势数�?*/
func (t *Trending) MovieDay(page int) ([]interface{}, error) {
	return t.trending("movie", "day", page)
}

// MovieWeek 获取电影周趋�?/*
Get movie weekly trending
:param page: 页码
:return: 电影周趋势数�?*/
func (t *Trending) MovieWeek(page int) ([]interface{}, error) {
	return t.trending("movie", "week", page)
}

// TVDay 获取电视节目日趋�?/*
Get tv daily trending
:param page: 页码
:return: 电视节目日趋势数�?*/
func (t *Trending) TVDay(page int) ([]interface{}, error) {
	return t.trending("tv", "day", page)
}

// TVWeek 获取电视节目周趋�?/*
Get tv weekly trending
:param page: 页码
:return: 电视节目周趋势数�?*/
func (t *Trending) TVWeek(page int) ([]interface{}, error) {
	return t.trending("tv", "week", page)
}

// PersonDay 获取人物日趋�?/*
Get person daily trending
:param page: 页码
:return: 人物日趋势数�?*/
func (t *Trending) PersonDay(page int) ([]interface{}, error) {
	return t.trending("person", "day", page)
}

// PersonWeek 获取人物周趋�?/*
Get person weekly trending
:param page: 页码
:return: 人物周趋势数�?*/
func (t *Trending) PersonWeek(page int) ([]interface{}, error) {
	return t.trending("person", "week", page)
}
