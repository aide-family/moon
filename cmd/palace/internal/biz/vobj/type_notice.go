package vobj

// NoticeType notice type
//
//go:generate stringer -type=NoticeType -linecomment -output=type_notice.string.go
type NoticeType int8

const (
	NoticeTypeUnknown NoticeType = 0               // NONE
	NoticeTypeEmail   NoticeType = 1 << (iota - 1) // Email
	NoticeTypeSMS                                  // SMS
	NoticeTypeVoice                                // Voice
)

// IsContainsEmail is contains email notice type
func (n NoticeType) IsContainsEmail() bool {
	return n&NoticeTypeEmail == NoticeTypeEmail
}

// IsContainsSMS is contains sms notice type
func (n NoticeType) IsContainsSMS() bool {
	return n&NoticeTypeSMS == NoticeTypeSMS
}

// IsContainsVoice is contains voice notice type
func (n NoticeType) IsContainsVoice() bool {
	return n&NoticeTypeVoice == NoticeTypeVoice
}

// IsContainsAll is contains all notice type
func (n NoticeType) IsContainsAll() bool {
	return n.IsContainsEmail() && n.IsContainsSMS() && n.IsContainsVoice()
}

// List is list of notice type
func (n NoticeType) List() []NoticeType {
	notices := make([]NoticeType, 0, 3)
	if n.IsContainsEmail() {
		notices = append(notices, NoticeTypeEmail)
	}
	if n.IsContainsSMS() {
		notices = append(notices, NoticeTypeSMS)
	}
	if n.IsContainsVoice() {
		notices = append(notices, NoticeTypeVoice)
	}
	return notices
}
