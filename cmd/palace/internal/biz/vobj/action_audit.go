// Package vobj is a value object package for kratos.
package vobj

//go:generate stringer -type=AuditAction -linecomment -output=action_audit.string.go
type AuditAction int8

const (
	AuditActionUnknown AuditAction = iota // Unknown
	AuditActionJoin                       // Join
	AuditActionLeave                      // Leave
)
