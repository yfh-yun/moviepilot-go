// Package dashboard 仪表板服务
package dashboard



// Service 仪表板服务接口
type Service interface {
	// GetOverview 获取仪表板总览
	GetOverview() (interface{}, error)
	// GetStatistics 获取详细统计
	GetStatistics(statType, period string) (interface{}, error)
	// GetRecentActivities 获取最近活动
	GetRecentActivities(limit int) (interface{}, error)
	// GetChartsData 获取图表数据
	GetChartsData(chartType, period string) (interface{}, error)
}