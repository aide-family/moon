package vobj

// DatasourceType 数据源类型
//
//go:generate stringer -type=DatasourceType -linecomment
type DatasourceType int

const (
	DataSourceTypeUnknown DatasourceType = iota // 未知
	DataSourceTypeMetrics                       // 监控指标
	DataSourceTypeTrace                         // 链路追踪
	DataSourceTypeLog                           // 日志
)
