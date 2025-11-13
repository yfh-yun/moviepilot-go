package models

// Message 消息表
type Message struct {
	// ID
	ID uint `json:"id,omitempty" gorm:"primaryKey;autoIncrement"`
	
	// 消息渠道
	Channel string `json:"channel,omitempty"`
	
	// 消息来源
	Source string `json:"source,omitempty"`
	
	// 消息类型
	MType string `json:"mtype,omitempty"`
	
	// 标题
	Title string `json:"title,omitempty"`
	
	// 文本内容
	Text string `json:"text,omitempty"`
	
	// 图片
	Image string `json:"image,omitempty"`
	
	// 链接
	Link string `json:"link,omitempty"`
	
	// 用户ID
	UserID string `json:"userid,omitempty"`
	
	// 登记时间
	RegTime string `json:"reg_time,omitempty" gorm:"index"`
	
	// 消息方向：0-接收息，1-发送消息
	Action int `json:"action,omitempty"`
	
	// 附件json
	Note map[string]interface{} `json:"note,omitempty" gorm:"serializer:json"`
}

// TableName 设置表名
func (Message) TableName() string {
	return "message"
}