package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

type AlertRecord interface {
	CreateAlertRecord(ctx context.Context, req *bo.CreateAlertRecordBo) (snowflake.ID, error)
	GetAlertRecord(ctx context.Context, uid snowflake.ID) (*bo.AlertRecordItemBo, error)
	ListAlertRecord(ctx context.Context, req *bo.ListAlertRecordBo) (*bo.PageResponseBo[*bo.AlertRecordItemBo], error)
}

type AlertSubscription interface {
	GetAlertSubscriptionByName(ctx context.Context, name string) (*bo.AlertSubscriptionItemBo, error)
	CreateAlertSubscription(ctx context.Context, req *bo.CreateAlertSubscriptionBo) (snowflake.ID, error)
	UpdateAlertSubscription(ctx context.Context, req *bo.UpdateAlertSubscriptionBo) error
	DeleteAlertSubscription(ctx context.Context, uid snowflake.ID) error
	GetAlertSubscription(ctx context.Context, uid snowflake.ID) (*bo.AlertSubscriptionItemBo, error)
	ListAlertSubscription(ctx context.Context, req *bo.ListAlertSubscriptionBo) (*bo.PageResponseBo[*bo.AlertSubscriptionItemBo], error)
	ListEnabledAlertSubscriptions(ctx context.Context) ([]*bo.AlertSubscriptionItemBo, error)
	UpdateAlertSubscriptionStatus(ctx context.Context, req *bo.UpdateAlertSubscriptionStatusBo) error
}
