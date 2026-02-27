package biz

import (
	"context"
	"errors"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

func NewWebhook(
	webhookConfigBiz *WebhookConfig,
	messageLogBiz *MessageLog,
	templateBiz *Template,
	helper *klog.Helper,
) *Webhook {
	return &Webhook{
		webhookConfigBiz: webhookConfigBiz,
		messageLogBiz:    messageLogBiz,
		templateBiz:      templateBiz,
		helper:           klog.NewHelper(klog.With(helper.Logger(), "biz", "webhook")),
	}
}

type Webhook struct {
	webhookConfigBiz *WebhookConfig
	messageLogBiz    *MessageLog
	templateBiz      *Template
	helper           *klog.Helper
}

func (w *Webhook) AppendWebhookMessage(ctx context.Context, req *bo.SendWebhookBo) (snowflake.ID, error) {
	// 获取webhook配置
	webhookConfig, err := w.webhookConfigBiz.GetWebhook(ctx, req.UID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, merr.ErrorParams("webhook config not found")
		}
		w.helper.Errorw("msg", "get webhook config failed", "error", err)
		return 0, merr.ErrorInternalServer("get webhook config failed").WithCause(err)
	}
	messageLog, err := req.ToMessageLog(webhookConfig)
	if err != nil {
		w.helper.Errorw("msg", "create message log failed", "error", err)
		return 0, merr.ErrorInternalServer("generate message log failed").WithCause(err)
	}
	uid, err := w.messageLogBiz.createMessageLog(ctx, messageLog)
	if err != nil {
		w.helper.Errorw("msg", "create message log failed", "error", err)
		return 0, merr.ErrorInternalServer("create message log failed").WithCause(err)
	}

	return uid, nil
}

func (w *Webhook) AppendWebhookMessageWithTemplate(ctx context.Context, req *bo.SendWebhookWithTemplateBo) (snowflake.ID, error) {
	// 获取模板
	templateDo, err := w.templateBiz.GetTemplate(ctx, req.TemplateUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, merr.ErrorParams("template not found")
		}
		w.helper.Errorw("msg", "get template failed", "error", err)
		return 0, merr.ErrorInternalServer("get template failed")
	}
	sendWebhookBo, err := req.ToSendWebhookBo(templateDo)
	if err != nil {
		w.helper.Errorw("msg", "convert template to webhook template data failed", "error", err)
		return 0, merr.ErrorInternalServer("convert template to webhook template data failed")
	}
	return w.AppendWebhookMessage(ctx, sendWebhookBo)
}
