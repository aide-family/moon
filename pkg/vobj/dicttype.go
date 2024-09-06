package vobj

// DictType 字典类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=DictType -linecomment
type DictType int

const (
	// DictTypeUnknown 未知
	DictTypeUnknown DictType = iota // 未知

	// DictTypeStrategyCategory 策略类目
	DictTypeStrategyCategory // 策略类目

	// DictTypeStrategyGroupCategory 策略组类目
	DictTypeStrategyGroupCategory // 策略组类目

	// DictTypeAlarmLevel 告警级别
	DictTypeAlarmLevel // 告警级别

	// DictTypeAlarmPage 告警页面
	DictTypeAlarmPage // 告警页面
)
