package objs

import (
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Credit 信用对象
type Credit struct {
	tmdb *tmdbv3api.TMDb
}

// NewCredit 创建Credit实例
func NewCredit(tmdb *tmdbv3api.TMDb) *Credit {
	return &Credit{
		tmdb: tmdb,
	}
}

// Details 获取电影或电视节目的信用详情
/*
Get a movie or TV credit details by id.
*/
func (c *Credit) Details(creditID int) (map[string]interface{}, error) {
	action := "/credit/" + strconv.Itoa(creditID)
	
	result, err := c.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
