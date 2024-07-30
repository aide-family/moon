package vobj

// MultiDatasourceSustain 持续类型定义
//
//	m时间内出现n次
//	m时间内最多出现n次
//	m时间内最少出现n次
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=MultiDatasourceSustain -linecomment
type MultiDatasourceSustain int

const (
	// MultiDatasourceSustainTypeUnknown 未知
	MultiDatasourceSustainTypeUnknown MultiDatasourceSustain = iota // 未知

	// MultiDatasourceSustainTypeAnd 所有数据告警集合一致
	MultiDatasourceSustainTypeAnd // 同时满足 所有数据告警集合一致

	// MultiDatasourceSustainTypeOr 其中一个满足 数据告警集合其中一个完全满足
	MultiDatasourceSustainTypeOr // 其中一个满足 数据告警集合其中一个完全满足

	// MultiDatasourceSustainTypeAndOr 共同满足 所有数据告警集合合并起来后满足
	MultiDatasourceSustainTypeAndOr // 共同满足 所有数据告警集合合并起来后满足
)
