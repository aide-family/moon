package phone

import (
	"context"

	"github.com/aide-family/moon/pkg/notify"
	"github.com/go-kratos/kratos/v2/log"
)

type (
	// Phone is a phone service.
	Phone interface {
		notify.Notify
		Call(ctx context.Context) error
	}

	p struct {
	}
)

func (l *p) Send(ctx context.Context, msg notify.Msg) error {
	log.Debugw("send phone", "phone", l, "msg", msg)
	return nil
}

func (l *p) Call(ctx context.Context) error {
	log.Debugw("call phone", "phone", l)
	return nil
}
