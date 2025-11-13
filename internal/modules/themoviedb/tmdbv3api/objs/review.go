package objs

import (
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Review 评论对象
type Review struct {
	tmdb *tmdbv3api.TMDb
}

// NewReview 创建Review实例
func NewReview(tmdb *tmdbv3api.TMDb) *Review {
	return &Review{
		tmdb: tmdb,
	}
}

// Details 通过ID获取评论详情
/*
Get the primary person details by id.
*/
func (r *Review) Details(reviewID int) (map[string]interface{}, error) {
	action := "/review/" + strconv.Itoa(reviewID)
	
	result, err := r.tmdb.RequestObj(action, "", "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
