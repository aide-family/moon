package vobj

// DatasourceType 数据源类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=DatasourceType -linecomment
type DatasourceType int

const (
	DatasourceTypeUnknown DatasourceType = iota // 未知
	DatasourceTypeMetrics                       // 监控指标
	DatasourceTypeTrace                         // 链路追踪
	DatasourceTypeLog                           // 日志
)
