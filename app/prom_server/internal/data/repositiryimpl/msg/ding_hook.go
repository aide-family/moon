package msg

import (
	"context"
	"io"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/httpx"
)

var _ HookNotify = (*dingNotify)(nil)

type dingNotify struct{}

func (l *dingNotify) Alarm(ctx context.Context, url string, msg *HookNotifyMsg) error {
	response, err := httpx.NewHttpX().POSTWithContext(ctx, url, []byte(msg.Content))
	body := response.Body
	resBytes, err := io.ReadAll(body)
	defer body.Close()
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes))
	return err
}

func NewDingNotify() HookNotify {
	return &dingNotify{}
}
