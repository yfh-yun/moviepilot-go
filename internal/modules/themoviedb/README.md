# TheMovieDB Module

TheMovieDB模块用于从TheMovieDB API获取电影和电视剧相关信息。

## 功能特性

- 通过TMDB API获取电影和电视剧详细信息
- 支持搜索电影、电视剧和人物
- 自动处理认证和请求
- 支持二级分类
- 支持元数据刮削
- 缓存机制提高性能

## 配置

在配置文件中需要设置以下参数：

- `TMDB_API_KEY`: TMDB API密钥
- `TMDB_API_DOMAIN`: TMDB API域名
- `TMDB_IMAGE_DOMAIN`: TMDB图片域名
- `TMDB_LOCALE`: TMDB语言设置

## 主要方法

### RecognizeMedia(meta *models.MetaBase, mtype models.MediaType, tmdbid *int, episodeGroup *string, cache bool, kwargs map[string]interface{}) *models.MediaInfo

识别媒体信息。

### SearchMedias(meta *models.MetaBase) []*models.MediaInfo

搜索媒体信息。

### TmdbInfo(tmdbid int, mtype models.MediaType, season *int) map[string]interface{}

获取TMDB信息。

### MediaCategory() map[string][]string

获取媒体分类。

## 使用示例

```go
module := NewTheMovieDbModule()
module.InitModule()

// 识别媒体信息
mediaInfo := module.RecognizeMedia(meta, models.MediaTypeMovie, nil, nil, true, nil)

// 搜索媒体
medias := module.SearchMedias(meta)

// 获取TMDB信息
info := module.TmdbInfo(550, models.MediaTypeMovie, nil)
```