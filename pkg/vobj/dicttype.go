package vobj

// DictType 字典类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=DictType -linecomment
type DictType int32

const (
	DictTypeUnknown DictType = iota // 未知

	DictTypePromLabel // 规则标签

	DictTypePromAnnotation // 规则告警描述

	DictTypePromStrategy // 规则

	DictTypePromStrategyGroup // 规则组

	DictTypeAlarmLevel // 告警级别

	DictTypeAlarmStatus // 告警状态

	DictTypeNotifyType // 通知类型

	DictTypeAlarmPage // 告警页面
)
