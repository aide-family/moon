package email

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/util/email"
	"github.com/aide-family/moon/pkg/util/format"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/notify"
)

var _ notify.Notify = (*e)(nil)

// New a notify service with options.
func New(cfg email.Config, to *api.ReceiverEmail) notify.Notify {
	cli := email.New(cfg)
	cli.SetAttach(to.GetAttachUrl()...).
		SetTo(to.GetTo()).
		SetCc(to.GetCc()...).
		SetSubject(to.GetSubject())

	return &e{
		Config: cfg,
		to:     to,
		cli:    cli,
	}
}

type e struct {
	email.Config
	to  *api.ReceiverEmail
	cli email.Interface
}

func (l *e) Send(ctx context.Context, msg notify.Msg) error {
	log.Debugw("send email", "email", l, "msg", msg)
	body := l.to.GetContent()
	if l.to.GetTemplate() != "" {
		body = l.to.GetTemplate()
	}
	body = format.Formatter(body, msg)
	l.cli.SetBody(body, l.to.GetContentType())
	return l.cli.Send()
}
