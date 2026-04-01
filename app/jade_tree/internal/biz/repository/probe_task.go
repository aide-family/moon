package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/jade_tree/internal/biz/bo"
)

type ProbeTask interface {
	Create(ctx context.Context, in *bo.CreateProbeTaskBo) (*bo.ProbeTaskItemBo, error)
	Update(ctx context.Context, in *bo.UpdateProbeTaskBo) (*bo.ProbeTaskItemBo, error)
	UpdateStatus(ctx context.Context, in *bo.UpdateProbeTaskStatusBo) (*bo.ProbeTaskItemBo, error)
	Delete(ctx context.Context, uid snowflake.ID) error
	Get(ctx context.Context, uid snowflake.ID) (*bo.ProbeTaskItemBo, error)
	List(ctx context.Context, req *bo.ListProbeTasksBo) (*bo.PageResponseBo[*bo.ProbeTaskItemBo], error)
	ListEnabled(ctx context.Context) ([]*bo.ProbeTaskItemBo, error)
}
