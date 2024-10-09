package vobj

// InviteType 邀请团队状态
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=InviteType -linecomment
type InviteType int

const (
	// InviteTypeUnknown 未知
	InviteTypeUnknown InviteType = iota // 未知

	// InviteTypeJoined 已加入
	InviteTypeJoined // 已加入

	// InviteTypeUnderReview 邀请中
	InviteTypeUnderReview // 邀请中

	// InviteTypeRejected 已拒绝
	InviteTypeRejected // 已拒绝
)

// GetInviteTypeJoined 获取已加入状态
func GetInviteTypeJoined() InviteType {
	return InviteTypeJoined
}

// GetInviteTypeUnderReview 获取邀请中状态
func GetInviteTypeUnderReview() InviteType {
	return InviteTypeUnderReview
}

// GetInviteTypeRejected 获取已拒绝状态
func GetInviteTypeRejected() InviteType {
	return InviteTypeRejected
}
