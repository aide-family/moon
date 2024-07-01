package vobj

// AlertStatus 告警数据状态
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=AlertStatus -linecomment
type AlertStatus int

const (
	AlertStatusUnknown AlertStatus = iota // 未知

	AlertStatusFiring // firing

	AlertStatusResolved // resolved

	AlertStatusSilenced // Silenced
)
