package msg

import (
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type HookNotifyMsg struct {
	Content   string      `json:"content"`
	Title     string      `json:"title"`
	AlarmInfo *bo.AlertBo `json:"-"`
	HookBytes []byte      `json:"-"`
	Secret    string      `json:"-"`
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
	case vo.NotifyAppFeiShu:
		return NewFeishuNotify()
	default:
		return NewOtherNotify()
	}
}
