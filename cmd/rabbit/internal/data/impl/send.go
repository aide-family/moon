package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/cmd/rabbit/internal/data"
	"github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/email"
	"github.com/aide-family/moon/pkg/plugin/hook"
	"github.com/aide-family/moon/pkg/plugin/sms"
	"github.com/aide-family/moon/pkg/plugin/sms/ali"
)

func NewSendRepo(d *data.Data, logger log.Logger) repository.Send {
	return &sendImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.send")),
	}
}

type sendImpl struct {
	*data.Data
	helper *log.Helper
}

func (s *sendImpl) Email(_ context.Context, params bo.SendEmailParams) error {
	if len(params.GetEmails()) == 0 {
		return merr.ErrorParamsError("No email is available")
	}
	emailInstance, ok := s.GetEmail(params.GetConfig().GetName())
	if !ok {
		emailInstance = email.New(params.GetConfig())
		s.SetEmail(params.GetConfig().GetName(), emailInstance)
	}

	emailInstance.SetTo(params.GetEmails()...).
		SetSubject(params.GetSubject()).
		SetBody(params.GetBody())
	if params.GetAttachment() != "" {
		emailInstance.SetAttach(params.GetAttachment())
	}
	if len(params.GetCc()) > 0 {
		emailInstance.SetCc(params.GetCc()...)
	}
	return emailInstance.Send()
}

func (s *sendImpl) SMS(ctx context.Context, params bo.SendSMSParams) error {
	if len(params.GetPhoneNumbers()) == 0 {
		return merr.ErrorParamsError("No phone number is available")
	}
	var err error
	smsInstance, ok := s.GetSms(params.GetConfig().GetName())
	if !ok {
		smsInstance, err = s.newSms(params.GetConfig())
		if err != nil {
			return err
		}
		s.SetSms(params.GetConfig().GetName(), smsInstance)
	}
	message := sms.Message{
		TemplateCode:  params.GetTemplateCode(),
		TemplateParam: params.GetGetTemplateParam(),
	}
	if len(params.GetPhoneNumbers()) == 1 {
		return smsInstance.Send(ctx, params.GetPhoneNumbers()[0], message)
	}
	return smsInstance.SendBatch(ctx, params.GetPhoneNumbers(), message)
}

func (s *sendImpl) Hook(ctx context.Context, params bo.SendHookParams) error {
	var err error
	hooks := make(map[string]hook.Sender)
	for _, configItem := range params.GetConfigs() {
		hookInstance, ok := s.GetHook(configItem.GetName())
		if !ok {
			hookInstance, err = s.newHook(configItem)
			if err != nil {
				s.helper.WithContext(ctx).Warnw("method", "newHook", "err", err)
				continue
			}
			s.SetHook(configItem.GetName(), hookInstance)
		}
		hooks[configItem.GetName()] = hookInstance
	}

	if len(hooks) == 0 {
		return merr.ErrorParamsError("No hook is available")
	}
	bodyMap := params.GetBody()
	eg := new(errgroup.Group)
	for _, body := range bodyMap {
		bodyItem := body
		eg.Go(func() error {
			sender, ok := hooks[bodyItem.AppName]
			if !ok {
				return merr.ErrorParamsError("No hook is available")
			}
			return sender.Send(ctx, bodyItem.Body)
		})
	}

	return eg.Wait()
}

func (s *sendImpl) newSms(config bo.SMSConfig) (sms.Sender, error) {
	switch config.GetType() {
	case common.SMSConfig_ALIYUN:
		return ali.NewAliyun(config, ali.WithAliyunLogger(s.helper.Logger()))
	default:
		return nil, merr.ErrorParamsError("No SMS configuration is available")
	}
}

func (s *sendImpl) newHook(config bo.HookConfig) (hook.Sender, error) {
	switch config.GetApp() {
	case common.HookAPP_OTHER:
		opts := []hook.OtherHookOption{
			hook.WithOtherBasicAuth(config.GetUsername(), config.GetPassword()),
			hook.WithOtherLogger(s.helper.Logger()),
			hook.WithOtherHeader(config.GetHeaders()),
		}
		return hook.NewOtherHook(config.GetUrl(), opts...), nil
	case common.HookAPP_DINGTALK:
		opts := []hook.DingTalkHookOption{
			hook.WithDingTalkLogger(s.helper.Logger()),
		}
		return hook.NewDingTalkHook(config.GetUrl(), config.GetSecret(), opts...), nil
	case common.HookAPP_WECHAT:
		opts := []hook.WechatHookOption{
			hook.WithWechatLogger(s.helper.Logger()),
		}
		return hook.NewWechatHook(config.GetUrl(), opts...), nil
	case common.HookAPP_FEISHU:
		opts := []hook.FeishuHookOption{
			hook.WithFeishuLogger(s.helper.Logger()),
		}
		return hook.NewFeishuHook(config.GetUrl(), config.GetSecret(), opts...), nil
	default:
		return nil, merr.ErrorParamsError("No hook configuration is available")
	}
}
