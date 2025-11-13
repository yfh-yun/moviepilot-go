package indexer

// SiteParserBase 站点解析器基础接口
type SiteParserBase interface {
	// Parse 解析站点数据
	Parse() error
	
	// Clear 清理资源
	Clear()
	
	// GetUserID 获取用户ID
	GetUserID() string
	
	// GetUsername 获取用户�?	GetUsername() string
	
	// GetUserLevel 获取用户等级
	GetUserLevel() string
	
	// GetJoinAt 获取加入时间
	GetJoinAt() string
	
	// GetUpload 获取上传�?	GetUpload() int
	
	// GetDownload 获取下载�?	GetDownload() int
	
	// GetRatio 获取分享�?	GetRatio() float64
	
	// GetBonus 获取积分
	GetBonus() float64
	
	// GetSeeding 获取做种�?	GetSeeding() int
	
	// GetLeeching 获取下载�?	GetLeeching() int
	
	// GetSeedingSize 获取做种体积
	GetSeedingSize() int
	
	// GetLeechingSize 获取下载体积
	GetLeechingSize() int
	
	// GetSeedingInfo 获取做种信息
	GetSeedingInfo() []interface{}
	
	// GetMessageUnread 获取未读消息�?	GetMessageUnread() int
	
	// GetMessageUnreadContents 获取未读消息内容
	GetMessageUnreadContents() []interface{}
	
	// GetErrMsg 获取错误信息
	GetErrMsg() string
	
	// GetSchema 获取站点解析器模�?	GetSchema() interface{}
	
	// SiteSchema 获取站点解析模型
	SiteSchema() SiteSchema
	
	// 以下是需要子类实现的抽象方法
	
	// parseMessageUnreadLinks 获取未阅读消息链�?	parseMessageUnreadLinks(htmlText string, msgLinks []string) string
	
	// parseSitePage 解析站点相关信息页面
	parseSitePage(htmlText string)
	
	// parseUserBaseInfo 解析用户基础信息
	parseUserBaseInfo(htmlText string)
	
	// parseUserTrafficInfo 解析用户的上传，下载，分享率等信�?	parseUserTrafficInfo(htmlText string)
	
	// parseUserTorrentSeedingInfo 解析用户的做种相关信�?	parseUserTorrentSeedingInfo(htmlText string, multiPage bool) string
	
	// parseUserDetailInfo 解析用户的详细信�?	parseUserDetailInfo(htmlText string)
	
	// parseMessageContent 解析短消息内�?	parseMessageContent(htmlText string) (string, string, string)
}
