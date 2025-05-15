package vobj

// SMSProviderType sms provider
//
//go:generate stringer -type=SMSProviderType -linecomment -output=type_provider_sms.string.go
type SMSProviderType int8

const (
	SMSProviderTypeUnknown SMSProviderType = iota // Unknown
	SMSProviderTypeAliyun                         // Aliyun
	SMSProviderTypeTencent                        // Tencent
	SMSProviderTypeTwilio                         // Twilio
)
