package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/jade_tree/internal/biz/bo"
)

// SSHCommand persists approved SSH command templates.
type SSHCommand interface {
	Create(ctx context.Context, in *bo.SSHCommandCreateRepoBo) (*bo.SSHCommandItemBo, error)
	Update(ctx context.Context, in *bo.SSHCommandUpdateRepoBo) error
	Get(ctx context.Context, uid snowflake.ID) (*bo.SSHCommandItemBo, error)
	List(ctx context.Context, req *bo.ListSSHCommandsBo) (*bo.PageResponseBo[*bo.SSHCommandItemBo], error)
	CountByName(ctx context.Context, in *bo.SSHCommandCountByNameBo) (int64, error)
}
