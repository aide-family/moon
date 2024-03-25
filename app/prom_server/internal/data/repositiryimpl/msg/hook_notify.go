package msg

import (
	"context"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

type HookNotifyMsg struct {
	Content   string      `json:"content"`
	AlarmInfo *bo.AlertBo `json:"-"`
	Secret    string      `json:"-"`
}

type HookNotify interface {
	Alarm(ctx context.Context, url string, msg *HookNotifyMsg) error
}

func NewHookNotify(app vobj.NotifyApp) HookNotify {
	switch app {
	case vobj.NotifyAppWeChatWork:
		return NewWechatNotify()
	case vobj.NotifyAppDingDing:
		return NewDingNotify()
	case vobj.NotifyAppFeiShu:
		return NewFeishuNotify()
	default:
		return NewOtherNotify()
	}
}
