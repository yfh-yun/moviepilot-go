package objs

import (
	"fmt"
	"os"
	"strconv"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Account 账户对象
type Account struct {
	tmdb *tmdbv3api.TMDb
}

// NewAccount 创建Account实例
func NewAccount(tmdb *tmdbv3api.TMDb) *Account {
	return &Account{
		tmdb: tmdb,
	}
}

// AccountID 获取账户ID
func (a *Account) AccountID() (string, error) {
	accountID := os.Getenv("TMDB_ACCOUNT_ID")
	if accountID == "" {
		details, err := a.Details()
		if err != nil {
			return "", err
		}
		
		idFloat, ok := details["id"].(float64)
		if !ok {
			return "", &tmdbv3api.TMDbException{Message: "无法获取账户ID"}
		}
		
		accountID = strconv.FormatFloat(idFloat, 'f', 0, 64)
		os.Setenv("TMDB_ACCOUNT_ID", accountID)
	}
	
	return accountID, nil
}

// Details 获取账户详情
func (a *Account) Details() (map[string]interface{}, error) {
	sessionID, err := a.tmdb.SessionID()
	if err != nil {
		return nil, err
	}
	
	params := fmt.Sprintf("session_id=%s", sessionID)
	result, err := a.tmdb.RequestObj("/account", params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// CreatedLists 获取创建的列�?func (a *Account) CreatedLists(page int) ([]interface{}, error) {
	accountID, err := a.AccountID()
	if err != nil {
		return nil, err
	}
	
	sessionID, err := a.tmdb.SessionID()
	if err != nil {
		return nil, err
	}
	
	url := fmt.Sprintf("/account/%s/lists", accountID)
	params := fmt.Sprintf("session_id=%s&page=%d", sessionID, page)
	key := "results"
	result, err := a.tmdb.RequestObj(url, params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// getLists 获取列表的通用方法
func (a *Account) getLists(url string, ascSort bool, page int) ([]interface{}, error) {
	accountID, err := a.AccountID()
	if err != nil {
		return nil, err
	}
	
	sessionID, err := a.tmdb.SessionID()
	if err != nil {
		return nil, err
	}
	
	params := fmt.Sprintf("session_id=%s&page=%d", sessionID, page)
	if !ascSort {
		params += "&sort_by=created_at.desc"
	}
	
	key := "results"
	result, err := a.tmdb.RequestObj(fmt.Sprintf(url, accountID), params, "GET", nil, nil, &key)
	if err != nil {
		return nil, err
	}
	
	if resultList, ok := result.([]interface{}); ok {
		return resultList, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// FavoriteMovies 获取收藏的电�?func (a *Account) FavoriteMovies(ascSort bool, page int) ([]interface{}, error) {
	return a.getLists("/account/%s/favorite/movies", ascSort, page)
}

// FavoriteTVShows 获取收藏的电视剧
func (a *Account) FavoriteTVShows(ascSort bool, page int) ([]interface{}, error) {
	return a.getLists("/account/%s/favorite/tv", ascSort, page)
}

// MarkAsFavorite 标记为收�?func (a *Account) MarkAsFavorite(mediaID int, mediaType string, favorite bool) error {
	if mediaType != "tv" && mediaType != "movie" {
		return &tmdbv3api.TMDbException{Message: "Media Type should be tv or movie."}
	}
	
	accountID, err := a.AccountID()
	if err != nil {
		return err
	}
	
	sessionID, err := a.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("/account/%s/favorite", accountID)
	params := fmt.Sprintf("session_id=%s", sessionID)
	
	jsonData := map[string]interface{}{
		"media_type": mediaType,
		"media_id":   mediaID,
		"favorite":   favorite,
	}
	
	_, err = a.tmdb.RequestObj(url, params, "POST", nil, jsonData, nil)
	return err
}

// UnmarkAsFavorite 取消收藏
func (a *Account) UnmarkAsFavorite(mediaID int, mediaType string) error {
	return a.MarkAsFavorite(mediaID, mediaType, false)
}

// RatedMovies 获取已评分的电影
func (a *Account) RatedMovies(ascSort bool, page int) ([]interface{}, error) {
	return a.getLists("/account/%s/rated/movies", ascSort, page)
}

// RatedTVShows 获取已评分的电视�?func (a *Account) RatedTVShows(ascSort bool, page int) ([]interface{}, error) {
	return a.getLists("/account/%s/rated/tv", ascSort, page)
}

// RatedEpisodes 获取已评分的剧集
func (a *Account) RatedEpisodes(ascSort bool, page int) ([]interface{}, error) {
	return a.getLists("/account/%s/rated/tv/episodes", ascSort, page)
}

// MovieWatchlist 获取电影观看列表
func (a *Account) MovieWatchlist(ascSort bool, page int) ([]interface{}, error) {
	return a.getLists("/account/%s/watchlist/movies", ascSort, page)
}

// TVShowWatchlist 获取电视剧观看列�?func (a *Account) TVShowWatchlist(ascSort bool, page int) ([]interface{}, error) {
	return a.getLists("/account/%s/watchlist/tv", ascSort, page)
}

// AddToWatchlist 添加到观看列�?func (a *Account) AddToWatchlist(mediaID int, mediaType string, watchlist bool) error {
	if mediaType != "tv" && mediaType != "movie" {
		return &tmdbv3api.TMDbException{Message: "Media Type should be tv or movie."}
	}
	
	accountID, err := a.AccountID()
	if err != nil {
		return err
	}
	
	sessionID, err := a.tmdb.SessionID()
	if err != nil {
		return err
	}
	
	url := fmt.Sprintf("/account/%s/watchlist", accountID)
	params := fmt.Sprintf("session_id=%s", sessionID)
	
	jsonData := map[string]interface{}{
		"media_type": mediaType,
		"media_id":   mediaID,
		"watchlist":  watchlist,
	}
	
	_, err = a.tmdb.RequestObj(url, params, "POST", nil, jsonData, nil)
	return err
}

// RemoveFromWatchlist 从观看列表移�?func (a *Account) RemoveFromWatchlist(mediaID int, mediaType string) error {
	return a.AddToWatchlist(mediaID, mediaType, false)
}
