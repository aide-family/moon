package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/do"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/rabbit/internal/service/build"
	"github.com/aide-family/moon/pkg/api/rabbit/common"
	rabbitv1 "github.com/aide-family/moon/pkg/api/rabbit/v1"
	"github.com/aide-family/moon/pkg/util/pointer"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type SendService struct {
	rabbitv1.UnimplementedSendServer
	configBiz *biz.Config
	emailBiz  *biz.Email
	smsBiz    *biz.SMS
	hookBiz   *biz.Hook
	lockBiz   *biz.Lock
	helper    *log.Helper
}

func NewSendService(
	configBiz *biz.Config,
	emailBiz *biz.Email,
	smsBiz *biz.SMS,
	hookBiz *biz.Hook,
	lockBiz *biz.Lock,
	logger log.Logger,
) *SendService {
	return &SendService{
		configBiz: configBiz,
		emailBiz:  emailBiz,
		smsBiz:    smsBiz,
		hookBiz:   hookBiz,
		lockBiz:   lockBiz,
		helper:    log.NewHelper(log.With(logger, "module", "service.send")),
	}
}

func (s *SendService) Email(ctx context.Context, req *rabbitv1.SendEmailRequest) (*common.EmptyReply, error) {
	if !s.lockBiz.LockByAPP(ctx, req.GetRequestId(), vobj.APPEmail) {
		return &common.EmptyReply{}, nil
	}
	params := &bo.GetEmailConfigParams{
		TeamID:             req.GetTeamId(),
		Name:               req.ConfigName,
		DefaultEmailConfig: req.EmailConfig,
	}
	emailConfig := s.configBiz.GetEmailConfig(ctx, params)
	if validate.IsNil(emailConfig) || !emailConfig.GetEnable() {
		// no email config
		return &common.EmptyReply{}, nil
	}
	opts := []bo.SendEmailParamsOption{
		bo.WithSendEmailParamsOptionEmail(req.GetEmails()...),
		bo.WithSendEmailParamsOptionBody(req.GetBody()),
		bo.WithSendEmailParamsOptionSubject(req.GetSubject()),
		bo.WithSendEmailParamsOptionContentType(req.GetContentType()),
		bo.WithSendEmailParamsOptionAttachment(req.GetAttachment()),
		bo.WithSendEmailParamsOptionCc(req.GetCc()...),
	}
	sendEmailParams, err := bo.NewSendEmailParams(emailConfig, opts...)
	if err != nil {
		return nil, err
	}
	if err := s.emailBiz.Send(ctx, sendEmailParams); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SendService) Sms(ctx context.Context, req *rabbitv1.SendSmsRequest) (*common.EmptyReply, error) {
	if !s.lockBiz.LockByAPP(ctx, req.GetRequestId(), vobj.APPSms) {
		return &common.EmptyReply{}, nil
	}
	smsConfig, _ := build.ToSMSConfig(req.GetSmsConfig())
	params := &bo.GetSMSConfigParams{
		TeamID:           req.GetTeamId(),
		Name:             req.ConfigName,
		DefaultSMSConfig: smsConfig,
	}
	smsConfig = s.configBiz.GetSMSConfig(ctx, params)
	if validate.IsNil(smsConfig) || !smsConfig.GetEnable() {
		// no sms config
		return &common.EmptyReply{}, nil
	}
	opts := []bo.SendSMSParamsOption{
		bo.WithSendSMSParamsOptionPhoneNumbers(req.GetPhones()...),
		bo.WithSendSMSParamsOptionTemplateParam(req.GetTemplateParameters()),
		bo.WithSendSMSParamsOptionTemplateCode(req.GetTemplateCode()),
	}
	sendSMSParams, err := bo.NewSendSMSParams(smsConfig, opts...)
	if err != nil {
		return nil, err
	}
	if err := s.smsBiz.Send(ctx, sendSMSParams); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SendService) Hook(ctx context.Context, req *rabbitv1.SendHookRequest) (*common.EmptyReply, error) {
	hookConfigs := slices.MapFilter(req.GetHooks(), func(hookItem *common.HookConfig) (bo.HookConfig, bool) {
		if !s.lockBiz.LockByAPP(ctx, req.GetRequestId(), build.ToSendHookAPP(hookItem.GetApp())) {
			return nil, false
		}
		opts := []do.HookConfigOption{
			do.WithHookConfigOptionApp(hookItem.App),
			do.WithHookConfigOptionEnable(hookItem.Enable),
			do.WithHookConfigOptionHeaders(hookItem.Headers),
			do.WithHookConfigOptionName(hookItem.Name),
			do.WithHookConfigOptionPassword(hookItem.Password),
			do.WithHookConfigOptionSecret(hookItem.Secret),
			do.WithHookConfigOptionToken(hookItem.Token),
			do.WithHookConfigOptionUsername(hookItem.Username),
		}
		hookConfig, err := do.NewHookConfig(hookItem.Url, opts...)
		if err != nil {
			return nil, false
		}
		params := &bo.GetHookConfigParams{
			TeamID:            req.GetTeamId(),
			Name:              pointer.Of(hookItem.Name),
			DefaultHookConfig: hookConfig,
		}
		hookConfigDo := s.configBiz.GetHookConfig(ctx, params)
		if validate.IsNil(hookConfigDo) || !hookConfigDo.GetEnable() {
			return nil, false
		}
		return hookConfig, true
	})
	if len(hookConfigs) == 0 || len(req.GetBody()) == 0 {
		return &common.EmptyReply{}, nil
	}
	bodyMap := make([]*bo.HookBody, 0, len(req.GetBody()))
	for _, body := range req.GetBody() {
		bodyMap = append(bodyMap, &bo.HookBody{
			AppName: body.GetAppName(),
			Body:    []byte(body.GetBody()),
		})
	}
	opts := []bo.SendHookParamsOption{
		bo.WithSendHookParamsOptionBody(bodyMap),
	}
	sendHookParams, err := bo.NewSendHookParams(hookConfigs, opts...)
	if err != nil {
		return nil, err
	}
	if err := s.hookBiz.Send(ctx, sendHookParams); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}
