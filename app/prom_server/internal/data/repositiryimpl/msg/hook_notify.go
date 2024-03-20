package msg

import (
	"context"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type HookNotifyMsg struct {
	Content   string      `json:"content"`
	AlarmInfo *bo.AlertBo `json:"-"`
	Secret    string      `json:"-"`
}

type HookNotify interface {
	Alarm(ctx context.Context, url string, msg *HookNotifyMsg) error
}

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
