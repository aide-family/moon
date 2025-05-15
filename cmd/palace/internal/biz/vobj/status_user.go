package vobj

// UserStatus user status
//
//go:generate stringer -type=UserStatus -linecomment -output=status_user.string.go
type UserStatus int8

const (
	UserStatusUnknown   UserStatus = iota // unknown
	UserStatusNormal                      // normal
	UserStatusForbidden                   // forbidden
	UserStatusDeleted                     // deleted
)
