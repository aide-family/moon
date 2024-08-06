package vobj

// NotifyType 通知类型
type NotifyType int

const (
	// Phone 电话通知
	Phone = 1 << iota

	// SMS 短信通知
	SMS

	// Email 邮件通知
	Email
)

// IsPhone 是否电话通知
func (n NotifyType) IsPhone() bool {
	return n&Phone != 0
}

// IsSMS 是否短信通知
func (n NotifyType) IsSMS() bool {
	return n&SMS != 0
}

// IsEmail 是否邮件通知
func (n NotifyType) IsEmail() bool {
	return n&Email != 0
}

func (n NotifyType) String() string {
	switch n {
	case Phone:
		return "Phone"
	case SMS:
		return "SMS"
	case Email:
		return "Email"
	default:
		return "Unknown"
	}
}
