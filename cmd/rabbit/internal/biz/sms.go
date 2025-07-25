package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewSMS(sendRepo repository.Send, logger log.Logger) *SMS {
	return &SMS{
		helper: log.NewHelper(log.With(logger, "module", "biz.sms")),
		send:   sendRepo,
	}
}

type SMS struct {
	helper *log.Helper
	send   repository.Send
}

func (s *SMS) Send(ctx context.Context, params bo.SendSMSParams) error {
	if validate.IsNil(params.GetConfig()) {
		return merr.ErrorParams("No sms configuration is available")
	}
	return s.send.SMS(ctx, params)
}
