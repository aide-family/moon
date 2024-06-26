package vobj

// Role 系统全局角色
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Role -linecomment
type Role int

const (
	RoleAll Role = iota // 未知

	RoleSuperAdmin // 超级管理员

	RoleAdmin // 管理员

	RoleUser // 普通用户
)
