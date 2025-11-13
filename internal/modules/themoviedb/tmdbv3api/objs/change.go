package objs

import (
	"fmt"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Change 变更对象
type Change struct {
	tmdb *tmdbv3api.TMDb
}

// NewChange 创建Change实例
func NewChange(tmdb *tmdbv3api.TMDb) *Change {
	return &Change{
		tmdb: tmdb,
	}
}

// changeList 获取变更列表的通用方法
/*
内部方法，用于获取不同类型的变更列表
*/
func (c *Change) changeList(changeType, startDate, endDate string, page int) ([]interface{}, error) {
	params := "page=" + strconv.Itoa(page)
	if startDate != "" {
		params += "&start_date=" + startDate
	}
	if endDate != "" {
		params += "&end_date=" + endDate
	}
	
	key := "results"
	result, err := c.tmdb.RequestObj("/"+changeType+"/changes", params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// MovieChangeList 获取电影变更列表
/*
Get the changes for a movie. By default only the last 24 hours are returned.
You can query up to 14 days in a single query by using the start_date and end_date query parameters.
*/
func (c *Change) MovieChangeList(startDate, endDate string, page int) ([]interface{}, error) {
	return c.changeList("movie", startDate, endDate, page)
}

// TVChangeList 获取电视节目变更列表
/*
Get a list of all of the TV show ids that have been changed in the past 24 hours.
You can query up to 14 days in a single query by using the start_date and end_date query parameters.
*/
func (c *Change) TVChangeList(startDate, endDate string, page int) ([]interface{}, error) {
	return c.changeList("tv", startDate, endDate, page)
}

// PersonChangeList 获取人物变更列表
/*
Get a list of all of the person ids that have been changed in the past 24 hours.
You can query up to 14 days in a single query by using the start_date and end_date query parameters.
*/
func (c *Change) PersonChangeList(startDate, endDate string, page int) ([]interface{}, error) {
	return c.changeList("person", startDate, endDate, page)
}
