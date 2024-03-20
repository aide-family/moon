package msg

import (
	"context"
	"io"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/httpx"
)

var _ HookNotify = (*otherNotify)(nil)

type otherNotify struct{}

func (l *otherNotify) Alarm(ctx context.Context, url string, msg *HookNotifyMsg) error {
	response, err := httpx.NewHttpX().POSTWithContext(ctx, url, msg.AlarmInfo.Bytes())
	body := response.Body
	resBytes, err := io.ReadAll(body)
	defer body.Close()
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes))
	return err
}

func NewOtherNotify() HookNotify {
	return &otherNotify{}
}
