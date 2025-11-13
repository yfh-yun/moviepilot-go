package models

// Site 站点信息模型
type Site struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 站点名称
	Name string `json:"name,omitempty"`
	
	// 域名Key
	Domain string `json:"domain,omitempty" gorm:"index"`
	
	// 站点地址
	URL string `json:"url,omitempty"`
	
	// 站点优先级
	Pri int `json:"pri,omitempty"`
	
	// RSS地址
	RSS string `json:"rss,omitempty"`
	
	// Cookie
	Cookie string `json:"cookie,omitempty"`
	
	// User-Agent
	UA string `json:"ua,omitempty"`
	
	// ApiKey
	APIKey string `json:"apikey,omitempty"`
	
	// Token
	Token string `json:"token,omitempty"`
	
	// 是否使用代理 0-否，1-是
	Proxy int `json:"proxy,omitempty"`
	
	// 过滤规则
	Filter string `json:"filter,omitempty"`
	
	// 是否渲染
	Render int `json:"render,omitempty"`
	
	// 是否公开站点
	Public int `json:"public,omitempty"`
	
	// 附加信息
	Note map[string]interface{} `json:"note,omitempty" gorm:"serializer:json"`
	
	// 流控单位周期
	LimitInterval int `json:"limit_interval,omitempty"`
	
	// 流控次数
	LimitCount int `json:"limit_count,omitempty"`
	
	// 流控间隔
	LimitSeconds int `json:"limit_seconds,omitempty"`
	
	// 超时时间
	Timeout int `json:"timeout,omitempty"`
	
	// 是否启用
	IsActive bool `json:"is_active,omitempty"`
	
	// 创建时间
	LstModDate string `json:"lst_mod_date,omitempty"`
	
	// 下载器
	Downloader string `json:"downloader,omitempty"`
}

// TableName 设置表名
func (Site) TableName() string {
	return "site"
}

// SiteStatistic 站点统计表
type SiteStatistic struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 域名Key
	Domain string `json:"domain,omitempty" gorm:"index"`
	
	// 成功次数
	Success int `json:"success,omitempty"`
	
	// 失败次数
	Fail int `json:"fail,omitempty"`
	
	// 平均耗时 秒
	Seconds int `json:"seconds,omitempty"`
	
	// 最后一次访问状态 0-成功 1-失败
	LstState int `json:"lst_state,omitempty"`
	
	// 最后访问时间
	LstModDate string `json:"lst_mod_date,omitempty"`
	
	// 耗时记录 Json
	Note map[string]interface{} `json:"note,omitempty" gorm:"serializer:json"`
}

// TableName 设置表名
func (SiteStatistic) TableName() string {
	return "sitestatistic"
}

// SiteUserData 站点数据表
type SiteUserData struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 站点域名
	Domain string `json:"domain,omitempty" gorm:"index"`
	
	// 站点名称
	Name string `json:"name,omitempty"`
	
	// 用户名
	Username string `json:"username,omitempty"`
	
	// 用户ID
	UserID string `json:"userid,omitempty"`
	
	// 用户等级
	UserLevel string `json:"user_level,omitempty"`
	
	// 加入时间
	JoinAt string `json:"join_at,omitempty"`
	
	// 积分
	Bonus float64 `json:"bonus,omitempty"`
	
	// 上传量
	Upload float64 `json:"upload,omitempty"`
	
	// 下载量
	Download float64 `json:"download,omitempty"`
	
	// 分享率
	Ratio float64 `json:"ratio,omitempty"`
	
	// 做种数
	Seeding float64 `json:"seeding,omitempty"`
	
	// 下载数
	Leeching float64 `json:"leeching,omitempty"`
	
	// 做种体积
	SeedingSize float64 `json:"seeding_size,omitempty"`
	
	// 下载体积
	LeechingSize float64 `json:"leeching_size,omitempty"`
	
	// 做种人数, 种子大小 JSON
	SeedingInfo []interface{} `json:"seeding_info,omitempty" gorm:"serializer:json"`
	
	// 未读消息
	MessageUnread int `json:"message_unread,omitempty"`
	
	// 未读消息内容 JSON
	MessageUnreadContents []interface{} `json:"message_unread_contents,omitempty" gorm:"serializer:json"`
	
	// 错误信息
	ErrMsg string `json:"err_msg,omitempty"`
	
	// 更新日期
	UpdatedDay string `json:"updated_day,omitempty" gorm:"index"`
	
	// 更新时间
	UpdatedTime string `json:"updated_time,omitempty"`
}

// TableName 设置表名
func (SiteUserData) TableName() string {
	return "siteuserdata"
}

// SiteIcon 站点图标表
type SiteIcon struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 站点名称
	Name string `json:"name,omitempty"`
	
	// 域名Key
	Domain string `json:"domain,omitempty" gorm:"index"`
	
	// 图标地址
	URL string `json:"url,omitempty"`
	
	// 图标Base64
	Base64 string `json:"base64,omitempty"`
}

// TableName 设置表名
func (SiteIcon) TableName() string {
	return "siteicon"
}