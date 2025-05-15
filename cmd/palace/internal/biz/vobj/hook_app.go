package vobj

// HookApp is the hook type for app
//
//go:generate stringer -type=HookApp -linecomment -output=hook_app.string.go
type HookApp int8

const (
	HookAppUnknown  HookApp = iota // unknown
	HookAppOther                   // other
	HookAppDingTalk                // dingtalk
	HookAppWechat                  // wechat
	HookAppFeiShu                  // feishu
)
