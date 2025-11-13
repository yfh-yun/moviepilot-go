package models

// Statistic 统计信息
type Statistic struct {
	// 电影数量
	MovieCount int `json:"movie_count"`
	// 电视剧数�?	TvCount int `json:"tv_count"`
	// 用户数量
	UserCount int `json:"user_count,omitempty"`
}

// NewStatistic 创建一个新�?Statistic 实例
func NewStatistic() *Statistic {
	return &Statistic{}
}
