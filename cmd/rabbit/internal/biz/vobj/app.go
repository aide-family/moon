package vobj

// APP is the application object
//
//go:generate stringer -type=APP -linecomment -output=app.string.go
type APP int8

const (
	APPUnknown      APP = iota // unknown
	APPEmail                   // email
	APPSms                     // sms
	APPHookOther               // hook:other
	APPHookDingTalk            // hook:dingtalk
	APPHookWechat              // hook:wechat
	APPHookFeiShu              // hook:feishu
)
