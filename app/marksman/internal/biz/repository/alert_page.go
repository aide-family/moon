package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type AlertPage interface {
	AlertPageNameTaken(ctx context.Context, name string, excludeUID snowflake.ID) (bool, error)
	CreateAlertPage(ctx context.Context, req *bo.CreateAlertPageBo) (snowflake.ID, error)
	UpdateAlertPage(ctx context.Context, req *bo.UpdateAlertPageBo) error
	DeleteAlertPage(ctx context.Context, uid snowflake.ID) error
	GetAlertPage(ctx context.Context, uid snowflake.ID) (*bo.AlertPageItemBo, error)
	ListAlertPage(ctx context.Context, req *bo.ListAlertPageBo) (*bo.PageResponseBo[*bo.AlertPageItemBo], error)
	// CountAlertPagesByUIDs returns how many of the given UIDs exist in the current namespace.
	CountAlertPagesByUIDs(ctx context.Context, uids []snowflake.ID) (int64, error)
	// GetAlertPagesByUIDs returns alert pages for the given UIDs in current namespace (order not guaranteed).
	GetAlertPagesByUIDs(ctx context.Context, uids []snowflake.ID) ([]*bo.AlertPageItemBo, error)
}
