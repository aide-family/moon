package vobj

// AlertStatus alert status
//
//go:generate stringer -type=AlertStatus -linecomment -output=status_alert.string.go
type AlertStatus int8

const (
	AlertStatusUnknown AlertStatus = iota
	AlertStatusPending
	AlertStatusFiring
	AlertStatusResolved
)
