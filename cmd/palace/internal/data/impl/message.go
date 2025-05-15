package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/conf"
	"github.com/moon-monitor/moon/pkg/api/rabbit/common"
	rabbitv1 "github.com/moon-monitor/moon/pkg/api/rabbit/v1"
	"github.com/moon-monitor/moon/pkg/plugin/email"
)

func NewSendMessageRepo(
	bc *conf.Bootstrap,
	rabbitRepo repository.Rabbit,
	logger log.Logger,
) repository.SendMessage {
	emailConfig := bc.GetEmail()
	return &sendMessageImpl{
		helper:     log.NewHelper(log.With(logger, "module", "data.impl.sendMessage")),
		rabbitRepo: rabbitRepo,
		emailConfig: &common.EmailConfig{
			User:   emailConfig.GetUser(),
			Pass:   emailConfig.GetPass(),
			Host:   emailConfig.GetHost(),
			Port:   emailConfig.GetPort(),
			Enable: true,
			Name:   emailConfig.GetName(),
		},
	}
}

type sendMessageImpl struct {
	helper      *log.Helper
	rabbitRepo  repository.Rabbit
	emailConfig *common.EmailConfig
}

// SendEmail implements repository.SendMessage.
func (s *sendMessageImpl) SendEmail(ctx context.Context, params *bo.SendEmailParams) error {
	sendClient, ok := s.rabbitRepo.Send()
	if !ok {
		// call local send email
		return s.localSendEmail(ctx, params)
	}
	// call rabbit server send email
	return s.rabbitSendEmail(ctx, sendClient, params)
}

func (s *sendMessageImpl) localSendEmail(ctx context.Context, params *bo.SendEmailParams) error {
	emailInstance := email.New(s.emailConfig)
	emailInstance.SetTo(params.Email).
		SetSubject(params.Subject).
		SetBody(params.Body, params.ContentType)
	if err := emailInstance.Send(); err != nil {
		s.helper.WithContext(ctx).Warnw("method", "local send email error", "params", params, "error", err)
		return err
	}
	return nil
}

func (s *sendMessageImpl) rabbitSendEmail(ctx context.Context, client repository.SendClient, params *bo.SendEmailParams) error {
	reply, err := client.Email(ctx, &rabbitv1.SendEmailRequest{
		Emails:      []string{params.Email},
		Body:        params.Body,
		Subject:     params.Subject,
		ContentType: params.ContentType,
		EmailConfig: s.emailConfig,
		RequestId:   "",
		Attachment:  "",
		Cc:          []string{},
		ConfigName:  new(string),
		TeamId:      "",
	})
	if err != nil {
		s.helper.WithContext(ctx).Warnw("method", "rabbit send email error", "params", params, "error", err)
		return err
	}
	s.helper.WithContext(ctx).Debugf("send email reply: %v", reply)
	return nil
}
