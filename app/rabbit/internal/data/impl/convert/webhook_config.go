package convert

import (
	"context"
	"maps"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func ToWebhookConfigDO(ctx context.Context, req *bo.CreateWebhookBo) *do.WebhookConfig {
	if pointer.IsNil(req.Headers) {
		req.Headers = make(map[string]string)
	}
	model := &do.WebhookConfig{
		App:          req.App,
		NamespaceUID: contextx.GetNamespace(ctx),
		Name:         req.Name,
		URL:          req.URL,
		Method:       req.Method,
		Headers:      safety.NewMap(maps.Clone(req.Headers)),
		Secret:       strutil.EncryptString(req.Secret),
		Status:       enum.GlobalStatus_ENABLED,
	}
	model.WithCreator(contextx.GetUserUID(ctx))
	return model
}

func ToWebhookConfigItemBo(webhookConfigDO *do.WebhookConfig) *bo.WebhookItemBo {
	if pointer.IsNil(webhookConfigDO.Headers) {
		webhookConfigDO.Headers = safety.NewMap(make(map[string]string))
	}
	return &bo.WebhookItemBo{
		UID:       webhookConfigDO.ID,
		App:       webhookConfigDO.App,
		Name:      webhookConfigDO.Name,
		URL:       webhookConfigDO.URL,
		Method:    webhookConfigDO.Method,
		Headers:   webhookConfigDO.Headers.Map(),
		Secret:    string(webhookConfigDO.Secret),
		Status:    webhookConfigDO.Status,
		CreatedAt: webhookConfigDO.CreatedAt,
		UpdatedAt: webhookConfigDO.UpdatedAt,
	}
}

func ToWebhookConfigItemSelectBo(webhookConfigDO *do.WebhookConfig) *bo.WebhookItemSelectBo {
	return &bo.WebhookItemSelectBo{
		UID:      webhookConfigDO.ID,
		Name:     webhookConfigDO.Name,
		Status:   webhookConfigDO.Status,
		Disabled: webhookConfigDO.Status == enum.GlobalStatus_DISABLED || webhookConfigDO.DeletedAt.Valid,
		Tooltip:  webhookConfigDO.Name,
		App:      webhookConfigDO.App,
	}
}
