package vobj

// SendMessageStatus send message status
//
//go:generate stringer -type=SendMessageStatus -linecomment -output=status_send_message.string.go
type SendMessageStatus int8

const (
	SendMessageStatusUnknown SendMessageStatus = iota // unknown
	SendMessageStatusSuccess                          // success
	SendMessageStatusFailed                           // failed
	SendMessageStatusPending                          // pending
	SendMessageStatusSending                          // sending
	SendMessageStatusRetry                            // retry
)
