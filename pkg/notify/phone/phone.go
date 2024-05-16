package phone

import (
	"context"

	"github.com/aide-cloud/moon/pkg/notify"
	"github.com/go-kratos/kratos/v2/log"
)

var _ Phone = (*Type)(nil)

type (
	Phone interface {
		notify.Notify
		Call(ctx context.Context) error
	}

	Type string
)

func (l Type) Send(ctx context.Context, msg notify.Msg) error {
	log.Debugw("send phone", "phone", l, "msg", msg)
	return nil
}

func (l Type) Call(ctx context.Context) error {
	log.Debugw("call phone", "phone", l)
	return nil
}
