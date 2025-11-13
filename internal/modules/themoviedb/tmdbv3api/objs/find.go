package objs

import (
	"net/url"
	"strings"

	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Find 查找对象
type Find struct {
	tmdb *tmdbv3api.TMDb
}

// NewFind 创建Find实例
func NewFind(tmdb *tmdbv3api.TMDb) *Find {
	return &Find{
		tmdb: tmdb,
	}
}

// Find 通过外部ID查找对象
/*
The find method makes it easy to search for objects in our database by an external id. For example, an IMDB ID.
*/
func (f *Find) Find(externalID, externalSource string) (map[string]interface{}, error) {
	// 替换"/"�?%2F"
	escapedID := strings.ReplaceAll(externalID, "/", "%2F")
	action := "/find/" + escapedID
	params := "external_source=" + externalSource
	
	result, err := f.tmdb.RequestObj(action, params, "GET", nil, nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	
	return nil, &tmdbv3api.TMDbException{Message: "返回数据格式不正�?}
}

// FindByImdbID 通过IMDB ID查找对象
/*
The find method makes it easy to search for objects in our database by an IMDB ID.
*/
func (f *Find) FindByImdbID(imdbID string) (map[string]interface{}, error) {
	return f.Find(imdbID, "imdb_id")
}

// FindByTvdbID 通过TVDB ID查找对象
/*
The find method makes it easy to search for objects in our database by a TVDB ID.
*/
func (f *Find) FindByTvdbID(tvdbID int) (map[string]interface{}, error) {
	return f.Find(url.QueryEscape(string(rune(tvdbID))), "tvdb_id")
}

// FindByFreebaseMid 通过Freebase MID查找对象
/*
The find method makes it easy to search for objects in our database by a Freebase MID.
*/
func (f *Find) FindByFreebaseMid(freebaseMid string) (map[string]interface{}, error) {
	return f.Find(freebaseMid, "freebase_mid")
}

// FindByFreebaseID 通过Freebase ID查找对象
/*
The find method makes it easy to search for objects in our database by a Freebase ID.
*/
func (f *Find) FindByFreebaseID(freebaseID string) (map[string]interface{}, error) {
	return f.Find(freebaseID, "freebase_id")
}

// FindByTvrageID 通过TVRage ID查找对象
/*
The find method makes it easy to search for objects in our database by a TVRage ID.
*/
func (f *Find) FindByTvrageID(tvrageID string) (map[string]interface{}, error) {
	return f.Find(tvrageID, "tvrage_id")
}

// FindByFacebookID 通过Facebook ID查找对象
/*
The find method makes it easy to search for objects in our database by a Facebook ID.
*/
func (f *Find) FindByFacebookID(facebookID string) (map[string]interface{}, error) {
	return f.Find(facebookID, "facebook_id")
}

// FindByInstagramID 通过Instagram ID查找对象
/*
The find method makes it easy to search for objects in our database by a Instagram ID.
*/
func (f *Find) FindByInstagramID(instagramID string) (map[string]interface{}, error) {
	return f.Find(instagramID, "instagram_id")
}

// FindByTwitterID 通过Twitter ID查找对象
/*
The find method makes it easy to search for objects in our database by a Twitter ID.
*/
func (f *Find) FindByTwitterID(twitterID string) (map[string]interface{}, error) {
	return f.Find(twitterID, "twitter_id")
}
