# TMDb v3 API

TMDb v3 API的Go语言实现，提供了对The Movie Database API的完整访问。

## 功能特性

- 完整的TMDb API访问
- 支持电影、电视剧、人物、搜索等所有主要功能
- 内置缓存机制
- 频率限制处理
- 错误处理
- 支持代理和自定义配置

## 主要对象

- **TMDb**: 核心API客户端
- **Movie**: 电影相关操作
- **TV**: 电视剧相关操作
- **Season**: 季相关操作
- **Episode**: 集相关操作
- **Person**: 人物相关操作
- **Search**: 搜索功能
- **Discover**: 发现功能
- **Trending**: 趋势功能
- **Collection**: 合集功能

## 使用示例

```go
// 创建TMDb实例
tmdb := tmdbv3api.NewTMDb(true, nil)

// 创建电影对象
movie := objs.NewMovie(tmdb)

// 获取电影详情
details, err := movie.Details(550, "images,credits")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("电影标题: %s\n", details["title"])
```

## 配置

API会自动从项目配置中读取以下参数：
- TMDB_API_KEY: API密钥
- TMDB_LOCALE: 语言设置
- PROXY: 代理设置
- TMDB_API_DOMAIN: API域名