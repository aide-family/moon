package msg

import (
	"prometheus-manager/pkg/httpx"
)

var _ HookNotify = (*otherNotify)(nil)

type otherNotify struct{}

func (l *otherNotify) Alarm(url string, msg *HookNotifyMsg) error {
	_, err := httpx.NewHttpX().POST(url, msg.HookBytes)
	return err
}

func NewOtherNotify() HookNotify {
	return &otherNotify{}
}
