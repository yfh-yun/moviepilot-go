package site

// Site 站点信息
type Site struct {
	// ID
	ID int `json:"id,omitempty"`
	// 站点名称
	Name string `json:"name,omitempty"`
	// 站点主域名Key
	Domain string `json:"domain,omitempty"`
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
	ApiKey string `json:"apikey,omitempty"`
	// Token
	Token string `json:"token,omitempty"`
	// 是否使用代理
	Proxy int `json:"proxy,omitempty"`
	// 过滤规则
	Filter string `json:"filter,omitempty"`
	// 是否渲染
	Render int `json:"render,omitempty"`
	// 是否公开站点
	Public int `json:"public,omitempty"`
	// 备注
	Note any `json:"note,omitempty"`
	// 超时时间
	Timeout int `json:"timeout,omitempty"`
	// 流控单位周期
	LimitInterval *int `json:"limit_interval,omitempty"`
	// 流控次数
	LimitCount *int `json:"limit_count,omitempty"`
	// 流控间隔
	LimitSeconds *int `json:"limit_seconds,omitempty"`
	// 是否启用
	IsActive bool `json:"is_active,omitempty"`
	// 下载器
	Downloader string `json:"downloader,omitempty"`
}

// SiteStatistic 站点统计
type SiteStatistic struct {
	// 站点ID
	Domain string `json:"domain,omitempty"`
	// 成功次数
	Success int `json:"success,omitempty"`
	// 失败次数
	Fail int `json:"fail,omitempty"`
	// 平均响应时间
	Seconds int `json:"seconds,omitempty"`
	// 最后状态
	LstState int `json:"lst_state,omitempty"`
	// 最后修改时间
	LstModDate string `json:"lst_mod_date,omitempty"`
	// 备注
	Note any `json:"note,omitempty"`
}

// SiteUserData 站点用户数据
type SiteUserData struct {
	// 站点域名
	Domain string `json:"domain,omitempty"`
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
	Upload int64 `json:"upload,omitempty"`
	// 下载量
	Download int64 `json:"download,omitempty"`
	// 分享率
	Ratio float64 `json:"ratio,omitempty"`
	// 做种数
	Seeding int `json:"seeding,omitempty"`
	// 下载数
	Leeching int `json:"leeching,omitempty"`
	// 做种体积
	SeedingSize int64 `json:"seeding_size,omitempty"`
	// 下载体积
	LeechingSize int64 `json:"leeching_size,omitempty"`
	// 做种人数, 种子大小
	SeedingInfo []any `json:"seeding_info,omitempty"`
	// 未读消息
	MessageUnread int `json:"message_unread,omitempty"`
	// 未读消息内容
	MessageUnreadContents []any `json:"message_unread_contents,omitempty"`
	// 错误信息
	ErrMsg string `json:"err_msg,omitempty"`
	// 更新日期
	UpdatedDay string `json:"updated_day,omitempty"`
	// 更新时间
	UpdatedTime string `json:"updated_time,omitempty"`
}

// SiteAuth 站点认证
type SiteAuth struct {
	Site   string         `json:"site,omitempty"`
	Params map[string]any `json:"params,omitempty"`
}

// SiteCategory 站点分类
type SiteCategory struct {
	ID   int    `json:"id,omitempty"`
	Cat  string `json:"cat,omitempty"`
	Desc string `json:"desc,omitempty"`
}
