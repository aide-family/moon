// Package vobj provides the value objects for the SMS.
package vobj

// SMSProviderType sms provider type
//
//go:generate stringer -type=SMSProviderType -linecomment -output=sms_type.string.go
type SMSProviderType int8

const (
	SMSProviderTypeUnknown SMSProviderType = iota // unknown
	SMSProviderTypeAliyun                         // aliyun
)
