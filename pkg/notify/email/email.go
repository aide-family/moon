package email

import (
	"context"

	"github.com/aide-cloud/moon/pkg/notify"
	"github.com/go-kratos/kratos/v2/log"
)

var _ notify.Notify = (*Type)(nil)

type Type string

func (l Type) Send(ctx context.Context, msg string) error {
	log.Debugw("send email", "email", l, "msg", msg)
	return nil
}
