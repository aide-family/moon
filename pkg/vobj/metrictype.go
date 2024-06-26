package vobj

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
