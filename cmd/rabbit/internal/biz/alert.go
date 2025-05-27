package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/template"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewAlert(
	configRepo repository.Config,
	sendRepo repository.Send,
	logger log.Logger,
) *Alert {
	return &Alert{
		configRepo: configRepo,
		sendRepo:   sendRepo,
		helper:     log.NewHelper(log.With(logger, "module", "biz.alert")),
	}
}

type Alert struct {
	configRepo repository.Config
	sendRepo   repository.Send

	helper *log.Helper
}

func (a *Alert) SendAlert(ctx context.Context, alert *bo.AlertsItem) error {
	if validate.IsNil(alert) {
		return merr.ErrorParams("No alert is available")
	}
	receivers := alert.GetReceiver()
	if len(receivers) == 0 {
		return merr.ErrorParams("No receiver is available")
	}

	for _, receiver := range receivers {
		noticeGroupConfig, ok := a.configRepo.GetNoticeGroupConfig(ctx, alert.GetTeamID(), receiver)
		if !ok || validate.IsNil(noticeGroupConfig) {
			continue
		}
		safety.Go(ctx, func(ctx context.Context) error {
			a.sendEmail(ctx, noticeGroupConfig, alert)
			return nil
		})
		safety.Go(ctx, func(ctx context.Context) error {
			a.sendSms(ctx, noticeGroupConfig, alert)
			return nil
		})
		safety.Go(ctx, func(ctx context.Context) error {
			a.sendHook(ctx, noticeGroupConfig, alert)
			return nil
		})
	}
	return nil
}

func (a *Alert) sendEmail(ctx context.Context, noticeGroupConfig bo.NoticeGroup, alert *bo.AlertsItem) {
	emailNames := noticeGroupConfig.GetEmailUserNames()
	if len(emailNames) == 0 {
		return
	}
	userConfigs, err := a.configRepo.GetNoticeUserConfigs(ctx, alert.GetTeamID(), emailNames...)
	if err != nil {
		a.helper.WithContext(ctx).Warnw("method", "GetNoticeUserConfigs", "err", err)
		return
	}
	emails := slices.MapFilter(userConfigs, func(userConfig bo.NoticeUser) (string, bool) {
		if email := userConfig.GetEmail(); validate.TextIsNotNull(email) {
			return email, true
		}
		return "", false
	})
	if len(emails) == 0 {
		return
	}
	emailConfig, ok := a.configRepo.GetEmailConfig(ctx, alert.GetTeamID(), noticeGroupConfig.GetEmailConfigName())
	if !ok || validate.IsNil(emailConfig) {
		return
	}
	emailTemplate := noticeGroupConfig.GetEmailTemplate()
	for _, alertItem := range alert.Alerts {
		opts := []bo.SendEmailParamsOption{
			bo.WithSendEmailParamsOptionEmail(emails...),
			bo.WithSendEmailParamsOptionSubject(a.getEmailSubject(emailTemplate.GetSubject(), alertItem)),
			bo.WithSendEmailParamsOptionBody(a.getEmailBody(emailTemplate.GetTemplate(), alertItem)),
		}
		sendEmailParams, err := bo.NewSendEmailParams(emailConfig, opts...)
		if err != nil {
			a.helper.WithContext(ctx).Warnw("method", "NewSendEmailParams", "err", err)
			continue
		}
		safety.Go(ctx, func(ctx context.Context) error {
			return a.sendRepo.Email(ctx, sendEmailParams)
		})
	}
}

func (a *Alert) sendSms(ctx context.Context, noticeGroupConfig bo.NoticeGroup, alert *bo.AlertsItem) {
	smsNames := noticeGroupConfig.GetSmsUserNames()
	if len(smsNames) == 0 {
		return
	}
	userConfigs, err := a.configRepo.GetNoticeUserConfigs(ctx, alert.GetTeamID(), smsNames...)
	if err != nil {
		a.helper.WithContext(ctx).Warnw("method", "GetNoticeUserConfigs", "err", err)
		return
	}
	phoneNumbers := slices.MapFilter(userConfigs, func(userConfig bo.NoticeUser) (string, bool) {
		if phone := userConfig.GetPhone(); validate.TextIsNotNull(phone) {
			return phone, true
		}
		return "", false
	})
	if len(phoneNumbers) == 0 {
		return
	}
	smsConfig, ok := a.configRepo.GetSMSConfig(ctx, alert.GetTeamID(), noticeGroupConfig.GetSmsConfigName())
	if !ok || validate.IsNil(smsConfig) {
		return
	}
	smsTemplate := noticeGroupConfig.GetSmsTemplate()
	for _, alertItem := range alert.Alerts {
		opts := []bo.SendSMSParamsOption{
			bo.WithSendSMSParamsOptionPhoneNumbers(phoneNumbers...),
			bo.WithSendSMSParamsOptionTemplateParam(a.getSmsBody(smsTemplate.GetTemplateParameters(), alertItem)),
			bo.WithSendSMSParamsOptionTemplateCode(smsTemplate.GetTemplate()),
		}
		sendSMSParams, err := bo.NewSendSMSParams(smsConfig, opts...)
		if err != nil {
			a.helper.WithContext(ctx).Warnw("method", "NewSendSMSParams", "err", err)
			continue
		}
		safety.Go(ctx, func(ctx context.Context) error {
			return a.sendRepo.SMS(ctx, sendSMSParams)
		})
	}
}
func (a *Alert) sendHook(ctx context.Context, noticeGroupConfig bo.NoticeGroup, alert *bo.AlertsItem) {
	hookNames := noticeGroupConfig.GetHookConfigNames()
	if len(hookNames) == 0 {
		return
	}
	hookConfigs := make([]bo.HookConfig, 0, len(hookNames))
	body := make([]*bo.HookBody, 0, len(hookNames)*len(alert.Alerts))

	for _, hookName := range hookNames {
		hookConfig, ok := a.configRepo.GetHookConfig(ctx, alert.GetTeamID(), hookName)
		if !ok || validate.IsNil(hookConfig) {
			continue
		}
		hookConfigs = append(hookConfigs, hookConfig)
		for _, alertItem := range alert.Alerts {
			body = append(body, &bo.HookBody{
				AppName: hookConfig.GetName(),
				Body:    a.getHookBody(noticeGroupConfig.GetHookTemplate(hookConfig.GetApp()), alertItem),
			})
		}
	}
	if len(body) == 0 {
		return
	}
	sendParamsOpts := []bo.SendHookParamsOption{
		bo.WithSendHookParamsOptionBody(body),
	}
	sendHookParams, err := bo.NewSendHookParams(hookConfigs, sendParamsOpts...)
	if err != nil {
		a.helper.WithContext(ctx).Warnw("method", "NewSendHookParams", "err", err)
		return
	}
	if err := a.sendRepo.Hook(ctx, sendHookParams); err != nil {
		a.helper.WithContext(ctx).Warnw("method", "sendRepo.Hook", "err", err)
	}
}

func (a *Alert) getHookBody(temp string, alert *bo.AlertItem) []byte {
	return []byte(template.TextFormatterX(temp, alert))
}

func (a *Alert) getEmailBody(temp string, alert *bo.AlertItem) string {
	return template.HtmlFormatterX(temp, alert)
}

func (a *Alert) getEmailSubject(temp string, alert *bo.AlertItem) string {
	return template.TextFormatterX(temp, alert)
}

func (a *Alert) getSmsBody(temp string, alert *bo.AlertItem) string {
	return template.TextFormatterX(temp, alert)
}
