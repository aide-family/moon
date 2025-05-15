package vobj

// MemberStatus team member status
//
//go:generate stringer -type=MemberStatus -linecomment -output=status_member.string.go
type MemberStatus int8

const (
	MemberStatusUnknown        MemberStatus = iota // unknown
	MemberStatusNormal                             // normal
	MemberStatusForbidden                          // forbidden
	MemberStatusDeleted                            // deleted
	MemberStatusPendingConfirm                     // pending-confirm
	MemberStatusDeparted                           // departed
)
