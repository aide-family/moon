package vobj

import (
	"strings"
)

// AlertStatus 告警数据状态
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=AlertStatus -linecomment
type AlertStatus int

const (
	// AlertStatusUnknown 未知
	AlertStatusUnknown AlertStatus = iota // 未知

	// AlertStatusFiring firing
	AlertStatusFiring // firing

	// AlertStatusResolved resolved
	AlertStatusResolved // resolved

	// AlertStatusSilenced silenced
	AlertStatusSilenced // Silenced
)

// ToAlertStatus convert
func ToAlertStatus(s string) AlertStatus {
	switch strings.ToLower(s) {
	case "firing":
		return AlertStatusFiring
	case "resolved":
		return AlertStatusResolved
	case "silenced":
		return AlertStatusSilenced
	default:
		return AlertStatusUnknown
	}
}
