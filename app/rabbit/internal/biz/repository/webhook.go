package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

type WebhookConfig interface {
	CreateWebhookConfig(ctx context.Context, req *bo.CreateWebhookBo) (snowflake.ID, error)
	UpdateWebhookConfig(ctx context.Context, req *bo.UpdateWebhookBo) error
	UpdateWebhookStatus(ctx context.Context, req *bo.UpdateWebhookStatusBo) error
	DeleteWebhookConfig(ctx context.Context, uid snowflake.ID) error
	GetWebhookConfig(ctx context.Context, uid snowflake.ID) (*bo.WebhookItemBo, error)
	GetWebhookConfigByName(ctx context.Context, name string) (*bo.WebhookItemBo, error)
	ListWebhookConfig(ctx context.Context, req *bo.ListWebhookBo) (*bo.PageResponseBo[*bo.WebhookItemBo], error)
	SelectWebhookConfig(ctx context.Context, req *bo.SelectWebhookBo) (*bo.SelectWebhookBoResult, error)
}
