package vobj

// InviteType 邀请团队状态
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=InviteType -linecomment
type InviteType int

const (
	// InviteTypeUnknown 未知
	InviteTypeUnknown InviteType = iota // 未知

	// InviteTypeJoined 加入
	InviteTypeJoined // 加入

	// InviteTypeUnderReview 邀请中
	InviteTypeUnderReview // 邀请中

	// InviteTypeRejected 已拒绝
	InviteTypeRejected // 已拒绝
)
