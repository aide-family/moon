package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

func NewEmail(
	emailConfigBiz *EmailConfig,
	templateBiz *Template,
	messageLogBiz *MessageLog,
	helper *klog.Helper,
) *Email {
	return &Email{
		emailConfigBiz: emailConfigBiz,
		messageLogBiz:  messageLogBiz,
		templateBiz:    templateBiz,
		helper:         klog.NewHelper(klog.With(helper.Logger(), "biz", "email")),
	}
}

type Email struct {
	emailConfigBiz *EmailConfig
	templateBiz    *Template
	messageLogBiz  *MessageLog
	helper         *klog.Helper
}

func (e *Email) AppendEmailMessage(ctx context.Context, req *bo.SendEmailBo) (snowflake.ID, error) {
	// 获取邮箱配置
	emailConfig, err := e.emailConfigBiz.GetEmailConfig(ctx, req.UID)
	if err != nil {
		return 0, err
	}
	messageLog, err := req.ToMessageLog(emailConfig)
	if err != nil {
		e.helper.Errorw("msg", "create message log failed", "error", err)
		return 0, merr.ErrorInternalServer("generate message log failed").WithCause(err)
	}
	uid, err := e.messageLogBiz.createMessageLog(ctx, messageLog)
	if err != nil {
		e.helper.Errorw("msg", "create message log failed", "error", err)
		return 0, merr.ErrorInternalServer("create message log failed").WithCause(err)
	}
	return uid, nil
}

func (e *Email) AppendEmailMessageWithTemplate(ctx context.Context, req *bo.SendEmailWithTemplateBo) (snowflake.ID, error) {
	// 获取模板
	templateBo, err := e.templateBiz.GetTemplate(ctx, req.TemplateUID)
	if err != nil {
		return 0, err
	}
	sendEmailBo, err := req.ToSendEmailBo(templateBo)
	if err != nil {
		e.helper.Errorw("msg", "convert template to email template data failed", "error", err)
		return 0, merr.ErrorInternalServer("convert template to email template data failed")
	}
	return e.AppendEmailMessage(ctx, sendEmailBo)
}
