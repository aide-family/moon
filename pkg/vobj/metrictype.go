package vobj

import (
	"github.com/go-kratos/kratos/v2/log"
)

// MetricType 数据源类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=MetricType -linecomment
type MetricType int

const (
	MetricTypeUnknown MetricType = iota // 未知

	MetricTypeCounter // 计数器

	MetricTypeGauge // 仪表盘

	MetricTypeHistogram // 直方图

	MetricTypeSummary // 摘要
)

// GetMetricType 获取指标类型
func GetMetricType(metricType string) MetricType {
	switch metricType {
	case "counter":
		return MetricTypeCounter
	case "histogram":
		return MetricTypeHistogram
	case "gauge":
		return MetricTypeGauge
	case "summary":
		return MetricTypeSummary
	default:
		log.Warnw("method", "GetMetricType", "metricType", metricType)
		return MetricTypeUnknown
	}
}
