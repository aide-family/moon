package vobj

// MetricType is the type of metric.
//
//go:generate stringer -type=MetricType -linecomment -output=metric_type.string.go
type MetricType int8

const (
	MetricTypeUnknown   MetricType = iota // unknown
	MetricTypeCounter                     // counter
	MetricTypeGauge                       // gauge
	MetricTypeHistogram                   // histogram
	MetricTypeSummary                     // summary
)
