package objs

import (
	"fmt"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Group 组对�?type Group struct {
	tmdb *tmdbv3api.TMDb
}

// NewGroup 创建Group实例
func NewGroup(tmdb *tmdbv3api.TMDb) *Group {
	return &Group{
		tmdb: tmdb,
	}
}

// Details 获取TV剧集组的详情
/*
Get the details of a TV episode group.
*/
func (g *Group) Details(groupID int) ([]interface{}, error) {
	action := fmt.Sprintf("/tv/episode_group/%d", groupID)
	key := "groups"
	
	result, err := g.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}
