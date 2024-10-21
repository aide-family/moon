package vobj

// HookAPP hook app 类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=HookAPP -linecomment
type HookAPP int

const (
	// HookAPPUnknown 未知
	HookAPPUnknown HookAPP = iota // 未知

	// HookAPPWebHook 自定义
	HookAPPWebHook // 自定义

	// HookAPPDingTalk 钉钉
	HookAPPDingTalk // 钉钉

	// HookAPPWeChat 企业微信
	HookAPPWeChat // 企业微信

	// HookAPPFeiShu 飞书
	HookAPPFeiShu // 飞书
)

func (h HookAPP) EnUSString() string {
	switch h {
	case HookAPPDingTalk:
		return "dingtalk"
	case HookAPPWeChat:
		return "wechat"
	case HookAPPFeiShu:
		return "feishu"
	default:
		return "other"
	}
}
