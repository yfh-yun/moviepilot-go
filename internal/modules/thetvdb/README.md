# TheTVDB Module

TheTVDB模块用于从TheTVDB API获取电视剧相关信息。

## 功能特性

- 通过TVDB API获取电视剧详细信息
- 支持搜索电视剧
- 自动处理认证和令牌刷新
- 线程安全的会话管理
- 错误处理和重试机制

## 配置

在配置文件中需要设置以下参数：

- `TVDB_V4_API_KEY`: TVDB API密钥
- `TVDB_V4_API_PIN`: TVDB API PIN（可选）
- `PROXY`: 代理设置（可选）

## 主要方法

### TvdbInfo(tvdbid int) map[string]interface{}

根据TVDB ID获取电视剧详细信息。

### SearchTvdb(title string) []map[string]interface{}

根据标题搜索电视剧。

### Test() (bool, string)

测试模块连接性。

## 使用示例

```go
module := NewTheTvDbModule()
module.InitModule()

// 获取电视剧信息
info := module.TvdbInfo(81189)

// 搜索电视剧
results := module.SearchTvdb("The Simpsons")
```