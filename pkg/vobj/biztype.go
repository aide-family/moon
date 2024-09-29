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
)
