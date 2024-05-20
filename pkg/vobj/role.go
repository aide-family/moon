package vobj

import (
	"database/sql/driver"
)

var _ driver.Valuer = Role(0)

// Role 系统全局角色
//
//go:generate stringer -type=Role -linecomment
type Role int

func (i Role) Value() (driver.Value, error) {
	return i, nil
}

const (
	RoleAll Role = iota // 未知

	RoleSuperAdmin // 超级管理员

	RoleAdmin // 管理员

	RoleUser // 普通用户
)
