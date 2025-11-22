package monitor

import (
	"moviepilot-go/pkg/logger"
	"go.uber.org/zap"
)

// BaseCollector 基础收集器，包含通用方法
type BaseCollector struct {
	MetricType MetricType
}

// NewBaseCollector 创建基础收集器
func NewBaseCollector(metricType MetricType) *BaseCollector {
	logger.Debug("Creating new BaseCollector instance", 
		zap.String("func", "NewBaseCollector"),
		zap.String("metric_type", string(metricType)))
	return &BaseCollector{
		MetricType: metricType,
	}
}

// GetMetricType 获取指标类型
func (bc *BaseCollector) GetMetricType() MetricType {
	return bc.MetricType
}

// LogCollectionStart 记录收集开始日志
func (bc *BaseCollector) LogCollectionStart(collectorName string) {
	logger.Debug("Starting metrics collection", 
		zap.String("func", collectorName+".Collect"),
		zap.String("metric_type", string(bc.MetricType)))
}

// LogCollectionError 记录收集错误日志
func (bc *BaseCollector) LogCollectionError(collectorName string, err error, operation string) {
	logger.Error("Failed to collect metrics", 
		zap.String("func", collectorName+".Collect"),
		zap.String("operation", operation),
		zap.String("metric_type", string(bc.MetricType)),
		zap.Error(err))
}

// LogCollectionSuccess 记录收集成功日志
func (bc *BaseCollector) LogCollectionSuccess(collectorName string, metricCount int) {
	logger.Debug("Successfully collected metrics", 
		zap.String("func", collectorName+".Collect"),
		zap.String("metric_type", string(bc.MetricType)),
		zap.Int("metric_count", metricCount))
}