package vobj

import (
	"database/sql/driver"
)

//var _ driver.Valuer = Status(0)

// Status 数据状态
//
//go:generate stringer -type=Status -linecomment
type Status int

func (i Status) Value() (driver.Value, error) {
	return i, nil
}

const (
	// StatusUnknown 未知
	StatusUnknown Status = iota // 未知

	// StatusEnable 启用
	StatusEnable // 启用

	// StatusDisable 禁用
	StatusDisable // 禁用
)
