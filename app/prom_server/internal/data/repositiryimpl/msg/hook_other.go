package msg

import (
	"context"
	"io"

	"github.com/aide-family/moon/pkg/httpx"
	"github.com/go-kratos/kratos/v2/log"
)

var _ HookNotify = (*otherNotify)(nil)

type otherNotify struct{}

func (l *otherNotify) Alarm(ctx context.Context, url string, msg *HookNotifyMsg) error {
	response, err := httpx.NewHttpX().POSTWithContext(ctx, url, msg.AlarmInfo.Bytes())
	if err != nil {
		return err
	}
	body := response.Body
	defer body.Close()
	resBytes, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes))
	return err
}

func NewOtherNotify() HookNotify {
	return &otherNotify{}
}
