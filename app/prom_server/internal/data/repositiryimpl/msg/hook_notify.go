package msg

import (
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type HookNotifyMsg struct {
	Context   string      `json:"context"`
	Title     string      `json:"title"`
	AlarmInfo *bo.AlertBo `json:"-"`
	HookBytes []byte      `json:"-"`
}

type HookNotify interface {
	Alarm(url string, msg *HookNotifyMsg) error
}

const (
	markdown = "markdown"
)

func NewHookNotify(app vo.NotifyApp) HookNotify {
	switch app {
	case vo.NotifyAppWeChatWork:
		return NewWechatNotify()
	case vo.NotifyAppDingDing:
		return NewDingNotify()
	default:
		return NewOtherNotify()
	}
}
