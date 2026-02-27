package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewWebhookConfig(
	webhookConfigRepo repository.WebhookConfig,
	helper *klog.Helper,
) *WebhookConfig {
	return &WebhookConfig{
		webhookConfigRepo: webhookConfigRepo,
		helper:            klog.NewHelper(klog.With(helper.Logger(), "biz", "webhookConfig")),
	}
}

type WebhookConfig struct {
	helper            *klog.Helper
	webhookConfigRepo repository.WebhookConfig
}

func (w *WebhookConfig) CreateWebhook(ctx context.Context, req *bo.CreateWebhookBo) (snowflake.ID, error) {
	if webhookConfig, err := w.webhookConfigRepo.GetWebhookConfigByName(ctx, req.Name); err == nil {
		return 0, merr.ErrorParams("webhook config %s already exists, uid: %s", req.Name, webhookConfig.UID)
	} else if !merr.IsNotFound(err) {
		w.helper.Errorw("msg", "check webhook config exists failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create webhook config %s failed", req.Name).WithCause(err)
	}
	uid, err := w.webhookConfigRepo.CreateWebhookConfig(ctx, req)
	if err != nil {
		w.helper.Errorw("msg", "create webhook config failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create webhook config %s failed", req.Name).WithCause(err)
	}
	return uid, nil
}

func (w *WebhookConfig) UpdateWebhook(ctx context.Context, req *bo.UpdateWebhookBo) error {
	existWebhookConfig, err := w.webhookConfigRepo.GetWebhookConfigByName(ctx, req.Name)
	if err != nil && !merr.IsNotFound(err) {
		w.helper.Errorw("msg", "check webhook config exists failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("update webhook config %s failed", req.Name).WithCause(err)
	} else if existWebhookConfig != nil && existWebhookConfig.UID != req.UID {
		return merr.ErrorParams("webhook config %s already exists", req.Name)
	}
	if err := w.webhookConfigRepo.UpdateWebhookConfig(ctx, req); err != nil {
		w.helper.Errorw("msg", "update webhook config failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update webhook config %s failed", req.UID).WithCause(err)
	}
	return nil
}

func (w *WebhookConfig) UpdateWebhookStatus(ctx context.Context, req *bo.UpdateWebhookStatusBo) error {
	if err := w.webhookConfigRepo.UpdateWebhookStatus(ctx, req); err != nil {
		w.helper.Errorw("msg", "update webhook status failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update webhook status %s failed", req.UID).WithCause(err)
	}
	return nil
}

func (w *WebhookConfig) DeleteWebhook(ctx context.Context, uid snowflake.ID) error {
	if err := w.webhookConfigRepo.DeleteWebhookConfig(ctx, uid); err != nil {
		w.helper.Errorw("msg", "delete webhook config failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete webhook config %s failed", uid).WithCause(err)
	}
	return nil
}

func (w *WebhookConfig) GetWebhook(ctx context.Context, uid snowflake.ID) (*bo.WebhookItemBo, error) {
	webhookConfigBo, err := w.webhookConfigRepo.GetWebhookConfig(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("webhook config %s not found", uid)
		}
		w.helper.Errorw("msg", "get webhook config failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get webhook config %s failed", uid).WithCause(err)
	}
	return webhookConfigBo, nil
}

func (w *WebhookConfig) ListWebhook(ctx context.Context, req *bo.ListWebhookBo) (*bo.PageResponseBo[*bo.WebhookItemBo], error) {
	pageResponseBo, err := w.webhookConfigRepo.ListWebhookConfig(ctx, req)
	if err != nil {
		w.helper.Errorw("msg", "list webhook config failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list webhook config failed").WithCause(err)
	}
	return pageResponseBo, nil
}

func (w *WebhookConfig) SelectWebhook(ctx context.Context, req *bo.SelectWebhookBo) (*bo.SelectWebhookBoResult, error) {
	result, err := w.webhookConfigRepo.SelectWebhookConfig(ctx, req)
	if err != nil {
		w.helper.Errorw("msg", "select webhook config failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select webhook config failed").WithCause(err)
	}
	return result, nil
}
