package vobj

// Allow 允许范围
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Allow -linecomment
type Allow int

const (
	// AllowNone 禁止
	AllowNone Allow = iota + 1

	// AllowSystem 系统控制
	AllowSystem // 系统控制

	// AllowTeam 团队控制
	AllowTeam

	// AllowUser 用户控制
	AllowUser

	// AllowRBAC RBAC控制
	AllowRBAC // RBAC控制
)
