package vobj

// BizType 业务类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=BizType -linecomment
type BizType int

const (
	// BizTypeUnknown 未知
	BizTypeUnknown BizType = iota // unknown

	// BizTypeInvitation 邀请
	BizTypeInvitation // invitation

	// BizTypeInvitationRejected 邀请被拒绝
	BizTypeInvitationRejected // invitation_rejected

	// BizTypeInvitationAccepted 邀请被接受
	BizTypeInvitationAccepted // invitation_accepted
)
