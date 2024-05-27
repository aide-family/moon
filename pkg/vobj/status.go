package vobj

// Status 数据状态
//
//go:generate stringer -type=Status -linecomment
type Status int

const (
	// StatusUnknown 未知
	StatusUnknown Status = iota // 未知

	// StatusEnable 启用
	StatusEnable // 启用

	// StatusDisable 禁用
	StatusDisable // 禁用
)
