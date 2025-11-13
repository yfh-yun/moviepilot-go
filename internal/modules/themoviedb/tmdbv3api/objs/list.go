package objs

import (
	"fmt"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// List 列表对象
type List struct {
	tmdb *tmdbv3api.TMDb
}

// NewList 创建List实例
func NewList(tmdb *tmdbv3api.TMDb) *List {
	return &List{
		tmdb: tmdb,
	}
}

// Details 通过ID获取列表详情
/*
Get list details by id.
*/
func (l *List) Details(listID int) ([]interface{}, error) {
	action := fmt.Sprintf("/list/%d", listID)
	key := "items"
	
	result, err := l.tmdb.RequestObj(action, "", "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// CheckItemStatus 检查电影是否已添加到列表中
/*
You can use this method to check if a movie has already been added to the list.
*/
func (l *List) CheckItemStatus(listID, movieID int) (bool, error) {
	action := fmt.Sprintf("/list/%d/item_status", listID)
	params := fmt.Sprintf("movie_id=%d", movieID)
	
	result, err := l.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return false, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		if itemPresent, exists := resultMap["item_present"]; exists {
			if itemPresentBool, ok := itemPresent.(bool); ok {
				return itemPresentBool, nil
			}
		}
	}
	
	return false, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// CreateList 创建列表
/*
You can use this method to check if a movie has already been added to the list.
*/
func (l *List) CreateList(name, description string) (int, error) {
	sessionID, err := l.tmdb.SessionID()
	if err != nil {
		return 0, err
	}
	
	action := "/list"
	params := fmt.Sprintf("session_id=%s", sessionID)
	jsonData := map[string]interface{}{
		"name":        name,
		"description": description,
		"language":    l.tmdb.Language(),
	}
	
	result, err := l.tmdb.RequestObj(action, params, "POST", nil, jsonData, nil)
	if err != nil {
		return 0, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		if listID, exists := resultMap["list_id"]; exists {
			if listIDFloat, ok := listID.(float64); ok {
				return int(listIDFloat), nil
			}
		}
	}
	
	return 0, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// AddMovie 添加电影到列�?/*
Add a movie to a list.
*/
func (l *List) AddMovie(listID, movieID int) error {
	sessionID, err := l.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	action := fmt.Sprintf("/list/%d/add_item", listID)
	params := fmt.Sprintf("session_id=%s", sessionID)
	jsonData := map[string]interface{}{
		"media_id": movieID,
	}
	
	_, err = l.tmdb.RequestObj(action, params, "POST", nil, jsonData, nil)
	return err
}

// RemoveMovie 从列表中移除电影
/*
Remove a movie from a list.
*/
func (l *List) RemoveMovie(listID, movieID int) error {
	sessionID, err := l.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	action := fmt.Sprintf("/list/%d/remove_item", listID)
	params := fmt.Sprintf("session_id=%s", sessionID)
	jsonData := map[string]interface{}{
		"media_id": movieID,
	}
	
	_, err = l.tmdb.RequestObj(action, params, "POST", nil, jsonData, nil)
	return err
}

// ClearList 清空列表中的所有项�?/*
Clear all of the items from a list.
*/
func (l *List) ClearList(listID int) error {
	sessionID, err := l.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	action := fmt.Sprintf("/list/%d/clear", listID)
	params := fmt.Sprintf("session_id=%s&confirm=true", sessionID)
	
	_, err = l.tmdb.RequestObj(action, params, "POST", nil, nil, nil)
	return err
}

// DeleteList 删除列表
/*
Delete a list.
*/
func (l *List) DeleteList(listID int) error {
	sessionID, err := l.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	action := fmt.Sprintf("/list/%d", listID)
	params := fmt.Sprintf("session_id=%s", sessionID)
	
	_, err = l.tmdb.RequestObj(action, params, "DELETE", nil, nil, nil)
	return err
}
